package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestConfigurationTemplatesUseHelper(t *testing.T) {
	helmtestutil.RequireTemplateValueAccessViaHelper(t, "../templates", "_configuration.tpl", `\.Values\.stackstate\.instance\b`, "anomaly-detection.stackstate.instance")
}

func TestInstanceStandalone(t *testing.T) {
	for _, instance := range []string{"https://analysis.example", "https://{{ .Release.Name }}.example"} {
		t.Run(instance, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "orders", &helm.Options{
				SetValues: map[string]string{
					"stackstate.instance":   instance,
					"global.receiverApiKey": "test-key",
				},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			require.Len(t, resources.Deployments, 2)
			expected := "https://analysis.example"
			if instance != expected {
				expected = "https://orders.example"
			}
			for _, deployment := range resources.Deployments {
				found := false
				for _, c := range deployment.Spec.Template.Spec.Containers {
					for i, arg := range c.Args {
						if arg == "--instance" {
							require.Less(t, i+1, len(c.Args))
							assert.Equal(t, expected, c.Args[i+1])
							found = true
						}
					}
				}
				assert.True(t, found)
			}
		})
	}
}
