package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// expectedGlobalLabelsHA are the expected global labels for HA mode
var expectedGlobalLabelsHA = map[string]string{
	"environment": "production",
	"team":        "platform",
	"deployment":  "ha",
}

// expectedGlobalLabelsNonHA are the expected global labels for Non-HA mode
var expectedGlobalLabelsNonHA = map[string]string{
	"environment": "test",
	"team":        "platform",
}

func TestGlobalLabelsHAMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/workload_ha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Test that global labels are applied to all workloads
	testGlobalLabelsOnAllWorkloads(t, &resources, expectedGlobalLabelsHA, "HA")
}

func TestGlobalLabelsNonHAMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/workload_nonha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Test that global labels are applied to all workloads
	testGlobalLabelsOnAllWorkloads(t, &resources, expectedGlobalLabelsNonHA, "Non-HA")
}

func testGlobalLabelsOnAllWorkloads(t *testing.T, resources *helmtestutil.KubernetesResources, expectedLabels map[string]string, mode string) {
	// Test ALL Deployments - check both workload metadata and pod template labels
	for _, deployment := range resources.Deployments {
		testDeploymentGlobalLabels(t, deployment, expectedLabels, mode)
	}

	// Test ALL StatefulSets - check both workload metadata and pod template labels
	for _, statefulset := range resources.Statefulsets {
		testStatefulSetGlobalLabels(t, statefulset, expectedLabels, mode)
	}

	// Test ALL Jobs - check both workload metadata and pod template labels
	for _, job := range resources.Jobs {
		testJobGlobalLabels(t, job, expectedLabels, mode)
	}

	// Test ALL CronJobs - check workload metadata, job template, and pod template labels
	for _, cronjob := range resources.CronJobs {
		testCronJobGlobalLabels(t, cronjob, expectedLabels, mode)
	}

	// Test backup ConfigMap Job templates - unmarshal YAML and validate Job structures
	testBackupConfigMapJobTemplates(t, resources, expectedLabels, mode)
}

func testDeploymentGlobalLabels(t *testing.T, deployment appsv1.Deployment, expectedLabels map[string]string, mode string) {
	// Test deployment metadata labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := deployment.Labels[labelKey]
		assert.True(t, exists, "%s mode: Deployment '%s' should have global label '%s'", mode, deployment.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: Deployment '%s' global label '%s' should have value '%s', got '%s'", mode, deployment.Name, labelKey, expectedValue, actualValue)
	}

	// Test pod template labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := deployment.Spec.Template.Labels[labelKey]
		assert.True(t, exists, "%s mode: Deployment '%s' pod template should have global label '%s'", mode, deployment.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: Deployment '%s' pod template global label '%s' should have value '%s', got '%s'", mode, deployment.Name, labelKey, expectedValue, actualValue)
	}
}

func testStatefulSetGlobalLabels(t *testing.T, statefulset appsv1.StatefulSet, expectedLabels map[string]string, mode string) {
	// Test statefulset metadata labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := statefulset.Labels[labelKey]
		assert.True(t, exists, "%s mode: StatefulSet '%s' should have global label '%s'", mode, statefulset.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: StatefulSet '%s' global label '%s' should have value '%s', got '%s'", mode, statefulset.Name, labelKey, expectedValue, actualValue)
	}

	// Test pod template labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := statefulset.Spec.Template.Labels[labelKey]
		assert.True(t, exists, "%s mode: StatefulSet '%s' pod template should have global label '%s'", mode, statefulset.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: StatefulSet '%s' pod template global label '%s' should have value '%s', got '%s'", mode, statefulset.Name, labelKey, expectedValue, actualValue)
	}
}

func testJobGlobalLabels(t *testing.T, job batchv1.Job, expectedLabels map[string]string, mode string) {
	// Test job metadata labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := job.Labels[labelKey]
		assert.True(t, exists, "%s mode: Job '%s' should have global label '%s'", mode, job.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: Job '%s' global label '%s' should have value '%s', got '%s'", mode, job.Name, labelKey, expectedValue, actualValue)
	}

	// Test pod template labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := job.Spec.Template.Labels[labelKey]
		assert.True(t, exists, "%s mode: Job '%s' pod template should have global label '%s'", mode, job.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: Job '%s' pod template global label '%s' should have value '%s', got '%s'", mode, job.Name, labelKey, expectedValue, actualValue)
	}
}

func testCronJobGlobalLabels(t *testing.T, cronjob interface{}, expectedLabels map[string]string, mode string) {
	// Handle both batch/v1 and batch/v1beta1 CronJob versions
	switch cj := cronjob.(type) {
	case batchv1.CronJob:
		testCronJobV1GlobalLabels(t, cj, expectedLabels, mode)
	case batchv1beta1.CronJob:
		testCronJobV1Beta1GlobalLabels(t, cj, expectedLabels, mode)
	}
}

func testCronJobV1GlobalLabels(t *testing.T, cronjob batchv1.CronJob, expectedLabels map[string]string, mode string) {
	// Test cronjob metadata labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := cronjob.Labels[labelKey]
		assert.True(t, exists, "%s mode: CronJob '%s' should have global label '%s'", mode, cronjob.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: CronJob '%s' global label '%s' should have value '%s', got '%s'", mode, cronjob.Name, labelKey, expectedValue, actualValue)
	}

	// Test job template labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := cronjob.Spec.JobTemplate.Labels[labelKey]
		assert.True(t, exists, "%s mode: CronJob '%s' job template should have global label '%s'", mode, cronjob.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: CronJob '%s' job template global label '%s' should have value '%s', got '%s'", mode, cronjob.Name, labelKey, expectedValue, actualValue)
	}

	// Test pod template labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := cronjob.Spec.JobTemplate.Spec.Template.Labels[labelKey]
		assert.True(t, exists, "%s mode: CronJob '%s' pod template should have global label '%s'", mode, cronjob.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: CronJob '%s' pod template global label '%s' should have value '%s', got '%s'", mode, cronjob.Name, labelKey, expectedValue, actualValue)
	}
}

func testCronJobV1Beta1GlobalLabels(t *testing.T, cronjob batchv1beta1.CronJob, expectedLabels map[string]string, mode string) {
	// Test cronjob metadata labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := cronjob.Labels[labelKey]
		assert.True(t, exists, "%s mode: CronJob '%s' should have global label '%s'", mode, cronjob.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: CronJob '%s' global label '%s' should have value '%s', got '%s'", mode, cronjob.Name, labelKey, expectedValue, actualValue)
	}

	// Test job template labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := cronjob.Spec.JobTemplate.Labels[labelKey]
		assert.True(t, exists, "%s mode: CronJob '%s' job template should have global label '%s'", mode, cronjob.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: CronJob '%s' job template global label '%s' should have value '%s', got '%s'", mode, cronjob.Name, labelKey, expectedValue, actualValue)
	}

	// Test pod template labels
	for labelKey, expectedValue := range expectedLabels {
		actualValue, exists := cronjob.Spec.JobTemplate.Spec.Template.Labels[labelKey]
		assert.True(t, exists, "%s mode: CronJob '%s' pod template should have global label '%s'", mode, cronjob.Name, labelKey)
		assert.Equal(t, expectedValue, actualValue, "%s mode: CronJob '%s' pod template global label '%s' should have value '%s', got '%s'", mode, cronjob.Name, labelKey, expectedValue, actualValue)
	}
}

func testBackupConfigMapJobTemplates(t *testing.T, resources *helmtestutil.KubernetesResources, expectedLabels map[string]string, mode string) {
	// Find the backup-restore-scripts ConfigMap
	var backupConfigMap *corev1.ConfigMap
	for _, cm := range resources.ConfigMaps {
		if strings.HasSuffix(cm.Name, "-backup-restore-scripts") {
			backupConfigMap = &cm
			break
		}
	}

	// Skip test if backup is not enabled (ConfigMap not found)
	if backupConfigMap == nil {
		return
	}

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

	// Test each Job template in the ConfigMap data
	for _, jobTemplateKey := range expectedJobTemplateKeys {
		jobYaml, exists := backupConfigMap.Data[jobTemplateKey]
		if !exists {
			continue // Skip if this particular job template is not present (conditional rendering)
		}

		// Unmarshal the YAML into a Job struct
		var job batchv1.Job
		err := yaml.Unmarshal([]byte(jobYaml), &job)
		assert.NoError(t, err, "%s mode: Job template '%s' should be valid YAML", mode, jobTemplateKey)

		// Test Job metadata labels contain global labels
		testJobGlobalLabels(t, job, expectedLabels, mode)
	}
}
