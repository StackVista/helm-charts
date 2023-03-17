# k8s-spot-termination-handler

![Version: 1.4.11](https://img.shields.io/badge/Version-1.4.11-informational?style=flat-square) ![AppVersion: 1.14.3](https://img.shields.io/badge/AppVersion-1.14.3-informational?style=flat-square)

The K8s Spot Termination handler handles draining AWS Spot Instances in response to termination requests.

Current chart version is `1.4.11`

**Homepage:** <https://gitlab.com/stackvista/devops/kube-spot-termination-handler>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| rbeckers | <rbeckers@stackstate.com> |  |
| viliakov | <viliakov@stackstate.com> |  |

## Source Code

* <https://gitlab.com/stackvista/devops/kube-spot-termination-handler>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| detachAsg | bool | `false` |  |
| enableLogspout | bool | `false` |  |
| gitlab.runnersnamespace | string | `nil` | By default the namespace `gitlab-runner` is inspected, this can be overriden by setting an alternative namespace here |
| gitlab.token | string | `nil` | If a gitlab token is  provided this enables Gitlab job restarts for jobs running on spot instances. The token is required to cancel/restart a job |
| gracePeriod | int | `120` |  |
| hostNetwork | bool | `true` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"quay.io/stackstate/k8s-spot-termination-handler"` |  |
| image.tag | string | `"1.14.3"` |  |
| imagePullSecrets | list | `[]` |  |
| maxUnavailable | int | `1` |  |
| nodeSelector | object | `{}` |  |
| noticeUrl | string | `"http://169.254.169.254/latest/meta-data/spot/termination-time"` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| pollInterval | int | `5` |  |
| priorityClassName | string | `""` |  |
| rbac.create | bool | `true` |  |
| resources | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `nil` |  |
| slackUrl | string | `nil` | Send notifications to a Slack webhook URL, example slackUrl: https://hooks.slack.com/services/EXAMPLE123/EXAMPLE123/example1234567 |
| tolerations | list | `[]` |  |
| updateStrategy | string | `"RollingUpdate"` |  |
