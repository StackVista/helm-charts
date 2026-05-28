package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
)

func TestRegularImageNonSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]string)
	expectedDeployments["suse-observability-server"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"

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
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]string)
	expectedDeployments["suse-observability-api"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["suse-observability-checks"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["suse-observability-health-sync"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["suse-observability-initializer"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["suse-observability-notification"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["suse-observability-state"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"
	expectedDeployments["suse-observability-sync"] = ".*stackstate-server:\\d\\.0\\.0-snapshot.*-2\\.5"

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

func TestHbaseHdfsUsesFullImageTag(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/hbase12_enabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	expectedStatefulSets := map[string]string{
		"suse-observability-hbase-hdfs-dn":  "my.registry.com/stackstate/hadoop:3.4.3-so1",
		"suse-observability-hbase-hdfs-nn":  "my.registry.com/stackstate/hadoop:3.4.3-so1",
		"suse-observability-hbase-hdfs-snn": "my.registry.com/stackstate/hadoop:3.4.3-so1",
	}

	for statefulSetName, expectedImage := range expectedStatefulSets {
		statefulSet, ok := resources.Statefulsets[statefulSetName]
		assert.True(t, ok, "expected StatefulSet %s to render", statefulSetName)
		if !ok {
			continue
		}
		assert.Equal(t, expectedImage, statefulSet.Spec.Template.Spec.Containers[0].Image)
	}
}
