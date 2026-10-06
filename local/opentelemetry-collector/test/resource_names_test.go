package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

func TestCollectorResourcesFollowDedicatedHelpers(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates", "_names.tpl")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	definitions := string(data)
	kinds := map[string]string{
		"deployment": "Deployment", "daemonset": "DaemonSet", "statefulset": "StatefulSet",
		"service": "Service", "grpc.service": "Service",
		"deployment.configmap": "ConfigMap", "daemonset.configmap": "ConfigMap", "statefulset.configmap": "ConfigMap",
		"serviceaccount": "ServiceAccount", "clusterrole": "ClusterRole", "clusterrolebinding": "ClusterRoleBinding",
		"horizontalpodautoscaler": "HorizontalPodAutoscaler", "poddisruptionbudget": "PodDisruptionBudget",
		"podmonitor": "PodMonitor", "servicemonitor": "ServiceMonitor",
		"prometheusrule": "PrometheusRule", "networkpolicy": "NetworkPolicy",
		"ingress": "Ingress", "httproute": "HTTPRoute", "grpcroute": "GRPCRoute",
	}
	name := func(resource string) string { return "explicit-" + strings.ReplaceAll(resource, ".", "-") }
	for resource := range kinds {
		helper := "opentelemetry-collector." + resource + ".fullname"
		header := `define "` + helper + `"`
		require.Equal(t, 1, strings.Count(definitions, header))
		// Keep nested template blocks intact while replacing the public helper.
		definitions = strings.Replace(definitions, header, `define "test.original.`+helper+`"`, 1)
		body := name(resource)
		if resource == "ingress" || resource == "httproute" || resource == "grpcroute" {
			body += `{{ with .name }}-{{ . }}{{ end }}`
		}
		definitions += "\n{{- " + header + " -}}" + body + "{{- end -}}\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(definitions), 0600))
	seen := map[string]bool{}
	for _, mode := range []string{"deployment", "daemonset", "statefulset"} {
		for _, externalConfig := range []bool{false, true} {
			for _, gateway := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/external-config=%t/gateway=%t", mode, externalConfig, gateway), func(t *testing.T) {
					values := map[string]string{
						"mode": mode, "service.enabled": "true", "autoscaling.enabled": "true",
						"podDisruptionBudget.enabled": "true", "podDisruptionBudget.minAvailable": "1",
						"podMonitor.enabled": "true", "serviceMonitor.enabled": "true",
						"prometheusRule.enabled": "true", "networkPolicy.enabled": "true",
						"clusterRole.create": "true", "presets.logsCollection.enabled": "true",
						"ingress.enabled": "true", "ingress.additionalIngresses[0].name": "additional",
						"ingress.hosts[0].host":          "collector.example.test",
						"ingress.hosts[0].paths[0].path": "/", "ingress.hosts[0].paths[0].pathType": "Prefix",
						"ingress.hosts[0].paths[0].port": "4318",
						"ingress.hosts[0].paths[1].path": "/customer", "ingress.hosts[0].paths[1].pathType": "Prefix",
						"ingress.hosts[0].paths[1].port": "4318", "ingress.hosts[0].paths[1].serviceName": "customer-service",
						"gateway.enabled": "true", "gateway.parentRefs[0].name": "customer-gateway",
						"gateway.additionalGateways[0].name": "grpc", "gateway.additionalGateways[0].kind": "GRPCRoute",
						"gateway.additionalGateways[0].parentRefs[0].name":                    "customer-gateway",
						"statefulset.volumeClaimTemplates[0].metadata.name":                   "data",
						"statefulset.volumeClaimTemplates[0].spec.accessModes[0]":             "ReadWriteOnce",
						"statefulset.volumeClaimTemplates[0].spec.resources.requests.storage": "1Gi",
					}
					values["ingress.enabled"] = fmt.Sprint(!gateway)
					values["gateway.enabled"] = fmt.Sprint(gateway)
					if externalConfig {
						values["configMap.create"] = "false"
						values["configMap.existingName"] = "customer-config"
					}
					opts := &helm.Options{SetValues: values}
					before := collectorNameDocuments(t, helmtestutil.RenderHelmTemplateOptsNoError(t, "orders", opts))
					output, err := helm.RenderTemplateE(t, opts, chart, "orders", nil)
					require.NoError(t, err)
					after := collectorNameDocuments(t, output)
					get := func(kind, resource string) map[string]interface{} {
						key := kind + "/" + name(resource)
						require.Contains(t, after, key)
						return after[key]
					}
					for resource, kind := range kinds {
						key := kind + "/" + name(resource)
						if resource == "grpcroute" {
							key += "-grpc"
						}
						if _, exists := after[key]; exists {
							seen[resource] = true
						}
					}
					workload := get(kinds[mode], mode)
					pod := collectorNameField(t, workload, "spec", "template", "spec").(map[string]interface{})
					assert.Equal(t, name("serviceaccount"), pod["serviceAccountName"])
					if mode != "daemonset" {
						hpa := get("HorizontalPodAutoscaler", "horizontalpodautoscaler")
						assert.Equal(t, name(mode), collectorNameField(t, hpa, "spec", "scaleTargetRef", "name"))
					}
					if mode == "statefulset" {
						assert.Equal(t, name("service"), collectorNameField(t, workload, "spec", "serviceName"))
						assert.Equal(t, collectorNameField(t, before["StatefulSet/"+fullName], "spec", "volumeClaimTemplates"),
							collectorNameField(t, workload, "spec", "volumeClaimTemplates"))
					}
					configName := name(mode + ".configmap")
					if externalConfig {
						configName = "customer-config"
						for key := range after {
							assert.False(t, strings.HasPrefix(key, "ConfigMap/"))
						}
					} else {
						cm := get("ConfigMap", mode+".configmap")
						config := collectorNameField(t, cm, "data", "relay").(string)
						assert.Contains(t, config, "_"+name(mode)+"*_*/", "self-log exclusion follows the workload")
					}
					volumes := pod["volumes"].([]interface{})
					foundConfig := false
					for _, volume := range volumes {
						v := volume.(map[string]interface{})
						if v["name"] == "opentelemetry-collector-configmap" {
							assert.Equal(t, configName, collectorNameField(t, v, "configMap", "name"))
							foundConfig = true
						}
					}
					require.True(t, foundConfig)
					binding := get("ClusterRoleBinding", "clusterrolebinding")
					assert.Equal(t, name("clusterrole"), collectorNameField(t, binding, "roleRef", "name"))
					subject := binding["subjects"].([]interface{})[0].(map[string]interface{})
					assert.Equal(t, name("serviceaccount"), subject["name"])
					if gateway {
						for kind, resource := range map[string]string{"HTTPRoute": "httproute", "GRPCRoute": "grpcroute"} {
							key := kind + "/" + name(resource)
							service := name("service")
							if kind == "GRPCRoute" {
								key += "-grpc"
								service = name("grpc.service")
							}
							require.Contains(t, after, key)
							rule := collectorNameField(t, after[key], "spec", "rules").([]interface{})[0].(map[string]interface{})
							backend := rule["backendRefs"].([]interface{})[0].(map[string]interface{})
							assert.Equal(t, service, backend["name"])
						}
					} else {
						ingress := get("Ingress", "ingress")
						rule := collectorNameField(t, ingress, "spec", "rules").([]interface{})[0].(map[string]interface{})
						paths := collectorNameField(t, rule, "http", "paths").([]interface{})
						assert.Equal(t, name("service"), collectorNameField(t, paths[0].(map[string]interface{}), "backend", "service", "name"))
						assert.Equal(t, "customer-service", collectorNameField(t, paths[1].(map[string]interface{}), "backend", "service", "name"))
						require.Contains(t, after, "Ingress/"+name("ingress")+"-additional")
					}
					oldWorkloadName := fullName
					if mode == "daemonset" {
						oldWorkloadName += "-agent"
					}
					oldWorkload := before[kinds[mode]+"/"+oldWorkloadName]
					assert.Equal(t, collectorNameField(t, oldWorkload, "spec", "selector"),
						collectorNameField(t, workload, "spec", "selector"))
					assert.Equal(t, collectorNameField(t, oldWorkload, "spec", "template", "metadata", "labels"),
						collectorNameField(t, workload, "spec", "template", "metadata", "labels"))
					oldHash := collectorNameField(t, oldWorkload, "spec", "template", "metadata", "annotations", "checksum/config")
					newHash := collectorNameField(t, workload, "spec", "template", "metadata", "annotations", "checksum/config")
					if externalConfig {
						assert.Equal(t, oldHash, newHash)
					} else {
						assert.NotEqual(t, oldHash, newHash, "fixture-only ConfigMap renames update the existing checksum")
					}
				})
			}
		}
	}
	for resource := range kinds {
		assert.True(t, seen[resource], "resource helper was not exercised: %s", resource)
	}
}

func collectorNameDocuments(t *testing.T, output string) map[string]map[string]interface{} {
	t.Helper()
	result := map[string]map[string]interface{}{}
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(output), 4096)
	for {
		var object map[string]interface{}
		err := decoder.Decode(&object)
		if err == io.EOF {
			return result
		}
		require.NoError(t, err)
		if len(object) == 0 {
			continue
		}
		name := collectorNameField(t, object, "metadata", "name").(string)
		result[object["kind"].(string)+"/"+name] = object
	}
}

func collectorNameField(t *testing.T, object map[string]interface{}, path ...string) interface{} {
	t.Helper()
	var current interface{} = object
	for _, field := range path {
		m, ok := current.(map[string]interface{})
		require.True(t, ok, "expected object at %s", field)
		require.Contains(t, m, field)
		current = m[field]
	}
	return current
}
