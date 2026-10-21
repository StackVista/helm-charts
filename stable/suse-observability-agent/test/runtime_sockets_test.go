package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestRuntimeSocketMounts(t *testing.T) {
	legacySockets := map[string]string{
		"containerdsocket": "/var/run/containerd/containerd.sock",
		"crisocket":        "/var/run/crio/crio.sock",
		"dockersocket":     "/var/run/docker.sock",
	}
	for _, tc := range []struct {
		name, kubeVersion, customSocket, envSocket, mountDefaults string
	}{
		{name: "default"},
		{name: "explicit-defaults", mountDefaults: "true"},
		{name: "rke2-defaults", kubeVersion: "v1.36.3+rke2r1"},
		{name: "k3s-defaults", kubeVersion: "v1.34.1+k3s1"},
		{name: "custom-additive", customSocket: "/custom/containerd.sock"},
		{name: "custom-explicit-defaults", customSocket: "/custom/containerd.sock", mountDefaults: "true"},
		{name: "cri-dockerd-additive", customSocket: "/run/cri-dockerd.sock"},
		{name: "custom-directory-additive", customSocket: "/var/run/containerd"},
		{name: "containerd-only", customSocket: "/var/run/containerd/containerd.sock", mountDefaults: "false"},
		{name: "crio-only", customSocket: "/var/run/crio/crio.sock", mountDefaults: "false"},
		{name: "docker-only", customSocket: "/var/run/docker.sock", mountDefaults: "false"},
		{name: "rke2-containerd-only", kubeVersion: "v1.36.3+rke2r1", customSocket: "/run/k3s/containerd/containerd.sock", mountDefaults: "false"},
		{name: "k3s-containerd-only", kubeVersion: "v1.34.1+k3s1", customSocket: "/run/k3s/containerd/containerd.sock", mountDefaults: "false"},
		{name: "no-sockets", mountDefaults: "false"},
		{name: "explicit-env", envSocket: "/var/run/containerd/containerd.sock"},
		{name: "rke2-explicit-env", kubeVersion: "v1.36.3+rke2r1", envSocket: "/var/run/containerd/containerd.sock"},
	} {
		for _, processEnabled := range []string{"true", "false"} {
			t.Run(tc.name+"/process="+processEnabled, func(t *testing.T) {
				values := map[string]string{
					processAgentEnabledKey:                        processEnabled,
					"nodeAgent.containerRuntime.customSocketPath": tc.customSocket,
				}
				if tc.mountDefaults != "" {
					values["nodeAgent.containerRuntime.mountDefaultSockets"] = tc.mountDefaults
				}
				if tc.envSocket != "" {
					values["nodeAgent.containers.agent.env.STS_CRI_SOCKET_PATH"] = tc.envSocket
					values["nodeAgent.containers.processAgent.env.STS_CRI_SOCKET_PATH"] = tc.envSocket
				}
				var args []string
				if tc.kubeVersion != "" {
					args = []string{"--kube-version", tc.kubeVersion}
				}
				output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, releaseName, &helm.Options{
					ValuesFiles: []string{"values/minimal.yaml"}, SetValues: values,
				}, args...)
				resources := helmtestutil.NewKubernetesResources(t, output)
				ds, ok := resources.DaemonSets[fullNodeAgentName]
				require.True(t, ok)

				volumes := map[string]corev1.Volume{}
				for _, volume := range ds.Spec.Template.Spec.Volumes {
					volumes[volume.Name] = volume
				}
				assert.NotContains(t, volumes, "runtimesockets", "avoid exposing all host runtime sockets")
				expectedMounts := map[string]string{}
				for name, path := range legacySockets {
					if tc.mountDefaults == "false" {
						assert.NotContains(t, volumes, name)
					} else {
						expectedMounts[name] = path
					}
				}
				if tc.customSocket != "" {
					expectedMounts["customcrisocket"] = tc.customSocket
				} else {
					assert.NotContains(t, volumes, "customcrisocket")
				}
				for name, path := range expectedMounts {
					require.Contains(t, volumes, name)
					require.NotNil(t, volumes[name].HostPath)
					assert.Equal(t, path, volumes[name].HostPath.Path)
					expectedType := corev1.HostPathUnset
					if name == "customcrisocket" && tc.mountDefaults == "false" {
						expectedType = corev1.HostPathSocket
					}
					actualType := corev1.HostPathUnset
					if volumes[name].HostPath.Type != nil {
						actualType = *volumes[name].HostPath.Type
					}
					assert.Equal(t, expectedType, actualType, name)
				}

				expectedContainers := 1
				if processEnabled == "true" {
					expectedContainers = 2
				}
				require.Len(t, ds.Spec.Template.Spec.Containers, expectedContainers)
				for _, container := range ds.Spec.Template.Spec.Containers {
					for name, path := range expectedMounts {
						assert.Contains(t, container.VolumeMounts, corev1.VolumeMount{Name: name, MountPath: path, ReadOnly: true}, container.Name)
					}
					for _, mount := range container.VolumeMounts {
						assert.Contains(t, volumes, mount.Name, container.Name)
					}
					env := map[string]string{}
					for _, variable := range container.Env {
						env[variable.Name] = variable.Value
					}
					if tc.envSocket != "" {
						assert.Equal(t, tc.envSocket, env["STS_CRI_SOCKET_PATH"], container.Name)
					} else if tc.customSocket != "" {
						assert.Equal(t, tc.customSocket, env["STS_CRI_SOCKET_PATH"], container.Name)
					} else {
						assert.NotContains(t, env, "STS_CRI_SOCKET_PATH", container.Name)
					}
				}
			})
		}
	}
}

func TestRuntimeSocketOverridesKeepDefaultMounts(t *testing.T) {
	for name, values := range map[string]map[string]string{
		"global-env":      {"global.extraEnv.open.STS_CRI_SOCKET_PATH": "/var/run/containerd/containerd.sock"},
		"secret-env":      {"global.extraEnv.secret.STS_CRI_SOCKET_PATH": "/var/run/containerd/containerd.sock"},
		"process-env":     {"nodeAgent.containers.processAgent.env.DD_CRI_SOCKET_PATH": "/var/run/containerd/containerd.sock"},
		"docker-env":      {"nodeAgent.containers.agent.env.DOCKER_HOST": "unix:///var/run/docker.sock"},
		"config-override": {"nodeAgent.config.override[0].name": "stackstate.yaml", "nodeAgent.config.override[0].path": "/etc/stackstate-agent", "nodeAgent.config.override[0].data": "cri_socket_path: /var/run/containerd/containerd.sock"},
	} {
		t.Run(name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, releaseName, &helm.Options{
				ValuesFiles: []string{"values/minimal.yaml"}, SetValues: values,
			}, "--kube-version", "v1.36.3+rke2r1")
			resources := helmtestutil.NewKubernetesResources(t, output)
			ds, ok := resources.DaemonSets[fullNodeAgentName]
			require.True(t, ok)
			var names []string
			for _, volume := range ds.Spec.Template.Spec.Volumes {
				names = append(names, volume.Name)
			}
			assert.NotContains(t, names, "customcrisocket")
			assert.Subset(t, names, []string{"crisocket", "containerdsocket", "dockersocket"})
		})
	}
}

func TestRuntimeSocketMountDefaultSocketsRequiresBoolean(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, releaseName, &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml"},
		SetStrValues: map[string]string{
			"nodeAgent.containerRuntime.mountDefaultSockets": "false",
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mountDefaultSockets")
}
