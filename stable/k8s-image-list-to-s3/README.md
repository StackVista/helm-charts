# k8s-image-list-to-s3

![Version: 0.0.2](https://img.shields.io/badge/Version-0.0.2-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)
Get the list of the Docker images deployed to the K8s cluster and uploads it to S3 bucket
**Homepage:** <https://gitlab.com/stackvista/devops/helm-charts.git>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| backoffLimit | int | `3` | For failed jobs, how many times to retry. |
| failedJobsHistoryLimit | int | `5` | The number of failed CronJob executions that are saved. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| image.pullPolicy | string | `"Always"` | Default container image pull policy. |
| image.repository | string | `"quay.io/stackstate/sts-ci-images"` | Base container image registry. Any image with kubectl, jq, aws-cli and gsutil will do. |
| image.tag | string | `"stackstate-devops-9a5f8c3b"` | Default container image tag. |
| imagePullSecrets | list | `[]` | Extra secrets / credentials needed for container image registry. |
| nameOverride | string | `""` | Override the name of the chart. |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| podAnnotations | object | `{}` | Annotations for the `Job` pod. |
| reports.aws.bucket | object | `{"name":null,"prefix":"/","region":null}` | The name of the AWS S3 bucket to store reports |
| reports.aws.enabled | bool | `true` | True if the reports should be stored to AWS S3 bucket |
| reports.gcp.bucket | object | `{"name":null,"prefix":"/"}` | The name of the GCP Storage bucket to store reports |
| reports.gcp.enabled | bool | `false` | True if the reports should be stored to GCP Storage bucket. Ignored if `reports.aws.enabled` is true |
| resources.limits.cpu | string | `"100m"` | CPU resource limits. |
| resources.limits.memory | string | `"256Mi"` | Memory resource limits. |
| resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| resources.requests.memory | string | `"256Mi"` | Memory resource requests. |
| restartPolicy | string | `"Never"` | For failed jobs, how to handle restarts. |
| scan.ignoreNamespaceRegex | string | `""` | Skip the namespaces whose names match the regex used by https://jqlang.github.io/jq/manual/#test |
| scan.ignoreResourceNameRegex | string | `""` | Skip the pods whose names match the regex used by https://jqlang.github.io/jq/manual/#test |
| schedule | string | `"0 1 * * *"` | Default schedule for this CronJob. |
| serviceAccount.annotations | object | `{}` | Extra annotations for the `ServiceAccount` object. |
| successfulJobsHistoryLimit | int | `5` | The number of successful CronJob executions that are saved. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |

## Overview
k8s-image-list-to-s3 collects the list of the Docker images deployed to the Kubernetes cluster and uploads it to either AWS S3 or GCP Storage bucket. The list can be passed to the Docker image scanning tool for security scan.

To authenticate to AWS or GCP IAM roles for serviceaccounts are used. For that matter the serviceaccounts have to be configured with the proper annotations.
