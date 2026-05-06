package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestOtelRouterRouteEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	routerConfigMap, ok := resources.ConfigMaps["suse-observability-router-active"]
	require.True(t, ok, "Active router configmap should exist")
	assert.Contains(t, routerConfigMap.Data["listeners.yaml"], "prefix: \"/receiver/otel/\"")
	assert.Contains(t, routerConfigMap.Data["clusters.yaml"], "name: \"suse-observability-otel-collector\"")
	assert.Contains(t, routerConfigMap.Data["clusters.yaml"], "port_value: 4318")
}
