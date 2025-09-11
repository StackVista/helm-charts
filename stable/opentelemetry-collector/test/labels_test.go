package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// TestOpenTelemetryCollectorGlobalCommonLabelsDeploymentMode tests that global.commonLabels are applied in Deployment mode
func TestOpenTelemetryCollectorGlobalCommonLabelsDeploymentMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "otel", "values/global-common-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have Deployments in deployment mode
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")

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

// TestOpenTelemetryCollectorGlobalCommonLabelsDaemonSetMode tests that global.commonLabels are applied in DaemonSet mode
func TestOpenTelemetryCollectorGlobalCommonLabelsDaemonSetMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "otel", "values/global-common-labels-daemonset.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have DaemonSets in daemonset mode
	assert.Greater(t, len(resources.DaemonSets), 0, "Should have DaemonSets")

	// Test DaemonSets have global common labels
	for _, daemonset := range resources.DaemonSets {
		// Check DaemonSet labels
		assert.Contains(t, daemonset.Labels, "global.label1", "DaemonSet %s should have global.label1", daemonset.Name)
		assert.Equal(t, "global-value1", daemonset.Labels["global.label1"], "DaemonSet %s global.label1 should have correct value", daemonset.Name)
		assert.Contains(t, daemonset.Labels, "global.label2", "DaemonSet %s should have global.label2", daemonset.Name)
		assert.Equal(t, "global-value2", daemonset.Labels["global.label2"], "DaemonSet %s global.label2 should have correct value", daemonset.Name)

		// Check Pod template labels
		podLabels := daemonset.Spec.Template.Labels
		assert.Contains(t, podLabels, "global.label1", "Pod template in %s should have global.label1", daemonset.Name)
		assert.Equal(t, "global-value1", podLabels["global.label1"], "Pod template in %s global.label1 should have correct value", daemonset.Name)
		assert.Contains(t, podLabels, "global.label2", "Pod template in %s should have global.label2", daemonset.Name)
		assert.Equal(t, "global-value2", podLabels["global.label2"], "Pod template in %s global.label2 should have correct value", daemonset.Name)
	}
}

// TestOpenTelemetryCollectorGlobalCommonLabelsStatefulSetMode tests that global.commonLabels are applied in StatefulSet mode
func TestOpenTelemetryCollectorGlobalCommonLabelsStatefulSetMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "otel", "values/global-common-labels-statefulset.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have StatefulSets in statefulset mode
	assert.Greater(t, len(resources.Statefulsets), 0, "Should have StatefulSets")

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
}

// TestOpenTelemetryCollectorDefaultLabels tests that default chart labels are always present
func TestOpenTelemetryCollectorDefaultLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "otel", "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have Deployments (default mode)
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")

	// Test all Deployments have default labels
	for _, deployment := range resources.Deployments {
		// Check that default component labels are present on Deployment
		assert.Contains(t, deployment.Labels, "app.kubernetes.io/name", "Deployment %s should have app.kubernetes.io/name label", deployment.Name)
		assert.Contains(t, deployment.Labels, "app.kubernetes.io/instance", "Deployment %s should have app.kubernetes.io/instance label", deployment.Name)

		// Check that default component labels are present on Pod template
		podLabels := deployment.Spec.Template.Labels
		assert.Contains(t, podLabels, "app.kubernetes.io/name", "Pod template in %s should have app.kubernetes.io/name label", deployment.Name)
		assert.Contains(t, podLabels, "app.kubernetes.io/instance", "Pod template in %s should have app.kubernetes.io/instance label", deployment.Name)
		assert.Contains(t, podLabels, "component", "Pod template in %s should have component label", deployment.Name)
	}
}

// TestOpenTelemetryCollectorGlobalCommonLabelsDoesNotOverrideDefaults tests that global common labels don't override required defaults
func TestOpenTelemetryCollectorGlobalCommonLabelsDoesNotOverrideDefaults(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "otel", "values/global-common-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have Deployments
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")

	// Test all Deployments still have default labels
	for _, deployment := range resources.Deployments {
		// Even with custom global labels, default chart labels should still be present and correct
		assert.Contains(t, deployment.Labels, "app.kubernetes.io/name", "Deployment %s should have app.kubernetes.io/name label", deployment.Name)
		assert.Equal(t, "opentelemetry-collector", deployment.Labels["app.kubernetes.io/name"], "Deployment %s app.kubernetes.io/name should have correct value", deployment.Name)
		assert.Contains(t, deployment.Labels, "app.kubernetes.io/instance", "Deployment %s should have app.kubernetes.io/instance label", deployment.Name)

		// Check Pod template as well
		podLabels := deployment.Spec.Template.Labels
		assert.Contains(t, podLabels, "app.kubernetes.io/name", "Pod template in %s should have app.kubernetes.io/name label", deployment.Name)
		assert.Equal(t, "opentelemetry-collector", podLabels["app.kubernetes.io/name"], "Pod template %s app.kubernetes.io/name should have correct value", deployment.Name)
		assert.Contains(t, podLabels, "app.kubernetes.io/instance", "Pod template in %s should have app.kubernetes.io/instance label", deployment.Name)
		assert.Contains(t, podLabels, "component", "Pod template in %s should have component label", deployment.Name)
		assert.Equal(t, "standalone-collector", podLabels["component"], "Pod template %s component should have correct value", deployment.Name)
	}
}

// TestOpenTelemetryCollectorGlobalCommonLabelsDeploymentPrecedence tests that local additionalLabels take precedence over global.commonLabels
func TestOpenTelemetryCollectorGlobalCommonLabelsDeploymentPrecedence(t *testing.T) {
	// Create a test with both global.commonLabels and additionalLabels that have overlapping keys
	output := helmtestutil.RenderHelmTemplate(t, "otel", "values/global-common-labels-precedence.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have Deployments
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")

	// Test that additionalLabels takes precedence over global.commonLabels
	for _, deployment := range resources.Deployments {
		// local additionalLabels should override global.commonLabels
		assert.Contains(t, deployment.Labels, "global.label1", "Deployment %s should have global.label1", deployment.Name)
		assert.Equal(t, "local-override", deployment.Labels["global.label1"], "Deployment %s global.label1 should be overridden by local value", deployment.Name)
		// global.label2 should still be from global.commonLabels since it's not overridden
		assert.Contains(t, deployment.Labels, "global.label2", "Deployment %s should have global.label2", deployment.Name)
		assert.Equal(t, "global-value2", deployment.Labels["global.label2"], "Deployment %s global.label2 should have global value", deployment.Name)

		// Check Pod template labels
		podLabels := deployment.Spec.Template.Labels
		assert.Contains(t, podLabels, "global.label1", "Pod template in %s should have global.label1", deployment.Name)
		assert.Equal(t, "local-override", podLabels["global.label1"], "Pod template in %s global.label1 should be overridden by local value", deployment.Name)
		assert.Contains(t, podLabels, "global.label2", "Pod template in %s should have global.label2", deployment.Name)
		assert.Equal(t, "global-value2", podLabels["global.label2"], "Pod template in %s global.label2 should have global value", deployment.Name)
	}
}
