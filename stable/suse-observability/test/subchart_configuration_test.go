package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func renderConnectionConfiguration(t *testing.T, release, values string, upgrade bool) helmtestutil.KubernetesResources {
	t.Helper()
	path := filepath.Join(t.TempDir(), "connection.yaml")
	require.NoError(t, os.WriteFile(path, []byte(values), 0600))
	var args []string
	if upgrade {
		args = append(args, "--is-upgrade")
	}
	output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, release, &helm.Options{
		ValuesFiles:    []string{"values/full.yaml", path},
		KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
	}, args...)
	return helmtestutil.NewKubernetesResources(t, output)
}

func TestSubchartConnectionsUsePlatformConfiguration(t *testing.T) {
	for _, tc := range []struct{ name, release, values, secret string }{
		{"default", "suse-observability", "", "suse-observability-s3proxy"},
		{"custom-release", "nightly", "", "suse-observability-s3proxy"},
		{"external-credentials", "nightly", "global:\n  s3proxy:\n    credentials:\n      fromExternalSecret: customer-s3proxy\n", "customer-s3proxy"},
		{"ignored-saved-inputs", "nightly", `anomaly-detection:
  enabled: true
  stackstate:
    instance: '{{ fail "must not evaluate saved instance" }}'
victoria-metrics-0:
  backup:
    awsSecrets: '{{ fail "must not evaluate saved secret" }}'
    overrideS3Endpoint: '{{ fail "must not evaluate saved endpoint" }}'
victoria-metrics-1:
  backup:
    awsSecrets: old-secret
    overrideS3Endpoint: https://old-storage.example
clickhouse:
  backup:
    s3:
      endpoint: '{{ fail "must not evaluate saved endpoint" }}'
      secretName: '{{ fail "must not evaluate saved secret" }}'
`, "suse-observability-s3proxy"},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				// full.yaml already enables backups; add anomaly detection for every case.
				values := "anomaly-detection:\n  enabled: true\n" + tc.values
				if tc.name == "ignored-saved-inputs" {
					values = tc.values
				}
				resources := renderConnectionConfiguration(t, tc.release, values, upgrade)
				for _, name := range []string{"suse-observability-victoria-metrics-0", "suse-observability-victoria-metrics-1"} {
					require.Contains(t, resources.Statefulsets, name)
					pod := resources.Statefulsets[name].Spec.Template.Spec
					assertConnectionEnv(t, pod.InitContainers, "S3_ENDPOINT", "http://suse-observability-s3proxy:9000")
					assertConnectionCredentials(t, pod.Containers, []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}, tc.secret)
				}
				require.Contains(t, resources.Statefulsets, "suse-observability-clickhouse-shard0")
				assertConnectionCredentials(t, resources.Statefulsets["suse-observability-clickhouse-shard0"].Spec.Template.Spec.Containers, []string{"S3_ACCESS_KEY", "S3_SECRET_KEY"}, tc.secret)
				configs := 0
				for _, cm := range resources.ConfigMaps {
					if _, ok := cm.Data["backup_enabled"]; ok {
						assert.Contains(t, cm.Data["config.yaml"], `endpoint: "http://suse-observability-s3proxy:9000"`)
						configs++
					}
				}
				assert.Equal(t, 1, configs)
				router := "suse-observability-router"
				if tc.release != "suse-observability" {
					router = tc.release + "-" + router
				}
				clients := 0
				for _, deployment := range resources.Deployments {
					for _, c := range deployment.Spec.Template.Spec.Containers {
						for i, arg := range c.Args {
							if arg == "--instance" {
								require.Less(t, i+1, len(c.Args))
								assert.Equal(t, "http://"+router+":8080", c.Args[i+1])
								clients++
							}
						}
					}
				}
				assert.Equal(t, 2, clients)
			})
		}
	}
}

func assertConnectionEnv(t *testing.T, containers []corev1.Container, key, expected string) {
	t.Helper()
	found := false
	for _, container := range containers {
		for _, env := range container.Env {
			if env.Name == key {
				assert.Equal(t, expected, env.Value)
				found = true
			}
		}
	}
	require.True(t, found, "missing environment variable %s", key)
}

func assertConnectionCredentials(t *testing.T, containers []corev1.Container, keys []string, secret string) {
	t.Helper()
	for _, key := range keys {
		found := false
		for _, container := range containers {
			for _, env := range container.Env {
				if env.Name == key {
					require.NotNil(t, env.ValueFrom)
					require.NotNil(t, env.ValueFrom.SecretKeyRef)
					assert.Equal(t, secret, env.ValueFrom.SecretKeyRef.Name)
					found = true
				}
			}
		}
		require.True(t, found, "missing credential %s", key)
	}
}

func TestBackupConnectionChecksumsUseEffectiveConfiguration(t *testing.T) {
	baseline := renderConnectionConfiguration(t, "nightly", "", false)
	ignored := renderConnectionConfiguration(t, "nightly", `victoria-metrics-0:
  backup:
    overrideS3Endpoint: https://old.example
    awsSecrets: old-secret
victoria-metrics-1:
  backup:
    overrideS3Endpoint: https://old.example
    awsSecrets: old-secret
clickhouse:
  backup:
    s3:
      endpoint: old.example
      secretName: old-secret
`, true)
	external := renderConnectionConfiguration(t, "nightly", "global:\n  s3proxy:\n    credentials:\n      fromExternalSecret: customer-secret\n", true)
	for _, name := range []string{"suse-observability-victoria-metrics-0", "suse-observability-victoria-metrics-1", "suse-observability-clickhouse-shard0"} {
		original := baseline.Statefulsets[name].Spec.Template.Annotations["checksum/backup-config"]
		require.NotEmpty(t, original)
		assert.Equal(t, original, ignored.Statefulsets[name].Spec.Template.Annotations["checksum/backup-config"], name)
		assert.NotEqual(t, original, external.Statefulsets[name].Spec.Template.Annotations["checksum/backup-config"])
		assert.Equal(t, baseline.Statefulsets[name].Spec.VolumeClaimTemplates, ignored.Statefulsets[name].Spec.VolumeClaimTemplates)
	}
}

func TestOtelInstrumentationNamespaceConfiguration(t *testing.T) {
	for _, tc := range []struct{ name, values, expected string }{
		{"computed-default", "", "suse-observability-observability"},
		{"shared-override", "        serviceNamespace: shared\n", "shared"},
		{"shared-template", "        serviceNamespace: '{{ .Release.Name }}-tracing'\n", "nightly-tracing"},
		{"component-override", "        serviceNamespace: shared\n    api:\n      otelInstrumentation:\n        serviceNamespace: api-tracing\n", "api-tracing"},
		{"component-template", "    api:\n      otelInstrumentation:\n        serviceNamespace: '{{ .Release.Namespace }}-api'\n", "observability-api"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := "stackstate:\n  components:\n    all:\n      otelInstrumentation:\n        enabled: true\n" + tc.values
			resources := renderConnectionConfiguration(t, "nightly", values, false)
			require.Contains(t, resources.Deployments, "nightly-suse-observability-api")
			assertConnectionEnv(t, resources.Deployments["nightly-suse-observability-api"].Spec.Template.Spec.Containers, "OTEL_RESOURCE_ATTRIBUTES", "service.namespace="+tc.expected+",service.instance.id=$(POD_NAME)")
		})
	}
}

func TestOtelInstrumentationDisablesKafkaClientSpans(t *testing.T) {
	enabled := "stackstate:\n  components:\n    all:\n      otelInstrumentation:\n        enabled: true\n"
	for _, tc := range []struct{ name, values, expected string }{
		{"default", enabled, "false"},
		{"extra-env-override", enabled + "    api:\n      extraEnv:\n        open:\n          OTEL_INSTRUMENTATION_KAFKA_CLIENTS_ENABLED: \"true\"\n", "true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := renderConnectionConfiguration(t, "nightly", tc.values, false)
			require.Contains(t, resources.Deployments, "nightly-suse-observability-api")
			assertConnectionEnv(t, resources.Deployments["nightly-suse-observability-api"].Spec.Template.Spec.Containers, "OTEL_INSTRUMENTATION_KAFKA_CLIENTS_ENABLED", tc.expected)
		})
	}
}

func TestOtelInstrumentationEnablesKafkaClientMetrics(t *testing.T) {
	enabled := "stackstate:\n  components:\n    all:\n      otelInstrumentation:\n        enabled: true\n"
	for _, tc := range []struct{ name, values, expected string }{
		{"default", enabled, "true"},
		{"extra-env-override", enabled + "    api:\n      extraEnv:\n        open:\n          OTEL_INSTRUMENTATION_KAFKA_CLIENTS_METRICS_ENABLED: \"false\"\n", "false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := renderConnectionConfiguration(t, "nightly", tc.values, false)
			require.Contains(t, resources.Deployments, "nightly-suse-observability-api")
			assertConnectionEnv(t, resources.Deployments["nightly-suse-observability-api"].Spec.Template.Spec.Containers, "OTEL_INSTRUMENTATION_KAFKA_CLIENTS_METRICS_ENABLED", tc.expected)
		})
	}
}

func TestPlatformDefaultValuesContainNoTemplateExpressions(t *testing.T) {
	values, err := os.ReadFile("../values.yaml")
	require.NoError(t, err)
	for i, line := range strings.Split(string(values), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			assert.NotContains(t, line, "{{", "values.yaml:%d: compute internal defaults in helpers", i+1)
		}
	}
}
