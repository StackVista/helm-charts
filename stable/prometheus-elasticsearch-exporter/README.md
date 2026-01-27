# prometheus-elasticsearch-exporter

![Version: 5.8.0-suse-observability.3](https://img.shields.io/badge/Version-5.8.0--suse--observability.3-informational?style=flat-square) ![AppVersion: v1.7.0](https://img.shields.io/badge/AppVersion-v1.7.0-informational?style=flat-square)

Elasticsearch stats exporter for Prometheus

**Homepage:** <https://github.com/prometheus-community/elasticsearch_exporter>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| svenmueller | <sven.mueller@commercetools.com> |  |
| desaintmartin | <cedric@desaintmartin.fr> |  |
| zeritti | <rootsandtrees@posteo.de> |  |

## Source Code

* <https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus-elasticsearch-exporter>

## Requirements

Kubernetes: `>=1.10.0-0`

| Repository | Name | Version |
|------------|------|---------|
| file://../suse-observability-sizing | suse-observability-sizing | 0.1.2 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| deployment.annotations | object | `{}` |  |
| deployment.labels | object | `{}` |  |
| deployment.metricsPort.name | string | `"http"` |  |
| dnsConfig | object | `{}` |  |
| env | object | `{}` |  |
| envFromSecret | string | `""` |  |
| es.aliases | bool | `false` |  |
| es.all | bool | `true` |  |
| es.cluster_settings | bool | `false` |  |
| es.data_stream | bool | `false` |  |
| es.ilm | bool | `false` |  |
| es.indices | bool | `true` |  |
| es.indices_mappings | bool | `true` |  |
| es.indices_settings | bool | `true` |  |
| es.shards | bool | `true` |  |
| es.slm | bool | `false` |  |
| es.snapshots | bool | `true` |  |
| es.ssl.ca.path | string | `"/ssl/ca.pem"` |  |
| es.ssl.client.enabled | bool | `true` |  |
| es.ssl.client.keyPath | string | `"/ssl/client.key"` |  |
| es.ssl.client.pemPath | string | `"/ssl/client.pem"` |  |
| es.ssl.enabled | bool | `false` |  |
| es.ssl.useExistingSecrets | bool | `false` |  |
| es.sslSkipVerify | bool | `false` |  |
| es.timeout | string | `"30s"` |  |
| es.uri | string | `"http://localhost:9200"` |  |
| extraArgs | list | `[]` |  |
| extraEnvSecrets | object | `{}` |  |
| extraVolumeMounts | list | `[]` |  |
| extraVolumes | list | `[]` |  |
| global.commonLabels | object | `{}` |  |
| global.imagePullSecrets | list | `[]` |  |
| global.imageRegistry | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.pullSecret | string | `""` |  |
| image.registry | string | `""` |  |
| image.repository | string | `"quay.io/prometheuscommunity/elasticsearch-exporter"` |  |
| image.tag | string | `""` |  |
| initContainers | list | `[]` |  |
| log.format | string | `"logfmt"` |  |
| log.level | string | `"info"` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| podMonitor.apiVersion | string | `"monitoring.coreos.com/v1"` |  |
| podMonitor.enabled | bool | `false` |  |
| podMonitor.honorLabels | bool | `true` |  |
| podMonitor.interval | string | `"60s"` |  |
| podMonitor.labels | object | `{}` |  |
| podMonitor.metricRelabelings | list | `[]` |  |
| podMonitor.namespace | string | `""` |  |
| podMonitor.relabelings | list | `[]` |  |
| podMonitor.scheme | string | `"http"` |  |
| podMonitor.scrapeTimeout | string | `"10s"` |  |
| podSecurityContext.runAsNonRoot | bool | `true` |  |
| podSecurityContext.runAsUser | int | `1000` |  |
| podSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| podSecurityPolicies.enabled | bool | `false` |  |
| priorityClassName | string | `""` |  |
| prometheusRule.enabled | bool | `false` |  |
| prometheusRule.labels | object | `{}` |  |
| prometheusRule.rules | list | `[]` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| restartPolicy | string | `"Always"` |  |
| secretMounts | list | `[]` |  |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| service.annotations | object | `{}` |  |
| service.enabled | bool | `true` |  |
| service.httpPort | int | `9108` |  |
| service.labels | object | `{}` |  |
| service.metricsPort.name | string | `"http"` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automountServiceAccountToken | bool | `true` |  |
| serviceAccount.create | bool | `false` |  |
| serviceAccount.name | string | `"default"` |  |
| serviceMonitor.apiVersion | string | `"monitoring.coreos.com/v1"` |  |
| serviceMonitor.enabled | bool | `false` |  |
| serviceMonitor.interval | string | `"10s"` |  |
| serviceMonitor.jobLabel | string | `""` |  |
| serviceMonitor.labels | object | `{}` |  |
| serviceMonitor.metricRelabelings | list | `[]` |  |
| serviceMonitor.relabelings | list | `[]` |  |
| serviceMonitor.sampleLimit | int | `0` |  |
| serviceMonitor.scheme | string | `"http"` |  |
| serviceMonitor.scrapeTimeout | string | `"10s"` |  |
| serviceMonitor.targetLabels | list | `[]` |  |
| tolerations | list | `[]` |  |
| web.path | string | `"/metrics"` |  |

