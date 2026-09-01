#!/usr/bin/env sh
set -eu

MODE="${1:-hourly}"
MIN_HISTORY_SECONDS="${MIN_HISTORY_SECONDS:-86400}"

# Refuse the hourly backup when this instance looks freshly built while the destination
# still holds a good backup.
#
# vmbackup MIRRORS local storage onto the destination, deleting remote parts that are
# absent locally. Backing up an empty instance therefore does not add nothing, it erases
# the backup - which is what happened on 2026-08-18, within an hour of a from-scratch
# redeploy and before anyone had restored into it.
#
# Two things must be true at once before refusing: this instance has nothing worth backing
# up, AND the destination holds something worth keeping. Either one alone is harmless, so
# either one alone proceeds:
#
#   does local storage hold data older than MIN_HISTORY_SECONDS?
#     yes -> back up; ordinary running instance
#     no  -> does $S3_PREFIX-latest/ hold a COMPLETED backup?
#              empty, or partial only -> back up; nothing restorable there to lose
#              yes                    -> refuse
#              cannot tell            -> refuse
#
# The local check runs first so the steady-state path makes no S3 calls at all, and an S3
# or credentials fault cannot stop backups on healthy instances.
assert_not_fresh_instance() {
  # Operator-supplied, and the neighbouring server.retentionPeriod takes durations like
  # "3d", so a duration reaching here is a live mistake rather than a theoretical one.
  # Under set -eu comparing one numerically aborts the script, which would stop backups
  # on a healthy instance; screen it and fall through to the pre-guard behaviour instead.
  case "$MIN_HISTORY_SECONDS" in
    ''|*[!0-9]*)
      echo "WARNING: MIN_HISTORY_SECONDS='$MIN_HISTORY_SECONDS' is not a whole number of"
      echo "         seconds (durations such as '1d' are not accepted); skipping the"
      echo "         fresh-instance check and backing up as before"
      return 0
      ;;
  esac

  # Escape hatch for an instance deliberately started empty.
  if [ "$MIN_HISTORY_SECONDS" -le 0 ]; then
    return 0
  fi

  HISTORY_END=$(( $(date +%s) - MIN_HISTORY_SECONDS ))
  # The window must stay at or under 40 days. Up to and including v1.109.0, VictoriaMetrics
  # answers label queries from a per-day index only while the range spans at most
  # maxDaysForPerDaySearch (40) days, and beyond that silently falls back to the global
  # metric-name index, which ignores the time range and reports every metric the instance
  # has ever seen. Widening this to be more permissive therefore disables the guard rather
  # than relaxing it. Later versions drop that fast path, so the cap is correct on both;
  # keep it, because it is what makes this work on the version we pin. 40 days still covers
  # anything restorable, since retention is well under it.
  HISTORY=$(curl -sf -G "http://localhost:8428/api/v1/label/__name__/values" \
    --data-urlencode "start=$(( HISTORY_END - 3456000 ))" \
    --data-urlencode "end=$HISTORY_END" \
    --data-urlencode "limit=1") || {
      # Fail open: with the probe broken we cannot tell whether this instance is fresh, and
      # backups always ran before this guard existed. Refusing on a transient VM blip would
      # stop backups on every healthy instance instead.
      echo "WARNING: could not read local history from VictoriaMetrics, continuing with backup"
      return 0
    }

  HISTORY_COUNT=$(printf '%s' "$HISTORY" | jq '.data | length' 2>/dev/null) || HISTORY_COUNT=""
  if [ -z "$HISTORY_COUNT" ]; then
    echo "WARNING: could not parse local history response, continuing with backup"
    return 0
  fi
  if [ "$HISTORY_COUNT" -gt 0 ]; then
    return 0
  fi

  # Listed rather than head'd because an absent object and an unauthorised read both exit
  # 1, while a listing separates them: empty output means the prefix really is empty.
  DESTINATION=$(sts-toolbox aws s3 ls \
    --endpoint "$S3_ENDPOINT" \
    --region us-east-1 \
    --bucket "$BUCKET_NAME" \
    --prefix "$S3_PREFIX-latest/" 2>/dev/null) || {
      # Fail closed, opposite to the probe above, because what is known differs: local
      # storage is already confirmed empty, so only "is there anything to lose" is still
      # open. A skipped backup from an instance this young costs almost nothing, while
      # overwriting a real backup cannot be undone.
      echo "ERROR: local storage holds no data older than ${MIN_HISTORY_SECONDS}s and the contents"
      echo "       of s3://$BUCKET_NAME/$S3_PREFIX-latest could not be listed."
      exit 1
    }

  # Only a completed backup is worth protecting; a partial upload is not restorable.
  if ! printf '%s\n' "$DESTINATION" | grep -qxF 'backup_complete.ignore'; then
    return 0
  fi

  echo "ERROR: local storage holds no data older than ${MIN_HISTORY_SECONDS}s, but"
  echo "       s3://$BUCKET_NAME/$S3_PREFIX-latest holds a completed backup."
  echo "       Refusing to overwrite it: this instance looks freshly created and would"
  echo "       replace the backup with its own short history."
  echo "       Restore this instance first. If it was intentionally started empty, set"
  echo "       victoria-metrics-<n>.backup.minHistorySeconds=0 and restart the pod, or"
  echo "       remove the stale backup. Setting the variable on this container has no"
  echo "       effect: the crontab is written by the create-crontab init container."
  exit 1
}

case "$MODE" in
  hourly)
    assert_not_fresh_instance

    /vmbackup-prod \
      -storageDataPath=/storage/ \
      -snapshot.createURL=http://localhost:8428/snapshot/create \
      -dst=s3://"$BUCKET_NAME"/"$S3_PREFIX"-latest \
      -customS3Endpoint="$S3_ENDPOINT"
    ;;

  daily)
    TODAY=$(date +"%Y%m%d%H%M%S")

    # Server-side copy from latest to dated folder
    /vmbackup-prod \
      -origin=s3://"$BUCKET_NAME"/"$S3_PREFIX"-latest \
      -dst=s3://"$BUCKET_NAME"/"$S3_PREFIX-$TODAY" \
      -customS3Endpoint="$S3_ENDPOINT"

    # Retention policy (two tiers):
    #   1. Keep the KEEP_DAILY most recent backups unconditionally.
    #   2. From older backups, group by 7-day bucket and keep the most
    #      recent backup per bucket (newest-first), up to KEEP_WEEKLY.
    #   3. Delete everything else.
    #
    # Example with KEEP_DAILY=7, KEEP_WEEKLY=4:
    #   - days 1-7:  kept as daily backups
    #   - days 8-35: one backup per 7-day bucket kept (up to 4 buckets)
    #   - older:     deleted
    ALL_DAILIES=$(sts-toolbox aws s3 ls \
      --endpoint "$S3_ENDPOINT" \
      --region us-east-1 \
      --bucket "$BUCKET_NAME" \
      --prefix "$S3_PREFIX" \
      | grep -E '^victoria-metrics-[0-9]+-[0-9]{14}/$' \
      | sort -r)

    echo "Found backups: $ALL_DAILIES"

    DAILY_COUNT=0
    WEEKLY_COUNT=0
    SEEN_WEEK_KEYS=""
    for DAILY in $ALL_DAILIES; do
      DAILY_COUNT=$((DAILY_COUNT + 1))

      # Tier 1: always keep the most recent KEEP_DAILY backups
      if [ "$DAILY_COUNT" -le "$KEEP_DAILY" ]; then
        echo "Keeping daily backup: $DAILY (daily $DAILY_COUNT of $KEEP_DAILY)"
        continue
      fi

      # Tier 2: for older backups, group by 7-day bucket and keep the
      # most recent backup per bucket (since we iterate newest-first),
      # up to KEEP_WEEKLY buckets.
      # Extract YYYYMMDD from the folder name, convert to YYYY-MM-DD
      # for BusyBox date, then divide epoch seconds by 604800 (7 days)
      # to get a week bucket number.
      DATE_PART=$(echo "$DAILY" | grep -oE '[0-9]{14}' | cut -c1-8)
      FORMATTED_DATE=$(echo "$DATE_PART" | sed 's/\(.\{4\}\)\(.\{2\}\)\(.\{2\}\)/\1-\2-\3/')
      WEEK_KEY=$(( $(date -d "$FORMATTED_DATE" +%s) / 604800 ))

      KEEP=false
      case "$SEEN_WEEK_KEYS" in
        *"$WEEK_KEY"*)
          # Already kept a backup for this week — skip
          ;;
        *)
          WEEKLY_COUNT=$((WEEKLY_COUNT + 1))
          if [ "$WEEKLY_COUNT" -le "$KEEP_WEEKLY" ]; then
            SEEN_WEEK_KEYS="$SEEN_WEEK_KEYS $WEEK_KEY"
            KEEP=true
          fi
          ;;
      esac

      if [ "$KEEP" = true ]; then
        echo "Keeping weekly backup: $DAILY (week key $WEEK_KEY, weekly $WEEKLY_COUNT of $KEEP_WEEKLY)"
      else
        echo "Deleting old backup: $DAILY"
        sts-toolbox aws s3 delete \
          --endpoint "$S3_ENDPOINT" \
          --region us-east-1 \
          --bucket "$BUCKET_NAME" \
          --key "$(dirname "$S3_PREFIX")/${DAILY}" \
          --recursive
      fi
    done
    ;;

  *)
    echo "Usage: $0 {hourly|daily}"
    exit 1
    ;;
esac
