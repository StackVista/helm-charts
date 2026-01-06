# elasticsearch

![Version: 8.5.1](https://img.shields.io/badge/Version-8.5.1-informational?style=flat-square) ![AppVersion: 8.5.1](https://img.shields.io/badge/AppVersion-8.5.1-informational?style=flat-square)

Official Elastic helm chart for Elasticsearch

**Homepage:** <https://github.com/elastic/helm-charts>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Elastic | <helm-charts@elastic.co> |  |

## Source Code

* <https://github.com/elastic/elasticsearch>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| antiAffinity | string | `"hard"` |  |
| antiAffinityTopologyKey | string | `"kubernetes.io/hostname"` |  |
| clusterHealthCheckParams | string | `"wait_for_status=green&timeout=1s"` |  |
| clusterName | string | `"elasticsearch"` |  |
| createCert | bool | `true` |  |
| enableServiceLinks | bool | `true` |  |
| envFrom | list | `[]` |  |
| esConfig | object | `{}` |  |
| esJavaOpts | string | `""` |  |
| esJvmOptions | object | `{}` |  |
| esMajorVersion | string | `""` |  |
| extraContainers | list | `[]` |  |
| extraEnvs | list | `[]` |  |
| extraInitContainers | list | `[]` |  |
| extraVolumeMounts | list | `[]` |  |
| extraVolumes | list | `[]` |  |
| fullnameOverride | string | `""` |  |
| healthNameOverride | string | `""` |  |
| hostAliases | list | `[]` |  |
| httpPort | int | `9200` |  |
| image | string | `"docker.elastic.co/elasticsearch/elasticsearch"` |  |
| imagePullPolicy | string | `"IfNotPresent"` |  |
| imagePullSecrets | list | `[]` |  |
| imageTag | string | `"8.5.1"` |  |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `"nginx"` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.pathtype | string | `"ImplementationSpecific"` |  |
| ingress.tls | list | `[]` |  |
| initResources | object | `{}` |  |
| keystore | list | `[]` |  |
| labels | object | `{}` |  |
| lifecycle | object | `{}` |  |
| masterService | string | `""` |  |
| maxUnavailable | int | `1` |  |
| minimumMasterNodes | int | `2` |  |
| nameOverride | string | `""` |  |
| networkHost | string | `"0.0.0.0"` |  |
| networkPolicy.http.enabled | bool | `false` |  |
| networkPolicy.transport.enabled | bool | `false` |  |
| nodeAffinity | object | `{}` |  |
| nodeGroup | string | `"master"` |  |
| nodeSelector | object | `{}` |  |
| persistence.annotations | object | `{}` |  |
| persistence.enabled | bool | `true` |  |
| persistence.labels.enabled | bool | `false` |  |
| podAnnotations | object | `{}` |  |
| podManagementPolicy | string | `"Parallel"` |  |
| podSecurityContext.fsGroup | int | `1000` |  |
| podSecurityContext.runAsUser | int | `1000` |  |
| podSecurityPolicy.create | bool | `false` |  |
| podSecurityPolicy.name | string | `""` |  |
| podSecurityPolicy.spec.fsGroup.rule | string | `"RunAsAny"` |  |
| podSecurityPolicy.spec.privileged | bool | `true` |  |
| podSecurityPolicy.spec.runAsUser.rule | string | `"RunAsAny"` |  |
| podSecurityPolicy.spec.seLinux.rule | string | `"RunAsAny"` |  |
| podSecurityPolicy.spec.supplementalGroups.rule | string | `"RunAsAny"` |  |
| podSecurityPolicy.spec.volumes[0] | string | `"secret"` |  |
| podSecurityPolicy.spec.volumes[1] | string | `"configMap"` |  |
| podSecurityPolicy.spec.volumes[2] | string | `"persistentVolumeClaim"` |  |
| podSecurityPolicy.spec.volumes[3] | string | `"emptyDir"` |  |
| priorityClassName | string | `""` |  |
| protocol | string | `"https"` |  |
| rbac.automountToken | bool | `true` |  |
| rbac.create | bool | `false` |  |
| rbac.serviceAccountAnnotations | object | `{}` |  |
| rbac.serviceAccountName | string | `""` |  |
| readinessProbe.failureThreshold | int | `3` |  |
| readinessProbe.initialDelaySeconds | int | `10` |  |
| readinessProbe.periodSeconds | int | `10` |  |
| readinessProbe.successThreshold | int | `3` |  |
| readinessProbe.timeoutSeconds | int | `5` |  |
| replicas | int | `3` |  |
| resources.limits.cpu | string | `"1000m"` |  |
| resources.limits.memory | string | `"2Gi"` |  |
| resources.requests.cpu | string | `"1000m"` |  |
| resources.requests.memory | string | `"2Gi"` |  |
| roles[0] | string | `"master"` |  |
| roles[1] | string | `"data"` |  |
| roles[2] | string | `"data_content"` |  |
| roles[3] | string | `"data_hot"` |  |
| roles[4] | string | `"data_warm"` |  |
| roles[5] | string | `"data_cold"` |  |
| roles[6] | string | `"ingest"` |  |
| roles[7] | string | `"ml"` |  |
| roles[8] | string | `"remote_cluster_client"` |  |
| roles[9] | string | `"transform"` |  |
| schedulerName | string | `""` |  |
| secret.enabled | bool | `true` |  |
| secret.password | string | `""` |  |
| secretMounts | list | `[]` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `1000` |  |
| service.annotations | object | `{}` |  |
| service.enabled | bool | `true` |  |
| service.externalTrafficPolicy | string | `""` |  |
| service.httpPortName | string | `"http"` |  |
| service.labels | object | `{}` |  |
| service.labelsHeadless | object | `{}` |  |
| service.loadBalancerIP | string | `""` |  |
| service.loadBalancerSourceRanges | list | `[]` |  |
| service.nodePort | string | `""` |  |
| service.publishNotReadyAddresses | bool | `false` |  |
| service.transportPortName | string | `"transport"` |  |
| service.type | string | `"ClusterIP"` |  |
| sysctlInitContainer.enabled | bool | `true` |  |
| sysctlVmMaxMapCount | int | `262144` |  |
| terminationGracePeriod | int | `120` |  |
| tests.enabled | bool | `true` |  |
| tolerations | list | `[]` |  |
| transportPort | int | `9300` |  |
| updateStrategy | string | `"RollingUpdate"` |  |
| volumeClaimTemplate.accessModes[0] | string | `"ReadWriteOnce"` |  |
| volumeClaimTemplate.resources.requests.storage | string | `"30Gi"` |  |

