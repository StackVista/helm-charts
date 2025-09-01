package test

import (
	"testing"
)

const expectedStackPackConfig = `stackstate.stackPacks {
  localStackPacksUri = "hdfs://suse-observability-hbase-hdfs-nn-headful:9000/stackpacks"
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

  upgradeOnStartUp = ["kubernetes-v2","stackstate-k8s-agent-v2","open-telemetry","aad-v2","stackstate","test-stackpack-1"]
  installOnStartUp += "prime-kubernetes"
  upgradeOnStartUp += "prime-kubernetes"
}`

const expectedStackPackConfigMonoHbase = `stackstate.stackPacks {
  localStackPacksUri = "file:///var/stackpacks_local"
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

  upgradeOnStartUp = ["kubernetes-v2","stackstate-k8s-agent-v2","open-telemetry","aad-v2","stackstate","test-stackpack-1"]
  installOnStartUp += "prime-kubernetes"
  upgradeOnStartUp += "prime-kubernetes"
}`

func TestStackPackConfigRenderingApi(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/stackpack_config.yaml"}, expectedStackPackConfig)
}

func TestStackPackConfigRenderingServer(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/stackpack_config.yaml", "values/split_disabled.yaml"}, expectedStackPackConfig)
}

func TestStackPackConfigRenderingApiWithMonoHbase(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/stackpack_config_mono_hbase.yaml"}, expectedStackPackConfigMonoHbase)
}

func TestStackPackConfigRenderingServerWithMonoHbase(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/stackpack_config_mono_hbase.yaml", "values/split_disabled.yaml"}, expectedStackPackConfigMonoHbase)
}
