package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestAPIResourceNamesPreserveLegacyConfiguration(t *testing.T) {
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
	} {
		for _, split := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/split=%t", tc.name, split), func(t *testing.T) {
				values := apiResourceNameTestValues()
				for key, value := range tc.values {
					values[key] = value
				}
				values["stackstate.features.server.split"] = fmt.Sprint(split)
				output := helmtestutil.RenderHelmTemplateOptsNoError(t, tc.release, apiResourceNameTestOptions(values))
				resources := helmtestutil.NewKubernetesResources(t, output)
				api := tc.prefix + "-api"
				if split {
					assertAPIConfigurationReferences(t, resources, api, api, api+"-log", api)
					// An unmigrated component still resolves its legacy names.
					sync := tc.prefix + "-sync"
					require.Contains(t, resources.Deployments, sync)
					assertLegacyConfigurationReferences(t, resources, sync)
				} else {
					assert.NotContains(t, resources.Deployments, api)
					assert.NotContains(t, resources.ConfigMaps, api)
					assert.NotContains(t, resources.ConfigMaps, api+"-log")
					assert.NotContains(t, resources.Secrets, api)
					server := tc.prefix + "-server"
					require.Contains(t, resources.Deployments, server)
					assertLegacyConfigurationReferences(t, resources, server)
				}
			})
		}
	}
}

// Give the three helpers different outputs to prove both producers and consumers
// use the dedicated names, rather than independently rebuilding the legacy name.
func TestAPIConfigurationReferencesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	helperPath := filepath.Join(chart, "templates", "_names.tpl")
	original, err := os.ReadFile(helperPath)
	require.NoError(t, err)
	for _, tc := range []struct{ name, emptyHelper, errorMessage string }{
		{name: "distinct-names"},
		{"empty-config", "stackstate.api.configmap.fullname", "configMapName must not be empty"},
		{"empty-log-config", "stackstate.api.log.configmap.fullname", "logConfigMapName must not be empty"},
		{"empty-secret", "stackstate.api.secret.fullname", "extraEnvSecretName must not be empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := string(original)
			for helper, value := range map[string]string{
				"stackstate.api.configmap.fullname":     "explicit-api-config",
				"stackstate.api.log.configmap.fullname": "explicit-api-logging",
				"stackstate.api.secret.fullname":        "explicit-api-environment",
			} {
				if helper == tc.emptyHelper {
					value = ""
				}
				definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
				require.Len(t, definition.FindAllString(content, -1), 1, helper)
				content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}`+value+`{{- end -}}`)
			}
			require.NoError(t, os.WriteFile(helperPath, []byte(content), 0600))
			options := apiResourceNameTestOptions(apiResourceNameTestValues())
			// This render uses the temporary chart, not the source directory.
			valuesFile, err := filepath.Abs("values/full.yaml")
			require.NoError(t, err)
			options.ValuesFiles = []string{valuesFile}
			output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
			if tc.errorMessage != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				return
			}
			require.NoError(t, err)
			resources := helmtestutil.NewKubernetesResources(t, output)
			assertAPIConfigurationReferences(t, resources, "nightly-suse-observability-api", "explicit-api-config", "explicit-api-logging", "explicit-api-environment")
			assert.NotContains(t, resources.ConfigMaps, "nightly-suse-observability-api")
			assert.NotContains(t, resources.ConfigMaps, "nightly-suse-observability-api-log")
			assert.NotContains(t, resources.Secrets, "nightly-suse-observability-api")
			assertLegacyConfigurationReferences(t, resources, "nightly-suse-observability-sync")
		})
	}
}

func apiResourceNameTestValues() map[string]string {
	return map[string]string{
		"stackstate.features.server.split":                         "true",
		"stackstate.components.all.extraEnv.secret.SHARED_SETTING": "shared-value",
		"stackstate.components.api.extraEnv.secret.SHARED_SETTING": "api-value",
		"stackstate.components.api.extraEnv.secret.API_ONLY":       "api-only-value",
	}
}

func apiResourceNameTestOptions(values map[string]string) *helm.Options {
	return &helm.Options{
		ValuesFiles: []string{"values/full.yaml"}, SetValues: values,
		KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"}, Logger: logger.Discard,
	}
}

func assertAPIConfigurationReferences(t *testing.T, resources helmtestutil.KubernetesResources, deployment, config, logging, secret string) {
	t.Helper()
	assertConfigurationReferences(t, resources, deployment, config, logging, secret, map[string]string{"SHARED_SETTING": "api-value", "API_ONLY": "api-only-value"})
}

func assertConfigurationReferences(t *testing.T, resources helmtestutil.KubernetesResources, deployment, config, logging, secret string, expectedSecretData map[string]string) {
	t.Helper()
	require.Contains(t, resources.Deployments, deployment)
	require.Contains(t, resources.ConfigMaps, config)
	require.Contains(t, resources.ConfigMaps, logging)
	require.Contains(t, resources.Secrets, secret)
	assertConfigurationVolumes(t, resources.Deployments[deployment].Spec.Template.Spec.Volumes, config, logging)
	found := map[string]bool{}
	for _, container := range resources.Deployments[deployment].Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if expected, ok := expectedSecretData[env.Name]; ok {
				require.NotNil(t, env.ValueFrom)
				require.NotNil(t, env.ValueFrom.SecretKeyRef)
				assert.Equal(t, secret, env.ValueFrom.SecretKeyRef.Name)
				assert.Equal(t, env.Name, env.ValueFrom.SecretKeyRef.Key)
				assert.Equal(t, expected, string(resources.Secrets[secret].Data[env.Name]))
				found[env.Name] = true
			}
		}
	}
	assert.Len(t, found, len(expectedSecretData))
}

func assertLegacyConfigurationReferences(t *testing.T, resources helmtestutil.KubernetesResources, component string) {
	t.Helper()
	require.Contains(t, resources.Deployments, component)
	require.Contains(t, resources.ConfigMaps, component)
	require.Contains(t, resources.ConfigMaps, component+"-log")
	require.Contains(t, resources.Secrets, component)
	pod := resources.Deployments[component].Spec.Template.Spec
	assertConfigurationVolumes(t, pod.Volumes, component, component+"-log")
	found := false
	for _, container := range pod.Containers {
		for _, env := range container.Env {
			if env.Name == "SHARED_SETTING" {
				require.NotNil(t, env.ValueFrom)
				require.NotNil(t, env.ValueFrom.SecretKeyRef)
				assert.Equal(t, component, env.ValueFrom.SecretKeyRef.Name)
				assert.Equal(t, "shared-value", string(resources.Secrets[component].Data[env.ValueFrom.SecretKeyRef.Key]))
				found = true
			}
		}
	}
	assert.True(t, found)
}

func assertConfigurationVolumes(t *testing.T, volumes []corev1.Volume, config, logging string) {
	t.Helper()
	found := map[string]bool{}
	for _, volume := range volumes {
		if expected, ok := map[string]string{"config-volume": config, "config-volume-log": logging}[volume.Name]; ok {
			require.NotNil(t, volume.ConfigMap)
			assert.Equal(t, expected, volume.ConfigMap.Name)
			found[volume.Name] = true
		}
	}
	assert.Len(t, found, 2)
}
