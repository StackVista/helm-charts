package test

import (
	"testing"
)

const expectedStackPackConfig = `stackstate.stackPacks {
  installOnStartUp += "test-stackpack-1"

  installOnStartUpConfig {
    test-stackpack-1 =    {
      "bool_value": true,
      "number_value": 10,
      "string_value": "one"
    }
  }
}`

func TestStackPackConfigRenderingApi(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-api", []string{"values/stackpack_config.yaml"}, expectedStackPackConfig)
}

func TestStackPackConfigRenderingServer(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-server", []string{"values/stackpack_config.yaml", "values/split_disabled.yaml"}, expectedStackPackConfig)
}
