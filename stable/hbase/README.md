hbase
=====
Helm chart for StackState HBase -- includes Zookeeper, and Hadoop for persistent storage.

Current chart version is `0.1.15`

Source code can be found [here](https://gitlab.com/stackvista/devops/helm-charts.git)

## Chart Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com | zookeeper | 5.1.1 |
| https://helm.stackstate.io/ | common | 0.1.8 |

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| all.affinity | object | `{}` | Affinity settings for pod assignment on all components. |
| all.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods for all components. |
| all.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object for all components. |
| all.metrics.enabled | bool | `true` | Enable metrics port. |
| all.metrics.servicemonitor.additionalLabels | object | `{}` | Additional labels for targeting Prometheus operator instances. |
| all.metrics.servicemonitor.enabled | bool | `true` | Enable `ServiceMonitor` object; `all.metrics.enabled` *must* be enabled. |
| all.nodeSelector | object | `{}` | Node labels for pod assignment on all components. |
| all.tolerations | list | `[]` | Toleration labels for pod assignment on all components. |
| console.affinity | object | `{}` | Affinity settings for pod assignment. |
| console.enabled | bool | `true` | Enable / disable deployment of the stackgraph-console for debugging. |
| console.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| console.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| console.image.pullPolicy | string | `"Always"` | Pull policy for Tephra pods. |
| console.image.repository | string | `"quay.io/stackstate/stackgraph-console"` | Base container image repository for Tephra pods. |
| console.image.tag | string | `"1.0.0"` | Default container image tag for Tephra pods. |
| console.nodeSelector | object | `{}` | Node labels for pod assignment. |
| console.resources | object | `{}` | Resources to allocate for HDFS data nodes. |
| console.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hbase.master.affinity | object | `{}` | Affinity settings for pod assignment. |
| hbase.master.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hbase.master.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hbase.master.image.pullPolicy | string | `"Always"` | Pull policy for HBase masters. |
| hbase.master.image.repository | string | `"quay.io/stackstate/hbase-master"` | Base container image repository for HBase masters. |
| hbase.master.image.tag | string | `"1.0.0"` | Default container image tag for HBase masters. |
| hbase.master.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hbase.master.replicaCount | int | `1` | Number of pods for HBase masters. |
| hbase.master.resources | object | `{"limits":{"memory":"1Gi"},"requests":{"memory":"1Gi"}}` | Resources to allocate for HBase masters. |
| hbase.master.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hbase.regionserver.affinity | object | `{}` | Affinity settings for pod assignment. |
| hbase.regionserver.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hbase.regionserver.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hbase.regionserver.image.pullPolicy | string | `"Always"` | Pull policy for HBase region servers. |
| hbase.regionserver.image.repository | string | `"quay.io/stackstate/hbase-regionserver"` | Base container image repository for HBase region servers. |
| hbase.regionserver.image.tag | string | `"1.0.0"` | Default container image tag for HBase region servers. |
| hbase.regionserver.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hbase.regionserver.replicaCount | int | `1` | Number of HBase regionserver nodes. |
| hbase.regionserver.resources | object | `{"limits":{"memory":"3Gi"},"requests":{"memory":"2Gi"}}` | Resources to allocate for HBase region servers. |
| hbase.regionserver.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hbase.zookeeper.quorum | string | `"hbase"` | Zookeeper quorum used for single-node Zookeeper installations; not used if `zookeeper.replicaCount` is more than `1`. |
| hdfs.datanode.affinity | object | `{}` | Affinity settings for pod assignment. |
| hdfs.datanode.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hdfs.datanode.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hdfs.datanode.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hdfs.datanode.persistence.accessModes | list | `["ReadWriteOnce"]` | Access mode for HDFS data nodes. |
| hdfs.datanode.persistence.enabled | bool | `true` | Enable persistence for HDFS data nodes. |
| hdfs.datanode.persistence.size | string | `"250Gi"` | Size of volume for HDFS data nodes. |
| hdfs.datanode.persistence.storageClass | string | `"default"` | Storage class of the volume for HDFS data nodes. |
| hdfs.datanode.replicaCount | int | `1` | Number of HDFS data nodes. |
| hdfs.datanode.resources | object | `{"limits":{"memory":"4Gi"},"requests":{"memory":"2Gi"}}` | Resources to allocate for HDFS data nodes. |
| hdfs.datanode.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hdfs.image.pullPolicy | string | `"Always"` | Pull policy for HDFS datanode. |
| hdfs.image.repository | string | `"quay.io/stackstate/hadoop"` | Base container image repository for HDFS datanode. |
| hdfs.image.tag | string | `"2.9.2-java11"` | Default container image tag for HDFS datanode. |
| hdfs.namenode.affinity | object | `{}` | Affinity settings for pod assignment. |
| hdfs.namenode.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hdfs.namenode.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hdfs.namenode.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hdfs.namenode.persistence.accessModes | list | `["ReadWriteOnce"]` | Access mode for HDFS name nodes. |
| hdfs.namenode.persistence.enabled | bool | `true` | Enable persistence for HDFS name nodes. |
| hdfs.namenode.persistence.size | string | `"20Gi"` | Size of volume for HDFS name nodes. |
| hdfs.namenode.persistence.storageClass | string | `"default"` | Storage class of the volume for HDFS name nodes. |
| hdfs.namenode.resources | object | `{"limits":{"memory":"1Gi"},"requests":{"memory":"1Gi"}}` | Resources to allocate for HDFS name nodes. |
| hdfs.namenode.tolerations | list | `[]` | Toleration labels for pod assignment. |
| hdfs.secondarynamenode.affinity | object | `{}` | Affinity settings for pod assignment. |
| hdfs.secondarynamenode.enabled | bool | `false` | Enable / disable secondary name nodes. |
| hdfs.secondarynamenode.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| hdfs.secondarynamenode.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| hdfs.secondarynamenode.nodeSelector | object | `{}` | Node labels for pod assignment. |
| hdfs.secondarynamenode.persistence.accessModes | list | `["ReadWriteOnce"]` | Access mode for HDFS secondary name nodes. |
| hdfs.secondarynamenode.persistence.enabled | bool | `true` | Enable persistence for HDFS secondary name nodes. |
| hdfs.secondarynamenode.persistence.size | string | `"20Gi"` | Size of volume for HDFS secondary name nodes. |
| hdfs.secondarynamenode.persistence.storageClass | string | `"default"` | Storage class of the volume for HDFS secondary name nodes. |
| hdfs.secondarynamenode.resources | object | `{"limits":{"memory":"1Gi"},"requests":{"memory":"1Gi"}}` | Resources to allocate for HDFS secondary name nodes. |
| hdfs.secondarynamenode.tolerations | list | `[]` | Toleration labels for pod assignment. |
| securityContext.runAsGroup | int | `65534` | GID of the Linux group to use for all containers. |
| securityContext.runAsUser | int | `65534` | UID of the Linux user to use for all containers. |
| statefulset.antiAffinity.strategy | string | `"soft"` | AntiAffinity strategy to use for all StatefulSets. |
| statefulset.antiAffinity.topologyKey | string | `"kubernetes.io/hostname"` | AntiAffinity topology key to use for all StatefulSets. |
| tephra.affinity | object | `{}` | Affinity settings for pod assignment. |
| tephra.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| tephra.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| tephra.image.pullPolicy | string | `"Always"` | Pull policy for Tephra pods. |
| tephra.image.repository | string | `"quay.io/stackstate/tephra-server"` | Base container image repository for Tephra pods. |
| tephra.image.tag | string | `"1.0.0"` | Default container image tag for Tephra pods. |
| tephra.nodeSelector | object | `{}` | Node labels for pod assignment. |
| tephra.replicaCount | int | `1` | Number of pods for Tephra pods. |
| tephra.resources | object | `{"limits":{"memory":"3Gi"},"requests":{"memory":"2Gi"}}` | Resources to allocate for Tephra pods. |
| tephra.tolerations | list | `[]` | Toleration labels for pod assignment. |
| zookeeper.enabled | bool | `false` | Enable / disable chart-based Zookeeper. |
| zookeeper.externalServers | string | `""` | If `zookeeper.enabled` is set to `false`, use this list of external Zookeeper servers instead. |
| zookeeper.fourlwCommandsWhitelist | string | `"mntr, ruok, stat, srvr"` | Zookeeper four-letter-word (FLW) commands that are enabled. |
| zookeeper.metrics.enabled | bool | `true` | Enable / disable Zookeeper Prometheus metrics. |
| zookeeper.metrics.serviceMonitor.enabled | bool | `true` | Enable creation of `ServiceMonitor` objects for Prometheus operator. |
| zookeeper.metrics.serviceMonitor.selector | object | `{"release":"prometheus-operator"}` | Default selector to use to target a certain Prometheus instance. |
| zookeeper.replicaCount | int | `1` | Default amount of Zookeeper replicas to provision. |
