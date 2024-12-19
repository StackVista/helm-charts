package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestAllServicesRenderedSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/dummy_trust_store.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	podsToCheck := append([]string{"api", "checks", "correlate", "initializer", "receiver", "slicing", "state", "sync", "e2es"})

	assertTrustStoreOnPods(t, resources, podsToCheck)
}

func TestAllServicesRenderedServer(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/dummy_trust_store.yaml", "values/split_disabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	podsToCheck := append([]string{"server", "correlate", "receiver", "e2es"})

	assertTrustStoreOnPods(t, resources, podsToCheck)
}

func assertTrustStoreOnPods(t *testing.T, resources helmtestutil.KubernetesResources, podsToCheck []string) {
	checked := []string{}
	for _, deployment := range resources.Deployments {
		var podName string
		for _, name := range podsToCheck {
			if ("suse-observability-" + name) == deployment.Name {
				checked = append(checked, name)
				podName = name
			}
		}
		if podName != "" {
			assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Args, "-Djavax.net.ssl.trustStore=/opt/docker/secrets/java-cacerts", "For pod "+podName)
			assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Args, "-Djavax.net.ssl.trustStorePassword=$(JAVA_TRUSTSTORE_PASSWORD)", "For pod "+podName)
			assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Args, "-Djavax.net.ssl.trustStoreType=jks", "For pod "+podName)
			assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Args, "-Dlogback.configurationFile=/opt/docker/etc_log/logback.xml", "For pod "+podName)

			var logVolumePath string
			var secretVolumePath string
			for _, volume := range deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
				if volume.Name == "config-volume-log" {
					logVolumePath = volume.MountPath
				}
				if volume.Name == "service-secrets-volume" {
					secretVolumePath = volume.MountPath
				}
			}
			assert.Equal(t, "/opt/docker/etc_log", logVolumePath, "For pod "+podName)
			assert.Equal(t, "/opt/docker/secrets", secretVolumePath, "For pod "+podName)

			var logConfigMapName string
			var secretName string
			for _, volume := range deployment.Spec.Template.Spec.Volumes {
				if volume.Name == "config-volume-log" {
					logConfigMapName = volume.ConfigMap.Name
				}
				if volume.Name == "service-secrets-volume" {
					secretName = volume.Secret.SecretName
				}
			}
			assert.Equal(t, "suse-observability-"+podName+"-log", logConfigMapName, "For pod "+podName)
			assert.Equal(t, "suse-observability-common", secretName, "For pod "+podName)
		}
	}
	assert.ElementsMatch(t, podsToCheck, checked, "Not all expected deployments were found")
}
