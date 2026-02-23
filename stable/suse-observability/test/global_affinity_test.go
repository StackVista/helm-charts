package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	v1 "k8s.io/api/core/v1"
)

// TestGlobalHAAffinityNodeAffinity tests that global.suseObservability.affinity.nodeAffinity is applied to all components in HA mode
func TestGlobalHAAffinityNodeAffinity(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_affinity_node_affinity.yaml", "values/workload_global_ha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	validateGlobalNodeAffinity(t, resources)
}

// TestGlobalNonHAAffinityNodeAffinity tests that global.suseObservability.affinity.nodeAffinity is applied to all components in non-HA mode
func TestGlobalNonHAAffinityNodeAffinity(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_affinity_node_affinity.yaml", "values/workload_global_nonha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	validateGlobalNodeAffinity(t, resources)
}

// TestGlobalNonHAAffinityNodeAffinity tests that global.suseObservability.affinity.nodeAffinity is applied to all components in non-HA mode
func TestGlobalAffinityNodeAffinityInBackupRestoreScriptsConfigMap(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_affinity_node_affinity.yaml", "values/workload_global_ha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Check all Jobs
	for backupJobName, job := range testJobsFromBackupRestoreScriptsConfigMap(t, &resources) {
		t.Run("BackupRestoreScriptsConfigMap/"+backupJobName, func(t *testing.T) {
			// Check that nodeAffinity is present
			require.NotNil(t, job.Spec.Template.Spec.Affinity, "Affinity should be set for the job from the backup-restore-scripts configmap %s", backupJobName)
			require.NotNil(t, job.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for the job from the backup-restore-scripts configmap %s", backupJobName)
			require.NotNil(t, job.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "RequiredDuringSchedulingIgnoredDuringExecution should be set for the job from the backup-restore-scripts configmap %s", backupJobName)

			// Verify the nodeAffinity matches what we set in global.suseObservability.affinity.nodeAffinity
			nodeSelectorTerms := job.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			require.Len(t, nodeSelectorTerms, 1, "Should have one node selector term for the job from the backup-restore-scripts configmap %s", backupJobName)
			require.Len(t, nodeSelectorTerms[0].MatchExpressions, 1, "Should have one match expression for the job from the backup-restore-scripts configmap %s", backupJobName)
			assert.Equal(t, "node-type", nodeSelectorTerms[0].MatchExpressions[0].Key, "Match expression key should be node-type for the job from the backup-restore-scripts configmap %s", backupJobName)
			assert.Equal(t, v1.NodeSelectorOpIn, nodeSelectorTerms[0].MatchExpressions[0].Operator, "Match expression operator should be In for the job from the backup-restore-scripts configmap %s", backupJobName)
			assert.Equal(t, []string{"observability"}, nodeSelectorTerms[0].MatchExpressions[0].Values, "Match expression values should be [observability] for the job from the backup-restore-scripts configmap %s", backupJobName)
		})
	}
}

// validateGlobalNodeAffinity validates that all Deployments, StatefulSets, CronJobs, and Jobs
// have the expected global nodeAffinity configuration applied
func validateGlobalNodeAffinity(t *testing.T, resources helmtestutil.KubernetesResources) {
	// Check all Deployments
	for deploymentName, deployment := range resources.Deployments {
		t.Run("Deployment/"+deploymentName, func(t *testing.T) {
			// Check that nodeAffinity is present
			require.NotNil(t, deployment.Spec.Template.Spec.Affinity, "Affinity should be set for deployment %s", deploymentName)
			require.NotNil(t, deployment.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for deployment %s", deploymentName)
			require.NotNil(t, deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "RequiredDuringSchedulingIgnoredDuringExecution should be set for deployment %s", deploymentName)

			// Verify the nodeAffinity matches what we set in global.suseObservability.affinity.nodeAffinity
			nodeSelectorTerms := deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			require.Len(t, nodeSelectorTerms, 1, "Should have one node selector term for deployment %s", deploymentName)
			require.Len(t, nodeSelectorTerms[0].MatchExpressions, 1, "Should have one match expression for deployment %s", deploymentName)
			assert.Equal(t, "node-type", nodeSelectorTerms[0].MatchExpressions[0].Key, "Match expression key should be node-type for deployment %s", deploymentName)
			assert.Equal(t, v1.NodeSelectorOpIn, nodeSelectorTerms[0].MatchExpressions[0].Operator, "Match expression operator should be In for deployment %s", deploymentName)
			assert.Equal(t, []string{"observability"}, nodeSelectorTerms[0].MatchExpressions[0].Values, "Match expression values should be [observability] for deployment %s", deploymentName)
		})
	}

	// Check all StatefulSets
	for statefulsetName, statefulset := range resources.Statefulsets {
		t.Run("Statefulset/"+statefulsetName, func(t *testing.T) {
			// Check that nodeAffinity is present
			require.NotNil(t, statefulset.Spec.Template.Spec.Affinity, "Affinity should be set for statefulset %s", statefulsetName)
			require.NotNil(t, statefulset.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for statefulset %s", statefulsetName)
			require.NotNil(t, statefulset.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "RequiredDuringSchedulingIgnoredDuringExecution should be set for statefulset %s", statefulsetName)

			// Verify the nodeAffinity matches what we set in global.suseObservability.affinity.nodeAffinity
			nodeSelectorTerms := statefulset.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			require.Len(t, nodeSelectorTerms, 1, "Should have one node selector term for statefulset %s", statefulsetName)
			require.Len(t, nodeSelectorTerms[0].MatchExpressions, 1, "Should have one match expression for statefulset %s", statefulsetName)
			assert.Equal(t, "node-type", nodeSelectorTerms[0].MatchExpressions[0].Key, "Match expression key should be node-type for statefulset %s", statefulsetName)
			assert.Equal(t, v1.NodeSelectorOpIn, nodeSelectorTerms[0].MatchExpressions[0].Operator, "Match expression operator should be In for statefulset %s", statefulsetName)
			assert.Equal(t, []string{"observability"}, nodeSelectorTerms[0].MatchExpressions[0].Values, "Match expression values should be [observability] for statefulset %s", statefulsetName)
		})
	}

	// Check all CronJobs
	for cronjobName, cronjob := range resources.CronJobs {
		t.Run("CronJob/"+cronjobName, func(t *testing.T) {
			// Check that nodeAffinity is present
			require.NotNil(t, cronjob.Spec.JobTemplate.Spec.Template.Spec.Affinity, "Affinity should be set for cronjob %s", cronjobName)
			require.NotNil(t, cronjob.Spec.JobTemplate.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for cronjob %s", cronjobName)
			require.NotNil(t, cronjob.Spec.JobTemplate.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "RequiredDuringSchedulingIgnoredDuringExecution should be set for cronjob %s", cronjobName)

			// Verify the nodeAffinity matches what we set in global.suseObservability.affinity.nodeAffinity
			nodeSelectorTerms := cronjob.Spec.JobTemplate.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			require.Len(t, nodeSelectorTerms, 1, "Should have one node selector term for cronjob %s", cronjobName)
			require.Len(t, nodeSelectorTerms[0].MatchExpressions, 1, "Should have one match expression for cronjob %s", cronjobName)
			assert.Equal(t, "node-type", nodeSelectorTerms[0].MatchExpressions[0].Key, "Match expression key should be node-type for cronjob %s", cronjobName)
			assert.Equal(t, v1.NodeSelectorOpIn, nodeSelectorTerms[0].MatchExpressions[0].Operator, "Match expression operator should be In for cronjob %s", cronjobName)
			assert.Equal(t, []string{"observability"}, nodeSelectorTerms[0].MatchExpressions[0].Values, "Match expression values should be [observability] for cronjob %s", cronjobName)
		})
	}

	// Check all Jobs
	for jobName, job := range resources.Jobs {
		t.Run("Job/"+jobName, func(t *testing.T) {
			// Check that nodeAffinity is present
			require.NotNil(t, job.Spec.Template.Spec.Affinity, "Affinity should be set for job %s", jobName)
			require.NotNil(t, job.Spec.Template.Spec.Affinity.NodeAffinity, "NodeAffinity should be set for job %s", jobName)
			require.NotNil(t, job.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "RequiredDuringSchedulingIgnoredDuringExecution should be set for job %s", jobName)

			// Verify the nodeAffinity matches what we set in global.suseObservability.affinity.nodeAffinity
			nodeSelectorTerms := job.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			require.Len(t, nodeSelectorTerms, 1, "Should have one node selector term for job %s", jobName)
			require.Len(t, nodeSelectorTerms[0].MatchExpressions, 1, "Should have one match expression for job %s", jobName)
			assert.Equal(t, "node-type", nodeSelectorTerms[0].MatchExpressions[0].Key, "Match expression key should be node-type for job %s", jobName)
			assert.Equal(t, v1.NodeSelectorOpIn, nodeSelectorTerms[0].MatchExpressions[0].Operator, "Match expression operator should be In for job %s", jobName)
			assert.Equal(t, []string{"observability"}, nodeSelectorTerms[0].MatchExpressions[0].Values, "Match expression values should be [observability] for job %s", jobName)
		})
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
