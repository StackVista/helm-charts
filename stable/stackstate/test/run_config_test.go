package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

// RunSecretsConfigTest takes the standard full.yaml values and additional values.yaml files and extracts the specified configmap
// for verification
func RunSecretsConfigTest(t *testing.T, secretKey string, extraValues []string, expectedInConfig ...string) {
	values := append([]string{"values/full.yaml"}, extraValues...)

	output := helmtestutil.RenderHelmTemplate(t, "stackstate", values...)

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateSecret corev1.Secret

	for _, secret := range resources.Secrets {
		if secret.Name == secretKey {
			stackstateSecret = secret
		}
	}
	require.NotNil(t, stackstateSecret)

	for _, expected := range expectedInConfig {
		require.Contains(t, stackstateSecret.StringData["application_stackstate.conf"], expected)
	}
}
