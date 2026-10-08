package test

import (
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

func TestRouterModeReferencesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	for suffix, name := range map[string]string{
		"active.configmap.fullname":      "explicit-active-config",
		"maintenance.configmap.fullname": "explicit-maintenance-config",
		"automatic.configmap.fullname":   "explicit-automatic-config",
		"scripts.configmap.fullname":     "explicit-hook-scripts",
		"serviceaccount.fullname":        "explicit-hook-account",
		"role.fullname":                  "explicit-hook-role",
		"rolebinding.fullname":           "explicit-hook-binding",
		"active.job.fullname":            "explicit-active-job",
		"maintenance.job.fullname":       "explicit-maintenance-job",
		"active.job.generateName":        "explicit-active-generated-",
		"maintenance.job.generateName":   "explicit-maintenance-generated-",
	} {
		helper := "stackstate.router.mode." + suffix
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAllString(content, -1), 1, helper)
		content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}`+name+`{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, mode := range []string{"active", "maintenance", "automatic"} {
		for _, argo := range []bool{false, true} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/argo=%t/upgrade=%t", mode, argo, upgrade), func(t *testing.T) {
					options := apiResourceNameTestOptions(map[string]string{
						"stackstate.components.router.mode.status":                   mode,
						"deployment.compatibleWithArgoCD":                            fmt.Sprint(argo),
						"stackstate.components.router.mode.jobAnnotations.test-hook": "value",
						"stackstate.components.router.mode.podAnnotations.test-pod":  "value",
						"stackstate.components.router.mode.podLabels.test-label":     "value",
					})
					options.ValuesFiles = []string{valuesFile}
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					beforeOutput := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "nightly", options, args...)
					afterOutput, err := helm.RenderTemplateE(t, options, chart, "nightly", nil, args...)
					require.NoError(t, err)
					before := helmtestutil.NewKubernetesResources(t, beforeOutput)
					after := helmtestutil.NewKubernetesResources(t, afterOutput)
					legacy := "nightly-suse-observability-router"
					workload := "suse-observability-router"
					config := "explicit-" + mode + "-config"
					require.Contains(t, before.Deployments, workload)
					require.Contains(t, after.Deployments, workload)
					deployment := before.Deployments[workload]
					expectedDeployment := deployment.DeepCopy()
					mounts := 0
					for i := range expectedDeployment.Spec.Template.Spec.Volumes {
						volume := &expectedDeployment.Spec.Template.Spec.Volumes[i]
						if volume.Name == "router-mode" {
							require.NotNil(t, volume.ConfigMap)
							require.Equal(t, legacy+"-"+mode, volume.ConfigMap.Name)
							volume.ConfigMap.Name = config
							mounts++
						}
					}
					require.Equal(t, 1, mounts)
					assert.Equal(t, *expectedDeployment, after.Deployments[workload])
					delete(before.Deployments, workload)
					delete(after.Deployments, workload)
					for _, state := range []string{"active", "maintenance", "automatic"} {
						oldName, newName := legacy+"-"+state, "explicit-"+state+"-config"
						if mode == state && mode != "automatic" {
							require.Contains(t, before.ConfigMaps, oldName)
							require.Contains(t, after.ConfigMaps, newName)
							expected := before.ConfigMaps[oldName]
							expected.Name = newName
							assert.Equal(t, expected, after.ConfigMaps[newName], "Mode configuration must not change")
							delete(before.ConfigMaps, oldName)
							delete(after.ConfigMaps, newName)
						} else {
							assert.NotContains(t, before.ConfigMaps, oldName)
							assert.NotContains(t, after.ConfigMaps, newName)
						}
						assert.NotContains(t, after.ConfigMaps, oldName)
					}
					beforeJobs := routerModeJobsByAction(t, beforeOutput)
					afterJobs := routerModeJobsByAction(t, afterOutput)
					if mode == "automatic" {
						assertRouterModeHookResources(t, &before, &after)
						require.Len(t, beforeJobs, 2)
						require.Len(t, afterJobs, 2)
						for _, action := range []string{"active", "maintenance"} {
							require.Contains(t, beforeJobs, action)
							require.Contains(t, afterJobs, action)
							job := beforeJobs[action]
							expected := job.DeepCopy()
							if argo {
								assert.Empty(t, expected.Name)
								assert.Equal(t, "set-"+action+"-", expected.GenerateName)
								expected.GenerateName = "explicit-" + action + "-generated-"
							} else {
								assert.Regexp(t, "^nightly-suse-observability-set-"+action+"-[0-9]{2}t[0-9]{6}$", expected.Name)
								assert.Empty(t, expected.GenerateName)
								expected.Name = "explicit-" + action + "-job"
							}
							require.Equal(t, legacy+"-mode-scripts", expected.Spec.Template.Spec.ServiceAccountName)
							expected.Spec.Template.Spec.ServiceAccountName = "explicit-hook-account"
							require.Len(t, expected.Spec.Template.Spec.Volumes, 1)
							volume := &expected.Spec.Template.Spec.Volumes[0]
							require.NotNil(t, volume.ConfigMap)
							require.Equal(t, legacy+"-mode-scripts", volume.ConfigMap.Name)
							volume.ConfigMap.Name = "explicit-hook-scripts"
							assert.Equal(t, *expected, afterJobs[action], "Hook ordering, deadlines, permissions, pull secrets and pod-local volume names must stay unchanged")
						}
					} else {
						assert.Empty(t, beforeJobs)
						assert.Empty(t, afterJobs)
						assert.NotContains(t, before.ConfigMaps, legacy+"-mode-scripts")
						assert.NotContains(t, after.ConfigMaps, "explicit-hook-scripts")
						assert.NotContains(t, after.ServiceAccounts, "explicit-hook-account")
						assert.NotContains(t, after.Roles, "explicit-hook-role")
						assert.NotContains(t, after.RoleBindings, "explicit-hook-binding")
					}
					assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
					assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
					assert.Equal(t, before.Roles, after.Roles)
					assert.Equal(t, before.RoleBindings, after.RoleBindings)
					assert.Equal(t, before.Deployments, after.Deployments)
					assert.Equal(t, before.Services, after.Services)
					assert.Equal(t, before.Secrets, after.Secrets)
					assert.Equal(t, before.Statefulsets, after.Statefulsets)
					assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
				})
			}
		}
	}
}

func TestRouterModeRejectsUnsupportedValues(t *testing.T) {
	for _, mode := range []string{"custom", "Active", ""} {
		for _, argo := range []bool{false, true} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("mode=%q/argo=%t/upgrade=%t", mode, argo, upgrade), func(t *testing.T) {
					options := apiResourceNameTestOptions(map[string]string{
						"stackstate.components.router.mode.status": mode,
						"deployment.compatibleWithArgoCD":          fmt.Sprint(argo),
					})
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					_, err := helmtestutil.RenderHelmTemplateOptsWithArgs(t, "nightly", options, args...)
					require.Error(t, err)
					assert.Contains(t, err.Error(), "stackstate.components.router.mode.status must be one of: active, maintenance, automatic")
				})
			}
		}
	}
}

func assertRouterModeHookResources(t *testing.T, before, after *helmtestutil.KubernetesResources) {
	t.Helper()
	legacy := "nightly-suse-observability-router"
	require.Contains(t, before.ConfigMaps, legacy+"-mode-scripts")
	require.Contains(t, after.ConfigMaps, "explicit-hook-scripts")
	scripts := before.ConfigMaps[legacy+"-mode-scripts"]
	scripts.Name = "explicit-hook-scripts"
	scripts.Data = maps.Clone(scripts.Data)
	for _, key := range []string{"set-active.sh", "set-maintenance.sh"} {
		require.Contains(t, scripts.Data, key)
		oldTarget := "  name: " + legacy + "-automatic\n"
		require.Equal(t, 1, strings.Count(scripts.Data[key], oldTarget))
		scripts.Data[key] = strings.ReplaceAll(scripts.Data[key], oldTarget, "  name: explicit-automatic-config\n")
	}
	assert.Equal(t, scripts, after.ConfigMaps["explicit-hook-scripts"], "Both scripts must write the mounted ConfigMap without changing its contents or their control flow")
	delete(before.ConfigMaps, legacy+"-mode-scripts")
	delete(after.ConfigMaps, "explicit-hook-scripts")

	require.Contains(t, before.ServiceAccounts, legacy+"-mode-scripts")
	require.Contains(t, after.ServiceAccounts, "explicit-hook-account")
	account := before.ServiceAccounts[legacy+"-mode-scripts"]
	account.Name = "explicit-hook-account"
	assert.Equal(t, account, after.ServiceAccounts["explicit-hook-account"])
	delete(before.ServiceAccounts, legacy+"-mode-scripts")
	delete(after.ServiceAccounts, "explicit-hook-account")

	require.Contains(t, before.Roles, legacy+"-mode")
	require.Contains(t, after.Roles, "explicit-hook-role")
	role := before.Roles[legacy+"-mode"]
	role.Name = "explicit-hook-role"
	assert.Equal(t, role, after.Roles["explicit-hook-role"])
	delete(before.Roles, legacy+"-mode")
	delete(after.Roles, "explicit-hook-role")

	require.Contains(t, before.RoleBindings, legacy+"-mode")
	require.Contains(t, after.RoleBindings, "explicit-hook-binding")
	binding := before.RoleBindings[legacy+"-mode"]
	binding.Name = "explicit-hook-binding"
	require.Equal(t, legacy+"-mode", binding.RoleRef.Name)
	binding.RoleRef.Name = "explicit-hook-role"
	binding.Subjects = renamedServiceAccountSubject(t, binding.Subjects, legacy+"-mode-scripts", "explicit-hook-account")
	assert.Equal(t, binding, after.RoleBindings["explicit-hook-binding"])
	delete(before.RoleBindings, legacy+"-mode")
	delete(after.RoleBindings, "explicit-hook-binding")
}

// The generic resource map uses metadata.name, so it cannot distinguish the
// two Argo CD Jobs that only have generateName. Read both by their fixed action.
func routerModeJobsByAction(t *testing.T, output string) map[string]batchv1.Job {
	t.Helper()
	jobs := map[string]batchv1.Job{}
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(output), 4096)
	for {
		var object unstructured.Unstructured
		err := decoder.Decode(&object)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if object.GetKind() != "Job" || object.GetLabels()["app.kubernetes.io/component"] != "router-mode" {
			continue
		}
		var job batchv1.Job
		require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, &job))
		require.Len(t, job.Spec.Template.Spec.Containers, 1)
		action := strings.TrimPrefix(job.Spec.Template.Spec.Containers[0].Name, "router-mode-")
		require.NotContains(t, jobs, action)
		jobs[action] = job
	}
	return jobs
}
