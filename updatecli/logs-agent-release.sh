#!/usr/bin/env bash
set -euo pipefail

candidate="${1:?Usage: logs-agent-release.sh <candidate> <current>}"
current="${2:?Usage: logs-agent-release.sh <candidate> <current>}"
release='^v[0-9]+\.[0-9]+\.[0-9]+-agent$'
if [[ ! "$candidate" =~ $release || "$current" == null ]]; then
  echo "Expected an agent release candidate and a current logs-agent image tag" >&2
  exit 1
fi

# The first compatible release must be adopted after chart/binary validation.
if [[ ! "$current" =~ $release ]]; then
  printf '%s' "$current"
  exit 0
fi

printf '%s\n%s\n' "$candidate" "$current" | sort -V | tail -n 1
