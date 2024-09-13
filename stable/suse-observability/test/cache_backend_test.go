package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestSyncWithInMemoryCache(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/sync_inmemory.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-k8s-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_cacheStorage_backend", Value: "inmemory"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expected)
}

func TestSyncWithMapDbCache(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/sync_mapdb.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-k8s-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_cacheStorage_backend", Value: "mapdb"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expected)
}

func TestSyncWithRocksDbCache(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/sync_rocksdb.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-k8s-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_cacheStorage_backend", Value: "rocksdb"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expected)
	expectedBytes := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_rocksdb_cacheSizeBytes", Value: "1717984000"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expectedBytes)
}

func TestHealthSyncWithInMemoryCache(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/healthsync_inmemory.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-k8s-health-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_cacheStorage_backend", Value: "inmemory"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expected)
}

func TestHealthSyncWithMapDbCache(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/healthsync_mapdb.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-k8s-health-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_cacheStorage_backend", Value: "mapdb"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expected)
}

func TestHealthSyncWithRocksDbCache(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/healthsync_rocksdb.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsSyncDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-k8s-health-sync" {
			stsSyncDeployment = deployment
		}
	}
	assert.NotNil(t, stsSyncDeployment)
	expected := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_cacheStorage_backend", Value: "rocksdb"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expected)
	expectedBytes := v1.EnvVar{Name: "CONFIG_FORCE_stackstate_rocksdb_cacheSizeBytes", Value: "2018508800"}
	assert.Contains(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expectedBytes)
}
