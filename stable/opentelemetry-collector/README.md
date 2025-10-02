# opentelemetry-collector

![Version: 0.108.0-stackstate.2](https://img.shields.io/badge/Version-0.108.0--stackstate.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.111.0](https://img.shields.io/badge/AppVersion-0.111.0-informational?style=flat-square)

OpenTelemetry Collector Helm chart for Kubernetes

**Homepage:** <https://opentelemetry.io/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Source Code

* <https://github.com/open-telemetry/opentelemetry-collector>
* <https://github.com/open-telemetry/opentelemetry-collector-contrib>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| additionalLabels | object | `{}` |  |
| affinity | object | `{}` |  |
| alternateConfig | object | `{}` |  |
| annotations | object | `{}` |  |
| autoscaling.behavior | object | `{}` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `10` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| clusterRole.annotations | object | `{}` |  |
| clusterRole.clusterRoleBinding.annotations | object | `{}` |  |
| clusterRole.clusterRoleBinding.name | string | `""` |  |
| clusterRole.create | bool | `false` |  |
| clusterRole.name | string | `""` |  |
| clusterRole.rules | list | `[]` |  |
| command.extraArgs | list | `[]` |  |
| command.name | string | `""` |  |
| config.exporters.debug | object | `{}` |  |
| config.extensions.health_check.endpoint | string | `"${env:MY_POD_IP}:13133"` |  |
| config.processors.batch | object | `{}` |  |
| config.processors.memory_limiter.check_interval | string | `"5s"` |  |
| config.processors.memory_limiter.limit_percentage | int | `80` |  |
| config.processors.memory_limiter.spike_limit_percentage | int | `25` |  |
| config.receivers.jaeger.protocols.grpc.endpoint | string | `"${env:MY_POD_IP}:14250"` |  |
| config.receivers.jaeger.protocols.thrift_compact.endpoint | string | `"${env:MY_POD_IP}:6831"` |  |
| config.receivers.jaeger.protocols.thrift_http.endpoint | string | `"${env:MY_POD_IP}:14268"` |  |
| config.receivers.otlp.protocols.grpc.endpoint | string | `"${env:MY_POD_IP}:4317"` |  |
| config.receivers.otlp.protocols.http.endpoint | string | `"${env:MY_POD_IP}:4318"` |  |
| config.receivers.prometheus.config.scrape_configs[0].job_name | string | `"opentelemetry-collector"` |  |
| config.receivers.prometheus.config.scrape_configs[0].scrape_interval | string | `"10s"` |  |
| config.receivers.prometheus.config.scrape_configs[0].static_configs[0].targets[0] | string | `"${env:MY_POD_IP}:8888"` |  |
| config.receivers.zipkin.endpoint | string | `"${env:MY_POD_IP}:9411"` |  |
| config.service.extensions[0] | string | `"health_check"` |  |
| config.service.pipelines.logs.exporters[0] | string | `"debug"` |  |
| config.service.pipelines.logs.processors[0] | string | `"memory_limiter"` |  |
| config.service.pipelines.logs.processors[1] | string | `"batch"` |  |
| config.service.pipelines.logs.receivers[0] | string | `"otlp"` |  |
| config.service.pipelines.metrics.exporters[0] | string | `"debug"` |  |
| config.service.pipelines.metrics.processors[0] | string | `"memory_limiter"` |  |
| config.service.pipelines.metrics.processors[1] | string | `"batch"` |  |
| config.service.pipelines.metrics.receivers[0] | string | `"otlp"` |  |
| config.service.pipelines.metrics.receivers[1] | string | `"prometheus"` |  |
| config.service.pipelines.traces.exporters[0] | string | `"debug"` |  |
| config.service.pipelines.traces.processors[0] | string | `"memory_limiter"` |  |
| config.service.pipelines.traces.processors[1] | string | `"batch"` |  |
| config.service.pipelines.traces.receivers[0] | string | `"otlp"` |  |
| config.service.pipelines.traces.receivers[1] | string | `"jaeger"` |  |
| config.service.pipelines.traces.receivers[2] | string | `"zipkin"` |  |
| config.service.telemetry.metrics.address | string | `"${env:MY_POD_IP}:8888"` |  |
| configMap.create | bool | `true` |  |
| configMap.existingName | string | `""` |  |
| dnsConfig | object | `{}` |  |
| dnsPolicy | string | `""` |  |
| extraContainers | list | `[]` |  |
| extraEnvs | list | `[]` |  |
| extraEnvsFrom | list | `[]` |  |
| extraVolumeMounts | list | `[]` |  |
| extraVolumes | list | `[]` |  |
| fullnameOverride | string | `""` |  |
| global.commonLabels | object | `{}` |  |
| global.features | object | `{"enableStackPacks2":false}` | Feature switches for SUSE Observability. |
| global.features.enableStackPacks2 | bool | `false` | Enable StackPacks 2.0 to signal to all components that they should support the StackPacks 2.0 spec. |
| hostAliases | list | `[]` |  |
| hostNetwork | bool | `false` |  |
| image.digest | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `""` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.additionalIngresses | list | `[]` |  |
| ingress.enabled | bool | `false` |  |
| initContainers | list | `[]` |  |
| lifecycleHooks | object | `{}` |  |
| livenessProbe.httpGet.path | string | `"/"` |  |
| livenessProbe.httpGet.port | int | `13133` |  |
| mode | string | `""` |  |
| nameOverride | string | `""` |  |
| namespaceOverride | string | `""` |  |
| networkPolicy.allowIngressFrom | list | `[]` |  |
| networkPolicy.annotations | object | `{}` |  |
| networkPolicy.egressRules | list | `[]` |  |
| networkPolicy.enabled | bool | `false` |  |
| networkPolicy.extraIngressRules | list | `[]` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podDisruptionBudget.enabled | bool | `false` |  |
| podLabels | object | `{}` |  |
| podMonitor.enabled | bool | `false` |  |
| podMonitor.extraLabels | object | `{}` |  |
| podMonitor.metricsEndpoints[0].port | string | `"metrics"` |  |
| podSecurityContext | object | `{}` |  |
| ports.jaeger-compact.containerPort | int | `6831` |  |
| ports.jaeger-compact.enabled | bool | `true` |  |
| ports.jaeger-compact.hostPort | int | `6831` |  |
| ports.jaeger-compact.protocol | string | `"UDP"` |  |
| ports.jaeger-compact.servicePort | int | `6831` |  |
| ports.jaeger-grpc.containerPort | int | `14250` |  |
| ports.jaeger-grpc.enabled | bool | `true` |  |
| ports.jaeger-grpc.hostPort | int | `14250` |  |
| ports.jaeger-grpc.protocol | string | `"TCP"` |  |
| ports.jaeger-grpc.servicePort | int | `14250` |  |
| ports.jaeger-thrift.containerPort | int | `14268` |  |
| ports.jaeger-thrift.enabled | bool | `true` |  |
| ports.jaeger-thrift.hostPort | int | `14268` |  |
| ports.jaeger-thrift.protocol | string | `"TCP"` |  |
| ports.jaeger-thrift.servicePort | int | `14268` |  |
| ports.metrics.containerPort | int | `8888` |  |
| ports.metrics.enabled | bool | `false` |  |
| ports.metrics.protocol | string | `"TCP"` |  |
| ports.metrics.servicePort | int | `8888` |  |
| ports.otlp-http.containerPort | int | `4318` |  |
| ports.otlp-http.enabled | bool | `true` |  |
| ports.otlp-http.hostPort | int | `4318` |  |
| ports.otlp-http.protocol | string | `"TCP"` |  |
| ports.otlp-http.servicePort | int | `4318` |  |
| ports.otlp.appProtocol | string | `"grpc"` |  |
| ports.otlp.containerPort | int | `4317` |  |
| ports.otlp.enabled | bool | `true` |  |
| ports.otlp.hostPort | int | `4317` |  |
| ports.otlp.protocol | string | `"TCP"` |  |
| ports.otlp.servicePort | int | `4317` |  |
| ports.zipkin.containerPort | int | `9411` |  |
| ports.zipkin.enabled | bool | `true` |  |
| ports.zipkin.hostPort | int | `9411` |  |
| ports.zipkin.protocol | string | `"TCP"` |  |
| ports.zipkin.servicePort | int | `9411` |  |
| presets.clusterMetrics.enabled | bool | `false` |  |
| presets.hostMetrics.enabled | bool | `false` |  |
| presets.kubeletMetrics.enabled | bool | `false` |  |
| presets.kubernetesAttributes.enabled | bool | `false` |  |
| presets.kubernetesAttributes.extractAllPodAnnotations | bool | `false` |  |
| presets.kubernetesAttributes.extractAllPodLabels | bool | `false` |  |
| presets.kubernetesEvents.enabled | bool | `false` |  |
| presets.logsCollection.enabled | bool | `false` |  |
| presets.logsCollection.includeCollectorLogs | bool | `false` |  |
| presets.logsCollection.maxRecombineLogSize | int | `102400` |  |
| presets.logsCollection.storeCheckpoints | bool | `false` |  |
| priorityClassName | string | `""` |  |
| prometheusRule.defaultRules.enabled | bool | `false` |  |
| prometheusRule.enabled | bool | `false` |  |
| prometheusRule.extraLabels | object | `{}` |  |
| prometheusRule.groups | list | `[]` |  |
| readinessProbe.httpGet.path | string | `"/"` |  |
| readinessProbe.httpGet.port | int | `13133` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| revisionHistoryLimit | int | `10` |  |
| rollout.rollingUpdate | object | `{}` |  |
| rollout.strategy | string | `"RollingUpdate"` |  |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| service.annotations | object | `{}` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| serviceMonitor.enabled | bool | `false` |  |
| serviceMonitor.extraLabels | object | `{}` |  |
| serviceMonitor.metricRelabelings | list | `[]` |  |
| serviceMonitor.metricsEndpoints[0].port | string | `"metrics"` |  |
| serviceMonitor.relabelings | list | `[]` |  |
| shareProcessNamespace | bool | `false` |  |
| startupProbe | object | `{}` |  |
| statefulset.persistentVolumeClaimRetentionPolicy.enabled | bool | `false` |  |
| statefulset.persistentVolumeClaimRetentionPolicy.whenDeleted | string | `"Retain"` |  |
| statefulset.persistentVolumeClaimRetentionPolicy.whenScaled | string | `"Retain"` |  |
| statefulset.podManagementPolicy | string | `"Parallel"` |  |
| statefulset.volumeClaimTemplates | list | `[]` |  |
| tolerations | list | `[]` |  |
| topologySpreadConstraints | list | `[]` |  |
| useGOMEMLIMIT | bool | `true` |  |

