package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestRoleGetPodsExists(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	role, ok := resources.Roles["suse-observability-get-pods"]
	require.True(t, ok, "Role 'suse-observability-get-pods' should exist")
	assert.Equal(t, "server", role.Labels["app.kubernetes.io/component"])

	require.Len(t, role.Rules, 1)
	assert.Equal(t, []string{""}, role.Rules[0].APIGroups)
	assert.Equal(t, []string{"pods"}, role.Rules[0].Resources)
	assert.Equal(t, []string{"get", "list"}, role.Rules[0].Verbs)
}

func TestRoleRbacAgentEnabledByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	// k8sAuthorization is enabled by default in full.yaml via the rbac-agent subchart
	_, roleExists := resources.Roles["suse-observability-rbac-agent"]
	_, bindingExists := resources.RoleBindings["suse-observability-rbac-agent"]
	assert.True(t, roleExists, "Role 'suse-observability-rbac-agent' should exist when k8sAuthorization is enabled")
	assert.True(t, bindingExists, "RoleBinding 'suse-observability-rbac-agent' should exist when k8sAuthorization is enabled")
}

func TestRoleRbacAgentDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"stackstate.k8sAuthorization.enabled": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	_, roleExists := resources.Roles["suse-observability-rbac-agent"]
	_, bindingExists := resources.RoleBindings["suse-observability-rbac-agent"]
	assert.False(t, roleExists, "Role 'suse-observability-rbac-agent' should not exist when k8sAuthorization is disabled")
	assert.False(t, bindingExists, "RoleBinding 'suse-observability-rbac-agent' should not exist when k8sAuthorization is disabled")
}

func TestRoleRbacAgentRules(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"stackstate.k8sAuthorization.enabled": "true",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	role, ok := resources.Roles["suse-observability-rbac-agent"]
	require.True(t, ok, "Role 'suse-observability-rbac-agent' should exist")

	require.Len(t, role.Rules, 1)
	assert.Equal(t, []string{""}, role.Rules[0].APIGroups)
	assert.Equal(t, []string{"ping"}, role.Rules[0].Resources)
	assert.Equal(t, []string{"get"}, role.Rules[0].Verbs)

	binding, ok := resources.RoleBindings["suse-observability-rbac-agent"]
	require.True(t, ok, "RoleBinding 'suse-observability-rbac-agent' should exist")

	assert.Equal(t, rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "suse-observability-rbac-agent",
	}, binding.RoleRef)
}

func TestRoleWorkloadObserverEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	role, ok := resources.Roles["suse-observability-workload-observer"]
	require.True(t, ok, "Role 'suse-observability-workload-observer' should exist")

	require.Len(t, role.Rules, 1)
	assert.Equal(t, []string{"apps"}, role.Rules[0].APIGroups)
	assert.Equal(t, []string{"statefulsets", "deployments"}, role.Rules[0].Resources)
	assert.Equal(t, []string{"get", "list", "watch"}, role.Rules[0].Verbs)

	binding, ok := resources.RoleBindings["suse-observability-workload-observer"]
	require.True(t, ok, "RoleBinding 'suse-observability-workload-observer' should exist")

	assert.Equal(t, rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "suse-observability-workload-observer",
	}, binding.RoleRef)

	require.Len(t, binding.Subjects, 1)
	assert.Equal(t, "ServiceAccount", binding.Subjects[0].Kind)
	assert.Equal(t, "suse-observability-workload-observer", binding.Subjects[0].Name)
}

func TestRoleWorkloadObserverDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"stackstate.components.workloadObserver.enabled": "false",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	_, roleExists := resources.Roles["suse-observability-workload-observer"]
	_, bindingExists := resources.RoleBindings["suse-observability-workload-observer"]
	assert.False(t, roleExists, "Role 'suse-observability-workload-observer' should not exist when workloadObserver is disabled")
	assert.False(t, bindingExists, "RoleBinding 'suse-observability-workload-observer' should not exist when workloadObserver is disabled")
}
