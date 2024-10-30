package test

import (
	"testing"

	"github.com/caspr-io/yamlpath"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"gopkg.in/yaml.v3"
)

func TestGenerateValuesNoReceiverApiKey(t *testing.T) {
	values := renderAsYaml(t, []string{"values/baseConfig.yaml"})
	v, err := yamlpath.YamlPath(values, "global.receiverApiKey")
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
	v, err := yamlpath.YamlPath(values, "global.receiverApiKey")
	assert.NoError(t, err)
	assert.Equal(t, "stackstate-api-key-1234", v)
}

func TestGenerateValuesForSizing(t *testing.T) {
	values := renderAsYaml(t, []string{"values/sizing.yaml"})
	v, err := yamlpath.YamlPath(values, "clickhouse.replicaCount")
	assert.NoError(t, err)
	assert.NotEmpty(t, v)
}

func TestGenerateValuesForAffinity(t *testing.T) {
	values := renderAsYaml(t, []string{"values/affinity.yaml"})
	v, err := yamlpath.YamlPath(values, "clickhouse.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution")
	assert.NoError(t, err)
	assert.NotEmpty(t, v)
}

func renderAsYaml(t *testing.T, values []string) map[string]interface{} {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "stackstate-values", &helm.Options{
		ValuesFiles: values,
	})

	var valuesMap map[string]interface{}
	err := yaml.Unmarshal([]byte(output), &valuesMap)
	assert.NoError(t, err)

	// fmt.Printf("valuesMap: %v\n", valuesMap)
	return valuesMap
}
