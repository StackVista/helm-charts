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
)

func TestCollectorServiceReferencesFollowDedicatedHelper(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates", "_names.tpl")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	require.NoError(t, err)
	_, err = file.WriteString(`
{{- define "opentelemetry-collector.service.fullname" -}}explicit-collector-service{{- end -}}
`)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	values, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, mode := range []string{"deployment", "daemonset", "statefulset"} {
		for _, collector := range []string{"suse-observability-otel-collector", "customer-collector"} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/upgrade=%t", mode, collector, upgrade), func(t *testing.T) {
					options := apiResourceNameTestOptions(map[string]string{
						"opentelemetry-collector.mode":             mode,
						"opentelemetry-collector.fullnameOverride": collector,
						"opentelemetry-collector.service.enabled":  "true",
					})
					options.ValuesFiles = []string{values}
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					before := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "nightly", options, args...)
					after, err := helm.RenderTemplateE(t, options, chart, "nightly", nil, args...)
					require.NoError(t, err)
					seen := map[string]bool{}
					expected := backupNamingDocuments(t, before, map[string]string{
						"Service/" + collector: "explicit-collector-service",
					}, seen)
					require.True(t, seen["Service/"+collector])
					rewriteCollectorServiceReferences(expected, collector)
					assert.Equal(t, expected, backupNamingDocuments(t, after, nil, nil))
				})
			}
		}
	}
}

func rewriteCollectorServiceReferences(value interface{}, old string) {
	switch node := value.(type) {
	case map[string]interface{}:
		for key, child := range node {
			if key == "serviceName" && child == old {
				node[key] = "explicit-collector-service"
			} else if key == "clusters.yaml" {
				// Envoy identifiers retain the old name; only the DNS target changes.
				node[key] = strings.ReplaceAll(child.(string),
					`address: "`+old+`"`, `address: "explicit-collector-service"`)
			}
			rewriteCollectorServiceReferences(child, old)
		}
	case []interface{}:
		for _, child := range node {
			rewriteCollectorServiceReferences(child, old)
		}
	}
}
