package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"gopkg.in/yaml.v3"
)

const (
	releaseName = "otel"
	// fullName matches the subchart's default fullnameOverride.
	fullName = "suse-observability-otel-collector"
)

func TestOpenTelemetryCollectorBasicTemplate(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have Deployments (default mode)
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")
}

func TestOpenTelemetryCollectorDeploymentMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/deployment-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// In Deployment mode, should have a Deployment
	expectedDeployments := []string{
		fullName,
	}

	assert.Len(t, resources.Deployments, len(expectedDeployments), "Should have exactly %d Deployments in Deployment mode", len(expectedDeployments))

	// Check each expected Deployment exists
	for _, expectedName := range expectedDeployments {
		found := false
		for _, deployment := range resources.Deployments {
			if deployment.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "Deployment %s should exist in Deployment mode", expectedName)
	}

	// Should not have DaemonSets or StatefulSets
	assert.Len(t, resources.DaemonSets, 0, "Should not have DaemonSets in Deployment mode")
	assert.Len(t, resources.Statefulsets, 0, "Should not have StatefulSets in Deployment mode")
}

func TestOpenTelemetryCollectorDaemonSetMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/daemonset-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// In DaemonSet mode, should have a DaemonSet
	expectedDaemonSets := []string{
		fullName + "-agent",
	}

	assert.Len(t, resources.DaemonSets, len(expectedDaemonSets), "Should have exactly %d DaemonSets in DaemonSet mode", len(expectedDaemonSets))

	// Check each expected DaemonSet exists
	for _, expectedName := range expectedDaemonSets {
		found := false
		for _, daemonset := range resources.DaemonSets {
			if daemonset.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "DaemonSet %s should exist in DaemonSet mode", expectedName)
	}

	// Should not have Deployments or StatefulSets
	assert.Len(t, resources.Deployments, 0, "Should not have Deployments in DaemonSet mode")
	assert.Len(t, resources.Statefulsets, 0, "Should not have StatefulSets in DaemonSet mode")
}

func TestOpenTelemetryCollectorStatefulSetMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/statefulset-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// In StatefulSet mode, should have a StatefulSet
	expectedStatefulSets := []string{
		fullName,
	}

	assert.Len(t, resources.Statefulsets, len(expectedStatefulSets), "Should have exactly %d StatefulSets in StatefulSet mode", len(expectedStatefulSets))

	// Check each expected StatefulSet exists
	for _, expectedName := range expectedStatefulSets {
		found := false
		for _, statefulset := range resources.Statefulsets {
			if statefulset.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "StatefulSet %s should exist in StatefulSet mode", expectedName)
	}

	// Should not have Deployments or DaemonSets
	assert.Len(t, resources.Deployments, 0, "Should not have Deployments in StatefulSet mode")
	assert.Len(t, resources.DaemonSets, 0, "Should not have DaemonSets in StatefulSet mode")
}

func TestOpenTelemetryCollectorConfigSelection(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	defaultCollectorConfig := resources.ConfigMaps[fullName].Data["relay"]

	assert.Contains(t, defaultCollectorConfig, "sts_settings_provider")
	assert.Contains(t, defaultCollectorConfig, "topology")
	assert.NotContains(t, defaultCollectorConfig, "ststopology")
	assert.Contains(t, defaultCollectorConfig, "sts_kafka_exporter")
	assert.Contains(t, defaultCollectorConfig, "trace_statements")
}

func TestCollectorConfigDefaultsAndOverrides(t *testing.T) {
	for _, mode := range []string{"deployment", "daemonset", "statefulset"} {
		for _, override := range []bool{false, true} {
			name := mode + "/defaults"
			if override {
				name = mode + "/overrides"
			}
			t.Run(name, func(t *testing.T) {
				values := map[string]string{"mode": mode}
				limit, batchSize, timeout := 80, 10000, "30s"
				if override {
					values["config.processors.memory_limiter.limit_percentage"] = "70"
					values["config.processors.batch.send_batch_size"] = "5000"
					values["config.exporters.prometheusremotewrite/victoria-metrics.timeout"] = "45s"
					limit, batchSize, timeout = 70, 5000, "45s"
				}
				output := helmtestutil.RenderHelmTemplateOptsNoError(t, releaseName, &helm.Options{
					ValuesFiles: []string{"values/default.yaml"},
					SetValues:   values,
				})
				resources := helmtestutil.NewKubernetesResources(t, output)
				configName := fullName
				if mode == "daemonset" {
					configName += "-agent"
				} else if mode == "statefulset" {
					configName += "-statefulset"
				}
				require.Contains(t, resources.ConfigMaps, configName)
				var config struct {
					Processors map[string]map[string]interface{} `yaml:"processors"`
					Exporters  map[string]map[string]interface{} `yaml:"exporters"`
					Service    struct {
						Pipelines map[string]struct {
							Processors []string `yaml:"processors"`
							Exporters  []string `yaml:"exporters"`
						} `yaml:"pipelines"`
					} `yaml:"service"`
				}
				require.NoError(t, yaml.Unmarshal([]byte(resources.ConfigMaps[configName].Data["relay"]), &config))
				assert.Equal(t, map[string]interface{}{
					"check_interval": "1s", "limit_percentage": limit, "spike_limit_percentage": 25,
				}, config.Processors["memory_limiter"])
				assert.Equal(t, batchSize, config.Processors["batch"]["send_batch_size"])
				assert.Equal(t, "2s", config.Processors["batch"]["timeout"])
				assert.Equal(t, timeout, config.Exporters["prometheusremotewrite/victoria-metrics"]["timeout"])
				for _, pipeline := range []string{"traces", "metrics", "metrics/internal", "logs"} {
					processors := config.Service.Pipelines[pipeline].Processors
					require.NotEmpty(t, processors, pipeline)
					assert.Equal(t, "memory_limiter", processors[0], pipeline)
				}
				for _, pipeline := range []string{"traces", "metrics", "logs", "traces/clickhouse", "metrics/internal", "metrics/victoria-metrics", "metrics/topology"} {
					processors := config.Service.Pipelines[pipeline].Processors
					require.NotEmpty(t, processors, pipeline)
					assert.Equal(t, "batch", processors[len(processors)-1], pipeline)
				}
				for _, pipeline := range []string{"traces/topology", "metrics/topology", "logs/topology_input"} {
					assert.Equal(t, []string{"topology"}, config.Service.Pipelines[pipeline].Exporters, pipeline)
					assert.Contains(t, config.Service.Pipelines[pipeline].Processors, "resource/removeStsApiKey", pipeline)
					assert.Contains(t, config.Service.Pipelines[pipeline].Processors, "attributes/removeStsApiKey", pipeline)
				}
				assert.Equal(t, []string{"sts_kafka_exporter"}, config.Service.Pipelines["logs/topology"].Exporters)
			})
		}
	}
}

func TestSendingQueueNoEnabledField(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configYaml := resources.ConfigMaps[fullName].Data["relay"]
	require.NotEmpty(t, configYaml, "ConfigMap relay data should not be empty")

	var config struct {
		Exporters map[string]map[string]interface{} `yaml:"exporters"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(configYaml), &config))

	// OTel Collector v0.149.0+ removed 'enabled' from QueueBatchConfig — the queue
	// is enabled by its presence, not by an internal flag. Verify no exporter uses it.
	for name, exporter := range config.Exporters {
		if sq, ok := exporter["sending_queue"].(map[string]interface{}); ok {
			_, hasEnabled := sq["enabled"]
			assert.False(t, hasEnabled, "exporters.%s.sending_queue must not contain 'enabled'", name)
		}
	}
}
