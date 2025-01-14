# minio

![Version: 8.0.10-stackstate.10](https://img.shields.io/badge/Version-8.0.10--stackstate.10-informational?style=flat-square) ![AppVersion: master](https://img.shields.io/badge/AppVersion-master-informational?style=flat-square)

High Performance, Kubernetes Native Object Storage

Current chart version is `8.0.10-stackstate.10`

**Homepage:** <https://min.io>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| MinIO, Inc | <dev@minio.io> |  |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../common/ | common | * |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| DeploymentUpdate.maxSurge | string | `"100%"` |  |
| DeploymentUpdate.maxUnavailable | int | `0` |  |
| DeploymentUpdate.type | string | `"RollingUpdate"` |  |
| StatefulSetUpdate.updateStrategy | string | `"RollingUpdate"` |  |
| accessKey | string | `""` |  |
| additionalAnnotations | list | `[]` |  |
| additionalLabels | list | `[]` |  |
| affinity | object | `{}` |  |
| azuregateway.enabled | bool | `false` |  |
| azuregateway.replicas | int | `4` |  |
| bucketRoot | string | `""` |  |
| buckets | list | `[]` |  |
| certsPath | string | `"/etc/minio/certs/"` |  |
| clusterDomain | string | `"cluster.local"` |  |
| configPathmc | string | `"/etc/minio/mc/"` |  |
| defaultBucket.enabled | bool | `false` |  |
| defaultBucket.name | string | `"bucket"` |  |
| defaultBucket.policy | string | `"none"` |  |
| defaultBucket.purge | bool | `false` |  |
| drivesPerNode | int | `1` |  |
| environment | object | `{}` |  |
| etcd.clientCert | string | `""` |  |
| etcd.clientCertKey | string | `""` |  |
| etcd.corednsPathPrefix | string | `""` |  |
| etcd.endpoints | list | `[]` |  |
| etcd.pathPrefix | string | `""` |  |
| existingSecret | string | `""` |  |
| extraArgs | list | `[]` |  |
| fullnameOverride | string | `""` |  |
| gcsgateway.enabled | bool | `false` |  |
| gcsgateway.gcsKeyJson | string | `""` |  |
| gcsgateway.projectId | string | `""` |  |
| gcsgateway.replicas | int | `4` |  |
| global.imageRegistry | string | `""` |  |
| global.storageClass | string | `nil` | StorageClass for all PVCs created by the chart. Can be overriden per PVC. |
| helmKubectlJqImage.pullPolicy | string | `"IfNotPresent"` |  |
| helmKubectlJqImage.registry | string | `"docker.io"` |  |
| helmKubectlJqImage.repository | string | `"bskim45/helm-kubectl-jq"` |  |
| helmKubectlJqImage.tag | string | `"3.1.0"` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.registry | string | `"quay.io"` |  |
| image.repository | string | `"stackstate/minio"` |  |
| image.tag | string | `"RELEASE.2025-01-13T16-22-00Z-4ae4220f"` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0] | string | `"chart-example.local"` |  |
| ingress.labels | object | `{}` |  |
| ingress.path | string | `"/"` |  |
| ingress.tls | list | `[]` |  |
| makeBucketJob.annotations | string | `nil` |  |
| makeBucketJob.podAnnotations | string | `nil` |  |
| makeBucketJob.resources.requests.memory | string | `"128Mi"` |  |
| makeBucketJob.securityContext.enabled | bool | `false` |  |
| makeBucketJob.securityContext.fsGroup | int | `1000` |  |
| makeBucketJob.securityContext.runAsGroup | int | `1000` |  |
| makeBucketJob.securityContext.runAsUser | int | `1000` |  |
| mcImage.pullPolicy | string | `"IfNotPresent"` |  |
| mcImage.registry | string | `"docker.io"` |  |
| mcImage.repository | string | `"minio/mc"` |  |
| mcImage.tag | string | `"RELEASE.2021-02-14T04-28-06Z"` |  |
| metrics.serviceMonitor.additionalLabels | object | `{}` |  |
| metrics.serviceMonitor.enabled | bool | `false` |  |
| metrics.serviceMonitor.relabelConfigs | object | `{}` |  |
| mode | string | `"standalone"` |  |
| mountPath | string | `"/export"` |  |
| nameOverride | string | `""` |  |
| nasgateway.enabled | bool | `false` |  |
| nasgateway.pv | string | `nil` |  |
| nasgateway.replicas | int | `4` |  |
| networkPolicy.allowExternal | bool | `true` |  |
| networkPolicy.enabled | bool | `false` |  |
| nodeSelector | object | `{}` |  |
| persistence.VolumeName | string | `""` |  |
| persistence.accessMode | string | `"ReadWriteOnce"` |  |
| persistence.enabled | bool | `true` |  |
| persistence.existingClaim | string | `""` |  |
| persistence.size | string | `"500Gi"` |  |
| persistence.storageClass | string | `""` |  |
| persistence.subPath | string | `""` |  |
| podAnnotations | object | `{}` |  |
| podDisruptionBudget.enabled | bool | `false` |  |
| podDisruptionBudget.maxUnavailable | int | `1` |  |
| podLabels | object | `{}` |  |
| priorityClassName | string | `""` |  |
| replicas | int | `4` |  |
| resources.limits.cpu | int | `1` |  |
| resources.limits.memory | string | `"4Gi"` |  |
| resources.requests.cpu | string | `"500m"` |  |
| resources.requests.memory | string | `"4Gi"` |  |
| s3gateway.accessKey | string | `""` |  |
| s3gateway.enabled | bool | `false` |  |
| s3gateway.replicas | int | `4` |  |
| s3gateway.secretKey | string | `""` |  |
| s3gateway.serviceEndpoint | string | `""` |  |
| secretKey | string | `""` |  |
| securityContext.enabled | bool | `true` |  |
| securityContext.fsGroup | int | `1000` |  |
| securityContext.runAsGroup | int | `1000` |  |
| securityContext.runAsUser | int | `1000` |  |
| service.annotations | object | `{}` |  |
| service.clusterIP | string | `nil` |  |
| service.externalIPs | list | `[]` |  |
| service.nodePort | int | `32000` |  |
| service.port | int | `9000` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `nil` |  |
| tls.certSecret | string | `""` |  |
| tls.enabled | bool | `false` |  |
| tls.privateKey | string | `"private.key"` |  |
| tls.publicCrt | string | `"public.crt"` |  |
| tolerations | list | `[]` |  |
| trustedCertsSecret | string | `""` |  |
| updatePrometheusJob.annotations | string | `nil` |  |
| updatePrometheusJob.podAnnotations | string | `nil` |  |
| updatePrometheusJob.securityContext.enabled | bool | `false` |  |
| updatePrometheusJob.securityContext.fsGroup | int | `1000` |  |
| updatePrometheusJob.securityContext.runAsGroup | int | `1000` |  |
| updatePrometheusJob.securityContext.runAsUser | int | `1000` |  |
| zones | int | `1` |  |
