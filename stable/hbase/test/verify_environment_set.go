package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHelmBasicRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml")

	// Parse all resources into their corresponding types for validation and further inspection
	resources := helmtestutil.NewKubernetesResources(t, output)

	for _, statefulset := range resources.Statefulsets {
		for _, container := range statefulset.Spec.Template.Spec.Containers {
			assert.NotEmpty(t, container.Env)
		}
	}
}
