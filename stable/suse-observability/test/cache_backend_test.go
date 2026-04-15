package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestSyncCacheBackendIsMapDb(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "suse-observability-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_cacheStorage_backend", Value: "mapdb"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expected)
}

func TestHealthSyncCacheBackendIsMapDb(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "suse-observability-health-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_cacheStorage_backend", Value: "mapdb"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expected)
}
