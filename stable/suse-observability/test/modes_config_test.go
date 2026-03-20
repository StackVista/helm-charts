package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestModesDefaultApi(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", nil, `stackstate.applicationDomains = ["Observability"]`)
}

func TestModesDefaultServer(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/split_disabled.yaml"}, `stackstate.applicationDomains = ["Observability"]`)
}

func TestModesCustomApi(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"global.suseObservability.applicationDomains[0]": "Observability",
			"global.suseObservability.applicationDomains[1]": "Security",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, ok := resources.ConfigMaps["suse-observability-api"]
	require.True(t, ok, "API configmap should exist")
	configData := configMap.Data["application_stackstate.conf"]
	assert.Contains(t, configData, `"Observability"`)
	assert.Contains(t, configData, `"Security"`)
}

func TestModesEmptyModes(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/modes_empty.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "global.suseObservability.applicationDomains must contain at least one application domain")
}

func TestModesInvalidMode(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
		},
		SetValues: map[string]string{
			"global.suseObservability.applicationDomains[0]": "InvalidMode",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `Invalid application domain "InvalidMode"`)
}

func TestModesCustomServer(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{
			"values/full.yaml",
			"values/split_disabled.yaml",
		},
		SetValues: map[string]string{
			"global.suseObservability.applicationDomains[0]": "Observability",
			"global.suseObservability.applicationDomains[1]": "Security",
		},
		KubectlOptions: &k8s.KubectlOptions{
			Namespace: "suse-observability",
		},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, ok := resources.ConfigMaps["suse-observability-server"]
	require.True(t, ok, "Server configmap should exist")
	configData := configMap.Data["application_stackstate.conf"]
	assert.Contains(t, configData, `"Observability"`)
	assert.Contains(t, configData, `"Security"`)
}
