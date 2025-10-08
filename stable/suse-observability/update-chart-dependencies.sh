#!/bin/bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "${dir}/../.." || exit


for SUBCHART in clickhouse  elasticsearch hbase kafka kafkaup-operator minio suse-observability victoria-metrics-single zookeeper
do
  rm -Rf "stable/${SUBCHART}/charts/*"
done


yq e '.dependencies[] | select (.repository == "file*").repository | sub("^file://","")' "stable/suse-observability/Chart.yaml"  | xargs -I % helm dependencies build --skip-refresh stable/suse-observability/%
helm dependencies update  stable/suse-observability
