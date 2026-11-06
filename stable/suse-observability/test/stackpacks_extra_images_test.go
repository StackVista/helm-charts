package test

import (
	"slices"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	v1 "k8s.io/api/core/v1"
)

func TestExtraStackpackImages(t *testing.T) {
	for _, tc := range []struct {
		name       string
		deployment string
		values     []string
		prime      bool
	}{
		{"PrimeSplit", "suse-observability-api", nil, true},
		{"PrimeNonSplit", "suse-observability-server", []string{"values/split_disabled.yaml"}, true},
		{"CommunitySplit", "suse-observability-api", []string{"values/community.yaml"}, false},
		{"CommunityNonSplit", "suse-observability-server", []string{"values/community.yaml", "values/split_disabled.yaml"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := append([]string{"values/full.yaml", "values/extra_stackpacks.yaml"}, tc.values...)
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", values...)
			resources := helmtestutil.NewKubernetesResources(t, output)
			deployment, ok := resources.Deployments[tc.deployment]
			require.True(t, ok)
			pod := deployment.Spec.Template.Spec

			var stackpacks []v1.Container
			for _, container := range pod.InitContainers {
				if slices.Contains(container.Command, "/stackpack-scripts/copy-stackpacks.sh") {
					stackpacks = append(stackpacks, container)
				}
			}
			expected := []string{"init-stackpacks-v1", "init-stackpacks-core"}
			if tc.prime {
				expected = append(expected, "init-stackpacks-suse")
			}
			expected = append(expected, "init-stackpacks-community-example")
			if tc.prime {
				expected = append(expected, "init-stackpacks-private-example")
			}
			expected = append(expected, "init-stackpacks-inherited-registry")
			var names []string
			for _, container := range stackpacks {
				names = append(names, container.Name)
			}
			require.Equal(t, expected, names)
			assert.Contains(t, stackpacks[1].Image, "my.registry.com/stackstate/core-stackpacks:")
			assert.Contains(t, stackpacks[0].Args, "--clear")

			for _, container := range stackpacks[1:] {
				assert.NotContains(t, container.Args, "--clear")
			}
			extraIndex := slices.IndexFunc(stackpacks, func(container v1.Container) bool {
				return container.Name == "init-stackpacks-community-example"
			})
			require.GreaterOrEqual(t, extraIndex, 0)
			extra := stackpacks[extraIndex]
			assert.Equal(t, "ghcr.io/stackvista/contrib-stackpacks/my-stackpack:0.0.1", extra.Image)
			assert.Equal(t, v1.PullIfNotPresent, extra.ImagePullPolicy)
			assert.Equal(t, []string{"/bin/sh", "/stackpack-scripts/copy-stackpacks.sh"}, extra.Command)
			assert.Equal(t, []string{"/var/stackpacks"}, extra.Args)
			assert.Equal(t, stackpacks[0].SecurityContext, extra.SecurityContext)
			assert.Equal(t, stackpacks[0].VolumeMounts, extra.VolumeMounts)
			assert.Contains(t, pod.ImagePullSecrets, v1.LocalObjectReference{Name: "stackpack-registry"})
			assert.Equal(t, "my.registry.com/integrations/inherited-stackpack:2.0.0", stackpacks[len(stackpacks)-1].Image)

			if tc.prime {
				private := stackpacks[len(stackpacks)-2]
				assert.Equal(t, "registry.example.com/integrations/private-stackpack:1.2.3", private.Image)
				assert.Equal(t, v1.PullAlways, private.ImagePullPolicy)
			}
		})
	}
}

func TestCoreStackpackImageRegistry(t *testing.T) {
	for _, tc := range []struct {
		name     string
		values   string
		registry string
		expected string
	}{
		{"Default", "values/full.yaml", "", "quay.io"},
		{"SaaSGlobalRegistry", "values/workload_global_nonha.yaml", "quay.io", "quay.io"},
		{"GlobalProfile", "values/workload_global_nonha.yaml", "", "registry.rancher.com"},
		{"CustomGlobalRegistry", "values/workload_global_nonha.yaml", "mirror.example.com", "mirror.example.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
				ValuesFiles: []string{tc.values},
				SetValues:   map[string]string{"global.imageRegistry": tc.registry},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			deployment := resources.Deployments["suse-observability-api"]
			if tc.name != "Default" {
				deployment = resources.Deployments["suse-observability-server"]
			}
			require.NotNil(t, deployment)
			idx := slices.IndexFunc(deployment.Spec.Template.Spec.InitContainers, func(container v1.Container) bool {
				return container.Name == "init-stackpacks-core"
			})
			require.GreaterOrEqual(t, idx, 0)
			assert.Contains(t, deployment.Spec.Template.Spec.InitContainers[idx].Image, tc.expected+"/stackstate/core-stackpacks:")
		})
	}
}
