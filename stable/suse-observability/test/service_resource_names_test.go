package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestAPIServiceReferencesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	for component, suffix := range map[string]string{
		"api": "api", "server": "server", "initializer": "initializer", "authorizationSync": "authorization-sync",
	} {
		helper := "stackstate." + component + ".service.fullname"
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAllString(content, -1), 1, helper)
		content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}explicit-`+suffix+`-service{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, serverSplit := range []bool{false, true} {
		for _, receiverSplit := range []bool{false, true} {
			for _, backups := range []bool{false, true} {
				for _, hbaseMode := range []string{"Distributed", "Mono"} {
					t.Run(fmt.Sprintf("server=%t/receiver=%t/backups=%t/hbase=%s", serverSplit, receiverSplit, backups, hbaseMode), func(t *testing.T) {
						options := apiResourceNameTestOptions(map[string]string{
							"stackstate.features.server.split":             fmt.Sprint(serverSplit),
							"stackstate.components.receiver.split.enabled": fmt.Sprint(receiverSplit),
							"hbase.deployment.mode":                        hbaseMode,
							"global.backup.enabled":                        fmt.Sprint(backups),
						})
						before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
						options.ValuesFiles = []string{valuesFile}
						output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
						require.NoError(t, err)
						after := helmtestutil.NewKubernetesResources(t, output)

						component, inactive := "server", "api"
						if serverSplit {
							component, inactive = "api", "server"
						}
						prefix := "nightly-suse-observability-"
						legacy := prefix + component + "-headless"
						explicit := "explicit-" + component + "-service"
						require.Contains(t, before.Services, legacy)
						require.Contains(t, after.Services, explicit)
						assert.NotContains(t, after.Services, "explicit-"+inactive+"-service")
						expected := before.Services[legacy]
						expected.Name = explicit
						assert.Equal(t, expected, after.Services[explicit], "Service ports and selectors must stay unchanged")
						delete(before.Services, legacy)
						delete(after.Services, explicit)
						if serverSplit {
							require.Contains(t, before.Services, prefix+"authorization-sync")
							require.Contains(t, after.Services, "explicit-authorization-sync-service")
							expectedAuthorizationSync := before.Services[prefix+"authorization-sync"]
							expectedAuthorizationSync.Name = "explicit-authorization-sync-service"
							assert.Equal(t, expectedAuthorizationSync, after.Services["explicit-authorization-sync-service"])
							delete(before.Services, prefix+"authorization-sync")
							delete(after.Services, "explicit-authorization-sync-service")
							require.Contains(t, before.Services, prefix+"initializer")
							require.Contains(t, after.Services, "explicit-initializer-service")
							expectedInitializer := before.Services[prefix+"initializer"]
							expectedInitializer.Name = "explicit-initializer-service"
							assert.Equal(t, expectedInitializer, after.Services["explicit-initializer-service"])
							delete(before.Services, prefix+"initializer")
							delete(after.Services, "explicit-initializer-service")
							for _, component := range []string{"api", "authorization-sync", "checks", "health-sync", "notification", "slicing", "state", "sync"} {
								name := prefix + component
								require.Contains(t, after.Deployments, name)
								init := after.Deployments[name].Spec.Template.Spec.InitContainers
								waitIndex := slices.IndexFunc(init, func(container corev1.Container) bool {
									return container.Name == "server-init"
								})
								require.NotEqual(t, -1, waitIndex, name)
								assert.Contains(t, strings.Join(init[waitIndex].Command, " "), ",explicit-initializer-service:1618 -t 300")
							}
						} else {
							assert.NotContains(t, after.Services, "explicit-initializer-service")
							assert.NotContains(t, after.Services, "explicit-authorization-sync-service")
						}
						assert.Equal(t, before.Services, after.Services)
						// Every occurrence of these DNS names must follow the Service helpers.
						assert.NotContains(t, output, prefix+"api-headless")
						assert.NotContains(t, output, prefix+"server-headless")
						assert.NotContains(t, output, prefix+"initializer:1618")
						assert.NotContains(t, output, prefix+"authorization-sync:7075")

						receiverSuffixes := []string{""}
						if receiverSplit {
							receiverSuffixes = []string{"-base", "-logs", "-process-agent"}
						}
						authService := explicit
						if serverSplit {
							authService = "explicit-authorization-sync-service"
						}
						for _, suffix := range receiverSuffixes {
							name := prefix + "receiver" + suffix
							require.Contains(t, after.Deployments, name)
							env := after.Deployments[name].Spec.Template.Spec.Containers[0].Env
							assert.Contains(t, env, corev1.EnvVar{
								Name:  "CONFIG_FORCE_stackstate_receiver_agentLeases_agentServiceBaseUri",
								Value: "http://" + explicit + ":7070/internal/api/agents",
							})
							assert.Contains(t, env, corev1.EnvVar{
								Name:  "CONFIG_FORCE_stackstate_receiver_authorizationService_authorizationServiceBaseUri",
								Value: "http://" + explicit + ":7070/api/user/authorization",
							})
							assert.Contains(t, env, corev1.EnvVar{
								Name:  "CONFIG_FORCE_stackstate_receiver_authorizationSyncApi_authorizationSyncApiBaseUri",
								Value: "http://" + authService + ":7075/rbac",
							})
						}

						require.Contains(t, after.Deployments, "suse-observability-mcp")
						assert.Contains(t, after.Deployments["suse-observability-mcp"].Spec.Template.Spec.Containers[0].Args, "http://"+explicit+":7070")
						require.Contains(t, after.ConfigMaps, prefix+"router-active")
						routerBefore := before.ConfigMaps[prefix+"router-active"]
						routerAfter := after.ConfigMaps[prefix+"router-active"]
						assert.Contains(t, routerAfter.Data["clusters.yaml"], `address: "`+explicit+`"`)
						assert.Equal(t, strings.ReplaceAll(routerBefore.Data["clusters.yaml"], legacy, explicit), routerAfter.Data["clusters.yaml"])
						assert.Equal(t, routerBefore.Data["listeners.yaml"], routerAfter.Data["listeners.yaml"], "Envoy cluster identifiers must stay unchanged")

						backupTarget := explicit + ":7070"
						if serverSplit {
							backupTarget = "explicit-initializer-service:1618"
						}
						for _, name := range []string{"backup-sg", "backup-sg-v2", "backup-conf"} {
							cronJobName := prefix + name
							if name == "backup-sg" {
								cronJobName = "suse-observability-backup-sg"
							}
							if !backups && name != "backup-conf" {
								assert.NotContains(t, after.CronJobs, cronJobName)
								continue
							}
							require.Contains(t, after.CronJobs, cronJobName)
							init := after.CronJobs[cronJobName].Spec.JobTemplate.Spec.Template.Spec.InitContainers
							require.NotEmpty(t, init)
							assert.Contains(t, strings.Join(init[0].Command, " "), ","+backupTarget+" -t 300")
						}
						assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
						assert.Equal(t, before.Statefulsets, after.Statefulsets)
					})
				}
			}
		}
	}
}

func TestSplitServiceReferencesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	namesPath := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(namesPath)
	require.NoError(t, err)
	content := string(data)
	variants := map[string][]string{
		"receiver":  {"", "-base", "-logs", "-process-agent"},
		"correlate": {"", "-connection", "-http-tracing", "-aggregator"},
	}
	receiverNames := map[string]string{
		"": "standalone-ingest", "-base": "primary-ingest",
		"-logs": "log-ingest", "-process-agent": "process-ingest",
	}
	// Unrelated outputs ensure split Services never derive their names from
	// the unsplit Service helper, or from one another.
	for helper, body := range map[string]string{
		"stackstate.receiver.service.fullname":              receiverNames[""],
		"stackstate.receiver.base.service.fullname":         receiverNames["-base"],
		"stackstate.receiver.logs.service.fullname":         receiverNames["-logs"],
		"stackstate.receiver.processAgent.service.fullname": receiverNames["-process-agent"],
		"stackstate.correlate.service.fullname":             `explicit-correlate-service{{ template "stackstate.correlate.name.postfix" . }}`,
	} {
		definition := regexp.MustCompile(`(?s)\{\{- define "` + regexp.QuoteMeta(helper) + `" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, definition.FindAllString(content, -1), 1, helper)
		content = definition.ReplaceAllString(content, `{{- define "`+helper+`" -}}`+body+`{{- end -}}`)
	}
	require.NoError(t, os.WriteFile(namesPath, []byte(content), 0600))
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, serverSplit := range []bool{false, true} {
		for _, receiverSplit := range []bool{false, true} {
			for _, correlateSplit := range []bool{false, true} {
				for _, routerMode := range []string{"active", "maintenance"} {
					t.Run(fmt.Sprintf("server=%t/receiver=%t/correlate=%t/router=%s", serverSplit, receiverSplit, correlateSplit, routerMode), func(t *testing.T) {
						values := componentResourceNameTestValues()
						values["stackstate.features.server.split"] = fmt.Sprint(serverSplit)
						values["stackstate.components.receiver.split.enabled"] = fmt.Sprint(receiverSplit)
						values["stackstate.components.correlate.split.enabled"] = fmt.Sprint(correlateSplit)
						values["stackstate.components.router.mode.status"] = routerMode
						values["stackstate.components.all.metrics.servicemonitor.enabled"] = "true"
						options := apiResourceNameTestOptions(values)
						before := helmtestutil.NewKubernetesResources(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", options))
						options.ValuesFiles = []string{valuesFile}
						output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil)
						require.NoError(t, err)
						after := helmtestutil.NewKubernetesResources(t, output)
						prefix := "nightly-suse-observability-"
						routerName := prefix + "router-" + routerMode
						require.Contains(t, before.ConfigMaps, routerName)
						require.Contains(t, after.ConfigMaps, routerName)
						expectedClusters := before.ConfigMaps[routerName].Data["clusters.yaml"]
						for component, suffixes := range variants {
							split := receiverSplit
							if component == "correlate" {
								split = correlateSplit
							}
							for _, suffix := range suffixes {
								legacy := prefix + component + suffix
								explicit := "explicit-" + component + "-service" + suffix
								if component == "receiver" {
									explicit = receiverNames[suffix]
								}
								if split != (suffix != "") {
									assert.NotContains(t, before.Services, legacy)
									assert.NotContains(t, after.Services, explicit)
									continue
								}
								require.Contains(t, before.Services, legacy)
								require.Contains(t, after.Services, explicit)
								assert.NotContains(t, after.Services, legacy)
								expected := before.Services[legacy]
								expected.Name = explicit
								assert.Equal(t, expected, after.Services[explicit], "Service ports, labels and selectors must stay unchanged")
								delete(before.Services, legacy)
								delete(after.Services, explicit)
								assert.Equal(t, before.Deployments[legacy], after.Deployments[legacy])
								if component == "receiver" {
									oldAddress := `address: "` + legacy + `"`
									newAddress := `address: "` + explicit + `"`
									assert.Contains(t, expectedClusters, oldAddress)
									expectedClusters = strings.ReplaceAll(expectedClusters, oldAddress, newAddress)
								}
							}
						}
						assert.Equal(t, before.Services, after.Services)
						// Replace only DNS addresses in the expectation, preserving Envoy identifiers.
						assert.Equal(t, expectedClusters, after.ConfigMaps[routerName].Data["clusters.yaml"])
						assert.Equal(t, before.ConfigMaps[routerName].Data["listeners.yaml"], after.ConfigMaps[routerName].Data["listeners.yaml"])
						target := receiverNames[""]
						if receiverSplit {
							target = receiverNames["-base"]
						}
						require.Contains(t, after.ConfigMaps, "suse-observability-otel-collector")
						assert.Equal(t, map[string]string{
							"api.url":    "http://" + target + ":7077/stsAgent/api/v1/validate",
							"intake.url": "http://" + target + ":7077/stsAgent/intake",
						}, after.ConfigMaps["suse-observability-otel-collector"].Data)
						assert.Equal(t, before.ServiceMonitors, after.ServiceMonitors)
						assert.Equal(t, before.Secrets, after.Secrets)
						assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
						assert.Equal(t, before.Pdbs, after.Pdbs)
						assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
						assert.Equal(t, before.Statefulsets, after.Statefulsets)
					})
				}
			}
		}
	}
}
