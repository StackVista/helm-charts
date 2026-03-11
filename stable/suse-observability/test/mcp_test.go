package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestMcpServerEnabledByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-mcp"]
	require.True(t, ok, "MCP server deployment should exist")

	service, ok := resources.Services["suse-observability-mcp"]
	require.True(t, ok, "MCP server service should exist")

	container := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "mcp", container.Name)
	assert.Contains(t, container.Args, "-url")
	assert.Contains(t, container.Args, "http://suse-observability-api-headless:7070")
	assert.Contains(t, container.Args, "-http")
	assert.Contains(t, container.Args, ":8080")
	require.Len(t, service.Spec.Ports, 1)
	assert.Equal(t, int32(8080), service.Spec.Ports[0].Port)

	apiDeployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")
	assert.Contains(t, apiDeployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "CONFIG_FORCE_stackstate_featureSwitches_enableAi",
		Value: "true",
	})
}

func TestMcpServerDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"ai.assistant.enabled": "false",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	assert.NotContains(t, resources.Deployments, "suse-observability-mcp")
	assert.NotContains(t, resources.Services, "suse-observability-mcp")

	apiDeployment, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok, "API deployment should exist")
	assert.NotContains(t, apiDeployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "CONFIG_FORCE_stackstate_featureSwitches_enableAi",
		Value: "true",
	})
}
