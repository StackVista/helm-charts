package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

const (
	normalPullSecretName = "suse-observability-pull-secret"
	hookPullSecretName   = "suse-observability-pull-secret-hook"
)

// TestPullSecretHookHasDistinctName verifies that the hook-managed pull secret and the normal pull secret
// render as two separate resources with different names. They previously shared a name, which made GitOps
// tooling (ArgoCD/Flux) and `helm template | kubectl apply` see two resources with the same identity.
func TestPullSecretHookHasDistinctName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_suse_observability_pull_secret_nonha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	normal, ok := resources.Secrets[normalPullSecretName]
	require.True(t, ok, "normal pull secret %q should exist", normalPullSecretName)
	assert.NotContains(t, normal.Annotations, "helm.sh/hook", "normal pull secret should not be a Helm hook")

	hook, ok := resources.Secrets[hookPullSecretName]
	require.True(t, ok, "hook pull secret %q should exist", hookPullSecretName)

	require.NotEqual(t, normal.Name, hook.Name, "hook and normal pull secret must not share a name")
	assert.Equal(t, "pre-install,pre-upgrade,post-delete", hook.Annotations["helm.sh/hook"],
		"hook pull secret should run on pre-install/pre-upgrade/post-delete")
	assert.Equal(t, normal.Type, hook.Type, "hook pull secret should be the same dockerconfigjson type as the normal one")
	assert.Equal(t, normal.Data, hook.Data, "hook pull secret should carry the same credentials as the normal one")
}

// TestPullSecretHookUsedByRouterModeJobs verifies that, in automatic router mode, the pre-install/PreSync
// set-maintenance job references the hook-managed pull secret (the normal one does not exist yet at that
// phase), while the post-install set-active job references the normal pull secret.
func TestPullSecretHookUsedByRouterModeJobs(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/global_suse_observability_pull_secret_nonha.yaml"},
		SetValues: map[string]string{
			"stackstate.components.router.mode.status": "automatic",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	maintenance := findJob(&resources, "set-maintenance")
	require.NotNil(t, maintenance, "set-maintenance job should render in automatic router mode")
	assertSingleImagePullSecret(t, maintenance.Spec.Template.Spec.ImagePullSecrets, hookPullSecretName)

	active := findJob(&resources, "set-active")
	require.NotNil(t, active, "set-active job should render in automatic router mode")
	assertSingleImagePullSecret(t, active.Spec.Template.Spec.ImagePullSecrets, normalPullSecretName)
}

// TestPullSecretHookArgoCDCompatibility verifies that with deployment.compatibleWithArgoCD the hook pull
// secret becomes an ArgoCD PreSync hook (not a Helm hook) and keeps its distinct name, so ArgoCD does not
// report a duplicate Secret/suse-observability-pull-secret.
func TestPullSecretHookArgoCDCompatibility(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/global_suse_observability_pull_secret_nonha.yaml"},
		SetValues: map[string]string{
			"deployment.compatibleWithArgoCD": "true",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	_, ok := resources.Secrets[normalPullSecretName]
	require.True(t, ok, "normal pull secret %q should exist", normalPullSecretName)

	hook, ok := resources.Secrets[hookPullSecretName]
	require.True(t, ok, "hook pull secret %q should exist", hookPullSecretName)
	assert.Equal(t, "PreSync", hook.Annotations["argocd.argoproj.io/hook"], "hook pull secret should be an ArgoCD PreSync hook")
	assert.NotContains(t, hook.Annotations, "helm.sh/hook", "under ArgoCD the hook pull secret must not use Helm hooks")
}

func assertSingleImagePullSecret(t *testing.T, secrets []corev1.LocalObjectReference, expected string) {
	require.Len(t, secrets, 1, "expected exactly one imagePullSecret")
	assert.Equal(t, expected, secrets[0].Name)
}
