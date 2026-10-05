package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestAnomalyResourceReferencesPreservePlatformConfiguration(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	require.NoError(t, os.WriteFile(filepath.Join(chart, "templates", "_anomaly-resource-names.tpl"), []byte(`
{{- define "anomaly-detection.manager.service.fullname" -}}explicit-anomaly-manager{{- end -}}
{{- define "anomaly-detection.serviceaccount.fullname" -}}explicit-anomaly-account{{- end -}}
{{- define "anomaly-detection.authentication.role.fullname" -}}explicit-anomaly-role{{- end -}}
`), 0600))
	values, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, enabled := range []bool{false, true} {
		for _, split := range []bool{false, true} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("enabled=%t/split=%t/upgrade=%t", enabled, split, upgrade), func(t *testing.T) {
					options := apiResourceNameTestOptions(map[string]string{
						"anomaly-detection.enabled":                        fmt.Sprint(enabled),
						"anomaly-detection.metrics.serviceMonitor.enabled": "true",
						"anomaly-detection.stackstate.instance":            "https://unused-customer-instance",
						"stackstate.features.server.split":                 fmt.Sprint(split),
					})
					options.ValuesFiles = []string{values}
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					before := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "nightly", options, args...)
					after, err := helm.RenderTemplateE(t, options, chart, "nightly", nil, args...)
					require.NoError(t, err)
					const base = "nightly-anomaly-detection"
					seen := map[string]bool{}
					expected := backupNamingDocuments(t, before, map[string]string{
						"Service/" + base + "-spotlight-manager": "explicit-anomaly-manager",
						"ServiceAccount/" + base + "-sa":         "explicit-anomaly-account",
						"Role/stackstate-aad":                    "explicit-anomaly-role",
					}, seen)
					require.Equal(t, enabled, seen["Role/stackstate-aad"])
					require.Equal(t, enabled, seen["Service/"+base+"-spotlight-manager"])
					require.Equal(t, enabled, seen["ServiceAccount/"+base+"-sa"])
					rewritePlatformAnomalyReferences(expected)
					// Platform router URLs, staticSubjects.stackstate-aad, account
					// annotations, selectors, monitors, checksums and PVCs stay equal.
					assert.Equal(t, expected, backupNamingDocuments(t, after, nil, nil))
				})
			}
		}
	}
}

func rewritePlatformAnomalyReferences(value interface{}) {
	switch node := value.(type) {
	case map[string]interface{}:
		if node["serviceAccountName"] == "nightly-anomaly-detection-sa" {
			node["serviceAccountName"] = "explicit-anomaly-account"
		}
		if subjects, ok := node["subjects"].([]interface{}); ok {
			for _, subject := range subjects {
				ref := subject.(map[string]interface{})
				if ref["kind"] == "ServiceAccount" && ref["name"] == "nightly-anomaly-detection-sa" {
					ref["name"] = "explicit-anomaly-account"
				}
			}
		}
		if ref, ok := node["roleRef"].(map[string]interface{}); ok {
			if ref["kind"] == "Role" && ref["name"] == "stackstate-aad" {
				ref["name"] = "explicit-anomaly-role"
			}
		}
		if args, ok := node["args"].([]interface{}); ok {
			for i, arg := range args {
				if arg == "--manager-host" && i+1 < len(args) && args[i+1] == "nightly-anomaly-detection-spotlight-manager" {
					args[i+1] = "explicit-anomaly-manager"
				}
			}
		}
		for _, child := range node {
			rewritePlatformAnomalyReferences(child)
		}
	case []interface{}:
		for _, child := range node {
			rewritePlatformAnomalyReferences(child)
		}
	}
}
