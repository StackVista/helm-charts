package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

func TestAnomalyResourceNamesPreserveLegacyConfiguration(t *testing.T) {
	for _, scenario := range []struct {
		name, release, namespace, base, clusterBase string
		values                                      map[string]string
	}{
		{
			name: "custom-release", release: "orders", namespace: "observability",
			base: "orders-anomaly-detection", clusterBase: "observability-orders-anomaly-detection",
		},
		{
			name: "release-matches-chart-and-namespace", release: "anomaly-detection", namespace: "anomaly-detection",
			base: "anomaly-detection", clusterBase: "anomaly-detection",
		},
		{
			name: "override-with-prefixes-and-suffixes", release: "orders", namespace: "customer",
			base: "global-local-custom-local-global", clusterBase: "global-local-custom-local-global",
			values: map[string]string{
				"fullnameOverride": "CUSTOM", "global.fullnamePrefix": "GLOBAL-", "fullnamePrefix": "LOCAL-",
				"fullnameSuffix": "-LOCAL", "global.fullnameSuffix": "-GLOBAL",
			},
		},
		{
			name: "long-release", release: strings.Repeat("r", 53), namespace: "customer",
			base: strings.Repeat("r", 53), clusterBase: "customer-" + strings.Repeat("r", 45),
		},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", scenario.name, upgrade), func(t *testing.T) {
				values := map[string]string{
					"stackstate.instance": "https://analysis.example", "global.receiverApiKey": "test-key",
					"image.pullSecretUsername": "test-user", "image.pullSecretPassword": "test-password",
					"metrics.serviceMonitor.enabled": "true", "ingress.enabled": "true",
					"ingress.hosts[0].host": "analysis.example", "ingress.tls[0].hosts[0]": "analysis.example",
				}
				for key, value := range scenario.values {
					values[key] = value
				}
				options := &helm.Options{
					SetValues: values, KubectlOptions: &k8s.KubectlOptions{Namespace: scenario.namespace},
				}
				args := []string{"--api-versions", "networking.k8s.io/v1/Ingress"}
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, scenario.release, options, args...)
				documents := anomalyNamingDocuments(t, output, nil, nil)
				expected := []string{
					"Deployment/" + scenario.base + "-spotlight-manager",
					"Deployment/" + scenario.base + "-spotlight-worker",
					"Service/" + scenario.base + "-spotlight-manager",
					"ServiceMonitor/suse-observability-spotlight-manager",
					"ServiceAccount/" + scenario.base + "-sa",
					"Role/stackstate-aad", "RoleBinding/stackstate-aad",
					"ClusterRoleBinding/" + scenario.clusterBase + "-aad-authentication",
					"Secret/" + scenario.base + "-stackstate-auth-secret",
					"Secret/" + scenario.base + "-pull-secret",
					"ConfigMap/" + scenario.base + "-spotlight-config-base",
					"PersistentVolumeClaim/spotlight-artifacts-volume-claim",
					"Ingress/" + scenario.base,
					"PodDisruptionBudget/suse-observability-anomaly-detection",
				}
				require.Len(t, documents, len(expected))
				for _, key := range expected {
					require.Contains(t, documents, key)
				}
				resources := helmtestutil.NewKubernetesResources(t, output)
				worker := resources.Deployments[scenario.base+"-spotlight-worker"]
				manager := resources.Deployments[scenario.base+"-spotlight-manager"]
				assert.Equal(t, scenario.base+"-spotlight-worker", worker.Spec.Selector.MatchLabels["name"])
				assert.Equal(t, scenario.base+"-spotlight-manager", manager.Spec.Selector.MatchLabels["name"])
				assert.Equal(t, manager.Spec.Selector.MatchLabels, resources.Services[scenario.base+"-spotlight-manager"].Spec.Selector)
				assert.Equal(t, "stackstate-aad", resources.ServiceAccounts[scenario.base+"-sa"].Annotations["stackstate.io/roles"])
				// These legacy Ingress references do not identify generated resources:
				// its backend differs from the manager Service, and TLS is external.
				ingress := resources.Ingresses[scenario.base]
				require.Len(t, ingress.Spec.Rules, 1)
				require.Len(t, ingress.Spec.Rules[0].HTTP.Paths, 1)
				assert.Equal(t, scenario.base, ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name)
				require.Len(t, ingress.Spec.TLS, 1)
				assert.Equal(t, scenario.base+"-ingress-tls", ingress.Spec.TLS[0].SecretName)
			})
		}
	}
}

func TestAnomalyResourcesFollowDedicatedHelpers(t *testing.T) {
	const base = "orders-anomaly-detection"
	fixtures := []struct {
		helper, kind, legacy string
	}{
		{"manager.deployment", "Deployment", base + "-spotlight-manager"},
		{"worker.deployment", "Deployment", base + "-spotlight-worker"},
		{"manager.service", "Service", base + "-spotlight-manager"},
		{"manager.servicemonitor", "ServiceMonitor", "suse-observability-spotlight-manager"},
		{"serviceaccount", "ServiceAccount", base + "-sa"},
		{"authentication.role", "Role", "stackstate-aad"},
		{"authentication.rolebinding", "RoleBinding", "stackstate-aad"},
		{"authentication.clusterrolebinding", "ClusterRoleBinding", "observability-" + base + "-aad-authentication"},
		{"authentication.secret", "Secret", base + "-stackstate-auth-secret"},
		{"pull.secret", "Secret", base + "-pull-secret"},
		{"config.configmap", "ConfigMap", base + "-spotlight-config-base"},
		{"manager.artifacts.persistentvolumeclaim", "PersistentVolumeClaim", "spotlight-artifacts-volume-claim"},
		{"ingress", "Ingress", base},
		{"pdb", "PodDisruptionBudget", "suse-observability-anomaly-detection"},
	}
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	definitions := string(data)
	renames := map[string]string{}
	for _, fixture := range fixtures {
		header := `define "anomaly-detection.` + fixture.helper + `.fullname"`
		require.Equal(t, 1, strings.Count(definitions, header))
		definitions = strings.Replace(definitions, header, `define "test.original.anomaly-detection.`+fixture.helper+`.fullname"`, 1)
		name := "explicit-" + strings.ReplaceAll(fixture.helper, ".", "-")
		definitions += "\n{{- " + header + " -}}" + name + "{{- end -}}\n"
		renames[fixture.kind+"/"+fixture.legacy] = name
	}
	require.NoError(t, os.WriteFile(path, []byte(definitions), 0600))
	seen := map[string]bool{}
	for _, scenario := range []struct {
		name   string
		values map[string]string
	}{
		{"token", map[string]string{
			"metrics.serviceMonitor.enabled": "true", "ingress.enabled": "true",
			"ingress.hosts[0].host": "analysis.example", "ingress.tls[0].hosts[0]": "analysis.example",
			"image.pullSecretUsername": "test-user", "image.pullSecretPassword": "test-password",
		}},
		{"no-cluster-binding-or-receiver-key", map[string]string{
			"cluster-role.enabled": "false", "global.receiverApiKey": "",
		}},
		{"cookie-external-pull-secret", map[string]string{
			"stackstate.authType": "cookie", "stackstate.username": "test-user", "stackstate.password": "test-password",
			"image.pullSecretName":       "customer-registry",
			"global.imagePullSecrets[0]": "\\{\\{ .Release.Name }}-registry",
		}},
		{"api-token", map[string]string{
			"stackstate.authType": "api-token", "stackstate.apiToken": "test-token",
		}},
		{"platform-pull-secret", map[string]string{
			"image.pullSecretUsername": "local-user", "image.pullSecretPassword": "local-password",
			"global.suseObservability.pullSecret.username": "platform-user",
			"global.suseObservability.pullSecret.password": "platform-password",
		}},
		{"docker-config-pull-reference", map[string]string{
			"image.pullSecretDockerConfigJson": "test-docker-config",
		}},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", scenario.name, upgrade), func(t *testing.T) {
				values := map[string]string{
					"stackstate.instance": "https://analysis.example", "global.receiverApiKey": "test-key",
					"manager.persistentStorage.size": "17Gi", "manager.persistentStorage.storageClass": "artifacts",
				}
				for key, value := range scenario.values {
					values[key] = value
				}
				options := &helm.Options{
					SetValues: values, KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
				}
				args := []string{"--api-versions", "networking.k8s.io/v1/Ingress"}
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				before := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "orders", options, args...)
				after, err := helm.RenderTemplateE(t, options, chart, "orders", nil, args...)
				require.NoError(t, err)
				actual := anomalyNamingDocuments(t, after, nil, nil)
				expected := anomalyNamingDocuments(t, before, renames, seen)
				for key, object := range expected {
					if object["kind"] != "Deployment" {
						continue
					}
					annotations := anomalyDeploymentAnnotations(object)
					require.Contains(t, actual, key)
					actualAnnotations := anomalyDeploymentAnnotations(actual[key])
					const checksum = "checksum/anomaly-detection-config-base"
					assert.NotEqual(t, annotations[checksum], actualAnnotations[checksum])
					annotations[checksum] = actualAnnotations[checksum]
				}
				// Entire resources must match: selectors, application role annotations,
				// external references, credentials, PVC specification and configuration.
				// Only the existing ConfigMap checksum responds to the fixture rename.
				assert.Equal(t, expected, actual)
			})
		}
	}
	for key := range renames {
		assert.True(t, seen[key], "helper not exercised: %s", key)
	}
}

func anomalyNamingDocuments(t *testing.T, output string, renames map[string]string, seen map[string]bool) map[string]map[string]interface{} {
	t.Helper()
	documents := map[string]map[string]interface{}{}
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(output), 4096)
	for {
		var object map[string]interface{}
		err := decoder.Decode(&object)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if len(object) == 0 {
			continue
		}
		metadata := object["metadata"].(map[string]interface{})
		key := object["kind"].(string) + "/" + metadata["name"].(string)
		if name, ok := renames[key]; ok {
			metadata["name"] = name
			seen[key] = true
		}
		rewriteAnomalyNamingReferences(object, renames)
		key = object["kind"].(string) + "/" + metadata["name"].(string)
		require.NotContains(t, documents, key, "duplicate resource")
		documents[key] = object
	}
	return documents
}

func rewriteAnomalyNamingReferences(value interface{}, renames map[string]string) {
	switch node := value.(type) {
	case map[string]interface{}:
		for key, child := range node {
			switch key {
			case "serviceAccountName":
				if name, ok := renames["ServiceAccount/"+child.(string)]; ok {
					node[key] = name
				}
			case "secretKeyRef", "configMap", "persistentVolumeClaim", "roleRef":
				ref := child.(map[string]interface{})
				kind, field := "ConfigMap", "name"
				switch key {
				case "secretKeyRef":
					kind = "Secret"
				case "persistentVolumeClaim":
					kind, field = "PersistentVolumeClaim", "claimName"
				case "roleRef":
					kind = ref["kind"].(string)
				}
				if name, ok := renames[kind+"/"+ref[field].(string)]; ok {
					ref[field] = name
				}
			case "imagePullSecrets", "subjects":
				for _, entry := range child.([]interface{}) {
					ref := entry.(map[string]interface{})
					kind := "Secret"
					if key == "subjects" {
						kind = ref["kind"].(string)
					}
					if name, ok := renames[kind+"/"+ref["name"].(string)]; ok {
						ref["name"] = name
					}
				}
			case "args":
				args := child.([]interface{})
				for i, arg := range args {
					if arg == "--manager-host" && i+1 < len(args) {
						if name, ok := renames["Service/"+args[i+1].(string)]; ok {
							args[i+1] = name
						}
					}
				}
			}
			rewriteAnomalyNamingReferences(child, renames)
		}
	case []interface{}:
		for _, child := range node {
			rewriteAnomalyNamingReferences(child, renames)
		}
	}
}

func anomalyDeploymentAnnotations(object map[string]interface{}) map[string]interface{} {
	spec := object["spec"].(map[string]interface{})
	template := spec["template"].(map[string]interface{})
	metadata := template["metadata"].(map[string]interface{})
	return metadata["annotations"].(map[string]interface{})
}
