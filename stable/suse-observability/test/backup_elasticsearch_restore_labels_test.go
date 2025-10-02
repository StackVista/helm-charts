package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
)

// expectedESRestoreScaleDownLabels are the expected labels for ES restore scale-down
var expectedESRestoreScaleDownLabels = map[string]string{
	"observability.suse.com/scalable-during-es-restore": "true",
}

func TestElasticsearchRestoreScaleDownLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Deployments that should have ES restore scale-down labels
	expectedDeployments := []string{
		"suse-observability-e2es",
		"suse-observability-receiver",
	}

	// Test that the expected deployments have the ES restore scale-down labels
	for _, deploymentName := range expectedDeployments {
		deployment, exists := resources.Deployments[deploymentName]
		assert.True(t, exists, "Deployment '%s' should exist", deploymentName)
		if exists {
			testDeploymentHasESRestoreLabels(t, deployment, expectedESRestoreScaleDownLabels)
		}
	}
}

func TestElasticsearchRestoreScaleDownLabelsSplitMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/receiver_split_enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Deployments that should have ES restore scale-down labels in split mode
	expectedDeployments := []string{
		"suse-observability-e2es",
		"suse-observability-receiver-base",
		"suse-observability-receiver-logs",
	}

	// Deployments that should NOT have ES restore scale-down labels in split mode
	unexpectedDeployments := []string{
		"suse-observability-receiver-process-agent",
	}

	// Test that the expected deployments have the ES restore scale-down labels
	for _, deploymentName := range expectedDeployments {
		deployment, exists := resources.Deployments[deploymentName]
		t.Logf("Deployment: %s", deployment.Name)
		assert.True(t, exists, "Deployment '%s' should exist", deploymentName)
		if exists {
			testDeploymentHasESRestoreLabels(t, deployment, expectedESRestoreScaleDownLabels)
		}
	}

	// Test that the unexpected deployments do NOT have the ES restore scale-down labels
	for _, deploymentName := range unexpectedDeployments {
		deployment, exists := resources.Deployments[deploymentName]
		assert.True(t, exists, "Deployment '%s' should exist", deploymentName)
		if exists {
			testDeploymentDoesNotHaveESRestoreLabels(t, deployment, expectedESRestoreScaleDownLabels)
		}
	}
}

func testDeploymentHasESRestoreLabels(t *testing.T, deployment appsv1.Deployment, expectedLabels map[string]string) {
	// Test deployment metadata labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := deployment.Labels[labelKey]
		assert.True(t, exists, "Deployment '%s' should have ES restore label '%s'", deployment.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "Deployment '%s' ES restore label '%s' should have value '%s', got '%s'", deployment.Name, labelKey, expectedValue, actualValue)
	}
}

func testDeploymentDoesNotHaveESRestoreLabels(t *testing.T, deployment appsv1.Deployment, unexpectedLabels map[string]string) {
	// Test deployment metadata labels should NOT contain the ES restore labels
	for labelKey := range unexpectedLabels {
		_, exists := deployment.Labels[labelKey]
		assert.False(t, exists, "Deployment '%s' should NOT have ES restore label '%s'", deployment.Name, labelKey)
	}
}
