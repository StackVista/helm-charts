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

func TestPDBNamesUseDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	components := map[string]string{
		"api": "api", "authorizationSync": "authorization-sync", "checks": "checks",
		"correlate": "correlate", "e2es": "e2es", "healthSync": "health-sync",
		"notification": "notification", "receiver": "receiver", "router": "router",
		"server": "server", "state": "state", "sync": "sync", "ui": "ui",
		"vmagent": "vmagent", "workloadObserver": "workload-observer", "mcp": "mcp",
		"aiAssistant": "ai-assistant", "victoriametrics": "victoriametrics",
	}
	for component, suffix := range components {
		helper := "stackstate." + component + ".poddisruptionbudget.fullname"
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAll(data, -1), 1, helper)
		data = definition.ReplaceAll(data, []byte(`{{- define "`+helper+`" -}}explicit-`+suffix+`-budget{{- end -}}`))
	}
	require.NoError(t, os.WriteFile(namesPath, data, 0600))
	fullValues, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	seen := map[string]bool{}
	for _, tc := range []struct {
		name                                                         string
		split, workers, assistant, mcp, observer, authorization, vm1 bool
		budget                                                       string
	}{
		{"split", true, false, true, true, true, true, true, "0"},
		{"mono", false, false, true, true, true, true, true, "40%"},
		{"split-workers", true, true, true, true, true, true, true, "40%"},
		{"optional-disabled", true, false, false, false, false, false, false, "0"},
		{"mcp-only", true, false, false, true, false, true, true, "0"},
		{"assistant-enables-mcp", false, true, true, false, true, true, true, "40%"},
	} {
		for _, policyV1 := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/policy-v1=%t", tc.name, policyV1), func(t *testing.T) {
				values := componentResourceNameTestValues()
				values["stackstate.features.server.split"] = fmt.Sprint(tc.split)
				values["stackstate.components.receiver.split.enabled"] = fmt.Sprint(tc.workers)
				values["stackstate.components.correlate.split.enabled"] = fmt.Sprint(tc.workers)
				values["ai.assistant.enabled"] = fmt.Sprint(tc.assistant)
				values["ai.mcp.enabled"] = fmt.Sprint(tc.mcp)
				values["stackstate.components.workloadObserver.enabled"] = fmt.Sprint(tc.observer)
				values["stackstate.k8sAuthorization.enabled"] = fmt.Sprint(tc.authorization)
				values["victoria-metrics-1.enabled"] = fmt.Sprint(tc.vm1)
				values["poddisruptionbudget.annotations.naming-test"] = "retained"
				for component := range components {
					if component != "victoriametrics" {
						values["stackstate.components."+component+".poddisruptionbudget.maxUnavailable"] = tc.budget
					}
				}
				options := apiResourceNameTestOptions(values)
				options.ValuesFiles = []string{fullValues}
				var args []string
				apiVersion := "policy/v1beta1"
				if policyV1 {
					args = []string{"--api-versions", "policy/v1/PodDisruptionBudget"}
					apiVersion = "policy/v1"
				}
				before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "nightly", options, args...))
				output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil, args...)
				require.NoError(t, err)
				after := helmtestutil.NewKubernetesResources(t, output)
				for component, suffix := range components {
					legacy, explicit := "nightly-suse-observability-"+suffix, "explicit-"+suffix+"-budget"
					enabled := true
					switch component {
					case "api", "checks", "healthSync", "notification", "state", "sync":
						enabled = tc.split
					case "server":
						enabled = !tc.split
					case "authorizationSync":
						enabled = tc.split && tc.authorization
					case "aiAssistant":
						enabled = tc.assistant
					case "mcp":
						enabled = tc.assistant || tc.mcp
					case "workloadObserver":
						enabled = tc.observer
					case "victoriametrics":
						enabled = tc.vm1
					}
					if !enabled {
						assert.NotContains(t, before.Pdbs, legacy)
						assert.NotContains(t, after.Pdbs, explicit)
						continue
					}
					require.Contains(t, before.Pdbs, legacy)
					require.Contains(t, after.Pdbs, explicit)
					expected := before.Pdbs[legacy]
					expected.Name = explicit
					assert.Equal(t, apiVersion, expected.APIVersion)
					assert.Equal(t, "retained", expected.Annotations["naming-test"])
					require.NotNil(t, expected.Spec.MaxUnavailable)
					if component != "victoriametrics" {
						assert.Equal(t, tc.budget, expected.Spec.MaxUnavailable.String())
					}
					// Include the complete selector, budgets, annotations, labels and API version.
					// In split mode receiver/correlate still have a single unchanged PDB each.
					assert.Equal(t, expected, after.Pdbs[explicit])
					assert.NotContains(t, after.Pdbs, legacy)
					delete(before.Pdbs, legacy)
					delete(after.Pdbs, explicit)
					seen[component] = true
				}
				assert.Equal(t, before.Pdbs, after.Pdbs, "subchart-owned PDBs must remain unchanged")
				assert.Equal(t, before.Deployments, after.Deployments)
				assert.Equal(t, before.Statefulsets, after.Statefulsets)
				assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
				assert.Equal(t, before.Services, after.Services)
				assert.Equal(t, before.ServiceMonitors, after.ServiceMonitors)
				assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
				assert.Equal(t, before.Secrets, after.Secrets)
			})
		}
	}
	for component := range components {
		assert.True(t, seen[component], "PDB was not exercised: %s", component)
	}
}

func TestPDBRequiresResolvedName(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	definition := regexp.MustCompile(`(?s)\{\{- define "stackstate.api.poddisruptionbudget.fullname" -\}\}.*?\{\{- end -\}\}`)
	require.Len(t, definition.FindAll(data, -1), 1)
	data = definition.ReplaceAll(data, []byte(`{{- define "stackstate.api.poddisruptionbudget.fullname" -}}{{- end -}}`))
	require.NoError(t, os.WriteFile(namesPath, data, 0600))
	values, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	options := apiResourceNameTestOptions(map[string]string{"stackstate.features.server.split": "true"})
	options.ValuesFiles = []string{values}
	_, err = helm.RenderTemplateE(t, options, chart, "nightly", nil)
	require.ErrorContains(t, err, "stackstate.service.spec.poddisruptionbudget: PdbFullname must not be empty")
}
