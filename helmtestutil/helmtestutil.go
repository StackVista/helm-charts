package helmtestutil

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/require"
)

// RenderHelmTemplate renders a helm template assuming it lives in the parent directory while
// only specifying values files for helm options
func RenderHelmTemplate(t *testing.T, releaseName string, valuesFiles ...string) string {
	helmOpts := &helm.Options{
		ValuesFiles: valuesFiles,
	}

	return RenderHelmTemplateOptsNoError(t, releaseName, helmOpts)
}

// RenderHelmTemplateError renders a helm template assuming it lives in the parent directory while
// only specifying values files for helm options. It expects an error and returns that error for further inspection
func RenderHelmTemplateError(t *testing.T, releaseName string, valuesFiles ...string) error {
	helmOpts := &helm.Options{
		ValuesFiles: valuesFiles,
		Logger:      logger.Discard,
	}

	out, err := RenderHelmTemplateOpts(t, releaseName, helmOpts)
	fmt.Print(out)
	require.NotNil(t, err)

	return err
}

// RenderHelmTemplateOptsNoError renders a helm template assuming it lives in the parent directory, asserts that
// no error happened during rendering
func RenderHelmTemplateOptsNoError(t *testing.T, releaseName string, helmOpts *helm.Options) string {
	helmOpts.Logger = logger.Discard
	output, err := RenderHelmTemplateOpts(t, releaseName, helmOpts)
	require.NoError(t, err)

	return output
}

// RenderHelmTemplateOpts renders a helm template assuming it lives in the parent directory
func RenderHelmTemplateOpts(t *testing.T, releaseName string, helmOpts *helm.Options) (string, error) {
	if helmOpts.Logger == nil {
		helmOpts.Logger = logger.Discard
	}

	helmChartPath, pathErr := filepath.Abs("..")
	require.NoError(t, pathErr)

	output, err := helm.RenderTemplateE(t, helmOpts, helmChartPath, releaseName, []string{})

	return output, err
}
