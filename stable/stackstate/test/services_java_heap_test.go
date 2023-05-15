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
	expectedDeployments["stackstate-server"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=2428m -Xmx5664m -Xms5664m"}
	expectedDeployments["stackstate-receiver"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1399m -Xmx2597m -Xms2597m"}
	expectedDeployments["stackstate-correlate"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=420m -Xmx780m -Xms780m"}
	expectedDeployments["stackstate-mm2es"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=387m -Xmx387m -Xms387m"}
	expectedDeployments["stackstate-kafka2prom-0"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=387m -Xmx387m -Xms387m"}
	expectedDeployments["stackstate-kafka2prom-1"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=387m -Xmx387m -Xms387m"}

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
	expectedDeployments["stackstate-api"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1750m -Xmx1750m -Xms1750m"}
	expectedDeployments["stackstate-checks"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1400m -Xmx2100m -Xms2100m"}
	expectedDeployments["stackstate-state"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=510m -Xmx1190m -Xms1190m"}
	expectedDeployments["stackstate-problem-producer"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=340m -Xmx1360m -Xms1360m"}
	expectedDeployments["stackstate-sync"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1240m -Xmx1860m -Xms1860m"}
	expectedDeployments["stackstate-slicing"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=520m -Xmx780m -Xms780m"}
	expectedDeployments["stackstate-view-health"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1035m -Xmx1265m -Xms1265m"}
	expectedDeployments["stackstate-health-sync"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=853m -Xmx697m -Xms697m"}
	expectedDeployments["stackstate-initializer"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=140m -Xmx260m -Xms260m"}

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

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=2428m -Xmx5664m -Xms5664m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005"}

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

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=2428m -Xmx5664m -Xms5664m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

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
	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=2428m -Xmx5664m -Xms5664m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func AssertJavaOpts(t *testing.T, env []v1.EnvVar, expectedVar v1.EnvVar) {
	assert.Contains(t, env, expectedVar)
}
