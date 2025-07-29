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
	customStorageClass = "TestStorageClass"
	customVolumeSize   = "9999Ti"
)

var components = []string{
	"suse-observability-api",
	"suse-observability-authorization-sync",
	"suse-observability-checks",
	"suse-observability-health-sync",
	"suse-observability-notification",
	"suse-observability-state",
	"suse-observability-sync",
}

func TestFeaturesStoreTransactionLogsToPVCDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.k8sAuthorization.enabled": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	for _, component := range components {
		deploymentDoesntHaveTransactionLogPVC(t, &resources, component)
		pvcName := component + "-txlog"
		assert.NotContains(t, resources.PersistentVolumeClaims, pvcName, "PVC '%s' should not exist when storeTransactionLogsToPVC is disabled", pvcName)
	}
}

func TestFeaturesStoreTransactionLogsToPVCEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.storeTransactionLogsToPVC.enabled": "true",
			"stackstate.k8sAuthorization.enabled":                   "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	for _, component := range components {
		deploymentHasTransactionLogPVC(t, &resources, component)
		pvcName := component + "-txlog"
		assert.Contains(t, resources.PersistentVolumeClaims, pvcName, "PVC '%s' should exist when storeTransactionLogsToPVC is enabled", pvcName)
	}
}

func TestFeaturesStoreTransactionLogsToPVCExperimentalOverFeatures(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.k8sAuthorization.enabled":                       "true",
			"stackstate.features.storeTransactionLogsToPVC.enabled":     "true",
			"stackstate.experimental.storeTransactionLogsToPVC.enabled": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	for _, component := range components {
		deploymentDoesntHaveTransactionLogPVC(t, &resources, component)
		pvcName := component + "-txlog"
		assert.NotContains(t, resources.PersistentVolumeClaims, pvcName, "PVC '%s' should not exist when experimental override disables storeTransactionLogsToPVC", pvcName)
	}
}

func TestFeaturesStoreTransactionLogsToPVCCustomSettings(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"stackstate.features.storeTransactionLogsToPVC.enabled":      "true",
			"stackstate.features.storeTransactionLogsToPVC.volumeSize":   customVolumeSize,
			"stackstate.features.storeTransactionLogsToPVC.storageClass": customStorageClass,
			"stackstate.k8sAuthorization.enabled":                        "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	for _, component := range components {
		pvcName := component + "-txlog"
		if ok := assert.Contains(t, resources.PersistentVolumeClaims, pvcName); ok {
			pvc := resources.PersistentVolumeClaims[pvcName]
			storageRequest, ok := pvc.Spec.Resources.Requests[v1.ResourceStorage]
			require.True(t, ok, "PVC '%s' should have storage resource request defined", pvcName)
			assert.Equal(t, customVolumeSize, storageRequest.String(), "PVC '%s' should have the expected storage size", pvcName)
			if ok := assert.NotNil(t, pvc.Spec.StorageClassName, "PVC '%s' should have the storageClass", pvcName); ok {
				assert.Equal(t, customStorageClass, *(pvc.Spec.StorageClassName), "PVC '%s' should have the expected storage class", pvcName)
			}
		}
	}
}

// deploymentHasTransactionLogPVC validates that a deployment has the expected volume and mount configuration
func deploymentHasTransactionLogPVC(t *testing.T, resources *helmtestutil.KubernetesResources, deploymentName string) {
	deployment, ok := resources.Deployments[deploymentName]
	require.True(t, ok, "Deployment '%s' should exist", deploymentName)

	expectedClaimName := fmt.Sprintf("%s-txlog", deploymentName)
	expectedVolume := v1.Volume{
		Name: "application-log",
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: expectedClaimName,
			},
		},
	}
	assert.Contains(t, deployment.Spec.Template.Spec.Volumes, expectedVolume, "Deployment '%s' should have the transaction log PVC volume", deploymentName)

	expectedVolumeMount := v1.VolumeMount{
		Name:      "application-log",
		MountPath: "/opt/docker/logs",
	}
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts, expectedVolumeMount, "Deployment '%s' container should have the transaction log volume mount", deploymentName)

	var foundInitContainer bool
	for _, initContainer := range deployment.Spec.Template.Spec.InitContainers {
		if initContainer.Name == "clean-transaction-logs-directory" {
			foundInitContainer = true
			assert.Contains(t, initContainer.VolumeMounts, expectedVolumeMount, "Init container 'clean-transaction-logs-directory' should have the transaction log volume mount")
			break
		}
	}
	assert.True(t, foundInitContainer, "Deployment '%s' should have 'clean-transaction-logs-directory' init container when using transaction log PVC", deploymentName)
}

// deploymentDoesntHaveTransactionLogPVC validates that a deployment does not have the expected volume and mount configuration
func deploymentDoesntHaveTransactionLogPVC(t *testing.T, resources *helmtestutil.KubernetesResources, deploymentName string) {
	deployment, ok := resources.Deployments[deploymentName]
	require.True(t, ok, "Deployment '%s' should exist", deploymentName)

	expectedClaimName := fmt.Sprintf("%s-txlog", deploymentName)
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == expectedClaimName {
			t.Errorf("Deployment '%s' should not have a volume with claim name '%s'", deploymentName, expectedClaimName)
		}
	}

	mountPathName := "application-log"

	for _, volumeMount := range deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
		if volumeMount.Name == mountPathName {
			t.Errorf("Deployment '%s' should not have a mountPath '%s'", deploymentName, mountPathName)
		}
	}

	for _, initContainer := range deployment.Spec.Template.Spec.InitContainers {
		if initContainer.Name == "clean-transaction-logs-directory" {
			t.Errorf("Deployment '%s' should not have a init container '%s'", deploymentName, initContainer.Name)
			break
		}
	}
}
