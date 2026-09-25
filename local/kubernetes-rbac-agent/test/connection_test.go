package test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
)

// Keep connection reads behind the helper so parent configuration applies to
// both the generated ConfigMaps and every consumer, including new templates.
func TestConnectionTemplatesUseConfigurationHelper(t *testing.T) {
	templatesDir := filepath.Join("..", "templates")
	directAccess := regexp.MustCompile(`\.Values\.(url|clusterName)\b`)
	comments := regexp.MustCompile(`(?s)\{\{-?\s*/\*.*?\*/\s*-?\}\}`)

	err := filepath.WalkDir(templatesDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || path == filepath.Join(templatesDir, "_configuration.tpl") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, access := range directAccess.FindAll(comments.ReplaceAll(content, nil), -1) {
			t.Errorf("%s: direct %s access bypasses parent configuration; resolve kubernetes-rbac-agent.connection with fromYaml and use $configuration instead", path, access)
		}
		return nil
	})
	require.NoError(t, err)
}

func TestConnectionStandaloneValues(t *testing.T) {
	for _, tc := range []struct {
		name, values, url, cluster, urlRef, clusterRef string
		externalOnly                                   bool
	}{
		{name: "literal", values: "url:\n  value: https://existing.example\nclusterName:\n  value: existing\n", url: "https://existing.example", cluster: "existing"},
		{name: "template", values: "url:\n  value: 'https://{{ .Release.Name }}.example'\nclusterName:\n  value: '{{ .Release.Namespace }}'\n", url: "https://orders.example", cluster: "observability"},
		{name: "external", values: "url:\n  fromConfigMap: '{{ .Release.Name }}-url'\nclusterName:\n  fromConfigMap: '{{ .Release.Name }}-cluster'\n", urlRef: "orders-url", clusterRef: "orders-cluster", externalOnly: true},
		{name: "external-and-inline", values: "url:\n  value: https://existing.example\n  fromConfigMap: external-url\nclusterName:\n  value: existing\n  fromConfigMap: external-cluster\n", url: "https://existing.example", cluster: "existing", urlRef: "external-url", clusterRef: "external-cluster"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "values.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tc.values), 0600))
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "orders", &helm.Options{
				ValuesFiles: []string{path}, KubectlOptions: &k8s.KubectlOptions{Namespace: "observability"},
			})
			resources := helmtestutil.NewKubernetesResources(t, output)
			if tc.externalOnly {
				assert.NotContains(t, resources.ConfigMaps, "orders-rbac-agent-url")
				assert.NotContains(t, resources.ConfigMaps, "orders-rbac-agent-cluster-name")
			} else {
				assert.Equal(t, map[string]string{"STS_URL": tc.url}, resources.ConfigMaps["orders-rbac-agent-url"].Data)
				assert.Equal(t, map[string]string{"STS_CLUSTER_NAME": tc.cluster}, resources.ConfigMaps["orders-rbac-agent-cluster-name"].Data)
			}
			urlRef, clusterRef := tc.urlRef, tc.clusterRef
			if urlRef == "" {
				urlRef = "orders-rbac-agent-url"
			}
			if clusterRef == "" {
				clusterRef = "orders-rbac-agent-cluster-name"
			}
			require.Contains(t, resources.Deployments, "orders-rbac-agent")
			refs := resources.Deployments["orders-rbac-agent"].Spec.Template.Spec.Containers[0].EnvFrom
			for _, name := range []string{urlRef, clusterRef} {
				assert.Contains(t, refs, corev1.EnvFromSource{
					ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: name}},
				})
			}
		})
	}
}

func TestConnectionStandaloneRequiresValues(t *testing.T) {
	for _, tc := range []struct{ name, values, message string }{
		{"missing-url", "clusterName:\n  value: example\n", "Url not defined"},
		{"missing-cluster", "url:\n  value: https://existing.example\n", "ClusterName not defined"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "values.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tc.values), 0600))
			err := helmtestutil.RenderHelmTemplateError(t, "orders", path)
			assert.Contains(t, err.Error(), tc.message)
		})
	}
}
