package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

func TestCollectorEndpointEnvironmentUpgradePatches(t *testing.T) {
	ref := func(name, config, key string) corev1.EnvVar {
		return corev1.EnvVar{Name: name, ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: config}, Key: key,
			},
		}}
	}
	defaults := []corev1.EnvVar{
		ref("API_URL", fullName, "api.url"),
		ref("INTAKE_URL", fullName, "intake.url"),
	}
	literals := []corev1.EnvVar{
		{Name: "API_URL", Value: "https://customer.example/api"},
		{Name: "INTAKE_URL", Value: "https://customer.example/intake"},
		{Name: "TOKEN", Value: "customer-token"},
	}
	ordered := []corev1.EnvVar{
		{Name: "TOKEN", Value: "customer-token"},
		{Name: "API_URL", Value: "$(TOKEN)/api"},
		{Name: "INTAKE_URL", Value: "$(API_URL)/intake"},
		{Name: "OTHER", Value: "$(TOKEN)/other"},
	}
	external := []corev1.EnvVar{
		ref("API_URL", "customer-endpoints", "api.url"),
		{Name: "INTAKE_URL", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "customer-endpoints"}, Key: "intake.url",
			},
		}},
	}
	for _, mode := range []string{"deployment", "daemonset", "statefulset"} {
		base := renderCollectorEnvironmentWorkload(t, mode, []corev1.EnvVar{})
		var builtins []corev1.EnvVar
		for _, env := range collectorEnvironmentTemplate(base).Spec.Containers[0].Env {
			if env.Name != "API_URL" && env.Name != "INTAKE_URL" {
				builtins = append(builtins, env)
			}
		}
		for _, tc := range []struct {
			name     string
			previous []corev1.EnvVar
			updates  [][]corev1.EnvVar
		}{
			{"ignored-endpoint-overrides", defaults, [][]corev1.EnvVar{literals}},
			{"ordered-additions", defaults, [][]corev1.EnvVar{ordered, defaults, {}}},
			{"existing-custom-endpoints", literals, [][]corev1.EnvVar{literals, external, literals, {}}},
			{"saved-defaults-migration", defaults, [][]corev1.EnvVar{defaults, literals, {}}},
		} {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				// The base chart supplied endpoints only through extraEnvs.
				// Reconstruct its rendered env list, keeping the other built-ins.
				original := base.DeepCopyObject()
				collectorEnvironmentTemplate(original).Spec.Containers[0].Env = append(
					append([]corev1.EnvVar{}, builtins...), tc.previous...)
				current := original.DeepCopyObject()
				collectorEnvironmentTemplate(current).Annotations["example.test/live"] = "preserved"
				for step, extra := range tc.updates {
					t.Run(fmt.Sprintf("step-%d", step), func(t *testing.T) {
						target := renderCollectorEnvironmentWorkload(t, mode, extra)
						oldJSON, err := json.Marshal(original)
						require.NoError(t, err)
						currentJSON, err := json.Marshal(current)
						require.NoError(t, err)
						newJSON, err := json.Marshal(target)
						require.NoError(t, err)
						schema, err := strategicpatch.NewPatchMetaFromStruct(target)
						require.NoError(t, err)
						// Helm 3 uses this three-way strategic merge for apps/v1 workloads.
						patch, err := strategicpatch.CreateThreeWayMergePatch(oldJSON, newJSON, currentJSON, schema, true)
						require.NoError(t, err)
						patchedJSON, err := strategicpatch.StrategicMergePatch(currentJSON, patch, target)
						require.NoError(t, err)
						patched := target.DeepCopyObject()
						require.NoError(t, json.Unmarshal(patchedJSON, patched))
						pod := collectorEnvironmentTemplate(patched)
						envs := pod.Spec.Containers[0].Env
						assert.Equal(t, collectorEnvironmentTemplate(target).Spec.Containers[0].Env, envs)
						seen := map[string]bool{}
						for _, env := range envs {
							assert.False(t, seen[env.Name], "duplicate environment variable: %s", env.Name)
							seen[env.Name] = true
							assert.False(t, env.Value != "" && env.ValueFrom != nil,
								"%s must not have both value and valueFrom", env.Name)
						}
						assert.True(t, seen["API_URL"])
						assert.True(t, seen["INTAKE_URL"])
						assert.Contains(t, envs, defaults[0])
						assert.Contains(t, envs, defaults[1])
						assert.Equal(t, "preserved", pod.Annotations["example.test/live"])
						original, current = target, patched
					})
				}
			})
		}
	}
}

func renderCollectorEnvironmentWorkload(t *testing.T, mode string, extra []corev1.EnvVar) runtime.Object {
	t.Helper()
	values, err := json.Marshal(map[string]interface{}{"extraEnvs": extra})
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "values.json")
	require.NoError(t, os.WriteFile(path, values, 0600))
	output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "orders", &helm.Options{
		ValuesFiles: []string{path}, SetValues: map[string]string{"mode": mode},
	}, "--is-upgrade")
	resources := helmtestutil.NewKubernetesResources(t, output)
	switch mode {
	case "deployment":
		require.Contains(t, resources.Deployments, fullName)
		workload := resources.Deployments[fullName]
		return &workload
	case "daemonset":
		require.Contains(t, resources.DaemonSets, fullName+"-agent")
		workload := resources.DaemonSets[fullName+"-agent"]
		return &workload
	default:
		require.Contains(t, resources.Statefulsets, fullName)
		workload := resources.Statefulsets[fullName]
		return &workload
	}
}

func collectorEnvironmentTemplate(workload runtime.Object) *corev1.PodTemplateSpec {
	switch resource := workload.(type) {
	case *appsv1.Deployment:
		return &resource.Spec.Template
	case *appsv1.DaemonSet:
		return &resource.Spec.Template
	case *appsv1.StatefulSet:
		return &resource.Spec.Template
	default:
		panic(fmt.Sprintf("unexpected collector workload: %T", workload))
	}
}
