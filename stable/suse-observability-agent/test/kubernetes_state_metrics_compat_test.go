package test

import (
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
	assert.NotContains(t, section(t, config, "processors"), "memory_limiter")
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
