package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func collectorConfigMap(t *testing.T, resources helmtestutil.KubernetesResources) map[string]any {
	t.Helper()
	configMap, ok := resources.ConfigMaps[resourceCollectorConfigName]
	require.True(t, ok, "k8s-resource-collector config map was not found")
	var config map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(configMap.Data["config.yaml"]), &config))
	return config
}

func section(t *testing.T, config map[string]any, keys ...string) map[string]any {
	t.Helper()
	current := config
	for _, key := range keys {
		next, ok := current[key].(map[string]any)
		require.True(t, ok, "%v not found", keys)
		current = next
	}
	return current
}

func clusterAgentEventsConf(t *testing.T, resources helmtestutil.KubernetesResources) string {
	t.Helper()
	configMap, ok := resources.ConfigMaps[clusterAgentName]
	require.True(t, ok, "cluster agent config map was not found")
	return configMap.Data["kubernetes_api_events_conf"]
}

func TestKubernetesEventsCompatDisabledByDefault(t *testing.T) {
	resources := renderCompat(t, &helm.Options{ValuesFiles: []string{"values/k8s-resource-collector-enabled.yaml"}})
	config := collectorConfigMap(t, resources)
	assert.NotContains(t, section(t, config, "receivers"), "k8sobjects/kubernetes-events")
	assert.NotContains(t, section(t, config, "exporters"), "stsk8sevents")
	assert.NotContains(t, section(t, config, "service", "pipelines"), "logs/kubernetes-events")
	assert.Contains(t, clusterAgentEventsConf(t, resources), "collect_events: true")
}

func TestKubernetesEventsCompatEnabled(t *testing.T) {
	resources := renderCompat(t, &helm.Options{
		ValuesFiles: []string{"values/kubernetes-events-compat.yaml"},
		SetValues:   map[string]string{"clusterAgent.config.events.categories.BackOff": "Alerts"},
	})
	config := collectorConfigMap(t, resources)

	receiver := section(t, config, "receivers", "k8sobjects/kubernetes-events")
	assert.Equal(t, "k8s_leader_elector", receiver["k8s_leader_elector"])
	objects, ok := receiver["objects"].([]any)
	require.True(t, ok)
	names := []string{}
	for _, object := range objects {
		watch, ok := object.(map[string]any)
		require.True(t, ok)
		names = append(names, watch["name"].(string))
		assert.Equal(t, "watch", watch["mode"])
		assert.Equal(t, []any{"DELETED"}, watch["exclude_watch_type"])
	}
	assert.ElementsMatch(t, []string{"events", "pods"}, names)

	exporter := section(t, config, "exporters", "stsk8sevents")
	assert.Equal(t, "${env:RECEIVER_INTAKE_URL}", exporter["endpoint"])
	assert.Equal(t, "${env:STS_API_KEY}", exporter["api_key"])
	assert.Equal(t, "${env:K8S_CLUSTER_NAME}", exporter["cluster_name"])
	assert.Equal(t, "kubernetes", exporter["cluster_type"])
	assert.Equal(t, map[string]any{"BackOff": "Alerts"}, exporter["event_categories"])

	pipeline := section(t, config, "service", "pipelines", "logs/kubernetes-events")
	assert.Equal(t, []any{"k8sobjects/kubernetes-events"}, pipeline["receivers"])
	assert.Equal(t, []any{"stsk8sevents"}, pipeline["exporters"])
	assert.NotContains(t, pipeline, "processors")

	env, ok := containerEnv(t, resources, resourceCollectorName, "RECEIVER_INTAKE_URL")
	require.True(t, ok)
	assert.Equal(t, "https://my-suse-observability-instance.com/receiver/stsAgent/intake", env.Value)
	assert.Contains(t, clusterAgentEventsConf(t, resources), "collect_events: false")

}

func TestKubernetesEventsCompatRestrictedRBAC(t *testing.T) {
	resources := renderCompat(t, &helm.Options{
		ValuesFiles: []string{"values/kubernetes-events-compat.yaml"},
		SetValues: map[string]string{
			"otel.k8sResourceCollector.rbac.useWildcard":             "false",
			"clusterAgent.collection.kubernetesResources.namespaces": "false",
		},
	})
	clusterRole, ok := resources.ClusterRoles[resourceCollectorName]
	require.True(t, ok)
	granted := map[string]bool{}
	for _, rule := range clusterRole.Rules {
		for _, group := range rule.APIGroups {
			if group != "" {
				continue
			}
			for _, resource := range rule.Resources {
				granted[resource] = true
			}
			assert.ElementsMatch(t, []string{"get", "list", "watch"}, rule.Verbs)
		}
	}
	assert.True(t, granted["events"], "events must be watchable")
	assert.True(t, granted["pods"], "pods must be watchable for OOM events")
	assert.True(t, granted["namespaces"], "kube-system must be readable for orch_cluster_id")
}

func TestKubernetesEventsCompatRequiresLeaderElection(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/kubernetes-events-compat.yaml"},
		SetValues: map[string]string{
			"otel.k8sResourceCollector.leaderElection.enabled": "false",
			"otel.k8sResourceCollector.replicaCount":           "1",
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires otel.k8sResourceCollector.leaderElection.enabled")
}

func TestKubernetesEventsCompatRequiresStsAgentUrl(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/kubernetes-events-compat.yaml"},
		SetValues:   map[string]string{"stackstate.url": "https://my-suse-observability-instance.com/receiver"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "to end in /stsAgent")
}

func TestKubernetesClusterAgentFlagMovesTopologyEventsAndMetrics(t *testing.T) {
	resources := renderCompat(t, &helm.Options{ValuesFiles: []string{"values/kubernetes-events-compat.yaml"}})
	config := collectorConfigMap(t, resources)
	exporters := section(t, config, "exporters")
	assert.Contains(t, exporters, "stsk8sevents")
	assert.Contains(t, exporters, "stsk8stopology")
	pipelines := section(t, config, "service", "pipelines")
	assert.Contains(t, pipelines, "logs/kubernetes-events")
	assert.Contains(t, pipelines, "logs/kubernetes-topology")
	assert.Contains(t, clusterAgentEventsConf(t, resources), "collect_events: false")
	env, ok := containerEnv(t, resources, clusterAgentName, "STS_COLLECT_KUBERNETES_TOPOLOGY")
	require.True(t, ok)
	assert.Equal(t, "false", env.Value)
	env, ok = containerEnv(t, resources, clusterAgentName, "STS_COLLECT_KUBERNETES_METRICS")
	require.True(t, ok)
	assert.Equal(t, "false", env.Value, "its only output, the OOM gauge, is dropped")
	assert.Contains(t, pipelines, "metrics/kubernetes-state")
}

func compatClusterTypes(t *testing.T, opts *helm.Options, args ...string) (topology, events any) {
	t.Helper()
	output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "suse-observability-agent", opts, args...)
	config := collectorConfigMap(t, helmtestutil.NewKubernetesResources(t, output))
	return section(t, config, "exporters", "stsk8stopology")["cluster_type"], section(t, config, "exporters", "stsk8sevents")["cluster_type"]
}

func TestKubernetesCompatDetectsOpenShift(t *testing.T) {
	opts := &helm.Options{ValuesFiles: []string{"values/kubernetes-topology-compat.yaml"}}
	topology, events := compatClusterTypes(t, opts, "--api-versions", "quota.openshift.io/v1")
	assert.Equal(t, "openshift", topology)
	assert.Equal(t, "openshift", events)
}

func TestKubernetesCompatClusterTypeOverridesDetection(t *testing.T) {
	opts := &helm.Options{
		ValuesFiles: []string{"values/kubernetes-topology-compat.yaml"},
		SetValues:   map[string]string{"otel.k8sResourceCollector.kubernetesTopology.clusterType": "kubernetes"},
	}
	topology, events := compatClusterTypes(t, opts, "--api-versions", "quota.openshift.io/v1")
	assert.Equal(t, "kubernetes", topology)
	assert.Equal(t, "kubernetes", events)
}

func TestKubernetesCompatRejectsUnknownClusterType(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/kubernetes-topology-compat.yaml"},
		SetValues:   map[string]string{"otel.k8sResourceCollector.kubernetesTopology.clusterType": "eks"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "clusterType must be kubernetes, openshift or empty")
}
