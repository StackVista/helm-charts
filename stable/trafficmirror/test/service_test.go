package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

func TestServiceAppProtocolUnsetByDefault(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "trafficmirror")
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Len(t, resources.Services, 1, "exactly one Service should be rendered")
	for _, svc := range resources.Services {
		require.Len(t, svc.Spec.Ports, 1)
		assert.Nil(t, svc.Spec.Ports[0].AppProtocol, "appProtocol should be unset by default")
	}
}

func TestServiceAppProtocolH2C(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "trafficmirror", &helm.Options{
		SetValues: map[string]string{"service.appProtocol": "kubernetes.io/h2c"},
	})
	resources := helmtestutil.NewKubernetesResources(t, output)

	require.Len(t, resources.Services, 1, "exactly one Service should be rendered")
	for _, svc := range resources.Services {
		require.Len(t, svc.Spec.Ports, 1)
		require.NotNil(t, svc.Spec.Ports[0].AppProtocol, "appProtocol should be set")
		assert.Equal(t, "kubernetes.io/h2c", *svc.Spec.Ports[0].AppProtocol)
	}
}
