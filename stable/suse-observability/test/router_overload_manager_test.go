package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"sigs.k8s.io/yaml"
)

type envoyTriggeredAction struct {
	Name     string `json:"name"`
	Triggers []struct {
		Name      string `json:"name"`
		Threshold struct {
			Value float64 `json:"value"`
		} `json:"threshold"`
	} `json:"triggers"`
}

type envoyBootstrap struct {
	OverloadManager *struct {
		ResourceMonitors []struct {
			Name        string                 `json:"name"`
			TypedConfig map[string]interface{} `json:"typed_config"`
		} `json:"resource_monitors"`
		Actions []envoyTriggeredAction `json:"actions"`
	} `json:"overload_manager"`
}

type envoyResources struct {
	Resources []map[string]interface{} `json:"resources"`
}

func renderRouter(t *testing.T, setValues map[string]string) helmtestutil.KubernetesResources {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues:   setValues,
	})
	return helmtestutil.NewKubernetesResources(t, output)
}

func routerBootstrap(t *testing.T, resources helmtestutil.KubernetesResources) envoyBootstrap {
	cm, ok := resources.ConfigMaps["suse-observability-router"]
	require.True(t, ok, "router configmap should exist")
	var bootstrap envoyBootstrap
	require.NoError(t, yaml.Unmarshal([]byte(cm.Data["envoy.yaml"]), &bootstrap))
	return bootstrap
}

func routerDynamic(t *testing.T, resources helmtestutil.KubernetesResources, key string) envoyResources {
	return routerDynamicConfigMap(t, resources, "suse-observability-router-active", key)
}

func routerDynamicConfigMap(t *testing.T, resources helmtestutil.KubernetesResources, name, key string) envoyResources {
	cm, ok := resources.ConfigMaps[name]
	require.True(t, ok, "router configmap %s should exist", name)
	var parsed envoyResources
	require.NoError(t, yaml.Unmarshal([]byte(cm.Data[key]), &parsed))
	require.NotEmpty(t, parsed.Resources)
	return parsed
}

func TestRouterOverloadManagerPerProfile(t *testing.T) {
	for _, profile := range []string{"", "10-nonha", "150-ha", "250-ha", "500-ha", "4000-ha"} {
		t.Run("profile="+profile, func(t *testing.T) {
			t.Parallel()
			values := map[string]string{
				"global.suseObservability.adminPassword": "test-password",
				"global.suseObservability.baseUrl":       "http://localhost",
				"global.suseObservability.license":       "test-license",
			}
			if profile != "" {
				values["global.suseObservability.sizing.profile"] = profile
			}
			om := routerBootstrap(t, renderRouter(t, values)).OverloadManager
			require.NotNil(t, om)

			monitors := map[string]map[string]interface{}{}
			for _, m := range om.ResourceMonitors {
				monitors[m.Name] = m.TypedConfig
			}
			assert.Contains(t, monitors, "envoy.resource_monitors.cgroup_memory")
			require.Contains(t, monitors, "envoy.resource_monitors.global_downstream_max_connections")
			assert.EqualValues(t, 50000, monitors["envoy.resource_monitors.global_downstream_max_connections"]["max_active_downstream_connections"])

			thresholds := map[string]float64{}
			for _, a := range om.Actions {
				require.Len(t, a.Triggers, 1)
				assert.Equal(t, "envoy.resource_monitors.cgroup_memory", a.Triggers[0].Name)
				thresholds[a.Name] = a.Triggers[0].Threshold.Value
			}
			assert.Equal(t, map[string]float64{
				"envoy.overload_actions.shrink_heap":                 0.5,
				"envoy.overload_actions.disable_http_keepalive":      0.8,
				"envoy.overload_actions.stop_accepting_requests":     0.8,
				"envoy.overload_actions.reject_incoming_connections": 0.9,
			}, thresholds)
		})
	}
}

func TestRouterBufferLimits(t *testing.T) {
	for mode, configMap := range map[string]string{
		"active":      "suse-observability-router-active",
		"maintenance": "suse-observability-router-maintenance",
	} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			resources := renderRouter(t, map[string]string{
				"stackstate.components.router.mode.status": mode,
			})
			listener := routerDynamicConfigMap(t, resources, configMap, "listeners.yaml").Resources[0]
			assert.EqualValues(t, 65536, listener["per_connection_buffer_limit_bytes"])
			listenerOptions := listener["socket_options"].([]interface{})
			require.Len(t, listenerOptions, 1)
			rcvbuf := listenerOptions[0].(map[string]interface{})
			assert.EqualValues(t, 1, rcvbuf["level"], "SOL_SOCKET")
			assert.EqualValues(t, 8, rcvbuf["name"], "SO_RCVBUF")
			assert.EqualValues(t, 65536, rcvbuf["int_value"])

			for _, cluster := range routerDynamicConfigMap(t, resources, configMap, "clusters.yaml").Resources {
				name := cluster["name"]
				assert.EqualValues(t, 65536, cluster["per_connection_buffer_limit_bytes"], "cluster %s", name)
				bind, ok := cluster["upstream_bind_config"].(map[string]interface{})
				require.True(t, ok, "cluster %s should bound its upstream socket send buffer", name)
				options := bind["socket_options"].([]interface{})
				require.Len(t, options, 1)
				option := options[0].(map[string]interface{})
				assert.EqualValues(t, 1, option["level"], "SOL_SOCKET")
				assert.EqualValues(t, 7, option["name"], "SO_SNDBUF")
				assert.EqualValues(t, 65536, option["int_value"])
			}
		})
	}
}

func TestRouterBufferLimitsAutomaticMode(t *testing.T) {
	t.Parallel()
	resources := renderRouter(t, map[string]string{
		"stackstate.components.router.mode.status": "automatic",
	})
	scripts, ok := resources.ConfigMaps["suse-observability-router-mode-scripts"]
	require.True(t, ok)
	for _, script := range []string{"set-active.sh", "set-maintenance.sh"} {
		assert.Contains(t, scripts.Data[script], "per_connection_buffer_limit_bytes: 65536", script)
		assert.Contains(t, scripts.Data[script], "description: SO_SNDBUF", script)
		assert.Contains(t, scripts.Data[script], "description: SO_RCVBUF", script)
	}
	assert.NotNil(t, routerBootstrap(t, resources).OverloadManager)
}
