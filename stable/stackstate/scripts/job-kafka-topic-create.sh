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
commonCreateFlags="--create --replication-factor ${defaultReplicationFactor}"

function createOrUpdateTopic() {
  set -euxo pipefail

  local topic=$1
  local partitions=$2
  local property=$3

  # shellcheck disable=SC2086
  if kafka-topics.sh ${commonFlags} --topic "${topic}" --describe 2>/dev/null; then
    printf -- "Topic '%s' already exists, updating settings...\n" "${topic}"
    kafka-configs.sh ${commonFlags} --alter --entity-type topics --entity-name "${topic}" --add-config ${property}
  else
    printf -- "Creating topic '%s'...\n" "${topic}"
    # shellcheck disable=SC2046
    # shellcheck disable=SC2086
    kafka-topics.sh ${commonFlags} ${commonCreateFlags} --partitions $partitions --topic "${topic}" --config ${property}
  fi
}

# For ephemeral data we can do time-based retention, to use less resources
ephemeralRetention="retention.ms=86400000"
# For persistent topics (which are required for consistency) we disable retention
persistentRetention="retention.ms=-1"

PIDS=()
createOrUpdateTopic "sts_correlated_connections" "10" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_correlate_endpoints" "10" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_trace_events" "10" "${ephemeralRetention}" &
PIDS+=($!)

createOrUpdateTopic "sts_health_sync" "10" "${persistentRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_intake_health" "10" "${persistentRetention}" &
PIDS+=($!)

createOrUpdateTopic "sts_connection_beat_events" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_topology_events" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_generic_events" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_internal_events" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_multi_metrics" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_metrics_v2" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_state_events" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_topo_process_agents" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_topo_trace_agents" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_internal_topology" "1" "${ephemeralRetention}" &
PIDS+=($!)

for pid in "${PIDS[@]}"; do
  wait "$pid"
done
