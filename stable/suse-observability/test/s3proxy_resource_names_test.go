package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Give each managed resource a distinct identity in a copied chart. PVC renames
// here only test wiring; real installations must retain their existing claims.
func TestS3ProxyResourcesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	renames := map[string]string{}
	for _, resource := range []struct{ helper, kind, old, replacement string }{
		{"deployment", "Deployment", "suse-observability-s3proxy", "explicit-s3proxy-deployment"},
		{"service", "Service", "suse-observability-s3proxy", "explicit-s3proxy-service"},
		{"secret", "Secret", "suse-observability-s3proxy", "explicit-s3proxy-credentials"},
		{"serviceaccount", "ServiceAccount", "suse-observability-s3proxy", "explicit-s3proxy-account"},
		{"configmap", "ConfigMap", "suse-observability-s3proxy-config", "explicit-s3proxy-config"},
		{"extraEnvSecret", "Secret", "suse-observability-s3proxy-extra-env", "explicit-s3proxy-env"},
		{"settings.persistentvolumeclaim", "PersistentVolumeClaim", "nightly-suse-observability-backup-settings-data", "explicit-s3proxy-settings"},
		{"main.persistentvolumeclaim", "PersistentVolumeClaim", "suse-observability-minio", "explicit-s3proxy-main"},
	} {
		helper := "stackstate.s3proxy." + resource.helper + ".fullname"
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAll(data, -1), 1, helper)
		data = definition.ReplaceAll(data, []byte(`{{- define "`+helper+`" -}}`+resource.replacement+`{{- end -}}`))
		renames[resource.kind+"/"+resource.old] = resource.replacement
	}
	require.NoError(t, os.WriteFile(namesPath, data, 0600))
	fullValues, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	seen := map[string]bool{}
	for _, tc := range []struct {
		name   string
		values map[string]string
	}{
		{"pvc", nil},
		{"settings-only", map[string]string{"global.backup.enabled": "false"}},
		{"disabled", map[string]string{"global.backup.enabled": "false", "backup.configuration.enabled": "false"}},
		{"external-credentials", map[string]string{"global.s3proxy.credentials.fromExternalSecret": "customer-proxy-credentials"}},
		{"s3", map[string]string{
			"backup.storage.backend.pvc.enabled": "false", "backup.storage.backend.s3.enabled": "true",
			"backup.storage.backend.s3.accessKey": "fixture-access", "backup.storage.backend.s3.secretKey": "fixture-secret",
		}},
		{"s3-external", map[string]string{
			"backup.storage.backend.pvc.enabled": "false", "backup.storage.backend.s3.enabled": "true",
			"backup.storage.backend.s3.fromExternalSecret":  "customer-s3-backend",
			"global.s3proxy.credentials.fromExternalSecret": "customer-proxy-credentials",
		}},
		{"azure", map[string]string{
			"backup.storage.backend.pvc.enabled": "false", "backup.storage.backend.azure.enabled": "true",
			"backup.storage.backend.azure.accountName": "fixtureaccount", "backup.storage.backend.azure.accountKey": "fixture-key",
		}},
		{"azure-external", map[string]string{
			"backup.storage.backend.pvc.enabled": "false", "backup.storage.backend.azure.enabled": "true",
			"backup.storage.backend.azure.fromExternalSecret": "customer-azure-backend",
			"global.s3proxy.credentials.fromExternalSecret":   "customer-proxy-credentials",
		}},
		{"legacy-s3", map[string]string{
			"backup.storage.backend.pvc.enabled": "false", "minio.s3gateway.enabled": "true",
			"minio.s3gateway.accessKey": "legacy-access", "minio.s3gateway.secretKey": "legacy-secret",
		}},
		{"legacy-azure", map[string]string{
			"backup.storage.backend.pvc.enabled": "false", "minio.azuregateway.enabled": "true",
			"minio.accessKey": "legacyaccount", "minio.secretKey": "legacy-key",
		}},
		{"custom-account", map[string]string{"s3proxy.serviceAccount.name": "customer-account"}},
		{"external-account", map[string]string{"s3proxy.serviceAccount.name": "customer-account", "s3proxy.serviceAccount.create": "false"}},
		{"legacy-account", map[string]string{"minio.serviceAccount.name": "legacy-account"}},
		{"legacy-external-account", map[string]string{"minio.serviceAccount.name": "legacy-account", "minio.serviceAccount.create": "false"}},
		{"account-precedence", map[string]string{
			"s3proxy.serviceAccount.name": "customer-account", "minio.serviceAccount.name": "legacy-account",
			"minio.serviceAccount.annotations.fixture": "legacy", "s3proxy.serviceAccount.annotations.fixture": "current",
		}},
	} {
		for _, extraEnv := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/extra-env=%t", tc.name, extraEnv), func(t *testing.T) {
				values := map[string]string{
					"global.backup.enabled": "true", "backup.configuration.enabled": "true",
					"backup.storage.settingsPvc.size": "9Gi", "backup.storage.settingsPvc.storageClass": "retained-settings",
					"backup.storage.backend.pvc.size": "13Gi", "backup.storage.backend.pvc.storageClass": "retained-main",
				}
				for key, value := range tc.values {
					values[key] = value
				}
				if extraEnv {
					values["s3proxy.extraEnv.secret.FIXTURE_SECRET"] = "retained-value"
				}
				options := apiResourceNameTestOptions(values)
				options.ValuesFiles = []string{fullValues}
				before := helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options)
				after, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
				require.NoError(t, err)
				expected := backupNamingDocuments(t, before, renames, seen)
				actual := backupNamingDocuments(t, after, nil, nil)
				rewriteS3ProxyReferences(expected, renames)
				assert.Len(t, actual, len(expected))
				for key, objects := range expected {
					require.Contains(t, actual, key)
					require.Len(t, actual[key], len(objects.([]interface{})))
					for i, object := range objects.([]interface{}) {
						fixS3ProxyFixtureChecksums(t, object.(map[string]interface{}), actual[key].([]interface{})[i].(map[string]interface{}))
					}
					// Includes Secret contents/external keys, account annotations,
					// PVC specs/keep policy, selectors, subcharts and embedded Jobs.
					assert.Equal(t, objects, actual[key], key)
				}
			})
		}
	}
	for key := range renames {
		assert.True(t, seen[key], "resource was not exercised: %s", key)
	}
}

func rewriteS3ProxyReferences(value interface{}, renames map[string]string) {
	switch node := value.(type) {
	case map[string]interface{}:
		for key, value := range node {
			switch key {
			case "secretName", "serviceAccountName":
				kind := "Secret"
				if key == "serviceAccountName" {
					kind = "ServiceAccount"
				}
				if name, ok := value.(string); ok {
					if replacement, ok := renames[kind+"/"+name]; ok {
						node[key] = replacement
					}
				}
			case "secretKeyRef":
				if ref, ok := value.(map[string]interface{}); ok {
					if name, ok := ref["name"].(string); ok {
						if replacement, ok := renames["Secret/"+name]; ok {
							ref["name"] = replacement
						}
					}
				}
			case "service":
				// The backup CLI's storage.service is a port-forward target.
				if ref, ok := value.(map[string]interface{}); ok && ref["name"] == "suse-observability-s3proxy" {
					ref["name"] = renames["Service/suse-observability-s3proxy"]
				}
			}
			if text, ok := node[key].(string); ok {
				node[key] = strings.ReplaceAll(text, "suse-observability-s3proxy:9000", renames["Service/suse-observability-s3proxy"]+":9000")
			}
			rewriteS3ProxyReferences(node[key], renames)
		}
	case []interface{}:
		for i, child := range node {
			if text, ok := child.(string); ok {
				node[i] = strings.ReplaceAll(text, "suse-observability-s3proxy:9000", renames["Service/suse-observability-s3proxy"]+":9000")
			}
			rewriteS3ProxyReferences(child, renames)
		}
	}
}

func fixS3ProxyFixtureChecksums(t *testing.T, expected, actual map[string]interface{}) {
	t.Helper()
	object := unstructured.Unstructured{Object: expected}
	var keys []string
	if object.GetKind() == "Deployment" && object.GetName() == "explicit-s3proxy-deployment" {
		keys = []string{"checksum/config", "checksum/secret", "checksum/extra-env-secret"}
	} else if object.GetKind() == "StatefulSet" && (strings.HasPrefix(object.GetName(), "suse-observability-clickhouse-shard") ||
		object.GetName() == "suse-observability-victoria-metrics-0" || object.GetName() == "suse-observability-victoria-metrics-1") {
		keys = []string{"checksum/backup-config"}
	}
	for _, key := range keys {
		path := []string{"spec", "template", "metadata", "annotations", key}
		old, exists, err := unstructured.NestedString(expected, path...)
		require.NoError(t, err)
		if !exists {
			continue
		}
		replacement, exists, err := unstructured.NestedString(actual, path...)
		require.NoError(t, err)
		require.True(t, exists)
		require.NotEmpty(t, replacement)
		if key != "checksum/secret" {
			assert.NotEqual(t, old, replacement, "the fixture rename must affect the existing %s annotation", key)
		} // An external credential Secret leaves the empty-template checksum unchanged.
		require.NoError(t, unstructured.SetNestedField(expected, replacement, path...))
	}
}
