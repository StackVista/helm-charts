package test

import (
	"testing"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHelmBasicRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml")

	// Parse all resources into their corresponding types for validation and further inspection
	helmtestutil.NewKubernetesResources(t, output)
}
