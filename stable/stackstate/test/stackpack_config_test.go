package test

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestStackPackConfigRendering(t *testing.T) {
	helmChartPath, pathErr := filepath.Abs("..")
	require.NoError(t, pathErr)

	helmOpts := &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
			"values/stackpack_config.yaml",
		},
	}

	output, err := helm.RenderTemplateE(t, helmOpts, helmChartPath, "stackstate", []string{})
	require.NoError(t, err)

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateServerConfigMap corev1.ConfigMap

	for _, configMap := range resources.ConfigMaps {
		if configMap.Name == "stackstate-server" {
			stackstateServerConfigMap = configMap
		}
	}
	require.NotNil(t, stackstateServerConfigMap)

	t.Log("Data: " + stackstateServerConfigMap.Data["application_stackstate.conf"])

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
