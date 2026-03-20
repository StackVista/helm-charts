package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestGatewayHTTPRouteEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/gateway-http.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Len(t, resources.HTTPRoutes, 1, "Should have exactly 1 HTTPRoute")

	routeName := releaseName + "-opentelemetry-collector"
	route, ok := resources.HTTPRoutes[routeName]
	require.True(t, ok, "HTTPRoute %s should exist", routeName)

	require.Len(t, route.Spec.ParentRefs, 1, "Should have 1 parentRef")
	assert.Equal(t, gatewayv1.ObjectName("my-gateway"), route.Spec.ParentRefs[0].Name)
	require.NotNil(t, route.Spec.ParentRefs[0].Namespace)
	assert.Equal(t, gatewayv1.Namespace("gateway-namespace"), *route.Spec.ParentRefs[0].Namespace)

	require.Len(t, route.Spec.Hostnames, 1, "Should have 1 hostname")
	assert.Equal(t, gatewayv1.Hostname("collector.example.com"), route.Spec.Hostnames[0])

	require.Len(t, route.Spec.Rules, 1, "Should have 1 rule")
	require.Len(t, route.Spec.Rules[0].BackendRefs, 1, "Should have 1 backendRef")

	backendRef := route.Spec.Rules[0].BackendRefs[0]
	assert.Equal(t, gatewayv1.ObjectName(routeName), backendRef.Name)
	require.NotNil(t, backendRef.Port)
	assert.Equal(t, gatewayv1.PortNumber(4318), *backendRef.Port)
}

func TestGatewayGRPCRouteEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/gateway-grpc.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Len(t, resources.GRPCRoutes, 1, "Should have exactly 1 GRPCRoute")

	routeName := releaseName + "-opentelemetry-collector"
	route, ok := resources.GRPCRoutes[routeName]
	require.True(t, ok, "GRPCRoute %s should exist", routeName)

	require.Len(t, route.Spec.ParentRefs, 1, "Should have 1 parentRef")
	assert.Equal(t, gatewayv1.ObjectName("my-gateway"), route.Spec.ParentRefs[0].Name)

	require.Len(t, route.Spec.Hostnames, 1, "Should have 1 hostname")
	assert.Equal(t, gatewayv1.Hostname("collector-grpc.example.com"), route.Spec.Hostnames[0])

	require.Len(t, route.Spec.Rules, 1, "Should have 1 rule")
	backendRef := route.Spec.Rules[0].BackendRefs[0]
	assert.Equal(t, gatewayv1.ObjectName(routeName), backendRef.Name)
	require.NotNil(t, backendRef.Port)
	assert.Equal(t, gatewayv1.PortNumber(4317), *backendRef.Port)
}

func TestGatewayBothHTTPAndGRPCRoutes(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/gateway-both.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Len(t, resources.HTTPRoutes, 1, "Should have exactly 1 HTTPRoute")
	require.Len(t, resources.GRPCRoutes, 1, "Should have exactly 1 GRPCRoute")

	httpRouteName := releaseName + "-opentelemetry-collector"
	httpRoute, ok := resources.HTTPRoutes[httpRouteName]
	require.True(t, ok, "HTTPRoute %s should exist", httpRouteName)
	require.Len(t, httpRoute.Spec.Rules, 1)
	assert.Equal(t, gatewayv1.PortNumber(4318), *httpRoute.Spec.Rules[0].BackendRefs[0].Port)

	grpcRouteName := releaseName + "-opentelemetry-collector-grpc"
	grpcRoute, ok := resources.GRPCRoutes[grpcRouteName]
	require.True(t, ok, "GRPCRoute %s should exist", grpcRouteName)
	require.Len(t, grpcRoute.Spec.Rules, 1)
	assert.Equal(t, gatewayv1.PortNumber(4317), *grpcRoute.Spec.Rules[0].BackendRefs[0].Port)
}

func TestGatewayDisabledByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.HTTPRoutes, 0, "Default should not have HTTPRoutes")
	assert.Len(t, resources.GRPCRoutes, 0, "Default should not have GRPCRoutes")
}
