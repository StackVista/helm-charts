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
	corev1 "k8s.io/api/core/v1"
)

// Independent test names distinguish resources that normally share a canonical
// identity. Renaming the StatefulSet here is only a wiring probe in a copied
// chart; a real rename would require migrating its generated persistent claims.
func TestUIAndAIResourcesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	renames := map[string]map[string]string{}
	for component, resources := range map[string][]string{
		"ui":          {"deployment", "service", "secret"},
		"mcp":         {"deployment", "service", "secret", "serviceaccount"},
		"aiAssistant": {"statefulset", "service", "secret", "serviceaccount"},
	} {
		suffix := component
		if component == "aiAssistant" {
			suffix = "ai-assistant"
		}
		for _, resource := range resources {
			helper := "stackstate." + component + "." + resource + ".fullname"
			name := "explicit-" + suffix + "-" + resource
			definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
			require.Len(t, definition.FindAll(data, -1), 1, helper)
			data = definition.ReplaceAll(data, []byte(`{{- define "`+helper+`" -}}`+name+`{{- end -}}`))
			if renames[resource] == nil {
				renames[resource] = map[string]string{}
			}
			renames[resource]["suse-observability-"+suffix] = name
		}
	}
	require.NoError(t, os.WriteFile(namesPath, data, 0600))
	fullValues, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, split := range []bool{false, true} {
		for _, tc := range []struct {
			name, provider, externalSecret string
			assistant, mcp, extraEnv       bool
		}{
			{"bedrock", "bedrock", "", true, true, true},
			{"bedrock-no-extra-env", "bedrock", "", true, true, false},
			{"anthropic", "anthropic", "", true, true, true},
			{"anthropic-no-extra-env", "anthropic", "", true, true, false},
			{"anthropic-external", "anthropic", "customer-anthropic", true, true, true},
			{"assistant-enables-mcp", "bedrock", "", true, false, true},
			{"mcp-only", "bedrock", "", false, true, true},
			{"disabled", "bedrock", "", false, false, true},
		} {
			t.Run(fmt.Sprintf("%s/split=%t", tc.name, split), func(t *testing.T) {
				values := map[string]string{
					"stackstate.features.server.split":                                     fmt.Sprint(split),
					"ai.assistant.enabled":                                                 fmt.Sprint(tc.assistant),
					"ai.mcp.enabled":                                                       fmt.Sprint(tc.mcp),
					"ai.assistant.provider":                                                tc.provider,
					"ai.assistant.anthropic.apiKey":                                        "fixture-key",
					"ai.assistant.anthropic.fromExternalSecret.name":                       tc.externalSecret,
					"stackstate.components.aiAssistant.persistence.size":                   "7Gi",
					"stackstate.components.aiAssistant.persistence.storageClass":           "retained-storage",
					"stackstate.components.all.metrics.servicemonitor.enabled":             "true",
					"stackstate.components.aiAssistant.serviceAccount.annotations.fixture": "retained",
				}
				if tc.externalSecret != "" {
					values["ai.assistant.anthropic.fromExternalSecret.key"] = "customer-key"
				}
				if tc.extraEnv {
					for _, component := range []string{"ui", "mcp", "aiAssistant"} {
						values["stackstate.components."+component+".extraEnv.secret.TEST_PRIVATE"] = "fixture-" + component
					}
				}
				options := apiResourceNameTestOptions(values)
				options.ValuesFiles = []string{fullValues}
				before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
				output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
				require.NoError(t, err)
				after := helmtestutil.NewKubernetesResources(t, output)
				assert.Contains(t, after.Deployments, "explicit-ui-deployment")
				_, mcpExists := after.Deployments["explicit-mcp-deployment"]
				assert.Equal(t, tc.assistant || tc.mcp, mcpExists)
				_, assistantExists := after.Statefulsets["explicit-ai-assistant-statefulset"]
				assert.Equal(t, tc.assistant, assistantExists)
				for component, enabled := range map[string]bool{"ui": tc.extraEnv, "mcp": (tc.assistant || tc.mcp) && tc.extraEnv, "ai-assistant": tc.assistant && (tc.extraEnv || (tc.provider == "anthropic" && tc.externalSecret == ""))} {
					_, exists := after.Secrets["explicit-"+component+"-secret"]
					assert.Equal(t, enabled, exists, component+" Secret creation condition")
				}
				assert.Len(t, after.Secrets, len(before.Secrets))
				for name, secret := range before.Secrets {
					expected := secret.DeepCopy()
					if replacement, ok := renames["secret"][name]; ok {
						expected.Name = replacement
						assert.NotContains(t, after.Secrets, name)
					}
					require.Contains(t, after.Secrets, expected.Name)
					assert.Equal(t, *expected, after.Secrets[expected.Name])
				}
				assert.Len(t, after.ServiceAccounts, len(before.ServiceAccounts))
				for name, account := range before.ServiceAccounts {
					expected := account.DeepCopy()
					if replacement, ok := renames["serviceaccount"][name]; ok {
						expected.Name = replacement
						assert.NotContains(t, after.ServiceAccounts, name)
					}
					require.Contains(t, after.ServiceAccounts, expected.Name)
					assert.Equal(t, *expected, after.ServiceAccounts[expected.Name])
				}
				assert.Len(t, after.Services, len(before.Services))
				for name, service := range before.Services {
					expected := service.DeepCopy()
					if replacement, ok := renames["service"][name]; ok {
						expected.Name = replacement
						assert.NotContains(t, after.Services, name)
					}
					require.Contains(t, after.Services, expected.Name)
					assert.Equal(t, *expected, after.Services[expected.Name])
				}
				assert.Len(t, after.Deployments, len(before.Deployments))
				for name, deployment := range before.Deployments {
					expected := deployment.DeepCopy()
					if replacement, ok := renames["deployment"][name]; ok {
						expected.Name = replacement
						assert.NotContains(t, after.Deployments, name)
					}
					require.Contains(t, after.Deployments, expected.Name)
					actual := after.Deployments[expected.Name]
					rewriteUIAndAIPod(t, &expected.Spec.Template, actual.Spec.Template, renames, after.Secrets)
					assert.Equal(t, *expected, actual)
				}
				assert.Len(t, after.Statefulsets, len(before.Statefulsets))
				for name, sts := range before.Statefulsets {
					expected := sts.DeepCopy()
					if replacement, ok := renames["statefulset"][name]; ok {
						expected.Name = replacement
						assert.NotContains(t, after.Statefulsets, name)
					}
					if replacement, ok := renames["service"][expected.Spec.ServiceName]; ok {
						expected.Spec.ServiceName = replacement
					}
					require.Contains(t, after.Statefulsets, expected.Name)
					actual := after.Statefulsets[expected.Name]
					rewriteUIAndAIPod(t, &expected.Spec.Template, actual.Spec.Template, renames, after.Secrets)
					// This includes claim templates, mount paths, replica counts and selectors.
					assert.Equal(t, *expected, actual)
				}
				assert.Len(t, after.ConfigMaps, len(before.ConfigMaps))
				for name, config := range before.ConfigMaps {
					expected := config.DeepCopy()
					for key, text := range expected.Data {
						for old, replacement := range renames["service"] {
							// Only Envoy DNS targets change. Cluster/listener/route identifiers stay put.
							text = strings.ReplaceAll(text, `address: "`+old+`"`, `address: "`+replacement+`"`)
						}
						expected.Data[key] = text
					}
					require.Contains(t, after.ConfigMaps, name)
					assert.Equal(t, *expected, after.ConfigMaps[name])
				}
				assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
				assert.Equal(t, before.Pdbs, after.Pdbs)
				assert.Equal(t, before.ServiceMonitors, after.ServiceMonitors)
				assert.Equal(t, before.Ingresses, after.Ingresses)
				assert.Equal(t, before.HTTPRoutes, after.HTTPRoutes)
			})
		}
	}
}

func rewriteUIAndAIPod(t *testing.T, expected *corev1.PodTemplateSpec, actual corev1.PodTemplateSpec, renames map[string]map[string]string, secrets map[string]corev1.Secret) {
	t.Helper()
	if replacement, ok := renames["serviceaccount"][expected.Spec.ServiceAccountName]; ok {
		expected.Spec.ServiceAccountName = replacement
	}
	for _, containers := range [][]corev1.Container{expected.Spec.InitContainers, expected.Spec.Containers} {
		for i := range containers {
			for j := range containers[i].Env {
				env := &containers[i].Env[j]
				if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
					if replacement, ok := renames["secret"][env.ValueFrom.SecretKeyRef.Name]; ok {
						env.ValueFrom.SecretKeyRef.Name = replacement
					}
				}
				if env.Name == "AI_SERVICE_URL" || env.Name == "MCP_SERVER_URL" {
					for old, replacement := range renames["service"] {
						env.Value = strings.ReplaceAll(env.Value, "http://"+old+":", "http://"+replacement+":")
					}
				}
			}
		}
	}
	for _, component := range []string{"ui", "mcp", "ai-assistant"} {
		checksum := "checksum/" + component + "-env"
		if old, ok := expected.Annotations[checksum]; ok {
			require.NotEmpty(t, actual.Annotations[checksum])
			if _, exists := secrets[renames["secret"]["suse-observability-"+component]]; exists {
				assert.NotEqual(t, old, actual.Annotations[checksum], "the test-only Secret rename must change its existing checksum")
			} else {
				assert.Equal(t, old, actual.Annotations[checksum], "an absent Secret retains its existing empty-template checksum")
			}
			expected.Annotations[checksum] = actual.Annotations[checksum]
		}
	}
}
