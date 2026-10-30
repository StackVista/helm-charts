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
	for _, source := range []string{"", "source", "source/team/nested"} {
		for _, destination := range []string{"repo", "mirror/team"} {
			t.Run(fmt.Sprintf("source=%s/destination=%s", source, destination), func(t *testing.T) {
				testChangeImageSourceScriptRewritesImages(t, source, destination)
			})
		}
	}
}

func testChangeImageSourceScriptRewritesImages(t *testing.T, source, targetRepo string) {
	t.Helper()
	curDir, err := os.Getwd()
	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp("/tmp", "install-scripts")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = exec.Command("cp", "-a", filepath.Join(curDir, ".."), tmpDir).Run()
	require.NoError(t, err)

	helmDir := filepath.Join(tmpDir, "suse-observability")

	repositoryPaths := []string{
		".global.wait.image.repository",
		".stackstate.components.api.image.repository",
		".stackstate.components.checks.image.repository",
		".stackstate.components.correlate.image.repository",
		".stackstate.components.authorizationSync.image.repository",
		".stackstate.components.healthSync.image.repository",
		".stackstate.components.initializer.image.repository",
		".stackstate.components.mcp.image.repository",
		".stackstate.components.aiAssistant.image.repository",
		".stackstate.components.e2es.image.repository",
		".stackstate.components.notification.image.repository",
		".stackstate.components.receiver.image.repository",
		".stackstate.components.router.image.repository",
		".stackstate.components.router.mode.image.repository",
		".stackstate.components.slicing.image.repository",
		".stackstate.components.state.image.repository",
		".stackstate.components.sync.image.repository",
		".stackstate.components.server.image.repository",
		".stackstate.components.ui.image.repository",
		".stackstate.components.clickhouseCleanup.image.repository",
		".stackstate.components.kafkaTopicCreate.image.repository",
		".stackstate.components.nginxPrometheusExporter.image.repository",
		".stackstate.components.containerTools.image.repository",
		".stackstate.components.vmagent.image.repository",
		".stackstate.components.workloadObserver.image.repository",
		".stackstate.stackpacks.images[0].repository",
		".stackstate.stackpacks.images[1].repository",
		".stackstate.stackpacks.images[2].repository",
		".elasticsearch.imageRepository",
		".elasticsearch.prometheus-elasticsearch-exporter.image.repository",
		".hbase.stackgraph.image.repository",
		".hbase.console.image.repository",
		".hbase.wait.image.repository",
		".hbase.hbase.master.image.repository",
		".hbase.hbase.regionserver.image.repository",
		".hbase.hdfs.image.repository",
		".hbase.tephra.image.repository",
		".hbase.zookeeper.image.repository",
		".kafkaup-operator.image.repository",
		".kafka.image.repository",
		".kafka.metrics.jmx.image.repository",
		".s3proxy.image.repository",
		".zookeeper.image.repository",
		".anomaly-detection.image.spotlightRepository",
		".victoria-metrics-0.server.image.repository",
		".victoria-metrics-1.server.image.repository",
		".victoria-metrics-0.backup.vmbackup.image.repository",
		".victoria-metrics-1.backup.vmbackup.image.repository",
		".victoria-metrics-0.backup.setupCron.image.repository",
		".victoria-metrics-1.backup.setupCron.image.repository",
		".victoria-metrics.restore.image.repository",
		".clickhouse.image.repository",
		".clickhouse.backup.image.repository",
		".opentelemetry-collector.image.repository",
		".kubernetes-rbac-agent.containers.rbacAgent.image.repository",
	}
	if source != "" {
		var overrides []string
		for i, path := range repositoryPaths {
			overrides = append(overrides, fmt.Sprintf(`%s = "%s/custom-%d"`, path, source, i))
		}
		output, err := exec.Command("yq", "-i", strings.Join(overrides, " | "), filepath.Join(helmDir, "values.yaml")).CombinedOutput()
		require.NoError(t, err, "%s", output)
	}

	targetReg := "mirror.example.com:5000"
	currentImages := RunGetImagesScript(t, helmDir)
	expectedImages := strings.Split(strings.TrimSpace(currentImages), "\n")
	for i, image := range expectedImages {
		expectedImages[i] = targetReg + "/" + targetRepo + "/" + image[strings.LastIndex(image, "/")+1:]
	}
	slices.Sort(expectedImages)
	RunChangeImageScript(t, helmDir, targetReg, targetRepo)
	renamedImages := strings.Split(strings.TrimSpace(RunGetImagesScript(t, helmDir)), "\n")

	require.Equal(t, expectedImages, renamedImages)
	if source != "" {
		output, err := exec.Command("yq", "["+strings.Join(repositoryPaths, ", ")+"] | .[]", filepath.Join(helmDir, "values.yaml")).CombinedOutput()
		require.NoError(t, err, "%s", output)
		repositories := strings.Split(strings.TrimSpace(string(output)), "\n")
		require.Len(t, repositories, len(repositoryPaths))
		for i, repository := range repositories {
			require.Equal(t, fmt.Sprintf("%s/custom-%d", targetRepo, i), repository, repositoryPaths[i])
		}
	}
}

func RunGetImagesScript(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command(filepath.Join(dir, "installation", "o11y-get-images.sh"))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()

	require.NoError(t, err, "%s", stderr.String())
	require.NotEmpty(t, strings.TrimSpace(string(stdout)), "%s", stderr.String())
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
