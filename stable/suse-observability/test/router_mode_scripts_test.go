package test

import (
	"bytes"
	"context"
	"fmt"
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
// connection is made. A watched pending rollout blocks until the test releases
// it; --watch=false returns success even while that rollout remains incomplete.
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
			{name: "wait-for-overlap", names: []string{legacy, canonical}, behavior: "pending"},
			{name: "deleted-while-waiting", names: []string{legacy}, behavior: "disappear", fails: true},
			{name: "rollout-timeout", names: []string{canonical}, behavior: "timeout", fails: true},
			{name: "watch-error", names: []string{canonical}, behavior: "watch-error", fails: true},
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
				var result bytes.Buffer
				cmd.Stdout, cmd.Stderr = &result, &result
				require.NoError(t, cmd.Start())
				finished := make(chan error, 1)
				go func() { finished <- cmd.Wait() }()
				if tc.behavior == "pending" {
					require.Eventually(t, func() bool {
						_, err := os.Stat(filepath.Join(dir, "pending"))
						return err == nil
					}, 2*time.Second, 10*time.Millisecond, "pending rollout path was not exercised")
					require.FileExists(t, filepath.Join(dir, "watching"), "status must watch the pending rollout")
					select {
					case err := <-finished:
						t.Fatalf("hook exited before the rollout completed: %v\n%s", err, result.String())
					default:
					}
					require.NoError(t, os.WriteFile(filepath.Join(dir, "ready"), nil, 0600))
				}
				err := <-finished
				if tc.fails {
					require.Error(t, err)
					assert.NotContains(t, result.String(), "+ echo 'Router mode", "failed rollouts must not report success")
				} else {
					require.NoError(t, err, "%s", result.String())
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
				if len(tc.names) == 0 || tc.behavior == "list-error" {
					assert.True(t, os.IsNotExist(err), "no router should be restarted")
				} else {
					require.NoError(t, err)
					assert.Equal(t, tc.names, strings.Fields(string(restarted)))
				}
				if len(tc.names) > 0 && tc.behavior != "list-error" {
					data, err := os.ReadFile(filepath.Join(dir, "status-targets"))
					require.NoError(t, err)
					assert.Equal(t, tc.names, strings.Fields(string(data)), "status must cover every restarted router")
				}
				if tc.behavior == "pending" {
					assert.FileExists(t, filepath.Join(dir, "completed"))
					assert.Contains(t, result.String(), "Router mode")
				}
			})
		}
	}
}

func TestRouterModeRolloutWatchPermissions(t *testing.T) {
	for _, argo := range []bool{false, true} {
		t.Run(fmt.Sprintf("argo=%t", argo), func(t *testing.T) {
			values := routerNamingFixtureValues("automatic")
			values["deployment.compatibleWithArgoCD"] = fmt.Sprint(argo)
			output := helmtestutil.RenderHelmTemplateOptsNoError(t, "nightly", routerNamingTestOptions(t, values))
			resources := helmtestutil.NewKubernetesResources(t, output)
			role, ok := resources.Roles["nightly-suse-observability-router-mode"]
			require.True(t, ok)
			require.Len(t, role.Rules, 2)
			assert.Equal(t, []string{"deployments"}, role.Rules[1].Resources)
			assert.ElementsMatch(t, []string{"get", "patch", "list", "watch"}, role.Rules[1].Verbs)
			jobs := routerModeJobsByAction(t, output)
			require.Len(t, jobs, 2)
			weightKey := "helm.sh/hook-weight"
			if argo {
				weightKey = "argocd.argoproj.io/sync-wave"
			}
			assert.Equal(t, "-3", role.Annotations[weightKey])
			for _, job := range jobs {
				assert.Equal(t, "-1", job.Annotations[weightKey], "watch permissions must precede the hook job")
			}
			require.NotNil(t, jobs["maintenance"].Spec.ActiveDeadlineSeconds)
			assert.Greater(t, *jobs["maintenance"].Spec.ActiveDeadlineSeconds, int64(120), "rollout timeout must be below the job deadline")
		})
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
  shift
fi
[[ "$1" == deployment || "$1" == deployments ]]
shift
namespace=""
selector=""
watch=true
timeout=""
while (( $# )); do
  case "$1" in
    -n) namespace="$2"; shift 2 ;;
    -l) selector="$2"; shift 2 ;;
    -o) [[ "$2" == name ]]; shift 2 ;;
    --watch=false) watch=false; shift ;;
    --timeout=*) timeout="${1#--timeout=}"; shift ;;
    *) exit 9 ;;
  esac
done
[[ "$namespace" == observability ]]
[[ "$selector" == app.kubernetes.io/component=router,app.kubernetes.io/instance=nightly ]]
if [[ "$operation" == get && "$ROUTER_FIXTURE_BEHAVIOR" == list-error ]]; then
  exit 7
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
  while read -r target; do
    [[ -z "$target" ]] || printf '%s\n' "${target#deployment/}" >> "$ROUTER_FIXTURE_DIR/status-targets"
  done <<< "$selected"
  if [[ "$watch" == false ]]; then
    touch "$ROUTER_FIXTURE_DIR/pending"
    exit 0
  fi
  [[ "$timeout" == 120s ]]
  touch "$ROUTER_FIXTURE_DIR/watching"
  case "$ROUTER_FIXTURE_BEHAVIOR" in
    pending)
      touch "$ROUTER_FIXTURE_DIR/pending"
      while [[ ! -f "$ROUTER_FIXTURE_DIR/ready" ]]; do sleep 0.01; done
      ;;
    disappear) echo "error: deployment was deleted while waiting" >&2; exit 1 ;;
    timeout) echo "error: timed out waiting for the condition" >&2; exit 1 ;;
    watch-error) echo "error: cannot watch deployments" >&2; exit 1 ;;
  esac
  touch "$ROUTER_FIXTURE_DIR/completed"
else
  exit 9
fi
`
