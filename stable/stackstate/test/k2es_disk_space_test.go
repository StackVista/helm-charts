package test

import (
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestK2ESDiskSpaceRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateMm2EsDeployment appsv1.Deployment
	var stackstateReceiverDeployment appsv1.Deployment

	for _, deploymentK2ES := range resources.Deployments {
		if deploymentK2ES.Name == "stackstate-mm2es" {
			stackstateMm2EsDeployment = deploymentK2ES
		}

		if deploymentK2ES.Name == "stackstate-receiver" {
			stackstateReceiverDeployment = deploymentK2ES
		}
	}

	require.NotNil(t, stackstateMm2EsDeployment)
	require.NotNil(t, stackstateReceiverDeployment)

	expectedDiskMm2EsSpace := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_elasticsearchDiskSpaceMB", Value: "281925"}
	expectedDiskReceiverSpace := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_receiver_elasticsearchDiskSpaceMB", Value: "120825"}

	require.Contains(t, stackstateMm2EsDeployment.Spec.Template.Spec.Containers[0].Env, expectedDiskMm2EsSpace)
	require.Contains(t, stackstateReceiverDeployment.Spec.Template.Spec.Containers[0].Env, expectedDiskReceiverSpace)
}

func TestUnknownK2ESDiskSpaceRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/unknown_disk_unit.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateMm2EsDeployment appsv1.Deployment

	for _, deploymentK2ES := range resources.Deployments {
		if deploymentK2ES.Name == "stackstate-mm2es" {
			stackstateMm2EsDeployment = deploymentK2ES
		}
	}

	require.NotNil(t, stackstateMm2EsDeployment)


	var stackstateDiskSpaceVarEntry v1.EnvVar

	for _, envVarEntry := range stackstateMm2EsDeployment.Spec.Template.Spec.Containers[0].Env {
		if 	envVarEntry.Name == "CONFIG_FORCE_stackstate_elasticsearchDiskSpaceMB" {
			stackstateDiskSpaceVarEntry = envVarEntry
		}
	}

	expectedDiskSpace := v1.EnvVar{Name: "", Value: ""}

	require.Equal(t, stackstateDiskSpaceVarEntry, expectedDiskSpace)

}
