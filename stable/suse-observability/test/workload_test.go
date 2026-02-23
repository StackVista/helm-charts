package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// expectedWorkloadsHA defines ALL expected workloads when in HA mode (from ha_workload file)
var expectedWorkloadsHA = []string{
	// Deployments
	"Deployment/suse-observability-anomaly-detection-spotlight-manager",
	"Deployment/suse-observability-anomaly-detection-spotlight-worker",
	"Deployment/suse-observability-prometheus-elasticsearch-exporter",
	"Deployment/suse-observability-hbase-console",
	"Deployment/suse-observability-kafkaup-operator-kafkaup",
	"Deployment/suse-observability-rbac-agent",
	"Deployment/suse-observability-minio",
	"Deployment/suse-observability-api",
	"Deployment/suse-observability-authorization-sync",
	"Deployment/suse-observability-checks",
	"Deployment/suse-observability-correlate",
	"Deployment/suse-observability-health-sync",
	"Deployment/suse-observability-initializer",
	"Deployment/suse-observability-e2es",
	"Deployment/suse-observability-notification",
	"Deployment/suse-observability-receiver-base",
	"Deployment/suse-observability-receiver-logs",
	"Deployment/suse-observability-receiver-process-agent",
	"Deployment/suse-observability-router",
	"Deployment/suse-observability-slicing",
	"Deployment/suse-observability-state",
	"Deployment/suse-observability-sync",
	"Deployment/suse-observability-ui",
	// StatefulSets
	"StatefulSet/suse-observability-clickhouse-shard0",
	"StatefulSet/suse-observability-elasticsearch-master",
	"StatefulSet/suse-observability-hbase-hbase-master",
	"StatefulSet/suse-observability-hbase-hdfs-nn",
	"StatefulSet/suse-observability-hbase-hbase-rs",
	"StatefulSet/suse-observability-hbase-hdfs-dn",
	"StatefulSet/suse-observability-hbase-hdfs-snn",
	"StatefulSet/suse-observability-hbase-tephra",
	"StatefulSet/suse-observability-kafka",
	"StatefulSet/suse-observability-otel-collector",
	"StatefulSet/suse-observability-victoria-metrics-0",
	"StatefulSet/suse-observability-victoria-metrics-1",
	"StatefulSet/suse-observability-zookeeper",
	"StatefulSet/suse-observability-vmagent",
	"StatefulSet/suse-observability-workload-observer",
	// Jobs (dynamic names will be handled with pattern matching)
	"Job/suse-observability-backup-conf-*",
	"Job/suse-observability-backup-init-*",
	"Job/suse-observability-topic-create-*",
	"Job/suse-observability-ch-clean*",
	// CronJobs
	"CronJob/suse-observability-backup-sg",
	"CronJob/suse-observability-backup-conf",
	"CronJob/suse-observability-backup-init",
	"CronJob/suse-observability-clickhouse-incremental-backup",
	"CronJob/suse-observability-clickhouse-full-backup",
}

// expectedWorkloadsNonHA defines ALL expected workloads when in Non-HA mode (from nonha_workload file)
var expectedWorkloadsNonHA = []string{
	// Deployments
	"Deployment/suse-observability-anomaly-detection-spotlight-manager",
	"Deployment/suse-observability-anomaly-detection-spotlight-worker",
	"Deployment/suse-observability-prometheus-elasticsearch-exporter",
	"Deployment/suse-observability-hbase-console",
	"Deployment/suse-observability-kafkaup-operator-kafkaup",
	"Deployment/suse-observability-rbac-agent",
	"Deployment/suse-observability-minio",
	"Deployment/suse-observability-correlate",
	"Deployment/suse-observability-e2es",
	"Deployment/suse-observability-receiver",
	"Deployment/suse-observability-router",
	"Deployment/suse-observability-server",
	"Deployment/suse-observability-ui",
	// StatefulSets
	"StatefulSet/suse-observability-clickhouse-shard0",
	"StatefulSet/suse-observability-elasticsearch-master",
	"StatefulSet/suse-observability-hbase-stackgraph",
	"StatefulSet/suse-observability-hbase-tephra-mono",
	"StatefulSet/suse-observability-kafka",
	"StatefulSet/suse-observability-otel-collector",
	"StatefulSet/suse-observability-victoria-metrics-0",
	"StatefulSet/suse-observability-zookeeper",
	"StatefulSet/suse-observability-vmagent",
	"StatefulSet/suse-observability-workload-observer",
	// Jobs (dynamic names will be handled with pattern matching)
	"Job/suse-observability-backup-conf-*",
	"Job/suse-observability-backup-init-*",
	"Job/suse-observability-topic-create-*",
	"Job/suse-observability-ch-clean*",
	// CronJobs
	"CronJob/suse-observability-backup-sg",
	"CronJob/suse-observability-backup-conf",
	"CronJob/suse-observability-backup-init",
	"CronJob/suse-observability-clickhouse-incremental-backup",
	"CronJob/suse-observability-clickhouse-full-backup",
}

func TestWorkloadHARendering(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/workload_ha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Test that ONLY expected workloads are rendered (no more, no less)
	testExactWorkloadsRendered(t, &resources, expectedWorkloadsHA, "HA")
}

func TestWorkloadNonHARendering(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/workload_nonha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Test that ONLY expected workloads are rendered (no more, no less)
	testExactWorkloadsRendered(t, &resources, expectedWorkloadsNonHA, "Non-HA")
}

func TestWorkloadGlobalHARendering(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/workload_global_ha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Test that ONLY expected workloads are rendered (no more, no less)
	testExactWorkloadsRendered(t, &resources, expectedWorkloadsHA, "HA")
}

func TestWorkloadGlobalNonHARendering(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/workload_global_nonha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Test that ONLY expected workloads are rendered (no more, no less)
	testExactWorkloadsRendered(t, &resources, expectedWorkloadsNonHA, "Non-HA")
}

func testExactWorkloadsRendered(t *testing.T, resources *helmtestutil.KubernetesResources, expected []string, mode string) {
	// Collect all rendered workloads
	actualWorkloads := make(map[string]bool)

	// Add all deployments
	for _, deployment := range resources.Deployments {
		actualWorkloads["Deployment/"+deployment.Name] = true
	}

	// Add all statefulsets
	for _, statefulset := range resources.Statefulsets {
		actualWorkloads["StatefulSet/"+statefulset.Name] = true
	}

	// Add all jobs
	for _, job := range resources.Jobs {
		actualWorkloads["Job/"+job.Name] = true
	}

	// Add all cronjobs
	for _, cronjob := range resources.CronJobs {
		actualWorkloads["CronJob/"+cronjob.Name] = true
	}

	// Check that all expected workloads are present
	expectedSet := make(map[string]bool)
	for _, expected := range expected {
		expectedSet[expected] = true
		found := false

		// Handle pattern matching for dynamic job names (containing *)
		if strings.Contains(expected, "*") {
			pattern := strings.ReplaceAll(expected, "*", "")
			for actual := range actualWorkloads {
				if strings.HasPrefix(actual, pattern) {
					found = true
					break
				}
			}
		} else {
			// Exact match
			found = actualWorkloads[expected]
		}

		assert.True(t, found, "%s mode: Expected workload '%s' to be rendered but was not found", mode, expected)
	}

	// Check that no unexpected workloads are present
	for actual := range actualWorkloads {
		found := false

		// Check exact match first
		if expectedSet[actual] {
			found = true
		} else {
			// Check pattern match for dynamic job names
			for _, expected := range expected {
				if strings.Contains(expected, "*") {
					pattern := strings.ReplaceAll(expected, "*", "")
					if strings.HasPrefix(actual, pattern) {
						found = true
						break
					}
				}
			}
		}

		assert.True(t, found, "%s mode: Unexpected workload '%s' was rendered but not expected", mode, actual)
	}
}
