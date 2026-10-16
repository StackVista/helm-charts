package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

// Before this step: 2d7562ce7006d406969ce05a56caa165d410c1c9.
// The independent-step check uses the extraction baseline
// c851a02ae1f2f7eb64c5538a148194ea36eee4b4, without intervening renames.
func TestCollectorPeripheralNamingCompatibility(t *testing.T) {
	beforeChart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(beforeChart, os.DirFS("..")))
	path := filepath.Join(beforeChart, "templates/_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	for _, resource := range collectorPeripheralNamingInventory() {
		pattern := collectorPeripheralNamingDefinition(resource.helper)
		require.Len(t, pattern.FindAllString(content, -1), 1)
		content = pattern.ReplaceAllString(content, "")
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	checkCollectorPeripheralNaming(t, beforeChart, "..")
}

func TestCollectorPeripheralNamingGitBaselineCompatibility(t *testing.T) {
	beforeChart := os.Getenv("COLLECTOR_PERIPHERAL_BASELINE_CHART")
	if beforeChart == "" {
		t.Skip("set COLLECTOR_PERIPHERAL_BASELINE_CHART to the built pre-change chart")
	}
	// Apply only this step's current helpers to the actual Git baseline.
	// Comparing directly with ".." would include unrelated later changes.
	checkCollectorPeripheralNaming(t, beforeChart, collectorPeripheralNamingStageChart(t, beforeChart))
}

func TestCollectorPeripheralNamingIndependentOfPreviousSteps(t *testing.T) {
	beforeChart := os.Getenv("COLLECTOR_PERIPHERAL_EXTRACTION_CHART")
	if beforeChart == "" {
		t.Skip("set COLLECTOR_PERIPHERAL_EXTRACTION_CHART to the built extraction baseline")
	}
	checkCollectorPeripheralNaming(t, beforeChart, collectorPeripheralNamingStageChart(t, beforeChart))
}

// Freeze the supplied chart and its dependency graph, adding only the five
// current peripheral name definitions. The ordinary compatibility test checks
// their wiring against the current chart; historical checks isolate this step.
func collectorPeripheralNamingStageChart(t *testing.T, beforeChart string) string {
	t.Helper()
	afterChart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(afterChart, os.DirFS(beforeChart)))
	current, err := os.ReadFile("../templates/_names.tpl")
	require.NoError(t, err)
	path := filepath.Join(afterChart, "templates/_names.tpl")
	original, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(original)
	for _, resource := range collectorPeripheralNamingInventory() {
		pattern := collectorPeripheralNamingDefinition(resource.helper)
		require.Empty(t, pattern.FindAllString(content, -1), "This step must not rely on an earlier parent override")
		definitions := pattern.FindAllString(string(current), -1)
		require.Len(t, definitions, 1)
		content += "\n" + definitions[0] + "\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	return afterChart
}

func collectorPeripheralNamingInventory() []struct{ helper, kind, suffix string } {
	return []struct{ helper, kind, suffix string }{
		{"poddisruptionbudget", "PodDisruptionBudget", ""},
		{"servicemonitor", "ServiceMonitor", ""},
		{"podmonitor", "PodMonitor", "-agent"},
		{"prometheusrule", "PrometheusRule", ""},
		{"networkpolicy", "NetworkPolicy", ""},
	}
}

func collectorPeripheralNamingDefinition(resource string) *regexp.Regexp {
	header := `{{- define "opentelemetry-collector.` + resource + `.fullname" -}}`
	return regexp.MustCompile(`(?s)` + regexp.QuoteMeta(header) + `.*?\{\{- end -\}\}`)
}

func checkCollectorPeripheralNaming(t *testing.T, beforeChart, afterChart string) {
	t.Helper()
	const collector = "opentelemetry-collector."
	const canonical = "suse-observability-otel-collector"
	for _, tc := range []struct {
		name, release, mode, legacy, namespace, collectorNamespace, sizing, fixture string
		values                                                                      map[string]string
	}{
		{name: "default-deployment", release: "suse-observability", mode: "deployment"},
		{name: "default-daemonset", release: "suse-observability", mode: "daemonset"},
		{name: "default-statefulset", release: "suse-observability", mode: "statefulset"},
		{name: "nightly-deployment", mode: "deployment"},
		{name: "nightly-daemonset", mode: "daemonset"},
		{name: "nightly-statefulset", mode: "statefulset"},
		{name: "custom-deployment", mode: "deployment", legacy: "customer-collector"},
		{name: "custom-daemonset", mode: "daemonset", legacy: "customer-collector"},
		{name: "custom-statefulset", mode: "statefulset", legacy: "customer-collector"},
		{name: "installation-namespace", mode: "deployment", legacy: "customer-collector", namespace: "tenant-b"},
		{name: "collector-namespace", mode: "statefulset", legacy: "customer-collector", collectorNamespace: "telemetry"},
		{name: "cleared-fullname", mode: "deployment", legacy: "nightly-opentelemetry-collector", values: map[string]string{
			collector + "fullnameOverride": "",
		}},
		{name: "name-override", mode: "daemonset", legacy: "nightly-telemetry", values: map[string]string{
			collector + "fullnameOverride": "", collector + "nameOverride": "telemetry",
		}},
		{name: "parent-overrides-and-affixes", mode: "deployment", legacy: "customer-collector", values: map[string]string{
			"fullnameOverride": "customer-platform", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnameOverride": "customer-service", "global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "long-fullname", mode: "deployment", legacy: strings.Repeat("c", 63), values: map[string]string{
			collector + "fullnameOverride": strings.Repeat("c", 70),
		}},
		{name: "external-config-and-account", mode: "statefulset", legacy: "customer-collector", values: map[string]string{
			collector + "configMap.create": "false", collector + "configMap.existingName": "customer-collector-config",
			collector + "serviceAccount.create": "false", collector + "serviceAccount.name": "customer-collector-account",
		}},
		{name: "persistent-claims", mode: "statefulset", legacy: "customer-collector", values: map[string]string{
			collector + "statefulset.volumeClaimTemplates[0].metadata.name":                   "customer-data",
			collector + "statefulset.volumeClaimTemplates[0].spec.accessModes[0]":             "ReadWriteOnce",
			collector + "statefulset.volumeClaimTemplates[0].spec.resources.requests.storage": "3Gi",
			collector + "extraVolumeMounts[0].name":                                           "customer-data",
			collector + "extraVolumeMounts[0].mountPath":                                      "/data",
		}},
		{name: "autoscaling", mode: "deployment", legacy: "customer-collector", values: map[string]string{collector + "autoscaling.enabled": "true"}},
		{name: "argo", mode: "deployment", legacy: "customer-collector", values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
		{name: "ha", mode: "deployment", legacy: "customer-collector", sizing: "global_sizing_150_ha.yaml"},
		{name: "nonha-monolithic", mode: "statefulset", legacy: "customer-collector", sizing: "global_sizing_50_nonha.yaml", values: map[string]string{
			"stackstate.features.server.split": "false", "hbase.deployment.mode": "Mono",
		}},
		{name: "custom-policy-and-monitoring", mode: "deployment", legacy: "customer-collector", values: map[string]string{
			collector + "networkPolicy.annotations.customer":                                     "preserved",
			collector + "networkPolicy.allowIngressFrom[0].namespaceSelector.matchLabels.tenant": "customer",
			collector + "networkPolicy.egressRules[0].to[0].ipBlock.cidr":                        "192.0.2.0/24",
			collector + "serviceMonitor.metricsEndpoints[0].interval":                            "37s",
			collector + "serviceMonitor.extraLabels.customer":                                    "preserved",
			collector + "serviceMonitor.relabelings[0].sourceLabels[0]":                          "__meta_kubernetes_namespace",
			collector + "serviceMonitor.relabelings[0].targetLabel":                              "namespace",
			collector + "prometheusRule.extraLabels.customer":                                    "preserved",
			collector + "podDisruptionBudget.minAvailable":                                       "2",
		}},
		{name: "monitors-disabled", mode: "daemonset", legacy: "customer-collector", values: map[string]string{
			collector + "serviceMonitor.enabled": "false", collector + "podMonitor.enabled": "false",
		}},
		{name: "policy-and-budget-disabled", mode: "deployment", legacy: "customer-collector", values: map[string]string{
			collector + "networkPolicy.enabled": "false", collector + "podDisruptionBudget.enabled": "false",
		}},
		{name: "empty-monitor-endpoints", mode: "daemonset", legacy: "customer-collector", fixture: "collector-peripheral-empty-endpoints.yaml"},
		{name: "service-disabled", mode: "daemonset", legacy: "customer-collector", values: map[string]string{collector + "service.enabled": "false"}},
	} {
		if tc.release == "" {
			tc.release = "nightly"
		}
		if tc.legacy == "" {
			tc.legacy = canonical
		}
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				values := componentResourceNameTestValues()
				for key, value := range map[string]string{
					"mode": tc.mode, "fullnameOverride": tc.legacy, "service.enabled": "true", "ports.metrics.enabled": "true",
					"podDisruptionBudget.enabled": "true", "podDisruptionBudget.minAvailable": "1",
					"serviceMonitor.enabled": "true", "podMonitor.enabled": "true",
					"prometheusRule.enabled": "true", "prometheusRule.defaultRules.enabled": "true", "networkPolicy.enabled": "true",
				} {
					values[collector+key] = value
				}
				for key, value := range tc.values {
					values[key] = value
				}
				options := apiResourceNameTestOptions(values)
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				namespace := options.KubectlOptions.Namespace
				if tc.collectorNamespace != "" {
					namespace = tc.collectorNamespace
					values[collector+"namespaceOverride"] = namespace
				}
				if tc.sizing != "" {
					options.ValuesFiles = append(options.ValuesFiles, filepath.Join("values", tc.sizing))
				}
				if tc.fixture != "" {
					options.ValuesFiles = append(options.ValuesFiles, filepath.Join("values", tc.fixture))
				}
				render := func(chart string) string {
					copied := *options
					copied.Logger = logger.Discard
					copied.ValuesFiles = make([]string, len(options.ValuesFiles))
					for i, file := range options.ValuesFiles {
						path, err := filepath.Abs(file)
						require.NoError(t, err)
						copied.ValuesFiles[i] = path
					}
					args := []string{"--kube-version", "1.32.0", "--api-versions", "policy/v1", "--api-versions", "monitoring.coreos.com/v1"}
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					output, err := helm.RenderTemplateE(t, &copied, chart, tc.release, nil, args...)
					require.NoError(t, err)
					return output
				}
				beforeOutput, afterOutput := render(beforeChart), render(afterChart)
				before := backupNamingDocuments(t, beforeOutput, nil, nil)
				after := backupNamingDocuments(t, afterOutput, nil, nil)
				renames := map[string]string{}
				for _, resource := range collectorPeripheralNamingInventory() {
					old, target := tc.legacy+resource.suffix, canonical+resource.suffix
					oldKey := resource.kind + "/" + namespace + "/" + old
					targetKey := resource.kind + "/" + namespace + "/" + target
					if _, found := before[oldKey]; found {
						require.Contains(t, after, targetKey)
						if old != target {
							renames[resource.kind+"/"+old] = target
							assert.NotContains(t, after, oldKey)
						}
					} else {
						assert.NotContains(t, after, targetKey, "Creation conditions must stay unchanged")
					}
				}
				seen := map[string]bool{}
				expected := backupNamingDocuments(t, beforeOutput, renames, seen)
				assert.Equal(t, expected, after, "Only the five reviewed metadata names may change; no later step may be needed to repair consumers")
				require.Len(t, seen, len(renames))
				original := helmtestutil.NewKubernetesResources(t, beforeOutput)
				target := helmtestutil.NewKubernetesResources(t, afterOutput)
				assert.Equal(t, original.Deployments, target.Deployments)
				assert.Equal(t, original.Statefulsets, target.Statefulsets)
				assert.Equal(t, original.DaemonSets, target.DaemonSets)
				assert.Equal(t, original.PersistentVolumeClaims, target.PersistentVolumeClaims)
				assert.Equal(t, original.ConfigMaps, target.ConfigMaps)
				assert.Equal(t, original.Secrets, target.Secrets)
				assert.Equal(t, original.Services, target.Services)
				assert.Equal(t, original.ServiceAccounts, target.ServiceAccounts)
				var pod corev1.PodTemplateSpec
				switch tc.mode {
				case "deployment":
					pod = target.Deployments[tc.legacy].Spec.Template
				case "daemonset":
					pod = target.DaemonSets[tc.legacy+"-agent"].Spec.Template
				case "statefulset":
					pod = target.Statefulsets[tc.legacy].Spec.Template
				}
				require.Equal(t, tc.release, pod.Labels["app.kubernetes.io/instance"])
				for _, resource := range collectorPeripheralNamingInventory() {
					key := resource.kind + "/" + namespace + "/" + canonical + resource.suffix
					if docs, found := after[key]; found {
						object := docs.([]interface{})[0].(map[string]interface{})
						metadata := object["metadata"].(map[string]interface{})
						require.Equal(t, tc.release, metadata["labels"].(map[string]interface{})["app.kubernetes.io/instance"])
						spec := object["spec"].(map[string]interface{})
						selectorField := "selector"
						if resource.kind == "NetworkPolicy" {
							selectorField = "podSelector"
						}
						if selector, found := spec[selectorField].(map[string]interface{}); found {
							for key, value := range selector["matchLabels"].(map[string]interface{}) {
								assert.Equal(t, value, pod.Labels[key], "Peripheral objects must still select the retained collector Pods")
							}
						}
					}
				}
			})
		}
	}
}
