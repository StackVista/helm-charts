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

func TestDeploymentNamesUseDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	components := map[string]string{
		"api": "api", "checks": "checks", "notification": "notification",
		"healthSync": "health-sync", "authorizationSync": "authorization-sync",
		"state": "state", "sync": "sync", "slicing": "slicing",
		"server": "server", "initializer": "initializer", "e2es": "e2es",
	}
	for component, suffix := range components {
		helper := "stackstate." + component + ".deployment.fullname"
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAllString(content, -1), 1, helper)
		content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}explicit-`+suffix+`-deployment{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, tc := range []struct {
		name, split, authorization string
	}{
		{"split", "true", "true"},
		{"monolithic", "false", "true"},
		{"authorization-disabled", "true", "false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := apiResourceNameTestOptions(map[string]string{
				"stackstate.features.server.split":                         tc.split,
				"stackstate.k8sAuthorization.enabled":                      tc.authorization,
				"stackstate.components.all.extraEnv.secret.SHARED_SETTING": "shared-value",
				"stackstate.features.storeTransactionLogsToPVC.enabled":    "true",
			})
			before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
			options.ValuesFiles = []string{valuesFile}
			output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
			require.NoError(t, err)
			after := helmtestutil.NewKubernetesResources(t, output)

			for component, suffix := range components {
				legacy := "nightly-suse-observability-" + suffix
				explicit := "explicit-" + suffix + "-deployment"
				enabled := tc.split == "true"
				switch component {
				case "server":
					enabled = tc.split == "false"
				case "e2es":
					enabled = true
				case "authorizationSync":
					enabled = enabled && tc.authorization == "true"
				}
				if !enabled {
					assert.NotContains(t, before.Deployments, legacy)
					assert.NotContains(t, after.Deployments, explicit)
					continue
				}
				require.Contains(t, before.Deployments, legacy)
				require.Contains(t, after.Deployments, explicit)
				assert.NotContains(t, after.Deployments, legacy)
				expected := before.Deployments[legacy]
				expected.Name = explicit
				// Changing only the Deployment helper must not alter selectors,
				// checksums, pod templates or other resource references.
				assert.Equal(t, expected, after.Deployments[explicit])
			}
			assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
			assert.Equal(t, before.Secrets, after.Secrets)
			assert.Equal(t, before.Services, after.Services)
			assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
			assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
			assert.Equal(t, before.Statefulsets, after.Statefulsets)
		})
	}
}

func TestSplitDeploymentNamesUseDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	variants := map[string][]string{
		"receiver":  {"", "-base", "-logs", "-process-agent"},
		"correlate": {"", "-connection", "-http-tracing", "-aggregator"},
	}
	for component := range variants {
		helper := "stackstate." + component + ".deployment.fullname"
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAllString(content, -1), 1, helper)
		content = definition.ReplaceAllString(content,
			`{{- define "`+helper+`" -}}explicit-`+component+`-deployment{{ template "stackstate.`+component+`.name.postfix" . }}{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, serverSplit := range []bool{false, true} {
		for _, receiverSplit := range []bool{false, true} {
			for _, correlateSplit := range []bool{false, true} {
				t.Run(fmt.Sprintf("server=%t/receiver=%t/correlate=%t", serverSplit, receiverSplit, correlateSplit), func(t *testing.T) {
					values := componentResourceNameTestValues()
					values["stackstate.features.server.split"] = fmt.Sprint(serverSplit)
					values["stackstate.components.receiver.split.enabled"] = fmt.Sprint(receiverSplit)
					values["stackstate.components.correlate.split.enabled"] = fmt.Sprint(correlateSplit)
					values["stackstate.components.all.metrics.servicemonitor.enabled"] = "true"
					options := apiResourceNameTestOptions(values)
					before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
					options.ValuesFiles = []string{valuesFile}
					output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
					require.NoError(t, err)
					after := helmtestutil.NewKubernetesResources(t, output)

					for component, suffixes := range variants {
						split := receiverSplit
						if component == "correlate" {
							split = correlateSplit
						}
						for _, suffix := range suffixes {
							legacy := "nightly-suse-observability-" + component + suffix
							explicit := "explicit-" + component + "-deployment" + suffix
							if split != (suffix != "") {
								assert.NotContains(t, before.Deployments, legacy)
								assert.NotContains(t, after.Deployments, explicit)
								continue
							}
							require.Contains(t, before.Deployments, legacy)
							require.Contains(t, after.Deployments, explicit)
							assert.NotContains(t, after.Deployments, legacy)
							expected := before.Deployments[legacy]
							expected.Name = explicit
							assert.Equal(t, expected, after.Deployments[explicit])
						}
					}
					assert.Len(t, after.Deployments, len(before.Deployments))
					assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
					assert.Equal(t, before.Secrets, after.Secrets)
					assert.Equal(t, before.Services, after.Services)
					assert.Equal(t, before.ServiceMonitors, after.ServiceMonitors)
					assert.Equal(t, before.Pdbs, after.Pdbs)
					assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
					assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
					assert.Equal(t, before.Statefulsets, after.Statefulsets)
				})
			}
		}
	}
}
