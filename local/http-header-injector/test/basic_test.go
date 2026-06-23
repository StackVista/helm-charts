package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestProvidedTlsRendersInjectorResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "http-header-injector", "values/provided-tls.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Contains(t, resources.Secrets, "http-header-injector-http-injector-cert")
	require.Contains(t, resources.Services, "http-header-injector-http-header-injector")
	require.Contains(t, resources.Deployments, "http-header-injector-http-header-injector")
	require.Contains(t, resources.ConfigMaps, "http-header-injector-http-header-injector-config")
	require.Contains(t, resources.MutatingWebhookConfigs, "http-header-injector-http-header-injector-mutatingwebhook")

	deployment := resources.Deployments["http-header-injector-http-header-injector"]
	require.Len(t, deployment.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, "http-header-injector", deployment.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "quay.io/stackstate/generic-sidecar-injector:15e759aa-1009-release", deployment.Spec.Template.Spec.Containers[0].Image)

	configMap := resources.ConfigMaps["http-header-injector-http-header-injector-config"]
	sidecarConfig := configMap.Data["sidecarconfig.yaml"]
	assert.Contains(t, sidecarConfig, `name: http-header-proxy-init`)
	assert.Contains(t, sidecarConfig, `image: "quay.io/stackstate/http-header-injector-proxy-init:1.0.0-so2"`)
	assert.Contains(t, sidecarConfig, `name: http-header-proxy`)
	assert.Contains(t, sidecarConfig, `image: "quay.io/stackstate/http-header-injector-proxy:1.38.2-so2"`)
	assert.Contains(t, sidecarConfig, `- NET_ADMIN`)
	assert.Contains(t, sidecarConfig, `- NET_RAW`)
	assert.Contains(t, sidecarConfig, `readOnlyRootFilesystem: true`)
	assert.Contains(t, sidecarConfig, `runAsUser: {% if index .Annotations "config.http-header-injector.stackstate.io/proxy-uid" %}{% index .Annotations "config.http-header-injector.stackstate.io/proxy-uid" %}{% else %}2103{% end %}`)

	mutationConfig := configMap.Data["mutationconfig.yaml"]
	assert.Contains(t, mutationConfig, `annotationNamespace: "http-header-injector.stackstate.io"`)
	assert.Contains(t, mutationConfig, `initContainersBeforePodInitContainers: [ "http-header-proxy-init" ]`)
	assert.Contains(t, mutationConfig, `containers: [ "http-header-proxy" ]`)
}

func TestGeneratedTlsRendersCertificateHooks(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "http-header-injector")
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Contains(t, resources.ClusterRoles, "http-header-injector-http-injector-cert-cluster-role")
	require.Contains(t, resources.ClusterRoleBindings, "http-header-injector-http-injector-cert-sa")
	require.Contains(t, resources.ServiceAccounts, "http-header-injector-http-injector-cert-sa")
	require.Contains(t, resources.Jobs, "http-header-injector-header-injector-generate-cert")
	require.Contains(t, resources.Jobs, "http-header-injector-header-injector-patch-cabundle")
	require.Contains(t, resources.Jobs, "http-header-injector-header-injector-delete-cert")
}
