package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
)

// TestKafkaupOperatorGlobalCommonLabels tests that global.commonLabels are applied to Deployment and Pod templates
func TestKafkaupOperatorGlobalCommonLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafkaup-operator", "values/global-common-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, d := range resources.Deployments {
		deployment = d
		break
	}

	// Check Deployment labels
	assert.Contains(t, deployment.Labels, "global.label1", "Deployment should have global.label1")
	assert.Equal(t, "global-value1", deployment.Labels["global.label1"], "Deployment global.label1 should have correct value")
	assert.Contains(t, deployment.Labels, "global.label2", "Deployment should have global.label2")
	assert.Equal(t, "global-value2", deployment.Labels["global.label2"], "Deployment global.label2 should have correct value")

	// Check Pod template labels
	podLabels := deployment.Spec.Template.Labels
	assert.Contains(t, podLabels, "global.label1", "Pod template should have global.label1")
	assert.Equal(t, "global-value1", podLabels["global.label1"], "Pod template global.label1 should have correct value")
	assert.Contains(t, podLabels, "global.label2", "Pod template should have global.label2")
	assert.Equal(t, "global-value2", podLabels["global.label2"], "Pod template global.label2 should have correct value")
}

// TestKafkaupOperatorCommonLabels tests that commonLabels are applied to Deployment and Pod templates
func TestKafkaupOperatorCommonLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafkaup-operator", "values/common-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, d := range resources.Deployments {
		deployment = d
		break
	}

	// Check Deployment labels
	assert.Contains(t, deployment.Labels, "common.label1", "Deployment should have common.label1")
	assert.Equal(t, "common-value1", deployment.Labels["common.label1"], "Deployment common.label1 should have correct value")
	assert.Contains(t, deployment.Labels, "common.label2", "Deployment should have common.label2")
	assert.Equal(t, "common-value2", deployment.Labels["common.label2"], "Deployment common.label2 should have correct value")

	// Check Pod template labels
	podLabels := deployment.Spec.Template.Labels
	assert.Contains(t, podLabels, "common.label1", "Pod template should have common.label1")
	assert.Equal(t, "common-value1", podLabels["common.label1"], "Pod template common.label1 should have correct value")
	assert.Contains(t, podLabels, "common.label2", "Pod template should have common.label2")
	assert.Equal(t, "common-value2", podLabels["common.label2"], "Pod template common.label2 should have correct value")
}

// TestKafkaupOperatorLabelPrecedence tests that commonLabels take precedence over global.commonLabels
func TestKafkaupOperatorLabelPrecedence(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafkaup-operator", "values/precedence-test.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, d := range resources.Deployments {
		deployment = d
		break
	}

	// Check Deployment labels
	// Should have unique global label
	assert.Contains(t, deployment.Labels, "global-unique", "Deployment should have global-unique")
	assert.Equal(t, "global-value", deployment.Labels["global-unique"], "Deployment global-unique should have correct value")

	// Should have unique common label
	assert.Contains(t, deployment.Labels, "common-unique", "Deployment should have common-unique")
	assert.Equal(t, "common-value", deployment.Labels["common-unique"], "Deployment common-unique should have correct value")

	// Test override: commonLabels should override global
	assert.Contains(t, deployment.Labels, "override-test", "Deployment should have override-test")
	assert.Equal(t, "from-common", deployment.Labels["override-test"], "Deployment override-test should be from commonLabels (overriding global)")

	// Check Pod template labels
	podLabels := deployment.Spec.Template.Labels

	// Should have unique global label
	assert.Contains(t, podLabels, "global-unique", "Pod template should have global-unique")
	assert.Equal(t, "global-value", podLabels["global-unique"], "Pod template global-unique should have correct value")

	// Should have unique common label
	assert.Contains(t, podLabels, "common-unique", "Pod template should have common-unique")
	assert.Equal(t, "common-value", podLabels["common-unique"], "Pod template common-unique should have correct value")

	// Test override: commonLabels should override global
	assert.Contains(t, podLabels, "override-test", "Pod template should have override-test")
	assert.Equal(t, "from-common", podLabels["override-test"], "Pod template override-test should be from commonLabels (overriding global)")
}

// TestKafkaupOperatorDefaultLabels tests that default chart labels are always present
func TestKafkaupOperatorDefaultLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafkaup-operator", "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, d := range resources.Deployments {
		deployment = d
		break
	}

	// Check that default Helm chart labels are present on Deployment
	assert.Contains(t, deployment.Labels, "app.kubernetes.io/name", "Deployment should have app.kubernetes.io/name label")
	assert.Contains(t, deployment.Labels, "app.kubernetes.io/instance", "Deployment should have app.kubernetes.io/instance label")
	assert.Contains(t, deployment.Labels, "app.kubernetes.io/component", "Deployment should have app.kubernetes.io/component label")

	// Check that default Helm chart labels are present on Pod template
	podLabels := deployment.Spec.Template.Labels
	assert.Contains(t, podLabels, "app", "Pod template should have app label")
}

// TestKafkaupOperatorLabelPrecedenceDoesNotOverrideDefaults tests that custom labels don't override required defaults
func TestKafkaupOperatorLabelPrecedenceDoesNotOverrideDefaults(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafkaup-operator", "values/precedence-test.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, d := range resources.Deployments {
		deployment = d
		break
	}

	// Even with custom labels, default chart labels should still be present and correct
	assert.Contains(t, deployment.Labels, "app.kubernetes.io/name", "Deployment should have app.kubernetes.io/name label")
	assert.Equal(t, "kafkaup-operator", deployment.Labels["app.kubernetes.io/name"], "app.kubernetes.io/name should be 'kafkaup-operator'")

	assert.Contains(t, deployment.Labels, "app.kubernetes.io/instance", "Deployment should have app.kubernetes.io/instance label")
	assert.Equal(t, "kafkaup-operator", deployment.Labels["app.kubernetes.io/instance"], "app.kubernetes.io/instance should be 'kafkaup-operator' (release name)")

	// Check Pod template as well
	podLabels := deployment.Spec.Template.Labels
	assert.Contains(t, podLabels, "app", "Pod template should have app label")
}
