package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestClusterRoleDeployedToSameNamespaceAsChartName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "stackstate-k8s", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "stackstate-k8s",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 3, len(resources.ClusterRoleBindings))
	require.Contains(t, resources.ClusterRoleBindings, "stackstate-k8s-authentication")
	require.Contains(t, resources.ClusterRoleBindings, "stackstate-k8s-authorization")
	require.Contains(t, resources.ClusterRoleBindings, "stackstate-k8s-victoria-metrics-cluster-clusterrolebinding")
	require.Equal(t, 2, len(resources.ClusterRoles))
	require.Contains(t, resources.ClusterRoles, "stackstate-k8s-authorization")
	require.Contains(t, resources.ClusterRoles, "stackstate-k8s-victoria-metrics-cluster-clusterrole")
}
func TestClusterRoleDeployedToDifferentNamespaceAsChartName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "stackstate-k8s", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "devver",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Equal(t, 3, len(resources.ClusterRoleBindings))
	require.Contains(t, resources.ClusterRoleBindings, "devver-stackstate-k8s-authentication")
	require.Contains(t, resources.ClusterRoleBindings, "devver-stackstate-k8s-authorization")
	require.Contains(t, resources.ClusterRoleBindings, "stackstate-k8s-victoria-metrics-cluster-clusterrolebinding")
	require.Equal(t, 2, len(resources.ClusterRoles))
	require.Contains(t, resources.ClusterRoles, "devver-stackstate-k8s-authorization")
	require.Contains(t, resources.ClusterRoles, "stackstate-k8s-victoria-metrics-cluster-clusterrole")
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
	require.Contains(t, resources.ClusterRoleBindings, "devver-stacky-stackstate-k8s-authentication")
	require.Contains(t, resources.ClusterRoleBindings, "devver-stacky-stackstate-k8s-authorization")
	require.Contains(t, resources.ClusterRoleBindings, "stacky-victoria-metrics-cluster-clusterrolebinding")
	require.Equal(t, 2, len(resources.ClusterRoles))
	require.Contains(t, resources.ClusterRoles, "devver-stacky-stackstate-k8s-authorization")
	require.Contains(t, resources.ClusterRoles, "stacky-victoria-metrics-cluster-clusterrole")
}

func TestResourcesNamesLength(t *testing.T) {
	// 23 chars is the max release name length
	// 12 chars is the max tenant name length
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "maximum-length-name", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	for _, element := range resources.ClusterRoles {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.ClusterRoleBindings {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.ConfigMaps {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.CronJobs {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 52)
	}
	for _, element := range resources.DaemonSets {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Deployments {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Ingresses {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Jobs {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 63)
	}
	for _, element := range resources.PersistentVolumeClaims {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Roles {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.RoleBindings {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Secrets {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Services {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 63)
	}
	for _, element := range resources.ServiceAccounts {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Statefulsets {
		for _, label := range element.Labels {
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
}
