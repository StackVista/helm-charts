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

var backupManuallyEnabledDeployments = []string{}

// S3Proxy renders whenever settings backup is enabled (default true), so with default
// values it is present even when global.backup.enabled=false. It is NOT truly "always"
// enabled — see TestAllBackupsDisabledNoS3Proxy for the case where it disappears.
var backupAlwaysEnabledDeployments = []string{
	"suse-observability-s3proxy",
}

// S3Proxy-related resources present under the default values (settings backup enabled).
var backupAlwaysEnabledSecrets = []string{
	"suse-observability-s3proxy",
}

var s3proxyAlwaysEnabledConfigMaps = []string{
	"suse-observability-s3proxy-config",
}

var s3proxyAlwaysEnabledPVCs = []string{
	"suse-observability-backup-settings-data",
}

// S3Proxy-related resources that are only present when global.backup.enabled=true
var s3proxyManuallyEnabledPVCs = []string{
	"suse-observability-minio",
}

var backupAlwaysPresentSecrets = []string{
	"suse-observability-backup-config",
}

var backupAlwaysEnabledConfigMaps = []string{
	"suse-observability-sts-backup-conf",
	"suse-observability-backup-config",
	"suse-observability-backup-restore-scripts",
	"suse-observability-clickhouse-backup",
	"suse-observability-clickhouse-backup-scripts",
}

var backupAlwaysEnabledPVCs = []string{}

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

	for _, component := range backupAlwaysEnabledDeployments {
		_, ok := resources.Deployments[component]
		assert.Equal(t, true, ok, "%s Deployment should exist", component)
	}

	for _, component := range backupAlwaysEnabledSecrets {
		_, ok := resources.Secrets[component]
		assert.Equal(t, true, ok, "%s Secret should exist", component)
	}

	for _, component := range s3proxyAlwaysEnabledConfigMaps {
		_, ok := resources.ConfigMaps[component]
		assert.Equal(t, true, ok, "%s ConfigMap should exist", component)
	}

	for _, component := range s3proxyAlwaysEnabledPVCs {
		_, ok := resources.PersistentVolumeClaims[component]
		assert.Equal(t, true, ok, "%s PersistentVolumeClaim should exist", component)
	}

	for _, component := range s3proxyManuallyEnabledPVCs {
		_, ok := resources.PersistentVolumeClaims[component]
		assert.Equal(t, true, ok, "%s PersistentVolumeClaim should exist", component)
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

	for _, component := range backupManuallyEnabledCronjobs {
		_, ok := resources.CronJobs[component]
		assert.Equal(t, false, ok, "%s Cronjob should not exist", component)
	}

	// S3Proxy deployment is always enabled for settings backup
	for _, component := range backupAlwaysEnabledDeployments {
		_, ok := resources.Deployments[component]
		assert.Equal(t, true, ok, "%s Deployment should exist", component)
	}

	for _, component := range backupManuallyEnabledDeployments {
		_, ok := resources.Deployments[component]
		assert.Equal(t, false, ok, "%s Deployment should not exist", component)
	}

	for _, component := range backupAlwaysPresentSecrets {
		_, ok := resources.Secrets[component]
		assert.Equal(t, true, ok, "%s Secret should exist", component)
	}

	for _, component := range backupAlwaysEnabledSecrets {
		_, ok := resources.Secrets[component]
		assert.Equal(t, true, ok, "%s Secret should exist", component)
	}

	for _, component := range backupAlwaysEnabledConfigMaps {
		_, ok := resources.ConfigMaps[component]
		assert.Equal(t, true, ok, "%s ConfigMap should exist", component)
	}

	for _, component := range s3proxyAlwaysEnabledConfigMaps {
		_, ok := resources.ConfigMaps[component]
		assert.Equal(t, true, ok, "%s ConfigMap should exist", component)
	}

	for _, component := range backupAlwaysEnabledPVCs {
		_, ok := resources.PersistentVolumeClaims[component]
		assert.Equal(t, true, ok, "%s PersistentVolumeClaim should exist", component)
	}

	for _, component := range s3proxyAlwaysEnabledPVCs {
		_, ok := resources.PersistentVolumeClaims[component]
		assert.Equal(t, true, ok, "%s PersistentVolumeClaim should exist", component)
	}

	for _, component := range backupManuallyEnabledPVCs {
		_, ok := resources.PersistentVolumeClaims[component]
		assert.Equal(t, false, ok, "%s PersistentVolumeClaim should not exist", component)
	}

	for _, component := range s3proxyManuallyEnabledPVCs {
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

func TestBackupStackpacksServiceSplitEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":            "true",
			"stackstate.features.server.split": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	service, ok := resources.Services["suse-observability-backup-stackpacks"]
	require.True(t, ok, "backup-stackpacks Service should exist")
	assert.Equal(t, "None", service.Spec.ClusterIP, "backup-stackpacks Service should be headless")
	require.Len(t, service.Spec.Ports, 1)
	assert.Equal(t, int32(7090), service.Spec.Ports[0].Port)
	assert.Equal(t, "api", service.Spec.Selector["app.kubernetes.io/component"], "should select api component when server.split is enabled")
}

func TestBackupStackpacksServiceSplitDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":            "true",
			"stackstate.features.server.split": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	service, ok := resources.Services["suse-observability-backup-stackpacks"]
	require.True(t, ok, "backup-stackpacks Service should exist")
	assert.Equal(t, "server", service.Spec.Selector["app.kubernetes.io/component"], "should select server component when server.split is disabled")
}

func TestBackupStackpacksEnvVars(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	expectedEnvVars := []corev1.EnvVar{
		{Name: "BACKUP_STACKPACKS_SERVICE_URL", Value: "http://suse-observability-backup-stackpacks:7090"},
		{Name: "BACKUP_STACKPACKS_DIR", Value: "stackpacks/"},
	}

	// Verify env vars on stackgraph backup cronjob
	sgCronjob, ok := resources.CronJobs["suse-observability-backup-sg"]
	require.True(t, ok, "backup-sg CronJob should exist")
	sgEnv := sgCronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env
	for _, expected := range expectedEnvVars {
		assert.Contains(t, sgEnv, expected, "backup-sg should have env var %s", expected.Name)
	}

	// Verify env vars on configuration backup cronjob
	confCronjob, ok := resources.CronJobs["suse-observability-backup-conf"]
	require.True(t, ok, "backup-conf CronJob should exist")
	confEnv := confCronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env
	for _, expected := range expectedEnvVars {
		assert.Contains(t, confEnv, expected, "backup-conf should have env var %s", expected.Name)
	}
}

func TestBackupStackpacksVolumeMounts(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	for _, cronjobName := range []string{"suse-observability-backup-sg", "suse-observability-backup-conf"} {
		cronjob, ok := resources.CronJobs[cronjobName]
		require.True(t, ok, "%s CronJob should exist", cronjobName)

		volumeMounts := cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts
		var hasConfigVolume bool
		for _, vm := range volumeMounts {
			if vm.Name == "config-volume" && vm.SubPath == "application_stackstate.conf" {
				hasConfigVolume = true
				break
			}
		}
		assert.True(t, hasConfigVolume, "%s should mount config-volume with application_stackstate.conf subPath", cronjobName)

		volumes := cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes
		var hasConfigVolumeSource bool
		for _, v := range volumes {
			if v.Name == "config-volume" && v.ConfigMap != nil && strings.HasSuffix(v.ConfigMap.Name, "-sts-backup-conf") {
				hasConfigVolumeSource = true
				break
			}
		}
		assert.True(t, hasConfigVolumeSource, "%s should have config-volume sourced from sts-backup-conf ConfigMap", cronjobName)
	}
}

func TestBackupConfigmapMonoMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/backup_config_default.yaml",
			"values/mono_hbase.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	backupConfigMap, ok := resources.ConfigMaps["suse-observability-backup-config"]
	require.True(t, ok, "ConfigMap 'suse-observability-backup-config' should exist")

	configData, ok := backupConfigMap.Data["config"]
	require.True(t, ok, "ConfigMap should have 'config' key")

	var config map[string]interface{}
	err := yaml.Unmarshal([]byte(configData), &config)
	require.NoError(t, err, "config data should be valid YAML")

	stackpacks, ok := config["stackpacks"].(map[string]interface{})
	require.True(t, ok, "config should have 'stackpacks' section")

	assert.Equal(t, "file:///var/stackpacks_local", stackpacks["localStackPacksUri"],
		"localStackPacksUri should use file:// scheme in Mono mode")
	assert.Equal(t, "suse-observability-stackpacks-local", stackpacks["pvc"],
		"stackpacks.pvc should reference the local stackpacks PVC in Mono mode")
}

func TestBackupConfigmapDistributedModeNoPvc(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles:    []string{"values/backup_config_default.yaml"},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	backupConfigMap, ok := resources.ConfigMaps["suse-observability-backup-config"]
	require.True(t, ok, "ConfigMap 'suse-observability-backup-config' should exist")

	configData, ok := backupConfigMap.Data["config"]
	require.True(t, ok, "ConfigMap should have 'config' key")

	var config map[string]interface{}
	err := yaml.Unmarshal([]byte(configData), &config)
	require.NoError(t, err, "config data should be valid YAML")

	stackpacks, ok := config["stackpacks"].(map[string]interface{})
	require.True(t, ok, "config should have 'stackpacks' section")

	assert.Equal(t, "hdfs://suse-observability-hbase-hdfs-nn-headful:9000/stackpacks", stackpacks["localStackPacksUri"],
		"localStackPacksUri should use hdfs:// scheme in Distributed mode")
	assert.Nil(t, stackpacks["pvc"], "stackpacks.pvc should not be present in Distributed mode")
}

// TestAllBackupsDisabledNoS3Proxy verifies that when both global.backup.enabled and
// backup.configuration.enabled (settings backup) are false, no backup component
// is enabled and s3proxy (with all its dependent resources) does not render at all.
func TestAllBackupsDisabledNoS3Proxy(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":        "false",
			"backup.configuration.enabled": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	_, ok := resources.Deployments["suse-observability-s3proxy"]
	assert.False(t, ok, "S3Proxy deployment should NOT exist when all backups are disabled")

	_, ok = resources.Services["suse-observability-s3proxy"]
	assert.False(t, ok, "S3Proxy service should NOT exist when all backups are disabled")

	_, ok = resources.Secrets["suse-observability-s3proxy"]
	assert.False(t, ok, "S3Proxy secret should NOT exist when all backups are disabled")

	_, ok = resources.ConfigMaps["suse-observability-s3proxy-config"]
	assert.False(t, ok, "S3Proxy ConfigMap should NOT exist when all backups are disabled")

	_, ok = resources.PersistentVolumeClaims["suse-observability-backup-settings-data"]
	assert.False(t, ok, "Settings backup data PVC should NOT exist when all backups are disabled")

	_, ok = resources.PersistentVolumeClaims["suse-observability-minio"]
	assert.False(t, ok, "Main data PVC should NOT exist when all backups are disabled")

	for _, cronjob := range []string{
		"suse-observability-backup-sg",
		"suse-observability-backup-conf",
		"suse-observability-clickhouse-full-backup",
		"suse-observability-clickhouse-incremental-backup",
	} {
		_, ok = resources.CronJobs[cronjob]
		assert.False(t, ok, "%s CronJob should NOT exist when all backups are disabled", cronjob)
	}
}

// TestSettingsOnlyBackupRendersS3Proxy verifies that settings/configuration backup alone
// (global.backup.enabled=false, backup.configuration.enabled=true) is enough to
// keep s3proxy and the settings backup CronJob rendering, while remote-only backups stay off.
func TestSettingsOnlyBackupRendersS3Proxy(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":        "false",
			"backup.configuration.enabled": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	_, ok := resources.Deployments["suse-observability-s3proxy"]
	assert.True(t, ok, "S3Proxy deployment should exist when settings backup alone is enabled")

	_, ok = resources.CronJobs["suse-observability-backup-conf"]
	assert.True(t, ok, "backup-conf CronJob should exist when settings backup alone is enabled")

	_, ok = resources.CronJobs["suse-observability-backup-sg"]
	assert.False(t, ok, "backup-sg CronJob should NOT exist when global.backup.enabled=false")

	for _, cronjob := range []string{
		"suse-observability-clickhouse-full-backup",
		"suse-observability-clickhouse-incremental-backup",
	} {
		_, ok = resources.CronJobs[cronjob]
		assert.False(t, ok, "%s CronJob should NOT exist when global.backup.enabled=false", cronjob)
	}

	_, ok = resources.PersistentVolumeClaims["suse-observability-minio"]
	assert.False(t, ok, "Main data PVC should NOT exist when global.backup.enabled=false")
}

// TestStackGraphBackupDisabled verifies that disabling backup.stackGraph.enabled
// removes only the stackgraph backup CronJob, leaving settings backup and s3proxy intact.
func TestStackGraphBackupDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":     "true",
			"backup.stackGraph.enabled": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	_, ok := resources.CronJobs["suse-observability-backup-sg"]
	assert.False(t, ok, "backup-sg CronJob should NOT exist when backup.stackGraph.enabled=false")

	_, ok = resources.CronJobs["suse-observability-backup-conf"]
	assert.True(t, ok, "backup-conf CronJob should exist")

	_, ok = resources.Deployments["suse-observability-s3proxy"]
	assert.True(t, ok, "S3Proxy deployment should exist")
}

// TestClickHouseBackupDisabled verifies that disabling clickhouse.backup.enabled removes
// only the ClickHouse backup CronJobs, leaving other backup components and s3proxy intact.
func TestClickHouseBackupDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":     "true",
			"clickhouse.backup.enabled": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	for _, cronjob := range []string{
		"suse-observability-clickhouse-full-backup",
		"suse-observability-clickhouse-incremental-backup",
	} {
		_, ok := resources.CronJobs[cronjob]
		assert.False(t, ok, "%s CronJob should NOT exist when clickhouse.backup.enabled=false", cronjob)
	}

	_, ok := resources.Deployments["suse-observability-s3proxy"]
	assert.True(t, ok, "S3Proxy deployment should exist")

	_, ok = resources.CronJobs["suse-observability-backup-sg"]
	assert.True(t, ok, "backup-sg CronJob should exist")
}

// TestVictoriaMetricsBackupDisabledPerInstance verifies that disabling
// victoria-metrics-0.backup.enabled removes only that instance's vmbackup sidecar container
// from its StatefulSet, leaving victoria-metrics-1's vmbackup sidecar in place.
func TestVictoriaMetricsBackupDisabledPerInstance(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.backup.enabled":             "true",
			"victoria-metrics-0.backup.enabled": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "suse-observability"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	hasContainer := func(sts *corev1.PodSpec, name string) bool {
		for _, c := range sts.Containers {
			if c.Name == name {
				return true
			}
		}
		return false
	}

	vm0, ok := resources.Statefulsets["suse-observability-victoria-metrics-0"]
	require.True(t, ok, "victoria-metrics-0 StatefulSet should exist")
	assert.False(t, hasContainer(&vm0.Spec.Template.Spec, "vmbackup"), "victoria-metrics-0 should NOT have a vmbackup container when its backup is disabled")

	vm1, ok := resources.Statefulsets["suse-observability-victoria-metrics-1"]
	require.True(t, ok, "victoria-metrics-1 StatefulSet should exist")
	assert.True(t, hasContainer(&vm1.Spec.Template.Spec, "vmbackup"), "victoria-metrics-1 should still have a vmbackup container")
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
