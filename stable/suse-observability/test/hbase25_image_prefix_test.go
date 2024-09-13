package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
)

func TestRegularImageNonSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]string)
	expectedDeployments["stackstate-k8s-server"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"

	var foundDeployments = make(map[string]appsv1.Deployment)

	for _, deployment := range resources.Deployments {
		if _, ok := expectedDeployments[deployment.Name]; ok {
			foundDeployments[deployment.Name] = deployment
		}
	}

	assert.Equal(t, len(expectedDeployments), len(foundDeployments))

	for deploymentName, expectedImageRegex := range expectedDeployments {
		assert.Regexp(t, expectedImageRegex, foundDeployments[deploymentName].Spec.Template.Spec.Containers[0].Image)
	}
}

func TestRegularImageSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]string)
	expectedDeployments["stackstate-k8s-api"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-checks"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-health-sync"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-initializer"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-notification"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-state"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-sync"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"

	var foundDeployments = make(map[string]appsv1.Deployment)

	for _, deployment := range resources.Deployments {
		if _, ok := expectedDeployments[deployment.Name]; ok {
			foundDeployments[deployment.Name] = deployment
		}
	}

	assert.Equal(t, len(expectedDeployments), len(foundDeployments))

	for deploymentName, expectedImageRegex := range expectedDeployments {
		assert.Regexp(t, expectedImageRegex, foundDeployments[deploymentName].Spec.Template.Spec.Containers[0].Image)
	}
}

func TestHbase12ImageSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/hbase12_enabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]string)
	expectedDeployments["stackstate-k8s-api"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-checks"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-health-sync"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-initializer"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-notification"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-state"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["stackstate-k8s-sync"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"

	var foundDeployments = make(map[string]appsv1.Deployment)

	for _, deployment := range resources.Deployments {
		if _, ok := expectedDeployments[deployment.Name]; ok {
			foundDeployments[deployment.Name] = deployment
		}
	}

	assert.Equal(t, len(expectedDeployments), len(foundDeployments))

	for deploymentName, expectedImageRegex := range expectedDeployments {
		assert.NotRegexp(t, expectedImageRegex, foundDeployments[deploymentName].Spec.Template.Spec.Containers[0].Image)
	}
}
