package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestAgentEnabledRendersResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Contains(t, resources.Deployments, "suse-observability-agent-cluster-agent")
	assert.Contains(t, resources.Services, "suse-observability-agent-cluster-agent")
	assert.Contains(t, resources.ConfigMaps, "suse-observability-agent-cluster-agent")
	assert.Contains(t, resources.ServiceAccounts, "suse-observability-agent")
	assert.Contains(t, resources.ClusterRoles, "suse-observability-agent")
	assert.Contains(t, resources.ClusterRoleBindings, "suse-observability-agent")
	assert.Contains(t, resources.Roles, "suse-observability-agent")
	assert.Contains(t, resources.RoleBindings, "suse-observability-agent")
	assert.Contains(t, resources.Pdbs, "suse-observability-agent")

	assert.Contains(t, resources.DaemonSets, "suse-observability-agent-node-agent")
	assert.Contains(t, resources.ServiceAccounts, "suse-observability-agent-node-agent")
	assert.Contains(t, resources.ClusterRoles, "suse-observability-agent-node-agent")
	assert.Contains(t, resources.Services, "suse-observability-agent-node-agent")
}

func TestAgentFullyDisabledRendersNoAgentResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/agent-fully-disabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.NotContains(t, resources.Deployments, "suse-observability-agent-cluster-agent")
	assert.NotContains(t, resources.Services, "suse-observability-agent-cluster-agent")
	assert.NotContains(t, resources.ConfigMaps, "suse-observability-agent-cluster-agent")
	assert.NotContains(t, resources.ServiceAccounts, "suse-observability-agent")
	assert.NotContains(t, resources.ClusterRoles, "suse-observability-agent")
	assert.NotContains(t, resources.ClusterRoleBindings, "suse-observability-agent")
	assert.NotContains(t, resources.Roles, "suse-observability-agent")
	assert.NotContains(t, resources.RoleBindings, "suse-observability-agent")
	assert.NotContains(t, resources.Pdbs, "suse-observability-agent")

	assert.NotContains(t, resources.DaemonSets, "suse-observability-agent-node-agent")
	assert.NotContains(t, resources.ServiceAccounts, "suse-observability-agent-node-agent")
	assert.NotContains(t, resources.ClusterRoles, "suse-observability-agent-node-agent")
	assert.NotContains(t, resources.Services, "suse-observability-agent-node-agent")
}

// The agent is switched off as a whole rather than component by component. Leaving the node agent
// on without a cluster agent makes every node agent read cluster-wide metadata from the API server
// itself, which scales with node count.
//
// The checks agent is disabled here so only the node-agent validation can fire. Without that, both
// validations are eligible and the test passes on whichever runs first, so it would keep passing if
// the node-agent guard were removed and the ordering happened to change.
func TestNodeAgentWithoutClusterAgentIsRejected(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability-agent", "values/minimal.yaml", "values/cluster-and-checks-agent-disabled.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nodeAgent.enabled is true but clusterAgent.enabled is false")
}

// checksAgent.enabled defaults to true, so this is what a user gets by disabling the cluster agent
// and the node agent but forgetting the checks agent.
func TestChecksAgentWithoutClusterAgentIsRejected(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability-agent", "values/minimal.yaml", "values/cluster-agent-disabled.yaml", "values/node-agent-disabled.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksAgent.enabled is true but clusterAgent.enabled is false")
}

func TestNodeAgentDisabledOnItsOwnIsAllowed(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/node-agent-disabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.NotContains(t, resources.DaemonSets, "suse-observability-agent-node-agent")
	assert.Contains(t, resources.Deployments, "suse-observability-agent-cluster-agent")
}

// ConfigMap, SCC and VPA are off by default, so a disabled-state assertion that does not turn them
// on would pass even with a broken or missing guard on those three templates.
func TestNodeAgentDisabledSuppressesItsOptionalResourcesToo(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/agent-fully-disabled.yaml"},
		SetValues: map[string]string{
			"nodeAgent.scc.enabled":             "true",
			"nodeAgent.autoScalingEnabled":      "true",
			"nodeAgent.config.override[0].name": "sts.yaml",
			"nodeAgent.config.override[0].path": "/etc/stackstate-agent",
			"nodeAgent.config.override[0].data": "log_level: debug",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.NotContains(t, resources.ConfigMaps, "suse-observability-agent-node-agent")
	assert.NotContains(t, resources.ClusterRoleBindings, "suse-observability-agent-node-agent")
	assert.NotContains(t, output, "SecurityContextConstraints")
	assert.NotContains(t, output, "VerticalPodAutoscaler")
}

// The checks agent binds to the node agent's ClusterRole rather than owning one, so that role has
// to survive the node agent being disabled or the binding points at nothing and the checks agent
// silently loses its permissions.
func TestChecksAgentKeepsItsRoleWhenNodeAgentDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/node-agent-disabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	binding, found := resources.ClusterRoleBindings["suse-observability-agent-checks-agent"]
	require.True(t, found, "checks agent ClusterRoleBinding should still be rendered")
	assert.Contains(t, resources.ClusterRoles, binding.RoleRef.Name,
		"checks agent binding references ClusterRole %q, which is not rendered", binding.RoleRef.Name)
}

// The role is retained for the checks agent, so its contents must not keep granting node-agent
// permissions. Pod correlation is a process-agent feature; with no node agent there is no process
// agent, and the checks agent service account would otherwise get cluster-wide pods list/watch.
func TestRetainedNodeAgentRoleDropsProcessAgentPermissions(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/node-agent-disabled.yaml"},
		SetValues: map[string]string{
			"processAgent.podCorrelation.enabled":     "true",
			"processAgent.podCorrelation.remoteCache": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	role, found := resources.ClusterRoles["suse-observability-agent-node-agent"]
	require.True(t, found, "role should be retained for the checks agent")
	assertRuleExistence(t, role.Rules, "pods+get,list,watch", false)
}

// The remote kube cache exists only to serve the node agent's process-agent pod correlation, so it
// must not outlive the node agent. It watches the API, which is the cost of leaving it running.
func TestRemoteKubeCacheGoesWithTheNodeAgent(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/agent-fully-disabled.yaml"},
		SetValues: map[string]string{
			"processAgent.podCorrelation.enabled":     "true",
			"processAgent.podCorrelation.remoteCache": "true",
		},
	})

	assert.NotContains(t, output, "remote-kube-cache")
}
