package test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	v1 "k8s.io/api/core/v1"
)

func TestStackpacks2EnabledNonSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/split_disabled.yaml", "values/stackpacks2_enabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	server := resources.Deployments["suse-observability-server"]

	// Stackpacks 2 feature flag Env var should be set
	assert.GreaterOrEqual(t, slices.IndexFunc(server.Spec.Template.Spec.Containers[0].Env, func(env v1.EnvVar) bool {
		t.Log(env.Name, ":", env.Value)
		return env.Name == "CONFIG_FORCE_stackstate_featureSwitches_enableStackPacks2"
	}), 0)

	// Stackpacks 2 is auto-upgraded
	serverConfigmap := resources.ConfigMaps["suse-observability-server"]
	assert.Regexp(t, "upgradeOnStartUp = \\[.*,\"open-telemetry-2\"", serverConfigmap.Data["application_stackstate.conf"])

	// Stackpacks 1 Docker image should be mounted
	stackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-v1" })
	assert.Regexp(t, ".*/stackstate/stackpacks:.*", server.Spec.Template.Spec.InitContainers[stackpacksInitIdx].Image)
	// Stackpacks 2 Contrib Docker image should be mounted
	contribStackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-contrib" })
	assert.Regexp(t, ".*/stackstate/contrib-stackpacks:.*", server.Spec.Template.Spec.InitContainers[contribStackpacksInitIdx].Image)
	// Stackpacks 2 SUSE Docker image should be mounted
	suseStackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-suse" })
	assert.Regexp(t, ".*/stackstate/suse-stackpacks:.*", server.Spec.Template.Spec.InitContainers[suseStackpacksInitIdx].Image)
}

func TestStackpacks2DisabledNonSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	server := resources.Deployments["suse-observability-server"]

	// Stackpacks 2 feature flag Env var should not be set
	assert.Equal(t, -1, slices.IndexFunc(server.Spec.Template.Spec.Containers[0].Env, func(env v1.EnvVar) bool {
		return env.Name == "CONFIG_FORCE_stackstate_featureSwitches_enableStackPacks2"
	}))

	// Stackpacks 2 is not auto-upgraded
	serverConfigmap := resources.ConfigMaps["suse-observability-server"]
	assert.NotRegexp(t, "upgradeOnStartUp = \\[.*,\"open-telemetry-2\"", serverConfigmap.Data["application_stackstate.conf"])

	// Stackpacks 1 Docker image should be mounted
	stackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return (container.Name == "init-stackpacks-v1") })
	assert.Regexp(t, ".*/stackstate/stackpacks:.*", server.Spec.Template.Spec.InitContainers[stackpacksInitIdx].Image)

	// Stackpacks 2 Contrib docker image should be absent
	assert.NotContains(t, server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-contrib" })
}

func TestStackpacks2EnabledSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/stackpacks2_enabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	api := resources.Deployments["suse-observability-api"]

	// Stackpacks 2 feature flag Env var should be set
	assert.GreaterOrEqual(t, slices.IndexFunc(api.Spec.Template.Spec.Containers[0].Env, func(env v1.EnvVar) bool {
		return env.Name == "CONFIG_FORCE_stackstate_featureSwitches_enableStackPacks2"
	}), 0)

	// Stackpacks 2 is auto-upgraded
	serverConfigmap := resources.ConfigMaps["suse-observability-api"]
	assert.Regexp(t, "upgradeOnStartUp = \\[.*,\"open-telemetry-2\"", serverConfigmap.Data["application_stackstate.conf"])

	// Stackpacks 1 Docker image should be mounted
	stackpacksInitIdx := slices.IndexFunc(api.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-v1" })
	assert.Regexp(t, ".*/stackstate/stackpacks:.*", api.Spec.Template.Spec.InitContainers[stackpacksInitIdx].Image)
	// Stackpacks 2 Contrib Docker image should be mounted
	contribStackpacksInitIdx := slices.IndexFunc(api.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-contrib" })
	assert.Regexp(t, ".*/stackstate/contrib-stackpacks:.*", api.Spec.Template.Spec.InitContainers[contribStackpacksInitIdx].Image)
}

func TestStackpacks2DisabledSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	api := resources.Deployments["suse-observability-api"]

	// Stackpacks 2 feature flag Env var should not be set
	assert.Equal(t, -1, slices.IndexFunc(api.Spec.Template.Spec.Containers[0].Env, func(env v1.EnvVar) bool {
		return env.Name == "CONFIG_FORCE_stackstate_featureSwitches_enableStackPacks2"
	}))

	// Stackpacks 2 is not auto-upgraded
	serverConfigmap := resources.ConfigMaps["suse-observability-api"]
	assert.NotRegexp(t, "upgradeOnStartUp = \\[.*,\"open-telemetry-2\"", serverConfigmap.Data["application_stackstate.conf"])

	// Stackpacks 1 Docker image should be mounted
	stackpacksInitIdx := slices.IndexFunc(api.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-v1" })
	assert.Regexp(t, ".*/stackstate/stackpacks:.*", api.Spec.Template.Spec.InitContainers[stackpacksInitIdx].Image)

	// Stackpacks 2 Contrib docker image should be absent
	assert.NotContains(t, api.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-contrib" })
}

func TestStackpacks2EnabledNonSplitCommunity(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/split_disabled.yaml", "values/stackpacks2_enabled.yaml", "values/community.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	server := resources.Deployments["suse-observability-server"]

	// Stackpacks 2 feature flag Env var should be set
	assert.GreaterOrEqual(t, slices.IndexFunc(server.Spec.Template.Spec.Containers[0].Env, func(env v1.EnvVar) bool {
		t.Log(env.Name, ":", env.Value)
		return env.Name == "CONFIG_FORCE_stackstate_featureSwitches_enableStackPacks2"
	}), 0)

	// Stackpacks 2 is auto-upgraded
	serverConfigmap := resources.ConfigMaps["suse-observability-server"]
	assert.Regexp(t, "upgradeOnStartUp = \\[.*,\"open-telemetry-2\"", serverConfigmap.Data["application_stackstate.conf"])

	// Stackpacks 1 Docker image should be mounted
	stackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-v1" })
	assert.Regexp(t, ".*/stackstate/stackpacks:.*", server.Spec.Template.Spec.InitContainers[stackpacksInitIdx].Image)
	// Stackpacks 2 Contrib Docker image should be mounted
	contribStackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-contrib" })
	assert.Regexp(t, ".*/stackstate/contrib-stackpacks:.*", server.Spec.Template.Spec.InitContainers[contribStackpacksInitIdx].Image)

	// Stackpacks 2 SUSE Docker image should be absent
	assert.NotContains(t, server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-suse" })
}
