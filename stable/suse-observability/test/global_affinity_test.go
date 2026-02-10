package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	v1 "k8s.io/api/core/v1"
)

// TestGlobalAffinityNodeAffinity tests that global.suseObservability.affinity.nodeAffinity is applied to all components
func TestGlobalAffinityNodeAffinity(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_affinity_node_affinity.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// List of components that should have affinity applied
	componentNames := []string{
		"suse-observability-api",
		"suse-observability-checks",
		"suse-observability-correlate-aggregator",
		"suse-observability-correlate-connection",
		"suse-observability-correlate-http-tracing",
		"suse-observability-health-sync",
		"suse-observability-initializer",
		"suse-observability-k2es-events-to-es",
		"suse-observability-notification",
		"suse-observability-receiver-base",
		"suse-observability-receiver-logs",
		"suse-observability-receiver-process-agent",
		"suse-observability-router",
		"suse-observability-slicing",
		"suse-observability-state",
		"suse-observability-sync",
		"suse-observability-ui",
		"suse-observability-authorization-sync",
	}

	// Verify each component has the global nodeAffinity applied
	for _, componentName := range componentNames {
		deployment, exists := resources.Deployments[componentName]
		if !exists {
			t.Logf("Skipping %s (not found in rendered output)", componentName)
			continue
		}

		// Check that nodeAffinity is present
		require.NotNil(t, deployment.Spec.Template.Spec.Affinity, "Affinity should be set for %s", componentName)
		require.NotNil(t, deployment.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for %s", componentName)
		require.NotNil(t, deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "RequiredDuringSchedulingIgnoredDuringExecution should be set for %s", componentName)

		// Verify the nodeAffinity matches what we set in global.suseObservability.affinity.nodeAffinity
		nodeSelectorTerms := deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		require.Len(t, nodeSelectorTerms, 1, "Should have one node selector term for %s", componentName)
		require.Len(t, nodeSelectorTerms[0].MatchExpressions, 1, "Should have one match expression for %s", componentName)
		assert.Equal(t, "node-type", nodeSelectorTerms[0].MatchExpressions[0].Key, "Match expression key should be node-type for %s", componentName)
		assert.Equal(t, v1.NodeSelectorOpIn, nodeSelectorTerms[0].MatchExpressions[0].Operator, "Match expression operator should be In for %s", componentName)
		assert.Equal(t, []string{"observability"}, nodeSelectorTerms[0].MatchExpressions[0].Values, "Match expression values should be [observability] for %s", componentName)
	}

	// Also check statefulsets
	statefulsetNames := []string{
		"suse-observability-vmagent",
		"suse-observability-workload-observer",
	}

	for _, statefulsetName := range statefulsetNames {
		statefulset, exists := resources.Statefulsets[statefulsetName]
		if !exists {
			t.Logf("Skipping %s (not found in rendered output)", statefulsetName)
			continue
		}

		// Check that nodeAffinity is present
		require.NotNil(t, statefulset.Spec.Template.Spec.Affinity, "Affinity should be set for %s", statefulsetName)
		require.NotNil(t, statefulset.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for %s", statefulsetName)
		require.NotNil(t, statefulset.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "RequiredDuringSchedulingIgnoredDuringExecution should be set for %s", statefulsetName)

		// Verify the nodeAffinity matches what we set in global.suseObservability.affinity.nodeAffinity
		nodeSelectorTerms := statefulset.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		require.Len(t, nodeSelectorTerms, 1, "Should have one node selector term for %s", statefulsetName)
		require.Len(t, nodeSelectorTerms[0].MatchExpressions, 1, "Should have one match expression for %s", statefulsetName)
		assert.Equal(t, "node-type", nodeSelectorTerms[0].MatchExpressions[0].Key, "Match expression key should be node-type for %s", statefulsetName)
	}
}

// TestGlobalAffinityComponentOverride tests that component-specific affinity overrides global affinity
func TestGlobalAffinityComponentOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_affinity_with_override.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// API component should have the override affinity (api-specific)
	apiDeployment, exists := resources.Deployments["suse-observability-api"]
	require.True(t, exists, "API deployment should exist")
	require.NotNil(t, apiDeployment.Spec.Template.Spec.Affinity, "Affinity should be set for API")
	require.NotNil(t, apiDeployment.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for API")

	nodeSelectorTerms := apiDeployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	require.Len(t, nodeSelectorTerms, 1, "Should have one node selector term for API")
	require.Len(t, nodeSelectorTerms[0].MatchExpressions, 1, "Should have one match expression for API")
	assert.Equal(t, "node-type", nodeSelectorTerms[0].MatchExpressions[0].Key, "Match expression key should be node-type")
	assert.Equal(t, []string{"api-specific"}, nodeSelectorTerms[0].MatchExpressions[0].Values, "API should use override value 'api-specific'")

	// Sync component should have the global affinity (observability)
	syncDeployment, exists := resources.Deployments["suse-observability-sync"]
	require.True(t, exists, "Sync deployment should exist")
	require.NotNil(t, syncDeployment.Spec.Template.Spec.Affinity, "Affinity should be set for Sync")
	require.NotNil(t, syncDeployment.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for Sync")

	syncNodeSelectorTerms := syncDeployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	require.Len(t, syncNodeSelectorTerms, 1, "Should have one node selector term for Sync")
	require.Len(t, syncNodeSelectorTerms[0].MatchExpressions, 1, "Should have one match expression for Sync")
	assert.Equal(t, "node-type", syncNodeSelectorTerms[0].MatchExpressions[0].Key, "Match expression key should be node-type")
	assert.Equal(t, []string{"observability"}, syncNodeSelectorTerms[0].MatchExpressions[0].Values, "Sync should use global value 'observability'")
}

// TestGlobalAffinityAppComponentsNoPodAntiAffinity tests that application components don't get podAntiAffinity
// from the simplified format (only infrastructure components get it)
func TestGlobalAffinityAppComponentsNoPodAntiAffinity(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_affinity_infrastructure.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Check that API deployment doesn't have podAntiAffinity from the simplified format
	// (simplified format only applies to infrastructure components)
	apiDeployment, exists := resources.Deployments["suse-observability-api"]
	require.True(t, exists, "API deployment should exist")
	require.NotNil(t, apiDeployment.Spec.Template.Spec.Affinity, "Affinity should be set for API")
	require.NotNil(t, apiDeployment.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for API")

	// PodAntiAffinity should be nil for application components (simplified format only for infra)
	assert.Nil(t, apiDeployment.Spec.Template.Spec.Affinity.PodAntiAffinity, "PodAntiAffinity should not be set for application components from simplified format")
}

// TestLegacyAffinityBackwardsCompatibility tests that legacy stackstate.components.*.affinity still works
func TestLegacyAffinityBackwardsCompatibility(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Verify that deployments without global.suseObservability.affinity can still render
	// This ensures we didn't break backward compatibility
	_, syncExists := resources.Deployments["suse-observability-sync"]
	require.True(t, syncExists, "Sync deployment should exist in legacy mode")

	_, apiExists := resources.Deployments["suse-observability-api"]
	require.True(t, apiExists, "API deployment should exist in legacy mode")
}

// TestGlobalAffinityInfrastructureNodeAffinity tests that nodeAffinity is applied to infrastructure components
func TestGlobalAffinityInfrastructureNodeAffinity(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_affinity_infrastructure.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Infrastructure statefulsets that should have nodeAffinity from global config
	infraStatefulsets := []string{
		"suse-observability-kafka",
		"suse-observability-zookeeper",
		"suse-observability-clickhouse-shard0",
		"suse-observability-victoria-metrics-0",
		"suse-observability-victoria-metrics-1",
		"suse-observability-elasticsearch-master",
	}

	for _, stsName := range infraStatefulsets {
		sts, exists := resources.Statefulsets[stsName]
		if !exists {
			t.Logf("Skipping %s (not found in rendered output)", stsName)
			continue
		}

		// Check that nodeAffinity is present
		require.NotNil(t, sts.Spec.Template.Spec.Affinity, "Affinity should be set for %s", stsName)
		require.NotNil(t, sts.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for %s", stsName)
		require.NotNil(t, sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "RequiredDuringSchedulingIgnoredDuringExecution should be set for %s", stsName)

		// Verify the nodeAffinity matches what we set in global.suseObservability.affinity.nodeAffinity
		nodeSelectorTerms := sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		require.Len(t, nodeSelectorTerms, 1, "Should have one node selector term for %s", stsName)
		require.Len(t, nodeSelectorTerms[0].MatchExpressions, 1, "Should have one match expression for %s", stsName)
		assert.Equal(t, "node-type", nodeSelectorTerms[0].MatchExpressions[0].Key, "Match expression key should be node-type for %s", stsName)
		assert.Equal(t, v1.NodeSelectorOpIn, nodeSelectorTerms[0].MatchExpressions[0].Operator, "Match expression operator should be In for %s", stsName)
		assert.Equal(t, []string{"observability"}, nodeSelectorTerms[0].MatchExpressions[0].Values, "Match expression values should be [observability] for %s", stsName)
	}
}

// TestGlobalAffinityInfrastructureCustomTopologyKey tests that custom topologyKey is applied to infrastructure podAntiAffinity
func TestGlobalAffinityInfrastructureCustomTopologyKey(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_affinity_infrastructure.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Infrastructure statefulsets that should have podAntiAffinity with custom topologyKey
	infraStatefulsets := []string{
		"suse-observability-kafka",
		"suse-observability-zookeeper",
		"suse-observability-clickhouse-shard0",
		"suse-observability-victoria-metrics-0",
		"suse-observability-victoria-metrics-1",
	}

	for _, stsName := range infraStatefulsets {
		sts, exists := resources.Statefulsets[stsName]
		if !exists {
			t.Logf("Skipping %s (not found in rendered output)", stsName)
			continue
		}

		// Check that podAntiAffinity is present with custom topologyKey
		require.NotNil(t, sts.Spec.Template.Spec.Affinity, "Affinity should be set for %s", stsName)
		require.NotNil(t, sts.Spec.Template.Spec.Affinity.PodAntiAffinity, "PodAntiAffinity should be set for %s", stsName)
		require.NotNil(t, sts.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "RequiredDuringSchedulingIgnoredDuringExecution should be set for %s", stsName)
		require.Len(t, sts.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1, "Should have one anti-affinity term for %s", stsName)

		// Verify custom topologyKey (topology.kubernetes.io/zone)
		topologyKey := sts.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey
		assert.Equal(t, "topology.kubernetes.io/zone", topologyKey, "TopologyKey should be topology.kubernetes.io/zone for %s", stsName)
	}
}

// TestGlobalAffinitySoftAntiAffinity tests that soft (preferred) anti-affinity is applied when configured
func TestGlobalAffinitySoftAntiAffinity(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_affinity_soft_antiaffinity.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Infrastructure statefulsets that should have soft podAntiAffinity
	infraStatefulsets := []string{
		"suse-observability-kafka",
		"suse-observability-zookeeper",
		"suse-observability-clickhouse-shard0",
		"suse-observability-victoria-metrics-0",
		"suse-observability-victoria-metrics-1",
	}

	for _, stsName := range infraStatefulsets {
		sts, exists := resources.Statefulsets[stsName]
		if !exists {
			t.Logf("Skipping %s (not found in rendered output)", stsName)
			continue
		}

		// Check that podAntiAffinity uses preferredDuringSchedulingIgnoredDuringExecution (soft)
		require.NotNil(t, sts.Spec.Template.Spec.Affinity, "Affinity should be set for %s", stsName)
		require.NotNil(t, sts.Spec.Template.Spec.Affinity.PodAntiAffinity, "PodAntiAffinity should be set for %s", stsName)

		// Should have preferred (soft) anti-affinity, not required (hard)
		assert.Nil(t, sts.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "RequiredDuringSchedulingIgnoredDuringExecution should NOT be set for soft anti-affinity on %s", stsName)
		require.NotNil(t, sts.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, "PreferredDuringSchedulingIgnoredDuringExecution should be set for soft anti-affinity on %s", stsName)
		require.Len(t, sts.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1, "Should have one preferred anti-affinity term for %s", stsName)

		// Verify topologyKey
		topologyKey := sts.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey
		assert.Equal(t, "kubernetes.io/hostname", topologyKey, "TopologyKey should be kubernetes.io/hostname for %s", stsName)
	}
}
