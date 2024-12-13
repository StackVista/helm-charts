package test

import (
	"encoding/base64"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"gopkg.in/yaml.v2"
)

func TestPullSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	CheckPullSecret(t, resources, "suse-observability-pull-secret", "test", "secret", "my.registry.com")
}

func TestPullSecretGlobal(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/pull_secret_global.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	CheckPullSecret(t, resources, "suse-observability-pull-secret", "test", "secret", "my.registry.com")
}

func TestPullSecretGlobalNamed(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/pull_secret_global_named.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	deploymentsToCheck := []string{"api", "checks", "correlate", "initializer", "receiver", "slicing", "state", "sync", "e2es"}

	CheckDeploymentsForPullSecret(t, resources, deploymentsToCheck, "my-existing-secret")
	for _, secret := range resources.Secrets {
		if secret.Name == "my-existing-secret" {
			assert.Fail(t, "Expected that no dockerconfigjson is being created")
		}
	}
}

func CheckDeploymentsForPullSecret(t *testing.T, resources helmtestutil.KubernetesResources, deploymentsToCheck []string, pullSecretName ...string) {
	checked := []string{}
	for _, deployment := range resources.Deployments {
		for _, name := range deploymentsToCheck {
			if ("suse-observability-" + name) == deployment.Name {
				checked = append(checked, name)
				assert.Len(t, deployment.Spec.Template.Spec.ImagePullSecrets, len(pullSecretName))
				for _, secret := range deployment.Spec.Template.Spec.ImagePullSecrets {
					assert.Contains(t, pullSecretName, secret.Name)
				}
			}
		}
	}

	assert.ElementsMatch(t, checked, deploymentsToCheck)
}

func CheckPullSecret(t *testing.T, resources helmtestutil.KubernetesResources, secretName string, user string, password string, registry string) {
	for _, secret := range resources.Secrets {
		if secret.Type == "kubernetes.io/dockerconfigjson" && secret.Name == secretName {
			assert.Len(t, secret.Data, 1)

			assert.Contains(t, secret.Data, ".dockerconfigjson")
			dockerConfigJson := secret.Data[".dockerconfigjson"]
			dockerConfig := map[interface{}]interface{}{}
			assert.NoError(t, yaml.Unmarshal(dockerConfigJson, &dockerConfig))
			assert.Contains(t, dockerConfig, "auths")
			auths := dockerConfig["auths"].(map[interface{}]interface{})
			assert.Contains(t, auths, registry)
			registryAuth := auths[registry].(map[interface{}]interface{})["auth"].(string)
			decodedBytes, err := base64.StdEncoding.DecodeString(registryAuth)
			if err != nil {
				log.Fatalf("Error decoding base64 string: %v", err)
			}

			assert.Equal(t, fmt.Sprintf("%s:%s", user, password), string(decodedBytes))
			return
		}
	}
	assert.Fail(t, "No pull secret found")
}
