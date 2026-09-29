package test

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestAuthSecretLookupPreservesCredentials(t *testing.T) {
	const storedPassword = "$2b$10$N9qo8uLOickgx2ZMRZoMye.IUIrCxGvz9/6pN.XMlqL9hzTgDaPGy"
	const replacementPassword = "098f6bcd4621d373cade4e832627b4f6"
	renamedChart := authSecretLookupTestChart(t, true)
	chart := authSecretLookupTestChart(t, false)
	valuesFile, err := filepath.Abs("values/full.yaml")
	require.NoError(t, err)

	for _, tc := range []struct {
		name, release, namespace, secretName, provided string
		values                                         map[string]string
		renamed, missing, external                     bool
	}{
		{name: "default", release: "suse-observability", secretName: "suse-observability-auth"},
		{name: "custom-release-and-namespace", release: "nightly", namespace: "customer-space", secretName: "nightly-suse-observability-auth"},
		{name: "fullname-override", secretName: "custom-auth", values: map[string]string{"fullnameOverride": "custom"}},
		{name: "prefix-suffix", secretName: "global-pre-custom-post-end-auth", values: map[string]string{
			"fullnameOverride": "custom", "fullnamePrefix": "pre-", "fullnameSuffix": "-post",
			"global.fullnamePrefix": "global-", "global.fullnameSuffix": "-end",
		}},
		{name: "truncated", secretName: strings.Repeat("a", 54) + "-auth", values: map[string]string{"fullnameOverride": strings.Repeat("a", 70)}},
		{name: "dedicated-helper", secretName: "explicit-auth-secret", renamed: true},
		{name: "same-password", provided: storedPassword},
		{name: "explicit-replacement", provided: replacementPassword},
		{name: "fresh-with-password", missing: true, provided: replacementPassword},
		{name: "missing-secret-and-password", missing: true},
		{name: "external-secret", external: true, renamed: true},
	} {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", tc.name, upgrade), func(t *testing.T) {
				release, namespace, secretName := tc.release, tc.namespace, tc.secretName
				if release == "" {
					release = "nightly"
				}
				if namespace == "" {
					namespace = "observability"
				}
				if secretName == "" {
					secretName = "nightly-suse-observability-auth"
				}
				secretPath := "/api/v1/namespaces/" + namespace + "/secrets/" + secretName
				var existing *corev1.Secret
				if !tc.missing {
					existing = &corev1.Secret{
						TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
						ObjectMeta: metav1.ObjectMeta{
							Name: secretName, Namespace: namespace,
							Labels: map[string]string{"app.kubernetes.io/managed-by": "Helm"},
							Annotations: map[string]string{
								"meta.helm.sh/release-name": release, "meta.helm.sh/release-namespace": namespace,
							},
						},
						Data: map[string][]byte{"default_password": []byte(storedPassword)},
					}
				}
				kubeconfig, requests := authSecretLookupTestAPI(t, secretPath, existing)
				values := map[string]string{"stackstate.authentication.adminPassword": tc.provided}
				maps.Copy(values, tc.values)
				if tc.external {
					values["stackstate.authentication.fromExternalSecret"] = "customer-auth"
				}
				options := apiResourceNameTestOptions(values)
				options.KubectlOptions.Namespace = namespace
				options.ValuesFiles = []string{valuesFile}
				selectedChart := chart
				if tc.renamed {
					selectedChart = renamedChart
				}
				// Always override kubeconfig explicitly: this test only talks to
				// its read-only mock API, never the developer's current cluster.
				args := []string{"--dry-run=server", "--disable-openapi-validation", "--kubeconfig", kubeconfig, "--kube-context", "fixture"}
				if upgrade {
					args = append(args, "--is-upgrade")
				}
				output, err := helm.RenderTemplateE(t, options, selectedChart, release, nil, args...)
				if tc.external {
					require.NoError(t, err)
					for _, path := range requests() {
						assert.NotContains(t, path, "/secrets/"+secretName)
						assert.NotContains(t, path, "/secrets/explicit-auth-secret")
						assert.NotContains(t, path, "/secrets/customer-auth")
					}
					resources := helmtestutil.NewKubernetesResources(t, output)
					assert.NotContains(t, resources.Secrets, secretName)
					assert.NotContains(t, resources.Secrets, "explicit-auth-secret")
					assert.NotContains(t, resources.Secrets, "customer-auth")
					return
				}
				require.Contains(t, requests(), secretPath, "Lookup must use the same name and namespace as the Secret declaration")
				if tc.missing && tc.provided == "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), "Admin password is required for new installations")
					return
				}
				require.NoError(t, err)
				resources := helmtestutil.NewKubernetesResources(t, output)
				require.Contains(t, resources.Secrets, secretName)
				expectedPassword := storedPassword
				if tc.provided != "" {
					expectedPassword = tc.provided
				}
				assert.Equal(t, expectedPassword, string(resources.Secrets[secretName].Data["default_password"]))
			})
		}
	}
}

// Keep the real Secret template, root helpers, chart metadata/defaults and
// packaged common dependency. Other platform workloads are covered by the
// full-chart reference tests and need no API simulation for this lookup test.
func authSecretLookupTestChart(t *testing.T, rename bool) string {
	t.Helper()
	chart := filepath.Join(t.TempDir(), "chart")
	copyFile := func(relative string) {
		data, err := os.ReadFile(filepath.Join("..", relative))
		require.NoError(t, err)
		target := filepath.Join(chart, relative)
		require.NoError(t, os.MkdirAll(filepath.Dir(target), 0700))
		require.NoError(t, os.WriteFile(target, data, 0600))
	}
	copyFile("values.yaml")
	copyFile("templates/global/secret-auth.yaml")
	helpers, err := filepath.Glob("../templates/*.tpl")
	require.NoError(t, err)
	for _, path := range helpers {
		copyFile(filepath.Join("templates", filepath.Base(path)))
	}
	common, err := filepath.Glob("../charts/common-*.tgz")
	require.NoError(t, err)
	require.Len(t, common, 1, "Build chart dependencies before running lookup tests")
	copyFile(filepath.Join("charts", filepath.Base(common[0])))
	data, err := os.ReadFile("../Chart.yaml")
	require.NoError(t, err)
	var metadata map[string]interface{}
	require.NoError(t, yaml.Unmarshal(data, &metadata))
	dependencies := metadata["dependencies"].([]interface{})
	var retained []interface{}
	for _, dependency := range dependencies {
		if dependency.(map[string]interface{})["name"] == "common" {
			retained = append(retained, dependency)
		}
	}
	require.Len(t, retained, 1)
	metadata["dependencies"] = retained
	data, err = yaml.Marshal(metadata)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(chart, "Chart.yaml"), data, 0600))
	if rename {
		replaceAuthSecretName(t, chart)
	}
	return chart
}

// Minimal discovery and Secret GET API for exercising Helm's real lookup.
// Unsupported reads return NotFound; mutations are rejected.
func authSecretLookupTestAPI(t *testing.T, secretPath string, existing *corev1.Secret) (string, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			t.Errorf("Unexpected API mutation: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var result interface{}
		switch r.URL.Path {
		case "/version":
			result = map[string]string{"major": "1", "minor": "32", "gitVersion": "v1.32.0"}
		case "/api":
			result = metav1.APIVersions{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "APIVersions"}, Versions: []string{"v1"}}
		case "/apis":
			result = metav1.APIGroupList{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "APIGroupList"}, Groups: []metav1.APIGroup{}}
		case "/api/v1":
			result = metav1.APIResourceList{
				TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "APIResourceList"}, GroupVersion: "v1",
				APIResources: []metav1.APIResource{{Name: "secrets", SingularName: "secret", Namespaced: true, Kind: "Secret", Verbs: metav1.Verbs{"get", "list"}}},
			}
		case secretPath:
			if existing != nil {
				result = existing
			}
		}
		if result == nil {
			w.WriteHeader(http.StatusNotFound)
			result = metav1.Status{
				TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Status"},
				Status:   "Failure", Reason: metav1.StatusReasonNotFound, Code: http.StatusNotFound,
			}
		}
		if err := json.NewEncoder(w).Encode(result); err != nil {
			t.Errorf("Mock API response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "kubeconfig.yaml")
	config := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: fixture
  cluster:
    server: %s
contexts:
- name: fixture
  context:
    cluster: fixture
    user: fixture
current-context: fixture
users:
- name: fixture
  user: {}
`, server.URL)
	require.NoError(t, os.WriteFile(path, []byte(config), 0600))
	return path, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), paths...)
	}
}
