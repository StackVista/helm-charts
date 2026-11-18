package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

const base64IntrinsicsFlag = "-XX:-UseBASE64Intrinsics"

func TestJvmsDisableBase64Intrinsics(t *testing.T) {
	tests := []struct {
		name       string
		values     []string
		containers []string
	}{
		{
			name:   "ha",
			values: []string{"values/full.yaml"},
			containers: []string{
				"suse-observability-api/api",
				"suse-observability-receiver/receiver",
				"suse-observability-correlate/correlate",
				"suse-observability-e2es/e2es",
				"suse-observability-s3proxy/s3proxy",
				"suse-observability-hbase-console/console",
				"suse-observability-hbase-hbase-master/master",
				"suse-observability-hbase-hbase-rs/regionserver",
				"suse-observability-hbase-hdfs-dn/datanode",
				"suse-observability-hbase-hdfs-nn/namenode",
				"suse-observability-hbase-hdfs-snn/namenode",
				"suse-observability-hbase-tephra/tephra",
				"suse-observability-elasticsearch-master/elasticsearch",
				"suse-observability-kafka/kafka",
				"suse-observability-kafka/jmx-exporter",
				"suse-observability-zookeeper/zookeeper",
			},
		},
		{
			name:   "nonha",
			values: []string{"values/full.yaml", "values/default_nonha_sizing.yaml"},
			containers: []string{
				"suse-observability-server/server",
				"suse-observability-hbase-stackgraph/stackgraph",
				"suse-observability-hbase-tephra-mono/tephra",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", tc.values...)
			containers := workloadContainers(helmtestutil.NewKubernetesResources(t, output))
			for _, name := range tc.containers {
				c, ok := containers[name]
				if assert.True(t, ok, "container %s should exist", name) {
					assert.True(t, carriesFlag(c), "container %s should pass %s to the JVM", name, base64IntrinsicsFlag)
				}
			}
		})
	}
}

func TestMonoStackgraphSetsJdkJavaOptions(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/default_nonha_sizing.yaml")
	c := workloadContainers(helmtestutil.NewKubernetesResources(t, output))["suse-observability-hbase-stackgraph/stackgraph"]
	assert.Contains(t, c.Env, corev1.EnvVar{Name: "JDK_JAVA_OPTIONS", Value: base64IntrinsicsFlag})
}

func workloadContainers(resources helmtestutil.KubernetesResources) map[string]corev1.Container {
	result := map[string]corev1.Container{}
	add := func(owner string, spec corev1.PodSpec) {
		for _, c := range spec.Containers {
			result[owner+"/"+c.Name] = c
		}
	}
	for name, d := range resources.Deployments {
		add(name, d.Spec.Template.Spec)
	}
	for name, s := range resources.Statefulsets {
		add(name, s.Spec.Template.Spec)
	}
	return result
}

func carriesFlag(c corev1.Container) bool {
	for _, e := range c.Env {
		if strings.Contains(e.Value, base64IntrinsicsFlag) {
			return true
		}
	}
	for _, a := range c.Args {
		if a == base64IntrinsicsFlag {
			return true
		}
	}
	return false
}
