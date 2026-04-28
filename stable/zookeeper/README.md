# zookeeper

![Version: 8.1.2-suse-observability.25](https://img.shields.io/badge/Version-8.1.2--suse--observability.25-informational?style=flat-square) ![AppVersion: 3.7.0](https://img.shields.io/badge/AppVersion-3.7.0-informational?style=flat-square)
Apache ZooKeeper provides a reliable, centralized register of configuration data and services for distributed applications.
**Homepage:** <https://github.com/bitnami/charts/tree/master/bitnami/zookeeper>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Bitnami | <containers@bitnami.com> |  |
## Source Code

* <https://github.com/bitnami/bitnami-docker-zookeeper>
* <https://zookeeper.apache.org/>
## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../suse-observability-sizing | suse-observability-sizing | 0.1.16 |
| file://charts/common | common | 1.x.x |
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| args | list | `[]` |  |
| auth.clientPassword | string | `""` |  |
| auth.clientUser | string | `""` |  |
| auth.enabled | bool | `false` |  |
| auth.existingSecret | string | `""` |  |
| auth.serverPasswords | string | `""` |  |
| auth.serverUsers | string | `""` |  |
| autopurge.purgeInterval | int | `3` |  |
| autopurge.snapRetainCount | int | `5` |  |
| clusterDomain | string | `"cluster.local"` |  |
| command[0] | string | `"/scripts/setup.sh"` |  |
| commonAnnotations | object | `{}` |  |
| commonLabels."app.kubernetes.io/part-of" | string | `"suse-observability"` |  |
| configuration | string | `""` |  |
| containerPorts.client | int | `2181` |  |
| containerPorts.election | int | `3888` |  |
| containerPorts.follower | int | `2888` |  |
| containerPorts.tls | int | `3181` |  |
| containerSecurityContext.allowPrivilegeEscalation | bool | `false` |  |
| containerSecurityContext.capabilities.drop[0] | string | `"ALL"` |  |
| containerSecurityContext.enabled | bool | `true` |  |
| containerSecurityContext.runAsNonRoot | bool | `true` |  |
| containerSecurityContext.runAsUser | int | `1000` |  |
| containerSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| customLivenessProbe.exec.command[0] | string | `"/bin/bash"` |  |
| customLivenessProbe.exec.command[1] | string | `"-c"` |  |
| customLivenessProbe.exec.command[2] | string | `"exec 3<>/dev/tcp/127.0.0.1/2181; printf ruok >&3; read -r -t 2 -u 3 r; [ \"$r\" = imok ]"` |  |
| customLivenessProbe.failureThreshold | int | `6` |  |
| customLivenessProbe.initialDelaySeconds | int | `30` |  |
| customLivenessProbe.periodSeconds | int | `10` |  |
| customLivenessProbe.successThreshold | int | `1` |  |
| customLivenessProbe.timeoutSeconds | int | `5` |  |
| customReadinessProbe.exec.command[0] | string | `"/bin/bash"` |  |
| customReadinessProbe.exec.command[1] | string | `"-c"` |  |
| customReadinessProbe.exec.command[2] | string | `"exec 3<>/dev/tcp/127.0.0.1/2181; printf ruok >&3; read -r -t 2 -u 3 r; [ \"$r\" = imok ]"` |  |
| customReadinessProbe.failureThreshold | int | `6` |  |
| customReadinessProbe.initialDelaySeconds | int | `5` |  |
| customReadinessProbe.periodSeconds | int | `10` |  |
| customReadinessProbe.successThreshold | int | `1` |  |
| customReadinessProbe.timeoutSeconds | int | `5` |  |
| customStartupProbe | object | `{}` |  |
| dataDir | string | `"/data/zookeeper/data"` |  |
| dataLogDir | string | `""` |  |
| diagnosticMode.args[0] | string | `"infinity"` |  |
| diagnosticMode.command[0] | string | `"sleep"` |  |
| diagnosticMode.enabled | bool | `false` |  |
| existingConfigmap | string | `""` |  |
| extraDeploy | list | `[]` |  |
| extraEnvVars | list | `[]` |  |
| extraEnvVarsCM | string | `""` |  |
| extraEnvVarsSecret | string | `""` |  |
| extraVolumeMounts | list | `[]` |  |
| extraVolumes | list | `[]` |  |
| fourlwCommandsWhitelist | string | `"mntr, ruok, stat, srvr"` |  |
| fullnameOverride | string | `"suse-observability-zookeeper"` |  |
| global.commonLabels | object | `{}` |  |
| global.imagePullSecrets | list | `[]` |  |
| global.imageRegistry | string | `""` |  |
| global.storageClass | string | `""` |  |
| heapSize | int | `400` |  |
| hostAliases | list | `[]` |  |
| image.debug | bool | `false` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.pullSecrets | list | `[]` |  |
| image.registry | string | `"quay.io"` |  |
| image.repository | string | `"stackstate/zookeeper"` |  |
| image.tag | string | `"3.9.5-3.2-release"` |  |
| initContainers | list | `[]` |  |
| initLimit | int | `10` |  |
| jvmFlags | string | `"-Djute.maxbuffer=2097150"` |  |
| kubeVersion | string | `""` |  |
| lifecycleHooks | object | `{}` |  |
| listenOnAllIPs | bool | `false` |  |
| livenessProbe.enabled | bool | `false` |  |
| livenessProbe.failureThreshold | int | `6` |  |
| livenessProbe.initialDelaySeconds | int | `30` |  |
| livenessProbe.periodSeconds | int | `10` |  |
| livenessProbe.probeCommandTimeout | int | `2` |  |
| livenessProbe.successThreshold | int | `1` |  |
| livenessProbe.timeoutSeconds | int | `5` |  |
| logLevel | string | `"ERROR"` |  |
| maxClientCnxns | int | `60` |  |
| maxSessionTimeout | int | `40000` |  |
| metrics.containerPort | int | `9141` |  |
| metrics.enabled | bool | `true` |  |
| metrics.prometheusRule.additionalLabels | object | `{}` |  |
| metrics.prometheusRule.enabled | bool | `false` |  |
| metrics.prometheusRule.namespace | string | `""` |  |
| metrics.prometheusRule.rules | list | `[]` |  |
| metrics.service.annotations."prometheus.io/path" | string | `"/metrics"` |  |
| metrics.service.annotations."prometheus.io/port" | string | `"{{ .Values.metrics.service.port }}"` |  |
| metrics.service.annotations."prometheus.io/scrape" | string | `"true"` |  |
| metrics.service.port | int | `9141` |  |
| metrics.service.type | string | `"ClusterIP"` |  |
| metrics.serviceMonitor.additionalLabels | object | `{}` |  |
| metrics.serviceMonitor.enabled | bool | `false` |  |
| metrics.serviceMonitor.honorLabels | bool | `false` |  |
| metrics.serviceMonitor.interval | string | `""` |  |
| metrics.serviceMonitor.jobLabel | string | `""` |  |
| metrics.serviceMonitor.metricRelabelings | list | `[]` |  |
| metrics.serviceMonitor.namespace | string | `""` |  |
| metrics.serviceMonitor.relabelings | list | `[]` |  |
| metrics.serviceMonitor.scrapeTimeout | string | `""` |  |
| metrics.serviceMonitor.selector | object | `{}` |  |
| minServerId | int | `1` |  |
| mountPath | string | `"/data/zookeeper"` |  |
| nameOverride | string | `""` |  |
| namespaceOverride | string | `""` |  |
| networkPolicy.allowExternal | bool | `true` |  |
| networkPolicy.enabled | bool | `false` |  |
| nodeAffinityPreset.key | string | `""` |  |
| nodeAffinityPreset.type | string | `""` |  |
| nodeAffinityPreset.values | list | `[]` |  |
| nodeSelector | object | `{}` |  |
| pdb.create | bool | `true` |  |
| pdb.maxUnavailable | int | `1` |  |
| pdb.minAvailable | string | `""` |  |
| persistence.accessModes[0] | string | `"ReadWriteOnce"` |  |
| persistence.annotations | object | `{}` |  |
| persistence.dataLogDir.existingClaim | string | `""` |  |
| persistence.dataLogDir.selector | object | `{}` |  |
| persistence.dataLogDir.size | string | `"8Gi"` |  |
| persistence.enabled | bool | `true` |  |
| persistence.existingClaim | string | `""` |  |
| persistence.selector | object | `{}` |  |
| persistence.size | string | `nil` |  |
| persistence.storageClass | string | `""` |  |
| podAffinityPreset | string | `""` |  |
| podAnnotations."ad.stackstate.com/zookeeper.check_names" | string | `"[\"openmetrics\"]"` |  |
| podAnnotations."ad.stackstate.com/zookeeper.init_configs" | string | `"[{}]"` |  |
| podAnnotations."ad.stackstate.com/zookeeper.instances" | string | `"[ { \"prometheus_url\": \"http://%%host%%:9141/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"*\"] } ]"` |  |
| podAntiAffinityPreset | string | `"soft"` |  |
| podLabels."app.kubernetes.io/part-of" | string | `"suse-observability"` |  |
| podManagementPolicy | string | `"Parallel"` |  |
| podSecurityContext.enabled | bool | `true` |  |
| podSecurityContext.fsGroup | int | `1000` |  |
| preAllocSize | int | `65536` |  |
| priorityClassName | string | `""` |  |
| readinessProbe.enabled | bool | `false` |  |
| readinessProbe.failureThreshold | int | `6` |  |
| readinessProbe.initialDelaySeconds | int | `5` |  |
| readinessProbe.periodSeconds | int | `10` |  |
| readinessProbe.probeCommandTimeout | int | `2` |  |
| readinessProbe.successThreshold | int | `1` |  |
| readinessProbe.timeoutSeconds | int | `5` |  |
| replicaCount | string | `nil` |  |
| resources | object | `{}` |  |
| schedulerName | string | `""` |  |
| service.annotations | object | `{}` |  |
| service.clusterIP | string | `""` |  |
| service.disableBaseClientPort | bool | `false` |  |
| service.externalTrafficPolicy | string | `"Cluster"` |  |
| service.extraPorts | list | `[]` |  |
| service.headless.annotations | object | `{}` |  |
| service.headless.publishNotReadyAddresses | bool | `true` |  |
| service.loadBalancerIP | string | `""` |  |
| service.loadBalancerSourceRanges | list | `[]` |  |
| service.nodePorts.client | string | `""` |  |
| service.nodePorts.tls | string | `""` |  |
| service.ports.client | int | `2181` |  |
| service.ports.election | int | `3888` |  |
| service.ports.follower | int | `2888` |  |
| service.ports.tls | int | `3181` |  |
| service.sessionAffinity | string | `"None"` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automountServiceAccountToken | bool | `true` |  |
| serviceAccount.create | bool | `false` |  |
| serviceAccount.name | string | `""` |  |
| sidecars | list | `[]` |  |
| snapCount | int | `100000` |  |
| startupProbe.enabled | bool | `false` |  |
| startupProbe.failureThreshold | int | `15` |  |
| startupProbe.initialDelaySeconds | int | `30` |  |
| startupProbe.periodSeconds | int | `10` |  |
| startupProbe.successThreshold | int | `1` |  |
| startupProbe.timeoutSeconds | int | `1` |  |
| syncLimit | int | `5` |  |
| tickTime | int | `2000` |  |
| tls.client.autoGenerated | bool | `false` |  |
| tls.client.enabled | bool | `false` |  |
| tls.client.existingSecret | string | `""` |  |
| tls.client.keystorePassword | string | `""` |  |
| tls.client.keystorePath | string | `"/opt/bitnami/zookeeper/config/certs/client/zookeeper.keystore.jks"` |  |
| tls.client.passwordsSecretName | string | `""` |  |
| tls.client.truststorePassword | string | `""` |  |
| tls.client.truststorePath | string | `"/opt/bitnami/zookeeper/config/certs/client/zookeeper.truststore.jks"` |  |
| tls.quorum.autoGenerated | bool | `false` |  |
| tls.quorum.enabled | bool | `false` |  |
| tls.quorum.existingSecret | string | `""` |  |
| tls.quorum.keystorePassword | string | `""` |  |
| tls.quorum.keystorePath | string | `"/opt/bitnami/zookeeper/config/certs/quorum/zookeeper.keystore.jks"` |  |
| tls.quorum.passwordsSecretName | string | `""` |  |
| tls.quorum.truststorePassword | string | `""` |  |
| tls.quorum.truststorePath | string | `"/opt/bitnami/zookeeper/config/certs/quorum/zookeeper.truststore.jks"` |  |
| tls.resources.limits | object | `{}` |  |
| tls.resources.requests | object | `{}` |  |
| tolerations | list | `[]` |  |
| topologySpreadConstraints | object | `{}` |  |
| updateStrategy.rollingUpdate | object | `{}` |  |
| updateStrategy.type | string | `"RollingUpdate"` |  |
| volumePermissions.containerSecurityContext.runAsUser | int | `0` |  |
| volumePermissions.enabled | bool | `false` |  |
| volumePermissions.image.pullPolicy | string | `"IfNotPresent"` |  |
| volumePermissions.image.pullSecrets | list | `[]` |  |
| volumePermissions.image.registry | string | `"docker.io"` |  |
| volumePermissions.image.repository | string | `"bitnami/bitnami-shell"` |  |
| volumePermissions.image.tag | string | `"10-debian-10-r367"` |  |
| volumePermissions.resources.limits | object | `{}` |  |
| volumePermissions.resources.requests | object | `{}` |  |
