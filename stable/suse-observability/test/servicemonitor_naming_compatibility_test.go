package test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	"k8s.io/apimachinery/pkg/labels"
)

// Frozen monitors come from 0524ff4ca, before ServiceMonitor renaming. Only names
// become canonical; discovery labels, namespace selection and scrape specs stay
// compatible, including with legacy fullname settings.
func TestServiceMonitorNamingCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, release, namespace, fixture, valuesFile string
		set                                           map[string]string
	}{
		{name: "default", release: "suse-observability"},
		{name: "nightly", release: "nightly"},
		{name: "second-namespace", release: "nightly", namespace: "tenant-a"},
		{name: "fullname-override", release: "nightly",
			set: map[string]string{"fullnameOverride": "Customer"}},
		{name: "prefix-suffix", release: "nightly", set: map[string]string{
			"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "local-prefix", release: "nightly", set: map[string]string{"fullnamePrefix": "pre-"}},
		{name: "local-suffix", release: "nightly", set: map[string]string{"fullnameSuffix": "-post"}},
		{name: "global-prefix", release: "nightly", set: map[string]string{"global.fullnamePrefix": "g-"}},
		{name: "global-suffix", release: "nightly", set: map[string]string{"global.fullnameSuffix": "-end"}},
		{name: "global-override", release: "nightly", set: map[string]string{"global.fullnameOverride": "global-only"}},
		{name: "long-release", release: strings.Repeat("x", 50)},
		{name: "long-override", release: "nightly", set: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "monolithic", release: "nightly", fixture: "mono",
			set: map[string]string{"stackstate.features.server.split": "false"}},
		{name: "split-workers", release: "nightly", fixture: "workers", set: map[string]string{
			"stackstate.components.receiver.split.enabled": "true", "stackstate.components.correlate.split.enabled": "true",
		}},
		{name: "ha", release: "nightly", valuesFile: "values/global_sizing_150_ha.yaml"},
		{name: "nonha", release: "nightly", fixture: "mono", valuesFile: "values/global_sizing_50_nonha.yaml"},
		{name: "argo", release: "nightly", set: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				fixture := tc.fixture
				if fixture == "" {
					fixture = "split"
				}
				data, err := os.ReadFile(filepath.Join("testdata/servicemonitor-renaming", fixture+".json"))
				require.NoError(t, err)
				var expected map[string]monitoringv1.ServiceMonitor
				require.NoError(t, json.Unmarshal(data, &expected))

				values := serviceMonitorCompatibilityValues()
				maps.Copy(values, tc.set)
				options := apiResourceNameTestOptions(values)
				if tc.valuesFile != "" {
					options.ValuesFiles = append(options.ValuesFiles, tc.valuesFile)
				}
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				namespace := options.KubectlOptions.Namespace
				for component, monitor := range expected {
					monitor.Name = "suse-observability-" + component
					monitor.Labels["app.kubernetes.io/instance"] = tc.release
					if _, found := monitor.Spec.Selector.MatchLabels["app.kubernetes.io/instance"]; found {
						monitor.Spec.Selector.MatchLabels["app.kubernetes.io/instance"] = tc.release
					}
					monitor.Spec.NamespaceSelector.MatchNames = []string{namespace}
					expected[component] = monitor
				}

				var args []string
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, tc.release, options, args...)
				resources := helmtestutil.NewKubernetesResources(t, output)
				actual := map[string]monitoringv1.ServiceMonitor{}
				for _, monitor := range resources.ServiceMonitors {
					component := monitor.Labels["app.kubernetes.io/component"]
					// Subchart monitors have different app names or components.
					if monitor.Labels["app.kubernetes.io/name"] != "suse-observability" {
						continue
					}
					require.NotContains(t, actual, component, "Each component must have one monitor")
					delete(monitor.Labels, "helm.sh/chart")
					delete(monitor.Labels, "app.kubernetes.io/version")
					actual[component] = monitor
				}
				assert.Equal(t, expected, actual, "Only reviewed monitor metadata names may change")

				// Prometheus selects these objects by labels, independently of names.
				for _, target := range []string{"platform-monitoring", "backup-monitoring"} {
					selector := labels.SelectorFromSet(labels.Set{
						"prometheus": target, "fixture-scope": "preserved",
					})
					for component, monitor := range expected {
						require.Contains(t, actual, component)
						assert.Equal(t, selector.Matches(labels.Set(monitor.Labels)),
							selector.Matches(labels.Set(actual[component].Labels)), component)
					}
				}
			})
		}
	}
}

func serviceMonitorCompatibilityValues() map[string]string {
	values := componentResourceNameTestValues()
	maps.Copy(values, map[string]string{
		"stackstate.components.all.metrics.enabled":                                    "true",
		"stackstate.components.all.metrics.servicemonitor.enabled":                     "true",
		"stackstate.components.all.metrics.servicemonitor.additionalLabels.prometheus": "platform-monitoring",
		"s3proxy.metrics.enabled":                                                      "true",
		"s3proxy.metrics.servicemonitor.enabled":                                       "true",
		"s3proxy.metrics.servicemonitor.additionalLabels.prometheus":                   "backup-monitoring",
		"stackstate.k8sAuthorization.enabled":                                          "true",
		"global.backup.enabled":                                                        "true",
		"backup.configuration.enabled":                                                 "true",
		"servicemonitor.namespace":                                                     "monitoring",
		"servicemonitor.annotations.fixture-owner":                                     "observability",
		"servicemonitor.labels.fixture-scope":                                          "preserved",
	})
	return values
}
