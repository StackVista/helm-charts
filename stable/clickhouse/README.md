# clickhouse

![Version: 3.6.9-suse-observability.5](https://img.shields.io/badge/Version-3.6.9--suse--observability.5-informational?style=flat-square) ![AppVersion: 23.7.4](https://img.shields.io/badge/AppVersion-23.7.4-informational?style=flat-square)

ClickHouse is an open-source column-oriented OLAP database management system. Use it to boost your database performance while providing linear scalability and hardware efficiency.

**Homepage:** <https://bitnami.com>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| VMware, Inc. |  | <https://github.com/bitnami/charts> |

## Source Code

* <https://github.com/bitnami/charts/tree/main/bitnami/clickhouse>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../suse-observability-sizing | suse-observability-sizing | 0.1.1 |
| file://charts/common | common | 2.x.x |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| args | list | `[]` |  |
| auth.existingSecret | string | `""` |  |
| auth.existingSecretKey | string | `""` |  |
| auth.password | string | `""` |  |
| auth.username | string | `"default"` |  |
| clusterDomain | string | `"cluster.local"` |  |
| command[0] | string | `"/scripts/setup.sh"` |  |
| commonAnnotations | object | `{}` |  |
| commonLabels | object | `{}` |  |
| containerPorts.http | int | `8123` |  |
| containerPorts.https | int | `8443` |  |
| containerPorts.interserver | int | `9009` |  |
| containerPorts.keeper | int | `2181` |  |
| containerPorts.keeperInter | int | `9444` |  |
| containerPorts.keeperSecure | int | `3181` |  |
| containerPorts.metrics | int | `8001` |  |
| containerPorts.mysql | int | `9004` |  |
| containerPorts.postgresql | int | `9005` |  |
| containerPorts.tcp | int | `9000` |  |
| containerPorts.tcpSecure | int | `9440` |  |
| containerSecurityContext.allowPrivilegeEscalation | bool | `false` |  |
| containerSecurityContext.capabilities.drop[0] | string | `"ALL"` |  |
| containerSecurityContext.enabled | bool | `true` |  |
| containerSecurityContext.runAsNonRoot | bool | `true` |  |
| containerSecurityContext.runAsUser | int | `1001` |  |
| customLivenessProbe | object | `{}` |  |
| customReadinessProbe | object | `{}` |  |
| customStartupProbe | object | `{}` |  |
| defaultConfigurationOverrides | string | `"<clickhouse>\n  <!-- Macros -->\n  <macros>\n    <shard from_env=\"CLICKHOUSE_SHARD_ID\"></shard>\n    <replica from_env=\"CLICKHOUSE_REPLICA_ID\"></replica>\n    <layer>{{ include \"common.names.fullname\" . }}</layer>\n  </macros>\n  <!-- Log Level -->\n  <logger>\n    <level>{{ .Values.logLevel }}</level>\n  </logger>\n  {{- $sizingReplicaCount := include \"common.sizing.clickhouse.replicaCount\" . | trim -}}\n  {{- $chartDefaultReplicas := \"1\" -}}{{/* Backwards-compatibility default matching pre-global-mode parent chart */}}\n  {{- $effectiveReplicaCount := ternary .Values.replicaCount ($sizingReplicaCount | default $chartDefaultReplicas) (not (kindIs \"invalid\" .Values.replicaCount)) | int -}}\n  {{- if or (ne (int .Values.shards) 1) (ne $effectiveReplicaCount 1)}}\n  <!-- Cluster configuration - Any update of the shards and replicas requires helm upgrade -->\n  <remote_servers>\n    <default>\n      {{- $shards := $.Values.shards | int }}\n      {{- range $shard, $e := until $shards }}\n      <shard>\n          {{- range $i, $_e := until $effectiveReplicaCount }}\n          <replica>\n              <host>{{ printf \"%s-shard%d-%d.%s.%s.svc.%s\" (include \"common.names.fullname\" $ ) $shard $i (include \"clickhouse.headlessServiceName\" $) (include \"common.names.namespace\" $) $.Values.clusterDomain }}</host>\n              <port>{{ $.Values.service.ports.tcp }}</port>\n              <user from_env=\"CLICKHOUSE_ADMIN_USER\"></user>\n              <password from_env=\"CLICKHOUSE_ADMIN_PASSWORD\"></password>\n          </replica>\n          {{- end }}\n      </shard>\n      {{- end }}\n    </default>\n  </remote_servers>\n  {{- end }}\n  {{- if .Values.keeper.enabled }}\n  <!-- keeper configuration -->\n  <keeper_server>\n    {{/*ClickHouse keeper configuration using the helm chart */}}\n    <tcp_port>{{ $.Values.containerPorts.keeper }}</tcp_port>\n    {{- if .Values.tls.enabled }}\n    <tcp_port_secure>{{ $.Values.containerPorts.keeperSecure }}</tcp_port_secure>\n    {{- end }}\n    <server_id from_env=\"KEEPER_SERVER_ID\"></server_id>\n    <log_storage_path>/bitnami/clickhouse/keeper/coordination/log</log_storage_path>\n    <snapshot_storage_path>/bitnami/clickhouse/keeper/coordination/snapshots</snapshot_storage_path>\n\n    <coordination_settings>\n        <operation_timeout_ms>10000</operation_timeout_ms>\n        <session_timeout_ms>30000</session_timeout_ms>\n        <raft_logs_level>trace</raft_logs_level>\n    </coordination_settings>\n\n    <raft_configuration>\n    {{- range $node, $e := until $effectiveReplicaCount }}\n    <server>\n      <id>{{ $node | int }}</id>\n      <hostname from_env=\"{{ printf \"KEEPER_NODE_%d\" $node }}\"></hostname>\n      <port>{{ $.Values.service.ports.keeperInter }}</port>\n    </server>\n    {{- end }}\n    </raft_configuration>\n  </keeper_server>\n  {{- end }}\n  {{- if or .Values.keeper.enabled .Values.externalZookeeper.servers }}\n  <!-- Zookeeper configuration -->\n  <zookeeper>\n    {{- if .Values.keeper.enabled }}\n    {{- range $node, $e := until $effectiveReplicaCount }}\n    <node>\n      <host from_env=\"{{ printf \"KEEPER_NODE_%d\" $node }}\"></host>\n      <port>{{ $.Values.service.ports.keeper }}</port>\n    </node>\n    {{- end }}\n    {{- else if .Values.externalZookeeper.servers }}\n    {{/* Zookeeper configuration using an external instance */}}\n    {{- range $node :=.Values.externalZookeeper.servers }}\n    <node>\n      <host>{{ $node }}</host>\n      <port>{{ $.Values.externalZookeeper.port }}</port>\n    </node>\n    {{- end }}\n    {{- end }}\n  </zookeeper>\n  {{- end }}\n  {{- if .Values.tls.enabled }}\n  <!-- TLS configuration -->\n  <tcp_port_secure from_env=\"CLICKHOUSE_TCP_SECURE_PORT\"></tcp_port_secure>\n  <https_port from_env=\"CLICKHOUSE_HTTPS_PORT\"></https_port>\n  <openSSL>\n      <server>\n          {{- $certFileName := default \"tls.crt\" .Values.tls.certFilename }}\n          {{- $keyFileName := default \"tls.key\" .Values.tls.certKeyFilename }}\n          <certificateFile>/bitnami/clickhouse/certs/{{$certFileName}}</certificateFile>\n          <privateKeyFile>/bitnami/clickhouse/certs/{{$keyFileName}}</privateKeyFile>\n          <verificationMode>none</verificationMode>\n          <cacheSessions>true</cacheSessions>\n          <disableProtocols>sslv2,sslv3</disableProtocols>\n          <preferServerCiphers>true</preferServerCiphers>\n          {{- if or .Values.tls.autoGenerated .Values.tls.certCAFilename }}\n          {{- $caFileName := default \"ca.crt\" .Values.tls.certCAFilename }}\n          <caConfig>/bitnami/clickhouse/certs/{{$caFileName}}</caConfig>\n          {{- else }}\n          <loadDefaultCAFile>true</loadDefaultCAFile>\n          {{- end }}\n      </server>\n      <client>\n          <loadDefaultCAFile>true</loadDefaultCAFile>\n          <cacheSessions>true</cacheSessions>\n          <disableProtocols>sslv2,sslv3</disableProtocols>\n          <preferServerCiphers>true</preferServerCiphers>\n          <verificationMode>none</verificationMode>\n          <invalidCertificateHandler>\n              <name>AcceptCertificateHandler</name>\n          </invalidCertificateHandler>\n      </client>\n  </openSSL>\n  {{- end }}\n  {{- if .Values.metrics.enabled }}\n   <!-- Prometheus metrics -->\n   <prometheus>\n      <endpoint>/metrics</endpoint>\n      <port from_env=\"CLICKHOUSE_METRICS_PORT\"></port>\n      <metrics>true</metrics>\n      <events>true</events>\n      <asynchronous_metrics>true</asynchronous_metrics>\n  </prometheus>\n  {{- end }}\n</clickhouse>\n"` |  |
| diagnosticMode.args[0] | string | `"infinity"` |  |
| diagnosticMode.command[0] | string | `"sleep"` |  |
| diagnosticMode.enabled | bool | `false` |  |
| existingOverridesConfigmap | string | `""` |  |
| externalAccess.enabled | bool | `false` |  |
| externalAccess.service.annotations | object | `{}` |  |
| externalAccess.service.extraPorts | list | `[]` |  |
| externalAccess.service.labels | object | `{}` |  |
| externalAccess.service.loadBalancerAnnotations | list | `[]` |  |
| externalAccess.service.loadBalancerIPs | list | `[]` |  |
| externalAccess.service.loadBalancerSourceRanges | list | `[]` |  |
| externalAccess.service.nodePorts.http | list | `[]` |  |
| externalAccess.service.nodePorts.https | list | `[]` |  |
| externalAccess.service.nodePorts.interserver | list | `[]` |  |
| externalAccess.service.nodePorts.keeper | list | `[]` |  |
| externalAccess.service.nodePorts.keeperInter | list | `[]` |  |
| externalAccess.service.nodePorts.keeperSecure | list | `[]` |  |
| externalAccess.service.nodePorts.metrics | list | `[]` |  |
| externalAccess.service.nodePorts.mysql | list | `[]` |  |
| externalAccess.service.nodePorts.postgresql | list | `[]` |  |
| externalAccess.service.nodePorts.tcp | list | `[]` |  |
| externalAccess.service.nodePorts.tcpSecure | list | `[]` |  |
| externalAccess.service.ports.http | int | `80` |  |
| externalAccess.service.ports.https | int | `443` |  |
| externalAccess.service.ports.interserver | int | `9009` |  |
| externalAccess.service.ports.keeper | int | `2181` |  |
| externalAccess.service.ports.keeperInter | int | `9444` |  |
| externalAccess.service.ports.keeperSecure | int | `3181` |  |
| externalAccess.service.ports.metrics | int | `8001` |  |
| externalAccess.service.ports.mysql | int | `9004` |  |
| externalAccess.service.ports.postgresql | int | `9005` |  |
| externalAccess.service.ports.tcp | int | `9000` |  |
| externalAccess.service.ports.tcpSecure | int | `9440` |  |
| externalAccess.service.type | string | `"LoadBalancer"` |  |
| externalZookeeper.port | int | `2888` |  |
| externalZookeeper.servers | list | `[]` |  |
| extraDeploy | list | `[]` |  |
| extraEnvVars | list | `[]` |  |
| extraEnvVarsCM | string | `""` |  |
| extraEnvVarsSecret | string | `""` |  |
| extraOverrides | string | `""` |  |
| extraOverridesConfigmap | string | `""` |  |
| extraOverridesSecret | string | `""` |  |
| extraVolumeMounts | list | `[]` |  |
| extraVolumes | list | `[]` |  |
| fullnameOverride | string | `""` |  |
| global.imagePullSecrets | list | `[]` |  |
| global.imageRegistry | string | `""` |  |
| global.storageClass | string | `""` |  |
| hostAliases | list | `[]` |  |
| image.debug | bool | `false` |  |
| image.digest | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.pullSecrets | list | `[]` |  |
| image.registry | string | `"docker.io"` |  |
| image.repository | string | `"bitnami/clickhouse"` |  |
| image.tag | string | `"23.7.4-debian-11-r5"` |  |
| ingress.annotations | object | `{}` |  |
| ingress.apiVersion | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.extraHosts | list | `[]` |  |
| ingress.extraPaths | list | `[]` |  |
| ingress.extraRules | list | `[]` |  |
| ingress.extraTls | list | `[]` |  |
| ingress.hostname | string | `"clickhouse.local"` |  |
| ingress.ingressClassName | string | `""` |  |
| ingress.path | string | `"/"` |  |
| ingress.pathType | string | `"ImplementationSpecific"` |  |
| ingress.secrets | list | `[]` |  |
| ingress.selfSigned | bool | `false` |  |
| ingress.tls | bool | `false` |  |
| initContainers | list | `[]` |  |
| initdbScripts | object | `{}` |  |
| initdbScriptsSecret | string | `""` |  |
| keeper.enabled | bool | `false` |  |
| kubeVersion | string | `""` |  |
| lifecycleHooks | object | `{}` |  |
| livenessProbe.enabled | bool | `true` |  |
| livenessProbe.failureThreshold | int | `3` |  |
| livenessProbe.initialDelaySeconds | int | `10` |  |
| livenessProbe.periodSeconds | int | `10` |  |
| livenessProbe.successThreshold | int | `1` |  |
| livenessProbe.timeoutSeconds | int | `1` |  |
| logLevel | string | `"information"` |  |
| metrics.enabled | bool | `false` |  |
| metrics.podAnnotations."prometheus.io/port" | string | `"{{ .Values.containerPorts.metrics }}"` |  |
| metrics.podAnnotations."prometheus.io/scrape" | string | `"true"` |  |
| metrics.prometheusRule.additionalLabels | object | `{}` |  |
| metrics.prometheusRule.enabled | bool | `false` |  |
| metrics.prometheusRule.namespace | string | `""` |  |
| metrics.prometheusRule.rules | list | `[]` |  |
| metrics.serviceMonitor.annotations | object | `{}` |  |
| metrics.serviceMonitor.enabled | bool | `false` |  |
| metrics.serviceMonitor.honorLabels | bool | `false` |  |
| metrics.serviceMonitor.interval | string | `""` |  |
| metrics.serviceMonitor.jobLabel | string | `""` |  |
| metrics.serviceMonitor.labels | object | `{}` |  |
| metrics.serviceMonitor.metricRelabelings | list | `[]` |  |
| metrics.serviceMonitor.namespace | string | `""` |  |
| metrics.serviceMonitor.relabelings | list | `[]` |  |
| metrics.serviceMonitor.scrapeTimeout | string | `""` |  |
| metrics.serviceMonitor.selector | object | `{}` |  |
| nameOverride | string | `""` |  |
| namespaceOverride | string | `""` |  |
| nodeAffinityPreset.key | string | `""` |  |
| nodeAffinityPreset.type | string | `""` |  |
| nodeAffinityPreset.values | list | `[]` |  |
| nodeSelector | object | `{}` |  |
| persistence.accessModes[0] | string | `"ReadWriteOnce"` |  |
| persistence.annotations | object | `{}` |  |
| persistence.dataSource | object | `{}` |  |
| persistence.enabled | bool | `true` |  |
| persistence.existingClaim | string | `""` |  |
| persistence.labels | object | `{}` |  |
| persistence.selector | object | `{}` |  |
| persistence.size | string | `"8Gi"` |  |
| persistence.storageClass | string | `""` |  |
| podAffinityPreset | string | `""` |  |
| podAnnotations | object | `{}` |  |
| podAntiAffinityPreset | string | `"soft"` |  |
| podLabels | object | `{}` |  |
| podManagementPolicy | string | `"Parallel"` |  |
| podSecurityContext.enabled | bool | `true` |  |
| podSecurityContext.fsGroup | int | `1001` |  |
| podSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| priorityClassName | string | `""` |  |
| readinessProbe.enabled | bool | `true` |  |
| readinessProbe.failureThreshold | int | `3` |  |
| readinessProbe.initialDelaySeconds | int | `10` |  |
| readinessProbe.periodSeconds | int | `10` |  |
| readinessProbe.successThreshold | int | `1` |  |
| readinessProbe.timeoutSeconds | int | `1` |  |
| replicaCount | string | `nil` |  |
| resources.limits | object | `{}` |  |
| resources.requests | object | `{}` |  |
| schedulerName | string | `""` |  |
| service.annotations | object | `{}` |  |
| service.clusterIP | string | `""` |  |
| service.externalTrafficPolicy | string | `"Cluster"` |  |
| service.extraPorts | list | `[]` |  |
| service.headless.annotations | object | `{}` |  |
| service.loadBalancerIP | string | `""` |  |
| service.loadBalancerSourceRanges | list | `[]` |  |
| service.nodePorts.http | string | `""` |  |
| service.nodePorts.https | string | `""` |  |
| service.nodePorts.interserver | string | `""` |  |
| service.nodePorts.keeper | string | `""` |  |
| service.nodePorts.keeperInter | string | `""` |  |
| service.nodePorts.keeperSecure | string | `""` |  |
| service.nodePorts.metrics | string | `""` |  |
| service.nodePorts.mysql | string | `""` |  |
| service.nodePorts.postgresql | string | `""` |  |
| service.nodePorts.tcp | string | `""` |  |
| service.nodePorts.tcpSecure | string | `""` |  |
| service.ports.http | int | `8123` |  |
| service.ports.https | int | `443` |  |
| service.ports.interserver | int | `9009` |  |
| service.ports.keeper | int | `2181` |  |
| service.ports.keeperInter | int | `9444` |  |
| service.ports.keeperSecure | int | `3181` |  |
| service.ports.metrics | int | `8001` |  |
| service.ports.mysql | int | `9004` |  |
| service.ports.postgresql | int | `9005` |  |
| service.ports.tcp | int | `9000` |  |
| service.ports.tcpSecure | int | `9440` |  |
| service.sessionAffinity | string | `"None"` |  |
| service.sessionAffinityConfig | object | `{}` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automountServiceAccountToken | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| shards | int | `2` |  |
| sidecars | list | `[]` |  |
| startdbScripts | object | `{}` |  |
| startdbScriptsSecret | string | `""` |  |
| startupProbe.enabled | bool | `false` |  |
| startupProbe.failureThreshold | int | `3` |  |
| startupProbe.initialDelaySeconds | int | `10` |  |
| startupProbe.periodSeconds | int | `10` |  |
| startupProbe.successThreshold | int | `1` |  |
| startupProbe.timeoutSeconds | int | `1` |  |
| terminationGracePeriodSeconds | string | `""` |  |
| tls.autoGenerated | bool | `false` |  |
| tls.certCAFilename | string | `""` |  |
| tls.certFilename | string | `""` |  |
| tls.certKeyFilename | string | `""` |  |
| tls.certificatesSecret | string | `""` |  |
| tls.enabled | bool | `false` |  |
| tolerations | list | `[]` |  |
| topologySpreadConstraints | list | `[]` |  |
| updateStrategy.type | string | `"RollingUpdate"` |  |
| usersExtraOverrides | string | `""` |  |
| usersExtraOverridesConfigmap | string | `""` |  |
| usersExtraOverridesSecret | string | `""` |  |
| volumePermissions.containerSecurityContext.runAsUser | int | `0` |  |
| volumePermissions.enabled | bool | `false` |  |
| volumePermissions.image.pullPolicy | string | `"IfNotPresent"` |  |
| volumePermissions.image.pullSecrets | list | `[]` |  |
| volumePermissions.image.registry | string | `"docker.io"` |  |
| volumePermissions.image.repository | string | `"bitnami/os-shell"` |  |
| volumePermissions.image.tag | string | `"11-debian-11-r40"` |  |
| volumePermissions.resources.limits | object | `{}` |  |
| volumePermissions.resources.requests | object | `{}` |  |

