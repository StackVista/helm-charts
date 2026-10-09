package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// Pre-change source: 7979fb76f8d9edd5f27d119265bf2f947fe38dba. Independent extraction baseline:
// c851a02ae1f2f7eb64c5538a148194ea36eee4b4.
func TestInternalRoleNamingCompatibility(t *testing.T) {
	before := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(before, os.DirFS("..")))
	restoreInternalRoleNames(t, before)
	checkInternalRoleNaming(t, before, "..")
}

func TestInternalRoleNamingGitBaselineCompatibility(t *testing.T) {
	before := os.Getenv("INTERNAL_ROLE_BASELINE_CHART")
	if before == "" {
		t.Skip("set INTERNAL_ROLE_BASELINE_CHART to the built pre-change chart")
	}
	checkInternalRoleNaming(t, before, "..")
}

func TestInternalRoleNamingIndependentOfPreviousSteps(t *testing.T) {
	before := os.Getenv("INTERNAL_ROLE_EXTRACTION_CHART")
	if before == "" {
		t.Skip("set INTERNAL_ROLE_EXTRACTION_CHART to the built extraction baseline")
	}
	after := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(after, os.DirFS(before)))
	current, err := os.ReadFile("../templates/_names.tpl")
	require.NoError(t, err)
	path := filepath.Join(after, "templates/_names.tpl")
	original, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(original)
	for helper, legacy := range internalRoleLegacyDefinitions() {
		pattern := internalRoleNamingDefinition(helper)
		definitions := pattern.FindAllString(string(current), -1)
		require.Len(t, definitions, 1, helper)
		if legacy == "" {
			require.Empty(t, pattern.FindAllString(content, -1))
			content += "\n" + definitions[0] + "\n"
		} else {
			require.Len(t, pattern.FindAllString(content, -1), 1)
			content = pattern.ReplaceAllString(content, definitions[0])
		}
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	checkInternalRoleNaming(t, before, after)
}

func internalRoleLegacyDefinitions() map[string]string {
	return map[string]string{
		"stackstate.router.mode.role.fullname":        `{{ template "common.fullname.short" . }}-router-mode`,
		"stackstate.router.mode.rolebinding.fullname": `{{ template "common.fullname.short" . }}-router-mode`,
		"kafkaup-operator.role.fullname":              "",
		"kafkaup-operator.rolebinding.fullname":       "",
	}
}

func internalRoleNamingDefinition(helper string) *regexp.Regexp {
	return regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
}

func restoreInternalRoleNames(t *testing.T, chart string) {
	t.Helper()
	path := filepath.Join(chart, "templates/_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	for helper, legacy := range internalRoleLegacyDefinitions() {
		pattern := internalRoleNamingDefinition(helper)
		require.Len(t, pattern.FindAllString(content, -1), 1, helper)
		definition := ""
		if legacy != "" {
			definition = `{{- define "` + helper + `" -}}` + legacy + `{{- end -}}`
		}
		content = pattern.ReplaceAllString(content, definition)
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}

func checkInternalRoleNaming(t *testing.T, beforeChart, afterChart string) {
	t.Helper()
	for _, tc := range []struct {
		name, release, namespace, sizing string
		values                           map[string]string
	}{
		{name: "default", release: "suse-observability"},
		{name: "nightly"},
		{name: "second-namespace", namespace: "tenant-b"},
		{name: "root-fullname", values: map[string]string{"fullnameOverride": "customer-platform"}},
		{name: "global-fullname", values: map[string]string{"global.fullnameOverride": "customer-global"}},
		{name: "affixes", values: map[string]string{
			"fullnamePrefix": "team-", "fullnameSuffix": "-local",
			"global.fullnamePrefix": "global-", "global.fullnameSuffix": "-end",
		}},
		{name: "operator-fullname-and-name", values: map[string]string{
			"kafkaup-operator.fullnameOverride": "customer-operator", "kafkaup-operator.nameOverride": "ignored-name",
		}},
		{name: "operator-name-and-affixes", values: map[string]string{
			"kafkaup-operator.nameOverride":   "custom-operator",
			"kafkaup-operator.fullnamePrefix": "team-", "kafkaup-operator.fullnameSuffix": "-operator",
		}},
		{name: "long-fullnames", values: map[string]string{
			"fullnameOverride": strings.Repeat("p", 70), "kafkaup-operator.fullnameOverride": strings.Repeat("o", 70),
		}},
		{name: "argo", values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
		{name: "ha", sizing: "global_sizing_150_ha.yaml"},
		{name: "nonha-monolithic", sizing: "global_sizing_50_nonha.yaml", values: map[string]string{
			"stackstate.features.server.split": "false", "hbase.deployment.mode": "Mono",
		}},
		{name: "external-kafka", values: map[string]string{
			"kafkaup-operator.kafkaSelectors.statefulSetName": "customer-{{ .Release.Name }}-kafka",
			"kafkaup-operator.kafkaSelectors.podLabel.key":    "customer", "kafkaup-operator.kafkaSelectors.podLabel.value": "external-kafka",
		}},
		{name: "operator-disabled", values: map[string]string{"kafkaup-operator.enabled": "false"}},
		{name: "active-router", values: map[string]string{"stackstate.components.router.mode.status": "active"}},
		{name: "maintenance-router-and-no-authorization", values: map[string]string{
			"stackstate.components.router.mode.status": "maintenance", "stackstate.k8sAuthorization.enabled": "false",
		}},
	} {
		if tc.release == "" {
			tc.release = "nightly"
		}
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				values := componentResourceNameTestValues()
				values["stackstate.components.router.mode.status"] = "automatic"
				values["kafkaup-operator.enabled"] = "true"
				values["stackstate.k8sAuthorization.enabled"] = "true"
				for key, value := range tc.values {
					values[key] = value
				}
				options := apiResourceNameTestOptions(values)
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				if tc.sizing != "" {
					options.ValuesFiles = append(options.ValuesFiles, filepath.Join("values", tc.sizing))
				}
				beforeOutput := renderInternalRoleChart(t, beforeChart, tc.release, options, upgrade)
				afterOutput := renderInternalRoleChart(t, afterChart, tc.release, options, upgrade)
				original := helmtestutil.NewKubernetesResources(t, beforeOutput)
				target := helmtestutil.NewKubernetesResources(t, afterOutput)
				operatorAccount := ""
				for _, deployment := range original.Deployments {
					for _, container := range deployment.Spec.Template.Spec.Containers {
						if container.Name == "kafkaup-operator" {
							operatorAccount = deployment.Spec.Template.Spec.ServiceAccountName
						}
					}
				}
				renames := map[string]string{}
				refs := map[string]string{}
				for _, pair := range []struct{ component, name string }{
					{"router-mode", "suse-observability-router-mode"},
					{"kafkaup-operator", "suse-observability-kafkaup"},
				} {
					var old rbacv1.RoleBinding
					found := false
					for _, binding := range original.RoleBindings {
						matches := binding.Labels["app.kubernetes.io/component"] == pair.component
						if pair.component == "kafkaup-operator" && operatorAccount != "" {
							for _, subject := range binding.Subjects {
								matches = matches || (subject.Kind == "ServiceAccount" && subject.Name == operatorAccount)
							}
						}
						if matches {
							require.False(t, found, "Each selected family must have exactly one binding")
							old, found = binding, true
						}
					}
					bindingName := pair.name
					if pair.component == "kafkaup-operator" {
						bindingName += "-binding"
						require.Equal(t, values["kafkaup-operator.enabled"] == "true", found)
					} else {
						require.Equal(t, values["stackstate.components.router.mode.status"] == "automatic", found)
					}
					if !found {
						assert.NotContains(t, target.Roles, pair.name)
						assert.NotContains(t, target.RoleBindings, bindingName)
						continue
					}
					require.Contains(t, original.Roles, old.RoleRef.Name)
					require.Contains(t, target.Roles, pair.name)
					require.Contains(t, target.RoleBindings, bindingName)
					binding := target.RoleBindings[bindingName]
					expected := *old.DeepCopy()
					expected.Name, expected.RoleRef.Name = bindingName, pair.name
					assert.Equal(t, expected, binding, "Subjects, permissions, namespaces, labels and hook ordering must stay unchanged")
					role := original.Roles[old.RoleRef.Name]
					role.Name = pair.name
					assert.Equal(t, role, target.Roles[pair.name])
					for _, subject := range binding.Subjects {
						require.Equal(t, "ServiceAccount", subject.Kind)
						require.Contains(t, target.ServiceAccounts, subject.Name, "Each binding must resolve its retained account")
					}
					if old.RoleRef.Name != pair.name {
						renames["Role/"+old.RoleRef.Name] = pair.name
						assert.NotContains(t, target.Roles, old.RoleRef.Name)
					}
					if old.Name != bindingName {
						renames["RoleBinding/"+old.Name] = bindingName
						assert.NotContains(t, target.RoleBindings, old.Name)
					}
					refs[bindingName] = pair.name
				}
				seen := map[string]bool{}
				expected := backupNamingDocuments(t, beforeOutput, renames, seen)
				for name, role := range refs {
					for _, namespace := range []string{"", options.KubectlOptions.Namespace} {
						if documents, ok := expected["RoleBinding/"+namespace+"/"+name]; ok {
							for _, raw := range documents.([]interface{}) {
								raw.(map[string]interface{})["roleRef"].(map[string]interface{})["name"] = role
							}
						}
					}
				}
				require.Len(t, seen, len(renames))
				assert.Equal(t, expected, backupNamingDocuments(t, afterOutput, nil, nil),
					"Only these internal Role/binding names and exact roleRefs may change; no other rename step supplies consumers")
			})
		}
	}
}

func renderInternalRoleChart(t *testing.T, chart, release string, options *helm.Options, upgrade bool) string {
	t.Helper()
	copied := *options
	copied.Logger = logger.Discard
	copied.ValuesFiles = make([]string, len(options.ValuesFiles))
	for i, path := range options.ValuesFiles {
		file, err := filepath.Abs(path)
		require.NoError(t, err)
		copied.ValuesFiles[i] = file
	}
	args := []string{"--kube-version", "1.32.0", "--api-versions", "policy/v1", "--api-versions", "monitoring.coreos.com/v1"}
	if upgrade {
		args = append(args, "--is-upgrade")
	}
	output, err := helm.RenderTemplateE(t, &copied, chart, release, nil, args...)
	require.NoError(t, err)
	return output
}

func TestInternalRoleNamingSubsequentUpgradePatches(t *testing.T) {
	options := apiResourceNameTestOptions(map[string]string{
		"stackstate.components.router.mode.status": "automatic",
		"kafkaup-operator.enabled":                 "true", "fullnameOverride": "first-platform",
		"kafkaup-operator.fullnameOverride": "first-operator",
	})
	old := helmtestutil.NewKubernetesResources(t, renderInternalRoleChart(t, "..", "nightly", options, true))
	options.SetValues["fullnameOverride"] = "second-platform"
	options.SetValues["kafkaup-operator.fullnameOverride"] = "second-operator"
	new := helmtestutil.NewKubernetesResources(t, renderInternalRoleChart(t, "..", "nightly", options, true))
	for _, name := range []string{"suse-observability-router-mode", "suse-observability-kafkaup-binding"} {
		require.Contains(t, old.RoleBindings, name)
		require.Contains(t, new.RoleBindings, name)
		for _, rollback := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/rollback=%t", name, rollback), func(t *testing.T) {
				before, after := old.RoleBindings[name], new.RoleBindings[name]
				if rollback {
					before, after = after, before
				}
				require.Equal(t, before.RoleRef, after.RoleRef, "Canonical bindings must never patch immutable roleRef")
				original, err := json.Marshal(before)
				require.NoError(t, err)
				modified, err := json.Marshal(after)
				require.NoError(t, err)
				before.ResourceVersion = "123"
				current, err := json.Marshal(before)
				require.NoError(t, err)
				schema, err := strategicpatch.NewPatchMetaFromStruct(rbacv1.RoleBinding{})
				require.NoError(t, err)
				patch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, current, schema, true)
				require.NoError(t, err)
				var fields map[string]interface{}
				require.NoError(t, json.Unmarshal(patch, &fields))
				assert.NotContains(t, fields, "roleRef")
				applied, err := strategicpatch.StrategicMergePatch(current, patch, rbacv1.RoleBinding{})
				require.NoError(t, err)
				var got rbacv1.RoleBinding
				require.NoError(t, json.Unmarshal(applied, &got))
				after.ResourceVersion = "123"
				assert.Equal(t, after, got, "Mutable account subjects follow naming changes in upgrade and rollback patches")
			})
		}
	}
}
