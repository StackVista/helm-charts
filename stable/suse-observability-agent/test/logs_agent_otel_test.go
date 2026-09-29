package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const logsAgentName = "suse-observability-agent-logs-agent"

func TestLogsAgentOtelBinaryValidation(t *testing.T) {
	binary := os.Getenv("OTEL_AGENT_BINARY")
	if binary == "" {
		t.Skip("set OTEL_AGENT_BINARY to validate rendered configs with the collector")
	}
	binary, err := exec.LookPath(binary)
	require.NoError(t, err)
	binary, err = filepath.Abs(binary)
	require.NoError(t, err)

	for _, scenario := range []struct {
		name   string
		values map[string]string
	}{
		{name: "literal values", values: map[string]string{
			"stackstate.url": "https://127.0.0.1:18443/receiver/stsAgent",
		}},
		{name: "Secret-only configuration", values: map[string]string{
			"global.url.fromSecret":         "receiver-url",
			"global.clusterName.fromSecret": "cluster-name",
			"stackstate.url":                "null",
			"stackstate.cluster.name":       "null",
		}},
		{name: "Secrets override literals", values: map[string]string{
			"global.url.fromSecret":         "receiver-url",
			"global.clusterName.fromSecret": "cluster-name",
			"stackstate.url":                "unused-url",
			"stackstate.cluster.name":       "unused-cluster",
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			validateLogsOtelConfig(t, binary, renderLogsOtel(t, scenario.values))
		})
	}
}

func validateLogsOtelConfig(t *testing.T, binary string, resources helmtestutil.KubernetesResources) {
	t.Helper()
	config := resources.ConfigMaps[logsAgentName].Data["otel-logs.yaml"]
	dir := t.TempDir()
	configPath := filepath.Join(dir, "otel-logs.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(config), 0600))

	container := logsContainer(t, resources)
	args := append([]string{"validate"}, container.Args...)
	require.Contains(t, args, "--config=/etc/otel/otel-logs.yaml")
	for i, arg := range args {
		if arg == "--config=/etc/otel/otel-logs.yaml" {
			args[i] = "--config=" + configPath
		}
	}
	env := []string{"PATH=/usr/bin:/bin", "TMPDIR=" + dir, "KUBERNETES_SERVICE_HOST=127.0.0.1"}
	for _, variable := range container.Env {
		value := variable.Value
		if variable.ValueFrom != nil {
			switch {
			case variable.ValueFrom.SecretKeyRef != nil:
				secretValues := map[string]string{
					"STS_API_KEY":      "synthetic-otel-logs-validation",
					"STS_URL":          "https://127.0.0.1:18443/receiver/stsAgent",
					"STS_CLUSTER_NAME": "secret-cluster-$name",
				}
				var found bool
				value, found = secretValues[variable.ValueFrom.SecretKeyRef.Key]
				require.True(t, found, "unsupported Secret key %s", variable.ValueFrom.SecretKeyRef.Key)
			case variable.ValueFrom.FieldRef != nil:
				fieldValues := map[string]string{
					"spec.nodeName": "validation-node", "metadata.name": "validation-logs-agent",
					"metadata.namespace": "validation", "status.podIP": "127.0.0.1",
				}
				var found bool
				value, found = fieldValues[variable.ValueFrom.FieldRef.FieldPath]
				require.True(t, found, "unsupported Downward API field %s", variable.ValueFrom.FieldRef.FieldPath)
			default:
				t.Fatalf("unsupported environment source for %s", variable.Name)
			}
		}
		value = strings.ReplaceAll(value, "$(KUBERNETES_SERVICE_HOST)", "127.0.0.1")
		value = strings.ReplaceAll(value, "$(STS_URL)", "https://127.0.0.1:18443/receiver/stsAgent")
		env = append(env, variable.Name+"="+value)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = dir
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "collector validate: %s", output)
}

func renderLogsAgent(t *testing.T, values map[string]string, fixtures ...string) helmtestutil.KubernetesResources {
	t.Helper()
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: append([]string{"values/minimal.yaml", "values/logs-otel-base.yaml"}, fixtures...),
		SetValues:   values,
	})
	assertUniqueLogsManifests(t, output)
	return helmtestutil.NewKubernetesResources(t, output)
}

func renderLogsOtel(t *testing.T, values map[string]string) helmtestutil.KubernetesResources {
	t.Helper()
	return renderLogsAgent(t, values, "values/logs-otel-enabled.yaml")
}

func logsOtelConfig(t *testing.T, resources helmtestutil.KubernetesResources) map[string]interface{} {
	t.Helper()
	cm, ok := resources.ConfigMaps[logsAgentName]
	require.True(t, ok, "logs ConfigMap")
	require.Contains(t, cm.Data, "otel-logs.yaml")
	assert.NotContains(t, cm.Data, "promtail.yaml")
	var config map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(cm.Data["otel-logs.yaml"]), &config))
	return config
}

func logsConfigMap(t *testing.T, config map[string]interface{}, path ...string) map[string]interface{} {
	t.Helper()
	for _, key := range path {
		child, ok := config[key].(map[string]interface{})
		require.True(t, ok, "configuration map %s", strings.Join(path, "."))
		config = child
	}
	return config
}

func logsContainer(t *testing.T, resources helmtestutil.KubernetesResources) corev1.Container {
	t.Helper()
	ds, ok := resources.DaemonSets[logsAgentName]
	require.True(t, ok, "logs DaemonSet")
	require.Len(t, ds.Spec.Template.Spec.Containers, 1)
	container := ds.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "logs-agent", container.Name)
	return container
}

func TestLogsAgentOtelGraphAndBounds(t *testing.T) {
	resources := renderLogsOtel(t, nil)
	config := logsOtelConfig(t, resources)
	assert.Len(t, logsConfigMap(t, config, "receivers"), 1)
	assert.Len(t, logsConfigMap(t, config, "processors"), 4)
	assert.NotContains(t, config, "connectors")
	assert.Len(t, logsConfigMap(t, config, "exporters"), 1)
	assert.Len(t, logsConfigMap(t, config, "service", "pipelines"), 1)
	for id, expected := range map[string]map[string]interface{}{
		"logs/input": {
			"receivers":  []interface{}{"file_log/pods"},
			"processors": []interface{}{"memory_limiter", "transform/static_pod", "k8s_attributes", "transform/cluster"},
			"exporters":  []interface{}{"stsk8slogs/promtail"},
		},
	} {
		assert.Equal(t, expected, logsConfigMap(t, config, "service", "pipelines", id), id)
	}
	assert.Equal(t, map[string]interface{}{
		"controller_extension": "stslogsagent/logs",
		"max_concurrent_calls": 8,
		"max_record_bytes":     262144,
		"max_request_bytes":    1048576,
		"export_lifetime":      "90s",
	}, logsConfigMap(t, config, "exporters", "stsk8slogs/promtail", "delivery"))
	assert.Equal(t, map[string]interface{}{
		"directory": "/var/lib/otelcol/checkpoints", "create_directory": true, "recreate": false,
	}, logsConfigMap(t, config, "extensions", "file_storage/logs"))
	assert.Equal(t, map[string]interface{}{
		"health_endpoint": "0.0.0.0:13133",
	}, logsConfigMap(t, config, "extensions", "stslogsagent/logs"))
	assert.ElementsMatch(t, []string{"file_storage/logs", "stslogsagent/logs"},
		logsConfigMap(t, config, "service")["extensions"])
	assert.Len(t, logsConfigMap(t, config, "extensions"), 2)
	for _, exporter := range []string{"stsk8slogs/promtail"} {
		export := logsConfigMap(t, config, "exporters", exporter)
		assert.Equal(t, "5s", export["timeout"], exporter)
		assert.Equal(t, map[string]interface{}{
			"enabled": true, "initial_interval": "1s", "max_interval": "5s", "max_elapsed_time": "30s",
		}, logsConfigMap(t, export, "retry_on_failure"), exporter)
		assert.Equal(t, map[string]interface{}{
			"enabled": false,
		}, logsConfigMap(t, export, "sending_queue"), exporter)
	}
	assert.Equal(t, "${env:PROMTAIL_LOGS_URL}", logsConfigMap(t, config, "exporters", "stsk8slogs/promtail")["endpoint"])
	assert.Equal(t, "${env:STS_API_KEY}", logsConfigMap(t, config, "exporters", "stsk8slogs/promtail")["api_key"])
	assert.Equal(t, "${env:CLUSTER_NAME}", logsConfigMap(t, config, "exporters", "stsk8slogs/promtail")["cluster_name"])
	assertLogsOtelBounds(t, resources, config)
}

func assertLogsOtelBounds(t *testing.T, resources helmtestutil.KubernetesResources, config map[string]interface{}) {
	t.Helper()
	delivery := logsConfigMap(t, config, "exporters", "stsk8slogs/promtail", "delivery")
	calls := delivery["max_concurrent_calls"].(int)
	files := logsConfigMap(t, config, "receivers", "file_log/pods")["max_concurrent_files"].(int)
	lifetime, err := time.ParseDuration(delivery["export_lifetime"].(string))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, calls, files+2)
	for name, raw := range logsConfigMap(t, config, "exporters") {
		export := raw.(map[string]interface{})
		assert.Equal(t, map[string]interface{}{"enabled": false}, logsConfigMap(t, export, "sending_queue"), name)
		retry, err := time.ParseDuration(logsConfigMap(t, export, "retry_on_failure")["max_elapsed_time"].(string))
		require.NoError(t, err)
		timeout, err := time.ParseDuration(export["timeout"].(string))
		require.NoError(t, err)
		assert.GreaterOrEqual(t, lifetime, retry+timeout+20*time.Second, name)
	}
	grace := resources.DaemonSets[logsAgentName].Spec.Template.Spec.TerminationGracePeriodSeconds
	require.NotNil(t, grace)
	assert.Equal(t, int64((lifetime+time.Second-1)/time.Second)+30, *grace)
}

func TestLogsAgentOtelFilelogAndEnrichment(t *testing.T) {
	config := logsOtelConfig(t, renderLogsOtel(t, nil))
	receiver := logsConfigMap(t, config, "receivers", "file_log/pods")
	assert.Equal(t, []interface{}{"/var/log/pods/*/*/*.log"}, receiver["include"])
	assert.Contains(t, receiver["exclude"], "/var/log/pods/${env:POD_NAMESPACE}_${env:POD_NAME}_*/logs-agent/*.log")
	for key, expected := range map[string]interface{}{
		"include_file_path": true, "start_at": "beginning", "preserve_trailing_whitespaces": true,
		"storage": "file_storage/logs", "max_concurrent_files": 4, "max_log_size": "256KiB",
	} {
		assert.Equal(t, expected, receiver[key], key)
	}
	assert.Equal(t, map[string]interface{}{"enabled": false}, logsConfigMap(t, receiver, "retry_on_failure"))
	assert.Equal(t, []interface{}{map[string]interface{}{"type": "container", "id": "container"}}, receiver["operators"])
	k8s := logsConfigMap(t, config, "processors", "k8s_attributes")
	assert.Equal(t, "serviceAccount", k8s["auth_type"])
	assert.Equal(t, true, k8s["wait_for_metadata"])
	assert.Equal(t, map[string]interface{}{"node_from_env_var": "K8S_NODE_NAME"}, logsConfigMap(t, k8s, "filter"))
	assert.ElementsMatch(t, []string{
		"k8s.namespace.name", "k8s.pod.name", "k8s.pod.uid", "k8s.node.name", "k8s.container.name",
		"container.id", "container.image.name", "container.image.tag",
	}, logsConfigMap(t, k8s, "extract")["metadata"])
	source := func(name string) interface{} {
		return map[string]interface{}{"from": "resource_attribute", "name": name}
	}
	assert.Equal(t, []interface{}{
		map[string]interface{}{"sources": []interface{}{source("k8s.pod.uid")}},
		map[string]interface{}{"sources": []interface{}{source("k8s.pod.name"), source("k8s.namespace.name")}},
	}, k8s["pod_association"])
	static := logsConfigMap(t, config, "processors", "transform/static_pod")
	assert.Equal(t, "propagate", static["error_mode"])
	assert.Equal(t, []interface{}{map[string]interface{}{
		"context":    "resource",
		"statements": []interface{}{`delete_key(resource.attributes, "k8s.pod.uid") where resource.attributes["k8s.pod.uid"] != nil and not IsMatch(resource.attributes["k8s.pod.uid"], "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")`},
	}}, static["log_statements"])
	assert.Equal(t, []interface{}{map[string]interface{}{
		"context": "resource", "statements": []interface{}{`set(resource.attributes["k8s.cluster.name"], "${env:CLUSTER_NAME}")`},
	}}, logsConfigMap(t, config, "processors", "transform/cluster")["log_statements"])
}

func TestLogsAgentOtelPodAndRBAC(t *testing.T) {
	resources := renderLogsOtel(t, nil)
	container := logsContainer(t, resources)
	pod := resources.DaemonSets[logsAgentName].Spec.Template.Spec
	assert.Equal(t, logsAgentName, pod.ServiceAccountName)
	require.NotNil(t, pod.TerminationGracePeriodSeconds)
	assert.Equal(t, int64(120), *pod.TerminationGracePeriodSeconds)
	require.NotNil(t, container.SecurityContext)
	security := container.SecurityContext
	require.NotNil(t, security.RunAsUser)
	require.NotNil(t, security.RunAsGroup)
	require.NotNil(t, security.RunAsNonRoot)
	require.NotNil(t, security.Privileged)
	require.NotNil(t, security.SELinuxOptions)
	assert.Equal(t, int64(0), *security.RunAsUser)
	assert.Equal(t, int64(0), *security.RunAsGroup)
	assert.False(t, *security.RunAsNonRoot)
	assert.False(t, *security.Privileged)
	assert.Equal(t, "spc_t", security.SELinuxOptions.Type)
	if container.Lifecycle != nil {
		assert.Nil(t, container.Lifecycle.PreStop)
	}
	for _, mount := range []corev1.VolumeMount{
		{Name: "logs", MountPath: "/var/log", ReadOnly: true},
		{Name: "varlibdockercontainers", MountPath: "/var/lib/docker/containers", ReadOnly: true},
		{Name: "logs-agent-config", MountPath: "/etc/otel", ReadOnly: true},
		{Name: "logs-agent-state", MountPath: "/var/lib/otelcol"},
	} {
		assert.Contains(t, container.VolumeMounts, mount)
	}
	state := requireVolume(t, pod.Volumes, "logs-agent-state")
	require.NotNil(t, state.EmptyDir)
	assert.Nil(t, state.HostPath)
	assert.Nil(t, state.PersistentVolumeClaim)
	assert.Equal(t, corev1.StorageMediumDefault, state.EmptyDir.Medium)
	configVolume := requireVolume(t, pod.Volumes, "logs-agent-config")
	require.NotNil(t, configVolume.ConfigMap)
	assert.Equal(t, logsAgentName, configVolume.ConfigMap.Name)
	for name, path := range map[string]string{"logs": "/var/log", "varlibdockercontainers": "/var/lib/docker/containers"} {
		volume := requireVolume(t, pod.Volumes, name)
		require.NotNil(t, volume.HostPath)
		assert.Equal(t, path, volume.HostPath.Path)
	}
	for name, expected := range map[string]int32{"health": 13133, "metrics": 8888} {
		found := false
		for _, port := range container.Ports {
			if port.Name == name {
				found = true
				assert.Equal(t, expected, port.ContainerPort)
			}
		}
		assert.True(t, found, "container port %s", name)
	}
	for _, tc := range []struct {
		name      string
		probe     *corev1.Probe
		path      string
		threshold int32
	}{
		{"startup", container.StartupProbe, "/live", 12},
		{"liveness", container.LivenessProbe, "/live", 3},
		{"readiness", container.ReadinessProbe, "/ready", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.probe)
			require.NotNil(t, tc.probe.HTTPGet)
			assert.Equal(t, tc.path, tc.probe.HTTPGet.Path)
			assert.Equal(t, intstr.FromString("health"), tc.probe.HTTPGet.Port)
			assert.Equal(t, int32(5), tc.probe.PeriodSeconds)
			assert.Equal(t, int32(2), tc.probe.TimeoutSeconds)
			assert.Equal(t, tc.threshold, tc.probe.FailureThreshold)
		})
	}
	role, ok := resources.ClusterRoles[logsAgentName]
	require.True(t, ok)
	require.Len(t, role.Rules, 1)
	for _, expected := range []rbacv1.PolicyRule{
		{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get", "list", "watch"}},
	} {
		found := false
		for _, rule := range role.Rules {
			if assert.ObjectsAreEqual(expected.APIGroups, rule.APIGroups) {
				found = true
				assert.ElementsMatch(t, expected.Resources, rule.Resources)
				assert.ElementsMatch(t, expected.Verbs, rule.Verbs)
				assert.Empty(t, rule.NonResourceURLs)
				assert.Empty(t, rule.ResourceNames)
			}
		}
		assert.True(t, found, "API groups %v", expected.APIGroups)
	}
	binding := resources.ClusterRoleBindings[logsAgentName]
	assert.Equal(t, rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: logsAgentName}, binding.RoleRef)
	assert.Contains(t, binding.Subjects, rbacv1.Subject{
		Kind: "ServiceAccount", Name: logsAgentName, Namespace: resources.DaemonSets[logsAgentName].Namespace,
	})
	metrics := logsConfigMap(t, logsOtelConfig(t, resources), "service", "telemetry", "metrics")
	assert.Equal(t, []interface{}{map[string]interface{}{
		"pull": map[string]interface{}{"exporter": map[string]interface{}{"prometheus": map[string]interface{}{"host": "0.0.0.0", "port": 8888}}},
	}}, metrics["readers"])
	annotations := resources.DaemonSets[logsAgentName].Spec.Template.Annotations
	assert.JSONEq(t, `["openmetrics"]`, annotations["ad.stackstate.com/logs-agent.check_names"])
	var instances []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(annotations["ad.stackstate.com/logs-agent.instances"]), &instances))
	require.Len(t, instances, 1)
	assert.Equal(t, "http://%%host%%:8888/metrics", instances[0]["prometheus_url"])
	assert.Contains(t, instances[0]["metrics"], "*")
}

func TestLogsAgentOtelEndpointsAndSecret(t *testing.T) {
	for _, tc := range []struct{ name, http, grpc string }{
		{"default", "", ""},
		{"shared-http", "https://metrics.example.test/ingest", ""},
		{"shared-grpc", "", "metrics.example.test:443"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := renderLogsOtel(t, map[string]string{
				"stackstate.url":                   "https://my-suse-observability-instance.com/receiver/stsAgent/",
				"otel.platformHttpOtlpEndpoint":    tc.http,
				"otel.platformGrpcOtlpEndpoint":    tc.grpc,
				"stackstate.manageOwnSecrets":      "true",
				"stackstate.customSecretName":      "logs-api-key",
				"stackstate.customApiKeySecretKey": "receiver-key",
			})
			container := logsContainer(t, resources)
			env := envVarsByName(container.Env)
			assert.NotContains(t, env, "NATIVE_OTLP_URL")
			assert.NotContains(t, env, "RECEIVER_URL")
			assert.Equal(t, "https://my-suse-observability-instance.com/receiver/stsAgent/logs/k8s", env["PROMTAIL_LOGS_URL"])
			assert.Equal(t, "some-k8s-cluster", env["CLUSTER_NAME"])
			assert.Contains(t, container.Env, corev1.EnvVar{
				Name: "STS_API_KEY", ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "logs-api-key"}, Key: "receiver-key",
					},
				},
			})
			assertEnvFromFieldRef(t, container.Env, "K8S_NODE_NAME", "spec.nodeName")
			assertEnvFromFieldRef(t, container.Env, "POD_NAME", "metadata.name")
			assertEnvFromFieldRef(t, container.Env, "POD_NAMESPACE", "metadata.namespace")
			config := logsOtelConfig(t, resources)
			assertLogsOtelBounds(t, resources, config)
			exporters := logsConfigMap(t, config, "exporters")
			assert.Len(t, exporters, 1)
			assert.Contains(t, exporters, "stsk8slogs/promtail")
			assert.NotContains(t, resources.ConfigMaps[logsAgentName].Data["otel-logs.yaml"], "foobar")
			assert.NotContains(t, resources.Secrets, "logs-api-key")
		})
	}
	resources := renderLogsOtel(t, map[string]string{"global.apiKey.fromSecret": "existing-receiver-key"})
	assert.Contains(t, logsContainer(t, resources).Env, corev1.EnvVar{
		Name: "STS_API_KEY", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "existing-receiver-key"}, Key: "STS_API_KEY",
			},
		},
	})
}

func TestLogsAgentOtelExternalConfiguration(t *testing.T) {
	for _, scenario := range []struct {
		name          string
		clusterSecret bool
		urlSecret     bool
		keepLiterals  bool
	}{
		{name: "all external", clusterSecret: true, urlSecret: true},
		{name: "cluster external", clusterSecret: true},
		{name: "URL external", urlSecret: true},
		{name: "Secrets take precedence", clusterSecret: true, urlSecret: true, keepLiterals: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			values := map[string]string{}
			if scenario.clusterSecret {
				values["global.clusterName.fromSecret"] = "\\{\\{ .Release.Name }}-cluster"
				values["stackstate.cluster.name"] = "null"
				if scenario.keepLiterals {
					values["stackstate.cluster.name"] = "unused-cluster"
				}
			}
			if scenario.urlSecret {
				values["global.url.fromSecret"] = "\\{\\{ .Release.Name }}-url"
				values["stackstate.url"] = "null"
				if scenario.keepLiterals {
					values["stackstate.url"] = "unused-url"
				}
			}
			resources := renderLogsOtel(t, values)
			container := logsContainer(t, resources)
			assertConfigEnv(t, container.Env, "CLUSTER_NAME", "STS_CLUSTER_NAME", "suse-observability-agent-cluster", "some-k8s-cluster", scenario.clusterSecret)
			endpoint := configEnv(t, container.Env, "PROMTAIL_LOGS_URL")
			if scenario.urlSecret {
				assertConfigEnv(t, container.Env, "STS_URL", "STS_URL", "suse-observability-agent-url", "", true)
				assert.Equal(t, "$(STS_URL)/logs/k8s", endpoint.Value)
				assert.Less(t, envPosition(container.Env, "STS_URL"), envPosition(container.Env, endpoint.Name))
			} else {
				assert.Equal(t, "https://my-suse-observability-instance.com/receiver/stsAgent/logs/k8s", endpoint.Value)
			}
			config := logsOtelConfig(t, resources)
			assert.Equal(t, "${env:CLUSTER_NAME}", logsConfigMap(t, config, "exporters", "stsk8slogs/promtail")["cluster_name"])
			assert.Equal(t, []interface{}{map[string]interface{}{
				"context": "resource", "statements": []interface{}{`set(resource.attributes["k8s.cluster.name"], "${env:CLUSTER_NAME}")`},
			}}, logsConfigMap(t, config, "processors", "transform/cluster")["log_statements"])
			assert.NotContains(t, resources.ConfigMaps[logsAgentName].Data["otel-logs.yaml"], "unused")
			assert.NotContains(t, resources.Secrets, "suse-observability-agent-cluster")
			assert.NotContains(t, resources.Secrets, "suse-observability-agent-url")
		})
	}
}

func TestLogsAgentOtelSelectedImageAndPullSecrets(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			resources := renderLogsAgent(t, map[string]string{
				"global.features.experimentalOtelLogsAgent": fmt.Sprint(enabled),
				"global.imageRegistry":                      "registry.example.test",
				"logsAgent.image.repository":                "test/promtail",
				"logsAgent.image.tag":                       "promtail-test",
				"logsAgent.image.pullPolicy":                "Always",
				"logsAgent.image.pullSecretName":            "promtail-pull",
				"otelLogsAgent.image.repository":            "test/collector",
				"otelLogsAgent.image.tag":                   "collector-test",
				"otelLogsAgent.image.pullPolicy":            "Never",
				"otelLogsAgent.image.pullSecretName":        "otel-pull",
				"global.imagePullSecrets[0]":                "global-pull",
				"global.imagePullSecrets[1]":                "custom-{{ .Release.Name }}-pull",
				"global.imagePullSecrets[2]":                "global-pull",
				"all.image.pullSecretName":                  "common-pull",
			})
			container := logsContainer(t, resources)
			secrets := resources.DaemonSets[logsAgentName].Spec.Template.Spec.ImagePullSecrets
			expected := []corev1.LocalObjectReference{
				{Name: "global-pull"},
				{Name: "custom-suse-observability-agent-pull"},
				{Name: "common-pull"},
				{Name: "suse-observability-agent-pull-secret"},
			}
			if enabled {
				assert.Equal(t, "registry.example.test/test/collector:collector-test", container.Image)
				assert.Equal(t, corev1.PullNever, container.ImagePullPolicy)
				expected = append(expected, corev1.LocalObjectReference{Name: "otel-pull"})
			} else {
				assert.Equal(t, "registry.example.test/test/promtail:promtail-test", container.Image)
				assert.Equal(t, corev1.PullAlways, container.ImagePullPolicy)
				expected = append(expected, corev1.LocalObjectReference{Name: "promtail-pull"})
			}
			assert.ElementsMatch(t, expected, secrets)
		})
	}
}

func TestLogsAgentOtelTLSAndProxy(t *testing.T) {
	for _, proxy := range []string{"http://proxy.example.test:3128", "https://proxy.example.test:8443"} {
		for _, ca := range []string{"none", "inline", "external"} {
			t.Run(fmt.Sprintf("proxy=%s/ca=%s", proxy, ca), func(t *testing.T) {
				values := map[string]string{"global.proxy.url": proxy}
				if ca != "none" {
					values["global.customCertificates.enabled"] = "true"
					if ca == "inline" {
						values["global.customCertificates.pemData"] = "synthetic-ca-bundle"
					} else {
						values["global.customCertificates.configMapName"] = "external-ca"
					}
				}
				resources := renderLogsOtel(t, values)
				container := logsContainer(t, resources)
				env := envVarsByName(container.Env)
				assert.Equal(t, proxy, env["PROXY_URL"])
				config := logsOtelConfig(t, resources)
				for _, client := range []map[string]interface{}{
					logsConfigMap(t, config, "exporters", "stsk8slogs/promtail"),
				} {
					tls := logsConfigMap(t, client, "tls")
					assert.Equal(t, false, tls["insecure_skip_verify"])
					assert.NotContains(t, tls, "ca_file")
					assert.NotContains(t, tls, "include_system_ca_certs_pool")
				}
				for _, client := range []map[string]interface{}{
					logsConfigMap(t, config, "exporters", "stsk8slogs/promtail"),
				} {
					assert.Equal(t, "${env:PROXY_URL}", client["proxy_url"])
				}
				assert.NotContains(t, env, "HTTPS_PROXY")
				assert.NotContains(t, env, "NO_PROXY")
				if ca != "none" {
					volume := requireVolume(t, resources.DaemonSets[logsAgentName].Spec.Template.Spec.Volumes, "custom-certificates")
					require.NotNil(t, volume.ConfigMap)
					assert.Empty(t, volume.ConfigMap.Items, "mount every certificate filename from the ConfigMap")
					assert.Contains(t, container.VolumeMounts, corev1.VolumeMount{
						Name: "custom-certificates", MountPath: "/etc/pki/tls/certs", ReadOnly: true,
					})
					if ca == "inline" {
						assert.Equal(t, "suse-observability-agent-custom-certificates", volume.ConfigMap.Name)
						assert.Equal(t, "synthetic-ca-bundle", strings.TrimSpace(resources.ConfigMaps[volume.ConfigMap.Name].Data["tls.pem"]))
					} else {
						assert.Equal(t, "external-ca", volume.ConfigMap.Name)
						assert.NotContains(t, resources.ConfigMaps, "suse-observability-agent-custom-certificates")
						assert.NotContains(t, resources.DaemonSets[logsAgentName].Spec.Template.Annotations, "checksum/custom-certificates")
					}
				}
			})
		}
	}
}

func TestLogsAgentOtelSkipTLSValidation(t *testing.T) {
	for _, global := range []bool{false, true} {
		for _, local := range []bool{false, true} {
			t.Run(fmt.Sprintf("global=%t/local=%t", global, local), func(t *testing.T) {
				config := logsOtelConfig(t, renderLogsOtel(t, map[string]string{
					"global.skipSslValidation":        fmt.Sprint(global),
					"otelLogsAgent.skipSslValidation": fmt.Sprint(local),
				}))
				assert.Equal(t, global || local, logsConfigMap(t, config, "exporters", "stsk8slogs/promtail", "tls")["insecure_skip_verify"])
			})
		}
	}
}

func TestLogsAgentOtelChecksumTracksSelectedConfigAndInlineCA(t *testing.T) {
	for _, otel := range []bool{false, true} {
		t.Run(fmt.Sprint(otel), func(t *testing.T) {
			render := func(overrides map[string]string) map[string]string {
				values := map[string]string{
					"global.features.experimentalOtelLogsAgent": fmt.Sprint(otel),
					"global.customCertificates.enabled":         "true",
					"global.customCertificates.pemData":         "synthetic-ca-one",
				}
				for key, value := range overrides {
					values[key] = value
				}
				return renderLogsAgent(t, values).DaemonSets[logsAgentName].Spec.Template.Annotations
			}
			base := render(nil)
			require.NotEmpty(t, base["checksum/override-configmap"])
			require.NotEmpty(t, base["checksum/custom-certificates"])
			repeated := render(nil)
			for _, key := range []string{"checksum/override-configmap", "checksum/custom-certificates"} {
				assert.Equal(t, base[key], repeated[key], "identical inputs must produce stable %s", key)
			}
			selected := "logsAgent"
			if otel {
				selected = "otelLogsAgent"
			}
			tls := render(map[string]string{selected + ".skipSslValidation": "true"})
			assert.NotEqual(t, base["checksum/override-configmap"], tls["checksum/override-configmap"])
			changedOtel := render(map[string]string{"otelLogsAgent.resources.limits.memory": "384Mi"})
			if otel {
				assert.NotEqual(t, base["checksum/override-configmap"], changedOtel["checksum/override-configmap"])
			} else {
				assert.Equal(t, base["checksum/override-configmap"], changedOtel["checksum/override-configmap"])
			}
			ca := render(map[string]string{"global.customCertificates.pemData": "synthetic-ca-two"})
			assert.NotEqual(t, base["checksum/custom-certificates"], ca["checksum/custom-certificates"])
			switched := render(map[string]string{"global.features.experimentalOtelLogsAgent": fmt.Sprint(!otel)})
			assert.NotEqual(t, base["checksum/override-configmap"], switched["checksum/override-configmap"])
		})
	}
}

func TestLogsAgentOtelMemoryLimiterFollowsResources(t *testing.T) {
	for _, tc := range []struct {
		memory       string
		limit, spike int
	}{
		{"192Mi", 128, 32},
		{"384Mi", 256, 64},
		{"0.375Gi", 256, 64},
		{"402653184", 256, 64},
		{".5Gi", 341, 85},
		{"+512Mi", 341, 85},
	} {
		t.Run(tc.memory, func(t *testing.T) {
			resources := renderLogsOtel(t, map[string]string{
				"otelLogsAgent.resources.limits.memory":   tc.memory,
				"otelLogsAgent.resources.requests.memory": "120Mi",
				"otelLogsAgent.resources.limits.cpu":      "500m",
			})
			container := logsContainer(t, resources)
			expectedMemory := resource.MustParse(tc.memory)
			assert.Equal(t, expectedMemory.Value(), container.Resources.Limits.Memory().Value())
			assert.Equal(t, int64(120*1024*1024), container.Resources.Requests.Memory().Value())
			assert.Equal(t, int64(500), container.Resources.Limits.Cpu().MilliValue())
			limiter := logsConfigMap(t, logsOtelConfig(t, resources), "processors", "memory_limiter")
			assert.Equal(t, "1s", limiter["check_interval"])
			assert.Equal(t, tc.limit, limiter["limit_mib"])
			assert.Equal(t, tc.spike, limiter["spike_limit_mib"])
			assert.NotContains(t, limiter, "limit_percentage")
			assert.NotContains(t, limiter, "spike_limit_percentage")
		})
	}
}

func TestLogsAgentOtelRejectsInvalidMemoryLimit(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/logs-otel-base.yaml", "values/logs-otel-enabled.yaml"},
		SetValues:   map[string]string{"otelLogsAgent.resources.limits.memory": "0Mi"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "otelLogsAgent.resources.limits.memory")
}

func TestLogsAgentOtelIgnoresSharedOtlpSettings(t *testing.T) {
	baseline := renderLogsOtel(t, map[string]string{"otel.enabled": "false"})
	changed := renderLogsOtel(t, map[string]string{
		"otel.enabled":                  "false",
		"otel.platformHttpOtlpEndpoint": "https://metrics.example.test/ingest",
		"otel.platformGrpcOtlpEndpoint": "metrics.example.test:443",
	})
	assert.Equal(t, baseline.ConfigMaps[logsAgentName], changed.ConfigMaps[logsAgentName])
	assert.Equal(t, baseline.DaemonSets[logsAgentName], changed.DaemonSets[logsAgentName])
}
