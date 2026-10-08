package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/validation"
)

type clusterBindingNamingFixture struct {
	Role     rbacv1.ClusterRole                   `json:"role"`
	Bindings map[string]rbacv1.ClusterRoleBinding `json:"bindings"`
}

// Complete objects captured before renaming from
// 91d4b4b016a40d233d05f6285d2dc932fb094bd9. Only the two binding names change.
func TestClusterBindingNamingFrozenCompatibility(t *testing.T) {
	checkClusterBindingNamingFixtures(t, "..", true)
}

func TestClusterBindingNamingBaselineReproduction(t *testing.T) {
	chart := os.Getenv("CLUSTER_BINDING_BASELINE_CHART")
	if chart == "" {
		t.Skip("set CLUSTER_BINDING_BASELINE_CHART to the pre-rename chart with built dependencies")
	}
	checkClusterBindingNamingFixtures(t, chart, false)
}

func checkClusterBindingNamingFixtures(t *testing.T, chart string, renamed bool) {
	t.Helper()
	for _, mode := range []string{"split", "monolithic"} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", mode, upgrade), func(t *testing.T) {
				data, err := os.ReadFile(filepath.Join("testdata/cluster-binding-renaming", mode+".json"))
				require.NoError(t, err)
				var expected clusterBindingNamingFixture
				require.NoError(t, json.Unmarshal(data, &expected))
				options := apiResourceNameTestOptions(map[string]string{"stackstate.features.server.split": fmt.Sprint(mode == "split")})
				output := renderRBACNamingChart(t, chart, "nightly", options, upgrade)
				resources := helmtestutil.NewKubernetesResources(t, output)
				require.Contains(t, resources.ClusterRoles, expected.Role.Name)
				actual := clusterBindingNamingFixture{Role: resources.ClusterRoles[expected.Role.Name], Bindings: map[string]rbacv1.ClusterRoleBinding{}}
				for _, purpose := range []string{"authentication", "authorization"} {
					binding := expected.Bindings[purpose]
					if renamed {
						binding.Name = "suse-observability-observability-" + purpose
						expected.Bindings[purpose] = binding
					}
					require.Contains(t, resources.ClusterRoleBindings, binding.Name)
					actual.Bindings[purpose] = resources.ClusterRoleBindings[binding.Name]
				}
				for _, labels := range []map[string]string{actual.Role.Labels, actual.Bindings["authentication"].Labels, actual.Bindings["authorization"].Labels} {
					delete(labels, "helm.sh/chart")
					delete(labels, "app.kubernetes.io/version")
				}
				assert.Equal(t, expected, actual)
			})
		}
	}
}

func TestClusterBindingNamingPreservesPlatformResources(t *testing.T) {
	beforeChart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(beforeChart, os.DirFS("..")))
	path := filepath.Join(beforeChart, "templates/_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	for _, purpose := range []string{"authentication", "authorization"} {
		header := `{{- define "stackstate.` + purpose + `.clusterrolebinding.fullname" -}}`
		pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(header) + `.*?\{\{- end -\}\}`)
		require.Len(t, pattern.FindAllString(content, -1), 1)
		content = pattern.ReplaceAllString(content, header+`{{ template "common.fullname.cluster.unique" . }}-`+purpose+`{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	for _, tc := range []struct {
		name, release, namespace, legacyPrefix, targetPrefix, sizing string
		disabled                                                     bool
		values                                                       map[string]string
	}{
		{name: "product-namespace", release: "suse-observability", namespace: "suse-observability", legacyPrefix: "suse-observability", targetPrefix: "suse-observability"},
		{name: "nightly-product-namespace", release: "nightly", namespace: "suse-observability", legacyPrefix: "suse-observability-nightly-suse-observability", targetPrefix: "suse-observability"},
		{name: "product-release", release: "suse-observability", legacyPrefix: "observability-suse-observability", targetPrefix: "suse-observability-observability"},
		{name: "nightly", release: "nightly", legacyPrefix: "observability-nightly-suse-observability", targetPrefix: "suse-observability-observability"},
		{name: "same-release-other-namespace", release: "suse-observability", namespace: "tenant-b", legacyPrefix: "tenant-b-suse-observability", targetPrefix: "suse-observability-tenant-b"},
		{name: "namespace-matches-legacy-base", release: "nightly", namespace: "nightly-suse-observability", legacyPrefix: "nightly-suse-observability", targetPrefix: "suse-observability-nightly-suse-observability"},
		{name: "fullname", release: "nightly", legacyPrefix: "customer", targetPrefix: "suse-observability-observability", values: map[string]string{"fullnameOverride": "CUSTOMER"}},
		{name: "affixes", release: "nightly", legacyPrefix: "g-pre-customer-post-end", targetPrefix: "suse-observability-observability", values: map[string]string{
			"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "global-override", release: "nightly", legacyPrefix: "observability-nightly-suse-observability", targetPrefix: "suse-observability-observability", values: map[string]string{"global.fullnameOverride": "customer"}},
		{name: "long-release", release: strings.Repeat("x", 50), legacyPrefix: "observability-" + strings.Repeat("x", 40), targetPrefix: "suse-observability-observability"},
		{name: "long-override", release: "nightly", legacyPrefix: strings.Repeat("a", 54), targetPrefix: "suse-observability-observability", values: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "long-namespace", release: "nightly", namespace: strings.Repeat("n", 63), legacyPrefix: strings.Repeat("n", 54), targetPrefix: "suse-observability-" + strings.Repeat("n", 63)},
		{name: "already-canonical", release: "nightly", legacyPrefix: "suse-observability-observability", targetPrefix: "suse-observability-observability", values: map[string]string{"fullnameOverride": "suse-observability-observability"}},
		{name: "monolithic", release: "nightly", legacyPrefix: "observability-nightly-suse-observability", targetPrefix: "suse-observability-observability", values: map[string]string{"stackstate.features.server.split": "false"}},
		{name: "ha", release: "nightly", legacyPrefix: "observability-nightly-suse-observability", targetPrefix: "suse-observability-observability", sizing: "global_sizing_150_ha.yaml"},
		{name: "nonha", release: "nightly", legacyPrefix: "observability-nightly-suse-observability", targetPrefix: "suse-observability-observability", sizing: "global_sizing_50_nonha.yaml"},
		{name: "argo", release: "nightly", legacyPrefix: "observability-nightly-suse-observability", targetPrefix: "suse-observability-observability", values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
		{name: "disabled", release: "nightly", legacyPrefix: "observability-nightly-suse-observability", targetPrefix: "suse-observability-observability", disabled: true, values: map[string]string{"cluster-role.enabled": "false"}},
		{name: "custom-metadata", release: "nightly", legacyPrefix: "observability-nightly-suse-observability", targetPrefix: "suse-observability-observability", values: map[string]string{
			"clusterrolebinding.annotations.fixture-note": "preserved", "commonLabels.fixture-scope": "preserved",
		}},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				options := apiResourceNameTestOptions(componentResourceNameTestValues())
				for key, value := range tc.values {
					options.SetValues[key] = value
				}
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				if tc.sizing != "" {
					options.ValuesFiles = append(options.ValuesFiles, filepath.Join("values", tc.sizing))
				}
				beforeOutput := renderRBACNamingChart(t, beforeChart, tc.release, options, upgrade)
				afterOutput := renderRBACNamingChart(t, "..", tc.release, options, upgrade)
				beforeResources := helmtestutil.NewKubernetesResources(t, beforeOutput)
				afterResources := helmtestutil.NewKubernetesResources(t, afterOutput)
				renames := map[string]string{}
				for _, purpose := range []string{"authentication", "authorization"} {
					old, target := tc.legacyPrefix+"-"+purpose, tc.targetPrefix+"-"+purpose
					if tc.disabled {
						assert.NotContains(t, beforeResources.ClusterRoleBindings, old)
						assert.NotContains(t, afterResources.ClusterRoleBindings, target)
						continue
					}
					require.Contains(t, beforeResources.ClusterRoleBindings, old)
					require.Contains(t, afterResources.ClusterRoleBindings, target)
					expected := beforeResources.ClusterRoleBindings[old]
					expected.Name = target
					actual := afterResources.ClusterRoleBindings[target]
					assert.Equal(t, expected, actual, "Role reference, subjects and all other fields must stay unchanged")
					assert.Empty(t, validation.IsDNS1123Subdomain(target))
					require.Len(t, actual.Subjects, 1)
					assert.Equal(t, options.KubectlOptions.Namespace, actual.Subjects[0].Namespace)
					require.Contains(t, afterResources.ServiceAccounts, actual.Subjects[0].Name)
					if purpose == "authentication" {
						assert.Equal(t, "system:auth-delegator", actual.RoleRef.Name)
					} else {
						require.Contains(t, afterResources.ClusterRoles, actual.RoleRef.Name)
					}
					if old != target {
						assert.NotContains(t, afterResources.ClusterRoleBindings, old)
					}
					renames["ClusterRoleBinding/"+old] = target
				}
				for name, binding := range beforeResources.ClusterRoleBindings {
					if after, found := afterResources.ClusterRoleBindings[name]; found {
						assertClusterBindingUpgradePatchUnchanged(t, binding, after)
					}
				}
				seen := map[string]bool{}
				before := backupNamingDocuments(t, beforeOutput, renames, seen)
				assert.Equal(t, before, backupNamingDocuments(t, afterOutput, nil, nil), "Only the two main ClusterRoleBinding metadata names may change")
				for name := range renames {
					require.True(t, seen[name], name)
				}
				assert.Equal(t, beforeResources.ClusterRoles, afterResources.ClusterRoles)
				assert.Equal(t, beforeResources.ServiceAccounts, afterResources.ServiceAccounts)
				assert.Equal(t, beforeResources.Deployments, afterResources.Deployments)
				assert.Equal(t, beforeResources.Statefulsets, afterResources.Statefulsets)
				assert.Equal(t, beforeResources.ConfigMaps, afterResources.ConfigMaps)
				assert.Equal(t, beforeResources.PersistentVolumeClaims, afterResources.PersistentVolumeClaims)
			})
		}
	}
}

func TestClusterBindingNamingKeepsNamespacesDistinct(t *testing.T) {
	names := map[string]bool{}
	for _, namespace := range []string{"suse-observability", "tenant-a", "tenant-b", strings.Repeat("n", 62) + "a", strings.Repeat("n", 62) + "b"} {
		options := apiResourceNameTestOptions(nil)
		options.KubectlOptions.Namespace = namespace
		output := renderRBACNamingChart(t, "..", "suse-observability", options, false)
		resources := helmtestutil.NewKubernetesResources(t, output)
		prefix := "suse-observability-" + namespace
		if namespace == "suse-observability" {
			prefix = "suse-observability"
		}
		for _, purpose := range []string{"authentication", "authorization"} {
			name := prefix + "-" + purpose
			require.Contains(t, resources.ClusterRoleBindings, name)
			assert.False(t, names[name], "Different namespaces must never share a main cluster binding")
			names[name] = true
			assert.Equal(t, namespace, resources.ClusterRoleBindings[name].Subjects[0].Namespace)
		}
	}
}

func assertClusterBindingUpgradePatchUnchanged(t *testing.T, before, after rbacv1.ClusterRoleBinding) {
	t.Helper()
	require.Equal(t, before.RoleRef, after.RoleRef, before.Name)
	current := before.DeepCopy()
	if current.Annotations == nil {
		current.Annotations = map[string]string{}
	}
	current.Annotations["fixture.customer"] = "keep"
	current.ResourceVersion = "fixture-version"
	original, err := json.Marshal(before)
	require.NoError(t, err)
	modified, err := json.Marshal(after)
	require.NoError(t, err)
	live, err := json.Marshal(current)
	require.NoError(t, err)
	metadata, err := strategicpatch.NewPatchMetaFromStruct(rbacv1.ClusterRoleBinding{})
	require.NoError(t, err)
	patch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, live, metadata, true)
	require.NoError(t, err)
	assert.JSONEq(t, "{}", string(patch), before.Name)
}
