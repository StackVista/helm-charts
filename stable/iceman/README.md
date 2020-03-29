iceman
======
Iceman -- Export configuration for all StackState instances in a cluster to an S3 bucket as backup (i.e. freeze their configuration state).

Current chart version is `0.1.5`

Source code can be found [here](https://gitlab.com/stackvista/devops/iceman.git)



## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| backoffLimit | int | `3` | For failed jobs, how many times to retry. |
| extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| failedJobsHistoryLimit | int | `5` | The number of failed CronJob executions that are saved. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| iceman.awsRegion | string | `"eu-west-1"` | Default AWS region where the S3 bucket resides. |
| iceman.logLevel | string | `"info"` | Log level of the Iceman application. |
| iceman.s3Bucket | string | `nil` | **REQUIRED** S3 bucket name to place the backup configuration. |
| iceman.stackstatePassword | string | `nil` | **REQUIRED** Administrator password for the Admin API running on all StackStates. |
| iceman.stackstateUsername | string | `nil` | **REQUIRED** Administrator username for the Admin API running on all StackStates. |
| image.pullPolicy | string | `"Always"` | Default container image pull policy. |
| image.repository | string | `"quay.io/stackstate/iceman"` | Base container image registry. |
| image.tag | string | `"master"` | Default container image tag. |
| imagePullSecrets | list | `[]` | Extra secrets / credentials needed for container image registry. |
| nameOverride | string | `""` | Override the name of the chart. |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| podAnnotations | object | `{}` | Annotations to inject into `Job` pods. |
| rbac.serviceAccountAnnotations | object | `{}` | Additional `ServiceAccount` annotations. |
| resources.limits.cpu | string | `"100m"` | CPU resource limits. |
| resources.limits.memory | string | `"128Mi"` | Memory resource limits. |
| resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| resources.requests.memory | string | `"128Mi"` | Memory resource requests. |
| restartPolicy | string | `"OnFailure"` | For failed jobs, how to handle restarts. |
| schedule | string | `"0 */4 * * *"` | Default schedule for this CronJob. |
| securityContext | object | `{"fsGroup":65534}` | Security context for the `CronJob` object. |
| successfulJobsHistoryLimit | int | `5` | The number of successful CronJob executions that are saved. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
