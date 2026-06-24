package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	gatewayChartName   = "trafficmirror"
	gatewayBackendPort = gatewayv1.PortNumber(8080)
)

var gatewayBaseValues = []string{}

func gatewayRender(t *testing.T, files ...string) helmtestutil.KubernetesResources {
	output := helmtestutil.RenderHelmTemplate(t, gatewayChartName, append(append([]string{}, gatewayBaseValues...), files...)...)
	return helmtestutil.NewKubernetesResources(t, output)
}

func onlyHTTPRoute(t *testing.T, r helmtestutil.KubernetesResources) gatewayv1.HTTPRoute {
	require.Len(t, r.HTTPRoutes, 1, "exactly one HTTPRoute should be rendered")
	for _, route := range r.HTTPRoutes {
		return route
	}
	return gatewayv1.HTTPRoute{}
}

func TestGatewayDisabledByDefault(t *testing.T) {
	resources := gatewayRender(t)
	assert.Empty(t, resources.HTTPRoutes, "HTTPRoutes should be empty when gateway is disabled")
}

func TestGatewayEnabled(t *testing.T) {
	resources := gatewayRender(t, "values/gateway_enabled.yaml")
	route := onlyHTTPRoute(t, resources)

	require.Len(t, route.Spec.ParentRefs, 1, "ParentRefs should have one entry")
	assert.Equal(t, gatewayv1.ObjectName("my-gateway"), route.Spec.ParentRefs[0].Name)
	require.NotNil(t, route.Spec.ParentRefs[0].Namespace)
	assert.Equal(t, gatewayv1.Namespace("gateway-namespace"), *route.Spec.ParentRefs[0].Namespace)
	require.NotNil(t, route.Spec.ParentRefs[0].SectionName)
	assert.Equal(t, gatewayv1.SectionName("https"), *route.Spec.ParentRefs[0].SectionName)

	require.Len(t, route.Spec.Hostnames, 1, "Hostnames should have one entry")
	assert.Equal(t, gatewayv1.Hostname("test.example.com"), route.Spec.Hostnames[0])

	require.Len(t, route.Spec.Rules, 1, "Rules should have one entry")
	rule := route.Spec.Rules[0]
	require.Len(t, rule.Matches, 1, "Matches should have one entry")
	require.NotNil(t, rule.Matches[0].Path)
	assert.Equal(t, gatewayv1.PathMatchPathPrefix, *rule.Matches[0].Path.Type)
	assert.Equal(t, "/", *rule.Matches[0].Path.Value)

	require.Len(t, rule.BackendRefs, 1, "BackendRefs should have one entry")
	require.NotNil(t, rule.BackendRefs[0].Port)
	assert.Equal(t, gatewayBackendPort, *rule.BackendRefs[0].Port)

	assert.Empty(t, resources.Ingresses, "Ingress should not exist when gateway mode is active")
}

func TestGatewayFullConfiguration(t *testing.T) {
	resources := gatewayRender(t, "values/gateway_full.yaml")
	route := onlyHTTPRoute(t, resources)

	assert.Equal(t, "value", route.Annotations["gateway.example.com/custom"])

	require.Len(t, route.Spec.ParentRefs, 1)
	require.NotNil(t, route.Spec.ParentRefs[0].Port)
	assert.Equal(t, gatewayv1.PortNumber(443), *route.Spec.ParentRefs[0].Port)

	require.Len(t, route.Spec.Hostnames, 2, "Should have 2 hostnames")
	assert.Equal(t, gatewayv1.Hostname("test.example.com"), route.Spec.Hostnames[0])
	assert.Equal(t, gatewayv1.Hostname("*.test.example.com"), route.Spec.Hostnames[1])

	rule := route.Spec.Rules[0]
	assert.Equal(t, "/api", *rule.Matches[0].Path.Value)

	require.NotNil(t, rule.Timeouts)
	require.NotNil(t, rule.Timeouts.Request)
	assert.Equal(t, gatewayv1.Duration("30s"), *rule.Timeouts.Request)
	require.NotNil(t, rule.Timeouts.BackendRequest)
	assert.Equal(t, gatewayv1.Duration("60s"), *rule.Timeouts.BackendRequest)

	require.NotNil(t, rule.BackendRefs[0].Weight)
	assert.Equal(t, int32(100), *rule.BackendRefs[0].Weight)
}

func TestGatewayAndIngressMutualExclusion(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, gatewayChartName, append(append([]string{}, gatewayBaseValues...), "values/gateway_with_ingress.yaml")...)
	require.Contains(t, err.Error(), "Cannot configure both ingress.enabled and gateway.enabled simultaneously")
}

func TestGatewayMissingParentRefs(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, gatewayChartName, append(append([]string{}, gatewayBaseValues...), "values/gateway_missing_parentrefs.yaml")...)
	require.Contains(t, err.Error(), "Gateway API requires gateway.parentRefs to be set")
}
