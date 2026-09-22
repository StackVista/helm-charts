package test

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

const namingBaselineCommit = "fb9ac7bdd"

// Fields are stored independently so a failure identifies the resource and field,
// rather than printing a diff of every rendered manifest.
type namingUpgradeContract map[string]map[string]interface{}

func TestResourceNamingUpgradeCompatibility(t *testing.T) {
	for _, scenario := range []struct {
		name, release, legacyPrefix string
		values                      map[string]string
	}{
		{name: "default", release: "suse-observability", legacyPrefix: "suse-observability"},
		{name: "nightly", release: "nightly", legacyPrefix: "nightly-suse-observability"},
		{
			name: "override-mono", release: "nightly", legacyPrefix: "existing-installation",
			values: map[string]string{
				"fullnameOverride":                 "existing-installation",
				"hbase.deployment.mode":            "Mono",
				"stackstate.features.server.split": "false",
			},
		},
		{
			name: "prefix-suffix", release: "suse-observability", legacyPrefix: "global-team-suse-observability-prod-end",
			values: map[string]string{
				"global.fullnamePrefix": "global-", "fullnamePrefix": "team-",
				"fullnameSuffix": "-prod", "global.fullnameSuffix": "-end",
			},
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			setValues := map[string]string{
				"stackstate.features.server.split":                         "true",
				"stackstate.components.replicationChecker.enabled":         "true",
				"stackstate.components.ui.extraEnv.secret.TEST_SECRET":     "example",
				"stackstate.components.all.metrics.servicemonitor.enabled": "true",
				// The baseline cannot render S3Proxy monitoring (missing helper).
				// TestResourceNamingStatelessComponents covers the corrected monitor.
				"s3proxy.metrics.servicemonitor.enabled": "false",
			}
			for key, value := range scenario.values {
				setValues[key] = value
			}
			chartPath, err := filepath.Abs("..")
			require.NoError(t, err)
			// Explicit fixture generation only; normal tests need no old checkout
			// or Git history. See testdata/resource-naming-upgrade/README.md.
			baselineChart := os.Getenv("RESOURCE_NAMING_BASELINE_CHART")
			if baselineChart != "" {
				chartPath = baselineChart
			}
			output, err := helm.RenderTemplateE(t, &helm.Options{
				ValuesFiles:    []string{filepath.Join(chartPath, "test/values/full.yaml")},
				SetValues:      setValues,
				KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
				Logger:         logger.Discard,
			}, chartPath, scenario.release, nil)
			require.NoError(t, err)
			actual := resourceNamingUpgradeContract(t, output)
			fixturePath := filepath.Join("testdata/resource-naming-upgrade", scenario.name+".json")
			if baselineChart != "" {
				data, err := json.MarshalIndent(actual, "", "  ")
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(fixturePath, append(data, '\n'), 0644))
				t.Logf("Wrote baseline from %s; source must be commit %s", baselineChart, namingBaselineCommit)
				return
			}
			data, err := os.ReadFile(fixturePath)
			require.NoError(t, err)
			var expected namingUpgradeContract
			require.NoError(t, json.Unmarshal(data, &expected))
			allowResourceNamingMigration(t, expected, scenario.legacyPrefix)

			for resource, fields := range expected {
				actualFields, found := actual[resource]
				if assert.True(t, found, "resource removed: %s", resource) {
					assert.Equal(t, fields, actualFields, "%s upgrade contract changed", resource)
				}
			}
			for resource := range actual {
				_, found := expected[resource]
				assert.True(t, found, "resource added: %s", resource)
			}
		})
	}
}

func resourceNamingUpgradeContract(t *testing.T, output string) namingUpgradeContract {
	t.Helper()
	contract := namingUpgradeContract{}
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(output), 4096)
	jobTimestamp := regexp.MustCompile(`[0-9]{2}t[0-9]{6}$`)
	testPodRandomSuffix := regexp.MustCompile(`-[a-z]{5}-test$`)
	for {
		var object unstructured.Unstructured
		err := decoder.Decode(&object)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if object.GetKind() == "" {
			continue
		}
		name := object.GetName()
		if object.GetKind() == "Job" {
			name = jobTimestamp.ReplaceAllString(name, "<timestamp>")
		}
		if object.GetKind() == "Pod" && object.GetAnnotations()["helm.sh/hook"] == "test-success" {
			name = testPodRandomSuffix.ReplaceAllString(name, "-<random>-test")
		}
		key := object.GetKind() + "/" + name
		// A chart can render the same Secret as both a regular resource and a
		// hook. Keep both entries, including their separate lifecycle identities.
		if hook := object.GetAnnotations()["helm.sh/hook"]; hook != "" {
			key += " [hook=" + hook + "]"
		}
		_, exists := contract[key]
		require.False(t, exists, "duplicate resource identity: %s", key)
		fields := map[string]interface{}{"namespace": object.GetNamespace()}
		capture := func(label string, path ...string) {
			value, found, err := unstructured.NestedFieldNoCopy(object.Object, path...)
			require.NoError(t, err)
			if found {
				fields[label] = value
			}
		}
		capture("selector", "spec", "selector")
		capture("roleRef", "roleRef")
		capture("subjects", "subjects")
		switch object.GetKind() {
		case "StatefulSet":
			capture("serviceName", "spec", "serviceName")
			capture("volumeClaimTemplates", "spec", "volumeClaimTemplates")
		case "PersistentVolumeClaim", "PodDisruptionBudget":
			capture("spec", "spec")
		case "Service":
			capture("ports", "spec", "ports")
			capture("clusterIP", "spec", "clusterIP")
		case "Secret":
			capture("type", "type")
			// Freeze credential Secret identities and keys, not random values or
			// salted hashes. Runtime lookup reuse needs a separate upgrade test.
			for _, field := range []string{"data", "stringData"} {
				if object.Object[field] == nil {
					continue
				}
				values, found, err := unstructured.NestedMap(object.Object, field)
				require.NoError(t, err)
				if found {
					keys := make([]string, 0, len(values))
					for key := range values {
						keys = append(keys, key)
					}
					sort.Strings(keys)
					fields[field+"Keys"] = keys
				}
			}
		}
		podPath := []string{"spec", "template", "spec"}
		if object.GetKind() == "CronJob" {
			podPath = []string{"spec", "jobTemplate", "spec", "template", "spec"}
		} else if object.GetKind() == "Pod" {
			podPath = []string{"spec"}
		}
		for _, field := range []string{"serviceAccountName", "imagePullSecrets", "volumes"} {
			capture(field, append(podPath, field)...)
		}
		for _, containerType := range []string{"containers", "initContainers"} {
			value, _, err := unstructured.NestedFieldNoCopy(object.Object, append(podPath, containerType)...)
			require.NoError(t, err)
			if value == nil {
				continue
			}
			containers, ok := value.([]interface{})
			require.True(t, ok, "%s must be a container list", containerType)
			for _, rawContainer := range containers {
				container := rawContainer.(map[string]interface{})
				prefix := containerType + "/" + container["name"].(string)
				if envFrom, found := container["envFrom"]; found {
					fields[prefix+"/envFrom"] = envFrom
				}
				if container["env"] == nil {
					continue
				}
				env, _, err := unstructured.NestedSlice(container, "env")
				require.NoError(t, err)
				for _, rawVariable := range env {
					variable := rawVariable.(map[string]interface{})
					if valueFrom, found := variable["valueFrom"]; found {
						fields[prefix+"/env/"+variable["name"].(string)] = valueFrom
					}
				}
			}
		}
		contract[key] = fields
	}
	// Canonical JSON types keep fresh renders and deserialized fixtures equal.
	data, err := json.Marshal(contract)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &contract))
	return contract
}

func allowResourceNamingMigration(t *testing.T, contract namingUpgradeContract, legacyPrefix string) {
	t.Helper()
	const canonical = "suse-observability"
	if legacyPrefix == canonical {
		return
	}
	// Exact resource allowlist: do not globally rewrite the old prefix, since
	// that would conceal accidental renames of storage or credential resources.
	for component, kinds := range map[string][]string{
		"ui":                  {"Deployment", "Service", "Secret", "ServiceMonitor"},
		"replication-checker": {"Deployment", "ServiceAccount", "Role", "RoleBinding"},
		"vmagent":             {"ConfigMap"},
	} {
		for _, kind := range kinds {
			oldKey := kind + "/" + legacyPrefix + "-" + component
			newKey := kind + "/" + canonical + "-" + component
			require.Contains(t, contract, oldKey)
			require.NotContains(t, contract, newKey)
			contract[newKey] = contract[oldKey]
			delete(contract, oldKey)
		}
	}
	// Only these fields may change their reference to a renamed resource.
	for resource, fields := range map[string][]string{
		"Deployment/suse-observability-ui":                   {"containers/ui/env/TEST_SECRET"},
		"Deployment/suse-observability-replication-checker":  {"serviceAccountName"},
		"RoleBinding/suse-observability-replication-checker": {"roleRef", "subjects"},
		"StatefulSet/suse-observability-vmagent":             {"volumes"},
	} {
		for _, field := range fields {
			require.Contains(t, contract[resource], field)
			data, err := json.Marshal(contract[resource][field])
			require.NoError(t, err)
			for _, component := range []string{"ui", "replication-checker", "vmagent"} {
				data = []byte(strings.ReplaceAll(string(data),
					`"`+legacyPrefix+"-"+component+`"`, `"`+canonical+"-"+component+`"`))
			}
			var value interface{}
			require.NoError(t, json.Unmarshal(data, &value))
			contract[resource][field] = value
		}
	}
}
