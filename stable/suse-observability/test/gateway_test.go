package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestGatewayDisabledByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Empty(t, resources.HTTPRoutes, "HTTPRoutes should be empty when gateway is disabled")
}

func TestGatewayEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/gateway_enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Contains(t, resources.HTTPRoutes, "suse-observability", "suse-observability HTTPRoute should exist")

	httpRoute, ok := resources.HTTPRoutes["suse-observability"]
	require.True(t, ok, "HTTPRoute named 'suse-observability' should exist")

	// Verify parentRefs from gateway config
	require.Len(t, httpRoute.Spec.ParentRefs, 1, "ParentRefs should have one entry")
	assert.Equal(t, gatewayv1.ObjectName("my-gateway"), httpRoute.Spec.ParentRefs[0].Name)
	require.NotNil(t, httpRoute.Spec.ParentRefs[0].Namespace)
	assert.Equal(t, gatewayv1.Namespace("gateway-namespace"), *httpRoute.Spec.ParentRefs[0].Namespace)
	require.NotNil(t, httpRoute.Spec.ParentRefs[0].SectionName)
	assert.Equal(t, gatewayv1.SectionName("https"), *httpRoute.Spec.ParentRefs[0].SectionName)

	// Verify hostnames derived from baseUrl
	require.Len(t, httpRoute.Spec.Hostnames, 1, "Hostnames should have one entry")
	assert.Equal(t, gatewayv1.Hostname("test.example.com"), httpRoute.Spec.Hostnames[0])

	// Verify rules
	require.Len(t, httpRoute.Spec.Rules, 1, "Rules should have one entry")
	rule := httpRoute.Spec.Rules[0]

	require.Len(t, rule.Matches, 1, "Matches should have one entry")
	require.NotNil(t, rule.Matches[0].Path)
	assert.Equal(t, gatewayv1.PathMatchPathPrefix, *rule.Matches[0].Path.Type)
	assert.Equal(t, "/", *rule.Matches[0].Path.Value)

	require.Len(t, rule.BackendRefs, 1, "BackendRefs should have one entry")
	assert.Equal(t, gatewayv1.ObjectName("suse-observability-router"), rule.BackendRefs[0].Name)
	assert.Equal(t, gatewayv1.PortNumber(8080), *rule.BackendRefs[0].Port)

	// Should NOT create an Ingress for suse-observability
	_, hasIngress := resources.Ingresses["suse-observability"]
	assert.False(t, hasIngress, "Ingress should not exist when gateway mode is active")
}

func TestGatewayFullConfiguration(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/gateway_full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	httpRoute, ok := resources.HTTPRoutes["suse-observability"]
	require.True(t, ok, "HTTPRoute should exist")

	// Verify annotations
	assert.Equal(t, "value", httpRoute.Annotations["gateway.example.com/custom"])

	// Verify parentRefs with port
	require.Len(t, httpRoute.Spec.ParentRefs, 1)
	require.NotNil(t, httpRoute.Spec.ParentRefs[0].Port)
	assert.Equal(t, gatewayv1.PortNumber(443), *httpRoute.Spec.ParentRefs[0].Port)

	// Verify multiple hostnames
	require.Len(t, httpRoute.Spec.Hostnames, 2, "Should have 2 hostnames")
	assert.Equal(t, gatewayv1.Hostname("test.example.com"), httpRoute.Spec.Hostnames[0])
	assert.Equal(t, gatewayv1.Hostname("*.test.example.com"), httpRoute.Spec.Hostnames[1])

	// Verify custom path
	rule := httpRoute.Spec.Rules[0]
	assert.Equal(t, "/api", *rule.Matches[0].Path.Value)

	// Verify timeouts
	require.NotNil(t, rule.Timeouts)
	require.NotNil(t, rule.Timeouts.Request)
	assert.Equal(t, gatewayv1.Duration("30s"), *rule.Timeouts.Request)
	require.NotNil(t, rule.Timeouts.BackendRequest)
	assert.Equal(t, gatewayv1.Duration("60s"), *rule.Timeouts.BackendRequest)

	// Verify backend weight
	assert.Equal(t, int32(100), *rule.BackendRefs[0].Weight)
}

func TestGatewayMissingParentRefs(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/gateway_missing_parentrefs.yaml")
	require.Contains(t, err.Error(), "Gateway API requires gateway.parentRefs to be set")
}

func TestGatewayUseBaseUrl(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/gateway_use_baseurl.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	httpRoute, ok := resources.HTTPRoutes["suse-observability"]
	require.True(t, ok, "HTTPRoute should exist")

	assert.Equal(t, gatewayv1.ObjectName("my-gateway"), httpRoute.Spec.ParentRefs[0].Name)
	assert.Equal(t, gatewayv1.Hostname("test.example.com"), httpRoute.Spec.Hostnames[0])
	assert.Equal(t, gatewayv1.PathMatchPathPrefix, *httpRoute.Spec.Rules[0].Matches[0].Path.Type)
	assert.Equal(t, "/observability", *httpRoute.Spec.Rules[0].Matches[0].Path.Value)
}

func TestGatewayOverrideBaseUrlPath(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/gateway_override_baseurl_path.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	httpRoute, ok := resources.HTTPRoutes["suse-observability"]
	require.True(t, ok, "HTTPRoute should exist")

	assert.Equal(t, gatewayv1.ObjectName("my-gateway"), httpRoute.Spec.ParentRefs[0].Name)
	assert.Equal(t, gatewayv1.Hostname("test.example.com"), httpRoute.Spec.Hostnames[0])
	assert.Equal(t, "/api", *httpRoute.Spec.Rules[0].Matches[0].Path.Value)
}

func TestGatewayOverrideBaseUrlHost(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/gateway_override_baseurl_host.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	httpRoute, ok := resources.HTTPRoutes["suse-observability"]
	require.True(t, ok, "HTTPRoute should exist")

	assert.Equal(t, gatewayv1.ObjectName("my-gateway"), httpRoute.Spec.ParentRefs[0].Name)
	assert.Equal(t, gatewayv1.Hostname("override.example.com"), httpRoute.Spec.Hostnames[0])
}

func TestIngressStillWorks(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"ingress.enabled":       "true",
			"ingress.hosts[0].host": "test.example.com",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Len(t, resources.Ingresses, 1, "Ingress should exist with ingress.enabled=true")
	assert.Empty(t, resources.HTTPRoutes, "HTTPRoutes should be empty")
}
