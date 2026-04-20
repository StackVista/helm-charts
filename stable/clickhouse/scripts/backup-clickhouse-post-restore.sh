#!/usr/bin/env bash
set -Eeuo pipefail

# ClickHouse has a problem with Mark cache after deletion of a table, it can throw an error 'Unknown codec family code'
# The solution is to drop the cache on ALL nodes

clickhouse-client -u "$CLICKHOUSE_ADMIN_USER" --password "$CLICKHOUSE_ADMIN_PASSWORD" --query "SYSTEM DROP MARK CACHE"
