package test

import (
	"slices"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestExternalConfiguration(t *testing.T) {
	for _, scenario := range []struct {
		name          string
		clusterSecret bool
		urlSecret     bool
		keepLiterals  bool
	}{
		{name: "literal values"},
		{name: "all external", clusterSecret: true, urlSecret: true},
		{name: "cluster external", clusterSecret: true},
		{name: "URL external", urlSecret: true},
		{name: "secrets take precedence", clusterSecret: true, urlSecret: true, keepLiterals: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			values := map[string]string{
				"global.apiKey.fromSecret":      "\\{\\{ .Release.Name }}-config",
				"stackstate.apiKey":             "null",
				"logsAgent.enabled":             "true",
				"checksAgent.enabled":           "true",
				"otel.enabled":                  "true",
				"otel.telemetryGateway.enabled": "true",
				"otel.prometheusScraping.targetAllocator.mtlsEnabled": "false",
			}
			if scenario.clusterSecret {
				values["global.clusterName.fromSecret"] = "\\{\\{ .Release.Name }}-config"
				if !scenario.keepLiterals {
					values["stackstate.cluster.name"] = "null"
				}
			}
			if scenario.urlSecret {
				values["global.url.fromSecret"] = "\\{\\{ .Release.Name }}-url-secret"
				if !scenario.keepLiterals {
					values["stackstate.url"] = "null"
				}
			}
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
				ValuesFiles: []string{"values/minimal.yaml"},
				SetValues:   values,
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			_, clusterConfigExists := resources.ConfigMaps["suse-observability-agent-cluster-name"]
			assert.Equal(t, !scenario.clusterSecret, clusterConfigExists)
			_, urlConfigExists := resources.ConfigMaps["suse-observability-agent-url"]
			assert.Equal(t, !scenario.urlSecret, urlConfigExists)
			assert.NotContains(t, resources.Secrets, "suse-observability-agent-config")
			assert.NotContains(t, resources.Secrets, "suse-observability-agent-url-secret")

			containers := map[string]corev1.Container{}
			for _, deployment := range resources.Deployments {
				for _, container := range deployment.Spec.Template.Spec.Containers {
					containers[container.Name] = container
				}
			}
			for _, daemonset := range resources.DaemonSets {
				for _, container := range daemonset.Spec.Template.Spec.Containers {
					containers[container.Name] = container
				}
			}
			for _, statefulset := range resources.Statefulsets {
				for _, container := range statefulset.Spec.Template.Spec.Containers {
					containers[container.Name] = container
				}
			}
			for _, containerName := range []string{"node-agent", "process-agent", "cluster-agent", "suse-observability-agent", "logs-agent", "kubernetes-rbac-agent", "k8s-resource-collector", "otel-metrics-scraper", "otel-telemetry-gateway"} {
				t.Run(containerName, func(t *testing.T) {
					container, exists := containers[containerName]
					require.True(t, exists)
					isCollector := slices.Contains([]string{"k8s-resource-collector", "otel-metrics-scraper", "otel-telemetry-gateway"}, containerName)
					clusterEnvName := "STS_CLUSTER_NAME"
					if isCollector {
						clusterEnvName = "K8S_CLUSTER_NAME"
					}
					if scenario.clusterSecret || containerName != "kubernetes-rbac-agent" {
						assertConfigEnv(t, container.Env, clusterEnvName, "STS_CLUSTER_NAME", "suse-observability-agent-config", "some-k8s-cluster", scenario.clusterSecret)
					}
					urlEnvName := "STS_STS_URL"
					if containerName == "process-agent" {
						urlEnvName = "STS_PROCESS_AGENT_URL"
					} else if isCollector || containerName == "logs-agent" || containerName == "kubernetes-rbac-agent" {
						urlEnvName = "STS_URL"
					}
					if scenario.urlSecret || urlEnvName != "STS_URL" {
						assertConfigEnv(t, container.Env, urlEnvName, "STS_URL", "suse-observability-agent-url-secret", "https://my-suse-observability-instance.com/receiver", scenario.urlSecret)
					}
					if isCollector {
						endpoint := configEnv(t, container.Env, "PLATFORM_OTLP_ENDPOINT")
						if scenario.urlSecret {
							assert.Equal(t, "$(STS_URL)/otel", endpoint.Value)
							assert.Less(t, envPosition(container.Env, "STS_URL"), envPosition(container.Env, endpoint.Name))
						} else {
							assert.Equal(t, "https://my-suse-observability-instance.com/receiver/otel", endpoint.Value)
						}
					}
					if envPosition(container.Env, "STS_HOSTNAME") >= 0 {
						hostname := configEnv(t, container.Env, "STS_HOSTNAME")
						if scenario.clusterSecret {
							assert.Equal(t, "$(KUBERNETES_HOSTNAME)-$(STS_CLUSTER_NAME)", hostname.Value)
						} else {
							assert.Equal(t, "$(KUBERNETES_HOSTNAME)-some-k8s-cluster", hostname.Value)
						}
						assert.Less(t, envPosition(container.Env, "STS_CLUSTER_NAME"), envPosition(container.Env, hostname.Name))
						assert.Less(t, envPosition(container.Env, "KUBERNETES_HOSTNAME"), envPosition(container.Env, hostname.Name))
					}
					if containerName == "kubernetes-rbac-agent" {
						assert.Contains(t, container.EnvFrom, corev1.EnvFromSource{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "suse-observability-agent-config"}}})
						for _, source := range []struct {
							name     string
							external bool
						}{
							{"suse-observability-agent-cluster-name", scenario.clusterSecret},
							{"suse-observability-agent-url", scenario.urlSecret},
						} {
							assert.Equal(t, !source.external, slices.ContainsFunc(container.EnvFrom, func(actual corev1.EnvFromSource) bool {
								return actual.ConfigMapRef != nil && actual.ConfigMapRef.Name == source.name
							}))
						}
					} else {
						assertConfigEnv(t, container.Env, "STS_API_KEY", "STS_API_KEY", "suse-observability-agent-config", "", true)
					}
				})
			}
			logsConfig := resources.ConfigMaps["suse-observability-agent-logs-agent"].Data["promtail.yaml"]
			collectorConfig := resources.ConfigMaps["suse-observability-agent-k8s-resource-collector-config"].Data["config.yaml"]
			if scenario.clusterSecret {
				assert.Contains(t, logsConfig, `sts_cluster_name: "${STS_CLUSTER_NAME}"`)
				assert.Contains(t, collectorConfig, `cluster_name: "${env:K8S_CLUSTER_NAME}"`)
			} else {
				assert.Contains(t, logsConfig, `sts_cluster_name: "some-k8s-cluster"`)
				assert.Contains(t, collectorConfig, `cluster_name: "some-k8s-cluster"`)
			}
			if scenario.urlSecret {
				assert.Contains(t, logsConfig, "url: ${STS_URL}/logs/k8s")
			} else {
				assert.Contains(t, logsConfig, "url: https://my-suse-observability-instance.com/receiver/logs/k8s")
			}
		})
	}
}

func TestExternalConfigurationOtlpOverrides(t *testing.T) {
	for _, setting := range []string{"otel.platformHttpOtlpEndpoint", "otel.platformGrpcOtlpEndpoint"} {
		t.Run(setting, func(t *testing.T) {
			endpoint := "otlp.example.com:443"
			if setting == "otel.platformHttpOtlpEndpoint" {
				endpoint = "https://" + endpoint
			}
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
				SetValues: map[string]string{
					"global.apiKey.fromSecret":                            "config",
					"global.clusterName.fromSecret":                       "config",
					"global.url.fromSecret":                               "config",
					"otel.enabled":                                        "true",
					"otel.telemetryGateway.enabled":                       "true",
					"otel.prometheusScraping.targetAllocator.mtlsEnabled": "false",
					setting: endpoint,
				},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			for _, deploymentName := range []string{"suse-observability-agent-k8s-resource-collector", "suse-observability-agent-otel-telemetry-gateway"} {
				deployment, exists := resources.Deployments[deploymentName]
				require.True(t, exists)
				assert.Equal(t, endpoint, configEnv(t, deployment.Spec.Template.Spec.Containers[0].Env, "PLATFORM_OTLP_ENDPOINT").Value)
			}
			scraper, exists := resources.Statefulsets[otelMetricsScraperName]
			require.True(t, exists)
			assert.Equal(t, endpoint, configEnv(t, scraper.Spec.Template.Spec.Containers[0].Env, "PLATFORM_OTLP_ENDPOINT").Value)
		})
	}
}

func TestExternalConfigurationMissingSource(t *testing.T) {
	for _, field := range []string{"stackstate.cluster.name", "stackstate.url"} {
		t.Run(field, func(t *testing.T) {
			_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability-agent", &helm.Options{
				ValuesFiles: []string{"values/minimal.yaml"},
				SetValues:   map[string]string{field: "null"},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "schema")
		})
	}
}

func envPosition(env []corev1.EnvVar, name string) int {
	return slices.IndexFunc(env, func(variable corev1.EnvVar) bool { return variable.Name == name })
}

func configEnv(t *testing.T, env []corev1.EnvVar, name string) corev1.EnvVar {
	t.Helper()
	position := envPosition(env, name)
	require.NotEqual(t, -1, position, "missing environment variable %s", name)
	assert.Equal(t, -1, envPosition(env[position+1:], name), "duplicate environment variable %s", name)
	return env[position]
}

func assertConfigEnv(t *testing.T, env []corev1.EnvVar, name, key, secret, literal string, external bool) {
	t.Helper()
	expected := corev1.EnvVar{Name: name, Value: literal}
	if external {
		expected.Value = ""
		expected.ValueFrom = &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: secret},
			Key:                  key,
		}}
	}
	assert.Equal(t, expected, configEnv(t, env, name))
}
