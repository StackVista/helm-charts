package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestK8sResourceCollectorEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Deployment should exist
	deployment, exists := resources.Deployments["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector deployment was not found")

	// ServiceAccount should exist
	_, exists = resources.ServiceAccounts["suse-observability-agent-k8s-resource-collector"]
	assert.True(t, exists, "k8s-resource-collector service account was not found")

	// ClusterRole should exist
	_, exists = resources.ClusterRoles["suse-observability-agent-k8s-resource-collector"]
	assert.True(t, exists, "k8s-resource-collector cluster role was not found")

	// ClusterRoleBinding should exist
	_, exists = resources.ClusterRoleBindings["suse-observability-agent-k8s-resource-collector"]
	assert.True(t, exists, "k8s-resource-collector cluster role binding was not found")

	// ConfigMap should exist
	configMap, exists := resources.ConfigMaps["suse-observability-agent-k8s-resource-collector-config"]
	require.True(t, exists, "k8s-resource-collector config map was not found")

	// Service should exist
	service, exists := resources.Services["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector service was not found")

	// Headless service for peer sync should exist (leaderElection defaults to enabled)
	headlessSvc, exists := resources.Services["suse-observability-agent-k8s-resource-collector-headless"]
	require.True(t, exists, "k8s-resource-collector headless service was not found")
	assert.Equal(t, "None", string(headlessSvc.Spec.ClusterIP))
	assert.Equal(t, "peer-sync", headlessSvc.Spec.Ports[0].Name)

	// Verify deployment has correct image
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Image, "sts-opentelemetry-collector")

	// Verify POD_IP env var is set (required for peer sync self-exclusion)
	container := deployment.Spec.Template.Spec.Containers[0]
	var hasPodIP bool
	for _, env := range container.Env {
		if env.Name == "POD_IP" && env.ValueFrom != nil && env.ValueFrom.FieldRef != nil {
			hasPodIP = true
			assert.Equal(t, "status.podIP", env.ValueFrom.FieldRef.FieldPath)
		}
	}
	assert.True(t, hasPodIP, "POD_IP env var from downward API not found")

	// Verify metrics port is exposed for Prometheus scraping
	var hasMetricsPort bool
	for _, port := range container.Ports {
		if port.Name == "metrics" {
			hasMetricsPort = true
			assert.Equal(t, int32(8888), port.ContainerPort)
		}
	}
	assert.True(t, hasMetricsPort, "metrics port (8888) not found in container ports")

	// Verify autodiscovery annotations for StackState Agent metrics scraping
	annotations := deployment.Spec.Template.Annotations
	assert.Contains(t, annotations, "ad.stackstate.com/k8s-resource-collector.check_names")
	assert.Contains(t, annotations["ad.stackstate.com/k8s-resource-collector.instances"], "8888/metrics")

	// Verify service port name
	assert.Equal(t, "health", service.Spec.Ports[0].Name)

	// Verify ConfigMap contains k8sresource receiver configuration
	assert.Contains(t, configMap.Data, "config.yaml")
	configData := configMap.Data["config.yaml"]
	assert.Contains(t, configData, "k8sresource:")
	assert.Contains(t, configData, "snapshot_interval: 5m")
	assert.Contains(t, configData, "discovery_mode: api_groups")

	// Verify peer sync configuration
	assert.Contains(t, configData, "peer_sync_port: 4319")
	assert.Contains(t, configData, "peer_sync_dns:")
}

func TestK8sResourceCollectorDisabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-disabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// All k8s-resource-collector resources should NOT exist
	_, exists := resources.Deployments["suse-observability-agent-k8s-resource-collector"]
	assert.False(t, exists, "k8s-resource-collector deployment should not exist when disabled")

	_, exists = resources.ServiceAccounts["suse-observability-agent-k8s-resource-collector"]
	assert.False(t, exists, "k8s-resource-collector service account should not exist when disabled")

	_, exists = resources.ClusterRoles["suse-observability-agent-k8s-resource-collector"]
	assert.False(t, exists, "k8s-resource-collector cluster role should not exist when disabled")

	_, exists = resources.ConfigMaps["suse-observability-agent-k8s-resource-collector-config"]
	assert.False(t, exists, "k8s-resource-collector config map should not exist when disabled")

	_, exists = resources.Services["suse-observability-agent-k8s-resource-collector"]
	assert.False(t, exists, "k8s-resource-collector service should not exist when disabled")

	_, exists = resources.Services["suse-observability-agent-k8s-resource-collector-headless"]
	assert.False(t, exists, "k8s-resource-collector headless service should not exist when disabled")
}

func TestK8sResourceCollectorWildcardRBAC(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	clusterRole, exists := resources.ClusterRoles["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector cluster role was not found")

	// Should have wildcard permissions for custom resources
	var hasWildcardRule bool
	for _, rule := range clusterRole.Rules {
		for _, apiGroup := range rule.APIGroups {
			if apiGroup == "*" {
				hasWildcardRule = true
				// Verify verbs
				assert.Contains(t, rule.Verbs, "get")
				assert.Contains(t, rule.Verbs, "list")
				assert.Contains(t, rule.Verbs, "watch")
				break
			}
		}
	}
	assert.True(t, hasWildcardRule, "wildcard RBAC rule not found in cluster role")

	// Should also have specific CRD permissions
	var hasCRDRule bool
	for _, rule := range clusterRole.Rules {
		for _, apiGroup := range rule.APIGroups {
			if apiGroup == "apiextensions.k8s.io" {
				hasCRDRule = true
				assert.Contains(t, rule.Resources, "customresourcedefinitions")
				break
			}
		}
	}
	assert.True(t, hasCRDRule, "CRD rule not found in cluster role")
}

func TestK8sResourceCollectorRestrictedRBAC(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-restricted-rbac.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	clusterRole, exists := resources.ClusterRoles["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector cluster role was not found")

	// Should NOT have wildcard permissions
	for _, rule := range clusterRole.Rules {
		for _, apiGroup := range rule.APIGroups {
			assert.NotEqual(t, "*", apiGroup, "should not have wildcard RBAC when restricted mode is enabled")
		}
	}

	// Should have specific API group permissions
	expectedAPIGroups := []string{"policies.kubewarden.io", "longhorn.io", "cert-manager.io"}
	foundAPIGroups := make(map[string]bool)

	for _, rule := range clusterRole.Rules {
		for _, apiGroup := range rule.APIGroups {
			for _, expected := range expectedAPIGroups {
				if apiGroup == expected {
					foundAPIGroups[expected] = true
				}
			}
		}
	}

	for _, expected := range expectedAPIGroups {
		assert.True(t, foundAPIGroups[expected], "expected API group %s not found in cluster role", expected)
	}

	// Should still have CRD permissions
	var hasCRDRule bool
	for _, rule := range clusterRole.Rules {
		for _, apiGroup := range rule.APIGroups {
			if apiGroup == "apiextensions.k8s.io" {
				hasCRDRule = true
				break
			}
		}
	}
	assert.True(t, hasCRDRule, "CRD rule not found in cluster role")
}

// TestK8sResourceCollectorObjects verifies that k8sResourceCollector.objects and
// deniedObjects are rendered into the ConfigMap (with snake_case keys), and that
// auto-derived RBAC rules in useWildcard=false mode group by API group and dedupe
// resource names (uniq) when the same resource is declared multiple times.
func TestK8sResourceCollectorObjects(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-objects.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// ---- ConfigMap: objects + denied_objects render ----
	configMap, exists := resources.ConfigMaps["suse-observability-agent-k8s-resource-collector-config"]
	require.True(t, exists, "k8s-resource-collector config map was not found")
	configData := configMap.Data["config.yaml"]

	assert.Contains(t, configData, "objects:")
	assert.Contains(t, configData, "denied_objects:")
	// Snake-case keys produced from camelCase values.
	assert.Contains(t, configData, "label_selector:")
	assert.Contains(t, configData, "field_selector:")
	assert.Contains(t, configData, "namespaces:")
	// Sampled resource names from the override.
	assert.Contains(t, configData, "pods")
	assert.Contains(t, configData, "deployments")
	assert.Contains(t, configData, "statefulsets")
	assert.Contains(t, configData, "certificates")

	// ---- ClusterRole: auto-derived rules ----
	clusterRole, exists := resources.ClusterRoles["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector cluster role was not found")

	for _, rule := range clusterRole.Rules {
		for _, apiGroup := range rule.APIGroups {
			assert.NotEqual(t, "*", apiGroup, "should not have wildcard RBAC when restricted mode is enabled")
		}
	}

	// Index rules by API group → set of resource names.
	rulesByGroup := map[string]map[string]int{}
	for _, rule := range clusterRole.Rules {
		for _, g := range rule.APIGroups {
			if _, ok := rulesByGroup[g]; !ok {
				rulesByGroup[g] = map[string]int{}
			}
			for _, r := range rule.Resources {
				rulesByGroup[g][r]++
			}
		}
	}

	// Core group rule: must include pods.
	coreResources, hasCore := rulesByGroup[""]
	require.True(t, hasCore, "expected an auto-derived rule for the core API group")
	assert.Contains(t, coreResources, "pods", "pods should appear under core group")

	// apps group rule: must include deployments AND statefulsets — proves
	// multiple object entries with the same group end up in the same rule.
	appsResources, hasApps := rulesByGroup["apps"]
	require.True(t, hasApps, "expected an auto-derived rule for the apps API group")
	assert.Contains(t, appsResources, "deployments", "deployments should appear under apps group")
	assert.Contains(t, appsResources, "statefulsets", "statefulsets should appear under apps group")
}

// TestK8sResourceCollectorObjectsMergeAcrossValues proves that the dict-shaped
// objects / crdApiGroups values merge per-key when an operator layers extra
// values via `-f extra.yaml`. With list-shaped values the second file would
// override the first entirely; with dict-shaped values both sets coexist.
func TestK8sResourceCollectorObjectsMergeAcrossValues(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(
		t,
		"suse-observability-agent",
		"values/minimal.yaml",
		"values/k8s-resource-collector-objects.yaml",
		"values/k8s-resource-collector-objects-extra.yaml",
	)
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps["suse-observability-agent-k8s-resource-collector-config"]
	require.True(t, exists, "k8s-resource-collector config map was not found")
	configData := configMap.Data["config.yaml"]

	// Base file's objects.
	assert.Contains(t, configData, "name: \"pods\"", "base 'pods' object should survive override")
	assert.Contains(t, configData, "name: \"deployments\"", "base 'deployments' object should survive override")
	assert.Contains(t, configData, "name: \"statefulsets\"", "base 'statefulsets' object should survive override")
	// Extra file's object — only present if dict merge worked.
	assert.Contains(t, configData, "name: \"services\"", "extra 'services' object proves per-key merge across -f files")

	// ClusterRole: both base group and extra group rules must be present.
	clusterRole, exists := resources.ClusterRoles["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector cluster role was not found")

	apiGroups := map[string]bool{}
	for _, rule := range clusterRole.Rules {
		for _, g := range rule.APIGroups {
			apiGroups[g] = true
		}
	}
	assert.True(t, apiGroups["policies.kubewarden.io"], "base crdApiGroups entry should survive override")
	assert.True(t, apiGroups["longhorn.io"], "extra crdApiGroups entry proves per-key merge across -f files")
}

// TestK8sResourceCollectorApiGroupDisableViaOverride proves that setting a
// default dict key to `false` in an override values file removes the entry
// from the rendered output (the disable semantics for dict-shaped values).
func TestK8sResourceCollectorApiGroupDisableViaOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(
		t,
		"suse-observability-agent",
		"values/minimal.yaml",
		"values/k8s-resource-collector-restricted-rbac.yaml",
		"values/k8s-resource-collector-disable-override.yaml",
	)
	resources := helmtestutil.NewKubernetesResources(t, output)

	clusterRole, exists := resources.ClusterRoles["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector cluster role was not found")

	apiGroups := map[string]bool{}
	for _, rule := range clusterRole.Rules {
		for _, g := range rule.APIGroups {
			apiGroups[g] = true
		}
	}
	// Disabled by override — must NOT appear.
	assert.False(t, apiGroups["longhorn.io"], "longhorn.io was set to false in override and should not render")
	// Other base entries must still be present.
	assert.True(t, apiGroups["policies.kubewarden.io"], "policies.kubewarden.io should still render")
	assert.True(t, apiGroups["cert-manager.io"], "cert-manager.io should still render")
}

func TestK8sResourceCollectorDiscoveryModeAll(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-all-mode.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps["suse-observability-agent-k8s-resource-collector-config"]
	require.True(t, exists, "k8s-resource-collector config map was not found")

	configData := configMap.Data["config.yaml"]
	assert.Contains(t, configData, "discovery_mode: all")
	// When mode is "all", crd_api_group_filters should not be present
	assert.NotContains(t, configData, "crd_api_group_filters:")
}

func TestK8sResourceCollectorConfigMapContent(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-restricted-rbac.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps["suse-observability-agent-k8s-resource-collector-config"]
	require.True(t, exists, "k8s-resource-collector config map was not found")

	configData := configMap.Data["config.yaml"]

	// Verify receiver configuration
	assert.Contains(t, configData, "receivers:")
	assert.Contains(t, configData, "k8sresource:")
	assert.Contains(t, configData, "auth_type: serviceAccount")
	assert.Contains(t, configData, "snapshot_interval: 5m")
	assert.Contains(t, configData, "discovery_mode: api_groups")

	// Verify API group filters
	assert.Contains(t, configData, "crd_api_group_filters:")
	assert.Contains(t, configData, "include:")
	assert.Contains(t, configData, "policies.kubewarden.io")
	assert.Contains(t, configData, "longhorn.io")
	assert.Contains(t, configData, "exclude:")
	assert.Contains(t, configData, "test.suse.com")

	// Verify exporters
	assert.Contains(t, configData, "exporters:")
	assert.Contains(t, configData, "otlp_http/suse-observability:")
	assert.Contains(t, configData, "endpoint: ${PLATFORM_OTLP_ENDPOINT}")
	assert.Contains(t, configData, "authenticator: bearertokenauth")

	// Verify extensions
	assert.Contains(t, configData, "extensions:")
	assert.Contains(t, configData, "health_check:")
	assert.Contains(t, configData, "bearertokenauth:")
	assert.Contains(t, configData, "scheme: SUSEObservability")

	// Verify service pipeline
	assert.Contains(t, configData, "service:")
	assert.Contains(t, configData, "pipelines:")
	assert.Contains(t, configData, "logs/k8s-resource:")
}

func TestK8sResourceCollectorPlatformOtlpEndpointDerived(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, exists := resources.Deployments["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector deployment was not found")

	container := deployment.Spec.Template.Spec.Containers[0]

	// Collect environment variables
	envVars := make(map[string]string)
	for _, env := range container.Env {
		if env.Value != "" {
			envVars[env.Name] = env.Value
		}
	}

	// Verify PLATFORM_OTLP_ENDPOINT is derived from stackstate.url by appending /otel.
	// stackstate.url: https://my-suse-observability-instance.com/stsAgent
	// Expected derivation: https://my-suse-observability-instance.com/stsAgent/otel
	// (matches the platform router's /stsAgent/otel/ route → otel-collector.)
	assert.Equal(t, "https://my-suse-observability-instance.com/receiver/stsAgent/otel", envVars["PLATFORM_OTLP_ENDPOINT"])
}

func TestK8sResourceCollectorPlatformOtlpEndpointExplicitOverride(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-explicit-otlp.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, exists := resources.Deployments["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector deployment was not found")

	container := deployment.Spec.Template.Spec.Containers[0]

	// Collect environment variables
	envVars := make(map[string]string)
	for _, env := range container.Env {
		if env.Value != "" {
			envVars[env.Name] = env.Value
		}
	}

	// Verify PLATFORM_OTLP_ENDPOINT uses explicit override
	assert.Equal(t, "https://custom-otlp.example.com:4318", envVars["PLATFORM_OTLP_ENDPOINT"])
}

func TestK8sResourceCollectorGrpcProtocol(t *testing.T) {
	// A platformOtlpEndpoint without an http(s):// scheme is inferred as gRPC.
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-grpc.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	// Verify configmap uses otlp (gRPC) exporter instead of otlp_http
	configMap, exists := resources.ConfigMaps["suse-observability-agent-k8s-resource-collector-config"]
	require.True(t, exists, "k8s-resource-collector config map was not found")

	configData := configMap.Data["config.yaml"]
	assert.Contains(t, configData, "otlp/suse-observability:")
	assert.NotContains(t, configData, "otlp_http/suse-observability:")

	// Verify pipeline references gRPC exporter
	assert.Contains(t, configData, "exporters: [otlp/suse-observability]")

	// Verify endpoint uses explicit platformOtlpEndpoint
	deployment, exists := resources.Deployments["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector deployment was not found")

	envVars := make(map[string]string)
	for _, env := range deployment.Spec.Template.Spec.Containers[0].Env {
		if env.Value != "" {
			envVars[env.Name] = env.Value
		}
	}
	assert.Equal(t, "otlp-my-suse-observability-instance.com:443", envVars["PLATFORM_OTLP_ENDPOINT"])
}

func TestK8sResourceCollectorGrpcEndpointRequiresPort443(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-grpc-missing-port.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "treated as gRPC and must include port :443")
}

// TestK8sResourceCollectorMultiNodeAffinityAndPDB asserts the default (replicaCount: 2)
// renders preferred (not required) hostname-level pod anti-affinity and a PDB that
// allows one pod down at a time. Preferred lets sub-3-node clusters still install
// and roll upgrades; required would wedge them.
func TestK8sResourceCollectorMultiNodeAffinityAndPDB(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-enabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	deployment, exists := resources.Deployments["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector deployment was not found")

	require.NotNil(t, deployment.Spec.Template.Spec.Affinity, "deployment should have affinity (auto-injected when leaderElection enabled)")
	require.NotNil(t, deployment.Spec.Template.Spec.Affinity.PodAntiAffinity, "deployment should have pod anti-affinity")

	antiAffinity := deployment.Spec.Template.Spec.Affinity.PodAntiAffinity
	assert.Empty(t, antiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
		"anti-affinity must be preferred, not required, so 1- and 2-node clusters can install/upgrade")
	require.Len(t, antiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1,
		"expected exactly one preferred anti-affinity term")

	preferred := antiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0]
	assert.Equal(t, int32(100), preferred.Weight, "weight 100 keeps spread behaviour effectively required on multi-node clusters")
	assert.Equal(t, "kubernetes.io/hostname", preferred.PodAffinityTerm.TopologyKey)
	require.NotNil(t, preferred.PodAffinityTerm.LabelSelector)
	assert.Equal(t, "k8s-resource-collector", preferred.PodAffinityTerm.LabelSelector.MatchLabels["app.kubernetes.io/component"])

	pdb, exists := resources.Pdbs["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector PDB should be rendered when replicaCount > 1")
	require.NotNil(t, pdb.Spec.MaxUnavailable)
	assert.Equal(t, int32(1), pdb.Spec.MaxUnavailable.IntVal,
		"PDB allows one pod down at a time, keeping the other available during voluntary disruptions")
	require.NotNil(t, pdb.Spec.Selector)
	assert.Equal(t, "k8s-resource-collector", pdb.Spec.Selector.MatchLabels["app.kubernetes.io/component"])
}

// TestK8sResourceCollectorSingleReplicaSkipsPDB asserts that when replicaCount is 1
// (e.g. operator-set on a single-node cluster) the PDB is NOT rendered. A PDB
// with maxUnavailable: 1 over a single replica would still allow the only pod
// to be evicted, providing no protection — better to omit it than mislead.
func TestK8sResourceCollectorSingleReplicaSkipsPDB(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability-agent", "values/minimal.yaml", "values/k8s-resource-collector-single-replica.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	_, exists := resources.Deployments["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector deployment was not found")

	_, exists = resources.Pdbs["suse-observability-agent-k8s-resource-collector"]
	assert.False(t, exists, "PDB should be skipped when replicaCount is 1")
}

// TestK8sResourceCollectorIntegrationFlags verifies that enabling integration
// flags merges the corresponding API groups into the rendered ConfigMap
// (crd_api_group_filters.include). All four flags are set in a single values
// file so one render covers every integration's groups.
func TestK8sResourceCollectorIntegrationFlags(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t,
		"suse-observability-agent",
		"values/minimal.yaml",
		"values/k8s-resource-collector-integrations.yaml",
	)
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps["suse-observability-agent-k8s-resource-collector-config"]
	require.True(t, exists, "k8s-resource-collector config map was not found")
	configData := configMap.Data["config.yaml"]

	require.Contains(t, configData, "crd_api_group_filters:")
	require.Contains(t, configData, "include:")

	allGroups := []string{
		// suseRuntimeEnforcer
		"security.rancher.io",
		// suseAdmissionController — includes both current and successor report groups
		"policies.kubewarden.io", "wgpolicyk8s.io", "openreports.io",
		// suseVirtualization
		"kubevirt.io",
		// suseSbomScanner
		"sbomscanner.kubewarden.io", "storage.sbomscanner.kubewarden.io", "postgresql.cnpg.io",
	}
	for _, g := range allGroups {
		assert.Contains(t, configData, g, "expected integration group %q in include filter", g)
	}

	// Base operator-supplied wildcard must also survive the merge.
	assert.Contains(t, configData, `"*"`, "operator-supplied wildcard should survive integration merge")
}

// TestK8sResourceCollectorIntegrationFlagsRestrictedRBAC verifies that
// integration flags also populate the ClusterRole under restricted RBAC
// (useWildcard=false), and that they merge with operator-supplied crdApiGroups
// rather than replacing them.
func TestK8sResourceCollectorIntegrationFlagsRestrictedRBAC(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t,
		"suse-observability-agent",
		"values/minimal.yaml",
		"values/k8s-resource-collector-integrations-restricted-rbac.yaml",
	)
	resources := helmtestutil.NewKubernetesResources(t, output)

	clusterRole, exists := resources.ClusterRoles["suse-observability-agent-k8s-resource-collector"]
	require.True(t, exists, "k8s-resource-collector cluster role was not found")

	apiGroups := map[string]bool{}
	for _, rule := range clusterRole.Rules {
		for _, g := range rule.APIGroups {
			apiGroups[g] = true
		}
	}

	// Integration-added groups must appear.
	for _, g := range []string{"sbomscanner.kubewarden.io", "storage.sbomscanner.kubewarden.io", "postgresql.cnpg.io"} {
		assert.True(t, apiGroups[g], "integration group %q should appear in ClusterRole under restricted RBAC", g)
	}

	// Operator-supplied base entry must survive the merge.
	assert.True(t, apiGroups["policies.kubewarden.io"], "operator-supplied crdApiGroups entry should survive integration merge")
}

// TestK8sResourceCollectorIntegrationFlagsDisabledByDefault verifies that when
// all integration flags are explicitly set to false, no integration groups appear
// in the rendered output. Uses a dedicated values file rather than relying on
// chart defaults (which ship three flags enabled out of the box).
func TestK8sResourceCollectorIntegrationFlagsDisabledByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t,
		"suse-observability-agent",
		"values/minimal.yaml",
		"values/k8s-resource-collector-integrations-disabled.yaml",
	)
	resources := helmtestutil.NewKubernetesResources(t, output)

	configMap, exists := resources.ConfigMaps["suse-observability-agent-k8s-resource-collector-config"]
	require.True(t, exists, "k8s-resource-collector config map was not found")
	configData := configMap.Data["config.yaml"]

	integrationGroups := []string{
		"security.rancher.io", "policies.kubewarden.io", "wgpolicyk8s.io",
		"kubevirt.io", "sbomscanner.kubewarden.io", "storage.sbomscanner.kubewarden.io", "postgresql.cnpg.io",
	}
	for _, g := range integrationGroups {
		assert.NotContains(t, configData, g, "integration group %q should not appear when flags are disabled", g)
	}
}
