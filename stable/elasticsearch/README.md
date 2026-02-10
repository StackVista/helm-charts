# elasticsearch

![Version: 8.19.4-stackstate.8](https://img.shields.io/badge/Version-8.19.4--stackstate.8-informational?style=flat-square) ![AppVersion: 8.19.4](https://img.shields.io/badge/AppVersion-8.19.4-informational?style=flat-square)
Official Elastic helm chart for Elasticsearch
**Homepage:** <https://github.com/elastic/helm-charts>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| SUSE Observability Ops Team | <suse-observability-ops@suse.com> |  |
## Source Code

* <https://github.com/elastic/elasticsearch>
## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../prometheus-elasticsearch-exporter | prometheus-elasticsearch-exporter | 5.8.0-suse-observability.7 |
| file://../suse-observability-sizing | suse-observability-sizing | 0.1.4 |
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| antiAffinity | string | `"hard"` |  |
| antiAffinityTopologyKey | string | `"kubernetes.io/hostname"` |  |
| clusterHealthCheckParams | string | `"local=true"` |  |
| clusterName | string | `"elasticsearch"` |  |
| commonLabels | object | `{}` |  |
| createCert | bool | `false` |  |
| esConfig | object | `{}` |  |
| esJavaOpts | string | `"-Xmx1g -Xms1g"` |  |
| esMajorVersion | string | `""` |  |
| extraContainers | string | `""` |  |
| extraEnvs | list | `[]` |  |
| extraInitContainers | string | `""` |  |
| extraVolumeMounts | string | `""` |  |
| extraVolumes | string | `""` |  |
| fsGroup | string | `""` |  |
| fullnameOverride | string | `""` |  |
| global.commonLabels | object | `{}` |  |
| httpPort | int | `9200` |  |
| imagePullPolicy | string | `"IfNotPresent"` |  |
| imagePullSecrets | list | `[]` |  |
| imageRegistry | string | `"quay.io"` |  |
| imageRepository | string | `"stackstate/elasticsearch"` |  |
| imageTag | string | `"8.19.4-8fb32031"` |  |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0] | string | `"chart-example.local"` |  |
| ingress.path | string | `"/"` |  |
| ingress.tls | list | `[]` |  |
| initResources | object | `{}` |  |
| keystore | list | `[]` |  |
| labels | object | `{}` |  |
| lifecycle | object | `{}` |  |
| livenessProbe.failureThreshold | int | `6` |  |
| livenessProbe.initialDelaySeconds | int | `60` |  |
| livenessProbe.periodSeconds | int | `10` |  |
| livenessProbe.timeoutSeconds | int | `5` |  |
| masterService | string | `""` |  |
| masterTerminationFix | bool | `false` |  |
| maxUnavailable | int | `1` |  |
| minimumMasterNodes | int | `2` |  |
| nameOverride | string | `""` |  |
| networkHost | string | `"0.0.0.0"` |  |
| nodeAffinity | object | `{}` |  |
| nodeGroup | string | `"master"` |  |
| nodeSelector | object | `{}` |  |
| persistence.annotations | object | `{}` |  |
| persistence.enabled | bool | `true` |  |
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
| priorityClassName | string | `""` |  |
| prometheus-elasticsearch-exporter.enabled | bool | `false` | Enable to expose prometheus metrics |
| prometheus-elasticsearch-exporter.es.uri | string | `"http://elasticsearch-master:9200"` | URI of Elasticsearch to monitor, override when changing clusterName or nodeGroup (format is <protocol>://<clusterName>-<nodegroup>:<httpPort>) |
| prometheus-elasticsearch-exporter.image.registry | string | `"quay.io"` |  |
| prometheus-elasticsearch-exporter.image.repository | string | `"stackstate/elasticsearch-exporter"` | Elastichsearch Prometheus exporter image repository |
| prometheus-elasticsearch-exporter.image.tag | string | `"v1.8.0-d2aa61ab"` | Elastichsearch Prometheus exporter image tag |
| prometheus-elasticsearch-exporter.podAnnotations | object | `{}` | custom annotations on the pod |
| prometheus-elasticsearch-exporter.servicemonitor.enabled | bool | `false` | enable to create a servicemonitor for prometheus operator |
| protocol | string | `"http"` |  |
| rbac.create | bool | `false` |  |
| rbac.serviceAccountName | string | `""` |  |
| readinessProbe.failureThreshold | int | `5` |  |
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
| roles[1] | string | `"ingest"` |  |
| roles[2] | string | `"data"` |  |
| schedulerName | string | `""` |  |
| secret.enabled | bool | `false` |  |
| secret.password | string | `""` |  |
| secretMounts | list | `[]` |  |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.enabled | bool | `true` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `1000` |  |
| securityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| service.annotations | object | `{}` |  |
| service.httpPortName | string | `"http"` |  |
| service.labels | object | `{}` |  |
| service.labelsHeadless | object | `{}` |  |
| service.loadBalancerIP | string | `""` |  |
| service.loadBalancerSourceRanges | list | `[]` |  |
| service.nodePort | string | `""` |  |
| service.transportPortName | string | `"transport"` |  |
| service.type | string | `"ClusterIP"` |  |
| sidecarResources | object | `{}` |  |
| startupProbe.failureThreshold | int | `17` |  |
| startupProbe.initialDelaySeconds | int | `90` |  |
| startupProbe.periodSeconds | int | `30` |  |
| startupProbe.timeoutSeconds | int | `5` |  |
| sysctlInitContainer.enabled | bool | `true` |  |
| sysctlVmMaxMapCount | int | `262144` |  |
| terminationGracePeriod | int | `120` |  |
| tolerations | list | `[]` |  |
| transportPort | int | `9300` |  |
| updateStrategy | string | `"RollingUpdate"` |  |
| volumeClaimTemplate.accessModes[0] | string | `"ReadWriteOnce"` |  |
| volumeClaimTemplate.resources.requests.storage | string | `"30Gi"` |  |
