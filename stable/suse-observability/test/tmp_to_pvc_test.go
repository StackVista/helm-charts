package test

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	v1 "k8s.io/api/core/v1"
)

const (
	customSyncTmpPvcStorageClass       = "TestSyncStorageClass"
	customSyncTmpPvcVolumeSize         = "999Ti"
	customHealthSyncTmpPvcStorageClass = "TestHealthSyncStorageClass"
	customHealthSyncTmpPvcVolumeSize   = "666Ti"
	customStateTmpPvcStorageClass      = "TestStateStorageClass"
	customStateTmpPvcVolumeSize        = "333Ti"
	customChecksTmpPvcStorageClass     = "TestChecksStorageClass"
	customChecksTmpPvcVolumeSize       = "222Ti"
)

var tmpPvcComponents = []string{
	"suse-observability-checks",
	"suse-observability-health-sync",
	"suse-observability-state",
	"suse-observability-sync",
}

func TestTmpToPVCDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	for _, component := range tmpPvcComponents {
		deploymentHasTmpPVC(t, &resources, component)
		pvcName := component + "-tmp"
		assert.Contains(t, resources.PersistentVolumeClaims, pvcName, "PVC '%s' should exist", pvcName)
	}
}

func TestTmpToPVCDefaultSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.server.split": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	for _, component := range tmpPvcComponents {
		deploymentHasTmpPVC(t, &resources, component)
		pvcName := component + "-tmp"
		assert.Contains(t, resources.PersistentVolumeClaims, pvcName, "PVC '%s' should exist", pvcName)
	}
}

func TestTmpToPVCDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.server.split":          "true",
			"stackstate.components.checks.tmpToPVC":     "",
			"stackstate.components.healthSync.tmpToPVC": "",
			"stackstate.components.state.tmpToPVC":      "",
			"stackstate.components.sync.tmpToPVC":       "",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	for _, component := range tmpPvcComponents {
		deploymentDoesntHaveTmpPVC(t, &resources, component)
		pvcName := component + "-tmp"
		assert.NotContains(t, resources.PersistentVolumeClaims, pvcName, "PVC '%s' should not exist", pvcName)
	}
}

func TestTmpToPVCMonolith(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.server.split": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	for _, component := range tmpPvcComponents {
		pvcName := component + "-tmp"
		assert.NotContains(t, resources.PersistentVolumeClaims, pvcName, "PVC '%s' should not exist", pvcName)
	}
}

func TestTmpToPVCCustomSettings(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.server.split":                       "true",
			"stackstate.components.checks.tmpToPVC.volumeSize":       customChecksTmpPvcVolumeSize,
			"stackstate.components.checks.tmpToPVC.storageClass":     customChecksTmpPvcStorageClass,
			"stackstate.components.healthSync.tmpToPVC.volumeSize":   customHealthSyncTmpPvcVolumeSize,
			"stackstate.components.healthSync.tmpToPVC.storageClass": customHealthSyncTmpPvcStorageClass,
			"stackstate.components.state.tmpToPVC.volumeSize":        customStateTmpPvcVolumeSize,
			"stackstate.components.state.tmpToPVC.storageClass":      customStateTmpPvcStorageClass,
			"stackstate.components.sync.tmpToPVC.volumeSize":         customSyncTmpPvcVolumeSize,
			"stackstate.components.sync.tmpToPVC.storageClass":       customSyncTmpPvcStorageClass,
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	pvcName := "suse-observability-checks-tmp"
	if ok := assert.Contains(t, resources.PersistentVolumeClaims, pvcName); ok {
		checkPvc(t, resources.PersistentVolumeClaims[pvcName], customChecksTmpPvcStorageClass, customChecksTmpPvcVolumeSize)
	}
	pvcName = "suse-observability-health-sync-tmp"
	if ok := assert.Contains(t, resources.PersistentVolumeClaims, pvcName); ok {
		checkPvc(t, resources.PersistentVolumeClaims[pvcName], customHealthSyncTmpPvcStorageClass, customHealthSyncTmpPvcVolumeSize)
	}
	pvcName = "suse-observability-state-tmp"
	if ok := assert.Contains(t, resources.PersistentVolumeClaims, pvcName); ok {
		checkPvc(t, resources.PersistentVolumeClaims[pvcName], customStateTmpPvcStorageClass, customStateTmpPvcVolumeSize)
	}
	pvcName = "suse-observability-sync-tmp"
	if ok := assert.Contains(t, resources.PersistentVolumeClaims, pvcName); ok {
		checkPvc(t, resources.PersistentVolumeClaims[pvcName], customSyncTmpPvcStorageClass, customSyncTmpPvcVolumeSize)
	}
}

// deploymentHasTmpPVC validates that a deployment has the expected volume and mount configuration
func deploymentHasTmpPVC(t *testing.T, resources *helmtestutil.KubernetesResources, deploymentName string) {
	deployment, ok := resources.Deployments[deploymentName]
	require.True(t, ok, "Deployment '%s' should exist", deploymentName)

	expectedClaimName := fmt.Sprintf("%s-tmp", deploymentName)
	expectedVolume := v1.Volume{
		Name: "tmp-volume",
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: expectedClaimName,
			},
		},
	}
	assert.Contains(t, deployment.Spec.Template.Spec.Volumes, expectedVolume, "Deployment '%s' should have the tmp PVC volume", deploymentName)

	expectedVolumeMount := v1.VolumeMount{
		Name:      "tmp-volume",
		MountPath: "/tmp",
	}
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts, expectedVolumeMount, "Deployment '%s' container should have the tmp volume mount", deploymentName)

	var foundInitContainer bool
	for _, initContainer := range deployment.Spec.Template.Spec.InitContainers {
		if initContainer.Name == "clean-tmp-directory" {
			foundInitContainer = true
			assert.Contains(t, initContainer.VolumeMounts, expectedVolumeMount, "Init container 'clean-transaction-logs-directory' should have the tmp volume mount")
			break
		}
	}
	assert.True(t, foundInitContainer, "Deployment '%s' should have 'clean-tmp-directory' init container when using tmp PVC", deploymentName)
}

// deploymentDoesntHaveTmpPVC validates that a deployment does not have the expected volume and mount configuration
func deploymentDoesntHaveTmpPVC(t *testing.T, resources *helmtestutil.KubernetesResources, deploymentName string) {
	deployment, ok := resources.Deployments[deploymentName]
	require.True(t, ok, "Deployment '%s' should exist", deploymentName)

	expectedClaimName := fmt.Sprintf("%s-tmp", deploymentName)
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == expectedClaimName {
			t.Errorf("Deployment '%s' should not have a volume with claim name '%s'", deploymentName, expectedClaimName)
		}
	}

	mountPathName := "tmp-volume"

	for _, volumeMount := range deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
		if volumeMount.Name == mountPathName {
			t.Errorf("Deployment '%s' should not have a mountPath '%s'", deploymentName, mountPathName)
		}
	}

	for _, initContainer := range deployment.Spec.Template.Spec.InitContainers {
		if initContainer.Name == "clean-tmp-directory" {
			t.Errorf("Deployment '%s' should not have a init container '%s'", deploymentName, initContainer.Name)
			break
		}
	}
}

func checkPvc(t *testing.T, pvc v1.PersistentVolumeClaim, expectedStorageClass, expectedVolumeSize string) {
	storageRequest, ok := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	require.True(t, ok, "PVC '%s' should have storage resource request defined", pvc.Name)
	assert.Equal(t, expectedVolumeSize, storageRequest.String(), "PVC '%s' should have the expected storage size", pvc.Name)
	if ok := assert.NotNil(t, pvc.Spec.StorageClassName, "PVC '%s' should have the storageClass", pvc.Name); ok {
		assert.Equal(t, expectedStorageClass, *(pvc.Spec.StorageClassName), "PVC '%s' should have the expected storage class", pvc.Name)
	}
}
