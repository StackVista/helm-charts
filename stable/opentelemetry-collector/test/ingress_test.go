package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// TestOpenTelemetryCollectorIngressDefaultServiceName tests that the ingress uses the default service name when serviceName is not specified
func TestOpenTelemetryCollectorIngressDefaultServiceName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml", "values/ingress-default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	ingress, ok := resources.Ingresses[releaseName+"-opentelemetry-collector"]
	require.True(t, ok, "Ingress should exist")

	require.Len(t, ingress.Spec.Rules, 1, "Should have exactly 1 rule")
	require.Len(t, ingress.Spec.Rules[0].HTTP.Paths, 1, "Should have exactly 1 path")

	backend := ingress.Spec.Rules[0].HTTP.Paths[0].Backend
	assert.Equal(t, releaseName+"-opentelemetry-collector", backend.Service.Name, "Backend service name should default to the main service")
	assert.Equal(t, int32(4318), backend.Service.Port.Number, "Backend service port should match")
}

// TestOpenTelemetryCollectorIngressCustomServiceName tests that the ingress uses a custom service name when serviceName is specified
func TestOpenTelemetryCollectorIngressCustomServiceName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml", "values/ingress-service-name.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	ingress, ok := resources.Ingresses[releaseName+"-opentelemetry-collector"]
	require.True(t, ok, "Ingress should exist")

	require.Len(t, ingress.Spec.Rules, 1, "Should have exactly 1 rule")
	require.Len(t, ingress.Spec.Rules[0].HTTP.Paths, 2, "Should have exactly 2 paths")

	// First path has a custom serviceName
	grpcPath := ingress.Spec.Rules[0].HTTP.Paths[0]
	assert.Equal(t, "/otlp-grpc", grpcPath.Path, "First path should be /otlp-grpc")
	assert.Equal(t, "otel-opentelemetry-collector-grpc", grpcPath.Backend.Service.Name, "First path should use the custom service name")
	assert.Equal(t, int32(4317), grpcPath.Backend.Service.Port.Number, "First path port should match")

	// Second path has no serviceName, should default to the main service
	httpPath := ingress.Spec.Rules[0].HTTP.Paths[1]
	assert.Equal(t, "/otlp-http", httpPath.Path, "Second path should be /otlp-http")
	assert.Equal(t, releaseName+"-opentelemetry-collector", httpPath.Backend.Service.Name, "Second path should default to the main service name")
	assert.Equal(t, int32(4318), httpPath.Backend.Service.Port.Number, "Second path port should match")
}
