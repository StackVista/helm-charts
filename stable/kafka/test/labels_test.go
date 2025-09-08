package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
)

// TestKafkaGlobalCommonLabels tests that global.commonLabels are applied to StatefulSet and Pod templates
func TestKafkaGlobalCommonLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafka", "values/global-common-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Check StatefulSet labels
	assert.Contains(t, statefulset.Labels, "global.label1", "StatefulSet should have global.label1")
	assert.Equal(t, "global-value1", statefulset.Labels["global.label1"], "StatefulSet global.label1 should have correct value")
	assert.Contains(t, statefulset.Labels, "global.label2", "StatefulSet should have global.label2")
	assert.Equal(t, "global-value2", statefulset.Labels["global.label2"], "StatefulSet global.label2 should have correct value")

	// Check Pod template labels
	podLabels := statefulset.Spec.Template.Labels
	assert.Contains(t, podLabels, "global.label1", "Pod template should have global.label1")
	assert.Equal(t, "global-value1", podLabels["global.label1"], "Pod template global.label1 should have correct value")
	assert.Contains(t, podLabels, "global.label2", "Pod template should have global.label2")
	assert.Equal(t, "global-value2", podLabels["global.label2"], "Pod template global.label2 should have correct value")
}

// TestKafkaCommonLabels tests that commonLabels are applied to StatefulSet and Pod templates
func TestKafkaCommonLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafka", "values/common-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Check StatefulSet labels
	assert.Contains(t, statefulset.Labels, "common.label1", "StatefulSet should have common.label1")
	assert.Equal(t, "common-value1", statefulset.Labels["common.label1"], "StatefulSet common.label1 should have correct value")
	assert.Contains(t, statefulset.Labels, "common.label2", "StatefulSet should have common.label2")
	assert.Equal(t, "common-value2", statefulset.Labels["common.label2"], "StatefulSet common.label2 should have correct value")

	// Check Pod template labels
	podLabels := statefulset.Spec.Template.Labels
	assert.Contains(t, podLabels, "common.label1", "Pod template should have common.label1")
	assert.Equal(t, "common-value1", podLabels["common.label1"], "Pod template common.label1 should have correct value")
	assert.Contains(t, podLabels, "common.label2", "Pod template should have common.label2")
	assert.Equal(t, "common-value2", podLabels["common.label2"], "Pod template common.label2 should have correct value")
}

// TestKafkaPodLabels tests that podLabels are applied only to Pod templates, not StatefulSet
func TestKafkaPodLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafka", "values/pod-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Check StatefulSet labels - podLabels should NOT be applied to StatefulSet
	assert.NotContains(t, statefulset.Labels, "pod.label1", "StatefulSet should NOT have pod.label1")
	assert.NotContains(t, statefulset.Labels, "pod.label2", "StatefulSet should NOT have pod.label2")

	// Check Pod template labels - podLabels should be applied to Pod templates
	podLabels := statefulset.Spec.Template.Labels
	assert.Contains(t, podLabels, "pod.label1", "Pod template should have pod.label1")
	assert.Equal(t, "pod-value1", podLabels["pod.label1"], "Pod template pod.label1 should have correct value")
	assert.Contains(t, podLabels, "pod.label2", "Pod template should have pod.label2")
	assert.Equal(t, "pod-value2", podLabels["pod.label2"], "Pod template pod.label2 should have correct value")
}

// TestKafkaAllLabelsCombined tests the precedence and combination of all label types
func TestKafkaAllLabelsCombined(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafka", "values/all-labels-combined.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Check StatefulSet labels (should have global + common, but NOT pod labels)
	// Should have unique global label
	assert.Contains(t, statefulset.Labels, "global-unique", "StatefulSet should have global-unique")
	assert.Equal(t, "global-value", statefulset.Labels["global-unique"], "StatefulSet global-unique should have correct value")

	// Should have unique common label
	assert.Contains(t, statefulset.Labels, "common-unique", "StatefulSet should have common-unique")
	assert.Equal(t, "common-value", statefulset.Labels["common-unique"], "StatefulSet common-unique should have correct value")

	// Should NOT have pod-only labels
	assert.NotContains(t, statefulset.Labels, "pod-unique", "StatefulSet should NOT have pod-unique")

	// Test override at common level: commonLabels should override global
	assert.Contains(t, statefulset.Labels, "override-at-common", "StatefulSet should have override-at-common")
	assert.Equal(t, "from-common", statefulset.Labels["override-at-common"], "StatefulSet override-at-common should be from commonLabels (overriding global)")

	// Test override that will be overridden by pod: should still be from common at StatefulSet level
	assert.Contains(t, statefulset.Labels, "override-at-pod", "StatefulSet should have override-at-pod")
	assert.Equal(t, "from-common", statefulset.Labels["override-at-pod"], "StatefulSet override-at-pod should be from commonLabels (overriding global)")

	// Check Pod template labels (should have global + common + pod labels)
	podLabels := statefulset.Spec.Template.Labels

	// Should have unique global label
	assert.Contains(t, podLabels, "global-unique", "Pod template should have global-unique")
	assert.Equal(t, "global-value", podLabels["global-unique"], "Pod template global-unique should have correct value")

	// Should have unique common label
	assert.Contains(t, podLabels, "common-unique", "Pod template should have common-unique")
	assert.Equal(t, "common-value", podLabels["common-unique"], "Pod template common-unique should have correct value")

	// Should have unique pod label
	assert.Contains(t, podLabels, "pod-unique", "Pod template should have pod-unique")
	assert.Equal(t, "pod-value", podLabels["pod-unique"], "Pod template pod-unique should have correct value")

	// Test override at common level: commonLabels should override global
	assert.Contains(t, podLabels, "override-at-common", "Pod template should have override-at-common")
	assert.Equal(t, "from-common", podLabels["override-at-common"], "Pod template override-at-common should be from commonLabels (overriding global)")

	// Test override at pod level: podLabels should have highest precedence
	assert.Contains(t, podLabels, "override-at-pod", "Pod template should have override-at-pod")
	assert.Equal(t, "from-pod", podLabels["override-at-pod"], "Pod template override-at-pod should be from podLabels (highest precedence)")
}

// TestKafkaDefaultLabels tests that default chart labels are always present
func TestKafkaDefaultLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafka", "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Check that default Helm chart labels are present on StatefulSet
	assert.Contains(t, statefulset.Labels, "app.kubernetes.io/name", "StatefulSet should have app.kubernetes.io/name label")
	assert.Contains(t, statefulset.Labels, "app.kubernetes.io/instance", "StatefulSet should have app.kubernetes.io/instance label")
	assert.Contains(t, statefulset.Labels, "app.kubernetes.io/component", "StatefulSet should have app.kubernetes.io/component label")

	// Check that default Helm chart labels are present on Pod template
	podLabels := statefulset.Spec.Template.Labels
	assert.Contains(t, podLabels, "app.kubernetes.io/name", "Pod template should have app.kubernetes.io/name label")
	assert.Contains(t, podLabels, "app.kubernetes.io/instance", "Pod template should have app.kubernetes.io/instance label")
	assert.Contains(t, podLabels, "app.kubernetes.io/component", "Pod template should have app.kubernetes.io/component label")
}

// TestKafkaLabelPrecedenceDoesNotOverrideDefaults tests that custom labels don't override required defaults
func TestKafkaLabelPrecedenceDoesNotOverrideDefaults(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "kafka", "values/all-labels-combined.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Even with custom labels, default chart labels should still be present and correct
	assert.Contains(t, statefulset.Labels, "app.kubernetes.io/name", "StatefulSet should have app.kubernetes.io/name label")
	assert.Equal(t, "kafka", statefulset.Labels["app.kubernetes.io/name"], "app.kubernetes.io/name should be 'kafka'")

	assert.Contains(t, statefulset.Labels, "app.kubernetes.io/instance", "StatefulSet should have app.kubernetes.io/instance label")
	assert.Equal(t, "kafka", statefulset.Labels["app.kubernetes.io/instance"], "app.kubernetes.io/instance should be 'kafka' (release name)")

	// Check Pod template as well
	podLabels := statefulset.Spec.Template.Labels
	assert.Contains(t, podLabels, "app.kubernetes.io/name", "Pod template should have app.kubernetes.io/name label")
	assert.Equal(t, "kafka", podLabels["app.kubernetes.io/name"], "Pod template app.kubernetes.io/name should be 'kafka'")
}
