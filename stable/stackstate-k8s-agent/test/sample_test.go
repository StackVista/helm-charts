package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestHelmBasicRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s-agent", "values/minimal.yaml")

	// Parse all resources into their corresponding types for validation and further inspection
	helmtestutil.NewKubernetesResources(t, output)
}

func TestClusterNameValidation(t *testing.T) {
	testCases := []struct {
		Name        string
		ClusterName string
		IsValid     bool
	}{
		{"not allowed end with special character [.]", "name.", false},
		{"not allowed end with special character [-]", "name.", false},
		{"not allowed start with special character [-]", "-name", false},
		{"not allowed start with special character [.]", ".name", false},
		{"special characters are allowed in the middle", "euwest1-prod.cool-company.com", true},
		{"underscore is allowed everywhere", "_why_7_", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			output, err := helmtestutil.RenderHelmTemplateOpts(
				t, "cluster-agent",
				&helm.Options{
					ValuesFiles: []string{"values/minimal.yaml"},
					SetStrValues: map[string]string{
						"stackstate.cluster.name": testCase.ClusterName,
					},
				})
			if testCase.IsValid {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Contains(t, output, "stackstate.cluster.name: Does not match pattern")
			}
		})
	}
}
