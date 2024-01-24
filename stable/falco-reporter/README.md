# falco-reporter

![Version: 0.0.1](https://img.shields.io/badge/Version-0.0.1-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)
Process Falco output: remove sensitive data, store to Cloud and send sanitized report to Slack
**Homepage:** <https://gitlab.com/stackvista/devops/helm-charts.git>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| aws.bucket | string | `nil` | AWS S3 bucket to store Falco reports |
| aws.region | string | `nil` | AWS region where the bucket resides |
| baseUrl | string | `nil` | Is used to build links in Slack reports. If not set ingress.host is used. If ingress.host is not set then the Kubernetes service is used. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| gcp.bucket | string | `nil` | GCP project where the bucket resides |
| gcp.project | string | `nil` |  |
| global.imagePullSecrets | list | `[]` | Globally add image pull secrets that are used. |
| global.imageRegistry | string | `nil` | Globally override the image registry that is used. Can be overridden by specific containers. Defaults to quay.io |
| image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| image.registry | string | `nil` | Registry containing the image for the Redirector |
| image.repository | string | `"stackstate/falco-reporter"` | Base container image registry. Any image with kubectl, jq, aws-cli and gsutil will do. |
| image.tag | string | `"6bf28661"` | Default container image tag. |
| imagePullSecrets | list | `[]` | Extra secrets / credentials needed for container image registry. |
| ingress.annotations | string | `nil` |  |
| ingress.enabled | bool | `false` | Whether to deploy Ingress resource. |
| ingress.host | string | `nil` | HTTP host for the ingress. |
| ingress.tls.enabled | bool | `false` | Whether to enable TLS for ingress. |
| ingress.tls.secretName | string | `nil` | The name of K8s secrets containing SSL certificate for ingress. |
| nameOverride | string | `""` | Override the name of the chart. |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| replicaCount | int | `1` | number of replicas to serve webhook |
| resources.limits.cpu | string | `"100m"` | CPU resource limits. |
| resources.limits.memory | string | `"256Mi"` | Memory resource limits. |
| resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| resources.requests.memory | string | `"256Mi"` | Memory resource requests. |
| serviceAccount.annotations | object | `{}` | Extra annotations for the `ServiceAccount` object. |
| slackWebhook | string | `nil` | Slack incoming webhook to send notifications to. |
| storageProvider | string | `nil` | Where to save Falco reports: gcp or aws |
| tolerations | list | `[]` | Toleration labels for pod assignment. |

## Overview
falco-reporter accepts an HTTP webhook from Falco, stores the Falco finding to a cloud storage and send a Falco finding with sensitive data removed to Slack.

To authenticate to AWS or GCP IAM roles for serviceaccounts are used. For that matter the serviceaccounts have to be configured with the proper annotations.
