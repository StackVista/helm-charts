package test

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

const (
	otelTelemetryGatewayName       = "suse-observability-agent-otel-telemetry-gateway"
	otelTelemetryGatewayConfigName = "suse-observability-agent-otel-telemetry-gateway-config"
)

func TestOtelTelemetryGatewayDisabledByDefault(t *testing.T) {
	// otel.enabled defaults to false so no gateway resources should render.
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assertOtelTelemetryGatewayResourcesExist(t, resources, false)
	_, exists := resources.Pdbs[otelTelemetryGatewayName]
	assert.False(t, exists, "gateway PDB should not exist when disabled")
}

func TestOtelTelemetryGatewayDisabledWhenFlagFalse(t *testing.T) {
	// otel.enabled=true but otel.telemetryGateway.enabled=false (default) — no gateway resources.
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml"},
		SetValues: map[string]string{
			"otel.enabled":                  "true",
			"otel.telemetryGateway.enabled": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)
	assertOtelTelemetryGatewayResourcesExist(t, resources, false)
}

func TestOtelMarkerCrdSkippedWhenNoOtelIntegrationPathActive(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml"},
		SetValues: map[string]string{
			"otel.enabled":                    "true",
			"otel.telemetryGateway.enabled":   "false",
			"otel.prometheusScraping.enabled": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	_, exists := resources.Unmapped["suseobservabilityagents.observability.suse.com"]
	assert.False(t, exists, "marker CRD should not render when otel.enabled=true but all product-facing OTel integration paths are disabled")
}

func TestOtelMarkerCrdRenderedWhenTelemetryGatewayEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	markerCrd, exists := resources.Unmapped["suseobservabilityagents.observability.suse.com"]
	require.True(t, exists, "marker CRD should render when telemetry gateway is enabled")
	assert.Contains(t, markerCrd, "kind: CustomResourceDefinition")
	assert.Contains(t, markerCrd, "group: observability.suse.com")
	assert.Contains(t, markerCrd, "plural: suseobservabilityagents")
	assert.Contains(t, markerCrd, "capability signal only, not endpoint")
}

func TestOtelTelemetryGatewayRendersResources(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assertOtelTelemetryGatewayResourcesExist(t, resources, true)

	// No PDB at default replicaCount=1.
	_, exists := resources.Pdbs[otelTelemetryGatewayName]
	assert.False(t, exists, "gateway PDB should be skipped at default replicaCount=1")

	deployment, exists := resources.Deployments[otelTelemetryGatewayName]
	require.True(t, exists, "gateway Deployment should exist")
	require.Len(t, deployment.Spec.Template.Spec.Containers, 1)
	container := deployment.Spec.Template.Spec.Containers[0]
	assert.Contains(t, container.Image, "stackstate/sts-opentelemetry-collector")
	assert.Equal(t, int32(1), *deployment.Spec.Replicas)
}

func TestOtelTelemetryGatewayDeploymentEnvVars(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, exists := resources.Deployments[otelTelemetryGatewayName]
	require.True(t, exists, "gateway Deployment should exist")
	container := deployment.Spec.Template.Spec.Containers[0]

	envVars := envVarsByName(container.Env)
	assert.Equal(t, "https://my-suse-observability-instance.com/receiver/otel", envVars["PLATFORM_OTLP_ENDPOINT"])
	assert.Equal(t, "some-k8s-cluster", envVars["K8S_CLUSTER_NAME"])
	assert.Equal(t, "false", envVars["SKIP_SSL_VALIDATION"])
	assertEnvFromFieldRef(t, container.Env, "POD_NAME", "metadata.name")
	assertEnvFromFieldRef(t, container.Env, "POD_NAMESPACE", "metadata.namespace")
	assertEnvFromFieldRef(t, container.Env, "K8S_NAMESPACE_NAME", "metadata.namespace")
	assertEnvFromFieldRef(t, container.Env, "POD_IP", "status.podIP")
	assertEnvFromFieldRef(t, container.Env, "POD_UID", "metadata.uid")
}

func TestOtelTelemetryGatewayHardenedSecurityContext(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, exists := resources.Deployments[otelTelemetryGatewayName]
	require.True(t, exists, "gateway Deployment should exist")
	assertHardenedPodAndContainerSecurityContext(t, deployment.Spec.Template.Spec, deployment.Spec.Template.Spec.Containers[0])
}

func TestOtelTelemetryGatewayServicePorts(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	svc, exists := resources.Services[otelTelemetryGatewayName]
	require.True(t, exists, "gateway Service should exist")

	portsByName := make(map[string]int32)
	for _, p := range svc.Spec.Ports {
		portsByName[p.Name] = p.Port
	}
	assert.Equal(t, int32(4317), portsByName["otlp-grpc"], "OTLP gRPC port should be 4317")
	assert.Equal(t, int32(4318), portsByName["otlp-http"], "OTLP HTTP port should be 4318")
	assert.Equal(t, int32(13133), portsByName["health"], "health port should be 13133")
	assert.Equal(t, int32(8888), portsByName["metrics"], "metrics port should be 8888")
}

func TestOtelTelemetryGatewayCollectorConfig(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTelemetryGatewayConfigName]
	require.True(t, exists, "gateway ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	// Receivers.
	assert.Contains(t, configData, "otlp:")
	assert.Contains(t, configData, "grpc:")
	assert.Contains(t, configData, "endpoint: 0.0.0.0:4317")
	assert.Contains(t, configData, "http:")
	assert.Contains(t, configData, "endpoint: 0.0.0.0:4318")
	assert.Contains(t, configData, "prometheus/self:")
	assert.Contains(t, configData, "job_name: opentelemetry-collector")
	assert.Contains(t, configData, "${env:POD_IP}:8888")
	assert.Contains(t, configData, "k8s.pod.uid: ${env:POD_UID}")

	// Processors.
	assert.Contains(t, configData, "memory_limiter:")
	assert.Contains(t, configData, "limit_percentage: 80")
	assert.Contains(t, configData, "spike_limit_percentage: 25")
	assert.Contains(t, configData, "transform/pre-k8sattributes:")
	assert.Contains(t, configData, "k8s_attributes:")
	assert.Contains(t, configData, "key_regex: (.*)")
	assert.Contains(t, configData, "tag_name: $$1")
	assert.Contains(t, configData, "k8s.namespace.name")
	assert.Contains(t, configData, "k8s.pod.uid")
	assert.Contains(t, configData, "from: connection")
	assert.Contains(t, configData, "transform/enrich-resource:")
	assert.Contains(t, configData, "filter/dropMissingK8sAttributes:")
	assert.Contains(t, configData, "tail_sampling:")
	assert.Contains(t, configData, "max_total_spans_per_second: 500")
	assert.Contains(t, configData, "transform/self-metrics:")
	assert.NotContains(t, configData, "stsusage:")
	assert.Contains(t, configData, "batch: {}")

	// span_metrics connector (new snake_case name).
	assert.Contains(t, configData, "routing/traces:")
	assert.Contains(t, configData, "pipelines: [traces/sampling, traces/spanmetrics]")
	assert.Contains(t, configData, "span_metrics:")
	assert.NotContains(t, configData, "\n  spanmetrics:", "legacy connector name should not appear")
	assert.Contains(t, configData, "aggregation_cardinality_limit: 5000")
	assert.Contains(t, configData, "resource_metrics_key_attributes:")
	assert.Contains(t, configData, "- service.name")
	assert.NotContains(t, configData, "- service.instance.id", "span metrics should aggregate at service level, not pod instance level")
	assert.Contains(t, configData, "- k8s.cluster.name")
	assert.Contains(t, configData, "unit: s")
	assert.Contains(t, configData, "metrics_expiration: 5m")
	assert.Contains(t, configData, "namespace: otel_span")

	// Exporter and extensions.
	assert.Contains(t, configData, "otlp_http/suse-observability:")
	assert.Contains(t, configData, "endpoint: ${env:PLATFORM_OTLP_ENDPOINT}")
	assert.Contains(t, configData, "insecure_skip_verify: ${env:SKIP_SSL_VALIDATION}")
	assert.Contains(t, configData, "bearertokenauth:")
	assert.Contains(t, configData, "scheme: SUSEObservability")
	assert.Contains(t, configData, "token: \"${env:STS_API_KEY}\"")
	assert.Contains(t, configData, "pprof:")
	assert.Contains(t, configData, "endpoint: 0.0.0.0:1777")
	assert.Contains(t, configData, "extensions: [health_check, pprof, bearertokenauth]")

	// Customer-data pipelines plus dedicated self-metrics pipeline.
	assert.Contains(t, configData, "traces/gateway:")
	assert.Contains(t, configData, "metrics/gateway:")
	assert.Contains(t, configData, "metrics/spanmetrics:")
	assert.Contains(t, configData, "metrics/self:")
	assert.Contains(t, configData, "logs/gateway:")
	assert.NotContains(t, configData, "sts_api_key")
}

func TestOtelTelemetryGatewayPprofCanBeDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml"},
		SetValues: map[string]string{
			"otel.telemetryGateway.pprof.enabled": "false",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTelemetryGatewayConfigName]
	require.True(t, exists, "gateway ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	assert.NotContains(t, configData, "pprof:", "pprof extension must be omitted when pprof.enabled=false")
	assert.NotContains(t, configData, "endpoint: 0.0.0.0:1777")
	assert.Contains(t, configData, "extensions: [health_check, bearertokenauth]",
		"service.extensions must drop pprof when pprof.enabled=false")
}

func TestOtelTelemetryGatewaySpanMetricsCardinalityLimitOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml"},
		SetValues: map[string]string{
			"otel.telemetryGateway.spanMetrics.aggregationCardinalityLimit": "2000",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTelemetryGatewayConfigName]
	require.True(t, exists, "gateway ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	assert.Contains(t, configData, "aggregation_cardinality_limit: 2000")
	assert.NotContains(t, configData, "aggregation_cardinality_limit: 5000")
}

func TestOtelTelemetryGatewaySecurityEnforcement(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTelemetryGatewayConfigName]
	require.True(t, exists, "gateway ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	// Pre-k8sattributes: both attributes deleted for all three signals.
	assert.Contains(t, configData, "delete_key(resource.attributes, \"k8s.cluster.name\")")
	assert.Contains(t, configData, "delete_key(resource.attributes, \"k8s.namespace.name\")")
	assert.Contains(t, configData, "trace_statements:")
	assert.Contains(t, configData, "metric_statements:")
	assert.Contains(t, configData, "log_statements:")

	// Post-k8sattributes transform: cluster name overwritten authoritatively.
	assert.Contains(t, configData, "set(resource.attributes[\"k8s.cluster.name\"], \"${env:K8S_CLUSTER_NAME}\")")
	assert.Contains(t, configData, "set(resource.attributes[\"service.instance.id\"], resource.attributes[\"k8s.pod.uid\"])")
	assert.Contains(t, configData, "set(resource.attributes[\"service.namespace\"], resource.attributes[\"k8s.namespace.name\"])")
}

func TestOtelTelemetryGatewayAllThreeSignalPipelines(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTelemetryGatewayConfigName]
	require.True(t, exists, "gateway ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	// traces/gateway: enriches and filters once before routing fan-out.
	assert.Contains(t, configData, "exporters: [routing/traces]")
	assert.Contains(t, configData,
		"processors: [memory_limiter, transform/pre-k8sattributes, k8s_attributes, transform/enrich-resource, filter/dropMissingK8sAttributes]")

	// traces/spanmetrics: consumes already-enriched traces and generates span metrics.
	assert.Contains(t, configData, "traces/spanmetrics:")
	assert.Contains(t, configData, "receivers: [routing/traces]")
	assert.Contains(t, configData, "exporters: [span_metrics]")

	// traces/sampling: applies the shared collector sampling policy before platform export.
	assert.Contains(t, configData, "traces/sampling:")
	assert.Contains(t, configData, "receivers: [routing/traces]")
	assert.Contains(t, configData, "processors: [tail_sampling, batch]")
	assert.Contains(t, configData, "exporters: [otlp_http/suse-observability]")

	// metrics/gateway: customer OTLP only; prometheus/self and span metrics are separate.
	assert.Contains(t, configData, "metrics/gateway:")
	assert.Contains(t, configData, "receivers: [otlp]")
	assert.NotContains(t, configData, "receivers: [otlp, span_metrics]")
	assert.NotContains(t, configData, "receivers: [otlp, prometheus/self]")

	// metrics/spanmetrics: span metrics already inherit enriched trace resource attributes.
	assert.Contains(t, configData, "metrics/spanmetrics:")
	assert.Contains(t, configData, "receivers: [span_metrics]")
	assert.Contains(t, configData, "processors: [memory_limiter, batch]")

	// metrics/self: dedicated pipeline for collector self-metrics; skips security enforcement.
	assert.Contains(t, configData, "metrics/self:")
	assert.Contains(t, configData, "receivers: [prometheus/self]")
	assert.Contains(t, configData, "processors: [memory_limiter, transform/self-metrics, batch]")
	assert.Contains(t, configData, "set(resource.attributes[\"k8s.namespace.name\"], \"${env:K8S_NAMESPACE_NAME}\")")

	// logs/gateway.
	assert.Contains(t, configData, "logs/gateway:")
}

func TestOtelTelemetryGatewayMinimalRbac(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	clusterRole, exists := resources.ClusterRoles[otelTelemetryGatewayName]
	require.True(t, exists, "gateway ClusterRole should exist")
	assertClusterRoleHasResource(t, clusterRole, "", "pods", "get", "list", "watch")
	assertClusterRoleHasResource(t, clusterRole, "", "namespaces", "get", "list", "watch")

	for _, rule := range clusterRole.Rules {
		for _, res := range rule.Resources {
			assert.Contains(t, []string{"pods", "namespaces"}, res,
				"gateway ClusterRole granted unexpected resource %q; only pods+namespaces are needed for k8sattributes enrichment", res)
		}
		assert.Empty(t, rule.NonResourceURLs, "gateway ClusterRole should not grant any nonResourceURLs")
	}
}

func TestOtelTelemetryGatewayGrpcExporter(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-grpc.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTelemetryGatewayConfigName]
	require.True(t, exists, "gateway ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	assert.Contains(t, configData, "otlp/suse-observability:")
	assert.NotContains(t, configData, "otlp_http/suse-observability:")
	assert.Contains(t, configData, "exporters: [otlp/suse-observability]")
	assert.Contains(t, configData, "exporters: [routing/traces]")

	deployment, exists := resources.Deployments[otelTelemetryGatewayName]
	require.True(t, exists, "gateway Deployment should exist")
	envVars := envVarsByName(deployment.Spec.Template.Spec.Containers[0].Env)
	assert.Equal(t, "otlp-my-suse-observability-instance.com:443", envVars["PLATFORM_OTLP_ENDPOINT"])
}

func TestOtelTelemetryGatewayDebugExporter(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-debug.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTelemetryGatewayConfigName]
	require.True(t, exists, "gateway ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	assert.Contains(t, configData, "\n  debug:\n    verbosity: detailed\n")
	assert.Contains(t, configData, "exporters: [otlp_http/suse-observability, debug]")
	assert.Contains(t, configData, "exporters: [otlp_http/suse-observability, debug]")
}

func TestOtelTelemetryGatewayDebugExporterDisabledByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTelemetryGatewayConfigName]
	require.True(t, exists, "gateway ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	assert.NotContains(t, configData, "debug:")
	assert.NotContains(t, configData, ", debug]")
}

func TestOtelTelemetryGatewayPdbRenderedAtMultipleReplicas(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-multi-replica.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	pdb, exists := resources.Pdbs[otelTelemetryGatewayName]
	require.True(t, exists, "gateway PDB should be present when replicaCount > 1")
	require.NotNil(t, pdb.Spec.MaxUnavailable)
	assert.Equal(t, int32(1), pdb.Spec.MaxUnavailable.IntVal,
		"PDB allows one gateway pod down at a time during voluntary disruptions")
	require.NotNil(t, pdb.Spec.Selector)
	assert.Equal(t, "otel-telemetry-gateway", pdb.Spec.Selector.MatchLabels["app.kubernetes.io/component"])
}

func TestOtelTelemetryGatewayCustomCertificates(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: []string{"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml"},
		SetValues: map[string]string{
			"global.customCertificates.enabled":       "true",
			"global.customCertificates.configMapName": "external-ca-bundle",
		},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, exists := resources.Deployments[otelTelemetryGatewayName]
	require.True(t, exists, "gateway Deployment should exist")
	volume := requireVolume(t, deployment.Spec.Template.Spec.Volumes, "custom-certificates")
	require.NotNil(t, volume.ConfigMap, "custom-certificates volume should use a ConfigMap")
	assert.Equal(t, "external-ca-bundle", volume.ConfigMap.Name,
		"gateway should mount the user-provided external custom certificate ConfigMap")
}

func TestOtelTelemetryGatewayConfigChecksumAnnotation(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, exists := resources.Deployments[otelTelemetryGatewayName]
	require.True(t, exists, "gateway Deployment should exist")
	annotations := deployment.Spec.Template.Annotations
	checksumVal, hasChecksum := annotations["checksum/configmap"]
	assert.True(t, hasChecksum, "pod template should carry a checksum/configmap annotation for auto-rollout on config change")
	assert.NotEmpty(t, checksumVal, "checksum/configmap annotation must not be empty")
}

func assertOtelTelemetryGatewayResourcesExist(t *testing.T, resources helmtestutil.KubernetesResources, shouldExist bool) {
	t.Helper()
	_, exists := resources.Deployments[otelTelemetryGatewayName]
	assert.Equal(t, shouldExist, exists, "gateway Deployment existence")
	_, exists = resources.ConfigMaps[otelTelemetryGatewayConfigName]
	assert.Equal(t, shouldExist, exists, "gateway ConfigMap existence")
	_, exists = resources.Services[otelTelemetryGatewayName]
	assert.Equal(t, shouldExist, exists, "gateway Service existence")
	_, exists = resources.ServiceAccounts[otelTelemetryGatewayName]
	assert.Equal(t, shouldExist, exists, "gateway ServiceAccount existence")
	_, exists = resources.ClusterRoles[otelTelemetryGatewayName]
	assert.Equal(t, shouldExist, exists, "gateway ClusterRole existence")
	_, exists = resources.ClusterRoleBindings[otelTelemetryGatewayName]
	assert.Equal(t, shouldExist, exists, "gateway ClusterRoleBinding existence")
}

func TestOtelTelemetryGatewayDoesNotUseStsUsage(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps[otelTelemetryGatewayConfigName]
	require.True(t, exists, "gateway ConfigMap should exist")
	configData := configMap.Data["config.yaml"]

	assert.NotContains(t, configData, "stsusage:")
	assert.Equal(t, 0, strings.Count(configData, "stsusage"), "usage accounting is intentionally deferred")
}

func TestOtelTelemetryGatewayDeploymentStrategy(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent",
		"values/minimal.yaml", "values/otel-telemetry-gateway-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, exists := resources.Deployments[otelTelemetryGatewayName]
	require.True(t, exists, "gateway Deployment should exist")
	assert.Equal(t, "RollingUpdate", string(deployment.Spec.Strategy.Type))
}
