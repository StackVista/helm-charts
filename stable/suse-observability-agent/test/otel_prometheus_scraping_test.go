package test

import (
	"slices"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	otelMetricsScraperName        = "suse-observability-agent-otel-metrics-scraper"
	otelMetricsScraperConfigName  = "suse-observability-agent-otel-metrics-scraper-config"
	otelTargetAllocatorName       = "suse-observability-agent-otel-target-allocator"
	otelTargetAllocatorConfigName = "suse-observability-agent-otel-target-allocator-config"
)

func TestOtelPrometheusScrapingDisabledByDefault(t *testing.T) {
	// otel.enabled defaults to false; otel.prometheusScraping.enabled defaults to true.
	// The master switch must be on for any OTel resources to be rendered.
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assertOtelPrometheusScrapingResourcesExist(t, resources, false)
	_, exists := resources.Pdbs[otelMetricsScraperName]
	assert.False(t, exists, "metrics scraper PDB should not exist when disabled")
	_, exists = resources.Pdbs[otelTargetAllocatorName]
	assert.False(t, exists, "target allocator PDB should not exist when disabled")
	_, exists = resources.Unmapped["servicemonitors.monitoring.coreos.com"]
	assert.False(t, exists, "ServiceMonitor CRD should not exist when disabled")
	_, exists = resources.Unmapped["podmonitors.monitoring.coreos.com"]
	assert.False(t, exists, "PodMonitor CRD should not exist when disabled")
}

func TestOtelPrometheusScrapingDisabledWhenScrapingFlagFalse(t *testing.T) {
	// otel.enabled=true but otel.prometheusScraping.enabled=false: scraping is
	// explicitly opted out even though the master OTel switch is on.
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml"},
		SetValues: map[string]string{
			"otel.enabled":                    "true",
			"otel.prometheusScraping.enabled": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	assertOtelPrometheusScrapingResourcesExist(t, resources, false)
}

func TestOtelPrometheusScrapingEnabledRendersResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assertOtelPrometheusScrapingResourcesExist(t, resources, true)
	serviceMonitorCRD, exists := resources.Unmapped["servicemonitors.monitoring.coreos.com"]
	assert.True(t, exists, "ServiceMonitor CRD should exist when enabled")
	assert.Contains(t, serviceMonitorCRD, "helm.sh/resource-policy: keep",
		"ServiceMonitor CRD should carry the keep resource-policy annotation by default so it survives helm uninstall")
	podMonitorCRD, exists := resources.Unmapped["podmonitors.monitoring.coreos.com"]
	assert.True(t, exists, "PodMonitor CRD should exist when enabled")
	assert.Contains(t, podMonitorCRD, "helm.sh/resource-policy: keep",
		"PodMonitor CRD should carry the keep resource-policy annotation by default so it survives helm uninstall")
	_, exists = resources.Pdbs[otelMetricsScraperName]
	assert.False(t, exists, "metrics scraper PDB should be skipped at default replicaCount=1")
	_, exists = resources.Pdbs[otelTargetAllocatorName]
	assert.False(t, exists, "target allocator PDB should be skipped at default replicaCount=1")

	statefulSet, exists := resources.Statefulsets[otelMetricsScraperName]
	require.True(t, exists, "metrics scraper StatefulSet should exist")
	require.Len(t, statefulSet.Spec.Template.Spec.Containers, 1)
	container := statefulSet.Spec.Template.Spec.Containers[0]
	assert.Contains(t, container.Image, "stackstate/sts-opentelemetry-collector")
	assert.Empty(t, container.Command, "image's default entrypoint is used; no explicit command override")
	assert.Equal(t, int32(1), *statefulSet.Spec.Replicas)

	envVars := envVarsByName(container.Env)
	assert.Equal(t, "https://my-suse-observability-instance.com/receiver/otel", envVars["PLATFORM_OTLP_ENDPOINT"])
	assert.Equal(t, "some-k8s-cluster", envVars["K8S_CLUSTER_NAME"])
	assert.Equal(t, "false", envVars["SKIP_SSL_VALIDATION"])
	assertEnvFromFieldRef(t, container.Env, "POD_NAME", "metadata.name")
	assertEnvFromFieldRef(t, container.Env, "POD_NAMESPACE", "metadata.namespace")
	assertEnvFromFieldRef(t, container.Env, "POD_IP", "status.podIP")
	assertEnvFromFieldRef(t, container.Env, "POD_UID", "metadata.uid")

	allocator, exists := resources.Deployments[otelTargetAllocatorName]
	require.True(t, exists, "target allocator Deployment should exist")
	assert.Contains(t, allocator.Spec.Template.Spec.Containers[0].Image, "stackstate/opentelemetry-target-allocator")
	assert.True(t, strings.HasPrefix(allocator.Spec.Template.Spec.Containers[0].Image, "quay.io/"), "target allocator image should inherit global.imageRegistry by default")
	allocatorEnv := envVarsByName(allocator.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "opentelemetry-target-allocator", allocatorEnv["OTEL_SERVICE_NAME"])
	assert.Equal(t, "k8s.pod.uid=$(POD_UID),k8s.pod.name=$(POD_NAME)", allocatorEnv["OTEL_RESOURCE_ATTRIBUTES"])
	assertEnvFromFieldRef(t, allocator.Spec.Template.Spec.Containers[0].Env, "POD_NAME", "metadata.name")
	assertEnvFromFieldRef(t, allocator.Spec.Template.Spec.Containers[0].Env, "POD_UID", "metadata.uid")
}

func TestOtelPrometheusScrapingCrdsCanBeDisabled(t *testing.T) {
	output := renderOtelPrometheusScrapingWithPrometheusCRDs(t, "values/otel-prometheus-scraping-crds-disabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assertOtelPrometheusScrapingResourcesExist(t, resources, true)
	_, exists := resources.Unmapped["servicemonitors.monitoring.coreos.com"]
	assert.False(t, exists, "ServiceMonitor CRD should not exist when CRD installation is disabled")
	_, exists = resources.Unmapped["podmonitors.monitoring.coreos.com"]
	assert.False(t, exists, "PodMonitor CRD should not exist when CRD installation is disabled")
}

func TestOtelPrometheusScrapingMonitorCrdsKeepCanBeDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml"},
		SetValues: map[string]string{
			"otel.prometheusScraping.monitorCrds.keep": "false",
		},
	},
		"--api-versions", "monitoring.coreos.com/v1/ServiceMonitor",
		"--api-versions", "monitoring.coreos.com/v1/PodMonitor",
	)
	resources := helmtestutil.NewKubernetesResources(t, output)

	serviceMonitorCRD, exists := resources.Unmapped["servicemonitors.monitoring.coreos.com"]
	require.True(t, exists, "ServiceMonitor CRD should still be installed when keep=false")
	assert.NotContains(t, serviceMonitorCRD, "helm.sh/resource-policy: keep",
		"ServiceMonitor CRD must not carry the keep annotation when keep=false")
	podMonitorCRD, exists := resources.Unmapped["podmonitors.monitoring.coreos.com"]
	require.True(t, exists, "PodMonitor CRD should still be installed when keep=false")
	assert.NotContains(t, podMonitorCRD, "helm.sh/resource-policy: keep",
		"PodMonitor CRD must not carry the keep annotation when keep=false")
}

func TestOtelPrometheusScrapingCrdsDisabledWarnsOnMissingCRDs(t *testing.T) {
	// Missing CRDs produce a warning in NOTES.txt, not a hard failure.
	helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-crds-disabled.yaml")
}

func TestOtelPrometheusScrapingCollectorConfig(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelMetricsScraperConfigName]
	require.True(t, exists, "metrics scraper ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	assert.Contains(t, configData, "prometheus:")
	assert.Contains(t, configData, "target_allocator:")
	assert.Contains(t, configData, "endpoint: \"http://suse-observability-agent-otel-target-allocator:8080\"")
	assert.Contains(t, configData, "collector_id: ${env:POD_NAME}")
	assert.Contains(t, configData, "prometheus/self:")
	assert.Contains(t, configData, "job_name: opentelemetry-collector")
	assert.Contains(t, configData, "${env:POD_IP}:8888")
	assert.Contains(t, configData, "k8s.pod.uid: ${env:POD_UID}")
	// The opentelemetry-target-allocator scrape job is now distributed by the target
	// allocator itself (see TestOtelPrometheusScrapingTargetAllocatorConfig) so that
	// it is assigned to exactly one collector replica.
	assert.NotContains(t, configData, "prometheus/target-allocator-metrics:")
	assert.NotContains(t, configData, "job_name: opentelemetry-target-allocator")
	assert.NotContains(t, configData, "receivers: [prometheus, prometheus/self]")
	assert.Contains(t, configData, "limit_percentage: 80")
	assert.Contains(t, configData, "spike_limit_percentage: 25")
	assert.Contains(t, configData, "transform/pre-k8sattributes:")
	assert.Contains(t, configData, "delete_key(attributes, \"k8s.namespace.name\")")
	assert.Contains(t, configData, "k8sattributes:")
	assert.Contains(t, configData, "key_regex: (.*)")
	assert.Contains(t, configData, "tag_name: $$1")
	assert.Contains(t, configData, "k8s.namespace.name")
	assert.Contains(t, configData, "k8s.pod.uid")
	assert.Contains(t, configData, "from: connection")
	assert.Contains(t, configData, "transform:")
	assert.Contains(t, configData, "error_mode: ignore")
	assert.Contains(t, configData, "metric_statements:")
	assert.Contains(t, configData, "set(attributes[\"k8s.cluster.name\"], \"${env:K8S_CLUSTER_NAME}\")")
	assert.Contains(t, configData, "set(attributes[\"service.instance.id\"], attributes[\"k8s.pod.uid\"])")
	assert.Contains(t, configData, "set(attributes[\"service.namespace\"], attributes[\"k8s.namespace.name\"])")
	assert.Contains(t, configData, "transform/self-metrics:")
	assert.Contains(t, configData, "set(attributes[\"service.name\"], \"suse-observability-agent-otel-metrics-scraper\")")
	assert.Contains(t, configData, "metrics/prometheus-scraping:")
	assert.Contains(t, configData, "otlphttp/suse-observability:")
	assert.Contains(t, configData, "endpoint: ${env:PLATFORM_OTLP_ENDPOINT}")
	assert.Contains(t, configData, "insecure_skip_verify: ${env:SKIP_SSL_VALIDATION}")
	assert.Contains(t, configData, "bearertokenauth:")
	assert.Contains(t, configData, "scheme: SUSEObservability")
	assert.Contains(t, configData, "token: \"${env:STS_API_KEY}\"")
	assert.Contains(t, configData, "pprof:")
	assert.Contains(t, configData, "endpoint: 0.0.0.0:1777")
	assert.Contains(t, configData, "extensions: [health_check, pprof, bearertokenauth]")
	assert.Contains(t, configData, "readers:")
	assert.Contains(t, configData, "host: ${env:POD_IP}")
	assert.Contains(t, configData, "port: 8888")
	assert.Contains(t, configData, "receivers: [prometheus]")
	assert.Contains(t, configData, "processors: [memory_limiter, transform/pre-k8sattributes, k8sattributes, transform, batch]")
	assert.Contains(t, configData, "metrics/self:")
	assert.Contains(t, configData, "receivers: [prometheus/self]")
	assert.Contains(t, configData, "processors: [memory_limiter, transform/pre-k8sattributes, k8sattributes, transform, transform/self-metrics, batch]")
	assert.NotContains(t, configData, "sts_api_key")
}

func TestOtelPrometheusScrapingPprofCanBeDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml"},
		SetValues: map[string]string{
			"otel.prometheusScraping.collector.pprof.enabled": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelMetricsScraperConfigName]
	require.True(t, exists, "metrics scraper ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	assert.NotContains(t, configData, "pprof:", "pprof extension must be omitted when pprof.enabled=false")
	assert.NotContains(t, configData, "endpoint: 0.0.0.0:1777")
	assert.Contains(t, configData, "extensions: [health_check, bearertokenauth]",
		"service.extensions must drop pprof when pprof.enabled=false")
}

func TestOtelPrometheusScrapingTargetAllocatorConfig(t *testing.T) {
	output := renderOtelPrometheusScrapingWithCertManager(t, "values/otel-prometheus-scraping-auth-secrets.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTargetAllocatorConfigName]
	require.True(t, exists, "target allocator ConfigMap should exist")
	configData := configMap.Data["targetallocator.yaml"]

	assert.Contains(t, configData, "collector_namespace: \"default\"")
	assert.Contains(t, configData, "collector_selector:")
	assert.Contains(t, configData, "listen_addr: \":8080\"")
	assert.Contains(t, configData, "allocation_strategy: consistent-hashing")
	assert.Contains(t, configData, "filter_strategy: relabel-config")
	assert.Contains(t, configData, "allow_insecure_auth_secrets: false")
	// Static scrape_configs distributed by the target allocator (avoids duplicate
	// scrapes when collector.replicaCount > 1).
	assert.Contains(t, configData, "config:")
	assert.Contains(t, configData, "scrape_configs:")
	assert.Contains(t, configData, "job_name: opentelemetry-target-allocator")
	assert.Contains(t, configData, "kubernetes_sd_configs:")
	assert.Contains(t, configData, "app.kubernetes.io/component=otel-target-allocator")
	assert.Contains(t, configData, "https:")
	assert.Contains(t, configData, "enabled: true")
	assert.Contains(t, configData, "listen_addr: \":8443\"")
	assert.Contains(t, configData, "ca_file_path: /etc/otel-target-allocator/ca/ca.crt")
	assert.Contains(t, configData, "prometheus_cr:")
	assert.Contains(t, configData, "enabled: true")
	assert.Contains(t, configData, "service_monitor_selector:")
	assert.Contains(t, configData, "pod_monitor_selector:")
	assert.Contains(t, configData, "matchLabels:")
	assert.Contains(t, configData, "observability.suse.com/agent: scrape")
	assert.Contains(t, configData, "secret_namespaces:")
	assert.Contains(t, configData, "- monitoring")
	assert.Contains(t, configData, "- application-metrics")
	assert.Contains(t, configData, "deny_fs_access_through_sms: true")
}

func TestOtelPrometheusScrapingMtlsEnabledRequiresCertManager(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-auth-secrets.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires cert-manager CRDs")
}

func TestOtelPrometheusScrapingMtlsEnabledConfiguresMTLS(t *testing.T) {
	output := renderOtelPrometheusScrapingWithCertManager(t, "values/otel-prometheus-scraping-auth-secrets.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelMetricsScraperConfigName]
	require.True(t, exists, "metrics scraper ConfigMap should exist")
	configData := configMap.Data["config.yaml"]
	assert.Contains(t, configData, "endpoint: \"https://suse-observability-agent-otel-target-allocator:8443\"")
	assert.Contains(t, configData, "ca_file: /etc/otel-target-allocator/ca/ca.crt")
	assert.Contains(t, configData, "cert_file: /etc/otel-target-allocator/tls/tls.crt")
	assert.Contains(t, configData, "key_file: /etc/otel-target-allocator/tls/tls.key")
	assert.Contains(t, configData, "reload_interval: 1m")

	allocator, exists := resources.Deployments[otelTargetAllocatorName]
	require.True(t, exists, "target allocator Deployment should exist")
	allocatorEnvVars := envVarsByName(allocator.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "false", allocatorEnvVars["ALLOW_INSECURE_AUTH_SECRETS"])
	assert.Len(t, allocator.Spec.Template.Spec.Volumes, 3)
	assert.Contains(t, resources.Unmapped, "suse-observability-agent-otel-target-allocator-selfsigned")
	assert.Contains(t, resources.Unmapped, "suse-observability-agent-otel-target-allocator-ca")
	assert.Contains(t, resources.Unmapped, "suse-observability-agent-otel-target-allocator-tls")
	assert.Contains(t, resources.Unmapped, "suse-observability-agent-otel-metrics-scraper-tls")

	service, exists := resources.Services[otelTargetAllocatorName]
	require.True(t, exists, "target allocator Service should exist")
	assert.Len(t, service.Spec.Ports, 2)
	assert.Equal(t, "https", service.Spec.Ports[1].Name)
}

func TestOtelPrometheusScrapingInsecureAuthSecretsUseHTTP(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml"},
		SetValues: map[string]string{
			"otel.prometheusScraping.targetAllocator.allowInsecureAuthSecrets": "true",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTargetAllocatorConfigName]
	require.True(t, exists, "target allocator ConfigMap should exist")
	targetAllocatorConfig := configMap.Data["targetallocator.yaml"]
	assert.Contains(t, targetAllocatorConfig, "allow_insecure_auth_secrets: true")
	assert.NotContains(t, targetAllocatorConfig, "https:")

	allocator, exists := resources.Deployments[otelTargetAllocatorName]
	require.True(t, exists, "target allocator Deployment should exist")
	allocatorEnvVars := envVarsByName(allocator.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "true", allocatorEnvVars["ALLOW_INSECURE_AUTH_SECRETS"])

	configMap, exists = resources.ConfigMaps[otelMetricsScraperConfigName]
	require.True(t, exists, "metrics scraper ConfigMap should exist")
	collectorConfig := configMap.Data["config.yaml"]
	assert.Contains(t, collectorConfig, "endpoint: \"http://suse-observability-agent-otel-target-allocator:8080\"")
	assert.NotContains(t, collectorConfig, "ca_file: /etc/otel-target-allocator/ca/ca.crt")
}

func TestOtelPrometheusScrapingTargetAllocatorRBAC(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	clusterRole, exists := resources.ClusterRoles[otelTargetAllocatorName]
	require.True(t, exists, "target allocator ClusterRole should exist")
	assertClusterRoleHasResource(t, clusterRole, "monitoring.coreos.com", "servicemonitors", "get", "list", "watch")
	assertClusterRoleHasResource(t, clusterRole, "monitoring.coreos.com", "podmonitors", "get", "list", "watch")
	assertClusterRoleHasResource(t, clusterRole, "", "nodes", "get", "list", "watch")
	// configmaps is read on demand for scrape-config templating; upstream
	// grants get only (no list/watch). Lock that scope in.
	assertClusterRoleHasResource(t, clusterRole, "", "configmaps", "get")
	for _, rule := range clusterRole.Rules {
		for _, res := range rule.Resources {
			if res == "configmaps" {
				assert.ElementsMatch(t, []string{"get"}, rule.Verbs,
					"target allocator configmaps verbs should be get-only")
			}
		}
		// secrets are granted per-namespace via Role/RoleBinding, never cluster-wide.
		assert.NotContains(t, rule.Resources, "secrets",
			"target allocator ClusterRole must not grant cluster-wide secrets access; use prometheusCR.secretNamespaces for per-namespace Roles")
		// statefulsets dropped — upstream allocator discovers collectors via the
		// configured pod selector, not by reading the owning workload.
		assert.NotContains(t, rule.APIGroups, "apps",
			"target allocator should not grant access to apps/* resources")
	}
}

func TestOtelPrometheusScrapingCollectorRBAC(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	clusterRole, exists := resources.ClusterRoles[otelMetricsScraperName]
	require.True(t, exists, "metrics scraper ClusterRole should exist")
	assertClusterRoleHasResource(t, clusterRole, "", "pods", "get", "list", "watch")
	assertClusterRoleHasResource(t, clusterRole, "", "namespaces", "get", "list", "watch")

	// Target discovery is delegated to the Target Allocator, so the scraper
	// should never accumulate broader read access (services, endpoints,
	// configmaps, ingresses, etc.). Lock the narrow scope in.
	for _, rule := range clusterRole.Rules {
		for _, res := range rule.Resources {
			assert.Contains(t, []string{"pods", "namespaces"}, res,
				"scraper ClusterRole granted unexpected resource %q; only pods+namespaces are needed for k8sattributes enrichment", res)
		}
		assert.Empty(t, rule.NonResourceURLs,
			"scraper ClusterRole should not grant any nonResourceURLs")
	}
}

func TestOtelPrometheusScrapingGrpcProtocol(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-grpc.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelMetricsScraperConfigName]
	require.True(t, exists, "metrics scraper ConfigMap should exist")
	configData := configMap.Data["config.yaml"]
	assert.Contains(t, configData, "otlp/suse-observability:")
	assert.NotContains(t, configData, "otlphttp/suse-observability:")
	assert.Contains(t, configData, "exporters: [otlp/suse-observability]")

	statefulSet, exists := resources.Statefulsets[otelMetricsScraperName]
	require.True(t, exists, "metrics scraper StatefulSet should exist")
	envVars := envVarsByName(statefulSet.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "otlp-my-suse-observability-instance.com:443", envVars["PLATFORM_OTLP_ENDPOINT"])
}

func TestOtelPrometheusScrapingSkipSslValidation(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-skip-ssl.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	statefulSet, exists := resources.Statefulsets[otelMetricsScraperName]
	require.True(t, exists, "metrics scraper StatefulSet should exist")
	envVars := envVarsByName(statefulSet.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "true", envVars["SKIP_SSL_VALIDATION"])
}

func TestOtelPrometheusScrapingExternalCustomCertificatesConfigMap(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml"},
		SetValues: map[string]string{
			"global.customCertificates.enabled":       "true",
			"global.customCertificates.configMapName": "external-ca-bundle",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	statefulSet, exists := resources.Statefulsets[otelMetricsScraperName]
	require.True(t, exists, "metrics scraper StatefulSet should exist")
	volume := requireVolume(t, statefulSet.Spec.Template.Spec.Volumes, "custom-certificates")
	require.NotNil(t, volume.ConfigMap, "custom-certificates volume should use a ConfigMap")
	assert.Equal(t, "external-ca-bundle", volume.ConfigMap.Name,
		"metrics scraper should mount the user-provided external custom certificate ConfigMap")
}

// See: https://github.com/open-telemetry/opentelemetry-operator/blob/main/cmd/otel-allocator/internal/config/config.go#L440
func TestOtelPrometheusScrapingTargetAllocatorNamespaceFiltersAreMutuallyExclusive(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOpts(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml"},
		SetValues: map[string]string{
			"otel.prometheusScraping.targetAllocator.prometheusCR.allowNamespaces[0]": "team-a",
			"otel.prometheusScraping.targetAllocator.prometheusCR.denyNamespaces[0]":  "team-b",
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "allowNamespaces")
	assert.Contains(t, err.Error(), "denyNamespaces")
}

func TestOtelPrometheusScrapingMtlsAndInsecureAuthSecretsAreMutuallyExclusive(t *testing.T) {
	_, err := helmtestutil.RenderHelmTemplateOptsWithArgs(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/otel-prometheus-scraping-auth-secrets.yaml"},
		SetValues: map[string]string{
			"otel.prometheusScraping.targetAllocator.allowInsecureAuthSecrets": "true",
		},
	},
		"--api-versions", "cert-manager.io/v1/Certificate",
		"--api-versions", "cert-manager.io/v1/Issuer",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mtlsEnabled")
	assert.Contains(t, err.Error(), "allowInsecureAuthSecrets")
}

func TestOtelPrometheusScrapingHardenedSecurityContext(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	statefulSet, exists := resources.Statefulsets[otelMetricsScraperName]
	require.True(t, exists, "metrics scraper StatefulSet should exist")
	assertHardenedPodAndContainerSecurityContext(t, statefulSet.Spec.Template.Spec, statefulSet.Spec.Template.Spec.Containers[0])

	allocator, exists := resources.Deployments[otelTargetAllocatorName]
	require.True(t, exists, "target allocator Deployment should exist")
	assertHardenedPodAndContainerSecurityContext(t, allocator.Spec.Template.Spec, allocator.Spec.Template.Spec.Containers[0])
}

func assertHardenedPodAndContainerSecurityContext(t *testing.T, podSpec corev1.PodSpec, container corev1.Container) {
	t.Helper()

	require.NotNil(t, podSpec.SecurityContext, "pod securityContext should be set")
	require.NotNil(t, podSpec.SecurityContext.RunAsNonRoot)
	assert.True(t, *podSpec.SecurityContext.RunAsNonRoot, "pod must run as non-root")
	require.NotNil(t, podSpec.SecurityContext.SeccompProfile)
	assert.Equal(t, corev1.SeccompProfileTypeRuntimeDefault, podSpec.SecurityContext.SeccompProfile.Type)

	require.NotNil(t, container.SecurityContext, "container securityContext should be set")
	require.NotNil(t, container.SecurityContext.AllowPrivilegeEscalation)
	assert.False(t, *container.SecurityContext.AllowPrivilegeEscalation, "privilege escalation must be disabled")
	require.NotNil(t, container.SecurityContext.ReadOnlyRootFilesystem)
	assert.True(t, *container.SecurityContext.ReadOnlyRootFilesystem, "root filesystem must be read-only")
	require.NotNil(t, container.SecurityContext.Capabilities)
	assert.ElementsMatch(t, []corev1.Capability{"ALL"}, container.SecurityContext.Capabilities.Drop, "all Linux capabilities must be dropped")
	assert.Empty(t, container.SecurityContext.Capabilities.Add, "no capabilities should be added")
}

func TestOtelPrometheusScrapingMultiReplicaRendersPDBs(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-multi-replica.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	scraperPDB, exists := resources.Pdbs[otelMetricsScraperName]
	require.True(t, exists, "metrics scraper PDB should be present when collector.replicaCount > 1")
	require.NotNil(t, scraperPDB.Spec.MaxUnavailable)
	assert.Equal(t, int32(1), scraperPDB.Spec.MaxUnavailable.IntVal,
		"PDB allows one scraper pod down at a time during voluntary disruptions")
	require.NotNil(t, scraperPDB.Spec.Selector)
	assert.Equal(t, "otel-metrics-scraper", scraperPDB.Spec.Selector.MatchLabels["app.kubernetes.io/component"])

	allocatorPDB, exists := resources.Pdbs[otelTargetAllocatorName]
	require.True(t, exists, "target allocator PDB should be present when targetAllocator.replicaCount > 1")
	require.NotNil(t, allocatorPDB.Spec.MaxUnavailable)
	assert.Equal(t, int32(1), allocatorPDB.Spec.MaxUnavailable.IntVal,
		"PDB allows one allocator pod down at a time during voluntary disruptions")
	require.NotNil(t, allocatorPDB.Spec.Selector)
	assert.Equal(t, "otel-target-allocator", allocatorPDB.Spec.Selector.MatchLabels["app.kubernetes.io/component"])
}

func assertOtelPrometheusScrapingResourcesExist(t *testing.T, resources helmtestutil.KubernetesResources, shouldExist bool) {
	_, exists := resources.Statefulsets[otelMetricsScraperName]
	assert.Equal(t, shouldExist, exists, "metrics scraper StatefulSet existence")
	_, exists = resources.Deployments[otelTargetAllocatorName]
	assert.Equal(t, shouldExist, exists, "target allocator Deployment existence")
	_, exists = resources.ConfigMaps[otelMetricsScraperConfigName]
	assert.Equal(t, shouldExist, exists, "metrics scraper ConfigMap existence")
	_, exists = resources.ConfigMaps[otelTargetAllocatorConfigName]
	assert.Equal(t, shouldExist, exists, "target allocator ConfigMap existence")
	_, exists = resources.Services[otelMetricsScraperName]
	assert.Equal(t, shouldExist, exists, "metrics scraper Service existence")
	_, exists = resources.Services[otelTargetAllocatorName]
	assert.Equal(t, shouldExist, exists, "target allocator Service existence")
	_, exists = resources.ServiceAccounts[otelMetricsScraperName]
	assert.Equal(t, shouldExist, exists, "metrics scraper ServiceAccount existence")
	_, exists = resources.ServiceAccounts[otelTargetAllocatorName]
	assert.Equal(t, shouldExist, exists, "target allocator ServiceAccount existence")
	_, exists = resources.ClusterRoles[otelMetricsScraperName]
	assert.Equal(t, shouldExist, exists, "metrics scraper ClusterRole existence")
	_, exists = resources.ClusterRoles[otelTargetAllocatorName]
	assert.Equal(t, shouldExist, exists, "target allocator ClusterRole existence")
	_, exists = resources.ClusterRoleBindings[otelMetricsScraperName]
	assert.Equal(t, shouldExist, exists, "metrics scraper ClusterRoleBinding existence")
	_, exists = resources.ClusterRoleBindings[otelTargetAllocatorName]
	assert.Equal(t, shouldExist, exists, "target allocator ClusterRoleBinding existence")
}

func assertClusterRoleHasResource(t *testing.T, role rbacv1.ClusterRole, apiGroup string, resource string, verbs ...string) {
	for _, rule := range role.Rules {
		if slices.Contains(rule.APIGroups, apiGroup) && slices.Contains(rule.Resources, resource) {
			for _, verb := range verbs {
				assert.Contains(t, rule.Verbs, verb)
			}
			return
		}
	}
	t.Fatalf("expected RBAC rule apiGroup=%q resource=%q verbs=%v", apiGroup, resource, verbs)
}

func envVarsByName(envVars []corev1.EnvVar) map[string]string {
	result := make(map[string]string)
	for _, env := range envVars {
		if env.Value != "" {
			result[env.Name] = env.Value
		}
	}
	return result
}

func assertEnvFromFieldRef(t *testing.T, envVars []corev1.EnvVar, name string, fieldPath string) {
	for _, env := range envVars {
		if env.Name == name {
			require.NotNil(t, env.ValueFrom)
			require.NotNil(t, env.ValueFrom.FieldRef)
			assert.Equal(t, fieldPath, env.ValueFrom.FieldRef.FieldPath)
			return
		}
	}
	t.Fatalf("expected env var %s", name)
}

func requireVolume(t *testing.T, volumes []corev1.Volume, name string) corev1.Volume {
	for _, volume := range volumes {
		if volume.Name == name {
			return volume
		}
	}
	t.Fatalf("expected volume %s", name)
	return corev1.Volume{}
}

func TestOtelPrometheusScrapingDebugExporterEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-debug.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelMetricsScraperConfigName]
	require.True(t, exists, "metrics scraper ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	// Lock indent: debug must be a sibling of otlphttp/suse-observability under
	// exporters — otherwise the collector rejects the config as invalid YAML.
	// (The ConfigMap parser strips the block-scalar indent, so the sibling
	// exporters live at 2 spaces here.)
	assert.Contains(t, configData, "\n  debug:\n    verbosity: detailed\n")
	assert.Contains(t, configData, "exporters: [otlphttp/suse-observability, debug]")
}

func TestOtelPrometheusScrapingDebugExporterDisabledByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/otel-prometheus-scraping-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelMetricsScraperConfigName]
	require.True(t, exists, "metrics scraper ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	assert.NotContains(t, configData, "debug:")
	assert.NotContains(t, configData, ", debug]")
}

func renderOtelPrometheusScrapingWithPrometheusCRDs(t *testing.T, valuesFiles ...string) string {
	allValuesFiles := append([]string{"values/minimal.yaml"}, valuesFiles...)
	return helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: allValuesFiles,
	},
		"--api-versions", "monitoring.coreos.com/v1/ServiceMonitor",
		"--api-versions", "monitoring.coreos.com/v1/PodMonitor",
	)
}

func renderOtelPrometheusScrapingWithCertManager(t *testing.T, valuesFiles ...string) string {
	allValuesFiles := append([]string{"values/minimal.yaml"}, valuesFiles...)
	return helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: allValuesFiles,
	},
		"--api-versions", "cert-manager.io/v1/Certificate",
		"--api-versions", "cert-manager.io/v1/Issuer",
	)
}
