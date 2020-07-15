stackstate
==========
Helm chart for StackState

Current chart version is `0.4.62`

Source code can be found [here](https://gitlab.com/stackvista/stackstate.git)

## Chart Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | kafka | 11.5.1 |
| https://charts.bitnami.com/bitnami | zookeeper | 5.16.0 |
| https://helm.stackstate.io | anomaly-detection | 4.1.6 |
| https://helm.stackstate.io | cluster-agent | 0.4.1 |
| https://helm.stackstate.io | common | 0.4.3 |
| https://helm.stackstate.io | elasticsearch | 7.6.2-stackstate.3 |
| https://helm.stackstate.io | hbase | 0.1.33 |

## Required Values

In order to successfully install this chart, you **must** provide the following variables:
* `stackstate.license.key`
* `stackstate.receiver.baseUrl`

Install them on the command line on Helm with the following command:

```shell
helm install \
--set stackstate.license.key=<your-license-key> \
--set stackstate.receiver.baseUrl=<your-base-url> \
stackstate/stackstate
```

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| anomaly-detection.enabled | bool | `false` | Enables anomaly detection chart |
| anomaly-detection.image.imagePullPolicy | string | `"Always"` | The default pullPolicy used for anomaly detection pods. |
| anomaly-detection.image.pullSecretName | string | `nil` | Name of ImagePullSecret to use for all pods. |
| anomaly-detection.image.pullSecretPassword | string | `nil` |  |
| anomaly-detection.image.pullSecretUsername | string | `nil` | Password used to login to the registry to pull Docker images of all pods. |
| anomaly-detection.image.registry | string | `"quay.io"` | Base container image registry for all containers, except for the wait container |
| anomaly-detection.image.spotlightRepository | string | `"stackstate/spotlight"` | Repository of the spotlight Docker image. |
| anomaly-detection.image.tag | string | `"latest"` | the chart image tag, e.g. 4.1.0-latest |
| anomaly-detection.ingress | object | `{"annotations":{},"enabled":false,"hostname":null,"hosts":[],"port":8090,"tls":null}` | Status interface ingress |
| anomaly-detection.ingress.enabled | bool | `false` | Enables ingress controller for status interface |
| anomaly-detection.ingress.hostname | string | `nil` | Status interface hostname e.g. spotlight.local.domain |
| anomaly-detection.stackstate.authRoleName | string | `"stackstate-admin"` | Stackstate Role that used by spotlight for authentication, it is mapped to the stackstate role with the same name. |
| anomaly-detection.stackstate.instance | string | `"http://stackstate-server-headless:7070"` | **Required Stackstate instance URL, e.g http://stackstate-server-headless:7070 |
| anomaly-detection.threadWorkers | int | `5` | The number of worker threads. |
| caspr.enabled | bool | `false` | Enable CASPR compatible values. |
| cluster-agent.enabled | bool | `false` |  |
| cluster-agent.stackstate.cluster.authToken | string | `nil` |  |
| cluster-agent.stackstate.cluster.name | string | `nil` |  |
| cluster-agent.stackstate.url | string | `"http://stackstate-router:8080/receiver/stsAgent"` |  |
| elasticsearch.clusterHealthCheckParams | string | `"wait_for_status=yellow&timeout=1s"` | The Elasticsearch cluster health status params that will be used by readinessProbe command |
| elasticsearch.clusterName | string | `"stackstate-elasticsearch"` | Name override for Elasticsearch child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| elasticsearch.enabled | bool | `true` | Enable / disable chart-based Elasticsearch. |
| elasticsearch.esJavaOpts | string | `"-Xmx2g -Xms2g"` | JVM options |
| elasticsearch.extraEnvs | list | `[{"name":"action.auto_create_index","value":"true"},{"name":"indices.query.bool.max_clause_count","value":"10000"}]` | Extra settings that StackState uses for Elasticsearch. |
| elasticsearch.imageTag | string | `"7.4.1"` | Elasticsearch version. |
| elasticsearch.minimumMasterNodes | int | `2` | Minimum number of Elasticsearch master nodes. |
| elasticsearch.nodeGroup | string | `"master"` |  |
| elasticsearch.replicas | int | `3` | Number of Elasticsearch replicas. |
| elasticsearch.resources | object | `{"limits":{"cpu":"2000m","memory":"3Gi"},"requests":{"cpu":"2000m","memory":"3Gi"}}` | Override Elasticsearch resources |
| elasticsearch.volumeClaimTemplate | object | `{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"250Gi"}}}` | PVC template defaulting to 250Gi default volumes |
| global.receiverApiKey | string | `""` | API key to be used by the Receiver; if no key is provided, a random one will be generated for you. |
| hbase.enabled | bool | `true` | Enable / disable chart-based HBase. |
| hbase.hbase.master.replicaCount | int | `2` | Number of HBase master node replicas. |
| hbase.hbase.regionserver.replicaCount | int | `3` | Number of HBase regionserver node replicas. |
| hbase.hdfs.datanode.replicaCount | int | `3` | Number of HDFS datanode replicas. |
| hbase.hdfs.secondarynamenode.enabled | bool | `true` |  |
| hbase.stackgraph.image.tag | string | `"1.7.0"` | The StackGraph server version, must be compatible with the StackState version |
| hbase.tephra.replicaCount | int | `2` | Number of Tephra replicas. |
| hbase.zookeeper.enabled | bool | `false` | Disable Zookeeper from the HBase chart **Don't change unless otherwise specified**. |
| hbase.zookeeper.externalServers | string | `"stackstate-zookeeper-headless"` | External Zookeeper if not used bundled Zookeeper chart **Don't change unless otherwise specified**. |
| ingress.annotations | object | `{}` | Annotations for ingress objects. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| ingress.hosts | list | `[]` | List of ingress hostnames; the paths are fixed to StackState backend services |
| ingress.path | string | `"/"` |  |
| ingress.tls | list | `[]` | List of ingress TLS certificates to use. |
| kafka.command[0] | string | `"bash"` |  |
| kafka.command[1] | string | `"-ec"` |  |
| kafka.command[2] | string | `"KAFKA_CFG_BROKER_ID=\"${MY_POD_NAME#\"{{- include \"kafka.fullname\" . }}-\"}\"\nif [[ -f /bitnami/kafka/data/meta.properties ]] && ! grep -q -e \"^broker.id=${KAFKA_CFG_BROKER_ID}$\" /bitnami/kafka/data/meta.properties; then\n  echo \"Forcing broker.id in /bitnami/kafka/data/meta.properties to ${KAFKA_CFG_BROKER_ID}\"\n  sed -i \"s/^broker.id=.*$/broker.id=${KAFKA_CFG_BROKER_ID}/\" /bitnami/kafka/data/meta.properties\nfi\n/scripts/setup.sh\n"` |  |
| kafka.enabled | bool | `true` | Enable / disable chart-based Kafka. |
| kafka.externalZookeeper.servers | string | `"stackstate-zookeeper-headless"` | External Zookeeper if not used bundled Zookeeper chart **Don't change unless otherwise specified**. |
| kafka.fullnameOverride | string | `"stackstate-kafka"` | Name override for Kafka child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| kafka.image.tag | string | `"2.3.1-debian-9-r41"` | Default tag used for Kafka. **Since StackState relies on this specific version, it's advised NOT to change this.** |
| kafka.livenessProbe.initialDelaySeconds | int | `45` | Delay before readiness probe is initiated. |
| kafka.logRetentionHours | int | `24` | The minimum age of a log file to be eligible for deletion due to age. |
| kafka.metrics.jmx.enabled | bool | `true` | Whether or not to expose JMX metrics to Prometheus. |
| kafka.metrics.kafka.enabled | bool | `true` | Whether or not to create a standalone Kafka exporter to expose Kafka metrics. |
| kafka.metrics.serviceMonitor.enabled | bool | `false` | If `true`, creates a Prometheus Operator `ServiceMonitor` (also requires `kafka.metrics.kafka.enabled` or `kafka.metrics.jmx.enabled` to be `true`). |
| kafka.metrics.serviceMonitor.interval | string | `"20s"` | How frequently to scrape metrics. |
| kafka.metrics.serviceMonitor.selector | object | `{}` | Selector to target Prometheus instance. |
| kafka.persistence.size | string | `"50Gi"` | Size of persistent volume for each Kafka pod |
| kafka.readinessProbe.initialDelaySeconds | int | `45` | Delay before readiness probe is initiated. |
| kafka.replicaCount | int | `3` | Number of Kafka replicas. |
| kafka.resources | object | `{"limits":{"memory":"2Gi"},"requests":{"memory":"2Gi"}}` | Kafka resources per pods. |
| kafka.zookeeper.enabled | bool | `false` | Disable Zookeeper from the Kafka chart **Don't change unless otherwise specified**. |
| networkPolicy.enabled | bool | `false` | Enable creating of `NetworkPolicy` object and associated rules for StackState. |
| networkPolicy.spec | object | `{"ingress":[{"from":[{"podSelector":{}}]}],"podSelector":{"matchLabels":{}},"policyTypes":["Ingress"]}` | `NetworkPolicy` rules for StackState. |
| stackstate.admin.authentication.password | string | `nil` | Password used for maintenance "admin" api's (low-level tools) of the various services, username: admin |
| stackstate.authentication | object | `{"adminPassword":null,"ldap":{}}` | (Secret) authentication settings for StackState |
| stackstate.authentication.ldap | object | `{}` | LDAP settings for StackState. See [Configuring LDAP](#configuring-ldap) |
| stackstate.components.all.affinity | object | `{}` | Affinity settings for pod assignment on all components. |
| stackstate.components.all.elasticsearchEndpoint | string | `""` | **Required if `elasticsearch.enabled` is `false`** Endpoint for shared Elasticsearch cluster. |
| stackstate.components.all.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods for all components. |
| stackstate.components.all.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object for all components. |
| stackstate.components.all.image.pullPolicy | string | `"Always"` | The default pullPolicy used for all stateless components of StackState; invividual service `pullPolicy`s can be overriden (see below). |
| stackstate.components.all.image.pullSecretName | string | `nil` | Name of ImagePullSecret to use for all pods. |
| stackstate.components.all.image.pullSecretPassword | string | `nil` |  |
| stackstate.components.all.image.pullSecretUsername | string | `nil` |  |
| stackstate.components.all.image.registry | string | `"quay.io"` | Base container image registry for all containers, except for the wait container |
| stackstate.components.all.image.repositorySuffix | string | `"-stable"` |  |
| stackstate.components.all.image.tag | string | `"sts-v4-0-1"` | The default tag used for all stateless components of StackState; invividual service `tag`s can be overriden (see below). |
| stackstate.components.all.kafkaEndpoint | string | `""` | **Required if `elasticsearch.enabled` is `false`** Endpoint for shared Kafka broker. |
| stackstate.components.all.metrics.enabled | bool | `true` | Enable metrics port. |
| stackstate.components.all.metrics.servicemonitor.additionalLabels | object | `{}` | Additional labels for targeting Prometheus operator instances. |
| stackstate.components.all.metrics.servicemonitor.enabled | bool | `false` | Enable `ServiceMonitor` object; `all.metrics.enabled` *must* be enabled. |
| stackstate.components.all.nodeSelector | object | `{}` | Node labels for pod assignment on all components. |
| stackstate.components.all.tolerations | list | `[]` | Toleration labels for pod assignment on all components. |
| stackstate.components.all.zookeeperEndpoint | string | `""` | **Required if `zookeeper.enabled` is `false`** Endpoint for shared Zookeeper nodes. |
| stackstate.components.api.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.api.config | string | `""` | Configuration file contents to customize the default StackState api configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.api.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.api.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.api.image.pullPolicy | string | `""` | `pullPolicy` used for the `api` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.api.image.repository | string | `"stackstate/stackstate-server"` | Repository of the api component Docker image. |
| stackstate.components.api.image.tag | string | `""` | Tag used for the `api` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.api.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.api.resources | object | `{"limits":{"memory":"2Gi"},"requests":{"memory":"2Gi"}}` | Resource allocation for `api` pods. |
| stackstate.components.api.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.checks.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.checks.config | string | `""` | Configuration file contents to customize the default StackState state configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.checks.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.checks.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.checks.image.pullPolicy | string | `""` | `pullPolicy` used for the `state` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.checks.image.repository | string | `"stackstate/stackstate-server"` | Repository of the sync component Docker image. |
| stackstate.components.checks.image.tag | string | `""` | Tag used for the `state` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.checks.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.checks.resources | object | `{"limits":{"memory":"3Gi"},"requests":{"memory":"3Gi"}}` | Resource allocation for `state` pods. |
| stackstate.components.checks.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.correlate.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.correlate.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.correlate.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.correlate.image.pullPolicy | string | `""` | `pullPolicy` used for the `correlate` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.correlate.image.repository | string | `"stackstate/stackstate-correlate"` | Repository of the correlate component Docker image. |
| stackstate.components.correlate.image.tag | string | `""` | Tag used for the `correlate` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.correlate.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.correlate.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `correlate` pods. |
| stackstate.components.correlate.replicaCount | int | `1` | Number of `correlate` replicas. |
| stackstate.components.correlate.resources | object | `{"limits":{"memory":"2Gi"},"requests":{"memory":"2Gi"}}` | Resource allocation for `correlate` pods. |
| stackstate.components.correlate.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.initializer.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.initializer.config | string | `""` | Configuration file contents to customize the default StackState initializer configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.initializer.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.initializer.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.initializer.image.pullPolicy | string | `""` | `pullPolicy` used for the `initializer` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.initializer.image.repository | string | `"stackstate/stackstate-server"` | Repository of the initializer component Docker image. |
| stackstate.components.initializer.image.tag | string | `""` | Tag used for the `initializer` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.initializer.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.initializer.resources | object | `{"limits":{"memory":"1Gi"},"requests":{"memory":"128Mi"}}` | Resource allocation for `initializer` pods. |
| stackstate.components.initializer.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.k2es.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.k2es.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.k2es.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.k2es.image.pullPolicy | string | `""` | `pullPolicy` used for the `k2es` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.k2es.image.repository | string | `"stackstate/stackstate-kafka-to-es"` | Repository of the k2es component Docker image. |
| stackstate.components.k2es.image.tag | string | `""` | Tag used for the `k2es` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.k2es.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.k2es.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `k2es` pods. |
| stackstate.components.k2es.replicaCount | int | `1` | Number of `k2es` replicas. |
| stackstate.components.k2es.resources | object | `{"limits":{"memory":"1Gi"},"requests":{"memory":"1Gi"}}` | Resource allocation for `k2es` pods. |
| stackstate.components.k2es.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.k2es.traces | object | `{"enabled":false}` | Trace-related `k2es` settings. |
| stackstate.components.kafkaTopicCreate.image.registry | string | `"docker.io"` | Base container image registry for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.repository | string | `"bitnami/kafka"` | Base container image repository for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.tag | string | `"latest"` | Container image tag for kafka-topic-create containers. |
| stackstate.components.nginxPrometheusExporter.image.registry | string | `"docker.io"` | Base container image registry for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.repository | string | `"nginx/nginx-prometheus-exporter"` | Base container image repository for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.tag | string | `"0.4.2"` | Container image tag for nginx-prometheus-exporter containers. |
| stackstate.components.receiver.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.receiver.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.receiver.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.receiver.image.pullPolicy | string | `""` | `pullPolicy` used for the `receiver` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.receiver.image.repository | string | `"stackstate/stackstate-receiver"` | Repository of the receiver component Docker image. |
| stackstate.components.receiver.image.tag | string | `""` | Tag used for the `receiver` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.receiver.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.receiver.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `receiver` pods. |
| stackstate.components.receiver.replicaCount | int | `1` | Number of `receiver` replicas. |
| stackstate.components.receiver.resources | object | `{"limits":{"memory":"2Gi"},"requests":{"memory":"1Gi"}}` | Resource allocation for `receiver` pods. |
| stackstate.components.receiver.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.router.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.router.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.router.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.router.image.pullPolicy | string | `"Always"` | `pullPolicy` used for the `router` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.router.image.registry | string | `"docker.io"` | Registry of the router component Docker image. |
| stackstate.components.router.image.repository | string | `"envoyproxy/envoy-alpine"` | Repository of the router component Docker image. |
| stackstate.components.router.image.tag | string | `"v1.12.1"` | Tag used for the `router` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.router.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.router.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `router` pods. |
| stackstate.components.router.replicaCount | int | `1` | Number of `router` replicas. |
| stackstate.components.router.resources | object | `{"limits":{"cpu":"100m","memory":"128Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | Resource allocation for `router` pods. |
| stackstate.components.router.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.server.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.server.config | string | `""` | Configuration file contents to customize the default StackState configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.server.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.server.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.server.image.pullPolicy | string | `""` | `pullPolicy` used for the `server` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.server.image.repository | string | `"stackstate/stackstate-server"` | Repository of the server component Docker image. |
| stackstate.components.server.image.tag | string | `""` | Tag used for the `server` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.server.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.server.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `server` pods. |
| stackstate.components.server.resources | object | `{"limits":{"memory":"8Gi"},"requests":{"memory":"8Gi"}}` | Resource allocation for `server` pods. |
| stackstate.components.server.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.state.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.state.config | string | `""` | Configuration file contents to customize the default StackState state configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.state.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.state.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.state.image.pullPolicy | string | `""` | `pullPolicy` used for the `state` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.state.image.repository | string | `"stackstate/stackstate-server"` | Repository of the sync component Docker image. |
| stackstate.components.state.image.tag | string | `""` | Tag used for the `state` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.state.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.state.resources | object | `{"limits":{"memory":"2Gi"},"requests":{"memory":"2Gi"}}` | Resource allocation for `state` pods. |
| stackstate.components.state.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.sync.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.sync.config | string | `""` | Configuration file contents to customize the default StackState sync configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.sync.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.sync.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.sync.image.pullPolicy | string | `""` | `pullPolicy` used for the `sync` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.sync.image.repository | string | `"stackstate/stackstate-server"` | Repository of the sync component Docker image. |
| stackstate.components.sync.image.tag | string | `""` | Tag used for the `sync` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.sync.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.sync.resources | object | `{"limits":{"memory":"4Gi"},"requests":{"memory":"4Gi"}}` | Resource allocation for `sync` pods. |
| stackstate.components.sync.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.ui.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.ui.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.ui.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.ui.image.pullPolicy | string | `""` | `pullPolicy` used for the `ui` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.ui.image.repository | string | `"stackstate/stackstate-ui"` | Repository of the ui component Docker image. |
| stackstate.components.ui.image.tag | string | `""` | Tag used for the `ui` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.ui.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.ui.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `ui` pods. |
| stackstate.components.ui.replicaCount | int | `2` | Number of `ui` replicas. |
| stackstate.components.ui.resources | object | `{"limits":{"cpu":"50m","memory":"64Mi"},"requests":{"cpu":"50m","memory":"64Mi"}}` | Resource allocation for `ui` pods. |
| stackstate.components.ui.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.viewHealth.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.viewHealth.config | string | `""` | Configuration file contents to customize the default StackState viewHealth configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.viewHealth.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.viewHealth.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.viewHealth.image.pullPolicy | string | `""` | `pullPolicy` used for the `viewHealth` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.viewHealth.image.repository | string | `"stackstate/stackstate-server"` | Repository of the sync component Docker image. |
| stackstate.components.viewHealth.image.tag | string | `""` | Tag used for the `viewHealth` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.viewHealth.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.viewHealth.resources | object | `{"limits":{"memory":"2Gi"},"requests":{"memory":"2Gi"}}` | Resource allocation for `viewHealth` pods. |
| stackstate.components.viewHealth.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.wait.image.registry | string | `"docker.io"` | Base container image registry for wait containers. |
| stackstate.components.wait.image.repository | string | `"dokkupaas/wait"` | Base container image repository for wait containers. |
| stackstate.components.wait.image.tag | string | `"latest"` | Container image tag for wait containers. |
| stackstate.experimental.server.split | bool | `false` | (boolean) Run a single service server or split in multiple sub services as api, state .... |
| stackstate.java | object | `{"trustStore":null,"trustStorePassword":null}` | Extra Java configuration for StackState |
| stackstate.java.trustStore | string | `nil` | Java TrustStore (cacerts) file to use |
| stackstate.java.trustStorePassword | string | `nil` | Password to access the Java TrustStore (cacerts) file |
| stackstate.license.key | string | `nil` | **PROVIDE YOUR LICENSE KEY HERE** The StackState license key needed to start the server. |
| stackstate.receiver.baseUrl | string | `nil` | **PROVIDE YOUR BASE URL HERE** Externally visible baseUrl of the StackState endpoints. |
| zookeeper.enabled | bool | `true` | Enable / disable chart-based Zookeeper. |
| zookeeper.externalServers | string | `""` | If `zookeeper.enabled` is set to `false`, use this list of external Zookeeper servers instead. |
| zookeeper.fourlwCommandsWhitelist | string | `"mntr, ruok, stat, srvr"` | Zookeeper four-letter-word (FLW) commands that are enabled. |
| zookeeper.fullnameOverride | string | `"stackstate-zookeeper"` | Name override for Zookeeper child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| zookeeper.metrics.enabled | bool | `true` | Enable / disable Zookeeper Prometheus metrics. |
| zookeeper.metrics.serviceMonitor.enabled | bool | `false` | Enable creation of `ServiceMonitor` objects for Prometheus operator. |
| zookeeper.metrics.serviceMonitor.selector | object | `{}` | Default selector to use to target a certain Prometheus instance. |
| zookeeper.replicaCount | int | `3` | Default amount of Zookeeper replicas to provision. |

## Configuring LDAP

When using LDAP, a number of (secret) values can be passed through the Helm values. You can provide the following values to configure LDAP:

* `stackstate.authentication.ldap.bind.dn`: The bind DN to use to authenticate to LDAP
* `stackstate.authentication.ldap.bind.password`: The bind password to use to authenticate to LDAP
* `stackstate.authentication.ldap.ssl.type`: The SSL Connection type to use to connect to LDAP (Either `ssl` or `starttls`)
* `stackstate.authentication.ldap.ssl.trustStore`: The Certificate Truststore to verify server certificates against
* `stackstate.authentication.ldap.ssl.trustCertificates`: The client Certificate trusted by the server

The `trustStore` and `trustCertificates` values need to be set from the command line, as they typically contain binary data. A sample command for this looks like:

```shell
helm install \
--set-file stackstate.authentication.ldap.ssl.trustStore=./ldap-cacerts \
--set-file stackstate.authentication.ldap.ssl.trustCertificates=./ldap-certificate.pem \
... \
stackstate/stackstate
