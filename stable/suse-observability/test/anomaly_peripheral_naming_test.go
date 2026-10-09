package test

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// Restore only the two legacy expressions in a temporary chart to compare the
// entire platform. Frozen subchart snapshots independently protect this delta.
func TestAnomalyPeripheralNamingPreservesPlatformResources(t *testing.T) {
	beforeChart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(beforeChart, os.DirFS("..")))
	require.NoError(t, os.WriteFile(filepath.Join(beforeChart, "templates", "_anomaly-peripheral-baseline.tpl"), []byte(`
{{- define "anomaly-detection.pdb.fullname" -}}{{ template "common.fullname.short" . }}-anomaly-detection{{- end -}}
{{- define "anomaly-detection.manager.servicemonitor.fullname" -}}{{ template "common.fullname.short" . }}-spotlight-manager{{- end -}}
`), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	afterChart, err := filepath.Abs("..")
	require.NoError(t, err)
	for _, tc := range []struct {
		name, release, namespace, base, sizing string
		values                                 map[string]string
	}{
		{name: "product-release", release: "suse-observability", base: "suse-observability-anomaly-detection"},
		{name: "custom-release", release: "nightly", base: "nightly-anomaly-detection"},
		{name: "same-release-other-namespace", release: "suse-observability", namespace: "tenant-b", base: "suse-observability-anomaly-detection"},
		{name: "root-override", release: "nightly", base: "nightly-anomaly-detection", values: map[string]string{"fullnameOverride": "customer-platform"}},
		{name: "subchart-affixes", release: "nightly", base: "g-pre-customer-post-end", values: map[string]string{
			"anomaly-detection.fullnameOverride": "customer",
			"anomaly-detection.fullnamePrefix":   "pre-", "anomaly-detection.fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "long-name", release: "nightly", base: strings.Repeat("a", 54), values: map[string]string{"anomaly-detection.fullnameOverride": strings.Repeat("a", 70)}},
		{name: "monolithic", release: "nightly", base: "nightly-anomaly-detection", values: map[string]string{"stackstate.features.server.split": "false"}},
		{name: "ha", release: "nightly", base: "nightly-anomaly-detection", sizing: "global_sizing_150_ha.yaml"},
		{name: "nonha", release: "nightly", base: "nightly-anomaly-detection", sizing: "global_sizing_50_nonha.yaml"},
		{name: "argo", release: "nightly", base: "nightly-anomaly-detection", values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
		{name: "monitor-disabled", release: "nightly", base: "nightly-anomaly-detection", values: map[string]string{"anomaly-detection.metrics.serviceMonitor.enabled": "false"}},
		{name: "component-disabled", release: "nightly", base: "nightly-anomaly-detection", values: map[string]string{"anomaly-detection.enabled": "false"}},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				values := map[string]string{
					"anomaly-detection.enabled": "true", "anomaly-detection.metrics.serviceMonitor.enabled": "true",
					"anomaly-detection.manager.persistentStorage.size":         "17Gi",
					"anomaly-detection.manager.persistentStorage.storageClass": "customer-artifacts",
				}
				maps.Copy(values, tc.values)
				options := apiResourceNameTestOptions(values)
				options.ValuesFiles = []string{valuesFile}
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				if tc.sizing != "" {
					sizing, err := filepath.Abs(filepath.Join("values", tc.sizing))
					require.NoError(t, err)
					options.ValuesFiles = append(options.ValuesFiles, sizing)
				}
				args := []string{"--api-versions", "policy/v1/PodDisruptionBudget"}
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				beforeOutput, err := helm.RenderTemplateE(t, options, beforeChart, tc.release, nil, args...)
				require.NoError(t, err)
				afterOutput, err := helm.RenderTemplateE(t, options, afterChart, tc.release, nil, args...)
				require.NoError(t, err)
				renames := map[string]string{}
				if values["anomaly-detection.enabled"] == "true" {
					renames["PodDisruptionBudget/"+tc.base+"-anomaly-detection"] = "suse-observability-anomaly-detection"
					if values["anomaly-detection.metrics.serviceMonitor.enabled"] == "true" {
						renames["ServiceMonitor/"+tc.base+"-spotlight-manager"] = "suse-observability-spotlight-manager"
					}
				}
				seen := map[string]bool{}
				before := backupNamingDocuments(t, beforeOutput, renames, seen)
				after := backupNamingDocuments(t, afterOutput, nil, nil)
				assert.Equal(t, before, after, "Only anomaly detection's PDB and monitor metadata names may change")
				beforeContract := resourceNamingUpgradeContract(t, beforeOutput)
				for old, name := range renames {
					require.True(t, seen[old], "expected resource missing: %s", old)
					require.Contains(t, beforeContract, old)
					kind := strings.SplitN(old, "/", 2)[0]
					fields := beforeContract[old]
					delete(beforeContract, old)
					require.NotContains(t, beforeContract, kind+"/"+name)
					beforeContract[kind+"/"+name] = fields
				}
				assert.Equal(t, beforeContract, resourceNamingUpgradeContract(t, afterOutput))
				resources := helmtestutil.NewKubernetesResources(t, afterOutput)
				if values["anomaly-detection.enabled"] == "true" {
					require.Contains(t, resources.Pdbs, "suse-observability-anomaly-detection")
					require.Contains(t, resources.PersistentVolumeClaims, "spotlight-artifacts-volume-claim")
				} else {
					assert.NotContains(t, resources.Pdbs, "suse-observability-anomaly-detection")
					assert.NotContains(t, resources.PersistentVolumeClaims, "spotlight-artifacts-volume-claim")
				}
				if _, ok := renames["ServiceMonitor/"+tc.base+"-spotlight-manager"]; ok {
					require.Contains(t, resources.ServiceMonitors, "suse-observability-spotlight-manager")
				} else {
					assert.NotContains(t, resources.ServiceMonitors, "suse-observability-spotlight-manager")
				}
			})
		}
	}
}
