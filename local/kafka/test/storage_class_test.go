package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestKafkaStorageClassPrecedence(t *testing.T) {
	testCases := []struct {
		name     string
		values   map[string]string
		expected map[string]*string
	}{
		{
			name: "cluster default",
			values: map[string]string{
				"logPersistence.enabled": "true",
			},
			expected: map[string]*string{
				"data": nil,
				"logs": nil,
			},
		},
		{
			name: "global fallback",
			values: map[string]string{
				"global.storageClass":    "global",
				"logPersistence.enabled": "true",
			},
			expected: map[string]*string{
				"data": stringPointer("global"),
				"logs": stringPointer("global"),
			},
		},
		{
			name: "per volume overrides global",
			values: map[string]string{
				"global.storageClass":         "global",
				"persistence.storageClass":    "data",
				"logPersistence.enabled":      "true",
				"logPersistence.storageClass": "logs",
			},
			expected: map[string]*string{
				"data": stringPointer("data"),
				"logs": stringPointer("logs"),
			},
		},
		{
			name: "static provisioning overrides global",
			values: map[string]string{
				"global.storageClass":         "global",
				"persistence.storageClass":    "-",
				"logPersistence.enabled":      "true",
				"logPersistence.storageClass": "-",
			},
			expected: map[string]*string{
				"data": stringPointer(""),
				"logs": stringPointer(""),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "kafka", &helm.Options{
				ValuesFiles: []string{"values/default.yaml"},
				SetValues:   tc.values,
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			statefulSet, ok := resources.Statefulsets["suse-observability-kafka"]
			require.True(t, ok)
			require.Len(t, statefulSet.Spec.VolumeClaimTemplates, len(tc.expected))

			for _, volumeClaim := range statefulSet.Spec.VolumeClaimTemplates {
				expected, ok := tc.expected[volumeClaim.Name]
				require.True(t, ok, "unexpected volume claim %q", volumeClaim.Name)
				assert.Equal(t, expected, volumeClaim.Spec.StorageClassName)
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}
