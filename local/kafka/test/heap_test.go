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

func TestKafkaHeapFromMemoryRequest(t *testing.T) {
	tests := []struct {
		name          string
		values        map[string]string
		memoryRequest string
		heapOpts      string
	}{
		{
			name:          "default",
			memoryRequest: "2Gi",
			heapOpts:      "-Xmx1024m -Xms1024m",
		},
		{
			name: "request-override",
			values: map[string]string{
				"global.suseObservability.sizing.profile": "150-ha",
				"resources.requests.memory":               "2Gi",
			},
			memoryRequest: "2Gi",
			heapOpts:      "-Xmx1024m -Xms1024m",
		},
		{
			name: "limit-only-override",
			values: map[string]string{
				"global.suseObservability.sizing.profile": "150-ha",
				"resources.limits.memory":                 "8Gi",
			},
			memoryRequest: "3Gi",
			heapOpts:      "-Xmx1536m -Xms1536m",
		},
		{
			name: "cpu-only-override",
			values: map[string]string{
				"global.suseObservability.sizing.profile": "4000-ha",
				"resources.requests.cpu":                  "5",
			},
			memoryRequest: "6Gi",
			heapOpts:      "-Xmx3072m -Xms3072m",
		},
		{
			name: "mebibytes",
			values: map[string]string{
				"resources.requests.memory": "1537Mi",
			},
			memoryRequest: "1537Mi",
			heapOpts:      "-Xmx768m -Xms768m",
		},
		{
			name: "fractional-gibibytes",
			values: map[string]string{
				"resources.requests.memory": "1.5Gi",
			},
			memoryRequest: "1.5Gi",
			heapOpts:      "-Xmx768m -Xms768m",
		},
		{
			name: "decimal-megabytes",
			values: map[string]string{
				"resources.requests.memory": "2000M",
			},
			memoryRequest: "2000M",
			heapOpts:      "-Xmx953m -Xms953m",
		},
		{
			name: "bytes",
			values: map[string]string{
				"resources.requests.memory": "2147483648",
			},
			memoryRequest: "2147483648",
			heapOpts:      "-Xmx1024m -Xms1024m",
		},
		{
			name: "scientific-notation",
			values: map[string]string{
				"resources.requests.memory": "2e9",
			},
			memoryRequest: "2e9",
			heapOpts:      "-Xmx953m -Xms953m",
		},
		{
			name: "explicit-heap",
			values: map[string]string{
				"global.suseObservability.sizing.profile": "4000-ha",
				"heapOpts": "-Xmx768m -Xms512m",
			},
			memoryRequest: "6Gi",
			heapOpts:      "-Xmx768m -Xms512m",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "kafka", &helm.Options{
				SetValues: test.values,
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			require.Len(t, resources.Statefulsets, 1)
			for _, statefulset := range resources.Statefulsets {
				container := statefulset.Spec.Template.Spec.Containers[0]
				assert.Equal(t, resource.MustParse(test.memoryRequest), container.Resources.Requests[corev1.ResourceMemory])
				assert.Contains(t, container.Env, corev1.EnvVar{Name: "KAFKA_HEAP_OPTS", Value: test.heapOpts})
			}
		})
	}
}

func TestKafkaHeapRejectsMemoryRequestTooSmall(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "kafka", &helm.Options{
		SetValues: map[string]string{
			"resources.requests.memory": "1Mi",
		},
	})
	require.ErrorContains(t, err, "Kafka memory request must be at least 2Mi")
}
