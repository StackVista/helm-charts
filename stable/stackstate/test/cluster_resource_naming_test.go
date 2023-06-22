package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestClusterRoleDeployedToSameNamespaceAsChartName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "stackstate", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "stackstate",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 3, len(resources.ClusterRoleBindings))
	require.Contains(t, resources.ClusterRoleBindings, "stackstate-authentication")
	require.Contains(t, resources.ClusterRoleBindings, "stackstate-authorization")
	require.Equal(t, 1, len(resources.ClusterRoles))
	require.Contains(t, resources.ClusterRoles, "stackstate-authorization")
}
func TestClusterRoleDeployedToDifferentNamespaceAsChartName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "stackstate", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "devver",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 3, len(resources.ClusterRoleBindings))
	require.Contains(t, resources.ClusterRoleBindings, "devver-stackstate-authentication")
	require.Contains(t, resources.ClusterRoleBindings, "devver-stackstate-authorization")
	require.Equal(t, 1, len(resources.ClusterRoles))
	require.Contains(t, resources.ClusterRoles, "devver-stackstate-authorization")
}

func TestClusterRoleNameWhenNamespaceReleaseNameAndChartNameAllDifferent(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "stacky", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "devver",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 3, len(resources.ClusterRoleBindings))
	require.Contains(t, resources.ClusterRoleBindings, "devver-stacky-stackstate-authentication")
	require.Contains(t, resources.ClusterRoleBindings, "devver-stacky-stackstate-authorization")
	require.Equal(t, 1, len(resources.ClusterRoles))
	require.Contains(t, resources.ClusterRoles, "devver-stacky-stackstate-authorization")
}
