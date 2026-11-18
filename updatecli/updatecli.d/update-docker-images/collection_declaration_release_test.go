package dockerimages_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectionDeclarationReleaseAdoption(t *testing.T) {
	script := filepath.Join(repoRoot(t), "updatecli/collection-declaration-release.sh")
	for _, tc := range []struct {
		name, tag string
		status    int
		wantError bool
		dryRun    bool
		unchanged bool
	}{
		{"release", "v0.0.100-agent", http.StatusOK, false, false, false},
		{"dry-run", "v0.0.100-agent", http.StatusOK, false, true, false},
		{"unchanged", "v0.0.100-agent", http.StatusOK, false, false, true},
		{"missing-export", "v0.0.100-agent", http.StatusNotFound, true, false, false},
		{"server-image", "v0.0.100", http.StatusOK, true, false, false},
		{"candidate-image", "test-sts-828-agent", http.StatusOK, true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const canonical = "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\n"
			var requests atomic.Uint32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				assert.Equal(t, "/v0.0.100/receiver/k8sresourcereceiver/collectiondeclaration/collectiondeclaration-crd.yaml", r.URL.Path)
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(canonical))
			}))
			defer server.Close()

			dir := t.TempDir()
			output := filepath.Join(dir, "collectiondeclaration.yaml")
			previous := "previous canonical export\n"
			if tc.unchanged {
				previous = canonical
			}
			require.NoError(t, os.WriteFile(output, []byte(previous), 0o644))
			command := exec.Command("bash", script, tc.tag, output)
			dryRun := "false"
			if tc.dryRun {
				dryRun = "true"
			}
			command.Env = append(os.Environ(), "COLLECTOR_SOURCE_URL="+server.URL, "DRY_RUN="+dryRun)
			logs, err := command.CombinedOutput()
			result, readErr := os.ReadFile(output)
			require.NoError(t, readErr)
			if tc.wantError {
				require.Error(t, err, "%s", logs)
				require.Equal(t, previous, string(result))
			} else {
				require.NoError(t, err, "%s", logs)
				if tc.dryRun {
					require.Equal(t, previous, string(result))
				} else {
					require.Equal(t, canonical, string(result))
				}
				require.Equal(t, uint32(1), requests.Load())
				if tc.unchanged {
					require.Empty(t, logs)
				} else {
					require.NotEmpty(t, logs)
				}
			}
			if tc.tag != "v0.0.100-agent" {
				require.Zero(t, requests.Load())
			}
		})
	}
}
