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
	corev1 "k8s.io/api/core/v1"
)

func TestRouterServiceNamePreservesGlobalNaming(t *testing.T) {
	chart, valuesFile := routerServiceNameTestChart(t, false)
	for _, tc := range []struct {
		name, release, prefix string
		values                map[string]string
	}{
		{"default", "suse-observability", "suse-observability", nil},
		{"custom-release", "nightly", "nightly-suse-observability", nil},
		{"local-overrides", "nightly", "nightly-suse-observability", map[string]string{
			"fullnameOverride": "local-platform", "fullnamePrefix": "local-", "fullnameSuffix": "-local",
			"kubernetes-rbac-agent.fullnameOverride": "local-agent", "anomaly-detection.fullnameOverride": "local-anomaly",
		}},
		{"global-override", "nightly", "shared", map[string]string{"global.fullnameOverride": "SHARED"}},
		{"global-affixes", "nightly", "global-nightly-suse-observability-end", map[string]string{
			"global.fullnamePrefix": "GLOBAL-", "global.fullnameSuffix": "-END",
		}},
		{"combined-overrides", "nightly", "global-shared-end", map[string]string{
			"global.fullnameOverride": "SHARED", "global.fullnamePrefix": "GLOBAL-", "global.fullnameSuffix": "-END",
			"fullnameOverride": "local-platform", "fullnamePrefix": "local-", "fullnameSuffix": "-local",
		}},
		{"truncated", "nightly", strings.Repeat("a", 54), map[string]string{"global.fullnameOverride": strings.Repeat("A", 70)}},
		{"trim-truncated-dash", "nightly", strings.Repeat("a", 53), map[string]string{"global.fullnameOverride": strings.Repeat("a", 53) + "-tail"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{"anomaly-detection.enabled": "true"}
			maps.Copy(values, tc.values)
			options := apiResourceNameTestOptions(values)
			options.ValuesFiles = []string{valuesFile}
			output, err := helm.RenderTemplateE(t, options, chart, tc.release, nil)
			require.NoError(t, err)
			resources := helmtestutil.NewKubernetesResources(t, output)
			name := tc.prefix + "-router"
			require.Contains(t, resources.Services, name)
			require.Contains(t, resources.ConfigMaps, "router-name-probe")
			assert.Equal(t, map[string]string{
				"legacyName": name, "legacyPrefix": tc.prefix,
			}, resources.ConfigMaps["router-name-probe"].Data)
			assertRouterServiceClientEndpoints(t, resources, name, false)
		})
	}
}

func TestRouterServiceReferencesFollowDedicatedHelper(t *testing.T) {
	chart, valuesFile := routerServiceNameTestChart(t, true)
	for _, edge := range []string{"ingress", "gateway"} {
		for _, split := range []bool{false, true} {
			for _, externalURL := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/split=%t/external-url=%t", edge, split, externalURL), func(t *testing.T) {
					values := map[string]string{
						"anomaly-detection.enabled":        "true",
						"stackstate.features.server.split": fmt.Sprint(split),
						"ingress.enabled":                  fmt.Sprint(edge == "ingress"),
						"gateway.enabled":                  fmt.Sprint(edge == "gateway"),
						"gateway.parentRefs[0].name":       "test-gateway",
					}
					options := apiResourceNameTestOptions(values)
					options.ValuesFiles = []string{valuesFile}
					if externalURL {
						externalValues := filepath.Join(t.TempDir(), "external-url.yaml")
						require.NoError(t, os.WriteFile(externalValues, []byte(`global:
  url:
    fromSecret: '{{ include "stackstate.router.name" . }}-connection'
`), 0600))
						options.ValuesFiles = append(options.ValuesFiles, externalValues)
					}
					args := []string{"--api-versions", "networking.k8s.io/v1/Ingress"}
					before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "nightly", options, args...))
					output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil, args...)
					require.NoError(t, err)
					after := helmtestutil.NewKubernetesResources(t, output)
					legacy := "nightly-suse-observability-router"
					explicit := "explicit-router-service"
					require.Contains(t, before.Services, legacy)
					require.Contains(t, after.Services, explicit)
					assert.NotContains(t, after.Services, legacy)
					expected := before.Services[legacy]
					expected.Name = explicit
					assert.Equal(t, expected, after.Services[explicit], "Service ports, selectors and metadata must stay unchanged")
					delete(before.Services, legacy)
					delete(after.Services, explicit)
					assert.Equal(t, before.Services, after.Services)

					require.Contains(t, after.ConfigMaps, "router-name-probe")
					assert.Equal(t, map[string]string{
						"legacyName": legacy, "legacyPrefix": "nightly-suse-observability",
					}, after.ConfigMaps["router-name-probe"].Data)
					assertRouterServiceClientEndpoints(t, after, explicit, externalURL)
					// Catch stale endpoints in embedded restore scripts as well as
					// regular manifests. Legacy helper output above is only a name.
					assert.NotContains(t, output, "http://"+legacy+":8080")
					if edge == "ingress" {
						require.Len(t, after.Ingresses, 1)
						for _, ingress := range after.Ingresses {
							require.NotEmpty(t, ingress.Spec.Rules)
							for _, rule := range ingress.Spec.Rules {
								require.NotNil(t, rule.HTTP)
								require.NotEmpty(t, rule.HTTP.Paths)
								for _, path := range rule.HTTP.Paths {
									require.NotNil(t, path.Backend.Service)
									assert.Equal(t, explicit, path.Backend.Service.Name)
									assert.EqualValues(t, 8080, path.Backend.Service.Port.Number)
								}
							}
						}
					} else {
						require.Len(t, after.HTTPRoutes, 1)
						for _, route := range after.HTTPRoutes {
							require.Len(t, route.Spec.Rules, 1)
							require.Len(t, route.Spec.Rules[0].BackendRefs, 1)
							backend := route.Spec.Rules[0].BackendRefs[0]
							assert.Equal(t, explicit, string(backend.Name))
							require.NotNil(t, backend.Port)
							assert.EqualValues(t, 8080, *backend.Port)
						}
					}
					require.Contains(t, before.Deployments, "suse-observability-router")
					require.Contains(t, after.Deployments, "suse-observability-router")
					assert.Equal(t, before.Deployments["suse-observability-router"], after.Deployments["suse-observability-router"], "Changing the Service helper must not rename or restart the router")
					assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
					assert.Equal(t, before.Secrets, after.Secrets)
					assert.Equal(t, before.Roles, after.Roles)
					assert.Equal(t, before.RoleBindings, after.RoleBindings)
					assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
					assert.Equal(t, before.Statefulsets, after.Statefulsets)
				})
			}
		}
	}
}

func routerServiceNameTestChart(t *testing.T, rename bool) (string, string) {
	t.Helper()
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	if rename {
		path := filepath.Join(chart, "templates", "_names.tpl")
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		definition := regexp.MustCompile(`(?s)\{\{- define "stackstate.router.service.fullname" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAll(data, -1), 1)
		data = definition.ReplaceAll(data, []byte(`{{- define "stackstate.router.service.fullname" -}}explicit-router-service{{- end -}}`))
		require.NoError(t, os.WriteFile(path, data, 0600))
	}
	require.NoError(t, os.WriteFile(filepath.Join(chart, "templates", "router-name-probe.yaml"), []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: router-name-probe
data:
  legacyName: {{ include "stackstate.router.name" . | quote }}
  legacyPrefix: {{ include "stackstate.hostname.prefix" . | quote }}
`), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	return chart, valuesFile
}

func assertRouterServiceClientEndpoints(t *testing.T, resources helmtestutil.KubernetesResources, service string, externalURL bool) {
	t.Helper()
	endpoint := "http://" + service + ":8080"
	anomalyClients, agentConfigs, agentSecrets, backupConsumers := 0, 0, 0, 0
	for _, deployment := range resources.Deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			for i, arg := range container.Args {
				if arg == "--instance" {
					require.Less(t, i+1, len(container.Args))
					assert.Equal(t, endpoint, container.Args[i+1])
					anomalyClients++
				}
			}
			if externalURL {
				for _, env := range container.Env {
					if env.Name == "STS_URL" {
						require.NotNil(t, env.ValueFrom)
						require.NotNil(t, env.ValueFrom.SecretKeyRef)
						assert.Equal(t, "nightly-suse-observability-router-connection", env.ValueFrom.SecretKeyRef.Name)
						agentSecrets++
					}
				}
			}
		}
	}
	for _, config := range resources.ConfigMaps {
		if url, ok := config.Data["STS_URL"]; ok {
			assert.Equal(t, endpoint+"/receiver/stsAgent", url)
			agentConfigs++
		}
	}
	assert.Equal(t, 2, anomalyClients)
	if externalURL {
		assert.Zero(t, agentConfigs)
		assert.Equal(t, 1, agentSecrets)
	} else {
		assert.Equal(t, 1, agentConfigs)
	}
	checkBackup := func(containers []corev1.Container) {
		for _, container := range containers {
			for _, env := range container.Env {
				if env.Name == "STACKSTATE_ROUTER_ENDPOINT" {
					assert.Equal(t, endpoint, env.Value)
					backupConsumers++
				}
			}
		}
	}
	for _, job := range resources.Jobs {
		checkBackup(job.Spec.Template.Spec.Containers)
	}
	for _, job := range resources.CronJobs {
		checkBackup(job.Spec.JobTemplate.Spec.Template.Spec.Containers)
	}
	assert.Positive(t, backupConsumers)
}
