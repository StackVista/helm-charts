package test

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
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

func testJobsFromBackupRestoreScriptsConfigMap(t *testing.T, resources *helmtestutil.KubernetesResources) map[string]batchv1.Job {
	// Find the backup-restore-scripts ConfigMap
	var backupConfigMap *corev1.ConfigMap
	for _, cm := range resources.ConfigMaps {
		if strings.HasSuffix(cm.Name, "-backup-restore-scripts") {
			backupConfigMap = &cm
			break
		}
	}
	require.NotNil(t, backupConfigMap, "backup-restore-scripts ConfigMap should exist")

	// List of expected Job template keys in the ConfigMap
	expectedJobTemplateKeys := []string{
		"job-elasticsearch-list-snapshots.yaml",
		"job-elasticsearch-restore-snapshot.yaml",
		"job-stackgraph-list-backups.yaml",
		"job-stackgraph-restore-backup.yaml",
		"job-configuration-list-backups.yaml",
		"job-configuration-restore-backup.yaml",
		"job-configuration-download-backup.yaml",
		"job-configuration-upload-backup.yaml",
		"job-victoria-metrics-list-backups.yaml",
		"job-victoria-metrics-restore-backup.yaml",
	}

	backupJobs := map[string]batchv1.Job{}
	// Test each Job template in the ConfigMap data
	for _, jobTemplateKey := range expectedJobTemplateKeys {
		jobYaml, exists := backupConfigMap.Data[jobTemplateKey]
		require.True(t, exists, "ConfigMap should have %s key", jobTemplateKey)

		// Unmarshal the YAML into a Job struct
		var job batchv1.Job
		err := yaml.Unmarshal([]byte(jobYaml), &job)
		assert.NoError(t, err, "Job template '%s' should be valid YAML", jobTemplateKey)
		backupJobs[jobTemplateKey] = job
	}
	return backupJobs
}
