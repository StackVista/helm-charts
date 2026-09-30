package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// Keep the real Secret template, root helpers, chart metadata/defaults and
// packaged common dependency. Other platform workloads are covered by the
// full-chart reference tests and need no API simulation for this lookup test.
func secretLookupTestChart(t *testing.T, templatePath string) string {
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
	copyFile(templatePath)
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
	return chart
}

// Minimal discovery and Secret GET API for exercising Helm's real lookup.
// Unsupported reads return NotFound; mutations are rejected.
func secretLookupTestAPI(t *testing.T, secretPath string, existing *corev1.Secret) (string, func() []string) {
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
