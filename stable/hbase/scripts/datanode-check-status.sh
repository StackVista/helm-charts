#!/usr/bin/env bash
# Exit on error. Append "|| true" if you expect an error.
set -o errexit
# Exit on error inside any functions or subshells.
set -o errtrace
# Do not allow use of undefined vars. Use ${VAR:-} to use an undefined VAR
set -o nounset
# Catch an error in command pipes.
set -o pipefail
# Turn on traces, useful while debugging.
set -o xtrace

# Needed as the image already defines a java agent that causes a port collision when running hdfs commands
# shellcheck disable=SC2034
HADOOP_OPTS=-XX:MaxRAMPercentage=75.0

function clusterIdCheck() {
  # Check if datanode registered with the namenode and got non-null cluster ID.
  _PORTS="50075 1006"
  _URL_PATH="jmx?qry=Hadoop:service=DataNode,name=DataNodeInfo"
  _CLUSTER_ID=""
  for _PORT in $_PORTS; do
  _CLUSTER_ID+=$(curl -s http://localhost:"${_PORT}"/"$_URL_PATH" |  \
      grep ClusterId) || true
  done
  echo "$_CLUSTER_ID" | grep -q -v null
}

function waitForReplication() {
  _REPLICATION_FACTOR="${HDFS_CONF_dfs_replication:-3}"
  _UNDER_REPLICATED=1

  while [ "$_UNDER_REPLICATED" -gt 0 ]
  do
     # Do we still have pending blocks?
     hdfs fsck / -maintenance | grep 'Under replicated' | awk -F':' '{print $1}' > /tmp/under_replicated_files
     _UNDER_REPLICATED=$(wc -l < /tmp/under_replicated_files)
     # Set replication factor on the under replicated files to trigger the missing replication
     xargs -n 500 hdfs dfs -setrep "$_REPLICATION_FACTOR" < /tmp/under_replicated_files
     sleep 15
  done
  echo "No pending replication"
}

case "$1" in
  waitForReplication)
    waitForReplication
    ;;
  clusterIdCheck)
    clusterIdCheck
    ;;
  *)
    echo $"Usage: $0 {waitForReplication|clusterIdCheck}"
esac
