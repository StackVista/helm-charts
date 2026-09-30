package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

func TestCollectorEndpointsFollowDedicatedConfigMapHelper(t *testing.T) {
	chart := filepath.Join(t.TempDir(), "chart")
	require.NoError(t, os.CopyFS(chart, os.DirFS("..")))
	path := filepath.Join(chart, "templates", "_names.tpl")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	require.NoError(t, err)
	_, err = file.WriteString(`
{{- define "stackstate.otelCollector.endpoints.configmap.fullname" -}}explicit-collector-endpoints{{- end -}}
`)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	values, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)
	for _, mode := range []string{"deployment", "daemonset", "statefulset"} {
		for _, split := range []bool{false, true} {
			for _, upgrade := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/split=%t/upgrade=%t", mode, split, upgrade), func(t *testing.T) {
					options := apiResourceNameTestOptions(map[string]string{
						"opentelemetry-collector.mode":             mode,
						"opentelemetry-collector.fullnameOverride": "custom-collector",
						"stackstate.features.server.split":         fmt.Sprint(split),
					})
					options.ValuesFiles = []string{values}
					var args []string
					if upgrade {
						args = append(args, "--is-upgrade")
					}
					before := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "nightly", options, args...)
					after, err := helm.RenderTemplateE(t, options, chart, "nightly", nil, args...)
					require.NoError(t, err)
					seen := map[string]bool{}
					expected := backupNamingDocuments(t, before, map[string]string{
						"ConfigMap/suse-observability-otel-collector": "explicit-collector-endpoints",
					}, seen)
					require.True(t, seen["ConfigMap/suse-observability-otel-collector"])
					rewriteCollectorEndpointReferences(expected)
					actual := backupNamingDocuments(t, after, nil, nil)
					// Includes router identifiers/addresses, workload configuration
					// checksums, selectors, StatefulSet names and volume references.
					assert.Equal(t, expected, actual)
				})
			}
		}
	}
}

func rewriteCollectorEndpointReferences(value interface{}) {
	switch node := value.(type) {
	case map[string]interface{}:
		if ref, ok := node["configMapKeyRef"].(map[string]interface{}); ok {
			if ref["name"] == "suse-observability-otel-collector" {
				ref["name"] = "explicit-collector-endpoints"
			}
		}
		for _, child := range node {
			rewriteCollectorEndpointReferences(child)
		}
	case []interface{}:
		for _, child := range node {
			rewriteCollectorEndpointReferences(child)
		}
	}
}

func TestCollectorSavedEnvironmentEntriesAreAppended(t *testing.T) {
	savedValues := filepath.Join(t.TempDir(), "saved-values.yaml")
	require.NoError(t, os.WriteFile(savedValues, []byte(`opentelemetry-collector:
  extraEnvs:
    - name: API_URL
      valueFrom:
        configMapKeyRef:
          name: suse-observability-otel-collector
          key: api.url
    - name: INTAKE_URL
      valueFrom:
        configMapKeyRef:
          name: suse-observability-otel-collector
          key: intake.url
`), 0600))
	for _, mode := range []string{"deployment", "daemonset", "statefulset"} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", mode, upgrade), func(t *testing.T) {
				options := apiResourceNameTestOptions(map[string]string{
					"opentelemetry-collector.mode": mode,
				})
				var args []string
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				options.ValuesFiles = append(options.ValuesFiles, savedValues)
				saved := helmtestutil.RenderHelmTemplateOptsNoErrorWithArgs(t, "nightly", options, args...)
				resources := helmtestutil.NewKubernetesResources(t, saved)
				var pod corev1.PodSpec
				switch mode {
				case "deployment":
					pod = resources.Deployments["suse-observability-otel-collector"].Spec.Template.Spec
				case "daemonset":
					pod = resources.DaemonSets["suse-observability-otel-collector-agent"].Spec.Template.Spec
				case "statefulset":
					pod = resources.Statefulsets["suse-observability-otel-collector"].Spec.Template.Spec
				}
				require.NotEmpty(t, pod.Containers)
				counts := map[string]int{}
				for _, env := range pod.Containers[0].Env {
					if env.Name == "API_URL" || env.Name == "INTAKE_URL" {
						counts[env.Name]++
					}
				}
				assert.Equal(t, map[string]int{"API_URL": 2, "INTAKE_URL": 2}, counts)
			})
		}
	}
}
