# aws-amicleaner

![Version: 0.2.8](https://img.shields.io/badge/Version-0.2.8-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)
AMI cleaner -- Clean old AWS AMI images.
**Homepage:** <https://gitlab.com/stackvista/devops/helm-charts.git>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | ops@stackstate.com |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| aws.configFile | string | `"[default]"` | The AWS config file contents. |
| aws.credentialsFile | string | `""` | The AWS credentials file contents. |
| aws.mountPath | string | `"/home/stackstate/.aws"` | The mount path of the AWS config and credentials file. |
| backoffLimit | int | `3` | For failed jobs, how many times to retry. |
| failedJobsHistoryLimit | int | `5` | The number of failed CronJob executions that are saved. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| image.pullPolicy | string | `"Always"` | Default container image pull policy. |
| image.repository | string | `"quay.io/stackstate/aws-amicleaner"` | Base container image registry. |
| image.tag | string | `"master"` | Default container image tag. |
| imagePullSecrets | list | `[]` | Extra secrets / credentials needed for container image registry. |
| nameOverride | string | `""` | Override the name of the chart. |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| podAnnotations | object | `{}` | Annotations for the `Job` pod. |
| resources.limits.cpu | string | `"100m"` | CPU resource limits. |
| resources.limits.memory | string | `"128Mi"` | Memory resource limits. |
| resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| resources.requests.memory | string | `"128Mi"` | Memory resource requests. |
| restartPolicy | string | `"OnFailure"` | For failed jobs, how to handle restarts. |
| schedule | string | `"17 * * * *"` | Default schedule for this CronJob. |
| serviceAccount.annotations | object | `{}` | Extra annotations for the `ServiceAccount` object. |
| serviceAccount.create | bool | `true` | Create the `ServiceAccount` object. |
| successfulJobsHistoryLimit | int | `5` | The number of successful CronJob executions that are saved. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
