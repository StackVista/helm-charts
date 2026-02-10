package test

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestDeprecatedBackupEnableShouldFail(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/deprecated_backup_enabled.yaml")
	require.Contains(t, err.Error(), "Please use `global.backup.enabled=true` instead of `backup.enabled=true`")
}

var backupAlwaysEnabledCronjobs = []string{
	"suse-observability-backup-conf",
}

var backupManuallyEnabledCronjobs = []string{
	"suse-observability-backup-sg",
	"suse-observability-backup-init",
	"suse-observability-clickhouse-full-backup",
	"suse-observability-clickhouse-incremental-backup",
}

var backupManuallyEnabledDeployments = []string{
	"suse-observability-minio",
}

var backupAlwaysPresentSecrets = []string{
	"suse-observability-backup-config",
}

var backupAlwaysEnabledConfigMaps = []string{
	"suse-observability-backup-config",
	"suse-observability-backup-restore-scripts",
	"suse-observability-clickhouse-backup",
	"suse-observability-clickhouse-backup-scripts",
}

var backupAlwaysEnabledPVCs = []string{
	"suse-observability-settings-backup-data",
}

var backupManuallyEnabledPVCs = []string{
	"suse-observability-backup-stackgraph-tmp-data",
}

func TestGlobalBackupEnabledEnsureResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"global.backup.enabled": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	expectedCronjobs := append(backupAlwaysEnabledCronjobs, backupManuallyEnabledCronjobs...)
	for _, component := range expectedCronjobs {
		_, ok := resources.CronJobs[component]
		assert.Equal(t, true, ok, "%s Cronjob should exist", component)
	}

	for _, component := range backupManuallyEnabledDeployments {
		_, ok := resources.Deployments[component]
		assert.Equal(t, true, ok, "%s Deployment should exist", component)
	}

	for _, component := range backupAlwaysPresentSecrets {
		_, ok := resources.Secrets[component]
		assert.Equal(t, true, ok, "%s Secret should exist", component)
	}

	for _, component := range backupAlwaysEnabledConfigMaps {
		_, ok := resources.ConfigMaps[component]
		assert.Equal(t, true, ok, "%s ConfigMap should exist", component)
	}

	expectedPVCs := append(backupAlwaysEnabledPVCs, backupManuallyEnabledPVCs...)
	for _, component := range expectedPVCs {
		_, ok := resources.PersistentVolumeClaims[component]
		assert.Equal(t, true, ok, "%s PersistentVolumeClaim should exist", component)
	}
}

func TestGlobalBackupDisabledEnsureResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"global.backup.enabled": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	for _, component := range backupAlwaysEnabledCronjobs {
		_, ok := resources.CronJobs[component]
		assert.Equal(t, true, ok, "%s Cronjob should exist", component)
	}

	for _, component := range backupAlwaysEnabledCronjobs {
		_, ok := resources.CronJobs[component]
		assert.Equal(t, true, ok, "%s Cronjob should not exist", component)
	}

	for _, component := range backupManuallyEnabledDeployments {
		_, ok := resources.Deployments[component]
		assert.Equal(t, false, ok, "%s Deployment should not exist", component)
	}

	for _, component := range backupAlwaysPresentSecrets {
		_, ok := resources.Secrets[component]
		assert.Equal(t, true, ok, "%s Secret should exist", component)
	}

	for _, component := range backupAlwaysEnabledConfigMaps {
		_, ok := resources.ConfigMaps[component]
		assert.Equal(t, true, ok, "%s ConfigMap should exist", component)
	}

	for _, component := range backupAlwaysEnabledPVCs {
		_, ok := resources.PersistentVolumeClaims[component]
		assert.Equal(t, true, ok, "%s PersistentVolumeClaim should exist", component)
	}

	for _, component := range backupManuallyEnabledPVCs {
		_, ok := resources.PersistentVolumeClaims[component]
		assert.Equal(t, false, ok, "%s PersistentVolumeClaim should not exist", component)
	}
}

func TestBackupConfigmapFull(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/backup_config_full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	backupConfigMap, ok := resources.ConfigMaps["suse-observability-backup-config"]
	require.True(t, ok, "ConfigMap 'suse-observability-backup-config' should exist")

	configData, ok := backupConfigMap.Data["config"]
	require.True(t, ok, "ConfigMap should have 'config' key")

	expectedConfig, err := os.ReadFile("test_data/backup-config-full.yaml")
	require.NoError(t, err, "Should be able to read expected config file")

	assert.Equal(t, string(expectedConfig), configData, "ConfigMap 'config' data should match expected backup-config-full.yaml")
}

func TestBackupConfigmapDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/backup_config_default.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	backupConfigMap, ok := resources.ConfigMaps["suse-observability-backup-config"]
	require.True(t, ok, "ConfigMap 'suse-observability-backup-config' should exist")

	configData, ok := backupConfigMap.Data["config"]
	require.True(t, ok, "ConfigMap should have 'config' key")

	expectedConfig, err := os.ReadFile("test_data/backup-config-default.yaml")
	require.NoError(t, err, "Should be able to read expected config file")

	assert.Equal(t, string(expectedConfig), configData, "ConfigMap 'config' data should match expected backup-config-default.yaml")
}
