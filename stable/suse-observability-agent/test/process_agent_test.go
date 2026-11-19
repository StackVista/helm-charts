package test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

const (
	releaseName            = "suse-observability-agent"
	nodeAgentName          = "node-agent"
	processAgentName       = "process-agent"
	processAgentEnabledKey = "nodeAgent.containers.processAgent.enabled"

	POD_CORRELATION_ENABLED                = "STS_POD_CORRELATION_ENABLED"
	POD_CORRELATION_PROTOCOL_METRICS       = "STS_POD_CORRELATION_PROTOCOL_METRICS"
	POD_CORRELATION_PARTIAL_CORRELATION    = "STS_POD_CORRELATION_PARTIAL_CORRELATION"
	POD_CORRELATION_ATTRIBUTES_KEYS        = "STS_POD_CORRELATION_ATTRIBUTES_KEYS"
	POD_CORRELATION_EXPORTER_TYPE          = "STS_POD_CORRELATION_EXPORTER_TYPE"
	POD_CORRELATION_EXPORTER_OTLP_ENDPOINT = "STS_POD_CORRELATION_EXPORTER_OTLP_ENDPOINT"
	POD_CORRELATION_EXPORTER_INTERVAL      = "STS_POD_CORRELATION_EXPORTER_INTERVAL"

	DISABLED_PROTOCOLS = "STS_DISABLED_PROTOCOLS"
)

var (
	fullNodeAgentName = fmt.Sprintf("%s-%s", releaseName, nodeAgentName)
)

func TestProcessAgentEnvVars(t *testing.T) {
	tests := []struct {
		name      string
		setValues map[string]string
		expectEnv map[string]string
	}{
		{
			// env vars should have default values
			name: "pod correlation disabled",
			setValues: map[string]string{
				"processAgent.podCorrelation.enabled": "false",
			},
			expectEnv: map[string]string{
				POD_CORRELATION_ENABLED:                "false",
				POD_CORRELATION_PROTOCOL_METRICS:       "false",
				POD_CORRELATION_PARTIAL_CORRELATION:    "false",
				POD_CORRELATION_EXPORTER_TYPE:          "",
				POD_CORRELATION_EXPORTER_OTLP_ENDPOINT: "",
				POD_CORRELATION_EXPORTER_INTERVAL:      "30",
				POD_CORRELATION_ATTRIBUTES_KEYS:        "",
			},
		},
		{
			// env vars should have default values
			name: "pod correlation enabled",
			setValues: map[string]string{
				"processAgent.podCorrelation.enabled":       "true",
				"processAgent.podCorrelation.attributes[0]": "example0",
				"processAgent.podCorrelation.attributes[1]": "example1",
			},
			expectEnv: map[string]string{
				POD_CORRELATION_ENABLED:             "true",
				POD_CORRELATION_PROTOCOL_METRICS:    "false",
				POD_CORRELATION_PARTIAL_CORRELATION: "false",
				POD_CORRELATION_EXPORTER_INTERVAL:   "30",
				POD_CORRELATION_ATTRIBUTES_KEYS:     "example0,example1",
			},
		},
		{
			name: "pod correlation enabled with otlp exporter",
			setValues: map[string]string{
				"processAgent.podCorrelation.enabled":           "true",
				"processAgent.podCorrelation.exporter.type":     "otlp",
				"processAgent.podCorrelation.exporter.endpoint": "otel-collector:4317",
				"processAgent.podCorrelation.attributes[0]":     "example0",
			},
			expectEnv: map[string]string{
				POD_CORRELATION_ENABLED:                "true",
				POD_CORRELATION_PROTOCOL_METRICS:       "false",
				POD_CORRELATION_PARTIAL_CORRELATION:    "false",
				POD_CORRELATION_EXPORTER_TYPE:          "otlp",
				POD_CORRELATION_EXPORTER_OTLP_ENDPOINT: "otel-collector:4317",
				POD_CORRELATION_EXPORTER_INTERVAL:      "30",
				POD_CORRELATION_ATTRIBUTES_KEYS:        "example0",
			},
		},
		{
			name: "all protocols enabled",
			setValues: map[string]string{
				"processAgent.disabledProtocols": "",
			},
			expectEnv: map[string]string{
				DISABLED_PROTOCOLS: "",
			},
		},
		{
			name: "disable http and amqp protocols",
			setValues: map[string]string{
				"processAgent.disabledProtocols[0]": "amqp",
				"processAgent.disabledProtocols[1]": "http",
			},
			expectEnv: map[string]string{
				DISABLED_PROTOCOLS: "amqp,http",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Add the process agent key to each test
			tt.setValues[processAgentEnabledKey] = "true"

			helmOpts := &helm.Options{
				ValuesFiles: []string{"values/minimal.yaml"},
				// We need to use `SetValues` instead of `SetStrValues` otherwise boolean will be interpreted as strings causing errors.
				// it is the same difference between `--set` and `--set-string`
				SetValues: tt.setValues,
			}

			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", helmOpts)
			resources := helmtestutil.NewKubernetesResources(t, output)

			///////////////////
			// Check process-agent presence and env vars
			///////////////////

			assert.Greater(t, len(resources.DaemonSets), 0)
			ds, ok := resources.DaemonSets[fullNodeAgentName]
			assert.True(t, ok)

			// We should have both containers
			assert.Len(t, ds.Spec.Template.Spec.Containers, 2)
			processAgentContainer := ds.Spec.Template.Spec.Containers[1]
			assert.Equal(t, processAgentName, processAgentContainer.Name)

			// Build env var
			envMap := map[string]string{}
			for _, env := range processAgentContainer.Env {
				envMap[env.Name] = env.Value
			}

			for k, v := range tt.expectEnv {
				assert.Contains(t, envMap, k)
				assert.Equal(t, v, envMap[k], "Wrong value for env var %s", k)
			}
			assert.NotContains(t, envMap, "STS_POD_CORRELATION_REMOTE_CACHE_ADDR")
		})
	}

}

func TestProcessAgentPodCorrelation(t *testing.T) {
	tests := []struct {
		name      string
		setValues map[string]string
	}{
		{
			// we don't expect the process agent container
			name: "process agent disabled",
			setValues: map[string]string{
				processAgentEnabledKey:                "false",
				"processAgent.podCorrelation.enabled": "true",
			},
		},
		{
			name: "pod correlation disabled",
			setValues: map[string]string{
				processAgentEnabledKey:                "true",
				"processAgent.podCorrelation.enabled": "false",
			},
		},
		{
			name: "pod correlation enabled",
			setValues: map[string]string{
				processAgentEnabledKey:                "true",
				"processAgent.podCorrelation.enabled": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()

			helmOpts := &helm.Options{
				ValuesFiles: []string{"values/minimal.yaml"},
				// We need to use `SetValues` instead of `SetStrValues` otherwise boolean will be interpreted as strings causing errors.
				// it is the same difference between `--set` and `--set-string`
				SetValues: tt.setValues,
			}

			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", helmOpts)
			resources := helmtestutil.NewKubernetesResources(t, output)

			assert.NotContains(t, output, "remote-kube-cache")

			///////////////////
			// Check node-agent Clusterrole permissions
			///////////////////

			assert.Greater(t, len(resources.ClusterRoles), 0)
			fullNodeAgentName := fmt.Sprintf("%s-%s", releaseName, nodeAgentName)
			nodeAgentClusterRole, ok := resources.ClusterRoles[fullNodeAgentName]
			assert.True(t, ok)
			assert.Len(t, nodeAgentClusterRole.Rules, 1)
			containsPods := slices.Contains(nodeAgentClusterRole.Rules[0].Resources, "pods")
			containsWatch := slices.Contains(nodeAgentClusterRole.Rules[0].Verbs, "watch")

			if tt.setValues[processAgentEnabledKey] == "true" &&
				tt.setValues["processAgent.podCorrelation.enabled"] == "true" {
				assert.True(t, containsWatch)
				assert.True(t, containsPods)
			} else {
				assert.False(t, containsWatch)
				assert.False(t, containsPods)
			}

			///////////////////
			// Check process-agent presence and env vars
			///////////////////

			assert.Greater(t, len(resources.DaemonSets), 0)
			ds, ok := resources.DaemonSets[fullNodeAgentName]
			assert.True(t, ok)

			if tt.setValues[processAgentEnabledKey] != "true" {
				// There should be only the node agent
				assert.Len(t, ds.Spec.Template.Spec.Containers, 1)
			} else {
				// We should have both containers
				assert.Len(t, ds.Spec.Template.Spec.Containers, 2)
				processAgentContainer := ds.Spec.Template.Spec.Containers[1]
				assert.Equal(t, processAgentName, processAgentContainer.Name)
			}

		})
	}
}
