package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// TestGlobalSizingProfilesRender tests that all 9 sizing profiles render successfully
func TestGlobalSizingProfilesRender(t *testing.T) {
	profiles := []struct {
		name      string
		valuesFile string
	}{
		{"trial", "values/global_sizing_trial.yaml"},
		{"10-nonha", "values/global_sizing_10_nonha.yaml"},
		{"20-nonha", "values/global_sizing_20_nonha.yaml"},
		{"50-nonha", "values/global_sizing_50_nonha.yaml"},
		{"100-nonha", "values/global_sizing_100_nonha.yaml"},
		{"150-ha", "values/global_sizing_150_ha.yaml"},
		{"250-ha", "values/global_sizing_250_ha.yaml"},
		{"500-ha", "values/global_sizing_500_ha.yaml"},
		{"4000-ha", "values/global_sizing_4000_ha.yaml"},
	}

	for _, profile := range profiles {
		t.Run(profile.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", profile.valuesFile)
			resources := helmtestutil.NewKubernetesResources(t, output)

			// Verify basic resources exist
			assert.NotEmpty(t, resources.Deployments, "Profile %s should have Deployments", profile.name)
			assert.NotEmpty(t, resources.Statefulsets, "Profile %s should have StatefulSets", profile.name)
			assert.NotEmpty(t, resources.Services, "Profile %s should have Services", profile.name)

			// Verify critical deployments exist
			// For non-HA profiles (trial, *-nonha), server is not split, so we have "server" instead of "api" and "checks"
			// For HA profiles, server is split into "api", "checks", etc.
			hasServer := false
			hasApi := false
			if _, exists := resources.Deployments["suse-observability-server"]; exists {
				hasServer = true
			}
			if _, exists := resources.Deployments["suse-observability-api"]; exists {
				hasApi = true
			}
			assert.True(t, hasServer || hasApi, "Profile %s should have either server or API deployment", profile.name)

			// Verify critical statefulsets exist
			assert.Contains(t, resources.Statefulsets, "suse-observability-kafka", "Profile %s should have Kafka", profile.name)
			assert.Contains(t, resources.Statefulsets, "suse-observability-zookeeper", "Profile %s should have Zookeeper", profile.name)
		})
	}
}

// TestGlobalSizingReceiverSplitMode tests that receiver split mode is controlled by sizing profile
func TestGlobalSizingReceiverSplitMode(t *testing.T) {
	testCases := []struct {
		name                  string
		valuesFile            string
		expectSplit           bool
		expectedDeployments   []string
	}{
		{
			name:        "trial-no-split",
			valuesFile:  "values/global_sizing_trial.yaml",
			expectSplit: false,
			expectedDeployments: []string{"suse-observability-receiver"},
		},
		{
			name:        "10-nonha-no-split",
			valuesFile:  "values/global_sizing_10_nonha.yaml",
			expectSplit: false,
			expectedDeployments: []string{"suse-observability-receiver"},
		},
		{
			name:        "20-nonha-no-split",
			valuesFile:  "values/global_sizing_20_nonha.yaml",
			expectSplit: false,
			expectedDeployments: []string{"suse-observability-receiver"},
		},
		{
			name:        "50-nonha-no-split",
			valuesFile:  "values/global_sizing_50_nonha.yaml",
			expectSplit: false,
			expectedDeployments: []string{"suse-observability-receiver"},
		},
		{
			name:        "100-nonha-no-split",
			valuesFile:  "values/global_sizing_100_nonha.yaml",
			expectSplit: false,
			expectedDeployments: []string{"suse-observability-receiver"},
		},
		{
			name:        "150-ha-split",
			valuesFile:  "values/global_sizing_150_ha.yaml",
			expectSplit: true,
			expectedDeployments: []string{
				"suse-observability-receiver-base",
				"suse-observability-receiver-logs",
				"suse-observability-receiver-process-agent",
			},
		},
		{
			name:        "250-ha-split",
			valuesFile:  "values/global_sizing_250_ha.yaml",
			expectSplit: true,
			expectedDeployments: []string{
				"suse-observability-receiver-base",
				"suse-observability-receiver-logs",
				"suse-observability-receiver-process-agent",
			},
		},
		{
			name:        "500-ha-split",
			valuesFile:  "values/global_sizing_500_ha.yaml",
			expectSplit: true,
			expectedDeployments: []string{
				"suse-observability-receiver-base",
				"suse-observability-receiver-logs",
				"suse-observability-receiver-process-agent",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", tc.valuesFile)
			resources := helmtestutil.NewKubernetesResources(t, output)

			for _, deploymentName := range tc.expectedDeployments {
				deployment, exists := resources.Deployments[deploymentName]
				assert.True(t, exists, "Deployment %s should exist for %s", deploymentName, tc.name)

				if exists {
					// Verify resources are set
					containers := deployment.Spec.Template.Spec.Containers
					require.NotEmpty(t, containers, "Deployment %s should have containers", deploymentName)

					resources := containers[0].Resources
					assert.NotNil(t, resources.Limits.Memory(), "Deployment %s should have memory limits", deploymentName)
					assert.NotNil(t, resources.Requests.Memory(), "Deployment %s should have memory requests", deploymentName)
					assert.NotNil(t, resources.Limits.Cpu(), "Deployment %s should have CPU limits", deploymentName)
					assert.NotNil(t, resources.Requests.Cpu(), "Deployment %s should have CPU requests", deploymentName)
				}
			}

			// Verify the opposite deployments don't exist
			if tc.expectSplit {
				assert.NotContains(t, resources.Deployments, "suse-observability-receiver",
					"Non-split receiver should not exist for %s", tc.name)
			} else {
				assert.NotContains(t, resources.Deployments, "suse-observability-receiver-base",
					"Split receiver-base should not exist for %s", tc.name)
				assert.NotContains(t, resources.Deployments, "suse-observability-receiver-logs",
					"Split receiver-logs should not exist for %s", tc.name)
			}
		})
	}
}

// TestGlobalSizingVictoriaMetrics1Enablement tests victoria-metrics-1 conditional enablement
func TestGlobalSizingVictoriaMetrics1Enablement(t *testing.T) {
	testCases := []struct {
		name                    string
		valuesFile              string
		expectVM1Enabled        bool
	}{
		{
			name:             "trial-vm1-disabled",
			valuesFile:       "values/global_sizing_trial.yaml",
			expectVM1Enabled: false,
		},
		{
			name:             "10-nonha-vm1-disabled",
			valuesFile:       "values/global_sizing_10_nonha.yaml",
			expectVM1Enabled: false,
		},
		{
			name:             "20-nonha-vm1-disabled",
			valuesFile:       "values/global_sizing_20_nonha.yaml",
			expectVM1Enabled: false,
		},
		{
			name:             "50-nonha-vm1-disabled",
			valuesFile:       "values/global_sizing_50_nonha.yaml",
			expectVM1Enabled: false,
		},
		{
			name:             "100-nonha-vm1-disabled",
			valuesFile:       "values/global_sizing_100_nonha.yaml",
			expectVM1Enabled: false,
		},
		{
			name:             "150-ha-vm1-enabled",
			valuesFile:       "values/global_sizing_150_ha.yaml",
			expectVM1Enabled: true,
		},
		{
			name:             "250-ha-vm1-enabled",
			valuesFile:       "values/global_sizing_250_ha.yaml",
			expectVM1Enabled: true,
		},
		{
			name:             "500-ha-vm1-enabled",
			valuesFile:       "values/global_sizing_500_ha.yaml",
			expectVM1Enabled: true,
		},
		{
			name:             "4000-ha-vm1-enabled",
			valuesFile:       "values/global_sizing_4000_ha.yaml",
			expectVM1Enabled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", tc.valuesFile)
			resources := helmtestutil.NewKubernetesResources(t, output)

			vm1StatefulSet := "suse-observability-victoria-metrics-1"
			vm1Service := "suse-observability-victoria-metrics-1"

			if tc.expectVM1Enabled {
				assert.Contains(t, resources.Statefulsets, vm1StatefulSet,
					"victoria-metrics-1 StatefulSet should exist for %s", tc.name)
				assert.Contains(t, resources.Services, vm1Service,
					"victoria-metrics-1 Service should exist for %s", tc.name)
			} else {
				assert.NotContains(t, resources.Statefulsets, vm1StatefulSet,
					"victoria-metrics-1 StatefulSet should not exist for %s", tc.name)
				assert.NotContains(t, resources.Services, vm1Service,
					"victoria-metrics-1 Service should not exist for %s", tc.name)
			}

			// victoria-metrics-0 should always exist
			assert.Contains(t, resources.Statefulsets, "suse-observability-victoria-metrics-0",
				"victoria-metrics-0 should always exist for %s", tc.name)
		})
	}
}

// TestGlobalSizingHbaseDeploymentMode tests hbase deployment mode (Mono vs Distributed)
func TestGlobalSizingHbaseDeploymentMode(t *testing.T) {
	testCases := []struct {
		name                    string
		valuesFile              string
		expectMonoMode          bool
		expectedStatefulSets    []string
		unexpectedStatefulSets  []string
	}{
		{
			name:           "trial-mono-mode",
			valuesFile:     "values/global_sizing_trial.yaml",
			expectMonoMode: true,
			expectedStatefulSets: []string{
				"suse-observability-hbase-stackgraph",
				"suse-observability-hbase-tephra-mono",
			},
			unexpectedStatefulSets: []string{
				"suse-observability-hbase-hbase-master",
				"suse-observability-hbase-hbase-rs",
				"suse-observability-hbase-hdfs-nn",
				"suse-observability-hbase-hdfs-dn",
			},
		},
		{
			name:           "10-nonha-mono-mode",
			valuesFile:     "values/global_sizing_10_nonha.yaml",
			expectMonoMode: true,
			expectedStatefulSets: []string{
				"suse-observability-hbase-stackgraph",
				"suse-observability-hbase-tephra-mono",
			},
			unexpectedStatefulSets: []string{
				"suse-observability-hbase-hbase-master",
				"suse-observability-hbase-hbase-rs",
				"suse-observability-hbase-hdfs-nn",
				"suse-observability-hbase-hdfs-dn",
			},
		},
		{
			name:           "20-nonha-mono-mode",
			valuesFile:     "values/global_sizing_20_nonha.yaml",
			expectMonoMode: true,
			expectedStatefulSets: []string{
				"suse-observability-hbase-stackgraph",
				"suse-observability-hbase-tephra-mono",
			},
			unexpectedStatefulSets: []string{
				"suse-observability-hbase-hbase-master",
				"suse-observability-hbase-hbase-rs",
				"suse-observability-hbase-hdfs-nn",
				"suse-observability-hbase-hdfs-dn",
			},
		},
		{
			name:           "50-nonha-mono-mode",
			valuesFile:     "values/global_sizing_50_nonha.yaml",
			expectMonoMode: true,
			expectedStatefulSets: []string{
				"suse-observability-hbase-stackgraph",
				"suse-observability-hbase-tephra-mono",
			},
			unexpectedStatefulSets: []string{
				"suse-observability-hbase-hbase-master",
				"suse-observability-hbase-hbase-rs",
				"suse-observability-hbase-hdfs-nn",
				"suse-observability-hbase-hdfs-dn",
			},
		},
		{
			name:           "100-nonha-mono-mode",
			valuesFile:     "values/global_sizing_100_nonha.yaml",
			expectMonoMode: true,
			expectedStatefulSets: []string{
				"suse-observability-hbase-stackgraph",
				"suse-observability-hbase-tephra-mono",
			},
			unexpectedStatefulSets: []string{
				"suse-observability-hbase-hbase-master",
				"suse-observability-hbase-hbase-rs",
				"suse-observability-hbase-hdfs-nn",
				"suse-observability-hbase-hdfs-dn",
			},
		},
		{
			name:           "150-ha-tephra-mode",
			valuesFile:     "values/global_sizing_150_ha.yaml",
			expectMonoMode: false,
			expectedStatefulSets: []string{
				"suse-observability-hbase-tephra",
			},
			unexpectedStatefulSets: []string{
				"suse-observability-hbase-stackgraph",
				"suse-observability-hbase-tephra-mono",
			},
		},
		{
			name:           "250-ha-tephra-mode",
			valuesFile:     "values/global_sizing_250_ha.yaml",
			expectMonoMode: false,
			expectedStatefulSets: []string{
				"suse-observability-hbase-tephra",
			},
			unexpectedStatefulSets: []string{
				"suse-observability-hbase-stackgraph",
				"suse-observability-hbase-tephra-mono",
			},
		},
		{
			name:           "500-ha-tephra-mode",
			valuesFile:     "values/global_sizing_500_ha.yaml",
			expectMonoMode: false,
			expectedStatefulSets: []string{
				"suse-observability-hbase-tephra",
			},
			unexpectedStatefulSets: []string{
				"suse-observability-hbase-stackgraph",
				"suse-observability-hbase-tephra-mono",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", tc.valuesFile)
			resources := helmtestutil.NewKubernetesResources(t, output)

			// Check expected StatefulSets exist
			for _, ssName := range tc.expectedStatefulSets {
				assert.Contains(t, resources.Statefulsets, ssName,
					"StatefulSet %s should exist for %s", ssName, tc.name)
			}

			// Check unexpected StatefulSets don't exist
			for _, ssName := range tc.unexpectedStatefulSets {
				assert.NotContains(t, resources.Statefulsets, ssName,
					"StatefulSet %s should not exist for %s", ssName, tc.name)
			}
		})
	}
}

// TestGlobalSizingResourcesAreSet tests that resources are properly set from sizing profiles
func TestGlobalSizingResourcesAreSet(t *testing.T) {
	testCases := []struct {
		name          string
		valuesFile    string
		componentName string
		componentType string // "deployment" or "statefulset"
	}{
		{
			name:          "trial-server-deployment",
			valuesFile:    "values/global_sizing_trial.yaml",
			componentName: "suse-observability-server",
			componentType: "deployment",
		},
		{
			name:          "150-ha-correlate-deployment",
			valuesFile:    "values/global_sizing_150_ha.yaml",
			componentName: "suse-observability-correlate",
			componentType: "deployment",
		},
		{
			name:          "trial-kafka-statefulset",
			valuesFile:    "values/global_sizing_trial.yaml",
			componentName: "suse-observability-kafka",
			componentType: "statefulset",
		},
		{
			name:          "500-ha-elasticsearch-statefulset",
			valuesFile:    "values/global_sizing_500_ha.yaml",
			componentName: "suse-observability-elasticsearch-master",
			componentType: "statefulset",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", tc.valuesFile)
			resources := helmtestutil.NewKubernetesResources(t, output)

			var containers []corev1.Container
			var found bool

			if tc.componentType == "deployment" {
				if deployment, exists := resources.Deployments[tc.componentName]; exists {
					containers = deployment.Spec.Template.Spec.Containers
					found = true
				}
			} else if tc.componentType == "statefulset" {
				if statefulSet, exists := resources.Statefulsets[tc.componentName]; exists {
					containers = statefulSet.Spec.Template.Spec.Containers
					found = true
				}
			}

			require.True(t, found, "Component %s of type %s should exist", tc.componentName, tc.componentType)
			require.NotEmpty(t, containers, "Component %s should have containers", tc.componentName)

			// Verify first container has resources set
			res := containers[0].Resources
			assert.NotNil(t, res.Limits.Memory(), "Component %s should have memory limits", tc.componentName)
			assert.NotNil(t, res.Requests.Memory(), "Component %s should have memory requests", tc.componentName)
			assert.NotNil(t, res.Limits.Cpu(), "Component %s should have CPU limits", tc.componentName)
			assert.NotNil(t, res.Requests.Cpu(), "Component %s should have CPU requests", tc.componentName)

			// Verify they're not zero/empty
			memLimit := res.Limits.Memory().Value()
			memRequest := res.Requests.Memory().Value()
			assert.Greater(t, memLimit, int64(0), "Component %s memory limit should be > 0", tc.componentName)
			assert.Greater(t, memRequest, int64(0), "Component %s memory request should be > 0", tc.componentName)
		})
	}
}

// TestGlobalSizingReplicaCounts tests that replica counts are set correctly by sizing profiles
func TestGlobalSizingReplicaCounts(t *testing.T) {
	testCases := []struct {
		name               string
		valuesFile         string
		componentName      string
		expectedMinReplicas int32
	}{
		{
			name:               "trial-kafka-single-replica",
			valuesFile:         "values/global_sizing_trial.yaml",
			componentName:      "suse-observability-kafka",
			expectedMinReplicas: 1,
		},
		{
			name:               "150-ha-kafka-multiple-replicas",
			valuesFile:         "values/global_sizing_150_ha.yaml",
			componentName:      "suse-observability-kafka",
			expectedMinReplicas: 3,
		},
		{
			name:               "150-ha-elasticsearch-multiple-replicas",
			valuesFile:         "values/global_sizing_150_ha.yaml",
			componentName:      "suse-observability-elasticsearch-master",
			expectedMinReplicas: 3,
		},
		{
			name:               "500-ha-correlate-multiple-replicas",
			valuesFile:         "values/global_sizing_500_ha.yaml",
			componentName:      "suse-observability-correlate",
			expectedMinReplicas: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", tc.valuesFile)
			resources := helmtestutil.NewKubernetesResources(t, output)

			// Try StatefulSet first
			if statefulSet, exists := resources.Statefulsets[tc.componentName]; exists {
				require.NotNil(t, statefulSet.Spec.Replicas, "StatefulSet %s should have replicas set", tc.componentName)
				assert.GreaterOrEqual(t, *statefulSet.Spec.Replicas, tc.expectedMinReplicas,
					"StatefulSet %s should have at least %d replicas", tc.componentName, tc.expectedMinReplicas)
				return
			}

			// Try Deployment
			if deployment, exists := resources.Deployments[tc.componentName]; exists {
				require.NotNil(t, deployment.Spec.Replicas, "Deployment %s should have replicas set", tc.componentName)
				assert.GreaterOrEqual(t, *deployment.Spec.Replicas, tc.expectedMinReplicas,
					"Deployment %s should have at least %d replicas", tc.componentName, tc.expectedMinReplicas)
				return
			}

			t.Fatalf("Component %s not found in StatefulSets or Deployments", tc.componentName)
		})
	}
}

// TestGlobalSizingProfileWinsOverDefaults verifies that when a sizing profile is active,
// profile resources are used directly (not merged with/overridden by values.yaml defaults).
func TestGlobalSizingProfileWinsOverDefaults(t *testing.T) {
	// Render 500-ha profile WITHOUT any resource overrides
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_sizing_500_ha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Check that Kafka resources are from profile, not from values.yaml defaults
	t.Run("kafka-uses-profile-resources", func(t *testing.T) {
		ss, exists := resources.Statefulsets["suse-observability-kafka"]
		require.True(t, exists, "kafka StatefulSet should exist")
		containers := ss.Spec.Template.Spec.Containers
		require.NotEmpty(t, containers)

		// Profile should set non-empty resources
		memLimit := containers[0].Resources.Limits[corev1.ResourceMemory]
		assert.True(t, memLimit.Value() > 0,
			"kafka memory limit should be > 0 from profile, got %s", memLimit.String())
	})

	// Check that elasticsearch resources come from profile, not from subchart defaults
	// Subchart default (elasticsearch/values.yaml) is 2Gi memory / 1000m CPU
	// Profile should set something larger for 500-ha
	t.Run("elasticsearch-uses-profile-resources", func(t *testing.T) {
		ss, exists := resources.Statefulsets["suse-observability-elasticsearch-master"]
		require.True(t, exists, "elasticsearch StatefulSet should exist")
		containers := ss.Spec.Template.Spec.Containers
		require.NotEmpty(t, containers)

		// elasticsearch subchart default memory limit is 2Gi
		// Profile should set it higher
		memLimit := containers[0].Resources.Limits[corev1.ResourceMemory]
		subchartDefaultMemLimit := resource.MustParse("2Gi")
		assert.True(t, memLimit.Cmp(subchartDefaultMemLimit) > 0,
			"elasticsearch memory limit should be greater than subchart default 2Gi, got %s", memLimit.String())
	})

	// Check that a stackstate component (correlate) gets profile resources
	t.Run("correlate-uses-profile-resources", func(t *testing.T) {
		dep, exists := resources.Deployments["suse-observability-correlate"]
		require.True(t, exists, "correlate deployment should exist")
		containers := dep.Spec.Template.Spec.Containers
		require.NotEmpty(t, containers)

		memLimit := containers[0].Resources.Limits[corev1.ResourceMemory]
		assert.True(t, memLimit.Value() > 0,
			"correlate memory limit should be > 0 from profile, got %s", memLimit.String())
	})
}

type expectedResources struct {
	memoryLimit string
	cpuLimit    string
}

// TestGlobalSizingUserResourceOverrides tests that user-specified resource overrides take precedence over sizing profile defaults
func TestGlobalSizingUserResourceOverrides(t *testing.T) {
	testCases := []struct {
		name         string
		valuesFile   string
		deployments  map[string]expectedResources
		statefulsets map[string]expectedResources
	}{
		{
			name:       "10-nonha",
			valuesFile: "values/global_sizing_10_nonha_resource_override.yaml",
			deployments: map[string]expectedResources{
				"suse-observability-server":   {memoryLimit: "10Gi", cpuLimit: "6000m"},
				"suse-observability-receiver": {memoryLimit: "2000Mi", cpuLimit: "4000m"},
				"suse-observability-correlate": {memoryLimit: "2500Mi", cpuLimit: "2000m"},
				"suse-observability-e2es":     {memoryLimit: "1024Mi", cpuLimit: "1000m"},
				"suse-observability-router":   {memoryLimit: "256Mi", cpuLimit: "200m"},
				"suse-observability-ui":       {memoryLimit: "128Mi", cpuLimit: "100m"},
			},
			statefulsets: map[string]expectedResources{
				"suse-observability-kafka":               {memoryLimit: "4Gi", cpuLimit: "3200m"},
				"suse-observability-zookeeper":            {memoryLimit: "1Gi", cpuLimit: "1000m"},
				"suse-observability-elasticsearch-master": {memoryLimit: "5Gi", cpuLimit: "2000m"},
				"suse-observability-hbase-stackgraph":     {memoryLimit: "4500Mi", cpuLimit: "3000m"},
				"suse-observability-hbase-tephra-mono":    {memoryLimit: "1Gi", cpuLimit: "1000m"},
				"suse-observability-victoria-metrics-0":   {memoryLimit: "3500Mi", cpuLimit: "2000m"},
				"suse-observability-clickhouse-shard0":    {memoryLimit: "8Gi", cpuLimit: "2000m"},
				"suse-observability-vmagent":              {memoryLimit: "1024Mi", cpuLimit: "400m"},
			},
		},
		{
			name:       "500-ha",
			valuesFile: "values/global_sizing_500_ha_resource_override.yaml",
			deployments: map[string]expectedResources{
				"suse-observability-api":                    {memoryLimit: "12Gi", cpuLimit: "8000m"},
				"suse-observability-checks":                 {memoryLimit: "10Gi"},
				"suse-observability-state":                  {memoryLimit: "10Gi"},
				"suse-observability-correlate":              {memoryLimit: "6000Mi", cpuLimit: "8000m"},
				"suse-observability-health-sync":            {memoryLimit: "16Gi", cpuLimit: "16000m"},
				"suse-observability-notification":           {memoryLimit: "6000Mi", cpuLimit: "6000m"},
				"suse-observability-sync":                   {memoryLimit: "16Gi", cpuLimit: "16000m"},
				"suse-observability-ui":                     {memoryLimit: "128Mi", cpuLimit: "100m"},
				"suse-observability-initializer":            {memoryLimit: "3000Mi", cpuLimit: "3000m"},
				"suse-observability-router":                 {memoryLimit: "256Mi", cpuLimit: "480m"},
				"suse-observability-slicing":                {memoryLimit: "4000Mi", cpuLimit: "3000m"},
				"suse-observability-e2es":                   {memoryLimit: "1024Mi", cpuLimit: "1000m"},
				"suse-observability-authorization-sync":     {memoryLimit: "2Gi", cpuLimit: "3000m"},
				"suse-observability-receiver-base":          {memoryLimit: "14Gi", cpuLimit: "24000m"},
				"suse-observability-receiver-logs":          {memoryLimit: "6Gi", cpuLimit: "4000m"},
				"suse-observability-receiver-process-agent": {memoryLimit: "6Gi", cpuLimit: "4000m"},
			},
			statefulsets: map[string]expectedResources{
				"suse-observability-kafka":               {memoryLimit: "8Gi", cpuLimit: "8"},
				"suse-observability-zookeeper":            {memoryLimit: "1Gi", cpuLimit: "1000m"},
				"suse-observability-elasticsearch-master": {memoryLimit: "8Gi", cpuLimit: "4000m"},
				"suse-observability-hbase-hbase-master":   {memoryLimit: "2Gi", cpuLimit: "2000m"},
				"suse-observability-hbase-hbase-rs":       {memoryLimit: "12Gi", cpuLimit: "16000m"},
				"suse-observability-hbase-tephra":         {memoryLimit: "2Gi", cpuLimit: "2000m"},
				"suse-observability-hbase-hdfs-dn":        {memoryLimit: "2Gi", cpuLimit: "2400m"},
				"suse-observability-hbase-hdfs-nn":        {memoryLimit: "2Gi", cpuLimit: "800m"},
				"suse-observability-victoria-metrics-0":   {memoryLimit: "20Gi", cpuLimit: "12"},
				"suse-observability-victoria-metrics-1":   {memoryLimit: "20Gi", cpuLimit: "12"},
				"suse-observability-clickhouse-shard0":    {memoryLimit: "8Gi", cpuLimit: "2000m"},
				"suse-observability-vmagent":              {memoryLimit: "4Gi", cpuLimit: "10000m"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", tc.valuesFile)
			resources := helmtestutil.NewKubernetesResources(t, output)

			for name, expected := range tc.deployments {
				t.Run("deployment-"+name, func(t *testing.T) {
					dep, exists := resources.Deployments[name]
					require.True(t, exists, "Deployment %s should exist", name)
					containers := dep.Spec.Template.Spec.Containers
					require.NotEmpty(t, containers)
					memLimit := containers[0].Resources.Limits[corev1.ResourceMemory]
					assert.Equal(t, resource.MustParse(expected.memoryLimit), memLimit,
						"Deployment %s memory limit should be %s", name, expected.memoryLimit)
					if expected.cpuLimit != "" {
						cpuLimit := containers[0].Resources.Limits[corev1.ResourceCPU]
						assert.Equal(t, resource.MustParse(expected.cpuLimit), cpuLimit,
							"Deployment %s CPU limit should be %s", name, expected.cpuLimit)
					}
				})
			}

			for name, expected := range tc.statefulsets {
				t.Run("statefulset-"+name, func(t *testing.T) {
					ss, exists := resources.Statefulsets[name]
					require.True(t, exists, "StatefulSet %s should exist", name)
					containers := ss.Spec.Template.Spec.Containers
					require.NotEmpty(t, containers)
					memLimit := containers[0].Resources.Limits[corev1.ResourceMemory]
					assert.Equal(t, resource.MustParse(expected.memoryLimit), memLimit,
						"StatefulSet %s memory limit should be %s", name, expected.memoryLimit)
					if expected.cpuLimit != "" {
						cpuLimit := containers[0].Resources.Limits[corev1.ResourceCPU]
						assert.Equal(t, resource.MustParse(expected.cpuLimit), cpuLimit,
							"StatefulSet %s CPU limit should be %s", name, expected.cpuLimit)
					}
				})
			}
		})
	}
}
