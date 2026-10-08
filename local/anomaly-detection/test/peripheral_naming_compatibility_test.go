package test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	anomalyCanonicalPDB     = "suse-observability-anomaly-detection"
	anomalyCanonicalMonitor = "suse-observability-spotlight-manager"
)

// Frozen objects were rendered before changing the two naming helpers, from
// 9ed0a1cb4837895c16e7e1c1b23ed2dfa8e69c13. No field except metadata.name changes.
func TestAnomalyPeripheralNamingFrozenCompatibility(t *testing.T) {
	checkAnomalyPeripheralFixtures(t, anomalyPeripheralTestChart(t, "..", false), true)
}

func TestAnomalyPeripheralNamingBaselineReproduction(t *testing.T) {
	source := os.Getenv("ANOMALY_PERIPHERAL_BASELINE_CHART")
	if source == "" {
		t.Skip("set ANOMALY_PERIPHERAL_BASELINE_CHART to a pre-rename chart with built dependencies")
	}
	checkAnomalyPeripheralFixtures(t, anomalyPeripheralTestChart(t, source, false), false)
}

func checkAnomalyPeripheralFixtures(t *testing.T, chart string, renamed bool) {
	t.Helper()
	for _, fixture := range []string{"default", "custom-budget"} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", fixture, upgrade), func(t *testing.T) {
				data, err := os.ReadFile(filepath.Join("testdata/peripheral-renaming", fixture+".json"))
				require.NoError(t, err)
				var expected map[string]map[string]interface{}
				require.NoError(t, json.Unmarshal(data, &expected))
				if renamed {
					renameAnomalyPeripheralDocuments(t, expected, "nightly-anomaly-detection", true)
				}
				values := anomalyPeripheralValues()
				if fixture == "custom-budget" {
					maps.Copy(values, anomalyPeripheralCustomBudgetValues())
				}
				output := renderAnomalyPeripheralChart(t, chart, "nightly", "observability", values, upgrade)
				actual := anomalyNamingDocuments(t, output, nil, nil)
				for key, object := range actual {
					if object["kind"] != "PodDisruptionBudget" && object["kind"] != "ServiceMonitor" {
						delete(actual, key)
					}
				}
				expectedJSON, err := json.Marshal(expected)
				require.NoError(t, err)
				actualJSON, err := json.Marshal(actual)
				require.NoError(t, err)
				assert.JSONEq(t, string(expectedJSON), string(actualJSON))
			})
		}
	}
}

func TestAnomalyPeripheralNamingPreservesOtherResources(t *testing.T) {
	beforeChart := anomalyPeripheralTestChart(t, "..", true)
	afterChart := anomalyPeripheralTestChart(t, "..", false)
	for _, tc := range []struct {
		name, release, namespace, base string
		values                         map[string]string
	}{
		{name: "product-release", release: "suse-observability", base: "suse-observability-anomaly-detection"},
		{name: "custom-release", release: "nightly", base: "nightly-anomaly-detection"},
		{name: "same-release-other-namespace", release: "suse-observability", namespace: "tenant-b", base: "suse-observability-anomaly-detection"},
		{name: "release-matches-chart", release: "anomaly-detection", base: "anomaly-detection"},
		{name: "fullname", release: "nightly", base: "customer", values: map[string]string{"fullnameOverride": "CUSTOMER"}},
		{name: "affixes", release: "nightly", base: "g-pre-customer-post-end", values: map[string]string{
			"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "global-override", release: "nightly", base: "nightly-anomaly-detection", values: map[string]string{"global.fullnameOverride": "global-service"}},
		{name: "long-release", release: strings.Repeat("r", 53), base: strings.Repeat("r", 53)},
		{name: "already-canonical", release: "nightly", base: "suse-observability", values: map[string]string{"fullnameOverride": "suse-observability"}},
		{name: "monitor-disabled", release: "nightly", base: "nightly-anomaly-detection", values: map[string]string{"metrics.serviceMonitor.enabled": "false"}},
		{name: "custom-budget", release: "nightly", base: "nightly-anomaly-detection", values: anomalyPeripheralCustomBudgetValues()},
		{name: "custom-pdb-selector", release: "nightly", base: "nightly-anomaly-detection", values: map[string]string{
			"pdb.selector.matchLabels.name": "nightly-anomaly-detection-spotlight-manager",
		}},
		{name: "credentials-storage-and-accounts", release: "nightly", base: "nightly-anomaly-detection", values: map[string]string{
			"manager.persistentStorage.size": "17Gi", "manager.persistentStorage.storageClass": "customer-artifacts",
			"stackstate.authType": "cookie", "stackstate.username": "fixture-user", "stackstate.password": "fixture-password",
			"image.pullSecretName": "customer-registry", "global.commonLabels.fixture-scope": "preserved",
			"cluster-role.enabled": "false",
		}},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				values := anomalyPeripheralValues()
				maps.Copy(values, tc.values)
				namespace := tc.namespace
				if namespace == "" {
					namespace = "observability"
				}
				beforeOutput := renderAnomalyPeripheralChart(t, beforeChart, tc.release, namespace, values, upgrade)
				afterOutput := renderAnomalyPeripheralChart(t, afterChart, tc.release, namespace, values, upgrade)
				before := anomalyNamingDocuments(t, beforeOutput, nil, nil)
				after := anomalyNamingDocuments(t, afterOutput, nil, nil)
				monitor := values["metrics.serviceMonitor.enabled"] == "true"
				renameAnomalyPeripheralDocuments(t, before, tc.base, monitor)
				assert.Equal(t, before, after, "Only the two reviewed metadata names may change")
				resources := helmtestutil.NewKubernetesResources(t, afterOutput)
				require.Contains(t, resources.Pdbs, anomalyCanonicalPDB)
				if tc.name == "custom-pdb-selector" {
					pdb := resources.Pdbs[anomalyCanonicalPDB]
					require.NotNil(t, pdb.Spec.Selector)
					selector := labels.SelectorFromSet(pdb.Spec.Selector.MatchLabels)
					require.Contains(t, resources.Deployments, tc.base+"-spotlight-manager")
					assert.True(t, selector.Matches(labels.Set(resources.Deployments[tc.base+"-spotlight-manager"].Spec.Template.Labels)))
				}
				if monitor {
					require.Contains(t, resources.ServiceMonitors, anomalyCanonicalMonitor)
					sm := resources.ServiceMonitors[anomalyCanonicalMonitor]
					assert.Equal(t, namespace, sm.Namespace)
					assert.Equal(t, []string{namespace}, sm.Spec.NamespaceSelector.MatchNames)
					selector := labels.SelectorFromSet(sm.Spec.Selector.MatchLabels)
					require.Contains(t, resources.Services, tc.base+"-spotlight-manager")
					assert.True(t, selector.Matches(labels.Set(resources.Services[tc.base+"-spotlight-manager"].Labels)))
				} else {
					assert.Empty(t, resources.ServiceMonitors)
				}
			})
		}
	}
}

func renameAnomalyPeripheralDocuments(t *testing.T, documents map[string]map[string]interface{}, base string, monitor bool) {
	t.Helper()
	renames := map[string]string{"PodDisruptionBudget/" + base + "-anomaly-detection": anomalyCanonicalPDB}
	if monitor {
		renames["ServiceMonitor/"+base+"-spotlight-manager"] = anomalyCanonicalMonitor
	}
	for old, target := range renames {
		require.Contains(t, documents, old)
		object := documents[old]
		object["metadata"].(map[string]interface{})["name"] = target
		delete(documents, old)
		kind := object["kind"].(string)
		require.NotContains(t, documents, kind+"/"+target, "target name collision")
		documents[kind+"/"+target] = object
	}
}

func anomalyPeripheralValues() map[string]string {
	return map[string]string{
		"stackstate.instance": "https://analysis.example", "global.receiverApiKey": "test-key",
		"metrics.serviceMonitor.enabled": "true",
	}
}

func anomalyPeripheralCustomBudgetValues() map[string]string {
	return map[string]string{
		"pdb.maxUnavailable": "25%", "commonLabels.fixture-scope": "preserved",
		"poddisruptionbudget.annotations.fixture-note": "preserved",
	}
}

func renderAnomalyPeripheralChart(t *testing.T, chart, release, namespace string, values map[string]string, upgrade bool) string {
	t.Helper()
	args := []string{"--api-versions", "policy/v1/PodDisruptionBudget"}
	if upgrade {
		args = append(args, "--is-upgrade")
	}
	output, err := helm.RenderTemplateE(t, &helm.Options{
		SetValues: values, KubectlOptions: &k8s.KubectlOptions{Namespace: namespace}, Logger: logger.Discard,
	}, chart, release, nil, args...)
	require.NoError(t, err)
	return output
}

func anomalyPeripheralTestChart(t *testing.T, source string, legacy bool) string {
	t.Helper()
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS(source)))
	path := filepath.Join(chart, "Chart.yaml")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	for key, value := range map[string]string{"version": "0.0.0", "appVersion": "naming-fixture"} {
		pattern := regexp.MustCompile(`(?m)^` + key + `:.*$`)
		require.Len(t, pattern.FindAllString(content, -1), 1)
		content = pattern.ReplaceAllString(content, key+": "+value)
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	if legacy {
		path := filepath.Join(chart, "templates", "_names.tpl")
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		content := string(data)
		for helper, suffix := range map[string]string{"pdb": "anomaly-detection", "manager.servicemonitor": "spotlight-manager"} {
			header := `{{- define "anomaly-detection.` + helper + `.fullname" -}}`
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(header) + `.*?\{\{- end -\}\}`)
			require.Len(t, pattern.FindAllString(content, -1), 1)
			content = pattern.ReplaceAllString(content, header+`{{ template "common.fullname.short" . }}-`+suffix+`{{- end -}}`)
		}
		require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	}
	return chart
}
