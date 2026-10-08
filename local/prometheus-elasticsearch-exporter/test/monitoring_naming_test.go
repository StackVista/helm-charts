package test

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"k8s.io/apimachinery/pkg/util/validation"
)

func TestExporterMonitoringNamespacesPreserveIsolation(t *testing.T) {
	for _, monitor := range []struct{ value, kind string }{
		{"serviceMonitor", "ServiceMonitor"}, {"podMonitor", "PodMonitor"},
	} {
		t.Run(monitor.kind, func(t *testing.T) {
			seen := map[string]bool{}
			for _, tc := range []struct{ namespace, monitoringNamespace string }{
				{namespace: "suse-observability"},
				{namespace: "tenant-a"},
				{namespace: "tenant-b"},
				{namespace: strings.Repeat("n", 62) + "a"},
				{namespace: strings.Repeat("n", 62) + "b"},
				{namespace: "tenant-a", monitoringNamespace: "monitoring"},
			} {
				namespace := tc.namespace
				objectNamespace := namespace
				if tc.monitoringNamespace != "" {
					objectNamespace = tc.monitoringNamespace
				}
				values := map[string]string{
					"fullnameOverride":       "existing-exporter",
					"serviceMonitor.enabled": "false", "podMonitor.enabled": "false",
					monitor.value + ".enabled": "true", monitor.value + ".namespace": objectNamespace,
					"prometheusRule.enabled": "true", "prometheusRule.namespace": objectNamespace,
					"prometheusRule.rules[0].alert": "FixtureAlert", "prometheusRule.rules[0].expr": "vector(1)",
				}
				output := helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", &helm.Options{
					SetValues: values, KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
				})
				documents := exporterNamingDocuments(t, output, nil, nil)
				name := "suse-observability-prometheus-elasticsearch-exporter"
				assert.Empty(t, validation.IsDNS1123Subdomain(name))
				for _, kind := range []string{monitor.kind, "PrometheusRule"} {
					require.Contains(t, documents, kind+"/"+name)
					object := documents[kind+"/"+name]
					metadata := object["metadata"].(map[string]interface{})
					assert.Equal(t, objectNamespace, metadata["namespace"])
					assert.Equal(t, "nightly", metadata["labels"].(map[string]interface{})["release"])
					key := kind + "/" + objectNamespace + "/" + name
					assert.False(t, seen[key], "Installations must not compete for the same monitoring object")
					seen[key] = true
					spec := object["spec"].(map[string]interface{})
					if kind != "PrometheusRule" {
						assert.Equal(t, []interface{}{namespace}, spec["namespaceSelector"].(map[string]interface{})["matchNames"])
						assert.Equal(t, "nightly", spec["selector"].(map[string]interface{})["matchLabels"].(map[string]interface{})["release"])
					}
				}
				require.Contains(t, documents, "Service/existing-exporter", "Service DNS must retain the existing name")
			}
			assert.Len(t, seen, 12)
		})
	}
}
