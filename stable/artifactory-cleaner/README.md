# artifactory-cleaner

![Version: 0.0.2](https://img.shields.io/badge/Version-0.0.2-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)
Artifactory Cleaner -- Clean up artifacts according to retention policies.
**Homepage:** <https://gitlab.com/stackvista/devops/helm-charts.git>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment. |
| artifactory.password | string | `nil` | The password for the Artifactory user |
| artifactory.url | string | `nil` | The Artifactory URL |
| artifactory.user | string | `nil` | An Artifactory user with the permissions to delete artifacts |
| backoffLimit | int | `3` | For failed jobs, how many times to retry. |
| failedJobsHistoryLimit | int | `5` | The number of failed CronJob executions that are saved. |
| fullnameOverride | string | `""` | Override the fullname of the chart. |
| image.pullPolicy | string | `"Always"` | Default container image pull policy. |
| image.repository | string | `"quay.io/stackstate/sts-ci-images"` | Base container image registry. Any image with curl and jq will do. |
| image.tag | string | `"stackstate-devops-d3ac6ed1"` | Default container image tag. |
| imagePullSecrets | list | `[]` | Extra secrets / credentials needed for container image registry. |
| nameOverride | string | `""` | Override the name of the chart. |
| nodeSelector | object | `{}` | Node labels for pod assignment. |
| podAnnotations | object | `{}` | Annotations for the `Job` pod. |
| resources.limits.cpu | string | `"100m"` | CPU resource limits. |
| resources.limits.memory | string | `"64Mi"` | Memory resource limits. |
| resources.requests.cpu | string | `"100m"` | CPU resource requests. |
| resources.requests.memory | string | `"64Mi"` | Memory resource requests. |
| restartPolicy | string | `"Never"` | For failed jobs, how to handle restarts. |
| retentionPolicies | list | `[]` | The retention policies. Check the Chart's Readme for more info. |
| schedule | string | `"0 23 * * *"` | Default schedule for this CronJob. |
| serviceAccount.annotations | object | `{}` | Extra annotations for the `ServiceAccount` object. |
| serviceAccount.create | bool | `false` | Create the `ServiceAccount` object. |
| successfulJobsHistoryLimit | int | `5` | The number of successful CronJob executions that are saved. |
| tolerations | list | `[]` | Toleration labels for pod assignment. |

## Overview
The cleaner uses the [artifactory query language](https://www.jfrog.com/confluence/display/JFROG/Artifactory+Query+Language), AQL, to discover artifacts for deletion.
The template for the AQL request rendered based on the values from Retention Policies manifest passed as a Helm value.

```java
// The AQL request template
// ${repo} - is a mandatory parameter which describes the name of the Artifactory repository
// ${until}, ${keep} - one of these is mandatory depending on the policy type
items
  .find({
    "repo":{"\$eq":"${repo}"},
    "path":{"\$match" : "${repo_path}/*x"},
    "path":{"\$nmatch" : "${exclude}"},
    "created":{"\$lt":"${until}"},
    "name":{"\$match":"${name_filter}"},
    "type":"file"
  })
  .include("name", "repo", "path", "created")
  .sort({"\$desc": ["created"]}).offset(${keep})
```

There are types of policies supported:
- type: **time-based** To delete artifacts based on their creation times
- type: **keep-last** To keep the recent artifacts only

Below the examples of policies that have been tested with the script:

```
- type: time-based
  age: 12 weeks ago
  repo: libs-candidates-local
  nameFilter: "*.pom"
- type: time-based
  age: 8 weeks ago
  path: zips/stackgraph-distr
  exclude: zips/stackgraph-distr/*.*.*
  repo: libs
- type: keep-last
  artifactsToKeep: 100
  repo: libs-trunk-local
  nameFilter: "*.pom"
- type: time-based
  age: 1 week ago
  repo: libs-snapshot-local
  path: com/stackstate
  nameFilter: "*.pom"
- type: time-based
  age: 2 weeks ago
  path: com/stackstate/stackstate-distr
  repo: libs
- type: time-based
  age: 6 weeks ago
  path: com/stackstate/stackstate-distr
  repo: release-candidates
```
