package test

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
)

func TestHelmBasicRender(t *testing.T) {
	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	helmOpts := &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
	}

	_, err = helm.RenderTemplateE(t, helmOpts, helmChartPath, "sample-render", []string{})
	require.NoError(t, err)

}
