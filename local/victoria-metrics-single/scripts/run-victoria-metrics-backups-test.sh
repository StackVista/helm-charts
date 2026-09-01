#!/usr/bin/env bash
# Self-check for the fresh-instance guard in run-victoria-metrics-backups.sh.
# Stubs curl, sts-toolbox and vmbackup-prod, then asserts each guard branch.
# shellcheck disable=SC2016 # stub bodies are written verbatim, expansion happens when they run
set -euo pipefail

HERE=$(cd "$(dirname "$0")" && pwd)
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

# The script calls /vmbackup-prod by absolute path, which cannot be stubbed on PATH.
sed 's|/vmbackup-prod|vmbackup-prod|g' "$HERE/run-victoria-metrics-backups.sh" >"$WORK/backup.sh"
chmod +x "$WORK/backup.sh"

mkdir -p "$WORK/bin"
printf '#!/usr/bin/env sh\necho BACKUP_RAN\n' >"$WORK/bin/vmbackup-prod"
printf '#!/usr/bin/env sh\nfor a in "$@"; do echo "$a"; done >"$CURL_ARGS"\n[ "${STUB_CURL_FAIL:-0}" = 1 ] && exit 22\nprintf %%s "$STUB_HISTORY"\n' >"$WORK/bin/curl"
printf '#!/usr/bin/env sh\nprintf %%s "${STUB_LISTING:-}"\nexit "${STUB_LS_STATUS:-0}"\n' >"$WORK/bin/sts-toolbox"
chmod +x "$WORK"/bin/*

export PATH="$WORK/bin:$PATH"
export BUCKET_NAME=bucket S3_PREFIX=victoria-metrics-0 S3_ENDPOINT=http://s3proxy:9000
export CURL_ARGS="$WORK/curl-args"

HAS_HISTORY='{"status":"success","data":["some_metric"]}'
NO_HISTORY='{"status":"success","data":[]}'

# The guard reads the history probe through jq and fails open when it cannot parse the
# response, so with jq missing or broken every case here still "proceeds" and only the two
# refusals fail. That reads as two broken assertions rather than an absent dependency, so
# say so instead. Exercised rather than looked up on PATH, because a jq that resolves but
# does not run produces the same misleading result.
printf '%s' "$NO_HISTORY" | jq -e '.data | length == 0' >/dev/null 2>&1 || {
  echo "FAIL - working jq required to exercise the guard; without it the guard fails open"
  exit 1
}

failures=0
check() { # check <ran|refused> <label> [env assignments...]
  local want=$1 label=$2 out status=0
  shift 2
  out=$(env "$@" "$WORK/backup.sh" hourly 2>&1) || status=$?
  local got=ran
  [ "$status" -eq 0 ] || got=refused
  if [ "$want" = ran ] && ! printf '%s' "$out" | grep -q BACKUP_RAN; then got=refused; fi
  if [ "$got" = "$want" ]; then
    echo "ok   - $label"
  else
    echo "FAIL - $label (wanted $want, got $got)"
    printf '%s\n' "$out" | sed 's/^/       /'
    failures=$((failures + 1))
  fi
}

# Listing shapes as observed from sts-toolbox against a real backup prefix.
COMPLETE=$'backup_complete.ignore\nbackup_metadata.ignore\ndata/\nmetadata/'
PARTIAL=$'backup_metadata.ignore\ndata/'

# Every hourly run on a healthy instance. Returns before reaching S3, so a broken bucket
# or expired credentials cannot stop backups on instances that have real data.
check ran     "local history present, backup proceeds" \
  "STUB_HISTORY=$HAS_HISTORY"

# The 2026-08-18 incident: rebuilt instance, good backup still at the destination, and
# vmbackup about to mirror an hour of data over it. The one case the guard exists for.
check refused "fresh instance over a completed backup is refused" \
  "STUB_HISTORY=$NO_HISTORY" "STUB_LISTING=$COMPLETE"

# A brand-new tenant looks identical to a wiped one from local storage alone. The empty
# destination is what separates them, so its first backup must not be blocked.
check ran     "fresh instance with an empty destination proceeds" \
  "STUB_HISTORY=$NO_HISTORY" STUB_LISTING=

# Data at the destination but no backup_complete.ignore means an interrupted upload. It
# cannot be restored from, so there is nothing of value to refuse on behalf of.
check ran     "fresh instance over an incomplete backup proceeds" \
  "STUB_HISTORY=$NO_HISTORY" "STUB_LISTING=$PARTIAL"

# The next two probes fail in opposite directions on purpose, according to what is already
# known when the failure happens.
#
# Here local storage is confirmed empty, so the destructive half is established and only
# "is there anything to lose" is unknown. Losing one backup from an instance this young is
# cheap; overwriting a real one is not.
check refused "fresh instance with an unlistable destination is refused" \
  "STUB_HISTORY=$NO_HISTORY" STUB_LS_STATUS=1

# Here nothing is established: with the local probe broken the instance may well be
# healthy. Backups always ran before this guard existed, so a transient VM fault must not
# halt them estate-wide. Note the destination holds a completed backup and it still runs.
check ran     "unreadable local history falls back to proceeding" \
  STUB_CURL_FAIL=1 "STUB_LISTING=$COMPLETE"

# The documented escape hatch for an instance deliberately started empty: 0 skips the guard
# even in the exact configuration that case 2 refuses.
check ran     "MIN_HISTORY_SECONDS=0 disables the guard" \
  "STUB_HISTORY=$NO_HISTORY" "STUB_LISTING=$COMPLETE" MIN_HISTORY_SECONDS=0

# A duration is an easy thing to write here, because server.retentionPeriod next door takes
# one. It must degrade to the pre-guard behaviour: comparing it numerically under set -eu
# aborts the script, which stops backups on a healthy instance rather than protecting one.
check ran     "a duration-style threshold warns instead of aborting the run" \
  "STUB_HISTORY=$NO_HISTORY" "STUB_LISTING=$COMPLETE" MIN_HISTORY_SECONDS=1d

# Everything above stubs curl, so it proves the branches wire up but never that the query
# asks the right question - a probe returning every metric regardless of range satisfies all
# seven. That is not hypothetical: VictoriaMetrics serves label queries from a per-day index
# only while the range spans at most 40 days (maxDaysForPerDaySearch) and past that silently
# switches to the global index, which ignores the range entirely. A 400-day window shipped
# in review for exactly this reason, so assert the width the stub was asked for.
: >"$CURL_ARGS"
env "STUB_HISTORY=$HAS_HISTORY" "$WORK/backup.sh" hourly >/dev/null 2>&1 || true
probe_start=$(sed -n 's/^start=//p' "$CURL_ARGS")
probe_end=$(sed -n 's/^end=//p' "$CURL_ARGS")
if [ -z "$probe_start" ] || [ -z "$probe_end" ]; then
  echo "FAIL - history probe recorded no start/end window"
  failures=$((failures + 1))
elif [ "$(( probe_end - probe_start ))" -gt 3456000 ]; then
  echo "FAIL - history probe window is $(( (probe_end - probe_start) / 86400 ))d, over the 40d per-day index limit"
  failures=$((failures + 1))
else
  echo "ok   - history probe window stays within the 40d per-day index limit"
fi

[ "$failures" -eq 0 ] || exit 1
echo "all guard checks passed"
