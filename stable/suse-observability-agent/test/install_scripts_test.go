package test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetImages(t *testing.T) {
	curDir, err := os.Getwd()
	require.NoError(t, err)
	images := strings.Split(RunGetImagesScript(t, filepath.Join(curDir, "..")), "\n")
	images = slices.DeleteFunc(images, func(e string) bool { return e == "" })
	require.GreaterOrEqual(t, len(images), 11, images)
	require.True(t, slices.IsSorted(images), images)
	require.Equal(t, images, slices.Compact(slices.Clone(images)), "duplicate images")
}

func TestInstallationImagesLogsModes(t *testing.T) {
	chartDir, err := filepath.Abs("..")
	require.NoError(t, err)
	for _, tc := range []struct {
		name         string
		promtailRepo string
		promtailTag  string
		otelRepo     string
		otelTag      string
	}{
		{"different repositories", "stackstate/test-promtail", "test", "stackstate/test-filelog", "test"},
		{"different tags", "stackstate/test-logs", "promtail", "stackstate/test-logs", "filelog"},
		{"same image", "stackstate/test-logs", "shared", "stackstate/test-logs", "shared"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, inputMode := range []string{"false", "true"} {
				t.Run(inputMode, func(t *testing.T) {
					valuesFile := filepath.Join(t.TempDir(), "custom values.yaml")
					values := fmt.Sprintf(`otel:
  enabled: false
global:
  features:
    experimentalOtelLogsAgent: %s
logsAgent:
  enabled: false
  image:
    repository: %s
    tag: %q
otelLogsAgent:
  enabled: false
  image:
    repository: %s
    tag: %q
`, inputMode, tc.promtailRepo, tc.promtailTag, tc.otelRepo, tc.otelTag)
					require.NoError(t, os.WriteFile(valuesFile, []byte(values), 0600))
					expected := []string{
						"quay.io/" + tc.promtailRepo + ":" + tc.promtailTag,
						"quay.io/" + tc.otelRepo + ":" + tc.otelTag,
					}
					for _, script := range installationImageScripts(chartDir) {
						t.Run(script.name, func(t *testing.T) {
							args := append(slices.Clone(script.args), "--", "-f", valuesFile, "--set", "otel.enabled=false,logsAgent.enabled=false,otelLogsAgent.enabled=false,global.features.experimentalOtelLogsAgent="+inputMode)
							images := runInstallationImageScript(t, chartDir, script.name, args...)
							for _, image := range expected {
								require.Contains(t, images, image)
							}
							require.Equal(t, images, slices.Compact(slices.Clone(images)), "duplicate images")
						})
					}
				})
			}
		})
	}
}

func TestInstallationImagesChartArchive(t *testing.T) {
	chartDir, err := filepath.Abs("..")
	require.NoError(t, err)
	archiveDir := t.TempDir()
	output, err := exec.Command("helm", "package", chartDir, "--destination", archiveDir).CombinedOutput()
	require.NoError(t, err, "%s", output)
	archives, err := filepath.Glob(filepath.Join(archiveDir, "*.tgz"))
	require.NoError(t, err)
	require.Len(t, archives, 1)

	expected := runInstallationImageScript(t, chartDir, "o11y-agent-get-images.sh", "-d", chartDir)
	for _, script := range installationImageScripts(archives[0]) {
		t.Run(script.name, func(t *testing.T) {
			if script.name == "o11y-agent-get-images.sh" {
				script.args = []string{"-f", archives[0]}
			}
			require.Equal(t, expected, runInstallationImageScript(t, chartDir, script.name, script.args...))
		})
	}
}

func TestInstallationImagesRenderFailure(t *testing.T) {
	chartDir, err := filepath.Abs("..")
	require.NoError(t, err)
	binDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "helm"), []byte(`#!/usr/bin/env bash
for arg in "$@"; do
  if [[ "$arg" == "global.features.experimentalOtelLogsAgent=${IMAGE_TEST_FAIL_MODE}" ]]; then
    echo "synthetic render failure" >&2
    exit 1
  fi
done
echo 'image: quay.io/stackstate/test-image:test'
`), 0700))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, mode := range []string{"false", "true"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("IMAGE_TEST_FAIL_MODE", mode)
			for _, script := range installationImageScripts(chartDir) {
				t.Run(script.name, func(t *testing.T) {
					cmd := exec.Command(filepath.Join(chartDir, "installation", script.name), script.args...)
					var stdout, stderr bytes.Buffer
					cmd.Stdout, cmd.Stderr = &stdout, &stderr
					require.Error(t, cmd.Run())
					require.Empty(t, stdout.String(), "must not emit a partial image list or start copying")
					require.Contains(t, stderr.String(), "synthetic render failure")
					require.Contains(t, stderr.String(), "global.features.experimentalOtelLogsAgent="+mode)
				})
			}
		})
	}
}

func installationImageScripts(chart string) []struct {
	name string
	args []string
} {
	return []struct {
		name string
		args []string
	}{
		{"o11y-agent-get-images.sh", []string{"-d", chart}},
		{"copy_images.sh", []string{"-c", chart, "-d", "registry.example.com", "-t"}},
		{"backup.sh", []string{"-c", chart, "-t"}},
	}
}

func runInstallationImageScript(t *testing.T, chartDir, name string, args ...string) []string {
	t.Helper()
	cmd := exec.Command(filepath.Join(chartDir, "installation", name), args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	require.NoError(t, err, "%s", stderr.String())
	var images []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		switch {
		case name == "o11y-agent-get-images.sh":
			images = append(images, line)
		case strings.HasPrefix(line, "Copying "):
			images = append(images, strings.Fields(line)[1])
		case strings.HasPrefix(line, "Backing up "):
			images = append(images, strings.Fields(line)[2])
		}
	}
	require.NotEmpty(t, images)
	require.True(t, slices.IsSorted(images), images)
	return images
}

func TestChangeImageRepo(t *testing.T) {
	curDir, err := os.Getwd()
	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp("/tmp", "install-scripts")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = exec.Command("cp", "-a", filepath.Join(curDir, ".."), tmpDir).Run()
	require.NoError(t, err)

	helmDir := filepath.Join(tmpDir, "suse-observability-agent")

	targetReg := "reg"
	targetRepo := "repo"
	currentImages := RunGetImagesScript(t, helmDir)
	expectedImages := strings.ReplaceAll(currentImages, "quay.io/stackstate", fmt.Sprintf("%s/%s", targetReg, targetRepo))
	RunChangeImageScript(t, helmDir, targetReg, targetRepo)
	renamedImages := RunGetImagesScript(t, helmDir)
	require.Equal(t, expectedImages, renamedImages)
}

func RunGetImagesScript(t *testing.T, dir string) string {
	cmd := exec.Command(filepath.Join(dir, "installation", "o11y-agent-get-images.sh"))
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	stdout, err := cmd.Output()

	require.NoError(t, err)
	return string(stdout)
}

func RunChangeImageScript(t *testing.T, dir, registry, repository string) {
	cmd := exec.Command(filepath.Join(dir, "maintenance", "change-image-source.sh"), "-g", registry, "-p", repository)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Error running change-image-source.sh: %v\nstderr: %s\nstdout: %s", err, errb.String(), outb.String())
	}
	require.NoError(t, err)
}
