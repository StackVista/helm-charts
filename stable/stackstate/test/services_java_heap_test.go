package test

import (
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestServerJavaHeapRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]v1.EnvVar)
	expectedDeployments["stackstate-server"] =  v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m"}
	expectedDeployments["stackstate-receiver"] =  v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1086m -Xms280m"}
	expectedDeployments["stackstate-correlate"] =  v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx799m -Xms799m"}
	expectedDeployments["stackstate-mm2es"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx402m -Xms402m"}

	var foundDeployments = make(map[string]appsv1.Deployment)

	for _, deployment := range resources.Deployments {
		if _, ok := expectedDeployments[deployment.Name]; ok {
			foundDeployments[deployment.Name] = deployment
		}
	}

	require.Equal(t, len(expectedDeployments), len(foundDeployments))

	for deploymentName, expectedJavaOpts := range expectedDeployments {
		AssertJavaOpts(t, foundDeployments[deploymentName].Spec.Template.Spec.Containers[0].Env, expectedJavaOpts)
	}
}

func TestSplitServicesJavaHeapRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/split_enabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var expectedDeployments = make(map[string]v1.EnvVar)
	expectedDeployments["stackstate-api"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1750m -Xms1750m"}
	expectedDeployments["stackstate-checks"] =  v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx2450m -Xms2450m"}
	expectedDeployments["stackstate-state"] =  v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1200m -Xms1200m"}
	expectedDeployments["stackstate-sync"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx960m -Xms960m"}
	expectedDeployments["stackstate-slicing"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx800m -Xms800m"}
	expectedDeployments["stackstate-view-health"] = v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1210m -Xms1210m"}

	var foundDeployments = make(map[string]appsv1.Deployment)

	for _, deployment := range resources.Deployments {
		if _, ok := expectedDeployments[deployment.Name]; ok {
			foundDeployments[deployment.Name] = deployment
		}
	}

	require.Equal(t, len(expectedDeployments), len(foundDeployments))

	for deploymentName, expectedJavaOpts := range expectedDeployments {
		AssertJavaOpts(t, foundDeployments[deploymentName].Spec.Template.Spec.Containers[0].Env, expectedJavaOpts)
	}
}

func TestServerJavaHeapRenderWithAllJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/components_all_javaopts.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-server" {
			stsServerDeployment = deployment
		}
	}

	require.NotNil(t, stsServerDeployment)

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func TestServerJavaHeapRenderWithServerJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/components_server_javaopts.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deploymentK2ES := range resources.Deployments {
		if deploymentK2ES.Name == "stackstate-server" {
			stsServerDeployment = deploymentK2ES
		}
	}

	require.NotNil(t, stsServerDeployment)

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func TestServerJavaHeapRenderWithBothJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/components_all_javaopts.yaml", "values/components_server_javaopts.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deploymentK2ES := range resources.Deployments {
		if deploymentK2ES.Name == "stackstate-server" {
			stsServerDeployment = deploymentK2ES
		}
	}

	require.NotNil(t, stsServerDeployment)

	// The service specific overrides the common JAVA_OPTS
	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
}

func AssertJavaOpts(t *testing.T, env []v1.EnvVar, expectedVar v1.EnvVar) {
	require.Contains(t, env, expectedVar)
}
