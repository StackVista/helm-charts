package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

// TestS3ProxyAlwaysEnabled verifies that S3Proxy is deployed even when global.backup.enabled=false
// because it's needed for settings-local-backup
func TestS3ProxyAlwaysEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// S3Proxy deployment should exist
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist even when backup is disabled")

	// Should only mount settings-data PVC (not main-data)
	foundSettingsVolume := false
	foundMainVolume := false
	for _, vol := range deployment.Spec.Template.Spec.Volumes {
		if vol.Name == "settings-data" {
			foundSettingsVolume = true
		}
		if vol.Name == "main-data" {
			foundMainVolume = true
		}
	}
	assert.True(t, foundSettingsVolume, "Settings data volume should be present")
	assert.False(t, foundMainVolume, "Main data volume should NOT be present when backup is disabled")

	// Settings PVC should exist
	_, ok = resources.PersistentVolumeClaims["suse-observability-backup-settings-data"]
	require.True(t, ok, "Settings data PVC should exist")

	// Main data PVC should NOT exist
	_, ok = resources.PersistentVolumeClaims["suse-observability-minio"]
	assert.False(t, ok, "Main data PVC should NOT exist when backup is disabled")
}

// TestS3ProxyWithBackupEnabledPVC verifies S3Proxy deployment with PVC backend when backup is enabled
func TestS3ProxyWithBackupEnabledPVC(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":              "true",
			"backup.storage.backend.pvc.enabled": "true",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// S3Proxy deployment should exist
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// Should mount both settings-data and main-data PVCs
	foundSettingsVolume := false
	foundMainVolume := false
	for _, vol := range deployment.Spec.Template.Spec.Volumes {
		if vol.Name == "settings-data" {
			foundSettingsVolume = true
		}
		if vol.Name == "main-data" {
			foundMainVolume = true
		}
	}
	assert.True(t, foundSettingsVolume, "Settings data volume should be present")
	assert.True(t, foundMainVolume, "Main data volume should be present with PVC backend")

	// Both PVCs should exist
	_, ok = resources.PersistentVolumeClaims["suse-observability-backup-settings-data"]
	require.True(t, ok, "Settings data PVC should exist")

	mainPvc, ok := resources.PersistentVolumeClaims["suse-observability-minio"]
	require.True(t, ok, "Main data PVC should exist with PVC backend")
	assert.Equal(t, "500Gi", mainPvc.Spec.Resources.Requests.Storage().String(), "Main PVC should have default size")
}

// TestS3ProxyWithBackupEnabledS3 verifies S3Proxy deployment with S3 backend
func TestS3ProxyWithBackupEnabledS3(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":               "true",
			"backup.storage.backend.pvc.enabled":  "false",
			"backup.storage.backend.s3.enabled":   "true",
			"backup.storage.backend.s3.region":    "eu-west-1",
			"backup.storage.backend.s3.accessKey": "test-access-key",
			"backup.storage.backend.s3.secretKey": "test-secret-key",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// S3Proxy deployment should exist
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// Should only mount settings-data volume (not main-data since S3 backend is used)
	foundSettingsVolume := false
	foundMainVolume := false
	for _, vol := range deployment.Spec.Template.Spec.Volumes {
		if vol.Name == "settings-data" {
			foundSettingsVolume = true
		}
		if vol.Name == "main-data" {
			foundMainVolume = true
		}
	}
	assert.True(t, foundSettingsVolume, "Settings data volume should be present")
	assert.False(t, foundMainVolume, "Main data volume should NOT be present with S3 backend")

	// Settings PVC should exist
	_, ok = resources.PersistentVolumeClaims["suse-observability-backup-settings-data"]
	require.True(t, ok, "Settings data PVC should exist")

	// Main data PVC should NOT exist with S3 backend
	_, ok = resources.PersistentVolumeClaims["suse-observability-minio"]
	assert.False(t, ok, "Main data PVC should NOT exist with S3 backend")

	// Secret should contain backend credentials
	secret, ok := resources.Secrets["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy secret should exist")
	_, hasBackendAccessKey := secret.Data["backendAccessKey"]
	_, hasBackendSecretKey := secret.Data["backendSecretKey"]
	assert.True(t, hasBackendAccessKey, "Secret should contain backendAccessKey for S3")
	assert.True(t, hasBackendSecretKey, "Secret should contain backendSecretKey for S3")

	// ConfigMap should contain S3 backend config
	configMap, ok := resources.ConfigMaps["suse-observability-s3proxy-config"]
	require.True(t, ok, "S3Proxy ConfigMap should exist")
	assert.Contains(t, configMap.Data["s3proxy-main.properties"], "jclouds.provider=aws-s3-sdk", "ConfigMap should have S3 provider")

	// Region should be set via environment variable on the deployment
	expectedRegionEnv := corev1.EnvVar{Name: "AWS_REGION", Value: "eu-west-1"}
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, expectedRegionEnv, "Deployment should have AWS_REGION env var set to the configured region")
}

// TestS3ProxyWithBackupEnabledAzure verifies S3Proxy deployment with Azure backend
func TestS3ProxyWithBackupEnabledAzure(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":                    "true",
			"backup.storage.backend.pvc.enabled":       "false",
			"backup.storage.backend.azure.enabled":     "true",
			"backup.storage.backend.azure.accountName": "mystorageaccount",
			"backup.storage.backend.azure.accountKey":  "secret-key-here",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// S3Proxy deployment should exist
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// Should only mount settings-data volume (not main-data since Azure backend is used)
	foundMainVolume := false
	for _, vol := range deployment.Spec.Template.Spec.Volumes {
		if vol.Name == "main-data" {
			foundMainVolume = true
		}
	}
	assert.False(t, foundMainVolume, "Main data volume should NOT be present with Azure backend")

	// Secret should contain Azure credentials
	secret, ok := resources.Secrets["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy secret should exist")
	_, hasAzureAccountName := secret.Data["azureAccountName"]
	_, hasAzureAccountKey := secret.Data["azureAccountKey"]
	assert.True(t, hasAzureAccountName, "Secret should contain azureAccountName")
	assert.True(t, hasAzureAccountKey, "Secret should contain azureAccountKey")

	// ConfigMap should contain Azure backend config
	configMap, ok := resources.ConfigMaps["suse-observability-s3proxy-config"]
	require.True(t, ok, "S3Proxy ConfigMap should exist")
	assert.Contains(t, configMap.Data["s3proxy-main.properties"], "jclouds.provider=azureblob-sdk", "ConfigMap should have Azure provider")
	assert.Contains(t, configMap.Data["s3proxy-main.properties"], "mystorageaccount.blob.core.windows.net", "ConfigMap should have correct Azure endpoint")
}

// TestS3ProxyLegacyS3GatewayBackwardCompatibility verifies backward compatibility with minio.s3gateway values
func TestS3ProxyLegacyS3GatewayBackwardCompatibility(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":              "true",
			"backup.storage.backend.pvc.enabled": "false",
			"minio.s3gateway.enabled":            "true",
			"minio.s3gateway.serviceEndpoint":    "https://custom-s3.example.com",
			"minio.s3gateway.accessKey":          "legacy-access-key",
			"minio.s3gateway.secretKey":          "legacy-secret-key",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// S3Proxy deployment should exist
	_, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// ConfigMap should contain the legacy S3 endpoint
	configMap, ok := resources.ConfigMaps["suse-observability-s3proxy-config"]
	require.True(t, ok, "S3Proxy ConfigMap should exist")
	assert.Contains(t, configMap.Data["s3proxy-main.properties"], "jclouds.provider=aws-s3-sdk", "ConfigMap should have S3 provider from legacy values")
	assert.Contains(t, configMap.Data["s3proxy-main.properties"], "jclouds.endpoint=https://custom-s3.example.com", "ConfigMap should have legacy S3 endpoint")

	// Secret should contain legacy credentials
	secret, ok := resources.Secrets["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy secret should exist")
	_, hasBackendAccessKey := secret.Data["backendAccessKey"]
	_, hasBackendSecretKey := secret.Data["backendSecretKey"]
	assert.True(t, hasBackendAccessKey, "Secret should contain backendAccessKey from legacy values")
	assert.True(t, hasBackendSecretKey, "Secret should contain backendSecretKey from legacy values")
}

// TestS3ProxyLegacyAzureGatewayBackwardCompatibility verifies backward compatibility with minio.azuregateway values
func TestS3ProxyLegacyAzureGatewayBackwardCompatibility(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":              "true",
			"backup.storage.backend.pvc.enabled": "false",
			"minio.azuregateway.enabled":         "true",
			"minio.accessKey":                    "legacystorageaccount",
			"minio.secretKey":                    "legacy-azure-key",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// S3Proxy deployment should exist
	_, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// ConfigMap should contain Azure backend config from legacy values
	configMap, ok := resources.ConfigMaps["suse-observability-s3proxy-config"]
	require.True(t, ok, "S3Proxy ConfigMap should exist")
	assert.Contains(t, configMap.Data["s3proxy-main.properties"], "jclouds.provider=azureblob-sdk", "ConfigMap should have Azure provider from legacy values")
	assert.Contains(t, configMap.Data["s3proxy-main.properties"], "legacystorageaccount.blob.core.windows.net", "ConfigMap should have Azure endpoint from legacy values")

	// Secret should contain Azure credentials from legacy values
	secret, ok := resources.Secrets["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy secret should exist")
	_, hasAzureAccountName := secret.Data["azureAccountName"]
	_, hasAzureAccountKey := secret.Data["azureAccountKey"]
	assert.True(t, hasAzureAccountName, "Secret should contain azureAccountName from legacy values")
	assert.True(t, hasAzureAccountKey, "Secret should contain azureAccountKey from legacy values")
}

// TestS3ProxyCredentialsFallbackToMinioAccessKey verifies that when minio.accessKey/secretKey are set
// and global.s3proxy.credentials are at their defaults, the minio values are used as fallback
func TestS3ProxyCredentialsFallbackToMinioAccessKey(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"minio.accessKey": "minio-legacy-access",
			"minio.secretKey": "minio-legacy-secret",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Secret should contain the minio legacy credentials as the s3proxy access/secret keys
	secret, ok := resources.Secrets["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy secret should exist")
	assert.Equal(t, []byte("minio-legacy-access"), secret.Data["accesskey"], "accesskey should fall back to minio.accessKey")
	assert.Equal(t, []byte("minio-legacy-secret"), secret.Data["secretkey"], "secretkey should fall back to minio.secretKey")
}

// TestS3ProxyCredentialsExplicitOverridesMinioFallback verifies that when global.s3proxy.credentials
// are explicitly set, they take precedence over minio.accessKey/secretKey
func TestS3ProxyCredentialsExplicitOverridesMinioFallback(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.s3proxy.credentials.accessKey": "explicit-access",
			"global.s3proxy.credentials.secretKey": "explicit-secret",
			"minio.accessKey":                      "minio-legacy-access",
			"minio.secretKey":                      "minio-legacy-secret",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Secret should contain the explicit s3proxy credentials, not the minio ones
	secret, ok := resources.Secrets["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy secret should exist")
	assert.Equal(t, []byte("explicit-access"), secret.Data["accesskey"], "accesskey should use explicit global value, not minio fallback")
	assert.Equal(t, []byte("explicit-secret"), secret.Data["secretkey"], "secretkey should use explicit global value, not minio fallback")
}

// TestS3ProxyConfigMapBuckets verifies the bucket-locator configuration in the ConfigMap
func TestS3ProxyConfigMapBuckets(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled": "true",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, ok := resources.ConfigMaps["suse-observability-s3proxy-config"]
	require.True(t, ok, "S3Proxy ConfigMap should exist")

	// Settings properties should have settings-local-backup bucket
	assert.Contains(t, configMap.Data["s3proxy-settings.properties"], "s3proxy.bucket-locator.1=local-settings-backup",
		"Settings properties should route local-settings-backup bucket")

	// Main properties should have all backup buckets
	mainProps := configMap.Data["s3proxy-main.properties"]
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.1=sts-configuration-backup", "Main properties should route configuration backup bucket")
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.2=sts-stackgraph-backup", "Main properties should route stackgraph backup bucket")
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.3=sts-elasticsearch-backup", "Main properties should route elasticsearch backup bucket")
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.4=sts-clickhouse-backup", "Main properties should route clickhouse backup bucket")
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.5=sts-victoria-metrics-backup", "Main properties should route victoria-metrics-0 backup bucket")
}

// TestS3ProxyService verifies the S3Proxy service configuration
func TestS3ProxyService(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	service, ok := resources.Services["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy service should exist")
	assert.Equal(t, int32(9000), service.Spec.Ports[0].Port, "Service port should be 9000")
	assert.Equal(t, "http", service.Spec.Ports[0].Name, "Service port name should be http")
}

// TestS3ProxyResources verifies resource configuration
func TestS3ProxyResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.resources.limits.cpu":      "1000m",
			"s3proxy.resources.limits.memory":   "1Gi",
			"s3proxy.resources.requests.cpu":    "200m",
			"s3proxy.resources.requests.memory": "512Mi",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	container := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "1", container.Resources.Limits.Cpu().String(), "CPU limit should be set")
	assert.Equal(t, "1Gi", container.Resources.Limits.Memory().String(), "Memory limit should be set")
	assert.Equal(t, "200m", container.Resources.Requests.Cpu().String(), "CPU request should be set")
	assert.Equal(t, "512Mi", container.Resources.Requests.Memory().String(), "Memory request should be set")
}

// TestS3ProxyDefaultArgs verifies the exact default commandline arguments for the s3proxy container
func TestS3ProxyDefaultArgs(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	container := deployment.Spec.Template.Spec.Containers[0]

	// Default values in full.yaml have global.backup.enabled=true,
	// so we expect both settings and main properties files
	expectedArgs := []string{
		"--properties",
		"/etc/s3proxy/s3proxy-settings.properties",
		"--properties",
		"/etc/s3proxy/s3proxy-main.properties",
	}
	assert.Equal(t, expectedArgs, container.Args, "Default args should load both settings and main properties files")

	// Container should not have a custom command (uses image entrypoint)
	assert.Empty(t, container.Command, "Container should use the default image entrypoint (no custom command)")
}

// TestS3ProxyDeploymentArgs verifies the deployment args based on backup.enabled
func TestS3ProxyDeploymentArgs(t *testing.T) {
	t.Run("backup disabled", func(t *testing.T) {
		output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
			ValuesFiles: []string{"values/full.yaml"},
			SetValues: map[string]string{
				"global.backup.enabled": "false",
			},
		})
		resources := helmtestutil.NewKubernetesResources(t, output)

		deployment, ok := resources.Deployments["suse-observability-s3proxy"]
		require.True(t, ok, "S3Proxy deployment should exist")
		container := deployment.Spec.Template.Spec.Containers[0]

		// Should only have settings properties
		assert.Contains(t, container.Args, "--properties", "Should have --properties arg")
		assert.Contains(t, container.Args, "/etc/s3proxy/s3proxy-settings.properties", "Should have settings properties")
		assert.NotContains(t, container.Args, "/etc/s3proxy/s3proxy-main.properties", "Should NOT have main properties when backup disabled")
	})

	t.Run("backup enabled", func(t *testing.T) {
		output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
			ValuesFiles: []string{"values/full.yaml"},
			SetValues: map[string]string{
				"global.backup.enabled": "true",
			},
		})
		resources := helmtestutil.NewKubernetesResources(t, output)

		deployment, ok := resources.Deployments["suse-observability-s3proxy"]
		require.True(t, ok, "S3Proxy deployment should exist")
		container := deployment.Spec.Template.Spec.Containers[0]

		// Should have both settings and main properties
		assert.Contains(t, container.Args, "/etc/s3proxy/s3proxy-settings.properties", "Should have settings properties")
		assert.Contains(t, container.Args, "/etc/s3proxy/s3proxy-main.properties", "Should have main properties when backup enabled")
	})
}

// TestS3ProxyFromExternalSecret verifies using an externally-managed secret
func TestS3ProxyFromExternalSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.s3proxy.credentials.fromExternalSecret": "my-existing-secret",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// S3Proxy managed secret should NOT exist
	_, ok := resources.Secrets["suse-observability-s3proxy"]
	assert.False(t, ok, "S3Proxy managed secret should NOT exist when using fromExternalSecret")

	// Deployment should reference the existing secret
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	container := deployment.Spec.Template.Spec.Containers[0]

	// Find the S3PROXY_IDENTITY env var
	var accessKeyEnv *corev1.EnvVar
	for i := range container.Env {
		if container.Env[i].Name == "S3PROXY_IDENTITY" {
			accessKeyEnv = &container.Env[i]
			break
		}
	}
	require.NotNil(t, accessKeyEnv, "S3PROXY_IDENTITY env var should exist")
	assert.Equal(t, "my-existing-secret", accessKeyEnv.ValueFrom.SecretKeyRef.Name, "Should reference existing secret")
}

// TestS3ProxyFromExternalSecretCredentialsNotRequired verifies that when fromExternalSecret is set,
// the accessKey and secretKey values are not required and can be empty
func TestS3ProxyFromExternalSecretCredentialsNotRequired(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.s3proxy.credentials.fromExternalSecret": "my-external-secret",
			"global.s3proxy.credentials.accessKey":          "",
			"global.s3proxy.credentials.secretKey":          "",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// S3Proxy managed secret should NOT exist
	_, ok := resources.Secrets["suse-observability-s3proxy"]
	assert.False(t, ok, "S3Proxy managed secret should NOT exist when using fromExternalSecret")

	// Deployment should exist and reference the external secret
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	container := deployment.Spec.Template.Spec.Containers[0]

	// Both S3PROXY_IDENTITY and S3PROXY_CREDENTIAL should reference the external secret
	var identityEnv, credentialEnv *corev1.EnvVar
	for i := range container.Env {
		if container.Env[i].Name == "S3PROXY_IDENTITY" {
			identityEnv = &container.Env[i]
		}
		if container.Env[i].Name == "S3PROXY_CREDENTIAL" {
			credentialEnv = &container.Env[i]
		}
	}
	require.NotNil(t, identityEnv, "S3PROXY_IDENTITY env var should exist")
	assert.Equal(t, "my-external-secret", identityEnv.ValueFrom.SecretKeyRef.Name, "S3PROXY_IDENTITY should reference external secret")
	assert.Equal(t, "accesskey", identityEnv.ValueFrom.SecretKeyRef.Key, "S3PROXY_IDENTITY should use accesskey key")

	require.NotNil(t, credentialEnv, "S3PROXY_CREDENTIAL env var should exist")
	assert.Equal(t, "my-external-secret", credentialEnv.ValueFrom.SecretKeyRef.Name, "S3PROXY_CREDENTIAL should reference external secret")
	assert.Equal(t, "secretkey", credentialEnv.ValueFrom.SecretKeyRef.Key, "S3PROXY_CREDENTIAL should use secretkey key")
}

// TestS3ProxyS3BackendFromExternalSecret verifies that when fromExternalSecret is set for the S3 backend,
// the managed secret does not contain backend credentials and the deployment references the external secret
func TestS3ProxyS3BackendFromExternalSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":                        "true",
			"backup.storage.backend.pvc.enabled":           "false",
			"backup.storage.backend.s3.enabled":            "true",
			"backup.storage.backend.s3.fromExternalSecret": "my-s3-backend-secret",
			"backup.storage.backend.s3.region":             "eu-west-1",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Managed secret should still exist (for s3proxy credentials) but should NOT contain backend keys
	secret, ok := resources.Secrets["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy managed secret should still exist for s3proxy credentials")
	_, hasBackendAccessKey := secret.Data["backendAccessKey"]
	_, hasBackendSecretKey := secret.Data["backendSecretKey"]
	assert.False(t, hasBackendAccessKey, "Managed secret should NOT contain backendAccessKey when S3 fromExternalSecret is set")
	assert.False(t, hasBackendSecretKey, "Managed secret should NOT contain backendSecretKey when S3 fromExternalSecret is set")

	// Deployment should reference the external secret for JCLOUDS env vars
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	container := deployment.Spec.Template.Spec.Containers[0]

	var identityEnv, credentialEnv *corev1.EnvVar
	for i := range container.Env {
		if container.Env[i].Name == "JCLOUDS_IDENTITY" {
			identityEnv = &container.Env[i]
		}
		if container.Env[i].Name == "JCLOUDS_CREDENTIAL" {
			credentialEnv = &container.Env[i]
		}
	}
	require.NotNil(t, identityEnv, "JCLOUDS_IDENTITY env var should exist")
	assert.Equal(t, "my-s3-backend-secret", identityEnv.ValueFrom.SecretKeyRef.Name, "JCLOUDS_IDENTITY should reference S3 external secret")
	assert.Equal(t, "backendAccessKey", identityEnv.ValueFrom.SecretKeyRef.Key, "JCLOUDS_IDENTITY should use backendAccessKey key")

	require.NotNil(t, credentialEnv, "JCLOUDS_CREDENTIAL env var should exist")
	assert.Equal(t, "my-s3-backend-secret", credentialEnv.ValueFrom.SecretKeyRef.Name, "JCLOUDS_CREDENTIAL should reference S3 external secret")
	assert.Equal(t, "backendSecretKey", credentialEnv.ValueFrom.SecretKeyRef.Key, "JCLOUDS_CREDENTIAL should use backendSecretKey key")
}

// TestS3ProxyS3BackendFromExternalSecretCredentialsNotRequired verifies that when fromExternalSecret
// is set for the S3 backend, the accessKey and secretKey values are not required
func TestS3ProxyS3BackendFromExternalSecretCredentialsNotRequired(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":                        "true",
			"backup.storage.backend.pvc.enabled":           "false",
			"backup.storage.backend.s3.enabled":            "true",
			"backup.storage.backend.s3.fromExternalSecret": "my-s3-backend-secret",
			"backup.storage.backend.s3.accessKey":          "",
			"backup.storage.backend.s3.secretKey":          "",
			"backup.storage.backend.s3.region":             "eu-west-1",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Managed secret should NOT contain backend keys
	secret, ok := resources.Secrets["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy managed secret should still exist for s3proxy credentials")
	_, hasBackendAccessKey := secret.Data["backendAccessKey"]
	_, hasBackendSecretKey := secret.Data["backendSecretKey"]
	assert.False(t, hasBackendAccessKey, "Managed secret should NOT contain backendAccessKey")
	assert.False(t, hasBackendSecretKey, "Managed secret should NOT contain backendSecretKey")

	// Deployment should reference the external secret
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	container := deployment.Spec.Template.Spec.Containers[0]

	var identityEnv, credentialEnv *corev1.EnvVar
	for i := range container.Env {
		if container.Env[i].Name == "JCLOUDS_IDENTITY" {
			identityEnv = &container.Env[i]
		}
		if container.Env[i].Name == "JCLOUDS_CREDENTIAL" {
			credentialEnv = &container.Env[i]
		}
	}
	require.NotNil(t, identityEnv, "JCLOUDS_IDENTITY env var should exist")
	assert.Equal(t, "my-s3-backend-secret", identityEnv.ValueFrom.SecretKeyRef.Name, "JCLOUDS_IDENTITY should reference external secret")

	require.NotNil(t, credentialEnv, "JCLOUDS_CREDENTIAL env var should exist")
	assert.Equal(t, "my-s3-backend-secret", credentialEnv.ValueFrom.SecretKeyRef.Name, "JCLOUDS_CREDENTIAL should reference external secret")
}

// TestS3ProxyAzureBackendFromExternalSecret verifies that when fromExternalSecret is set for the Azure backend,
// the managed secret does not contain Azure credentials and the deployment references the external secret
func TestS3ProxyAzureBackendFromExternalSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":                           "true",
			"backup.storage.backend.pvc.enabled":              "false",
			"backup.storage.backend.azure.enabled":            "true",
			"backup.storage.backend.azure.fromExternalSecret": "my-azure-backend-secret",
			"backup.storage.backend.azure.endpoint":           "https://mystorageaccount.blob.core.windows.net",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Managed secret should still exist (for s3proxy credentials) but should NOT contain Azure keys
	secret, ok := resources.Secrets["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy managed secret should still exist for s3proxy credentials")
	_, hasAzureAccountName := secret.Data["azureAccountName"]
	_, hasAzureAccountKey := secret.Data["azureAccountKey"]
	assert.False(t, hasAzureAccountName, "Managed secret should NOT contain azureAccountName when Azure fromExternalSecret is set")
	assert.False(t, hasAzureAccountKey, "Managed secret should NOT contain azureAccountKey when Azure fromExternalSecret is set")

	// Deployment should reference the external secret for JCLOUDS env vars
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	container := deployment.Spec.Template.Spec.Containers[0]

	var identityEnv, credentialEnv *corev1.EnvVar
	for i := range container.Env {
		if container.Env[i].Name == "JCLOUDS_IDENTITY" {
			identityEnv = &container.Env[i]
		}
		if container.Env[i].Name == "JCLOUDS_CREDENTIAL" {
			credentialEnv = &container.Env[i]
		}
	}
	require.NotNil(t, identityEnv, "JCLOUDS_IDENTITY env var should exist")
	assert.Equal(t, "my-azure-backend-secret", identityEnv.ValueFrom.SecretKeyRef.Name, "JCLOUDS_IDENTITY should reference Azure external secret")
	assert.Equal(t, "azureAccountName", identityEnv.ValueFrom.SecretKeyRef.Key, "JCLOUDS_IDENTITY should use azureAccountName key")

	require.NotNil(t, credentialEnv, "JCLOUDS_CREDENTIAL env var should exist")
	assert.Equal(t, "my-azure-backend-secret", credentialEnv.ValueFrom.SecretKeyRef.Name, "JCLOUDS_CREDENTIAL should reference Azure external secret")
	assert.Equal(t, "azureAccountKey", credentialEnv.ValueFrom.SecretKeyRef.Key, "JCLOUDS_CREDENTIAL should use azureAccountKey key")
}

// TestS3ProxyAzureBackendFromExternalSecretCredentialsNotRequired verifies that when fromExternalSecret
// is set for the Azure backend, the accountName and accountKey values are not required
func TestS3ProxyAzureBackendFromExternalSecretCredentialsNotRequired(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":                           "true",
			"backup.storage.backend.pvc.enabled":              "false",
			"backup.storage.backend.azure.enabled":            "true",
			"backup.storage.backend.azure.fromExternalSecret": "my-azure-backend-secret",
			"backup.storage.backend.azure.accountName":        "",
			"backup.storage.backend.azure.accountKey":         "",
			"backup.storage.backend.azure.endpoint":           "https://mystorageaccount.blob.core.windows.net",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Managed secret should NOT contain Azure keys
	secret, ok := resources.Secrets["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy managed secret should still exist for s3proxy credentials")
	_, hasAzureAccountName := secret.Data["azureAccountName"]
	_, hasAzureAccountKey := secret.Data["azureAccountKey"]
	assert.False(t, hasAzureAccountName, "Managed secret should NOT contain azureAccountName")
	assert.False(t, hasAzureAccountKey, "Managed secret should NOT contain azureAccountKey")

	// Deployment should reference the external secret
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	container := deployment.Spec.Template.Spec.Containers[0]

	var identityEnv, credentialEnv *corev1.EnvVar
	for i := range container.Env {
		if container.Env[i].Name == "JCLOUDS_IDENTITY" {
			identityEnv = &container.Env[i]
		}
		if container.Env[i].Name == "JCLOUDS_CREDENTIAL" {
			credentialEnv = &container.Env[i]
		}
	}
	require.NotNil(t, identityEnv, "JCLOUDS_IDENTITY env var should exist")
	assert.Equal(t, "my-azure-backend-secret", identityEnv.ValueFrom.SecretKeyRef.Name, "JCLOUDS_IDENTITY should reference external secret")

	require.NotNil(t, credentialEnv, "JCLOUDS_CREDENTIAL env var should exist")
	assert.Equal(t, "my-azure-backend-secret", credentialEnv.ValueFrom.SecretKeyRef.Name, "JCLOUDS_CREDENTIAL should reference external secret")
}

// TestS3ProxyNodeSelector verifies node selector configuration
func TestS3ProxyNodeSelector(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.nodeSelector.disk-type": "ssd",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	assert.Equal(t, "ssd", deployment.Spec.Template.Spec.NodeSelector["disk-type"], "Node selector should be set")
}

// TestS3ProxyTolerations verifies tolerations configuration
func TestS3ProxyTolerations(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.tolerations[0].key":      "dedicated",
			"s3proxy.tolerations[0].operator": "Equal",
			"s3proxy.tolerations[0].value":    "storage",
			"s3proxy.tolerations[0].effect":   "NoSchedule",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	require.NotEmpty(t, deployment.Spec.Template.Spec.Tolerations, "Tolerations should be set")

	found := false
	for _, tol := range deployment.Spec.Template.Spec.Tolerations {
		if tol.Key == "dedicated" && tol.Value == "storage" {
			found = true
			break
		}
	}
	assert.True(t, found, "Custom toleration should be present")
}

// TestS3ProxySettingsPVCSize verifies custom settings PVC size
func TestS3ProxySettingsPVCSize(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"backup.storage.settingsPvc.size": "5Gi",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	pvc, ok := resources.PersistentVolumeClaims["suse-observability-backup-settings-data"]
	require.True(t, ok, "Settings PVC should exist")
	assert.Equal(t, "5Gi", pvc.Spec.Resources.Requests.Storage().String(), "Settings PVC size should be customized")
}

// TestS3ProxyMainPVCSize verifies custom main PVC size
func TestS3ProxyMainPVCSize(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":           "true",
			"backup.storage.backend.pvc.size": "1Ti",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	pvc, ok := resources.PersistentVolumeClaims["suse-observability-minio"]
	require.True(t, ok, "Main PVC should exist")
	assert.Equal(t, "1Ti", pvc.Spec.Resources.Requests.Storage().String(), "Main PVC size should be customized")
}

// TestS3ProxyInitContainerBackupDisabled verifies init container creates only settings bucket when backup is disabled
func TestS3ProxyInitContainerBackupDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// Should have init container
	require.Len(t, deployment.Spec.Template.Spec.InitContainers, 1, "Should have exactly one init container")
	initContainer := deployment.Spec.Template.Spec.InitContainers[0]
	assert.Equal(t, "create-bucket-dirs", initContainer.Name, "Init container should be named create-bucket-dirs")

	// Init container should only mount settings-data volume
	var settingsMount, mainMount *corev1.VolumeMount
	for i := range initContainer.VolumeMounts {
		if initContainer.VolumeMounts[i].Name == "settings-data" {
			settingsMount = &initContainer.VolumeMounts[i]
		}
		if initContainer.VolumeMounts[i].Name == "main-data" {
			mainMount = &initContainer.VolumeMounts[i]
		}
	}
	require.NotNil(t, settingsMount, "Init container should mount settings-data volume")
	assert.Equal(t, "/settings-data", settingsMount.MountPath, "Settings data should be mounted at /settings-data")
	assert.Nil(t, mainMount, "Init container should NOT mount main-data volume when backup is disabled")

	// Verify the command creates only settings bucket directory
	require.Len(t, initContainer.Command, 3, "Command should have 3 elements (sh, -c, script)")
	script := initContainer.Command[2]
	assert.Contains(t, script, "mkdir -p /settings-data/local-settings-backup", "Should create settings bucket directory")
	assert.NotContains(t, script, "/main-data/", "Should NOT create any main-data directories when backup is disabled")
}

// TestS3ProxyInitContainerBackupEnabledPVC verifies init container creates all bucket directories with PVC backend
func TestS3ProxyInitContainerBackupEnabledPVC(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":              "true",
			"backup.storage.backend.pvc.enabled": "true",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// Should have init container
	require.Len(t, deployment.Spec.Template.Spec.InitContainers, 1, "Should have exactly one init container")
	initContainer := deployment.Spec.Template.Spec.InitContainers[0]
	assert.Equal(t, "create-bucket-dirs", initContainer.Name, "Init container should be named create-bucket-dirs")

	// Init container should mount both volumes
	var settingsMount, mainMount *corev1.VolumeMount
	for i := range initContainer.VolumeMounts {
		if initContainer.VolumeMounts[i].Name == "settings-data" {
			settingsMount = &initContainer.VolumeMounts[i]
		}
		if initContainer.VolumeMounts[i].Name == "main-data" {
			mainMount = &initContainer.VolumeMounts[i]
		}
	}
	require.NotNil(t, settingsMount, "Init container should mount settings-data volume")
	assert.Equal(t, "/settings-data", settingsMount.MountPath, "Settings data should be mounted at /settings-data")
	require.NotNil(t, mainMount, "Init container should mount main-data volume when PVC backend is used")
	assert.Equal(t, "/main-data", mainMount.MountPath, "Main data should be mounted at /main-data")

	// Verify the command creates all bucket directories
	require.Len(t, initContainer.Command, 3, "Command should have 3 elements (sh, -c, script)")
	script := initContainer.Command[2]
	assert.Contains(t, script, "mkdir -p /settings-data/local-settings-backup", "Should create settings bucket directory")
	assert.Contains(t, script, "mkdir -p /main-data/sts-configuration-backup", "Should create configuration backup bucket directory")
	assert.Contains(t, script, "mkdir -p /main-data/sts-stackgraph-backup", "Should create stackgraph backup bucket directory")
	assert.Contains(t, script, "mkdir -p /main-data/sts-elasticsearch-backup", "Should create elasticsearch backup bucket directory")
	assert.Contains(t, script, "mkdir -p /main-data/sts-clickhouse-backup", "Should create clickhouse backup bucket directory")
	assert.Contains(t, script, "mkdir -p /main-data/sts-victoria-metrics-backup", "Should create victoria-metrics backup bucket directory")
}

// TestS3ProxyInitContainerBackupEnabledS3 verifies init container creates only settings bucket with S3 backend
func TestS3ProxyInitContainerBackupEnabledS3(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":               "true",
			"backup.storage.backend.pvc.enabled":  "false",
			"backup.storage.backend.s3.enabled":   "true",
			"backup.storage.backend.s3.region":    "eu-west-1",
			"backup.storage.backend.s3.accessKey": "test-access-key",
			"backup.storage.backend.s3.secretKey": "test-secret-key",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// Should have init container
	require.Len(t, deployment.Spec.Template.Spec.InitContainers, 1, "Should have exactly one init container")
	initContainer := deployment.Spec.Template.Spec.InitContainers[0]

	// Init container should only mount settings-data volume (not main-data since S3 backend is used)
	var settingsMount, mainMount *corev1.VolumeMount
	for i := range initContainer.VolumeMounts {
		if initContainer.VolumeMounts[i].Name == "settings-data" {
			settingsMount = &initContainer.VolumeMounts[i]
		}
		if initContainer.VolumeMounts[i].Name == "main-data" {
			mainMount = &initContainer.VolumeMounts[i]
		}
	}
	require.NotNil(t, settingsMount, "Init container should mount settings-data volume")
	assert.Nil(t, mainMount, "Init container should NOT mount main-data volume with S3 backend")

	// Verify the command creates only settings bucket directory
	require.Len(t, initContainer.Command, 3, "Command should have 3 elements (sh, -c, script)")
	script := initContainer.Command[2]
	assert.Contains(t, script, "mkdir -p /settings-data/local-settings-backup", "Should create settings bucket directory")
	assert.NotContains(t, script, "/main-data/", "Should NOT create any main-data directories with S3 backend")
}

// TestS3ProxyInitContainerBackupEnabledAzure verifies init container creates only settings bucket with Azure backend
func TestS3ProxyInitContainerBackupEnabledAzure(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":                    "true",
			"backup.storage.backend.pvc.enabled":       "false",
			"backup.storage.backend.azure.enabled":     "true",
			"backup.storage.backend.azure.accountName": "mystorageaccount",
			"backup.storage.backend.azure.accountKey":  "secret-key-here",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// Should have init container
	require.Len(t, deployment.Spec.Template.Spec.InitContainers, 1, "Should have exactly one init container")
	initContainer := deployment.Spec.Template.Spec.InitContainers[0]

	// Init container should only mount settings-data volume (not main-data since Azure backend is used)
	var mainMount *corev1.VolumeMount
	for i := range initContainer.VolumeMounts {
		if initContainer.VolumeMounts[i].Name == "main-data" {
			mainMount = &initContainer.VolumeMounts[i]
		}
	}
	assert.Nil(t, mainMount, "Init container should NOT mount main-data volume with Azure backend")

	// Verify the command creates only settings bucket directory
	require.Len(t, initContainer.Command, 3, "Command should have 3 elements (sh, -c, script)")
	script := initContainer.Command[2]
	assert.Contains(t, script, "mkdir -p /settings-data/local-settings-backup", "Should create settings bucket directory")
	assert.NotContains(t, script, "/main-data/", "Should NOT create any main-data directories with Azure backend")
}

// TestS3ProxyInitContainerCustomBucketNames verifies init container uses custom bucket names
func TestS3ProxyInitContainerCustomBucketNames(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":                "true",
			"backup.storage.backend.pvc.enabled":   "true",
			"backup.configuration.bucketName":      "custom-config-backup",
			"backup.stackGraph.bucketName":         "custom-stackgraph-backup",
			"backup.elasticsearch.bucketName":      "custom-es-backup",
			"clickhouse.backup.bucketName":         "custom-clickhouse-backup",
			"victoria-metrics-0.backup.bucketName": "custom-vm-backup",
			"victoria-metrics-1.backup.bucketName": "custom-vm-backup",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	initContainer := deployment.Spec.Template.Spec.InitContainers[0]
	require.Len(t, initContainer.Command, 3, "Command should have 3 elements (sh, -c, script)")
	script := initContainer.Command[2]

	// Verify custom bucket names are used
	assert.Contains(t, script, "mkdir -p /main-data/custom-config-backup", "Should create custom configuration backup bucket directory")
	assert.Contains(t, script, "mkdir -p /main-data/custom-stackgraph-backup", "Should create custom stackgraph backup bucket directory")
	assert.Contains(t, script, "mkdir -p /main-data/custom-es-backup", "Should create custom elasticsearch backup bucket directory")
	assert.Contains(t, script, "mkdir -p /main-data/custom-clickhouse-backup", "Should create custom clickhouse backup bucket directory")
	assert.Contains(t, script, "mkdir -p /main-data/custom-vm-backup", "Should create custom victoria-metrics backup bucket directory")
}

// TestS3ProxyInitContainerImage verifies init container uses the correct container-tools image
func TestS3ProxyInitContainerImage(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	initContainer := deployment.Spec.Template.Spec.InitContainers[0]
	assert.Contains(t, initContainer.Image, "container-tools", "Init container should use container-tools image")
}

// TestS3ProxyBackendValidationPVCAndS3 verifies that enabling both PVC and S3 backends fails
func TestS3ProxyBackendValidationPVCAndS3(t *testing.T) {
	// Now test with both PVC and S3 enabled
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":               "true",
			"backup.storage.backend.pvc.enabled":  "true",
			"backup.storage.backend.s3.enabled":   "true",
			"backup.storage.backend.s3.region":    "eu-west-1",
			"backup.storage.backend.s3.accessKey": "test-access-key",
			"backup.storage.backend.s3.secretKey": "test-secret-key",
		},
	})
	require.NotNil(t, err, "Should fail when both PVC and S3 backends are enabled")
	assert.Contains(t, err.Error(), "Only one backup storage backend can be enabled at a time", "Error message should mention mutually exclusive backends")
}

// TestS3ProxyBackendValidationPVCAndAzure verifies that enabling both PVC and Azure backends fails
func TestS3ProxyBackendValidationPVCAndAzure(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":                    "true",
			"backup.storage.backend.pvc.enabled":       "true",
			"backup.storage.backend.azure.enabled":     "true",
			"backup.storage.backend.azure.accountName": "mystorageaccount",
			"backup.storage.backend.azure.accountKey":  "secret-key-here",
		},
	})
	require.NotNil(t, err, "Should fail when both PVC and Azure backends are enabled")
	assert.Contains(t, err.Error(), "Only one backup storage backend can be enabled at a time", "Error message should mention mutually exclusive backends")
}

// TestS3ProxyBackendValidationS3AndAzure verifies that enabling both S3 and Azure backends fails
func TestS3ProxyBackendValidationS3AndAzure(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":                    "true",
			"backup.storage.backend.pvc.enabled":       "false",
			"backup.storage.backend.s3.enabled":        "true",
			"backup.storage.backend.s3.region":         "eu-west-1",
			"backup.storage.backend.s3.accessKey":      "test-access-key",
			"backup.storage.backend.s3.secretKey":      "test-secret-key",
			"backup.storage.backend.azure.enabled":     "true",
			"backup.storage.backend.azure.accountName": "mystorageaccount",
			"backup.storage.backend.azure.accountKey":  "secret-key-here",
		},
	})
	require.NotNil(t, err, "Should fail when both S3 and Azure backends are enabled")
	assert.Contains(t, err.Error(), "Only one backup storage backend can be enabled at a time", "Error message should mention mutually exclusive backends")
}

// TestS3ProxyBackendValidationAllThree verifies that enabling all three backends fails
func TestS3ProxyBackendValidationAllThree(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":                    "true",
			"backup.storage.backend.pvc.enabled":       "true",
			"backup.storage.backend.s3.enabled":        "true",
			"backup.storage.backend.s3.region":         "eu-west-1",
			"backup.storage.backend.s3.accessKey":      "test-access-key",
			"backup.storage.backend.s3.secretKey":      "test-secret-key",
			"backup.storage.backend.azure.enabled":     "true",
			"backup.storage.backend.azure.accountName": "mystorageaccount",
			"backup.storage.backend.azure.accountKey":  "secret-key-here",
		},
	})
	require.NotNil(t, err, "Should fail when all three backends are enabled")
	assert.Contains(t, err.Error(), "Only one backup storage backend can be enabled at a time", "Error message should mention mutually exclusive backends")
}

// TestS3ProxyBackendValidationSingleBackendAllowed verifies that enabling a single backend works
func TestS3ProxyBackendValidationSingleBackendAllowed(t *testing.T) {
	t.Run("PVC only", func(t *testing.T) {
		output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
			ValuesFiles: []string{"values/full.yaml"},
			SetValues: map[string]string{
				"global.backup.enabled":              "true",
				"backup.storage.backend.pvc.enabled": "true",
			},
		})
		resources := helmtestutil.NewKubernetesResources(t, output)
		_, ok := resources.Deployments["suse-observability-s3proxy"]
		assert.True(t, ok, "S3Proxy deployment should exist with PVC backend only")
	})

	t.Run("S3 only", func(t *testing.T) {
		output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
			ValuesFiles: []string{"values/full.yaml"},
			SetValues: map[string]string{
				"global.backup.enabled":               "true",
				"backup.storage.backend.pvc.enabled":  "false",
				"backup.storage.backend.s3.enabled":   "true",
				"backup.storage.backend.s3.region":    "eu-west-1",
				"backup.storage.backend.s3.accessKey": "test-access-key",
				"backup.storage.backend.s3.secretKey": "test-secret-key",
			},
		})
		resources := helmtestutil.NewKubernetesResources(t, output)
		_, ok := resources.Deployments["suse-observability-s3proxy"]
		assert.True(t, ok, "S3Proxy deployment should exist with S3 backend only")
	})

	t.Run("Azure only", func(t *testing.T) {
		output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
			ValuesFiles: []string{"values/full.yaml"},
			SetValues: map[string]string{
				"global.backup.enabled":                    "true",
				"backup.storage.backend.pvc.enabled":       "false",
				"backup.storage.backend.azure.enabled":     "true",
				"backup.storage.backend.azure.accountName": "mystorageaccount",
				"backup.storage.backend.azure.accountKey":  "secret-key-here",
			},
		})
		resources := helmtestutil.NewKubernetesResources(t, output)
		_, ok := resources.Deployments["suse-observability-s3proxy"]
		assert.True(t, ok, "S3Proxy deployment should exist with Azure backend only")
	})
}

// TestS3ProxyBackendValidationNotAppliedWhenBackupDisabled verifies validation is skipped when backup is disabled
func TestS3ProxyBackendValidationNotAppliedWhenBackupDisabled(t *testing.T) {
	// When backup is disabled, the validation should not fail even if multiple backends are "enabled"
	// because the backend settings are irrelevant
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":              "false",
			"backup.storage.backend.pvc.enabled": "true",
			"backup.storage.backend.s3.enabled":  "true",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	_, ok := resources.Deployments["suse-observability-s3proxy"]
	assert.True(t, ok, "S3Proxy deployment should exist even with conflicting backend settings when backup is disabled")
}

// TestS3ProxyJavaHeapDefaultValues verifies that JVM heap is calculated correctly with default values
func TestS3ProxyJavaHeapDefaultValues(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	container := deployment.Spec.Template.Spec.Containers[0]

	// Find JAVA_OPTS env var
	var javaOpts string
	for _, env := range container.Env {
		if env.Name == "JAVA_OPTS" {
			javaOpts = env.Value
			break
		}
	}
	require.NotEmpty(t, javaOpts, "JAVA_OPTS should be set")

	// With default values: memory limit 512Mi, base consumption 100Mi, heap fraction 70%
	// Available for heap: 512Mi * 1.048576 - 100Mi * 1.048576 = ~432MB
	// Heap: 432 * 70% = ~302MB for Xmx
	// Memory request 256Mi: (256 * 1.048576 - 100 * 1.048576) * 70% = ~114MB for Xms
	assert.Contains(t, javaOpts, "-Xmx", "JAVA_OPTS should contain -Xmx")
	assert.Contains(t, javaOpts, "-Xms", "JAVA_OPTS should contain -Xms")
}

// TestS3ProxyJavaHeapCustomValues verifies that JVM heap is calculated correctly with custom values
func TestS3ProxyJavaHeapCustomValues(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.resources.limits.memory":       "700Mi",
			"s3proxy.resources.requests.memory":     "700Mi",
			"s3proxy.sizing.baseMemoryConsumption":  "200Mi",
			"s3proxy.sizing.javaHeapMemoryFraction": "60",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	container := deployment.Spec.Template.Spec.Containers[0]

	// Find JAVA_OPTS env var
	var javaOpts string
	for _, env := range container.Env {
		if env.Name == "JAVA_OPTS" {
			javaOpts = env.Value
			break
		}
	}
	require.NotEmpty(t, javaOpts, "JAVA_OPTS should be set")

	// With custom values: memory limit 1Gi, base consumption 200Mi, heap fraction 60%
	// Available for heap: 1Gi * 1.07374 - 200Mi * 1.048576 = ~863MB
	// Heap: 863 * 60% = ~518MB for Xmx
	// Memory request 512Mi: (512 * 1.048576 - 200 * 1.048576) * 60% = ~196MB for Xms
	assert.Contains(t, javaOpts, "-Xmx", "JAVA_OPTS should contain -Xmx with custom memory settings")
	assert.Contains(t, javaOpts, "-Xms", "JAVA_OPTS should contain -Xms with custom memory settings")
}

// TestS3ProxyJavaHeapHighFraction verifies that JVM heap calculation handles high heap fractions
func TestS3ProxyJavaHeapHighFraction(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.resources.limits.memory":       "2Gi",
			"s3proxy.resources.requests.memory":     "1Gi",
			"s3proxy.sizing.baseMemoryConsumption":  "50Mi",
			"s3proxy.sizing.javaHeapMemoryFraction": "85",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	container := deployment.Spec.Template.Spec.Containers[0]

	// Find JAVA_OPTS env var
	var javaOpts string
	for _, env := range container.Env {
		if env.Name == "JAVA_OPTS" {
			javaOpts = env.Value
			break
		}
	}
	require.NotEmpty(t, javaOpts, "JAVA_OPTS should be set")

	// High heap fraction should still work
	assert.Contains(t, javaOpts, "-Xmx", "JAVA_OPTS should contain -Xmx with high heap fraction")
	assert.Contains(t, javaOpts, "-Xms", "JAVA_OPTS should contain -Xms with high heap fraction")
}

// TestS3ProxyCustomCATrustStore verifies that custom CA truststore is mounted and configured in JAVA_OPTS
func TestS3ProxyCustomCATrustStore(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml", "values/dummy_trust_store.yaml"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	container := deployment.Spec.Template.Spec.Containers[0]

	// Verify JAVA_OPTS contains truststore JVM arguments
	var javaOpts string
	for _, env := range container.Env {
		if env.Name == "JAVA_OPTS" {
			javaOpts = env.Value
			break
		}
	}
	require.NotEmpty(t, javaOpts, "JAVA_OPTS should be set")
	assert.Contains(t, javaOpts, "-Djavax.net.ssl.trustStore=/opt/s3proxy/secrets/java-cacerts", "JAVA_OPTS should contain trustStore path")
	assert.Contains(t, javaOpts, "-Djavax.net.ssl.trustStoreType=jks", "JAVA_OPTS should contain trustStoreType")
	assert.Contains(t, javaOpts, "-Djavax.net.ssl.trustStorePassword=$(JAVA_TRUSTSTORE_PASSWORD)", "JAVA_OPTS should contain trustStorePassword reference")

	// Verify JAVA_TRUSTSTORE_PASSWORD env var is set from the common secret
	var trustStorePasswordEnv *corev1.EnvVar
	for i := range container.Env {
		if container.Env[i].Name == "JAVA_TRUSTSTORE_PASSWORD" {
			trustStorePasswordEnv = &container.Env[i]
			break
		}
	}
	require.NotNil(t, trustStorePasswordEnv, "JAVA_TRUSTSTORE_PASSWORD env var should exist")
	assert.Equal(t, "suse-observability-common", trustStorePasswordEnv.ValueFrom.SecretKeyRef.Name, "Should reference the common secret")
	assert.Equal(t, "javaTrustStorePassword", trustStorePasswordEnv.ValueFrom.SecretKeyRef.Key, "Should reference the javaTrustStorePassword key")

	// Verify volume mount for common secrets
	var secretsMount *corev1.VolumeMount
	for i := range container.VolumeMounts {
		if container.VolumeMounts[i].Name == "common-secrets" {
			secretsMount = &container.VolumeMounts[i]
			break
		}
	}
	require.NotNil(t, secretsMount, "common-secrets volume mount should exist")
	assert.Equal(t, "/opt/s3proxy/secrets", secretsMount.MountPath, "Secrets should be mounted at /opt/s3proxy/secrets")
	assert.True(t, secretsMount.ReadOnly, "Secrets mount should be read-only")

	// Verify volume definition for common secrets
	var secretsVolume *corev1.Volume
	for i := range deployment.Spec.Template.Spec.Volumes {
		if deployment.Spec.Template.Spec.Volumes[i].Name == "common-secrets" {
			secretsVolume = &deployment.Spec.Template.Spec.Volumes[i]
			break
		}
	}
	require.NotNil(t, secretsVolume, "common-secrets volume should exist")
	require.NotNil(t, secretsVolume.Secret, "common-secrets volume should be a secret volume")
	assert.Equal(t, "suse-observability-common", secretsVolume.Secret.SecretName, "Should reference the common secret")
	require.Len(t, secretsVolume.Secret.Items, 1, "Should project exactly one item")
	assert.Equal(t, "javaTrustStore", secretsVolume.Secret.Items[0].Key, "Should project the javaTrustStore key")
	assert.Equal(t, "java-cacerts", secretsVolume.Secret.Items[0].Path, "Should project to java-cacerts path")

	// Verify common secret contains the truststore data
	commonSecret, ok := resources.Secrets["suse-observability-common"]
	require.True(t, ok, "Common secret should exist")
	_, hasTrustStore := commonSecret.Data["javaTrustStore"]
	assert.True(t, hasTrustStore, "Common secret should contain javaTrustStore")
	_, hasTrustStorePassword := commonSecret.Data["javaTrustStorePassword"]
	assert.True(t, hasTrustStorePassword, "Common secret should contain javaTrustStorePassword")

	// Verify checksum annotation for common secret
	annotations := deployment.Spec.Template.Annotations
	_, hasChecksumAnnotation := annotations["checksum/common-secret"]
	assert.True(t, hasChecksumAnnotation, "Deployment should have checksum annotation for common secret")
}

// TestS3ProxyCustomCATrustStoreBase64NoPassword verifies that base64-encoded truststore without password works
func TestS3ProxyCustomCATrustStoreBase64NoPassword(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"stackstate.java.trustStoreBase64Encoded": "c29tZSBkYXRh",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	container := deployment.Spec.Template.Spec.Containers[0]

	// Verify JAVA_OPTS contains truststore path but NOT password
	var javaOpts string
	for _, env := range container.Env {
		if env.Name == "JAVA_OPTS" {
			javaOpts = env.Value
			break
		}
	}
	require.NotEmpty(t, javaOpts, "JAVA_OPTS should be set")
	assert.Contains(t, javaOpts, "-Djavax.net.ssl.trustStore=/opt/s3proxy/secrets/java-cacerts", "JAVA_OPTS should contain trustStore path")
	assert.Contains(t, javaOpts, "-Djavax.net.ssl.trustStoreType=jks", "JAVA_OPTS should contain trustStoreType")
	assert.NotContains(t, javaOpts, "trustStorePassword", "JAVA_OPTS should NOT contain trustStorePassword when not set")

	// Verify JAVA_TRUSTSTORE_PASSWORD env var is NOT present
	for _, env := range container.Env {
		assert.NotEqual(t, "JAVA_TRUSTSTORE_PASSWORD", env.Name, "JAVA_TRUSTSTORE_PASSWORD should not be set when password is not configured")
	}

	// Verify volume mount and volume still exist
	var secretsMount *corev1.VolumeMount
	for i := range container.VolumeMounts {
		if container.VolumeMounts[i].Name == "common-secrets" {
			secretsMount = &container.VolumeMounts[i]
			break
		}
	}
	require.NotNil(t, secretsMount, "common-secrets volume mount should exist with base64-encoded truststore")

	var secretsVolume *corev1.Volume
	for i := range deployment.Spec.Template.Spec.Volumes {
		if deployment.Spec.Template.Spec.Volumes[i].Name == "common-secrets" {
			secretsVolume = &deployment.Spec.Template.Spec.Volumes[i]
			break
		}
	}
	require.NotNil(t, secretsVolume, "common-secrets volume should exist with base64-encoded truststore")
}

// TestS3ProxyNoCustomCA verifies that no truststore configuration is added when no custom CA is set
func TestS3ProxyNoCustomCA(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	container := deployment.Spec.Template.Spec.Containers[0]

	// Verify JAVA_OPTS does NOT contain truststore arguments
	var javaOpts string
	for _, env := range container.Env {
		if env.Name == "JAVA_OPTS" {
			javaOpts = env.Value
			break
		}
	}
	require.NotEmpty(t, javaOpts, "JAVA_OPTS should be set")
	assert.NotContains(t, javaOpts, "javax.net.ssl.trustStore", "JAVA_OPTS should NOT contain trustStore when no custom CA is set")

	// Verify no JAVA_TRUSTSTORE_PASSWORD env var
	for _, env := range container.Env {
		assert.NotEqual(t, "JAVA_TRUSTSTORE_PASSWORD", env.Name, "JAVA_TRUSTSTORE_PASSWORD should not be set when no custom CA is configured")
	}

	// Verify no common-secrets volume mount
	for _, vm := range container.VolumeMounts {
		assert.NotEqual(t, "common-secrets", vm.Name, "common-secrets volume mount should NOT exist when no custom CA is set")
	}

	// Verify no common-secrets volume
	for _, vol := range deployment.Spec.Template.Spec.Volumes {
		assert.NotEqual(t, "common-secrets", vol.Name, "common-secrets volume should NOT exist when no custom CA is set")
	}

	// Verify no checksum/common-secret annotation
	_, hasChecksumAnnotation := deployment.Spec.Template.Annotations["checksum/common-secret"]
	assert.False(t, hasChecksumAnnotation, "Should NOT have checksum annotation for common secret when no custom CA is set")
}

// TestS3ProxySecurityContextDefault verifies the default securityContext configuration
func TestS3ProxySecurityContextDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// Verify pod-level securityContext defaults
	podSecCtx := deployment.Spec.Template.Spec.SecurityContext
	require.NotNil(t, podSecCtx, "Pod securityContext should be set")
	assert.Equal(t, int64(65534), *podSecCtx.RunAsUser, "Pod runAsUser should be 65534")
	assert.Equal(t, int64(65534), *podSecCtx.RunAsGroup, "Pod runAsGroup should be 65534")
	assert.Equal(t, int64(65534), *podSecCtx.FSGroup, "Pod fsGroup should be 65534")
	assert.True(t, *podSecCtx.RunAsNonRoot, "Pod runAsNonRoot should be true")

	// Verify main container securityContext (from common.container defaults - no runAsUser/runAsGroup)
	container := deployment.Spec.Template.Spec.Containers[0]
	require.NotNil(t, container.SecurityContext, "Container securityContext should be set")
	assert.True(t, *container.SecurityContext.RunAsNonRoot, "Container runAsNonRoot should be true")
	assert.False(t, *container.SecurityContext.AllowPrivilegeEscalation, "Container allowPrivilegeEscalation should be false")
	require.NotNil(t, container.SecurityContext.Capabilities, "Container capabilities should be set")
	assert.Contains(t, container.SecurityContext.Capabilities.Drop, corev1.Capability("ALL"), "Container should drop ALL capabilities")
	require.NotNil(t, container.SecurityContext.SeccompProfile, "Container seccomp profile should be set")
	assert.Equal(t, corev1.SeccompProfileTypeRuntimeDefault, container.SecurityContext.SeccompProfile.Type, "Container seccomp profile should be RuntimeDefault")

	// Verify init container securityContext (from common.container defaults - no runAsUser/runAsGroup)
	require.Len(t, deployment.Spec.Template.Spec.InitContainers, 1, "Should have one init container")
	initContainer := deployment.Spec.Template.Spec.InitContainers[0]
	require.NotNil(t, initContainer.SecurityContext, "Init container securityContext should be set")
	assert.True(t, *initContainer.SecurityContext.RunAsNonRoot, "Init container runAsNonRoot should be true")
	assert.False(t, *initContainer.SecurityContext.AllowPrivilegeEscalation, "Init container allowPrivilegeEscalation should be false")
	require.NotNil(t, initContainer.SecurityContext.Capabilities, "Init container capabilities should be set")
	assert.Contains(t, initContainer.SecurityContext.Capabilities.Drop, corev1.Capability("ALL"), "Init container should drop ALL capabilities")
	require.NotNil(t, initContainer.SecurityContext.SeccompProfile, "Init container seccomp profile should be set")
	assert.Equal(t, corev1.SeccompProfileTypeRuntimeDefault, initContainer.SecurityContext.SeccompProfile.Type, "Init container seccomp profile should be RuntimeDefault")
}

// TestS3ProxySecurityContextDisabled verifies that pod-level securityContext can be disabled
func TestS3ProxySecurityContextDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.securityContext.enabled": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// Pod-level securityContext should be empty/nil when disabled
	podSecCtx := deployment.Spec.Template.Spec.SecurityContext
	if podSecCtx != nil {
		assert.Nil(t, podSecCtx.RunAsUser, "Pod runAsUser should not be set when securityContext is disabled")
		assert.Nil(t, podSecCtx.RunAsGroup, "Pod runAsGroup should not be set when securityContext is disabled")
		assert.Nil(t, podSecCtx.FSGroup, "Pod fsGroup should not be set when securityContext is disabled")
	}

	// Container-level securityContext should still be present (from common.container defaults)
	container := deployment.Spec.Template.Spec.Containers[0]
	require.NotNil(t, container.SecurityContext, "Container securityContext should still be set even when pod-level is disabled")
	assert.False(t, *container.SecurityContext.AllowPrivilegeEscalation, "Container allowPrivilegeEscalation should be false")
	assert.True(t, *container.SecurityContext.RunAsNonRoot, "Container runAsNonRoot should be true")
}

// TestS3ProxySecurityContextCustomValues verifies custom securityContext values
func TestS3ProxySecurityContextCustomValues(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.securityContext.enabled":    "true",
			"s3proxy.securityContext.runAsUser":  "1000",
			"s3proxy.securityContext.runAsGroup": "1000",
			"s3proxy.securityContext.fsGroup":    "1000",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	// Verify pod-level securityContext uses custom values
	podSecCtx := deployment.Spec.Template.Spec.SecurityContext
	require.NotNil(t, podSecCtx, "Pod securityContext should be set")
	assert.Equal(t, int64(1000), *podSecCtx.RunAsUser, "Pod runAsUser should be custom value 1000")
	assert.Equal(t, int64(1000), *podSecCtx.RunAsGroup, "Pod runAsGroup should be custom value 1000")
	assert.Equal(t, int64(1000), *podSecCtx.FSGroup, "Pod fsGroup should be custom value 1000")
}

// TestS3ProxyServiceAccountDefault verifies the default service account is created with backward-compatible name
func TestS3ProxyServiceAccountDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Service account should exist with backward-compatible name (same as old Minio subchart)
	sa, ok := resources.ServiceAccounts["suse-observability-minio"]
	require.True(t, ok, "S3Proxy service account should exist with backward-compatible name 'suse-observability-minio'")
	assert.Equal(t, "suse-observability-minio", sa.Name, "Service account name should match old Minio name for backward compatibility")

	// Deployment should reference the service account
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	assert.Equal(t, "suse-observability-minio", deployment.Spec.Template.Spec.ServiceAccountName, "Deployment should reference the backward-compatible service account name")
}

// TestS3ProxyServiceAccountCustomName verifies a custom service account name can be set
func TestS3ProxyServiceAccountCustomName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.serviceAccount.name": "my-custom-sa",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Service account should exist with custom name
	sa, ok := resources.ServiceAccounts["my-custom-sa"]
	require.True(t, ok, "S3Proxy service account should exist with custom name")
	assert.Equal(t, "my-custom-sa", sa.Name, "Service account name should match custom name")

	// Deployment should reference the custom service account
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	assert.Equal(t, "my-custom-sa", deployment.Spec.Template.Spec.ServiceAccountName, "Deployment should reference the custom service account name")
}

// TestS3ProxyServiceAccountAnnotations verifies annotations can be set on the service account (e.g. for IAM roles)
func TestS3ProxyServiceAccountAnnotations(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn": "arn:aws:iam::123456789012:role/s3-access-role",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	sa, ok := resources.ServiceAccounts["suse-observability-minio"]
	require.True(t, ok, "S3Proxy service account should exist")
	assert.Equal(t, "arn:aws:iam::123456789012:role/s3-access-role", sa.Annotations["eks.amazonaws.com/role-arn"], "Service account should have IAM role annotation")
}

// TestS3ProxyServiceAccountDisabled verifies the service account is not created when disabled
func TestS3ProxyServiceAccountDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.serviceAccount.create": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Service account should NOT exist
	_, ok := resources.ServiceAccounts["suse-observability-minio"]
	assert.False(t, ok, "S3Proxy service account should NOT exist when create is false")

	// Deployment should still reference the service account name (for pre-existing or externally managed SAs)
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	assert.Equal(t, "suse-observability-minio", deployment.Spec.Template.Spec.ServiceAccountName, "Deployment should still reference the service account name even when create is false")
}

// TestS3ProxyServiceAccountLegacyMinioName verifies backward compatibility with minio.serviceAccount.name
func TestS3ProxyServiceAccountLegacyMinioName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"minio.serviceAccount.name": "my-legacy-minio-sa",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Service account should exist with the legacy minio name
	sa, ok := resources.ServiceAccounts["my-legacy-minio-sa"]
	require.True(t, ok, "S3Proxy service account should use legacy minio.serviceAccount.name")
	assert.Equal(t, "my-legacy-minio-sa", sa.Name, "Service account name should match legacy minio value")

	// Deployment should reference the legacy name
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	assert.Equal(t, "my-legacy-minio-sa", deployment.Spec.Template.Spec.ServiceAccountName, "Deployment should reference the legacy minio service account name")
}

// TestS3ProxyServiceAccountS3ProxyOverridesLegacyMinio verifies that s3proxy.serviceAccount.name takes precedence over minio.serviceAccount.name
func TestS3ProxyServiceAccountS3ProxyOverridesLegacyMinio(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"minio.serviceAccount.name":   "my-legacy-minio-sa",
			"s3proxy.serviceAccount.name": "my-new-sa",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// s3proxy.serviceAccount.name should take precedence
	sa, ok := resources.ServiceAccounts["my-new-sa"]
	require.True(t, ok, "S3Proxy service account should use s3proxy.serviceAccount.name over legacy minio value")
	assert.Equal(t, "my-new-sa", sa.Name, "Service account name should match s3proxy value, not legacy minio value")

	// Legacy name should NOT exist
	_, ok = resources.ServiceAccounts["my-legacy-minio-sa"]
	assert.False(t, ok, "Legacy minio service account name should NOT be created when s3proxy.serviceAccount.name is set")

	// Deployment should reference the new name
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	assert.Equal(t, "my-new-sa", deployment.Spec.Template.Spec.ServiceAccountName, "Deployment should reference s3proxy service account name, not legacy minio")
}

// TestS3ProxyServiceAccountLegacyMinioCreateFalse verifies backward compatibility with minio.serviceAccount.create=false
func TestS3ProxyServiceAccountLegacyMinioCreateFalse(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"minio.serviceAccount.create": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Service account should NOT exist (legacy minio.serviceAccount.create=false should be respected)
	_, ok := resources.ServiceAccounts["suse-observability-minio"]
	assert.False(t, ok, "S3Proxy service account should NOT exist when legacy minio.serviceAccount.create is false")

	// Deployment should still reference the service account name
	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")
	assert.Equal(t, "suse-observability-minio", deployment.Spec.Template.Spec.ServiceAccountName, "Deployment should still reference the service account name even when create is false via legacy")
}

// TestS3ProxyServiceAccountCreateOverridesLegacyMinioCreate verifies s3proxy.serviceAccount.create=false takes precedence over minio.serviceAccount.create=true
func TestS3ProxyServiceAccountCreateOverridesLegacyMinioCreate(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"minio.serviceAccount.create":   "true",
			"s3proxy.serviceAccount.create": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// s3proxy.serviceAccount.create=false should override minio.serviceAccount.create=true
	_, ok := resources.ServiceAccounts["suse-observability-minio"]
	assert.False(t, ok, "S3Proxy service account should NOT exist when s3proxy.serviceAccount.create=false, even if minio.serviceAccount.create=true")
}

// TestS3ProxyServiceAccountLegacyMinioAnnotations verifies backward compatibility with minio.serviceAccount.annotations
func TestS3ProxyServiceAccountLegacyMinioAnnotations(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"minio.serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn": "arn:aws:iam::123456789012:role/legacy-minio-role",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	sa, ok := resources.ServiceAccounts["suse-observability-minio"]
	require.True(t, ok, "S3Proxy service account should exist")
	assert.Equal(t, "arn:aws:iam::123456789012:role/legacy-minio-role", sa.Annotations["eks.amazonaws.com/role-arn"], "Service account should have IAM role annotation from legacy minio values")
}

// TestS3ProxyServiceAccountAnnotationsMerge verifies that s3proxy and legacy minio annotations are merged correctly
func TestS3ProxyServiceAccountAnnotationsMerge(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"minio.serviceAccount.annotations.legacy-key":                       "legacy-value",
			"minio.serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn":   "arn:aws:iam::123456789012:role/legacy-role",
			"s3proxy.serviceAccount.annotations.new-key":                        "new-value",
			"s3proxy.serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn": "arn:aws:iam::123456789012:role/new-role",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	sa, ok := resources.ServiceAccounts["suse-observability-minio"]
	require.True(t, ok, "S3Proxy service account should exist")

	// Legacy-only key should be present (merged from minio)
	assert.Equal(t, "legacy-value", sa.Annotations["legacy-key"], "Legacy-only annotation should be present from minio values")

	// New-only key should be present (from s3proxy)
	assert.Equal(t, "new-value", sa.Annotations["new-key"], "New annotation should be present from s3proxy values")

	// Conflicting key should use s3proxy value (higher precedence)
	assert.Equal(t, "arn:aws:iam::123456789012:role/new-role", sa.Annotations["eks.amazonaws.com/role-arn"], "Conflicting annotation should use s3proxy value over legacy minio value")
}

// TestS3ProxyExtraEnvOpenDefault verifies no extra env vars are present by default
func TestS3ProxyExtraEnvOpenDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	mainContainer := deployment.Spec.Template.Spec.Containers[0]

	// Should not have any extra env vars beyond the defaults
	for _, env := range mainContainer.Env {
		assert.NotEqual(t, "MY_CUSTOM_VAR", env.Name, "Custom env var should not be present by default")
	}

	// Extra env secret should not exist
	_, ok = resources.Secrets["suse-observability-s3proxy-extra-env"]
	assert.False(t, ok, "Extra env secret should not exist by default")
}

// TestS3ProxyExtraEnvOpen verifies that extraEnv.open values are injected as plain env vars
func TestS3ProxyExtraEnvOpen(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.extraEnv.open.MY_CUSTOM_VAR":      "my-custom-value",
			"s3proxy.extraEnv.open.ANOTHER_CUSTOM_VAR": "another-value",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	mainContainer := deployment.Spec.Template.Spec.Containers[0]

	expectedMyCustomVar := corev1.EnvVar{Name: "MY_CUSTOM_VAR", Value: "my-custom-value"}
	expectedAnotherVar := corev1.EnvVar{Name: "ANOTHER_CUSTOM_VAR", Value: "another-value"}
	assert.Contains(t, mainContainer.Env, expectedMyCustomVar, "MY_CUSTOM_VAR should be present")
	assert.Contains(t, mainContainer.Env, expectedAnotherVar, "ANOTHER_CUSTOM_VAR should be present")
}

// TestS3ProxyExtraEnvSecret verifies that extraEnv.secret values are injected via secretKeyRef
func TestS3ProxyExtraEnvSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.extraEnv.secret.SECRET_VAR":         "secret-value",
			"s3proxy.extraEnv.secret.ANOTHER_SECRET_VAR": "another-secret",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	mainContainer := deployment.Spec.Template.Spec.Containers[0]

	// Verify env vars reference the secret
	expectedSecretVar := corev1.EnvVar{
		Name: "ANOTHER_SECRET_VAR",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "suse-observability-s3proxy-extra-env",
				},
				Key: "ANOTHER_SECRET_VAR",
			},
		},
	}
	expectedSecretVar2 := corev1.EnvVar{
		Name: "SECRET_VAR",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "suse-observability-s3proxy-extra-env",
				},
				Key: "SECRET_VAR",
			},
		},
	}
	assert.Contains(t, mainContainer.Env, expectedSecretVar, "ANOTHER_SECRET_VAR should reference secret")
	assert.Contains(t, mainContainer.Env, expectedSecretVar2, "SECRET_VAR should reference secret")

	// Verify the secret object exists with correct data
	secret, ok := resources.Secrets["suse-observability-s3proxy-extra-env"]
	require.True(t, ok, "Extra env secret should exist")
	assert.Contains(t, secret.Data, "SECRET_VAR", "Secret should contain SECRET_VAR key")
	assert.Contains(t, secret.Data, "ANOTHER_SECRET_VAR", "Secret should contain ANOTHER_SECRET_VAR key")
}

// TestS3ProxyExtraEnvOpenAndSecret verifies that both open and secret extra env vars work together
func TestS3ProxyExtraEnvOpenAndSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.extraEnv.open.OPEN_VAR":     "open-value",
			"s3proxy.extraEnv.secret.SECRET_VAR": "secret-value",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, ok := resources.Deployments["suse-observability-s3proxy"]
	require.True(t, ok, "S3Proxy deployment should exist")

	mainContainer := deployment.Spec.Template.Spec.Containers[0]

	// Verify open env var
	expectedOpenVar := corev1.EnvVar{Name: "OPEN_VAR", Value: "open-value"}
	assert.Contains(t, mainContainer.Env, expectedOpenVar, "OPEN_VAR should be present as plain env var")

	// Verify secret env var
	expectedSecretVar := corev1.EnvVar{
		Name: "SECRET_VAR",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "suse-observability-s3proxy-extra-env",
				},
				Key: "SECRET_VAR",
			},
		},
	}
	assert.Contains(t, mainContainer.Env, expectedSecretVar, "SECRET_VAR should reference secret")

	// Verify both secret and deployment exist
	_, ok = resources.Secrets["suse-observability-s3proxy-extra-env"]
	require.True(t, ok, "Extra env secret should exist when secret env vars are set")
}

// TestS3ProxyCredentialsChecksumAnnotation verifies that victoria-metrics and clickhouse
// pods include a checksum annotation for s3proxy credentials when backup is enabled,
// so they restart when credentials change.
func TestS3ProxyCredentialsChecksumAnnotation(t *testing.T) {
	outputEnabled := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled": "true",
		},
	})
	resourcesEnabled := helmtestutil.NewKubernetesResources(t, outputEnabled)

	// With backup enabled: Victoria Metrics 0 should have the checksum annotation
	vm0, ok := resourcesEnabled.Statefulsets["suse-observability-victoria-metrics-0"]
	require.True(t, ok, "Victoria Metrics 0 StatefulSet should exist")
	assert.Contains(t, vm0.Spec.Template.Annotations, "checksum/s3proxy-credentials",
		"Victoria Metrics 0 pods should have checksum/s3proxy-credentials annotation when backup enabled")

	// With backup enabled: Victoria Metrics 1 should have the checksum annotation
	vm1, ok := resourcesEnabled.Statefulsets["suse-observability-victoria-metrics-1"]
	require.True(t, ok, "Victoria Metrics 1 StatefulSet should exist")
	assert.Contains(t, vm1.Spec.Template.Annotations, "checksum/s3proxy-credentials",
		"Victoria Metrics 1 pods should have checksum/s3proxy-credentials annotation when backup enabled")

	// With backup enabled: Clickhouse should have the checksum annotation
	ch, ok := resourcesEnabled.Statefulsets["suse-observability-clickhouse-shard0"]
	require.True(t, ok, "ClickHouse StatefulSet should exist")
	assert.Contains(t, ch.Spec.Template.Annotations, "checksum/s3proxy-credentials",
		"ClickHouse pods should have checksum/s3proxy-credentials annotation when backup enabled")

	outputDisabled := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled": "false",
		},
	})
	resourcesDisabled := helmtestutil.NewKubernetesResources(t, outputDisabled)

	// With backup disabled: Victoria Metrics 0 should NOT have the checksum annotation
	vm0Disabled, ok := resourcesDisabled.Statefulsets["suse-observability-victoria-metrics-0"]
	require.True(t, ok, "Victoria Metrics 0 StatefulSet should exist")
	assert.NotContains(t, vm0Disabled.Spec.Template.Annotations, "checksum/s3proxy-credentials",
		"Victoria Metrics 0 pods should NOT have checksum/s3proxy-credentials annotation when backup disabled")

	// With backup disabled: Clickhouse should NOT have the checksum annotation
	chDisabled, ok := resourcesDisabled.Statefulsets["suse-observability-clickhouse-shard0"]
	require.True(t, ok, "ClickHouse StatefulSet should exist")
	assert.NotContains(t, chDisabled.Spec.Template.Annotations, "checksum/s3proxy-credentials",
		"ClickHouse pods should NOT have checksum/s3proxy-credentials annotation when backup disabled")
}
