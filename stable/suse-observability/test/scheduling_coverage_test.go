package test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// schedulingWorkload maps rendered workloads to the values path of their most specific
// scheduling layer. Every workload rendered by schedulingHA or schedulingNonHA must match
// exactly one entry, so a new workload fails TestSchedulingCoverageIsComplete until added.
type schedulingWorkload struct {
	match string
	layer string
	// elasticsearch exposes nodeAffinity directly instead of an affinity block.
	elasticsearch bool
}

const app = "stackstate.components."

var schedulingWorkloads = []schedulingWorkload{
	{match: `Deployment/suse-observability-api`, layer: app + "api"},
	{match: `Deployment/suse-observability-authorization-sync`, layer: app + "authorizationSync"},
	{match: `Deployment/suse-observability-checks`, layer: app + "checks"},
	{match: `Deployment/suse-observability-correlate`, layer: app + "correlate"},
	{match: `Deployment/suse-observability-correlate-aggregator`, layer: app + "correlate.split.aggregator"},
	{match: `Deployment/suse-observability-correlate-connection`, layer: app + "correlate.split.connection"},
	{match: `Deployment/suse-observability-correlate-http-tracing`, layer: app + "correlate.split.httpTracing"},
	{match: `Deployment/suse-observability-e2es`, layer: app + "e2es"},
	{match: `Deployment/suse-observability-health-sync`, layer: app + "healthSync"},
	{match: `Deployment/suse-observability-initializer`, layer: app + "initializer"},
	{match: `Deployment/suse-observability-mcp`, layer: app + "mcp"},
	{match: `Deployment/suse-observability-notification`, layer: app + "notification"},
	{match: `Deployment/suse-observability-receiver`, layer: app + "receiver"},
	{match: `Deployment/suse-observability-receiver-base`, layer: app + "receiver.split.base"},
	{match: `Deployment/suse-observability-receiver-logs`, layer: app + "receiver.split.logs"},
	{match: `Deployment/suse-observability-receiver-process-agent`, layer: app + "receiver.split.processAgent"},
	{match: `Deployment/suse-observability-replication-checker`, layer: app + "replicationChecker"},
	{match: `Deployment/suse-observability-router`, layer: app + "router"},
	{match: `Deployment/suse-observability-server`, layer: app + "server"},
	{match: `Deployment/suse-observability-slicing`, layer: app + "slicing"},
	{match: `Deployment/suse-observability-state`, layer: app + "state"},
	{match: `Deployment/suse-observability-sync`, layer: app + "sync"},
	{match: `Deployment/suse-observability-ui`, layer: app + "ui"},
	{match: `StatefulSet/suse-observability-ai-assistant`, layer: app + "aiAssistant"},
	{match: `StatefulSet/suse-observability-vmagent`, layer: app + "vmagent"},
	{match: `StatefulSet/suse-observability-workload-observer`, layer: app + "workloadObserver"},
	{match: `Deployment/suse-observability-s3proxy`, layer: "s3proxy"},

	{match: `Job/suse-observability-set-active`, layer: app + "router.mode"},
	{match: `Job/suse-observability-set-maintenance`, layer: app + "router.mode"},
	{match: `Job/suse-observability-topic-create`, layer: app + "kafkaTopicCreate"},
	{match: `Job/suse-observability-ch-clean`, layer: app + "clickhouseCleanup"},
	{match: `Job/suse-observability-backup-init`, layer: app + "backup"},
	{match: `Job/suse-observability-init-pvc`, layer: app + "configurationBackup"},
	{match: `CronJob/suse-observability-backup-(init|sg|sg-v2)`, layer: app + "backup"},
	{match: `CronJob/suse-observability-backup-conf`, layer: app + "configurationBackup"},
	{match: `ConfigMapJob/job-(elasticsearch|stackgraph|victoria-metrics)-.+`, layer: app + "backup"},
	{match: `ConfigMapJob/job-configuration-.+`, layer: app + "configurationBackup"},
	{match: `BackupConfig/(stackgraph|settings|victoriaMetrics)`, layer: app + "backup"},

	{match: `StatefulSet/suse-observability-kafka`, layer: "kafka"},
	{match: `Deployment/suse-observability-kafka-exporter`, layer: "kafka.metrics.kafka"},
	{match: `Job/suse-observability-kafka-provisioning`, layer: "kafka.provisioning"},
	{match: `StatefulSet/suse-observability-zookeeper`, layer: "zookeeper"},
	{match: `StatefulSet/suse-observability-clickhouse-shard0`, layer: "clickhouse"},
	{match: `CronJob/suse-observability-clickhouse-(full|incremental)-backup`, layer: "clickhouse.backup"},
	{match: `StatefulSet/suse-observability-victoria-metrics-0`, layer: "victoria-metrics-0.server"},
	{match: `StatefulSet/suse-observability-victoria-metrics-1`, layer: "victoria-metrics-1.server"},
	{match: `StatefulSet/suse-observability-otel-collector`, layer: "opentelemetry-collector"},
	{match: `StatefulSet/suse-observability-elasticsearch-master`, layer: "elasticsearch", elasticsearch: true},
	{match: `Pod/suse-observability-test`, layer: "elasticsearch", elasticsearch: true},
	{match: `Deployment/suse-observability-prometheus-elasticsearch-exporter`, layer: "elasticsearch.prometheus-elasticsearch-exporter"},
	{match: `Deployment/suse-observability-anomaly-detection-spotlight-worker`, layer: "anomaly-detection"},
	{match: `Deployment/suse-observability-anomaly-detection-spotlight-manager`, layer: "anomaly-detection.manager"},
	{match: `Deployment/suse-observability-rbac-agent`, layer: "kubernetes-rbac-agent.containers.rbacAgent"},
	{match: `Deployment/suse-observability-kafkaup-operator-kafkaup`, layer: "kafkaup-operator"},
	{match: `StatefulSet/suse-observability-hbase-hbase-master`, layer: "hbase.hbase.master"},
	{match: `StatefulSet/suse-observability-hbase-hbase-rs`, layer: "hbase.hbase.regionserver"},
	{match: `StatefulSet/suse-observability-hbase-hdfs-dn`, layer: "hbase.hdfs.datanode"},
	{match: `StatefulSet/suse-observability-hbase-hdfs-nn`, layer: "hbase.hdfs.namenode"},
	{match: `StatefulSet/suse-observability-hbase-hdfs-snn`, layer: "hbase.hdfs.secondarynamenode"},
	{match: `StatefulSet/suse-observability-hbase-tephra(-mono)?`, layer: "hbase.tephra"},
	{match: `StatefulSet/suse-observability-hbase-stackgraph`, layer: "hbase.stackgraph"},
	{match: `Deployment/suse-observability-hbase-console`, layer: "hbase.console"},
	{match: `CronJob/suse-observability-hbase-integrity`, layer: "hbase.console"},
}

func layerPodAffinity(layer string) *corev1.PodAffinity {
	return &corev1.PodAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
			Weight: 10,
			PodAffinityTerm: corev1.PodAffinityTerm{
				TopologyKey:   layer,
				LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"layer": layer}},
			},
		}},
	}
}

// allSchedulingSpecs returns every pod spec the chart renders, including restore Jobs
// that are only embedded in ConfigMaps.
func allSchedulingSpecs(t *testing.T, valuesFiles ...string) map[string]corev1.PodSpec {
	_, resources := renderScheduling(t, valuesFiles...)
	specs := map[string]corev1.PodSpec{}
	for key, spec := range podSpecs(resources) {
		key = generatedSuffix.ReplaceAllString(key, "$1")
		require.NotContains(t, specs, key)
		specs[key] = spec
	}
	for key, spec := range backupRestoreJobSpecs(t, resources) {
		specs[key] = spec
	}
	return specs
}

// generatedSuffix strips timestamp and random name parts so renders can be compared.
var generatedSuffix = regexp.MustCompile(`-?\d{2}t\d{6}$|-[a-z0-9]{5}(-test)$`)

func matchingWorkloads(key string) []schedulingWorkload {
	var matches []schedulingWorkload
	for _, w := range schedulingWorkloads {
		if regexp.MustCompile(`^` + w.match + `$`).MatchString(key) {
			matches = append(matches, w)
		}
	}
	return matches
}

func workloadFor(t *testing.T, key string) schedulingWorkload {
	t.Helper()
	matches := matchingWorkloads(key)
	require.Len(t, matches, 1, "%s must match exactly one schedulingWorkloads entry", key)
	return matches[0]
}

func setValuePath(values map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	for _, part := range parts[:len(parts)-1] {
		next, ok := values[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			values[part] = next
		}
		values = next
	}
	values[parts[len(parts)-1]] = value
}

// layerValuesFile writes a values file that sets, for every workload layer, the values
// returned by fn (keyed relative to the layer).
func layerValuesFile(t *testing.T, fn func(w schedulingWorkload) map[string]any) string {
	values := map[string]any{}
	for _, w := range schedulingWorkloads {
		for key, value := range fn(w) {
			setValuePath(values, w.layer+"."+key, value)
		}
	}
	out, err := yaml.Marshal(values)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "layers.yaml")
	require.NoError(t, os.WriteFile(path, out, 0600))
	return path
}

func nodeAffinityValues(w schedulingWorkload, nodeAffinity any) map[string]any {
	if w.elasticsearch {
		return map[string]any{"nodeAffinity": nodeAffinity}
	}
	return map[string]any{"affinity.nodeAffinity": nodeAffinity}
}

func merged(maps ...map[string]any) map[string]any {
	result := map[string]any{}
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

var schedulingProfiles = map[string][]string{"ha": schedulingHA, "nonha": schedulingNonHA}

func TestSchedulingCoverageIsComplete(t *testing.T) {
	used := map[string]bool{}
	for name, values := range schedulingProfiles {
		t.Run(name, func(t *testing.T) {
			specs := allSchedulingSpecs(t, values...)
			keys := make([]string, 0, len(specs))
			for key := range specs {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				matches := matchingWorkloads(key)
				if assert.Len(t, matches, 1, "%s must match exactly one schedulingWorkloads entry", key) {
					used[matches[0].match] = true
				}
			}
		})
	}
	for _, w := range schedulingWorkloads {
		assert.True(t, used[w.match], "schedulingWorkloads entry %s matches no rendered workload", w.match)
	}
}

func TestSchedulingOwnLayerOverridesEveryWorkload(t *testing.T) {
	layers := layerValuesFile(t, func(w schedulingWorkload) map[string]any {
		return merged(
			map[string]any{"tolerations": toleration(w.layer)},
			nodeAffinityValues(w, nodePool(w.layer)),
		)
	})
	for name, values := range schedulingProfiles {
		t.Run(name, func(t *testing.T) {
			for key, spec := range allSchedulingSpecs(t, with(values, "values/scheduling_global.yaml", layers)...) {
				w := workloadFor(t, key)
				assert.Equal(t, toleration(w.layer), spec.Tolerations, "tolerations of %s", key)
				assert.Equal(t, nodePool(w.layer), affinityOf(spec).NodeAffinity, "nodeAffinity of %s", key)
			}
		})
	}
}

func TestSchedulingOwnLayerPartialOverrideKeepsEveryOtherField(t *testing.T) {
	layers := layerValuesFile(t, func(w schedulingWorkload) map[string]any {
		if w.elasticsearch {
			return nodeAffinityValues(w, nodePool(w.layer))
		}
		return map[string]any{"affinity.podAffinity": layerPodAffinity(w.layer)}
	})
	for name, values := range schedulingProfiles {
		t.Run(name, func(t *testing.T) {
			baseline := allSchedulingSpecs(t, with(values, "values/scheduling_global.yaml")...)
			partial := allSchedulingSpecs(t, with(values, "values/scheduling_global.yaml", layers)...)
			require.Equal(t, len(baseline), len(partial))
			for key, spec := range partial {
				w := workloadFor(t, key)
				expected := affinityOf(baseline[key])
				if w.elasticsearch {
					expected.NodeAffinity = nodePool(w.layer)
				} else {
					expected.PodAffinity = layerPodAffinity(w.layer)
				}
				assert.Equal(t, expected, affinityOf(spec), "affinity of %s", key)
				assert.Equal(t, baseline[key].Tolerations, spec.Tolerations, "tolerations of %s", key)
			}
		})
	}
}

func TestSchedulingOwnLayerClearsEveryWorkload(t *testing.T) {
	layers := layerValuesFile(t, func(w schedulingWorkload) map[string]any {
		return merged(
			map[string]any{"tolerations": []any{}},
			nodeAffinityValues(w, map[string]any{}),
		)
	})
	for name, values := range schedulingProfiles {
		t.Run(name, func(t *testing.T) {
			for key, spec := range allSchedulingSpecs(t, with(values, "values/scheduling_global.yaml", layers)...) {
				assert.Empty(t, spec.Tolerations, "tolerations of %s", key)
				assert.Nil(t, affinityOf(spec).NodeAffinity, "nodeAffinity of %s", key)
			}
		})
	}
}

// templatedLayers lists the scheduling values that rendered template expressions before
// the shared resolvers, so the resolvers must keep rendering them.
var templatedLayers = map[string]struct {
	key                     string
	tolerations, affinities bool
}{
	"kafka":               {`StatefulSet/suse-observability-kafka`, true, true},
	"kafka.metrics.kafka": {`Deployment/suse-observability-kafka-exporter`, true, true},
	"kafka.provisioning":  {`Job/suse-observability-kafka-provisioning`, true, false},
	"zookeeper":           {`StatefulSet/suse-observability-zookeeper`, true, true},
	"clickhouse":          {`StatefulSet/suse-observability-clickhouse-shard0`, true, true},
	"kafkaup-operator":    {`Deployment/suse-observability-kafkaup-operator-kafkaup`, false, true},
	"elasticsearch.prometheus-elasticsearch-exporter": {`Deployment/suse-observability-prometheus-elasticsearch-exporter`, false, true},
}

func TestSchedulingTemplatedValuesAreRendered(t *testing.T) {
	values := map[string]any{}
	for layer, l := range templatedLayers {
		if l.tolerations {
			setValuePath(values, layer+".tolerations", toleration("{{ .Release.Name }}-"+layer))
		}
		if l.affinities {
			setValuePath(values, layer+".affinity.nodeAffinity", nodePool("{{ .Release.Name }}-"+layer))
		}
	}
	out, err := yaml.Marshal(values)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "templated.yaml")
	require.NoError(t, os.WriteFile(path, out, 0600))

	for name, profile := range schedulingProfiles {
		t.Run(name, func(t *testing.T) {
			specs := allSchedulingSpecs(t, with(profile, path)...)
			for layer, l := range templatedLayers {
				spec, ok := specs[l.key]
				require.True(t, ok, "%s should be rendered", l.key)
				if l.tolerations {
					assert.Equal(t, toleration(releaseName+"-"+layer), spec.Tolerations, "tolerations of %s", l.key)
				}
				if l.affinities {
					assert.Equal(t, nodePool(releaseName+"-"+layer), affinityOf(spec).NodeAffinity, "nodeAffinity of %s", l.key)
				}
			}
		})
	}
}
