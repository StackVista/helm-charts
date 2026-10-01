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

const (
	logsAgentName     = "suse-observability-agent-logs-agent"
	otelLogsAgentName = "suse-observability-agent-otel-logs-agent"
)

func TestLogsAgentOtelBinaryValidation(t *testing.T) {
	binary := os.Getenv("OTEL_AGENT_BINARY")
	if binary == "" {
		t.Skip("set OTEL_AGENT_BINARY to validate rendered configs with the collector")
	}
	binary, err := exec.LookPath(binary)
	require.NoError(t, err)
	binary, err = filepath.Abs(binary)
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		transport string
		ca        bool
		proxy     bool
		secrets   bool
		literals  bool
		mode      string
	}{
		{name: "default"},
		{name: "secret-default", secrets: true},
		{name: "secret-http", transport: "http", secrets: true},
		{name: "secret-grpc", transport: "grpc", secrets: true},
		{name: "secret-overrides-default", secrets: true, literals: true},
		{name: "secret-overrides-http", transport: "http", secrets: true, literals: true},
		{name: "secret-overrides-grpc", transport: "grpc", secrets: true, literals: true},
		{name: "http", transport: "http"},
		{name: "grpc", transport: "grpc"},
		{name: "http-custom-ca", transport: "http", ca: true},
		{name: "grpc-custom-ca", transport: "grpc", ca: true},
		{name: "http-proxy", transport: "http", proxy: true},
		{name: "grpc-proxy", transport: "grpc", proxy: true},
		{name: "http-custom-ca-proxy", transport: "http", ca: true, proxy: true},
		{name: "grpc-custom-ca-proxy", transport: "grpc", ca: true, proxy: true},
		{name: "fixed-promtail", mode: "promtail"},
		{name: "fixed-native-http", transport: "http", mode: "native"},
		{name: "fixed-native-grpc", transport: "grpc", mode: "native"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{
				"stackstate.url": "https://127.0.0.1:18443/receiver/stsAgent",
			}
			if tc.mode != "" {
				values["otelLogsAgent.exportMode"] = tc.mode
			}
			switch tc.transport {
			case "http":
				values["otel.platformHttpOtlpEndpoint"] = "https://127.0.0.1:14318/native"
			case "grpc":
				values["otel.platformGrpcOtlpEndpoint"] = "127.0.0.1:14317"
			}
			if tc.ca {
				values["global.customCertificates.enabled"] = "true"
				values["global.customCertificates.configMapName"] = "validation-ca"
			}
			if tc.proxy {
				values["global.proxy.url"] = "http://127.0.0.1:13128"
			}
			if tc.secrets {
				values["global.url.fromSecret"] = "receiver-url"
				values["global.clusterName.fromSecret"] = "cluster-name"
				values["stackstate.url"] = "null"
				values["stackstate.cluster.name"] = "null"
				if tc.literals {
					values["stackstate.url"] = "unused-url"
					values["stackstate.cluster.name"] = "unused-cluster"
				}
			}
			resources := renderLogsOtel(t, values)
			assertLogsOtelBounds(t, resources, logsOtelConfig(t, resources))
			validateLogsOtelConfig(t, binary, resources)
		})
	}
}

func validateLogsOtelConfig(t *testing.T, binary string, resources helmtestutil.KubernetesResources) {
	t.Helper()
	config := resources.ConfigMaps[otelLogsAgentName].Data["otel-logs.yaml"]
	require.NotContains(t, config, "ca_file:")
	dir := t.TempDir()
	configPath := filepath.Join(dir, "otel-logs.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(config), 0600))

	container := logsContainer(t, resources, otelLogsAgentName)
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
	cm, ok := resources.ConfigMaps[otelLogsAgentName]
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

func logsContainer(t *testing.T, resources helmtestutil.KubernetesResources, name string) corev1.Container {
	t.Helper()
	ds, ok := resources.DaemonSets[name]
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
	assert.Len(t, logsConfigMap(t, config, "connectors"), 1)
	assert.Len(t, logsConfigMap(t, config, "exporters"), 2)
	assert.Len(t, logsConfigMap(t, config, "service", "pipelines"), 3)
	for id, expected := range map[string]map[string]interface{}{
		"logs/input": {
			"receivers":  []interface{}{"file_log/pods"},
			"processors": []interface{}{"memory_limiter", "transform/static_pod", "k8s_attributes", "transform/cluster"},
			"exporters":  []interface{}{"stslogsroute/logs"},
		},
		"logs/promtail": {
			"receivers": []interface{}{"stslogsroute/logs"},
			"exporters": []interface{}{"stsk8slogs/promtail"},
		},
		"logs/native": {
			"receivers": []interface{}{"stslogsroute/logs"},
			"exporters": []interface{}{"otlp_http/native"},
		},
	} {
		assert.Equal(t, expected, logsConfigMap(t, config, "service", "pipelines", id), id)
	}
	route := logsConfigMap(t, config, "connectors", "stslogsroute/logs")
	for key, expected := range map[string]interface{}{
		"controller_extension": "stslogsagent/logs",
		"promtail_pipeline":    "logs/promtail",
		"native_pipeline":      "logs/native",
		"max_concurrent_calls": 8,
		"max_record_bytes":     262144,
		"max_request_bytes":    1048576,
		"export_lifetime":      "90s",
	} {
		assert.EqualValues(t, expected, route[key], key)
	}
	assert.Equal(t, map[string]interface{}{
		"directory": "/var/lib/otelcol/checkpoints", "create_directory": true, "recreate": false,
	}, logsConfigMap(t, config, "extensions", "file_storage/logs"))
	capability := logsConfigMap(t, config, "extensions", "stslogsagent/logs")
	for key, expected := range map[string]interface{}{
		"discovery_enabled": true,
		"receiver_url":      "${env:RECEIVER_URL}", "api_key": "${env:STS_API_KEY}",
		"state_directory": "/var/lib/otelcol/controller", "health_endpoint": "0.0.0.0:13133",
		"termination_message_path": "/dev/termination-log",
	} {
		assert.Equal(t, expected, capability[key], key)
	}
	assert.Equal(t, map[string]interface{}{
		"scheme": "SUSEObservability", "token": "${env:STS_API_KEY}",
	}, logsConfigMap(t, config, "extensions", "bearertokenauth/native"))
	assert.ElementsMatch(t, []string{"file_storage/logs", "stslogsagent/logs", "bearertokenauth/native"},
		logsConfigMap(t, config, "service")["extensions"])
	for _, exporter := range []string{"stsk8slogs/promtail", "otlp_http/native"} {
		export := logsConfigMap(t, config, "exporters", exporter)
		assert.NotContains(t, export, "delivery", exporter)
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
	assert.Equal(t, "${env:NATIVE_OTLP_URL}", logsConfigMap(t, config, "exporters", "otlp_http/native")["endpoint"])
	assert.Equal(t, "bearertokenauth/native", logsConfigMap(t, config, "exporters", "otlp_http/native", "auth")["authenticator"])
	assertLogsOtelBounds(t, resources, config)
}

func assertLogsOtelBounds(t *testing.T, resources helmtestutil.KubernetesResources, config map[string]interface{}) {
	t.Helper()
	route := logsConfigMap(t, config, "connectors", "stslogsroute/logs")
	calls := route["max_concurrent_calls"].(int)
	files := logsConfigMap(t, config, "receivers", "file_log/pods")["max_concurrent_files"].(int)
	lifetime, err := time.ParseDuration(route["export_lifetime"].(string))
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
	grace := resources.DaemonSets[otelLogsAgentName].Spec.Template.Spec.TerminationGracePeriodSeconds
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
		"k8s.deployment.name", "k8s.replicaset.name", "k8s.statefulset.name", "k8s.daemonset.name",
		"k8s.job.name", "k8s.cronjob.name",
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
	container := logsContainer(t, resources, otelLogsAgentName)
	pod := resources.DaemonSets[otelLogsAgentName].Spec.Template.Spec
	assert.Equal(t, otelLogsAgentName, pod.ServiceAccountName)
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
	assert.Equal(t, "/dev/termination-log", container.TerminationMessagePath)
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
	assert.Equal(t, otelLogsAgentName, configVolume.ConfigMap.Name)
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
	role, ok := resources.ClusterRoles[otelLogsAgentName]
	require.True(t, ok)
	assert.Equal(t, []rbacv1.PolicyRule{
		{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get", "list", "watch"}},
	}, role.Rules)
	binding := resources.ClusterRoleBindings[otelLogsAgentName]
	assert.Equal(t, rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: otelLogsAgentName}, binding.RoleRef)
	assert.Contains(t, binding.Subjects, rbacv1.Subject{
		Kind: "ServiceAccount", Name: otelLogsAgentName, Namespace: resources.DaemonSets[otelLogsAgentName].Namespace,
	})
	metrics := logsConfigMap(t, logsOtelConfig(t, resources), "service", "telemetry", "metrics")
	assert.Equal(t, []interface{}{map[string]interface{}{
		"pull": map[string]interface{}{"exporter": map[string]interface{}{"prometheus": map[string]interface{}{"host": "0.0.0.0", "port": 8888}}},
	}}, metrics["readers"])
	annotations := resources.DaemonSets[otelLogsAgentName].Spec.Template.Annotations
	assert.JSONEq(t, `["openmetrics"]`, annotations["ad.stackstate.com/logs-agent.check_names"])
	var instances []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(annotations["ad.stackstate.com/logs-agent.instances"]), &instances))
	require.Len(t, instances, 1)
	assert.Equal(t, "http://%%host%%:8888/metrics", instances[0]["prometheus_url"])
	assert.Contains(t, instances[0]["metrics"], "*")
}

func TestLogsAgentOtelEndpointsAndSecret(t *testing.T) {
	for _, tc := range []struct {
		name, http, grpc, endpoint, exporter string
	}{
		{"derived", "", "", "https://my-suse-observability-instance.com/receiver/stsAgent/otel", "otlp_http/native"},
		{"http", "https://native.example.test/ingest", "", "https://native.example.test/ingest", "otlp_http/native"},
		{"grpc", "", "native.example.test:443", "native.example.test:443", "otlp/native"},
		{"grpc-ipv6", "", "[2001:db8::1]:443", "[2001:db8::1]:443", "otlp/native"},
		{"grpc-ipv6-loopback", "", "[::1]:4317", "[::1]:4317", "otlp/native"},
		{"http-wins", "https://native.example.test/ingest", "unused.example.test:443", "https://native.example.test/ingest", "otlp_http/native"},
		{"http-wins-over-invalid-grpc", "https://native.example.test/ingest", ":", "https://native.example.test/ingest", "otlp_http/native"},
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
			container := logsContainer(t, resources, otelLogsAgentName)
			env := envVarsByName(container.Env)
			assert.Equal(t, tc.endpoint, env["NATIVE_OTLP_URL"])
			assert.Equal(t, "https://my-suse-observability-instance.com/receiver/stsAgent", env["RECEIVER_URL"])
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
			require.Contains(t, exporters, tc.exporter)
			assert.Len(t, exporters, 2)
			assert.Equal(t, []interface{}{tc.exporter}, logsConfigMap(t, config, "service", "pipelines", "logs/native")["exporters"])
			assert.Equal(t, "${env:NATIVE_OTLP_URL}", logsConfigMap(t, exporters, tc.exporter)["endpoint"])
			assert.NotContains(t, resources.ConfigMaps[otelLogsAgentName].Data["otel-logs.yaml"], "foobar")
			assert.NotContains(t, resources.Secrets, "logs-api-key")
		})
	}
	resources := renderLogsOtel(t, map[string]string{"global.apiKey.fromSecret": "existing-receiver-key"})
	assert.Contains(t, logsContainer(t, resources, otelLogsAgentName).Env, corev1.EnvVar{
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
			container := logsContainer(t, resources, otelLogsAgentName)
			assertConfigEnv(t, container.Env, "CLUSTER_NAME", "STS_CLUSTER_NAME", "suse-observability-agent-cluster", "some-k8s-cluster", scenario.clusterSecret)
			discovery := configEnv(t, container.Env, "RECEIVER_URL")
			native := configEnv(t, container.Env, "NATIVE_OTLP_URL")
			endpoint := configEnv(t, container.Env, "PROMTAIL_LOGS_URL")
			if scenario.urlSecret {
				assertConfigEnv(t, container.Env, "STS_URL", "STS_URL", "suse-observability-agent-url", "", true)
				assert.Equal(t, "$(STS_URL)/logs/k8s", endpoint.Value)
				assert.Equal(t, "$(STS_URL)", discovery.Value)
				assert.Equal(t, "$(STS_URL)/otel", native.Value)
				assert.Less(t, envPosition(container.Env, "STS_URL"), envPosition(container.Env, discovery.Name))
				assert.Less(t, envPosition(container.Env, "STS_URL"), envPosition(container.Env, native.Name))
				assert.Less(t, envPosition(container.Env, "STS_URL"), envPosition(container.Env, endpoint.Name))
			} else {
				assert.Equal(t, "https://my-suse-observability-instance.com/receiver/stsAgent/logs/k8s", endpoint.Value)
				assert.Equal(t, "https://my-suse-observability-instance.com/receiver/stsAgent", discovery.Value)
				assert.Equal(t, "https://my-suse-observability-instance.com/receiver/stsAgent/otel", native.Value)
			}
			config := logsOtelConfig(t, resources)
			assert.Equal(t, "${env:CLUSTER_NAME}", logsConfigMap(t, config, "exporters", "stsk8slogs/promtail")["cluster_name"])
			assert.Equal(t, []interface{}{map[string]interface{}{
				"context": "resource", "statements": []interface{}{`set(resource.attributes["k8s.cluster.name"], "${env:CLUSTER_NAME}")`},
			}}, logsConfigMap(t, config, "processors", "transform/cluster")["log_statements"])
			assert.NotContains(t, resources.ConfigMaps[otelLogsAgentName].Data["otel-logs.yaml"], "unused")
			assert.NotContains(t, resources.Secrets, "suse-observability-agent-cluster")
			assert.NotContains(t, resources.Secrets, "suse-observability-agent-url")
		})
	}
}

func TestLogsAgentOtelSelectedImageAndPullSecrets(t *testing.T) {
	for _, otel := range []bool{false, true} {
		t.Run(fmt.Sprint(otel), func(t *testing.T) {
			name := logsAgentName
			if otel {
				name = otelLogsAgentName
			}
			resources := renderLogsAgent(t, map[string]string{
				"global.features.experimentalOtelLogsAgent": "true",
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
			container := logsContainer(t, resources, name)
			secrets := resources.DaemonSets[name].Spec.Template.Spec.ImagePullSecrets
			expected := []corev1.LocalObjectReference{
				{Name: "global-pull"},
				{Name: "custom-suse-observability-agent-pull"},
				{Name: "common-pull"},
				{Name: "suse-observability-agent-pull-secret"},
			}
			if otel {
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
	for _, grpc := range []bool{false, true} {
		for _, ca := range []string{"none", "inline", "external"} {
			t.Run(fmt.Sprintf("grpc=%t/ca=%s", grpc, ca), func(t *testing.T) {
				values := map[string]string{"global.proxy.url": "http://proxy.example.test:3128"}
				exporter := "otlp_http/native"
				if grpc {
					values["otel.platformGrpcOtlpEndpoint"] = "native.example.test:443"
					exporter = "otlp/native"
				}
				if ca != "none" {
					values["global.customCertificates.enabled"] = "true"
					if ca == "inline" {
						values["global.customCertificates.pemData"] = "synthetic-ca-bundle"
					} else {
						values["global.customCertificates.configMapName"] = "external-ca"
					}
				}
				resources := renderLogsOtel(t, values)
				container := logsContainer(t, resources, otelLogsAgentName)
				env := envVarsByName(container.Env)
				assert.Equal(t, "http://proxy.example.test:3128", env["PROXY_URL"])
				config := logsOtelConfig(t, resources)
				for _, client := range []map[string]interface{}{
					logsConfigMap(t, config, "extensions", "stslogsagent/logs"),
					logsConfigMap(t, config, "exporters", "stsk8slogs/promtail"),
					logsConfigMap(t, config, "exporters", exporter),
				} {
					tls := logsConfigMap(t, client, "tls")
					assert.Equal(t, false, tls["insecure_skip_verify"])
					assert.NotContains(t, tls, "ca_file")
					assert.NotContains(t, tls, "include_system_ca_certs_pool")
				}
				for _, client := range []map[string]interface{}{
					logsConfigMap(t, config, "extensions", "stslogsagent/logs"),
					logsConfigMap(t, config, "exporters", "stsk8slogs/promtail"),
				} {
					assert.Equal(t, "${env:PROXY_URL}", client["proxy_url"])
				}
				native := logsConfigMap(t, config, "exporters", exporter)
				if grpc {
					assert.NotContains(t, native, "proxy_url")
					assert.Equal(t, "http://proxy.example.test:3128", env["HTTPS_PROXY"])
					for _, bypass := range []string{"$(KUBERNETES_SERVICE_HOST)", "localhost", "127.0.0.1", "::1", ".svc", ".svc.cluster.local"} {
						assert.Contains(t, strings.Split(env["NO_PROXY"], ","), bypass)
					}
				} else {
					assert.Equal(t, "${env:PROXY_URL}", native["proxy_url"])
					assert.Empty(t, env["HTTPS_PROXY"])
				}
				if ca != "none" {
					volume := requireVolume(t, resources.DaemonSets[otelLogsAgentName].Spec.Template.Spec.Volumes, "custom-certificates")
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
						assert.NotContains(t, resources.DaemonSets[otelLogsAgentName].Spec.Template.Annotations, "checksum/custom-certificates")
					}
				}
			})
		}
	}
}

func TestLogsAgentOtelSkipTLSValidation(t *testing.T) {
	for _, grpc := range []bool{false, true} {
		for _, global := range []bool{false, true} {
			for _, local := range []bool{false, true} {
				t.Run(fmt.Sprintf("grpc=%t/global=%t/local=%t", grpc, global, local), func(t *testing.T) {
					values := map[string]string{
						"global.skipSslValidation":        fmt.Sprint(global),
						"otelLogsAgent.skipSslValidation": fmt.Sprint(local),
					}
					exporter := "otlp_http/native"
					if grpc {
						values["otel.platformGrpcOtlpEndpoint"] = "native.example.test:443"
						exporter = "otlp/native"
					}
					config := logsOtelConfig(t, renderLogsOtel(t, values))
					for _, path := range [][]string{
						{"extensions", "stslogsagent/logs", "tls"},
						{"exporters", "stsk8slogs/promtail", "tls"},
						{"exporters", exporter, "tls"},
					} {
						assert.Equal(t, global || local, logsConfigMap(t, config, path...)["insecure_skip_verify"], path)
					}
				})
			}
		}
	}
}

func TestLogsAgentOtelHTTPSProxyTransportSelection(t *testing.T) {
	for _, tc := range []struct {
		name, http, grpc string
		wantError        bool
	}{
		{name: "derived-http"},
		{name: "explicit-http", http: "http://native.example.test/ingest"},
		{name: "explicit-https", http: "https://native.example.test/ingest"},
		{name: "grpc-only", grpc: "native.example.test:443", wantError: true},
		{name: "http-wins", http: "https://native.example.test/ingest", grpc: "native.example.test:443"},
		{name: "http-wins-over-invalid-grpc", http: "https://native.example.test/ingest", grpc: ":"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability-agent", &helm.Options{
				ValuesFiles: []string{"values/minimal.yaml", "values/logs-otel-base.yaml", "values/logs-otel-enabled.yaml"},
				SetValues: map[string]string{
					"global.proxy.url":              "https://proxy.example.test:8443",
					"otel.platformHttpOtlpEndpoint": tc.http,
					"otel.platformGrpcOtlpEndpoint": tc.grpc,
				},
			})
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assertUniqueLogsManifests(t, output)
			resources := helmtestutil.NewKubernetesResources(t, output)
			env := envVarsByName(logsContainer(t, resources, otelLogsAgentName).Env)
			assert.Equal(t, "https://proxy.example.test:8443", env["PROXY_URL"])
			assert.Empty(t, env["HTTPS_PROXY"])
			if tc.http != "" {
				assert.Equal(t, tc.http, env["NATIVE_OTLP_URL"])
			}
			config := logsOtelConfig(t, resources)
			assert.NotContains(t, logsConfigMap(t, config, "exporters"), "otlp/native")
			assert.Equal(t, "${env:PROXY_URL}", logsConfigMap(t, config, "exporters", "otlp_http/native")["proxy_url"])
			assert.Equal(t, []interface{}{"otlp_http/native"}, logsConfigMap(t, config, "service", "pipelines", "logs/native")["exporters"])
		})
	}
}

func TestLogsAgentOtelChecksumTracksOwnConfigAndInlineCA(t *testing.T) {
	for _, otel := range []bool{false, true} {
		t.Run(fmt.Sprint(otel), func(t *testing.T) {
			name, own, other := logsAgentName, "logsAgent", "otelLogsAgent"
			if otel {
				name, own, other = otelLogsAgentName, other, own
			}
			render := func(overrides map[string]string) map[string]string {
				values := map[string]string{
					"global.features.experimentalOtelLogsAgent": "true",
					"global.customCertificates.enabled":         "true",
					"global.customCertificates.pemData":         "synthetic-ca-one",
				}
				for key, value := range overrides {
					values[key] = value
				}
				return renderLogsAgent(t, values).DaemonSets[name].Spec.Template.Annotations
			}
			base := render(nil)
			require.NotEmpty(t, base["checksum/override-configmap"])
			require.NotEmpty(t, base["checksum/custom-certificates"])
			repeated := render(nil)
			for _, key := range []string{"checksum/override-configmap", "checksum/custom-certificates"} {
				assert.Equal(t, base[key], repeated[key], "identical inputs must produce stable %s", key)
			}
			tls := render(map[string]string{own + ".skipSslValidation": "true"})
			assert.NotEqual(t, base["checksum/override-configmap"], tls["checksum/override-configmap"])
			otherTLS := render(map[string]string{other + ".skipSslValidation": "true"})
			assert.Equal(t, base["checksum/override-configmap"], otherTLS["checksum/override-configmap"])
			memory := render(map[string]string{"otelLogsAgent.resources.limits.memory": "384Mi"})
			if otel {
				assert.NotEqual(t, base["checksum/override-configmap"], memory["checksum/override-configmap"])
			} else {
				assert.Equal(t, base["checksum/override-configmap"], memory["checksum/override-configmap"])
			}
			ca := render(map[string]string{"global.customCertificates.pemData": "synthetic-ca-two"})
			assert.NotEqual(t, base["checksum/custom-certificates"], ca["checksum/custom-certificates"])
			if otel {
				mode := render(map[string]string{"otelLogsAgent.exportMode": "native"})
				assert.NotEqual(t, base["checksum/override-configmap"], mode["checksum/override-configmap"])
			}
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
			container := logsContainer(t, resources, otelLogsAgentName)
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

func TestLogsAgentOtelSynchronousLifetime(t *testing.T) {
	for _, grpc := range []bool{false, true} {
		t.Run(fmt.Sprintf("grpc=%t", grpc), func(t *testing.T) {
			values := map[string]string{}
			if grpc {
				values["otel.platformGrpcOtlpEndpoint"] = "native.example.test:443"
			}
			resources := renderLogsOtel(t, values)
			config := logsOtelConfig(t, resources)
			assertLogsOtelBounds(t, resources, config)
			for name, raw := range logsConfigMap(t, config, "exporters") {
				assert.NotContains(t, raw.(map[string]interface{}), "delivery", name)
			}
		})
	}
}

func TestLogsAgentOtelDiscoveryUsesCollectorPolicy(t *testing.T) {
	for _, grpc := range []bool{false, true} {
		t.Run(fmt.Sprintf("grpc=%t", grpc), func(t *testing.T) {
			values := map[string]string{}
			if grpc {
				values["otel.platformGrpcOtlpEndpoint"] = "native.example.test:443"
			}
			config := logsOtelConfig(t, renderLogsOtel(t, values))
			discovery := logsConfigMap(t, config, "extensions", "stslogsagent/logs")
			assert.Equal(t, true, discovery["discovery_enabled"])
			for _, key := range []string{
				"query_timeout", "attempt_timeout", "max_attempts", "initial_backoff", "max_backoff",
				"poll_interval", "jitter", "stable_observations", "restart_cooldown",
			} {
				assert.NotContains(t, discovery, key)
			}
		})
	}
}

func TestLogsAgentOtelExportMode(t *testing.T) {
	for _, mode := range []string{"auto", "promtail", "native"} {
		for _, grpc := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/grpc=%t", mode, grpc), func(t *testing.T) {
				values := map[string]string{"otelLogsAgent.exportMode": mode}
				exporter := "otlp_http/native"
				if grpc {
					values["otel.platformGrpcOtlpEndpoint"] = "native.example.test:443"
					exporter = "otlp/native"
				}
				resources := renderLogsOtel(t, values)
				config := logsOtelConfig(t, resources)
				assertLogsOtelBounds(t, resources, config)
				controller := logsConfigMap(t, config, "extensions", "stslogsagent/logs")
				if mode == "auto" {
					assert.Equal(t, true, controller["discovery_enabled"])
					assert.NotContains(t, controller, "fixed_mode")
				} else {
					assert.Equal(t, false, controller["discovery_enabled"])
					assert.Equal(t, mode, controller["fixed_mode"])
				}
				route := logsConfigMap(t, config, "connectors", "stslogsroute/logs")
				pipelines := logsConfigMap(t, config, "service", "pipelines")
				exporters := logsConfigMap(t, config, "exporters")
				extensions := logsConfigMap(t, config, "service")["extensions"]
				env := envVarsByName(logsContainer(t, resources, otelLogsAgentName).Env)
				if mode == "promtail" {
					for _, name := range []string{"NATIVE_OTLP_URL", "HTTPS_PROXY", "NO_PROXY"} {
						assert.NotContains(t, env, name)
					}
					assert.NotContains(t, route, "native_pipeline")
					assert.Len(t, pipelines, 2)
					assert.NotContains(t, pipelines, "logs/native")
					assert.Equal(t, []string{"stsk8slogs/promtail"}, mapKeys(exporters))
					assert.NotContains(t, logsConfigMap(t, config, "extensions"), "bearertokenauth/native")
					assert.ElementsMatch(t, []string{"file_storage/logs", "stslogsagent/logs"}, extensions)
				} else {
					assert.Equal(t, "logs/native", route["native_pipeline"])
					assert.Len(t, pipelines, 3)
					assert.Equal(t, []interface{}{exporter}, logsConfigMap(t, pipelines, "logs/native")["exporters"])
					assert.ElementsMatch(t, []string{"stsk8slogs/promtail", exporter}, mapKeys(exporters))
					assert.ElementsMatch(t, []string{"file_storage/logs", "stslogsagent/logs", "bearertokenauth/native"}, extensions)
				}
			})
		}
	}
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/logs-otel-base.yaml", "values/logs-otel-enabled.yaml"},
		SetValues:   map[string]string{"otelLogsAgent.exportMode": "otel"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exportMode")
}

func TestLogsAgentOtelFixedPromtailIgnoresNativeSettings(t *testing.T) {
	for _, tc := range []struct{ name, grpc, proxy string }{
		{"grpc-https-proxy", "native.example.test:443", "https://proxy.example.test:8443"},
		{"invalid-grpc-endpoint", "native.example.test", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{
				"otel.platformGrpcOtlpEndpoint": tc.grpc,
				"global.proxy.url":              tc.proxy,
			}
			for _, mode := range []string{"auto", "native"} {
				values["otelLogsAgent.exportMode"] = mode
				_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability-agent", &helm.Options{
					ValuesFiles: []string{"values/minimal.yaml", "values/logs-otel-base.yaml", "values/logs-otel-enabled.yaml"},
					SetValues:   values,
				})
				require.Error(t, err, mode)
			}
			values["otelLogsAgent.exportMode"] = "promtail"
			resources := renderLogsOtel(t, values)
			env := envVarsByName(logsContainer(t, resources, otelLogsAgentName).Env)
			for _, name := range []string{"NATIVE_OTLP_URL", "HTTPS_PROXY", "NO_PROXY"} {
				assert.NotContains(t, env, name)
			}
			assert.Equal(t, tc.proxy, env["PROXY_URL"])
			config := logsOtelConfig(t, resources)
			assert.Equal(t, []string{"stsk8slogs/promtail"}, mapKeys(logsConfigMap(t, config, "exporters")))
		})
	}
}

func mapKeys(values map[string]interface{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func TestLogsAgentOtelRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct{ key, value, errorField string }{
		{"otelLogsAgent.resources.limits.memory", "0Mi", "otelLogsAgent.resources.limits.memory"},
		{"otel.platformHttpOtlpEndpoint", "native.example.test", "platformHttpOtlpEndpoint"},
		{"otel.platformGrpcOtlpEndpoint", "https://native.example.test:443", "platformGrpcOtlpEndpoint"},
		{"otel.platformGrpcOtlpEndpoint", "native.example.test", "platformGrpcOtlpEndpoint"},
		{"otel.platformGrpcOtlpEndpoint", ":", ""},
		{"otel.platformGrpcOtlpEndpoint", ":4317", ""},
		{"otel.platformGrpcOtlpEndpoint", "native.example.test:", ""},
		{"otel.platformGrpcOtlpEndpoint", "native.example.test:abc", ""},
		{"otel.platformGrpcOtlpEndpoint", "native.example.test:0", ""},
		{"otel.platformGrpcOtlpEndpoint", "native.example.test:-1", ""},
		{"otel.platformGrpcOtlpEndpoint", "native.example.test:65536", ""},
		{"otel.platformGrpcOtlpEndpoint", "::1:4317", ""},
	} {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability-agent", &helm.Options{
				ValuesFiles: []string{"values/minimal.yaml", "values/logs-otel-base.yaml", "values/logs-otel-enabled.yaml"},
				SetValues:   map[string]string{tc.key: tc.value},
			})
			require.Error(t, err)
			if tc.errorField != "" {
				assert.Contains(t, strings.ReplaceAll(err.Error(), "/", "."), tc.errorField)
			}
		})
	}
}
