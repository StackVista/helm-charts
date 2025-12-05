package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestTephraArchiveDiskSpaceRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateTephraStatefulset appsv1.StatefulSet
	var stackstateMasterStatefulSet appsv1.StatefulSet

	for _, statefulSet := range resources.Statefulsets {
		if statefulSet.Name == "suse-observability-hbase-tephra" {
			stackstateTephraStatefulset = statefulSet
		}
		if statefulSet.Name == "suse-observability-hbase-hbase-master" {
			stackstateMasterStatefulSet = statefulSet
		}
	}

	require.NotNil(t, stackstateTephraStatefulset)
	expectedTephraArchiveDiskSpace := v1.EnvVar{Name: "HBASE_CONF_tephra_tx_snapshot_archive_max_size_mb", Value: "26850"}
	require.Contains(t, stackstateTephraStatefulset.Spec.Template.Spec.Containers[0].Env, expectedTephraArchiveDiskSpace)

	require.NotNil(t, stackstateMasterStatefulSet)
	expectedMasterArchiveDiskSpace := v1.EnvVar{Name: "HBASE_CONF_hbase_master_stackstate_logcleaner_max_size_mb", Value: "13425"}
	require.Contains(t, stackstateMasterStatefulSet.Spec.Template.Spec.Containers[0].Env, expectedMasterArchiveDiskSpace)
}

func TestTephraMonoHbaseArchiveDiskSpaceRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/mono_hbase.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateTephraStatefulset appsv1.StatefulSet
	var stackstateMasterStatefulSet appsv1.StatefulSet

	for _, statefulSet := range resources.Statefulsets {
		if statefulSet.Name == "suse-observability-hbase-tephra-mono" {
			stackstateTephraStatefulset = statefulSet
		}

		if statefulSet.Name == "suse-observability-hbase-stackgraph" {
			stackstateMasterStatefulSet = statefulSet
		}
	}

	require.NotNil(t, stackstateTephraStatefulset)
	expectedTephraArchiveDiskSpace := v1.EnvVar{Name: "HBASE_CONF_tephra_tx_snapshot_archive_max_size_mb", Value: "107"}
	require.Contains(t, stackstateTephraStatefulset.Spec.Template.Spec.Containers[0].Env, expectedTephraArchiveDiskSpace)

	require.NotNil(t, stackstateMasterStatefulSet)
	expectedMasterArchiveDiskSpace := v1.EnvVar{Name: "HBASE_CONF_hbase_master_stackstate_logcleaner_max_size_mb", Value: "13425"}
	require.Contains(t, stackstateMasterStatefulSet.Spec.Template.Spec.Containers[0].Env, expectedMasterArchiveDiskSpace)
}
