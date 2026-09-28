package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestRbacAgentConnectionUsesPlatformConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name, release, values, url string
	}{
		{name: "default", release: "suse-observability", url: "http://suse-observability-router:8080/receiver/stsAgent"},
		{name: "custom-release", release: "nightly"},
		{name: "empty-maps", release: "nightly", values: "kubernetes-rbac-agent:\n  url: {}\n  clusterName: {}\n"},
		{name: "null-maps", release: "nightly", values: "kubernetes-rbac-agent:\n  url: null\n  clusterName: null\n"},
		{
			name: "naming-overrides", release: "nightly",
			values: "fullnameOverride: ignored-root\nglobal:\n  fullnameOverride: existing\n  fullnamePrefix: global-\n  fullnameSuffix: -end\n",
			url:    "http://global-existing-end-router:8080/receiver/stsAgent",
		},
		{
			name: "saved-literals-ignored", release: "nightly",
			values: "kubernetes-rbac-agent:\n  url:\n    value: http://nightly-suse-observability-receiver:7077/stsAgent\n  clusterName:\n    value: existing-cluster\n",
		},
		{
			name: "saved-default-templates", release: "nightly",
			values: `kubernetes-rbac-agent:
  url:
    value: '{{ include "stackstate.rbacAgent.url" . }}'
  clusterName:
    value: '{{ .Release.Name }}'
`,
		},
		{
			name: "external-configmaps-ignored", release: "nightly",
			values: "kubernetes-rbac-agent:\n  url:\n    fromConfigMap: external-url\n  clusterName:\n    fromConfigMap: external-cluster\n",
		},
		{
			name: "empty-inline-values", release: "nightly",
			values: "kubernetes-rbac-agent:\n  url:\n    value: ''\n  clusterName:\n    value: ''\n",
		},
		{
			name: "null-inline-values-with-external-configmaps", release: "nightly",
			values: "kubernetes-rbac-agent:\n  url:\n    value: null\n    fromConfigMap: external-url\n  clusterName:\n    value: null\n    fromConfigMap: external-cluster\n",
		},
		{
			name: "overridden-templates-never-evaluated", release: "nightly",
			values: `kubernetes-rbac-agent:
  url:
    value: '{{ fail "must not evaluate inline URL" }}'
    fromConfigMap: '{{ fail "must not evaluate external URL" }}'
  clusterName:
    value: '{{ fail "must not evaluate inline cluster" }}'
    fromConfigMap: '{{ fail "must not evaluate external cluster" }}'
`,
		},
	} {
		if tc.url == "" {
			tc.url = "http://nightly-suse-observability-router:8080/receiver/stsAgent"
		}
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				valuesPath := filepath.Join(t.TempDir(), "values.yaml")
				require.NoError(t, os.WriteFile(valuesPath, []byte(tc.values), 0600))
				var args []string
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, tc.release, &helm.Options{
					ValuesFiles:    []string{"values/full.yaml", valuesPath},
					KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
				}, args...)
				resources := helmtestutil.NewKubernetesResources(t, output)
				agent := tc.release + "-rbac-agent"
				for name, expected := range map[string]struct{ key, value string }{
					agent + "-url": {"STS_URL", tc.url}, agent + "-cluster-name": {"STS_CLUSTER_NAME", tc.release},
				} {
					require.Contains(t, resources.ConfigMaps, name)
					assert.Equal(t, map[string]string{expected.key: expected.value}, resources.ConfigMaps[name].Data)
				}
				require.Contains(t, resources.Deployments, agent)
				var configMapRefs []string
				for _, ref := range resources.Deployments[agent].Spec.Template.Spec.Containers[0].EnvFrom {
					if ref.ConfigMapRef != nil {
						configMapRefs = append(configMapRefs, ref.ConfigMapRef.Name)
					}
				}
				assert.ElementsMatch(t, []string{agent + "-url", agent + "-cluster-name"}, configMapRefs)
			})
		}
	}
}

func TestRbacAgentConnectionGlobalSecretsTakePrecedence(t *testing.T) {
	for _, tc := range []struct {
		name         string
		url, cluster bool
	}{
		{"url", true, false},
		{"cluster", false, true},
		{"both", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{}
			if tc.url {
				values["global.url.fromSecret"] = "\\{\\{ .Release.Name }}-connection"
			}
			if tc.cluster {
				values["global.clusterName.fromSecret"] = "\\{\\{ .Release.Name }}-connection"
			}
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", &helm.Options{
				ValuesFiles: []string{"values/full.yaml"}, SetValues: values,
				KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			require.Contains(t, resources.Deployments, "nightly-rbac-agent")
			container := resources.Deployments["nightly-rbac-agent"].Spec.Template.Spec.Containers[0]
			for _, field := range []struct {
				key, suffix string
				secret      bool
			}{
				{"STS_URL", "url", tc.url},
				{"STS_CLUSTER_NAME", "cluster-name", tc.cluster},
			} {
				configMap := "nightly-rbac-agent-" + field.suffix
				ref := corev1.EnvFromSource{ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: configMap},
				}}
				if field.secret {
					assert.NotContains(t, resources.ConfigMaps, configMap)
					assert.NotContains(t, container.EnvFrom, ref)
					assert.Contains(t, container.Env, corev1.EnvVar{
						Name: field.key,
						ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "nightly-connection"},
							Key:                  field.key,
						}},
					})
				} else {
					assert.Contains(t, resources.ConfigMaps, configMap)
					assert.Contains(t, container.EnvFrom, ref)
				}
			}
		})
	}
}
