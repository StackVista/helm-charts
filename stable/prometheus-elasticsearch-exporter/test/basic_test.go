package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// TestPrometheusElasticsearchExporterBasic tests basic functionality
func TestPrometheusElasticsearchExporterBasic(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "prometheus-elasticsearch-exporter", "values/default.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Deployments, 1, "Should have exactly one Deployment")
}
