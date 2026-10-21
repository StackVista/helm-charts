package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestKafkaHeapBySizingProfile(t *testing.T) {
	tests := []struct {
		profile       string
		memoryRequest string
		heapOpts      string
	}{
		{"trial", "2048Mi", "-Xmx1024m -Xms1024m"},
		{"10-nonha", "2048Mi", "-Xmx1024m -Xms1024m"},
		{"20-nonha", "2048Mi", "-Xmx1024m -Xms1024m"},
		{"50-nonha", "2048Mi", "-Xmx1024m -Xms1024m"},
		{"100-nonha", "3000Mi", "-Xmx1500m -Xms1500m"},
		{"150-ha", "3Gi", "-Xmx1536m -Xms1536m"},
		{"250-ha", "3Gi", "-Xmx1536m -Xms1536m"},
		{"500-ha", "3Gi", "-Xmx1536m -Xms1536m"},
		{"4000-ha", "6Gi", "-Xmx3072m -Xms3072m"},
	}

	for _, test := range tests {
		t.Run(test.profile, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{"values/global_sizing_trial.yaml"},
				SetValues: map[string]string{
					"global.suseObservability.sizing.profile": test.profile,
				},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			statefulset, ok := resources.Statefulsets["suse-observability-kafka"]
			require.True(t, ok)
			container := statefulset.Spec.Template.Spec.Containers[0]
			assert.Equal(t, resource.MustParse(test.memoryRequest), container.Resources.Requests[corev1.ResourceMemory])
			assert.Contains(t, container.Env, corev1.EnvVar{Name: "KAFKA_HEAP_OPTS", Value: test.heapOpts})
		})
	}
}

func TestKafkaHeapOverrides(t *testing.T) {
	tests := []struct {
		name     string
		values   map[string]string
		heapOpts string
	}{
		{
			name: "memory-request",
			values: map[string]string{
				"kafka.resources.requests.memory": "2Gi",
			},
			heapOpts: "-Xmx1024m -Xms1024m",
		},
		{
			name: "memory-limit",
			values: map[string]string{
				"kafka.resources.limits.memory": "8Gi",
			},
			heapOpts: "-Xmx1536m -Xms1536m",
		},
		{
			name: "explicit-heap",
			values: map[string]string{
				"kafka.resources.requests.memory": "2Gi",
				"kafka.heapOpts":                  "-Xmx768m -Xms512m",
			},
			heapOpts: "-Xmx768m -Xms512m",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.values["global.suseObservability.sizing.profile"] = "150-ha"
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{"values/global_sizing_trial.yaml"},
				SetValues:   test.values,
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			statefulset, ok := resources.Statefulsets["suse-observability-kafka"]
			require.True(t, ok)
			assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env,
				corev1.EnvVar{Name: "KAFKA_HEAP_OPTS", Value: test.heapOpts})
		})
	}
}
