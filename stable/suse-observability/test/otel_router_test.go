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
	listeners := routerConfigMap.Data["listeners.yaml"]

	//   /stsAgent/otel/          — direct stsAgent URL
	//   /receiver/stsAgent/otel/ — stsAgent URL proxied through /receiver/
	assert.Contains(t, listeners, "prefix: \"/stsAgent/otel/\"")
	assert.Contains(t, listeners, "prefix: \"/receiver/stsAgent/otel/\"")
	assert.NotContains(t, listeners, "prefix: \"/receiver/otel/\"",
		"old /receiver/otel/ route was based on a non-canonical URL shape and should not be present")

	assert.Contains(t, routerConfigMap.Data["clusters.yaml"], "name: \"suse-observability-otel-collector\"")
	assert.Contains(t, routerConfigMap.Data["clusters.yaml"], "port_value: 4318")
}
