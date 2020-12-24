# hbase

![Version: 0.1.61](https://img.shields.io/badge/Version-0.1.61-informational?style=flat-square) ![AppVersion: 1.2.6](https://img.shields.io/badge/AppVersion-1.2.6-informational?style=flat-square)

Helm chart for StackState HBase -- includes Zookeeper, and Hadoop for persistent storage.

**Homepage:** <https://gitlab.com/stackvista/devops/helm-charts.git>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Jeroen van Erp | jvanerp@stackstate.com |  |
| Remco Beckers | rbeckers@stackstate.com |  |
| Vincent Partington | vpartington@stackstate.com |  |

## Source Code

* <https://github.com/apache/hadoop>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | zookeeper | 5.3.4 |
| https://helm.stackstate.io/ | common | 0.4.8 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| all.affinity | object | `{}` | Affinity settings for pod assignment on all components. |
| all.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods for all components. |
| all.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object for all components. |
| all.image.pullSecretName | string | `nil` | Name of ImagePullSecret to use for all pods. |
| all.image.pullSecretPassword | string | `nil` | Password used to login to the registry to pull Docker images of all pods. |
| all.image.pullSecretUsername | string | `nil` | Username used to login to the registry to pull Docker images of all pods. |
| all.image.registry | string | `"quay.io"` | Base container image registry for all containers, except for the wait container |
| all.metrics.enabled | bool | `false` | Enable metrics port. |
| all.metrics.servicemonitor.additionalLabels | object | `{}` | Additional labels for targeting Prometheus operator instances. |
| all.metrics.servicemonitor.enabled | bool | `false` | Enable `ServiceMonitor` object; `all.metrics.enabled` *must* be enabled. |
| all.nodeSelector | object | `{}` | Node labels for pod assignment on all components. |
| all.tolerations | list | `[]` | Toleration labels for pod assignment on all components. |
| commonLabels | object | `{}` | Labels that will be applied to all resources created by this helm chart |
| console.affinity | object | `{}` | Affinity settings for pod assignment. |
| console.enabled | bool | `true` | Enable / disable deployment of the stackgraph-console for debugging. |
| console.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| console.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| console.image.pullPolicy | string | `nil` | Pull policy for console pods, defaults to `stackgraph.image.pullPolicy` |
| console.image.repository | string | `"stackstate/stackgraph-console"` | Base container image repository for console pods. |
| console.image.tag | string | `nil` | Container image tag for console pods, defaults to `stackgraph.image.tag` |
| console.nodeSelector | object | `{}` | Node labels for pod assignment. |
| console.resources | object | `{}` | Resources to allocate for HDFS data nodes. |
| console.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hbase.master.affinity | object | `{}` | Affinity settings for pod assignment. |
| hbase.master.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hbase.master.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hbase.master.image.pullPolicy | string | `nil` | Pull policy for HBase masters, defaults to `stackgraph.image.pullPolicy` |
| hbase.master.image.repository | string | `"stackstate/hbase-master"` | Base container image repository for HBase masters. |
| hbase.master.image.tag | string | `nil` | Container image tag for HBase masters, defaults to `stackgraph.image.tag` |
| hbase.master.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hbase.master.replicaCount | int | `1` | Number of pods for HBase masters. |
| hbase.master.resources | object | `{"limits":{"memory":"1Gi"},"requests":{"cpu":"50m","memory":"1Gi"}}` | Resources to allocate for HBase masters. |
| hbase.master.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hbase.regionserver.affinity | object | `{}` | Affinity settings for pod assignment. |
| hbase.regionserver.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hbase.regionserver.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hbase.regionserver.image.pullPolicy | string | `nil` | Pull policy for HBase region servers, defaults to `stackgraph.image.pullPolicy` |
| hbase.regionserver.image.repository | string | `"stackstate/hbase-regionserver"` | Base container image repository for HBase region servers. |
| hbase.regionserver.image.tag | string | `nil` | Container image tag for HBase region servers, defaults to `stackgraph.image.tag` |
| hbase.regionserver.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hbase.regionserver.replicaCount | int | `1` | Number of HBase regionserver nodes. |
| hbase.regionserver.resources | object | `{"limits":{"memory":"3Gi"},"requests":{"cpu":"2000m","memory":"2Gi"}}` | Resources to allocate for HBase region servers. |
| hbase.regionserver.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hbase.zookeeper.quorum | string | `"hbase"` | Zookeeper quorum used for single-node Zookeeper installations; not used if `zookeeper.replicaCount` is more than `1`. |
| hdfs.datanode.affinity | object | `{}` | Affinity settings for pod assignment. |
| hdfs.datanode.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hdfs.datanode.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hdfs.datanode.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hdfs.datanode.persistence.accessModes | list | `["ReadWriteOnce"]` | Access mode for HDFS data nodes. |
| hdfs.datanode.persistence.enabled | bool | `true` | Enable persistence for HDFS data nodes. |
| hdfs.datanode.persistence.size | string | `"250Gi"` | Size of volume for HDFS data nodes. |
| hdfs.datanode.persistence.storageClass | string | `nil` | Storage class of the volume for HDFS data nodes. |
| hdfs.datanode.replicaCount | int | `1` | Number of HDFS data nodes. |
| hdfs.datanode.resources | object | `{"limits":{"memory":"4Gi"},"requests":{"cpu":"50m","memory":"2Gi"}}` | Resources to allocate for HDFS data nodes. |
| hdfs.datanode.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hdfs.image.pullPolicy | string | `"Always"` | Pull policy for HDFS datanode. |
| hdfs.image.repository | string | `"stackstate/hadoop"` | Base container image repository for HDFS datanode. |
| hdfs.image.tag | string | `"2.9.2-java11"` | Default container image tag for HDFS datanode. |
| hdfs.namenode.affinity | object | `{}` | Affinity settings for pod assignment. |
| hdfs.namenode.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hdfs.namenode.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hdfs.namenode.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hdfs.namenode.persistence.accessModes | list | `["ReadWriteOnce"]` | Access mode for HDFS name nodes. |
| hdfs.namenode.persistence.enabled | bool | `true` | Enable persistence for HDFS name nodes. |
| hdfs.namenode.persistence.size | string | `"20Gi"` | Size of volume for HDFS name nodes. |
| hdfs.namenode.persistence.storageClass | string | `nil` | Storage class of the volume for HDFS name nodes. |
| hdfs.namenode.resources | object | `{"limits":{"memory":"1Gi"},"requests":{"cpu":"50m","memory":"1Gi"}}` | Resources to allocate for HDFS name nodes. |
| hdfs.namenode.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hdfs.scc.enabled | bool | `false` | Whether to create an OpenShift SecurityContextConfiguration (required when running on OpenShift) |
| hdfs.secondarynamenode.affinity | object | `{}` | Affinity settings for pod assignment. |
| hdfs.secondarynamenode.enabled | bool | `false` | Enable / disable secondary name nodes. |
| hdfs.secondarynamenode.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hdfs.secondarynamenode.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hdfs.secondarynamenode.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hdfs.secondarynamenode.persistence.accessModes | list | `["ReadWriteOnce"]` | Access mode for HDFS secondary name nodes. |
| hdfs.secondarynamenode.persistence.enabled | bool | `true` | Enable persistence for HDFS secondary name nodes. |
| hdfs.secondarynamenode.persistence.size | string | `"20Gi"` | Size of volume for HDFS secondary name nodes. |
| hdfs.secondarynamenode.persistence.storageClass | string | `nil` | Storage class of the volume for HDFS secondary name nodes. |
| hdfs.secondarynamenode.resources | object | `{"limits":{"memory":"1Gi"},"requests":{"cpu":"50m","memory":"1Gi"}}` | Resources to allocate for HDFS secondary name nodes. |
| hdfs.secondarynamenode.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hdfs.securityContext.enabled | bool | `true` | Whether to explicitly set the UID/GID of the pod. |
| hdfs.securityContext.runAsGroup | int | `65534` | GID of the Linux group to use for all pod. |
| hdfs.securityContext.runAsUser | int | `65534` | UID of the Linux user to use for all pod. |
| hdfs.volumePermissions.enabled | bool | `true` | Whether to explicitly change the volume permissions for the data/name nodes |
| hdfs.volumePermissions.securityContext.allowPrivilegeEscalation | bool | `true` | Run the volumePermissions init container with privilege escalation mode (Do not change unless instructed) |
| hdfs.volumePermissions.securityContext.enabled | bool | `true` | Whether to add a securityContext to the volumePermissions init container |
| hdfs.volumePermissions.securityContext.privileged | bool | `true` | Run the volumePermissions init container in privileged mode (required for plain K8s, not for OpenShift) |
| hdfs.volumePermissions.securityContext.runAsNonRoot | bool | `false` | Run the volumePermissions init container in non-root required mode (Do not change unless instructed) |
| hdfs.volumePermissions.securityContext.runAsUser | int | `0` | Run the volumePermissions init container with the specified UID (Do not change unless instructed) |
| securityContext.enabled | bool | `true` | Whether to explicitly set the UID/GID of the container. |
| securityContext.runAsGroup | int | `65534` | GID of the Linux group to use for all containers. |
| securityContext.runAsUser | int | `65534` | UID of the Linux user to use for all containers. |
| stackgraph.image.pullPolicy | string | `"Always"` | The default pullPolicy used for all components of hbase that are stackgraph version dependent; invividual service `pullPolicy`s can be overriden (see below). |
| stackgraph.image.tag | string | `"3.6.16"` | The default tag used for all omponents of hbase that are stackgraph version dependent; invividual service `tag`s can be overriden (see below). |
| statefulset.antiAffinity.strategy | string | `"soft"` | AntiAffinity strategy to use for all StatefulSets. |
| statefulset.antiAffinity.topologyKey | string | `"kubernetes.io/hostname"` | AntiAffinity topology key to use for all StatefulSets. |
| tephra.affinity | object | `{}` | Affinity settings for pod assignment. |
| tephra.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| tephra.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| tephra.image.pullPolicy | string | `nil` | Pull policy for Tephra pods, defaults to `stackgraph.image.pullPolicy` |
| tephra.image.repository | string | `"stackstate/tephra-server"` | Base container image repository for Tephra pods. |
| tephra.image.tag | string | `nil` | Container image tag for Tephra pods, defaults to `stackgraph.image.tag` |
| tephra.nodeSelector | object | `{}` | Node labels for pod assignment. |
| tephra.replicaCount | int | `1` | Number of pods for Tephra pods. |
| tephra.resources | object | `{"limits":{"memory":"3Gi"},"requests":{"cpu":"50m","memory":"2Gi"}}` | Resources to allocate for Tephra pods. |
| tephra.tolerations | list | `[]` | Toleration labels for pod assignment. |
| volumePermissions.enabled | bool | `true` | Whether to explicitly change the volume permissions for the data/name nodes |
| wait.image.registry | string | `"docker.io"` | Base container image registry for wait containers |
| wait.image.repository | string | `"dokkupaas/wait"` | Container image tag for wait containers |
| wait.image.tag | string | `"latest"` |  |
| zookeeper.enabled | bool | `true` | Enable / disable chart-based Zookeeper. |
| zookeeper.externalServers | string | `""` | If `zookeeper.enabled` is set to `false`, use this list of external Zookeeper servers instead. |
| zookeeper.fourlwCommandsWhitelist | string | `"mntr, ruok, stat, srvr"` | Zookeeper four-letter-word (FLW) commands that are enabled. |
| zookeeper.metrics.enabled | bool | `true` | Enable / disable Zookeeper Prometheus metrics. |
| zookeeper.metrics.serviceMonitor.enabled | bool | `false` | Enable creation of `ServiceMonitor` objects for Prometheus operator. |
| zookeeper.metrics.serviceMonitor.selector | object | `{"release":"prometheus-operator"}` | Default selector to use to target a certain Prometheus instance. |
| zookeeper.replicaCount | int | `1` | Default amount of Zookeeper replicas to provision. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.4.0](https://github.com/norwoodj/helm-docs/releases/v1.4.0)
