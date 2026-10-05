package test

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
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

func TestElasticsearchResourcesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	definitions := string(data)
	fixtures := []struct {
		helper, kind, suffix string
	}{
		{"statefulset", "StatefulSet", ""},
		{"service", "Service", ""},
		{"headless.service", "Service", "-headless"},
		{"config.configmap", "ConfigMap", "-config"},
		{"credentials.secret", "Secret", "-credentials"},
		{"certificates.secret", "Secret", "-certs"},
		{"serviceaccount", "ServiceAccount", ""},
		{"role", "Role", ""},
		{"rolebinding", "RoleBinding", ""},
		{"podsecuritypolicy", "PodSecurityPolicy", ""},
		{"pull.secret", "Secret", "-pull-secret"},
		{"pdb", "PodDisruptionBudget", "-pdb"},
		{"ingress", "Ingress", ""},
		{"test.pod", "Pod", ""},
	}
	for _, fixture := range fixtures {
		header := `define "elasticsearch.` + fixture.helper + `.fullname"`
		require.Equal(t, 1, strings.Count(definitions, header))
		definitions = strings.Replace(definitions, header, `define "test.original.elasticsearch.`+fixture.helper+`.fullname"`, 1)
		definitions += "\n{{- " + header + " -}}explicit-" + strings.ReplaceAll(fixture.helper, ".", "-") + "{{- end -}}\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(definitions), 0600))
	for _, scenario := range []struct {
		name, uname, service, master string
		values                       map[string]string
	}{
		{"default", "suse-observability-elasticsearch-master", "suse-observability-elasticsearch-master", "suse-observability-elasticsearch-master", nil},
		{"name-override", "customer-master", "customer-master", "customer-master", map[string]string{"nameOverride": "customer"}},
		{"fullname-override", "customer", "customer", "customer", map[string]string{"fullnameOverride": "customer"}},
		{"custom-master-service", "suse-observability-elasticsearch-master", "customer-master", "customer-master", map[string]string{"masterService": "customer-master"}},
		{"data-group", "suse-observability-elasticsearch-data", "suse-observability-elasticsearch-data", "external-master", map[string]string{"nodeGroup": "data", "masterService": "external-master"}},
		{"account-and-policy-overrides", "suse-observability-elasticsearch-master", "suse-observability-elasticsearch-master", "suse-observability-elasticsearch-master", map[string]string{
			"rbac.serviceAccountName": "customer-account", "podSecurityPolicy.name": "customer-policy",
		}},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", scenario.name, upgrade), func(t *testing.T) {
				values := map[string]string{
					"prometheus-elasticsearch-exporter.enabled": "false", "replicas": "2",
					"rbac.create": "true", "podSecurityPolicy.create": "true",
					"createCert": "true", "secret.enabled": "true", "secret.password": "test-password",
					"esConfig.elasticsearch\\.yml": "test: true", "ingress.enabled": "true",
					"ingress.hosts[0]": "search.example", "ingress.tls[0].secretName": "customer-tls",
					"ingress.tls[0].hosts[0]": "search.example",
					"pullSecretUsername":      "test-user", "pullSecretPassword": "test-password",
					"volumeClaimTemplate.resources.requests.storage": "17Gi",
					"volumeClaimTemplate.storageClassName":           "search-data",
					"secretMounts[0].name":                           "external", "secretMounts[0].secretName": "customer-secret",
					"secretMounts[0].path": "/external",
				}
				for key, value := range scenario.values {
					values[key] = value
				}
				options := &helm.Options{
					SetValues: values, KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
				}
				args := []string{"--api-versions", "policy/v1beta1/PodSecurityPolicy"}
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				beforeOutput := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "orders", options, args...)
				resources := helmtestutil.NewKubernetesResources(t, beforeOutput)
				require.Contains(t, resources.Services, scenario.service)
				require.Contains(t, resources.Ingresses, scenario.uname)
				backend := resources.Ingresses[scenario.uname].Spec.Rules[0].HTTP.Paths[0].Backend.Service
				require.NotNil(t, backend)
				assert.Equal(t, scenario.service, backend.Name, "Ingress should target the created Service, including custom masterService")
				require.Len(t, resources.Pods, 1)
				for _, pod := range resources.Pods {
					assert.Contains(t, strings.Join(pod.Spec.Containers[0].Command, "\n"),
						"'"+scenario.service+":9200/_cluster/health?", "Helm health test should target the created Service")
				}
				before := elasticsearchNamingDocuments(t, beforeOutput)
				output, err := helm.RenderTemplateE(t, options, chart, "orders", nil, args...)
				require.NoError(t, err)
				after := elasticsearchNamingDocuments(t, output)
				require.Len(t, before, len(fixtures))
				renames := map[string]string{}
				for _, fixture := range fixtures {
					old := scenario.uname + fixture.suffix
					if fixture.kind == "Service" {
						old = scenario.service + fixture.suffix
					}
					if fixture.helper == "test.pod" {
						old = "helm-test"
					}
					if fixture.helper == "serviceaccount" && scenario.values["rbac.serviceAccountName"] != "" {
						continue
					}
					if fixture.helper == "podsecuritypolicy" && scenario.values["podSecurityPolicy.name"] != "" {
						old = scenario.values["podSecurityPolicy.name"]
					}
					key := fixture.kind + "/" + old
					require.Contains(t, before, key)
					renames[key] = "explicit-" + strings.ReplaceAll(fixture.helper, ".", "-")
				}
				expected := map[string]map[string]interface{}{}
				for key, object := range before {
					metadata := object["metadata"].(map[string]interface{})
					if name, ok := renames[key]; ok {
						metadata["name"] = name
						key = object["kind"].(string) + "/" + name
					}
					rewriteElasticsearchNamingReferences(object, renames)
					if object["kind"] == "StatefulSet" {
						pod := elasticsearchNamingPod(object)
						// Even with a configured account, the old rbac.create path
						// selects uname. Preserve that existing selection.
						pod["serviceAccountName"] = "explicit-serviceaccount"
						object["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["name"] = "explicit-statefulset"
						annotations := object["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
						actualAnnotations := after[key]["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
						assert.NotEqual(t, annotations["configchecksum"], actualAnnotations["configchecksum"])
						annotations["configchecksum"] = actualAnnotations["configchecksum"]
						// Bootstrap node names historically ignore name/fullname overrides.
						if scenario.values["nameOverride"] == "" && scenario.values["fullnameOverride"] == "" {
							rewriteElasticsearchBootstrapNodes(pod, scenario.uname, "explicit-statefulset")
						}
					}
					if key == "Secret/explicit-certificates-secret" {
						// Generated certificates are random on client-side renders.
						// Validate their DNS identity separately before comparing metadata.
						service := "explicit-service"
						if values["nodeGroup"] == "data" {
							service = scenario.master
						}
						assertElasticsearchCertificateNames(t, after[key], service)
						object["data"] = after[key]["data"]
					}
					expected[key] = object
				}
				assert.Equal(t, expected, after, "Only fixture names/references, regenerated certificates and the existing config checksum may differ")
			})
		}
	}
}

func elasticsearchNamingDocuments(t *testing.T, output string) map[string]map[string]interface{} {
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
		if object["kind"] == "Pod" {
			containers := object["spec"].(map[string]interface{})["containers"].([]interface{})
			containers[0].(map[string]interface{})["name"] = "test-container"
			if metadata["name"] != "explicit-test-pod" {
				assert.Regexp(t, `^orders-[a-z]{5}-test$`, metadata["name"])
				metadata["name"] = "helm-test"
				key = "Pod/helm-test"
			}
		}
		require.NotContains(t, documents, key)
		documents[key] = object
	}
	return documents
}

func rewriteElasticsearchNamingReferences(value interface{}, renames map[string]string) {
	switch node := value.(type) {
	case map[string]interface{}:
		for key, child := range node {
			switch key {
			case "serviceName":
				if name, ok := renames["Service/"+child.(string)]; ok {
					node[key] = name
				}
			case "secretKeyRef", "secret", "configMap", "roleRef":
				ref := child.(map[string]interface{})
				kind, field := "Secret", "name"
				if key == "secret" {
					field = "secretName"
				} else if key == "configMap" {
					kind = "ConfigMap"
				} else if key == "roleRef" {
					kind = ref["kind"].(string)
				}
				if name, ok := renames[kind+"/"+ref[field].(string)]; ok {
					ref[field] = name
				}
			case "imagePullSecrets", "subjects":
				for _, item := range child.([]interface{}) {
					ref := item.(map[string]interface{})
					kind := "Secret"
					if key == "subjects" {
						kind = ref["kind"].(string)
					}
					if name, ok := renames[kind+"/"+ref["name"].(string)]; ok {
						ref["name"] = name
					}
				}
			case "resourceNames":
				items := child.([]interface{})
				for i, item := range items {
					if name, ok := renames["PodSecurityPolicy/"+item.(string)]; ok {
						items[i] = name
					}
				}
			case "value", "command":
				// DNS references in discovery, the shutdown sidecar and Helm test.
				// Replace endpoints only; the runtime master-node prefix stays unchanged.
				rewriteElasticsearchEndpointText(node, key, child, renames)
			case "service":
				if ref, ok := child.(map[string]interface{}); ok {
					if name, ok := renames["Service/"+ref["name"].(string)]; ok {
						ref["name"] = name
					}
				}
			}
			rewriteElasticsearchNamingReferences(child, renames)
		}
	case []interface{}:
		for _, child := range node {
			rewriteElasticsearchNamingReferences(child, renames)
		}
	}
}

func rewriteElasticsearchEndpointText(node map[string]interface{}, key string, value interface{}, renames map[string]string) {
	rewrite := func(text string) string {
		for ref, name := range renames {
			if strings.HasPrefix(ref, "Service/") {
				old := strings.TrimPrefix(ref, "Service/")
				if text == old {
					text = name
				}
				text = strings.ReplaceAll(text, old+":9200", name+":9200")
			}
		}
		return text
	}
	if text, ok := value.(string); ok {
		node[key] = rewrite(text)
	} else if items, ok := value.([]interface{}); ok {
		for i, item := range items {
			items[i] = rewrite(item.(string))
		}
	}
}

func rewriteElasticsearchBootstrapNodes(pod map[string]interface{}, old, name string) {
	for _, item := range pod["containers"].([]interface{}) {
		container := item.(map[string]interface{})
		for _, entry := range container["env"].([]interface{}) {
			env := entry.(map[string]interface{})
			if env["name"] == "cluster.initial_master_nodes" {
				env["value"] = strings.ReplaceAll(env["value"].(string), old+"-", name+"-")
			}
		}
	}
}

func elasticsearchNamingPod(object map[string]interface{}) map[string]interface{} {
	return object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})
}

func assertElasticsearchCertificateNames(t *testing.T, secret map[string]interface{}, service string) {
	t.Helper()
	data := secret["data"].(map[string]interface{})
	require.Len(t, data, 3)
	certPEM, err := base64.StdEncoding.DecodeString(data["tls.crt"].(string))
	require.NoError(t, err)
	block, _ := pem.Decode(certPEM)
	require.NotNil(t, block)
	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	assert.Equal(t, service, cert.Subject.CommonName)
	assert.Equal(t, []string{service, service + ".observability", service + ".observability.svc"}, cert.DNSNames)
}
