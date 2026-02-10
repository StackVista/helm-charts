package test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

var expectedImageRegistry = "global.image.registry.test"

func TestGlobalImageRegistryHAMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_image_registry_ha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	testGlobalImageRegistryOnAllWorkloads(t, &resources, expectedImageRegistry, "HA")
}

func TestGlobalImageRegistryNonHAMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_image_registry_nonha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	testGlobalImageRegistryOnAllWorkloads(t, &resources, expectedImageRegistry, "Non-HA")
}

func testGlobalImageRegistryOnAllWorkloads(t *testing.T, resources *helmtestutil.KubernetesResources, expectedImageRegistry string, mode string) {
	for _, deployment := range resources.Deployments {
		testName := fmt.Sprintf("%s_%s_%s", mode, "Deployment", deployment.Name)
		t.Run(testName, func(t *testing.T) {
			testImageRegistryInPodTemplateSpec(t, deployment.Spec.Template, expectedImageRegistry)
		})
	}

	for _, statefulset := range resources.Statefulsets {
		testName := fmt.Sprintf("%s_%s_%s", mode, "Statefulset", statefulset.Name)
		t.Run(testName, func(t *testing.T) {
			testImageRegistryInPodTemplateSpec(t, statefulset.Spec.Template, expectedImageRegistry)
		})
	}

	for _, job := range resources.Jobs {
		testName := fmt.Sprintf("%s_%s_%s", mode, "Job", job.Name)
		t.Run(testName, func(t *testing.T) {
			testImageRegistryInPodTemplateSpec(t, job.Spec.Template, expectedImageRegistry)
		})
	}

	for _, cronjob := range resources.CronJobs {
		testName := fmt.Sprintf("%s_%s_%s", mode, "Cronjob", cronjob.Name)
		t.Run(testName, func(t *testing.T) {
			testImageRegistryInPodTemplateSpec(t, cronjob.Spec.JobTemplate.Spec.Template, expectedImageRegistry)
		})
	}

	t.Run(mode+"backup_config_map", func(t *testing.T) {
		testBackupConfigMapJobTemplatesGlobalImageRegistry(t, resources, expectedImageRegistry, mode)
	})
}

func testBackupConfigMapJobTemplatesGlobalImageRegistry(t *testing.T, resources *helmtestutil.KubernetesResources, expectedImageRegistry string, mode string) {

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

		testImageRegistryInPodTemplateSpec(t, job.Spec.Template, expectedImageRegistry)
	}
}

func testImageRegistryInPodTemplateSpec(t *testing.T, spec corev1.PodTemplateSpec, expectedImageRegistry string) {
	startsWithRegex := "^" + expectedImageRegistry
	for _, container := range spec.Spec.InitContainers {
		assert.Regexp(t, regexp.MustCompile(startsWithRegex), container.Image, "InitContainer '%s' has wrong registry: %s, expected: %s", container.Name, container.Image, expectedImageRegistry)
	}
	for _, container := range spec.Spec.Containers {
		assert.Regexp(t, regexp.MustCompile(startsWithRegex), container.Image, "Container '%s' has wrong registry: %s, expected: %s", container.Name, container.Image, expectedImageRegistry)
	}
}
