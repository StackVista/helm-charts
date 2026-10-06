package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

const maxExcludeDatanodeEnv = "HBASE_CONF_hbase_regionserver_async_wal_max_exclude_datanode_count"

func distributedOptions(values map[string]string) *helm.Options {
	setValues := map[string]string{"deployment.mode": "Distributed"}
	for key, value := range values {
		setValues[key] = value
	}
	return &helm.Options{ValuesFiles: []string{"values/distributed-mode.yaml"}, SetValues: setValues, Logger: logger.Discard}
}

func TestHBaseMaxExcludeDatanodeCount(t *testing.T) {
	for _, tc := range []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{"chart-defaults", nil, "1"},
		{"150-ha", map[string]string{"global.suseObservability.sizing.profile": "150-ha"}, "1"},
		{"4000-ha", map[string]string{"global.suseObservability.sizing.profile": "4000-ha"}, "3"},
		{"single-datanode", map[string]string{"hdfs.datanode.replicaCount": "1"}, "0"},
		{"min-replication-override", map[string]string{"hdfs.datanode.replicaCount": "5", "hdfs.replication": "3", "hdfs.minReplication": "3"}, "2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, releaseName, distributedOptions(tc.values))
			resources := helmtestutil.NewKubernetesResources(t, output)
			for _, name := range []string{"hbase-master", "hbase-rs", "tephra"} {
				sts, ok := resources.Statefulsets[releaseName+"-hbase-"+name]
				require.True(t, ok, "StatefulSet %s should exist", name)
				value, found := "", false
				for _, env := range sts.Spec.Template.Spec.Containers[0].Env {
					if env.Name == maxExcludeDatanodeEnv {
						value, found = env.Value, true
					}
				}
				require.True(t, found, "%s should set %s", name, maxExcludeDatanodeEnv)
				assert.Equal(t, tc.expected, value, name)
			}
		})
	}
}

func TestHBaseDatanodeCountMustExceedReplication(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values map[string]string
	}{
		{"equal", map[string]string{"hdfs.datanode.replicaCount": "3", "hdfs.replication": "3"}},
		{"fewer-datanodes", map[string]string{"hdfs.datanode.replicaCount": "2", "hdfs.replication": "3"}},
		{"4000-ha-with-datanode-override", map[string]string{"global.suseObservability.sizing.profile": "4000-ha", "hdfs.datanode.replicaCount": "3"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := helmtestutil.RenderHelmTemplateOpts(t, releaseName, distributedOptions(tc.values))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "must be greater than hbase.hdfs.replication")
		})
	}
}

func TestHBaseMinReplicationMustNotExceedReplication(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values map[string]string
	}{
		{"replication-lowered-below-default-min", map[string]string{"hdfs.datanode.replicaCount": "2", "hdfs.replication": "1"}},
		{"min-replication-raised", map[string]string{"hdfs.datanode.replicaCount": "5", "hdfs.replication": "2", "hdfs.minReplication": "3"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := helmtestutil.RenderHelmTemplateOpts(t, releaseName, distributedOptions(tc.values))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "must not exceed hbase.hdfs.replication")
		})
	}
}

func TestHBaseDatanodeReplicationAllowed(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values map[string]string
	}{
		{"single-datanode", map[string]string{"hdfs.datanode.replicaCount": "1", "hdfs.replication": "3"}},
		{"replication-one", map[string]string{"hdfs.datanode.replicaCount": "2", "hdfs.replication": "1", "hdfs.minReplication": "1"}},
		{"spare-datanode", map[string]string{"hdfs.datanode.replicaCount": "4", "hdfs.replication": "3"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			helmtestutil.RenderHelmTemplateOptsNoError(t, releaseName, distributedOptions(tc.values))
		})
	}

	for _, profile := range []string{"150-ha", "250-ha", "500-ha", "4000-ha"} {
		t.Run(profile, func(t *testing.T) {
			helmtestutil.RenderHelmTemplateOptsNoError(t, releaseName, distributedOptions(map[string]string{"global.suseObservability.sizing.profile": profile}))
		})
	}

	t.Run("mono-ignores-datanodes", func(t *testing.T) {
		options := distributedOptions(map[string]string{"deployment.mode": "Mono", "hdfs.datanode.replicaCount": "2", "hdfs.replication": "3"})
		helmtestutil.RenderHelmTemplateOptsNoError(t, releaseName, options)
	})
}
