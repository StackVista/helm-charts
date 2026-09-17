package test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetImages(t *testing.T) {
	curDir, err := os.Getwd()
	require.NoError(t, err)
	images := strings.Split(RunGetImagesScript(t, filepath.Join(curDir, "..")), "\n")
	require.Equal(t, 37, len(
		slices.DeleteFunc(
			images,
			func(e string) bool {
				return e == ""
			}),
	), images)
}

func TestGetImagesIncludesReplicationChecker(t *testing.T) {
	image := "registry.example.com/stackstate/replication-checker:test"
	chartDir := chartWithDistinctCheckerImage(t, image)
	require.Contains(t, strings.Split(RunGetImagesScript(t, chartDir), "\n"), image)
}

func chartWithDistinctCheckerImage(t *testing.T, image string) string {
	t.Helper()
	tmpDir := t.TempDir()
	require.NoError(t, exec.Command("cp", "-a", "..", filepath.Join(tmpDir, "chart")).Run())
	chartDir := filepath.Join(tmpDir, "chart")
	templatePath := filepath.Join(chartDir, "templates", "replication-checker", "deployment-replication-checker.yaml")
	template, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	// A distinct image prevents other container-tools users from masking an omitted checker.
	imageLine := regexp.MustCompile(`(?m)^image:.*$`)
	require.Len(t, imageLine.FindAll(template, -1), 1)
	require.NoError(t, os.WriteFile(templatePath, imageLine.ReplaceAll(template, []byte("image: "+image)), 0o644))

	return chartDir
}

func TestCopyImages(t *testing.T) {
	image := "registry.example.com/stackstate/replication-checker:test"
	chartDir := chartWithDistinctCheckerImage(t, image)
	realHelm, err := exec.LookPath("helm")
	require.NoError(t, err)

	for _, tc := range []struct {
		name       string
		repository string
		failRender bool
	}{
		{name: "default repository"},
		{name: "repository override", repository: "https://charts.example.com"},
		{name: "render failure", failRender: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			binDir := t.TempDir()
			argsFile := filepath.Join(binDir, "helm-args")
			require.NoError(t, os.WriteFile(filepath.Join(binDir, "docker"), []byte("#!/bin/bash\nexit 0\n"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(binDir, "helm"), []byte(`#!/bin/bash
printf '%s\n' "$@" > "$COPY_TEST_HELM_ARGS"
if [[ "$COPY_TEST_FAIL_RENDER" == true ]]; then
  echo "chart lookup failed" >&2
  exit 1
fi
args=("$@")
exec "$COPY_TEST_REAL_HELM" "${args[@]:0:2}" "$COPY_TEST_CHART" "${args[@]:5}"
`), 0o755))
			t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("COPY_TEST_HELM_ARGS", argsFile)
			t.Setenv("COPY_TEST_CHART", chartDir)
			t.Setenv("COPY_TEST_REAL_HELM", realHelm)
			t.Setenv("COPY_TEST_FAIL_RENDER", fmt.Sprint(tc.failRender))
			t.Setenv("STS_REGISTRY_USERNAME", "test")
			t.Setenv("STS_REGISTRY_PASSWORD", "test")
			t.Setenv("DST_REGISTRY_USERNAME", "")
			t.Setenv("DST_REGISTRY_PASSWORD", "")

			args := []string{"-d", "mirror.example.com", "-t"}
			repository := "https://charts.rancher.com/server-charts/prime/suse-observability"
			if tc.repository != "" {
				args = append(args, "-r", tc.repository)
				repository = tc.repository
			}
			output, err := exec.Command(filepath.Join(chartDir, "installation", "copy_images.sh"), args...).CombinedOutput()
			invocation, readErr := os.ReadFile(argsFile)
			require.NoError(t, readErr)
			helmArgs := strings.Split(strings.TrimSpace(string(invocation)), "\n")
			require.GreaterOrEqual(t, len(helmArgs), 7)
			require.Equal(t, []string{"template", "suse-observability", "suse-observability", "--repo", repository, "--set"}, helmArgs[:6])
			if tc.failRender {
				require.Error(t, err)
				require.Contains(t, string(output), "chart lookup failed")
				require.NotContains(t, string(output), "Copying ")
				return
			}
			require.NoError(t, err, string(output))
			require.Contains(t, string(output), "Copying "+image+" to mirror.example.com/stackstate/replication-checker:test (dry-run)")
		})
	}
}

func TestChangeImageSourceScriptRewritesImages(t *testing.T) {
	curDir, err := os.Getwd()
	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp("/tmp", "install-scripts")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = exec.Command("cp", "-a", filepath.Join(curDir, ".."), tmpDir).Run()
	require.NoError(t, err)

	helmDir := filepath.Join(tmpDir, "suse-observability")

	targetReg := "reg"
	targetRepo := "repo"
	currentImages := RunGetImagesScript(t, helmDir)
	expectedImages := strings.ReplaceAll(currentImages, "quay.io/stackstate", fmt.Sprintf("%s/%s", targetReg, targetRepo))
	RunChangeImageScript(t, helmDir, targetReg, targetRepo)
	renamedImages := RunGetImagesScript(t, helmDir)

	require.NotContains(t, renamedImages, fmt.Sprintf("%s/stackstate/", targetReg))
	require.Equal(t, expectedImages, renamedImages)
}

func RunGetImagesScript(t *testing.T, dir string) string {
	cmd := exec.Command(filepath.Join(dir, "installation", "o11y-get-images.sh"))
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
