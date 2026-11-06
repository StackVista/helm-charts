package test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

const podMetadataKey = "test.stackstate.com/pod-metadata"

func TestPodMetadataEveryTemplate(t *testing.T) {
	for _, profile := range []string{"10-nonha", "150-ha"} {
		t.Run(profile, func(t *testing.T) {
			values := map[string]string{
				"global.suseObservability.adminPassword":                                "test-password",
				"global.suseObservability.baseUrl":                                      "http://localhost",
				"global.suseObservability.license":                                      "test-license",
				"global.suseObservability.sizing.profile":                               profile,
				"global.commonLabels." + strings.ReplaceAll(podMetadataKey, ".", "\\."): "configured",
				"hbase.console.integrity.enabled":                                       "true",
				"anomaly-detection.enabled":                                             "true",
			}
			addPodAnnotationValues(t, "../values.yaml", "", values)

			chartData, err := os.ReadFile("../Chart.yaml")
			require.NoError(t, err)
			var chart struct {
				Dependencies []struct {
					Name       string
					Alias      string
					Repository string
				}
			}
			require.NoError(t, yaml.Unmarshal(chartData, &chart))
			for _, dependency := range chart.Dependencies {
				name := dependency.Name
				if dependency.Alias != "" {
					name = dependency.Alias
				}
				require.True(t, strings.HasPrefix(dependency.Repository, "file://"))
				path := filepath.Join("..", strings.TrimPrefix(dependency.Repository, "file://"), "values.yaml")
				if _, err := os.Stat(path); os.IsNotExist(err) {
					continue
				}
				addPodAnnotationValues(t, path, name, values)
			}

			output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{"values/full.yaml"},
				SetValues:   values,
			}, "--skip-tests")
			resources := helmtestutil.NewKubernetesResources(t, output)
			templates := renderedPodTemplates(&resources)
			require.NotEmpty(t, templates)
			for name, pod := range templates {
				t.Run(name, func(t *testing.T) {
					annotated := false
					for key, value := range pod.Annotations {
						if strings.HasPrefix(key, podMetadataKey) && value == "configured" {
							annotated = true
						}
					}
					assert.True(t, annotated, "pod annotation")
					assert.Equal(t, "configured", pod.Labels[podMetadataKey], "pod label")
				})
			}
			require.Contains(t, templates, "Deployment/suse-observability-hbase-console")
			require.Contains(t, templates, "Deployment/suse-observability-kafkaup-operator-kafkaup")
			require.Contains(t, templates, "StatefulSet/suse-observability-vmagent")
			require.Contains(t, templates, "CronJob/suse-observability-hbase-integrity")
			require.Contains(t, templates, "Deployment/suse-observability-anomaly-detection-spotlight-manager")
			require.Contains(t, templates, "Deployment/suse-observability-anomaly-detection-spotlight-worker")
			if profile == "150-ha" {
				for _, component := range []string{"hbase-master", "hbase-rs", "hdfs-dn", "hdfs-nn", "hdfs-snn", "tephra"} {
					require.Contains(t, templates, "StatefulSet/suse-observability-hbase-"+component)
				}
			} else {
				require.Contains(t, templates, "StatefulSet/suse-observability-hbase-stackgraph")
				require.Contains(t, templates, "StatefulSet/suse-observability-hbase-tephra-mono")
			}
		})
	}
}

func TestPodMetadataComponentOverrides(t *testing.T) {
	hbaseComponents := map[string]string{
		"console":      "console",
		"stackgraph":   "stackgraph",
		"tephra":       "tephra",
		"hbase-master": "hbase.master",
		"hbase-rs":     "hbase.regionserver",
		"hdfs-dn":      "hdfs.datanode",
		"hdfs-nn":      "hdfs.namenode",
		"hdfs-snn":     "hdfs.secondarynamenode",
	}
	for _, profile := range []string{"10-nonha", "150-ha"} {
		t.Run(profile, func(t *testing.T) {
			values := map[string]string{
				"global.suseObservability.adminPassword":                  "test-password",
				"global.suseObservability.baseUrl":                        "http://localhost",
				"global.suseObservability.license":                        "test-license",
				"global.suseObservability.sizing.profile":                 profile,
				"global.commonLabels.test-global":                         "global",
				"global.commonLabels.test-override":                       "global",
				"hbase.console.integrity.enabled":                         "true",
				"hbase.all.podAnnotations.test-shared":                    "shared",
				"hbase.all.podAnnotations.test-override":                  "shared",
				"hbase.all.podLabels.test-shared":                         "shared",
				"hbase.all.podLabels.test-override":                       "shared",
				"stackstate.components.all.podAnnotations.test-shared":    "shared",
				"stackstate.components.vmagent.podAnnotations.test-local": "vmagent",
				"kafkaup-operator.podAnnotations.test-local":              "kafkaup",
				"kafkaup-operator.podLabels.test-local":                   "kafkaup",
				"kafkaup-operator.commonLabels.test-local":                "common",
				"kafkaup-operator.commonAnnotations.test-resource":        "resource",
			}
			for name, component := range hbaseComponents {
				for _, field := range []string{"podAnnotations", "podLabels"} {
					values["hbase."+component+"."+field+".test-override"] = name
					values["hbase."+component+"."+field+".test-local"] = name
				}
			}
			output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{"values/full.yaml"},
				SetValues:   values,
			}, "--skip-tests")
			resources := helmtestutil.NewKubernetesResources(t, output)
			checked := 0
			for name, pod := range renderedPodTemplates(&resources) {
				if !strings.Contains(name, "/suse-observability-hbase-") {
					continue
				}
				checked++
				component := strings.TrimPrefix(strings.SplitN(name, "/", 2)[1], "suse-observability-hbase-")
				if component == "integrity" {
					component = "console"
				}
				if component == "tephra-mono" {
					component = "tephra"
				}
				require.Contains(t, hbaseComponents, component)
				assert.Equal(t, "shared", pod.Annotations["test-shared"], name)
				assert.Equal(t, component, pod.Annotations["test-override"], name)
				assert.Equal(t, component, pod.Annotations["test-local"], name)
				assert.Equal(t, "global", pod.Labels["test-global"], name)
				assert.Equal(t, "shared", pod.Labels["test-shared"], name)
				assert.Equal(t, component, pod.Labels["test-override"], name)
				assert.Equal(t, component, pod.Labels["test-local"], name)
				if strings.HasPrefix(name, "StatefulSet/") {
					metricsContainer := map[string]string{"stackgraph": "stackgraph", "tephra": "tephra", "hbase-master": "master", "hbase-rs": "regionserver", "hdfs-dn": "datanode", "hdfs-nn": "namenode", "hdfs-snn": "namenode"}[component]
					assert.Equal(t, `["openmetrics"]`, pod.Annotations["ad.stackstate.com/"+metricsContainer+".check_names"], name)
					statefulset := resources.Statefulsets[strings.TrimPrefix(name, "StatefulSet/")]
					for key, value := range statefulset.Spec.Selector.MatchLabels {
						assert.Equal(t, value, pod.Labels[key], name)
					}
				}
			}
			if profile == "150-ha" {
				assert.Equal(t, 8, checked)
			} else {
				assert.Equal(t, 4, checked)
			}
			vmagent := resources.Statefulsets["suse-observability-vmagent"]
			assert.Equal(t, "shared", vmagent.Spec.Template.Annotations["test-shared"])
			assert.Equal(t, "vmagent", vmagent.Spec.Template.Annotations["test-local"])
			assert.NotEmpty(t, vmagent.Spec.Template.Annotations["checksum/vmagent-configmap"])
			kafkaup := resources.Deployments["suse-observability-kafkaup-operator-kafkaup"]
			assert.Equal(t, "kafkaup", kafkaup.Spec.Template.Annotations["test-local"])
			assert.Equal(t, "kafkaup", kafkaup.Spec.Template.Labels["test-local"])
			assert.NotEmpty(t, kafkaup.Spec.Template.Annotations["checksum/configmap"])
			assert.Equal(t, "resource", kafkaup.Annotations["test-resource"])
			assert.NotContains(t, kafkaup.Spec.Template.Annotations, "test-resource")
		})
	}
}

func TestPodMetadataWithoutGlobalLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"hbase.all.podAnnotations.test-shared":    "shared",
			"hbase.all.podLabels.test-shared":         "shared",
			"hbase.console.podAnnotations.test-local": "console",
			"hbase.console.podLabels.test-local":      "console",
			"kafkaup-operator.podLabels.test-local":   "kafkaup",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	console := resources.Deployments["suse-observability-hbase-console"]
	assert.Equal(t, "console", console.Spec.Template.Annotations["test-local"])
	assert.Equal(t, "console", console.Spec.Template.Labels["test-local"])
	for name, pod := range renderedPodTemplates(&resources) {
		if strings.Contains(name, "/suse-observability-hbase-") {
			assert.Equal(t, "shared", pod.Annotations["test-shared"], name)
			assert.Equal(t, "shared", pod.Labels["test-shared"], name)
			if name != "Deployment/suse-observability-hbase-console" {
				assert.NotContains(t, pod.Annotations, "test-local", name)
				assert.NotContains(t, pod.Labels, "test-local", name)
			}
		}
	}
	kafkaup := resources.Deployments["suse-observability-kafkaup-operator-kafkaup"]
	assert.Equal(t, "kafkaup", kafkaup.Spec.Template.Labels["test-local"])
}

func addPodAnnotationValues(t *testing.T, path, prefix string, values map[string]string) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var defaults map[string]interface{}
	require.NoError(t, yaml.Unmarshal(data, &defaults))
	var visit func(map[string]interface{}, string)
	visit = func(node map[string]interface{}, prefix string) {
		for key, value := range node {
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			if key == "podAnnotations" {
				// Existing helpers append shared and component maps, so use distinct keys.
				key := podMetadataKey + "-" + strconv.Itoa(len(values))
				values[path+"."+strings.ReplaceAll(key, ".", "\\.")] = "configured"
			} else if child, ok := value.(map[string]interface{}); ok {
				visit(child, path)
			}
		}
	}
	visit(defaults, prefix)
}

func renderedPodTemplates(resources *helmtestutil.KubernetesResources) map[string]corev1.PodTemplateSpec {
	templates := make(map[string]corev1.PodTemplateSpec)
	for name, resource := range resources.Deployments {
		templates["Deployment/"+name] = resource.Spec.Template
	}
	for name, resource := range resources.Statefulsets {
		templates["StatefulSet/"+name] = resource.Spec.Template
	}
	for name, resource := range resources.DaemonSets {
		templates["DaemonSet/"+name] = resource.Spec.Template
	}
	for name, resource := range resources.Jobs {
		templates["Job/"+name] = resource.Spec.Template
	}
	for name, resource := range resources.CronJobs {
		templates["CronJob/"+name] = resource.Spec.JobTemplate.Spec.Template
	}
	for name, resource := range resources.Pods {
		templates["Pod/"+name] = corev1.PodTemplateSpec{ObjectMeta: resource.ObjectMeta, Spec: resource.Spec}
	}
	return templates
}
