package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestCoreServiceAccountReferencesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	components := map[string]string{
		"api": "api", "server": "server", "checks": "checks",
		"notification": "notification", "healthSync": "health-sync",
		"authorizationSync": "authorization-sync", "initializer": "initializer",
		"slicing": "slicing", "state": "state", "sync": "sync",
	}
	for component, suffix := range components {
		helper := "stackstate." + component + ".serviceaccount.fullname"
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAllString(content, -1), 1, helper)
		content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}explicit-`+suffix+`-account{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, serverSplit := range []bool{false, true} {
		for _, clusterRBAC := range []bool{false, true} {
			for _, authorization := range []bool{false, true} {
				t.Run(fmt.Sprintf("server=%t/cluster-rbac=%t/authorization=%t", serverSplit, clusterRBAC, authorization), func(t *testing.T) {
					options := apiResourceNameTestOptions(map[string]string{
						"stackstate.features.server.split":    fmt.Sprint(serverSplit),
						"cluster-role.enabled":                fmt.Sprint(clusterRBAC),
						"stackstate.k8sAuthorization.enabled": fmt.Sprint(authorization),
					})
					before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
					options.ValuesFiles = []string{valuesFile}
					output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
					require.NoError(t, err)
					after := helmtestutil.NewKubernetesResources(t, output)

					prefix := "nightly-suse-observability-"
					bindingName := prefix + "get-pods"
					require.Contains(t, before.RoleBindings, bindingName)
					require.Contains(t, after.RoleBindings, bindingName)
					expectedBinding := before.RoleBindings[bindingName]
					for component, suffix := range components {
						legacy := prefix + suffix
						explicit := "explicit-" + suffix + "-account"
						inServerMode := serverSplit != (component == "server")
						if inServerMode {
							// The existing split-mode binding retains the authorization-sync
							// subject even when that component is disabled.
							expectedBinding.Subjects = renamedServiceAccountSubject(t, expectedBinding.Subjects, legacy, explicit)
						}
						enabled := inServerMode && (component != "authorizationSync" || authorization)
						if !enabled {
							assert.NotContains(t, before.ServiceAccounts, legacy)
							assert.NotContains(t, after.ServiceAccounts, explicit)
							assert.NotContains(t, before.Deployments, legacy)
							assert.NotContains(t, after.Deployments, legacy)
							continue
						}
						require.Contains(t, before.ServiceAccounts, legacy)
						require.Contains(t, after.ServiceAccounts, explicit)
						assert.NotContains(t, after.ServiceAccounts, legacy)
						expectedAccount := before.ServiceAccounts[legacy]
						expectedAccount.Name = explicit
						assert.Equal(t, expectedAccount, after.ServiceAccounts[explicit])
						delete(before.ServiceAccounts, legacy)
						delete(after.ServiceAccounts, explicit)

						require.Contains(t, before.Deployments, legacy)
						require.Contains(t, after.Deployments, legacy)
						expectedDeployment := before.Deployments[legacy]
						assert.Equal(t, legacy, expectedDeployment.Spec.Template.Spec.ServiceAccountName)
						expectedDeployment.Spec.Template.Spec.ServiceAccountName = explicit
						assert.Equal(t, expectedDeployment, after.Deployments[legacy])
						delete(before.Deployments, legacy)
						delete(after.Deployments, legacy)
					}
					assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
					assert.Equal(t, before.Deployments, after.Deployments)
					assert.Equal(t, expectedBinding, after.RoleBindings[bindingName])
					delete(before.RoleBindings, bindingName)
					delete(after.RoleBindings, bindingName)
					assert.Equal(t, before.RoleBindings, after.RoleBindings)

					apiComponent := "server"
					if serverSplit {
						apiComponent = "api"
					}
					for _, purpose := range []string{"authentication", "authorization"} {
						name := "observability-" + prefix + purpose
						if !clusterRBAC {
							assert.NotContains(t, before.ClusterRoleBindings, name)
							assert.NotContains(t, after.ClusterRoleBindings, name)
							continue
						}
						require.Contains(t, before.ClusterRoleBindings, name)
						require.Contains(t, after.ClusterRoleBindings, name)
						expected := before.ClusterRoleBindings[name]
						expected.Subjects = renamedServiceAccountSubject(t, expected.Subjects, prefix+apiComponent, "explicit-"+apiComponent+"-account")
						assert.Equal(t, expected, after.ClusterRoleBindings[name], "Binding identity, namespace and roleRef must stay unchanged")
						delete(before.ClusterRoleBindings, name)
						delete(after.ClusterRoleBindings, name)
					}
					assert.Equal(t, before.ClusterRoleBindings, after.ClusterRoleBindings)
					assert.Equal(t, before.Roles, after.Roles)
					assert.Equal(t, before.ClusterRoles, after.ClusterRoles)
					assert.Equal(t, before.Services, after.Services)
					assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
					assert.Equal(t, before.Secrets, after.Secrets)
					assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
					assert.Equal(t, before.Statefulsets, after.Statefulsets)
				})
			}
		}
	}
}

func renamedServiceAccountSubject(t *testing.T, subjects []rbacv1.Subject, oldName, newName string) []rbacv1.Subject {
	t.Helper()
	result := slices.Clone(subjects)
	matches := 0
	for i := range result {
		if result[i].Kind == "ServiceAccount" && result[i].Name == oldName {
			result[i].Name = newName
			matches++
		}
	}
	require.Equal(t, 1, matches, "Expected exactly one subject referencing %s", oldName)
	return result
}
