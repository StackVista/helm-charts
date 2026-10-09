package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"gopkg.in/yaml.v3"
)

func TestOtelLogsWithTracesEnabledRenders(t *testing.T) {
	// Test that OTel logs can be enabled when traces is enabled (the default)
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/otel_logs_with_traces_valid.yaml")
	helmtestutil.NewKubernetesResources(t, output)
}

func TestCollectorConfigOverrides(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"opentelemetry-collector.config.processors.memory_limiter.limit_percentage": "70",
			"opentelemetry-collector.config.processors.batch.send_batch_size":           "5000",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	var config struct {
		Processors map[string]map[string]interface{} `yaml:"processors"`
		Connectors map[string]interface{}            `yaml:"connectors"`
	}
	relay := resources.ConfigMaps["suse-observability-otel-collector-statefulset"].Data["relay"]
	require.NotEmpty(t, relay)
	require.NoError(t, yaml.Unmarshal([]byte(relay), &config))
	assert.Equal(t, 70, config.Processors["memory_limiter"]["limit_percentage"])
	assert.Equal(t, 5000, config.Processors["batch"]["send_batch_size"])
	assert.Contains(t, config.Connectors, "topology")
}
