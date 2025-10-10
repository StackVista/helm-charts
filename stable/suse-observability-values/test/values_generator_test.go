package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/imdario/mergo"
	"github.com/stretchr/testify/require"

	"github.com/caspr-io/yamlpath"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"gopkg.in/yaml.v3"
)

func TestGenerateValuesNoReceiverApiKey(t *testing.T) {
	values := renderAsYaml(t, []string{"values/baseConfig.yaml"})
	v, err := yamlpath.YamlPath(values, "stackstate.apiKey.key")
	assert.NoError(t, err)
	assert.NotEmpty(t, v)
}

func TestGenerateValuesPullSecretNotGenerated(t *testing.T) {
	values := renderAsYaml(t, []string{"values/baseConfig.yaml"})
	v, err := yamlpath.YamlPath(values, "pull-secret")
	assert.NoError(t, err)
	assert.Empty(t, v)
	v, err = yamlpath.YamlPath(values, "global.imagePullSecrets")
	assert.Empty(t, v)
	assert.NoError(t, err)
}

func TestGenerateValuesWithGeneratesBcryptPasswords(t *testing.T) {
	values := renderAsYaml(t, []string{"values/baseConfig.yaml"})
	v, err := yamlpath.YamlPath(values, "stackstate.authentication.adminPassword")
	assert.NoError(t, err)
	assert.Contains(t, v, "$2a$10$")

}

func TestGenerateValuesWithPullSecret(t *testing.T) {
	values := renderAsYaml(t, []string{"values/baseConfig.yaml", "values/pullsecret.yaml"})
	v, err := yamlpath.YamlPath(values, "global.imagePullSecrets[0]")
	assert.NoError(t, err)
	assert.Contains(t, v, "suse-observability-pull-secret")

	v, err = yamlpath.YamlPath(values, "pull-secret.credentials[0].username")
	assert.NoError(t, err)
	assert.Contains(t, v, "john")

	v, err = yamlpath.YamlPath(values, "pull-secret.credentials[0].password")
	assert.NoError(t, err)
	assert.Contains(t, v, "doe")

	v, err = yamlpath.YamlPath(values, "pull-secret.credentials[0].registry")
	assert.NoError(t, err)
	assert.Contains(t, v, "registry.rancher.com")
}

func TestGenerateValuesWithPlaintextPasswords(t *testing.T) {
	values := renderAsYaml(t, []string{"values/baseConfig.yaml", "values/plaintextpasswords.yaml"})
	v, err := yamlpath.YamlPath(values, "stackstate.authentication.adminPassword")
	assert.NoError(t, err)
	assert.Contains(t, v, "$2a$10$")
}

func TestGenerateValuesWithBcryptPasswords(t *testing.T) {
	values := renderAsYaml(t, []string{"values/baseConfig.yaml", "values/bcryptpasswords.yaml"})
	v, err := yamlpath.YamlPath(values, "stackstate.authentication.adminPassword")
	assert.NoError(t, err)
	assert.Equal(t, v, "$2a$10$qCWqX0H9E1crJ3tibX7ChuPSmkd2T6sBcSEzZc6gPBYH7Vm.qQKH.")
}

func TestGenerateValuesSetReceiverApiKey(t *testing.T) {
	values := renderAsYaml(t, []string{"values/baseConfig.yaml", "values/apikey.yaml"})
	v, err := yamlpath.YamlPath(values, "stackstate.apiKey.key")
	assert.NoError(t, err)
	assert.Equal(t, "stackstate-api-key-1234", v)
}

func TestGenerateValuesForSizing(t *testing.T) {
	values := renderAsYaml(t, []string{"values/sizing.yaml"})
	v, err := yamlpath.YamlPath(values, "clickhouse.replicaCount")
	assert.NoError(t, err)
	assert.NotEmpty(t, v)
}

func TestGenerateValuesForAffinity2(t *testing.T) {
	values := renderAsYaml(t, []string{"values/baseConfig.yaml", "values/nonHA_nodeAffinity.yaml"})
	v, err := yamlpath.YamlPath(values, "clickhouse.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution")
	assert.NoError(t, err)
	assert.NotEmpty(t, v)
}

func TestGenerateValuesForAffinity(t *testing.T) {

	baseConfigValues := "values/baseConfig.yaml"
	tests := []struct {
		name           string
		valuesPath     string
		expectedValues []string
		unwantedValues []string
	}{
		{
			name:       "nonHa profile with nodeAffinity configured",
			valuesPath: "values/nonHA_nodeAffinity.yaml",
			expectedValues: []string{
				"clickhouse.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"elasticsearch.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.all.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"kafka.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"opentelemetry-collector.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"stackstate.components.all.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-0.server.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-1.server.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"zookeeper.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
			},
			unwantedValues: []string{
				"clickhouse.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"elasticsearch.antiAffinity",
				"elasticsearch.antiAffinityTopologyKey",
				"hbase.hbase.master.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hbase.regionserver.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.datanode.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.namenode.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.secondarynamenode.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.tephra.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"kafka.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-0.server.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-1.server.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"zookeeper.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
			},
		},
		{
			name:           "nonHa profile with no nodeAffinity configured",
			valuesPath:     "values/nonHA_NoNodeAffinity.yaml",
			expectedValues: []string{},
			unwantedValues: []string{
				"clickhouse.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"elasticsearch.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.all.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"kafka.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"opentelemetry-collector.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"stackstate.components.all.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-0.server.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-1.server.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"zookeeper.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
			},
		},
		{
			name:           "nonHa profile with podAntiAffinity configured",
			valuesPath:     "values/nonHA_podAntiAffinity.yaml",
			expectedValues: []string{},
			unwantedValues: []string{
				"clickhouse.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"elasticsearch.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.all.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"kafka.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"opentelemetry-collector.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"stackstate.components.all.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-0.server.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-1.server.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"zookeeper.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
			},
		},
		{
			name:       "Ha profile with podAntiAffinity Configured",
			valuesPath: "values/HA_podAntiAffinity.yaml",
			expectedValues: []string{
				"clickhouse.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"elasticsearch.antiAffinity",
				"elasticsearch.antiAffinityTopologyKey",
				"hbase.hbase.master.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hbase.regionserver.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.datanode.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.namenode.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.secondarynamenode.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.tephra.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"kafka.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-0.server.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-1.server.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"zookeeper.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
			},
			unwantedValues: []string{
				"clickhouse.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"elasticsearch.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.all.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"kafka.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"opentelemetry-collector.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"stackstate.components.all.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-0.server.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-1.server.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"zookeeper.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
			},
		},
		{
			name:       "Ha profile with nodeAffinity and preferred podAntiAffinity Configured",
			valuesPath: "values/HA_nodeAffinity_podAntiAffinity.yaml",
			expectedValues: []string{
				"clickhouse.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"elasticsearch.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.all.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"kafka.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"opentelemetry-collector.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"stackstate.components.all.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-0.server.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-1.server.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"zookeeper.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"clickhouse.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
				"elasticsearch.antiAffinity",
				"elasticsearch.antiAffinityTopologyKey",
				"hbase.hbase.master.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
				"hbase.hbase.regionserver.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.datanode.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.namenode.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.secondarynamenode.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
				"hbase.tephra.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
				"kafka.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-0.server.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-1.server.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
				"zookeeper.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution",
			},
			unwantedValues: []string{
				"clickhouse.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"elasticsearch.antiAffinity",
				"elasticsearch.antiAffinityTopologyKey",
				"hbase.hbase.master.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hbase.regionserver.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.datanode.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.namenode.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.hdfs.secondarynamenode.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"hbase.tephra.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"kafka.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-0.server.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"victoria-metrics-1.server.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
				"zookeeper.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := renderAsYaml(t, []string{baseConfigValues, tt.valuesPath})
			for i, expectedValue := range tt.expectedValues {
				v, err := yamlpath.YamlPath(values, expectedValue)
				assert.NoErrorf(t, err, fmt.Sprintf("missing value %v: %s", i, expectedValue))
				assert.NotEmpty(t, v, fmt.Sprintf("expect not empty value %v: %s", i, expectedValue))

			}
		})
	}
}

func TestGenerateValuesWithInvalidBaseUrl(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "stackstate-values", &helm.Options{
		ValuesFiles: []string{"values/base_url_schema.yaml"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid URL format")
	assert.Contains(t, err.Error(), "must include a scheme")
}

func TestGenerateValuesWithoutBaseConfigAndMissingBaseUrl(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "stackstate-values", &helm.Options{
		ValuesFiles: []string{"values/missing_base_url_without_base_config.yaml"},
	})
	assert.NoError(t, err)
}

func renderAsYaml(t *testing.T, values []string) map[string]interface{} {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "stackstate-values", &helm.Options{
		ValuesFiles: values,
	})

	var valuesMap map[string]interface{}
	err := yaml.Unmarshal([]byte(output), &valuesMap)
	assert.NoError(t, err)

	decoder := yaml.NewDecoder(strings.NewReader(output))

	for {
		var doc map[string]interface{}
		err := decoder.Decode(&doc)
		if err != nil {
			if err.Error() == "EOF" {
				break // End of documents
			}
			assert.NoError(t, err)
		}

		err = mergo.Merge(&valuesMap, doc, mergo.WithOverride)
		require.NoErrorf(t, err, "Failed to merge map with values: %v", valuesMap)
	}

	return valuesMap
}
