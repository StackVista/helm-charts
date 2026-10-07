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

func TestExporterResourcesFollowDedicatedHelpers(t *testing.T) {
	fixtures := []struct{ helper, kind, suffix string }{
		{"deployment", "Deployment", ""},
		{"service", "Service", ""},
		{"serviceaccount", "ServiceAccount", ""},
		{"role", "Role", ""},
		{"rolebinding", "RoleBinding", ""},
		{"podsecuritypolicy", "PodSecurityPolicy", ""},
		{"certificates.secret", "Secret", "-cert"},
		{"servicemonitor", "ServiceMonitor", ""},
		{"podmonitor", "PodMonitor", ""},
		{"prometheusrule", "PrometheusRule", ""},
	}
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	definitions := string(data)
	for _, fixture := range fixtures {
		header := `define "elasticsearch-exporter.` + fixture.helper + `.fullname"`
		require.Equal(t, 1, strings.Count(definitions, header))
		definitions = strings.Replace(definitions, header, `define "test.original.elasticsearch-exporter.`+fixture.helper+`.fullname"`, 1)
		definitions += "\n{{- " + header + " -}}explicit-" + strings.ReplaceAll(fixture.helper, ".", "-") + "{{- end -}}\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(definitions), 0600))
	seen := map[string]bool{}
	for _, scenario := range []struct {
		name, fullname string
		values         map[string]string
	}{
		{"service-monitor", "orders-prometheus-elasticsearch-exporter", nil},
		{"pod-monitor", "orders-prometheus-elasticsearch-exporter", map[string]string{
			"serviceMonitor.enabled": "false", "podMonitor.enabled": "true",
		}},
		{"external-account-and-tls", "orders-prometheus-elasticsearch-exporter", map[string]string{
			"serviceAccount.create": "false", "serviceAccount.name": "customer-account",
			"es.ssl.useExistingSecrets": "true",
			"secretMounts[0].name":      "external-tls", "secretMounts[0].secretName": "customer-tls",
			"secretMounts[0].path": "/ssl", "envFromSecret": "customer-environment",
			"extraEnvSecrets.TEST_KEY.secret": "\\{\\{ .Release.Name }}-customer-key", "extraEnvSecrets.TEST_KEY.key": "key",
		}},
		{"disabled-service-and-external-pull", "orders-prometheus-elasticsearch-exporter", map[string]string{
			"service.enabled": "false", "serviceMonitor.enabled": "false",
			"image.pullSecret": "\\{\\{ .Release.Name }}-registry",
		}},
		{"fullname-override", "customer-exporter", map[string]string{"fullnameOverride": "customer-exporter"}},
		{"name-override", "orders-customer", map[string]string{"nameOverride": "customer"}},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", scenario.name, upgrade), func(t *testing.T) {
				values := map[string]string{
					"serviceAccount.create":  "true",
					"serviceMonitor.enabled": "true", "prometheusRule.enabled": "true",
					"podSecurityPolicies.enabled": "true", "es.ssl.enabled": "true",
					"es.ssl.ca.pem": "test-ca", "es.ssl.client.pem": "test-client", "es.ssl.client.key": "test-key",
					"es.uri":                        "https://customer-search:9200",
					"prometheusRule.rules[0].alert": "TestAlert", "prometheusRule.rules[0].expr": "vector(1)",
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
				before := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "orders", options, args...)
				output, err := helm.RenderTemplateE(t, options, chart, "orders", nil, args...)
				require.NoError(t, err)
				renames := map[string]string{}
				for _, fixture := range fixtures {
					renames[fixture.kind+"/"+scenario.fullname+fixture.suffix] = "explicit-" + strings.ReplaceAll(fixture.helper, ".", "-")
				}
				expected := exporterNamingDocuments(t, before, renames, seen)
				actual := exporterNamingDocuments(t, output, nil, nil)
				assert.Equal(t, expected, actual, "Selectors, external accounts/Secrets/URLs and Prometheus rule groups must stay unchanged")
			})
		}
	}
	for _, fixture := range fixtures {
		assert.True(t, seen[fixture.kind+"/orders-prometheus-elasticsearch-exporter"+fixture.suffix], "helper not exercised: %s", fixture.helper)
	}
}

func exporterNamingDocuments(t *testing.T, output string, renames map[string]string, seen map[string]bool) map[string]map[string]interface{} {
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
			key = object["kind"].(string) + "/" + name
		}
		rewriteExporterNamingReferences(object, renames)
		require.NotContains(t, documents, key)
		documents[key] = object
	}
	return documents
}

func rewriteExporterNamingReferences(value interface{}, renames map[string]string) {
	switch node := value.(type) {
	case map[string]interface{}:
		for key, child := range node {
			switch key {
			case "serviceAccountName":
				if name, ok := renames["ServiceAccount/"+child.(string)]; ok {
					node[key] = name
				}
			case "secret", "roleRef":
				ref := child.(map[string]interface{})
				kind, field := "Secret", "secretName"
				if key == "roleRef" {
					kind, field = ref["kind"].(string), "name"
				}
				if name, ok := renames[kind+"/"+ref[field].(string)]; ok {
					ref[field] = name
				}
			case "subjects":
				for _, item := range child.([]interface{}) {
					ref := item.(map[string]interface{})
					if name, ok := renames[ref["kind"].(string)+"/"+ref["name"].(string)]; ok {
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
			}
			rewriteExporterNamingReferences(child, renames)
		}
	case []interface{}:
		for _, child := range node {
			rewriteExporterNamingReferences(child, renames)
		}
	}
}
