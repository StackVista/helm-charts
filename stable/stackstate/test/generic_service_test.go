package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestAllServicesRenderedSplit(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/dummy_trust_store.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	podsToCheck := append([]string{"api", "checks", "correlate", "initializer", "receiver", "slicing", "state", "sync", "view-health", "mm2es", "e2es", "trace2es", "sts2es"})

	assertTrustStoreOnPods(t, resources, podsToCheck)
}

func TestAllServicesRenderedServer(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml", "values/dummy_trust_store.yaml", "values/split_disabled.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	podsToCheck := append([]string{"server", "correlate", "receiver", "mm2es", "e2es", "trace2es", "sts2es"})

	assertTrustStoreOnPods(t, resources, podsToCheck)
}

func assertTrustStoreOnPods(t *testing.T, resources helmtestutil.KubernetesResources, podsToCheck []string) {
	checked := []string{}
	for _, deployment := range resources.Deployments {
		var podName string
		for _, name := range podsToCheck {
			if ("stackstate-" + name) == deployment.Name {
				checked = append(checked, name)
				if strings.Contains(name, "2es") {
					podName = "k2es"
				} else {
					podName = name
				}
			}
		}
		if podName != "" {
			assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Args, "-Djavax.net.ssl.trustStore=/opt/docker/secrets/java-cacerts")
			assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Args, "-Djavax.net.ssl.trustStorePassword=$(JAVA_TRUSTSTORE_PASSWORD)")
			assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Args, "-Djavax.net.ssl.trustStoreType=jks")
			assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Args, "-Dlogback.configurationFile=/opt/docker/etc_log/logback.groovy")

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
			assert.Equal(t, "/opt/docker/etc_log", logVolumePath)
			assert.Equal(t, "/opt/docker/secrets", secretVolumePath)

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
			assert.Equal(t, "stackstate-"+podName+"-log", logConfigMapName)
			assert.Equal(t, "stackstate-common", secretName)
		}
	}
	assert.ElementsMatch(t, podsToCheck, checked, "Not all expected deployments were found")
}
