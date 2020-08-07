package helmtestutil

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
)

// RenderHelmTemplate renders a helm template assuming it lives in the parent directory while
// only specifying values files for helm options
func RenderHelmTemplate(t *testing.T, releaseName string, valuesFiles ...string) string {
	helmOpts := &helm.Options{
		ValuesFiles: valuesFiles,
	}

	return RenderHelmTemplateOpts(t, releaseName, helmOpts)
}

// RenderHelmTemplateOpts renders a helm template assuming it lives in the parent directory
func RenderHelmTemplateOpts(t *testing.T, releaseName string, helmOpts *helm.Options) string {
	helmChartPath, pathErr := filepath.Abs("..")
	require.NoError(t, pathErr)

	output, err := helm.RenderTemplateE(t, helmOpts, helmChartPath, releaseName, []string{})
	require.NoError(t, err)

	return output
}
