package test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
)

// Snapshots predate this rename. Compare the complete Deployment so selectors,
// probes, Elasticsearch URLs, security settings and credential references stay
// unchanged when Helm replaces the controller.
func TestExporterDeploymentNamingCompatibility(t *testing.T) {
	const canonical = "suse-observability-prometheus-elasticsearch-exporter"
	const prefix = "elasticsearch.prometheus-elasticsearch-exporter."
	for _, tc := range []struct {
		name, release, namespace, fullname, app string
		tls                                     bool
		values                                  map[string]string
	}{
		{name: "default", release: "suse-observability"},
		{name: "nightly", release: "nightly"},
		{name: "second-namespace", release: "nightly", namespace: "tenant-a"},
		{name: "parent-overrides", release: "nightly", values: map[string]string{
			"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end", "global.fullnameOverride": "global-only",
		}},
		{name: "monolithic", release: "nightly", values: map[string]string{"stackstate.features.server.split": "false"}},
		{name: "split-workers", release: "nightly", values: map[string]string{
			"stackstate.components.receiver.split.enabled": "true", "stackstate.components.correlate.split.enabled": "true",
		}},
		{name: "argo", release: "nightly", values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
		{name: "long-release", release: strings.Repeat("x", 50), fullname: strings.Repeat("x", 50) + "-prometheus-e"},
		{name: "release-contains-chart", release: "tenant-prometheus-elasticsearch-exporter", fullname: "tenant-prometheus-elasticsearch-exporter"},
		{name: "generated-account-and-tls", release: "nightly", tls: true},
		{name: "exporter-fullname", release: "nightly", fullname: "customer-exporter", tls: true,
			values: map[string]string{prefix + "fullnameOverride": "customer-exporter"}},
		{name: "exporter-name", release: "nightly", fullname: "nightly-customer", app: "customer", tls: true,
			values: map[string]string{prefix + "nameOverride": "customer"}},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				fixture := "default"
				values := componentResourceNameTestValues()
				values[prefix+"image.tag"] = "naming-fixture"
				if tc.tls {
					fixture = "tls"
					maps.Copy(values, exporterNamingTLSValues())
				}
				maps.Copy(values, tc.values)
				options := apiResourceNameTestOptions(values)
				if tc.namespace != "" {
					options.KubectlOptions.Namespace = tc.namespace
				}
				data, err := os.ReadFile(filepath.Join("testdata/exporter-renaming", fixture+".json"))
				require.NoError(t, err)
				var expected appsv1.Deployment
				require.NoError(t, json.Unmarshal(data, &expected))
				legacy := tc.fullname
				if legacy == "" {
					legacy = tc.release + "-prometheus-elasticsearch-exporter"
				}
				app := tc.app
				if app == "" {
					app = "prometheus-elasticsearch-exporter"
				}
				expected.Name = canonical
				expected.Labels["release"] = tc.release
				expected.Labels["app"] = app
				expected.Spec.Selector.MatchLabels["release"] = tc.release
				expected.Spec.Selector.MatchLabels["app"] = app
				expected.Spec.Template.Labels["release"] = tc.release
				expected.Spec.Template.Labels["app"] = app
				if tc.tls {
					expected.Spec.Template.Spec.ServiceAccountName = legacy
					for i := range expected.Spec.Template.Spec.Volumes {
						if expected.Spec.Template.Spec.Volumes[i].Name == "ssl" {
							expected.Spec.Template.Spec.Volumes[i].Secret.SecretName = legacy + "-cert"
						}
					}
				}
				var args []string
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				output := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, tc.release, options, args...)
				resources := helmtestutil.NewKubernetesResources(t, output)
				require.Contains(t, resources.Deployments, canonical)
				actual := resources.Deployments[canonical]
				delete(actual.Labels, "chart")
				assert.Equal(t, expected, actual, "Only the Deployment metadata name may change")
				if legacy != canonical {
					assert.NotContains(t, resources.Deployments, legacy)
				}
				require.Contains(t, resources.Services, legacy, "The existing Service DNS name is retained")
				for key, value := range resources.Services[legacy].Spec.Selector {
					assert.Equal(t, value, actual.Spec.Template.Labels[key], "Service must still select exporter Pods")
				}
				if tc.tls {
					require.Contains(t, resources.ServiceAccounts, legacy)
					require.Contains(t, resources.Secrets, legacy+"-cert")
					assert.Equal(t, map[string][]byte{
						"ca.pem": []byte("fixture-ca"), "client.pem": []byte("fixture-client"), "client.key": []byte("fixture-key"),
					}, resources.Secrets[legacy+"-cert"].Data)
				}
			})
		}
	}
}

func exporterNamingTLSValues() map[string]string {
	const prefix = "elasticsearch.prometheus-elasticsearch-exporter."
	return map[string]string{
		prefix + "serviceAccount.create":           "true",
		prefix + "es.ssl.enabled":                  "true",
		prefix + "es.ssl.client.enabled":           "true",
		prefix + "es.ssl.ca.pem":                   "fixture-ca",
		prefix + "es.ssl.client.pem":               "fixture-client",
		prefix + "es.ssl.client.key":               "fixture-key",
		prefix + "secretMounts[0].name":            "external-tls",
		prefix + "secretMounts[0].secretName":      "customer-tls",
		prefix + "secretMounts[0].path":            "/customer-tls",
		prefix + "envFromSecret":                   "customer-environment",
		prefix + "extraEnvSecrets.TEST_KEY.secret": "customer-key",
		prefix + "extraEnvSecrets.TEST_KEY.key":    "key",
	}
}
