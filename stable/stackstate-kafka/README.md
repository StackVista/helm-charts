stackstate-kafka
================
StackState Kafka Helm chart that can pre-create topics as well

Current chart version is `0.1.0`

Source code can be found [here](https://gitlab.com/stackvista/stackstate.git)

## Chart Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | kafka | 7.2.9 |
| https://helm.stackstate.io | common | 0.4.1 |

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| kafka.enabled | bool | `true` |  |
| kafka.externalZookeeper.servers | string | `"distributed-zookeeper-headless"` |  |
| kafka.fullnameOverride | string | `"distributed-kafka"` |  |
| kafka.zookeeper.enabled | bool | `false` |  |
| zookeeper.fullnameOverride | string | `"distributed-zookeeper"` |  |
