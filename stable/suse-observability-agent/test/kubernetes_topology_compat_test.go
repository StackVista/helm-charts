package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

const (
	resourceCollectorName       = "suse-observability-agent-k8s-resource-collector"
	resourceCollectorConfigName = "suse-observability-agent-k8s-resource-collector-config"
	clusterAgentName            = "suse-observability-agent-cluster-agent"
)

type collectorObjectWatch struct {
	Name          string `json:"name"`
	Group         string `json:"group"`
	LabelSelector string `json:"label_selector"`
}

type compatCollectorConfig struct {
	Receivers struct {
		K8sResource struct {
			EmitSnapshotBoundaries      bool                   `json:"emit_snapshot_boundaries"`
			MaxObjectTotalDataSizeBytes int                    `json:"max_object_total_data_size_bytes"`
			ConfigMapMaxDataSize        int                    `json:"configmap_max_datasize"`
			Objects                     []collectorObjectWatch `json:"objects"`
		} `json:"k8sresource"`
	} `json:"receivers"`
	Exporters map[string]map[string]any `json:"exporters"`
	Service   struct {
		Pipelines map[string]struct {
			Receivers  []string `json:"receivers"`
			Processors []string `json:"processors"`
			Exporters  []string `json:"exporters"`
		} `json:"pipelines"`
	} `json:"service"`
}

func renderCompat(t *testing.T, opts *helm.Options) helmtestutil.KubernetesResources {
	t.Helper()
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", opts)
	return helmtestutil.NewKubernetesResources(t, output)
}

func collectorConfig(t *testing.T, resources helmtestutil.KubernetesResources) compatCollectorConfig {
	t.Helper()
	configMap, ok := resources.ConfigMaps[resourceCollectorConfigName]
	require.True(t, ok, "k8s-resource-collector config map was not found")
	var config compatCollectorConfig
	require.NoError(t, yaml.Unmarshal([]byte(configMap.Data["config.yaml"]), &config))
	return config
}

func containerEnv(t *testing.T, resources helmtestutil.KubernetesResources, deployment, name string) (corev1.EnvVar, bool) {
	t.Helper()
	d, ok := resources.Deployments[deployment]
	require.True(t, ok, "%s deployment was not found", deployment)
	for _, env := range d.Spec.Template.Spec.Containers[0].Env {
		if env.Name == name {
			return env, true
		}
	}
	return corev1.EnvVar{}, false
}

func watchesByName(config compatCollectorConfig) map[string]collectorObjectWatch {
	out := map[string]collectorObjectWatch{}
	for _, watch := range config.Receivers.K8sResource.Objects {
		out[watch.Name] = watch
	}
	return out
}

func TestKubernetesTopologyCompatDisabledByDefault(t *testing.T) {
	resources := renderCompat(t, &helm.Options{ValuesFiles: []string{"values/k8s-resource-collector-enabled.yaml"}})
	config := collectorConfig(t, resources)

	assert.False(t, config.Receivers.K8sResource.EmitSnapshotBoundaries)
	assert.NotContains(t, config.Exporters, "stsk8stopology")
	assert.NotContains(t, config.Service.Pipelines, "logs/kubernetes-topology")
	_, ok := containerEnv(t, resources, resourceCollectorName, "RECEIVER_INTAKE_URL")
	assert.False(t, ok)
	env, ok := containerEnv(t, resources, clusterAgentName, "STS_COLLECT_KUBERNETES_TOPOLOGY")
	require.True(t, ok)
	assert.Equal(t, "true", env.Value)
}

func TestKubernetesTopologyCompatEnabled(t *testing.T) {
	resources := renderCompat(t, &helm.Options{ValuesFiles: []string{"values/kubernetes-topology-compat.yaml"}})
	config := collectorConfig(t, resources)

	receiver := config.Receivers.K8sResource
	assert.True(t, receiver.EmitSnapshotBoundaries)
	assert.Zero(t, receiver.MaxObjectTotalDataSizeBytes, "the object payload budget must be disabled")
	watches := watchesByName(config)
	for name, group := range map[string]string{
		"nodes": "", "namespaces": "", "pods": "", "services": "", "persistentvolumes": "",
		"persistentvolumeclaims": "", "volumeattachments": "storage.k8s.io", "deployments": "apps",
		"replicasets": "apps", "daemonsets": "apps", "statefulsets": "apps", "jobs": "batch",
		"cronjobs": "batch", "ingresses": "networking.k8s.io", "configmaps": "", "secrets": "",
	} {
		watch, ok := watches[name]
		if assert.True(t, ok, "%s must be watched", name) {
			assert.Equal(t, group, watch.Group, name)
		}
	}

	exporter := config.Exporters["stsk8stopology"]
	require.NotNil(t, exporter)
	assert.Equal(t, "${env:RECEIVER_INTAKE_URL}", exporter["endpoint"])
	assert.Equal(t, "${env:STS_API_KEY}", exporter["api_key"])
	assert.Equal(t, "${env:K8S_CLUSTER_NAME}", exporter["cluster_name"])
	assert.Equal(t, "kubernetes", exporter["cluster_type"])
	assert.Equal(t, "90s", exporter["interval"])
	switches, ok := exporter["resources"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, switches["configmaps"])
	assert.Equal(t, true, switches["secrets"])

	pipeline, ok := config.Service.Pipelines["logs/kubernetes-topology"]
	require.True(t, ok)
	assert.Equal(t, []string{"k8sresource"}, pipeline.Receivers)
	assert.Empty(t, pipeline.Processors, "records must reach the exporter in emission order")
	assert.Equal(t, []string{"stsk8stopology"}, pipeline.Exporters)
	assert.Equal(t, []string{"batch"}, config.Service.Pipelines["logs/k8s-resource"].Processors)

	env, ok := containerEnv(t, resources, resourceCollectorName, "RECEIVER_INTAKE_URL")
	require.True(t, ok)
	assert.Equal(t, "https://my-suse-observability-instance.com/receiver/stsAgent/intake", env.Value)
	env, ok = containerEnv(t, resources, clusterAgentName, "STS_COLLECT_KUBERNETES_TOPOLOGY")
	require.True(t, ok)
	assert.Equal(t, "false", env.Value)
}

func TestKubernetesTopologyCompatUrlFromSecret(t *testing.T) {
	resources := renderCompat(t, &helm.Options{
		ValuesFiles: []string{"values/kubernetes-topology-compat.yaml"},
		SetValues:   map[string]string{"global.url.fromSecret": "sts-url"},
	})
	env, ok := containerEnv(t, resources, resourceCollectorName, "RECEIVER_INTAKE_URL")
	require.True(t, ok)
	assert.Equal(t, "$(STS_URL)/intake", env.Value)
	_, ok = containerEnv(t, resources, resourceCollectorName, "STS_URL")
	assert.True(t, ok, "STS_URL must be defined for the dependent variable")
}

func TestKubernetesTopologyCompatFollowsClusterAgentResourceSwitches(t *testing.T) {
	resources := renderCompat(t, &helm.Options{
		ValuesFiles: []string{"values/kubernetes-topology-compat.yaml"},
		SetValues: map[string]string{
			"clusterAgent.collection.kubernetesResources.persistentvolumes":      "false",
			"clusterAgent.collection.kubernetesResources.persistentvolumeclaims": "false",
			"clusterAgent.collection.kubernetesResources.cronjobs":               "false",
			"clusterAgent.collection.kubernetesResources.secrets":                "false",
			"clusterAgent.config.topology.collectionInterval":                    "120",
			"clusterAgent.config.configMap.maxDataSize":                          "2048",
		},
	})
	config := collectorConfig(t, resources)
	watches := watchesByName(config)
	for _, name := range []string{"persistentvolumes", "persistentvolumeclaims", "volumeattachments", "cronjobs", "secrets"} {
		assert.NotContains(t, watches, name)
	}
	assert.Contains(t, watches, "jobs")

	exporter := config.Exporters["stsk8stopology"]
	assert.Equal(t, "120s", exporter["interval"])
	assert.Equal(t, 2048, config.Receivers.K8sResource.ConfigMapMaxDataSize)
	assert.NotContains(t, exporter, "configmap_max_datasize")
	switches, ok := exporter["resources"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, false, switches["persistentvolumes"])
	assert.Equal(t, false, switches["cronjobs"])
	assert.Equal(t, true, switches["jobs"])
	assert.Equal(t, false, switches["secrets"])
}

func TestKubernetesTopologyCompatRestrictedRBAC(t *testing.T) {
	resources := renderCompat(t, &helm.Options{
		ValuesFiles: []string{"values/kubernetes-topology-compat-restricted-rbac.yaml"},
	})
	config := collectorConfig(t, resources)
	pods := watchesByName(config)["pods"]
	assert.Empty(t, pods.LabelSelector, "topology needs every pod, not an integration's subset")

	clusterRole, ok := resources.ClusterRoles[resourceCollectorName]
	require.True(t, ok)
	granted := map[string]map[string]bool{}
	for _, rule := range clusterRole.Rules {
		for _, group := range rule.APIGroups {
			if granted[group] == nil {
				granted[group] = map[string]bool{}
			}
			for _, resource := range rule.Resources {
				granted[group][resource] = true
			}
		}
	}
	for _, watch := range config.Receivers.K8sResource.Objects {
		assert.True(t, granted[watch.Group][watch.Name] || granted[watch.Group]["*"],
			"%s/%s is watched but not granted", watch.Group, watch.Name)
	}
}

func TestKubernetesTopologyCompatRequiresResourceCollector(t *testing.T) {
	resources := renderCompat(t, &helm.Options{
		ValuesFiles: []string{"values/kubernetes-topology-compat.yaml"},
		SetValues:   map[string]string{"otel.enabled": "false"},
	})
	_, ok := resources.ConfigMaps[resourceCollectorConfigName]
	assert.False(t, ok)
	env, ok := containerEnv(t, resources, clusterAgentName, "STS_COLLECT_KUBERNETES_TOPOLOGY")
	require.True(t, ok)
	assert.Equal(t, "true", env.Value, "the cluster agent keeps topology when the collector is not deployed")
}

func TestKubernetesTopologyCompatRequiresStsAgentUrl(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability-agent",
		"values/kubernetes-topology-compat.yaml", "values/kubernetes-topology-compat-receiver-url.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires stackstate.url to end in /stsAgent")
}

func TestKubernetesTopologyCompatSnapshotMaxAgeFollowsIntervals(t *testing.T) {
	for _, tc := range []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{"defaults", map[string]string{}, "900s"},
		{"long collection interval", map[string]string{"clusterAgent.config.topology.collectionInterval": "1800"}, "1800s"},
		{"long snapshot interval", map[string]string{"otel.k8sResourceCollector.crDiscovery.snapshotInterval": "30m"}, "5400s"},
		{"compound snapshot interval", map[string]string{"otel.k8sResourceCollector.crDiscovery.snapshotInterval": "1h30m"}, "16200s"},
		{"zero snapshot interval uses the receiver default", map[string]string{"otel.k8sResourceCollector.crDiscovery.snapshotInterval": "0s"}, "900s"},
		{"bare zero snapshot interval", map[string]string{"otel.k8sResourceCollector.crDiscovery.snapshotInterval": "0"}, "900s"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := renderCompat(t, &helm.Options{
				ValuesFiles: []string{"values/kubernetes-topology-compat.yaml"},
				SetValues:   tc.values,
			})
			exporter := collectorConfig(t, resources).Exporters["stsk8stopology"]
			assert.Equal(t, tc.expected, exporter["snapshot_max_age"])
		})
	}
}

func TestKubernetesTopologyCompatRejectsUnsupportedSnapshotInterval(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability-agent",
		"values/kubernetes-topology-compat.yaml", "values/kubernetes-topology-compat-invalid-snapshot-interval.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unsupported duration "5min"`)
}
