package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"gopkg.in/yaml.v2"
)

func TestPullSecretGlobal(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/pull_secret_global.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	deploymentsToCheck := []string{"api", "checks", "correlate", "initializer", "receiver", "slicing", "state", "problem-producer", "sync", "view-health", "mm2es", "e2es", "trace2es"}

	CheckDeploymentsForPullSecret(t, resources, deploymentsToCheck, "stackstate-pull-secret")
	CheckPullSecret(t, resources, "stackstate-pull-secret", "global-user", "global-password", "my.registry.com")
}

func TestPullSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/full.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	deploymentsToCheck := []string{"api", "checks", "correlate", "initializer", "receiver", "slicing", "state", "problem-producer", "sync", "view-health", "mm2es", "e2es", "trace2es"}

	CheckDeploymentsForPullSecret(t, resources, deploymentsToCheck, "stackstate-pull-secret")
	CheckPullSecret(t, resources, "stackstate-pull-secret", "test", "secret", "quay.io")
}

func TestPullSecretGlobalNamed(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/pull_secret_global_named.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	deploymentsToCheck := []string{"api", "checks", "correlate", "initializer", "receiver", "slicing", "state", "problem-producer", "sync", "view-health", "mm2es", "e2es", "trace2es"}

	CheckDeploymentsForPullSecret(t, resources, deploymentsToCheck, "my-existing-secret")
	for _, secret := range resources.Secrets {
		if secret.Name == "my-existing-secret" {
			assert.Fail(t, "Expected that no dockerconfigjson is being created")
		}
	}
}
func TestImagePullSecretName(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/pull_secret_name.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	deploymentsToCheck := []string{"api", "checks", "correlate", "initializer", "receiver", "slicing", "state", "problem-producer", "sync", "view-health", "mm2es", "e2es", "trace2es"}

	CheckDeploymentsForPullSecret(t, resources, deploymentsToCheck, "my-pull-secret")
	for _, secret := range resources.Secrets {
		if secret.Name == "my-pull-secret" {
			assert.Fail(t, "Expected that no dockerconfigjson is being created")
		}
	}
}

func TestGlobalRegistryLocalPullSecret(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/pull_secret_global_registry.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	deploymentsToCheck := []string{"api", "checks", "correlate", "initializer", "receiver", "slicing", "state", "problem-producer", "sync", "view-health", "mm2es", "e2es", "trace2es"}

	CheckDeploymentsForPullSecret(t, resources, deploymentsToCheck, "stackstate-pull-secret")
	CheckPullSecret(t, resources, "stackstate-pull-secret", "test", "secret", "my.registry.com")
}

func TestGlobalOverridesLocalPullSecretDetails(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "stackstate", "values/pull_secret_both.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)
	deploymentsToCheck := []string{"api", "checks", "correlate", "initializer", "receiver", "slicing", "state", "problem-producer", "sync", "view-health", "mm2es", "e2es", "trace2es"}

	CheckDeploymentsForPullSecret(t, resources, deploymentsToCheck, "stackstate-pull-secret")
	CheckPullSecret(t, resources, "stackstate-pull-secret", "global-user", "global-password", "my.registry.com")
}

func CheckDeploymentsForPullSecret(t *testing.T, resources helmtestutil.KubernetesResources, deploymentsToCheck []string, pullSecretName string) {
	checked := []string{}
	for _, deployment := range resources.Deployments {
		for _, name := range deploymentsToCheck {
			if ("stackstate-" + name) == deployment.Name {
				checked = append(checked, name)
				assert.Len(t, deployment.Spec.Template.Spec.ImagePullSecrets, 1)
				assert.Equal(t, pullSecretName, deployment.Spec.Template.Spec.ImagePullSecrets[0].Name)
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
			registryAuth := auths[registry].(map[interface{}]interface{})
			assert.Equal(t, user, registryAuth["username"])
			assert.Equal(t, password, registryAuth["password"])
			return
		}
	}
	assert.Fail(t, "No pull secret found")
}
