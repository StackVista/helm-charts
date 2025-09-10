package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestElasticsearchBasicTemplate(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "elasticsearch", "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have a StatefulSet
	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")
}
