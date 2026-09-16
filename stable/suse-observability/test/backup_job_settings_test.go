package test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestBackupJobSettingsSharedOverrides(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/pod_scheduling_labels.yaml"},
		SetValues: map[string]string{
			"global.commonLabels.shared-label":                               "global",
			"stackstate.components.all.nodeSelector.shared-node":             "all",
			"stackstate.components.backup.nodeSelector.shared-node":          "backup",
			"stackstate.components.all.podAnnotations.shared-note":           "all",
			"stackstate.components.backup.podAnnotations.shared-note":        "backup",
			"stackstate.components.backup.podLabels.shared-label":            "backup",
			"stackstate.components.backup.resources.requests.memory":         "1234Mi",
			"stackstate.components.backup.resources.limits.memory":           "2345Mi",
			"stackstate.components.containerTools.resources.requests.memory": "64Mi",
			"stackstate.components.containerTools.resources.limits.memory":   "128Mi",
			"common.container.securityContext.readOnlyRootFilesystem":        "true",
			// Settings jobs keep their independent overrides.
			"stackstate.components.configurationBackup.nodeSelector.shared-node":   "settings",
			"stackstate.components.configurationBackup.podAnnotations.shared-note": "settings",
			"stackstate.components.configurationBackup.podLabels.shared-label":     "settings",
			"stackstate.components.configurationBackup.resources.requests.memory":  "99Mi",
			"stackstate.components.configurationBackup.resources.limits.memory":    "199Mi",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	cronJob, ok := resources.CronJobs["suse-observability-backup-conf"]
	require.True(t, ok)
	scheduled := cronJob.Spec.JobTemplate.Spec.Template
	stackgraph, ok := resources.CronJobs["suse-observability-backup-sg"]
	require.True(t, ok)
	initJob := findJob(&resources, "init-pvc")
	require.NotNil(t, initJob)
	v2Job, ok := resources.CronJobs["suse-observability-backup-sg-v2"]
	require.True(t, ok)
	assert.Equal(t, "backup-v2", v2Job.Spec.JobTemplate.Spec.Template.Labels["app.kubernetes.io/component"])
	require.Len(t, scheduled.Spec.Containers, 1)
	require.NotNil(t, scheduled.Spec.Containers[0].SecurityContext)
	require.NotNil(t, scheduled.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem)
	assert.True(t, *scheduled.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem)

	pods := map[string]corev1.PodTemplateSpec{
		"settings creation": scheduled,
		"settings PVC init": initJob.Spec.Template,
	}
	jobs := testJobsFromBackupRestoreScriptsConfigMap(t, &resources)
	for name, job := range jobs {
		pods[name] = job.Spec.Template
	}
	for name, pod := range pods {
		t.Run(name, func(t *testing.T) {
			isSettings := strings.HasPrefix(name, "settings") || strings.Contains(name, "configuration")
			expected, reference := "backup", stackgraph.Spec.JobTemplate.Spec.Template
			if isSettings {
				expected, reference = "settings", scheduled
			}
			assert.Equal(t, expected, pod.Spec.NodeSelector["shared-node"])
			assert.Equal(t, expected, pod.Annotations["shared-note"])
			assert.Equal(t, expected, pod.Labels["shared-label"])
			assert.Equal(t, "all-value", pod.Spec.NodeSelector["all-node"])
			assert.Equal(t, "backup", pod.Labels["app.kubernetes.io/component"])
			assert.Equal(t, reference.Spec.Affinity, pod.Spec.Affinity)
			assert.Equal(t, reference.Spec.Tolerations, pod.Spec.Tolerations)
			for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
				require.NotNil(t, container.SecurityContext, container.Name)
				assert.Equal(t, scheduled.Spec.Containers[0].SecurityContext, container.SecurityContext, container.Name)
			}
			require.Len(t, pod.Spec.Containers, 1)
			request, limit := "1234Mi", "2345Mi"
			if isSettings {
				request, limit = "99Mi", "199Mi"
			}
			if strings.Contains(name, "elasticsearch") {
				request, limit = "64Mi", "128Mi"
			} else if name == "settings PVC init" {
				request, limit = "100Mi", "100Mi"
			}
			assert.Equal(t, resource.MustParse(request), *pod.Spec.Containers[0].Resources.Requests.Memory())
			assert.Equal(t, resource.MustParse(limit), *pod.Spec.Containers[0].Resources.Limits.Memory())
		})
	}
}

func TestBackupJobSettingsSecurityContext(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run("enabled="+strconv.FormatBool(enabled), func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{"values/full.yaml"},
				SetValues: map[string]string{
					"backup.configuration.securityContext.enabled":    strconv.FormatBool(enabled),
					"backup.configuration.securityContext.runAsUser":  "12345",
					"backup.configuration.securityContext.runAsGroup": "12346",
					"backup.configuration.securityContext.fsGroup":    "12347",
					"backup.stackGraph.securityContext.enabled":       strconv.FormatBool(!enabled),
					"backup.stackGraph.securityContext.runAsUser":     "23456",
				},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			cronJob, ok := resources.CronJobs["suse-observability-backup-conf"]
			require.True(t, ok)
			initJob := findJob(&resources, "init-pvc")
			require.NotNil(t, initJob)
			pods := map[string]corev1.PodSpec{
				"settings creation": cronJob.Spec.JobTemplate.Spec.Template.Spec,
				"settings PVC init": initJob.Spec.Template.Spec,
			}
			for name, job := range testJobsFromBackupRestoreScriptsConfigMap(t, &resources) {
				if strings.Contains(name, "configuration") {
					pods[name] = job.Spec.Template.Spec
				}
			}
			require.Len(t, pods, 6)
			for name, pod := range pods {
				t.Run(name, func(t *testing.T) {
					if !enabled {
						assert.Nil(t, pod.SecurityContext)
						return
					}
					require.NotNil(t, pod.SecurityContext)
					require.NotNil(t, pod.SecurityContext.RunAsUser)
					require.NotNil(t, pod.SecurityContext.RunAsGroup)
					require.NotNil(t, pod.SecurityContext.FSGroup)
					assert.EqualValues(t, 12345, *pod.SecurityContext.RunAsUser)
					assert.EqualValues(t, 12346, *pod.SecurityContext.RunAsGroup)
					assert.EqualValues(t, 12347, *pod.SecurityContext.FSGroup)
				})
			}
		})
	}
}

func TestBackupJobSettingsEmptyOverridesStayIndependent(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"stackstate.components.all.nodeSelector.shared":          "all",
			"stackstate.components.all.podAnnotations.shared":        "all",
			"stackstate.components.backup.nodeSelector.shared":       "backup",
			"stackstate.components.backup.podAnnotations.shared":     "backup",
			"stackstate.components.backup.podLabels.backup-only":     "backup",
			"stackstate.components.backup.resources.requests.memory": "2345Mi",
			"stackstate.components.backup.tolerations[0].key":        "backup-only",
			"stackstate.components.backup.tolerations[0].operator":   "Exists",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	cronJob, ok := resources.CronJobs["suse-observability-backup-conf"]
	require.True(t, ok)
	initJob := findJob(&resources, "init-pvc")
	require.NotNil(t, initJob)
	pods := map[string]corev1.PodTemplateSpec{
		"settings creation": cronJob.Spec.JobTemplate.Spec.Template,
		"settings PVC init": initJob.Spec.Template,
	}
	for name, job := range testJobsFromBackupRestoreScriptsConfigMap(t, &resources) {
		if strings.Contains(name, "configuration") {
			pods[name] = job.Spec.Template
		}
	}
	for name, pod := range pods {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, map[string]string{"shared": "all"}, pod.Spec.NodeSelector)
			assert.Equal(t, map[string]string{"shared": "all"}, pod.Annotations)
			assert.NotContains(t, pod.Labels, "backup-only")
			assert.Empty(t, pod.Spec.Tolerations)
			if name != "settings PVC init" {
				assert.Equal(t, resource.MustParse("1000Mi"), *pod.Spec.Containers[0].Resources.Requests.Memory())
			}
		})
	}
}

func TestBackupJobSettingsDefaults(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	cronJob, ok := resources.CronJobs["suse-observability-backup-conf"]
	require.True(t, ok)
	pods := map[string]corev1.PodTemplateSpec{"settings creation": cronJob.Spec.JobTemplate.Spec.Template}
	for name, job := range testJobsFromBackupRestoreScriptsConfigMap(t, &resources) {
		pods[name] = job.Spec.Template
	}
	for name, pod := range pods {
		t.Run(name, func(t *testing.T) {
			assert.Empty(t, pod.Annotations)
			assert.Empty(t, pod.Spec.NodeSelector)
			assert.Empty(t, pod.Spec.Tolerations)
			if name == "settings creation" || strings.Contains(name, "configuration") {
				assert.Equal(t, resource.MustParse("1000Mi"), *pod.Spec.Containers[0].Resources.Requests.Memory())
				assert.Equal(t, resource.MustParse("1000Mi"), *pod.Spec.Containers[0].Resources.Limits.Memory())
				assert.Equal(t, resource.MustParse("1000m"), *pod.Spec.Containers[0].Resources.Limits.Cpu())
				assert.Equal(t, resource.MustParse("100Mi"), *pod.Spec.Containers[0].Resources.Requests.StorageEphemeral())
			}
		})
	}
	// Check the embedded jobs too: they are not top-level Helm resources.
	embedded := helmtestutil.KubernetesResources{
		Jobs: testJobsFromBackupRestoreScriptsConfigMap(t, &resources),
	}
	helmtestutil.AssertRestrictedSecurityContext(t, embedded, nil)
}

func TestBackupJobSettingsExplicitEmptyStrings(t *testing.T) {
	for _, emptyComponent := range []string{"backup", "configurationBackup"} {
		t.Run(emptyComponent, func(t *testing.T) {
			overrides := map[string]string{
				"global.commonLabels.shared":                         "global",
				"global.commonLabels.inherited":                      "global",
				"stackstate.components.all.nodeSelector.shared":      "all",
				"stackstate.components.all.nodeSelector.inherited":   "all",
				"stackstate.components.all.podAnnotations.shared":    "all",
				"stackstate.components.all.podAnnotations.inherited": "all",
			}
			expected := map[string]string{"backup": "backup", "configurationBackup": "settings"}
			expected[emptyComponent] = ""
			for component, value := range expected {
				for _, setting := range []string{"nodeSelector", "podAnnotations", "podLabels"} {
					overrides["stackstate.components."+component+"."+setting+".shared"] = value
				}
			}
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{"values/full.yaml"},
				SetValues:   overrides,
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			pods := map[string]corev1.PodTemplateSpec{}
			for _, name := range []string{"backup-sg", "backup-sg-v2", "backup-init", "backup-conf"} {
				job, ok := resources.CronJobs["suse-observability-"+name]
				require.True(t, ok, name)
				pods[name] = job.Spec.JobTemplate.Spec.Template
			}
			for _, name := range []string{"init-pvc", "backup-init"} {
				job := findJob(&resources, name)
				require.NotNil(t, job, name)
				pods["job-"+name] = job.Spec.Template
			}
			for name, job := range testJobsFromBackupRestoreScriptsConfigMap(t, &resources) {
				pods[name] = job.Spec.Template
			}
			for name, pod := range pods {
				t.Run(name, func(t *testing.T) {
					component := "backup"
					if name == "backup-conf" || name == "job-init-pvc" || strings.Contains(name, "configuration") {
						component = "configurationBackup"
					}
					assert.Equal(t, map[string]string{"shared": expected[component], "inherited": "all"}, pod.Spec.NodeSelector)
					assert.Equal(t, map[string]string{"shared": expected[component], "inherited": "all"}, pod.Annotations)
					assert.Contains(t, pod.Labels, "shared")
					assert.Equal(t, expected[component], pod.Labels["shared"])
					assert.Equal(t, "global", pod.Labels["inherited"])
					label := "backup"
					if name == "backup-sg-v2" {
						label = "backup-v2"
					}
					assert.Equal(t, label, pod.Labels["app.kubernetes.io/component"])
				})
			}
		})
	}
}

func TestBackupConfigurationTemporaryStorage(t *testing.T) {
	for _, readOnly := range []bool{true, false} {
		t.Run("readOnlyRootFilesystem="+strconv.FormatBool(readOnly), func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{"values/full.yaml"},
				SetValues: map[string]string{
					"common.container.securityContext.readOnlyRootFilesystem": strconv.FormatBool(readOnly),
				},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			jobs := testJobsFromBackupRestoreScriptsConfigMap(t, &resources)
			for _, name := range []string{"job-configuration-restore-backup.yaml", "job-configuration-download-backup.yaml"} {
				t.Run(name, func(t *testing.T) {
					job, ok := jobs[name]
					require.True(t, ok)
					pod := job.Spec.Template.Spec
					require.Len(t, pod.Containers, 1)
					container := pod.Containers[0]
					require.NotNil(t, container.SecurityContext)
					require.NotNil(t, container.SecurityContext.ReadOnlyRootFilesystem)
					assert.Equal(t, readOnly, *container.SecurityContext.ReadOnlyRootFilesystem)

					var tmpMounts []corev1.VolumeMount
					for _, mount := range container.VolumeMounts {
						if mount.MountPath == "/tmp" {
							tmpMounts = append(tmpMounts, mount)
						}
					}
					require.Len(t, tmpMounts, 1, "backup scripts need a writable /tmp")
					assert.False(t, tmpMounts[0].ReadOnly)
					var tmpVolumes []corev1.Volume
					for _, volume := range pod.Volumes {
						if volume.Name == tmpMounts[0].Name {
							tmpVolumes = append(tmpVolumes, volume)
						}
					}
					require.Len(t, tmpVolumes, 1)
					require.NotNil(t, tmpVolumes[0].EmptyDir, "/tmp must use temporary storage independent of the container root filesystem")
				})
			}
		})
	}
}

func TestBackupConfigurationUploadStorage(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"common.container.securityContext.readOnlyRootFilesystem": "true",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	job, ok := testJobsFromBackupRestoreScriptsConfigMap(t, &resources)["job-configuration-upload-backup.yaml"]
	require.True(t, ok)
	require.Len(t, job.Spec.Template.Spec.Containers, 1)
	container := job.Spec.Template.Spec.Containers[0]
	require.NotNil(t, container.SecurityContext)
	require.NotNil(t, container.SecurityContext.ReadOnlyRootFilesystem)
	assert.True(t, *container.SecurityContext.ReadOnlyRootFilesystem)

	// The upload wrapper copies to this directory, and deletes the Job on exit.
	// Keep the uploaded backup on the legacy settings PVC after that cleanup.
	var mounts []corev1.VolumeMount
	for _, mount := range container.VolumeMounts {
		if mount.MountPath == "/settings-backup-data" {
			mounts = append(mounts, mount)
		}
	}
	require.Len(t, mounts, 1)
	assert.False(t, mounts[0].ReadOnly)
	var volumes []corev1.Volume
	for _, volume := range job.Spec.Template.Spec.Volumes {
		if volume.Name == mounts[0].Name {
			volumes = append(volumes, volume)
		}
	}
	require.Len(t, volumes, 1)
	require.NotNil(t, volumes[0].PersistentVolumeClaim)
	assert.False(t, volumes[0].PersistentVolumeClaim.ReadOnly)
	assert.Equal(t, "suse-observability-settings-backup-data", volumes[0].PersistentVolumeClaim.ClaimName)
	_, ok = resources.PersistentVolumeClaims[volumes[0].PersistentVolumeClaim.ClaimName]
	assert.True(t, ok, "the upload destination PVC must be rendered")
}

func TestBackupManualJobContainerSecurityOverrides(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"common.container.securityContext.readOnlyRootFilesystem":           "true",
			"backup.manualJobs.containerSecurityContext.runAsNonRoot":           "false",
			"backup.manualJobs.containerSecurityContext.runAsUser":              "0",
			"backup.manualJobs.containerSecurityContext.runAsGroup":             "0",
			"backup.manualJobs.containerSecurityContext.readOnlyRootFilesystem": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	jobs := testJobsFromBackupRestoreScriptsConfigMap(t, &resources)
	require.NotEmpty(t, jobs)
	for name, job := range jobs {
		t.Run(name, func(t *testing.T) {
			pod := job.Spec.Template.Spec
			// Container overrides must take effect even when the pod still defaults
			// to non-root. Explicit false and UID/GID zero must survive the merge.
			require.NotNil(t, pod.SecurityContext)
			require.NotNil(t, pod.SecurityContext.RunAsNonRoot)
			assert.True(t, *pod.SecurityContext.RunAsNonRoot)
			for _, container := range append(pod.InitContainers, pod.Containers...) {
				sc := container.SecurityContext
				require.NotNil(t, sc, container.Name)
				require.NotNil(t, sc.RunAsNonRoot, container.Name)
				assert.False(t, *sc.RunAsNonRoot, container.Name)
				require.NotNil(t, sc.RunAsUser, container.Name)
				assert.EqualValues(t, 0, *sc.RunAsUser, container.Name)
				require.NotNil(t, sc.RunAsGroup, container.Name)
				assert.EqualValues(t, 0, *sc.RunAsGroup, container.Name)
				require.NotNil(t, sc.ReadOnlyRootFilesystem, container.Name)
				assert.False(t, *sc.ReadOnlyRootFilesystem, container.Name)
				require.NotNil(t, sc.AllowPrivilegeEscalation, container.Name)
				assert.False(t, *sc.AllowPrivilegeEscalation, "other common defaults must be retained")
				require.NotNil(t, sc.Capabilities, container.Name)
				assert.Equal(t, []corev1.Capability{"ALL"}, sc.Capabilities.Drop)
				require.NotNil(t, sc.SeccompProfile, container.Name)
				assert.Equal(t, corev1.SeccompProfileTypeRuntimeDefault, sc.SeccompProfile.Type)
			}
		})
	}
	// A manual-job exception must not mutate the common context or weaken other workloads.
	cronJob, ok := resources.CronJobs["suse-observability-backup-conf"]
	require.True(t, ok)
	api, ok := resources.Deployments["suse-observability-api"]
	require.True(t, ok)
	for _, pod := range []corev1.PodSpec{cronJob.Spec.JobTemplate.Spec.Template.Spec, api.Spec.Template.Spec} {
		sc := pod.Containers[0].SecurityContext
		require.NotNil(t, sc)
		require.NotNil(t, sc.RunAsNonRoot)
		assert.True(t, *sc.RunAsNonRoot)
		require.NotNil(t, sc.ReadOnlyRootFilesystem)
		assert.True(t, *sc.ReadOnlyRootFilesystem)
	}
}

func TestBackupSettingsInitializerWithoutContainerToolsResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/backup_container_tools_resources_null.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	job := findJob(&resources, "init-pvc")
	require.NotNil(t, job)
	require.Len(t, job.Spec.Template.Spec.Containers, 1)
	container := job.Spec.Template.Spec.Containers[0]
	assert.Equal(t, resource.MustParse("10m"), *container.Resources.Requests.Cpu())
	assert.Equal(t, resource.MustParse("10m"), *container.Resources.Limits.Cpu())
	assert.Equal(t, resource.MustParse("100Mi"), *container.Resources.Requests.Memory())
	assert.Equal(t, resource.MustParse("100Mi"), *container.Resources.Limits.Memory())
	require.NotNil(t, container.SecurityContext)
	require.NotNil(t, container.SecurityContext.RunAsNonRoot)
	assert.True(t, *container.SecurityContext.RunAsNonRoot)
	require.NotNil(t, container.SecurityContext.AllowPrivilegeEscalation)
	assert.False(t, *container.SecurityContext.AllowPrivilegeEscalation)
}
