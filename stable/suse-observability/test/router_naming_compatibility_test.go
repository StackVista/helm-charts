package test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const canonicalRouterDeployment = "suse-observability-router"

type routerNamingFixture struct {
	Deployment appsv1.Deployment `json:"deployment"`
	Scripts    *corev1.ConfigMap `json:"scripts,omitempty"`
}

// Frozen pre-rename snapshots protect the entire Pod specification, including
// static configuration checksums, credentials and dynamic-mode mount selection.
func TestRouterDeploymentNamingFrozenCompatibility(t *testing.T) {
	chart := routerNamingTestChart(t, "..", false)
	checkRouterNamingFixtures(t, chart, true)
}

// Optional independent reproduction; never rewrite historical snapshots from
// the implementation under test.
func TestRouterNamingBaselineReproduction(t *testing.T) {
	source := os.Getenv("ROUTER_NAMING_BASELINE_CHART")
	if source == "" {
		t.Skip("set ROUTER_NAMING_BASELINE_CHART to a pre-rename chart with built dependencies")
	}
	checkRouterNamingFixtures(t, routerNamingTestChart(t, source, false), false)
}

func checkRouterNamingFixtures(t *testing.T, chart string, renamed bool) {
	t.Helper()
	for _, mode := range []string{"active", "maintenance", "automatic"} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", mode, upgrade), func(t *testing.T) {
				data, err := os.ReadFile(filepath.Join("testdata/router-renaming", mode+".json"))
				require.NoError(t, err)
				var expected routerNamingFixture
				require.NoError(t, json.Unmarshal(data, &expected))
				if renamed {
					expected.Deployment.Name = canonicalRouterDeployment
					if expected.Scripts != nil {
						allowRouterScriptSelectorChange(t, expected.Scripts)
					}
				}
				options := routerNamingTestOptions(t, routerNamingFixtureValues(mode))
				var args []string
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				output, err := helm.RenderTemplateE(t, options, chart, "nightly", nil, args...)
				require.NoError(t, err)
				resources := helmtestutil.NewKubernetesResources(t, output)
				require.Contains(t, resources.Deployments, expected.Deployment.Name)
				assert.Equal(t, expected.Deployment, resources.Deployments[expected.Deployment.Name])
				if expected.Scripts != nil {
					require.Contains(t, resources.ConfigMaps, expected.Scripts.Name)
					assert.Equal(t, *expected.Scripts, resources.ConfigMaps[expected.Scripts.Name])
				}
			})
		}
	}
}

// Restoring only the old naming expression allows complete before/after
// comparisons for customer settings that are not covered by frozen snapshots.
func TestRouterDeploymentNamingCompatibility(t *testing.T) {
	beforeChart := routerNamingTestChart(t, "..", true)
	afterChart := routerNamingTestChart(t, "..", false)
	for _, tc := range []struct {
		name, release, namespace, prefix string
		values                           map[string]string
	}{
		{name: "default", release: "suse-observability", prefix: "suse-observability"},
		{name: "custom-release", release: "nightly", prefix: "nightly-suse-observability"},
		{name: "namespace", release: "nightly", namespace: "tenant-a", prefix: "nightly-suse-observability"},
		{name: "fullname", release: "nightly", prefix: "customer", values: map[string]string{"fullnameOverride": "customer"}},
		{name: "affixes", release: "nightly", prefix: "g-pre-customer-post-end", values: map[string]string{
			"fullnameOverride": "customer", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "g-", "global.fullnameSuffix": "-end",
		}},
		{name: "global-override", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{"global.fullnameOverride": "global-service"}},
		{name: "long-name", release: "nightly", prefix: strings.Repeat("a", 54), values: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "monolithic", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{"stackstate.features.server.split": "false"}},
		{name: "argo", release: "nightly", prefix: "nightly-suse-observability", values: map[string]string{"deployment.compatibleWithArgoCD": "true"}},
	} {
		for _, mode := range []string{"active", "maintenance", "automatic"} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/upgrade=%t", tc.name, mode, upgrade), func(t *testing.T) {
					values := routerNamingFixtureValues(mode)
					maps.Copy(values, tc.values)
					options := routerNamingTestOptions(t, values)
					if tc.namespace != "" {
						options.KubectlOptions.Namespace = tc.namespace
					}
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					beforeOutput, err := helm.RenderTemplateE(t, options, beforeChart, tc.release, nil, args...)
					require.NoError(t, err)
					afterOutput, err := helm.RenderTemplateE(t, options, afterChart, tc.release, nil, args...)
					require.NoError(t, err)
					before := helmtestutil.NewKubernetesResources(t, beforeOutput)
					after := helmtestutil.NewKubernetesResources(t, afterOutput)
					legacy := tc.prefix + "-router"
					require.Contains(t, before.Deployments, legacy)
					expected := before.Deployments[legacy]
					expected.Name = canonicalRouterDeployment
					if legacy != canonicalRouterDeployment {
						delete(before.Deployments, legacy)
						assert.NotContains(t, after.Deployments, legacy)
					}
					before.Deployments[canonicalRouterDeployment] = expected
					assert.Equal(t, before.Deployments, after.Deployments)
					assert.Equal(t, before.ConfigMaps, after.ConfigMaps)
					assert.Equal(t, before.Secrets, after.Secrets)
					assert.Equal(t, before.Services, after.Services)
					assert.Equal(t, before.ServiceAccounts, after.ServiceAccounts)
					assert.Equal(t, before.Roles, after.Roles)
					assert.Equal(t, before.RoleBindings, after.RoleBindings)
					assert.Equal(t, before.Statefulsets, after.Statefulsets)
					assert.Equal(t, before.PersistentVolumeClaims, after.PersistentVolumeClaims)
					assert.Equal(t, before.Ingresses, after.Ingresses)
					assert.Equal(t, before.HTTPRoutes, after.HTTPRoutes)
					beforeContract := resourceNamingUpgradeContract(t, beforeOutput)
					if legacy != canonicalRouterDeployment {
						old := "Deployment/" + legacy
						beforeContract["Deployment/"+canonicalRouterDeployment] = beforeContract[old]
						delete(beforeContract, old)
					}
					assert.Equal(t, beforeContract, resourceNamingUpgradeContract(t, afterOutput))
					for _, service := range after.Services {
						if service.Labels["app.kubernetes.io/component"] == "router" {
							for key, value := range service.Spec.Selector {
								assert.Equal(t, value, expected.Spec.Template.Labels[key])
							}
						}
					}
				})
			}
		}
	}
}

// Allow namespace/release selection and a bounded watched rollout wait.
// All configuration, hook identities and other script text is frozen.
func allowRouterScriptSelectorChange(t *testing.T, scripts *corev1.ConfigMap) {
	t.Helper()
	scripts.Data = maps.Clone(scripts.Data)
	get := `kubectl get deployments -n "observability" -l "$router_selector" -o name`
	replacements := [][2]string{
		{`# shellcheck disable=SC2140
if kubectl get deployment "nightly-suse-observability-router" -n "observability"; then`,
			`# Release labels identify the router before and after a Deployment rename.
router_selector="app.kubernetes.io/component=router,app.kubernetes.io/instance=nightly"
router_deployments=$(` + get + `)
if [[ -n "$router_deployments" ]]; then`},
		{`kubectl rollout restart "deployment/nightly-suse-observability-router" -n "observability"`,
			`kubectl rollout restart deployment -n "observability" -l "$router_selector"`},
		{`  while ! kubectl rollout status "deployment/nightly-suse-observability-router" -n "observability"; do
    echo "."
    if ! kubectl get deployment "nightly-suse-observability-router" -n "observability"; then
      echo "Deployment went away, exiting"
      exit 0
    fi
    sleep 1
  done`,
			`  kubectl rollout status deployment -n "observability" -l "$router_selector" --timeout=120s`},
	}
	for _, key := range []string{"set-active.sh", "set-maintenance.sh"} {
		script := scripts.Data[key]
		for _, replacement := range replacements {
			require.Equal(t, 1, strings.Count(script, replacement[0]), key)
			script = strings.Replace(script, replacement[0], replacement[1], 1)
		}
		scripts.Data[key] = script
	}
}

func routerNamingFixtureValues(mode string) map[string]string {
	return map[string]string{
		"stackstate.components.router.image.tag":                      "naming-fixture",
		"stackstate.components.router.mode.status":                    mode,
		"stackstate.components.router.extraEnv.open.ROUTER_PUBLIC":    "fixture-public",
		"stackstate.components.router.extraEnv.secret.ROUTER_PRIVATE": "fixture-private",
	}
}

func routerNamingTestOptions(t *testing.T, values map[string]string) *helm.Options {
	t.Helper()
	options := apiResourceNameTestOptions(values)
	path, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	options.ValuesFiles = []string{path}
	return options
}

func routerNamingTestChart(t *testing.T, source string, legacy bool) string {
	t.Helper()
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS(source)))
	metadataPath := filepath.Join(chart, "Chart.yaml")
	data, err := os.ReadFile(metadataPath)
	require.NoError(t, err)
	metadata := string(data)
	for key, value := range map[string]string{"version": "0.0.0", "appVersion": "naming-fixture"} {
		pattern := regexp.MustCompile(`(?m)^` + key + `:.*$`)
		require.Len(t, pattern.FindAllString(metadata, -1), 1)
		metadata = pattern.ReplaceAllString(metadata, key+": "+value)
	}
	require.NoError(t, os.WriteFile(metadataPath, []byte(metadata), 0600))
	if legacy {
		path := filepath.Join(chart, "templates", "_names.tpl")
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		pattern := regexp.MustCompile(`(?s)\{\{- define "stackstate.router.deployment.fullname" -\}\}.*?\{\{- end -\}\}`)
		require.Len(t, pattern.FindAllString(string(data), -1), 1)
		content := pattern.ReplaceAllString(string(data), `{{- define "stackstate.router.deployment.fullname" -}}{{ template "common.fullname.short" . }}-router{{- end -}}`)
		require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	}
	return chart
}
