package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestInternalRBACNamesUseDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	// Role and binding names deliberately differ to catch accidental coupling.
	for helper, name := range map[string]string{
		"stackstate.getPods.role.fullname":                      "explicit-pod-reader",
		"stackstate.getPods.rolebinding.fullname":               "explicit-pod-reader-binding",
		"stackstate.authorization.clusterrole.fullname":         "explicit-authorization-role",
		"stackstate.authorization.clusterrolebinding.fullname":  "explicit-authorization-binding",
		"stackstate.authentication.clusterrolebinding.fullname": "explicit-authentication-binding",
	} {
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAllString(content, -1), 1, helper)
		content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}`+name+`{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, serverSplit := range []bool{false, true} {
		for _, clusterRBAC := range []bool{false, true} {
			for _, namespace := range []string{"observability", "nightly-suse-observability"} {
				t.Run(fmt.Sprintf("server=%t/cluster-rbac=%t/namespace=%s", serverSplit, clusterRBAC, namespace), func(t *testing.T) {
					options := apiResourceNameTestOptions(map[string]string{
						"stackstate.features.server.split": fmt.Sprint(serverSplit),
						"cluster-role.enabled":             fmt.Sprint(clusterRBAC),
					})
					options.KubectlOptions.Namespace = namespace
					before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
					options.ValuesFiles = []string{valuesFile}
					output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
					require.NoError(t, err)
					after := helmtestutil.NewKubernetesResources(t, output)

					prefix := "nightly-suse-observability-"
					podReader := prefix + "get-pods"
					require.Contains(t, before.Roles, podReader)
					expectedRole := before.Roles[podReader]
					expectedRole.Name = "explicit-pod-reader"
					delete(before.Roles, podReader)
					before.Roles[expectedRole.Name] = expectedRole
					assert.Equal(t, before.Roles, after.Roles)

					require.Contains(t, before.RoleBindings, podReader)
					expectedBinding := before.RoleBindings[podReader]
					assert.Equal(t, podReader, expectedBinding.RoleRef.Name)
					expectedBinding.Name = "explicit-pod-reader-binding"
					expectedBinding.RoleRef.Name = "explicit-pod-reader"
					delete(before.RoleBindings, podReader)
					before.RoleBindings[expectedBinding.Name] = expectedBinding
					assert.Equal(t, before.RoleBindings, after.RoleBindings)

					clusterPrefix := prefix
					if namespace != strings.TrimSuffix(prefix, "-") {
						clusterPrefix = namespace + "-" + prefix
					}
					authorizationRole := clusterPrefix + "authorization"
					if clusterRBAC {
						require.Contains(t, before.ClusterRoles, authorizationRole)
						expected := before.ClusterRoles[authorizationRole]
						expected.Name = "explicit-authorization-role"
						delete(before.ClusterRoles, authorizationRole)
						before.ClusterRoles[expected.Name] = expected
					} else {
						assert.NotContains(t, before.ClusterRoles, authorizationRole)
						assert.NotContains(t, after.ClusterRoles, "explicit-authorization-role")
					}
					assert.Equal(t, before.ClusterRoles, after.ClusterRoles)

					for _, purpose := range []string{"authentication", "authorization"} {
						legacy := clusterPrefix + purpose
						explicit := "explicit-" + purpose + "-binding"
						if !clusterRBAC {
							assert.NotContains(t, before.ClusterRoleBindings, legacy)
							assert.NotContains(t, after.ClusterRoleBindings, explicit)
							continue
						}
						require.Contains(t, before.ClusterRoleBindings, legacy)
						expected := before.ClusterRoleBindings[legacy]
						expected.Name = explicit
						if purpose == "authorization" {
							assert.Equal(t, authorizationRole, expected.RoleRef.Name)
							expected.RoleRef.Name = "explicit-authorization-role"
						} else {
							assert.Equal(t, "system:auth-delegator", expected.RoleRef.Name)
						}
						delete(before.ClusterRoleBindings, legacy)
						before.ClusterRoleBindings[explicit] = expected
					}
					// Whole-resource comparisons preserve rules, subjects, namespaces
					// and references to Kubernetes-owned roles.
					assert.Equal(t, before.ClusterRoleBindings, after.ClusterRoleBindings)
					assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
					assert.Equal(t, before.Deployments, after.Deployments)
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

func TestRBACAgentNamesPreserveApplicationSubjects(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	for helper, name := range map[string]string{
		"stackstate.rbacAgent.role.fullname":        "explicit-agent-role",
		"stackstate.rbacAgent.rolebinding.fullname": "explicit-agent-binding",
	} {
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAllString(content, -1), 1, helper)
		content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}`+name+`{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, serverSplit := range []bool{false, true} {
		for _, authorization := range []bool{false, true} {
			for _, mode := range []string{"SelfHosted", "Saas"} {
				t.Run(fmt.Sprintf("server=%t/authorization=%t/mode=%s", serverSplit, authorization, mode), func(t *testing.T) {
					options := apiResourceNameTestOptions(map[string]string{
						"stackstate.features.server.split":    fmt.Sprint(serverSplit),
						"stackstate.k8sAuthorization.enabled": fmt.Sprint(authorization),
						"stackstate.deployment.mode":          mode,
					})
					before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
					options.ValuesFiles = []string{valuesFile}
					output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
					require.NoError(t, err)
					after := helmtestutil.NewKubernetesResources(t, output)
					legacy := "nightly-suse-observability-rbac-agent"
					if authorization {
						require.Contains(t, before.Roles, legacy)
						expectedRole := before.Roles[legacy]
						expectedRole.Name = "explicit-agent-role"
						delete(before.Roles, legacy)
						before.Roles[expectedRole.Name] = expectedRole
						require.Contains(t, before.RoleBindings, legacy)
						expectedBinding := before.RoleBindings[legacy]
						assert.Equal(t, legacy, expectedBinding.RoleRef.Name)
						expectedBinding.Name = "explicit-agent-binding"
						expectedBinding.RoleRef.Name = "explicit-agent-role"
						delete(before.RoleBindings, legacy)
						before.RoleBindings[expectedBinding.Name] = expectedBinding
					} else {
						assert.NotContains(t, before.Roles, legacy)
						assert.NotContains(t, after.Roles, "explicit-agent-role")
						assert.NotContains(t, before.RoleBindings, legacy)
						assert.NotContains(t, after.RoleBindings, "explicit-agent-binding")
					}
					assert.Equal(t, before.Roles, after.Roles)
					assert.Equal(t, before.RoleBindings, after.RoleBindings)

					component := "server"
					if serverSplit {
						component = "api"
					}
					configName := "nightly-suse-observability-" + component
					require.Contains(t, after.ConfigMaps, configName)
					config := after.ConfigMaps[configName].Data["application_stackstate.conf"]
					subject := "stackstate.authorization.staticSubjects."
					if mode == "Saas" {
						subject = "stackstate.api.authorization.staticSubjects."
					}
					if authorization {
						assert.Contains(t, config, subject+legacy+`: { systemPermissions: ["update-permissions"] }`)
					} else {
						assert.NotContains(t, config, subject+legacy)
					}
					assert.NotContains(t, config, "explicit-agent-role")
					assert.NotContains(t, config, "explicit-agent-binding")
					assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
					assert.Equal(t, before.Deployments, after.Deployments)
					assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
					assert.Equal(t, before.ClusterRoles, after.ClusterRoles)
					assert.Equal(t, before.ClusterRoleBindings, after.ClusterRoleBindings)
					assert.Equal(t, before.Services, after.Services)
					assert.Equal(t, before.Secrets, after.Secrets)
					assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
					assert.Equal(t, before.Statefulsets, after.Statefulsets)
				})
			}
		}
	}
}

func TestInstanceRBACNamesPreserveExternalGroups(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	roles := map[string]string{
		"admin": "admin", "observer": "observer", "troubleshooter": "troubleshooter",
		"basicAccess": "basic-access", "recommendedAccess": "recommended-access",
	}
	for role, suffix := range roles {
		for _, kind := range []string{"role", "rolebinding"} {
			helper := "stackstate.k8s.authorization.instance." + role + "." + kind + ".fullname"
			definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
			require.Len(t, definition.FindAllString(content, -1), 1, helper)
			content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}explicit-`+suffix+`-`+kind+`{{- end -}}`)
		}
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, serverSplit := range []bool{false, true} {
		for _, mode := range []string{"SelfHosted", "Saas"} {
			for _, instanceRoles := range []bool{false, true} {
				for _, authorization := range []bool{false, true} {
					t.Run(fmt.Sprintf("server=%t/mode=%s/instance-roles=%t/authorization=%t", serverSplit, mode, instanceRoles, authorization), func(t *testing.T) {
						options := apiResourceNameTestOptions(map[string]string{
							"stackstate.features.server.split":    fmt.Sprint(serverSplit),
							"stackstate.deployment.mode":          mode,
							"stackstate.features.role-k8s-authz":  fmt.Sprint(instanceRoles),
							"stackstate.k8sAuthorization.enabled": fmt.Sprint(authorization),
						})
						before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
						options.ValuesFiles = []string{valuesFile}
						output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
						require.NoError(t, err)
						after := helmtestutil.NewKubernetesResources(t, output)
						for role, suffix := range roles {
							legacy := "nightly-suse-observability-instance-" + suffix
							roleName := "explicit-" + suffix + "-role"
							bindingName := "explicit-" + suffix + "-rolebinding"
							enabled := instanceRoles
							if role == "basicAccess" {
								enabled = authorization
							}
							if !enabled {
								assert.NotContains(t, before.Roles, legacy)
								assert.NotContains(t, before.RoleBindings, legacy)
								assert.NotContains(t, after.Roles, roleName)
								assert.NotContains(t, after.RoleBindings, bindingName)
								continue
							}
							require.Contains(t, before.Roles, legacy)
							expectedRole := before.Roles[legacy]
							expectedRole.Name = roleName
							delete(before.Roles, legacy)
							before.Roles[roleName] = expectedRole

							require.Contains(t, before.RoleBindings, legacy)
							expectedBinding := before.RoleBindings[legacy]
							assert.Equal(t, legacy, expectedBinding.RoleRef.Name)
							expectedBinding.Name = bindingName
							expectedBinding.RoleRef.Name = roleName
							delete(before.RoleBindings, legacy)
							before.RoleBindings[bindingName] = expectedBinding
							require.Contains(t, after.RoleBindings, bindingName)
							subjects := after.RoleBindings[bindingName].Subjects
							require.Len(t, subjects, 1)
							assert.Equal(t, "Group", subjects[0].Kind)
							group := legacy
							if role == "basicAccess" || role == "recommendedAccess" {
								group = "system:authenticated"
							}
							assert.Equal(t, group, subjects[0].Name, "Group identity must not follow resource-name helpers")
						}
						assert.Equal(t, before.Roles, after.Roles)
						assert.Equal(t, before.RoleBindings, after.RoleBindings)
						assert.Equal(t, before.ClusterRoles, after.ClusterRoles)
						assert.Equal(t, before.ClusterRoleBindings, after.ClusterRoleBindings)
						assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
						assert.Equal(t, before.Deployments, after.Deployments)
						assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
						assert.Equal(t, before.Services, after.Services)
						assert.Equal(t, before.Secrets, after.Secrets)
						assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
						assert.Equal(t, before.Statefulsets, after.Statefulsets)
					})
				}
			}
		}
	}
}
