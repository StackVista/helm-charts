package test

import (
	"slices"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func clusterAgentKSMConf(t *testing.T, resources helmtestutil.KubernetesResources) (string, bool) {
	t.Helper()
	configMap, ok := resources.ConfigMaps[clusterAgentName]
	require.True(t, ok, "cluster agent config map was not found")
	conf, ok := configMap.Data["kube_state_metrics_core_conf"]
	return conf, ok
}

func clusterAgentMountsKSMConf(t *testing.T, resources helmtestutil.KubernetesResources) bool {
	t.Helper()
	for _, mount := range resources.Deployments[clusterAgentName].Spec.Template.Spec.Containers[0].VolumeMounts {
		if mount.SubPath == "kube_state_metrics_core_conf" {
			return true
		}
	}
	return false
}

func TestKubernetesStateMetricsCompatDisabledByDefault(t *testing.T) {
	resources := renderCompat(t, &helm.Options{ValuesFiles: []string{"values/k8s-resource-collector-enabled.yaml"}})
	config := collectorConfigMap(t, resources)
	assert.NotContains(t, section(t, config, "receivers"), "stsksm")
	assert.NotContains(t, section(t, config, "service", "pipelines"), "metrics/kubernetes-state")
	_, ok := clusterAgentKSMConf(t, resources)
	assert.True(t, ok)
	assert.True(t, clusterAgentMountsKSMConf(t, resources))
}

func TestKubernetesStateMetricsCompatEnabled(t *testing.T) {
	resources := renderCompat(t, &helm.Options{ValuesFiles: []string{"values/kubernetes-events-compat.yaml"}})
	config := collectorConfigMap(t, resources)

	receiver := section(t, config, "receivers", "stsksm")
	assert.Equal(t, "${env:K8S_CLUSTER_NAME}", receiver["cluster_name"])
	assert.Equal(t, "k8s_leader_elector", receiver["k8s_leader_elector"])
	collectors, ok := receiver["collectors"].([]any)
	require.True(t, ok)
	assert.Len(t, collectors, 22)
	assert.NotContains(t, receiver, "labels_as_tags")

	pipeline := section(t, config, "service", "pipelines", "metrics/kubernetes-state")
	assert.Equal(t, []any{"stsksm"}, pipeline["receivers"])
	assert.Equal(t, []any{"memory_limiter", "batch"}, pipeline["processors"])
	assert.Contains(t, section(t, config, "processors"), "memory_limiter")
	assert.Len(t, pipeline["exporters"], 1)

	_, ok = clusterAgentKSMConf(t, resources)
	assert.False(t, ok, "the cluster agent's KSM check is off")
	assert.False(t, clusterAgentMountsKSMConf(t, resources))
}

func TestKubernetesStateMetricsCompatFollowsClusterAgentSettings(t *testing.T) {
	resources := renderCompat(t, &helm.Options{
		ValuesFiles: []string{"values/kubernetes-events-compat.yaml"},
		SetValues: map[string]string{
			"clusterAgent.collection.kubernetesResources.secrets":            "false",
			"clusterAgent.collection.kubeStateMetrics.labelsAsTags.pods.app": "app",
		},
	})
	receiver := section(t, collectorConfigMap(t, resources), "receivers", "stsksm")
	collectors, ok := receiver["collectors"].([]any)
	require.True(t, ok)
	assert.Len(t, collectors, 21)
	assert.NotContains(t, collectors, "secrets")
	assert.Equal(t, map[string]any{"pods": map[string]any{"app": "app"}}, receiver["labels_as_tags"])
}

func TestKubernetesStateMetricsCompatOffWithoutKubeStateMetrics(t *testing.T) {
	resources := renderCompat(t, &helm.Options{
		ValuesFiles: []string{"values/kubernetes-events-compat.yaml"},
		SetValues:   map[string]string{"clusterAgent.collection.kubeStateMetrics.enabled": "false"},
	})
	assert.NotContains(t, section(t, collectorConfigMap(t, resources), "receivers"), "stsksm")
	_, ok := clusterAgentKSMConf(t, resources)
	assert.False(t, ok)
}

func TestKubernetesStateMetricsCompatRestrictedRBAC(t *testing.T) {
	resources := renderCompat(t, &helm.Options{
		ValuesFiles: []string{"values/kubernetes-events-compat.yaml"},
		SetValues:   map[string]string{"otel.k8sResourceCollector.rbac.useWildcard": "false"},
	})
	clusterRole, ok := resources.ClusterRoles[resourceCollectorName]
	require.True(t, ok)
	granted := map[string]bool{}
	for _, rule := range clusterRole.Rules {
		for _, group := range rule.APIGroups {
			for _, resource := range rule.Resources {
				granted[group+"/"+resource] = true
			}
		}
	}
	for _, want := range []string{
		"/nodes", "/pods", "/services", "/persistentvolumeclaims", "/persistentvolumes", "/namespaces",
		"/endpoints", "/secrets", "/resourcequotas", "/replicationcontrollers", "/limitranges",
		"apps/daemonsets", "apps/deployments", "apps/replicasets", "apps/statefulsets",
		"batch/cronjobs", "batch/jobs", "networking.k8s.io/ingresses",
		"autoscaling/horizontalpodautoscalers", "policy/poddisruptionbudgets",
		"storage.k8s.io/storageclasses", "storage.k8s.io/volumeattachments",
	} {
		assert.True(t, granted[want], "%s must be listable for kube-state-metrics", want)
	}
}

func TestClusterAgentFlagFollowsCollectionSwitches(t *testing.T) {
	for name, tc := range map[string]struct {
		values    map[string]string
		receivers []string
		pipelines []string
	}{
		"topology off": {
			values:    map[string]string{"clusterAgent.collection.kubernetesTopology": "false"},
			receivers: []string{"k8sobjects/kubernetes-events", "stsksm"},
			pipelines: []string{"logs/kubernetes-events", "metrics/kubernetes-state"},
		},
		"events off": {
			values:    map[string]string{"clusterAgent.collection.kubernetesEvents": "false"},
			receivers: []string{"stsksm"},
			pipelines: []string{"logs/kubernetes-topology", "metrics/kubernetes-state"},
		},
		"cluster agent off": {
			values: map[string]string{
				"clusterAgent.enabled": "false", "nodeAgent.enabled": "false", "checksAgent.enabled": "false",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			resources := renderCompat(t, &helm.Options{
				ValuesFiles: []string{"values/kubernetes-events-compat.yaml"},
				SetValues:   tc.values,
			})
			config := collectorConfigMap(t, resources)
			receivers := section(t, config, "receivers")
			for _, r := range []string{"k8sobjects/kubernetes-events", "stsksm"} {
				assert.Equal(t, slices.Contains(tc.receivers, r), receivers[r] != nil, r)
			}
			pipelines := section(t, config, "service", "pipelines")
			for _, p := range []string{"logs/kubernetes-topology", "logs/kubernetes-events", "metrics/kubernetes-state"} {
				assert.Equal(t, slices.Contains(tc.pipelines, p), pipelines[p] != nil, p)
			}
		})
	}
}

func TestMemoryLimiterOnlyWhereTheSourceResends(t *testing.T) {
	for name, values := range map[string]string{
		"defaults":           "values/k8s-resource-collector-enabled.yaml",
		"cluster agent flag": "values/kubernetes-events-compat.yaml",
	} {
		t.Run(name, func(t *testing.T) {
			config := collectorConfigMap(t, renderCompat(t, &helm.Options{ValuesFiles: []string{values}}))
			assert.Contains(t, section(t, config, "processors"), "memory_limiter")
			for pipeline, raw := range section(t, config, "service", "pipelines") {
				processors, _ := raw.(map[string]any)["processors"].([]any)
				limited := slices.Contains(processors, any("memory_limiter"))
				switch pipeline {
				case "logs/kubernetes-topology", "logs/kubernetes-events":
					assert.False(t, limited, "%s must not refuse data", pipeline)
				default:
					assert.True(t, limited, "%s should refuse data near the memory limit", pipeline)
					assert.Equal(t, "memory_limiter", processors[0], "%s: the limiter goes first", pipeline)
				}
			}
		})
	}
}
