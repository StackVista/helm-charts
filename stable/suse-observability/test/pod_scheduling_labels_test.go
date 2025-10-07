package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func TestPodSchedulingAndLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/pod_scheduling_labels.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Test that pod annotations, labels, nodeSelector, affinity, and tolerations are applied correctly
	testJobPodSchedulingAndLabels(t, &resources)
	testCronJobPodSchedulingAndLabels(t, &resources)
	testPodSchedulingBackupConfigMapJobTemplates(t, &resources)
}

func testJobPodSchedulingAndLabels(t *testing.T, resources *helmtestutil.KubernetesResources) {
	// Test backup init job
	backupInitJob := findJob(resources, "backup-init")
	if backupInitJob != nil {
		testPodAnnotations(t, backupInitJob.Spec.Template.Annotations, map[string]string{
			"all-annotation":    "all-value",
			"backup-annotation": "backup-value",
		}, "Job", backupInitJob.Name)

		testPodLabels(t, backupInitJob.Spec.Template.Labels, map[string]string{
			"all-label":    "all-value",
			"backup-label": "backup-value",
		}, "Job", backupInitJob.Name)

		testNodeSelector(t, backupInitJob.Spec.Template.Spec.NodeSelector, map[string]string{
			"all-node":    "all-value",
			"backup-node": "backup-value",
		}, "Job", backupInitJob.Name)

		testAffinity(t, backupInitJob.Spec.Template.Spec.Affinity, "Job", backupInitJob.Name)
		testTolerations(t, backupInitJob.Spec.Template.Spec.Tolerations, "Job", backupInitJob.Name)
	}

	// Test configuration backup init job
	configBackupJob := findJob(resources, "backup-conf")
	if configBackupJob != nil {
		testPodAnnotations(t, configBackupJob.Spec.Template.Annotations, map[string]string{
			"all-annotation":           "all-value",
			"config-backup-annotation": "config-value",
		}, "Job", configBackupJob.Name)

		testPodLabels(t, configBackupJob.Spec.Template.Labels, map[string]string{
			"all-label":           "all-value",
			"config-backup-label": "config-value",
		}, "Job", configBackupJob.Name)

		testNodeSelector(t, configBackupJob.Spec.Template.Spec.NodeSelector, map[string]string{
			"all-node":           "all-value",
			"config-backup-node": "config-value",
		}, "Job", configBackupJob.Name)
	}

	// Test clickhouse cleanup job
	clickhouseCleanupJob := findJob(resources, "ch-clean")
	if clickhouseCleanupJob != nil {
		testPodAnnotations(t, clickhouseCleanupJob.Spec.Template.Annotations, map[string]string{
			"all-annotation":         "all-value",
			"ch-cleanup-annotation":  "ch-cleanup-value",
		}, "Job", clickhouseCleanupJob.Name)

		testPodLabels(t, clickhouseCleanupJob.Spec.Template.Labels, map[string]string{
			"all-label":         "all-value",
			"ch-cleanup-label":  "ch-cleanup-value",
		}, "Job", clickhouseCleanupJob.Name)
	}

	// Test kafka topic create job
	kafkaTopicJob := findJob(resources, "topic-create")
	if kafkaTopicJob != nil {
		testPodAnnotations(t, kafkaTopicJob.Spec.Template.Annotations, map[string]string{
			"all-annotation":         "all-value",
			"kafka-topic-annotation": "kafka-value",
		}, "Job", kafkaTopicJob.Name)

		testPodLabels(t, kafkaTopicJob.Spec.Template.Labels, map[string]string{
			"all-label":         "all-value",
			"kafka-topic-label": "kafka-value",
		}, "Job", kafkaTopicJob.Name)
	}

	// Test router mode jobs
	routerModeActiveJob := findJob(resources, "set-active")
	if routerModeActiveJob != nil {
		testPodAnnotations(t, routerModeActiveJob.Spec.Template.Annotations, map[string]string{
			"all-annotation":         "all-value",
			"router-mode-annotation": "router-value",
		}, "Job", routerModeActiveJob.Name)

		testPodLabels(t, routerModeActiveJob.Spec.Template.Labels, map[string]string{
			"all-label":         "all-value",
			"router-mode-label": "router-value",
		}, "Job", routerModeActiveJob.Name)
	}

	routerModeMaintenanceJob := findJob(resources, "set-maintenance")
	if routerModeMaintenanceJob != nil {
		testPodAnnotations(t, routerModeMaintenanceJob.Spec.Template.Annotations, map[string]string{
			"all-annotation":         "all-value",
			"router-mode-annotation": "router-value",
		}, "Job", routerModeMaintenanceJob.Name)

		testPodLabels(t, routerModeMaintenanceJob.Spec.Template.Labels, map[string]string{
			"all-label":         "all-value",
			"router-mode-label": "router-value",
		}, "Job", routerModeMaintenanceJob.Name)
	}
}

func testCronJobPodSchedulingAndLabels(t *testing.T, resources *helmtestutil.KubernetesResources) {
	for _, cronjob := range resources.CronJobs {
		name := getCronJobName(cronjob)
		podTemplate := getCronJobPodTemplate(cronjob)

		// Test backup stackgraph cronjob
		if strings.Contains(name, "backup-sg") {
			testPodAnnotations(t, podTemplate.Annotations, map[string]string{
				"all-annotation":    "all-value",
				"backup-annotation": "backup-value",
			}, "CronJob", name)

			testPodLabels(t, podTemplate.Labels, map[string]string{
				"all-label":    "all-value",
				"backup-label": "backup-value",
			}, "CronJob", name)

			testNodeSelector(t, podTemplate.Spec.NodeSelector, map[string]string{
				"all-node":    "all-value",
				"backup-node": "backup-value",
			}, "CronJob", name)

			testAffinity(t, podTemplate.Spec.Affinity, "CronJob", name)
			testTolerations(t, podTemplate.Spec.Tolerations, "CronJob", name)
		}

		// Test configuration backup cronjob
		if strings.Contains(name, "backup-conf") {
			testPodAnnotations(t, podTemplate.Annotations, map[string]string{
				"all-annotation":           "all-value",
				"config-backup-annotation": "config-value",
			}, "CronJob", name)

			testPodLabels(t, podTemplate.Labels, map[string]string{
				"all-label":           "all-value",
				"config-backup-label": "config-value",
			}, "CronJob", name)
		}

		// Test clickhouse full backup cronjob
		if strings.Contains(name, "clickhouse-full-backup") {
			testPodAnnotations(t, podTemplate.Annotations, map[string]string{
				"all-annotation":               "all-value",
				"clickhouse-backup-annotation": "clickhouse-value",
			}, "CronJob", name)

			testPodLabels(t, podTemplate.Labels, map[string]string{
				"all-label":               "all-value",
				"clickhouse-backup-label": "clickhouse-value",
			}, "CronJob", name)

			testNodeSelector(t, podTemplate.Spec.NodeSelector, map[string]string{
				"all-node":               "all-value",
				"clickhouse-backup-node": "clickhouse-value",
			}, "CronJob", name)
		}

		// Test clickhouse incremental backup cronjob
		if strings.Contains(name, "clickhouse-incremental-backup") {
			testPodAnnotations(t, podTemplate.Annotations, map[string]string{
				"all-annotation":               "all-value",
				"clickhouse-backup-annotation": "clickhouse-value",
			}, "CronJob", name)

			testPodLabels(t, podTemplate.Labels, map[string]string{
				"all-label":               "all-value",
				"clickhouse-backup-label": "clickhouse-value",
			}, "CronJob", name)
		}
	}
}

func testPodSchedulingBackupConfigMapJobTemplates(t *testing.T, resources *helmtestutil.KubernetesResources) {
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

	// List of Job template keys that should have backup component settings
	backupJobTemplateKeys := []string{
		"job-elasticsearch-list-snapshots.yaml",
		"job-elasticsearch-restore-snapshot.yaml",
		"job-stackgraph-list-backups.yaml",
		"job-stackgraph-restore-backup.yaml",
	}

	// Test each backup Job template in the ConfigMap
	for _, jobTemplateKey := range backupJobTemplateKeys {
		jobYaml, exists := backupConfigMap.Data[jobTemplateKey]
		if !exists {
			continue
		}

		var job batchv1.Job
		err := yaml.Unmarshal([]byte(jobYaml), &job)
		assert.NoError(t, err, "Job template '%s' should be valid YAML", jobTemplateKey)

		testPodAnnotations(t, job.Spec.Template.Annotations, map[string]string{
			"all-annotation":    "all-value",
			"backup-annotation": "backup-value",
		}, "ConfigMap Job", jobTemplateKey)

		testPodLabels(t, job.Spec.Template.Labels, map[string]string{
			"all-label":    "all-value",
			"backup-label": "backup-value",
		}, "ConfigMap Job", jobTemplateKey)

		testNodeSelector(t, job.Spec.Template.Spec.NodeSelector, map[string]string{
			"all-node":    "all-value",
			"backup-node": "backup-value",
		}, "ConfigMap Job", jobTemplateKey)

		testAffinity(t, job.Spec.Template.Spec.Affinity, "ConfigMap Job", jobTemplateKey)
		testTolerations(t, job.Spec.Template.Spec.Tolerations, "ConfigMap Job", jobTemplateKey)
	}

	// List of Job template keys that should have configurationBackup component settings
	configBackupJobTemplateKeys := []string{
		"job-configuration-list-backups.yaml",
		"job-configuration-restore-backup.yaml",
		"job-configuration-download-backup.yaml",
		"job-configuration-upload-backup.yaml",
	}

	// Test each configuration backup Job template in the ConfigMap
	for _, jobTemplateKey := range configBackupJobTemplateKeys {
		jobYaml, exists := backupConfigMap.Data[jobTemplateKey]
		if !exists {
			continue
		}

		var job batchv1.Job
		err := yaml.Unmarshal([]byte(jobYaml), &job)
		assert.NoError(t, err, "Job template '%s' should be valid YAML", jobTemplateKey)

		testPodAnnotations(t, job.Spec.Template.Annotations, map[string]string{
			"all-annotation":           "all-value",
			"config-backup-annotation": "config-value",
		}, "ConfigMap Job", jobTemplateKey)

		testPodLabels(t, job.Spec.Template.Labels, map[string]string{
			"all-label":           "all-value",
			"config-backup-label": "config-value",
		}, "ConfigMap Job", jobTemplateKey)

		testNodeSelector(t, job.Spec.Template.Spec.NodeSelector, map[string]string{
			"all-node":           "all-value",
			"config-backup-node": "config-value",
		}, "ConfigMap Job", jobTemplateKey)
	}
}

// Helper functions

func findJob(resources *helmtestutil.KubernetesResources, nameSubstring string) *batchv1.Job {
	for _, job := range resources.Jobs {
		if strings.Contains(job.Name, nameSubstring) {
			return &job
		}
	}
	return nil
}

func getCronJobName(cronjob interface{}) string {
	switch cj := cronjob.(type) {
	case batchv1.CronJob:
		return cj.Name
	case batchv1beta1.CronJob:
		return cj.Name
	}
	return ""
}

func getCronJobPodTemplate(cronjob interface{}) *corev1.PodTemplateSpec {
	switch cj := cronjob.(type) {
	case batchv1.CronJob:
		return &cj.Spec.JobTemplate.Spec.Template
	case batchv1beta1.CronJob:
		return &cj.Spec.JobTemplate.Spec.Template
	}
	return nil
}

func testPodAnnotations(t *testing.T, annotations map[string]string, expectedAnnotations map[string]string, resourceType string, resourceName string) {
	for key, expectedValue := range expectedAnnotations {
		actualValue, exists := annotations[key]
		assert.True(t, exists, "%s '%s' pod template should have annotation '%s'", resourceType, resourceName, key)
		assert.Equal(t, expectedValue, actualValue, "%s '%s' pod annotation '%s' should have value '%s', got '%s'", resourceType, resourceName, key, expectedValue, actualValue)
	}
}

func testPodLabels(t *testing.T, labels map[string]string, expectedLabels map[string]string, resourceType string, resourceName string) {
	for key, expectedValue := range expectedLabels {
		actualValue, exists := labels[key]
		assert.True(t, exists, "%s '%s' pod template should have label '%s'", resourceType, resourceName, key)
		assert.Equal(t, expectedValue, actualValue, "%s '%s' pod label '%s' should have value '%s', got '%s'", resourceType, resourceName, key, expectedValue, actualValue)
	}
}

func testNodeSelector(t *testing.T, nodeSelector map[string]string, expectedNodeSelector map[string]string, resourceType string, resourceName string) {
	if len(expectedNodeSelector) == 0 {
		return
	}
	assert.NotNil(t, nodeSelector, "%s '%s' should have nodeSelector", resourceType, resourceName)
	for key, expectedValue := range expectedNodeSelector {
		actualValue, exists := nodeSelector[key]
		assert.True(t, exists, "%s '%s' nodeSelector should have key '%s'", resourceType, resourceName, key)
		assert.Equal(t, expectedValue, actualValue, "%s '%s' nodeSelector '%s' should have value '%s', got '%s'", resourceType, resourceName, key, expectedValue, actualValue)
	}
}

func testAffinity(t *testing.T, affinity *corev1.Affinity, resourceType string, resourceName string) {
	assert.NotNil(t, affinity, "%s '%s' should have affinity", resourceType, resourceName)
	if affinity == nil {
		return
	}
	// Just verify that some affinity is set (either nodeAffinity, podAffinity, or podAntiAffinity)
	hasAffinity := affinity.NodeAffinity != nil || affinity.PodAffinity != nil || affinity.PodAntiAffinity != nil
	assert.True(t, hasAffinity, "%s '%s' should have at least one type of affinity configured", resourceType, resourceName)
}

func testTolerations(t *testing.T, tolerations []corev1.Toleration, resourceType string, resourceName string) {
	assert.NotNil(t, tolerations, "%s '%s' should have tolerations", resourceType, resourceName)
	// We expect at least 2 tolerations: one from "all" and one from component-specific
	assert.GreaterOrEqual(t, len(tolerations), 2, "%s '%s' should have at least 2 tolerations", resourceType, resourceName)
}
