package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestShouldRenderConfigMapWithoutOverrides(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-agent", "values/minimal.yaml", "values/clustercheck_ksm_no_override.yaml")
	// Parse all resources into their corresponding types for validation and further inspection
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Contains(t, resources.ConfigMaps, "stackstate-agent-cluster-agent")
	cm := resources.ConfigMaps["stackstate-agent-cluster-agent"]
	assert.Contains(t, cm.Data, "kubernetes_state_conf")
	assert.Contains(t, cm.Data["kubernetes_state_conf"], "kube_state_url: http://stackstate-agent-kube-state-metrics.")
}

func TestShouldRenderConfigMapWithKSMWithCustomUrl(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-agent", "values/minimal.yaml", "values/clustercheck_ksm_custom_url.yaml")
	// Parse all resources into their corresponding types for validation and further inspection
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Contains(t, resources.ConfigMaps, "stackstate-agent-cluster-agent")
	cm := resources.ConfigMaps["stackstate-agent-cluster-agent"]
	assert.Contains(t, cm.Data, "kubernetes_state_conf")
	assert.Contains(t, cm.Data["kubernetes_state_conf"], "kube_state_url: http://my-custom-ksm-url.")
}
func TestShouldRenderConfigMapWithoutKSMWithCustomUrl(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-agent", "values/minimal.yaml", "values/clustercheck_no_ksm_custom_url.yaml")
	// Parse all resources into their corresponding types for validation and further inspection
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Contains(t, resources.ConfigMaps, "stackstate-agent-cluster-agent")
	cm := resources.ConfigMaps["stackstate-agent-cluster-agent"]
	assert.Contains(t, cm.Data, "kubernetes_state_conf")
	assert.Contains(t, cm.Data["kubernetes_state_conf"], "kube_state_url: http://my-custom-ksm-url.")
	assert.NotContains(t, resources.Deployments, "stackstate-agent-kube-state-metrics")
}
func TestShouldRenderConfigMapWithKSMWithOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate-agent", "values/minimal.yaml", "values/clustercheck_ksm_override.yaml")
	// Parse all resources into their corresponding types for validation and further inspection
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Contains(t, resources.ConfigMaps, "stackstate-agent-cluster-agent")
	cm := resources.ConfigMaps["stackstate-agent-cluster-agent"]
	assert.Contains(t, cm.Data, "_etc_stackstate-agent_conf.d_kubernetes_state.d_conf.yaml")
	assert.Contains(t, cm.Data["_etc_stackstate-agent_conf.d_kubernetes_state.d_conf.yaml"], "http://YOUR_KUBE_STATE_METRICS_SERVICE_NAME:8080/metrics")
}
