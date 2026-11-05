package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestParentHdfsNamenodeStartupProtectsSlowInitialization(t *testing.T) {
	for _, valuesFile := range []string{"values/full.yaml", "values/global_sizing_150_ha.yaml"} {
		t.Run(valuesFile, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{valuesFile},
				SetValues: map[string]string{
					"hbase.deployment.mode": "Distributed",
				},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			namenode, ok := resources.Statefulsets["suse-observability-hbase-hdfs-nn"]
			require.True(t, ok)
			container := namenode.Spec.Template.Spec.Containers[0]
			require.NotNil(t, container.StartupProbe)
			require.NotNil(t, container.StartupProbe.HTTPGet)
			require.NotNil(t, container.LivenessProbe)
			require.NotNil(t, container.ReadinessProbe)
			assert.Equal(t, container.LivenessProbe.HTTPGet, container.StartupProbe.HTTPGet)
			assert.Equal(t, container.ReadinessProbe.HTTPGet, container.StartupProbe.HTTPGet)
			// Protect slow initial formatting until the endpoint succeeds;
			// normal liveness checks then resume immediately.
			assert.Equal(t, int32(300), container.StartupProbe.FailureThreshold*container.StartupProbe.PeriodSeconds)
			for _, name := range []string{"nn", "dn", "snn"} {
				sts, ok := resources.Statefulsets["suse-observability-hbase-hdfs-"+name]
				require.True(t, ok)
				for _, init := range sts.Spec.Template.Spec.InitContainers {
					assert.Nil(t, init.StartupProbe, name+"/"+init.Name)
					assert.Nil(t, init.LivenessProbe, name+"/"+init.Name)
					assert.Nil(t, init.ReadinessProbe, name+"/"+init.Name)
				}
			}
		})
	}
}
