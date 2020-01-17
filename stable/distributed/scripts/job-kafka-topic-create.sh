#!/bin/bash

[[ -n "${TRACE+x}" ]] && set -x

KAFKA_REPLICAS="${KAFKA_REPLICAS:-1}"

if (( KAFKA_REPLICAS >= 5 )); then
  defaultReplicationFactor="3"
elif (( KAFKA_REPLICAS >= 3 )); then
  defaultReplicationFactor="2"
else
  defaultReplicationFactor="1"
fi

commonFlags="--bootstrap-server ${KAFKA_BROKERS} --config retention.ms=86400000 --create --force --replication-factor ${defaultReplicationFactor}"
topics=("sts_connection_beat_events" "sts_generic_events" "sts_internal_events" "sts_multi_metrics" "sts_state_events" "sts_topo_process_agents" "sts_topo_trace_agents")

for topic in "${topics[@]}"; do
  printf -- "Creating topic '%s'...\n" "${topic}"
  # shellcheck disable=SC2046
  # shellcheck disable=SC2086
  kafka-topics.sh ${commonFlags} --partitions 1 --topic "${topic}" 2>/dev/null || printf -- "Topic '%s' already exists, skipping...\n" "${topic}"
done

printf -- "Creating topic '%s'...\n" "sts_correlate_endpoints"
# shellcheck disable=SC2046
# shellcheck disable=SC2086
kafka-topics.sh ${commonFlags} --partitions 10 --topic sts_correlate_endpoints 2>/dev/null || printf -- "Topic '%s' already exists, skipping...\n" "sts_correlate_endpoints"
