#!/usr/bin/env bash
set -Eeuo pipefail
if [ -z "${INDICES_TO_RESTORE}" ] ; then
    INDICES_TO_RESTORE=${BACKUP_ELASTICSEARCH_SCHEDULED_INDICES}
fi

# Delete all indices if DELETE_ALL_INDICES is true
if [ "${DELETE_ALL_INDICES:-false}" = "true" ]; then
    echo "=== Deleting all indices matching pattern \"${INDICES_TO_RESTORE}\"..."

    # Get list of indices
    INDICES=$(curl -X GET "http://${ELASTICSEARCH_ENDPOINT}/_cat/indices/${INDICES_TO_RESTORE}?h=i" --silent)

    if [ -z "${INDICES}" ]; then
        echo "No indices found matching pattern \"${INDICES_TO_RESTORE}\""
    else
        echo "Found indices to delete:"
        echo "${INDICES}"

        # Check if sts_k8s_logs datastream indices exist and rollover if needed
        if echo "${INDICES}" | grep -q '^\.ds-sts_k8s_logs-'; then
            echo "sts_k8s_logs datastream indices found, rolling over..."
            SC=$(curl -X POST "http://${ELASTICSEARCH_ENDPOINT}/sts_k8s_logs/_rollover" --silent --output /dev/stderr --write-out "%{http_code}")
            if [ "$SC" -ne 200 ]; then
                echo "Failed to rollover datastream sts_k8s_logs (HTTP ${SC})"
                exit 1
            fi
            echo "Datastream sts_k8s_logs rolled over successfully"
        fi

        # Delete each index
        while IFS= read -r index; do
            if [ -n "${index}" ]; then
                echo "Deleting index: ${index}"
                SC=$(curl -X DELETE "http://${ELASTICSEARCH_ENDPOINT}/${index}" --silent --output /dev/stderr --write-out "%{http_code}")
                if [ "$SC" -ne 200 ]; then
                    echo "Failed to delete index ${index} (HTTP ${SC})"
                    exit 1
                fi

                # Verify index deletion (poll until index no longer exists)
                max_attempts=30
                attempt=0
                while [ $attempt -lt $max_attempts ]; do
                    SC=$(curl -X GET "http://${ELASTICSEARCH_ENDPOINT}/${index}" --silent --write-out "%{http_code}" --output /dev/null)
                    if [ "$SC" -eq 404 ]; then
                        echo "Index ${index} successfully deleted"
                        break
                    fi
                    attempt=$((attempt + 1))
                    if [ $attempt -ge $max_attempts ]; then
                        echo "Timeout waiting for index ${index} to be deleted"
                        exit 1
                    fi
                    sleep 1
                done
            fi
        done <<< "${INDICES}"

        echo "=== All indices deleted successfully"
    fi
fi

echo "=== Restoring ElasticSearch snapshot \"${SNAPSHOT_NAME}\" (indices = \"${INDICES_TO_RESTORE}\") from snapshot repository \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME}\" in bucket \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\"..."
SC=$(curl --request POST "http://${ELASTICSEARCH_ENDPOINT}/_snapshot/${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME}/${SNAPSHOT_NAME}/_restore?wait_for_completion=true&pretty" \
    --data "{\"indices\": \"${INDICES_TO_RESTORE}\"}" -H 'Content-Type: application/json' \
    --silent --output /dev/stderr --write-out "%{http_code}")
if [ "$SC" -ne 200 ]; then exit 1; fi
echo "==="
