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
					enabled := split && (component.key != "authorizationSync" || values["stackstate.k8sAuthorization.enabled"] != "false")
					if enabled {
						assertConfigurationReferences(t, resources, name, name, name+"-log", name, map[string]string{
							"SHARED_SETTING": component.key + "-override", "COMPONENT_ONLY": component.key + "-specific", "GLOBAL_ONLY": "global-value",
						})
					} else {
						assert.NotContains(t, resources.Deployments, name)
						assert.NotContains(t, resources.ConfigMaps, name)
						assert.NotContains(t, resources.ConfigMaps, name+"-log")
						assert.NotContains(t, resources.Secrets, name)
					}
				}
				fallback := "server"
				if split {
					fallback = "sync"
				}
				assertLegacyConfigurationReferences(t, resources, tc.prefix+"-"+fallback)
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
	options := apiResourceNameTestOptions(componentResourceNameTestValues())
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	options.ValuesFiles = []string{valuesFile}
	output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
	require.NoError(t, err)
	resources := helmtestutil.NewKubernetesResources(t, output)
	for _, component := range explicitConfigurationComponents {
		deployment := "nightly-suse-observability-" + component.suffix
		prefix := "explicit-" + component.suffix
		assertConfigurationReferences(t, resources, deployment, prefix+"-config", prefix+"-logging", prefix+"-environment", map[string]string{
			"SHARED_SETTING": component.key + "-override", "COMPONENT_ONLY": component.key + "-specific", "GLOBAL_ONLY": "global-value",
		})
		assert.NotContains(t, resources.ConfigMaps, deployment)
		assert.NotContains(t, resources.ConfigMaps, deployment+"-log")
		assert.NotContains(t, resources.Secrets, deployment)
	}
	assertLegacyConfigurationReferences(t, resources, "nightly-suse-observability-sync")
}

func componentResourceNameTestValues() map[string]string {
	values := map[string]string{
		"stackstate.features.server.split":                         "true",
		"stackstate.components.all.extraEnv.secret.SHARED_SETTING": "shared-value",
		"stackstate.components.all.extraEnv.secret.GLOBAL_ONLY":    "global-value",
	}
	for _, component := range explicitConfigurationComponents {
		prefix := "stackstate.components." + component.key + ".extraEnv.secret."
		values[prefix+"SHARED_SETTING"] = component.key + "-override"
		values[prefix+"COMPONENT_ONLY"] = component.key + "-specific"
	}
	return values
}
