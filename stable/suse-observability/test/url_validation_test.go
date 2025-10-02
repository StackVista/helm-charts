package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestUrlValidationWithValidHttpUrl(t *testing.T) {
	// Test that valid HTTP URLs pass validation
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/url_valid_http.yaml")
	helmtestutil.NewKubernetesResources(t, output)
}

func TestUrlValidationWithValidHttpsUrl(t *testing.T) {
	// Test that valid HTTPS URLs pass validation
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/url_valid_https.yaml")
	helmtestutil.NewKubernetesResources(t, output)
}

func TestUrlValidationMissingScheme(t *testing.T) {
	// Test that URLs without a scheme (e.g., "localhost") fail validation
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/url_invalid_no_scheme.yaml")
	require.Contains(t, err.Error(), "Invalid URL format")
	require.Contains(t, err.Error(), "The URL must include a scheme")
}

func TestUrlValidationEmptyUrl(t *testing.T) {
	// Test that empty/missing baseUrl fails validation with required error
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/url_invalid_empty.yaml")
	require.Contains(t, err.Error(), "Please provide your SUSE Observability base URL")
}
