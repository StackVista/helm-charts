package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"sigs.k8s.io/yaml"
)

// expectedImagePullSecretsName - the ImagePullSecrets all workloads should have
var expectedImagePullSecretsName = "suse-observability-pull-secret"

func TestGlobalPullSecretHAMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_suse_observability_pull_secret_ha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	testGlobalPullSecretOnAllWorkloads(t, &resources, expectedImagePullSecretsName, "HA")
}

func TestGlobalPullSecretNonHAMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_suse_observability_pull_secret_nonha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	testGlobalPullSecretOnAllWorkloads(t, &resources, expectedImagePullSecretsName, "Non-HA")
}

func testGlobalPullSecretOnAllWorkloads(t *testing.T, resources *helmtestutil.KubernetesResources, expectedImagePullSecretsName string, mode string) {
	for _, deployment := range resources.Deployments {
		t.Run("Deployment_"+deployment.Name, func(t *testing.T) {
			testDeploymentGlobalPullSecret(t, deployment, expectedImagePullSecretsName, mode)
		})
	}

	for _, statefulset := range resources.Statefulsets {
		t.Run("Statefulset_"+statefulset.Name, func(t *testing.T) {
			testStatefulSetGlobalPullSecret(t, statefulset, expectedImagePullSecretsName, mode)
		})
	}

	for _, job := range resources.Jobs {
		t.Run("Job_"+job.Name, func(t *testing.T) {
			testJobGlobalPullSecret(t, job, expectedImagePullSecretsName, mode)
		})
	}

	for _, cronjob := range resources.CronJobs {
		t.Run("Cronjob_"+cronjob.Name, func(t *testing.T) {
			testCronJobGlobalPullSecret(t, cronjob, expectedImagePullSecretsName, mode)
		})
	}

	testBackupConfigMapJobTemplatesGlobalPullSecret(t, resources, expectedImagePullSecretsName, mode)
}

func testDeploymentGlobalPullSecret(t *testing.T, deployment appsv1.Deployment, expectedImagePullSecretsName string, mode string) {
	assert.NotNil(t, deployment.Spec.Template.Spec.ImagePullSecrets, "%s mode: Deployment '%s' should have ImagePullSecrets", mode, deployment.Name)
	require.Len(t, deployment.Spec.Template.Spec.ImagePullSecrets, 1, "%s mode: Deployment '%s' should have exactly one ImagePullSecret", mode, deployment.Name)
	require.Equal(t, expectedImagePullSecretsName, deployment.Spec.Template.Spec.ImagePullSecrets[0].Name, "%s mode: Deployment '%s' should have ImagePullSecret named '%s'", mode, deployment.Name, expectedImagePullSecretsName)
}

func testStatefulSetGlobalPullSecret(t *testing.T, statefulset appsv1.StatefulSet, eexpectedImagePullSecretsName string, mode string) {
	assert.NotNil(t, statefulset.Spec.Template.Spec.ImagePullSecrets, "%s mode: StatefulSet '%s' should have ImagePullSecrets", mode, statefulset.Name)
	require.Len(t, statefulset.Spec.Template.Spec.ImagePullSecrets, 1, "%s mode: StatefulSet '%s' should have exactly one ImagePullSecret", mode, statefulset.Name)
	require.Equal(t, expectedImagePullSecretsName, statefulset.Spec.Template.Spec.ImagePullSecrets[0].Name, "%s mode: StatefulSet '%s' should have ImagePullSecret named '%s'", mode, statefulset.Name, expectedImagePullSecretsName)
}

func testJobGlobalPullSecret(t *testing.T, job batchv1.Job, expectedImagePullSecretsName string, mode string) {
	assert.NotNil(t, job.Spec.Template.Spec.ImagePullSecrets, "%s mode: Job '%s' should have ImagePullSecrets", mode, job.Name)
	require.Len(t, job.Spec.Template.Spec.ImagePullSecrets, 1, "%s mode: Job '%s' should have exactly one ImagePullSecret", mode, job.Name)
	require.Equal(t, expectedImagePullSecretsName, job.Spec.Template.Spec.ImagePullSecrets[0].Name, "%s mode: Job '%s' should have ImagePullSecret named '%s'", mode, job.Name, expectedImagePullSecretsName)
}

func testCronJobGlobalPullSecret(t *testing.T, cronjob batchv1beta1.CronJob, expectedImagePullSecretsName string, mode string) {
	assert.NotNil(t, cronjob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets, "%s mode: Job '%s' should have ImagePullSecrets", mode, cronjob.Name)
	require.Len(t, cronjob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets, 1, "%s mode: Job '%s' should have exactly one ImagePullSecret", mode, cronjob.Name)
	require.Equal(t, expectedImagePullSecretsName, cronjob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets[0].Name, "%s mode: Job '%s' should have ImagePullSecret named '%s'", mode, cronjob.Name, expectedImagePullSecretsName)
}

func testBackupConfigMapJobTemplatesGlobalPullSecret(t *testing.T, resources *helmtestutil.KubernetesResources, expectedImagePullSecretsName string, mode string) {

	backupConfigMapName := "suse-observability-backup-restore-scripts"

	backupConfigMap, ok := resources.ConfigMaps[backupConfigMapName]
	require.True(t, ok, "%s mode: ConfigMap '%s' should exist", mode, backupConfigMapName)

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

	for _, jobTemplateKey := range expectedJobTemplateKeys {
		jobYaml, ok := backupConfigMap.Data[jobTemplateKey]
		assert.True(t, ok, "%s mode: JobTemplate '%s' should exist", mode, jobTemplateKey)

		var job batchv1.Job
		err := yaml.Unmarshal([]byte(jobYaml), &job)
		assert.NoError(t, err, "%s mode: Job template '%s' should be valid YAML", mode, jobTemplateKey)

		testJobGlobalPullSecret(t, job, expectedImagePullSecretsName, mode)
	}
}
