package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
)

// TestVictoriaMetricsSingleGlobalCommonLabels tests that global.commonLabels are applied to StatefulSet and Pod templates
func TestVictoriaMetricsSingleGlobalCommonLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/global-common-labels.yaml")
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

// TestVictoriaMetricsSingleExtraLabels tests that server.extraLabels are applied to StatefulSet
func TestVictoriaMetricsSingleExtraLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/extra-labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Check StatefulSet labels
	assert.Contains(t, statefulset.Labels, "extra.label1", "StatefulSet should have extra.label1")
	assert.Equal(t, "extra-value1", statefulset.Labels["extra.label1"], "StatefulSet extra.label1 should have correct value")
	assert.Contains(t, statefulset.Labels, "extra.label2", "StatefulSet should have extra.label2")
	assert.Equal(t, "extra-value2", statefulset.Labels["extra.label2"], "StatefulSet extra.label2 should have correct value")
}

// TestVictoriaMetricsSinglePodLabels tests that server.podLabels are applied only to Pod templates, not StatefulSet
func TestVictoriaMetricsSinglePodLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/pod-labels.yaml")
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

// TestVictoriaMetricsSingleAllLabelsCombined tests the precedence and combination of all label types
func TestVictoriaMetricsSingleAllLabelsCombined(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/all-labels-combined.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Check StatefulSet labels (should have global + extra, but NOT pod labels)
	// Should have unique global label
	assert.Contains(t, statefulset.Labels, "global-unique", "StatefulSet should have global-unique")
	assert.Equal(t, "global-value", statefulset.Labels["global-unique"], "StatefulSet global-unique should have correct value")

	// Should have unique extra label
	assert.Contains(t, statefulset.Labels, "extra-unique", "StatefulSet should have extra-unique")
	assert.Equal(t, "extra-value", statefulset.Labels["extra-unique"], "StatefulSet extra-unique should have correct value")

	// Should NOT have pod-only labels
	assert.NotContains(t, statefulset.Labels, "pod-unique", "StatefulSet should NOT have pod-unique")

	// Test override at extra level: extraLabels should override global
	assert.Contains(t, statefulset.Labels, "override-at-extra", "StatefulSet should have override-at-extra")
	assert.Equal(t, "from-extra", statefulset.Labels["override-at-extra"], "StatefulSet override-at-extra should be from extraLabels (overriding global)")

	// Test override that will be overridden by pod: should still be from extra at StatefulSet level
	assert.Contains(t, statefulset.Labels, "override-at-pod", "StatefulSet should have override-at-pod")
	assert.Equal(t, "from-extra", statefulset.Labels["override-at-pod"], "StatefulSet override-at-pod should be from extraLabels (overriding global)")

	// Check Pod template labels (should have global + extra + pod labels)
	podLabels := statefulset.Spec.Template.Labels

	// Should have unique global label
	assert.Contains(t, podLabels, "global-unique", "Pod template should have global-unique")
	assert.Equal(t, "global-value", podLabels["global-unique"], "Pod template global-unique should have correct value")

	// Should NOT have extra-unique (extraLabels are not inherited by pods)
	assert.NotContains(t, podLabels, "extra-unique", "Pod template should NOT have extra-unique (extraLabels not inherited)")

	// Should have unique pod label
	assert.Contains(t, podLabels, "pod-unique", "Pod template should have pod-unique")
	assert.Equal(t, "pod-value", podLabels["pod-unique"], "Pod template pod-unique should have correct value")

	// Test override at extra level: should be from global in pod template (since extraLabels don't apply to pods)
	assert.Contains(t, podLabels, "override-at-extra", "Pod template should have override-at-extra")
	assert.Equal(t, "from-global", podLabels["override-at-extra"], "Pod template override-at-extra should be from global (extraLabels not inherited)")

	// Test override at pod level: podLabels should have highest precedence
	assert.Contains(t, podLabels, "override-at-pod", "Pod template should have override-at-pod")
	assert.Equal(t, "from-pod", podLabels["override-at-pod"], "Pod template override-at-pod should be from podLabels (highest precedence)")
}

// TestVictoriaMetricsSingleDefaultLabels tests that default chart labels are always present
func TestVictoriaMetricsSingleDefaultLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Check that default Helm chart labels are present on StatefulSet
	assert.Contains(t, statefulset.Labels, "app", "StatefulSet should have app label")
	assert.Equal(t, "server", statefulset.Labels["app"], "StatefulSet app should be 'server'")
	assert.Contains(t, statefulset.Labels, "app.kubernetes.io/name", "StatefulSet should have app.kubernetes.io/name label")
	assert.Equal(t, "victoria-metrics-single", statefulset.Labels["app.kubernetes.io/name"], "StatefulSet app.kubernetes.io/name should be chart name")
	assert.Contains(t, statefulset.Labels, "app.kubernetes.io/instance", "StatefulSet should have app.kubernetes.io/instance label")
	assert.Equal(t, releaseName, statefulset.Labels["app.kubernetes.io/instance"], "StatefulSet app.kubernetes.io/instance should be release name")

	// Check that default labels are present on Pod template
	podLabels := statefulset.Spec.Template.Labels
	assert.Contains(t, podLabels, "app", "Pod template should have app label")
	assert.Equal(t, "server", podLabels["app"], "Pod template app should be 'server'")
	assert.Contains(t, podLabels, "app.kubernetes.io/name", "Pod template should have app.kubernetes.io/name label")
	assert.Equal(t, "victoria-metrics-single", podLabels["app.kubernetes.io/name"], "Pod template app.kubernetes.io/name should be chart name")
	assert.Contains(t, podLabels, "app.kubernetes.io/instance", "Pod template should have app.kubernetes.io/instance label")
	assert.Equal(t, releaseName, podLabels["app.kubernetes.io/instance"], "Pod template app.kubernetes.io/instance should be release name")
}

// TestVictoriaMetricsSingleLabelPrecedenceDoesNotOverrideDefaults tests that custom labels don't override required defaults
func TestVictoriaMetricsSingleLabelPrecedenceDoesNotOverrideDefaults(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/all-labels-combined.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Even with custom labels, default chart labels should still be present and correct
	assert.Contains(t, statefulset.Labels, "app", "StatefulSet should have app label")
	assert.Equal(t, "server", statefulset.Labels["app"], "app should be 'server'")
	assert.Contains(t, statefulset.Labels, "app.kubernetes.io/instance", "StatefulSet should have app.kubernetes.io/instance label")
	assert.Equal(t, releaseName, statefulset.Labels["app.kubernetes.io/instance"], "app.kubernetes.io/instance should be release name")

	// Check Pod template as well
	podLabels := statefulset.Spec.Template.Labels
	assert.Contains(t, podLabels, "app", "Pod template should have app label")
	assert.Equal(t, "server", podLabels["app"], "Pod template app should be 'server'")
	assert.Contains(t, podLabels, "app.kubernetes.io/instance", "Pod template should have app.kubernetes.io/instance label")
	assert.Equal(t, releaseName, podLabels["app.kubernetes.io/instance"], "Pod template app.kubernetes.io/instance should be release name")
}
