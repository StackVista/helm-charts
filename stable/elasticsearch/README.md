# elasticsearch

![Version: 7.16.2-stackstate.0](https://img.shields.io/badge/Version-7.16.2--stackstate.26-informational?style=flat-square) ![AppVersion: 7.16.2](https://img.shields.io/badge/AppVersion-7.16.2-informational?style=flat-square)
Official Elastic helm chart for Elasticsearch
**Homepage:** <https://github.com/elastic/helm-charts>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Jeroen van Erp | jvanerp@stackstate.com |  |
| Remco Beckers | rbeckers@stackstate.com |  |
| Vincent Partington | vpartington@stackstate.com |  |
## Source Code

* <https://github.com/elastic/elasticsearch>
## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://prometheus-community.github.io/helm-charts | prometheus-elasticsearch-exporter | 4.4.0 |
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| antiAffinity | string | `"hard"` |  |
| antiAffinityTopologyKey | string | `"kubernetes.io/hostname"` |  |
| clusterName | string | `"elasticsearch"` |  |
| commonLabels | object | `{}` |  |
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
| httpPort | int | `9200` |  |
| imagePullPolicy | string | `"IfNotPresent"` |  |
| imagePullSecrets | list | `[]` |  |
| imageRegistry | string | `"quay.io"` |  |
| imageRepository | string | `"stackstate/elasticsearch"` |  |
| imageTag | string | `"7.6.2-sts.20211214.1906"` |  |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0] | string | `"chart-example.local"` |  |
| ingress.path | string | `"/"` |  |
| ingress.tls | list | `[]` |  |
| initResources | object | `{}` |  |
| keystore | list | `[]` |  |
| labels | object | `{}` |  |
| lifecycle | object | `{}` |  |
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
| prometheus-elasticsearch-exporter.image.repository | string | `"quay.io/stackstate/elasticsearch-exporter"` | Elastichsearch Prometheus exporter image repository |
| prometheus-elasticsearch-exporter.image.tag | string | `"v1.2.1"` | Elastichsearch Prometheus exporter image tag |
| prometheus-elasticsearch-exporter.podAnnotations | object | `{}` | custom annotations on the pod |
| prometheus-elasticsearch-exporter.securityContext.enabled | bool | `true` | Set to `false` for OpenShift compatibility |
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
| roles.data | string | `"true"` |  |
| roles.ingest | string | `"true"` |  |
| roles.master | string | `"true"` |  |
| schedulerName | string | `""` |  |
| secretMounts | list | `[]` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.enabled | bool | `true` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `1000` |  |
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
| sysctlInitContainer.enabled | bool | `true` |  |
| sysctlVmMaxMapCount | int | `262144` |  |
| terminationGracePeriod | int | `120` |  |
| tolerations | list | `[]` |  |
| transportPort | int | `9300` |  |
| updateStrategy | string | `"RollingUpdate"` |  |
| volumeClaimTemplate.accessModes[0] | string | `"ReadWriteOnce"` |  |
| volumeClaimTemplate.resources.requests.storage | string | `"30Gi"` |  |
