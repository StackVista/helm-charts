package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

// TestVictoriaMetricsScrapeAnnotationsMatchContainer verifies that every VictoriaMetrics
// instance carries autodiscovery annotations keyed for its own server container. Autodiscovery
// matches the key against the container name, so an instance annotated for another instance's
// container is never scraped and reports no metrics at all.
func TestVictoriaMetricsScrapeAnnotationsMatchContainer(t *testing.T) {
	for _, valuesFile := range []string{"values/global_sizing_150_ha.yaml", "values/global_sizing_50_nonha.yaml"} {
		t.Run(valuesFile, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", valuesFile)
			resources := helmtestutil.NewKubernetesResources(t, output)

			instances := 0
			for name, sts := range resources.Statefulsets {
				if !strings.Contains(name, "victoria-metrics") {
					continue
				}

				serverContainer := ""
				for _, c := range sts.Spec.Template.Spec.Containers {
					if strings.HasSuffix(c.Name, "-server") {
						serverContainer = c.Name
					}
				}
				require.NotEmpty(t, serverContainer, "%s should have a server container", name)

				annotations := sts.Spec.Template.Annotations
				instanceKey := fmt.Sprintf("ad.stackstate.com/%s.instances", serverContainer)
				assert.Contains(t, annotations, instanceKey,
					"%s must be scraped via its own container name %q", name, serverContainer)
				assert.Contains(t, annotations[instanceKey], ":8428/metrics",
					"%s scrape config should target the VictoriaMetrics server port", name)

				for key := range annotations {
					if !strings.HasPrefix(key, "ad.stackstate.com/") || !strings.HasSuffix(key, ".instances") {
						continue
					}
					container := strings.TrimSuffix(strings.TrimPrefix(key, "ad.stackstate.com/"), ".instances")
					if !strings.HasSuffix(container, "-server") {
						continue
					}
					assert.Equal(t, serverContainer, container,
						"%s carries a scrape annotation for another instance's container", name)
				}

				instances++
			}
			require.NotZero(t, instances, "expected at least one VictoriaMetrics StatefulSet")
		})
	}
}
