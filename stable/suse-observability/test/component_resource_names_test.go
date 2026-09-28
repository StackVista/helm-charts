package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

var explicitConfigurationComponents = []struct{ key, suffix string }{
	{"checks", "checks"},
	{"notification", "notification"},
	{"healthSync", "health-sync"},
	{"authorizationSync", "authorization-sync"},
	{"state", "state"},
	{"sync", "sync"},
	{"slicing", "slicing"},
	{"server", "server"},
	{"receiver", "receiver"},
	{"correlate", "correlate"},
	{"initializer", "initializer"},
	{"e2es", "e2es"},
}

func TestComponentResourceNamesPreserveLegacyConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name, release, prefix string
		values                map[string]string
	}{
		{"default", "suse-observability", "suse-observability", nil},
		{"custom-release", "nightly", "nightly-suse-observability", nil},
		{"fullname-override", "nightly", "custom", map[string]string{"fullnameOverride": "custom"}},
		{"prefix-suffix", "nightly", "g-pre-custom-post-end", map[string]string{
			"fullnameOverride": "custom", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{"global-override", "nightly", "nightly-suse-observability", map[string]string{"global.fullnameOverride": "global-only"}},
		{"long-name", "nightly", strings.Repeat("a", 54), map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{"authorization-disabled", "nightly", "nightly-suse-observability", map[string]string{"stackstate.k8sAuthorization.enabled": "false"}},
		{"split-workers", "nightly", "nightly-suse-observability", map[string]string{
			"stackstate.components.receiver.split.enabled":  "true",
			"stackstate.components.correlate.split.enabled": "true",
		}},
	} {
		for _, split := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/split=%t", tc.name, split), func(t *testing.T) {
				values := componentResourceNameTestValues()
				for key, value := range tc.values {
					values[key] = value
				}
				values["stackstate.features.server.split"] = fmt.Sprint(split)
				output := helmtestutil.RenderHelmTemplateOptsNoError(t, tc.release, apiResourceNameTestOptions(values))
				resources := helmtestutil.NewKubernetesResources(t, output)
				for _, component := range explicitConfigurationComponents {
					name := tc.prefix + "-" + component.suffix
					deployments := configurationComponentDeployments(component.key, values)
					if len(deployments) > 0 {
						for _, deployment := range deployments {
							assertConfigurationReferences(t, resources, tc.prefix+"-"+deployment, name, name+"-log", name, map[string]string{
								"SHARED_SETTING": component.key + "-override", "COMPONENT_ONLY": component.key + "-specific", "GLOBAL_ONLY": "global-value",
							})
						}
					} else {
						assert.NotContains(t, resources.Deployments, name)
						assert.NotContains(t, resources.ConfigMaps, name)
						assert.NotContains(t, resources.ConfigMaps, name+"-log")
						assert.NotContains(t, resources.Secrets, name)
					}
				}
			})
		}
	}
}

func TestComponentConfigurationReferencesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	helperPath := filepath.Join(chart, "templates", "_names.tpl")
	original, err := os.ReadFile(helperPath)
	require.NoError(t, err)
	content := string(original)
	for _, component := range explicitConfigurationComponents {
		for helperSuffix, nameSuffix := range map[string]string{
			"configmap": "config", "log.configmap": "logging", "secret": "environment",
		} {
			helper := "stackstate." + component.key + "." + helperSuffix + ".fullname"
			definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
			require.Len(t, definition.FindAllString(content, -1), 1, helper)
			name := "explicit-" + component.suffix + "-" + nameSuffix
			content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}`+name+`{{- end -}}`)
		}
	}
	require.NoError(t, os.WriteFile(helperPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, serverSplit := range []bool{false, true} {
		for _, workersSplit := range []bool{false, true} {
			t.Run(fmt.Sprintf("server-split=%t/workers-split=%t", serverSplit, workersSplit), func(t *testing.T) {
				values := componentResourceNameTestValues()
				values["stackstate.features.server.split"] = fmt.Sprint(serverSplit)
				values["stackstate.components.receiver.split.enabled"] = fmt.Sprint(workersSplit)
				values["stackstate.components.correlate.split.enabled"] = fmt.Sprint(workersSplit)
				options := apiResourceNameTestOptions(values)
				options.ValuesFiles = []string{valuesFile}
				output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
				require.NoError(t, err)
				resources := helmtestutil.NewKubernetesResources(t, output)
				for _, component := range explicitConfigurationComponents {
					prefix := "explicit-" + component.suffix
					deployments := configurationComponentDeployments(component.key, values)
					for _, deployment := range deployments {
						assertConfigurationReferences(t, resources, "nightly-suse-observability-"+deployment, prefix+"-config", prefix+"-logging", prefix+"-environment", map[string]string{
							"SHARED_SETTING": component.key + "-override", "COMPONENT_ONLY": component.key + "-specific", "GLOBAL_ONLY": "global-value",
						})
					}
					if len(deployments) == 0 {
						assert.NotContains(t, resources.ConfigMaps, prefix+"-config")
						assert.NotContains(t, resources.ConfigMaps, prefix+"-logging")
						assert.NotContains(t, resources.Secrets, prefix+"-environment")
					}
					legacy := "nightly-suse-observability-" + component.suffix
					assert.NotContains(t, resources.ConfigMaps, legacy)
					assert.NotContains(t, resources.ConfigMaps, legacy+"-log")
					assert.NotContains(t, resources.Secrets, legacy)
				}
			})
		}
	}
}

// Receiver and correlate workers share configuration even when their Deployments split.
func configurationComponentDeployments(key string, values map[string]string) []string {
	serverSplit := values["stackstate.features.server.split"] == "true"
	switch key {
	case "server":
		if serverSplit {
			return nil
		}
	case "receiver":
		if values["stackstate.components.receiver.split.enabled"] == "true" {
			return []string{"receiver-base", "receiver-logs", "receiver-process-agent"}
		}
	case "correlate":
		if values["stackstate.components.correlate.split.enabled"] == "true" {
			return []string{"correlate-connection", "correlate-http-tracing", "correlate-aggregator"}
		}
	case "e2es":
	default:
		if !serverSplit || (key == "authorizationSync" && values["stackstate.k8sAuthorization.enabled"] == "false") {
			return nil
		}
	}
	switch key {
	case "healthSync":
		return []string{"health-sync"}
	case "authorizationSync":
		return []string{"authorization-sync"}
	default:
		return []string{key}
	}
}

func componentResourceNameTestValues() map[string]string {
	values := map[string]string{
		"stackstate.features.server.split":                         "true",
		"stackstate.components.receiver.split.enabled":             "false",
		"stackstate.components.correlate.split.enabled":            "false",
		"stackstate.components.all.extraEnv.secret.SHARED_SETTING": "shared-value",
		"stackstate.components.all.extraEnv.secret.GLOBAL_ONLY":    "global-value",
	}
	for _, component := range explicitConfigurationComponents {
		prefix := "stackstate.components." + component.key + ".extraEnv.secret."
		values[prefix+"SHARED_SETTING"] = component.key + "-override"
		values[prefix+"COMPONENT_ONLY"] = component.key + "-specific"
	}
	// Split correlate workers currently need explicit memory settings and their
	// own extraEnv settings. All workers consume the shared correlate Secret.
	for _, worker := range []string{"connection", "httpTracing", "aggregator"} {
		prefix := "stackstate.components.correlate.split." + worker + "."
		values[prefix+"resources.limits.memory"] = "2Gi"
		values[prefix+"resources.requests.memory"] = "1Gi"
		values[prefix+"sizing.baseMemoryConsumption"] = "400Mi"
		values[prefix+"sizing.javaHeapMemoryFraction"] = "65"
		values[prefix+"extraEnv.secret.SHARED_SETTING"] = "correlate-override"
		values[prefix+"extraEnv.secret.COMPONENT_ONLY"] = "correlate-specific"
	}
	return values
}
