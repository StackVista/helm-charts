stackstate-standalone
=====================
Helm chart for StackState standlone -- all components running inside a single container.

Current chart version is `0.3.2`

Source code can be found [here](https://gitlab.com/stackvista/devops/helm-charts.git)



## Required Values

In order to successfully install this chart, you **must** provide the following variables:
* `stackstate.license.key`
* `stackstate.receiver.baseUrl`

Install them on the command line on Helm with the following command:

```shell
helm install \
--set stackstate.license.key=<your-license-key> \
--set stackstate.receiver.baseUrl=<your-base-url> \
stackstate/stackstate-standalone
```

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| deployment.annotations | object | `{}` | Annotations to attach to the `Deployment` object. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| gitlab.app | string | `""` | If CI is GitLab, specify the `app` for annotations. |
| gitlab.env | string | `""` | If CI is GitLab, specify the `env` for annotations. |
| image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| image.repository | string | `"508573134510.dkr.ecr.eu-west-1.amazonaws.com/stackstate"` | Base container image registry. |
| image.tag | string | `"sts-v1-14-10-1"` | Default container image tag. |
| imagePullSecrets | list | `[]` | Extra secrets / credentials needed for container image registry. |
| ingress.annotations | object | `{}` | Annotations for ingress objects. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| ingress.hosts | list | `[]` | List of ingress hostnames; the paths are fixed to StackState backend services |
| ingress.path.admin | string | `"/admin"` | Ingress path to the admin service. |
| ingress.path.receiver | string | `"/receiver"` | Ingress path to the receiver service. |
| ingress.path.ui | string | `"/"` | Ingress path to the base UI. |
| ingress.tls | list | `[]` | List of ingress TLS certificates to use. |
| livenessProbe.enabled | bool | `true` | Enable use of livenessProbe check. |
| livenessProbe.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| livenessProbe.initialDelaySeconds | int | `120` | `initialDelaySeconds` for the liveness probe. |
| livenessProbe.periodSeconds | int | `10` | `periodSeconds` for the liveness probe. |
| livenessProbe.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| livenessProbe.timeoutSeconds | int | `2` | `timeoutSeconds` for the liveness probe. |
| nameOverride | string | `""` | Override the name of the chart. |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| persistence.accessMode | string | `"ReadWriteOnce"` | Access mode of the persistent volume claim. |
| persistence.annotations | object | `{}` | Annotations to attach to the `PersistentVolumeClaim` object. |
| persistence.enabled | bool | `false` | Enable use of persistence. |
| persistence.size | string | `"20Gi"` | Size (in GiB) of the persistent volume. |
| persistence.storageClass | string | `"gp2"` | Name of the storage class to use for the persistent volume. |
| pod.annotations | object | `{}` | Annotations to attach to the `Pod` object(s). |
| readinessProbe.enabled | bool | `true` | Enable use of readinessProbe check. |
| readinessProbe.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| readinessProbe.initialDelaySeconds | int | `120` | `initialDelaySeconds` for the readiness probe. |
| readinessProbe.periodSeconds | int | `10` | `periodSeconds` for the readiness probe. |
| readinessProbe.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| readinessProbe.timeoutSeconds | int | `2` | `timeoutSeconds` for the readiness probe. |
| resources.limits.cpu | string | `"3"` | CPU resource limits. |
| resources.limits.memory | string | `"8Gi"` | Memory resource limits. |
| resources.requests.cpu | string | `"2"` | CPU resource requests. |
| resources.requests.memory | string | `"4Gi"` | Memory resource requests. |
| service.admin.port | int | `7071` | The default port for the StackState Administration area. |
| service.annotations | object | `{}` | Annotations to attach to the `Service` object. |
| service.receiver.port | int | `7077` | The default port for the StackState Receiver. |
| service.type | string | `"ClusterIP"` | The Kubernetes 'Service' type to use. |
| service.ui.port | int | `7070` | The default port for the StackState UI. |
| stackstate.demoData.enabled | bool | `false` | Whether or not to enable demo data for branch deploys. |
| stackstate.license.key | string | `nil` | **PROVIDE YOUR LICENSE KEY HERE** The StackState license key needed to start the server. |
| stackstate.receiver.apiKey | string | `""` | API key to be used by the Receiver; if no key is provided, a random one will be generated for you. |
| stackstate.receiver.baseUrl | string | `nil` | **PROVIDE YOUR BASE URL HERE** Externally visible baseUrl of the StackState endpoints. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
