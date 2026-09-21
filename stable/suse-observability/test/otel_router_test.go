package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestOtelRouterRouteEnabled(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	routerConfigMap, ok := resources.ConfigMaps["suse-observability-router-active"]
	require.True(t, ok, "Active router configmap should exist")
	listeners := routerConfigMap.Data["listeners.yaml"]

	//   /stsAgent/otel/          — direct stsAgent URL
	//   /receiver/stsAgent/otel/ — stsAgent URL proxied through /receiver/
	assert.Contains(t, listeners, "prefix: \"/stsAgent/otel/\"")
	assert.Contains(t, listeners, "prefix: \"/receiver/stsAgent/otel/\"")
	assert.NotContains(t, listeners, "prefix: \"/receiver/otel/\"",
		"old /receiver/otel/ route was based on a non-canonical URL shape and should not be present")

	assert.Contains(t, routerConfigMap.Data["clusters.yaml"], "name: \"suse-observability-otel-collector\"")
	assert.Contains(t, routerConfigMap.Data["clusters.yaml"], "port_value: 4318")
}

// TestOtelRouterClusterNameStableAcrossReleaseNames asserts the otel-collector cluster address
// uses the subchart's stable fullnameOverride regardless of the Helm release name. A non-default
// release name previously caused the router to reference a non-existent service (503 UH).
func TestOtelRouterClusterNameStableAcrossReleaseNames(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "prime-test", &helm.Options{
		ValuesFiles: []string{"values/full.yaml"},
	})

	resources := helmtestutil.NewKubernetesResources(t, output)

	routerConfigMap, ok := resources.ConfigMaps["prime-test-suse-observability-router-active"]
	require.True(t, ok, "Active router configmap should exist for non-default release name")

	clusters := routerConfigMap.Data["clusters.yaml"]
	assert.Contains(t, clusters, "address: \"suse-observability-otel-collector\"",
		"router must use the stable otel-collector service name regardless of release name")
	assert.NotContains(t, clusters, "address: \"prime-test-suse-observability-otel-collector\"",
		"router must not use release-prefixed name for otel-collector (service does not exist)")
}

func TestOtelRouterFullnameOverride(t *testing.T) {
	for _, release := range []string{"suse-observability", "nightly"} {
		t.Run(release, func(t *testing.T) {
			const collectorName = "custom-otel-collector"
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, release, &helm.Options{
				ValuesFiles: []string{"values/full.yaml"},
				SetValues: map[string]string{
					"opentelemetry-collector.fullnameOverride": collectorName,
				},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			require.Contains(t, resources.Services, collectorName)
			require.Contains(t, resources.Statefulsets, collectorName)
			assert.NotContains(t, resources.Services, "suse-observability-otel-collector")

			const configName = "suse-observability-otel-collector"
			require.Contains(t, resources.ConfigMaps, configName)
			collector := resources.Statefulsets[collectorName]
			require.NotEmpty(t, collector.Spec.Template.Spec.Containers)
			for envName, key := range map[string]string{"API_URL": "api.url", "INTAKE_URL": "intake.url"} {
				assert.NotEmpty(t, resources.ConfigMaps[configName].Data[key])
				assert.Contains(t, collector.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
					Name: envName,
					ValueFrom: &corev1.EnvVarSource{ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: configName},
						Key:                  key,
					}},
				})
			}

			routerName := "suse-observability-router-active"
			if release != "suse-observability" {
				routerName = release + "-" + routerName
			}
			require.Contains(t, resources.ConfigMaps, routerName)
			clusters := resources.ConfigMaps[routerName].Data["clusters.yaml"]
			assert.Contains(t, clusters, `name: "`+collectorName+`"`)
			assert.Contains(t, clusters, `address: "`+collectorName+`"`)
			assert.NotContains(t, clusters, `address: "suse-observability-otel-collector"`)
		})
	}
}
