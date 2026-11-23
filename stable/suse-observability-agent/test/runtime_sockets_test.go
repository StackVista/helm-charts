package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestNoContainerRuntimeMounts(t *testing.T) {
	runtimePaths := []string{
		"/var/run/containerd/containerd.sock",
		"/run/k3s/containerd/containerd.sock",
		"/var/run/crio/crio.sock",
		"/var/run/docker.sock",
		"/run/docker/netns",
		"/var/lib/docker/overlay2",
	}
	for name, values := range map[string]map[string]string{
		"default": {},
		"legacy-custom-socket": {
			"nodeAgent.containerRuntime.customSocketPath": "/run/k3s/containerd/containerd.sock",
		},
		"legacy-custom-socket-only": {
			"nodeAgent.containerRuntime.customSocketPath":    "/run/k3s/containerd/containerd.sock",
			"nodeAgent.containerRuntime.mountDefaultSockets": "false",
		},
		"legacy-mount-defaults": {
			"nodeAgent.containerRuntime.mountDefaultSockets": "true",
		},
	} {
		for _, processEnabled := range []string{"true", "false"} {
			t.Run(name+"/process="+processEnabled, func(t *testing.T) {
				setValues := map[string]string{processAgentEnabledKey: processEnabled}
				for k, v := range values {
					setValues[k] = v
				}
				output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, releaseName, &helm.Options{
					ValuesFiles: []string{"values/minimal.yaml"}, SetValues: setValues,
				}, "--kube-version", "v1.36.3+rke2r1")
				resources := helmtestutil.NewKubernetesResources(t, output)
				ds, ok := resources.DaemonSets[fullNodeAgentName]
				require.True(t, ok)

				volumes := map[string]bool{}
				for _, volume := range ds.Spec.Template.Spec.Volumes {
					volumes[volume.Name] = true
					if volume.HostPath != nil {
						assert.NotContains(t, runtimePaths, volume.HostPath.Path, volume.Name)
					}
				}

				expectedContainers := 1
				if processEnabled == "true" {
					expectedContainers = 2
				}
				require.Len(t, ds.Spec.Template.Spec.Containers, expectedContainers)
				for _, container := range ds.Spec.Template.Spec.Containers {
					for _, mount := range container.VolumeMounts {
						assert.True(t, volumes[mount.Name], "%s mounts undefined volume %s", container.Name, mount.Name)
						assert.NotContains(t, runtimePaths, mount.MountPath, container.Name)
					}
					for _, variable := range container.Env {
						assert.NotEqual(t, "STS_CRI_SOCKET_PATH", variable.Name, container.Name)
					}
				}
			})
		}
	}
}
