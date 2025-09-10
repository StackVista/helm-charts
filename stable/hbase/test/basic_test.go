package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

const releaseName = "olly"

func TestHBaseBasicTemplate(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Should have StatefulSets and Deployments
	assert.Greater(t, len(resources.Statefulsets), 0, "Should have StatefulSets")
	assert.Greater(t, len(resources.Deployments), 0, "Should have Deployments")
}

func TestHBaseDistributedMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/distributed-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// In Distributed mode, should have multiple StatefulSets (hdfs-snn is disabled by default)
	expectedStatefulSets := []string{
		releaseName + "-hbase-hbase-master",
		releaseName + "-hbase-hbase-rs",
		releaseName + "-hbase-hdfs-nn",
		releaseName + "-hbase-hdfs-dn",
		releaseName + "-hbase-tephra",
	}

	assert.Len(t, resources.Statefulsets, len(expectedStatefulSets), "Should have exactly %d StatefulSets in Distributed mode", len(expectedStatefulSets))

	// Check each expected StatefulSet exists
	for _, expectedName := range expectedStatefulSets {
		found := false
		for _, ss := range resources.Statefulsets {
			if ss.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "StatefulSet %s should exist in Distributed mode", expectedName)
	}

	// Should have console Deployment
	expectedDeployments := []string{
		releaseName + "-hbase-console",
	}

	assert.Len(t, resources.Deployments, len(expectedDeployments), "Should have exactly %d Deployments in Distributed mode", len(expectedDeployments))

	for _, expectedName := range expectedDeployments {
		found := false
		for _, deployment := range resources.Deployments {
			if deployment.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "Deployment %s should exist in Distributed mode", expectedName)
	}
}

func TestHBaseMonoMode(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, "values/mono-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// In Mono mode, should have fewer StatefulSets (only stackgraph and tephra-mono)
	expectedStatefulSets := []string{
		releaseName + "-hbase-stackgraph",
		releaseName + "-hbase-tephra-mono",
	}

	assert.Len(t, resources.Statefulsets, len(expectedStatefulSets), "Should have exactly %d StatefulSets in Mono mode", len(expectedStatefulSets))

	// Check each expected StatefulSet exists
	for _, expectedName := range expectedStatefulSets {
		found := false
		for _, ss := range resources.Statefulsets {
			if ss.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "StatefulSet %s should exist in Mono mode", expectedName)
	}

	// Should have console Deployment
	expectedDeployments := []string{
		releaseName + "-hbase-console",
	}

	assert.Len(t, resources.Deployments, len(expectedDeployments), "Should have exactly %d Deployments in Mono mode", len(expectedDeployments))

	for _, expectedName := range expectedDeployments {
		found := false
		for _, deployment := range resources.Deployments {
			if deployment.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "Deployment %s should exist in Mono mode", expectedName)
	}

	// Verify that Distributed-mode-only StatefulSets are NOT present
	distributedOnlyStatefulSets := []string{
		releaseName + "-hbase-hbase-master",
		releaseName + "-hbase-hbase-rs",
		releaseName + "-hbase-hdfs-nn",
		releaseName + "-hbase-hdfs-snn",
		releaseName + "-hbase-hdfs-dn",
	}

	for _, name := range distributedOnlyStatefulSets {
		found := false
		for _, ss := range resources.Statefulsets {
			if ss.Name == name {
				found = true
				break
			}
		}
		assert.False(t, found, "StatefulSet %s should NOT exist in Mono mode", name)
	}
}
