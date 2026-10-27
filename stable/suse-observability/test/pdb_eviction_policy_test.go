package test

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	policyv1 "k8s.io/api/policy/v1"
)

func TestPDBUnhealthyEvictionDefaults(t *testing.T) {
	for _, version := range []string{"1.26.0", "1.34.0", "1.35.0", "1.36.0"} {
		for _, split := range []string{"true", "false"} {
			t.Run(version+"/split="+split, func(t *testing.T) {
				output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "suse-observability", &helm.Options{
					ValuesFiles: []string{"values/full.yaml"},
					SetValues: map[string]string{
						"stackstate.features.server.split":                         split,
						"stackstate.components.workloadObserver.enabled":           "true",
						"stackstate.k8sAuthorization.enabled":                      "true",
						"ai.mcp.enabled":                                           "true",
						"opentelemetry-collector.mode":                             "deployment",
						"opentelemetry-collector.podDisruptionBudget.enabled":      "true",
						"opentelemetry-collector.podDisruptionBudget.minAvailable": "1",
					},
				}, "--kube-version", version, "--api-versions", "policy/v1/PodDisruptionBudget")

				applicationCount, dependencyCount, collectorCount := 0, 0, 0
				for _, document := range strings.Split(output, "\n---\n") {
					if !strings.Contains(document, "\nkind: PodDisruptionBudget\n") {
						continue
					}
					var pdb policyv1.PodDisruptionBudget
					require.NoError(t, helm.UnmarshalK8SYamlE(t, document, &pdb))
					application := strings.Contains(document, "# Source: suse-observability/templates/")
					collector := strings.Contains(document, "/charts/opentelemetry-collector/")
					component := pdb.Labels["app.kubernetes.io/component"]
					eligible := (application && component != "ai-assistant" && component != "victoriametrics") || collector
					if eligible {
						require.NotNil(t, pdb.Spec.UnhealthyPodEvictionPolicy, pdb.Name)
						assert.Equal(t, policyv1.AlwaysAllow, *pdb.Spec.UnhealthyPodEvictionPolicy, pdb.Name)
						assert.Equal(t, "policy/v1", pdb.APIVersion, pdb.Name)
					} else {
						assert.Nil(t, pdb.Spec.UnhealthyPodEvictionPolicy, pdb.Name)
					}
					if application {
						applicationCount++
						require.NotNil(t, pdb.Spec.MaxUnavailable, pdb.Name)
						assert.Equal(t, int32(1), pdb.Spec.MaxUnavailable.IntVal, pdb.Name)
						if component != "victoriametrics" {
							assert.Equal(t, component, pdb.Spec.Selector.MatchLabels["app.kubernetes.io/component"], pdb.Name)
						}
					} else {
						dependencyCount++
					}
					if collector {
						collectorCount++
					}
				}
				if split == "true" {
					assert.Equal(t, 17, applicationCount)
				} else {
					assert.Equal(t, 11, applicationCount)
				}
				assert.Equal(t, 1, collectorCount)
				assert.GreaterOrEqual(t, dependencyCount, 10, "Stateful dependency PDBs must be exercised")
			})
		}
	}
}

func TestPDBUnhealthyEvictionOverrides(t *testing.T) {
	for _, version := range []string{"1.26.0", "1.34.0", "1.35.0", "1.36.0"} {
		t.Run(version, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{"values/full.yaml"},
				SetValues: map[string]string{
					"stackstate.components.sync.poddisruptionbudget.unhealthyPodEvictionPolicy":     "IfHealthyBudget",
					"stackstate.components.sync.poddisruptionbudget.maxUnavailable":                 "0",
					"stackstate.components.receiver.poddisruptionbudget.unhealthyPodEvictionPolicy": "null",
					"stackstate.components.receiver.poddisruptionbudget.maxUnavailable":             "null",
					"stackstate.components.receiver.poddisruptionbudget.minAvailable":               "1",
					"opentelemetry-collector.mode":                                                  "deployment",
					"opentelemetry-collector.podDisruptionBudget.enabled":                           "true",
					"opentelemetry-collector.podDisruptionBudget.minAvailable":                      "1",
					"opentelemetry-collector.podDisruptionBudget.unhealthyPodEvictionPolicy":        "IfHealthyBudget",
				},
			}, "--kube-version", version, "--api-versions", "policy/v1/PodDisruptionBudget")
			resources := helmtestutil.NewKubernetesResources(t, output)
			sync := resources.Pdbs["suse-observability-sync"]
			require.NotNil(t, sync.Spec.MaxUnavailable)
			assert.Zero(t, sync.Spec.MaxUnavailable.IntVal)
			require.NotNil(t, sync.Spec.UnhealthyPodEvictionPolicy)
			assert.Equal(t, "IfHealthyBudget", string(*sync.Spec.UnhealthyPodEvictionPolicy))
			receiver := resources.Pdbs["suse-observability-receiver"]
			assert.Nil(t, receiver.Spec.UnhealthyPodEvictionPolicy)
			assert.Nil(t, receiver.Spec.MaxUnavailable)
			require.NotNil(t, receiver.Spec.MinAvailable)
			assert.Equal(t, int32(1), receiver.Spec.MinAvailable.IntVal)
			collectorFound := false
			for _, pdb := range resources.Pdbs {
				if pdb.Labels["app.kubernetes.io/name"] == "opentelemetry-collector" {
					collectorFound = true
					require.NotNil(t, pdb.Spec.UnhealthyPodEvictionPolicy)
					assert.Equal(t, "IfHealthyBudget", string(*pdb.Spec.UnhealthyPodEvictionPolicy))
				}
			}
			assert.True(t, collectorFound)
		})
	}
}
