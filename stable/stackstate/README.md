# stackstate

Helm chart for StackState

Current chart version is `5.1.4-snapshot.3`

**Homepage:** <https://gitlab.com/stackvista/stackstate.git>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../stackstate-agent/ | stackstate-agent | * |
| https://helm.stackstate.io | anomaly-detection | 5.1.4 |
| https://helm.stackstate.io | common | 0.4.23 |
| https://helm.stackstate.io | elasticsearch | 7.17.2-stackstate.6 |
| https://helm.stackstate.io | hbase | 0.1.152 |
| https://helm.stackstate.io | kafkaup-operator | 0.1.6 |
| https://helm.stackstate.io | minio | 8.0.10-stackstate.8 |
| https://helm.stackstate.io | pull-secret | 1.0.0 |
| https://raw.githubusercontent.com/bitnami/charts/eb5f9a9513d987b519f0ecd732e7031241c50328/bitnami | kafka | 15.5.1 |
| https://raw.githubusercontent.com/bitnami/charts/eb5f9a9513d987b519f0ecd732e7031241c50328/bitnami | zookeeper | 8.1.2 |

## Required Values

In order to successfully install this chart, you **must** provide the following variables:
* `stackstate.license.key`
* `stackstate.baseUrl`

Install them on the command line on Helm with the following command:

```shell
helm install \
--set stackstate.license.key=<your-license-key> \
--set stackstate.baseUrl=<your-base-url> \
stackstate/stackstate
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| anomaly-detection.cpu.limit | string | `"2000m"` | CPU resource limit |
| anomaly-detection.cpu.request | string | `"1000m"` | CPU resource request |
| anomaly-detection.enabled | bool | `true` | Enables anomaly detection chart |
| anomaly-detection.image.imagePullPolicy | string | `"IfNotPresent"` | The default pullPolicy used for anomaly detection pods. |
| anomaly-detection.image.pullSecretName | string | `nil` | Name of ImagePullSecret to use for all pods. |
| anomaly-detection.image.pullSecretPassword | string | `nil` |  |
| anomaly-detection.image.pullSecretUsername | string | `nil` | Password used to login to the registry to pull Docker images of all pods. |
| anomaly-detection.image.registry | string | `"quay.io"` | Base container image registry for all containers, except for the wait container |
| anomaly-detection.image.spotlightRepository | string | `"stackstate/spotlight"` | Repository of the spotlight Docker image. |
| anomaly-detection.ingress | object | `{"annotations":{},"enabled":false,"hostname":null,"hosts":[],"port":8090,"tls":null}` | Status interface ingress |
| anomaly-detection.ingress.enabled | bool | `false` | Enables ingress controller for status interface |
| anomaly-detection.ingress.hostname | string | `nil` | Status interface hostname e.g. spotlight.local.domain |
| anomaly-detection.memory.limit | string | `"3Gi"` | Memory resource limit |
| anomaly-detection.memory.request | string | `"3Gi"` | Memory resource request |
| anomaly-detection.pdb.maxUnavailable | int | `0` | Maximum number of pods that can be unavailable for the anomaly detection |
| anomaly-detection.stackstate.apiToken | string | `nil` | Stackstate Api token that used by spotlight for authentication, it is expected to be set only in case if authType = "api-token" |
| anomaly-detection.stackstate.authType | string | `"token"` | Type of authentication. There are three options 1) "token" - with service account token (default), 2) "api-token" - with Stackstate API Token, 3) "cookie" - username, password based auth. |
| anomaly-detection.stackstate.instance | string | `"http://{{ include \"stackstate.hostname.prefix\" . }}-router:8080"` | **Required Stackstate instance URL, e.g http://stackstate-router:8080 |
| anomaly-detection.stackstate.password | string | `nil` | Stackstate Password used by spotlight for authentication, it is expected to be set only in case if authType = "cookie" |
| anomaly-detection.stackstate.username | string | `nil` | Stackstate Username used by spotlight for authentication, it is expected to be set only in case if authType = "cookie" |
| anomaly-detection.threadWorkers | int | `3` | The number of worker threads. |
| backup.additionalLogging | string | `""` | Additional logback config for backup components |
| backup.configuration.bucketName | string | `"sts-configuration-backup"` | Name of the MinIO bucket to store configuration backups. |
| backup.configuration.restore.enabled | bool | `true` | Enable configuration backup restore functionality (if `backup.enabled` is set to `true`). |
| backup.configuration.scheduled.backupDatetimeParseFormat | string | `"%Y%m%d-%H%M"` | Format to parse date/time from configuration backup name. *Note:* This should match the value for `backupNameTemplate`. |
| backup.configuration.scheduled.backupNameParseRegexp | string | `"sts-backup-([0-9]*-[0-9]*).stj"` | Regular expression to retrieve date/time from configuration backup name. *Note:* This should match the value for `backupNameTemplate`. |
| backup.configuration.scheduled.backupNameTemplate | string | `"sts-backup-$(date +%Y%m%d-%H%M).stj"` | Template for the configuration backup name as a double-quoted shell string value. |
| backup.configuration.scheduled.backupRetentionTimeDelta | string | `"days = 365"` | Time to keep configuration backups in [Python timedelta format](https://docs.python.org/3/library/datetime.html#timedelta-objects). |
| backup.configuration.scheduled.enabled | bool | `true` | Enable scheduled configuration backups (if `backup.enabled` is set to `true`). |
| backup.configuration.scheduled.schedule | string | `"0 4 * * *"` | Cron schedule for automatic configuration backups in [Kubernetes cron schedule syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#cron-schedule-syntax). |
| backup.configuration.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| backup.configuration.securityContext.fsGroup | int | `65534` | The GID (group ID) of all files on all mounted volumes |
| backup.configuration.securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| backup.configuration.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| backup.configuration.securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| backup.elasticsearch.bucketName | string | `"sts-elasticsearch-backup"` | Name of the MinIO bucket where ElasticSearch snapshots are stored. |
| backup.elasticsearch.restore.enabled | bool | `true` | Enable ElasticSearch snapshot restore functionality (if `backup.enabled` is set to `true`). |
| backup.elasticsearch.scheduled.enabled | bool | `true` | Enable scheduled ElasticSearch snapshots (if `backup.enabled` is set to `true`). |
| backup.elasticsearch.scheduled.indices | string | `"sts*"` | ElasticSearch indices to snapshot in [JSON array format](https://www.w3schools.com/js/js_json_arrays.asp). |
| backup.elasticsearch.scheduled.schedule | string | `"0 0 3 * * ?"` | Cron schedule for automatic ElasticSearch snaphosts in [ElastichSearch cron schedule syntax](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/cron-expressions.html). |
| backup.elasticsearch.scheduled.snapshotNameTemplate | string | `"<sts-backup-{now{yyyyMMdd-HHmm}}>"` | Template for the ElasticSearch snapshot name in [ElasticSearch date math format](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/date-math-index-names.html). |
| backup.elasticsearch.scheduled.snapshotPolicyName | string | `"auto-sts-backup"` | Name of the ElasticSearch snapshot policy. |
| backup.elasticsearch.scheduled.snapshotRetentionExpireAfter | string | `"30d"` | Amount of time to keep ElasticSearch snapshots in [ElasticSearch time units](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/common-options.html#time-units). *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). |
| backup.elasticsearch.scheduled.snapshotRetentionMaxCount | string | `"30"` | Minimum number of ElasticSearch snapshots to keep. *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). |
| backup.elasticsearch.scheduled.snapshotRetentionMinCount | string | `"5"` | Minimum number of ElasticSearch snapshots to keep. *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). |
| backup.elasticsearch.snapshotRepositoryName | string | `"sts-backup"` | Name of the ElasticSearch snapshot repository. |
| backup.enabled | bool | `false` | Enables backup/restore, including the MinIO subsystem. |
| backup.poddisruptionbudget.maxUnavailable | int | `0` | Maximum number of pods that can be unavailable during the backup. |
| backup.stackGraph.bucketName | string | `"sts-stackgraph-backup"` | Name of the MinIO bucket to store StackGraph backups. |
| backup.stackGraph.restore.enabled | bool | `true` | Enable StackGraph backup restore functionality (if `backup.enabled` is set to `true`). |
| backup.stackGraph.restore.tempData.accessModes[0] | string | `"ReadWriteOnce"` |  |
| backup.stackGraph.restore.tempData.size | string | `"{{ .Values.hbase.hdfs.datanode.persistence.size }}"` |  |
| backup.stackGraph.restore.tempData.storageClass | string | `nil` |  |
| backup.stackGraph.scheduled.backupDatetimeParseFormat | string | `"%Y%m%d-%H%M"` | Format to parse date/time from StackGraph backup name. *Note:* This should match the value for `backupNameTemplate`. |
| backup.stackGraph.scheduled.backupNameParseRegexp | string | `"sts-backup-([0-9]*-[0-9]*).graph"` | Regular expression to retrieve date/time from StackGraph backup name. *Note:* This should match the value for `backupNameTemplate`. |
| backup.stackGraph.scheduled.backupNameTemplate | string | `"sts-backup-$(date +%Y%m%d-%H%M).graph"` | Template for the StackGraph backup name as a double-quoted shell string value. |
| backup.stackGraph.scheduled.backupRetentionTimeDelta | string | `"days = 30"` | Time to keep StackGraph backups in [Python timedelta format](https://docs.python.org/3/library/datetime.html#timedelta-objects). |
| backup.stackGraph.scheduled.enabled | bool | `true` | Enable scheduled StackGraph backups (if `backup.enabled` is set to `true`). |
| backup.stackGraph.scheduled.schedule | string | `"0 3 * * *"` | Cron schedule for automatic StackGraph backups in [Kubernetes cron schedule syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#cron-schedule-syntax). |
| backup.stackGraph.scheduled.tempData.accessModes[0] | string | `"ReadWriteOnce"` |  |
| backup.stackGraph.scheduled.tempData.size | string | `"{{ .Values.hbase.hdfs.datanode.persistence.size }}"` |  |
| backup.stackGraph.scheduled.tempData.storageClass | string | `nil` |  |
| backup.stackGraph.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| backup.stackGraph.securityContext.fsGroup | int | `65534` | The GID (group ID) of all files on all mounted volumes |
| backup.stackGraph.securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| backup.stackGraph.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| backup.stackGraph.securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| caspr.enabled | bool | `false` | Enable CASPR compatible values. |
| cluster-role.enabled | bool | `true` | Deploy the ClusterRole(s) and ClusterRoleBinding(s) together with the chart. Can be disabled if these need to be installed by an administrator of the Kubernetes cluster. |
| commonLabels | object | `{}` | Labels that will be added to all resources created by the stackstate chart (not the subcharts though) |
| elasticsearch.clusterHealthCheckParams | string | `"wait_for_status=yellow&timeout=1s"` | The Elasticsearch cluster health status params that will be used by readinessProbe command |
| elasticsearch.clusterName | string | `"stackstate-elasticsearch"` | Name override for Elasticsearch child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| elasticsearch.commonLabels | object | `{"app.kubernetes.io/part-of":"stackstate"}` | Add additional labels to all resources created for elasticsearch |
| elasticsearch.enabled | bool | `true` | Enable / disable chart-based Elasticsearch. |
| elasticsearch.esJavaOpts | string | `"-Xmx3g -Xms3g -Des.allow_insecure_settings=true"` | JVM options |
| elasticsearch.extraEnvs | list | `[{"name":"action.auto_create_index","value":"true"},{"name":"indices.query.bool.max_clause_count","value":"10000"}]` | Extra settings that StackState uses for Elasticsearch. |
| elasticsearch.minimumMasterNodes | int | `2` | Minimum number of Elasticsearch master nodes. |
| elasticsearch.nodeGroup | string | `"master"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.enabled | bool | `true` |  |
| elasticsearch.prometheus-elasticsearch-exporter.es.uri | string | `"http://stackstate-elasticsearch-master:9200"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.podAnnotations."ad.stackstate.com/exporter.check_names" | string | `"[\"openmetrics\"]"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.podAnnotations."ad.stackstate.com/exporter.init_configs" | string | `"[{}]"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.podAnnotations."ad.stackstate.com/exporter.instances" | string | `"[ { \"prometheus_url\": \"http://%%host%%:9108/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"elasticsearch_indices_store_*\", \"elasticsearch_cluster_health_*\"] } ]"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.limits.cpu | string | `"100m"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.limits.memory | string | `"100Mi"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.requests.cpu | string | `"100m"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.requests.memory | string | `"100Mi"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.serviceMonitor.enabled | bool | `false` |  |
| elasticsearch.prometheus-elasticsearch-exporter.serviceMonitor.labels | object | `{}` | Labels for the service monitor |
| elasticsearch.replicas | int | `3` | Number of Elasticsearch replicas. |
| elasticsearch.resources | object | `{"limits":{"cpu":"2000m","ephemeral-storage":"1Gi","memory":"4Gi"},"requests":{"cpu":"2000m","ephemeral-storage":"1Mi","memory":"4Gi"}}` | Override Elasticsearch resources |
| elasticsearch.volumeClaimTemplate | object | `{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"250Gi"}}}` | PVC template defaulting to 250Gi default volumes |
| global.imagePullSecrets | list | `[]` | List of image pull secret names to be used by all images across all charts. |
| global.receiverApiKey | string | `""` | API key to be used by the Receiver; if no key is provided, a random one will be generated for you. |
| hbase.all.metrics.agentAnnotationsEnabled | bool | `true` |  |
| hbase.all.metrics.enabled | bool | `true` |  |
| hbase.commonLabels | object | `{"app.kubernetes.io/part-of":"stackstate"}` | Add additional labels to all resources created for all hbase resources |
| hbase.console.enabled | bool | `false` | Enable / disable deployment of the stackgraph-console for debugging. |
| hbase.console.integrity.enabled | bool | `false` | Enable / disable periodic integrity check to run though a cronjob. |
| hbase.console.integrity.schedule | string | `"*/30 * * * *"` | Schedule at which the integrity check runs |
| hbase.enabled | bool | `true` | Enable / disable chart-based HBase. |
| hbase.hbase.master.replicaCount | int | `2` | Number of HBase master node replicas. |
| hbase.hbase.master.resources.limits.cpu | string | `"500m"` |  |
| hbase.hbase.master.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| hbase.hbase.master.resources.limits.memory | string | `"1Gi"` |  |
| hbase.hbase.master.resources.requests.cpu | string | `"50m"` |  |
| hbase.hbase.master.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| hbase.hbase.master.resources.requests.memory | string | `"1Gi"` |  |
| hbase.hbase.regionserver.replicaCount | int | `3` | Number of HBase regionserver node replicas. |
| hbase.hbase.regionserver.resources.limits.cpu | string | `"3000m"` |  |
| hbase.hbase.regionserver.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| hbase.hbase.regionserver.resources.limits.memory | string | `"3Gi"` |  |
| hbase.hbase.regionserver.resources.requests.cpu | string | `"1500m"` |  |
| hbase.hbase.regionserver.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| hbase.hbase.regionserver.resources.requests.memory | string | `"3Gi"` |  |
| hbase.hdfs.datanode.replicaCount | int | `3` | Number of HDFS datanode replicas. |
| hbase.hdfs.datanode.resources.limits.cpu | string | `"500m"` |  |
| hbase.hdfs.datanode.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| hbase.hdfs.datanode.resources.limits.memory | string | `"4Gi"` |  |
| hbase.hdfs.datanode.resources.requests.cpu | string | `"300m"` |  |
| hbase.hdfs.datanode.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| hbase.hdfs.datanode.resources.requests.memory | string | `"4Gi"` |  |
| hbase.hdfs.minReplication | int | `2` | Min number of copies we create from any data block. (If the hbase.hdfs.datanode.replicaCount is set to a lower value than this, we will use the replicaCount instead) |
| hbase.hdfs.namenode.resources.limits.cpu | string | `"500m"` |  |
| hbase.hdfs.namenode.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| hbase.hdfs.namenode.resources.limits.memory | string | `"1Gi"` |  |
| hbase.hdfs.namenode.resources.requests.cpu | string | `"100m"` |  |
| hbase.hdfs.namenode.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| hbase.hdfs.namenode.resources.requests.memory | string | `"1Gi"` |  |
| hbase.hdfs.secondarynamenode.enabled | bool | `true` |  |
| hbase.hdfs.secondarynamenode.resources.limits.cpu | string | `"500m"` |  |
| hbase.hdfs.secondarynamenode.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| hbase.hdfs.secondarynamenode.resources.limits.memory | string | `"1Gi"` |  |
| hbase.hdfs.secondarynamenode.resources.requests.cpu | string | `"50m"` |  |
| hbase.hdfs.secondarynamenode.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| hbase.hdfs.secondarynamenode.resources.requests.memory | string | `"1Gi"` |  |
| hbase.stackgraph.image.tag | string | `"4.9.2"` | The StackGraph server version, must be compatible with the StackState version |
| hbase.tephra.replicaCount | int | `2` | Number of Tephra replicas. |
| hbase.tephra.resources.limits.cpu | string | `"500m"` |  |
| hbase.tephra.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| hbase.tephra.resources.limits.memory | string | `"3Gi"` |  |
| hbase.tephra.resources.requests.cpu | string | `"250m"` |  |
| hbase.tephra.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| hbase.tephra.resources.requests.memory | string | `"2Gi"` |  |
| hbase.zookeeper.enabled | bool | `false` | Disable Zookeeper from the HBase chart **Don't change unless otherwise specified**. |
| hbase.zookeeper.externalServers | string | `"stackstate-zookeeper-headless"` | External Zookeeper if not used bundled Zookeeper chart **Don't change unless otherwise specified**. |
| ingress.annotations | object | `{}` | Annotations for ingress objects. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| ingress.hosts | list | `[]` | List of ingress hostnames; the paths are fixed to StackState backend services |
| ingress.path | string | `"/"` |  |
| ingress.tls | list | `[]` | List of ingress TLS certificates to use. |
| kafka.command | list | `["/scripts/custom-setup.sh"]` | Override kafka container command. |
| kafka.commonLabels | object | `{"app.kubernetes.io/part-of":"stackstate"}` | Add additional labels to all resources created for kafka |
| kafka.defaultReplicationFactor | int | `2` |  |
| kafka.enabled | bool | `true` | Enable / disable chart-based Kafka. |
| kafka.externalZookeeper.servers | string | `"stackstate-zookeeper-headless"` | External Zookeeper if not used bundled Zookeeper chart **Don't change unless otherwise specified**. |
| kafka.extraDeploy | list | `[{"apiVersion":"v1","data":{"custom-setup.sh":"#!/bin/bash\n\nID=\"${MY_POD_NAME#\"{{ include \"common.names.fullname\" . }}-\"}\"\n\nKAFKA_META_PROPERTIES=/bitnami/kafka/data/meta.properties\nif [[ -f ${KAFKA_META_PROPERTIES} ]]; then\n  ID=`grep -e ^broker.id= ${KAFKA_META_PROPERTIES} | sed 's/^broker.id=//'`\n  if [[ \"${ID}\" != \"\" ]] && [[ \"${ID}\" -gt 1000 ]]; then\n    echo \"Using broker ID ${ID} from ${KAFKA_META_PROPERTIES} for compatibility (STAC-9614)\"\n  fi\nfi\n\nexport KAFKA_CFG_BROKER_ID=\"$ID\"\n\nexec /entrypoint.sh /run.sh"},"kind":"ConfigMap","metadata":{"name":"kafka-custom-scripts"}}]` | Array of extra objects to deploy with the release |
| kafka.extraEnvVars | list | `[{"name":"KAFKA_CFG_RESERVED_BROKER_MAX_ID","value":"2000"},{"name":"KAFKA_CFG_TRANSACTIONAL_ID_EXPIRATION_MS","value":"2147483647"}]` | Extra environment variables to add to kafka pods. |
| kafka.extraVolumeMounts | list | `[{"mountPath":"/scripts/custom-setup.sh","name":"kafka-custom-scripts","subPath":"custom-setup.sh"}]` | Extra volumeMount(s) to add to Kafka containers. |
| kafka.extraVolumes | list | `[{"configMap":{"defaultMode":493,"name":"kafka-custom-scripts"},"name":"kafka-custom-scripts"}]` | Extra volume(s) to add to Kafka statefulset. |
| kafka.fullnameOverride | string | `"stackstate-kafka"` | Name override for Kafka child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| kafka.image.registry | string | `"quay.io"` | Kafka image registry |
| kafka.image.repository | string | `"stackstate/kafka"` | Kafka image repository |
| kafka.image.tag | string | `"2.8.1-2738720666"` | Kafka image tag. **Since StackState relies on this specific version, it's advised NOT to change this.** When changing this version, be sure to change the pod annotation stackstate.com/kafkaup-operator.kafka_version aswell, in order for the kafkaup operator to upgrade the inter broker protocol version |
| kafka.initContainers | list | `[{"args":["-c","while [ -z \"${KAFKA_CFG_INTER_BROKER_PROTOCOL_VERSION}\" ]; do echo \"KAFKA_CFG_INTER_BROKER_PROTOCOL_VERSION should be set by operator\"; sleep 1; done"],"command":["/bin/bash"],"image":"{{ include \"kafka.image\" . }}","imagePullPolicy":"","name":"check-inter-broker-protocol-version","resources":{"limits":{},"requests":{}}}]` | required to make the kafka versionup operator work |
| kafka.livenessProbe.initialDelaySeconds | int | `240` | Delay before readiness probe is initiated. |
| kafka.logRetentionHours | int | `24` | The minimum age of a log file to be eligible for deletion due to age. |
| kafka.metrics.jmx.enabled | bool | `true` | Whether or not to expose JMX metrics to Prometheus. |
| kafka.metrics.jmx.image.registry | string | `"quay.io"` | Kafka JMX exporter image registry |
| kafka.metrics.jmx.image.repository | string | `"stackstate/jmx-exporter"` | Kafka JMX exporter image repository |
| kafka.metrics.jmx.image.tag | string | `"0.17.0-2738680727"` | Kafka JMX exporter image tag |
| kafka.metrics.jmx.resources.limits.cpu | string | `"1"` |  |
| kafka.metrics.jmx.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| kafka.metrics.jmx.resources.limits.memory | string | `"300Mi"` |  |
| kafka.metrics.jmx.resources.requests.cpu | string | `"200m"` |  |
| kafka.metrics.jmx.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| kafka.metrics.jmx.resources.requests.memory | string | `"300Mi"` |  |
| kafka.metrics.kafka.enabled | bool | `false` | Whether or not to create a standalone Kafka exporter to expose Kafka metrics. |
| kafka.metrics.serviceMonitor.enabled | bool | `false` | If `true`, creates a Prometheus Operator `ServiceMonitor` (also requires `kafka.metrics.kafka.enabled` or `kafka.metrics.jmx.enabled` to be `true`). |
| kafka.metrics.serviceMonitor.interval | string | `"20s"` | How frequently to scrape metrics. |
| kafka.metrics.serviceMonitor.labels | object | `{}` | Add extra labels to target a specific prometheus instance |
| kafka.offsetsTopicReplicationFactor | int | `2` |  |
| kafka.pdb.create | bool | `true` |  |
| kafka.pdb.maxUnavailable | int | `1` |  |
| kafka.pdb.minAvailable | string | `""` |  |
| kafka.persistence.size | string | `"50Gi"` | Size of persistent volume for each Kafka pod |
| kafka.podAnnotations | object | `{"ad.stackstate.com/jmx-exporter.check_names":"[\"openmetrics\"]","ad.stackstate.com/jmx-exporter.init_configs":"[{}]","ad.stackstate.com/jmx-exporter.instances":"[ { \"prometheus_url\": \"http://%%host%%:5556/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"*\"] } ]","stackstate.com/kafkaup-operator.kafka_version":"2.8.1"}` | Kafka Pod annotations. |
| kafka.podLabels."app.kubernetes.io/part-of" | string | `"stackstate"` |  |
| kafka.readinessProbe.initialDelaySeconds | int | `45` | Delay before readiness probe is initiated. |
| kafka.replicaCount | int | `3` | Number of Kafka replicas. |
| kafka.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2Gi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Mi","memory":"2Gi"}}` | Kafka resources per pods. |
| kafka.topic.stsMetricsV2.partitionCount | int | `10` |  |
| kafka.transactionStateLogReplicationFactor | int | `2` |  |
| kafka.volumePermissions.enabled | bool | `false` |  |
| kafka.zookeeper.enabled | bool | `false` | Disable Zookeeper from the Kafka chart **Don't change unless otherwise specified**. |
| kafkaup-operator.enabled | bool | `true` |  |
| kafkaup-operator.image.pullPolicy | string | `""` |  |
| kafkaup-operator.image.registry | string | `"quay.io"` |  |
| kafkaup-operator.image.repository | string | `"stackstate/kafkaup-operator"` |  |
| kafkaup-operator.image.tag | string | `"0.0.2"` |  |
| kafkaup-operator.kafkaSelectors.podLabel.key | string | `"app.kubernetes.io/component"` |  |
| kafkaup-operator.kafkaSelectors.podLabel.value | string | `"kafka"` |  |
| kafkaup-operator.kafkaSelectors.statefulSetName | string | `"stackstate-kafka"` |  |
| kafkaup-operator.startVersion | string | `"2.3.1"` |  |
| minio.accessKey | string | `"setme"` | Secret key for MinIO. Default is set to an invalid value that will cause MinIO to not start up to ensure users of this Helm chart set an explicit value. |
| minio.azuregateway.replicas | int | `1` |  |
| minio.fullnameOverride | string | `"stackstate-minio"` | **N.B.: Do not change this value!** The fullname override for MinIO subchart is hardcoded so that the stackstate chart can refer to its components. |
| minio.image.registry | string | `"quay.io"` | MinIO image registry |
| minio.image.repository | string | `"stackstate/minio"` | MinIO image repository |
| minio.image.tag | string | `"RELEASE.2021-02-14T04-01-33Z-3118065624"` |  |
| minio.persistence.enabled | bool | `false` | Enables MinIO persistence. Must be enabled when MinIO is not configured as a gateway to AWS S3 or Azure Blob Storage. |
| minio.replicas | int | `1` | Number of MinIO replicas. |
| minio.s3gateway.replicas | int | `1` |  |
| minio.secretKey | string | `"setme"` |  |
| networkPolicy.enabled | bool | `false` | Enable creating of `NetworkPolicy` object and associated rules for StackState. |
| networkPolicy.spec | object | `{"ingress":[{"from":[{"podSelector":{}}]}],"podSelector":{"matchLabels":{}},"policyTypes":["Ingress"]}` | `NetworkPolicy` rules for StackState. |
| pull-secret.credentials | list | `[]` | Registry and assotiated credentials (username, password) that will be stored in the pull-secret |
| pull-secret.enabled | bool | `false` | Deploy the ImagePullSecret for the chart. |
| pull-secret.fullNameOverride | string | `""` | Name of the ImagePullSecret that will be created. This can be referenced by setting the `global.imagePullSecrets[0].name` value in the chart. |
| stackstate-agent.enabled | bool | `false` | Deploy the StackState Kubernetes Agent so StackState can monitor the cluster it runs in |
| stackstate-agent.stackstate.cluster.authToken | string | `nil` |  |
| stackstate-agent.stackstate.cluster.name | string | `nil` |  |
| stackstate-agent.stackstate.url | string | `"http://{{ include \"stackstate.hostname.prefix\" . }}-router:8080/receiver/stsAgent"` |  |
| stackstate.admin.authentication.password | string | `nil` | Password used for default platform "admin" api's (low-level tools) of the various services, username: platformadmin |
| stackstate.authentication | object | `{"adminPassword":null,"file":{},"keycloak":{},"ldap":{},"oidc":{},"roles":{"admin":[],"guest":[],"platformAdmin":[],"powerUser":[]},"serviceToken":{"bootstrap":{"roles":[],"token":"","ttl":"24h"}},"sessionLifetime":"7d"}` | Configure the authentication settings for StackState here. Only one of the authentication providers can be used, configuring multiple will result in an error. |
| stackstate.authentication.adminPassword | string | `nil` | Password for the 'admin' user that StackState creates by default |
| stackstate.authentication.file | object | `{}` | Configure users, their passwords and roles from (config) file |
| stackstate.authentication.keycloak | object | `{}` | Use Keycloak as authentication provider. See [Configuring Keycloak](#configuring-keycloak). |
| stackstate.authentication.ldap | object | `{}` | LDAP settings for StackState. See [Configuring LDAP](#configuring-ldap). |
| stackstate.authentication.oidc | object | `{}` | Use an OpenId Connect provider for authentication. See [Configuring OpenId Connect](#configuring-openid-connect). |
| stackstate.authentication.roles | object | `{"admin":[],"guest":[],"platformAdmin":[],"powerUser":[]}` | Extend the default role names in StackState |
| stackstate.authentication.roles.admin | list | `[]` | Extend the role names that have admin permissions (default: 'stackstate-admin') |
| stackstate.authentication.roles.guest | list | `[]` | Extend the role names that have guest permissions (default: 'stackstate-guest') |
| stackstate.authentication.roles.platformAdmin | list | `[]` | Extend the role names that have platform admin permissions (default: 'stackstate-platform-admin') |
| stackstate.authentication.roles.powerUser | list | `[]` | Extend the role names that have power user permissions (default: 'stackstate-power-user') |
| stackstate.authentication.serviceToken.bootstrap.roles | list | `[]` | The roles the service token assumes when it’s used for authentication |
| stackstate.authentication.serviceToken.bootstrap.token | string | `""` | The service token to set as bootstrap token |
| stackstate.authentication.serviceToken.bootstrap.ttl | string | `"24h"` | The amount of time the service token is valid for |
| stackstate.authentication.sessionLifetime | string | `"7d"` | Amount of time to keep a session when a user does not log in |
| stackstate.baseUrl | string | `nil` |  |
| stackstate.components.all.affinity | object | `{}` | Affinity settings for pod assignment on all components. |
| stackstate.components.all.deploymentStrategy.type | string | `"RecreateSingletonsOnly"` | Deployment strategy for StackState components. Possible values: `RollingUpdate`, `Recreate` and `RecreateSingletonsOnly`. `RecreateSingletonsOnly` uses `Recreate` for the singleton Deployments and `RollingUpdate` for the other Deployments. |
| stackstate.components.all.elasticsearchEndpoint | string | `""` | **Required if `elasticsearch.enabled` is `false`** Endpoint for shared Elasticsearch cluster. |
| stackstate.components.all.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods for all components. |
| stackstate.components.all.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object for all components. |
| stackstate.components.all.image.pullPolicy | string | `"IfNotPresent"` | The default pullPolicy used for all stateless components of StackState; invividual service `pullPolicy`s can be overriden (see below). |
| stackstate.components.all.image.pullSecretName | string | `nil` | Name of ImagePullSecret to use for all pods. |
| stackstate.components.all.image.pullSecretPassword | string | `nil` |  |
| stackstate.components.all.image.pullSecretUsername | string | `nil` |  |
| stackstate.components.all.image.registry | string | `"quay.io"` | Base container image registry for all StackState containers, except for the wait container and the container-tools container |
| stackstate.components.all.image.repositorySuffix | string | `"-stable"` |  |
| stackstate.components.all.image.tag | string | `"5.1.4"` | The default tag used for all stateless components of StackState; invividual service `tag`s can be overriden (see below). |
| stackstate.components.all.kafkaEndpoint | string | `""` | **Required if `elasticsearch.enabled` is `false`** Endpoint for shared Kafka broker. |
| stackstate.components.all.metrics.agentAnnotationsEnabled | bool | `true` | Put annotations on each pod to instruct the stackstate agent to scrape the metrics |
| stackstate.components.all.metrics.enabled | bool | `true` | Enable metrics port. |
| stackstate.components.all.metrics.servicemonitor.additionalLabels | object | `{}` | Additional labels for targeting Prometheus operator instances. |
| stackstate.components.all.metrics.servicemonitor.enabled | bool | `false` | Enable `ServiceMonitor` object; `all.metrics.enabled` *must* be enabled. |
| stackstate.components.all.nodeSelector | object | `{}` | Node labels for pod assignment on all components. |
| stackstate.components.all.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| stackstate.components.all.securityContext.fsGroup | int | `65534` | The GID (group ID) used to mount volumes |
| stackstate.components.all.securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| stackstate.components.all.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| stackstate.components.all.securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| stackstate.components.all.tolerations | list | `[]` | Toleration labels for pod assignment on all components. |
| stackstate.components.all.zookeeperEndpoint | string | `""` | **Required if `zookeeper.enabled` is `false`** Endpoint for shared Zookeeper nodes. |
| stackstate.components.api.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.api.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.api.config | string | `""` | Configuration file contents to customize the default StackState api configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.api.docslink | string | `""` | Documentation URL root to use in the product help page & tooltips. |
| stackstate.components.api.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.api.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.api.image.pullPolicy | string | `""` | `pullPolicy` used for the `api` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.api.image.repository | string | `"stackstate/stackstate-server"` | Repository of the api component Docker image. |
| stackstate.components.api.image.tag | string | `""` | Tag used for the `api` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.api.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.api.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `api` pods. |
| stackstate.components.api.replicaCount | int | `1` | Number of `api` replicas. |
| stackstate.components.api.resources | object | `{"limits":{"cpu":"2000m","ephemeral-storage":"2Gi","memory":"4000Mi"},"requests":{"cpu":"1500m","ephemeral-storage":"1Mi","memory":"4000Mi"}}` | Resource allocation for `api` pods. |
| stackstate.components.api.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.api.sizing.javaHeapMemoryFraction | string | `"50"` |  |
| stackstate.components.api.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.backup.resources.limits.cpu | string | `"3000m"` |  |
| stackstate.components.backup.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| stackstate.components.backup.resources.limits.memory | string | `"4000Mi"` |  |
| stackstate.components.backup.resources.requests.cpu | string | `"1000m"` |  |
| stackstate.components.backup.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| stackstate.components.backup.resources.requests.memory | string | `"4000Mi"` |  |
| stackstate.components.checks.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.checks.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.checks.config | string | `""` | Configuration file contents to customize the default StackState state configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.checks.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.checks.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.checks.image.pullPolicy | string | `""` | `pullPolicy` used for the `state` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.checks.image.repository | string | `"stackstate/stackstate-server"` | Repository of the sync component Docker image. |
| stackstate.components.checks.image.tag | string | `""` | Tag used for the `state` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.checks.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.checks.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `checks` pods. |
| stackstate.components.checks.replicaCount | int | `1` | Number of `checks` replicas. |
| stackstate.components.checks.resources | object | `{"limits":{"cpu":"2000m","ephemeral-storage":"1Gi","memory":"4000Mi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Mi","memory":"4000Mi"}}` | Resource allocation for `state` pods. |
| stackstate.components.checks.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.checks.sizing.javaHeapMemoryFraction | string | `"70"` |  |
| stackstate.components.checks.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.containerTools.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for container-tools containers. |
| stackstate.components.containerTools.image.registry | string | `"quay.io"` | Base container image registry for container-tools containers. |
| stackstate.components.containerTools.image.repository | string | `"stackstate/container-tools"` | Base container image repository for container-tools containers. |
| stackstate.components.containerTools.image.tag | string | `"1.1.4"` | Container image tag for container-tools containers. |
| stackstate.components.containerTools.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"2000Mi"}}` | Resource allocation for `kafkaTopicCreate` pods. |
| stackstate.components.correlate.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.correlate.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.correlate.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.correlate.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.correlate.image.pullPolicy | string | `""` | `pullPolicy` used for the `correlate` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.correlate.image.repository | string | `"stackstate/stackstate-correlate"` | Repository of the correlate component Docker image. |
| stackstate.components.correlate.image.tag | string | `""` | Tag used for the `correlate` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.correlate.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.correlate.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `correlate` pods. |
| stackstate.components.correlate.replicaCount | int | `1` | Number of `correlate` replicas. |
| stackstate.components.correlate.resources | object | `{"limits":{"cpu":"2","ephemeral-storage":"1Gi","memory":"1600Mi"},"requests":{"cpu":"2","ephemeral-storage":"1Mi","memory":"1600Mi"}}` | Resource allocation for `correlate` pods. |
| stackstate.components.correlate.sizing.baseMemoryConsumption | string | `"575Mi"` |  |
| stackstate.components.correlate.sizing.javaHeapMemoryFraction | string | `"78"` |  |
| stackstate.components.correlate.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.e2es.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.e2es.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.e2es.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.e2es.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.e2es.image.pullPolicy | string | `""` | `pullPolicy` used for the `e2es` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.e2es.image.repository | string | `"stackstate/stackstate-kafka-to-es"` | Repository of the e2es component Docker image. |
| stackstate.components.e2es.image.tag | string | `""` | Tag used for the `e2es` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.e2es.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.e2es.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `e2es` pods. |
| stackstate.components.e2es.replicaCount | int | `1` | Number of `e2es` replicas. |
| stackstate.components.e2es.resources | object | `{"limits":{"cpu":"500m","ephemeral-storage":"1Gi","memory":"1500Mi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"1500Mi"}}` | Resource allocation for `e2es` pods. |
| stackstate.components.e2es.sizing.baseMemoryConsumption | string | `"600Mi"` |  |
| stackstate.components.e2es.sizing.javaHeapMemoryFraction | string | `"85"` |  |
| stackstate.components.e2es.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.healthSync.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.healthSync.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.healthSync.cache.backend | string | `"mapdb"` | Type of cache backend used by the service, possible values are mapdb, rocksdb and inmemory |
| stackstate.components.healthSync.config | string | `""` | Configuration file contents to customize the default StackState healthSync configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.healthSync.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.healthSync.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.healthSync.image.pullPolicy | string | `""` | `pullPolicy` used for the `healthSync` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.healthSync.image.repository | string | `"stackstate/stackstate-server"` | Repository of the healthSync component Docker image. |
| stackstate.components.healthSync.image.tag | string | `""` | Tag used for the `healthSync` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.healthSync.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.healthSync.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `healthSync` pods. |
| stackstate.components.healthSync.replicaCount | int | `1` | Number of `healthSync` replicas. |
| stackstate.components.healthSync.resources | object | `{"limits":{"cpu":"1500m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Mi","memory":"2000Mi"}}` | Resource allocation for `healthSync` pods. |
| stackstate.components.healthSync.sizing.baseMemoryConsumption | string | `"1200Mi"` |  |
| stackstate.components.healthSync.sizing.javaHeapMemoryFraction | string | `"85"` |  |
| stackstate.components.healthSync.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.initializer.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.initializer.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.initializer.config | string | `""` | Configuration file contents to customize the default StackState initializer configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.initializer.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.initializer.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.initializer.image.pullPolicy | string | `""` | `pullPolicy` used for the `initializer` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.initializer.image.repository | string | `"stackstate/stackstate-server"` | Repository of the initializer component Docker image. |
| stackstate.components.initializer.image.tag | string | `""` | Tag used for the `initializer` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.initializer.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.initializer.resources | object | `{"limits":{"cpu":"500m","ephemeral-storage":"1Gi","memory":"700Mi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"700Mi"}}` | Resource allocation for `initializer` pods. |
| stackstate.components.initializer.sizing.baseMemoryConsumption | string | `"460Mi"` |  |
| stackstate.components.initializer.sizing.javaHeapMemoryFraction | string | `"65"` |  |
| stackstate.components.initializer.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.kafkaTopicCreate.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.kafkaTopicCreate.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.registry | string | `"quay.io"` | Base container image registry for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.repository | string | `"stackstate/kafka"` | Base container image repository for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.tag | string | `"2.8.1-2738720666"` | Container image tag for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.kafkaTopicCreate.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"2000Mi"}}` | Resource allocation for `kafkaTopicCreate` pods. |
| stackstate.components.kafkaTopicCreate.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| stackstate.components.kafkaTopicCreate.securityContext.fsGroup | int | `1001` | The GID (group ID) used to mount volumes |
| stackstate.components.kafkaTopicCreate.securityContext.runAsGroup | int | `1001` | The GID (group ID) of the owning user of the process |
| stackstate.components.kafkaTopicCreate.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| stackstate.components.kafkaTopicCreate.securityContext.runAsUser | int | `1001` | The UID (user ID) of the owning user of the process |
| stackstate.components.kafkaTopicCreate.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.mm2es.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.mm2es.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.mm2es.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.mm2es.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.mm2es.image.pullPolicy | string | `""` | `pullPolicy` used for the `mm2es` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.mm2es.image.repository | string | `"stackstate/stackstate-kafka-to-es"` | Repository of the mm2es component Docker image. |
| stackstate.components.mm2es.image.tag | string | `""` | Tag used for the `mm2es` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.mm2es.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.mm2es.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `mm2es` pods. |
| stackstate.components.mm2es.replicaCount | int | `1` | Number of `mm2es` replicas. |
| stackstate.components.mm2es.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"1Gi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Mi","memory":"1Gi"}}` | Resource allocation for `mm2es` pods. |
| stackstate.components.mm2es.sizing.baseMemoryConsumption | string | `"600Mi"` |  |
| stackstate.components.mm2es.sizing.javaHeapMemoryFraction | string | `"85"` |  |
| stackstate.components.mm2es.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.nginxPrometheusExporter.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.registry | string | `"quay.io"` | Base container image registry for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.repository | string | `"stackstate/nginx-prometheus-exporter"` | Base container image repository for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.tag | string | `"0.9.0-2738682730"` | Container image tag for nginx-prometheus-exporter containers. |
| stackstate.components.problemProducer.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.problemProducer.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.problemProducer.config | string | `""` | Configuration file contents to customize the default StackState problemProducer configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.problemProducer.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.problemProducer.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.problemProducer.image.pullPolicy | string | `""` | `pullPolicy` used for the `problemProducer` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.problemProducer.image.repository | string | `"stackstate/stackstate-server"` | Repository of the problemProducer component Docker image. |
| stackstate.components.problemProducer.image.tag | string | `""` | Tag used for the `problemProducer` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.problemProducer.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.problemProducer.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `problemProducer` pods. |
| stackstate.components.problemProducer.replicaCount | int | `1` | Number of `problemProducer` replicas. |
| stackstate.components.problemProducer.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"2000Mi"}}` | Resource allocation for `problemProducer` pods. |
| stackstate.components.problemProducer.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.problemProducer.sizing.javaHeapMemoryFraction | string | `"80"` |  |
| stackstate.components.problemProducer.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.receiver.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.receiver.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.receiver.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.receiver.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.receiver.image.pullPolicy | string | `""` | `pullPolicy` used for the `receiver` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.receiver.image.repository | string | `"stackstate/stackstate-receiver"` | Repository of the receiver component Docker image. |
| stackstate.components.receiver.image.tag | string | `""` | Tag used for the `receiver` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.receiver.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.receiver.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `receiver` pods. |
| stackstate.components.receiver.replicaCount | int | `1` | Number of `receiver` replicas. |
| stackstate.components.receiver.resources | object | `{"limits":{"cpu":"3000m","ephemeral-storage":"1Gi","memory":"4Gi"},"requests":{"cpu":"3000m","ephemeral-storage":"1Mi","memory":"4Gi"}}` | Resource allocation for `receiver` pods. |
| stackstate.components.receiver.sizing.baseMemoryConsumption | string | `"700Mi"` |  |
| stackstate.components.receiver.sizing.javaHeapMemoryFraction | string | `"75"` |  |
| stackstate.components.receiver.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.router.accesslog.enabled | bool | `false` | Enable access logging on the router |
| stackstate.components.router.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.router.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.router.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.router.image.pullPolicy | string | `""` | `pullPolicy` used for the `router` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.router.image.registry | string | `"quay.io"` | Registry of the router component Docker image. |
| stackstate.components.router.image.repository | string | `"stackstate/envoy"` | Repository of the router component Docker image. |
| stackstate.components.router.image.tag | string | `"v1.19.1-2738711656"` | Tag used for the `router` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.router.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.router.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `router` pods. |
| stackstate.components.router.replicaCount | int | `1` | Number of `router` replicas. |
| stackstate.components.router.resources | object | `{"limits":{"cpu":"100m","ephemeral-storage":"1Gi","memory":"128Mi"},"requests":{"cpu":"100m","ephemeral-storage":"1Mi","memory":"128Mi"}}` | Resource allocation for `router` pods. |
| stackstate.components.router.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.server.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.server.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.server.config | string | `""` | Configuration file contents to customize the default StackState configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.server.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.server.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.server.image.pullPolicy | string | `""` | `pullPolicy` used for the `server` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.server.image.repository | string | `"stackstate/stackstate-server"` | Repository of the server component Docker image. |
| stackstate.components.server.image.tag | string | `""` | Tag used for the `server` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.server.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.server.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `server` pods. |
| stackstate.components.server.replicaCount | int | `1` | Number of `server` replicas. |
| stackstate.components.server.resources | object | `{"limits":{"cpu":"3600m","ephemeral-storage":"1Gi","memory":"8Gi"},"requests":{"cpu":"3600m","ephemeral-storage":"1Mi","memory":"8Gi"}}` | Resource allocation for `server` pods. |
| stackstate.components.server.sizing.baseMemoryConsumption | string | `"1700Mi"` |  |
| stackstate.components.server.sizing.javaHeapMemoryFraction | string | `"85"` |  |
| stackstate.components.server.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.slicing.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.slicing.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.slicing.config | string | `""` | Configuration file contents to customize the default StackState slicing configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.slicing.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.slicing.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.slicing.image.pullPolicy | string | `""` | `pullPolicy` used for the `slicing` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.slicing.image.repository | string | `"stackstate/stackstate-server"` | Repository of the slicing component Docker image. |
| stackstate.components.slicing.image.tag | string | `""` | Tag used for the `slicing` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.slicing.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.slicing.replicaCount | int | `1` | Number of `slicing` replicas. |
| stackstate.components.slicing.resources | object | `{"limits":{"cpu":"1500m","ephemeral-storage":"1Gi","memory":"1800Mi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Mi","memory":"1800Mi"}}` | Resource allocation for `slicing` pods. |
| stackstate.components.slicing.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.slicing.sizing.javaHeapMemoryFraction | string | `"60"` |  |
| stackstate.components.slicing.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.state.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.state.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.state.config | string | `""` | Configuration file contents to customize the default StackState state configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.state.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.state.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.state.image.pullPolicy | string | `""` | `pullPolicy` used for the `state` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.state.image.repository | string | `"stackstate/stackstate-server"` | Repository of the sync component Docker image. |
| stackstate.components.state.image.tag | string | `""` | Tag used for the `state` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.state.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.state.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `state` pods. |
| stackstate.components.state.replicaCount | int | `1` | Number of `state` replicas. |
| stackstate.components.state.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"750m","ephemeral-storage":"1Mi","memory":"2000Mi"}}` | Resource allocation for `state` pods. |
| stackstate.components.state.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.state.sizing.javaHeapMemoryFraction | string | `"80"` |  |
| stackstate.components.state.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.sync.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.sync.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.sync.cache.backend | string | `"mapdb"` | Type of cache backend used by the service, possible values are mapdb, rocksdb and inmemory |
| stackstate.components.sync.config | string | `""` | Configuration file contents to customize the default StackState sync configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.sync.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.sync.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.sync.image.pullPolicy | string | `""` | `pullPolicy` used for the `sync` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.sync.image.repository | string | `"stackstate/stackstate-server"` | Repository of the sync component Docker image. |
| stackstate.components.sync.image.tag | string | `""` | Tag used for the `sync` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.sync.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.sync.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `sync` pods. |
| stackstate.components.sync.replicaCount | int | `1` | Number of `sync` replicas. |
| stackstate.components.sync.resources | object | `{"limits":{"cpu":"3000m","ephemeral-storage":"1Gi","memory":"3500Mi"},"requests":{"cpu":"2000m","ephemeral-storage":"1Mi","memory":"3500Mi"}}` | Resource allocation for `sync` pods. |
| stackstate.components.sync.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.sync.sizing.javaHeapMemoryFraction | string | `"60"` |  |
| stackstate.components.sync.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.trace2es.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.trace2es.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.trace2es.enabled | bool | `true` | Enable/disable the trace2es service |
| stackstate.components.trace2es.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.trace2es.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.trace2es.image.pullPolicy | string | `""` | `pullPolicy` used for the `trace2es` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.trace2es.image.repository | string | `"stackstate/stackstate-kafka-to-es"` | Repository of the trace2es component Docker image. |
| stackstate.components.trace2es.image.tag | string | `""` | Tag used for the `trace2es` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.trace2es.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.trace2es.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `trace2es` pods. |
| stackstate.components.trace2es.replicaCount | int | `1` | Number of `trace2es` replicas. |
| stackstate.components.trace2es.resources | object | `{"limits":{"cpu":"500m","ephemeral-storage":"1Gi","memory":"1Gi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"1Gi"}}` | Resource allocation for `trace2es` pods. |
| stackstate.components.trace2es.sizing.baseMemoryConsumption | string | `"600Mi"` |  |
| stackstate.components.trace2es.sizing.javaHeapMemoryFraction | string | `"85"` |  |
| stackstate.components.trace2es.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.ui.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.ui.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.ui.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.ui.image.pullPolicy | string | `""` | `pullPolicy` used for the `ui` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.ui.image.repository | string | `"stackstate/stackstate-ui"` | Repository of the ui component Docker image. |
| stackstate.components.ui.image.tag | string | `""` | Tag used for the `ui` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.ui.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.ui.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `ui` pods. |
| stackstate.components.ui.replicaCount | int | `2` | Number of `ui` replicas. |
| stackstate.components.ui.resources | object | `{"limits":{"cpu":"50m","ephemeral-storage":"1Gi","memory":"64Mi"},"requests":{"cpu":"50m","ephemeral-storage":"1Mi","memory":"64Mi"}}` | Resource allocation for `ui` pods. |
| stackstate.components.ui.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| stackstate.components.ui.securityContext.fsGroup | int | `101` | The GID (group ID) used to mount volumes |
| stackstate.components.ui.securityContext.runAsGroup | int | `101` | The GID (group ID) of the owning user of the process |
| stackstate.components.ui.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| stackstate.components.ui.securityContext.runAsUser | int | `101` | The UID (user ID) of the owning user of the process |
| stackstate.components.ui.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.viewHealth.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.viewHealth.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.viewHealth.config | string | `""` | Configuration file contents to customize the default StackState viewHealth configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.viewHealth.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.viewHealth.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.viewHealth.image.pullPolicy | string | `""` | `pullPolicy` used for the `viewHealth` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.viewHealth.image.repository | string | `"stackstate/stackstate-server"` | Repository of the viewHealth component Docker image. |
| stackstate.components.viewHealth.image.tag | string | `""` | Tag used for the `viewHealth` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.viewHealth.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.viewHealth.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `viewHealth` pods. |
| stackstate.components.viewHealth.replicaCount | int | `1` | Number of `viewHealth` replicas. |
| stackstate.components.viewHealth.resources | object | `{"limits":{"cpu":"2000m","ephemeral-storage":"1Gi","memory":"2700Mi"},"requests":{"cpu":"2000m","ephemeral-storage":"1Mi","memory":"2700Mi"}}` | Resource allocation for `viewHealth` pods. |
| stackstate.components.viewHealth.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.viewHealth.sizing.javaHeapMemoryFraction | string | `"55"` |  |
| stackstate.components.viewHealth.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.wait.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for wait containers. |
| stackstate.components.wait.image.registry | string | `"quay.io"` | Base container image registry for wait containers. |
| stackstate.components.wait.image.repository | string | `"stackstate/wait"` | Base container image repository for wait containers. |
| stackstate.components.wait.image.tag | string | `"1.0.7-2755960650"` | Container image tag for wait containers. |
| stackstate.experimental.server.split | boolean | `true` | Run a single service server or split in multiple sub services as api, state .... |
| stackstate.java | object | `{"trustStore":null,"trustStoreBase64Encoded":null,"trustStorePassword":null}` | Extra Java configuration for StackState |
| stackstate.java.trustStore | string | `nil` | Java TrustStore (cacerts) file to use |
| stackstate.java.trustStoreBase64Encoded | string | `nil` | Base64 encoded Java TrustStore (cacerts) file to use. Ignored if stackstate.java.trustStore is set. |
| stackstate.java.trustStorePassword | string | `nil` | Password to access the Java TrustStore (cacerts) file |
| stackstate.license.key | string | `nil` | **PROVIDE YOUR LICENSE KEY HERE** The StackState license key needed to start the server. |
| stackstate.receiver.baseUrl | string | `nil` | **DEPRECATED** Use stackstate.baseUrl instead |
| stackstate.stackpacks.installed | list | `[]` | Specify a list of stackpacks to be always installed including their configuration, for an example see [Auto-installing StackPacks](#auto-installing-stackpacks) |
| zookeeper.commonLabels."app.kubernetes.io/part-of" | string | `"stackstate"` |  |
| zookeeper.enabled | bool | `true` | Enable / disable chart-based Zookeeper. |
| zookeeper.externalServers | string | `""` | If `zookeeper.enabled` is set to `false`, use this list of external Zookeeper servers instead. |
| zookeeper.fourlwCommandsWhitelist | string | `"mntr, ruok, stat, srvr"` | Zookeeper four-letter-word (FLW) commands that are enabled. |
| zookeeper.fullnameOverride | string | `"stackstate-zookeeper"` | Name override for Zookeeper child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| zookeeper.image.registry | string | `"quay.io"` | ZooKeeper image registry |
| zookeeper.image.repository | string | `"stackstate/zookeeper"` | ZooKeeper image repository |
| zookeeper.image.tag | string | `"3.6.3-2738717608"` | ZooKeeper image tag |
| zookeeper.metrics.enabled | bool | `true` | Enable / disable Zookeeper Prometheus metrics. |
| zookeeper.metrics.serviceMonitor | object | `{"enabled":false,"selector":{}}` |  |
| zookeeper.metrics.serviceMonitor.enabled | bool | `false` | Enable creation of `ServiceMonitor` objects for Prometheus operator. |
| zookeeper.metrics.serviceMonitor.selector | object | `{}` | Default selector to use to target a certain Prometheus instance. |
| zookeeper.pdb.create | bool | `true` |  |
| zookeeper.pdb.maxUnavailable | int | `1` |  |
| zookeeper.pdb.minAvailable | string | `""` |  |
| zookeeper.podAnnotations | object | `{"ad.stackstate.com/zookeeper.check_names":"[\"openmetrics\"]","ad.stackstate.com/zookeeper.init_configs":"[{}]","ad.stackstate.com/zookeeper.instances":"[ { \"prometheus_url\": \"http://%%host%%:9141/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"*\"] } ]"}` | Annotations for ZooKeeper pod. |
| zookeeper.podLabels."app.kubernetes.io/part-of" | string | `"stackstate"` |  |
| zookeeper.replicaCount | int | `3` | Default amount of Zookeeper replicas to provision. |
| zookeeper.resources.limits.cpu | string | `"250m"` |  |
| zookeeper.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| zookeeper.resources.limits.memory | string | `"512Mi"` |  |
| zookeeper.resources.requests.cpu | string | `"100m"` |  |
| zookeeper.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| zookeeper.resources.requests.memory | string | `"512Mi"` |  |

## Authentication

For more details on configuring authentication please refer to the [StackState documentation](https://docs.stackstate.com).

### Configuring OpenId connect

Create a `oidc_values.yaml similar to the example below and add it as an additional argument to the installation:

```
helm install \
  --values oidc_values.yaml \
  ... \
stackstate/stackstate
```

Example:

```yaml
stackstate:
  authentication:
    oidc:
      clientId: stackstate-client-id
      secret: "some-secret"
      discoveryUri: http://oidc-provider/.well-known/openid-configuration
      authenticationMethod: client_secret_basic
      jwsAlgorithm: RS256
      scope: ["email", "fullname"]
      jwtClaims:
        usernameField: email
        groupsField: groups
      customParameters:
        access_type: offline # Custom request parameter
```

### Configuring Keycloak

Create a `keycloak_values.yaml similar to the example below and add it as an additional argument to the installation:

```
helm install \
  --values keycloak_values.yaml \
  ... \
stackstate/stackstate
```

Example:
```yaml
stackstate:
  authentication:
    keycloak:
      url: http://keycloak-server/auth
      realm: test
      clientId: stackstate-client-id
      secret: "some-secret"
      authenticationMethod: client_secret_basic
      jwsAlgorithm: RS256
```

If the defaults of Keycloak don't suit your needs you can extend the yaml to select a different field as username and add groups/roles from fields other than the `roles` in keycloak:
```
stackstate:
  authentication:
    keycloak:
      jwtClaims:
        usernameField: email
        groupsField: groups
```

### Configuring LDAP

To use LDAP create a ldap_values.yaml similar to the example below (update for your LDAP configuration of course). Next to these keys there are 2 optional values that can be set but usually need to be read from a file so you'd specify them on the helm command line (see below):
* `stackstate.authentication.ldap.ssl.trustStore`: The Certificate Truststore to verify server certificates against
* `stackstate.authentication.ldap.ssl.trustCertificates`: The client Certificate trusted by the server (supports PEM, DER and PKCS7 formats)

There are also Base64 Encoded analogues of these values. They are ignored if `trustCertificates` and/or `trustStore` are set:
* `stackstate.authentication.ldap.ssl.trustStoreBase64Encoded`
* `stackstate.authentication.ldap.ssl.trustCertificatesBase64Encoded`
**Note: The reason for `*Base64Encoded` values is this is the only way to upload binary files via KOTS Config**

Only one of `trustCertificates`/`trustCertificatesBase64Encoded` or `trustStore`/`trustStoreBase64Encoded` will be used, `trustCertificates` takes precedence over `trustStore`, and `trustStore`/`trustCertificates` takes precedence over `trustStoreBase64Encoded`/`trustCertificatesBase64Encoded` respectively.

In order to search the groups that a user belongs to and from those get the roles the user can have in StackState we need to config values `rolesKey` and `groupMemberKey`.

Those values in the end will be used to form a filter (and extract the relevant attribute) that looks like:
```({groupMemberKey}=email=admin@sts.com,ou=employees,dc=stackstate,dc=com)```

This returns an entry similar to this:
```dn: {rolesKey}=stackstate-admin,ou=Group,ou=employee,o=stackstate,cn=people,dc=stackstate,dc=com```

Via the {rolesKey} we will get `stackstate-admin` as the role.

Note that the order of the parameters is of importance.

```yaml
stackstate:
  authentication:
    ldap:
      host: ldap-server
      port: 10636 # Standard LDAP SSL port, 10398 for plain LDAP
      bind:
        dn: ou=acme,dc=com # The bind DN to use to authenticate to LDAP
        password: foobar   # The bind password to use to authenticate to LDAP
      ssl:
        type: ssl          # Optional: The SSL Connection type to use to connect to LDAP (Either `ssl` or `starttls`)
      userQuery:
        emailKey: email
        usernameKey: cn
        parameters:
          - ou: employees
          - dc: stackstate
          - dc: com
      groupQuery:
        groupMemberKey: member
        rolesKey: cn
        parameters:
          - ou: groups
          - dc: stackstate
          - dc: com
```

The `trustStore` and `trustCertificates` values need to be set from the command line, as they typically contain binary data. A sample command for this looks like:

```shell
helm install \
--set-file stackstate.authentication.ldap.ssl.trustStore=./ldap-cacerts \
--set-file stackstate.authentication.ldap.ssl.trustCertificates=./ldap-certificate.pem \
--values ldap_values.yaml \
... \
stackstate/stackstate
```

### Configuring file based authentication

If you don't have an external identity provider you can configure users and there roles directly in StackState via a configuration file (a change will result in a restart of the API).

To use this create a `file_auth_values.yaml similar to the example below and add it as an additional argument to the installation:

```
helm install \
  --values file_auth_values.yaml \
  ... \
stackstate/stackstate
```

Example that creates 4 different users with the standard roles provided by Stackstate (see our [docs](https://docs.stackstate.com)):
```
stackstate:
  authentication:
    file:
      logins:
        - username: administrator
          passwordHash: 098f6bcd4621d373cade4e832627b4f6
          roles: [stackstate-admin]
        - username: guest1
          passwordHash: 098f6bcd4621d373cade4e832627b4f6
          roles: [stackstate-guest]
        - username: guest2
          passwordHash: 098f6bcd4621d373cade4e832627b4f6
          roles: [ stackstate-guest ]
        - username: maintainer
          passwordHash: 098f6bcd4621d373cade4e832627b4f6
          roles: [stackstate-power-user, stackstate-guest]
```

## Auto-installing StackPacks
It can be useful to have some StackPacks always installed in StackState. For example the Kuberentes StackPack configured for the cluster running StackState. For that purpose the value `stackstate.stackpacks.installed` can be used to specify the StackPacks that should be installed (by name) and their (required) configuration. As an example here the Kubernetes StackPack will be pre-installed:
```
stackstate:
  stackpacks:
    installed:
      - name: kubernetes
        configuration:
          kubernetes_cluster_name: test
```
