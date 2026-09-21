package test

import (
	"encoding/json"
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

// TestVictoriaMetricsScrapeFilters verifies that VictoriaMetrics and vmagent collect the same
// metric families. Their allowlists are declared in two unrelated places - the subchart's
// values and stackstate.vmagent.agentMetricsFilter — so they drift silently, and a family
// missing from one of them leaves gaps in dashboards built for both.
func TestVictoriaMetricsScrapeFilters(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/global_sizing_150_ha.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	scrapeFilter := func(stsName, container string) []string {
		sts, ok := resources.Statefulsets[stsName]
		require.True(t, ok, "%s StatefulSet should exist", stsName)

		raw, ok := sts.Spec.Template.Annotations[fmt.Sprintf("ad.stackstate.com/%s.instances", container)]
		require.True(t, ok, "%s should carry a scrape config for container %q", stsName, container)

		var instances []struct {
			Metrics []string `json:"metrics"`
		}
		require.NoError(t, json.Unmarshal([]byte(raw), &instances), "%s scrape config should be valid JSON", stsName)
		require.Len(t, instances, 1, "%s should declare exactly one scrape instance", stsName)
		return instances[0].Metrics
	}

	// The process_* entries are named individually rather than globbed: the rest of that family
	// duplicates container_cpu_usage and container_memory_rss, which the node agent already collects.
	expected := []string{"vm*", "go*", "process_open_fds", "process_max_fds", "process_cpu_cores_available"}

	vmagent := scrapeFilter("suse-observability-vmagent", "vmagent")
	assert.ElementsMatch(t, expected, vmagent, "vmagent scrape filter")

	for _, instance := range []string{"0", "1"} {
		name := "suse-observability-victoria-metrics-" + instance
		assert.ElementsMatch(t, expected, scrapeFilter(name, "victoria-metrics-"+instance+"-server"),
			"%s scrape filter should match vmagent's", name)
	}
}
