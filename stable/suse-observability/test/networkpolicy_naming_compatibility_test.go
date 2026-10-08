package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	networkingv1 "k8s.io/api/networking/v1"
)

const canonicalNetworkPolicy = "suse-observability-networkpolicy"

// Frozen policies were rendered from the pre-rename Git revision documented in
// testdata/networkpolicy-renaming/README.md. Only metadata.name may change.
func TestNetworkPolicyNamingFrozenCompatibility(t *testing.T) {
	checkNetworkPolicyNamingFixtures(t, "..", true)
}

func TestNetworkPolicyNamingBaselineReproduction(t *testing.T) {
	chart := os.Getenv("NETWORK_POLICY_BASELINE_CHART")
	if chart == "" {
		t.Skip("set NETWORK_POLICY_BASELINE_CHART to the pre-rename chart with built dependencies")
	}
	checkNetworkPolicyNamingFixtures(t, chart, false)
}

func checkNetworkPolicyNamingFixtures(t *testing.T, chart string, renamed bool) {
	t.Helper()
	for _, fixture := range []string{"default", "custom"} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", fixture, upgrade), func(t *testing.T) {
				data, err := os.ReadFile(filepath.Join("testdata/networkpolicy-renaming", fixture+".json"))
				require.NoError(t, err)
				var expected networkingv1.NetworkPolicy
				require.NoError(t, json.Unmarshal(data, &expected))
				if renamed {
					expected.Name = canonicalNetworkPolicy
				}
				options := apiResourceNameTestOptions(map[string]string{"networkPolicy.enabled": "true"})
				if fixture == "custom" {
					options.ValuesFiles = append(options.ValuesFiles, "values/networkpolicy-custom.yaml")
				}
				output := renderNetworkPolicyNamingChart(t, chart, "nightly", options, upgrade)
				actual := networkPolicyNamingResource(t, output, expected.Name)
				delete(actual.Labels, "helm.sh/chart")
				delete(actual.Labels, "app.kubernetes.io/version")
				assert.Equal(t, expected, actual)
			})
		}
	}
}

func TestNetworkPolicyNamingPreservesPlatformResources(t *testing.T) {
	beforeChart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(beforeChart, os.DirFS("..")))
	policyTemplate := filepath.Join(beforeChart, "templates/global/networkpolicy.yaml")
	data, err := os.ReadFile(policyTemplate)
	require.NoError(t, err)
	// Restore the exact old producer: common.networkpolicy alone supplied the
	// name, including its common-value merging and long-name truncation rules.
	declaration := "{{- $_ := set $commonNetworkPolicy.metadata \"name\" (include \"stackstate.networkpolicy.fullname\" .) -}}\n"
	require.Equal(t, 1, strings.Count(string(data), declaration))
	require.NoError(t, os.WriteFile(policyTemplate, []byte(strings.Replace(string(data), declaration, "", 1)), 0600))

	for _, tc := range []struct {
		name, release, namespace, legacy, sizing string
		custom, disabled                         bool
		values                                   map[string]string
	}{
		{name: "product-release", release: "suse-observability", legacy: "suse-observability-suse-observability"},
		{name: "nightly", release: "nightly", legacy: "nightly-suse-observability"},
		{name: "same-release-other-namespace", release: "suse-observability", namespace: "tenant-b", legacy: "suse-observability-suse-observability"},
		{name: "fullname", release: "nightly", legacy: "customer", values: map[string]string{"fullnameOverride": "CUSTOMER"}},
		{name: "affixes", release: "nightly", legacy: "g-pre-customer-post-end", values: map[string]string{
			"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "global-override", release: "nightly", legacy: "nightly-suse-observability", values: map[string]string{"global.fullnameOverride": "customer"}},
		{name: "long-release", release: strings.Repeat("x", 50), legacy: strings.Repeat("x", 50) + "-sus"},
		{name: "long-override", release: "nightly", legacy: strings.Repeat("a", 54), values: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "already-canonical", release: "nightly", legacy: canonicalNetworkPolicy, values: map[string]string{"fullnameOverride": canonicalNetworkPolicy}},
		{name: "common-overrides", release: "nightly", legacy: "shared-identity", values: map[string]string{"common.fullnameOverride": "shared-identity"}},
		{name: "custom-rules-and-metadata", release: "nightly", legacy: "nightly-suse-observability", custom: true},
		{name: "ha", release: "nightly", legacy: "nightly-suse-observability", sizing: "global_sizing_150_ha.yaml"},
		{name: "nonha", release: "nightly", legacy: "nightly-suse-observability", sizing: "global_sizing_50_nonha.yaml"},
		{name: "split-workers", release: "nightly", legacy: "nightly-suse-observability", values: map[string]string{
			"stackstate.components.receiver.split.enabled": "true", "stackstate.components.correlate.split.enabled": "true",
		}},
		{name: "argo", release: "nightly", legacy: "nightly-suse-observability", values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
		{name: "collector-policy-independent", release: "nightly", legacy: "nightly-suse-observability", values: map[string]string{
			"opentelemetry-collector.networkPolicy.enabled": "true", "opentelemetry-collector.fullnameOverride": "customer-collector",
		}},
		{name: "disabled", release: "nightly", legacy: "nightly-suse-observability", disabled: true},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				values := componentResourceNameTestValues()
				values["networkPolicy.enabled"] = fmt.Sprint(!tc.disabled)
				options := apiResourceNameTestOptions(values)
				for key, value := range tc.values {
					options.SetValues[key] = value
				}
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				if tc.custom {
					options.ValuesFiles = append(options.ValuesFiles, "values/networkpolicy-custom.yaml")
				}
				if tc.sizing != "" {
					options.ValuesFiles = append(options.ValuesFiles, filepath.Join("values", tc.sizing))
				}
				beforeOutput := renderNetworkPolicyNamingChart(t, beforeChart, tc.release, options, upgrade)
				afterOutput := renderNetworkPolicyNamingChart(t, "..", tc.release, options, upgrade)
				renames := map[string]string{}
				if !tc.disabled {
					renames["NetworkPolicy/"+tc.legacy] = canonicalNetworkPolicy
					beforePolicy := networkPolicyNamingResource(t, beforeOutput, tc.legacy)
					afterPolicy := networkPolicyNamingResource(t, afterOutput, canonicalNetworkPolicy)
					beforePolicy.Name = canonicalNetworkPolicy
					assert.Equal(t, beforePolicy, afterPolicy, "Rules, selectors and all remaining metadata must be unchanged")
					assert.Equal(t, tc.release, afterPolicy.Labels["app.kubernetes.io/instance"])
				}
				seen := map[string]bool{}
				before := backupNamingDocuments(t, beforeOutput, renames, seen)
				after := backupNamingDocuments(t, afterOutput, nil, nil)
				assert.Equal(t, before, after, "Only the main NetworkPolicy metadata name may change")
				if !tc.disabled {
					require.True(t, seen["NetworkPolicy/"+tc.legacy])
				} else {
					assert.NotContains(t, after, "NetworkPolicy//"+canonicalNetworkPolicy)
					assert.NotContains(t, before, "NetworkPolicy//"+tc.legacy)
				}
				// Preserve explicit Pod/controller/storage checks alongside the raw
				// comparison, including configuration checksum and credential refs.
				beforeResources := helmtestutil.NewKubernetesResources(t, beforeOutput)
				afterResources := helmtestutil.NewKubernetesResources(t, afterOutput)
				assert.Equal(t, beforeResources.Deployments, afterResources.Deployments)
				assert.Equal(t, beforeResources.Statefulsets, afterResources.Statefulsets)
				assert.Equal(t, beforeResources.DaemonSets, afterResources.DaemonSets)
				assert.Equal(t, beforeResources.PersistentVolumeClaims, afterResources.PersistentVolumeClaims)
				assert.Equal(t, beforeResources.Services, afterResources.Services)
			})
		}
	}
}

func TestNetworkPolicyNamingUsesDedicatedHelper(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates/_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	header := `define "stackstate.networkpolicy.fullname"`
	require.Equal(t, 1, strings.Count(string(data), header))
	definitions := strings.Replace(string(data), header, `define "test.original.stackstate.networkpolicy.fullname"`, 1)
	definitions += "\n{{- " + header + " -}}fixture-policy{{- end -}}\n"
	require.NoError(t, os.WriteFile(path, []byte(definitions), 0600))
	options := apiResourceNameTestOptions(map[string]string{"networkPolicy.enabled": "true"})
	output := renderNetworkPolicyNamingChart(t, chart, "nightly", options, false)
	policy := networkPolicyNamingResource(t, output, "fixture-policy")
	assert.Equal(t, "fixture-policy", policy.Name)
}

func renderNetworkPolicyNamingChart(t *testing.T, chart, release string, options *helm.Options, upgrade bool) string {
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

func networkPolicyNamingResource(t *testing.T, output, name string) networkingv1.NetworkPolicy {
	t.Helper()
	resources := helmtestutil.NewKubernetesResources(t, output)
	require.Contains(t, resources.Unmapped, name)
	var policy networkingv1.NetworkPolicy
	helm.UnmarshalK8SYaml(t, resources.Unmapped[name], &policy)
	require.Equal(t, "NetworkPolicy", policy.Kind)
	return policy
}
