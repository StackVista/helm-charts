#!/usr/bin/env sh
set -euo

MODE="${1:-hourly}"

case "$MODE" in
  hourly)
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
