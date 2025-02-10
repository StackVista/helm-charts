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
| clickhouse.auth.backup.password | string | `"backup"` | Password to "backup" user used by the backup and restore tools. |
| clickhouse.auth.observability.password | string | `"observability"` | Password to "observability" user used by the app and otel collector |
| clickhouse.backup.bucketName | string | `"sts-clickhouse-backup"` | Name of the MinIO bucket where ClickHouse backups are stored. |
| clickhouse.backup.config.keep_remote | int | `308` | How many latest backup should be kept on remote storage, 0 means all uploaded backups will be stored on remote storage. Incremental backups are executed every one 1h so the value 308 = ~14 days. |
| clickhouse.backup.config.tables | string | `"otel.*"` | Create and upload backup only matched with table name patterns, separated by comma, allow ? and * as wildcard. |
| clickhouse.backup.enabled | bool | `false` | Enable scheduled backups of ClickHouse. It requires to be enabled MinIO 'backup.enabled'. |
| clickhouse.backup.image.registry | string | `""` | Registry where to get the image from, the default repository is defined in `global.imageRegistry` |
| clickhouse.backup.image.repository | string | `"stackstate/clickhouse-backup"` | Repository where to get the image from. |
| clickhouse.backup.image.tag | string | `"2.6.5-3c1fce8a"` | Container image tag for 'clickhouse-backup' containers. |
| clickhouse.backup.resources | object | `{"limit":{"cpu":"100m","memory":"250Mi"},"requests":{"cpu":"50m","memory":"250Mi"}}` | Resources of the backup tool. |
| clickhouse.backup.s3Prefix | string | `""` |  |
| clickhouse.backup.scheduled.full_schedule | string | `"45 0 * * *"` | Cron schedule for automatic full backups of ClickHouse. |
| clickhouse.backup.scheduled.incremental_schedule | string | `"45 3-23 * * *"` | Cron schedule for automatic incremental backups of ClickHouse. IMPORTANT: incremental and full backup CAN NOT overlap. |
| clickhouse.enabled | bool | `true` | Enable / disable Clickhouse installation |
| clickhouse.externalZookeeper.servers | list | `["suse-observability-zookeeper-headless"]` | List of zookeeper hosts |
| clickhouse.image.registry | string | `"docker.io"` | Registry where to get the image from, the default repository is defined in `global.imageRegistry` |
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
| deployment.compatibleWithArgoCD | bool | `false` | Whether to adjust the Chart to be compatible with ArgoCD. This feature is as of yet not deployed in the o11y-tenants and saas-tenants directories, so should be considered unfinished (see STAC-21445) |
| fullnameOverride | string | `""` |  |
| global.imageRegistry | string | `"quay.io"` |  |
| minio.accessKey | string | `"setme"` | Secret key for MinIO. Default is set to an invalid value that will cause MinIO to not start up to ensure users of this Helm chart set an explicit value. |
| minio.fullnameOverride | string | `"suse-observability-minio"` |  |
| minio.secretKey | string | `"setme"` |  |
| nameOverride | string | `""` |  |

