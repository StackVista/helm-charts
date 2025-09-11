package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

const releaseName = "victoria-metrics-single"

func TestVictoriaMetricsSingleBasicTemplate(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have StatefulSets
	assert.Greater(t, len(resources.Statefulsets), 0, "Should have StatefulSets")
}

func TestVictoriaMetricsSingleStatefulSetRendered(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have exactly one StatefulSet with the expected name
	expectedStatefulSets := []string{
		releaseName + "-server",
	}

	assert.Len(t, resources.Statefulsets, len(expectedStatefulSets), "Should have exactly %d StatefulSets", len(expectedStatefulSets))

	// Check each expected StatefulSet exists
	for _, expectedName := range expectedStatefulSets {
		_, exists := resources.Statefulsets[expectedName]
		assert.True(t, exists, "StatefulSet %s should exist", expectedName)
	}
}
