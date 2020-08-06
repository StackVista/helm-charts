package test

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHelmBasicRender(t *testing.T) {
	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	helmOpts := &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
	}

	output, renderErr := helm.RenderTemplateE(t, helmOpts, helmChartPath, "sample-render", []string{})
	require.NoError(t, renderErr)

	// Parse all resources into their corresponding types for validation and further inspection
	helmtestutil.NewKubernetesResources(t, output)
}
