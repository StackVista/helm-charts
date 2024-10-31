package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

// RunConfigMapTest takes the standard full.yaml values and additional values.yaml files and extracts the specified configmap
// for verification
func RunConfigMapTest(t *testing.T, configMapKey string, extraValues []string, expectedInConfig ...string) {
	RunConfigMapTestF(t, configMapKey, extraValues, func(stringData string) {
		for _, expected := range expectedInConfig {
			require.Contains(t, stringData, expected)
		}
	})
}

func RunConfigMapTestF(t *testing.T, configMapKey string, extraValues []string, f func(stringData string)) {
	values := append([]string{"values/full.yaml"}, extraValues...)

	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", values...)

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateConfigMap corev1.ConfigMap

	for _, configMap := range resources.ConfigMaps {
		if configMap.Name == configMapKey {
			stackstateConfigMap = configMap
		}
	}
	require.NotNil(t, stackstateConfigMap)

	f(stackstateConfigMap.Data["application_stackstate.conf"])
}

// RunConfigMapTest takes the standard full.yaml values and additional values.yaml files and extracts the specified configmap
// for verification
func RunSecretTest(t *testing.T, configMapKey string, extraValues []string, expectedInSecret map[string]string) {
	RunSecretTestF(t, configMapKey, extraValues, func(inSecret map[string]string) {
		require.Equal(t, expectedInSecret, inSecret)
	})
}

func RunSecretTestF(t *testing.T, secretKey string, extraValues []string, f func(inSecret map[string]string)) {
	values := append([]string{"values/full.yaml"}, extraValues...)

	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", values...)

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateSecret corev1.Secret

	for _, secret := range resources.Secrets {
		if secret.Name == secretKey {
			stackstateSecret = secret
		}
	}
	require.NotNil(t, stackstateSecret)

	var stringData = make(map[string]string)
	for k, v := range stackstateSecret.Data {
		stringData[k] = string(v)
	}
	f(stringData)
}
