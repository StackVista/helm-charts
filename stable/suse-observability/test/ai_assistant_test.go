package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func envVarByName(envVars []corev1.EnvVar, name string) *corev1.EnvVar {
	for i := range envVars {
		if envVars[i].Name == name {
			return &envVars[i]
		}
	}

	return nil
}

func TestAiAssistantEnabledWithAiAssistant(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	statefulSet, ok := resources.Statefulsets["suse-observability-ai-assistant"]
	require.True(t, ok, "AI Assistant StatefulSet should exist")

	service, ok := resources.Services["suse-observability-ai-assistant"]
	require.True(t, ok, "AI Assistant service should exist")

	serviceAccount, ok := resources.ServiceAccounts["suse-observability-ai-assistant"]
	require.True(t, ok, "AI Assistant service account should exist")

	assert.Equal(t, "suse-observability-ai-assistant", serviceAccount.Name)

	container := statefulSet.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "ai-assistant", container.Name)
	assert.Contains(t, container.Args, "serve")

	// Verify MCP_SERVER_URL env var points to MCP service
	assert.Contains(t, container.Env, corev1.EnvVar{
		Name:  "MCP_SERVER_URL",
		Value: "http://suse-observability-mcp:8080",
	})

	// Verify SQLite path env var
	assert.Contains(t, container.Env, corev1.EnvVar{
		Name:  "SQLITE_PATH",
		Value: "/data/ai-assistant.db",
	})

	// Verify AWS Bedrock is enabled
	assert.Contains(t, container.Env, corev1.EnvVar{
		Name:  "USE_BEDROCK",
		Value: "true",
	})
	assert.Contains(t, container.Env, corev1.EnvVar{
		Name:  "AWS_REGION",
		Value: "eu-west-1",
	})
	assert.Nil(t, envVarByName(container.Env, "ANTHROPIC_API_KEY"))

	// Verify service port
	require.Len(t, service.Spec.Ports, 1)
	assert.Equal(t, int32(8081), service.Spec.Ports[0].Port)

	// Verify service account is referenced in the StatefulSet
	assert.Equal(t, "suse-observability-ai-assistant", statefulSet.Spec.Template.Spec.ServiceAccountName)

	// Verify PVC is defined
	require.Len(t, statefulSet.Spec.VolumeClaimTemplates, 1)
	assert.Equal(t, "data", statefulSet.Spec.VolumeClaimTemplates[0].Name)

	// Verify volume mount for SQLite data
	require.NotEmpty(t, container.VolumeMounts)
	var dataMount *corev1.VolumeMount
	for i := range container.VolumeMounts {
		if container.VolumeMounts[i].Name == "data" {
			dataMount = &container.VolumeMounts[i]
			break
		}
	}
	require.NotNil(t, dataMount, "data volume mount should exist")
	assert.Equal(t, "/data", dataMount.MountPath)
}

func TestAiAssistantDisabledWhenAiAssistantDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"ai.assistant.enabled": "false",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)
	assert.NotContains(t, resources.Statefulsets, "suse-observability-ai-assistant")
	assert.NotContains(t, resources.Services, "suse-observability-ai-assistant")
	assert.NotContains(t, resources.ServiceAccounts, "suse-observability-ai-assistant")
}

func TestAiAssistantServiceAccountAnnotations(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"stackstate.components.aiAssistant.serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn": "arn:aws:iam::123456789012:role/my-ai-assistant-role",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	serviceAccount, ok := resources.ServiceAccounts["suse-observability-ai-assistant"]
	require.True(t, ok, "AI Assistant service account should exist")

	assert.Equal(t, "arn:aws:iam::123456789012:role/my-ai-assistant-role", serviceAccount.Annotations["eks.amazonaws.com/role-arn"])
}

func TestAiAssistantAnthropicProvider(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"ai.assistant.provider":         "anthropic",
			"ai.assistant.anthropic.apiKey": "sk-ant-test",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	statefulSet, ok := resources.Statefulsets["suse-observability-ai-assistant"]
	require.True(t, ok, "AI Assistant StatefulSet should exist")

	secret, ok := resources.Secrets["suse-observability-ai-assistant"]
	require.True(t, ok, "AI Assistant secret should exist")
	assert.Equal(t, []byte("sk-ant-test"), secret.Data["ANTHROPIC_API_KEY"])

	container := statefulSet.Spec.Template.Spec.Containers[0]
	assert.Contains(t, container.Env, corev1.EnvVar{
		Name:  "USE_BEDROCK",
		Value: "false",
	})
	assert.Nil(t, envVarByName(container.Env, "AWS_REGION"))

	anthropicAPIKey := envVarByName(container.Env, "ANTHROPIC_API_KEY")
	require.NotNil(t, anthropicAPIKey, "ANTHROPIC_API_KEY env var should exist")
	require.NotNil(t, anthropicAPIKey.ValueFrom, "ANTHROPIC_API_KEY should come from a secret")
	require.NotNil(t, anthropicAPIKey.ValueFrom.SecretKeyRef, "ANTHROPIC_API_KEY should use secretKeyRef")
	assert.Equal(t, "suse-observability-ai-assistant", anthropicAPIKey.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "ANTHROPIC_API_KEY", anthropicAPIKey.ValueFrom.SecretKeyRef.Key)
}

func TestAiAssistantAnthropicProviderWithExternalSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"ai.assistant.provider":                          "anthropic",
			"ai.assistant.anthropic.fromExternalSecret.name": "my-anthropic-secret",
			"ai.assistant.anthropic.fromExternalSecret.key":  "api-token",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	_, ok := resources.Secrets["suse-observability-ai-assistant"]
	assert.False(t, ok, "AI Assistant secret should not be created when using an external Anthropic secret")

	statefulSet, ok := resources.Statefulsets["suse-observability-ai-assistant"]
	require.True(t, ok, "AI Assistant StatefulSet should exist")

	container := statefulSet.Spec.Template.Spec.Containers[0]
	anthropicAPIKey := envVarByName(container.Env, "ANTHROPIC_API_KEY")
	require.NotNil(t, anthropicAPIKey, "ANTHROPIC_API_KEY env var should exist")
	require.NotNil(t, anthropicAPIKey.ValueFrom, "ANTHROPIC_API_KEY should come from a secret")
	require.NotNil(t, anthropicAPIKey.ValueFrom.SecretKeyRef, "ANTHROPIC_API_KEY should use secretKeyRef")
	assert.Equal(t, "my-anthropic-secret", anthropicAPIKey.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "api-token", anthropicAPIKey.ValueFrom.SecretKeyRef.Key)
}

func TestAiAssistantImageTagRequired(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability",
		"values/full.yaml",
		"values/ai_assistant_missing_image_tag.yaml",
	)
	require.Contains(t, err.Error(), "stackstate.components.aiAssistant.image.tag must be set")
}

func TestAiAssistantAnthropicAPIKeyRequired(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability",
		"values/full.yaml",
		"values/ai_assistant_anthropic_missing_api_key.yaml",
	)
	require.Contains(t, err.Error(), "ai.assistant.anthropic.apiKey must be set when ai.assistant.provider=anthropic and ai.assistant.anthropic.fromExternalSecret.name is empty")
}
