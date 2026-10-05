package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestElasticsearchHeadlessServiceReferencesFollowDedicatedHelper(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	require.NoError(t, os.WriteFile(filepath.Join(chart, "templates", "_elasticsearch-resource-names.tpl"), []byte(
		`{{- define "elasticsearch.headless.service.fullname" -}}explicit-search-headless{{- end -}}`), 0600))
	values, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, scenario := range []struct {
		name string
		set  map[string]string
	}{
		{"default", nil},
		{"custom-fullname", map[string]string{"elasticsearch.fullnameOverride": "customer-search"}},
		{"custom-master-service", map[string]string{"elasticsearch.masterService": "customer-master"}},
		{"data-group", map[string]string{"elasticsearch.nodeGroup": "data"}},
		{"custom-cluster", map[string]string{"elasticsearch.clusterName": "customer"}},
	} {
		for _, split := range []bool{false, true} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/split=%t/upgrade=%t", scenario.name, split, upgrade), func(t *testing.T) {
					set := map[string]string{"stackstate.features.server.split": fmt.Sprint(split)}
					for key, value := range scenario.set {
						set[key] = value
					}
					options := apiResourceNameTestOptions(set)
					options.ValuesFiles = []string{values}
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					before := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "nightly", options, args...)
					output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil, args...)
					require.NoError(t, err)
					old := "suse-observability-elasticsearch-master-headless"
					if scenario.name == "custom-fullname" {
						old = "customer-search-headless"
					} else if scenario.name == "custom-master-service" {
						old = "customer-master-headless"
					} else if scenario.name == "data-group" {
						old = "suse-observability-elasticsearch-data-headless"
					} else if scenario.name == "custom-cluster" {
						old = "customer-master-headless"
					}
					renames := map[string]string{"Service/" + old: "explicit-search-headless"}
					seen := map[string]bool{}
					expected := backupNamingDocuments(t, before, renames, seen)
					require.True(t, seen["Service/"+old])
					// Existing mismatches stay unchanged because their legacy
					// endpoint differs from the Service being deliberately renamed.
					rewritePlatformElasticsearchHeadless(expected, old)
					assert.Equal(t, expected, backupNamingDocuments(t, output, nil, nil))
				})
			}
		}
	}
}

func rewritePlatformElasticsearchHeadless(value interface{}, old string) {
	switch node := value.(type) {
	case map[string]interface{}:
		for key, child := range node {
			if text, ok := child.(string); ok {
				node[key] = strings.ReplaceAll(text, old, "explicit-search-headless")
			} else {
				rewritePlatformElasticsearchHeadless(child, old)
			}
		}
	case []interface{}:
		for i, child := range node {
			if text, ok := child.(string); ok {
				node[i] = strings.ReplaceAll(text, old, "explicit-search-headless")
			} else {
				rewritePlatformElasticsearchHeadless(child, old)
			}
		}
	}
}

func TestElasticsearchCertificateLookupUsesDedicatedHelper(t *testing.T) {
	for _, renamed := range []bool{false, true} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("renamed=%t/upgrade=%t", renamed, upgrade), func(t *testing.T) {
				chart := filepath.Join(t.TempDir(), "chart")
				require.NoError(t, os.CopyFS(chart, os.DirFS("../../../local/elasticsearch")))
				// Keep the real certificate template/helpers and dependencies.
				// The minimal mock API only supports Secret reads; full workload
				// manifests are exercised by the reference tests above.
				templates, err := filepath.Glob(filepath.Join(chart, "templates", "*.yaml"))
				require.NoError(t, err)
				for _, path := range templates {
					if filepath.Base(path) != "secret-cert.yaml" {
						require.NoError(t, os.Remove(path))
					}
				}
				require.NoError(t, os.RemoveAll(filepath.Join(chart, "templates", "test")))
				name := "suse-observability-elasticsearch-master-certs"
				if renamed {
					name = "explicit-search-certificates"
					path := filepath.Join(chart, "templates", "_names.tpl")
					data, err := os.ReadFile(path)
					require.NoError(t, err)
					header := `define "elasticsearch.certificates.secret.fullname"`
					require.Equal(t, 1, strings.Count(string(data), header))
					content := strings.Replace(string(data), header, `define "test.original.elasticsearch.certificates.secret.fullname"`, 1)
					content += "\n{{- " + header + " -}}" + name + "{{- end -}}\n"
					require.NoError(t, os.WriteFile(path, []byte(content), 0600))
				}
				existing := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
					ObjectMeta: metav1.ObjectMeta{
						Name: name, Namespace: "observability",
						Labels: map[string]string{"app.kubernetes.io/managed-by": "Helm"},
						Annotations: map[string]string{
							"meta.helm.sh/release-name": "nightly", "meta.helm.sh/release-namespace": "observability",
						},
					},
					Data: map[string][]byte{
						"tls.crt": []byte("existing-certificate"),
						"tls.key": []byte("existing-private-key"),
						"ca.crt":  []byte("existing-ca"),
					},
				}
				secretPath := "/api/v1/namespaces/observability/secrets/" + name
				kubeconfig, requests := secretLookupTestAPI(t, secretPath, existing)
				options := apiResourceNameTestOptions(map[string]string{
					"createCert": "true", "prometheus-elasticsearch-exporter.enabled": "false",
				})
				options.ValuesFiles = nil
				args := []string{"--dry-run=server", "--disable-openapi-validation", "--kubeconfig", kubeconfig, "--kube-context", "fixture", "--namespace", "observability"}
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				output, err := helm.RenderTemplateE(t, options, chart, "nightly", []string{"templates/secret-cert.yaml"}, args...)
				require.NoError(t, err)
				resources := helmtestutil.NewKubernetesResources(t, output)
				require.Contains(t, resources.Secrets, name)
				assert.Equal(t, existing.Data, resources.Secrets[name].Data)
				assert.Contains(t, requests(), secretPath, "The declaration and real lookup must share the same name")
			})
		}
	}
}
