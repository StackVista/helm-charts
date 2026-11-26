package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
)

// TestEsJavaOptsFix tests that esJavaOpts are correctly set in the StatefulSet, removing deprecated logging options
// This is needed because our small profiles in SUSE Observability override esJavaOpts and include the deprecated logging options
func TestEsJavaOptsFix(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "elasticsearch", "values/old-small-profile-es-java-opts.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	assert.Len(t, resources.Statefulsets, 1, "Should have exactly one StatefulSet")

	var statefulset appsv1.StatefulSet
	for _, ss := range resources.Statefulsets {
		statefulset = ss
		break
	}

	// Check container env for ES_JAVA_OPTS
	container := statefulset.Spec.Template.Spec.Containers[0]
	esJavaOptsFound := false
	for _, envVar := range container.Env {
		if envVar.Name == "ES_JAVA_OPTS" {
			esJavaOptsFound = true
			assert.Equal(t, "-Xmx1500m -Xms1500m -Des.allow_insecure_settings=true", envVar.Value, "Container ES_JAVA_OPTS should have correct value")
			break
		}
	}
	assert.True(t, esJavaOptsFound, "Container should have ES_JAVA_OPTS environment variable")
}
