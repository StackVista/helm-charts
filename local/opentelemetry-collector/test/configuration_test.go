package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestCollectorEndpointEnvironmentConfiguration(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	// A deliberately distinct name catches consumers still using the literal.
	require.NoError(t, os.WriteFile(filepath.Join(chart, "templates", "_names.tpl"), []byte(
		`{{- define "stackstate.otelCollector.endpoints.configmap.fullname" -}}explicit-endpoints{{- end -}}`), 0600))
	ref := func(name, config, key string) corev1.EnvVar {
		return corev1.EnvVar{Name: name, ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: config}, Key: key,
			},
		}}
	}
	for _, mode := range []string{"deployment", "daemonset", "statefulset"} {
		for _, tc := range []struct {
			name, values string
			additional   []corev1.EnvVar
		}{
			{"defaults", "", nil},
			{"empty", "extraEnvs: []\n", nil},
			{"ordered-additions", `extraEnvs:
- name: TOKEN
  value: customer-token
- name: API_URL
  value: $(TOKEN)/api
- name: INTAKE_URL
  value: $(API_URL)/intake
- name: OTHER
  value: $(TOKEN)/other
`, []corev1.EnvVar{
				{Name: "TOKEN", Value: "customer-token"},
				{Name: "OTHER", Value: "$(TOKEN)/other"},
			}},
			{"custom", `extraEnvs:
- name: API_URL
  value: https://customer.example/api
- name: INTAKE_URL
  valueFrom:
    configMapKeyRef:
      name: customer-endpoints
      key: intake.url
- name: TOKEN
  valueFrom:
    secretKeyRef:
      name: customer-token
      key: token
- name: OTHER
  valueFrom:
    configMapKeyRef:
      name: suse-observability-otel-collector
      key: api.url
`, []corev1.EnvVar{
				{Name: "TOKEN", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "customer-token"}, Key: "token",
				}}},
				ref("OTHER", "suse-observability-otel-collector", "api.url"),
			}},
			{"saved-defaults", `extraEnvs:
- name: API_URL
  valueFrom:
    configMapKeyRef:
      name: suse-observability-otel-collector
      key: api.url
- name: INTAKE_URL
  valueFrom:
    configMapKeyRef:
      name: suse-observability-otel-collector
      key: intake.url
`, nil},
			{"duplicate-names", `extraEnvs:
- name: API_URL
  value: https://customer.example
- name: API_URL
  valueFrom:
    configMapKeyRef:
      name: suse-observability-otel-collector
      key: api.url
      optional: true
- name: INTAKE_URL
  valueFrom:
    configMapKeyRef:
      name: suse-observability-otel-collector
      key: customer-key
- name: ADDITIONAL_URL
  valueFrom:
    configMapKeyRef:
      name: customer-config
      key: optional-url
      optional: true
`, []corev1.EnvVar{
				func() corev1.EnvVar {
					env := ref("ADDITIONAL_URL", "customer-config", "optional-url")
					optional := true
					env.ValueFrom.ConfigMapKeyRef.Optional = &optional
					return env
				}(),
			}},
		} {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				values := filepath.Join(t.TempDir(), "values.yaml")
				require.NoError(t, os.WriteFile(values, []byte(tc.values), 0600))
				output, err := helm.RenderTemplateE(t, &helm.Options{
					ValuesFiles: []string{values}, SetValues: map[string]string{"mode": mode},
				}, chart, "orders", nil)
				require.NoError(t, err)
				resources := helmtestutil.NewKubernetesResources(t, output)
				var pod corev1.PodSpec
				switch mode {
				case "deployment":
					pod = resources.Deployments[fullName].Spec.Template.Spec
				case "daemonset":
					pod = resources.DaemonSets[fullName+"-agent"].Spec.Template.Spec
				case "statefulset":
					pod = resources.Statefulsets[fullName].Spec.Template.Spec
				}
				require.NotEmpty(t, pod.Containers)
				var extra []corev1.EnvVar
				for _, env := range pod.Containers[0].Env {
					if env.Name != "MY_POD_IP" && env.Name != "GOMEMLIMIT" && env.Name != "K8S_NODE_NAME" {
						extra = append(extra, env)
					}
				}
				expected := append([]corev1.EnvVar{
					ref("API_URL", "explicit-endpoints", "api.url"),
					ref("INTAKE_URL", "explicit-endpoints", "intake.url"),
				}, tc.additional...)
				assert.Equal(t, expected, extra)
				assert.NotContains(t, resources.ConfigMaps, "explicit-endpoints",
					"standalone collectors reference an external endpoint ConfigMap")
			})
		}
	}
}

func TestCollectorReservedEnvironmentVariables(t *testing.T) {
	for _, mode := range []string{"deployment", "daemonset", "statefulset"} {
		for _, tc := range []struct {
			name                 string
			nodeName, gomemlimit bool
		}{
			{"automatic-memory", false, true},
			{"node-name", true, true},
			{"manual-memory", false, false},
		} {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				render := func(custom bool) corev1.PodSpec {
					options := &helm.Options{SetValues: map[string]string{
						"mode":                           mode,
						"presets.kubeletMetrics.enabled": fmt.Sprint(tc.nodeName),
						"useGOMEMLIMIT":                  fmt.Sprint(tc.gomemlimit),
						"resources.limits.memory":        "512Mi",
					}}
					if custom {
						options.ValuesFiles = []string{"values/reserved-environment.yaml"}
					}
					output := helmtestutil.RenderHelmTemplateOptsNoError(t, releaseName, options)
					resources := helmtestutil.NewKubernetesResources(t, output)
					switch mode {
					case "deployment":
						require.Contains(t, resources.Deployments, fullName)
						return resources.Deployments[fullName].Spec.Template.Spec
					case "daemonset":
						require.Contains(t, resources.DaemonSets, fullName+"-agent")
						return resources.DaemonSets[fullName+"-agent"].Spec.Template.Spec
					default:
						require.Contains(t, resources.Statefulsets, fullName)
						return resources.Statefulsets[fullName].Spec.Template.Spec
					}
				}
				baseline := render(false)
				actual := render(true)
				require.NotEmpty(t, baseline.Containers)
				require.NotEmpty(t, actual.Containers)
				expected := append([]corev1.EnvVar{}, baseline.Containers[0].Env...)
				if !tc.gomemlimit {
					expected = append(expected, corev1.EnvVar{Name: "GOMEMLIMIT", Value: "123MiB"})
				}
				optional := true
				expected = append(expected,
					corev1.EnvVar{Name: "TOKEN", Value: "customer-token"},
					corev1.EnvVar{Name: "OTHER", ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "customer-settings"},
							Key:                  "other", Optional: &optional,
						},
					}},
					corev1.EnvVar{Name: "DEPENDENT", Value: "$(TOKEN)/other"},
				)
				assert.Equal(t, expected, actual.Containers[0].Env)
				assert.Equal(t, []corev1.EnvFromSource{
					{ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "customer-settings"},
					}},
					{SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "customer-secrets"},
					}},
				}, actual.Containers[0].EnvFrom)
			})
		}
	}
}
