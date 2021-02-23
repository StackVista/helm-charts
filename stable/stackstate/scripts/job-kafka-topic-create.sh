#!/bin/bash

set -euxo pipefail

KAFKA_REPLICAS="${KAFKA_REPLICAS:-1}"

if (( KAFKA_REPLICAS >= 5 )); then
  defaultReplicationFactor="3"
elif (( KAFKA_REPLICAS >= 3 )); then
  defaultReplicationFactor="2"
else
  defaultReplicationFactor="1"
fi

commonFlags="--bootstrap-server ${KAFKA_BROKERS}"
commonCreateFlags="--config retention.ms=86400000 --create --force --replication-factor ${defaultReplicationFactor}"
extraPartitionTopics=("sts_correlate_endpoints" "sts_trace_events" "sts_alerts" "sts_external_alerts")
normalTopics=("sts_connection_beat_events" "sts_topology_events" "sts_generic_events" "sts_internal_events" "sts_multi_metrics" "sts_state_events" "sts_topo_process_agents" "sts_topo_trace_agents" "sts_internal_topology")

for topic in "${normalTopics[@]}"; do
  # shellcheck disable=SC2086
  if kafka-topics.sh ${commonFlags} --topic "${topic}" --describe 2>/dev/null; then
    printf -- "Topic '%s' already exists, skipping...\n" "${topic}"
  else
    printf -- "Creating topic '%s'...\n" "${topic}"
    # shellcheck disable=SC2046
    # shellcheck disable=SC2086
    kafka-topics.sh ${commonFlags} ${commonCreateFlags} --partitions 1 --topic "${topic}" 2>/dev/null
  fi
done

for topic in "${extraPartitionTopics[@]}"; do
  # shellcheck disable=SC2086
  if kafka-topics.sh ${commonFlags} --topic "${topic}" --describe 2>/dev/null; then
    printf -- "Topic '%s' already exists, skipping...\n" "${topic}"
  else
    printf -- "Creating topic '%s'...\n" "${topic}"
    # shellcheck disable=SC2046
    # shellcheck disable=SC2086
    kafka-topics.sh ${commonFlags} ${commonCreateFlags} --partitions 10 --topic "${topic}" 2>/dev/null
  fi
done
