distributed
===========
Helm chart for StackState distributed -- all components split into microservices.

Current chart version is `0.1.2`

Source code can be found [here](https://gitlab.com/stackvista/stackstate.git)

## Chart Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com | kafka | 6.1.X |
| https://charts.bitnami.com | zookeeper | 5.1.X |
| https://helm.elastic.co | elasticsearch | 7.4.X |
| https://helm.stackstate.io | common | 0.1.X |
| https://helm.stackstate.io | hbase | 0.1.X |

## Required Values

In order to successfully install this chart, you **must** provide the following variables:
* `stackstate.license.key`
* `stackstate.receiver.baseUrl`

Install them on the command line on Helm with the following command:

```shell
helm install \
--set stackstate.license.key=<your-license-key> \
--set stackstate.receiver.baseUrl=<your-base-url> \
stackstate/distributed
```

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| elasticsearch.clusterName | string | `"distributed-elasticsearch"` | Name override for Elasticsearch child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| elasticsearch.enabled | bool | `true` | Enable / disable chart-based Elasticsearch. |
| elasticsearch.extraEnvs | list | `[{"name":"action.auto_create_index","value":"true"},{"name":"indices.query.bool.max_clause_count","value":"10000"}]` | Extra settings that StackState uses for Elasticsearch. |
| elasticsearch.minimumMasterNodes | int | `1` | Minimum number of Elasticsearch master nodes. |
| elasticsearch.replicas | int | `1` | Number of Elasticsearch replicas. |
| global.imagePullSecrets | list | `[]` | imagePullSecrets for all containers. |
| hbase.enabled | bool | `true` | Enable / disable chart-based HBase. |
| hbase.hbase.master.replicaCount | int | `1` | Number of HBase master node replicas. |
| hbase.hbase.regionserver.replicaCount | int | `1` | Number of HBase regionserver node replicas. |
| hbase.hdfs.datanode.replicaCount | int | `1` | Number of HDFS datanode replicas. |
| hbase.tephra.replicaCount | int | `1` | Number of Tephra replicas. |
| hbase.zookeeper.enabled | bool | `false` | Disable Zookeeper from the HBase chart **Don't change unless otherwise specified**. |
| hbase.zookeeper.externalServers | string | `"distributed-zookeeper-headless"` | External Zookeeper if not used bundled Zookeeper chart **Don't change unless otherwise specified**. |
| ingress.annotations | object | `{}` | Annotations for ingress objects. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| ingress.hosts | list | `[]` | List of ingress hostnames; the paths are fixed to StackState backend services |
| ingress.tls | list | `[]` | List of ingress TLS certificates to use. |
| kafka.enabled | bool | `true` | Enable / disable chart-based Kafka. |
| kafka.externalZookeeper.servers | string | `"distributed-zookeeper-headless"` | External Zookeeper if not used bundled Zookeeper chart **Don't change unless otherwise specified**. |
| kafka.logRetentionHours | int | `24` | The minimum age of a log file to be eligible for deletion due to age. |
| kafka.replicaCount | int | `1` | Number of Kafka replicas. |
| kafka.resources | object | `{"limits":{"cpu":"500m","memory":"2Gi"},"requests":{"cpu":"500m","memory":"2Gi"}}` | Kafka resources per pods. |
| kafka.zookeeper.enabled | bool | `false` | Disable Zookeeper from the Kafka chart **Don't change unless otherwise specified**. |
| stackstate.components.all.affinity | object | `{}` | Affinity settings for pod assignment on all components. |
| stackstate.components.all.elasticsearchEndpoint | string | `""` | **Required if `elasticsearch.enabled` is `false`** Endpoint for shared Elasticsearch cluster. |
| stackstate.components.all.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods for all components. |
| stackstate.components.all.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object for all components. |
| stackstate.components.all.image.pullPolicy | string | `"IfNotPresent"` | The default pullPolicy used for all stateless components of StackState; invividual service `pullPolicy`s can be overriden (see below). |
| stackstate.components.all.image.tag | string | `"master"` | The default tag used for all stateless components of StackState; invividual service `tag`s can be overriden (see below). |
| stackstate.components.all.kafkaEndpoint | string | `""` | **Required if `elasticsearch.enabled` is `false`** Endpoint for shared Kafka broker. |
| stackstate.components.all.nodeSelector | object | `{}` | Node labels for pod assignment on all components. |
| stackstate.components.all.tolerations | list | `[]` | Toleration labels for pod assignment on all components. |
| stackstate.components.all.zookeeperEndpoint | string | `""` | **Required if `zookeeper.enabled` is `false`** Endpoint for shared Zookeeper nodes. |
| stackstate.components.correlate.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.correlate.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.correlate.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.correlate.image.pullPolicy | string | `""` | `pullPolicy` used for the `correlate` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.correlate.image.repository | string | `"quay.io/stackstate/stackstate-correlate"` | Repository of the correlate component Docker image. |
| stackstate.components.correlate.image.tag | string | `""` | Tag used for the `correlate` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.correlate.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.correlate.replicaCount | int | `1` | Number of `correlate` replicas. |
| stackstate.components.correlate.resources | object | `{"limits":{"cpu":"500m","memory":"2Gi"},"requests":{"cpu":"500m","memory":"2Gi"}}` | Resource allocation for `correlate` pods. |
| stackstate.components.correlate.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.k2es.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.k2es.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.k2es.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.k2es.image.pullPolicy | string | `""` | `pullPolicy` used for the `k2es` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.k2es.image.repository | string | `"quay.io/stackstate/stackstate-kafka-to-es"` | Repository of the k2es component Docker image. |
| stackstate.components.k2es.image.tag | string | `""` | Tag used for the `k2es` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.k2es.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.k2es.replicaCount | int | `1` | Number of `k2es` replicas. |
| stackstate.components.k2es.resources | object | `{"limits":{"cpu":"500m","memory":"1Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}` | Resource allocation for `k2es` pods. |
| stackstate.components.k2es.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.receiver.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.receiver.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.receiver.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.receiver.image.pullPolicy | string | `""` | `pullPolicy` used for the `receiver` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.receiver.image.repository | string | `"quay.io/stackstate/stackstate-receiver"` | Repository of the receiver component Docker image. |
| stackstate.components.receiver.image.tag | string | `""` | Tag used for the `receiver` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.receiver.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.receiver.replicaCount | int | `1` | Number of `receiver` replicas. |
| stackstate.components.receiver.resources | object | `{"limits":{"cpu":"1","memory":"2Gi"},"requests":{"cpu":"1","memory":"2Gi"}}` | Resource allocation for `receiver` pods. |
| stackstate.components.receiver.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.server.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.server.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.server.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.server.image.pullPolicy | string | `""` | `pullPolicy` used for the `server` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.server.image.repository | string | `"quay.io/stackstate/stackstate-server"` | Repository of the server component Docker image. |
| stackstate.components.server.image.tag | string | `""` | Tag used for the `server` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.server.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.server.replicaCount | int | `1` | Number of `server` replicas. |
| stackstate.components.server.resources | object | `{"limits":{"cpu":"3800m","memory":"8Gi"},"requests":{"cpu":"3800m","memory":"8Gi"}}` | Resource allocation for `server` pods. |
| stackstate.components.server.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.ui.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.ui.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.ui.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.ui.image.pullPolicy | string | `""` | `pullPolicy` used for the `ui` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.ui.image.repository | string | `"quay.io/stackstate/stackstate-ui"` | Repository of the ui component Docker image. |
| stackstate.components.ui.image.tag | string | `""` | Tag used for the `ui` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.ui.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.ui.replicaCount | int | `1` | Number of `ui` replicas. |
| stackstate.components.ui.resources | object | `{"limits":{"cpu":"50m","memory":"64Mi"},"requests":{"cpu":"50m","memory":"64Mi"}}` | Resource allocation for `ui` pods. |
| stackstate.components.ui.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.license.key | string | `nil` | **PROVIDE YOUR LICENSE KEY HERE** The StackState license key needed to start the server. |
| stackstate.receiver.apiKey | string | `""` | API key to be used by the Receiver; if no key is provided, a random one will be generated for you. |
| stackstate.receiver.baseUrl | string | `nil` | **PROVIDE YOUR BASE URL HERE** Externally visible baseUrl of the StackState endpoints. |
| zookeeper.enabled | bool | `true` | Enable / disable chart-based Zookeeper. |
| zookeeper.externalServers | string | `""` | If `zookeeper.enabled` is set to `false`, use this list of external Zookeeper servers instead. |
| zookeeper.fourlwCommandsWhitelist | string | `"mntr, ruok, stat, srvr"` | Zookeeper four-letter-word (FLW) commands that are enabled. |
| zookeeper.fullnameOverride | string | `"distributed-zookeeper"` | Name override for Zookeeper child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| zookeeper.metrics.enabled | bool | `true` | Enable / disable Zookeeper Prometheus metrics. |
| zookeeper.metrics.serviceMonitor.enabled | bool | `true` | Enable creation of `ServiceMonitor` objects for Prometheus operator. |
| zookeeper.metrics.serviceMonitor.selector | object | `{"release":"prometheus-operator"}` | Default selector to use to target a certain Prometheus instance. |
| zookeeper.replicaCount | int | `1` | Default amount of Zookeeper replicas to provision. |
