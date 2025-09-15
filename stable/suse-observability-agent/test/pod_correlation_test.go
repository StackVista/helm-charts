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
	releaseName                = "suse-observability-agent"
	nodeAgentName              = "node-agent"
	processAgentName           = "process-agent"
	remoteCacheName            = "remote-kube-cache"
	remoteCacheServiceGRPCPort = "50055"

	POD_CORRELATION_ENABLED                = "STS_POD_CORRELATION_ENABLED"
	POD_CORRELATION_REMOTE_CACHE_ADDR      = "STS_POD_CORRELATION_REMOTE_CACHE_ADDR"
	POD_CORRELATION_PROTOCOL_METRICS       = "STS_POD_CORRELATION_PROTOCOL_METRICS"
	POD_CORRELATION_PARTIAL_CORRELATION    = "STS_POD_CORRELATION_PARTIAL_CORRELATION"
	POD_CORRELATION_ATTRIBUTES_KEYS        = "STS_POD_CORRELATION_ATTRIBUTES_KEYS"
	POD_CORRELATION_EXPORTER_TYPE          = "STS_POD_CORRELATION_EXPORTER_TYPE"
	POD_CORRELATION_EXPORTER_OTLP_ENDPOINT = "STS_POD_CORRELATION_EXPORTER_OTLP_ENDPOINT"
	POD_CORRELATION_EXPORTER_INTERVAL      = "STS_POD_CORRELATION_EXPORTER_INTERVAL"
)

func TestProcessAgentPodCorrelation(t *testing.T) {
	tests := []struct {
		name      string
		setValues map[string]string
		expectEnv map[string]string
	}{
		{
			// we don't expect the process agent container
			name: "process agent disabled",
			setValues: map[string]string{
				"nodeAgent.containers.processAgent.enabled": "false",
				"processAgent.podCorrelation.enabled":       "true",
			},
		},
		{
			// env vars should have default values
			name: "pod correlation disabled",
			setValues: map[string]string{
				"nodeAgent.containers.processAgent.enabled": "true",
				"processAgent.podCorrelation.enabled":       "false",
			},
			expectEnv: map[string]string{
				POD_CORRELATION_ENABLED:             "false",
				POD_CORRELATION_PROTOCOL_METRICS:    "false",
				POD_CORRELATION_PARTIAL_CORRELATION: "false",
				// Note: if we provide an empty sting these 2 variables won't be defined.
				// POD_CORRELATION_EXPORTER_TYPE:          "",
				// POD_CORRELATION_EXPORTER_OTLP_ENDPOINT: "",
				POD_CORRELATION_EXPORTER_INTERVAL: "30",
				// if we don't define any attributes we should have an empty string, the variable won't be defined
				// POD_CORRELATION_ATTRIBUTES_KEYS: "",
			},
		},
		{
			// env vars should have default values
			name: "pod correlation enabled no remote cache",
			setValues: map[string]string{
				"nodeAgent.containers.processAgent.enabled": "true",
				"processAgent.podCorrelation.enabled":       "true",
				"processAgent.podCorrelation.remoteCache":   "false",
				"processAgent.podCorrelation.attributes[0]": "example0",
				"processAgent.podCorrelation.attributes[1]": "example1",
			},
			expectEnv: map[string]string{
				POD_CORRELATION_ENABLED:             "true",
				POD_CORRELATION_PROTOCOL_METRICS:    "false",
				POD_CORRELATION_PARTIAL_CORRELATION: "false",
				// Note: if we provide an empty sting these 2 variables won't be defined.
				// POD_CORRELATION_EXPORTER_TYPE:          "",
				// POD_CORRELATION_EXPORTER_OTLP_ENDPOINT: "",
				POD_CORRELATION_EXPORTER_INTERVAL: "30",
				POD_CORRELATION_ATTRIBUTES_KEYS:   "example0,example1",
			},
		},
		{
			name: "pod correlation enabled remote cache",
			setValues: map[string]string{
				"nodeAgent.containers.processAgent.enabled":     "true",
				"processAgent.podCorrelation.remoteCache":       "true",
				"processAgent.podCorrelation.enabled":           "true",
				"processAgent.podCorrelation.exporter.type":     "otlp",
				"processAgent.podCorrelation.exporter.endpoint": "otel-collector:4317",
				"processAgent.podCorrelation.attributes[0]":     "example0",
			},
			expectEnv: map[string]string{
				POD_CORRELATION_ENABLED:                "true",
				POD_CORRELATION_PROTOCOL_METRICS:       "false",
				POD_CORRELATION_PARTIAL_CORRELATION:    "false",
				POD_CORRELATION_REMOTE_CACHE_ADDR:      fmt.Sprintf("%s-%s:%s", releaseName, remoteCacheName, remoteCacheServiceGRPCPort),
				POD_CORRELATION_EXPORTER_TYPE:          "otlp",
				POD_CORRELATION_EXPORTER_OTLP_ENDPOINT: "otel-collector:4317",
				POD_CORRELATION_EXPORTER_INTERVAL:      "30",
				POD_CORRELATION_ATTRIBUTES_KEYS:        "example0",
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

			///////////////////
			// Check k8s remote cache presence
			///////////////////

			fullRemoteCacheName := fmt.Sprintf("%s-%s", releaseName, remoteCacheName)
			// we should have the remote kubernetes cache
			if tt.setValues["nodeAgent.containers.processAgent.enabled"] == "true" &&
				tt.setValues["processAgent.podCorrelation.enabled"] == "true" &&
				tt.setValues["processAgent.podCorrelation.remoteCache"] == "true" {
				// remote cache deployment
				assert.Greater(t, len(resources.Deployments), 0)
				// remote cache service
				assert.Greater(t, len(resources.Services), 0)

				_, ok := resources.Deployments[fullRemoteCacheName]
				assert.True(t, ok)
				_, ok = resources.Services[fullRemoteCacheName]
				assert.True(t, ok)
			} else {
				// in all other cases we shouldn't have the cache
				_, ok := resources.Deployments[fullRemoteCacheName]
				assert.False(t, ok)
				_, ok = resources.Services[fullRemoteCacheName]
				assert.False(t, ok)
			}

			///////////////////
			// Check node-agent Clusterrole permissions without the remote cache
			///////////////////

			assert.Greater(t, len(resources.ClusterRoles), 0)
			fullNodeAgentName := fmt.Sprintf("%s-%s", releaseName, nodeAgentName)
			nodeAgentClusterRole, ok := resources.ClusterRoles[fullNodeAgentName]
			assert.True(t, ok)
			assert.Len(t, nodeAgentClusterRole.Rules, 1)
			containsPods := slices.Contains(nodeAgentClusterRole.Rules[0].Resources, "pods")
			containsWatch := slices.Contains(nodeAgentClusterRole.Rules[0].Verbs, "watch")

			if tt.setValues["nodeAgent.containers.processAgent.enabled"] == "true" &&
				tt.setValues["processAgent.podCorrelation.enabled"] == "true" &&
				tt.setValues["processAgent.podCorrelation.remoteCache"] == "false" {
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

			if len(tt.expectEnv) == 0 {
				// There should be only the node agent
				assert.Len(t, ds.Spec.Template.Spec.Containers, 1)
				return
			}

			// We should have both containers
			assert.Len(t, ds.Spec.Template.Spec.Containers, 2)
			processAgentContainer := ds.Spec.Template.Spec.Containers[1]
			assert.Equal(t, processAgentName, processAgentContainer.Name)

			// Build env var
			envMap := map[string]string{}
			for _, env := range processAgentContainer.Env {
				if env.Value != "" {
					envMap[env.Name] = env.Value
				}
			}

			for k, v := range tt.expectEnv {
				assert.Contains(t, envMap, k)
				assert.Equal(t, v, envMap[k])
			}
		})
	}
}
