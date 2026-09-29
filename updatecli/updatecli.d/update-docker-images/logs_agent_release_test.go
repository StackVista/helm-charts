package dockerimages_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogsAgentReleaseAdoption(t *testing.T) {
	script := filepath.Join(repoRoot(t), "updatecli/logs-agent-release.sh")
	for _, tc := range []struct {
		name, candidate, current, want string
	}{
		{"retain-test-until-qualified", "v0.0.99-agent", "test-otel-logs-agent", "test-otel-logs-agent"},
		{"retain-digest-until-qualified", "v0.0.99-agent", "sha256:synthetic", "sha256:synthetic"},
		{"upgrade-qualified-release", "v0.0.100-agent", "v0.0.99-agent", "v0.0.100-agent"},
		{"retain-qualified-floor", "v0.0.9-agent", "v0.0.99-agent", "v0.0.99-agent"},
		{"same-release", "v0.0.99-agent", "v0.0.99-agent", "v0.0.99-agent"},
		{"reject-test-candidate", "test-otel-logs-agent", "v0.0.99-agent", ""},
		{"reject-server-candidate", "v0.0.100", "v0.0.99-agent", ""},
		{"reject-missing-current", "v0.0.100-agent", "null", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := exec.Command("bash", script, tc.candidate, tc.current).CombinedOutput()
			if tc.want == "" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err, "%s", output)
			require.Equal(t, tc.want, strings.TrimSpace(string(output)))
		})
	}
}
