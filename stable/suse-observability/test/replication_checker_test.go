package test

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestReplicationCheckerProfiles(t *testing.T) {
	for _, profile := range []string{"trial", "10-nonha", "20-nonha", "50-nonha", "100-nonha", "150-ha", "250-ha", "500-ha", "4000-ha"} {
		for _, source := range []string{"builtin", "legacy", "generated"} {
			t.Run(fmt.Sprintf("%s/%s", profile, source), func(t *testing.T) {
				values := []string{"values/full.yaml"}
				set := map[string]string{}
				if source == "generated" {
					values = append(values, "../../suse-observability-values/profiles/"+profile+".yaml")
				} else {
					values = append(values, "values/global_sizing_250_ha.yaml")
					set["global.suseObservability.sizing.profile"] = profile
					if source == "legacy" {
						set["global.suseObservability.sizing.profile"] = ""
						set["global.sizing.profile"] = profile
					}
				}
				output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{ValuesFiles: values, SetValues: set})
				resources := helmtestutil.NewKubernetesResources(t, output)
				expected := profile == "150-ha" || profile == "250-ha" || profile == "500-ha" || profile == "4000-ha"
				_, exists := resources.Deployments["suse-observability-replication-checker"]
				assert.Equal(t, expected, exists)
			})
		}
	}
}

func TestReplicationChecker(t *testing.T) {
	for _, tc := range []struct {
		name     string
		values   map[string]string
		observer bool
		checker  bool
	}{
		{name: "no profile", observer: true},
		{name: "explicit enable", observer: true, checker: true, values: map[string]string{
			"stackstate.components.replicationChecker.enabled": "true",
		}},
		{name: "disable HA default", observer: true, values: map[string]string{
			"global.suseObservability.sizing.profile":          "150-ha",
			"stackstate.components.replicationChecker.enabled": "false",
		}},
		{name: "enable non-HA", observer: true, checker: true, values: map[string]string{
			"global.suseObservability.sizing.profile":          "10-nonha",
			"stackstate.components.replicationChecker.enabled": "true",
		}},
		{name: "disable legacy HA default", observer: true, values: map[string]string{
			"global.sizing.profile":                            "250-ha",
			"stackstate.components.replicationChecker.enabled": "false",
		}},
		{name: "enable legacy non-HA", observer: true, checker: true, values: map[string]string{
			"global.sizing.profile":                            "10-nonha",
			"stackstate.components.replicationChecker.enabled": "true",
		}},
		{name: "canonical non-HA takes precedence", observer: true, values: map[string]string{
			"global.suseObservability.sizing.profile": "10-nonha",
			"global.sizing.profile":                   "250-ha",
		}},
		{name: "canonical HA takes precedence", observer: true, checker: true, values: map[string]string{
			"global.suseObservability.sizing.profile": "250-ha",
			"global.sizing.profile":                   "10-nonha",
		}},
		{name: "independent of observer", checker: true, values: map[string]string{
			"stackstate.components.workloadObserver.enabled":   "false",
			"stackstate.components.replicationChecker.enabled": "true",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := []string{"values/full.yaml"}
			if tc.values["global.suseObservability.sizing.profile"] != "" {
				values = append(values, "values/global_sizing_250_ha.yaml")
			}
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
				ValuesFiles: values, SetValues: tc.values,
				KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			observer, exists := resources.Statefulsets["suse-observability-workload-observer"]
			require.Equal(t, tc.observer, exists)
			if tc.observer {
				assert.Len(t, observer.Spec.Template.Spec.Containers, 1)
				assert.Equal(t, []rbacv1.PolicyRule{{APIGroups: []string{"apps"}, Resources: []string{"statefulsets", "deployments"}, Verbs: []string{"get", "list", "watch"}}}, resources.Roles[observer.Name].Rules)
			}
			name := "suse-observability-replication-checker"
			deployment, exists := resources.Deployments[name]
			require.Equal(t, tc.checker, exists)
			role, exists := resources.Roles[name]
			require.Equal(t, tc.checker, exists)
			binding, exists := resources.RoleBindings[name]
			require.Equal(t, tc.checker, exists)
			_, exists = resources.ServiceAccounts[name]
			require.Equal(t, tc.checker, exists)
			_, exists = resources.Pdbs[name]
			assert.False(t, exists)
			if !tc.checker {
				return
			}
			assert.Equal(t, []rbacv1.PolicyRule{
				{APIGroups: []string{"apps"}, Resources: []string{"statefulsets"}, Verbs: []string{"list"}},
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"list"}},
				{APIGroups: []string{""}, Resources: []string{"pods/exec"}, Verbs: []string{"create"}},
			}, role.Rules)
			require.Len(t, binding.Subjects, 1)
			assert.Equal(t, rbacv1.Subject{Kind: "ServiceAccount", Name: name, Namespace: "observability"}, binding.Subjects[0])
			assert.Equal(t, rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "Role", Name: name}, binding.RoleRef)
			pod := deployment.Spec.Template.Spec
			assert.Equal(t, name, pod.ServiceAccountName)
			assert.Empty(t, pod.Volumes)
			assert.Empty(t, pod.InitContainers)
			require.Len(t, pod.Containers, 1)
			checker := pod.Containers[0]
			assert.Equal(t, "replication-checker", checker.Name)
			assert.Regexp(t, `^my\.registry\.com/stackstate/container-tools:.+$`, checker.Image)
			assert.Equal(t, corev1.PullIfNotPresent, checker.ImagePullPolicy)
			assert.Equal(t, []string{"/bin/sh", "-c"}, checker.Command)
			assert.Equal(t, []string{"trap 'exit 0' TERM INT; sleep infinity & wait"}, checker.Args)
			assert.Empty(t, checker.VolumeMounts)
			assert.Empty(t, checker.EnvFrom)
			assert.Nil(t, checker.LivenessProbe)
			assert.Nil(t, checker.ReadinessProbe)
			require.NotNil(t, checker.SecurityContext)
			require.NotNil(t, checker.SecurityContext.ReadOnlyRootFilesystem)
			assert.True(t, *checker.SecurityContext.ReadOnlyRootFilesystem)
			assert.Equal(t, "64Mi", checker.Resources.Requests.Memory().String())
			assert.Equal(t, "256Mi", checker.Resources.Limits.Memory().String())
			assert.Equal(t, int32(1), *deployment.Spec.Replicas)
			helmtestutil.AssertRestrictedSecurityContext(t, helmtestutil.KubernetesResources{Deployments: map[string]appsv1.Deployment{name: deployment}}, nil)
		})
	}
}

func TestReplicationCheckerOverrides(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
		SetValues: map[string]string{
			"global.imagePullSecrets[0]":                                       "registry-secret",
			"global.imageRegistry":                                             "",
			"stackstate.components.replicationChecker.enabled":                 "true",
			"stackstate.components.replicationChecker.replicaCount":            "2",
			"stackstate.components.replicationChecker.resources.limits.memory": "512Mi",
			"stackstate.components.replicationChecker.resources.requests.cpu":  "25m",
			"stackstate.components.containerTools.image.registry":              "registry.example.com",
			"stackstate.components.containerTools.image.repository":            "custom/tools",
			"stackstate.components.containerTools.image.tag":                   "custom",
			"stackstate.components.containerTools.image.pullPolicy":            "Always",
			"stackstate.components.all.nodeSelector.pool":                      "observability",
			"stackstate.components.replicationChecker.nodeSelector.disk":       "ssd",
			"stackstate.components.all.tolerations[0].key":                     "all",
			"stackstate.components.all.tolerations[0].operator":                "Exists",
			"stackstate.components.replicationChecker.tolerations[0].key":      "checker",
			"stackstate.components.replicationChecker.tolerations[0].operator": "Exists",
			"stackstate.components.replicationChecker.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":      "zone",
			"stackstate.components.replicationChecker.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator": "Exists",
			"stackstate.components.all.podAnnotations.shared":                 "yes",
			"stackstate.components.replicationChecker.podAnnotations.checker": "yes",
			"global.commonLabels.team":                                        "operations",
			"stackstate.components.all.deploymentStrategy.type":               "RecreateSingletonsOnly",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	deployment := resources.Deployments["suse-observability-replication-checker"]
	require.Len(t, deployment.Spec.Template.Spec.Containers, 1)
	checker := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "registry.example.com/custom/tools:custom", checker.Image)
	assert.Equal(t, corev1.PullAlways, checker.ImagePullPolicy)
	assert.Equal(t, "512Mi", checker.Resources.Limits.Memory().String())
	assert.Equal(t, "25m", checker.Resources.Requests.Cpu().String())
	assert.Equal(t, int32(2), *deployment.Spec.Replicas)
	assert.Equal(t, appsv1.RollingUpdateDeploymentStrategyType, deployment.Spec.Strategy.Type)
	assert.Equal(t, map[string]string{"pool": "observability", "disk": "ssd"}, deployment.Spec.Template.Spec.NodeSelector)
	assert.ElementsMatch(t, []corev1.Toleration{{Key: "all", Operator: corev1.TolerationOpExists}, {Key: "checker", Operator: corev1.TolerationOpExists}}, deployment.Spec.Template.Spec.Tolerations)
	require.NotNil(t, deployment.Spec.Template.Spec.Affinity.NodeAffinity)
	assert.Equal(t, "zone", deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key)
	assert.Equal(t, "yes", deployment.Spec.Template.Annotations["shared"])
	assert.Equal(t, "yes", deployment.Spec.Template.Annotations["checker"])
	assert.Equal(t, "operations", deployment.Labels["team"])
	assert.Equal(t, "operations", deployment.Spec.Template.Labels["team"])
	assert.NotEmpty(t, deployment.Spec.Template.Spec.ImagePullSecrets)
}
