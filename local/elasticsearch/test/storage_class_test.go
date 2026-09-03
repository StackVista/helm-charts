package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestElasticsearchStaticProvisioningStorageClass(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "elasticsearch", &helm.Options{
		ValuesFiles: []string{"values/default.yaml"},
		SetValues: map[string]string{
			"global.storageClass":                            "global",
			"persistence.enabled":                            "true",
			"volumeClaimTemplate.resources.requests.storage": "1Gi",
			"volumeClaimTemplate.storageClassName":           "-",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	require.Len(t, resources.Statefulsets, 1)

	for _, statefulSet := range resources.Statefulsets {
		require.Len(t, statefulSet.Spec.VolumeClaimTemplates, 1)
		storageClass := statefulSet.Spec.VolumeClaimTemplates[0].Spec.StorageClassName
		require.NotNil(t, storageClass)
		assert.Empty(t, *storageClass)
	}
}
