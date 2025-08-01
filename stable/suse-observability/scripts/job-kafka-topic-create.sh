#!/bin/bash

set -euxo pipefail

KAFKA_REPLICAS="${KAFKA_REPLICAS:-1}"
KAFKA_TOPIC_RETENTION="${KAFKA_TOPIC_RETENTION:-86400000}"

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

  PARTITION_ENV="KAFKA_PARTITIONS_$1"

  local topic=$1
  local partitions=${!PARTITION_ENV:-$2}
  local property=$3

  # support to pass properties as a single arg like "p1=v1,...,pN=vN"
  IFS=',' read -ra props_array <<< "$property"

  # shellcheck disable=SC2086
  if kafka-topics.sh ${commonFlags} --topic "${topic}" --describe 2>/dev/null; then
    local config_args=()
    for p in "${props_array[@]}"; do
      config_args+=(--add-config "$p")
    done

    printf -- "Topic '%s' already exists, updating settings...\n" "${topic}"
    kafka-configs.sh ${commonFlags} --alter --entity-type topics --entity-name "${topic}" "${config_args[@]}"
  else
    local config_args=()
    for p in "${props_array[@]}"; do
      config_args+=(--config "$p")
    done

    printf -- "Creating topic '%s'...\n" "${topic}"
    # shellcheck disable=SC2046
    # shellcheck disable=SC2086
    kafka-topics.sh ${commonFlags} ${commonCreateFlags} --partitions $partitions --topic "${topic}" "${config_args[@]}"
  fi
}

# For ephemeral data we can do time-based retention, to use less resources
ephemeralRetention="retention.ms=$KAFKA_TOPIC_RETENTION"
# For persistent topics (which are required for consistency) we disable retention
persistentRetention="retention.ms=-1"

PIDS=()
createOrUpdateTopic "sts_correlated_connections" "10" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_correlate_endpoints" "10" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_correlate_http_trace_observations" "10" "${ephemeralRetention}" &
PIDS+=($!)

createOrUpdateTopic "sts_health_sync" "10" "${persistentRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_intake_health" "10" "${persistentRetention}" &
PIDS+=($!)

createOrUpdateTopic "sts_topology_events" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_internal_events" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_topo_process_agents" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_internal_topology" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_health_sync_settings" "1" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_topology_stream" "10" "${ephemeralRetention}" &
PIDS+=($!)
createOrUpdateTopic "sts_otel_mapping_sync_settings" "1" "${persistentRetention},cleanup.policy=compact,min.cleanable.dirty.ratio=0.2" &
PIDS+=($!)

for pid in "${PIDS[@]}"; do
  wait "$pid"
done
