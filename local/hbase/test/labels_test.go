package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// TestHBaseGlobalCommonLabelsDistributedMode tests that global.commonLabels are applied in Distributed mode
func TestHBaseGlobalCommonLabelsDistributedMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "hbase", "values/global-common-labels-distributed.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have multiple StatefulSets in distributed mode
	assert.Greater(t, len(resources.Statefulsets), 0, "Should have StatefulSets")
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")

	// Test StatefulSets have global common labels
	for _, statefulset := range resources.Statefulsets {
		// Check StatefulSet labels
		assert.Contains(t, statefulset.Labels, "global.label1", "StatefulSet %s should have global.label1", statefulset.Name)
		assert.Equal(t, "global-value1", statefulset.Labels["global.label1"], "StatefulSet %s global.label1 should have correct value", statefulset.Name)
		assert.Contains(t, statefulset.Labels, "global.label2", "StatefulSet %s should have global.label2", statefulset.Name)
		assert.Equal(t, "global-value2", statefulset.Labels["global.label2"], "StatefulSet %s global.label2 should have correct value", statefulset.Name)

		// Check Pod template labels
		podLabels := statefulset.Spec.Template.Labels
		assert.Contains(t, podLabels, "global.label1", "Pod template in %s should have global.label1", statefulset.Name)
		assert.Equal(t, "global-value1", podLabels["global.label1"], "Pod template in %s global.label1 should have correct value", statefulset.Name)
		assert.Contains(t, podLabels, "global.label2", "Pod template in %s should have global.label2", statefulset.Name)
		assert.Equal(t, "global-value2", podLabels["global.label2"], "Pod template in %s global.label2 should have correct value", statefulset.Name)
	}

	// Test Deployments have global common labels
	for _, deployment := range resources.Deployments {
		// Check Deployment labels
		assert.Contains(t, deployment.Labels, "global.label1", "Deployment %s should have global.label1", deployment.Name)
		assert.Equal(t, "global-value1", deployment.Labels["global.label1"], "Deployment %s global.label1 should have correct value", deployment.Name)
		assert.Contains(t, deployment.Labels, "global.label2", "Deployment %s should have global.label2", deployment.Name)
		assert.Equal(t, "global-value2", deployment.Labels["global.label2"], "Deployment %s global.label2 should have correct value", deployment.Name)

		// Check Pod template labels
		podLabels := deployment.Spec.Template.Labels
		assert.Contains(t, podLabels, "global.label1", "Pod template in %s should have global.label1", deployment.Name)
		assert.Equal(t, "global-value1", podLabels["global.label1"], "Pod template in %s global.label1 should have correct value", deployment.Name)
		assert.Contains(t, podLabels, "global.label2", "Pod template in %s should have global.label2", deployment.Name)
		assert.Equal(t, "global-value2", podLabels["global.label2"], "Pod template in %s global.label2 should have correct value", deployment.Name)
	}
}

// TestHBaseGlobalCommonLabelsMonoMode tests that global.commonLabels are applied in Mono mode
func TestHBaseGlobalCommonLabelsMonoMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "hbase", "values/global-common-labels-mono.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have StatefulSets and Deployments in mono mode
	assert.Greater(t, len(resources.Statefulsets), 0, "Should have StatefulSets")
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")

	// Test StatefulSets have global common labels
	for _, statefulset := range resources.Statefulsets {
		// Check StatefulSet labels
		assert.Contains(t, statefulset.Labels, "global.label1", "StatefulSet %s should have global.label1", statefulset.Name)
		assert.Equal(t, "global-value1", statefulset.Labels["global.label1"], "StatefulSet %s global.label1 should have correct value", statefulset.Name)
		assert.Contains(t, statefulset.Labels, "global.label2", "StatefulSet %s should have global.label2", statefulset.Name)
		assert.Equal(t, "global-value2", statefulset.Labels["global.label2"], "StatefulSet %s global.label2 should have correct value", statefulset.Name)

		// Check Pod template labels
		podLabels := statefulset.Spec.Template.Labels
		assert.Contains(t, podLabels, "global.label1", "Pod template in %s should have global.label1", statefulset.Name)
		assert.Equal(t, "global-value1", podLabels["global.label1"], "Pod template in %s global.label1 should have correct value", statefulset.Name)
		assert.Contains(t, podLabels, "global.label2", "Pod template in %s should have global.label2", statefulset.Name)
		assert.Equal(t, "global-value2", podLabels["global.label2"], "Pod template in %s global.label2 should have correct value", statefulset.Name)
	}

	// Test Deployments have global common labels
	for _, deployment := range resources.Deployments {
		// Check Deployment labels
		assert.Contains(t, deployment.Labels, "global.label1", "Deployment %s should have global.label1", deployment.Name)
		assert.Equal(t, "global-value1", deployment.Labels["global.label1"], "Deployment %s global.label1 should have correct value", deployment.Name)
		assert.Contains(t, deployment.Labels, "global.label2", "Deployment %s should have global.label2", deployment.Name)
		assert.Equal(t, "global-value2", deployment.Labels["global.label2"], "Deployment %s global.label2 should have correct value", deployment.Name)

		// Check Pod template labels
		podLabels := deployment.Spec.Template.Labels
		assert.Contains(t, podLabels, "global.label1", "Pod template in %s should have global.label1", deployment.Name)
		assert.Equal(t, "global-value1", podLabels["global.label1"], "Pod template in %s global.label1 should have correct value", deployment.Name)
		assert.Contains(t, podLabels, "global.label2", "Pod template in %s should have global.label2", deployment.Name)
		assert.Equal(t, "global-value2", podLabels["global.label2"], "Pod template in %s global.label2 should have correct value", deployment.Name)
	}
}

// TestHBaseDefaultLabels tests that default chart labels are always present
func TestHBaseDefaultLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "hbase", "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have multiple resources
	assert.Greater(t, len(resources.Statefulsets), 0, "Should have StatefulSets")
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")

	// Test all StatefulSets have default labels
	for _, statefulset := range resources.Statefulsets {
		// Check that default component labels are present on StatefulSet
		assert.Contains(t, statefulset.Labels, "app.kubernetes.io/component", "StatefulSet %s should have app.kubernetes.io/component label", statefulset.Name)

		// Check that default component labels are present on Pod template
		podLabels := statefulset.Spec.Template.Labels
		assert.Contains(t, podLabels, "app.kubernetes.io/component", "Pod template in %s should have app.kubernetes.io/component label", statefulset.Name)
	}

	// Test all Deployments have default labels
	for _, deployment := range resources.Deployments {
		// Check that default component labels are present on Deployment
		assert.Contains(t, deployment.Labels, "app.kubernetes.io/component", "Deployment %s should have app.kubernetes.io/component label", deployment.Name)

		// Check that default component labels are present on Pod template
		podLabels := deployment.Spec.Template.Labels
		assert.Contains(t, podLabels, "app.kubernetes.io/component", "Pod template in %s should have app.kubernetes.io/component label", deployment.Name)
	}
}

// TestHBaseGlobalCommonLabelsDoesNotOverrideDefaults tests that global common labels don't override required defaults
func TestHBaseGlobalCommonLabelsDoesNotOverrideDefaults(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "hbase", "values/global-common-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have multiple resources
	assert.Greater(t, len(resources.Statefulsets), 0, "Should have StatefulSets")
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")

	// Test all StatefulSets still have default component labels
	for _, statefulset := range resources.Statefulsets {
		// Even with custom global labels, default chart labels should still be present and correct
		assert.Contains(t, statefulset.Labels, "app.kubernetes.io/component", "StatefulSet %s should have app.kubernetes.io/component label", statefulset.Name)

		// Check Pod template as well
		podLabels := statefulset.Spec.Template.Labels
		assert.Contains(t, podLabels, "app.kubernetes.io/component", "Pod template in %s should have app.kubernetes.io/component label", statefulset.Name)
	}

	// Test all Deployments still have default component labels
	for _, deployment := range resources.Deployments {
		// Even with custom global labels, default chart labels should still be present and correct
		assert.Contains(t, deployment.Labels, "app.kubernetes.io/component", "Deployment %s should have app.kubernetes.io/component label", deployment.Name)

		// Check Pod template as well
		podLabels := deployment.Spec.Template.Labels
		assert.Contains(t, podLabels, "app.kubernetes.io/component", "Pod template in %s should have app.kubernetes.io/component label", deployment.Name)
	}
}
