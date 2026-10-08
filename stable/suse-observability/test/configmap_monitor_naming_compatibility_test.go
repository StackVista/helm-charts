package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// The pre-change Git source is 6871f337366782b3757de14d244a928bef06ec96.
// Normal tests restore only the reviewed helper bodies in a disposable chart.
// The optional Git baseline check compares against an independently built copy.
func TestApplicationConfigMapNamingCompatibility(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates/_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	for _, component := range applicationConfigMapNamingComponents() {
		for _, suffix := range []string{"configmap", "log.configmap"} {
			helper := "stackstate." + component.key + "." + suffix + ".fullname"
			header := `{{- define "` + helper + `" -}}`
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(header) + `.*?\{\{- end -\}\}`)
			require.Len(t, pattern.FindAllString(content, -1), 1, helper)
			postfix := component.suffix
			if suffix == "log.configmap" {
				postfix += "-log"
			}
			content = pattern.ReplaceAllString(content, header+`{{ template "common.fullname.short" . }}-`+postfix+`{{- end -}}`)
		}
	}
	for _, resource := range []string{"servicemonitor", "podmonitor", "prometheusrule"} {
		// Parent definitions take precedence over packaged subchart helpers.
		content += "\n" + `{{- define "elasticsearch-exporter.` + resource + `.fullname" -}}{{ include "elasticsearch-exporter.fullname" . }}{{- end -}}` + "\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	checkApplicationConfigMapNamingCompatibility(t, chart)
}

func TestApplicationConfigMapNamingGitBaselineCompatibility(t *testing.T) {
	chart := os.Getenv("CONFIGMAP_MONITOR_BASELINE_CHART")
	if chart == "" {
		t.Skip("set CONFIGMAP_MONITOR_BASELINE_CHART to the pre-change chart with built local dependencies")
	}
	checkApplicationConfigMapNamingCompatibility(t, chart)
}

func applicationConfigMapNamingComponents() []struct{ key, suffix string } {
	return append([]struct{ key, suffix string }{{"api", "api"}}, explicitConfigurationComponents...)
}

func checkApplicationConfigMapNamingCompatibility(t *testing.T, beforeChart string) {
	t.Helper()
	const exporter = "elasticsearch.prometheus-elasticsearch-exporter."
	for _, tc := range []struct {
		name, release, prefix, namespace, exporterName, sizing string
		values                                                 map[string]string
	}{
		{name: "default", release: "suse-observability", prefix: "suse-observability"},
		{name: "nightly", release: "nightly", prefix: "nightly-suse-observability"},
		{name: "namespace", release: "nightly", prefix: "nightly-suse-observability", namespace: "tenant-b"},
		{name: "fullname", release: "nightly", prefix: "customer", values: map[string]string{"fullnameOverride": "customer"}},
		{name: "affixes", release: "nightly", prefix: "g-pre-customer-post-end", values: map[string]string{
			"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "global-only", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{"global.fullnameOverride": "global-only"}},
		{name: "long-name", release: "nightly", prefix: strings.Repeat("a", 54), values: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "monolithic", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{"stackstate.features.server.split": "false"}},
		{name: "split-workers", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{
			"stackstate.components.receiver.split.enabled": "true", "stackstate.components.correlate.split.enabled": "true",
		}},
		{name: "nonha", release: "nightly", prefix: "nightly-suse-observability", sizing: "global_sizing_50_nonha.yaml"},
		{name: "ha", release: "nightly", prefix: "nightly-suse-observability", sizing: "global_sizing_150_ha.yaml"},
		{name: "argo", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
		{name: "external-secrets", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{
			"stackstate.apiKey.fromExternalSecret": "customer-api-key", "stackstate.authentication.fromExternalSecret": "customer-auth",
			"stackstate.components.all.envsFromExistingSecrets[0].name":       "CUSTOMER_SETTING",
			"stackstate.components.all.envsFromExistingSecrets[0].secretName": "customer-environment",
			"stackstate.components.all.envsFromExistingSecrets[0].secretKey":  "setting",
		}},
		{name: "exporter-certificates-and-account", release: "nightly", prefix: "nightly-suse-observability", values: exporterNamingTLSValues()},
		{name: "authorization-disabled", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{"stackstate.k8sAuthorization.enabled": "false"}},
		{name: "exporter-fullname", release: "nightly", prefix: "nightly-suse-observability", exporterName: "customer-exporter",
			values: map[string]string{exporter + "fullnameOverride": "customer-exporter"}},
		{name: "exporter-name", release: "nightly", prefix: "nightly-suse-observability", exporterName: "nightly-customer",
			values: map[string]string{exporter + "nameOverride": "customer"}},
		{name: "pod-monitor", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{
			exporter + "serviceMonitor.enabled": "false", exporter + "podMonitor.enabled": "true", exporter + "podMonitor.namespace": "monitoring",
		}},
		{name: "monitor-namespace-and-metadata", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{
			exporter + "serviceMonitor.namespace": "monitoring", exporter + "prometheusRule.namespace": "monitoring",
			exporter + "serviceMonitor.labels.customer": "preserved", exporter + "serviceMonitor.interval": "37s",
			"configmap.annotations.customer": "preserved",
		}},
		{name: "monitoring-disabled", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{
			exporter + "serviceMonitor.enabled": "false", exporter + "prometheusRule.enabled": "false",
		}},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				values := componentResourceNameTestValues()
				values[exporter+"serviceMonitor.enabled"] = "true"
				values[exporter+"prometheusRule.enabled"] = "true"
				values[exporter+"prometheusRule.rules[0].alert"] = "FixtureAlert"
				values[exporter+"prometheusRule.rules[0].expr"] = "vector(1)"
				for key, value := range tc.values {
					values[key] = value
				}
				options := apiResourceNameTestOptions(values)
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				if tc.sizing != "" {
					options.ValuesFiles = append(options.ValuesFiles, filepath.Join("values", tc.sizing))
				}
				beforeOutput := renderRBACNamingChart(t, beforeChart, tc.release, options, upgrade)
				afterOutput := renderRBACNamingChart(t, "..", tc.release, options, upgrade)
				before := helmtestutil.NewKubernetesResources(t, beforeOutput)
				after := helmtestutil.NewKubernetesResources(t, afterOutput)
				renames := map[string]string{}
				for _, component := range applicationConfigMapNamingComponents() {
					for _, suffix := range []string{"", "-log"} {
						old, target := tc.prefix+"-"+component.suffix+suffix, "suse-observability-"+component.suffix+suffix
						if config, found := before.ConfigMaps[old]; found {
							require.Contains(t, after.ConfigMaps, target)
							config.Name = target
							assert.Equal(t, config, after.ConfigMaps[target], "Configuration content and metadata must stay unchanged")
							if old != target {
								renames["ConfigMap/"+old] = target
								assert.NotContains(t, after.ConfigMaps, old)
							}
						}
					}
				}
				exporterName := tc.exporterName
				if exporterName == "" {
					exporterName = tc.release + "-prometheus-elasticsearch-exporter"
				}
				for _, kind := range []string{"ServiceMonitor", "PodMonitor", "PrometheusRule"} {
					target := "suse-observability-prometheus-elasticsearch-exporter"
					if exporterName != target {
						renames[kind+"/"+exporterName] = target
					}
				}
				seen := map[string]bool{}
				expected := backupNamingDocuments(t, beforeOutput, renames, seen)
				actual := backupNamingDocuments(t, afterOutput, nil, nil)
				for name, deployment := range before.Deployments {
					require.Contains(t, after.Deployments, name)
					target := after.Deployments[name]
					allowed := deployment.DeepCopy()
					for i, volume := range allowed.Spec.Template.Spec.Volumes {
						if volume.ConfigMap != nil {
							oldName := volume.ConfigMap.Name
							if replacement, found := renames["ConfigMap/"+oldName]; found {
								allowed.Spec.Template.Spec.Volumes[i].ConfigMap.Name = replacement
								for _, component := range applicationConfigMapNamingComponents() {
									if oldName == tc.prefix+"-"+component.suffix {
										key := "checksum/" + component.key + "-configmap"
										oldHash, newHash := deployment.Spec.Template.Annotations[key], target.Spec.Template.Annotations[key]
										require.Regexp(t, "^[0-9a-f]{64}$", oldHash)
										require.Regexp(t, "^[0-9a-f]{64}$", newHash)
										assert.NotEqual(t, oldHash, newHash, "Renaming application configuration must update its checksum")
										allowed.Spec.Template.Annotations[key] = newHash
										doc := expected["Deployment/"+deployment.Namespace+"/"+name].([]interface{})[0].(map[string]interface{})
										template := doc["spec"].(map[string]interface{})["template"].(map[string]interface{})
										template["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})[key] = newHash
									}
								}
							}
						}
					}
					assert.Equal(t, *allowed, target, "Only ConfigMap volume names and their checksum may change")
					assertApplicationConfigMapUpgradePatches(t, deployment, target)
				}
				assert.Equal(t, expected, actual, "Only the reviewed ConfigMaps, volume references, checksums and exporter monitor names may change")
				for old := range renames {
					if strings.HasPrefix(old, "ConfigMap/") {
						assert.True(t, seen[old], old)
					}
				}
				assert.Equal(t, before.Statefulsets, after.Statefulsets)
				assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
				assert.Equal(t, before.Secrets, after.Secrets)
				assert.Equal(t, before.Services, after.Services)
				assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
			})
		}
	}
}

func assertApplicationConfigMapUpgradePatches(t *testing.T, before, after appsv1.Deployment) {
	t.Helper()
	for _, pair := range [][2]appsv1.Deployment{{before, after}, {after, after}, {after, before}} {
		live := pair[0].DeepCopy()
		if live.Annotations == nil {
			live.Annotations = map[string]string{}
		}
		live.Annotations["fixture.customer"] = "preserved"
		live.ResourceVersion = "fixture-version"
		original, err := json.Marshal(pair[0])
		require.NoError(t, err)
		modified, err := json.Marshal(pair[1])
		require.NoError(t, err)
		current, err := json.Marshal(live)
		require.NoError(t, err)
		metadata, err := strategicpatch.NewPatchMetaFromStruct(appsv1.Deployment{})
		require.NoError(t, err)
		patch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, current, metadata, true)
		require.NoError(t, err)
		applied, err := strategicpatch.StrategicMergePatch(current, patch, appsv1.Deployment{})
		require.NoError(t, err)
		expected := pair[1].DeepCopy()
		if expected.Annotations == nil {
			expected.Annotations = map[string]string{}
		}
		expected.Annotations["fixture.customer"] = "preserved"
		expected.ResourceVersion = "fixture-version"
		expectedJSON, err := json.Marshal(expected)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), string(applied), "Upgrade/repeat/rollback patch must preserve volume entries and live metadata")
	}
}
