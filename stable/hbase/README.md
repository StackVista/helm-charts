# hbase

Helm chart for StackState HBase -- includes Zookeeper, and Hadoop for persistent storage.

Current chart version is `0.2.40`

**Homepage:** <https://gitlab.com/stackvista/devops/helm-charts.git>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../common/ | common | * |
| https://raw.githubusercontent.com/bitnami/charts/eb5f9a9513d987b519f0ecd732e7031241c50328/bitnami | zookeeper | 8.1.2 |

## Required Values

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
| all.metrics.agentAnnotationsEnabled | bool | `true` |  |
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
| console.image.tag | string | `nil` | Container image tag for console pods, defaults to `version`-`stackgraph.version` |
| console.nodeSelector | object | `{}` | Node labels for pod assignment. |
| console.replicaCount | int | `0` | Amount of console replicas to provision. Default of 0, |
| console.resources | object | `{"limits":{"cpu":"500m","memory":"1Gi"},"requests":{"cpu":"50m","memory":"512Mi"}}` | Resources to allocate for HDFS console. |
| console.securityContext.enabled | bool | `true` | Whether to explicitly set the UID/GID of the pod. |
| console.securityContext.runAsGroup | int | `65534` | GID of the Linux group to use for all pod. |
| console.securityContext.runAsUser | int | `65534` | UID of the Linux user to use for all pod. |
| console.strategy | object | `{"type":"RollingUpdate"}` | The strategy for the Deployment object. |
| console.tolerations | list | `[]` | Toleration labels for pod assignment. |
| deployment.mode | string | `"Distributed"` |  |
| global.storageClass | string | `nil` | StorageClass for all PVCs created by the chart. Can be overriden per PVC. |
| hbase.master.affinity | object | `{}` | Affinity settings for pod assignment. |
| hbase.master.experimental.execLivenessProbe.enabled | bool | `false` | Whether to use a new scripted livenessProbe instead of the original HTTP check. Requires >= 4.11.5 version of the StackGraph docker images |
| hbase.master.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hbase.master.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hbase.master.image.pullPolicy | string | `nil` | Pull policy for HBase masters, defaults to `stackgraph.image.pullPolicy` |
| hbase.master.image.repository | string | `"stackstate/hbase-master"` | Base container image repository for HBase masters. |
| hbase.master.image.tag | string | `nil` | Container image tag for HBase masters, defaults to `version`-`stackgraph.version` |
| hbase.master.livenessProbe.httpPort | int | `16010` | The port of the Hbase master service to perform HTTP health checks upon. |
| hbase.master.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hbase.master.replicaCount | int | `1` | Number of pods for HBase masters. |
| hbase.master.resources | object | `{"limits":{"memory":"1Gi"},"requests":{"cpu":"50m","memory":"1Gi"}}` | Resources to allocate for HBase masters. |
| hbase.master.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hbase.regionserver.affinity | object | `{}` | Affinity settings for pod assignment. |
| hbase.regionserver.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hbase.regionserver.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hbase.regionserver.image.pullPolicy | string | `nil` | Pull policy for HBase region servers, defaults to `stackgraph.image.pullPolicy` |
| hbase.regionserver.image.repository | string | `"stackstate/hbase-regionserver"` | Base container image repository for HBase region servers. |
| hbase.regionserver.image.tag | string | `nil` | Container image tag for HBase region servers, defaults to `version`-`stackgraph.version` |
| hbase.regionserver.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hbase.regionserver.replicaCount | int | `1` | Number of HBase regionserver nodes. |
| hbase.regionserver.resources | object | `{"limits":{"memory":"3Gi"},"requests":{"cpu":"2000m","memory":"2Gi"}}` | Resources to allocate for HBase region servers. |
| hbase.regionserver.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hbase.securityContext.enabled | bool | `true` | Whether to explicitly set the UID/GID of the pod. |
| hbase.securityContext.runAsGroup | int | `65534` | GID of the Linux group to use for all pod. |
| hbase.securityContext.runAsUser | int | `65534` | UID of the Linux user to use for all pod. |
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
| hdfs.datanode.terminationGracePeriodSeconds | int | `600` | Grace period to stop the pod. We give some time to fix under replicated blocks in Pre Stop hook |
| hdfs.datanode.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hdfs.image.pullPolicy | string | `"IfNotPresent"` | Pull policy for HDFS datanode. |
| hdfs.image.repository | string | `"stackstate/hadoop"` | Base container image repository for HDFS datanode. |
| hdfs.image.tag | string | `nil` | Default container image tag for HDFS datanode. |
| hdfs.minReplication | int | `1` | Sets the minimum synchronous replication that the namenode will enforce when writing a block. This gives guarantees about the amount of copies of a single block. (If hdfs.datanode.replicaCount is set to a value less than this, the replicationfactor will be equal to the replicaCount.) |
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
| hdfs.securityContext.fsGroup | int | `65534` |  |
| hdfs.securityContext.runAsGroup | int | `65534` | GID of the Linux group to use for all pod. |
| hdfs.securityContext.runAsUser | int | `65534` | UID of the Linux user to use for all pod. |
| hdfs.version | string | `"java11-8-949969df"` |  |
| hdfs.volumePermissions.enabled | bool | `false` | Whether to explicitly change the volume permissions for the data/name nodes. If permissions on volume mounts are not correct for whatever reason this can be used to set them properly. Usually also requires enabling the securityContext because root user is required. |
| hdfs.volumePermissions.securityContext.allowPrivilegeEscalation | bool | `true` | Run the volumePermissions init container with privilege escalation mode (Do not change unless instructed) |
| hdfs.volumePermissions.securityContext.enabled | bool | `false` | Whether to add a securityContext to the volumePermissions init container |
| hdfs.volumePermissions.securityContext.privileged | bool | `true` | Run the volumePermissions init container in privileged mode (required for plain K8s, not for OpenShift) |
| hdfs.volumePermissions.securityContext.runAsNonRoot | bool | `false` | Run the volumePermissions init container in non-root required mode (Do not change unless instructed) |
| hdfs.volumePermissions.securityContext.runAsUser | int | `0` | Run the volumePermissions init container with the specified UID (Do not change unless instructed) |
| serviceAccount.create | bool | `true` | Whether to create serviceAccounts and run the statefulsets under them |
| stackgraph.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackgraph.agentMetricsFilter | string | `""` | Configure metrics scraped by the agent |
| stackgraph.image.pullPolicy | string | `"IfNotPresent"` | The default pullPolicy used for all components of hbase that are stackgraph version dependent; invividual service `pullPolicy`s can be overriden (see below). |
| stackgraph.image.repository | string | `"stackstate/stackgraph-hbase"` | The default repository used for the single service stackgraph image |
| stackgraph.image.tag | string | `nil` | The default tag used for the single service stackgraph image |
| stackgraph.livenessProbe.httpPort | int | `16010` | The port of the Hbase master service to perform HTTP health checks upon. |
| stackgraph.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackgraph.persistence.accessModes | list | `["ReadWriteOnce"]` | Access mode for stackgraph. |
| stackgraph.persistence.enabled | bool | `true` | Enable persistence for HDFS data nodes. |
| stackgraph.persistence.size | string | `"250Gi"` | Size of volume for HDFS data nodes. |
| stackgraph.persistence.storageClass | string | `nil` | Storage class of the volume for HDFS data nodes. |
| stackgraph.resources | object | `{"limits":{"memory":"3Gi"},"requests":{"cpu":"1000m","memory":"2Gi"}}` | Resources to allocate for Stackgraph mono image. |
| stackgraph.securityContext.enabled | bool | `true` | Whether to explicitly set the UID/GID of the pod. |
| stackgraph.securityContext.fsGroup | int | `65534` | UID of the Linux user to use for all pod. |
| stackgraph.securityContext.runAsGroup | int | `65534` | GID of the Linux group to use for all pod. |
| stackgraph.securityContext.runAsUser | int | `65534` | UID of the Linux user to use for all pod. |
| stackgraph.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackgraph.version | string | `"7.7.6"` | Version of stackgraph to use |
| statefulset.antiAffinity.strategy | string | `"soft"` | AntiAffinity strategy to use for all StatefulSets. |
| statefulset.antiAffinity.topologyKey | string | `"kubernetes.io/hostname"` | AntiAffinity topology key to use for all StatefulSets. |
| tephra.affinity | object | `{}` | Affinity settings for pod assignment. |
| tephra.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| tephra.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| tephra.image.pullPolicy | string | `nil` | Pull policy for Tephra pods, defaults to `stackgraph.image.pullPolicy` |
| tephra.image.repository | string | `"stackstate/tephra-server"` | Base container image repository for Tephra pods. |
| tephra.image.tag | string | `nil` | Container image tag for Tephra pods, defaults to `version`-`stackgraph.version` |
| tephra.nodeSelector | object | `{}` | Node labels for pod assignment. |
| tephra.replicaCount | int | `1` | Number of pods for Tephra pods. |
| tephra.resources | object | `{"limits":{"memory":"3Gi"},"requests":{"cpu":"50m","memory":"2Gi"}}` | Resources to allocate for Tephra pods. |
| tephra.securityContext.enabled | bool | `true` | Whether to explicitly set the UID/GID of the pod. |
| tephra.securityContext.runAsGroup | int | `65534` | GID of the Linux group to use for all pod. |
| tephra.securityContext.runAsUser | int | `65534` | UID of the Linux user to use for all pod. |
| tephra.tolerations | list | `[]` | Toleration labels for pod assignment. |
| version | string | `"1.2"` | Version of hbase to use |
| wait.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for wait containers. |
| wait.image.registry | string | `"quay.io"` | Base container image registry for wait containers. |
| wait.image.repository | string | `"stackstate/wait"` | Container image tag for wait containers. |
| wait.image.tag | string | `"1.0.10-025450d9"` |  |
| zookeeper.enabled | bool | `true` | Enable / disable chart-based Zookeeper. |
| zookeeper.externalServers | string | `""` | If `zookeeper.enabled` is set to `false`, use this list of external Zookeeper servers instead. |
| zookeeper.fourlwCommandsWhitelist | string | `"mntr, ruok, stat, srvr"` | Zookeeper four-letter-word (FLW) commands that are enabled. |
| zookeeper.image.registry | string | `"quay.io"` | ZooKeeper image registry |
| zookeeper.image.repository | string | `"stackstate/zookeeper"` | ZooKeeper image repository |
| zookeeper.image.tag | string | `"3.6.3-5e3ee3c0"` | ZooKeeper image tag |
| zookeeper.metrics.enabled | bool | `true` | Enable / disable Zookeeper Prometheus metrics. |
| zookeeper.metrics.serviceMonitor.enabled | bool | `false` | Enable creation of `ServiceMonitor` objects for Prometheus operator. |
| zookeeper.metrics.serviceMonitor.selector | object | `{"release":"prometheus-operator"}` | Default selector to use to target a certain Prometheus instance. |
| zookeeper.replicaCount | int | `1` | Default amount of Zookeeper replicas to provision. |
