package stackpacks_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runCoreResolver(t *testing.T, pages [][]string, branch string) (string, error) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/repository/stackstate/core-stackpacks/tag/", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("onlyActiveTags"))
		if pages == nil {
			http.Error(w, "registry unavailable", http.StatusServiceUnavailable)
			return
		}
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if !assert.NoError(t, err) || page < 1 || page > len(pages) {
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}
		tags := make([]map[string]string, len(pages[page-1]))
		for i, tag := range pages[page-1] {
			tags[i] = map[string]string{"name": tag}
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
			"tags": tags, "has_additional": page < len(pages),
		}))
	}))
	defer server.Close()

	script, err := filepath.Abs("../../latest-core-stackpacks-tag.sh")
	require.NoError(t, err)
	command := exec.Command("bash", script)
	command.Env = append(os.Environ(),
		"QUAY_BASE_URL="+server.URL,
		"STACKPACKS_CORE_SOURCE_BRANCH="+branch,
	)
	output, err := command.Output()
	return strings.TrimSpace(string(output)), err
}

func TestLatestCoreStackpacksTag(t *testing.T) {
	base := "20261106090329-main-f1c8402"
	tests := []struct {
		name   string
		pages  [][]string
		branch string
		want   string
	}{
		{
			name:   "original tag without retries",
			pages:  [][]string{{base}},
			branch: "main",
			want:   base,
		},
		{
			name:   "numeric retry beats original and single digit",
			pages:  [][]string{{base + "-retry-10", base, base + "-retry-2"}},
			branch: "main",
			want:   base + "-retry-10",
		},
		{
			name:   "reversed registry order has the same result",
			pages:  [][]string{{base + "-retry-2", base, base + "-retry-10"}},
			branch: "main",
			want:   base + "-retry-10",
		},
		{
			name:   "retry on a later page wins",
			pages:  [][]string{{base + "-retry-2"}, {base + "-retry-10"}, {base}},
			branch: "main",
			want:   base + "-retry-10",
		},
		{
			name:   "newer timestamp wins before retry attempt",
			pages:  [][]string{{base + "-retry-999", "20261106100329-main-a1b2c3d"}},
			branch: "main",
			want:   "20261106100329-main-a1b2c3d",
		},
		{
			name: "other branches and malformed tags are excluded",
			pages: [][]string{{
				base,
				"20261106100329-sts-892-a1b2c3d-retry-10",
				"20261106100329-main-a1b2c3d-amd64",
				"20261106100329-main-a1b2c3d-arm64",
				"20261106100329-main-a1b2c3d-retry-x",
				"20261106100329-main-ZZZZZZZ",
				"9999-main-a1b2c3d",
				"latest",
			}},
			branch: "main",
			want:   base,
		},
		{
			name:   "hyphenated branch is matched exactly",
			pages:  [][]string{{base, "20261106100329-sts-892-a1b2c3d-retry-10"}},
			branch: "sts-892",
			want:   "20261106100329-sts-892-a1b2c3d-retry-10",
		},
		{
			name: "branch characters are literal",
			pages: [][]string{{
				"20261106100329-feature.fix-a1b2c3d",
				"20261106110329-featureXfix-a1b2c3d",
			}},
			branch: "feature.fix",
			want:   "20261106100329-feature.fix-a1b2c3d",
		},
		{
			name:   "equal numeric keys use a stable tag tie break",
			pages:  [][]string{{base, "20261106090329-main-a1b2c3d"}},
			branch: "main",
			want:   base,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := runCoreResolver(t, test.pages, test.branch)
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestLatestCoreStackpacksTagRejectsMissingBranch(t *testing.T) {
	got, err := runCoreResolver(t, [][]string{{"latest", "20261106090329-sts-892-f1c8402"}}, "main")
	require.Error(t, err)
	assert.Empty(t, got)
}

func TestLatestCoreStackpacksTagRejectsRegistryFailure(t *testing.T) {
	got, err := runCoreResolver(t, nil, "main")
	require.Error(t, err)
	assert.Empty(t, got)
}
