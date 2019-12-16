#!/bin/bash

[[ -n "${TRACE+x}" ]] && set -x

commonFlags="--bootstrap-server ${KAFKA_BROKERS} --config retention.ms=86400000 --create --force --replication-factor 1"
topics=("sts_connection_beat_events" "sts_generic_events" "sts_internal_events" "sts_multi_metrics" "sts_state_events" "sts_topo_process_agents" "sts_topo_trace_agents")

for topic in "${topics[@]}"; do
  # shellcheck disable=SC2046
  # shellcheck disable=SC2086
  kafka-topics.sh ${commonFlags} --partitions 1 --topic "${topic}"
done

# shellcheck disable=SC2046
# shellcheck disable=SC2086
kafka-topics.sh ${commonFlags} --partitions 10 --topic sts_correlate_endpoints
