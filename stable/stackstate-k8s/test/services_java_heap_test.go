package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestServerJavaHeapRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]v1.EnvVar)
	expectedDeployments["stackstate-k8s-server"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=2426m -Xmx5658m -Xms5658m"}
	expectedDeployments["stackstate-k8s-receiver"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1398m -Xmx2594m -Xms2594m"}
	expectedDeployments["stackstate-k8s-correlate"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=840m -Xmx1560m -Xms1560m"}

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
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]v1.EnvVar)
	expectedDeployments["stackstate-k8s-api"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=906m -Xmx740m -Xms740m"}
	expectedDeployments["stackstate-k8s-checks"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1400m -Xmx2100m -Xms2100m"}
	expectedDeployments["stackstate-k8s-state"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=510m -Xmx1105m -Xms803m"}
	expectedDeployments["stackstate-k8s-sync"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1557m -Xmx2335m -Xms1691m"}
	expectedDeployments["stackstate-k8s-slicing"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=520m -Xmx780m -Xms621m"}
	expectedDeployments["stackstate-k8s-view-health"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1035m -Xmx1265m -Xms960m"}
	expectedDeployments["stackstate-k8s-health-sync"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1678m -Xmx1372m -Xms1372m"}
	expectedDeployments["stackstate-k8s-notification"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=460m -Xmx632m -Xms632m"}
	expectedDeployments["stackstate-k8s-initializer"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=403m -Xmx747m -Xms105m"}

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
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/components_all_javaopts.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-k8s-server" {
			stsServerDeployment = deployment
		}
	}

	assert.NotNil(t, stsServerDeployment)

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=2426m -Xmx5658m -Xms5658m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func TestServerJavaHeapRenderWithServerJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/components_server_javaopts.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-k8s-server" {
			stsServerDeployment = deployment
		}
	}

	assert.NotNil(t, stsServerDeployment)

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=2426m -Xmx5658m -Xms5658m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func TestServerJavaHeapRenderWithBothJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-k8s", "values/full.yaml", "values/components_all_javaopts.yaml", "values/components_server_javaopts.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-k8s-server" {
			stsServerDeployment = deployment
		}
	}

	assert.NotNil(t, stsServerDeployment)

	// The service specific overrides the common JAVA_OPTS
	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=2426m -Xmx5658m -Xms5658m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func AssertJavaOpts(t *testing.T, env []v1.EnvVar, expectedVar v1.EnvVar) {
	assert.Contains(t, env, expectedVar)
}
