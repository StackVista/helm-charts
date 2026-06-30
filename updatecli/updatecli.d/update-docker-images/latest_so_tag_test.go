package dockerimages_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repoRoot walks up from this file's directory until it finds go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, parent, dir, "go.mod not found walking up from %s", file)
		dir = parent
	}
}

// quayResponse is the shape of a quay.io tag list API page.
type quayResponse struct {
	Tags          []quayTag `json:"tags"`
	HasAdditional bool      `json:"has_additional"`
}

type quayTag struct {
	Name string `json:"name"`
}

// mockQuayServer starts an httptest server that serves paginated quay.io-style
// tag list responses.  pages is a slice of tag-name slices — one entry per
// page.  The server expects requests to match
// /api/v1/repository/stackstate/<image>/tag/ and uses the ?page= query param
// to select the right page (1-indexed).
func mockQuayServer(t *testing.T, pages [][]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		pageIdx := 0
		if pageStr != "" {
			for i, p := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"} {
				if p == pageStr {
					pageIdx = i
					break
				}
			}
		}
		if pageIdx >= len(pages) {
			http.Error(w, "page out of range", http.StatusBadRequest)
			return
		}
		tags := make([]quayTag, len(pages[pageIdx]))
		for i, name := range pages[pageIdx] {
			tags[i] = quayTag{Name: name}
		}
		resp := quayResponse{
			Tags:          tags,
			HasAdditional: pageIdx < len(pages)-1,
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
}

// runScript executes latest-so-tag.sh <image> against a mock server and
// returns trimmed stdout.
func runScript(t *testing.T, server *httptest.Server, image string) (string, error) {
	t.Helper()
	script := filepath.Join(repoRoot(t), "updatecli", "latest-so-tag.sh")
	cmd := exec.Command("bash", script, image)
	cmd.Env = append(os.Environ(), "QUAY_BASE_URL="+server.URL)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func TestLatestSoTagScript(t *testing.T) {
	tests := []struct {
		name  string
		pages [][]string
		want  string
	}{
		{
			name:  "single-digit soN — picks numerically highest",
			pages: [][]string{{"3.9.5-so7", "3.9.5-so8", "3.9.5-so9"}},
			want:  "3.9.5-so9",
		},
		{
			// The original bug: lexicographic ordering makes "so10" < "so9".
			// Numeric ordering must select so11.
			name:  "double-digit soN beats single-digit (the bug case)",
			pages: [][]string{{"3.9.5-so8", "3.9.5-so9", "3.9.5-so10", "3.9.5-so11"}},
			want:  "3.9.5-so11",
		},
		{
			// When the upstream version bumps, soN resets to 1.
			// The new upstream version must win even though its soN is lower.
			name:  "higher upstream version beats higher soN on older version",
			pages: [][]string{{"3.9.5-so11", "3.9.6-so1"}},
			want:  "3.9.6-so1",
		},
		{
			// Platform-specific tags (-amd64, -arm64) and unrelated tags must
			// be filtered out so only the multi-arch manifest tag is selected.
			name: "non-soN tags are filtered out",
			pages: [][]string{{
				"3.9.5-so11",
				"3.9.5-so11-amd64",
				"3.9.5-so11-arm64",
				"latest",
				"main",
				"3.9.5",
			}},
			want: "3.9.5-so11",
		},
		{
			// Pagination: the script must fetch all pages and pick the global
			// maximum, not just the maximum from the first page.
			name: "pagination — correct tag appears on second page",
			pages: [][]string{
				{"3.9.5-so8", "3.9.5-so9"},
				{"3.9.5-so10", "3.9.5-so11"},
			},
			want: "3.9.5-so11",
		},
		{
			// Three-page scenario mixing upstream versions across pages.
			name: "pagination across three pages with upstream version change",
			pages: [][]string{
				{"3.9.4-so5", "3.9.4-so6"},
				{"3.9.5-so9", "3.9.5-so10"},
				{"3.9.6-so1", "3.9.5-so11"},
			},
			want: "3.9.6-so1",
		},
		{
			// Only one tag available — script must still return it.
			name:  "single tag",
			pages: [][]string{{"3.9.5-so1"}},
			want:  "3.9.5-so1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := mockQuayServer(t, tt.pages)
			defer srv.Close()

			got, err := runScript(t, srv, "testimage")
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
