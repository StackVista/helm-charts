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

	var stsServerDeployment appsv1.Deployment

	for _, deploymentK2ES := range resources.Deployments {
		if deploymentK2ES.Name == "stackstate-server" {
			stsServerDeployment = deploymentK2ES
		}
	}

	require.NotNil(t, stsServerDeployment)

	expectedDiskSpace := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m"}

	require.Contains(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedDiskSpace)
}

func TestServerJavaHeapRenderWithAllJavaOptsOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/components_all_javaopts.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsServerDeployment appsv1.Deployment

	for _, deploymentK2ES := range resources.Deployments {
		if deploymentK2ES.Name == "stackstate-server" {
			stsServerDeployment = deploymentK2ES
		}
	}

	require.NotNil(t, stsServerDeployment)

	expectedDiskSpace := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005"}

	require.Contains(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedDiskSpace)
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

	expectedDiskSpace := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	require.Contains(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedDiskSpace)
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
	expectedDiskSpace := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5000"}

	require.Contains(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedDiskSpace)
}
