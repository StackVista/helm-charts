# common

![Version: 0.4.26](https://img.shields.io/badge/Version-0.4.26-informational?style=flat-square) ![AppVersion: 0.4.25](https://img.shields.io/badge/AppVersion-0.4.25-informational?style=flat-square)
Common chartbuilding components and helpers
**Homepage:** <https://gitlab.com/stackvista/devops/helm-charts.git>
## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Stackstate Ops Team | <ops@stackstate.com> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| clusterrole.annotations | object | `{}` | Annotations for `Role` objects. |
| clusterrole.apiVersion | string | `"rbac.authorization.k8s.io/v1"` | Kubernetes apiVersion to use with a `Role` object. |
| clusterrolebinding.annotations | object | `{}` | Annotations for `ClusterRoleBinding` objects. |
| clusterrolebinding.apiVersion | string | `"rbac.authorization.k8s.io/v1"` | Kubernetes apiVersion to use with a `ClusterRoleBinding` object. |
| configmap.annotations | object | `{}` | Annotations for `ConfigMap` objects. |
| configmap.apiVersion | string | `"v1"` | Kubernetes apiVersion to use with a `ConfigMap` object. |
| container.livenessProbeDefaults.enabled | bool | `false` | Enable default options of `livenessProbe` check. |
| container.livenessProbeDefaults.failureThreshold | int | `3` | `failureThreshold` for the liveness probe. |
| container.livenessProbeDefaults.initialDelaySeconds | int | `10` | `initialDelaySeconds` for the liveness probe. |
| container.livenessProbeDefaults.periodSeconds | int | `10` | `periodSeconds` for the liveness probe. |
| container.livenessProbeDefaults.successThreshold | int | `1` | `successThreshold` for the liveness probe. |
| container.livenessProbeDefaults.timeoutSeconds | int | `2` | `timeoutSeconds` for the liveness probe. |
| container.readinessProbeDefaults.enabled | bool | `false` | Enable default options of `readinessProbe` check. |
| container.readinessProbeDefaults.failureThreshold | int | `3` | `failureThreshold` for the readiness probe. |
| container.readinessProbeDefaults.initialDelaySeconds | int | `10` | `initialDelaySeconds` for the readiness probe. |
| container.readinessProbeDefaults.periodSeconds | int | `10` | `periodSeconds` for the readiness probe. |
| container.readinessProbeDefaults.successThreshold | int | `1` | `successThreshold` for the readiness probe. |
| container.readinessProbeDefaults.timeoutSeconds | int | `2` | `timeoutSeconds` for the readiness probe. |
| container.resources | object | `{}` | Container resource requests / limits. |
| container.securityContext.allowPrivilegeEscalation | bool | `false` | Controls whether a process can gain more privileges than its parent process. |
| container.securityContext.capabilities.drop | list | `["ALL"]` | Drops all Linux capabilities by default. |
| container.securityContext.runAsNonRoot | bool | `true` |  |
| container.securityContext.seccompProfile.type | string | `"RuntimeDefault"` | Sets the Seccomp profile for a Container. |
| cronjob.annotations | object | `{}` | Annotations for `CronJob` objects. |
| cronjob.concurrencyPolicy | string | `"Forbid"` | Specifies how to treat concurrent executions of a job that is created by this `CronJob` object. |
| cronjob.dnsPolicy | string | `"ClusterFirst"` | DNS policy to apply to `Pod` objects. |
| cronjob.failedJobsHistoryLimit | int | `5` | Specifies how many failed `Job` objects created from a parent `CronJob` should be kept. |
| cronjob.jobTemplate.annotations | object | `{}` | Annotations for `Job` objects created from a parent `CronJob` object. |
| cronjob.nodeSelector | object | `{}` | Node labels for pod assignment. |
| cronjob.restartPolicy | string | `"OnFailure"` | Restart policy to apply to containers within a `Pod` object. |
| cronjob.schedule | string | `"0 0 * * *"` | A cron string, such as `0 * * * *` or `@hourly`, which specifies when `Job` objects are created and executed. |
| cronjob.successfulJobsHistoryLimit | int | `5` | Specifies how many successful `Job` objects created from a parent `CronJob` should be kept. |
| cronjob.suspend | bool | `false` | If set to `true`, all subsequent executions are suspended; this setting does not apply to already started executions. |
| cronjob.tolerations | list | `[]` | Toleration labels for `Pod` assignment. |
| daemonset.annotations | object | `{}` | Annotations for `DaemonSet` objects. |
| daemonset.apiVersion | string | `"apps/v1"` | Kubernetes apiVersion to use with a `DaemonSet` object. |
| daemonset.dnsPolicy | string | `"ClusterFirst"` | DNS policy to apply to `Pod` objects. |
| daemonset.nodeSelector | object | `{}` | Node labels for pod assignment. |
| daemonset.restartPolicy | string | `"Always"` | Restart policy to apply to containers within a `Pod` object. |
| daemonset.securityContext | object | `{}` | The security settings for all containers in a `Pod` object. |
| daemonset.terminationGracePeriodSeconds | int | `30` | Time (in seconds) to wait for a `Pod` object to shut down gracefully. |
| daemonset.tolerations | list | `[]` | Toleration labels for `Pod` assignment. |
| daemonset.updateStrategy | object | `{"type":"RollingUpdate"}` | Strategy to use for deploying `DaemonSet` objects. |
| deployment.affinity | object | `{}` | Affinity settings for pod assignment. |
| deployment.annotations | object | `{}` | Annotations for `Deployment` objects. |
| deployment.antiAffinity.strategy | string | `""` | Spread pods using simple `antiAffinity`; valid values are `hard` or `soft`. |
| deployment.antiAffinity.topologyKey | string | `"failure-domain.beta.kubernetes.io/zone"` | The `topologyKey` to use for simple `antiAffinity` rule. |
| deployment.apiVersion | string | `"apps/v1"` | Kubernetes apiVersion to use with a `Deployment` object. |
| deployment.dnsPolicy | string | `"ClusterFirst"` | DNS policy to apply to `Pod` objects. |
| deployment.nodeSelector | object | `{}` | Node labels for pod assignment. |
| deployment.progressDeadlineSeconds | int | `600` | Denotes the number of seconds the `Deployment` controller waits before indicating (in the `Deployment` status) that the `Deployment` progress has stalled. |
| deployment.replicaCount | int | `1` | Amount of replicas to create for the `Deployment` object. |
| deployment.restartPolicy | string | `"Always"` | Restart policy to apply to containers within a `Pod` object. |
| deployment.securityContext | object | `{}` | The security settings for all containers in a `Pod` object. |
| deployment.strategy | object | `{}` | Strategy to use for deploying `Deployment` objects. |
| deployment.terminationGracePeriodSeconds | int | `30` | Time (in seconds) to wait for a `Pod` object to shut down gracefully. |
| deployment.tolerations | list | `[]` | Toleration labels for `Pod` assignment. |
| extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| gitlab.app | string | `""` | If CI is GitLab, specify the `app` for annotations. |
| gitlab.env | string | `""` | If CI is GitLab, specify the `env` for annotations. |
| global | object | `{}` |  |
| image.pullPolicy | string | `"IfNotPresent"` | The pull policy for the Docker image. |
| image.repository | string | `"nginx"` | Repository of the Docker image. |
| image.tag | string | `"latest"` | Tag of the Docker image. |
| ingress.annotations | object | `{}` | Annotations for `Ingress` objects. |
| ingress.hosts | list | `[]` | List of ingress hostnames. |
| ingress.tls | list | `[]` | List of ingress TLS certificates to use. |
| job.annotations | object | `{}` | Annotations for `Job` objects. |
| job.apiVersion | string | `"batch/v1"` | Kubernetes apiVersion to use with a `Job` object. |
| job.backoffLimit | int | `5` | Number of retries before considering a `Job` as failed |
| job.dnsPolicy | string | `"ClusterFirst"` | DNS policy to apply to `Pod` objects. |
| job.nodeSelector | object | `{}` | Node labels for pod assignment. |
| job.restartPolicy | string | `"OnFailure"` | Restart policy to apply to containers within a `Pod` object. |
| job.suspend | bool | `false` | If set to `true`, all subsequent executions are suspended; this setting does not apply to already started executions. |
| job.tolerations | list | `[]` | Toleration labels for `Pod` assignment. |
| job.ttlSecondsAfterFinished | int | `300` | TTL for completed jobs before they are deleted. |
| networkpolicy.annotations | object | `{}` | Annotations for `NetworkPolicy` objects. |
| networkpolicy.apiVersion | string | `"networking.k8s.io/v1"` | Kubernetes apiVersion to use with a `NetworkPolicy` object. |
| persistentvolumeclaim.accessMode | string | `"ReadWriteOnce"` | Default type of access for mounted volume. |
| persistentvolumeclaim.annotations | object | `{}` | Annotations for `PersistentVolumeClaim` objects. |
| persistentvolumeclaim.apiVersion | string | `"v1"` | Kubernetes apiVersion to use with a `PersistentVolumeClaim` object. |
| persistentvolumeclaim.size | string | `"10Gi"` | Default size for `PersistentVolumeClaim` object. |
| persistentvolumeclaim.storageClass | string | `nil` | Default storage class for `PersistentVolumeClaim` object. If left blank, then the default provisioner is used. If set to "-", then `storageClass: ""`, which disabled dynamic provisioning. |
| pod.annotations | object | `{}` | Annotations for `Pod` objects. |
| poddisruptionbudget.annotations | object | `{}` | Annotations for `PodDisruptionBudget` objects. |
| role.annotations | object | `{}` |  |
| role.apiVersion | string | `"rbac.authorization.k8s.io/v1"` |  |
| rolebinding.annotations | object | `{}` | Annotations for `RoleBinding` objects. |
| rolebinding.apiVersion | string | `"rbac.authorization.k8s.io/v1"` | Kubernetes apiVersion to use with a `RoleBinding` object. |
| secret.annotations | object | `{}` | Annotations for `Secret` objects. |
| secret.apiVersion | string | `"v1"` | Kubernetes apiVersion to use with a `Secret` object. |
| service.annotations | object | `{}` | Annotations for `Service` objects. |
| service.apiVersion | string | `"v1"` | Kubernetes apiVersion to use with a `Service` object. |
| service.clusterIP | string | `""` |  |
| service.externalIPs | list | `[]` | List of external IP addresses to map to the `Service` object. |
| service.externalTrafficPolicy | string | `"Cluster"` | Denotes if this Service desires to route external traffic to node-local or cluster-wide endpoints; only enabled when `service.type` is set to `LoadBalancer` or `NodePort`. |
| service.loadBalancerIP | string | `""` | Specify the source IP range to allow traffic from; only enabled when `service.type` is set to `LoadBalancer`. |
| service.loadBalancerSourceRanges | list | `[]` |  |
| service.ports | list | `[]` | List of ports to apply to the `Service` object. |
| service.type | string | `"ClusterIP"` | Kubernetes 'Service' type to use. |
| serviceaccount.annotations | object | `{}` | Annotations for `ServiceAccount` objects. |
| serviceaccount.apiVersion | string | `"v1"` | Kubernetes apiVersion to use with a `ServiceAccount` object. |
| servicemonitor.annotations | object | `{}` | Annotations for `ServiceMonitor` objects. |
| servicemonitor.apiVersion | string | `"monitoring.coreos.com/v1"` | Kubernetes apiVersion to use with a `ServiceMonitor` object. |
| servicemonitor.interval | string | `"20s"` | Default interval used on `Service` objects for monitoring through a `ServiceMonitor` object. |
| servicemonitor.path | string | `"/metrics"` | Default path used on `Service` objects for monitoring through a `ServiceMonitor` object. |
| servicemonitor.port | string | `"metrics"` | Default port used on `Service` objects for monitoring through a `ServiceMonitor` object. |
| servicemonitor.scheme | string | `"http"` | Default scheme used on `Service` objects for monitoring through a `ServiceMonitor` object. |
| statefulset.affinity | object | `{}` | Affinity settings for pod assignment. |
| statefulset.annotations | object | `{}` | Annotations for `StatefulSet` objects. |
| statefulset.antiAffinity.strategy | string | `""` | Spread pods using simple `antiAffinity`; valid values are `hard` or `soft`. |
| statefulset.antiAffinity.topologyKey | string | `"failure-domain.beta.kubernetes.io/zone"` | The `topologyKey` to use for simple `antiAffinity` rule. |
| statefulset.apiVersion | string | `"apps/v1"` | Kubernetes apiVersion to use with a `StatefulSet` object. |
| statefulset.dnsPolicy | string | `"ClusterFirst"` | DNS policy to apply to `Pod` objects. |
| statefulset.nodeSelector | object | `{}` | Node labels for pod assignment. |
| statefulset.podManagementPolicy | string | `"OrderedReady"` | Policy to dictate how `StatefulSet` pods are updated; either `OrderedReady`, or `Parallel`. |
| statefulset.replicaCount | int | `1` | Amount of replicas to create for the `StatefulSet` object. |
| statefulset.restartPolicy | string | `"Always"` | Restart policy to apply to containers within a `Pod` object. |
| statefulset.terminationGracePeriodSeconds | int | `30` | Time (in seconds) to wait for a `Pod` object to shut down gracefully. |
| statefulset.tolerations | list | `[]` | Toleration labels for `Pod` assignment. |
| statefulset.updateStrategy | object | `{"type":"RollingUpdate"}` | Strategy to use for deploying `StatefulSet` objects. |
