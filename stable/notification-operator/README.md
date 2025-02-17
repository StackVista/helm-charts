# notification-operator

![Version: 0.0.3](https://img.shields.io/badge/Version-0.0.3-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

Notification Operator manages SuseObservability Notification Configurations with custom resources.

**Homepage:** <https://stackvista/devops/notification-operator.git>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| SUSE | <suse-observability-ops@suse.com> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules for pod scheduling |
| clusterDomain | string | `"cluster.local"` | The cluster domain name |
| fullnameOverride | string | `""` | Override the full name of the chart |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.repository | string | `"quay.io/stackstate/notification-operator-controller"` | Container image repository |
| image.tag | string | `"71dd134e"` | Overrides the image tag. Defaults to the chart appVersion. |
| imagePullSecrets | list | `[]` | List of secrets for pulling an image from a private repository |
| livenessProbe.httpGet.path | string | `"/healthz"` | HTTP path for liveness probe |
| livenessProbe.httpGet.port | int | `8081` | HTTP port for liveness probe |
| livenessProbe.initialDelaySeconds | int | `15` | Initial delay before liveness probe starts |
| livenessProbe.periodSeconds | int | `20` | Period between liveness probe checks |
| nameOverride | string | `""` | Override the chart name |
| nodeSelector | object | `{}` | Node selector for scheduling |
| podAnnotations | object | `{}` | Kubernetes annotations for the pod |
| podLabels | object | `{}` | Kubernetes labels for the pod |
| podSecurityContext | object | `{}` | Pod-level security context |
| readinessProbe.httpGet.path | string | `"/readyz"` | HTTP path for readiness probe |
| readinessProbe.httpGet.port | int | `8081` | HTTP port for readiness probe |
| readinessProbe.initialDelaySeconds | int | `5` | Initial delay before readiness probe starts |
| readinessProbe.periodSeconds | int | `10` | Period between readiness probe checks |
| replicaCount | int | `1` | Number of replicas for the deployment |
| resources.limits.cpu | string | `"500m"` | CPU limit for the container |
| resources.limits.memory | string | `"256Mi"` | Memory limit for the container |
| resources.requests.cpu | string | `"100m"` | CPU request for the container |
| resources.requests.memory | string | `"128Mi"` | Memory request for the container |
| securityContext.runAsNonRoot | bool | `true` | Ensure the container runs as a non-root user |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.automount | bool | `true` | Automatically mount API credentials to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | Name of the service account to use. Defaults to a generated name if left empty |
| tolerations | list | `[]` | Tolerations for pod scheduling |
| volumeMounts | list | `[]` | Additional volume mounts for the Deployment |
| volumes | list | `[]` | Additional volumes for the Deployment |

