package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
)

// TestMinioGlobalCommonLabelsStandalone tests that global.commonLabels are applied to Deployment and Pod templates in standalone mode
func TestMinioGlobalCommonLabelsStandalone(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/global-common-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment in standalone mode")
	assert.Len(t, resources.Statefulsets, 0, "Should have no StatefulSets in standalone mode")

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

// TestMinioGlobalCommonLabelsDistributed tests that global.commonLabels are applied to StatefulSet and Pod templates in distributed mode
func TestMinioGlobalCommonLabelsDistributed(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/distributed-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet in distributed mode")
	assert.Len(t, resources.Deployments, 0, "Should have no Deployments in distributed mode")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Check StatefulSet labels
	assert.Contains(t, statefulset.Labels, "mode.label1", "StatefulSet should have mode.label1")
	assert.Equal(t, "distributed-value1", statefulset.Labels["mode.label1"], "StatefulSet mode.label1 should have correct value")
	assert.Contains(t, statefulset.Labels, "mode.label2", "StatefulSet should have mode.label2")
	assert.Equal(t, "distributed-value2", statefulset.Labels["mode.label2"], "StatefulSet mode.label2 should have correct value")

	// Check Pod template labels
	podLabels := statefulset.Spec.Template.Labels
	assert.Contains(t, podLabels, "mode.label1", "Pod template should have mode.label1")
	assert.Equal(t, "distributed-value1", podLabels["mode.label1"], "Pod template mode.label1 should have correct value")
	assert.Contains(t, podLabels, "mode.label2", "Pod template should have mode.label2")
	assert.Equal(t, "distributed-value2", podLabels["mode.label2"], "Pod template mode.label2 should have correct value")
}

// TestMinioAdditionalLabels tests that additionalLabels are applied to Deployment and Pod templates
func TestMinioAdditionalLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/additional-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, d := range resources.Deployments {
		deployment = d
		break
	}

	// Check Deployment labels
	assert.Contains(t, deployment.Labels, "additional.label1", "Deployment should have additional.label1")
	assert.Equal(t, "additional-value1", deployment.Labels["additional.label1"], "Deployment additional.label1 should have correct value")
	assert.Contains(t, deployment.Labels, "additional.label2", "Deployment should have additional.label2")
	assert.Equal(t, "additional-value2", deployment.Labels["additional.label2"], "Deployment additional.label2 should have correct value")

	// Check Pod template labels
	podLabels := deployment.Spec.Template.Labels
	assert.Contains(t, podLabels, "additional.label1", "Pod template should have additional.label1")
	assert.Equal(t, "additional-value1", podLabels["additional.label1"], "Pod template additional.label1 should have correct value")
	assert.Contains(t, podLabels, "additional.label2", "Pod template should have additional.label2")
	assert.Equal(t, "additional-value2", podLabels["additional.label2"], "Pod template additional.label2 should have correct value")
}

// TestMinioPodLabels tests that podLabels are applied only to Pod templates, not Deployment
func TestMinioPodLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/pod-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, d := range resources.Deployments {
		deployment = d
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

// TestMinioAllLabelsCombined tests the precedence and combination of all label types
func TestMinioAllLabelsCombined(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/all-labels-combined.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, d := range resources.Deployments {
		deployment = d
		break
	}

	// Check Deployment labels (should have global + additional, but NOT pod labels)
	// Should have unique global label
	assert.Contains(t, deployment.Labels, "global-unique", "Deployment should have global-unique")
	assert.Equal(t, "global-value", deployment.Labels["global-unique"], "Deployment global-unique should have correct value")

	// Should have unique additional label
	assert.Contains(t, deployment.Labels, "additional-unique", "Deployment should have additional-unique")
	assert.Equal(t, "additional-value", deployment.Labels["additional-unique"], "Deployment additional-unique should have correct value")

	// Should NOT have pod-only labels
	assert.NotContains(t, deployment.Labels, "pod-unique", "Deployment should NOT have pod-unique")

	// Test override at additional level: additionalLabels should override global
	assert.Contains(t, deployment.Labels, "override-at-additional", "Deployment should have override-at-additional")
	assert.Equal(t, "from-additional", deployment.Labels["override-at-additional"], "Deployment override-at-additional should be from additionalLabels (overriding global)")

	// Test override that will be overridden by pod: should still be from additional at Deployment level
	assert.Contains(t, deployment.Labels, "override-at-pod", "Deployment should have override-at-pod")
	assert.Equal(t, "from-additional", deployment.Labels["override-at-pod"], "Deployment override-at-pod should be from additionalLabels (overriding global)")

	// Check Pod template labels (should have global + additional + pod labels)
	podLabels := deployment.Spec.Template.Labels

	// Should have unique global label
	assert.Contains(t, podLabels, "global-unique", "Pod template should have global-unique")
	assert.Equal(t, "global-value", podLabels["global-unique"], "Pod template global-unique should have correct value")

	// Should have unique additional label
	assert.Contains(t, podLabels, "additional-unique", "Pod template should have additional-unique")
	assert.Equal(t, "additional-value", podLabels["additional-unique"], "Pod template additional-unique should have correct value")

	// Should have unique pod label
	assert.Contains(t, podLabels, "pod-unique", "Pod template should have pod-unique")
	assert.Equal(t, "pod-value", podLabels["pod-unique"], "Pod template pod-unique should have correct value")

	// Test override at additional level: additionalLabels should override global
	assert.Contains(t, podLabels, "override-at-additional", "Pod template should have override-at-additional")
	assert.Equal(t, "from-additional", podLabels["override-at-additional"], "Pod template override-at-additional should be from additionalLabels (overriding global)")

	// Test override at pod level: podLabels should have highest precedence
	assert.Contains(t, podLabels, "override-at-pod", "Pod template should have override-at-pod")
	assert.Equal(t, "from-pod", podLabels["override-at-pod"], "Pod template override-at-pod should be from podLabels (highest precedence)")
}

// TestMinioJobLabels tests that global.commonLabels are applied to all Job resources and Pod templates
func TestMinioJobLabels(t *testing.T) {
	// Enable both bucket creation and ServiceMonitor to ensure all jobs are created
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/global-common-labels-with-jobs.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have both jobs when both features are enabled
	expectedJobs := map[string]string{
		releaseName + "-make-bucket-job":          "make-bucket job",
		releaseName + "-update-prometheus-secret": "prometheus-secret job",
	}

	assert.Len(t, resources.Jobs, len(expectedJobs), "Should have exactly %d Jobs when both features are enabled", len(expectedJobs))

	// Check labels for each job
	for expectedName, jobDescription := range expectedJobs {
		job, exists := resources.Jobs[expectedName]
		assert.True(t, exists, "Should have %s", jobDescription)

		// Check Job metadata labels
		assert.Contains(t, job.Labels, "global.label1", "%s should have global.label1", jobDescription)
		assert.Equal(t, "global-value1", job.Labels["global.label1"], "%s global.label1 should have correct value", jobDescription)
		assert.Contains(t, job.Labels, "global.label2", "%s should have global.label2", jobDescription)
		assert.Equal(t, "global-value2", job.Labels["global.label2"], "%s global.label2 should have correct value", jobDescription)

		// Check Pod template labels
		podLabels := job.Spec.Template.Labels
		assert.Contains(t, podLabels, "global.label1", "%s Pod template should have global.label1", jobDescription)
		assert.Equal(t, "global-value1", podLabels["global.label1"], "%s Pod template global.label1 should have correct value", jobDescription)
		assert.Contains(t, podLabels, "global.label2", "%s Pod template should have global.label2", jobDescription)
		assert.Equal(t, "global-value2", podLabels["global.label2"], "%s Pod template global.label2 should have correct value", jobDescription)
	}
}

// TestMinioDefaultLabels tests that default chart labels are always present
func TestMinioDefaultLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, d := range resources.Deployments {
		deployment = d
		break
	}

	// Check that default Helm chart labels are present on Deployment
	assert.Contains(t, deployment.Labels, "app", "Deployment should have app label")
	assert.Equal(t, "minio", deployment.Labels["app"], "Deployment app should be 'minio'")
	assert.Contains(t, deployment.Labels, "release", "Deployment should have release label")
	assert.Equal(t, releaseName, deployment.Labels["release"], "Deployment release should be release name")
	assert.Contains(t, deployment.Labels, "chart", "Deployment should have chart label")
	assert.Contains(t, deployment.Labels, "heritage", "Deployment should have heritage label")

	// Check that default labels are present on Pod template
	podLabels := deployment.Spec.Template.Labels
	assert.Contains(t, podLabels, "app", "Pod template should have app label")
	assert.Equal(t, "minio", podLabels["app"], "Pod template app should be 'minio'")
	assert.Contains(t, podLabels, "release", "Pod template should have release label")
	assert.Equal(t, releaseName, podLabels["release"], "Pod template release should be release name")
}

// TestMinioLabelPrecedenceDoesNotOverrideDefaults tests that custom labels don't override required defaults
func TestMinioLabelPrecedenceDoesNotOverrideDefaults(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/all-labels-combined.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")

	var deployment appsv1.Deployment
	for _, d := range resources.Deployments {
		deployment = d
		break
	}

	// Even with custom labels, default chart labels should still be present and correct
	assert.Contains(t, deployment.Labels, "app", "Deployment should have app label")
	assert.Equal(t, "minio", deployment.Labels["app"], "app should be 'minio'")
	assert.Contains(t, deployment.Labels, "release", "Deployment should have release label")
	assert.Equal(t, releaseName, deployment.Labels["release"], "release should be release name")

	// Check Pod template as well
	podLabels := deployment.Spec.Template.Labels
	assert.Contains(t, podLabels, "app", "Pod template should have app label")
	assert.Equal(t, "minio", podLabels["app"], "Pod template app should be 'minio'")
	assert.Contains(t, podLabels, "release", "Pod template should have release label")
	assert.Equal(t, releaseName, podLabels["release"], "Pod template release should be release name")
}
