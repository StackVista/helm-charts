package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func tenantmanagementContainer(t *testing.T, setValues map[string]string) corev1.Container {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "tenantmanagement", &helm.Options{SetValues: setValues})
	resources := helmtestutil.NewKubernetesResources(t, output)
	require.Len(t, resources.Deployments, 1)
	for _, d := range resources.Deployments {
		return d.Spec.Template.Spec.Containers[0]
	}
	return corev1.Container{}
}

func TestJavaOptsUnsetByDefault(t *testing.T) {
	for _, env := range tenantmanagementContainer(t, nil).Env {
		assert.NotEqual(t, "JAVA_OPTS", env.Name)
	}
}

func TestJavaOptsSet(t *testing.T) {
	container := tenantmanagementContainer(t, map[string]string{"javaOpts": "-XX:-UseBASE64Intrinsics"})
	assert.Contains(t, container.Env, corev1.EnvVar{Name: "JAVA_OPTS", Value: "-XX:-UseBASE64Intrinsics"})
}
