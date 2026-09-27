package test

import (
	"slices"
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

func TestApiKeyConfiguration(t *testing.T) {
	for _, scenario := range []struct {
		name       string
		values     map[string]string
		secretName string
	}{
		{
			name:       "external API key with literal configuration",
			values:     map[string]string{"global.apiKey.fromSecret": "\\{\\{ .Release.Name }}-shared"},
			secretName: "rbac-agent-shared",
		},
		{
			name:       "generated API key secret",
			values:     map[string]string{"apiKey": "test-api-key"},
			secretName: "rbac-agent-rbac-agent-api-key",
		},
		{
			name: "external API key takes precedence",
			values: map[string]string{
				"global.apiKey.fromSecret": "shared",
				"apiKey":                   "test-api-key",
			},
			secretName: "shared",
		},
		{name: "service account authentication"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "rbac-agent", &helm.Options{
				ValuesFiles: []string{"../linter_values.yaml"},
				SetValues:   scenario.values,
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			require.Len(t, resources.Deployments, 1)
			for _, deployment := range resources.Deployments {
				container := deployment.Spec.Template.Spec.Containers[0]
				require.Len(t, container.EnvFrom, 2)
				for _, source := range container.EnvFrom {
					assert.Nil(t, source.SecretRef, "API-key secrets must not import unrelated settings")
					require.NotNil(t, source.ConfigMapRef)
					assert.Contains(t, resources.ConfigMaps, source.ConfigMapRef.Name)
				}
				apiKeyPosition := slices.IndexFunc(container.Env, func(variable corev1.EnvVar) bool {
					return variable.Name == "STS_API_KEY"
				})
				serviceAccount := corev1.EnvVar{Name: "STS_K8S_SERVICE_ACCOUNT", Value: "true"}
				if scenario.secretName == "" {
					assert.Equal(t, -1, apiKeyPosition)
					assert.Contains(t, container.Env, serviceAccount)
				} else {
					require.NotEqual(t, -1, apiKeyPosition)
					assert.Equal(t, corev1.EnvVar{
						Name: "STS_API_KEY",
						ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: scenario.secretName},
							Key:                  "STS_API_KEY",
						}},
					}, container.Env[apiKeyPosition])
					assert.NotContains(t, container.Env, serviceAccount)
				}
			}
		})
	}
}
