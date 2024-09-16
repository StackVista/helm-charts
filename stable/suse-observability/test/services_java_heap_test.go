package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestServerJavaHeapRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]v1.EnvVar)
	expectedDeployments["suse-observability-server"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=3630m -Xmx4435m -Xms4435m"}
	expectedDeployments["suse-observability-receiver"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1393m -Xmx2587m -Xms2587m"}
	expectedDeployments["suse-observability-correlate"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=881m -Xmx1635m -Xms1635m"}

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
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]v1.EnvVar)
	expectedDeployments["suse-observability-api"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=893m -Xmx730m -Xms730m"}
	expectedDeployments["suse-observability-checks"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1468m -Xmx2202m -Xms2202m"}
	expectedDeployments["suse-observability-state"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=624m -Xmx1158m -Xms842m"}
	expectedDeployments["suse-observability-sync"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1550m -Xmx2325m -Xms1680m"}
	expectedDeployments["suse-observability-slicing"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=786m -Xmx786m -Xms681m"}
	expectedDeployments["suse-observability-health-sync"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=1759m -Xmx1439m -Xms1439m"}
	expectedDeployments["suse-observability-notification"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=543m -Xmx662m -Xms662m"}
	expectedDeployments["suse-observability-initializer"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=422m -Xmx783m -Xms109m"}

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
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/components_all_javaopts.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "suse-observability-server" {
			stsServerDeployment = deployment
		}
	}

	assert.NotNil(t, stsServerDeployment)

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=3630m -Xmx4435m -Xms4435m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func TestServerJavaHeapRenderWithServerJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/components_server_javaopts.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "suse-observability-server" {
			stsServerDeployment = deployment
		}
	}

	assert.NotNil(t, stsServerDeployment)

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=3630m -Xmx4435m -Xms4435m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func TestServerJavaHeapRenderWithBothJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/components_all_javaopts.yaml", "values/components_server_javaopts.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "suse-observability-server" {
			stsServerDeployment = deployment
		}
	}

	assert.NotNil(t, stsServerDeployment)

	// The service specific overrides the common JAVA_OPTS
	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:MaxDirectMemorySize=3630m -Xmx4435m -Xms4435m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func AssertJavaOpts(t *testing.T, env []v1.EnvVar, expectedVar v1.EnvVar) {
	assert.Contains(t, env, expectedVar)
}
