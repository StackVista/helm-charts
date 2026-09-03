package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

const globalStorageClass = "global"

func TestStorageClassOverrides(t *testing.T) {
	distributed := map[string]string{
		"pvc/spotlight-artifacts-volume-claim":                            "anomaly-detection",
		"pvc/suse-observability-api-txlog":                                "transaction-logs",
		"pvc/suse-observability-authorization-sync-txlog":                 "transaction-logs",
		"pvc/suse-observability-backup-settings-data":                     "backup-settings",
		"pvc/suse-observability-backup-stackgraph-tmp-data":               "backup-stackgraph-scheduled",
		"pvc/suse-observability-backup-stackgraph-v2-tmp-data":            "backup-stackgraph-v2",
		"pvc/suse-observability-checks-tmp":                               "checks-tmp",
		"pvc/suse-observability-checks-txlog":                             "transaction-logs",
		"pvc/suse-observability-health-sync-tmp":                          "health-sync-tmp",
		"pvc/suse-observability-health-sync-txlog":                        "transaction-logs",
		"pvc/suse-observability-minio":                                    "backup-backend",
		"pvc/suse-observability-notification-txlog":                       "transaction-logs",
		"pvc/suse-observability-settings-backup-data":                     "backup-configuration",
		"pvc/suse-observability-stackpacks":                               "stackpacks",
		"pvc/suse-observability-state-tmp":                                "state-tmp",
		"pvc/suse-observability-state-txlog":                              "transaction-logs",
		"pvc/suse-observability-sync-tmp":                                 "sync-tmp",
		"pvc/suse-observability-sync-txlog":                               "transaction-logs",
		"statefulset/suse-observability-ai-assistant/data":                "ai-assistant",
		"statefulset/suse-observability-clickhouse-shard0/data":           "clickhouse",
		"statefulset/suse-observability-elasticsearch-master/data":        "elasticsearch",
		"statefulset/suse-observability-hbase-hdfs-dn/data":               "hbase-datanode",
		"statefulset/suse-observability-hbase-hdfs-nn/data":               "hbase-namenode",
		"statefulset/suse-observability-hbase-hdfs-snn/data":              "hbase-secondarynamenode",
		"statefulset/suse-observability-kafka/data":                       "kafka-data",
		"statefulset/suse-observability-kafka/logs":                       "kafka-logs",
		"statefulset/suse-observability-otel-collector/data":              "opentelemetry",
		"statefulset/suse-observability-victoria-metrics-0/server-volume": "victoria-metrics-0",
		"statefulset/suse-observability-victoria-metrics-1/server-volume": "victoria-metrics-1",
		"statefulset/suse-observability-vmagent/tmpdata":                  "vmagent",
		"statefulset/suse-observability-workload-observer/data":           "workload-observer",
		"statefulset/suse-observability-zookeeper/data":                   "zookeeper",
	}

	mono := map[string]string{
		"pvc/spotlight-artifacts-volume-claim":                            "anomaly-detection",
		"pvc/suse-observability-backup-settings-data":                     "backup-settings",
		"pvc/suse-observability-backup-stackgraph-tmp-data":               "backup-stackgraph-scheduled",
		"pvc/suse-observability-backup-stackgraph-v2-tmp-data":            "backup-stackgraph-v2",
		"pvc/suse-observability-minio":                                    "backup-backend",
		"pvc/suse-observability-settings-backup-data":                     "backup-configuration",
		"pvc/suse-observability-stackpacks":                               "stackpacks",
		"pvc/suse-observability-stackpacks-local":                         "stackpacks-local",
		"statefulset/suse-observability-ai-assistant/data":                "ai-assistant",
		"statefulset/suse-observability-clickhouse-shard0/data":           "clickhouse",
		"statefulset/suse-observability-elasticsearch-master/data":        "elasticsearch",
		"statefulset/suse-observability-hbase-stackgraph/data":            "hbase-stackgraph",
		"statefulset/suse-observability-hbase-tephra-mono/snapshot":       "hbase-tephra",
		"statefulset/suse-observability-kafka/data":                       "kafka-data",
		"statefulset/suse-observability-kafka/logs":                       "kafka-logs",
		"statefulset/suse-observability-otel-collector/data":              "opentelemetry",
		"statefulset/suse-observability-victoria-metrics-0/server-volume": "victoria-metrics-0",
		"statefulset/suse-observability-vmagent/tmpdata":                  "vmagent",
		"statefulset/suse-observability-workload-observer/data":           "workload-observer",
		"statefulset/suse-observability-zookeeper/data":                   "zookeeper",
	}

	victoriaMetricsPvc := cloneStorageClasses(distributed)
	delete(victoriaMetricsPvc, "statefulset/suse-observability-victoria-metrics-0/server-volume")
	victoriaMetricsPvc["pvc/suse-observability-victoria-metrics-0"] = "victoria-metrics-0"

	testCases := []struct {
		name        string
		valuesFiles []string
		setValues   map[string]string
		expected    map[string]string
	}{
		{
			name:        "distributed",
			valuesFiles: []string{"values/full.yaml", "values/storage_class_overrides.yaml"},
			expected:    distributed,
		},
		{
			name:        "mono",
			valuesFiles: []string{"values/global_sizing_10_nonha.yaml", "values/storage_class_overrides.yaml"},
			setValues: map[string]string{
				"global.backup.enabled":      "true",
				"victoria-metrics-1.enabled": "true",
			},
			expected: mono,
		},
		{
			name:        "victoria metrics pvc",
			valuesFiles: []string{"values/full.yaml", "values/storage_class_overrides.yaml"},
			setValues: map[string]string{
				"victoria-metrics-0.server.statefulSet.enabled": "false",
			},
			expected: victoriaMetricsPvc,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
				ValuesFiles: tc.valuesFiles,
				SetValues:   tc.setValues,
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			actual := storageClasses(resources)

			assert.Equal(t, tc.expected, actual)
			for claim, storageClass := range actual {
				assert.NotEqual(t, globalStorageClass, storageClass, "%s should override global.storageClass", claim)
			}

			backupConfig, ok := resources.ConfigMaps["suse-observability-backup-config"]
			require.True(t, ok)
			assert.Contains(t, backupConfig.Data["config"], "storageClassName: backup-stackgraph-restore")

			restoreScripts, ok := resources.ConfigMaps["suse-observability-backup-restore-scripts"]
			require.True(t, ok)
			assert.Contains(t, restoreScripts.Data["pvc-stackgraph-restore-backup.yaml"], "storageClassName: backup-stackgraph-restore")
		})
	}
}

func storageClasses(resources helmtestutil.KubernetesResources) map[string]string {
	classes := make(map[string]string)
	for name, pvc := range resources.PersistentVolumeClaims {
		classes["pvc/"+name] = storageClassName(pvc.Spec.StorageClassName)
	}
	for statefulSetName, statefulSet := range resources.Statefulsets {
		for _, volumeClaim := range statefulSet.Spec.VolumeClaimTemplates {
			classes["statefulset/"+statefulSetName+"/"+volumeClaim.Name] = storageClassName(volumeClaim.Spec.StorageClassName)
		}
	}
	return classes
}

func storageClassName(storageClass *string) string {
	if storageClass == nil {
		return "<unset>"
	}
	return *storageClass
}

func cloneStorageClasses(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	for claim, storageClass := range source {
		clone[claim] = storageClass
	}
	return clone
}
