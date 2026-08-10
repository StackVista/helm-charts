package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestPodAnnotationsAndLabels(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/pod-annotations.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	podTemplates := map[string]corev1.PodTemplateSpec{}
	for _, name := range []string{
		"suse-observability-agent-cluster-agent",
		"suse-observability-agent-checks-agent",
		"suse-observability-agent-remote-kube-cache",
	} {
		deployment, exists := resources.Deployments[name]
		require.True(t, exists, "deployment %s was not found", name)
		podTemplates[name] = deployment.Spec.Template
	}
	for _, name := range []string{
		"suse-observability-agent-node-agent",
		"suse-observability-agent-logs-agent",
	} {
		daemonSet, exists := resources.DaemonSets[name]
		require.True(t, exists, "daemonset %s was not found", name)
		podTemplates[name] = daemonSet.Spec.Template
	}

	// Distinct owner/workload values per workload so a template wired to the wrong
	// values key fails here instead of passing on an identical value.
	expectedOwner := map[string]string{
		"suse-observability-agent-cluster-agent":     "cluster-agent",
		"suse-observability-agent-node-agent":        "node-agent",
		"suse-observability-agent-checks-agent":      "checks-agent",
		"suse-observability-agent-logs-agent":        "logs-agent",
		"suse-observability-agent-remote-kube-cache": "remote-kube-cache",
	}

	for name, expected := range expectedOwner {
		template := podTemplates[name]

		assert.Equal(t, expected, template.Annotations["example.com/owner"],
			"%s has wrong podAnnotations", name)
		assert.Equal(t, expected, template.Labels["example.com/workload"],
			"%s has wrong podLabels", name)

		assert.Equal(t, "global-value", template.Annotations["global.example.com/annotation"],
			"%s lost global.extraAnnotations", name)
		assert.Equal(t, "global-value", template.Labels["global.example.com/label"],
			"%s lost global.extraLabels", name)
	}

	// The monitor annotation from the originating issue must survive rendering intact,
	// and only on the workloads that set it.
	const throttlingAnnotation = "monitor.kubernetes-v2.stackstate.io/pod-cpu-throttling"
	assert.Equal(t, `{ "enabled": false}`,
		podTemplates["suse-observability-agent-cluster-agent"].Annotations[throttlingAnnotation])
	assert.Equal(t, `{ "enabled": false}`,
		podTemplates["suse-observability-agent-node-agent"].Annotations[throttlingAnnotation])
	assert.NotContains(t, podTemplates["suse-observability-agent-checks-agent"].Annotations, throttlingAnnotation)
}

// These two ConfigMaps pipe their body through fromYaml, so malformed YAML degrades into an
// error document instead of failing the render. A non-empty global.extraLabels or
// global.extraAnnotations used to trigger exactly that, silently dropping the cluster name
// and receiver URL.
func TestGlobalExtraMetadataKeepsConfigMapsValid(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/pod-annotations.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	clusterName, exists := resources.ConfigMaps["suse-observability-agent-cluster-name"]
	require.True(t, exists, "cluster-name config map was not found")
	assert.Equal(t, "some-k8s-cluster", clusterName.Data["STS_CLUSTER_NAME"])
	assert.Equal(t, "global-value", clusterName.Labels["global.example.com/label"])
	assert.Equal(t, "global-value", clusterName.Annotations["global.example.com/annotation"])

	url, exists := resources.ConfigMaps["suse-observability-agent-url"]
	require.True(t, exists, "url config map was not found")
	assert.NotEmpty(t, url.Data["STS_URL"])
	assert.Equal(t, "global-value", url.Labels["global.example.com/label"])
	assert.Equal(t, "global-value", url.Annotations["global.example.com/annotation"])
}
