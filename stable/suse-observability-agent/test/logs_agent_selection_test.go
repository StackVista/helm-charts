package test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func assertUniqueLogsManifests(t *testing.T, output string) map[string]int {
	t.Helper()
	decoder := yaml.NewDecoder(strings.NewReader(output))
	identities := map[string]bool{}
	logsKinds := map[string]int{}
	for {
		var document struct {
			APIVersion string `yaml:"apiVersion"`
			Kind       string `yaml:"kind"`
			Metadata   struct {
				Name      string            `yaml:"name"`
				Namespace string            `yaml:"namespace"`
				Labels    map[string]string `yaml:"labels"`
			} `yaml:"metadata"`
		}
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if document.Kind == "" {
			continue
		}
		group := ""
		if parts := strings.SplitN(document.APIVersion, "/", 2); len(parts) == 2 {
			group = parts[0]
		}
		namespace := document.Metadata.Namespace
		switch document.Kind {
		case "ClusterRole", "ClusterRoleBinding", "CustomResourceDefinition", "Namespace",
			"MutatingWebhookConfiguration", "ValidatingWebhookConfiguration":
			namespace = ""
		default:
			if namespace == "" {
				namespace = "default"
			}
		}
		identity := strings.Join([]string{group, document.Kind, namespace, document.Metadata.Name}, "/")
		require.False(t, identities[identity], "duplicate rendered resource %s", identity)
		identities[identity] = true
		if strings.Contains(document.Metadata.Name, "logs-agent") ||
			strings.Contains(document.Metadata.Labels["app.kubernetes.io/component"], "logs-agent") {
			assert.Equal(t, logsAgentName, document.Metadata.Name, "stable logs resource name")
			logsKinds[document.Kind]++
		}
	}
	return logsKinds
}

func TestLogsAgentOtelSelection(t *testing.T) {
	for _, selector := range []bool{false, true} {
		for _, promtailEnabled := range []bool{false, true} {
			for _, otelLogsEnabled := range []bool{false, true} {
				for _, globalOtel := range []bool{false, true} {
					t.Run(fmt.Sprintf("selector=%t/promtail=%t/otelLogs=%t/otel=%t",
						selector, promtailEnabled, otelLogsEnabled, globalOtel), func(t *testing.T) {
						enabled := promtailEnabled
						if selector {
							enabled = otelLogsEnabled
						}
						assertLogsSelection(t, map[string]string{
							"global.features.experimentalOtelLogsAgent": fmt.Sprint(selector),
							"logsAgent.enabled":                         fmt.Sprint(promtailEnabled),
							"otelLogsAgent.enabled":                     fmt.Sprint(otelLogsEnabled),
							"otel.enabled":                              fmt.Sprint(globalOtel),
						}, selector, enabled)
					})
				}
			}
		}
	}
}

func TestLogsAgentSelectionDefaults(t *testing.T) {
	for _, globalOtel := range []bool{false, true} {
		for _, tc := range []struct {
			name     string
			values   map[string]string
			selector bool
			enabled  bool
		}{
			{name: "all-defaults", enabled: true},
			{name: "omitted-selector-no-promtail-fallback", values: map[string]string{
				"logsAgent.enabled": "false", "otelLogsAgent.enabled": "true",
			}},
			{name: "omitted-selector-otel-disabled", values: map[string]string{
				"otelLogsAgent.enabled": "false",
			}, enabled: true},
			{name: "explicit-promtail-default-enables", values: map[string]string{
				"global.features.experimentalOtelLogsAgent": "false",
			}, enabled: true},
			{name: "explicit-otel-default-enables", values: map[string]string{
				"global.features.experimentalOtelLogsAgent": "true",
			}, selector: true, enabled: true},
			{name: "default-otel-enable-promtail-disabled", values: map[string]string{
				"global.features.experimentalOtelLogsAgent": "true", "logsAgent.enabled": "false",
			}, selector: true, enabled: true},
			{name: "otel-disabled-default-promtail-no-fallback", values: map[string]string{
				"global.features.experimentalOtelLogsAgent": "true", "otelLogsAgent.enabled": "false",
			}, selector: true},
		} {
			t.Run(fmt.Sprintf("%s/otel=%t", tc.name, globalOtel), func(t *testing.T) {
				values := map[string]string{"otel.enabled": fmt.Sprint(globalOtel)}
				for key, value := range tc.values {
					values[key] = value
				}
				assertLogsSelection(t, values, tc.selector, tc.enabled)
			})
		}
	}
}

func assertLogsSelection(t *testing.T, values map[string]string, otel, enabled bool) {
	t.Helper()
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/logs-otel-base.yaml"},
		SetValues:   values,
	})
	kinds := assertUniqueLogsManifests(t, output)
	expected := map[string]int{}
	if enabled {
		expected = map[string]int{
			"DaemonSet": 1, "ConfigMap": 1, "ClusterRole": 1, "ClusterRoleBinding": 1, "ServiceAccount": 1,
		}
	}
	require.Equal(t, expected, kinds, "complete, exclusive logs workload and RBAC set")
	resources := helmtestutil.NewKubernetesResources(t, output)
	if !enabled {
		return
	}
	assertLogsStableIdentity(t, resources, otel)
	container := logsContainer(t, resources)
	if otel {
		logsOtelConfig(t, resources)
		assert.Contains(t, container.Image, "/stackstate/sts-opentelemetry-collector:")
		assert.Equal(t, []string{"--config=/etc/otel/otel-logs.yaml", "--feature-gates=stanza.synchronousLogEmitter"}, container.Args)
		assert.Empty(t, container.Command)
	} else {
		cm := resources.ConfigMaps[logsAgentName]
		assert.Contains(t, cm.Data, "promtail.yaml")
		assert.NotContains(t, cm.Data, "otel-logs.yaml")
		assert.Contains(t, container.Image, "/stackstate/promtail:")
		assert.Equal(t, []string{"-config.expand-env=true", "-config.file=/etc/promtail/promtail.yaml"}, container.Args)
		assert.NotContains(t, cm.Data["promtail.yaml"], "stslogsroute")
		assert.Contains(t, cm.Data["promtail.yaml"], "https://my-suse-observability-instance.com/receiver/stsAgent/logs/k8s")
	}
}

func assertLogsStableIdentity(t *testing.T, resources helmtestutil.KubernetesResources, otel bool) {
	t.Helper()
	ds := resources.DaemonSets[logsAgentName]
	require.NotNil(t, ds.Spec.Selector)
	assert.Equal(t, map[string]string{
		"app.kubernetes.io/component": "logs-agent",
		"app.kubernetes.io/instance":  "suse-observability-agent",
		"app.kubernetes.io/name":      "suse-observability-agent",
	}, ds.Spec.Selector.MatchLabels)
	assert.Empty(t, ds.Spec.Selector.MatchExpressions)
	for key, value := range ds.Spec.Selector.MatchLabels {
		assert.Equal(t, value, ds.Spec.Template.Labels[key])
	}
	assert.Equal(t, logsAgentName, ds.Spec.Template.Spec.ServiceAccountName)
	assert.Equal(t, logsAgentName, resources.ServiceAccounts[logsAgentName].Name)
	assert.Equal(t, ds.Namespace, resources.ServiceAccounts[logsAgentName].Namespace)
	volume := requireVolume(t, ds.Spec.Template.Spec.Volumes, "logs-agent-config")
	require.NotNil(t, volume.ConfigMap)
	assert.Equal(t, logsAgentName, volume.ConfigMap.Name)
	binding := resources.ClusterRoleBindings[logsAgentName]
	assert.Equal(t, rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: logsAgentName,
	}, binding.RoleRef)
	assert.Equal(t, []rbacv1.Subject{{
		Kind: "ServiceAccount", Name: logsAgentName, Namespace: ds.Namespace,
	}}, binding.Subjects)
	rules := []rbacv1.PolicyRule{{
		APIGroups: []string{""}, Resources: []string{"nodes", "services", "pods"}, Verbs: []string{"get", "watch", "list"},
	}}
	if otel {
		rules = []rbacv1.PolicyRule{
			{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get", "list", "watch"}},
		}
	}
	assert.ElementsMatch(t, rules, resources.ClusterRoles[logsAgentName].Rules)
}

func logsWorkloadOverrides(prefix, owner string) map[string]string {
	return map[string]string{
		prefix + ".resources.limits.cpu":      "750m",
		prefix + ".resources.limits.memory":   "384Mi",
		prefix + ".resources.requests.cpu":    "150m",
		prefix + ".resources.requests.memory": "128Mi",
		prefix + ".priorityClassName":         owner + "-priority",
		prefix + ".nodeSelector.pool":         owner,
		prefix + ".tolerations[0].key":        owner,
		prefix + ".tolerations[0].operator":   "Exists",
		prefix + ".tolerations[0].effect":     "NoSchedule",
		prefix + ".affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":       "pool",
		prefix + ".affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":  "In",
		prefix + ".affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]": owner,
		prefix + ".updateStrategy.type":              "OnDelete",
		prefix + ".updateStrategy.rollingUpdate":     "null",
		prefix + ".podLabels.owner":                  owner,
		prefix + ".podAnnotations.owner":             owner,
		prefix + ".serviceaccount.annotations.owner": owner,
		prefix + ".skipSslValidation":                "true",
	}
}

func TestLogsAgentSelectedWorkloadOverrides(t *testing.T) {
	for _, otel := range []bool{false, true} {
		t.Run(fmt.Sprint(otel), func(t *testing.T) {
			selected, inactive := "logsAgent", "otelLogsAgent"
			if otel {
				selected, inactive = inactive, selected
			}
			values := logsWorkloadOverrides(selected, "selected")
			values["global.features.experimentalOtelLogsAgent"] = fmt.Sprint(otel)
			baseline := renderLogsAgent(t, values)
			ds := baseline.DaemonSets[logsAgentName]
			pod := ds.Spec.Template.Spec
			container := logsContainer(t, baseline)
			assert.Equal(t, corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("750m"), corev1.ResourceMemory: resource.MustParse("384Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("150m"), corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			}, container.Resources)
			assert.Equal(t, "selected-priority", pod.PriorityClassName)
			assert.Equal(t, map[string]string{"pool": "selected"}, pod.NodeSelector)
			assert.Equal(t, []corev1.Toleration{{
				Key: "selected", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule,
			}}, pod.Tolerations)
			assert.Equal(t, &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key: "pool", Operator: corev1.NodeSelectorOpIn, Values: []string{"selected"},
						}},
					}},
				},
			}}, pod.Affinity)
			assert.Equal(t, appsv1.DaemonSetUpdateStrategy{Type: appsv1.OnDeleteDaemonSetStrategyType}, ds.Spec.UpdateStrategy)
			assert.Equal(t, "selected", ds.Spec.Template.Labels["owner"])
			assert.Equal(t, "selected", ds.Spec.Template.Annotations["owner"])
			assert.Equal(t, "selected", baseline.ServiceAccounts[logsAgentName].Annotations["owner"])
			require.NotEmpty(t, ds.Spec.Template.Annotations["checksum/override-configmap"])
			assertLogsStableIdentity(t, baseline, otel)
			if otel {
				config := logsOtelConfig(t, baseline)
				for _, path := range [][]string{
					{"exporters", "stsk8slogs/promtail", "tls"},
				} {
					assert.Equal(t, true, logsConfigMap(t, config, path...)["insecure_skip_verify"])
				}
			} else {
				var config map[string]interface{}
				require.NoError(t, yaml.Unmarshal([]byte(baseline.ConfigMaps[logsAgentName].Data["promtail.yaml"]), &config))
				clients, ok := config["clients"].([]interface{})
				require.True(t, ok)
				require.Len(t, clients, 1)
				assert.Equal(t, true, logsConfigMap(t, clients[0].(map[string]interface{}), "tls_config")["insecure_skip_verify"])
			}
			for key, value := range logsWorkloadOverrides(inactive, "inactive") {
				values[key] = value
			}
			values[inactive+".resources.limits.memory"] = "768Mi"
			values[inactive+".resources.limits.cpu"] = "900m"
			values[inactive+".resources.requests.memory"] = "256Mi"
			values[inactive+".resources.requests.cpu"] = "200m"
			values[inactive+".skipSslValidation"] = "false"
			values[inactive+".image.repository"] = "test/inactive"
			values[inactive+".image.tag"] = "unused"
			if otel {
				values["logsAgent.image.pullSecretName"] = "inactive-pull"
			}
			changed := renderLogsAgent(t, values)
			assert.Equal(t, baseline.DaemonSets[logsAgentName], changed.DaemonSets[logsAgentName])
			assert.Equal(t, baseline.ConfigMaps[logsAgentName], changed.ConfigMaps[logsAgentName])
			assert.Equal(t, baseline.ServiceAccounts[logsAgentName], changed.ServiceAccounts[logsAgentName])
			assert.Equal(t, baseline.ClusterRoles[logsAgentName], changed.ClusterRoles[logsAgentName])
			assert.Equal(t, baseline.ClusterRoleBindings[logsAgentName], changed.ClusterRoleBindings[logsAgentName])
		})
	}
}
