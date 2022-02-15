package test

import (
	"testing"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHelmBasicRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "cluster-agent", "values/minimal.yaml")

	// Parse all resources into their corresponding types for validation and further inspection
	helmtestutil.NewKubernetesResources(t, output)
}
