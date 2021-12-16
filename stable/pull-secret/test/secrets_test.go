package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestRenderPullSecretWithOneRegistry(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "pull-secret", "values/one_registry.yaml")

	// Parse all resources into their corresponding types for validation and further inspection
	r := helmtestutil.NewKubernetesResources(t, output)
	assert.Len(t, r.Secrets, 1)
	s := r.Secrets[0]
	assert.Equal(t, "pull-secret", s.Name)
	assert.Equal(t, "kubernetes.io/dockerconfigjson", string(s.Type))
	json := s.Data[".dockerconfigjson"]
	assert.Equal(t, `{"auths":{"docker.io":{"auth":"am9objpkb2U="}}}`, string(json))
}

func TestRenderPullSecretWithTwoRegistries(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "pull-secret", "values/two_registries.yaml")

	// Parse all resources into their corresponding types for validation and further inspection
	r := helmtestutil.NewKubernetesResources(t, output)
	assert.Len(t, r.Secrets, 1)
	s := r.Secrets[0]
	assert.Equal(t, "pull-secret", s.Name)
	assert.Equal(t, "kubernetes.io/dockerconfigjson", string(s.Type))
	json := s.Data[".dockerconfigjson"]
	assert.Equal(t, `{"auths":{"docker.io":{"auth":"am9objpkb2U="},"quay.io":{"auth":"cGlldGplOnB1aw=="}}}`, string(json))
}

func TestRenderPullSecretWithGlobal(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "pull-secret", "values/global_secret.yaml")

	// Parse all resources into their corresponding types for validation and further inspection
	r := helmtestutil.NewKubernetesResources(t, output)
	assert.Len(t, r.Secrets, 1)
	s := r.Secrets[0]
	assert.Equal(t, "pull-secret", s.Name)
	assert.Equal(t, "kubernetes.io/dockerconfigjson", string(s.Type))
	json := s.Data[".dockerconfigjson"]
	assert.Equal(t, `{"auths":{"artifactory.example.com":{"auth":"amFuZTpkb2U="}}}`, string(json))
}
