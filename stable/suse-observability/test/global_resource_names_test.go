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
)

func TestGlobalResourcesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	renames := map[string]string{}
	for _, resource := range []struct{ helper, kind, old, name string }{
		{"stackstate.ingress.fullname", "Ingress", "nightly-suse-observability", "explicit-ingress"},
		{"stackstate.securitycontextconstraints.fullname", "SecurityContextConstraints", "nightly-suse-observability-observability", "explicit-scc"},
		{"stackstate.victoriametrics.service.fullname", "Service", "suse-observability-victoriametrics", "explicit-metrics-service"},
		{"stackstate.kafkaTopicCreate.job.fullname", "Job", "nightly-suse-observability-topic-create-<timestamp>", "explicit-topic-job"},
		{"stackstate.kafkaTopicCreate.job.generateName", "Job", "topic-create-", "explicit-topic-generated-"},
		{"suse-observability.pullSecret.hook.fullname", "Secret", "suse-observability-pull-secret-hook", "explicit-hook-pull-secret"},
		{"suse-observability.pullSecret.fullname", "Secret", "suse-observability-pull-secret", "explicit-normal-pull-secret"},
	} {
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(resource.helper) + `" -\}\}.*?\{\{- end -\}\}`)
		replacement := []byte(`{{- define "` + resource.helper + `" -}}` + resource.name + `{{- end -}}`)
		if resource.helper == "suse-observability.pullSecret.fullname" {
			// Override common in the parent. Bundled consumers must use this
			// shared interface, including subcharts without a common dependency.
			require.Empty(t, definition.FindAll(data, -1))
			data = append(data, append(replacement, '\n')...)
		} else {
			require.Len(t, definition.FindAll(data, -1), 1, resource.helper)
			data = definition.ReplaceAll(data, replacement)
		}
		renames[resource.kind+"/"+resource.old] = resource.name
	}
	require.NoError(t, os.WriteFile(path, data, 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	seen := map[string]bool{}
	for _, argo := range []bool{false, true} {
		for _, split := range []bool{false, true} {
			for _, pull := range []string{"automatic", "external", "legacy"} {
				for _, upgrade := range []bool{false, true} {
					t.Run(fmt.Sprintf("argo=%t/split=%t/pull=%s/upgrade=%t", argo, split, pull, upgrade), func(t *testing.T) {
						values := map[string]string{
							"deployment.compatibleWithArgoCD":          fmt.Sprint(argo),
							"stackstate.features.server.split":         fmt.Sprint(split),
							"stackstate.components.router.mode.status": "automatic",
							"anomaly-detection.enabled":                "true",
							"ingress.enabled":                          "true", "scc.enabled": "true",
						}
						if pull == "automatic" {
							values["pull-secret.enabled"] = "false"
							values["global.suseObservability.pullSecret.username"] = "fixture-user"
							values["global.suseObservability.pullSecret.password"] = "fixture-password"
							values["global.suseObservability.adminPasswordBcrypt"] = "$2b$10$N9qo8uLOickgx2ZMRZoMye.IUIrCxGvz9/6pN.XMlqL9hzTgDaPGy"
							values["global.suseObservability.license"] = "api01"
							values["global.suseObservability.baseUrl"] = "http://localhost"
						} else if pull == "external" {
							values["global.imagePullSecrets[0]"] = "customer-registry-secret"
						}
						options := apiResourceNameTestOptions(values)
						options.ValuesFiles = []string{valuesFile}
						var args []string
						if upgrade {
							args = append(args, "--is-upgrade")
						}
						before := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "nightly", options, args...)
						after, err := helm.RenderTemplateE(t, options, chart, "nightly", nil, args...)
						require.NoError(t, err)
						expected := backupNamingDocuments(t, before, renames, seen)
						actual := backupNamingDocuments(t, after, nil, nil)
						rewriteGlobalNameReferences(expected, renames)
						assert.Len(t, actual, len(expected))
						for key, objects := range expected {
							require.Contains(t, actual, key)
							assert.Equal(t, objects, actual[key], key)
						}
						// Full comparisons preserve hook lifecycle/credentials, SCC
						// groups, Service selectors, storage and existing checksums.
					})
				}
			}
		}
	}
	for key := range renames {
		assert.True(t, seen[key], "resource was not exercised: %s", key)
	}
}

func rewriteGlobalNameReferences(value interface{}, renames map[string]string) {
	switch node := value.(type) {
	case map[string]interface{}:
		if refs, ok := node["imagePullSecrets"].([]interface{}); ok {
			for _, entry := range refs {
				if ref, ok := entry.(map[string]interface{}); ok {
					if name, ok := ref["name"].(string); ok {
						if replacement, ok := renames["Secret/"+name]; ok {
							ref["name"] = replacement
						}
					}
				}
			}
		}
		for key, child := range node {
			if text, ok := child.(string); ok {
				node[key] = strings.ReplaceAll(text, "http://suse-observability-victoriametrics:8428", "http://explicit-metrics-service:8428")
			}
			rewriteGlobalNameReferences(child, renames)
		}
	case []interface{}:
		for _, child := range node {
			rewriteGlobalNameReferences(child, renames)
		}
	}
}
