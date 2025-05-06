package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestAuthProviderInGroupName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/role-k8s-authz.yaml")
  resources := helmtestutil.NewKubernetesResources(t, output)

  for _, binding := range resources.RoleBindings {
    if strings.Contains(binding.Name, "-instance-") {
      for _, subject := range binding.Subjects {
        require.Contains(t, subject.Name, "keycloak")
        require.Contains(t, subject.Name, binding.Name)
      }
    }
  }
}

func TestNoAuthProviderInDefaultGroupName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
  resources := helmtestutil.NewKubernetesResources(t, output)

  for _, binding := range resources.RoleBindings {
    if strings.Contains(binding.Name, "-instance-") {
      for _, subject := range binding.Subjects {
        require.Equal(t, binding.Name, subject.Name)
      }
    }
  }
}
