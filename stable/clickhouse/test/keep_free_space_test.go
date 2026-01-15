package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// TestClickhouseMergeTreeDefaults tests that default mergeTree settings are applied correctly
func TestClickhouseMergeTreeDefaults(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "clickhouse", "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Check ConfigMap exists
	configMap, ok := resources.ConfigMaps["clickhouse"]
	assert.True(t, ok, "clickhouse ConfigMap should exist")

	// Get the configuration data
	configData, ok := configMap.Data["00_default_overrides.xml"]
	assert.True(t, ok, "00_default_overrides.xml should exist in ConfigMap")

	// Verify storage_configuration is present with fixed 1 GiB
	// Note: Helm renders large integers in scientific notation
	assert.Contains(t, configData, "<storage_configuration>", "ConfigMap should contain storage_configuration")
	assert.Contains(t, configData, "<keep_free_space_bytes>1.073741824e+09</keep_free_space_bytes>",
		"keep_free_space_bytes should be 1 GiB (1.073741824e+09 in scientific notation)")

	// Verify merge_tree configuration is present with default 0.10 ratio
	assert.Contains(t, configData, "<merge_tree>", "ConfigMap should contain merge_tree section")
	assert.Contains(t, configData, "<min_free_disk_ratio_to_perform_insert>0.1</min_free_disk_ratio_to_perform_insert>",
		"min_free_disk_ratio_to_perform_insert should be 0.1")
}

// TestClickhouseMergeTreeCustomValues tests that custom mergeTree settings are applied correctly
func TestClickhouseMergeTreeCustomValues(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "clickhouse", "values/keep-free-space.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Check ConfigMap exists
	configMap, ok := resources.ConfigMaps["clickhouse"]
	assert.True(t, ok, "clickhouse ConfigMap should exist")

	// Get the configuration data
	configData, ok := configMap.Data["00_default_overrides.xml"]
	assert.True(t, ok, "00_default_overrides.xml should exist in ConfigMap")

	// Verify custom keep_free_space_bytes (2 GiB)
	// Note: Helm renders large integers in scientific notation
	assert.Contains(t, configData, "<storage_configuration>", "ConfigMap should contain storage_configuration")
	assert.Contains(t, configData, "<keep_free_space_bytes>2.147483648e+09</keep_free_space_bytes>",
		"keep_free_space_bytes should be 2 GiB (2.147483648e+09 in scientific notation)")

	// Verify custom min_free_disk_ratio_to_perform_insert (0.15)
	assert.Contains(t, configData, "<merge_tree>", "ConfigMap should contain merge_tree section")
	assert.Contains(t, configData, "<min_free_disk_ratio_to_perform_insert>0.15</min_free_disk_ratio_to_perform_insert>",
		"min_free_disk_ratio_to_perform_insert should be 0.15")
}
