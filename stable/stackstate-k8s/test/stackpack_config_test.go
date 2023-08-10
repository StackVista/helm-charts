package test

import (
	"testing"
)

const expectedStackPackConfig = `stackstate.stackPacks {
  localStackPacksUri = "hdfs://stackstate-k8s-hbase-hdfs-nn-headful:9000/stackpacks"
  latestVersionsStackPackStoreUri = "file:///var/stackpacks"

  updateStackPacksInterval = "5 minutes"
  installOnStartUp += "test-stackpack-1"

  installOnStartUpConfig {
    test-stackpack-1 =    {
      "bool_value": true,
      "number_value": 10,
      "string_value": "one"
    }
  }
  upgradeOnStartUp = ["test-stackpack-1"]
}`

func TestStackPackConfigRenderingApi(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-k8s-api", []string{"values/stackpack_config.yaml"}, expectedStackPackConfig)
}

func TestStackPackConfigRenderingServer(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-k8s-server", []string{"values/stackpack_config.yaml", "values/split_disabled.yaml"}, expectedStackPackConfig)
}
