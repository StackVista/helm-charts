package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Split receiver deployment names
var splitReceiverDeployments = []string{
	"suse-observability-receiver-base",
	"suse-observability-receiver-logs",
	"suse-observability-receiver-process-agent",
}

// Expected CPU requests for split mode (in millicores)
var splitReceiverCPURequests = map[string]int64{
	"suse-observability-receiver-base":          666000,
	"suse-observability-receiver-logs":          777000,
	"suse-observability-receiver-process-agent": 888000,
}

// Expected specific env vars for split mode
var splitReceiverEnvVars = map[string]corev1.EnvVar{
	"suse-observability-receiver-base":          {Name: "receiverBase", Value: "test"},
	"suse-observability-receiver-logs":          {Name: "receiverLogs", Value: "test"},
	"suse-observability-receiver-process-agent": {Name: "receiverProcessAgent", Value: "test"},
}

// Common env vars expected in all receiver deployments
var commonReceiverEnvVars = []corev1.EnvVar{
	{Name: "all", Value: "test"},
	{Name: "receiverRoot", Value: "test"},
}

// assertSplitReceiverDeployments verifies split receiver deployments exist with correct resources and env vars
func assertSplitReceiverDeployments(t *testing.T, resources helmtestutil.KubernetesResources) {
	for _, deploymentName := range splitReceiverDeployments {
		deployment, ok := resources.Deployments[deploymentName]
		require.True(t, ok, "Deployment %s should exist", deploymentName)

		assertReceiverContainerResources(t, deployment, deploymentName, splitReceiverCPURequests[deploymentName])
		assertReceiverContainerEnvVars(t, deployment, deploymentName, commonReceiverEnvVars)
		assertReceiverContainerEnvVars(t, deployment, deploymentName, []corev1.EnvVar{splitReceiverEnvVars[deploymentName]})
	}
}

// assertNonSplitReceiverDeployment verifies non-split receiver deployment exists with correct resources and env vars
func assertNonSplitReceiverDeployment(t *testing.T, resources helmtestutil.KubernetesResources) {
	deploymentName := "suse-observability-receiver"
	deployment, ok := resources.Deployments[deploymentName]
	require.True(t, ok, "Deployment %s should exist", deploymentName)

	// CPU request is 555 (555000 millicores)
	assertReceiverContainerResources(t, deployment, deploymentName, 555000)
	assertReceiverContainerEnvVars(t, deployment, deploymentName, commonReceiverEnvVars)
}

// assertReceiverContainerResources verifies the first container has the expected CPU request
func assertReceiverContainerResources(t *testing.T, deployment appsv1.Deployment, deploymentName string, expectedCPUMillicores int64) {
	containers := deployment.Spec.Template.Spec.Containers
	require.NotEmpty(t, containers, "Deployment %s should have containers", deploymentName)

	cpuRequest := containers[0].Resources.Requests.Cpu()
	require.NotNil(t, cpuRequest, "Deployment %s should have CPU request", deploymentName)
	assert.Equal(t, expectedCPUMillicores, cpuRequest.MilliValue(), "Deployment %s CPU request should match expected value", deploymentName)
}

// assertReceiverContainerEnvVars verifies the first container has the expected environment variables
func assertReceiverContainerEnvVars(t *testing.T, deployment appsv1.Deployment, deploymentName string, expectedEnvVars []corev1.EnvVar) {
	containers := deployment.Spec.Template.Spec.Containers
	require.NotEmpty(t, containers, "Deployment %s should have containers", deploymentName)

	env := containers[0].Env
	for _, expectedEnv := range expectedEnvVars {
		assert.Contains(t, env, expectedEnv, "Deployment %s should have env var %s=%s", deploymentName, expectedEnv.Name, expectedEnv.Value)
	}
}

func TestReceiverSplitLegacy(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/receiver_split_legacy.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assertSplitReceiverDeployments(t, resources)
}

func TestReceiverNonSplitLegacy(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/receiver_nonsplit_legacy.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assertNonSplitReceiverDeployment(t, resources)
}

func TestReceiverSplitGlobal(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/receiver_split_global.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assertSplitReceiverDeployments(t, resources)
}

func TestReceiverNonSplitGlobal(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/receiver_nonsplit_global.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assertNonSplitReceiverDeployment(t, resources)
}
