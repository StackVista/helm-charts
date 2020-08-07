package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestStackPackConfigRendering(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/stackpack_config.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateServerConfigMap corev1.ConfigMap

	for _, configMap := range resources.ConfigMaps {
		if configMap.Name == "stackstate-server" {
			stackstateServerConfigMap = configMap
		}
	}
	require.NotNil(t, stackstateServerConfigMap)

	expectedConfig := `stackstate.stackPacks {
  installOnStartUp += "test-stackpack-1"

  installOnStartUpConfig {
    test-stackpack-1 =    {
      "bool_value": true,
      "number_value": 10,
      "string_value": "one"
    }
  }
}`
	require.Contains(t, stackstateServerConfigMap.Data["application_stackstate.conf"], expectedConfig)
}
