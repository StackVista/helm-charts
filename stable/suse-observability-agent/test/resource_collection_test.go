package test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

const resourceCollectionCRDName = "resourcecollections.observability.suse.com"
const resourceCollectionCollectorName = "suse-observability-agent-k8s-resource-collector"

func renderResourceCollections(t *testing.T, overrides map[string]string, fixtures ...string) helmtestutil.KubernetesResources {
	t.Helper()
	values := map[string]string{"otel.enabled": "true", "otel.prometheusScraping.enabled": "false"}
	for key, value := range overrides {
		values[key] = value
	}
	files := append([]string{"values/minimal.yaml"}, fixtures...)
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability-agent", &helm.Options{
		ValuesFiles: files,
		SetValues:   values,
	})
	return helmtestutil.NewKubernetesResources(t, output)
}

func resourceCollectionReceiver(t *testing.T, resources helmtestutil.KubernetesResources) map[string]any {
	t.Helper()
	configMap, exists := resources.ConfigMaps[resourceCollectionCollectorName+"-config"]
	require.True(t, exists)
	var config struct {
		Receivers map[string]map[string]any `yaml:"receivers"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(configMap.Data["config.yaml"]), &config))
	require.Contains(t, config.Receivers, "k8sresource")
	return config.Receivers["k8sresource"]
}

func TestResourceCollectionGates(t *testing.T) {
	for _, tc := range []struct {
		name      string
		overrides map[string]string
		collector bool
		enabled   bool
	}{
		{"enabled-default", nil, true, true},
		{"crd-install-disabled", map[string]string{"otel.k8sResourceCollector.resourceCollections.enabled": "false"}, true, false},
		{"collector-disabled", map[string]string{"otel.k8sResourceCollector.enabled": "false"}, false, false},
		{"otel-disabled", map[string]string{"otel.enabled": "false"}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := renderResourceCollections(t, tc.overrides)
			_, exists := resources.Unmapped[resourceCollectionCRDName]
			assert.Equal(t, tc.enabled, exists)
			_, exists = resources.Deployments[resourceCollectionCollectorName]
			assert.Equal(t, tc.collector, exists)
			if tc.collector {
				receiver := resourceCollectionReceiver(t, resources)
				assert.NotContains(t, receiver, "resource_collections")
				role := resources.ClusterRoles[resourceCollectionCollectorName]
				for _, rule := range role.Rules {
					if slices.Contains(rule.APIGroups, "observability.suse.com") {
						assert.True(t, tc.enabled, "disabled feature must not retain explicit declaration/status permissions")
					}
				}
			}
		})
	}
}

func TestResourceCollectionKeepCanBeDisabled(t *testing.T) {
	resources := renderResourceCollections(t, map[string]string{"otel.k8sResourceCollector.resourceCollections.keep": "false"})
	var definition struct {
		Metadata struct{ Annotations map[string]string }
	}
	require.NoError(t, yaml.Unmarshal([]byte(resources.Unmapped[resourceCollectionCRDName]), &definition))
	assert.NotContains(t, definition.Metadata.Annotations, "helm.sh/resource-policy")
}

func TestResourceCollectionInstallFlagPreservesBaselineConfiguration(t *testing.T) {
	enabled := resourceCollectionReceiver(t, renderResourceCollections(t, nil))
	disabled := resourceCollectionReceiver(t, renderResourceCollections(t,
		map[string]string{"otel.k8sResourceCollector.resourceCollections.enabled": "false"}))
	assert.Equal(t, enabled, disabled)
}

func TestResourceCollectionDefinition(t *testing.T) {
	resources := renderResourceCollections(t, nil)
	definition, exists := resources.Unmapped[resourceCollectionCRDName]
	require.True(t, exists)
	var crd struct {
		Metadata struct {
			Annotations map[string]string
		}
		Spec struct {
			Group    string
			Scope    string
			Names    struct{ Kind string }
			Versions []struct {
				Name         string
				Served       bool
				Storage      bool
				Subresources map[string]any
				Schema       struct{ OpenAPIV3Schema map[string]any }
			}
		}
	}
	require.NoError(t, yaml.Unmarshal([]byte(definition), &crd))
	assert.Equal(t, "keep", crd.Metadata.Annotations["helm.sh/resource-policy"])
	assert.Equal(t, "observability.suse.com", crd.Spec.Group)
	assert.Equal(t, "Cluster", crd.Spec.Scope)
	assert.Equal(t, "ResourceCollection", crd.Spec.Names.Kind)
	require.Len(t, crd.Spec.Versions, 1)
	version := crd.Spec.Versions[0]
	assert.Equal(t, "v1alpha1", version.Name)
	assert.True(t, version.Served)
	assert.True(t, version.Storage)
	assert.Contains(t, version.Subresources, "status")
	properties := version.Schema.OpenAPIV3Schema["properties"].(map[string]any)
	spec := properties["spec"].(map[string]any)
	specProperties := spec["properties"].(map[string]any)
	selection := specProperties["resources"].(map[string]any)["items"].(map[string]any)
	selectionProperties := selection["properties"].(map[string]any)
	assert.Len(t, selectionProperties, 5)
	for _, field := range []string{"group", "resource", "namespaces", "labelSelector", "fieldSelector"} {
		assert.Contains(t, selectionProperties, field)
	}
	assert.NotContains(t, selectionProperties, "version")
	assert.Equal(t, float64(1), selectionProperties["group"].(map[string]any)["minLength"])
	labelProperties := selectionProperties["labelSelector"].(map[string]any)["properties"].(map[string]any)
	expressionProperties := labelProperties["matchExpressions"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
	assert.Equal(t, []any{"In", "NotIn", "Exists", "DoesNotExist"}, expressionProperties["operator"].(map[string]any)["enum"])
	statusProperties := properties["status"].(map[string]any)["properties"].(map[string]any)
	selectionStatus := statusProperties["resources"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
	assert.NotContains(t, selectionStatus, "version")
}

func TestResourceCollectionRBAC(t *testing.T) {
	for _, wildcard := range []string{"true", "false"} {
		t.Run("wildcard-"+wildcard, func(t *testing.T) {
			resources := renderResourceCollections(t, map[string]string{"otel.k8sResourceCollector.rbac.useWildcard": wildcard})
			role := resources.ClusterRoles[resourceCollectionCollectorName]
			declarationRules := map[string][]string{}
			for _, rule := range role.Rules {
				if slices.Contains(rule.APIGroups, "observability.suse.com") {
					require.Len(t, rule.Resources, 1)
					declarationRules[rule.Resources[0]] = rule.Verbs
				}
				for _, verb := range rule.Verbs {
					if verb == "create" || verb == "delete" || verb == "update" || verb == "patch" {
						assert.Equal(t, []string{"resourcecollections/status"}, rule.Resources)
						assert.Equal(t, []string{"observability.suse.com"}, rule.APIGroups)
					}
				}
				if wildcard == "false" {
					assert.NotContains(t, rule.APIGroups, "*")
				}
			}
			require.Len(t, declarationRules, 2)
			assert.ElementsMatch(t, []string{"get", "list", "watch"}, declarationRules["resourcecollections"])
			assert.ElementsMatch(t, []string{"patch", "update"}, declarationRules["resourcecollections/status"])
			binding := resources.ClusterRoleBindings[resourceCollectionCollectorName]
			assert.Equal(t, rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: role.Name}, binding.RoleRef)
			require.Len(t, binding.Subjects, 1)
			assert.Equal(t, resources.Deployments[resourceCollectionCollectorName].Spec.Template.Spec.ServiceAccountName, binding.Subjects[0].Name)
			assert.Equal(t, "ServiceAccount", binding.Subjects[0].Kind)
			assert.NotEmpty(t, binding.Subjects[0].Namespace)
			assert.NotContains(t, resources.Deployments, "suse-observability-agent-resource-collection-controller")
		})
	}
}

func TestResourceCollectionDeclarationsOnly(t *testing.T) {
	resources := renderResourceCollections(t, map[string]string{"otel.k8sResourceCollector.rbac.useWildcard": "false"}, "values/resource-collections-only.yaml")
	receiver := resourceCollectionReceiver(t, resources)
	groups := receiver["cr_api_groups"].(map[string]any)
	assert.Equal(t, []any{}, groups["include"], "empty must remain an explicit list, not null/default-all")
	assert.NotContains(t, receiver, "objects")
	role := resources.ClusterRoles[resourceCollectionCollectorName]
	require.Len(t, role.Rules, 3, "CR-only restricted mode grants only the source/status and CRD reads")
	for _, rule := range role.Rules {
		assert.NotContains(t, rule.APIGroups, "*")
	}

	disabled := renderResourceCollections(t,
		map[string]string{
			"otel.k8sResourceCollector.resourceCollections.enabled": "false",
			"otel.k8sResourceCollector.rbac.useWildcard": "false",
		},
		"values/resource-collections-only.yaml")
	assert.Equal(t, receiver, resourceCollectionReceiver(t, disabled))
	assert.Len(t, disabled.ClusterRoles[resourceCollectionCollectorName].Rules, 1)
}

func TestResourceCollectionPreservesIntegrationBaseline(t *testing.T) {
	for _, enabled := range []string{"true", "false"} {
		t.Run(enabled, func(t *testing.T) {
			resources := renderResourceCollections(t, map[string]string{"otel.k8sResourceCollector.resourceCollections.enabled": enabled})
			receiver := resourceCollectionReceiver(t, resources)
			groups := receiver["cr_api_groups"].(map[string]any)
			assert.Contains(t, groups["include"], "kubevirt.io")
			objects := receiver["objects"].([]any)
			require.Len(t, objects, 1)
			pods := objects[0].(map[string]any)
			assert.Equal(t, "pods", pods["name"])
			assert.Equal(t, "kubevirt.io in (virt-launcher, virt-controller, virt-api, virt-operator, virt-handler)", pods["label_selector"])
		})
	}
}

func TestResourceCollectionCanonicalArtifact(t *testing.T) {
	schema, err := os.ReadFile("../files/crds/" + resourceCollectionCRDName + ".yaml")
	require.NoError(t, err)
	provenanceBytes, err := os.ReadFile("../files/crds/" + resourceCollectionCRDName + ".provenance.json")
	require.NoError(t, err)
	var provenance struct {
		Repository string `json:"repository"`
		Path       string `json:"path"`
		SHA256     string `json:"sha256"`
	}
	require.NoError(t, json.Unmarshal(provenanceBytes, &provenance))
	assert.Equal(t, "StackVista/sts-opentelemetry-collector", provenance.Repository)
	assert.Equal(t, "receiver/k8sresourcereceiver/resourcecollection/resourcecollection-crd.yaml", provenance.Path)
	sum := sha256.Sum256(schema)
	assert.Equal(t, hex.EncodeToString(sum[:]), provenance.SHA256)
	assert.NotContains(t, string(schema), "\n#")

	if canonical := os.Getenv("RESOURCE_COLLECTION_CANONICAL_CRD"); canonical != "" {
		upstream, err := os.ReadFile(canonical)
		require.NoError(t, err)
		assert.Equal(t, upstream, schema, "chart schema must be the unchanged canonical artifact")
	}
}

func TestResourceCollectionBinaryValidation(t *testing.T) {
	binary := os.Getenv("OTEL_AGENT_BINARY")
	if binary == "" {
		if os.Getenv("RESOURCE_COLLECTION_REQUIRE_BINARY") == "true" {
			t.Fatal("OTEL_AGENT_BINARY is required for candidate ResourceCollection validation")
		}
		t.Skip("set OTEL_AGENT_BINARY to validate rendered resource-collector configuration")
	}
	binary, err := exec.LookPath(binary)
	require.NoError(t, err)
	binary, err = filepath.Abs(binary)
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		overrides map[string]string
		fixtures  []string
	}{
		{"disabled", map[string]string{"otel.k8sResourceCollector.resourceCollections.enabled": "false"}, nil},
		{"enabled", nil, nil},
		{"declarations-only", nil, []string{"values/resource-collections-only.yaml"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := renderResourceCollections(t, tc.overrides, tc.fixtures...)
			config := resources.ConfigMaps[resourceCollectionCollectorName+"-config"].Data["config.yaml"]
			dir := t.TempDir()
			configPath := filepath.Join(dir, "config.yaml")
			require.NoError(t, os.WriteFile(configPath, []byte(config), 0600))
			container := resources.Deployments[resourceCollectionCollectorName].Spec.Template.Spec.Containers[0]
			env := []string{"PATH=/usr/bin:/bin", "TMPDIR=" + dir, "KUBECONFIG=/dev/null", "KUBERNETES_SERVICE_HOST=127.0.0.1"}
			for _, variable := range container.Env {
				value := variable.Value
				if variable.ValueFrom != nil {
					switch {
					case variable.ValueFrom.SecretKeyRef != nil:
						ref := variable.ValueFrom.SecretKeyRef
						secret, found := resources.Secrets[ref.Name]
						require.True(t, found, "missing fixture Secret %s", ref.Name)
						if data, found := secret.Data[ref.Key]; found {
							value = string(data)
						} else {
							value, found = secret.StringData[ref.Key]
							require.True(t, found, "missing fixture Secret key %s", ref.Key)
						}
					case variable.ValueFrom.FieldRef != nil:
						fields := map[string]string{
							"status.podIP": "127.0.0.1", "metadata.uid": "validation-uid", "metadata.namespace": "validation",
						}
						var found bool
						value, found = fields[variable.ValueFrom.FieldRef.FieldPath]
						require.True(t, found, "unsupported field %s", variable.ValueFrom.FieldRef.FieldPath)
					default:
						t.Fatalf("unsupported environment source for %s", variable.Name)
					}
				}
				value = strings.ReplaceAll(value, "$(STS_URL)", "https://127.0.0.1:18443/receiver")
				env = append(env, variable.Name+"="+value)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary, "validate", "--config="+configPath)
			command.Dir = dir
			command.Env = env
			output, err := command.CombinedOutput()
			require.NoError(t, err, "rendered resource-collector validate: %s", output)
		})
	}
}
