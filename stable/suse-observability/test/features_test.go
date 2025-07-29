package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	v1 "k8s.io/api/core/v1"
)

const expectedClickhouseConfig = "stackstate.traces.clickHouse ="

func TestFeaturesTracesDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_traces", Value: "true"}
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, expected, "API deployment should have traces feature flag enabled by default")

	configMap, ok := resources.ConfigMaps["suse-observability-api"]
	require.True(t, ok, "API configmap should exist")
	assert.Contains(t, configMap.Data["application_stackstate.conf"], expectedClickhouseConfig, "API configmap should contain traces ClickHouse configuration by default")
}

func TestFeaturesTracesDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.traces": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	notExpected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_traces", Value: "true"}
	assert.NotContains(t, deployment.Spec.Template.Spec.Containers[0].Env, notExpected, "API deployment should not have traces feature flag when disabled")

	configMap, ok := resources.ConfigMaps["suse-observability-api"]
	require.True(t, ok, "API configmap should exist")
	assert.NotContains(t, configMap.Data["application_stackstate.conf"], expectedClickhouseConfig, "API configmap should not contain traces ClickHouse configuration when traces disabled")
}

func TestFeaturesTracesExperimentalOverFeatures(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.traces":     "true",
			"stackstate.experimental.traces": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	notExpected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_traces", Value: "true"}
	assert.NotContains(t, deployment.Spec.Template.Spec.Containers[0].Env, notExpected, "API deployment should not have traces feature flag when experimental override disables it")

	configMap, ok := resources.ConfigMaps["suse-observability-api"]
	require.True(t, ok, "API configmap should exist")
	assert.NotContains(t, configMap.Data["application_stackstate.conf"], expectedClickhouseConfig, "API configmap should not contain traces ClickHouse configuration when experimental override disables traces")
}

func TestFeaturesDashboardsDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	notExpected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_dashboards", Value: "true"}
	assert.NotContains(t, deployment.Spec.Template.Spec.Containers[0].Env, notExpected, "API deployment should not have dashboards feature flag enabled by default")
}

func TestFeaturesDashboardsDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.dashboards": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_dashboards", Value: "true"}
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, expected, "API deployment should have dashboards feature flag when enabled")
}

func TestFeaturesDashboardsExperimentalOverFeatures(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.dashboards":     "false",
			"stackstate.experimental.dashboards": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")

	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_dashboards", Value: "true"}
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, expected, "API deployment should have dashboards feature flag when experimental override enables it")
}
