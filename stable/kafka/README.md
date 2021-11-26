# kafka

![Version: 12.2.5-stackstate.2](https://img.shields.io/badge/Version-12.2.5--stackstate.2-informational?style=flat-square) ![AppVersion: 2.6.0](https://img.shields.io/badge/AppVersion-2.6.0-informational?style=flat-square)
Apache Kafka is a distributed streaming platform.
**Homepage:** <https://github.com/bitnami/charts/tree/master/bitnami/kafka>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Bitnami | containers@bitnami.com |  |
## Source Code

* <https://github.com/bitnami/bitnami-docker-kafka>
* <https://kafka.apache.org/>
## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | common | 1.1.1 |
| https://charts.bitnami.com/bitnami | zookeeper | 6.x.x |
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| advertisedListeners | list | `[]` |  |
| affinity | object | `{}` |  |
| allowPlaintextListener | bool | `true` |  |
| args | string | `nil` |  |
| auth.clientProtocol | string | `"plaintext"` |  |
| auth.interBrokerProtocol | string | `"plaintext"` |  |
| auth.jaas.clientPasswords | list | `[]` |  |
| auth.jaas.clientUsers[0] | string | `"user"` |  |
| auth.jaas.interBrokerPassword | string | `""` |  |
| auth.jaas.interBrokerUser | string | `"admin"` |  |
| auth.saslInterBrokerMechanism | string | `"plain"` |  |
| auth.saslMechanisms | string | `"plain,scram-sha-256,scram-sha-512"` |  |
| auth.tlsEndpointIdentificationAlgorithm | string | `"https"` |  |
| autoCreateTopicsEnable | bool | `true` |  |
| clusterDomain | string | `"cluster.local"` |  |
| command[0] | string | `"/scripts/setup.sh"` |  |
| commonAnnotations | object | `{}` |  |
| commonLabels | object | `{}` |  |
| containerSecurityContext | object | `{}` |  |
| customLivenessProbe | object | `{}` |  |
| customReadinessProbe | object | `{}` |  |
| defaultReplicationFactor | int | `1` |  |
| deleteTopicEnable | bool | `false` |  |
| externalAccess.autoDiscovery.enabled | bool | `false` |  |
| externalAccess.autoDiscovery.image.pullPolicy | string | `"IfNotPresent"` |  |
| externalAccess.autoDiscovery.image.pullSecrets | list | `[]` |  |
| externalAccess.autoDiscovery.image.registry | string | `"docker.io"` |  |
| externalAccess.autoDiscovery.image.repository | string | `"bitnami/kubectl"` |  |
| externalAccess.autoDiscovery.image.tag | string | `"1.17.13-debian-10-r21"` |  |
| externalAccess.autoDiscovery.resources.limits | object | `{}` |  |
| externalAccess.autoDiscovery.resources.requests | object | `{}` |  |
| externalAccess.enabled | bool | `false` |  |
| externalAccess.service.annotations | object | `{}` |  |
| externalAccess.service.loadBalancerIPs | list | `[]` |  |
| externalAccess.service.loadBalancerSourceRanges | list | `[]` |  |
| externalAccess.service.nodePorts | list | `[]` |  |
| externalAccess.service.port | int | `9094` |  |
| externalAccess.service.type | string | `"LoadBalancer"` |  |
| externalZookeeper.servers | list | `[]` |  |
| extraDeploy | list | `[]` |  |
| extraEnvVars | list | `[]` |  |
| extraVolumeMounts | list | `[]` |  |
| extraVolumes | list | `[]` |  |
| heapOpts | string | `"-Xmx1024m -Xms1024m"` |  |
| image.debug | bool | `false` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.pullSecrets | list | `[]` |  |
| image.registry | string | `"docker.io"` |  |
| image.repository | string | `"bitnami/kafka"` |  |
| image.tag | string | `"2.6.0-debian-10-r78"` |  |
| interBrokerListenerName | string | `"INTERNAL"` |  |
| listeners | list | `[]` |  |
| livenessProbe.enabled | bool | `true` |  |
| livenessProbe.initialDelaySeconds | int | `10` |  |
| livenessProbe.timeoutSeconds | int | `5` |  |
| logFlushIntervalMessages | int | `10000` |  |
| logFlushIntervalMs | int | `1000` |  |
| logPersistence.accessModes[0] | string | `"ReadWriteOnce"` |  |
| logPersistence.annotations | object | `{}` |  |
| logPersistence.enabled | bool | `false` |  |
| logPersistence.mountPath | string | `"/opt/bitnami/kafka/logs"` |  |
| logPersistence.size | string | `"8Gi"` |  |
| logRetentionBytes | string | `"_1073741824"` |  |
| logRetentionCheckIntervalMs | int | `300000` |  |
| logRetentionHours | int | `168` |  |
| logSegmentBytes | string | `"_1073741824"` |  |
| logsDirs | string | `"/bitnami/kafka/data"` |  |
| maxMessageBytes | string | `"_1000012"` |  |
| metrics.jmx.config | string | `"jmxUrl: service:jmx:rmi:///jndi/rmi://127.0.0.1:5555/jmxrmi\nlowercaseOutputName: true\nlowercaseOutputLabelNames: true\nssl: false\n{{- if .Values.metrics.jmx.whitelistObjectNames }}\nwhitelistObjectNames: [\"{{ join \"\\\",\\\"\" .Values.metrics.jmx.whitelistObjectNames }}\"]\n{{- end }}"` |  |
| metrics.jmx.enabled | bool | `false` |  |
| metrics.jmx.image.pullPolicy | string | `"IfNotPresent"` |  |
| metrics.jmx.image.pullSecrets | list | `[]` |  |
| metrics.jmx.image.registry | string | `"docker.io"` |  |
| metrics.jmx.image.repository | string | `"bitnami/jmx-exporter"` |  |
| metrics.jmx.image.tag | string | `"0.14.0-debian-10-r64"` |  |
| metrics.jmx.resources.limits | object | `{}` |  |
| metrics.jmx.resources.requests | object | `{}` |  |
| metrics.jmx.service.annotations."prometheus.io/path" | string | `"/"` |  |
| metrics.jmx.service.annotations."prometheus.io/port" | string | `"{{ .Values.metrics.jmx.service.port }}"` |  |
| metrics.jmx.service.annotations."prometheus.io/scrape" | string | `"true"` |  |
| metrics.jmx.service.loadBalancerSourceRanges | list | `[]` |  |
| metrics.jmx.service.nodePort | string | `""` |  |
| metrics.jmx.service.port | int | `5556` |  |
| metrics.jmx.service.type | string | `"ClusterIP"` |  |
| metrics.jmx.whitelistObjectNames[0] | string | `"kafka.controller:*"` |  |
| metrics.jmx.whitelistObjectNames[1] | string | `"kafka.server:*"` |  |
| metrics.jmx.whitelistObjectNames[2] | string | `"java.lang:*"` |  |
| metrics.jmx.whitelistObjectNames[3] | string | `"kafka.network:*"` |  |
| metrics.jmx.whitelistObjectNames[4] | string | `"kafka.log:*"` |  |
| metrics.kafka.enabled | bool | `false` |  |
| metrics.kafka.extraFlags | object | `{}` |  |
| metrics.kafka.image.pullPolicy | string | `"IfNotPresent"` |  |
| metrics.kafka.image.pullSecrets | list | `[]` |  |
| metrics.kafka.image.registry | string | `"docker.io"` |  |
| metrics.kafka.image.repository | string | `"bitnami/kafka-exporter"` |  |
| metrics.kafka.image.tag | string | `"1.2.0-debian-10-r277"` |  |
| metrics.kafka.resources.limits | object | `{}` |  |
| metrics.kafka.resources.requests | object | `{}` |  |
| metrics.kafka.service.annotations."prometheus.io/path" | string | `"/metrics"` |  |
| metrics.kafka.service.annotations."prometheus.io/port" | string | `"{{ .Values.metrics.kafka.service.port }}"` |  |
| metrics.kafka.service.annotations."prometheus.io/scrape" | string | `"true"` |  |
| metrics.kafka.service.loadBalancerSourceRanges | list | `[]` |  |
| metrics.kafka.service.nodePort | string | `""` |  |
| metrics.kafka.service.port | int | `9308` |  |
| metrics.kafka.service.type | string | `"ClusterIP"` |  |
| metrics.serviceMonitor.enabled | bool | `false` |  |
| nodeAffinityPreset.key | string | `""` |  |
| nodeAffinityPreset.type | string | `""` |  |
| nodeAffinityPreset.values | list | `[]` |  |
| nodeSelector | object | `{}` |  |
| numIoThreads | int | `8` |  |
| numNetworkThreads | int | `3` |  |
| numPartitions | int | `1` |  |
| numRecoveryThreadsPerDataDir | int | `1` |  |
| offsetsTopicReplicationFactor | int | `1` |  |
| pdb.create | bool | `false` |  |
| pdb.maxUnavailable | int | `1` |  |
| persistence.accessModes[0] | string | `"ReadWriteOnce"` |  |
| persistence.annotations | object | `{}` |  |
| persistence.enabled | bool | `true` |  |
| persistence.mountPath | string | `"/bitnami/kafka"` |  |
| persistence.size | string | `"8Gi"` |  |
| podAffinityPreset | string | `""` |  |
| podAnnotations | object | `{}` |  |
| podAntiAffinityPreset | string | `"soft"` |  |
| podLabels | object | `{}` |  |
| podSecurityContext.enabled | bool | `true` |  |
| podSecurityContext.fsGroup | int | `1001` |  |
| podSecurityContext.runAsUser | int | `1001` |  |
| priorityClassName | string | `""` |  |
| rbac.create | bool | `false` |  |
| readinessProbe.enabled | bool | `true` |  |
| readinessProbe.failureThreshold | int | `6` |  |
| readinessProbe.initialDelaySeconds | int | `5` |  |
| readinessProbe.timeoutSeconds | int | `5` |  |
| replicaCount | int | `1` |  |
| requireInterBrokerProtocolVersion.enabled | bool | `false` |  |
| requireInterBrokerProtocolVersion.resources.limits | object | `{}` |  |
| requireInterBrokerProtocolVersion.resources.requests | object | `{}` |  |
| resources.limits | object | `{}` |  |
| resources.requests | object | `{}` |  |
| service.annotations | object | `{}` |  |
| service.externalPort | int | `9094` |  |
| service.internalPort | int | `9093` |  |
| service.loadBalancerSourceRanges | list | `[]` |  |
| service.nodePorts.client | string | `""` |  |
| service.nodePorts.external | string | `""` |  |
| service.port | int | `9092` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.create | bool | `true` |  |
| sidecars | object | `{}` |  |
| socketReceiveBufferBytes | int | `102400` |  |
| socketRequestMaxBytes | string | `"_104857600"` |  |
| socketSendBufferBytes | int | `102400` |  |
| tolerations | list | `[]` |  |
| transactionStateLogMinIsr | int | `1` |  |
| transactionStateLogReplicationFactor | int | `1` |  |
| updateStrategy | string | `"RollingUpdate"` |  |
| volumePermissions.enabled | bool | `true` |  |
| volumePermissions.image.pullPolicy | string | `"Always"` |  |
| volumePermissions.image.pullSecrets | list | `[]` |  |
| volumePermissions.image.registry | string | `"docker.io"` |  |
| volumePermissions.image.repository | string | `"bitnami/minideb"` |  |
| volumePermissions.image.tag | string | `"buster"` |  |
| volumePermissions.resources.limits | object | `{}` |  |
| volumePermissions.resources.requests | object | `{}` |  |
| volumePermissions.securityContext.runAsUser | int | `0` |  |
| zookeeper.auth.enabled | bool | `false` |  |
| zookeeper.enabled | bool | `true` |  |
| zookeeperConnectionTimeoutMs | int | `6000` |  |
