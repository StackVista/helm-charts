package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHttpHeaderInjectorLocalDependencyRendersWhenEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/http-header-injector-enabled.yaml"},
		SetValues: map[string]string{
			"global.imageRegistry": "registry.example.test",
			"httpHeaderInjectorWebhook.sidecarInjector.image.repository": "stackstate/test-sidecar-injector",
			"httpHeaderInjectorWebhook.sidecarInjector.image.tag":        "sidecar-injector-test-tag",
			"httpHeaderInjectorWebhook.proxy.image.repository":           "stackstate/test-header-proxy",
			"httpHeaderInjectorWebhook.proxy.image.tag":                  "header-proxy-test-tag",
			"httpHeaderInjectorWebhook.proxyInit.image.repository":       "stackstate/test-header-proxy-init",
			"httpHeaderInjectorWebhook.proxyInit.image.tag":              "header-proxy-init-test-tag",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Contains(t, resources.Deployments, "suse-observability-agent-http-header-injector")
	require.Contains(t, resources.ConfigMaps, "suse-observability-agent-http-header-injector-config")
	require.Contains(t, resources.MutatingWebhookConfigs, "suse-observability-agent-http-header-injector-mutatingwebhook")

	deployment := resources.Deployments["suse-observability-agent-http-header-injector"]
	require.Len(t, deployment.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, "registry.example.test/stackstate/test-sidecar-injector:sidecar-injector-test-tag", deployment.Spec.Template.Spec.Containers[0].Image)

	configMap := resources.ConfigMaps["suse-observability-agent-http-header-injector-config"]
	sidecarConfig := configMap.Data["sidecarconfig.yaml"]
	assert.Contains(t, sidecarConfig, `image: "registry.example.test/stackstate/test-header-proxy:header-proxy-test-tag"`)
	assert.Contains(t, sidecarConfig, `image: "registry.example.test/stackstate/test-header-proxy-init:header-proxy-init-test-tag"`)
}
