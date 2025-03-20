#!/bin/bash

set -euo pipefail

clickhouse-client --host "$CLICKHOUSE_HOST" --user "$USERNAME" --password "$PASSWORD" --query "DROP TABLE IF EXISTS system.asynchronous_metric_log, system.asynchronous_metric_log_0,  system.backup_log, system.error_log, system.query_thread_log_0, system.metric_log, system.metric_log_0, system.part_log, system.part_log_0, system.processors_profile_log, system.query_log_0, system.query_log_1, system.query_views_log, system.query_views_log_0, system.text_log, system.session_log, system.trace_log, system.trace_log_0, system.crash_log, system.opentelemetry_span_log, system.zookeeper_log"
