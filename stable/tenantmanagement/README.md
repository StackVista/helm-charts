# tenantmanagement

![Version: 0.1.4](https://img.shields.io/badge/Version-0.1.4-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.0.0](https://img.shields.io/badge/AppVersion-1.0.0-informational?style=flat-square)
Manages all SaaS tenants
**Homepage:** <https://gitlab.com/stackvista/devops/helm-charts.git>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| SUSE Observability Team | <suse-observability-ops@suse.com> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| app.db.password | string | `""` |  |
| app.db.url | string | `""` |  |
| app.db.username | string | `""` |  |
| app.scaling.api_key | string | `""` |  |
| app.scaling.url | string | `""` |  |
| app.sqs.billing_queue | string | `""` |  |
| app.sqs.provisioning_queue | string | `""` |  |
| app.sqs.provisioning_status_queue | string | `""` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"quay.io/stackstate/tenant-management"` |  |
| image.tag | string | `"0.1.0-SNAPSHOT-3abd0d54"` |  |
| ingress.annotations | string | `nil` |  |
| ingress.enabled | bool | `false` | Whether to deploy Ingress resource. |
| ingress.host | string | `nil` | HTTP host for the ingress. |
| ingress.tls.enabled | bool | `false` | Whether to enable TLS for ingress. |
| ingress.tls.secretName | string | `nil` | The name of K8s secrets containing SSL certificate for ingress. |
| livenessProbe.httpGet.path | string | `"/status/live"` |  |
| livenessProbe.httpGet.port | string | `"http"` |  |
| livenessProbe.initialDelaySeconds | int | `20` |  |
| livenessProbe.timeoutSeconds | int | `5` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| pullSecret.password | string | `""` |  |
| pullSecret.username | string | `""` |  |
| readinessProbe.httpGet.path | string | `"/status/ready"` |  |
| readinessProbe.httpGet.port | string | `"http"` |  |
| readinessProbe.initialDelaySeconds | int | `20` |  |
| readinessProbe.timeoutSeconds | int | `5` |  |
| replicaCount | int | `1` |  |
| resources.limits.cpu | string | `"250m"` |  |
| resources.limits.memory | string | `"256Mi"` |  |
| resources.requests.cpu | string | `"100m"` |  |
| resources.requests.memory | string | `"256Mi"` |  |
| securityContext | object | `{}` |  |
| service.port | int | `80` |  |
| service.targetPort | int | `8080` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automount | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tolerations | list | `[]` |  |
| volumeMounts | list | `[]` |  |
| volumes | list | `[]` |  |

## Overview
tenantmanagement manages all SaaS tenants.
