package test

import (
	"fmt"
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
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", monitor.kind, upgrade), func(t *testing.T) {
				seen := map[string]bool{}
				for _, tc := range []struct{ release, namespace, monitoringNamespace, fullname string }{
					{namespace: "suse-observability", fullname: "existing-exporter"},
					{namespace: "tenant-a", fullname: "existing-exporter"},
					{namespace: "tenant-b", fullname: "existing-exporter", monitoringNamespace: "tenant-b"},
					{namespace: strings.Repeat("n", 62) + "a", fullname: "existing-exporter"},
					{namespace: strings.Repeat("n", 62) + "b", fullname: "existing-exporter"},
					{namespace: "tenant-a", fullname: "existing-exporter", monitoringNamespace: "monitoring"},
					{release: "installation-a", namespace: "tenant-a", monitoringNamespace: "monitoring"},
					{release: "installation-b", namespace: "tenant-b", monitoringNamespace: "monitoring"},
					{release: "custom-a", namespace: "tenant-a", monitoringNamespace: "monitoring", fullname: "customer-exporter-a"},
					{release: "custom-b", namespace: "tenant-b", monitoringNamespace: "monitoring", fullname: "customer-exporter-b"},
				} {
					release := tc.release
					if release == "" {
						release = "nightly"
					}
					legacyName := tc.fullname
					if legacyName == "" {
						legacyName = release + "-prometheus-elasticsearch-exporter"
					}
					namespace := tc.namespace
					objectNamespace := namespace
					if tc.monitoringNamespace != "" {
						objectNamespace = tc.monitoringNamespace
					}
					values := map[string]string{
						"fullnameOverride":       tc.fullname,
						"serviceMonitor.enabled": "false", "podMonitor.enabled": "false",
						monitor.value + ".enabled": "true", monitor.value + ".namespace": tc.monitoringNamespace,
						"prometheusRule.enabled": "true", "prometheusRule.namespace": tc.monitoringNamespace,
						"prometheusRule.rules[0].alert": "FixtureAlert", "prometheusRule.rules[0].expr": "vector(1)",
					}
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, release, &helm.Options{
						SetValues: values, KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
					}, args...)
					documents := exporterNamingDocuments(t, output, nil, nil)
					name := "suse-observability-prometheus-elasticsearch-exporter"
					if objectNamespace != namespace {
						name = legacyName
					}
					assert.Empty(t, validation.IsDNS1123Subdomain(name))
					for _, kind := range []string{monitor.kind, "PrometheusRule"} {
						require.Contains(t, documents, kind+"/"+name)
						object := documents[kind+"/"+name]
						metadata := object["metadata"].(map[string]interface{})
						if tc.monitoringNamespace == "" {
							assert.NotContains(t, metadata, "namespace", "An unset namespace must use the release namespace")
						} else {
							assert.Equal(t, objectNamespace, metadata["namespace"])
						}
						assert.Equal(t, release, metadata["labels"].(map[string]interface{})["release"])
						key := kind + "/" + objectNamespace + "/" + name
						assert.False(t, seen[key], "Installations must not compete for the same monitoring object")
						seen[key] = true
						spec := object["spec"].(map[string]interface{})
						if kind != "PrometheusRule" {
							assert.Equal(t, []interface{}{namespace}, spec["namespaceSelector"].(map[string]interface{})["matchNames"])
							assert.Equal(t, release, spec["selector"].(map[string]interface{})["matchLabels"].(map[string]interface{})["release"])
						}
					}
					require.Contains(t, documents, "Service/"+legacyName, "Service DNS must retain the existing name")
				}
				assert.Len(t, seen, 20)
			})
		}
	}
}
