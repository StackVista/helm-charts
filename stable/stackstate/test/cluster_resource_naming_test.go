package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestClusterRoleDeployedToSameNamespaceAsChartName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOpts(t, "stackstate", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "stackstate",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, len(resources.ClusterRoleBindings), 1)
	require.Equal(t, resources.ClusterRoleBindings[0].Name, "stackstate-server")
	require.Equal(t, len(resources.ClusterRoles), 1)
	require.Equal(t, resources.ClusterRoles[0].Name, "stackstate-server")
}
func TestClusterRoleDeployedToDifferentNamespaceAsChartName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOpts(t, "stackstate", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "devver",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, len(resources.ClusterRoleBindings), 1)
	require.Equal(t, resources.ClusterRoleBindings[0].Name, "devver-stackstate-server")
	require.Equal(t, len(resources.ClusterRoles), 1)
	require.Equal(t, resources.ClusterRoles[0].Name, "devver-stackstate-server")
}
func TestClusterRoleNameWhenNamespaceReleaseNameAndChartNameAllDifferent(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOpts(t, "stacky", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "devver",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, len(resources.ClusterRoleBindings), 1)
	require.Equal(t, resources.ClusterRoleBindings[0].Name, "devver-stacky-stackstate-server")
	require.Equal(t, len(resources.ClusterRoles), 1)
	require.Equal(t, resources.ClusterRoles[0].Name, "devver-stacky-stackstate-server")
}
