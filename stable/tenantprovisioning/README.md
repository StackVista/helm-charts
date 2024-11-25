# tenantprovisioning

![Version: 0.0.5](https://img.shields.io/badge/Version-0.0.5-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)
Create tenants manifests by Hubspot webhook.
**Homepage:** <https://gitlab.com/stackvista/devops/helm-charts.git>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| config.AWS.Region | string | `nil` | Region to connect to |
| config.AWS.SQSBaseEndpoint | string | `nil` | Url of incoming provisioning queue |
| config.AWS.TenantProvisioningInternalWorkQueueURL | string | `nil` | Url of internal progress queue |
| config.AWS.TenantProvisioningQueueURL | string | `nil` |  |
| config.AWS.TenantProvisioningStatusOutputQueueURL | string | `nil` | Url of output status |
| config.ArgoCD.AuthToken | string | `nil` | Authentication token. |
| config.ArgoCD.GitRepo | string | `nil` | Git repo argo will pull, should be the git repo url as in config.Git.RepoURL but then how argo pulls it. This setting is separate because sometimes one uses https:// and the other git@ ssh style |
| config.ArgoCD.Insecure | bool | `false` | Do we allow insecure access? |
| config.ArgoCD.ServerAddr | string | `nil` | Address of the argocd server of the form <server>:<port>, no protocol spec! |
| config.Clusters | list | `[{"AvailabilityZones":[],"Name":null}]` | Clusters configuration. |
| config.Clusters[0].AvailabilityZones | list | `[]` | cluster availability zones. |
| config.Clusters[0].Name | string | `nil` | cluster name. |
| config.GenericWebhookAuthToken | string | `nil` | Token to protect Generic webhook endpoint with. |
| config.Git.Auth.Password | string | `nil` | Password for Git authentication. |
| config.Git.Auth.Username | string | `nil` | Username for Git authentication. |
| config.Git.Branch | string | `nil` | Branch to check out. |
| config.Git.CommitAuthor.Email | string | `nil` | Email of the commit author. |
| config.Git.CommitAuthor.Name | string | `nil` | Name of the commit author. |
| config.Git.RepoURL | string | `nil` | URL of the Git repository. |
| config.ListenAddr | string | `"0.0.0.0:8080"` | Address and port to listen on. |
| config.PrivateGPGKeyBase64Encoded | string | `nil` | Base64-encoded private GPG key to sign commits. Must not be protected with passphrase. |
| config.TmpDir | string | `"/tmp"` | Temporary directory for the application. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| global.imagePullSecrets | list | `[]` | Globally add image pull secrets that are used. |
| global.imageRegistry | string | `nil` | Globally override the image registry that is used. Can be overridden by specific containers. Defaults to quay.io |
| image.pullPolicy | string | `"IfNotPresent"` | Default container image pull policy. |
| image.registry | string | `nil` | Registry containing the image for the Redirector |
| image.repository | string | `"stackstate/o11y-tooling"` | Base container image registry. Any image with kubectl, jq, aws-cli and gsutil will do. |
| image.tag | string | `"7468ec65"` | Default container image tag. |
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
| resources.limits.memory | string | `"384Mi"` | Memory resource limits. |
| resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| resources.requests.memory | string | `"384Mi"` | Memory resource requests. |
| securityContext.fsGroup | int | `1000` |  |
| securityContext.runAsGroup | int | `1000` |  |
| securityContext.runAsUser | int | `1000` |  |
| serviceAccount.annotations | object | `{}` | Extra annotations for the `ServiceAccount` object. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |

## Overview
tenantprovisioning accepts an SQS data, generates tenants manifests for o11y-tenants repository and push them to Git.
