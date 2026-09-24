package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestExternalConfiguration(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "rbac-agent", &helm.Options{
		SetValues: map[string]string{
			"global.clusterName.fromSecret": "\\{\\{ .Release.Name }}-config",
			"global.url.fromSecret":         "\\{\\{ .Release.Name }}-config",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	require.Len(t, resources.Deployments, 1)
	assert.Empty(t, resources.ConfigMaps)
	assert.NotContains(t, resources.Secrets, "rbac-agent-config")
	for _, deployment := range resources.Deployments {
		container := deployment.Spec.Template.Spec.Containers[0]
		assert.Empty(t, container.EnvFrom)
		for _, name := range []string{"STS_URL", "STS_CLUSTER_NAME"} {
			assert.Contains(t, container.Env, corev1.EnvVar{
				Name: name,
				ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "rbac-agent-config"},
					Key:                  name,
				}},
			})
		}
	}
}

func TestLiteralConfiguration(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "rbac-agent", "../linter_values.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	require.Len(t, resources.Deployments, 1)
	require.Len(t, resources.ConfigMaps, 2)
	for _, deployment := range resources.Deployments {
		container := deployment.Spec.Template.Spec.Containers[0]
		require.Len(t, container.EnvFrom, 2)
		for _, source := range container.EnvFrom {
			require.NotNil(t, source.ConfigMapRef)
			assert.Contains(t, resources.ConfigMaps, source.ConfigMapRef.Name)
		}
	}
}
