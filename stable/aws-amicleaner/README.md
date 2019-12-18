aws-amicleaner
==============
AMI cleaner -- Clean old AWS AMI images.

Current chart version is `0.2.1`

Source code can be found [here](https://gitlab.com/stackvista/devops/helm-charts.git)



## Chart Values

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
| successfulJobsHistoryLimit | int | `5` | The number of successful CronJob executions that are saved. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |
