#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
test_dir="$(mktemp -d)"
trap 'rm -rf "$test_dir"' EXIT
cat > "$test_dir/skopeo" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
[[ "$*" == 'inspect --no-tags --format {{ index .Labels "com.stackstate.stackgraph.version" }} docker://quay.io/stackstate/stackstate-server:server-test-2.5' ]]
[[ "${INSPECT_FAIL:-false}" != true ]] || exit 1
printf '%s\n' "$LABEL_VALUE"
MOCK
chmod +x "$test_dir/skopeo"
export PATH="$test_dir:$PATH"

for version in 8.3.12 8.3.13 8.3.14 8.3.15 8.4.0 branch-build; do
  export LABEL_VALUE="$version"
  expected="$version"
  [[ "$version" != 8.3.13 ]] || expected=8.3.14
  [[ "$(bash "$script_dir/stackgraph-image-version.sh" server-test-2.5)" == "$expected" ]]
done
for label in '' '<no value>' null; do
  export LABEL_VALUE="$label"
  if bash "$script_dir/stackgraph-image-version.sh" server-test-2.5 > /dev/null 2>&1; then
    echo "Missing label unexpectedly accepted" >&2
    exit 1
  fi
done
export LABEL_VALUE=8.3.13 INSPECT_FAIL=true
if bash "$script_dir/stackgraph-image-version.sh" server-test-2.5 > /dev/null 2>&1; then
  echo "Registry failure unexpectedly accepted" >&2
  exit 1
fi
echo "StackGraph image version tests passed"
