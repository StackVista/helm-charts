package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestMcpServerEnabledByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-mcp-server"]
	require.True(t, ok, "MCP server deployment should exist")

	service, ok := resources.Services["suse-observability-mcp-server"]
	require.True(t, ok, "MCP server service should exist")

	container := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "mcp-server", container.Name)
	assert.Contains(t, container.Args, "-url")
	assert.Contains(t, container.Args, "http://suse-observability-api-headless:7070")
	assert.Contains(t, container.Args, "-http")
	assert.Contains(t, container.Args, ":8080")
	require.Len(t, service.Spec.Ports, 1)
	assert.Equal(t, int32(8080), service.Spec.Ports[0].Port)
}

func TestMcpServerDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"stackstate.components.mcpServer.enabled": "false",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	assert.NotContains(t, resources.Deployments, "suse-observability-mcp-server")
	assert.NotContains(t, resources.Services, "suse-observability-mcp-server")
}
