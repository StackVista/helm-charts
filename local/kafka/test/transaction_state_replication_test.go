package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestTransactionStateReplicationFactorBySizingProfile(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		override string
		expected string
	}{
		{name: "10-nonha", profile: "10-nonha", expected: "1"},
		{name: "150-ha", profile: "150-ha", expected: "2"},
		{name: "250-ha", profile: "250-ha", expected: "2"},
		{name: "500-ha", profile: "500-ha", expected: "2"},
		{name: "4000-ha", profile: "4000-ha", expected: "2"},
		{name: "150-ha-explicit-override", profile: "150-ha", override: "1", expected: "1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setValues := map[string]string{
				"global.suseObservability.sizing.profile": test.profile,
			}
			if test.override != "" {
				setValues["transactionStateLogReplicationFactor"] = test.override
			}
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "kafka", &helm.Options{
				SetValues: setValues,
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			require.Len(t, resources.Statefulsets, 1)

			for _, statefulset := range resources.Statefulsets {
				for _, env := range statefulset.Spec.Template.Spec.Containers[0].Env {
					if env.Name == "KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR" {
						assert.Equal(t, test.expected, env.Value)
						return
					}
				}
			}
			t.Fatal("KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR was not rendered")
		})
	}
}
