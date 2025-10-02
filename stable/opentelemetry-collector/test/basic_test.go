package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

const releaseName = "otel"

func TestOpenTelemetryCollectorBasicTemplate(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have Deployments (default mode)
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")
}

func TestOpenTelemetryCollectorDeploymentMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/deployment-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// In Deployment mode, should have a Deployment
	expectedDeployments := []string{
		releaseName + "-opentelemetry-collector",
	}

	assert.Len(t, resources.Deployments, len(expectedDeployments), "Should have exactly %d Deployments in Deployment mode", len(expectedDeployments))

	// Check each expected Deployment exists
	for _, expectedName := range expectedDeployments {
		found := false
		for _, deployment := range resources.Deployments {
			if deployment.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "Deployment %s should exist in Deployment mode", expectedName)
	}

	// Should not have DaemonSets or StatefulSets
	assert.Len(t, resources.DaemonSets, 0, "Should not have DaemonSets in Deployment mode")
	assert.Len(t, resources.Statefulsets, 0, "Should not have StatefulSets in Deployment mode")
}

func TestOpenTelemetryCollectorDaemonSetMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/daemonset-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// In DaemonSet mode, should have a DaemonSet
	expectedDaemonSets := []string{
		releaseName + "-opentelemetry-collector-agent",
	}

	assert.Len(t, resources.DaemonSets, len(expectedDaemonSets), "Should have exactly %d DaemonSets in DaemonSet mode", len(expectedDaemonSets))

	// Check each expected DaemonSet exists
	for _, expectedName := range expectedDaemonSets {
		found := false
		for _, daemonset := range resources.DaemonSets {
			if daemonset.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "DaemonSet %s should exist in DaemonSet mode", expectedName)
	}

	// Should not have Deployments or StatefulSets
	assert.Len(t, resources.Deployments, 0, "Should not have Deployments in DaemonSet mode")
	assert.Len(t, resources.Statefulsets, 0, "Should not have StatefulSets in DaemonSet mode")
}

func TestOpenTelemetryCollectorStatefulSetMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/statefulset-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// In StatefulSet mode, should have a StatefulSet
	expectedStatefulSets := []string{
		releaseName + "-opentelemetry-collector",
	}

	assert.Len(t, resources.Statefulsets, len(expectedStatefulSets), "Should have exactly %d StatefulSets in StatefulSet mode", len(expectedStatefulSets))

	// Check each expected StatefulSet exists
	for _, expectedName := range expectedStatefulSets {
		found := false
		for _, statefulset := range resources.Statefulsets {
			if statefulset.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "StatefulSet %s should exist in StatefulSet mode", expectedName)
	}

	// Should not have Deployments or DaemonSets
	assert.Len(t, resources.Deployments, 0, "Should not have Deployments in StatefulSet mode")
	assert.Len(t, resources.DaemonSets, 0, "Should not have DaemonSets in StatefulSet mode")
}

func TestOpenTelemetryCollectorConfigSelection(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	defaultCollectorConfig := resources.ConfigMaps["otel-opentelemetry-collector"].Data["relay"]

	assert.NotContains(t, defaultCollectorConfig, "stssettingsextension")
	assert.NotContains(t, defaultCollectorConfig, "tracetotopo")
	assert.NotContains(t, defaultCollectorConfig, "stskafkaexporter")

	// with the global.features.enableStackPacks2 set to true
	output = helmtestutil.RenderHelmTemplate(t, releaseName, "values/enable-stackpacks2.yaml")
	resources = helmtestutil.NewKubernetesResources(t, output)

	stackPacks2CollectorConfig := resources.ConfigMaps["otel-opentelemetry-collector"].Data["relay"]

	assert.Contains(t, stackPacks2CollectorConfig, "stssettingsextension")
	assert.Contains(t, stackPacks2CollectorConfig, "tracetotopo")
	assert.Contains(t, stackPacks2CollectorConfig, "stskafkaexporter")
}
