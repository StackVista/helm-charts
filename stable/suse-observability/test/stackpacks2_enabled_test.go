package test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// otel-k8s-crd is auto-upgraded when StackPacks 2 is enabled
	serverConfigmap := resources.ConfigMaps["suse-observability-server"]
	assert.Regexp(t, "upgradeOnStartUp = \\[.*,\"otel-k8s-crd\"", serverConfigmap.Data["application_stackstate.conf"])

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

	// otel-k8s-crd is not auto-upgraded when StackPacks 2 is disabled
	serverConfigmap := resources.ConfigMaps["suse-observability-server"]
	assert.NotRegexp(t, "upgradeOnStartUp = \\[.*,\"otel-k8s-crd\"", serverConfigmap.Data["application_stackstate.conf"])

	// Stackpacks 1 Docker image should be mounted
	stackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return (container.Name == "init-stackpacks-v1") })
	assert.Regexp(t, ".*/stackstate/stackpacks:.*", server.Spec.Template.Spec.InitContainers[stackpacksInitIdx].Image)

	// Stackpacks 2 Contrib docker image should be absent
	assert.Equal(t, -1, slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool {
		return container.Name == "init-stackpacks-contrib"
	}))
}

func TestStackpacks2EnabledSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/stackpacks2_enabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)

	api := resources.Deployments["suse-observability-api"]

	// Stackpacks 2 feature flag Env var should be set
	assert.GreaterOrEqual(t, slices.IndexFunc(api.Spec.Template.Spec.Containers[0].Env, func(env v1.EnvVar) bool {
		return env.Name == "CONFIG_FORCE_stackstate_featureSwitches_enableStackPacks2"
	}), 0)

	// otel-k8s-crd is auto-upgraded when StackPacks 2 is enabled
	serverConfigmap := resources.ConfigMaps["suse-observability-api"]
	assert.Regexp(t, "upgradeOnStartUp = \\[.*,\"otel-k8s-crd\"", serverConfigmap.Data["application_stackstate.conf"])

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

	// otel-k8s-crd is not auto-upgraded when StackPacks 2 is disabled
	serverConfigmap := resources.ConfigMaps["suse-observability-api"]
	assert.NotRegexp(t, "upgradeOnStartUp = \\[.*,\"otel-k8s-crd\"", serverConfigmap.Data["application_stackstate.conf"])

	// Stackpacks 1 Docker image should be mounted
	stackpacksInitIdx := slices.IndexFunc(api.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-v1" })
	assert.Regexp(t, ".*/stackstate/stackpacks:.*", api.Spec.Template.Spec.InitContainers[stackpacksInitIdx].Image)

	// Stackpacks 2 Contrib docker image should be absent
	assert.Equal(t, -1, slices.IndexFunc(api.Spec.Template.Spec.InitContainers, func(container v1.Container) bool {
		return container.Name == "init-stackpacks-contrib"
	}))
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

	// otel-k8s-crd is auto-upgraded when StackPacks 2 is enabled
	serverConfigmap := resources.ConfigMaps["suse-observability-server"]
	assert.Regexp(t, "upgradeOnStartUp = \\[.*,\"otel-k8s-crd\"", serverConfigmap.Data["application_stackstate.conf"])

	// Stackpacks 1 Docker image should be mounted
	stackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-v1" })
	assert.Regexp(t, ".*/stackstate/stackpacks:.*", server.Spec.Template.Spec.InitContainers[stackpacksInitIdx].Image)
	// Stackpacks 2 Contrib Docker image should be mounted
	contribStackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-contrib" })
	assert.Regexp(t, ".*/stackstate/contrib-stackpacks:.*", server.Spec.Template.Spec.InitContainers[contribStackpacksInitIdx].Image)

	// Stackpacks 2 SUSE Docker image should be absent
	assert.Equal(t, -1, slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool {
		return container.Name == "init-stackpacks-suse"
	}))
}

func TestStackpacks2EnabledNonSplitPrimeWithInternalStackpacks(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/split_disabled.yaml", "values/stackpacks2_enabled.yaml", "values/internal_stackpacks_enabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)
	server := resources.Deployments["suse-observability-server"]

	stackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-v1" })
	require.GreaterOrEqual(t, stackpacksInitIdx, 0)
	assert.Regexp(t, ".*/stackstate/stackpacks:.*", server.Spec.Template.Spec.InitContainers[stackpacksInitIdx].Image)

	contribStackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-contrib" })
	require.GreaterOrEqual(t, contribStackpacksInitIdx, 0)
	assert.Regexp(t, ".*/stackstate/contrib-stackpacks:.*", server.Spec.Template.Spec.InitContainers[contribStackpacksInitIdx].Image)

	suseStackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-suse" })
	require.GreaterOrEqual(t, suseStackpacksInitIdx, 0)
	assert.Regexp(t, ".*/stackstate/suse-stackpacks:.*", server.Spec.Template.Spec.InitContainers[suseStackpacksInitIdx].Image)

	internalStackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-internal" })
	require.GreaterOrEqual(t, internalStackpacksInitIdx, 0)
	assert.Regexp(t, ".*/stackstate/internal-stackpacks:test-tag", server.Spec.Template.Spec.InitContainers[internalStackpacksInitIdx].Image)
}

func TestStackpacksInitContainersUseChartScript(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/split_disabled.yaml")

	resources := helmtestutil.NewKubernetesResources(t, output)
	server := resources.Deployments["suse-observability-server"]

	stackpacksInitIdx := slices.IndexFunc(server.Spec.Template.Spec.InitContainers, func(container v1.Container) bool { return container.Name == "init-stackpacks-v1" })
	require.GreaterOrEqual(t, stackpacksInitIdx, 0)
	stackpacksInitContainer := server.Spec.Template.Spec.InitContainers[stackpacksInitIdx]

	assert.Equal(t, []string{"/bin/sh", "/stackpack-scripts/copy-stackpacks.sh"}, stackpacksInitContainer.Command)
	assert.Contains(t, stackpacksInitContainer.Args, "/var/stackpacks")
	assert.Contains(t, stackpacksInitContainer.Args, "--clear")

	assert.GreaterOrEqual(t, slices.IndexFunc(stackpacksInitContainer.VolumeMounts, func(volumeMount v1.VolumeMount) bool {
		return volumeMount.Name == "stackpack-scripts" && volumeMount.MountPath == "/stackpack-scripts"
	}), 0)
	assert.GreaterOrEqual(t, slices.IndexFunc(server.Spec.Template.Spec.Volumes, func(volume v1.Volume) bool {
		return volume.Name == "stackpack-scripts" && volume.ConfigMap != nil && volume.ConfigMap.Name == "suse-observability-stackpacks-scripts"
	}), 0)

	stackpacksScripts := resources.ConfigMaps["suse-observability-stackpacks-scripts"]
	assert.Contains(t, stackpacksScripts.Data["copy-stackpacks.sh"], "Copying StackPacks from /stackpacks")
}
