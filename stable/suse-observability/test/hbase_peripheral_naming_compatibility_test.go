package test

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// Actual pre-change source: 225d8e44c4dca0910154c47aa20481a920cf8352.
// Independent extraction baseline: c851a02ae1f2f7eb64c5538a148194ea36eee4b4.
func TestHBasePeripheralNamingCompatibility(t *testing.T) {
	before, child := hbasePeripheralTestChart(t, "..")
	path := filepath.Join(child, "templates/_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	for _, resource := range hbasePeripheralInventory() {
		pattern := hbasePeripheralDefinition(resource.helper)
		require.Len(t, pattern.FindAllString(content, -1), 1)
		content = pattern.ReplaceAllString(content, `{{- define "`+resource.helper+`" -}}{{ include "common.fullname.short" . }}-`+resource.suffix+`{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	checkHBasePeripheralNaming(t, before, "..")
}

func TestHBasePeripheralNamingGitBaselineCompatibility(t *testing.T) {
	before := os.Getenv("HBASE_PERIPHERAL_BASELINE_CHART")
	if before == "" {
		t.Skip("set HBASE_PERIPHERAL_BASELINE_CHART to the built pre-change chart")
	}
	checkHBasePeripheralNaming(t, before, "..")
}

func TestHBasePeripheralNamingIndependentOfPreviousSteps(t *testing.T) {
	before := os.Getenv("HBASE_PERIPHERAL_EXTRACTION_CHART")
	if before == "" {
		t.Skip("set HBASE_PERIPHERAL_EXTRACTION_CHART to the built extraction baseline")
	}
	after, child := hbasePeripheralTestChart(t, before)
	data, err := os.ReadFile("../../../local/hbase/templates/_names.tpl")
	require.NoError(t, err)
	section, _, found := strings.Cut(string(data), "{{/*\nPDB names")
	require.True(t, found)
	path := filepath.Join(child, "templates/_names.tpl")
	original, err := os.ReadFile(path)
	require.True(t, err == nil || os.IsNotExist(err))
	require.NoError(t, os.WriteFile(path, append([]byte(section), original...), 0600))
	for _, resource := range hbasePeripheralInventory() {
		files := []string{resource.file}
		if resource.consumer != "" {
			files = append(files, resource.consumer)
		}
		for _, file := range files {
			path := filepath.Join(child, "templates", file)
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			old := `{{ template "common.fullname.short" . }}-` + resource.suffix
			require.Equal(t, 1, strings.Count(string(data), old), file)
			content := strings.Replace(string(data), old, `{{ include "`+resource.helper+`" . }}`, 1)
			require.NoError(t, os.WriteFile(path, []byte(content), 0600))
		}
	}
	checkHBasePeripheralNaming(t, before, after)
}

type hbasePeripheralResource struct{ kind, component, suffix, helper, file, consumer string }

func hbasePeripheralInventory() []hbasePeripheralResource {
	return []hbasePeripheralResource{
		{"ServiceMonitor", "hbase-master", "hbase-master", "hbase.hbase.master.servicemonitor.fullname", "hbase-master-servicemonitor.yaml", ""},
		{"ServiceMonitor", "hbase-rs", "hbase-rs", "hbase.hbase.regionserver.servicemonitor.fullname", "hbase-rs-servicemonitor.yaml", ""},
		{"ServiceMonitor", "hdfs-nn", "hdfs-nn", "hbase.hdfs.namenode.servicemonitor.fullname", "hdfs-nn-servicemonitor.yaml", ""},
		{"ServiceMonitor", "hdfs-snn", "hdfs-snn", "hbase.hdfs.secondarynamenode.servicemonitor.fullname", "hdfs-snn-servicemonitor.yaml", ""},
		{"ServiceMonitor", "hdfs-dn", "hdfs-dn", "hbase.hdfs.datanode.servicemonitor.fullname", "hdfs-dn-servicemonitor.yaml", ""},
		{"ServiceMonitor", "tephra", "tephra", "hbase.tephra.servicemonitor.fullname", "tephra-servicemonitor.yaml", ""},
		{"ConfigMap", "hbase-rs", "hbase-rs-scripts", "hbase.hbase.regionserver.scripts.configmap.fullname", "hbase-rs-configmap-scripts.yaml", "hbase-rs-statefulset.yaml"},
		{"ConfigMap", "hdfs-dn", "hdfs-dn-scripts", "hbase.hdfs.datanode.scripts.configmap.fullname", "hdfs-dn-configmap-scripts.yaml", "hdfs-dn-statefulset.yaml"},
	}
}

func hbasePeripheralDefinition(helper string) *regexp.Regexp {
	return regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
}

func TestHBasePeripheralNamingSharedNamespaceInstallations(t *testing.T) {
	identities := map[string]bool{}
	for _, tenant := range []string{"a", "b"} {
		release, namespace := "installation-"+tenant, "tenant-"+tenant
		options := apiResourceNameTestOptions(map[string]string{
			"hbase.deployment.mode": "Distributed", "hbase.all.metrics.servicemonitor.enabled": "true",
			"hbase.common.servicemonitor.namespace": "shared-monitoring",
		})
		options.KubectlOptions.Namespace = namespace
		resources := helmtestutil.NewKubernetesResources(t, renderHBasePeripheralChart(t, "..", release, options, true))
		count := 0
		for name, monitor := range resources.ServiceMonitors {
			if monitor.Labels["app.kubernetes.io/name"] != "hbase" {
				continue
			}
			assert.Equal(t, release+"-hbase-"+monitor.Labels["app.kubernetes.io/component"], name)
			assert.Equal(t, "shared-monitoring", monitor.Namespace)
			assert.Equal(t, []string{namespace}, monitor.Spec.NamespaceSelector.MatchNames)
			assert.Equal(t, release, monitor.Labels["app.kubernetes.io/instance"])
			key := monitor.Namespace + "/" + name
			require.False(t, identities[key], "Shared-namespace identities must stay distinct for existing installations")
			identities[key] = true
			count++
		}
		require.Equal(t, 6, count)
	}
	require.Len(t, identities, 12)
}

// Unpack only the HBase dependency in a disposable platform chart. Helm accepts
// unpacked dependencies, so tests can edit its templates without rebuilding or
// changing the checked-out source, lockfiles or other dependency archives.
func hbasePeripheralTestChart(t *testing.T, source string) (string, string) {
	t.Helper()
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS(source)))
	child := filepath.Join(chart, "charts/hbase")
	archives, err := filepath.Glob(filepath.Join(chart, "charts/hbase-*.tgz"))
	require.NoError(t, err)
	require.Len(t, archives, 1)
	file, err := os.Open(archives[0])
	require.NoError(t, err)
	defer file.Close()
	reader, err := gzip.NewReader(file)
	require.NoError(t, err)
	defer reader.Close()
	archive := tar.NewReader(reader)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		path := filepath.Join(chart, "charts", header.Name)
		require.True(t, strings.HasPrefix(path, child+string(os.PathSeparator)), header.Name)
		if header.Typeflag == tar.TypeDir {
			require.NoError(t, os.MkdirAll(path, 0700))
		} else {
			require.Equal(t, byte(tar.TypeReg), header.Typeflag)
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
			data, err := io.ReadAll(archive)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(path, data, 0600))
		}
	}
	require.NoError(t, os.Remove(archives[0]))
	return chart, child
}

func checkHBasePeripheralNaming(t *testing.T, beforeChart, afterChart string) {
	t.Helper()
	for _, tc := range []struct {
		name, release, namespace, sizing string
		mono                             bool
		values                           map[string]string
	}{
		{name: "default", release: "suse-observability"},
		{name: "nightly"},
		{name: "second-namespace", namespace: "tenant-b"},
		{name: "root-fullname", values: map[string]string{"fullnameOverride": "customer-platform"}},
		{name: "hbase-fullname", values: map[string]string{"hbase.fullnameOverride": "customer-database"}},
		{name: "affixes", values: map[string]string{
			"global.fullnamePrefix": "global-", "global.fullnameSuffix": "-end",
			"hbase.fullnamePrefix": "customer-", "hbase.fullnameSuffix": "-database",
		}},
		{name: "long-fullname", values: map[string]string{"hbase.fullnameOverride": strings.Repeat("h", 70)}},
		{name: "already-canonical", values: map[string]string{"hbase.fullnameOverride": "suse-observability"}},
		{name: "shared-monitoring", values: map[string]string{"hbase.servicemonitor.namespace": "monitoring"}},
		{name: "common-shared-monitoring", values: map[string]string{"hbase.common.servicemonitor.namespace": "monitoring"}},
		{name: "explicit-installation-namespace", values: map[string]string{
			"hbase.common.servicemonitor.namespace": "monitoring", "hbase.servicemonitor.namespace": "observability",
		}},
		{name: "secondary-namenode", values: map[string]string{"hbase.hdfs.secondarynamenode.enabled": "true"}},
		{name: "mono", mono: true, values: map[string]string{"hbase.deployment.mode": "Mono"}},
		{name: "metrics-disabled", values: map[string]string{"hbase.all.metrics.enabled": "false"}},
		{name: "monitors-disabled", values: map[string]string{"hbase.all.metrics.servicemonitor.enabled": "false"}},
		{name: "argo-and-router-hooks", values: map[string]string{
			"deployment.compatibleWithArgoCD": "true", "stackstate.components.router.mode.status": "automatic",
		}},
		{name: "ha", sizing: "global_sizing_150_ha.yaml"},
		{name: "nonha", mono: true, sizing: "global_sizing_50_nonha.yaml"},
		{name: "custom-storage-and-monitoring", values: map[string]string{
			"hbase.hdfs.datanode.persistence.size": "17Gi", "hbase.hdfs.datanode.persistence.storageClass": "customer-storage",
			"hbase.all.image.pullSecretName": "customer-registry", "hbase.all.extraEnv.secret.TEST_KEY": "synthetic-value",
			"hbase.all.metrics.servicemonitor.additionalLabels.prometheus": "customer",
			"hbase.servicemonitor.interval":                                "37s", "hbase.servicemonitor.path": "/custom-metrics",
			"hbase.servicemonitor.annotations.customer": "preserved", "global.backup.enabled": "true",
		}},
	} {
		if tc.release == "" {
			tc.release = "nightly"
		}
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				values := componentResourceNameTestValues()
				values["hbase.deployment.mode"] = "Distributed"
				values["hbase.all.metrics.enabled"] = "true"
				values["hbase.all.metrics.servicemonitor.enabled"] = "true"
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
				beforeOutput := renderHBasePeripheralChart(t, beforeChart, tc.release, options, upgrade)
				afterOutput := renderHBasePeripheralChart(t, afterChart, tc.release, options, upgrade)
				before := helmtestutil.NewKubernetesResources(t, beforeOutput)
				after := helmtestutil.NewKubernetesResources(t, afterOutput)
				renames := map[string]string{}
				for _, resource := range hbasePeripheralInventory() {
					if resource.kind == "ServiceMonitor" {
						for name, monitor := range before.ServiceMonitors {
							if monitor.Labels["app.kubernetes.io/name"] != "hbase" || monitor.Labels["app.kubernetes.io/component"] != resource.component {
								continue
							}
							target := "suse-observability-" + resource.suffix
							if monitor.Namespace != options.KubectlOptions.Namespace {
								target = name
							}
							require.Contains(t, after.ServiceMonitors, target)
							monitor.Name = target
							assert.Equal(t, monitor, after.ServiceMonitors[target])
							matched := 0
							for _, service := range after.Services {
								if service.Namespace == monitor.Spec.NamespaceSelector.MatchNames[0] || service.Namespace == "" {
									if labels.SelectorFromSet(monitor.Spec.Selector.MatchLabels).Matches(labels.Set(service.Labels)) {
										matched++
									}
								}
							}
							assert.Greater(t, matched, 0, "Monitor must retain its Service targets")
							if name != target {
								renames["ServiceMonitor/"+name] = target
								assert.NotContains(t, after.ServiceMonitors, name)
							}
						}
						continue
					}
					target := "suse-observability-" + resource.suffix
					found := false
					for name, sts := range before.Statefulsets {
						if sts.Labels["app.kubernetes.io/name"] != "hbase" || sts.Labels["app.kubernetes.io/component"] != resource.component {
							continue
						}
						found = true
						expected := sts.DeepCopy()
						references := 0
						for i := range expected.Spec.Template.Spec.Volumes {
							ref := expected.Spec.Template.Spec.Volumes[i].ConfigMap
							if ref != nil && strings.HasSuffix(ref.Name, "-"+resource.suffix) {
								old := ref.Name
								require.Contains(t, before.ConfigMaps, old)
								require.Contains(t, after.ConfigMaps, target)
								config := before.ConfigMaps[old]
								config.Name = target
								assert.Equal(t, config, after.ConfigMaps[target], "Probe scripts and other ConfigMap fields stay equal")
								if old != target {
									renames["ConfigMap/"+old] = target
									assert.NotContains(t, after.ConfigMaps, old)
								}
								ref.Name = target
								references++
							}
						}
						require.Equal(t, 1, references)
						require.Contains(t, after.Statefulsets, name)
						assert.Equal(t, *expected, after.Statefulsets[name], "Only the probe-script volume reference changes in this StatefulSet")
						assertHBasePeripheralPatches(t, sts, after.Statefulsets[name])
					}
					require.Equal(t, !tc.mono, found, "Sizing profiles take precedence over the explicit HBase mode")
					if !found {
						assert.NotContains(t, after.ConfigMaps, target)
					}
				}
				seen := map[string]bool{}
				expected := backupNamingDocuments(t, beforeOutput, renames, seen)
				require.Len(t, seen, len(renames))
				assert.Equal(t, expected, backupNamingDocuments(t, afterOutput, nil, nil),
					"Only eight reviewed names and two probe-script volume references may change; no later step repairs consumers")
			})
		}
	}
}

func renderHBasePeripheralChart(t *testing.T, chart, release string, options *helm.Options, upgrade bool) string {
	t.Helper()
	copied := *options
	copied.Logger = logger.Discard
	copied.ValuesFiles = make([]string, len(options.ValuesFiles))
	for i, file := range options.ValuesFiles {
		path, err := filepath.Abs(file)
		require.NoError(t, err)
		copied.ValuesFiles[i] = path
	}
	args := []string{"--kube-version", "1.32.0", "--api-versions", "policy/v1", "--api-versions", "monitoring.coreos.com/v1"}
	if upgrade {
		args = append(args, "--is-upgrade")
	}
	output, err := helm.RenderTemplateE(t, &copied, chart, release, nil, args...)
	require.NoError(t, err)
	return output
}

func assertHBasePeripheralPatches(t *testing.T, before, after appsv1.StatefulSet) {
	t.Helper()
	for _, operation := range []string{"upgrade", "repeat", "rollback"} {
		old, next := before, after
		if operation == "repeat" {
			old = after
		} else if operation == "rollback" {
			old, next = after, before
		}
		current := old.DeepCopy()
		if current.Annotations == nil {
			current.Annotations = map[string]string{}
		}
		current.Annotations["customer"] = "retained"
		current.ResourceVersion = "123"
		current.Status.ReadyReplicas = 7
		original, err := json.Marshal(old)
		require.NoError(t, err)
		modified, err := json.Marshal(next)
		require.NoError(t, err)
		live, err := json.Marshal(current)
		require.NoError(t, err)
		schema, err := strategicpatch.NewPatchMetaFromStruct(appsv1.StatefulSet{})
		require.NoError(t, err)
		patch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, live, schema, true)
		require.NoError(t, err)
		var fields map[string]interface{}
		require.NoError(t, json.Unmarshal(patch, &fields))
		if spec, ok := fields["spec"].(map[string]interface{}); ok {
			for _, immutable := range []string{"selector", "serviceName", "volumeClaimTemplates"} {
				assert.NotContains(t, spec, immutable, operation)
			}
		}
		patched, err := strategicpatch.StrategicMergePatch(live, patch, appsv1.StatefulSet{})
		require.NoError(t, err)
		var actual appsv1.StatefulSet
		require.NoError(t, json.Unmarshal(patched, &actual))
		// JSON round trips canonicalize equivalent CPU quantities (3000m -> 3).
		expectedSpec, err := json.Marshal(next.Spec)
		require.NoError(t, err)
		actualSpec, err := json.Marshal(actual.Spec)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedSpec), string(actualSpec), operation)
		assert.Equal(t, current.Status, actual.Status, operation)
		assert.Equal(t, current.ResourceVersion, actual.ResourceVersion)
		assert.Equal(t, "retained", actual.Annotations["customer"])
	}
}
