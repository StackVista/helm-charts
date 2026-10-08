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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

const canonicalPodReader = "suse-observability-get-pods"

type podReaderNamingFixture struct {
	Role    rbacv1.Role        `json:"Role"`
	Binding rbacv1.RoleBinding `json:"RoleBinding"`
}

// Complete pre-rename objects from c6030933f0d8971cb1b8c4ea5079d487089f0e84.
// Only the binding's metadata name changes; its Role and subjects are retained.
func TestPodReaderNamingFrozenCompatibility(t *testing.T) {
	checkPodReaderNamingFixtures(t, "..", true)
}

func TestPodReaderNamingBaselineReproduction(t *testing.T) {
	chart := os.Getenv("POD_READER_BASELINE_CHART")
	if chart == "" {
		t.Skip("set POD_READER_BASELINE_CHART to the pre-rename chart with built dependencies")
	}
	checkPodReaderNamingFixtures(t, chart, false)
}

func checkPodReaderNamingFixtures(t *testing.T, chart string, renamed bool) {
	t.Helper()
	for _, mode := range []string{"split", "monolithic"} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", mode, upgrade), func(t *testing.T) {
				data, err := os.ReadFile(filepath.Join("testdata/pod-reader-renaming", mode+".json"))
				require.NoError(t, err)
				var expected podReaderNamingFixture
				require.NoError(t, json.Unmarshal(data, &expected))
				if renamed {
					expected.Binding.Name = canonicalPodReader
				}
				options := apiResourceNameTestOptions(map[string]string{"stackstate.features.server.split": fmt.Sprint(mode == "split")})
				output := renderPodReaderNamingChart(t, chart, "nightly", options, upgrade)
				resources := helmtestutil.NewKubernetesResources(t, output)
				require.Contains(t, resources.Roles, expected.Role.Name)
				require.Contains(t, resources.RoleBindings, expected.Binding.Name)
				actual := podReaderNamingFixture{
					Role: resources.Roles[expected.Role.Name], Binding: resources.RoleBindings[expected.Binding.Name],
				}
				for _, labels := range []map[string]string{actual.Role.Labels, actual.Binding.Labels} {
					delete(labels, "helm.sh/chart")
					delete(labels, "app.kubernetes.io/version")
				}
				assert.Equal(t, expected, actual)
			})
		}
	}
}

func TestPodReaderNamingPreservesPlatformResources(t *testing.T) {
	beforeChart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(beforeChart, os.DirFS("..")))
	path := filepath.Join(beforeChart, "templates/_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	header := `{{- define "stackstate.getPods.rolebinding.fullname" -}}`
	pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(header) + `.*?\{\{- end -\}\}`)
	require.Len(t, pattern.FindAllString(content, -1), 1)
	content = pattern.ReplaceAllString(content, header+`{{ template "common.fullname.short" . }}-get-pods{{- end -}}`)
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	for _, tc := range []struct {
		name, release, namespace, legacyPrefix, sizing string
		values                                         map[string]string
	}{
		{name: "product-release", release: "suse-observability", legacyPrefix: "suse-observability"},
		{name: "nightly", release: "nightly", legacyPrefix: "nightly-suse-observability"},
		{name: "same-release-other-namespace", release: "suse-observability", namespace: "tenant-b", legacyPrefix: "suse-observability"},
		{name: "fullname", release: "nightly", legacyPrefix: "customer", values: map[string]string{"fullnameOverride": "CUSTOMER"}},
		{name: "affixes", release: "nightly", legacyPrefix: "g-pre-customer-post-end", values: map[string]string{
			"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "global-override", release: "nightly", legacyPrefix: "nightly-suse-observability", values: map[string]string{"global.fullnameOverride": "customer"}},
		{name: "long-release", release: strings.Repeat("x", 50), legacyPrefix: strings.Repeat("x", 50) + "-sus"},
		{name: "long-override", release: "nightly", legacyPrefix: strings.Repeat("a", 54), values: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "already-canonical", release: "nightly", legacyPrefix: "suse-observability", values: map[string]string{"fullnameOverride": "suse-observability"}},
		{name: "monolithic", release: "nightly", legacyPrefix: "nightly-suse-observability", values: map[string]string{"stackstate.features.server.split": "false"}},
		{name: "ha", release: "nightly", legacyPrefix: "nightly-suse-observability", sizing: "global_sizing_150_ha.yaml"},
		{name: "nonha", release: "nightly", legacyPrefix: "nightly-suse-observability", sizing: "global_sizing_50_nonha.yaml"},
		{name: "argo", release: "nightly", legacyPrefix: "nightly-suse-observability", values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
		{name: "authorization-disabled", release: "nightly", legacyPrefix: "nightly-suse-observability", values: map[string]string{"stackstate.k8sAuthorization.enabled": "false"}},
		{name: "cluster-rbac-disabled", release: "nightly", legacyPrefix: "nightly-suse-observability", values: map[string]string{"cluster-role.enabled": "false"}},
		{name: "custom-metadata", release: "nightly", legacyPrefix: "nightly-suse-observability", values: map[string]string{
			"role.annotations.fixture-note": "preserved", "rolebinding.annotations.fixture-note": "preserved",
			"commonLabels.fixture-scope": "preserved",
		}},
		{name: "split-workers", release: "nightly", legacyPrefix: "nightly-suse-observability", values: map[string]string{
			"stackstate.components.receiver.split.enabled": "true", "stackstate.components.correlate.split.enabled": "true",
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
				beforeOutput := renderPodReaderNamingChart(t, beforeChart, tc.release, options, upgrade)
				afterOutput := renderPodReaderNamingChart(t, "..", tc.release, options, upgrade)
				beforeResources := helmtestutil.NewKubernetesResources(t, beforeOutput)
				afterResources := helmtestutil.NewKubernetesResources(t, afterOutput)
				legacy := tc.legacyPrefix + "-get-pods"
				require.Contains(t, beforeResources.Roles, legacy)
				require.Contains(t, beforeResources.RoleBindings, legacy)
				require.Contains(t, afterResources.Roles, legacy)
				require.Contains(t, afterResources.RoleBindings, canonicalPodReader)
				assert.Equal(t, beforeResources.Roles, afterResources.Roles, "All Role identities and rules must stay unchanged")
				expectedBinding := beforeResources.RoleBindings[legacy]
				assert.Equal(t, legacy, expectedBinding.RoleRef.Name)
				expectedBinding.Name = canonicalPodReader
				assert.Equal(t, expectedBinding, afterResources.RoleBindings[canonicalPodReader])
				if legacy != canonicalPodReader {
					assert.NotContains(t, afterResources.RoleBindings, legacy)
				}
				for name, binding := range beforeResources.RoleBindings {
					if after, found := afterResources.RoleBindings[name]; found {
						assertPodReaderBindingUpgradePatchUnchanged(t, binding, after)
					}
				}

				renames := map[string]string{"RoleBinding/" + legacy: canonicalPodReader}
				seen := map[string]bool{}
				before := backupNamingDocuments(t, beforeOutput, renames, seen)
				assert.Equal(t, before, backupNamingDocuments(t, afterOutput, nil, nil), "Only the internal Pod-reader binding metadata name may change")
				require.True(t, seen["RoleBinding/"+legacy])
				assert.Equal(t, beforeResources.ServiceAccounts, afterResources.ServiceAccounts)
				assert.Equal(t, beforeResources.Deployments, afterResources.Deployments)
				assert.Equal(t, beforeResources.Statefulsets, afterResources.Statefulsets)
				assert.Equal(t, beforeResources.PersistentVolumeClaims, afterResources.PersistentVolumeClaims)
				assert.Equal(t, beforeResources.ConfigMaps, afterResources.ConfigMaps, "Application authorization subjects and configuration checksums must stay unchanged")
			})
		}
	}
}

// Helm 3 must not patch an immutable roleRef for bindings retaining their names.
// Unmanaged live annotations are preserved without changing the existing grants.
func assertPodReaderBindingUpgradePatchUnchanged(t *testing.T, before, after rbacv1.RoleBinding) {
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
	metadata, err := strategicpatch.NewPatchMetaFromStruct(rbacv1.RoleBinding{})
	require.NoError(t, err)
	patch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, live, metadata, true)
	require.NoError(t, err)
	assert.JSONEq(t, "{}", string(patch), before.Name)
}

func renderPodReaderNamingChart(t *testing.T, chart, release string, options *helm.Options, upgrade bool) string {
	t.Helper()
	copied := *options
	copied.ValuesFiles = make([]string, len(options.ValuesFiles))
	for i, file := range options.ValuesFiles {
		path, err := filepath.Abs(file)
		require.NoError(t, err)
		copied.ValuesFiles[i] = path
	}
	args := []string{"--api-versions", "policy/v1/PodDisruptionBudget", "--api-versions", "monitoring.coreos.com/v1"}
	if upgrade {
		args = append(args, "--is-upgrade")
	}
	output, err := helm.RenderTemplateE(t, &copied, chart, release, nil, args...)
	require.NoError(t, err)
	return output
}
