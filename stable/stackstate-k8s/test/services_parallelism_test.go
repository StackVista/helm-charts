package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestSyncParalellismConfigRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expectedSyncJavaOpts := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_sync_parallelWorkers", Value: "2"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expectedSyncJavaOpts)
}

func TestSyncParalellismConfigRenderOnMilliCpu(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/sync_parallelism_12001.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expectedSyncJavaOpts := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_sync_parallelWorkers", Value: "12"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expectedSyncJavaOpts)
}

func TestSyncParalellismConfigRenderOnCpu(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/sync_parallelism_12.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expectedSyncJavaOpts := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_sync_parallelWorkers", Value: "12"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expectedSyncJavaOpts)
}

func TestSyncParalellismConfigRenderNotOverridable(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/sync_parallelism_12.yaml", "values/sync_env_parallel_workers.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expectedSyncJavaOpts := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_sync_parallelWorkers", Value: "12"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expectedSyncJavaOpts)
}
