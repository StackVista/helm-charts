package test

import (
	"testing"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHelmBasicRender(t *testing.T) {
	t.Parallel()

	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")

	// Parse all resources into their corresponding types for validation and further inspection
	helmtestutil.NewKubernetesResources(t, output)
}
