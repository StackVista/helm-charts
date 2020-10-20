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
	var stsReceiverDeployment appsv1.Deployment
	var stsCorrelateDeployment appsv1.Deployment
	var stsMm2EsDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-server" {
			stsServerDeployment = deployment
		}
		if deployment.Name == "stackstate-receiver" {
			stsReceiverDeployment = deployment
		}
		if deployment.Name == "stackstate-correlate" {
			stsCorrelateDeployment = deployment
		}
		if deployment.Name == "stackstate-mm2es" {
			stsMm2EsDeployment = deployment
		}
	}

	require.NotNil(t, stsServerDeployment, stsReceiverDeployment, stsCorrelateDeployment, stsMm2EsDeployment)

	expectedServerJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx5858m -Xms5858m"}
	expectedReceiverJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1086m -Xms280m"}
	expectedCorrelateJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx799m -Xms799m"}
	expectedMm2EsJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx402m -Xms402m"}

	AssertJavaOpts(t, stsServerDeployment.Spec.Template.Spec.Containers[0].Env, expectedServerJavaOpts)
	AssertJavaOpts(t, stsReceiverDeployment.Spec.Template.Spec.Containers[0].Env, expectedReceiverJavaOpts)
	AssertJavaOpts(t, stsCorrelateDeployment.Spec.Template.Spec.Containers[0].Env, expectedCorrelateJavaOpts)
	AssertJavaOpts(t, stsMm2EsDeployment.Spec.Template.Spec.Containers[0].Env, expectedMm2EsJavaOpts)
}

func TestSplitServicesJavaHeapRender(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/split_enabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	var stsApiDeployment appsv1.Deployment
	var stsChecksDeployment appsv1.Deployment
	var stsStateDeployment appsv1.Deployment
	var stsSyncDeployment appsv1.Deployment
	var stsSlicingDeployment appsv1.Deployment
	var stsViewHealthDeployment appsv1.Deployment

	for _, deployment := range resources.Deployments {
		if deployment.Name == "stackstate-api" {
			stsApiDeployment = deployment
		}
		if deployment.Name == "stackstate-checks" {
			stsChecksDeployment = deployment
		}
		if deployment.Name == "stackstate-state" {
			stsStateDeployment = deployment
		}
		if deployment.Name == "stackstate-sync" {
			stsSyncDeployment = deployment
		}
		if deployment.Name == "stackstate-slicing" {
			stsSlicingDeployment = deployment
		}
		if deployment.Name == "stackstate-view-health" {
			stsViewHealthDeployment = deployment
		}
	}

	require.NotNil(t, stsApiDeployment, stsChecksDeployment, stsStateDeployment, stsSyncDeployment, stsSlicingDeployment, stsViewHealthDeployment)

	expectedApiJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx510m -Xms510m"}
	expectedChecksJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1620m -Xms1620m"}
	expectedStateJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx950m -Xms950m"}
	expectedSyncJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx2336m -Xms2336m"}
	expectedSlicingJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1000m -Xms1000m"}
	expectedViewHealthJavaOpts := v1.EnvVar{Name: "JAVA_OPTS", Value: "-Xmx1123m -Xms1123m"}

	AssertJavaOpts(t, stsApiDeployment.Spec.Template.Spec.Containers[0].Env, expectedApiJavaOpts)
	AssertJavaOpts(t, stsChecksDeployment.Spec.Template.Spec.Containers[0].Env, expectedChecksJavaOpts)
	AssertJavaOpts(t, stsStateDeployment.Spec.Template.Spec.Containers[0].Env, expectedStateJavaOpts)
	AssertJavaOpts(t, stsSyncDeployment.Spec.Template.Spec.Containers[0].Env, expectedSyncJavaOpts)
	AssertJavaOpts(t, stsSlicingDeployment.Spec.Template.Spec.Containers[0].Env, expectedSlicingJavaOpts)
	AssertJavaOpts(t, stsViewHealthDeployment.Spec.Template.Spec.Containers[0].Env, expectedViewHealthJavaOpts)
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
