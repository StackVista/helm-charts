package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

const releaseName = "minio"

func TestMinioBasicTemplate(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have Deployments and no StatefulSets in standalone mode (default)
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")
	assert.Len(t, resources.Statefulsets, 0, "Should have no StatefulSets in standalone mode")
}

func TestMinioStandaloneMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// In standalone mode (default), should have one Deployment and no StatefulSets
	expectedDeployments := []string{
		releaseName,
	}

	assert.Len(t, resources.Deployments, len(expectedDeployments), "Should have exactly %d Deployments in standalone mode", len(expectedDeployments))
	assert.Len(t, resources.Statefulsets, 0, "Should have no StatefulSets in standalone mode")

	// Check each expected Deployment exists
	for _, expectedName := range expectedDeployments {
		_, exists := resources.Deployments[expectedName]
		assert.True(t, exists, "Deployment %s should exist in standalone mode", expectedName)
	}
}

func TestMinioDistributedMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/distributed-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// In distributed mode, should have one StatefulSet and no Deployments
	expectedStatefulSets := []string{
		releaseName,
	}

	assert.Len(t, resources.Statefulsets, len(expectedStatefulSets), "Should have exactly %d StatefulSets in distributed mode", len(expectedStatefulSets))
	assert.Len(t, resources.Deployments, 0, "Should have no Deployments in distributed mode")

	// Check each expected StatefulSet exists
	for _, expectedName := range expectedStatefulSets {
		_, exists := resources.Statefulsets[expectedName]
		assert.True(t, exists, "StatefulSet %s should exist in distributed mode", expectedName)
	}
}

func TestMinioJobsRendered(t *testing.T) {
	// Test with bucket creation enabled to ensure jobs are rendered
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/bucket-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have make-bucket job when bucket creation is enabled
	expectedJobs := []string{
		releaseName + "-make-bucket-job",
	}

	assert.Len(t, resources.Jobs, len(expectedJobs), "Should have exactly %d Jobs when bucket creation is enabled", len(expectedJobs))

	// Check each expected Job exists
	for _, expectedName := range expectedJobs {
		_, exists := resources.Jobs[expectedName]
		assert.True(t, exists, "Job %s should exist when bucket creation is enabled", expectedName)
	}
}

func TestMinioPrometheusJobRendered(t *testing.T) {
	// Test with ServiceMonitor enabled to ensure prometheus job is rendered
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/servicemonitor-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have prometheus-secret job when ServiceMonitor is enabled
	expectedJobs := []string{
		releaseName + "-update-prometheus-secret",
	}

	assert.Len(t, resources.Jobs, len(expectedJobs), "Should have exactly %d Jobs when ServiceMonitor is enabled", len(expectedJobs))

	// Check each expected Job exists
	for _, expectedName := range expectedJobs {
		_, exists := resources.Jobs[expectedName]
		assert.True(t, exists, "Job %s should exist when ServiceMonitor is enabled", expectedName)
	}
}

func TestMinioAllJobsRendered(t *testing.T) {
	// Test with both bucket creation and ServiceMonitor enabled to ensure all jobs are rendered
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/all-jobs-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have both jobs when both features are enabled
	expectedJobs := []string{
		releaseName + "-make-bucket-job",
		releaseName + "-update-prometheus-secret",
	}

	assert.Len(t, resources.Jobs, len(expectedJobs), "Should have exactly %d Jobs when both features are enabled", len(expectedJobs))

	// Check each expected Job exists
	for _, expectedName := range expectedJobs {
		_, exists := resources.Jobs[expectedName]
		assert.True(t, exists, "Job %s should exist when both features are enabled", expectedName)
	}
}
