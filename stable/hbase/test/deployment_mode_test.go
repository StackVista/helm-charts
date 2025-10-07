package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// TestDistributedModeResourcesPresent verifies that resources with Distributed mode conditionals are rendered in Distributed mode
func TestDistributedModeResourcesPresent(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/distributed-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// PodDisruptionBudgets that should exist in Distributed mode
	expectedPDBs := []string{
		releaseName + "-hbase-hbase-master",
		releaseName + "-hbase-hbase-rs",
		releaseName + "-hbase-hdfs-nn",
		releaseName + "-hbase-hdfs-dn",
	}

	for _, expectedName := range expectedPDBs {
		found := false
		for _, pdb := range resources.Pdbs {
			if pdb.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "PodDisruptionBudget %s should exist in Distributed mode", expectedName)
	}

	// ConfigMaps for scripts that should exist in Distributed mode
	expectedConfigMaps := []string{
		releaseName + "-hbase-hbase-rs-scripts",
		releaseName + "-hbase-hdfs-dn-scripts",
	}

	for _, expectedName := range expectedConfigMaps {
		found := false
		for _, cm := range resources.ConfigMaps {
			if cm.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "ConfigMap %s should exist in Distributed mode", expectedName)
	}
}

// TestDistributedModeSecretsPresent verifies that Secrets with Distributed mode conditionals are rendered when extraEnv.secret is set
func TestDistributedModeSecretsPresent(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/distributed-mode-with-secrets.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Secrets that should exist in Distributed mode when extraEnv.secret is set
	expectedSecrets := []string{
		releaseName + "-hbase-hbase-master",
		releaseName + "-hbase-hbase-rs",
		releaseName + "-hbase-hdfs-nn",
		releaseName + "-hbase-hdfs-dn",
	}

	for _, expectedName := range expectedSecrets {
		found := false
		for _, secret := range resources.Secrets {
			if secret.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "Secret %s should exist in Distributed mode when extraEnv.secret is set", expectedName)
	}
}

// TestMonoModeResourcesAbsent verifies that resources with Distributed mode conditionals are NOT rendered in Mono mode
func TestMonoModeResourcesAbsent(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/mono-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// PodDisruptionBudgets that should NOT exist in Mono mode
	distributedOnlyPDBs := []string{
		releaseName + "-hbase-hbase-master",
		releaseName + "-hbase-hbase-rs",
		releaseName + "-hbase-hdfs-nn",
		releaseName + "-hbase-hdfs-dn",
		releaseName + "-hbase-hdfs-snn",
	}

	for _, name := range distributedOnlyPDBs {
		found := false
		for _, pdb := range resources.Pdbs {
			if pdb.Name == name {
				found = true
				break
			}
		}
		assert.False(t, found, "PodDisruptionBudget %s should NOT exist in Mono mode", name)
	}

	// ConfigMaps that should NOT exist in Mono mode
	distributedOnlyConfigMaps := []string{
		releaseName + "-hbase-hbase-rs-scripts",
		releaseName + "-hbase-hdfs-dn-scripts",
	}

	for _, name := range distributedOnlyConfigMaps {
		found := false
		for _, cm := range resources.ConfigMaps {
			if cm.Name == name {
				found = true
				break
			}
		}
		assert.False(t, found, "ConfigMap %s should NOT exist in Mono mode", name)
	}
}

// TestMonoModeSecretsAbsent verifies that Secrets with Distributed mode conditionals are NOT rendered in Mono mode
func TestMonoModeSecretsAbsent(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/mono-mode-with-secrets.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Secrets that should NOT exist in Mono mode
	distributedOnlySecrets := []string{
		releaseName + "-hbase-hbase-master",
		releaseName + "-hbase-hbase-rs",
		releaseName + "-hbase-hdfs-nn",
		releaseName + "-hbase-hdfs-dn",
		releaseName + "-hbase-hdfs-snn",
	}

	for _, name := range distributedOnlySecrets {
		found := false
		for _, secret := range resources.Secrets {
			if secret.Name == name {
				found = true
				break
			}
		}
		assert.False(t, found, "Secret %s should NOT exist in Mono mode", name)
	}
}

// TestDistributedModeSecondaryNameNodeResources verifies that hdfs-snn resources are rendered when enabled in Distributed mode
func TestDistributedModeSecondaryNameNodeResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/distributed-mode-with-snn.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	expectedName := releaseName + "-hbase-hdfs-snn"

	// Check PodDisruptionBudget
	found := false
	for _, pdb := range resources.Pdbs {
		if pdb.Name == expectedName {
			found = true
			break
		}
	}
	assert.True(t, found, "PodDisruptionBudget %s should exist in Distributed mode when hdfs.secondarynamenode.enabled=true", expectedName)

	// Check Secret
	found = false
	for _, secret := range resources.Secrets {
		if secret.Name == expectedName {
			found = true
			break
		}
	}
	assert.True(t, found, "Secret %s should exist in Distributed mode when hdfs.secondarynamenode.enabled=true and extraEnv.secret is set", expectedName)
}
