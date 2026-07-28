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

// TestOtelRouterClusterNameStableAcrossReleaseNames asserts the otel-collector cluster address
// uses the subchart's stable fullnameOverride regardless of the Helm release name. A non-default
// release name previously caused the router to reference a non-existent service (503 UH).
func TestOtelRouterClusterNameStableAcrossReleaseNames(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "prime-test", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	routerConfigMap, ok := resources.ConfigMaps["prime-test-suse-observability-router-active"]
	require.True(t, ok, "Active router configmap should exist for non-default release name")

	clusters := routerConfigMap.Data["clusters.yaml"]
	assert.Contains(t, clusters, "address: \"suse-observability-otel-collector\"",
		"router must use the stable otel-collector service name regardless of release name")
	assert.NotContains(t, clusters, "address: \"prime-test-suse-observability-otel-collector\"",
		"router must not use release-prefixed name for otel-collector (service does not exist)")
}
