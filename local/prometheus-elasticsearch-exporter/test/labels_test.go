package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
)

// TestPrometheusElasticsearchExporterGlobalCommonLabels tests that global.commonLabels are applied to Deployment and Pod templates
func TestPrometheusElasticsearchExporterGlobalCommonLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "prometheus-elasticsearch-exporter", "values/global-common-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, dep := range resources.Deployments {
		deployment = dep
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

// TestPrometheusElasticsearchExporterPodLabels tests that podLabels are applied only to Pod templates, not Deployment
func TestPrometheusElasticsearchExporterPodLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "prometheus-elasticsearch-exporter", "values/pod-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, dep := range resources.Deployments {
		deployment = dep
		break
	}

	// Check Deployment labels - podLabels should NOT be applied to Deployment
	assert.NotContains(t, deployment.Labels, "pod.label1", "Deployment should NOT have pod.label1")
	assert.NotContains(t, deployment.Labels, "pod.label2", "Deployment should NOT have pod.label2")

	// Check Pod template labels - podLabels should be applied to Pod templates
	podLabels := deployment.Spec.Template.Labels
	assert.Contains(t, podLabels, "pod.label1", "Pod template should have pod.label1")
	assert.Equal(t, "pod-value1", podLabels["pod.label1"], "Pod template pod.label1 should have correct value")
	assert.Contains(t, podLabels, "pod.label2", "Pod template should have pod.label2")
	assert.Equal(t, "pod-value2", podLabels["pod.label2"], "Pod template pod.label2 should have correct value")
}

// TestPrometheusElasticsearchExporterAllLabelsCombined tests the precedence and combination of all label types
func TestPrometheusElasticsearchExporterAllLabelsCombined(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "prometheus-elasticsearch-exporter", "values/all-labels-combined.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, dep := range resources.Deployments {
		deployment = dep
		break
	}

	// Check Deployment labels (should have global labels, but NOT pod labels)
	// Should have unique global label
	assert.Contains(t, deployment.Labels, "global-unique", "Deployment should have global-unique")
	assert.Equal(t, "global-value", deployment.Labels["global-unique"], "Deployment global-unique should have correct value")

	// Should NOT have pod-only labels
	assert.NotContains(t, deployment.Labels, "pod-unique", "Deployment should NOT have pod-unique")

	// Test override that will be overridden by pod: should still be from global at Deployment level
	assert.Contains(t, deployment.Labels, "override-at-pod", "Deployment should have override-at-pod")
	assert.Equal(t, "from-global", deployment.Labels["override-at-pod"], "Deployment override-at-pod should be from global")

	// Check Pod template labels (should have global + pod labels)
	podLabels := deployment.Spec.Template.Labels

	// Should have unique global label
	assert.Contains(t, podLabels, "global-unique", "Pod template should have global-unique")
	assert.Equal(t, "global-value", podLabels["global-unique"], "Pod template global-unique should have correct value")

	// Should have unique pod label
	assert.Contains(t, podLabels, "pod-unique", "Pod template should have pod-unique")
	assert.Equal(t, "pod-value", podLabels["pod-unique"], "Pod template pod-unique should have correct value")

	// Test override at pod level: podLabels should have highest precedence
	assert.Contains(t, podLabels, "override-at-pod", "Pod template should have override-at-pod")
	assert.Equal(t, "from-pod", podLabels["override-at-pod"], "Pod template override-at-pod should be from podLabels (highest precedence)")
}

// TestPrometheusElasticsearchExporterDefaultLabels tests that default chart labels are always present
func TestPrometheusElasticsearchExporterDefaultLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "prometheus-elasticsearch-exporter", "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, dep := range resources.Deployments {
		deployment = dep
		break
	}

	// Check that default Helm chart labels are present on Deployment
	assert.Contains(t, deployment.Labels, "app", "Deployment should have app label")
	assert.Contains(t, deployment.Labels, "release", "Deployment should have release label")
	assert.Contains(t, deployment.Labels, "chart", "Deployment should have chart label")

	// Check that default Helm chart labels are present on Pod template
	podLabels := deployment.Spec.Template.Labels
	assert.Contains(t, podLabels, "app", "Pod template should have app label")
	assert.Contains(t, podLabels, "release", "Pod template should have release label")
}

// TestPrometheusElasticsearchExporterLabelPrecedenceDoesNotOverrideDefaults tests that custom labels don't override required defaults
func TestPrometheusElasticsearchExporterLabelPrecedenceDoesNotOverrideDefaults(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "prometheus-elasticsearch-exporter", "values/all-labels-combined.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, dep := range resources.Deployments {
		deployment = dep
		break
	}

	// Even with custom labels, default chart labels should still be present and correct
	assert.Contains(t, deployment.Labels, "app", "Deployment should have app label")
	assert.Equal(t, "prometheus-elasticsearch-exporter", deployment.Labels["app"], "app should be 'prometheus-elasticsearch-exporter'")

	assert.Contains(t, deployment.Labels, "release", "Deployment should have release label")
	assert.Equal(t, "prometheus-elasticsearch-exporter", deployment.Labels["release"], "release should be 'prometheus-elasticsearch-exporter' (release name)")

	// Check Pod template as well
	podLabels := deployment.Spec.Template.Labels
	assert.Contains(t, podLabels, "app", "Pod template should have app label")
	assert.Equal(t, "prometheus-elasticsearch-exporter", podLabels["app"], "Pod template app should be 'prometheus-elasticsearch-exporter'")
}
