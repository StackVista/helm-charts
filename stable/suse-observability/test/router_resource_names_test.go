package test

import (
	"fmt"
	"maps"
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

func TestRouterWorkloadReferencesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	for resource, name := range map[string]string{
		"deployment":     "explicit-router-workload",
		"serviceaccount": "explicit-router-account",
		"configmap":      "explicit-router-bootstrap",
		"secret":         "explicit-router-environment",
	} {
		helper := "stackstate.router." + resource + ".fullname"
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAllString(content, -1), 1, helper)
		content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}`+name+`{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, mode := range []string{"active", "maintenance", "automatic"} {
		for _, split := range []bool{false, true} {
			for _, secret := range []bool{false, true} {
				t.Run(fmt.Sprintf("mode=%s/server-split=%t/secret=%t", mode, split, secret), func(t *testing.T) {
					values := map[string]string{
						"stackstate.features.server.split":                         fmt.Sprint(split),
						"stackstate.components.router.mode.status":                 mode,
						"stackstate.components.router.extraEnv.open.ROUTER_PUBLIC": "public-value",
					}
					if secret {
						values["stackstate.components.router.extraEnv.secret.ROUTER_PRIVATE"] = "private-value"
					}
					options := apiResourceNameTestOptions(values)
					before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
					options.ValuesFiles = []string{valuesFile}
					output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
					require.NoError(t, err)
					after := helmtestutil.NewKubernetesResources(t, output)

					legacy := "nightly-suse-observability-router"
					require.Contains(t, before.Deployments, legacy)
					require.Contains(t, after.Deployments, "explicit-router-workload")
					assert.NotContains(t, after.Deployments, legacy)
					originalDeployment := before.Deployments[legacy]
					expectedDeployment := originalDeployment.DeepCopy()
					actualDeployment := after.Deployments["explicit-router-workload"]
					expectedDeployment.Name = "explicit-router-workload"
					require.Equal(t, legacy, expectedDeployment.Spec.Template.Spec.ServiceAccountName)
					expectedDeployment.Spec.Template.Spec.ServiceAccountName = "explicit-router-account"
					staticVolumes, dynamicVolumes, secretRefs := 0, 0, 0
					for i := range expectedDeployment.Spec.Template.Spec.Volumes {
						volume := &expectedDeployment.Spec.Template.Spec.Volumes[i]
						if volume.Name == "config" {
							require.NotNil(t, volume.ConfigMap)
							require.Equal(t, legacy, volume.ConfigMap.Name)
							volume.ConfigMap.Name = "explicit-router-bootstrap"
							staticVolumes++
						}
						if volume.Name == "router-mode" {
							require.NotNil(t, volume.ConfigMap)
							require.Equal(t, legacy+"-"+mode, volume.ConfigMap.Name)
							dynamicVolumes++
						}
					}
					require.Equal(t, 1, staticVolumes)
					require.Equal(t, 1, dynamicVolumes)
					for i := range expectedDeployment.Spec.Template.Spec.Containers {
						for j := range expectedDeployment.Spec.Template.Spec.Containers[i].Env {
							env := &expectedDeployment.Spec.Template.Spec.Containers[i].Env[j]
							if env.Name == "ROUTER_PRIVATE" {
								require.NotNil(t, env.ValueFrom)
								require.NotNil(t, env.ValueFrom.SecretKeyRef)
								require.Equal(t, legacy, env.ValueFrom.SecretKeyRef.Name)
								env.ValueFrom.SecretKeyRef.Name = "explicit-router-environment"
								secretRefs++
							}
						}
					}
					// Deliberately changing the static ConfigMap name also changes
					// its rendered-manifest checksum. Normal extraction preserves it.
					checksum := "checksum/router-configmap"
					require.Contains(t, expectedDeployment.Spec.Template.Annotations, checksum)
					require.Contains(t, actualDeployment.Spec.Template.Annotations, checksum)
					assert.NotEqual(t, expectedDeployment.Spec.Template.Annotations[checksum], actualDeployment.Spec.Template.Annotations[checksum])
					expectedDeployment.Spec.Template.Annotations[checksum] = actualDeployment.Spec.Template.Annotations[checksum]
					assert.Equal(t, *expectedDeployment, actualDeployment)

					require.Contains(t, before.ConfigMaps, legacy)
					require.Contains(t, after.ConfigMaps, "explicit-router-bootstrap")
					assert.NotContains(t, after.ConfigMaps, legacy)
					expectedConfig := before.ConfigMaps[legacy]
					expectedConfig.Name = "explicit-router-bootstrap"
					assert.Equal(t, expectedConfig, after.ConfigMaps["explicit-router-bootstrap"])

					require.Contains(t, before.ServiceAccounts, legacy)
					require.Contains(t, after.ServiceAccounts, "explicit-router-account")
					assert.NotContains(t, after.ServiceAccounts, legacy)
					expectedAccount := before.ServiceAccounts[legacy]
					expectedAccount.Name = "explicit-router-account"
					assert.Equal(t, expectedAccount, after.ServiceAccounts["explicit-router-account"])

					if secret {
						require.Equal(t, 1, secretRefs)
						require.Contains(t, before.Secrets, legacy)
						require.Contains(t, after.Secrets, "explicit-router-environment")
						expectedSecret := before.Secrets[legacy]
						expectedSecret.Name = "explicit-router-environment"
						assert.Equal(t, expectedSecret, after.Secrets["explicit-router-environment"])
						delete(before.Secrets, legacy)
						delete(after.Secrets, "explicit-router-environment")
					} else {
						assert.Zero(t, secretRefs)
						assert.NotContains(t, before.Secrets, legacy)
						assert.NotContains(t, after.Secrets, "explicit-router-environment")
					}
					assert.NotContains(t, after.Secrets, legacy)
					if mode == "automatic" {
						scriptsName := legacy + "-mode-scripts"
						require.Contains(t, before.ConfigMaps, scriptsName)
						require.Contains(t, after.ConfigMaps, scriptsName)
						expectedScripts := before.ConfigMaps[scriptsName]
						expectedScripts.Data = maps.Clone(expectedScripts.Data)
						for _, key := range []string{"set-active.sh", "set-maintenance.sh"} {
							require.Contains(t, expectedScripts.Data, key)
							script := expectedScripts.Data[key]
							for _, prefix := range []string{`deployment "`, `"deployment/`} {
								previous := prefix + legacy + `"`
								require.Equal(t, 2, strings.Count(script, previous))
								script = strings.ReplaceAll(script, previous, prefix+`explicit-router-workload"`)
							}
							expectedScripts.Data[key] = script
						}
						assert.Equal(t, expectedScripts, after.ConfigMaps[scriptsName], "Only the Deployment targets in hook scripts may change")
						delete(before.ConfigMaps, scriptsName)
						delete(after.ConfigMaps, scriptsName)
					}
					delete(before.Deployments, legacy)
					delete(after.Deployments, "explicit-router-workload")
					delete(before.ConfigMaps, legacy)
					delete(after.ConfigMaps, "explicit-router-bootstrap")
					delete(before.ServiceAccounts, legacy)
					delete(after.ServiceAccounts, "explicit-router-account")
					assert.Equal(t, before.Deployments, after.Deployments)
					assert.Equal(t, before.ConfigMaps, after.ConfigMaps, "Dynamic mode configuration and hook scripts must stay unchanged")
					assert.Equal(t, before.Secrets, after.Secrets)
					assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
					assert.Equal(t, before.Services, after.Services)
					assert.Equal(t, before.Roles, after.Roles)
					assert.Equal(t, before.RoleBindings, after.RoleBindings)
					assert.Equal(t, before.Statefulsets, after.Statefulsets)
					assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
					assert.Equal(t, before.Pdbs, after.Pdbs)
				})
			}
		}
	}
}
