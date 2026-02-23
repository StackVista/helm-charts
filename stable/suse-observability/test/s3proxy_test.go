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
	assert.Contains(t, configMap.Data["s3proxy-main.properties"], "aws-s3-sdk.region=eu-west-1", "ConfigMap should have correct region")
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
	assert.Contains(t, configMap.Data["s3proxy-settings.properties"], "s3proxy.bucket-locator.1=settings-local-backup",
		"Settings properties should route settings-local-backup bucket")

	// Main properties should have all backup buckets
	mainProps := configMap.Data["s3proxy-main.properties"]
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.1=sts-configuration-backup", "Main properties should route configuration backup bucket")
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.2=sts-stackgraph-backup", "Main properties should route stackgraph backup bucket")
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.3=sts-elasticsearch-backup", "Main properties should route elasticsearch backup bucket")
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.4=sts-victoria-metrics-backup", "Main properties should route victoria-metrics-0 backup bucket")
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.5=sts-victoria-metrics-backup", "Main properties should route victoria-metrics-1 backup bucket")
	assert.Contains(t, mainProps, "s3proxy.bucket-locator.6=sts-clickhouse-backup", "Main properties should route clickhouse backup bucket")
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

		deployment := resources.Deployments["suse-observability-s3proxy"]
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

		deployment := resources.Deployments["suse-observability-s3proxy"]
		container := deployment.Spec.Template.Spec.Containers[0]

		// Should have both settings and main properties
		assert.Contains(t, container.Args, "/etc/s3proxy/s3proxy-settings.properties", "Should have settings properties")
		assert.Contains(t, container.Args, "/etc/s3proxy/s3proxy-main.properties", "Should have main properties when backup enabled")
	})
}

// TestS3ProxyExistingSecret verifies using an existing secret
func TestS3ProxyExistingSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"backup.storage.credentials.existingSecret": "my-existing-secret",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	// S3Proxy managed secret should NOT exist
	_, ok := resources.Secrets["suse-observability-s3proxy"]
	assert.False(t, ok, "S3Proxy managed secret should NOT exist when using existingSecret")

	// Deployment should reference the existing secret
	deployment := resources.Deployments["suse-observability-s3proxy"]
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

// TestS3ProxyNodeSelector verifies node selector configuration
func TestS3ProxyNodeSelector(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.nodeSelector.disk-type": "ssd",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment := resources.Deployments["suse-observability-s3proxy"]
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

	deployment := resources.Deployments["suse-observability-s3proxy"]
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
	assert.Contains(t, javaOpts, "-DLOG_LEVEL", "JAVA_OPTS should contain other s3proxy options")
}

// TestS3ProxyJavaHeapCustomValues verifies that JVM heap is calculated correctly with custom values
func TestS3ProxyJavaHeapCustomValues(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"s3proxy.resources.limits.memory":       "1Gi",
			"s3proxy.resources.requests.memory":     "512Mi",
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
