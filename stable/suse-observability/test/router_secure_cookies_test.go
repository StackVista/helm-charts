package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestRouterSecureCookiesEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues:   map[string]string{"stackstate.components.router.secureCookies.enabled": "true"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	routerConfigMap, ok := resources.ConfigMaps["suse-observability-router-active"]
	require.True(t, ok, "Active router configmap should exist")
	listeners := routerConfigMap.Data["listeners.yaml"]

	assert.Contains(t, listeners, "envoy.filters.http.lua",
		"the Lua filter should be present when secureCookies is enabled")
	assert.Contains(t, listeners, "x-forwarded-proto",
		"the filter should gate on the X-Forwarded-Proto header")
	assert.Contains(t, listeners, "; Secure",
		"the filter should append the Secure cookie attribute")
	assert.Contains(t, listeners, "envoy.filters.http.router",
		"the router filter must remain present (and last) in the chain")
}

func TestRouterSecureCookiesDisabledByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	routerConfigMap, ok := resources.ConfigMaps["suse-observability-router-active"]
	require.True(t, ok, "Active router configmap should exist")

	assert.NotContains(t, routerConfigMap.Data["listeners.yaml"], "envoy.filters.http.lua",
		"the Lua filter should be absent unless secureCookies is enabled")
}
