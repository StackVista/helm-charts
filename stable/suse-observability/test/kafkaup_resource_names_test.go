package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestKafkaupOperatorResourceNamesInPlatform(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	var definitions string
	for _, resource := range []string{"deployment", "serviceaccount", "role", "rolebinding", "configmap"} {
		definitions += fmt.Sprintf("{{- define \"kafkaup-operator.%s.fullname\" -}}explicit-%s{{- end -}}\n", resource, resource)
	}
	require.NoError(t, os.WriteFile(filepath.Join(chart, "templates", "_kafkaup-resource-names.tpl"), []byte(definitions), 0600))
	values, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, scenario := range []struct {
		release, base string
		set           map[string]string
	}{
		{"suse-observability", "suse-observability-kafkaup-operator", nil},
		{"nightly", "nightly-kafkaup-operator", nil},
		{"nightly", "customer-operator", map[string]string{
			"kafkaup-operator.fullnameOverride":               "customer-operator",
			"kafkaup-operator.kafkaSelectors.statefulSetName": "external-{{ .Release.Name }}-kafka",
		}},
	} {
		for _, enabled := range []bool{false, true} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/enabled=%t/upgrade=%t", scenario.release, scenario.base, enabled, upgrade), func(t *testing.T) {
					set := map[string]string{"kafkaup-operator.enabled": fmt.Sprint(enabled)}
					for key, value := range scenario.set {
						set[key] = value
					}
					options := apiResourceNameTestOptions(set)
					options.ValuesFiles = []string{values}
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					before := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, scenario.release, options, args...)
					output, err := helm.RenderTemplateE(t, options, chart, scenario.release, nil, args...)
					require.NoError(t, err)
					old := scenario.base + "-kafkaup"
					renames := map[string]string{
						"Deployment/" + old:               "explicit-deployment",
						"ServiceAccount/" + old:           "explicit-serviceaccount",
						"Role/" + old:                     "explicit-role",
						"RoleBinding/" + old + "-binding": "explicit-rolebinding",
						"ConfigMap/" + old + "-config":    "explicit-configmap",
					}
					seen := map[string]bool{}
					expected := backupNamingDocuments(t, before, renames, seen)
					actual := backupNamingDocuments(t, output, nil, nil)
					if enabled {
						require.Len(t, seen, len(renames))
						deployment := expected["Deployment/observability/explicit-deployment"].([]interface{})[0].(map[string]interface{})
						pod := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
						pod["spec"].(map[string]interface{})["serviceAccountName"] = "explicit-serviceaccount"
						annotations := pod["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
						actualDeployment := actual["Deployment/observability/explicit-deployment"].([]interface{})[0].(map[string]interface{})
						actualAnnotations := actualDeployment["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
						assert.NotEqual(t, annotations["checksum/configmap"], actualAnnotations["checksum/configmap"])
						annotations["checksum/configmap"] = actualAnnotations["checksum/configmap"]
						binding := expected["RoleBinding//explicit-rolebinding"].([]interface{})[0].(map[string]interface{})
						binding["roleRef"].(map[string]interface{})["name"] = "explicit-role"
						binding["subjects"].([]interface{})[0].(map[string]interface{})["name"] = "explicit-serviceaccount"
					} else {
						assert.Empty(t, seen)
					}
					assert.Equal(t, expected, actual, "Other platform resources, Kafka selection and registry references must remain unchanged")
				})
			}
		}
	}
}
