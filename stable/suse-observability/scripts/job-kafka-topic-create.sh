#!/bin/bash

set -euxo pipefail

KAFKA_REPLICAS="${KAFKA_REPLICAS:-1}"
KAFKA_TOPIC_RETENTION="${KAFKA_TOPIC_RETENTION:-86400000}"
KAFKA_TRANSACTION_STATE_REPLICATION_FACTOR="${KAFKA_TRANSACTION_STATE_REPLICATION_FACTOR:-1}"
KAFKA_TRANSACTION_STATE_REASSIGNMENT_ATTEMPTS="${KAFKA_TRANSACTION_STATE_REASSIGNMENT_ATTEMPTS:-30}"
KAFKA_TRANSACTION_STATE_REASSIGNMENT_RETRY_INTERVAL="${KAFKA_TRANSACTION_STATE_REASSIGNMENT_RETRY_INTERVAL:-10}"
KAFKA_TRANSACTION_STATE_REASSIGNMENT_TIMEOUT="${KAFKA_TRANSACTION_STATE_REASSIGNMENT_TIMEOUT:-1800}"

if (( KAFKA_REPLICAS >= 5 )); then
  defaultReplicationFactor="3"
elif (( KAFKA_REPLICAS >= 3 )); then
  defaultReplicationFactor="2"
else
  defaultReplicationFactor="1"
fi

commonFlags="--bootstrap-server ${KAFKA_BROKERS}"
commonCreateFlags="--create --replication-factor ${defaultReplicationFactor}"

function ensureTopicReplicationFactor() {
  local topic=$1
  local desiredReplicationFactor=$2
  local deadline=$3
  local describeOutput
  local brokerOutput
  local reassignmentFile
  local verifyOutput
  local remainingSeconds
  local sleepSeconds
  local partition
  local replicas
  local brokerId
  local assignedBrokerId
  local replicaIndex
  local liveBrokerIndex
  local startIndex
  local offset
  local candidateIndex
  local reassignmentCount=0
  local firstPartition=true
  local assigned=false
  local live=false
  local reassignmentStarted=false
  local -a liveBrokerIds
  local -a assignedBrokerIds

  if (( desiredReplicationFactor <= 1 )); then
    return
  fi

  if (( SECONDS >= deadline )); then
    return 2
  fi

  # shellcheck disable=SC2086
  if ! describeOutput="$(kafka-topics.sh ${commonFlags} --topic "${topic}" --describe 2>&1)"; then
    if grep -Eiq 'does not exist|UnknownTopicOrPartitionException' <<< "${describeOutput}"; then
      printf -- "Topic '%s' does not exist yet; broker defaults will apply when it is created.\n" "${topic}"
      return
    fi

    printf -- "Failed to describe topic '%s'.\n%s\n" "${topic}" "${describeOutput}" >&2
    return 1
  fi

  if awk -v desired="${desiredReplicationFactor}" '
    /Partition:/ {
      foundPartition = 1
      for (i = 1; i <= NF; i++) {
        if ($i == "Replicas:") {
          foundReplicas = 1
          count = split($(i + 1), replicaIds, ",")
          if (count < desired) {
            underReplicated = 1
          }
        }
      }
    }
    END {
      exit !(foundPartition && foundReplicas && !underReplicated)
    }
  ' <<< "${describeOutput}"; then
    printf -- "Topic '%s' already has replication factor %s or higher.\n" "${topic}" "${desiredReplicationFactor}"
    return
  fi

  # shellcheck disable=SC2086
  if ! brokerOutput="$(kafka-broker-api-versions.sh ${commonFlags} 2>&1)"; then
    printf -- "Failed to discover live Kafka brokers.\n%s\n" "${brokerOutput}" >&2
    return 1
  fi
  mapfile -t liveBrokerIds < <(
    sed -nE 's/.*\(id: ([-0-9]+).*/\1/p' <<< "${brokerOutput}" | sort -n -u
  )

  if (( ${#liveBrokerIds[@]} < desiredReplicationFactor )); then
    printf -- "Topic '%s' needs %s replicas, but only %s live brokers were found.\n" \
      "${topic}" "${desiredReplicationFactor}" "${#liveBrokerIds[@]}" >&2
    return 1
  fi

  reassignmentFile="$(mktemp)"
  printf '{"version":1,"partitions":[' > "${reassignmentFile}"

  # Build assignments directly to preserve existing replicas and preferred leaders.
  while IFS=$'\t' read -r partition replicas; do
    IFS=',' read -ra assignedBrokerIds <<< "${replicas}"
    if (( ${#assignedBrokerIds[@]} >= desiredReplicationFactor )); then
      continue
    fi

    startIndex=-1
    for assignedBrokerId in "${assignedBrokerIds[@]}"; do
      live=false
      for liveBrokerIndex in "${!liveBrokerIds[@]}"; do
        if [[ "${liveBrokerIds[liveBrokerIndex]}" == "${assignedBrokerId}" ]]; then
          live=true
          if (( startIndex == -1 )); then
            startIndex=${liveBrokerIndex}
          fi
          break
        fi
      done
      if [[ "${live}" == false ]]; then
        printf -- "Replica broker %s for %s-%s is not live; postponing reassignment.\n" \
          "${assignedBrokerId}" "${topic}" "${partition}" >&2
        rm -f "${reassignmentFile}"
        return 1
      fi
    done

    for ((offset = 1; offset <= ${#liveBrokerIds[@]}; offset += 1)); do
      candidateIndex=$(((startIndex + offset) % ${#liveBrokerIds[@]}))
      brokerId=${liveBrokerIds[candidateIndex]}
      assigned=false
      for assignedBrokerId in "${assignedBrokerIds[@]}"; do
        if [[ "${assignedBrokerId}" == "${brokerId}" ]]; then
          assigned=true
          break
        fi
      done
      if [[ "${assigned}" == false ]]; then
        assignedBrokerIds+=("${brokerId}")
      fi
      if (( ${#assignedBrokerIds[@]} == desiredReplicationFactor )); then
        break
      fi
    done

    if (( ${#assignedBrokerIds[@]} < desiredReplicationFactor )); then
      printf -- "Could not assign %s replicas to %s-%s.\n" \
        "${desiredReplicationFactor}" "${topic}" "${partition}" >&2
      rm -f "${reassignmentFile}"
      return 1
    fi

    if [[ "${firstPartition}" == false ]]; then
      printf ',' >> "${reassignmentFile}"
    fi
    firstPartition=false
    ((reassignmentCount += 1))
    printf '{"topic":"%s","partition":%s,"replicas":[' "${topic}" "${partition}" >> "${reassignmentFile}"
    for replicaIndex in "${!assignedBrokerIds[@]}"; do
      if (( replicaIndex > 0 )); then
        printf ',' >> "${reassignmentFile}"
      fi
      printf '%s' "${assignedBrokerIds[replicaIndex]}" >> "${reassignmentFile}"
    done
    printf ']}' >> "${reassignmentFile}"
  done < <(
    awk '
      /Partition:/ {
        partition = ""
        replicas = ""
        for (i = 1; i <= NF; i++) {
          if ($i == "Partition:") {
            partition = $(i + 1)
          } else if ($i == "Replicas:") {
            replicas = $(i + 1)
          }
        }
        if (partition != "" && replicas != "") {
          printf "%s\t%s\n", partition, replicas
        }
      }
    ' <<< "${describeOutput}"
  )
  printf ']}' >> "${reassignmentFile}"

  if (( reassignmentCount == 0 )); then
    printf -- "No partitions requiring reassignment were found for topic '%s'.\n" "${topic}" >&2
    rm -f "${reassignmentFile}"
    return 1
  fi

  printf -- "Increasing topic '%s' to replication factor %s.\n" "${topic}" "${desiredReplicationFactor}"
  while true; do
    # shellcheck disable=SC2086
    if ! verifyOutput="$(kafka-reassign-partitions.sh ${commonFlags} \
      --reassignment-json-file "${reassignmentFile}" \
      --verify \
      --preserve-throttles 2>&1)"; then
      printf -- "Failed to verify topic '%s' reassignment.\n%s\n" "${topic}" "${verifyOutput}" >&2
      rm -f "${reassignmentFile}"
      return 1
    fi
    printf '%s\n' "${verifyOutput}"

    if grep -q 'still in progress' <<< "${verifyOutput}"; then
      reassignmentStarted=true
    elif grep -q 'rather than' <<< "${verifyOutput}"; then
      if [[ "${reassignmentStarted}" == true ]]; then
        printf -- "Topic '%s' reassignment finished with an unexpected replica set.\n" "${topic}" >&2
        rm -f "${reassignmentFile}"
        return 1
      fi
      # shellcheck disable=SC2086
      if ! kafka-reassign-partitions.sh ${commonFlags} \
        --reassignment-json-file "${reassignmentFile}" \
        --execute; then
        printf -- "Could not start topic '%s' reassignment.\n" "${topic}" >&2
        rm -f "${reassignmentFile}"
        return 1
      fi
      reassignmentStarted=true
      continue
    else
      break
    fi

    if (( SECONDS >= deadline )); then
      printf -- "Timed out waiting for topic '%s' reassignment.\n" "${topic}" >&2
      rm -f "${reassignmentFile}"
      return 2
    fi
    remainingSeconds=$((deadline - SECONDS))
    sleepSeconds=5
    if (( remainingSeconds < sleepSeconds )); then
      sleepSeconds=${remainingSeconds}
    fi
    sleep "${sleepSeconds}"
  done

  rm -f "${reassignmentFile}"
}

function createOrUpdateTopic() {
  set -euxo pipefail

  PARTITION_ENV="KAFKA_PARTITIONS_$1"

  local topic=$1
  local partitions=${!PARTITION_ENV:-$2}
  local property=$3

  # support to pass properties as a single arg like "p1=v1|...|pN=vN"
  IFS='|' read -ra props_array <<< "$property"

  # shellcheck disable=SC2086
  if kafka-topics.sh ${commonFlags} --topic "${topic}" --describe 2>/dev/null; then
    # For updates: single --add-config with comma-separated properties, bracket-wrap values with commas
    local config_parts=()
    for p in "${props_array[@]}"; do
      # Check if the property value (after =) contains commas
      if [[ "$p" == *"="*","* ]]; then
        # Split on first = to get key and value
        local key="${p%%=*}"
        local value="${p#*=}"
        # Wrap value in brackets if it contains commas
        config_parts+=("${key}=[${value}]")
      else
        config_parts+=("$p")
      fi
    done

    # Join all config parts with commas
    local config_string=""
    for i in "${!config_parts[@]}"; do
      if [ $i -eq 0 ]; then
        config_string="${config_parts[i]}"
      else
        config_string="${config_string},${config_parts[i]}"
      fi
    done

    printf -- "Topic '%s' already exists, updating settings...\n" "${topic}"
    kafka-configs.sh ${commonFlags} --alter --entity-type topics --entity-name "${topic}" --add-config "${config_string}"
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
# Compact-only: metadata must survive restarts indefinitely (one record per mapping).
createOrUpdateTopic "sts_topology_stream_metadata" "1" "cleanup.policy=compact|min.cleanable.dirty.ratio=0.1|max.compaction.lag.ms=86400000" &
PIDS+=($!)

# Topic configuration for aggressive compaction and retention
# - cleanup.policy=compact,delete: Enable both log compaction (deduplication) and time-based deletion
# - min.cleanable.dirty.ratio=0.1: Trigger compaction when 10% of log is "dirty" (more aggressive than default 50%)
# - segment.ms=86400000: Roll log segments every 24 hours (86400000ms) to enable deletion of old segments
# - max.compaction.lag.ms=86400000: Force compaction within 24 hours regardless of dirty ratio
createOrUpdateTopic "sts_internal_settings" "1" "${ephemeralRetention}|cleanup.policy=compact,delete|min.cleanable.dirty.ratio=0.1|segment.ms=86400000|max.compaction.lag.ms=86400000"
PIDS+=($!)

for pid in "${PIDS[@]}"; do
  wait "$pid"
done

migrationAttempt=1
migrationDeadline=$((SECONDS + KAFKA_TRANSACTION_STATE_REASSIGNMENT_TIMEOUT))
while true; do
  if ensureTopicReplicationFactor \
    "__transaction_state" \
    "${KAFKA_TRANSACTION_STATE_REPLICATION_FACTOR}" \
    "${migrationDeadline}"; then
    break
  else
    migrationStatus=$?
  fi

  if (( migrationStatus == 2 \
    || SECONDS >= migrationDeadline \
    || migrationAttempt >= KAFKA_TRANSACTION_STATE_REASSIGNMENT_ATTEMPTS )); then
    printf -- "WARNING: __transaction_state replication was not updated; the next chart run will retry it.\n" >&2
    break
  fi

  migrationRetryDelay=$((migrationDeadline - SECONDS))
  if (( migrationRetryDelay <= 0 )); then
    printf -- "WARNING: __transaction_state replication was not updated; the next chart run will retry it.\n" >&2
    break
  fi
  if (( migrationRetryDelay > KAFKA_TRANSACTION_STATE_REASSIGNMENT_RETRY_INTERVAL )); then
    migrationRetryDelay=${KAFKA_TRANSACTION_STATE_REASSIGNMENT_RETRY_INTERVAL}
  fi

  printf -- "Retrying __transaction_state replication migration in %s seconds (%s/%s).\n" \
    "${migrationRetryDelay}" \
    "${migrationAttempt}" \
    "${KAFKA_TRANSACTION_STATE_REASSIGNMENT_ATTEMPTS}" >&2
  sleep "${migrationRetryDelay}"
  ((migrationAttempt += 1))
done
