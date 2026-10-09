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
	appsv1 "k8s.io/api/apps/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

var hbasePDBComponents = []string{"hbase-master", "hbase-rs", "hdfs-nn", "hdfs-snn", "hdfs-dn", "tephra"}

// Frozen pre-rename budgets from 3ea779a7170a4603240376c7ba2c18c344db53df
// protect complete specs, labels, annotations and mode-dependent creation.
func TestHBasePDBNamingFrozenCompatibility(t *testing.T) {
	checkHBasePDBFixtures(t, hbasePDBTestChart(t, "..", false), true)
}

func TestHBasePDBNamingBaselineReproduction(t *testing.T) {
	chart := os.Getenv("HBASE_PDB_BASELINE_CHART")
	if chart == "" {
		t.Skip("set HBASE_PDB_BASELINE_CHART to a pre-rename chart with built dependencies")
	}
	checkHBasePDBFixtures(t, hbasePDBTestChart(t, chart, false), false)
}

func checkHBasePDBFixtures(t *testing.T, chart string, renamed bool) {
	t.Helper()
	for _, tc := range []struct {
		fixture, mode string
		snn           bool
	}{
		{"mono", "Mono", true}, {"distributed", "Distributed", false}, {"distributed-snn", "Distributed", true},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.fixture, upgrade), func(t *testing.T) {
				data, err := os.ReadFile(filepath.Join("testdata/pdb-renaming", tc.fixture+".json"))
				require.NoError(t, err)
				var expected map[string]policyv1beta1.PodDisruptionBudget
				require.NoError(t, json.Unmarshal(data, &expected))
				byName := map[string]policyv1beta1.PodDisruptionBudget{}
				for _, pdb := range expected {
					if renamed {
						pdb.Name = "suse-observability-" + strings.TrimPrefix(pdb.Name, "nightly-hbase-")
					}
					byName[pdb.Name] = pdb
				}
				output := renderHBasePDBChart(t, chart, "nightly", "observability", map[string]string{
					"deployment.mode": tc.mode, "hdfs.secondarynamenode.enabled": fmt.Sprint(tc.snn),
				}, upgrade, true)
				resources := helmtestutil.NewKubernetesResources(t, output)
				assert.Equal(t, byName, resources.Pdbs)
			})
		}
	}
}

func TestHBasePDBNamingPreservesPersistentResources(t *testing.T) {
	beforeChart := hbasePDBTestChart(t, "..", true)
	afterChart := hbasePDBTestChart(t, "..", false)
	for _, tc := range []struct {
		name, release, namespace, base string
		values                         map[string]string
	}{
		{name: "product-release", release: "suse-observability", base: "suse-observability-hbase"},
		{name: "custom-release", release: "nightly", base: "nightly-hbase"},
		{name: "same-release-other-namespace", release: "suse-observability", namespace: "tenant-b", base: "suse-observability-hbase"},
		{name: "release-matches-chart", release: "hbase", base: "hbase"},
		{name: "fullname", release: "nightly", base: "customer", values: map[string]string{"fullnameOverride": "CUSTOMER"}},
		{name: "affixes", release: "nightly", base: "g-pre-customer-post-end", values: map[string]string{
			"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "global-override", release: "nightly", base: "nightly-hbase", values: map[string]string{"global.fullnameOverride": "customer-backup"}},
		{name: "long-name", release: "nightly", base: strings.Repeat("a", 54), values: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "already-canonical", release: "nightly", base: "suse-observability", values: map[string]string{"fullnameOverride": "suse-observability"}},
		{name: "metrics-disabled", release: "nightly", base: "nightly-hbase", values: map[string]string{"all.metrics.enabled": "false"}},
		{name: "shared-monitor-namespace", release: "nightly", base: "nightly-hbase", values: map[string]string{"servicemonitor.namespace": "shared-monitoring"}},
		{name: "metadata-and-storage", release: "nightly", base: "nightly-hbase", values: map[string]string{
			"commonLabels.fixture-scope": "preserved", "global.commonLabels.fixture-pod": "preserved",
			"poddisruptionbudget.annotations.fixture-note": "preserved",
			"hdfs.datanode.persistence.size":               "17Gi", "hdfs.datanode.persistence.storageClass": "customer-storage",
			"all.image.pullSecretName": "customer-registry", "all.extraEnv.secret.TEST_KEY": "fixture-value",
		}},
		{name: "legacy-api", release: "nightly", base: "nightly-hbase"},
	} {
		for _, mode := range []string{"Mono", "Distributed", "Distributed-SNN"} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/upgrade=%t", tc.name, mode, upgrade), func(t *testing.T) {
					values := map[string]string{
						"deployment.mode": "Distributed", "hdfs.secondarynamenode.enabled": fmt.Sprint(mode == "Distributed-SNN"),
						"all.metrics.servicemonitor.enabled": "true",
					}
					if mode == "Mono" {
						values["deployment.mode"] = "Mono"
					}
					maps.Copy(values, tc.values)
					namespace := tc.namespace
					if namespace == "" {
						namespace = "observability"
					}
					before := helmtestutil.NewKubernetesResources(t, renderHBasePDBChart(t, beforeChart, tc.release, namespace, values, upgrade, tc.name != "legacy-api"))
					after := helmtestutil.NewKubernetesResources(t, renderHBasePDBChart(t, afterChart, tc.release, namespace, values, upgrade, tc.name != "legacy-api"))
					renameHBasePDBs(t, before.Pdbs, tc.base)
					assert.Equal(t, before, after, "Only six reviewed PDB metadata names may change")
					for name, pdb := range after.Pdbs {
						assert.Equal(t, int32(1), pdb.Spec.MaxUnavailable.IntVal, name)
						selector := labels.SelectorFromSet(pdb.Spec.Selector.MatchLabels)
						matched := 0
						for _, sts := range after.Statefulsets {
							if selector.Matches(labels.Set(sts.Spec.Template.Labels)) {
								matched++
							}
						}
						assert.Greater(t, matched, 0, "%s must still select its storage Pods", name)
					}
					for name, sts := range before.Statefulsets {
						assertHBaseStatefulSetUpgradePatchPreserved(t, sts, after.Statefulsets[name])
					}
				})
			}
		}
	}
}

func renameHBasePDBs(t *testing.T, pdbs map[string]policyv1beta1.PodDisruptionBudget, base string) {
	t.Helper()
	for _, component := range hbasePDBComponents {
		old := base + "-" + component
		if pdb, found := pdbs[old]; found {
			target := "suse-observability-" + component
			pdb.Name = target
			delete(pdbs, old)
			require.NotContains(t, pdbs, target)
			pdbs[target] = pdb
		}
	}
}

// Helm 3 uses three-way strategic merge for StatefulSets. This rename must add
// no changes beyond an ordinary upgrade's reconciliation of live drift.
func assertHBaseStatefulSetUpgradePatchPreserved(t *testing.T, before, after appsv1.StatefulSet) {
	t.Helper()
	current := before.DeepCopy()
	current.Annotations = maps.Clone(current.Annotations)
	if current.Annotations == nil {
		current.Annotations = map[string]string{}
	}
	current.Annotations["fixture.customer"] = "keep"
	current.ResourceVersion = "fixture-version"
	replicas := *current.Spec.Replicas + 1
	current.Spec.Replicas = &replicas
	current.Status.ReadyReplicas = replicas
	original, err := json.Marshal(before)
	require.NoError(t, err)
	modified, err := json.Marshal(after)
	require.NoError(t, err)
	live, err := json.Marshal(current)
	require.NoError(t, err)
	metadata, err := strategicpatch.NewPatchMetaFromStruct(appsv1.StatefulSet{})
	require.NoError(t, err)
	patch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, live, metadata, true)
	require.NoError(t, err)
	ordinaryPatch, err := strategicpatch.CreateThreeWayMergePatch(original, original, live, metadata, true)
	require.NoError(t, err)
	assert.JSONEq(t, string(ordinaryPatch), string(patch), "%s upgrade patch must not change because PDBs were renamed", before.Name)
	patched, err := strategicpatch.StrategicMergePatch(live, patch, appsv1.StatefulSet{})
	require.NoError(t, err)
	var actual appsv1.StatefulSet
	require.NoError(t, json.Unmarshal(patched, &actual))
	assert.Equal(t, current.Spec.VolumeClaimTemplates, actual.Spec.VolumeClaimTemplates)
	assert.Equal(t, current.Spec.ServiceName, actual.Spec.ServiceName)
	assert.Equal(t, "keep", actual.Annotations["fixture.customer"])
	assert.Equal(t, current.Status, actual.Status)
}

func renderHBasePDBChart(t *testing.T, chart, release, namespace string, values map[string]string, upgrade, currentAPI bool) string {
	t.Helper()
	values = maps.Clone(values)
	values["zookeeper.externalServers"] = "test-zookeeper"
	args := []string{"--api-versions", "monitoring.coreos.com/v1"}
	if currentAPI {
		args = append(args, "--api-versions", "policy/v1/PodDisruptionBudget")
	}
	if upgrade {
		args = append(args, "--is-upgrade")
	}
	output, err := helm.RenderTemplateE(t, &helm.Options{
		SetValues: values, KubectlOptions: &k8s.KubectlOptions{Namespace: namespace}, Logger: logger.Discard,
	}, chart, release, nil, args...)
	require.NoError(t, err)
	return output
}

func hbasePDBTestChart(t *testing.T, source string, legacy bool) string {
	t.Helper()
	chart := filepath.Join(t.TempDir(), "chart")
	// Copy render inputs only; the chart's linter values are a symlink which
	// os.CopyFS rejects. Tests and lint configuration are not runtime inputs.
	for _, directory := range []string{"templates", "scripts", "charts"} {
		require.NoError(t, os.CopyFS(filepath.Join(chart, directory), os.DirFS(filepath.Join(source, directory))))
	}
	for _, filename := range []string{"Chart.yaml", "Chart.lock", "values.yaml", ".helmignore"} {
		data, err := os.ReadFile(filepath.Join(source, filename))
		require.NoError(t, err)
		if filename == "Chart.yaml" {
			content := string(data)
			for key, value := range map[string]string{"version": "0.0.0", "appVersion": "naming-fixture"} {
				pattern := regexp.MustCompile(`(?m)^` + key + `:.*$`)
				require.Len(t, pattern.FindAllString(content, -1), 1)
				content = pattern.ReplaceAllString(content, key+": "+value)
			}
			data = []byte(content)
		}
		require.NoError(t, os.WriteFile(filepath.Join(chart, filename), data, 0600))
	}
	if legacy {
		path := filepath.Join(chart, "templates", "_names.tpl")
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		content := string(data)
		for helper, suffix := range map[string]string{
			"hbase.master": "hbase-master", "hbase.regionserver": "hbase-rs",
			"hdfs.namenode": "hdfs-nn", "hdfs.secondarynamenode": "hdfs-snn",
			"hdfs.datanode": "hdfs-dn", "tephra": "tephra",
		} {
			header := `{{- define "hbase.` + helper + `.poddisruptionbudget.fullname" -}}`
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(header) + `.*?\{\{- end -\}\}`)
			require.Len(t, pattern.FindAllString(content, -1), 1)
			content = pattern.ReplaceAllString(content, header+`{{ template "common.fullname.short" . }}-`+suffix+`{{- end -}}`)
		}
		require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	}
	return chart
}
