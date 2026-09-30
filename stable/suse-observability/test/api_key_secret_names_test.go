package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestAPIKeySecretReferencesFollowDedicatedHelper(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	replaceAPIKeySecretName(t, chart)
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, split := range []bool{false, true} {
		for _, receiverSplit := range []bool{false, true} {
			for _, mode := range []string{"internal", "external", "absent"} {
				t.Run(fmt.Sprintf("%s/server=%t/receiver=%t", mode, split, receiverSplit), func(t *testing.T) {
					values := map[string]string{
						"stackstate.features.server.split":             fmt.Sprint(split),
						"stackstate.components.receiver.split.enabled": fmt.Sprint(receiverSplit),
						"global.receiverApiKey":                        "fixture-api-key",
						"global.suseObservability.receiverApiKey":      "",
						"stackstate.apiKey.key":                        "",
						"stackstate.receiver.apiKey":                   "",
					}
					if mode == "external" {
						values["stackstate.apiKey.fromExternalSecret"] = "customer-api-key"
					} else if mode == "absent" {
						values["global.receiverApiKey"] = ""
					}
					options := apiResourceNameTestOptions(values)
					options.ValuesFiles = []string{valuesFile}
					before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
					output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
					require.NoError(t, err)
					after := helmtestutil.NewKubernetesResources(t, output)
					legacy, explicit := "nightly-suse-observability-api-key", "explicit-api-key-secret"
					selected := explicit
					if mode == "internal" {
						require.Contains(t, before.Secrets, legacy)
						require.Contains(t, after.Secrets, explicit)
						expected := before.Secrets[legacy]
						expected.Name = explicit
						assert.Equal(t, "fixture-api-key", string(expected.Data["API_KEY"]))
						assert.Equal(t, expected, after.Secrets[explicit])
						delete(before.Secrets, legacy)
						delete(after.Secrets, explicit)
					} else {
						assert.NotContains(t, before.Secrets, legacy)
						assert.NotContains(t, after.Secrets, explicit)
						if mode == "external" {
							selected = "customer-api-key"
							assert.NotContains(t, after.Secrets, selected)
						}
					}
					assert.NotContains(t, after.Secrets, legacy)
					assert.Equal(t, before.Secrets, after.Secrets, "Other credentials must stay unchanged")
					references := 0
					for name, deployment := range before.Deployments {
						require.Contains(t, after.Deployments, name)
						expected := deployment.DeepCopy()
						if mode != "external" {
							for i := range expected.Spec.Template.Spec.Containers {
								for j := range expected.Spec.Template.Spec.Containers[i].EnvFrom {
									ref := expected.Spec.Template.Spec.Containers[i].EnvFrom[j].SecretRef
									if ref != nil && ref.Name == legacy {
										ref.Name = explicit
									}
								}
							}
						}
						if mode == "internal" {
							const checksum = "checksum/api-key-env"
							if old, ok := expected.Spec.Template.Annotations[checksum]; ok {
								actual := after.Deployments[name].Spec.Template.Annotations
								require.Contains(t, actual, checksum)
								assert.NotEqual(t, old, actual[checksum], "Deliberate test renames change the rendered Secret checksum")
								expected.Spec.Template.Annotations[checksum] = actual[checksum]
							}
						}
						assert.Equal(t, *expected, after.Deployments[name])
						for _, container := range after.Deployments[name].Spec.Template.Spec.Containers {
							for _, env := range container.EnvFrom {
								if env.SecretRef != nil && env.SecretRef.Name == selected {
									require.NotNil(t, env.SecretRef.Optional)
									assert.True(t, *env.SecretRef.Optional)
									references++
								}
							}
						}
					}
					expectedReferences := 2 // API/server plus the unsplit receiver.
					if receiverSplit {
						expectedReferences = 4 // API/server plus three receiver variants.
					}
					assert.Equal(t, expectedReferences, references)
					assert.Len(t, after.Deployments, len(before.Deployments))
					assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
					assert.Equal(t, before.Services, after.Services)
					assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
					assert.Equal(t, before.Roles, after.Roles)
					assert.Equal(t, before.RoleBindings, after.RoleBindings)
					assert.Equal(t, before.Statefulsets, after.Statefulsets)
					assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
				})
			}
		}
	}
}

func replaceAPIKeySecretName(t *testing.T, chart string) {
	t.Helper()
	path := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	definition := regexp.MustCompile(`(?s)\{\{- define "stackstate.apiKey.secret.fullname" -\}\}.*?\{\{- end -\}\}`)
	require.Len(t, definition.FindAll(data, -1), 1)
	data = definition.ReplaceAll(data, []byte(`{{- define "stackstate.apiKey.secret.fullname" -}}explicit-api-key-secret{{- end -}}`))
	require.NoError(t, os.WriteFile(path, data, 0600))
}
