package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
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

func TestResourcesNamesLength(t *testing.T) {
	// 23 chars is the max release name length
	// 12 chars is the max tenant name length
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "stackstate-abcd1234abcd", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	for _, element := range resources.ClusterRoles{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.ClusterRoleBindings{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.ConfigMaps{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.CronJobs{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 52)
	}
	for _, element := range resources.DaemonSets{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Deployments{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Ingresses{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Jobs{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 63)
	}
	for _, element := range resources.PersistentVolumeClaims{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Roles{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.RoleBindings{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Secrets{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Services{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 63)
	}
	for _, element := range resources.ServiceAccounts{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
	for _, element := range resources.Statefulsets{
		for _, label := range element.Labels{
			assert.LessOrEqual(t, len(label), 63)
		}
		assert.LessOrEqual(t, len(element.Name), 253)
	}
}
