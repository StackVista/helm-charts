package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestYamlMaxRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/split_disabled.yaml", "values/yaml_max.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]v1.EnvVar)
	expectedDeployments["stackstate-k8s-server"] = v1.EnvVar{Name: "CONFIG_FORCE_stackstate_yaml_codePointLimit", Value: "3145728"}

	var foundDeployments = make(map[string]appsv1.Deployment)

	for _, deployment := range resources.Deployments {
		if _, ok := expectedDeployments[deployment.Name]; ok {
			foundDeployments[deployment.Name] = deployment
		}
	}

	assert.Equal(t, len(expectedDeployments), len(foundDeployments))

	for deploymentName, expectedJavaOpts := range expectedDeployments {
		AssertEnvValue(t, foundDeployments[deploymentName].Spec.Template.Spec.Containers[0].Env, expectedJavaOpts)
	}
}

func TestSplitYamlMaxRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/yaml_max.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]v1.EnvVar)
	expectedDeployments["stackstate-k8s-api"] = v1.EnvVar{Name: "CONFIG_FORCE_stackstate_yaml_codePointLimit", Value: "3145728"}

	var foundDeployments = make(map[string]appsv1.Deployment)

	for _, deployment := range resources.Deployments {
		if _, ok := expectedDeployments[deployment.Name]; ok {
			foundDeployments[deployment.Name] = deployment
		}
	}

	assert.Equal(t, len(expectedDeployments), len(foundDeployments))

	for deploymentName, expectedJavaOpts := range expectedDeployments {
		AssertEnvValue(t, foundDeployments[deploymentName].Spec.Template.Spec.Containers[0].Env, expectedJavaOpts)
	}
}

func AssertEnvValue(t *testing.T, env []v1.EnvVar, expectedVar v1.EnvVar) {
	assert.Contains(t, env, expectedVar)
}
