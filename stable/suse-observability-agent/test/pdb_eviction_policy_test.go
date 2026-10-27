package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestPDBUnhealthyEvictionPolicy(t *testing.T) {
	policyKeys := map[string]string{
		"suse-observability-agent":                        "clusterAgent.podDisruptionBudget.unhealthyPodEvictionPolicy",
		"suse-observability-agent-checks-agent":           "checksAgent.podDisruptionBudget.unhealthyPodEvictionPolicy",
		otelMetricsScraperName:                            "otel.prometheusScraping.collector.podDisruptionBudget.unhealthyPodEvictionPolicy",
		otelTargetAllocatorName:                           "otel.prometheusScraping.targetAllocator.podDisruptionBudget.unhealthyPodEvictionPolicy",
		"suse-observability-agent-k8s-resource-collector": "otel.k8sResourceCollector.podDisruptionBudget.unhealthyPodEvictionPolicy",
		otelTelemetryGatewayName:                          "otel.telemetryGateway.podDisruptionBudget.unhealthyPodEvictionPolicy",
	}
	type scenario struct {
		name, version, override, expected, apiVersion, component string
	}
	scenarios := []scenario{
		{"legacy-api", "1.20.0", "", "", "policy/v1beta1", ""},
		{"stable-api", "1.21.0", "", "", "policy/v1", ""},
		{"alpha-disabled", "1.26.0", "", "", "policy/v1", ""},
		{"beta-default", "1.27.0", "", "AlwaysAllow", "policy/v1", ""},
		{"stable-default", "1.31.0", "", "AlwaysAllow", "policy/v1", ""},
		{"newer-default", "1.34.0", "", "AlwaysAllow", "policy/v1", ""},
		{"newer-default-1.35", "1.35.0", "", "AlwaysAllow", "policy/v1", ""},
		{"newer-default-1.36", "1.36.0", "", "AlwaysAllow", "policy/v1", ""},
		{"override", "1.31.0", "IfHealthyBudget", "IfHealthyBudget", "policy/v1", ""},
		{"unset", "1.31.0", "null", "", "policy/v1", ""},
		{"legacy-override", "1.26.0", "AlwaysAllow", "", "policy/v1", ""},
	}
	for name, key := range policyKeys {
		scenarios = append(scenarios, scenario{
			"override-" + name, "1.31.0", "IfHealthyBudget", "IfHealthyBudget", "policy/v1", key,
		})
	}
	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{
				"otel.enabled":                                         "true",
				"otel.k8sResourceCollector.replicaCount":               "2",
				"otel.prometheusScraping.collector.replicaCount":       "2",
				"otel.prometheusScraping.targetAllocator.replicaCount": "2",
				"otel.telemetryGateway.enabled":                        "true",
				"otel.telemetryGateway.replicaCount":                   "2",
			}
			if tc.override != "" {
				for _, key := range policyKeys {
					if tc.component == "" || tc.component == key {
						values[key] = tc.override
					}
				}
			}
			output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "suse-observability-agent", &helm.Options{
				ValuesFiles: []string{"values/minimal.yaml"},
				SetValues:   values,
			}, "--kube-version", tc.version)
			resources := helmtestutil.NewKubernetesResources(t, output)
			require.Len(t, resources.Pdbs, len(policyKeys))
			for name, key := range policyKeys {
				pdb, ok := resources.Pdbs[name]
				require.True(t, ok, name)
				assert.Equal(t, tc.apiVersion, pdb.APIVersion, name)
				expected := tc.expected
				if tc.component != "" && tc.component != key {
					expected = "AlwaysAllow"
				}
				if expected == "" {
					assert.Nil(t, pdb.Spec.UnhealthyPodEvictionPolicy, name)
				} else {
					require.NotNil(t, pdb.Spec.UnhealthyPodEvictionPolicy, name)
					assert.Equal(t, expected, string(*pdb.Spec.UnhealthyPodEvictionPolicy), name)
				}
				require.NotNil(t, pdb.Spec.MaxUnavailable, name)
				assert.Equal(t, int32(1), pdb.Spec.MaxUnavailable.IntVal, name)
				assert.Nil(t, pdb.Spec.MinAvailable, name)
				assert.Equal(t, "suse-observability-agent",
					pdb.Spec.Selector.MatchLabels["app.kubernetes.io/instance"], name)
				assert.NotEmpty(t, pdb.Spec.Selector.MatchLabels["app.kubernetes.io/component"], name)
			}
		})
	}
}
