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
	policyv1 "k8s.io/api/policy/v1beta1"
)

// These fixtures come from c851a02ae, before the next renaming phase. They
// freeze the complete main-chart PDBs, excluding chart/app version labels.
// Future renames must explicitly allow only metadata.name changes here.
func TestPDBNamingBaselineCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, release, prefix, namespace, fixture, valuesFile string
		set                                                   map[string]string
		disabled                                              []string
	}{
		{name: "default", release: "suse-observability", prefix: "suse-observability"},
		{name: "mono", release: "suse-observability", prefix: "suse-observability", fixture: "mono",
			set: map[string]string{"stackstate.features.server.split": "false"}},
		{name: "nightly", release: "nightly", prefix: "nightly-suse-observability"},
		{name: "nightly-mono", release: "nightly", prefix: "nightly-suse-observability", fixture: "mono",
			set: map[string]string{"stackstate.features.server.split": "false"}},
		{name: "second-namespace", release: "nightly", prefix: "nightly-suse-observability", namespace: "tenant-a"},
		{name: "fullname-override", release: "nightly", prefix: "customer",
			set: map[string]string{"fullnameOverride": "Customer"}},
		{name: "prefix-suffix", release: "nightly", prefix: "g-pre-customer-post-end",
			set: map[string]string{
				"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
				"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
			}},
		{name: "local-prefix", release: "nightly", prefix: "pre-nightly-suse-observability",
			set: map[string]string{"fullnamePrefix": "pre-"}},
		{name: "local-suffix", release: "nightly", prefix: "nightly-suse-observability-post",
			set: map[string]string{"fullnameSuffix": "-post"}},
		{name: "global-prefix", release: "nightly", prefix: "g-nightly-suse-observability",
			set: map[string]string{"global.fullnamePrefix": "g-"}},
		{name: "global-suffix", release: "nightly", prefix: "nightly-suse-observability-end",
			set: map[string]string{"global.fullnameSuffix": "-end"}},
		{name: "global-override", release: "nightly", prefix: "nightly-suse-observability",
			set: map[string]string{"global.fullnameOverride": "global-only"}},
		{name: "long-release", release: strings.Repeat("x", 50), prefix: strings.Repeat("x", 50) + "-sus"},
		{name: "long-override", release: "nightly", prefix: strings.Repeat("a", 54),
			set: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "split-workers", release: "nightly", prefix: "nightly-suse-observability",
			set: map[string]string{
				"stackstate.components.receiver.split.enabled":  "true",
				"stackstate.components.correlate.split.enabled": "true",
			}},
		{name: "optional-disabled", release: "nightly", prefix: "nightly-suse-observability",
			set: map[string]string{
				"ai.assistant.enabled": "false", "ai.mcp.enabled": "false",
				"stackstate.components.workloadObserver.enabled": "false",
				"stackstate.k8sAuthorization.enabled":            "false",
				"victoria-metrics-1.enabled":                     "false",
			},
			disabled: []string{"ai-assistant", "mcp", "workload-observer", "authorization-sync", "victoriametrics"}},
		{name: "ha", release: "nightly", prefix: "nightly-suse-observability", fixture: "split",
			valuesFile: "values/global_sizing_150_ha.yaml"},
		{name: "nonha", release: "nightly", prefix: "nightly-suse-observability", fixture: "nonha",
			valuesFile: "values/global_sizing_50_nonha.yaml"},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				fixture := tc.fixture
				if fixture == "" {
					fixture = "split"
				}
				data, err := os.ReadFile(filepath.Join("testdata/pdb-renaming", fixture+".json"))
				require.NoError(t, err)
				var expected map[string]policyv1.PodDisruptionBudget
				require.NoError(t, json.Unmarshal(data, &expected))
				for _, component := range tc.disabled {
					require.Contains(t, expected, component)
					delete(expected, component)
				}
				for component, pdb := range expected {
					require.True(t, strings.HasPrefix(pdb.Name, "suse-observability-"))
					pdb.Name = tc.prefix + strings.TrimPrefix(pdb.Name, "suse-observability")
					pdb.Labels["app.kubernetes.io/instance"] = tc.release
					if _, found := pdb.Spec.Selector.MatchLabels["app.kubernetes.io/instance"]; found {
						pdb.Spec.Selector.MatchLabels["app.kubernetes.io/instance"] = tc.release
					}
					expected[component] = pdb
				}

				values := map[string]string{
					"stackstate.features.server.split":               "true",
					"stackstate.components.receiver.split.enabled":   "false",
					"stackstate.components.correlate.split.enabled":  "false",
					"ai.assistant.enabled":                           "true",
					"ai.mcp.enabled":                                 "true",
					"stackstate.components.workloadObserver.enabled": "true",
					"stackstate.k8sAuthorization.enabled":            "true",
					"victoria-metrics-1.enabled":                     "true",
				}
				for key, value := range tc.set {
					values[key] = value
				}
				if values["stackstate.components.correlate.split.enabled"] == "true" {
					// Split workers require explicit JVM sizing in the legacy chart.
					for _, worker := range []string{"connection", "httpTracing", "aggregator"} {
						prefix := "stackstate.components.correlate.split." + worker + "."
						values[prefix+"resources.limits.memory"] = "2Gi"
						values[prefix+"resources.requests.memory"] = "1Gi"
						values[prefix+"sizing.baseMemoryConsumption"] = "400Mi"
						values[prefix+"sizing.javaHeapMemoryFraction"] = "65"
					}
				}
				options := apiResourceNameTestOptions(values)
				if tc.valuesFile != "" {
					options.ValuesFiles = append(options.ValuesFiles, tc.valuesFile)
				}
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				args := []string{"--api-versions", "policy/v1/PodDisruptionBudget"}
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, tc.release, options, args...)
				actual := map[string]policyv1.PodDisruptionBudget{}
				for _, document := range strings.Split(output, "\n---\n") {
					// Main-chart PDBs are emitted by templates/<component>/pdb-*.yaml.
					// Keep subchart budgets outside this rename allowlist.
					if !strings.Contains(document, "# Source: suse-observability/templates/") ||
						!strings.Contains(document, "\nkind: PodDisruptionBudget\n") {
						continue
					}
					var pdb policyv1.PodDisruptionBudget
					require.NoError(t, helm.UnmarshalK8SYamlE(t, document, &pdb))
					component := pdb.Labels["app.kubernetes.io/component"]
					require.NotContains(t, actual, component, "Each main-chart component must have only one PDB")
					delete(pdb.Labels, "app.kubernetes.io/version")
					delete(pdb.Labels, "helm.sh/chart")
					actual[component] = pdb
				}
				assert.Equal(t, expected, actual, "Only explicitly reviewed PDB identity changes may be allowed")
			})
		}
	}
}
