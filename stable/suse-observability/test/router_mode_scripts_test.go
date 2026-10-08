package test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// Execute the actual rendered hooks against a synthetic kubectl. No Kubernetes
// connection is made. The fixture models label filtering and a pending rollout,
// rather than only asserting that command strings contain the new name.
func TestRouterModeScriptsReachOldAndNewDeployments(t *testing.T) {
	output := helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly",
		routerNamingTestOptions(t, routerNamingFixtureValues("automatic")))
	resources := helmtestutil.NewKubernetesResources(t, output)
	scripts := resources.ConfigMaps["nightly-suse-observability-router-mode-scripts"]
	require.Len(t, scripts.Data, 2)
	legacy, canonical := "nightly-suse-observability-router", canonicalRouterDeployment
	for action, script := range scripts.Data {
		mode := strings.TrimSuffix(strings.TrimPrefix(action, "set-"), ".sh")
		expectedOutput := helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly",
			routerNamingTestOptions(t, routerNamingFixtureValues(mode)))
		expected := helmtestutil.NewKubernetesResources(t, expectedOutput)
		for _, tc := range []struct {
			name, behavior string
			names          []string
			fails          bool
		}{
			{name: "pre-upgrade-old", names: []string{legacy}},
			{name: "post-upgrade-new", names: []string{canonical}},
			{name: "overlap", names: []string{legacy, canonical}},
			{name: "fresh-install-empty"},
			{name: "wait-for-readiness", names: []string{canonical}, behavior: "pending"},
			{name: "deleted-while-waiting", names: []string{legacy}, behavior: "disappear"},
			{name: "list-error", names: []string{legacy}, behavior: "list-error", fails: true},
		} {
			t.Run(action+"/"+tc.name, func(t *testing.T) {
				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "kubectl"), []byte(routerModeKubectlFixture), 0700))
				scriptPath := filepath.Join(dir, action)
				require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0600))
				inventory := ""
				for _, name := range tc.names {
					inventory += name + "|observability|nightly|router\n"
				}
				// These must never be restarted even though they share one label.
				inventory += "foreign-release|observability|other|router\n"
				inventory += "foreign-namespace|tenant-a|nightly|router\n"
				inventory += "another-component|observability|nightly|api\n"
				require.NoError(t, os.WriteFile(filepath.Join(dir, "inventory"), []byte(inventory), 0600))
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, "bash", scriptPath)
				cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"),
					"ROUTER_FIXTURE_DIR="+dir, "ROUTER_FIXTURE_BEHAVIOR="+tc.behavior)
				result, err := cmd.CombinedOutput()
				if tc.fails {
					require.Error(t, err)
				} else {
					require.NoError(t, err, "%s", result)
				}
				require.NoError(t, ctx.Err(), "script did not finish")
				data, err := os.ReadFile(filepath.Join(dir, "applied.yaml"))
				require.NoError(t, err)
				var config corev1.ConfigMap
				require.NoError(t, yaml.Unmarshal(data, &config))
				assert.Equal(t, "nightly-suse-observability-router-automatic", config.Name)
				assert.Equal(t, "observability", config.Namespace)
				// Applying the chosen mode must preserve all generated Envoy data.
				assert.Equal(t, expected.ConfigMaps["nightly-suse-observability-router-"+mode].Data, config.Data)
				restarted, err := os.ReadFile(filepath.Join(dir, "restarted"))
				if len(tc.names) == 0 || tc.fails {
					assert.True(t, os.IsNotExist(err), "no router should be restarted")
				} else {
					require.NoError(t, err)
					assert.Equal(t, tc.names, strings.Fields(string(restarted)))
				}
				if tc.behavior == "pending" || tc.behavior == "disappear" {
					_, err := os.Stat(filepath.Join(dir, "pending"))
					require.NoError(t, err, "pending rollout path was not exercised")
					if tc.behavior == "pending" {
						assert.Contains(t, string(result), "Router mode")
					} else {
						assert.Contains(t, string(result), "Deployment went away")
					}
				}
			})
		}
	}
}

const routerModeKubectlFixture = `#!/bin/bash
set -eu
if [[ "$1" == apply ]]; then
  [[ "$*" == "apply -f -" ]]
  cat > "$ROUTER_FIXTURE_DIR/applied.yaml"
  exit 0
fi
operation="$1"
shift
if [[ "$operation" == rollout ]]; then
  action="$1"
  if [[ "$action" == status ]]; then
    [[ "$*" == *"--watch=false"* ]]
  fi
  shift
fi
[[ "$1" == deployment || "$1" == deployments ]]
shift
namespace=""
selector=""
while (( $# )); do
  case "$1" in
    -n) namespace="$2"; shift 2 ;;
    -l) selector="$2"; shift 2 ;;
    -o) [[ "$2" == name ]]; shift 2 ;;
    --watch=false) shift ;;
    *) exit 9 ;;
  esac
done
[[ "$namespace" == observability ]]
[[ "$selector" == app.kubernetes.io/component=router,app.kubernetes.io/instance=nightly ]]
if [[ "$operation" == get && "$ROUTER_FIXTURE_BEHAVIOR" == list-error ]]; then
  exit 7
fi
if [[ "$ROUTER_FIXTURE_BEHAVIOR" == disappear && -f "$ROUTER_FIXTURE_DIR/pending" ]]; then
  exit 0
fi
selected=""
while IFS='|' read -r name ns release component; do
  if [[ "$ns" == "$namespace" && "$release" == nightly && "$component" == router ]]; then
    selected+="deployment/$name"$'\n'
  fi
done < "$ROUTER_FIXTURE_DIR/inventory"
if [[ "$operation" == get ]]; then
  printf '%s' "$selected"
elif [[ "$action" == restart ]]; then
  while read -r target; do
    [[ -z "$target" ]] || printf '%s\n' "${target#deployment/}" >> "$ROUTER_FIXTURE_DIR/restarted"
  done <<< "$selected"
elif [[ "$action" == status ]]; then
  if [[ -n "$ROUTER_FIXTURE_BEHAVIOR" && ! -f "$ROUTER_FIXTURE_DIR/pending" ]]; then
    touch "$ROUTER_FIXTURE_DIR/pending"
    exit 1
  fi
else
  exit 9
fi
`
