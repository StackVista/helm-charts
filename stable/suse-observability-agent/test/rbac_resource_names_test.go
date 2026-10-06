package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestRbacAgentServiceAccountReferencesFollowDedicatedHelper(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	require.NoError(t, os.WriteFile(filepath.Join(chart, "templates", "_rbac-resource-names.tpl"), []byte(
		`{{- define "kubernetes-rbac-agent.serviceaccount.fullname" -}}explicit-rbac-account{{- end -}}`), 0600))
	values, err := filepath.Abs("values/minimal.yaml")
	require.NoError(t, err)
	for _, release := range []string{"suse-observability-agent", "nightly"} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", release, upgrade), func(t *testing.T) {
				options := &helm.Options{
					ValuesFiles: []string{values},
					SetValues: map[string]string{
						"stackstate.cluster.authToken": "test-cluster-token",
					},
					KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
				}
				var args []string
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				before := helmtestutil.NewKubernetesResources(t,
					helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, release, options, args...))
				output, err := helm.RenderTemplateE(t, options, chart, release, nil, args...)
				require.NoError(t, err)
				after := helmtestutil.NewKubernetesResources(t, output)
				accountName := release + "-rbac-agent"
				require.Contains(t, before.ServiceAccounts, accountName)
				account := before.ServiceAccounts[accountName]
				delete(before.ServiceAccounts, accountName)
				account.Name = "explicit-rbac-account"
				before.ServiceAccounts[account.Name] = account
				for name, deployment := range before.Deployments {
					if deployment.Spec.Template.Spec.ServiceAccountName == accountName {
						deployment.Spec.Template.Spec.ServiceAccountName = account.Name
						before.Deployments[name] = deployment
					}
				}
				for name, binding := range before.ClusterRoleBindings {
					for i := range binding.Subjects {
						if binding.Subjects[i].Kind == "ServiceAccount" && binding.Subjects[i].Name == accountName {
							binding.Subjects[i].Name = account.Name
						}
					}
					before.ClusterRoleBindings[name] = binding
				}
				assert.Equal(t, before, after)
			})
		}
	}
}
