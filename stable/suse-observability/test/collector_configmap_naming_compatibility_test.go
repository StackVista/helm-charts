package test

import (
	"encoding/json"
	"fmt"
	"maps"
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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// Pre-change source: ecb7ef9770d3685df9bc033ece48ce5577c35290.
// Independent extraction baseline: c851a02ae1f2f7eb64c5538a148194ea36eee4b4.
func TestCollectorConfigMapNamingCompatibility(t *testing.T) {
	before := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(before, os.DirFS("..")))
	path := filepath.Join(before, "templates/_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	for _, helper := range collectorConfigMapNamingHelpers() {
		pattern := collectorConfigMapNamingDefinition(helper)
		require.Len(t, pattern.FindAllString(content, -1), 1)
		content = pattern.ReplaceAllString(content, "")
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	checkCollectorConfigMapNaming(t, before, "..")
}

func TestCollectorConfigMapNamingGitBaselineCompatibility(t *testing.T) {
	before := os.Getenv("COLLECTOR_CONFIGMAP_BASELINE_CHART")
	if before == "" {
		t.Skip("set COLLECTOR_CONFIGMAP_BASELINE_CHART to the built pre-change chart")
	}
	checkCollectorConfigMapNaming(t, before, "..")
}

func TestCollectorConfigMapNamingIndependentOfPreviousSteps(t *testing.T) {
	before := os.Getenv("COLLECTOR_CONFIGMAP_EXTRACTION_CHART")
	if before == "" {
		t.Skip("set COLLECTOR_CONFIGMAP_EXTRACTION_CHART to the built extraction baseline")
	}
	after := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(after, os.DirFS(before)))
	current, err := os.ReadFile("../templates/_names.tpl")
	require.NoError(t, err)
	path := filepath.Join(after, "templates/_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	for _, helper := range collectorConfigMapNamingHelpers() {
		pattern := collectorConfigMapNamingDefinition(helper)
		require.Empty(t, pattern.FindAllString(content, -1))
		definitions := pattern.FindAllString(string(current), -1)
		require.Len(t, definitions, 1, helper)
		content += "\n" + definitions[0] + "\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	checkCollectorConfigMapNaming(t, before, after)
}

func collectorConfigMapNamingHelpers() []string {
	return []string{
		"opentelemetry-collector.deployment.configmap.fullname",
		"opentelemetry-collector.daemonset.configmap.fullname",
		"opentelemetry-collector.statefulset.configmap.fullname",
	}
}

func collectorConfigMapNamingDefinition(helper string) *regexp.Regexp {
	return regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
}

func checkCollectorConfigMapNaming(t *testing.T, beforeChart, afterChart string) {
	t.Helper()
	const collector = "opentelemetry-collector."
	for _, tc := range []struct {
		name, mode, release, namespace, sizing string
		values                                 map[string]string
	}{
		{name: "default-deployment", mode: "deployment", release: "suse-observability"},
		{name: "default-daemonset", mode: "daemonset", release: "suse-observability"},
		{name: "default-statefulset", mode: "statefulset", release: "suse-observability"},
		{name: "custom-deployment", mode: "deployment"},
		{name: "custom-daemonset", mode: "daemonset"},
		{name: "custom-statefulset", mode: "statefulset"},
		{name: "cleared-fullname-and-name-override", mode: "deployment", values: map[string]string{
			collector + "fullnameOverride": "", collector + "nameOverride": "telemetry",
		}},
		{name: "parent-and-global-naming", mode: "statefulset", values: map[string]string{
			"fullnameOverride": "customer-platform", "global.fullnameOverride": "customer-global",
			"fullnamePrefix": "team-", "fullnameSuffix": "-local",
			"global.fullnamePrefix": "global-", "global.fullnameSuffix": "-end",
		}},
		{name: "second-installation-namespace", mode: "daemonset", namespace: "tenant-b"},
		{name: "external-deployment-config", mode: "deployment", values: map[string]string{
			collector + "configMap.create": "false", collector + "configMap.existingName": "customer-configuration",
		}},
		{name: "external-daemonset-config", mode: "daemonset", values: map[string]string{
			collector + "configMap.create": "false", collector + "configMap.existingName": "customer-configuration",
		}},
		{name: "external-statefulset-config", mode: "statefulset", values: map[string]string{
			collector + "configMap.create": "false", collector + "configMap.existingName": "customer-configuration",
		}},
		{name: "configuration-disabled", mode: "deployment", values: map[string]string{collector + "configMap.create": "false"}},
		{name: "custom-relay-config", mode: "statefulset", values: map[string]string{
			collector + "config.receivers.otlp.protocols.grpc.endpoint": "0.0.0.0:14317",
			collector + "config.processors.batch.timeout":               "7s",
			collector + "config.exporters.debug.verbosity":              "detailed",
		}},
		{name: "persistent-claims", mode: "statefulset", values: map[string]string{
			collector + "statefulset.volumeClaimTemplates[0].metadata.name":                   "customer-data",
			collector + "statefulset.volumeClaimTemplates[0].spec.accessModes[0]":             "ReadWriteOnce",
			collector + "statefulset.volumeClaimTemplates[0].spec.resources.requests.storage": "3Gi",
			collector + "extraVolumeMounts[0].name":                                           "customer-data", collector + "extraVolumeMounts[0].mountPath": "/data",
		}},
		{name: "log-checkpoints", mode: "daemonset", values: map[string]string{
			collector + "presets.logsCollection.enabled": "true", collector + "presets.logsCollection.storeCheckpoints": "true",
		}},
		{name: "autoscaling", mode: "statefulset", values: map[string]string{collector + "autoscaling.enabled": "true"}},
		{name: "explicit-checksum-template", mode: "statefulset", values: map[string]string{
			collector + "configMap.create": "false", collector + "configMap.existingName": "customer-configuration",
			collector + "configMap.existingPath": "configmap-statefulset.yaml",
		}},
		{name: "ha", mode: "statefulset", sizing: "global_sizing_150_ha.yaml"},
		{name: "nonha", mode: "deployment", sizing: "global_sizing_50_nonha.yaml"},
		{name: "argo-and-router-hooks", mode: "statefulset", values: map[string]string{
			"deployment.compatibleWithArgoCD": "true", "stackstate.components.router.mode.status": "automatic",
		}},
		{name: "recreate", mode: "deployment", values: map[string]string{collector + "rollout.strategy": "Recreate"}},
	} {
		if tc.release == "" {
			tc.release = "nightly"
		}
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				values := componentResourceNameTestValues()
				values[collector+"mode"] = tc.mode
				values[collector+"fullnameOverride"] = "customer-collector"
				if strings.HasPrefix(tc.name, "default-") {
					values[collector+"fullnameOverride"] = "suse-observability-otel-collector"
				}
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
				beforeOutput := renderCollectorConfigMapChart(t, beforeChart, tc.release, options, upgrade)
				afterOutput := renderCollectorConfigMapChart(t, afterChart, tc.release, options, upgrade)
				expected := backupNamingDocuments(t, beforeOutput, nil, nil)
				actual := backupNamingDocuments(t, afterOutput, nil, nil)
				oldConfig, targetConfig := "", ""
				for key, documents := range maps.Clone(expected) {
					for i, raw := range documents.([]interface{}) {
						object := raw.(map[string]interface{})
						data, _ := object["data"].(map[string]interface{})
						if object["kind"] != "ConfigMap" || data["relay"] == nil {
							continue
						}
						require.Empty(t, oldConfig, "Exactly one relay ConfigMap is generated")
						metadata := object["metadata"].(map[string]interface{})
						oldConfig = metadata["name"].(string)
						targetConfig = "suse-observability-otel-collector"
						switch tc.mode {
						case "deployment":
							targetConfig += "-config"
						case "daemonset":
							targetConfig += "-agent"
						case "statefulset":
							targetConfig += "-statefulset"
						}
						if oldConfig == targetConfig {
							continue
						}
						metadata["name"] = targetConfig
						namespace, _ := metadata["namespace"].(string)
						newKey := "ConfigMap/" + namespace + "/" + targetConfig
						require.NotContains(t, expected, newKey)
						expected[newKey] = []interface{}{object}
						remaining := append(documents.([]interface{})[:i:i], documents.([]interface{})[i+1:]...)
						if len(remaining) == 0 {
							delete(expected, key)
						} else {
							expected[key] = remaining
						}
					}
				}
				generated := values[collector+"configMap.create"] != "false"
				require.Equal(t, generated, oldConfig != "")
				controllers := 0
				controllerName := ""
				for key, documents := range expected {
					if !strings.HasPrefix(key, "Deployment/") && !strings.HasPrefix(key, "DaemonSet/") && !strings.HasPrefix(key, "StatefulSet/") {
						continue
					}
					for _, raw := range documents.([]interface{}) {
						object := raw.(map[string]interface{})
						template := object["spec"].(map[string]interface{})["template"].(map[string]interface{})
						pod := template["spec"].(map[string]interface{})
						isCollector := false
						for _, raw := range pod["containers"].([]interface{}) {
							isCollector = isCollector || raw.(map[string]interface{})["name"] == "opentelemetry-collector"
						}
						if !isCollector {
							continue
						}
						controllers++
						controllerName = object["metadata"].(map[string]interface{})["name"].(string)
						references := 0
						volumes, _ := pod["volumes"].([]interface{})
						for _, raw := range volumes {
							volume := raw.(map[string]interface{})
							if volume["name"] != "opentelemetry-collector-configmap" {
								continue
							}
							ref := volume["configMap"].(map[string]interface{})
							if generated {
								require.Equal(t, oldConfig, ref["name"])
								ref["name"] = targetConfig
							} else {
								require.Equal(t, values[collector+"configMap.existingName"], ref["name"])
							}
							references++
						}
						usesConfig := generated || values[collector+"configMap.existingName"] != ""
						require.Equal(t, usesConfig, references == 1)
						require.Contains(t, actual, key)
						actualObject := actual[key].([]interface{})[0].(map[string]interface{})
						annotations := template["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
						newAnnotations := actualObject["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
						if oldConfig != targetConfig {
							assert.NotEqual(t, annotations["checksum/config"], newAnnotations["checksum/config"])
							annotations["checksum/config"] = newAnnotations["checksum/config"]
						}
					}
				}
				require.Equal(t, 1, controllers)
				assert.Equal(t, expected, actual, "Only generated relay names, their volume references and existing checksum annotations may change; endpoint maps and env references stay unchanged")
				// Resolve namespace defaults when checking duplicate identities:
				// the parent endpoint map omits metadata.namespace.
				identities := map[string]bool{}
				for key, documents := range actual {
					if !strings.HasPrefix(key, "ConfigMap/") {
						continue
					}
					for _, raw := range documents.([]interface{}) {
						object := raw.(map[string]interface{})
						metadata := object["metadata"].(map[string]interface{})
						namespace, _ := metadata["namespace"].(string)
						if namespace == "" {
							namespace = options.KubectlOptions.Namespace
						}
						identity := namespace + "/" + metadata["name"].(string)
						require.False(t, identities[identity], "ConfigMap identities must be distinct after upgrade")
						identities[identity] = true
					}
				}
				before := helmtestutil.NewKubernetesResources(t, beforeOutput)
				after := helmtestutil.NewKubernetesResources(t, afterOutput)
				switch tc.mode {
				case "deployment":
					deployment, updated := before.Deployments[controllerName], after.Deployments[controllerName]
					assertCollectorConfigMapPatches(t, &deployment, &updated)
				case "daemonset":
					daemonset, updated := before.DaemonSets[controllerName], after.DaemonSets[controllerName]
					assertCollectorConfigMapPatches(t, &daemonset, &updated)
				case "statefulset":
					statefulset, updated := before.Statefulsets[controllerName], after.Statefulsets[controllerName]
					assertCollectorConfigMapPatches(t, &statefulset, &updated)
				}
			})
		}
	}
}

func renderCollectorConfigMapChart(t *testing.T, chart, release string, options *helm.Options, upgrade bool) string {
	t.Helper()
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
	output, err := helm.RenderTemplateE(t, &copied, chart, release, nil, args...)
	require.NoError(t, err)
	return output
}

func assertCollectorConfigMapPatches(t *testing.T, before, after runtime.Object) {
	t.Helper()
	live := func(object runtime.Object) runtime.Object {
		copy := object.DeepCopyObject()
		metadata, err := meta.Accessor(copy)
		require.NoError(t, err)
		annotations := metadata.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations["customer"] = "preserved"
		metadata.SetAnnotations(annotations)
		metadata.SetResourceVersion("123")
		return copy
	}
	for _, pair := range [][2]runtime.Object{{before, after}, {after, after}, {after, before}} {
		original, err := json.Marshal(pair[0])
		require.NoError(t, err)
		modified, err := json.Marshal(pair[1])
		require.NoError(t, err)
		current, err := json.Marshal(live(pair[0]))
		require.NoError(t, err)
		schema, err := strategicpatch.NewPatchMetaFromStruct(pair[0])
		require.NoError(t, err)
		patch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, current, schema, true)
		require.NoError(t, err)
		var fields map[string]interface{}
		require.NoError(t, json.Unmarshal(patch, &fields))
		if spec, ok := fields["spec"].(map[string]interface{}); ok {
			for _, immutable := range []string{"selector", "serviceName", "volumeClaimTemplates"} {
				assert.NotContains(t, spec, immutable)
			}
		}
		applied, err := strategicpatch.StrategicMergePatch(current, patch, pair[0])
		require.NoError(t, err)
		expected, err := json.Marshal(live(pair[1]))
		require.NoError(t, err)
		assert.JSONEq(t, string(expected), string(applied), "Upgrade/repeat/rollback patches preserve complete objects and current metadata")
	}
}
