package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	v1 "k8s.io/api/core/v1"
)

// TestGlobalSuseObservabilityValidConfiguration tests that valid global.suseObservability configuration renders correctly
// This tests the auto-detection behavior - when any global.suseObservability value is set, the global mode is enabled
func TestGlobalSuseObservabilityValidConfiguration(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_suse_observability_valid.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Verify that the chart renders successfully with global.suseObservability auto-detection
	// This is the key test - proving the new mode works without errors

	// Verify license secret is created with value from global.suseObservability.license
	licenseSecret, exists := resources.Secrets["suse-observability-license"]
	require.True(t, exists, "License secret should be created")
	assert.NotEmpty(t, licenseSecret.Data["LICENSE_KEY"], "License key should be set")

	// Verify auth secret is created with value from global.suseObservability.adminPassword
	authSecret, exists := resources.Secrets["suse-observability-auth"]
	require.True(t, exists, "Auth secret should be created")
	assert.NotEmpty(t, authSecret.Data["default_password"], "Admin password should be set")

	// Verify pull secret is created with values from global.suseObservability.pullSecret
	// Using the comprehensive CheckPullSecret helper to validate registry, username, and password
	CheckPullSecret(t, resources, "suse-observability-pull-secret", "testuser", "testpassword", "my.registry.com")

	// Verify deployments exist and are using correct baseUrl from global.suseObservability.baseUrl
	apiDeployment, exists := resources.Deployments["suse-observability-api"]
	require.True(t, exists, "API deployment should exist")
	envVars := getEnvVars(apiDeployment.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "http://test.example.com", envVars["STACKSTATE_BASE_URL"], "Base URL should be set from global.suseObservability.baseUrl")
}

// TestGlobalSuseObservabilityMissingLicense tests that setting global values without license fails
// Auto-detection will enable global mode, but validation will fail
func TestGlobalSuseObservabilityMissingLicense(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/global_suse_observability_missing_license.yaml")
	require.Error(t, err, "Should fail when license is missing")
	require.Contains(t, err.Error(), "global.suseObservability.license is required", "Error should mention license is required")
}

// TestGlobalSuseObservabilityMissingBaseUrl tests that setting global values without baseUrl fails
// Auto-detection will enable global mode, but validation will fail
func TestGlobalSuseObservabilityMissingBaseUrl(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/global_suse_observability_missing_baseurl.yaml")
	require.Error(t, err, "Should fail when baseUrl is missing")
	require.Contains(t, err.Error(), "global.suseObservability.baseUrl is required", "Error should mention baseUrl is required")
}

// TestLegacyStackstateConfigBackwardsCompatibility tests that legacy stackstate.* configuration still works
// Uses the existing full.yaml which uses stackstate.* values without any global.suseObservability values
func TestLegacyStackstateConfigBackwardsCompatibility(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Verify secrets are created with values from stackstate.*
	licenseSecret, exists := resources.Secrets["suse-observability-license"]
	require.True(t, exists, "License secret should be created")
	assert.NotEmpty(t, licenseSecret.Data["LICENSE_KEY"], "License key should be set from stackstate.license.key")

	authSecret, exists := resources.Secrets["suse-observability-auth"]
	require.True(t, exists, "Auth secret should be created")
	assert.NotEmpty(t, authSecret.Data["default_password"], "Admin password should be set from stackstate.authentication.adminPassword")

	// Verify environment variables are set correctly with baseUrl from stackstate.baseUrl
	apiDeployment, exists := resources.Deployments["suse-observability-api"]
	require.True(t, exists, "API deployment should exist")
	envVars := getEnvVars(apiDeployment.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "http://localhost", envVars["STACKSTATE_BASE_URL"], "Base URL should be set from stackstate.baseUrl")
}

// TestGlobalSuseObservabilityHelperFallback tests that helpers fall back to stackstate.* values when no global.suseObservability values are set
func TestGlobalSuseObservabilityHelperFallback(t *testing.T) {
	// This test uses the full.yaml which has no global.suseObservability values set
	// so it should use stackstate.* values (legacy mode)
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Verify deployments exist (proving the chart renders successfully)
	_, syncExists := resources.Deployments["suse-observability-sync"]
	require.True(t, syncExists, "Sync deployment should exist")

	apiDeployment, apiExists := resources.Deployments["suse-observability-api"]
	require.True(t, apiExists, "API deployment should exist")

	// Verify environment variables use stackstate.baseUrl
	envVars := getEnvVars(apiDeployment.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "http://localhost", envVars["STACKSTATE_BASE_URL"], "Base URL should fall back to stackstate.baseUrl")
}

// TestGlobalSuseObservabilityPrecedence tests that global.suseObservability values take precedence when any global value is set
func TestGlobalSuseObservabilityPrecedence(t *testing.T) {
	// Create a values file that has both global.suseObservability and stackstate.* values
	// global.suseObservability should take precedence when auto-detected
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_suse_observability_valid.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	apiDeployment, exists := resources.Deployments["suse-observability-api"]
	require.True(t, exists, "API deployment should exist")

	envVars := getEnvVars(apiDeployment.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "http://test.example.com", envVars["STACKSTATE_BASE_URL"],
		"Base URL should use global.suseObservability.baseUrl when global mode is auto-detected")
}

// TestGlobalSuseObservabilityAutoDetectionWithProfile tests that auto-detection works when only sizing profile is set
func TestGlobalSuseObservabilityAutoDetectionWithProfile(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_suse_observability_profile_only.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Verify that the chart renders successfully with sizing profile
	// Note: 10-nonha profile uses monolithic server, not split deployments
	serverDeployment, exists := resources.Deployments["suse-observability-server"]
	require.True(t, exists, "Server deployment should exist")

	// Verify baseUrl is set from global.suseObservability.baseUrl
	envVars := getEnvVars(serverDeployment.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "http://test.example.com", envVars["STACKSTATE_BASE_URL"],
		"Base URL should be set from global.suseObservability.baseUrl when profile is set")

	// Verify license secret is created
	licenseSecret, exists := resources.Secrets["suse-observability-license"]
	require.True(t, exists, "License secret should be created")
	assert.NotEmpty(t, licenseSecret.Data["LICENSE_KEY"], "License key should be set")
}

// TestGlobalSuseObservabilityAutoDetectionMinimal tests that auto-detection works with minimal config (license + baseUrl)
func TestGlobalSuseObservabilityAutoDetectionMinimal(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_suse_observability_minimal_license.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Verify that the chart renders successfully with minimal config
	apiDeployment, exists := resources.Deployments["suse-observability-api"]
	require.True(t, exists, "API deployment should exist")

	// Verify baseUrl is set from global.suseObservability.baseUrl
	envVars := getEnvVars(apiDeployment.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "http://test.example.com", envVars["STACKSTATE_BASE_URL"],
		"Base URL should be set from global.suseObservability.baseUrl")

	// Verify license secret is created
	licenseSecret, exists := resources.Secrets["suse-observability-license"]
	require.True(t, exists, "License secret should be created")
	assert.NotEmpty(t, licenseSecret.Data["LICENSE_KEY"], "License key should be set")
}

// TestGlobalSuseObservabilityBcryptPassword tests that bcrypt hashed passwords are accepted and not re-hashed
func TestGlobalSuseObservabilityBcryptPassword(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_suse_observability_bcrypt_password.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Verify auth secret is created
	authSecret, exists := resources.Secrets["suse-observability-auth"]
	require.True(t, exists, "Auth secret should be created")
	require.Contains(t, authSecret.Data, "default_password", "Auth secret should contain default_password")

	// Decode the password and verify it's the same bcrypt hash we provided (not re-hashed)
	passwordBytes := authSecret.Data["default_password"]
	passwordStr := string(passwordBytes)

	// The bcrypt hash should start with $2b$ and be preserved as-is
	assert.Contains(t, passwordStr, "$2b$10$N9qo8uLOickgx2ZMRZoMye", "Bcrypt hash should be preserved without re-hashing")
}

// TestGlobalSuseObservabilityMissingAdminPassword tests that adminPassword is required in global mode
func TestGlobalSuseObservabilityMissingAdminPassword(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/global_suse_observability_missing_adminpassword.yaml")
	require.Error(t, err, "Should fail when adminPassword is missing in global mode")
	require.Contains(t, err.Error(), "global.suseObservability.adminPassword is required", "Error should mention adminPassword is required")
}

// Helper function to convert environment variables array to map for easier testing
func getEnvVars(envVars []v1.EnvVar) map[string]string {
	envMap := make(map[string]string)
	for _, env := range envVars {
		envMap[env.Name] = env.Value
	}
	return envMap
}
