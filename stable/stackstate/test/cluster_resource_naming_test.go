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

	require.Equal(t, 2, len(resources.ClusterRoleBindings))
	require.Equal(t, "stackstate-authentication", resources.ClusterRoleBindings[0].Name)
	require.Equal(t, "stackstate-authorization", resources.ClusterRoleBindings[1].Name)
	require.Equal(t, 1, len(resources.ClusterRoles))
	require.Equal(t, "stackstate-authorization", resources.ClusterRoles[0].Name)
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

	require.Equal(t, 2, len(resources.ClusterRoleBindings))
	require.Equal(t, "devver-stackstate-authentication", resources.ClusterRoleBindings[0].Name)
	require.Equal(t, "devver-stackstate-authorization", resources.ClusterRoleBindings[1].Name)
	require.Equal(t, 1, len(resources.ClusterRoles))
	require.Equal(t, "devver-stackstate-authorization", resources.ClusterRoles[0].Name)
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

	require.Equal(t, 2, len(resources.ClusterRoleBindings))
	require.Equal(t, "devver-stacky-stackstate-authentication", resources.ClusterRoleBindings[0].Name)
	require.Equal(t, "devver-stacky-stackstate-authorization", resources.ClusterRoleBindings[1].Name)
	require.Equal(t, 1, len(resources.ClusterRoles))
	require.Equal(t, "devver-stacky-stackstate-authorization", resources.ClusterRoles[0].Name)
}
