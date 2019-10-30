hbase
=====
Helm chart for StackState HBase -- includes Zookeeper, and Hadoop for persistent storage.

Current chart version is `0.1.0`

Source code can be found [here](https://gitlab.com/stackvista/devops/helm-charts.git)

## Chart Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com | zookeeper | 5.1.X |
| https://helm.stackstate.io | common | 0.1.7 |

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| hbase.master.image.pullPolicy | string | `"IfNotPresent"` | Pull policy for HBase masters. |
| hbase.master.image.repository | string | `"quay.io/stackstate/hbase-master"` | Base container image repository for HBase masters. |
| hbase.master.image.tag | string | `"master"` | Default container image tag for HBase masters. |
| hbase.master.replicaCount | int | `2` | Number of pods for HBase masters. |
| hbase.master.resources | object | `{"limits":{"cpu":"1","memory":"3Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}` | Resources to allocate for HBase masters. |
| hbase.regionserver.image.pullPolicy | string | `"IfNotPresent"` | Pull policy for HBase region servers. |
| hbase.regionserver.image.repository | string | `"quay.io/stackstate/hbase-regionserver"` | Base container image repository for HBase region servers. |
| hbase.regionserver.image.tag | string | `"master"` | Default container image tag for HBase region servers. |
| hbase.regionserver.replicaCount | int | `3` | Number of HBase regionserver nodes. |
| hbase.regionserver.resources | object | `{"limits":{"cpu":"1","memory":"3Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}` | Resources to allocate for HBase region servers. |
| hbase.zookeeper.quorum | string | `"hbase"` | Zookeeper quorum used for single-node Zookeeper installations; not used if `zookeeper.replicaCount` is more than `1`. |
| hdfs.datanode.persistence.accessModes | list | `["ReadWriteOnce"]` | Access mode for HDFS data nodes. |
| hdfs.datanode.persistence.enabled | bool | `true` | Enable persistence for HDFS data nodes. |
| hdfs.datanode.persistence.size | string | `"250Gi"` | Size of volume for HDFS data nodes. |
| hdfs.datanode.persistence.storageClass | string | `"default"` | Storage class of the volume for HDFS data nodes. |
| hdfs.datanode.replicaCount | int | `3` | Number of HDFS data nodes. |
| hdfs.datanode.resources | object | `{"limits":{"cpu":"1","memory":"3Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}` | Resources to allocate for HDFS data nodes. |
| hdfs.image.pullPolicy | string | `"IfNotPresent"` | Pull policy for HDFS datanode. |
| hdfs.image.repository | string | `"quay.io/stackstate/hadoop"` | Base container image repository for HDFS datanode. |
| hdfs.image.tag | string | `"hadoop2.9.2-java8"` | Default container image tag for HDFS datanode. |
| hdfs.namenode.persistence.accessModes | list | `["ReadWriteOnce"]` | Access mode for HDFS name nodes. |
| hdfs.namenode.persistence.enabled | bool | `true` | Enable persistence for HDFS name nodes. |
| hdfs.namenode.persistence.size | string | `"20Gi"` | Size of volume for HDFS name nodes. |
| hdfs.namenode.persistence.storageClass | string | `"default"` | Storage class of the volume for HDFS name nodes. |
| hdfs.namenode.resources | object | `{"limits":{"cpu":"1","memory":"3Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}` | Resources to allocate for HDFS name nodes. |
| hdfs.secondarynamenode.persistence.accessModes | list | `["ReadWriteOnce"]` | Access mode for HDFS secondary name nodes. |
| hdfs.secondarynamenode.persistence.enabled | bool | `true` | Enable persistence for HDFS secondary name nodes. |
| hdfs.secondarynamenode.persistence.size | string | `"20Gi"` | Size of volume for HDFS secondary name nodes. |
| hdfs.secondarynamenode.persistence.storageClass | string | `"default"` | Storage class of the volume for HDFS secondary name nodes. |
| hdfs.secondarynamenode.resources | object | `{"limits":{"cpu":"1","memory":"3Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}` | Resources to allocate for HDFS secondary name nodes. |
| securityContext.runAsGroup | int | `65534` | GID of the Linux group to use for all containers. |
| securityContext.runAsUser | int | `65534` | UID of the Linux user to use for all containers. |
| statefulset.antiAffinity.strategy | string | `"soft"` |  |
| statefulset.antiAffinity.topologyKey | string | `"kubernetes.io/hostname"` |  |
| tephra.image.pullPolicy | string | `"IfNotPresent"` | Pull policy for Tephra pods. |
| tephra.image.repository | string | `"quay.io/stackstate/tephra-server"` | Base container image repository for Tephra pods. |
| tephra.image.tag | string | `"master"` | Default container image tag for Tephra pods. |
| tephra.replicaCount | int | `2` | Number of pods for Tephra pods. |
| tephra.resources | object | `{"limits":{"cpu":"1","memory":"3Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}` | Resources to allocate for Tephra pods. |
| zookeeper.enabled | bool | `true` |  |
| zookeeper.fourlwCommandsWhitelist | string | `"mntr, ruok, stat, srvr"` |  |
| zookeeper.metrics.enabled | bool | `true` |  |
| zookeeper.replicaCount | int | `3` |  |
