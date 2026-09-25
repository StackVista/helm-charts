package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestRbacAgentConnectionUsesAgentConfiguration(t *testing.T) {
	for _, release := range []string{"suse-observability-agent", "nightly"} {
		for _, custom := range []bool{false, true} {
			name := release + "/default"
			values := map[string]string{}
			if custom {
				name = release + "/external"
				values["kubernetes-rbac-agent.url.fromConfigMap"] = "external-url"
				values["kubernetes-rbac-agent.clusterName.fromConfigMap"] = "external-cluster"
			}
			t.Run(name, func(t *testing.T) {
				output := helmtestutil.RenderHelmTemplateOptsNoError(t, release, &helm.Options{
					ValuesFiles: []string{"values/minimal.yaml"}, SetValues: values,
					KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
				})
				resources := helmtestutil.NewKubernetesResources(t, output)
				agent := release + "-rbac-agent"
				require.Contains(t, resources.Deployments, agent)
				// This parent supplies external ConfigMap references to the child,
				// and must never receive the platform parent's inline defaults.
				assert.NotContains(t, resources.ConfigMaps, agent+"-url")
				assert.NotContains(t, resources.ConfigMaps, agent+"-cluster-name")
				refs := resources.Deployments[agent].Spec.Template.Spec.Containers[0].EnvFrom
				if custom {
					for _, name := range []string{"external-url", "external-cluster"} {
						assert.Contains(t, refs, corev1.EnvFromSource{
							ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: name}},
						})
					}
				} else {
					data := map[string]string{}
					for _, ref := range refs {
						if ref.ConfigMapRef != nil {
							require.Contains(t, resources.ConfigMaps, ref.ConfigMapRef.Name)
							for key, value := range resources.ConfigMaps[ref.ConfigMapRef.Name].Data {
								data[key] = value
							}
						}
					}
					assert.Equal(t, "https://my-suse-observability-instance.com/receiver", data["STS_URL"])
					assert.Equal(t, "some-k8s-cluster", data["STS_CLUSTER_NAME"])
				}
			})
		}
	}
}
