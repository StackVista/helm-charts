package test

import (
	"fmt"
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

func TestResourceNamingStatelessComponents(t *testing.T) {
	for _, release := range []string{"suse-observability", "nightly"} {
		for _, split := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/split=%t", release, split), func(t *testing.T) {
				output := helmtestutil.RenderHelmTemplateOptsNoError(t, release, &helm.Options{
					ValuesFiles:    []string{"values/full.yaml"},
					KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
					SetValues: map[string]string{
						"stackstate.features.server.split":                         fmt.Sprint(split),
						"stackstate.components.replicationChecker.enabled":         "true",
						"stackstate.components.ui.extraEnv.secret.TEST_SECRET":     "example",
						"stackstate.components.all.metrics.servicemonitor.enabled": "true",
						"s3proxy.metrics.servicemonitor.enabled":                   "true",
					},
				})
				resources := helmtestutil.NewKubernetesResources(t, output)

				const uiName = "suse-observability-ui"
				require.Contains(t, resources.Deployments, uiName)
				require.Contains(t, resources.Services, uiName)
				require.Contains(t, resources.Secrets, uiName)
				uiPdbName := uiName
				if release != "suse-observability" {
					uiPdbName = release + "-" + uiName
					assert.NotContains(t, resources.Pdbs, uiName)
				}
				require.Contains(t, resources.Pdbs, uiPdbName)
				require.Contains(t, resources.ServiceMonitors, uiName)
				ui := resources.Deployments[uiName]
				assert.Equal(t, release, ui.Spec.Template.Labels["app.kubernetes.io/instance"])
				for key, value := range resources.Services[uiName].Spec.Selector {
					assert.Equal(t, value, ui.Spec.Template.Labels[key])
				}
				assert.Contains(t, ui.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
					Name: "TEST_SECRET",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: uiName},
						Key:                  "TEST_SECRET",
					}},
				})
				for key, value := range resources.Pdbs[uiPdbName].Spec.Selector.MatchLabels {
					assert.Equal(t, value, ui.Spec.Template.Labels[key])
				}
				for key, value := range resources.ServiceMonitors[uiName].Spec.Selector.MatchLabels {
					assert.Equal(t, value, resources.Services[uiName].Labels[key])
				}

				routerName := "suse-observability-router-active"
				if release != "suse-observability" {
					routerName = release + "-" + routerName
				}
				require.Contains(t, resources.ConfigMaps, routerName)
				router := resources.ConfigMaps[routerName]
				assert.Contains(t, router.Data["listeners.yaml"], `cluster: "`+uiName+`"`)
				assert.Contains(t, router.Data["clusters.yaml"], `name: "`+uiName+`"`)
				assert.Contains(t, router.Data["clusters.yaml"], `address: "`+uiName+`"`)
				assert.NotContains(t, router.Data["clusters.yaml"], `address: "nightly-suse-observability-ui"`)

				const checkerName = "suse-observability-replication-checker"
				require.Contains(t, resources.Deployments, checkerName)
				require.Contains(t, resources.ServiceAccounts, checkerName)
				require.Contains(t, resources.Roles, checkerName)
				require.Contains(t, resources.RoleBindings, checkerName)
				checker := resources.Deployments[checkerName]
				assert.Equal(t, checkerName, checker.Spec.Template.Spec.ServiceAccountName)
				binding := resources.RoleBindings[checkerName]
				assert.Equal(t, checkerName, binding.RoleRef.Name)
				require.Len(t, binding.Subjects, 1)
				assert.Equal(t, checkerName, binding.Subjects[0].Name)
				assert.Equal(t, "observability", binding.Subjects[0].Namespace)
				assert.Equal(t, release, checker.Spec.Template.Labels["app.kubernetes.io/instance"])

				const vmagentName = "suse-observability-vmagent"
				require.Contains(t, resources.ConfigMaps, vmagentName)
				require.Contains(t, resources.Statefulsets, vmagentName)
				assert.Contains(t, resources.Statefulsets[vmagentName].Spec.Template.Spec.Volumes, corev1.Volume{
					Name: "vmagent-config",
					VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: vmagentName},
					}},
				})

				const s3proxyName = "suse-observability-s3proxy"
				require.Contains(t, resources.ServiceMonitors, s3proxyName)
				require.Contains(t, resources.Services, s3proxyName)
				monitor := resources.ServiceMonitors[s3proxyName]
				for key, value := range monitor.Spec.Selector.MatchLabels {
					assert.Equal(t, value, resources.Services[s3proxyName].Labels[key])
				}
				require.Len(t, monitor.Spec.Endpoints, 1)
				assert.Equal(t, resources.Services[s3proxyName].Spec.Ports[0].Name, monitor.Spec.Endpoints[0].Port)
				assert.Equal(t, "/metrics", monitor.Spec.Endpoints[0].Path)

				assert.NotContains(t, resources.Deployments, "nightly-suse-observability-ui")
				assert.NotContains(t, resources.Deployments, "nightly-suse-observability-replication-checker")
				assert.NotContains(t, resources.ConfigMaps, "nightly-suse-observability-vmagent")
			})
		}
	}
}

func TestResourceNamingPreservesStorageIdentity(t *testing.T) {
	for _, mode := range []string{"Distributed", "Mono"} {
		t.Run(mode, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", &helm.Options{
				ValuesFiles: []string{"values/full.yaml"},
				SetValues:   map[string]string{"hbase.deployment.mode": mode},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			hbaseComponents := []string{"hbase-master", "hbase-rs", "hdfs-dn", "hdfs-nn", "hdfs-snn", "tephra"}
			if mode == "Mono" {
				hbaseComponents = []string{"stackgraph", "tephra-mono"}
			}
			for _, component := range hbaseComponents {
				name := "nightly-hbase-" + component
				require.Contains(t, resources.Statefulsets, name)
				statefulset := resources.Statefulsets[name]
				serviceName := name
				if component == "tephra-mono" {
					serviceName = "nightly-hbase-tephra"
				}
				assert.Equal(t, serviceName, statefulset.Spec.ServiceName)
				switch component {
				case "hdfs-dn", "hdfs-nn", "hdfs-snn", "stackgraph":
					require.Len(t, statefulset.Spec.VolumeClaimTemplates, 1)
					assert.Equal(t, "data", statefulset.Spec.VolumeClaimTemplates[0].Name)
				case "tephra-mono":
					require.Len(t, statefulset.Spec.VolumeClaimTemplates, 1)
					assert.Equal(t, "snapshot", statefulset.Spec.VolumeClaimTemplates[0].Name)
				default:
					assert.Empty(t, statefulset.Spec.VolumeClaimTemplates)
				}
			}
			for _, name := range []string{
				"nightly-suse-observability-stackpacks",
				"nightly-suse-observability-settings-backup-data",
				"nightly-suse-observability-backup-settings-data",
				"suse-observability-minio",
			} {
				assert.Contains(t, resources.PersistentVolumeClaims, name)
			}
			require.Contains(t, resources.Statefulsets, "suse-observability-vmagent")
			vmagent := resources.Statefulsets["suse-observability-vmagent"]
			require.Len(t, vmagent.Spec.VolumeClaimTemplates, 1)
			assert.Equal(t, "tmpdata", vmagent.Spec.VolumeClaimTemplates[0].Name)
		})
	}
}

func TestResourceNamingVictoriaMetricsInstances(t *testing.T) {
	for _, release := range []string{"suse-observability", "nightly"} {
		for _, secondInstance := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/secondInstance=%t", release, secondInstance), func(t *testing.T) {
				output := helmtestutil.RenderHelmTemplateOptsNoError(t, release, &helm.Options{
					ValuesFiles: []string{"values/full.yaml"},
					SetValues: map[string]string{
						"victoria-metrics-1.enabled": fmt.Sprint(secondInstance),
					},
				})
				resources := helmtestutil.NewKubernetesResources(t, output)
				require.Contains(t, resources.Statefulsets, "suse-observability-vmagent")
				pod := resources.Statefulsets["suse-observability-vmagent"].Spec.Template.Spec
				require.NotEmpty(t, pod.Containers)
				require.NotEmpty(t, pod.InitContainers)
				initCommand := strings.Join(pod.InitContainers[0].Command, " ")
				scriptsName := "suse-observability-backup-restore-scripts"
				if release != "suse-observability" {
					scriptsName = release + "-" + scriptsName
				}
				require.Contains(t, resources.ConfigMaps, scriptsName)
				restoreTemplate := resources.ConfigMaps[scriptsName].Data["job-victoria-metrics-restore-backup.yaml"]
				require.NotEmpty(t, restoreTemplate)
				for index := 0; index < 2; index++ {
					name := fmt.Sprintf("suse-observability-victoria-metrics-%d", index)
					endpoint := name + ":8428"
					argument := "-remoteWrite.url=http://" + endpoint + "/api/v1/write"
					if index == 0 || secondInstance {
						require.Contains(t, resources.Services, name)
						assert.Contains(t, pod.Containers[0].Args, argument)
						assert.Contains(t, initCommand, endpoint)

						require.Contains(t, resources.Statefulsets, name)
						vm := resources.Statefulsets[name]
						require.Len(t, vm.Spec.VolumeClaimTemplates, 1)
						claimName := vm.Spec.VolumeClaimTemplates[0].Name + "-" + vm.Name + "-0"
						var restoreJob batchv1.Job
						rendered := strings.ReplaceAll(restoreTemplate, "REPLACE_ME_VICTORIA_METRICS_INSTANCE_NAME", fmt.Sprintf("victoria-metrics-%d", index))
						require.NoError(t, yaml.Unmarshal([]byte(rendered), &restoreJob))
						var claims []string
						for _, volume := range restoreJob.Spec.Template.Spec.Volumes {
							if volume.PersistentVolumeClaim != nil {
								claims = append(claims, volume.PersistentVolumeClaim.ClaimName)
							}
						}
						assert.Contains(t, claims, claimName, "restore must mount the existing StatefulSet's PVC")
					} else {
						assert.NotContains(t, resources.Services, name)
						assert.NotContains(t, pod.Containers[0].Args, argument)
						assert.NotContains(t, initCommand, endpoint)
					}
				}
			})
		}
	}
}

func TestResourceNamingBackupAndHTTPRoute(t *testing.T) {
	for _, release := range []string{"suse-observability", "nightly"} {
		t.Run(release, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, release, "values/full.yaml", "values/gateway_enabled.yaml")
			resources := helmtestutil.NewKubernetesResources(t, output)

			require.Contains(t, resources.ConfigMaps, "suse-observability-backup-config")
			require.Contains(t, resources.Secrets, "suse-observability-backup-config")
			require.Contains(t, resources.HTTPRoutes, "suse-observability")
			route := resources.HTTPRoutes["suse-observability"]
			require.Len(t, route.Spec.Rules, 1)
			require.Len(t, route.Spec.Rules[0].BackendRefs, 1)
			backendName := string(route.Spec.Rules[0].BackendRefs[0].Name)
			require.Contains(t, resources.Services, backendName)
			assert.Equal(t, release, route.Labels["app.kubernetes.io/instance"])

			const clickhouseName = "suse-observability-clickhouse"
			require.Contains(t, resources.ConfigMaps, clickhouseName+"-backup")
			require.Contains(t, resources.ConfigMaps, clickhouseName+"-backup-scripts")
			require.Contains(t, resources.Services, clickhouseName+"-backup")
			require.Contains(t, resources.Statefulsets, clickhouseName+"-shard0")
			backupService := resources.Services[clickhouseName+"-backup"]
			statefulset := resources.Statefulsets[clickhouseName+"-shard0"]
			assert.Equal(t, statefulset.Name+"-0", backupService.Spec.Selector["statefulset.kubernetes.io/pod-name"])
			for _, job := range []string{"full-backup", "incremental-backup"} {
				require.Contains(t, resources.CronJobs, clickhouseName+"-"+job)
				pod := resources.CronJobs[clickhouseName+"-"+job].Spec.JobTemplate.Spec.Template.Spec
				require.Len(t, pod.Volumes, 1)
				require.NotNil(t, pod.Volumes[0].ConfigMap)
				assert.Equal(t, clickhouseName+"-backup-scripts", pod.Volumes[0].ConfigMap.Name)
			}
		})
	}
}

func TestResourceNamingSharedPullSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "nightly", "values/global_suse_observability_pull_secret_ha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	const secretName = "suse-observability-pull-secret"
	require.Contains(t, resources.Secrets, secretName)
	for _, deployment := range resources.Deployments {
		assert.Contains(t, deployment.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: secretName}, deployment.Name)
	}
	for _, statefulset := range resources.Statefulsets {
		assert.Contains(t, statefulset.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: secretName}, statefulset.Name)
	}
}
