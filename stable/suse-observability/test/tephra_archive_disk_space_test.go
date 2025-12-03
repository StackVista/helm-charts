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

	for _, statefulSetTephra := range resources.Statefulsets {
		if statefulSetTephra.Name == "suse-observability-hbase-tephra" {
			stackstateTephraStatefulset = statefulSetTephra
		}
	}

	require.NotNil(t, stackstateTephraStatefulset)
	expectedTephraArchiveDiskSpace := v1.EnvVar{Name: "HBASE_CONF_tephra_tx_snapshot_archive_max_size_mb", Value: "26850"}
	require.Contains(t, stackstateTephraStatefulset.Spec.Template.Spec.Containers[0].Env, expectedTephraArchiveDiskSpace)
}

func TestTephraMonoHbaseArchiveDiskSpaceRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/mono_hbase.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stackstateTephraStatefulset appsv1.StatefulSet

	for _, statefulSetTephra := range resources.Statefulsets {
		if statefulSetTephra.Name == "suse-observability-hbase-tephra-mono" {
			stackstateTephraStatefulset = statefulSetTephra
		}
	}

	require.NotNil(t, stackstateTephraStatefulset)
	expectedTephraArchiveDiskSpace := v1.EnvVar{Name: "HBASE_CONF_tephra_tx_snapshot_archive_max_size_mb", Value: "107"}
	require.Contains(t, stackstateTephraStatefulset.Spec.Template.Spec.Containers[0].Env, expectedTephraArchiveDiskSpace)
}
