package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestServerJavaHeapRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]v1.EnvVar)
	expectedDeployments["stackstate-server"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m"}
	expectedDeployments["stackstate-receiver"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx2697m -Xms2697m"}
	expectedDeployments["stackstate-correlate"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx799m -Xms799m"}
	expectedDeployments["stackstate-mm2es"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx402m -Xms402m"}
	expectedDeployments["stackstate-kafka2prom"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx402m -Xms402m"}

	var foundDeployments = make(map[string]appsv1.Deployment)

	for _, deployment := range resources.Deployments {
		if _, ok := expectedDeployments[deployment.Name]; ok {
			foundDeployments[deployment.Name] = deployment
		}
	}

	assert.Equal(t, len(expectedDeployments), len(foundDeployments))

	for deploymentName, expectedJavaOpts := range expectedDeployments {
		AssertJavaOpts(t, foundDeployments[deploymentName].Spec.Template.Spec.Containers[0].Env, expectedJavaOpts)
	}
}

func TestSplitServicesJavaHeapRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]v1.EnvVar)
	expectedDeployments["stackstate-api"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1750m -Xms1750m"}
	expectedDeployments["stackstate-checks"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx2450m -Xms2450m"}
	expectedDeployments["stackstate-state"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1200m -Xms1200m"}
	expectedDeployments["stackstate-problem-producer"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1200m -Xms1200m"}
	expectedDeployments["stackstate-sync"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1800m -Xms1800m"}
	expectedDeployments["stackstate-slicing"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx780m -Xms780m"}
	expectedDeployments["stackstate-view-health"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1210m -Xms1210m"}
	expectedDeployments["stackstate-health-sync"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx680m -Xms680m"}
	expectedDeployments["stackstate-initializer"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx260m -Xms260m"}

	var foundDeployments = make(map[string]appsv1.Deployment)

	for _, deployment := range resources.Deployments {
		if _, ok := expectedDeployments[deployment.Name]; ok {
			foundDeployments[deployment.Name] = deployment
		}
	}

	assert.Equal(t, len(expectedDeployments), len(foundDeployments))

	for deploymentName, expectedJavaOpts := range expectedDeployments {
		AssertJavaOpts(t, foundDeployments[deploymentName].Spec.Template.Spec.Containers[0].Env, expectedJavaOpts)
	}
}

func TestServerJavaHeapRenderWithAllJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/components_all_javaopts.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-server" {
			stsServerDeployment = deployment
		}
	}

	assert.NotNil(t, stsServerDeployment)

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func TestServerJavaHeapRenderWithServerJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/components_server_javaopts.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-server" {
			stsServerDeployment = deployment
		}
	}

	assert.NotNil(t, stsServerDeployment)

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func TestServerJavaHeapRenderWithBothJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/components_all_javaopts.yaml", "values/components_server_javaopts.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-server" {
			stsServerDeployment = deployment
		}
	}

	assert.NotNil(t, stsServerDeployment)

	// The service specific overrides the common JAVA_OPTS
	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func AssertJavaOpts(t *testing.T, env []v1.EnvVar, expectedVar v1.EnvVar) {
	assert.Contains(t, env, expectedVar)
}
