package test

import (
	"fmt"
	"maps"
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

func TestServiceMonitorNamesUseDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	monitors := map[string]string{
		"api": "api", "authorizationSync": "authorization-sync", "checks": "checks",
		"e2es": "e2es", "healthSync": "health-sync", "initializer": "initializer",
		"notification": "notification", "router": "router", "s3proxy": "s3proxy",
		"server": "server", "slicing": "slicing", "state": "state", "sync": "sync", "ui": "ui",
		"receiver": "receiver", "receiver.base": "receiver-base", "receiver.logs": "receiver-logs",
		"receiver.processAgent": "receiver-process-agent", "correlate": "correlate",
		"correlate.connection": "correlate-connection", "correlate.httpTracing": "correlate-http-tracing",
		"correlate.aggregator": "correlate-aggregator",
	}
	for component, suffix := range monitors {
		helper := "stackstate." + component + ".servicemonitor.fullname"
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAll(data, -1), 1, helper)
		data = definition.ReplaceAll(data, []byte(`{{- define "`+helper+`" -}}explicit-`+suffix+`-monitor{{- end -}}`))
	}
	require.NoError(t, os.WriteFile(namesPath, data, 0600))
	fullValues, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	scenarios := []struct {
		name   string
		values map[string]string
	}{
		{"metrics-disabled", map[string]string{"stackstate.components.all.metrics.enabled": "false"}},
		{"monitors-disabled", map[string]string{"stackstate.components.all.metrics.servicemonitor.enabled": "false"}},
		{"s3proxy-metrics-disabled", map[string]string{"s3proxy.metrics.enabled": "false"}},
		{"s3proxy-monitor-disabled", map[string]string{"s3proxy.metrics.servicemonitor.enabled": "false"}},
		{"all-backups-disabled", map[string]string{"global.backup.enabled": "false", "backup.configuration.enabled": "false"}},
		{"settings-only-backup", map[string]string{"global.backup.enabled": "false", "backup.configuration.enabled": "true"}},
		{"authorization-disabled", map[string]string{"stackstate.k8sAuthorization.enabled": "false"}},
	}
	for _, split := range []bool{false, true} {
		for _, receiver := range []bool{false, true} {
			for _, correlate := range []bool{false, true} {
				scenarios = append(scenarios, struct {
					name   string
					values map[string]string
				}{
					fmt.Sprintf("server=%t/receiver=%t/correlate=%t", split, receiver, correlate),
					map[string]string{"stackstate.features.server.split": fmt.Sprint(split), "stackstate.components.receiver.split.enabled": fmt.Sprint(receiver), "stackstate.components.correlate.split.enabled": fmt.Sprint(correlate)},
				})
			}
		}
	}
	seen := map[string]bool{}
	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			values := componentResourceNameTestValues()
			maps.Copy(values, map[string]string{
				"stackstate.components.all.metrics.enabled":                                    "true",
				"stackstate.components.all.metrics.servicemonitor.enabled":                     "true",
				"stackstate.components.all.metrics.servicemonitor.additionalLabels.prometheus": "platform-monitoring",
				"s3proxy.metrics.enabled":                                                      "true",
				"s3proxy.metrics.servicemonitor.enabled":                                       "true",
				"s3proxy.metrics.servicemonitor.additionalLabels.prometheus":                   "backup-monitoring",
				"stackstate.k8sAuthorization.enabled":                                          "true",
				"global.backup.enabled":                                                        "true",
				"backup.configuration.enabled":                                                 "true",
				"servicemonitor.namespace":                                                     "monitoring",
				"servicemonitor.annotations.fixture-owner":                                     "observability",
				"servicemonitor.labels.fixture-scope":                                          "preserved",
			})
			maps.Copy(values, tc.values)
			options := apiResourceNameTestOptions(values)
			options.ValuesFiles = []string{fullValues}
			before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
			output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
			require.NoError(t, err)
			after := helmtestutil.NewKubernetesResources(t, output)
			for component, suffix := range monitors {
				legacy := "nightly-suse-observability-" + suffix
				if component == "ui" || component == "s3proxy" {
					legacy = "suse-observability-" + suffix
				}
				explicit := "explicit-" + suffix + "-monitor"
				enabled := values["stackstate.components.all.metrics.enabled"] == "true" && values["stackstate.components.all.metrics.servicemonitor.enabled"] == "true"
				split := values["stackstate.features.server.split"] == "true"
				switch component {
				case "api", "checks", "healthSync", "initializer", "notification", "slicing", "state", "sync":
					enabled = enabled && split
				case "authorizationSync":
					enabled = enabled && split && values["stackstate.k8sAuthorization.enabled"] == "true"
				case "server":
					enabled = enabled && !split
				case "s3proxy":
					enabled = values["s3proxy.metrics.enabled"] == "true" && values["s3proxy.metrics.servicemonitor.enabled"] == "true" && (values["global.backup.enabled"] == "true" || values["backup.configuration.enabled"] == "true")
				}
				group, _, variant := strings.Cut(component, ".")
				if group == "receiver" || group == "correlate" {
					enabled = enabled && (variant == (values["stackstate.components."+group+".split.enabled"] == "true"))
				}
				if !enabled {
					assert.NotContains(t, before.ServiceMonitors, legacy)
					assert.NotContains(t, after.ServiceMonitors, explicit)
					continue
				}
				require.Contains(t, before.ServiceMonitors, legacy)
				require.Contains(t, after.ServiceMonitors, explicit)
				expected := before.ServiceMonitors[legacy]
				expected.Name = explicit
				// Full equality preserves Service selectors, namespace selection, ports,
				// paths, intervals, relabelings, labels and annotations.
				assert.Equal(t, expected, after.ServiceMonitors[explicit])
				assert.Equal(t, "monitoring", expected.Namespace)
				assert.Equal(t, []string{"observability"}, expected.Spec.NamespaceSelector.MatchNames)
				assert.Equal(t, "observability", expected.Annotations["fixture-owner"])
				assert.Equal(t, "preserved", expected.Labels["fixture-scope"])
				target := "platform-monitoring"
				if component == "s3proxy" {
					target = "backup-monitoring"
				}
				assert.Equal(t, target, expected.Labels["prometheus"])
				assert.NotContains(t, after.ServiceMonitors, legacy)
				delete(before.ServiceMonitors, legacy)
				delete(after.ServiceMonitors, explicit)
				seen[component] = true
			}
			assert.Equal(t, before.ServiceMonitors, after.ServiceMonitors, "subchart monitors must remain unchanged")
			assert.Equal(t, before.Services, after.Services)
			assert.Equal(t, before.Deployments, after.Deployments)
			assert.Equal(t, before.Statefulsets, after.Statefulsets)
			assert.Equal(t, before.Pdbs, after.Pdbs)
			assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
			assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
			assert.Equal(t, before.Secrets, after.Secrets)
		})
	}
	for component := range monitors {
		assert.True(t, seen[component], "monitor was not exercised: %s", component)
	}
}

func TestServiceMonitorRequiresResolvedVariantName(t *testing.T) {
	fullValues, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, component := range []string{"receiver", "correlate"} {
		t.Run(component, func(t *testing.T) {
			chart := filepath.Join(t.TempDir(), "chart")
			require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
			namesPath := filepath.Join(chart, "templates", "_names.tpl")
			data, err := os.ReadFile(namesPath)
			require.NoError(t, err)
			helper := "stackstate." + component + ".servicemonitor.fullname"
			definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
			require.Len(t, definition.FindAll(data, -1), 1)
			data = definition.ReplaceAll(data, []byte(`{{- define "`+helper+`" -}}{{- end -}}`))
			require.NoError(t, os.WriteFile(namesPath, data, 0600))
			options := apiResourceNameTestOptions(map[string]string{"stackstate.components.all.metrics.enabled": "true", "stackstate.components.all.metrics.servicemonitor.enabled": "true", "stackstate.components." + component + ".split.enabled": "false"})
			options.ValuesFiles = []string{fullValues}
			_, err = helm.RenderTemplateE(t, options, chart, "nightly", nil)
			require.ErrorContains(t, err, "stackstate."+component+".servicemonitor: ServiceMonitorFullname must not be empty")
		})
	}
}
