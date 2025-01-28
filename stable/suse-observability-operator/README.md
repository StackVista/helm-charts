# suse-observability-operator

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![AppVersion: 7.0.0-snapshot.20250122132112-master-94424f8](https://img.shields.io/badge/AppVersion-7.0.0--snapshot.20250122132112--master--94424f8-informational?style=flat-square)

Helm chart to install SUSE Observability using Operators

**Homepage:** <https://gitlab.com/stackvista/stackstate.git>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| SUSE | <ops@stackstate.com> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| clickhouse.auth.observability.password | string | `"observability"` | Password to "observability" user used by the app and otel collector |
| clickhouse.enabled | bool | `true` | Enable / disable Clickhouse installation |
| clickhouse.externalZookeeper.servers | list | `["suse-observability-zookeeper-headless"]` | List of zookeeper hosts |
| clickhouse.image.registry | string | `""` | Registry where to get the image from, the default repository is defined in `global.imageRegistry` |
| clickhouse.image.repository | string | `"clickhouse/clickhouse-server"` | Repository where to get the image from. |
| clickhouse.image.tag | string | `"24.8.12.28-alpine"` | Container image tag for 'clickhouse' containers. |
| clickhouse.persistence.data.size | string | `"50Gi"` | Size of persistent volume for ClickHouse data |
| clickhouse.persistence.data.storageClassName | string | `""` | Storage Class used by CH data. It should support expansion if you decide to expand it |
| clickhouse.replicaCount | int | `3` | Number of ClickHouse replicas per shard to deploy |
| clickhouse.resources.limits.cpu | string | `"1000m"` |  |
| clickhouse.resources.limits.memory | string | `"4Gi"` |  |
| clickhouse.resources.requests.cpu | string | `"500m"` |  |
| clickhouse.resources.requests.memory | string | `"4Gi"` |  |
| clickhouse.shards | int | `1` | Number of ClickHouse shards to deploy |
| fullnameOverride | string | `""` |  |
| global.imageRegistry | string | `"quay.io"` |  |
| nameOverride | string | `""` |  |

