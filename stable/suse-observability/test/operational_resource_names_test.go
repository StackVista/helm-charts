package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// Artificially rename each resource independently to detect references still
// using a shared helper. These fixture-only StatefulSet renames are not a
// migration procedure for the persistent claims of an installed release.
func TestOperationalResourcesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	renames := map[string]string{}
	for _, component := range []struct {
		key, suffix string
		resources   map[string]string
	}{
		{"replicationChecker", "replication-checker", map[string]string{
			"deployment": "Deployment", "serviceaccount": "ServiceAccount", "role": "Role", "rolebinding": "RoleBinding",
		}},
		{"workloadObserver", "workload-observer", map[string]string{
			"statefulset": "StatefulSet", "serviceaccount": "ServiceAccount", "role": "Role", "rolebinding": "RoleBinding",
		}},
		{"vmagent", "vmagent", map[string]string{
			"statefulset": "StatefulSet", "configmap": "ConfigMap", "service": "Service",
		}},
	} {
		for resource, kind := range component.resources {
			helper := "stackstate." + component.key + "." + resource + ".fullname"
			name := "explicit-" + component.suffix + "-" + resource
			definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
			require.Len(t, definition.FindAll(data, -1), 1, helper)
			data = definition.ReplaceAll(data, []byte(`{{- define "`+helper+`" -}}`+name+`{{- end -}}`))
			renames[kind+"/suse-observability-"+component.suffix] = name
		}
	}
	require.NoError(t, os.WriteFile(namesPath, data, 0600))
	fullValues, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	seen := map[string]bool{}
	for _, split := range []bool{false, true} {
		for _, workersSplit := range []bool{false, true} {
			for _, backups := range []bool{false, true} {
				for _, enabled := range []bool{false, true} {
					t.Run(fmt.Sprintf("server=%t/workers=%t/backups=%t/controllers=%t", split, workersSplit, backups, enabled), func(t *testing.T) {
						values := componentResourceNameTestValues()
						values["stackstate.features.server.split"] = fmt.Sprint(split)
						values["stackstate.components.receiver.split.enabled"] = fmt.Sprint(workersSplit)
						values["stackstate.components.correlate.split.enabled"] = fmt.Sprint(workersSplit)
						values["global.backup.enabled"] = fmt.Sprint(backups)
						values["backup.configuration.enabled"] = fmt.Sprint(backups)
						values["stackstate.components.replicationChecker.enabled"] = fmt.Sprint(enabled)
						values["stackstate.components.workloadObserver.enabled"] = fmt.Sprint(enabled)
						for _, component := range []string{"vmagent", "workloadObserver"} {
							values["stackstate.components."+component+".persistence.size"] = "9Gi"
							values["stackstate.components."+component+".persistence.storageClass"] = "retained-storage"
						}
						options := apiResourceNameTestOptions(values)
						options.ValuesFiles = []string{fullValues}
						before := helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options)
						after, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
						require.NoError(t, err)
						expected := backupNamingDocuments(t, before, renames, seen)
						actual := backupNamingDocuments(t, after, nil, nil)
						resources := helmtestutil.NewKubernetesResources(t, after)
						_, checkerExists := resources.Deployments["explicit-replication-checker-deployment"]
						assert.Equal(t, enabled, checkerExists)
						_, observerExists := resources.Statefulsets["explicit-workload-observer-statefulset"]
						assert.Equal(t, enabled, observerExists)
						require.Contains(t, resources.Statefulsets, "explicit-vmagent-statefulset")
						vmagent := resources.Statefulsets["explicit-vmagent-statefulset"]
						checksum := vmagent.Spec.Template.Annotations["checksum/vmagent-configmap"]
						require.NotEmpty(t, checksum)
						rewriteOperationalReferences(t, expected, renames, checksum)
						// Compare entire manifests, including RBAC rules/subjects, pod
						// selectors, claim templates/mounts, inherited serviceName fields,
						// and the backup restore Jobs embedded in ConfigMaps.
						assert.Equal(t, expected, actual)
					})
				}
			}
		}
	}
	for key := range renames {
		assert.True(t, seen[key], "resource was not exercised: %s", key)
	}
}

func rewriteOperationalReferences(t *testing.T, value interface{}, renames map[string]string, checksum string) {
	t.Helper()
	switch node := value.(type) {
	case map[string]interface{}:
		// ServiceAccount subjects and roleRef entries carry kind and name
		// directly; preserve their namespaces and all remaining RBAC fields.
		if kind, ok := node["kind"].(string); ok && (kind == "ServiceAccount" || kind == "Role") {
			if name, ok := node["name"].(string); ok {
				if replacement, ok := renames[kind+"/"+name]; ok {
					node["name"] = replacement
				}
			}
		}
		if name, ok := node["serviceAccountName"].(string); ok {
			if replacement, ok := renames["ServiceAccount/"+name]; ok {
				node["serviceAccountName"] = replacement
			}
		}
		if node["name"] == "PROMETHEUS_WRITE_ENDPOINT" && node["value"] == "http://suse-observability-vmagent:8429/api/v1/write" {
			node["value"] = "http://" + renames["Service/suse-observability-vmagent"] + ":8429/api/v1/write"
		}
		if old, ok := node["checksum/vmagent-configmap"]; ok {
			assert.NotEqual(t, old, checksum, "the fixture-only ConfigMap rename must affect its existing checksum")
			node["checksum/vmagent-configmap"] = checksum
		}
		for _, child := range node {
			rewriteOperationalReferences(t, child, renames, checksum)
		}
	case []interface{}:
		for _, child := range node {
			rewriteOperationalReferences(t, child, renames, checksum)
		}
	}
}
