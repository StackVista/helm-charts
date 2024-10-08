package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestK2ESDiskSpaceRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateE2esDeployment appsv1.Deployment
	var stackstateReceiverDeployment appsv1.Deployment

	for _, deploymentK2ES := range resources.Deployments {
		if deploymentK2ES.Name == "suse-observability-e2es" {
			stackstateE2esDeployment = deploymentK2ES
		}

		if deploymentK2ES.Name == "suse-observability-receiver" {
			stackstateReceiverDeployment = deploymentK2ES
		}
	}

	require.NotNil(t, stackstateE2esDeployment)
	require.NotNil(t, stackstateReceiverDeployment)

	expectedDiskE2esSpace := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_elasticsearchDiskSpaceMB", Value: "40265"}
	expectedTeWeightsPerIndex := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_kafkaTopologyEventsToES_elasticsearch_index_diskSpaceWeight", Value: "100"}
	expectedDiskReceiverSpace := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_receiver_elasticsearchDiskSpaceMB", Value: "362387"}

	require.Contains(t, stackstateE2esDeployment.Spec.Template.Spec.Containers[0].Env, expectedDiskE2esSpace)
	require.Contains(t, stackstateE2esDeployment.Spec.Template.Spec.Containers[0].Env, expectedTeWeightsPerIndex)
	require.Contains(t, stackstateE2esDeployment.Spec.Template.Spec.Containers[0].Env, expectedDiskE2esSpace)

	require.Contains(t, stackstateReceiverDeployment.Spec.Template.Spec.Containers[0].Env, expectedDiskReceiverSpace)
}

func TestK2ESDiskSpaceBudgetFail(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/e2es_diskWeight_fail.yaml")
	require.Contains(t, err.Error(), "The share of ElasticSearch disk on receiver.esDiskSpaceShare, e2es.esDiskSpaceShare should be 100.")
}

func TestUnknownK2ESDiskSpaceRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/unknown_disk_unit.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateE2esDeployment appsv1.Deployment

	for _, deploymentK2ES := range resources.Deployments {
		if deploymentK2ES.Name == "suse-observability-e2es" {
			stackstateE2esDeployment = deploymentK2ES
		}
	}

	require.NotNil(t, stackstateE2esDeployment)

	var stackstateDiskSpaceVarEntry v1.EnvVar

	for _, envVarEntry := range stackstateE2esDeployment.Spec.Template.Spec.Containers[0].Env {
		if envVarEntry.Name == "CONFIG_FORCE_stackstate_elasticsearchDiskSpaceMB" {
			stackstateDiskSpaceVarEntry = envVarEntry
		}
	}

	expectedDiskSpace := v1.EnvVar{Name: "", Value: ""}

	require.Equal(t, stackstateDiskSpaceVarEntry, expectedDiskSpace)

}
