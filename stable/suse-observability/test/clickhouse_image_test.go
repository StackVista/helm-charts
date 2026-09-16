package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// The cleanup client and server must advance together when consuming a new release.
func TestClickhouseCleanupUsesServerImage(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	var serverImage string
	for _, sts := range resources.StatefulSets {
		for _, container := range sts.Spec.Template.Spec.Containers {
			if container.Name == "clickhouse" {
				serverImage = container.Image
			}
		}
	}
	require.NotEmpty(t, serverImage, "ClickHouse server must be rendered")
	cleanup := findJob(resources, "ch-clean")
	require.NotNil(t, cleanup, "ClickHouse cleanup hook must be rendered")
	require.NotEmpty(t, cleanup.Spec.Template.Spec.Containers)
	require.Equal(t, serverImage, cleanup.Spec.Template.Spec.Containers[0].Image)
}
