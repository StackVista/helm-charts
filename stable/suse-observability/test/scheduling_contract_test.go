package test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	yamlv3 "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const releaseName = "suse-observability"

var (
	schedulingHA     = []string{"values/workload_global_ha.yaml", "values/replication_checker_enabled.yaml", "values/scheduling_all_workloads.yaml", "values/scheduling_correlate_split.yaml"}
	schedulingNonHA  = []string{"values/workload_global_nonha.yaml", "values/scheduling_all_workloads.yaml"}
	schedulingLegacy = []string{"values/full.yaml", "values/scheduling_all_workloads.yaml", "values/scheduling_correlate_split.yaml"}

	globalTolerations = []corev1.Toleration{{Key: "global-taint", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule}}
)

func nodePool(values ...string) *corev1.NodeAffinity {
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{{
				MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "node-pool", Operator: corev1.NodeSelectorOpIn, Values: values}},
			}},
		},
	}
}

var markerPodAffinity = &corev1.PodAffinity{
	PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
		Weight: 10,
		PodAffinityTerm: corev1.PodAffinityTerm{
			TopologyKey:   "local-marker",
			LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "marker"}},
		},
	}},
}

func toleration(key string) []corev1.Toleration {
	return []corev1.Toleration{{Key: key, Operator: corev1.TolerationOpExists}}
}

func with(base []string, extra ...string) []string {
	return append(append([]string{}, base...), extra...)
}

func renderScheduling(t *testing.T, valuesFiles ...string) (string, helmtestutil.KubernetesResources) {
	output := helmtestutil.RenderHelmTemplate(t, releaseName, valuesFiles...)
	return output, helmtestutil.NewKubernetesResources(t, output)
}

func podSpecs(resources helmtestutil.KubernetesResources) map[string]corev1.PodSpec {
	specs := map[string]corev1.PodSpec{}
	for _, ref := range helmtestutil.CollectPodSpecs(resources) {
		specs[ref.Kind+"/"+ref.Name] = ref.Spec
	}
	return specs
}

func podSpec(t *testing.T, specs map[string]corev1.PodSpec, kind, namePrefix string) corev1.PodSpec {
	t.Helper()
	exact := kind + "/" + releaseName + "-" + namePrefix
	if spec, ok := specs[exact]; ok {
		return spec
	}
	var found []string
	for key := range specs {
		if strings.HasPrefix(key, exact) {
			found = append(found, key)
		}
	}
	require.Len(t, found, 1, "expected exactly one %s with prefix %s", kind, exact)
	return specs[found[0]]
}

func affinityOf(spec corev1.PodSpec) corev1.Affinity {
	if spec.Affinity == nil {
		return corev1.Affinity{}
	}
	return *spec.Affinity
}

// backupRestoreJobSpecs returns the pod specs of the restore Jobs that the chart
// renders into ConfigMaps rather than as cluster resources.
func backupRestoreJobSpecs(t *testing.T, resources helmtestutil.KubernetesResources) map[string]corev1.PodSpec {
	specs := map[string]corev1.PodSpec{}
	for name, job := range testJobsFromBackupRestoreScriptsConfigMap(t, &resources) {
		specs["ConfigMapJob/"+name] = job.Spec.Template.Spec
	}

	configMap, ok := resources.ConfigMaps[releaseName+"-backup-config"]
	require.True(t, ok, "backup-config ConfigMap should exist")
	type restoreJob struct {
		Restore struct {
			Job struct {
				Tolerations []corev1.Toleration `json:"tolerations"`
				Affinity    *corev1.Affinity    `json:"affinity"`
			} `json:"job"`
		} `json:"restore"`
	}
	var config struct {
		Stackgraph      restoreJob `json:"stackgraph"`
		Settings        restoreJob `json:"settings"`
		VictoriaMetrics restoreJob `json:"victoriaMetrics"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(configMap.Data["config"]), &config))
	for section, cfg := range map[string]restoreJob{"stackgraph": config.Stackgraph, "settings": config.Settings, "victoriaMetrics": config.VictoriaMetrics} {
		job := cfg.Restore.Job
		specs["BackupConfig/"+section] = corev1.PodSpec{Tolerations: job.Tolerations, Affinity: job.Affinity}
	}
	return specs
}

func TestSchedulingGlobalAppliesToEveryWorkload(t *testing.T) {
	for name, values := range map[string][]string{"ha": schedulingHA, "nonha": schedulingNonHA, "legacy": schedulingLegacy} {
		t.Run(name, func(t *testing.T) {
			_, resources := renderScheduling(t, with(values, "values/scheduling_global.yaml")...)
			specs := podSpecs(resources)
			require.NotEmpty(t, specs)
			for key, spec := range backupRestoreJobSpecs(t, resources) {
				specs[key] = spec
			}
			for key, spec := range specs {
				assert.Equal(t, globalTolerations, spec.Tolerations, "tolerations of %s", key)
				assert.Equal(t, nodePool("observability"), affinityOf(spec).NodeAffinity, "nodeAffinity of %s", key)
			}
		})
	}
}

func TestSchedulingNothingConfiguredRendersNoTolerations(t *testing.T) {
	for name, values := range map[string][]string{"ha": schedulingHA, "nonha": schedulingNonHA} {
		t.Run(name, func(t *testing.T) {
			_, resources := renderScheduling(t, values...)
			for key, spec := range podSpecs(resources) {
				assert.Empty(t, spec.Tolerations, "tolerations of %s", key)
				assert.Nil(t, affinityOf(spec).NodeAffinity, "nodeAffinity of %s", key)
			}
		})
	}
}

func TestSchedulingMostSpecificLayerWins(t *testing.T) {
	_, resources := renderScheduling(t, with(schedulingHA, "values/scheduling_global.yaml", "values/scheduling_component_overrides.yaml")...)
	specs := podSpecs(resources)

	cases := []struct {
		kind, name   string
		tolerations  []corev1.Toleration
		nodeAffinity *corev1.NodeAffinity
	}{
		{"Deployment", "api", toleration("api-taint"), nodePool("api")},
		{"Deployment", "ui", toleration("all-taint"), nodePool("all")},
		{"Deployment", "router", toleration("all-taint"), nodePool("all")},
		{"Deployment", "correlate-connection", toleration("connection-taint"), nodePool("all")},
		{"Deployment", "correlate-http-tracing", toleration("correlate-taint"), nodePool("all")},
		{"Deployment", "receiver-logs", toleration("logs-taint"), nodePool("all")},
		{"Deployment", "receiver-base", toleration("receiver-taint"), nodePool("all")},
		{"Job", "set-active-", toleration("router-mode-taint"), nodePool("all")},
		{"Job", "set-maintenance-", toleration("router-mode-taint"), nodePool("all")},
		{"CronJob", "backup-sg", toleration("backup-taint"), nodePool("all")},
		{"StatefulSet", "kafka", toleration("kafka-taint"), nodePool("observability")},
		{"StatefulSet", "zookeeper", globalTolerations, nodePool("observability")},
		{"StatefulSet", "hbase-hbase-master", toleration("hbase-master-taint"), nodePool("observability")},
		{"StatefulSet", "hbase-hbase-rs", toleration("hbase-all-taint"), nodePool("observability")},
		{"CronJob", "clickhouse-full-backup", toleration("clickhouse-backup-taint"), nodePool("observability")},
		{"StatefulSet", "clickhouse-shard0", globalTolerations, nodePool("observability")},
	}
	for _, c := range cases {
		t.Run(c.kind+"/"+c.name, func(t *testing.T) {
			spec := podSpec(t, specs, c.kind, c.name)
			assert.Equal(t, c.tolerations, spec.Tolerations)
			assert.Equal(t, c.nodeAffinity, affinityOf(spec).NodeAffinity)
		})
	}

	backup := backupRestoreJobSpecs(t, resources)
	assert.Equal(t, toleration("backup-taint"), backup["BackupConfig/stackgraph"].Tolerations)
	assert.Equal(t, toleration("backup-taint"), backup["ConfigMapJob/job-stackgraph-restore-backup.yaml"].Tolerations)
}

func TestSchedulingPartialOverridesKeepInheritedFields(t *testing.T) {
	_, resources := renderScheduling(t, with(schedulingHA, "values/scheduling_global.yaml", "values/scheduling_partial_overrides.yaml")...)
	specs := podSpecs(resources)

	t.Run("app podAntiAffinity keeps global nodeAffinity", func(t *testing.T) {
		affinity := affinityOf(podSpec(t, specs, "Deployment", "api"))
		assert.Equal(t, nodePool("observability"), affinity.NodeAffinity)
		require.NotNil(t, affinity.PodAntiAffinity)
		assert.Equal(t, "local-marker", affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey)
	})

	infra := []struct{ kind, name string }{
		{"StatefulSet", "kafka"},
		{"StatefulSet", "zookeeper"},
		{"StatefulSet", "victoria-metrics-0"},
		{"StatefulSet", "victoria-metrics-1"},
		{"StatefulSet", "otel-collector"},
		{"StatefulSet", "hbase-hbase-rs"},
		{"StatefulSet", "hbase-hdfs-dn"},
		{"Deployment", "anomaly-detection-spotlight-worker"},
		{"Deployment", "anomaly-detection-spotlight-manager"},
		{"Deployment", "rbac-agent"},
		{"Deployment", "kafkaup-operator-kafkaup"},
		{"Deployment", "prometheus-elasticsearch-exporter"},
	}
	for _, w := range infra {
		t.Run(w.kind+"/"+w.name+" local podAffinity keeps nodeAffinity and HA spreading", func(t *testing.T) {
			affinity := affinityOf(podSpec(t, specs, w.kind, w.name))
			assert.Equal(t, markerPodAffinity, affinity.PodAffinity)
			assert.Equal(t, nodePool("observability"), affinity.NodeAffinity)
			assert.NotNil(t, affinity.PodAntiAffinity)
		})
	}

	for _, w := range []struct{ kind, name, pool string }{
		{"StatefulSet", "clickhouse-shard0", "clickhouse"},
		{"StatefulSet", "hbase-hbase-master", "hbase-master"},
	} {
		t.Run(w.kind+"/"+w.name+" local nodeAffinity keeps HA spreading", func(t *testing.T) {
			affinity := affinityOf(podSpec(t, specs, w.kind, w.name))
			assert.Equal(t, nodePool(w.pool), affinity.NodeAffinity)
			assert.NotNil(t, affinity.PodAntiAffinity)
		})
	}

	t.Run("hbase all and component layers combine per field", func(t *testing.T) {
		affinity := affinityOf(podSpec(t, specs, "StatefulSet", "hbase-hbase-master"))
		assert.Equal(t, markerPodAffinity, affinity.PodAffinity)
	})
}

func TestSchedulingPartialOverridesNonHA(t *testing.T) {
	_, resources := renderScheduling(t, with(schedulingNonHA, "values/scheduling_global.yaml", "values/scheduling_partial_overrides.yaml")...)
	specs := podSpecs(resources)
	for _, w := range []struct{ kind, name string }{
		{"StatefulSet", "kafka"},
		{"StatefulSet", "zookeeper"},
		{"StatefulSet", "hbase-stackgraph"},
	} {
		t.Run(w.kind+"/"+w.name, func(t *testing.T) {
			affinity := affinityOf(podSpec(t, specs, w.kind, w.name))
			assert.Equal(t, markerPodAffinity, affinity.PodAffinity)
			assert.Equal(t, nodePool("observability"), affinity.NodeAffinity)
		})
	}
}

func TestSchedulingEmptyValuesClearInheritedValues(t *testing.T) {
	_, resources := renderScheduling(t, with(schedulingHA, "values/scheduling_global.yaml", "values/scheduling_clear.yaml")...)
	specs := podSpecs(resources)

	api := podSpec(t, specs, "Deployment", "api")
	assert.Empty(t, api.Tolerations)
	assert.Nil(t, affinityOf(api).NodeAffinity)

	assert.Empty(t, podSpec(t, specs, "StatefulSet", "kafka").Tolerations)

	clickhouse := affinityOf(podSpec(t, specs, "StatefulSet", "clickhouse-shard0"))
	assert.Nil(t, clickhouse.PodAntiAffinity)
	assert.Equal(t, nodePool("observability"), clickhouse.NodeAffinity)

	assert.Equal(t, globalTolerations, podSpec(t, specs, "Deployment", "ui").Tolerations)
}

func TestSchedulingSoftAntiAffinityIsHonoured(t *testing.T) {
	_, resources := renderScheduling(t, with(schedulingHA, "values/scheduling_soft_antiaffinity.yaml")...)
	specs := podSpecs(resources)
	for _, w := range []struct{ kind, name string }{
		{"StatefulSet", "kafka"},
		{"StatefulSet", "elasticsearch-master"},
		{"StatefulSet", "clickhouse-shard0"},
	} {
		t.Run(w.kind+"/"+w.name, func(t *testing.T) {
			affinity := affinityOf(podSpec(t, specs, w.kind, w.name))
			require.NotNil(t, affinity.PodAntiAffinity)
			assert.Empty(t, affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
			assert.NotEmpty(t, affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
		})
	}
}

func TestSchedulingRendersNoDuplicateKeys(t *testing.T) {
	for name, values := range map[string][]string{
		"ha-overrides": with(schedulingHA, "values/scheduling_global.yaml", "values/scheduling_component_overrides.yaml"),
		"ha-partial":   with(schedulingHA, "values/scheduling_global.yaml", "values/scheduling_partial_overrides.yaml"),
		"nonha":        with(schedulingNonHA, "values/scheduling_global.yaml", "values/scheduling_partial_overrides.yaml"),
	} {
		t.Run(name, func(t *testing.T) {
			output, _ := renderScheduling(t, values...)
			decoder := yamlv3.NewDecoder(bytes.NewBufferString(output))
			for {
				var doc map[string]any
				err := decoder.Decode(&doc)
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)
			}
		})
	}
}

func TestSchedulingHdfsNameNodesAvoidEachOther(t *testing.T) {
	_, resources := renderScheduling(t, schedulingHA...)
	specs := podSpecs(resources)
	for name, other := range map[string]string{"hbase-hdfs-nn": "hdfs-snn", "hbase-hdfs-snn": "hdfs-nn"} {
		required := affinityOf(podSpec(t, specs, "StatefulSet", name)).PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		require.Len(t, required, 1, name)
		assert.Equal(t, map[string]string{"app.kubernetes.io/component": other}, required[0].LabelSelector.MatchLabels, name)
	}
}
