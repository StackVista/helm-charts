package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHttpHeaderInjectorLocalDependencyRendersWhenEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/http-header-injector-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Contains(t, resources.Deployments, "suse-observability-agent-http-header-injector")
	require.Contains(t, resources.ConfigMaps, "suse-observability-agent-http-header-injector-config")
	require.Contains(t, resources.MutatingWebhookConfigs, "suse-observability-agent-http-header-injector-mutatingwebhook")

	deployment := resources.Deployments["suse-observability-agent-http-header-injector"]
	require.Len(t, deployment.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, "quay.io/stackstate/generic-sidecar-injector:15e759aa-1009-release", deployment.Spec.Template.Spec.Containers[0].Image)

	configMap := resources.ConfigMaps["suse-observability-agent-http-header-injector-config"]
	sidecarConfig := configMap.Data["sidecarconfig.yaml"]
	assert.Contains(t, sidecarConfig, `image: "quay.io/stackstate/http-header-injector-proxy:1.38.2-so2"`)
	assert.Contains(t, sidecarConfig, `image: "quay.io/stackstate/http-header-injector-proxy-init:1.0.0-so2"`)
}
