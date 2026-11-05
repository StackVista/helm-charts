package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHdfsNamenodeStartupProtectsSlowInitialization(t *testing.T) {
	for _, profile := range []string{"", "150-ha"} {
		t.Run("profile="+profile, func(t *testing.T) {
			values := map[string]string{}
			if profile != "" {
				values["global.suseObservability.sizing.profile"] = profile
			}
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, releaseName, distributedOptions(values))
			resources := helmtestutil.NewKubernetesResources(t, output)
			sts, ok := resources.Statefulsets[releaseName+"-hbase-hdfs-nn"]
			require.True(t, ok)
			container := sts.Spec.Template.Spec.Containers[0]
			require.NotNil(t, container.StartupProbe)
			require.NotNil(t, container.StartupProbe.HTTPGet)
			require.NotNil(t, container.LivenessProbe)
			require.NotNil(t, container.ReadinessProbe)
			assert.Equal(t, container.LivenessProbe.HTTPGet, container.StartupProbe.HTTPGet)
			assert.Equal(t, container.ReadinessProbe.HTTPGet, container.StartupProbe.HTTPGet)
			// The observed startup was killed at 60 seconds. Allow five minutes,
			// while Kubernetes resumes the normal liveness checks immediately
			// after the startup endpoint first succeeds.
			assert.Equal(t, int32(300), container.StartupProbe.FailureThreshold*container.StartupProbe.PeriodSeconds)
			for _, init := range sts.Spec.Template.Spec.InitContainers {
				assert.Nil(t, init.StartupProbe, init.Name)
				assert.Nil(t, init.LivenessProbe, init.Name)
				assert.Nil(t, init.ReadinessProbe, init.Name)
			}
		})
	}
}
