package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestValidateAgentImagesLogsModes(t *testing.T) {
	chartDir, err := filepath.Abs("..")
	require.NoError(t, err)
	repoDir := filepath.Clean(filepath.Join(chartDir, "../.."))
	tmpDir := t.TempDir()
	chartCopy := filepath.Join(tmpDir, "suse-observability-agent")
	require.NoError(t, exec.Command("cp", "-a", chartDir, chartCopy).Run())
	valuesPath := filepath.Join(chartCopy, "values.yaml")
	rawValues, err := os.ReadFile(valuesPath)
	require.NoError(t, err)
	var values map[string]any
	require.NoError(t, yaml.Unmarshal(rawValues, &values))
	logsAgent := values["logsAgent"].(map[string]any)
	logsAgent["enabled"] = false
	promtail := logsAgent["image"].(map[string]any)
	promtail["repository"], promtail["tag"] = "stackstate/test-promtail", "promtail"
	otelLogs := values["otelLogsAgent"].(map[string]any)
	otelLogs["enabled"] = false
	filelog := otelLogs["image"].(map[string]any)
	filelog["repository"], filelog["tag"] = "stackstate/test-filelog", "filelog"

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "skopeo"), []byte(`#!/usr/bin/env bash
printf '%s\n' "${!#}" >> "${IMAGE_TEST_INSPECT_LOG}"
if [[ "${!#}" == "${IMAGE_TEST_MISSING_IMAGE:-}" ]]; then
  echo 'synthetic missing image' >&2
  exit 1
fi
`), 0700))
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	promtailImage := "docker://quay.io/stackstate/test-promtail:promtail"
	otelImage := "docker://quay.io/stackstate/test-filelog:filelog"

	for _, validator := range []string{"scripts/ci/validate_chart_images.sh", ".gitlab/validate_chart_images.sh"} {
		t.Run(validator, func(t *testing.T) {
			for _, mode := range []bool{false, true} {
				values["global"].(map[string]any)["features"].(map[string]any)["experimentalOtelLogsAgent"] = mode
				renderValues, err := yaml.Marshal(values)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(valuesPath, renderValues, 0600))
				for _, missingImage := range []string{"", otelImage} {
					name := "all images present"
					if missingImage != "" {
						name = "OTel image missing"
					}
					t.Run(strconv.FormatBool(mode)+"/"+name, func(t *testing.T) {
						inspectLog := filepath.Join(t.TempDir(), "inspected-images")
						t.Setenv("IMAGE_TEST_INSPECT_LOG", inspectLog)
						t.Setenv("IMAGE_TEST_MISSING_IMAGE", missingImage)
						output, err := exec.Command(filepath.Join(repoDir, validator), chartCopy).CombinedOutput()
						if missingImage == "" {
							require.NoError(t, err, "%s", output)
						} else {
							require.Error(t, err, "%s", output)
							require.Contains(t, string(output), "test-filelog:filelog")
							require.Contains(t, string(output), "synthetic missing image")
						}
						inspected, err := os.ReadFile(inspectLog)
						require.NoError(t, err)
						images := strings.Fields(string(inspected))
						require.Contains(t, images, promtailImage)
						require.Contains(t, images, otelImage)
						slices.Sort(images)
						require.Equal(t, images, slices.Compact(slices.Clone(images)), "duplicate inspections")
					})
				}
			}
		})
	}
}
