# suse-observability

Helm chart for SUSE Observability

Current chart version is `2.7.1-pre.87`

**Homepage:** <https://gitlab.com/stackvista/stackstate.git>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../clickhouse/ | clickhouse | 3.6.9-suse-observability.11 |
| file://../common/ | common | * |
| file://../elasticsearch/ | elasticsearch | 8.19.4-stackstate.8 |
| file://../hbase/ | hbase | 0.2.115 |
| file://../kafka/ | kafka | 19.1.3-suse-observability.10 |
| file://../kafkaup-operator/ | kafkaup-operator | 0.1.16 |
| file://../minio/ | minio | 8.0.10-stackstate.18 |
| file://../opentelemetry-collector | opentelemetry-collector | 0.108.0-stackstate.19 |
| file://../pull-secret/ | pull-secret | * |
| file://../suse-observability-sizing/ | suse-observability-sizing | 0.1.4 |
| file://../victoria-metrics-single/ | victoria-metrics-0(victoria-metrics-single) | 0.8.53-stackstate.36 |
| file://../victoria-metrics-single/ | victoria-metrics-1(victoria-metrics-single) | 0.8.53-stackstate.36 |
| file://../zookeeper/ | zookeeper | 8.1.2-suse-observability.8 |
| https://helm.stackstate.io | anomaly-detection | 5.2.0-snapshot.177 |
| https://helm.stackstate.io | kubernetes-rbac-agent | 0.0.24 |

## Required Values

In order to successfully install this chart, you **must** provide the following variables:
* `stackstate.license.key`
* `stackstate.baseUrl`

Install them on the command line on Helm with the following command:

```shell
helm install \
--set stackstate.license.key=<your-license-key> \
--set stackstate.baseUrl=<your-base-url> \
stackstate/stackstate
```

## Simplified Sizing Configuration

SUSE Observability now provides built-in sizing profiles that automatically configure all component resources, replica counts, storage sizes, and deployment modes with a single configuration value.

### Quick Start with Sizing Profiles

```yaml
# values.yaml
global:
  # Required: Set image registry for SUSE Observability images
  imageRegistry: "registry.rancher.com"
  suseObservability:
    sizing:
      profile: "150-ha"  # Single value configures everything!
    license: "<your-license-key>"
    baseUrl: "<your-base-url>"
    adminPassword: "<bcrypt-hashed-password>"
    pullSecret:
      username: "<registry-username>"
      password: "<registry-password>"

```

```shell
helm install suse-observability . -f values.yaml
```

#### Generating a bcrypt Password Hash

The `adminPassword` must be a bcrypt-hashed password. Generate one using either method:

```shell
# Using htpasswd (commonly available on Linux/macOS)
htpasswd -bnBC 10 "" "your-password" | tr -d ':\n'

# Using Python with bcrypt library
python3 -c "import bcrypt; print(bcrypt.hashpw(b'your-password', bcrypt.gensalt(10)).decode())"

# Using Docker
docker run --rm httpd:alpine htpasswd -bnBC 10 "" "your-password" | tr -d ':\n'
```

### Available Sizing Profiles

| Profile | Use Case | HA Mode | Components | VM Instances | Server Split |
|---------|----------|---------|------------|--------------|--------------|
| `trial` | Development/Testing | No | ~10 | 1 | No |
| `10-nonha` | Small non-HA | No | ~10 | 1 | No |
| `20-nonha` | Small non-HA | No | ~20 | 1 | No |
| `50-nonha` | Medium non-HA | No | ~50 | 1 | No |
| `100-nonha` | Large non-HA | No | ~100 | 1 | No |
| `150-ha` | Production HA | Yes | ~150 | 2 | Yes |
| `250-ha` | Production HA | Yes | ~250 | 2 | Yes |
| `500-ha` | Production HA | Yes | ~500 | 2 | Yes |
| `4000-ha` | Enterprise HA | Yes | ~4000 | 2 | Yes |

### What Gets Configured Automatically

A single sizing profile automatically configures:

**Infrastructure Components:**
- **ClickHouse**: Replicas, CPU/memory resources, storage size
- **Elasticsearch**: Replicas, CPU/memory resources, storage size
- **HBase**: Deployment mode (Mono/Distributed), resources for master/regionserver/datanode/namenode/tephra
- **Kafka**: Replicas, CPU/memory resources, storage size, partition counts
- **Zookeeper**: Replicas, CPU/memory resources
- **Victoria Metrics 0**: CPU/memory resources, storage size, retention period
- **Victoria Metrics 1**: Enablement (HA only), CPU/memory resources, storage size

**SUSE Observability Components:**
- **API/Server**: Split mode (HA profiles), replica counts, CPU/memory resources
- **Receiver**: Split mode (base/logs/process-agent for HA), replica counts, resources
- **Checks, Correlate, State, Sync, Health-Sync**: Replica counts, CPU/memory resources
- **UI, Notification, Slicing**: Replica counts, CPU/memory resources

**Supporting Services:**
- **Minio**: CPU/memory resources
- **KafkaUp Operator**: CPU/memory resources
- **Prometheus Elasticsearch Exporter**: CPU/memory resources

### Migrating from suse-observability-values Chart

> **⚠️ DEPRECATION NOTICE**
> The `suse-observability-values` chart is deprecated. Use the built-in sizing profiles instead.

**Old workflow (DEPRECATED - Two steps):**

```shell
# Step 1: Generate values file with suse-observability-values chart
helm template suse-observability-values \
  --set sizing.profile=150-ha \
  --set license=<your-license-key> \
  --set baseUrl=<your-base-url> \
  --set pullSecret.username=<username> \
  --set pullSecret.password=<password> \
  suse-observability/suse-observability-values > generated-values.yaml

# Step 2: Install suse-observability chart with generated values
helm install suse-observability . -f generated-values.yaml
```

**New workflow (Recommended - Single step):**

```yaml
# values.yaml
global:
  # Required: Set image registry for SUSE Observability images
  imageRegistry: "registry.rancher.com"
  suseObservability:
    sizing:
      profile: "150-ha"  # Replaces the entire values generation step!
    license: "<your-license-key>"
    baseUrl: "<your-base-url>"
    adminPassword: "<bcrypt-hashed-password>"
    pullSecret:
      username: "<username>"
      password: "<password>"

```

```shell
helm install suse-observability . -f values.yaml
```

**Migration checklist:**

1. Identify your current sizing profile (e.g., `150-ha`)
2. Create new values file with `global.suseObservability.sizing.profile`
3. Set `global.imageRegistry: "registry.rancher.com"` for SUSE Observability images
4. Move credentials to `global.suseObservability.*` section
5. Remove `helm template suse-observability-values` step from deployment scripts
6. Test installation in non-production environment first

### Upgrading Existing Deployments

If you have an existing SUSE Observability installation using the old `suse-observability-values` chart workflow, follow these steps:

**Before upgrading:**

1. **Backup your current values**: Save your existing generated values file and any custom overrides
   ```shell
   kubectl get configmap -n <namespace> -o yaml > backup-configmaps.yaml
   helm get values suse-observability -n <namespace> > current-values.yaml
   ```

2. **Identify your sizing profile**: Check your current `suse-observability-values` configuration to find the profile name (e.g., `150-ha`)

3. **Review resource differences**: The new profiles may have updated resource recommendations. Compare your current resources with the new profile defaults if you have custom overrides

**Upgrade procedure:**

```shell
# 1. Create your new values file with global.suseObservability configuration
#    (see Quick Start example above)

# 2. Perform helm upgrade with the new values
helm upgrade suse-observability suse-observability/suse-observability \
  -n <namespace> \
  -f new-values.yaml

# 3. Verify the upgrade
kubectl get pods -n <namespace>
helm get values suse-observability -n <namespace>
```

**Important considerations:**

- **No downtime expected**: The upgrade is a standard Helm upgrade; pods will be rolled incrementally
- **PVCs are preserved**: Existing persistent volume claims remain intact
- **Secrets are preserved**: Existing secrets (licenses, API keys) are not deleted
- **Rollback available**: Use `helm rollback suse-observability <revision>` if needed

### Overriding Sizing Profile Defaults

You can override specific values from the sizing profile when needed:

```yaml
global:
  suseObservability:
    sizing:
      profile: "150-ha"

# Override specific component resources
stackstate:
  components:
    api:
      resources:
        requests:
          memory: 16Gi  # Override profile's default of 12Gi

# Override storage sizes
elasticsearch:
  volumeClaimTemplate:
    resources:
      requests:
        storage: 500Gi  # Override profile's default
```

### Global Affinity Configuration

The `global.suseObservability.affinity` section allows you to configure pod scheduling constraints for all components:

```yaml
global:
  suseObservability:
    sizing:
      profile: "150-ha"
    affinity:
      # Node affinity - target specific nodes (applies to ALL components)
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
                - key: node-role.kubernetes.io/observability
                  operator: Exists

      # Pod affinity - co-locate application pods (does NOT apply to infrastructure)
      podAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/part-of: suse-observability
              topologyKey: kubernetes.io/hostname

      # Pod anti-affinity - spread infrastructure pods (HA profiles only)
      podAntiAffinity:
        # Use hard anti-affinity (pods MUST be on different nodes)
        requiredDuringSchedulingIgnoredDuringExecution: true
        # Spread across nodes (use topology.kubernetes.io/zone for zone spreading)
        topologyKey: "kubernetes.io/hostname"
```

**Affinity scope:**

| Affinity Type | Application Components | Infrastructure Components |
|---------------|------------------------|---------------------------|
| `nodeAffinity` | Yes | Yes |
| `podAffinity` | Yes | No |
| `podAntiAffinity` | No | Yes (HA profiles) |

**Pod anti-affinity modes:**

- `requiredDuringSchedulingIgnoredDuringExecution: true` - **Hard anti-affinity**: Pods will NOT schedule if they cannot be placed on separate nodes. Use when you have enough nodes.
- `requiredDuringSchedulingIgnoredDuringExecution: false` - **Soft anti-affinity**: Pods will prefer separate nodes but can co-locate if necessary. Use when node count is limited.

### Backward Compatibility

The traditional configuration method (without sizing profiles) is still supported for backward compatibility:

```yaml
# Traditional method - still works
stackstate:
  license:
    key: "<your-license-key>"
  baseUrl: "<your-base-url>"
  authentication:
    adminPassword: "<password>"
  components:
    api:
      resources:
        requests:
          cpu: "4000m"
          memory: 12Gi
    # ... manual configuration for each component
```

However, we strongly recommend migrating to sizing profiles for easier maintenance and upgrades.

### Troubleshooting Migration Issues

**Problem: Pods stuck in Pending state after migration**

This usually indicates a scheduling issue, often related to anti-affinity rules:

```shell
# Check pod events
kubectl describe pod <pod-name> -n <namespace>

# If anti-affinity is the issue, use soft anti-affinity
global:
  suseObservability:
    affinity:
      podAntiAffinity:
        requiredDuringSchedulingIgnoredDuringExecution: false  # Soft anti-affinity
```

**Problem: Resources differ from previous installation**

Sizing profiles may have updated resource recommendations. To preserve your previous settings:

```yaml
# Override specific resources while using the profile for everything else
global:
  suseObservability:
    sizing:
      profile: "150-ha"

# Your custom overrides
stackstate:
  components:
    api:
      resources:
        requests:
          memory: 8Gi  # Your previous value
```

**Problem: helm upgrade fails with validation errors**

Ensure you're not mixing old and new configuration styles:

```shell
# Check current values
helm get values suse-observability -n <namespace>

# Common issue: both stackstate.license.key AND global.suseObservability.license set
# Solution: Use only one configuration style
```

**Getting help:**

If you encounter issues not covered here:
1. Check pod logs: `kubectl logs <pod-name> -n <namespace>`
2. Review Helm release status: `helm status suse-observability -n <namespace>`
3. Compare rendered templates: `helm template suse-observability . -f values.yaml > rendered.yaml`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| anomaly-detection.cpu.limit | string | `"2000m"` | CPU resource limit |
| anomaly-detection.cpu.request | string | `"1000m"` | CPU resource request |
| anomaly-detection.enabled | bool | `false` | Enables anomaly detection chart |
| anomaly-detection.image.imagePullPolicy | string | `"IfNotPresent"` | The default pullPolicy used for anomaly detection pods. |
| anomaly-detection.image.pullSecretName | string | `nil` | Name of ImagePullSecret to use for all pods. |
| anomaly-detection.image.pullSecretPassword | string | `nil` |  |
| anomaly-detection.image.pullSecretUsername | string | `nil` | Password used to login to the registry to pull Docker images of all pods. |
| anomaly-detection.image.registry | string | `"quay.io"` | Base container image registry for all containers, except for the wait container |
| anomaly-detection.image.spotlightRepository | string | `"stackstate/spotlight"` | Repository of the spotlight Docker image. |
| anomaly-detection.ingress | object | `{"annotations":{},"enabled":false,"hostname":null,"hosts":[],"port":8090,"tls":null}` | Status interface ingress |
| anomaly-detection.ingress.enabled | bool | `false` | Enables ingress controller for status interface |
| anomaly-detection.ingress.hostname | string | `nil` | Status interface hostname e.g. spotlight.local.domain |
| anomaly-detection.memory.limit | string | `"3Gi"` | Memory resource limit |
| anomaly-detection.memory.request | string | `"3Gi"` | Memory resource request |
| anomaly-detection.pdb.maxUnavailable | int | `0` | Maximum number of pods that can be unavailable for the anomaly detection |
| anomaly-detection.stackstate.apiToken | string | `nil` | Stackstate Api token that used by spotlight for authentication, it is expected to be set only in case if authType = "api-token" |
| anomaly-detection.stackstate.authType | string | `"token"` | Type of authentication. There are three options 1) "token" - with service account token (default), 2) "api-token" - with Stackstate API Token, 3) "cookie" - username, password based auth. |
| anomaly-detection.stackstate.instance | string | `"{{ include \"stackstate.router.endpoint\" . }}"` | **Required Stackstate instance URL, e.g http://stackstate-router:8080 |
| anomaly-detection.stackstate.password | string | `nil` | Stackstate Password used by spotlight for authentication, it is expected to be set only in case if authType = "cookie" |
| anomaly-detection.stackstate.username | string | `nil` | Stackstate Username used by spotlight for authentication, it is expected to be set only in case if authType = "cookie" |
| anomaly-detection.threadWorkers | int | `3` | The number of worker threads. |
| backup.additionalLogging | string | `""` | Additional logback config for backup components |
| backup.configuration.bucketName | string | `"sts-configuration-backup"` | Name of the MinIO bucket to store configuration backups. |
| backup.configuration.maxLocalFiles | int | `10` | The maximum number of configuration backup files stored on the PVC for the configuration backup (which is only of limited size, see backup.configuration.scheduled.pvc.size. |
| backup.configuration.s3Prefix | string | `""` | Prefix (dir name) used to store backup files. |
| backup.configuration.scheduled.backupDatetimeParseFormat | string | `"%Y%m%d-%H%M"` | Format to parse date/time from configuration backup name. *Note:* This should match the value for `backupNameTemplate`. |
| backup.configuration.scheduled.backupNameParseRegexp | string | `"sts-backup-([0-9]*-[0-9]*).sty"` | Regular expression to retrieve date/time from configuration backup name. *Note:* This should match the value for `backupNameTemplate`. |
| backup.configuration.scheduled.backupNameTemplate | string | `"sts-backup-$(date +%Y%m%d-%H%M).sty"` | Template for the configuration backup name as a double-quoted shell string value. |
| backup.configuration.scheduled.backupRetentionTimeDelta | string | `"365 days ago"` | Time to keep configuration backups. The value is passed to GNU date tool to determine a specific date, and files older than this date will be deleted. |
| backup.configuration.scheduled.pvc.accessModes | list | `["ReadWriteOnce"]` | Access mode for settings backup data. |
| backup.configuration.scheduled.pvc.size | string | `"1Gi"` | Size of volume for settings backup |
| backup.configuration.scheduled.pvc.storageClass | string | `nil` | Storage class of the volume for settings backup data. |
| backup.configuration.scheduled.schedule | string | `"0 4 * * *"` | Cron schedule for automatic configuration backups in [Kubernetes cron schedule syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#cron-schedule-syntax). |
| backup.configuration.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| backup.configuration.securityContext.fsGroup | int | `65534` | The GID (group ID) of all files on all mounted volumes |
| backup.configuration.securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| backup.configuration.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| backup.configuration.securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| backup.configuration.yaml.maxSizeLimit | string | `"100Mi"` | Max size of the settings backup or installed via a stackpack |
| backup.elasticsearch.bucketName | string | `"sts-elasticsearch-backup"` | Name of the MinIO bucket where ElasticSearch snapshots are stored. |
| backup.elasticsearch.restore.scaleDownLabels | object | `{"observability.suse.com/scalable-during-es-restore":"true"}` | Labels used to identify deployments that should be scaled down during Elasticsearch restore procedure. |
| backup.elasticsearch.s3Prefix | string | `""` |  |
| backup.elasticsearch.scheduled.indices | string | `"sts*"` | ElasticSearch indices to snapshot in [JSON array format](https://www.w3schools.com/js/js_json_arrays.asp). |
| backup.elasticsearch.scheduled.schedule | string | `"0 0 3 * * ?"` | Cron schedule for automatic ElasticSearch snaphosts in [ElasticSearch cron schedule syntax](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/cron-expressions.html). |
| backup.elasticsearch.scheduled.snapshotNameTemplate | string | `"<sts-backup-{now{yyyyMMdd-HHmm}}>"` | Template for the ElasticSearch snapshot name in [ElasticSearch date math format](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/date-math-index-names.html). |
| backup.elasticsearch.scheduled.snapshotPolicyName | string | `"auto-sts-backup"` | Name of the ElasticSearch snapshot policy. |
| backup.elasticsearch.scheduled.snapshotRetentionExpireAfter | string | `"30d"` | Amount of time to keep ElasticSearch snapshots in [ElasticSearch time units](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/common-options.html#time-units). *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). |
| backup.elasticsearch.scheduled.snapshotRetentionMaxCount | string | `"30"` | Minimum number of ElasticSearch snapshots to keep. *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). |
| backup.elasticsearch.scheduled.snapshotRetentionMinCount | string | `"5"` | Minimum number of ElasticSearch snapshots to keep. *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). |
| backup.elasticsearch.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| backup.elasticsearch.securityContext.fsGroup | int | `65534` | The GID (group ID) of all files on all mounted volumes |
| backup.elasticsearch.securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| backup.elasticsearch.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| backup.elasticsearch.securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| backup.elasticsearch.snapshotRepositoryName | string | `"sts-backup"` | Name of the ElasticSearch snapshot repository. backup.elasticsearch.s3Prefix -- Prefix (dir name) used to store backup files. |
| backup.enabled | bool | `false` | Enables backup/restore, including the MinIO subsystem. |
| backup.initJobAnnotations | object | `{}` | Annotations for Backup-init Job. |
| backup.poddisruptionbudget.maxUnavailable | int | `0` | Maximum number of pods that can be unavailable during the backup. |
| backup.stackGraph.bucketName | string | `"sts-stackgraph-backup"` | Name of the MinIO bucket to store StackGraph backups. |
| backup.stackGraph.restore.tempData.accessModes[0] | string | `"ReadWriteOnce"` |  |
| backup.stackGraph.restore.tempData.size | string | `nil` |  |
| backup.stackGraph.restore.tempData.storageClass | string | `nil` |  |
| backup.stackGraph.s3Prefix | string | `""` | Prefix (dir name) used to store backup files. |
| backup.stackGraph.scheduled.backupDatetimeParseFormat | string | `"%Y%m%d-%H%M"` | Format to parse date/time from StackGraph backup name. *Note:* This should match the value for `backupNameTemplate`. |
| backup.stackGraph.scheduled.backupNameParseRegexp | string | `"sts-backup-([0-9]*-[0-9]*).graph"` | Regular expression to retrieve date/time from StackGraph backup name. *Note:* This should match the value for `backupNameTemplate`. |
| backup.stackGraph.scheduled.backupNameTemplate | string | `"sts-backup-$(date +%Y%m%d-%H%M).graph"` | Template for the StackGraph backup name as a double-quoted shell string value. |
| backup.stackGraph.scheduled.backupRetentionTimeDelta | string | `"30 days ago"` | Time to keep StackGraph backups in. The value is passed to GNU date tool  to determine a specific date, and files older than this date will be deleted. |
| backup.stackGraph.scheduled.schedule | string | `"0 3 * * *"` | Cron schedule for automatic StackGraph backups in [Kubernetes cron schedule syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#cron-schedule-syntax). |
| backup.stackGraph.scheduled.tempData.accessModes[0] | string | `"ReadWriteOnce"` |  |
| backup.stackGraph.scheduled.tempData.size | string | `nil` |  |
| backup.stackGraph.scheduled.tempData.storageClass | string | `nil` |  |
| backup.stackGraph.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| backup.stackGraph.securityContext.fsGroup | int | `65534` | The GID (group ID) of all files on all mounted volumes |
| backup.stackGraph.securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| backup.stackGraph.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| backup.stackGraph.securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| backup.stackGraph.splitArchiveSize | int | `0` | Split the Stackgraph dump into chunks of the specified size in bytes. Accepts an integer greater or equal to 0 with optional suffix K,M,G (powers of 1024) or KB,MB,GB (powers of 1000) If set to 0, the dump is not split. |
| clickhouse.auth.password | string | `"admin"` | ClickHouse Admin password. If left empty the random value is generated. |
| clickhouse.auth.username | string | `"admin"` | ClickHouse Admin username |
| clickhouse.backup.affinity | object | `{}` | Affinity settings for pod assignment. |
| clickhouse.backup.bucketName | string | `"sts-clickhouse-backup"` | Name of the MinIO bucket where ClickHouse backups are stored. |
| clickhouse.backup.config.keep_remote | int | `308` | How many latest backup should be kept on remote storage, 0 means all uploaded backups will be stored on remote storage. Incremental backups are executed every one 1h so the value 308 = ~14 days. |
| clickhouse.backup.config.tables | string | `"otel.*"` | Create and upload backup only matched with table name patterns, separated by comma, allow ? and * as wildcard. |
| clickhouse.backup.image.registry | string | `"quay.io"` | Registry where to get the image from. |
| clickhouse.backup.image.repository | string | `"stackstate/clickhouse-backup"` | Repository where to get the image from. |
| clickhouse.backup.image.tag | string | `"2.6.39-d0d7ba46-65"` | Container image tag for 'clickhouse' backup containers. |
| clickhouse.backup.nodeSelector | object | `{}` | Node labels for pod assignment. |
| clickhouse.backup.podAnnotations | object | `{}` | Extra annotations for ClickHouse backup pods. |
| clickhouse.backup.podLabels | object | `{}` | Extra labels for ClickHouse backup pods. |
| clickhouse.backup.resources | object | `{"limit":{"cpu":"100m","memory":"250Mi"},"requests":{"cpu":"50m","memory":"250Mi"}}` | Resources of the backup tool. |
| clickhouse.backup.s3Prefix | string | `""` |  |
| clickhouse.backup.scheduled.full_schedule | string | `"45 0 * * *"` | Cron schedule for automatic full backups of ClickHouse. |
| clickhouse.backup.scheduled.incremental_schedule | string | `"45 3-23 * * *"` | Cron schedule for automatic incremental backups of ClickHouse. IMPORTANT: incremental and full backup CAN NOT overlap. |
| clickhouse.backup.tolerations | list | `[]` | Toleration labels for pod assignment. |
| clickhouse.enabled | bool | `true` | Enable / disable chart-based Clickhouse. |
| clickhouse.externalZookeeper.port | int | `2181` |  |
| clickhouse.externalZookeeper.servers | list | `["suse-observability-zookeeper-headless"]` | External Zookeeper configuration. |
| clickhouse.extraOverrides | string | `"<clickhouse>\n  <!-- Recommended settings for low memory systems https://clickhouse.com/docs/operations/tips#ram -->\n  <mark_cache_size>1073741824</mark_cache_size>\n  <concurrent_threads_soft_limit_num>1</concurrent_threads_soft_limit_num>\n\n  <profiles>\n    <default>\n      <!-- Recommended settings for low memory systems https://clickhouse.com/docs/operations/tips#ram -->\n      <max_block_size>8192</max_block_size>\n      <max_download_threads>1</max_download_threads>\n      <input_format_parallel_parsing>0</input_format_parallel_parsing>\n      <output_format_parallel_formatting>0</output_format_parallel_formatting>\n    </default>\n  </profiles>\n\n  <!-- Disable unused logs to avoid filling up disks -->\n  <!-- For more details see https://kb.altinity.com/altinity-kb-setup-and-maintenance/altinity-kb-system-tables-eat-my-disk/ -->\n  <asynchronous_metric_log remove=\"1\"/>\n  <backup_log remove=\"1\"/>\n  <error_log remove=\"1\"/>\n  <metric_log remove=\"1\"/>\n  <query_metric_log remove=\"1\" />\n  <query_views_log remove=\"1\" />\n  <part_log remove=\"1\"/>\n  <session_log remove=\"1\"/>\n  <text_log remove=\"1\" />\n  <trace_log remove=\"1\"/>\n  <crash_log remove=\"1\"/>\n  <opentelemetry_span_log remove=\"1\"/>\n  <zookeeper_log remove=\"1\"/>\n  <processors_profile_log remove=\"1\"/>\n\n  <!-- keeping these for debugging purposes, but configuring TTL -->\n  <query_thread_log replace=\"1\">\n    <database>system</database>\n    <table>query_thread_log</table>\n    <engine>ENGINE = MergeTree PARTITION BY (event_date)\n      ORDER BY (event_time)\n      TTL event_date + INTERVAL 7 DAY DELETE\n    </engine>\n  </query_thread_log>\n\n  <query_log replace=\"1\">\n    <database>system</database>\n    <table>query_log</table>\n    <engine>ENGINE = MergeTree PARTITION BY (event_date)\n      ORDER BY (event_time)\n      TTL event_date + INTERVAL 7 DAY DELETE\n    </engine>\n  </query_log>\n\n  <aggregated_zookeeper_log replace=\"1\">\n    <database>system</database>\n    <table>aggregated_zookeeper_log</table>\n    <engine>ENGINE = MergeTree PARTITION BY (event_date)\n      ORDER BY (event_date, event_time)\n      TTL event_date + INTERVAL 3 DAY DELETE\n    </engine>\n  </aggregated_zookeeper_log>\n\n  {{- $effectiveReplicaCount := include \"common.sizing.clickhouse.effectiveReplicaCount\" . | int -}}\n  <!-- Cluster configuration - Any update of the shards and replicas requires helm upgrade -->\n  <remote_servers>\n    <default>\n      {{- $shards := $.Values.shards | int }}\n      {{- range $shard, $e := until $shards }}\n      <shard>\n          {{- range $i, $_e := until $effectiveReplicaCount }}\n          <replica>\n              <host>{{ printf \"%s-shard%d-%d.%s.%s.svc.%s\" (include \"common.names.fullname\" $ ) $shard $i (include \"clickhouse.headlessServiceName\" $) (include \"common.names.namespace\" $) $.Values.clusterDomain }}</host>\n              <port>{{ $.Values.service.ports.tcp }}</port>\n              <user from_env=\"CLICKHOUSE_ADMIN_USER\"></user>\n              <password from_env=\"CLICKHOUSE_ADMIN_PASSWORD\"></password>\n          </replica>\n          {{- end }}\n      </shard>\n      {{- end }}\n    </default>\n  </remote_servers>\n</clickhouse>\n"` | Extra configuration overrides (evaluated as a template) apart from the default. This configuration deploys ClickHouse in the cluster mode even if there is only one node. |
| clickhouse.extraVolumeMounts | list | `[{"mountPath":"/app/post_restore.sh","name":"clickhouse-backup-scripts","subPath":"post_restore.sh"}]` | extra VolumeMounts for the ClickHouse container |
| clickhouse.extraVolumes | list | `[{"configMap":{"name":"suse-observability-clickhouse-backup"},"name":"clickhouse-backup-config"},{"configMap":{"defaultMode":360,"name":"suse-observability-clickhouse-backup"},"name":"clickhouse-backup-scripts"}]` | extra volumes for ClickHouse Pods |
| clickhouse.fullnameOverride | string | `"suse-observability-clickhouse"` | Name override for clickhouse child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| clickhouse.image.registry | string | `"quay.io"` | Registry where to get the image from |
| clickhouse.image.repository | string | `"stackstate/clickhouse"` | Repository where to get the image from. |
| clickhouse.image.tag | string | `"25.9.5-85835861-137"` | Container image tag for 'clickhouse' containers. |
| clickhouse.metrics.enabled | bool | `true` |  |
| clickhouse.persistence.size | string | `"50Gi"` | Size of persistent volume for each clickhouse pod |
| clickhouse.podAnnotations."ad.stackstate.com/backup.check_names" | string | `"[\"openmetrics\"]"` |  |
| clickhouse.podAnnotations."ad.stackstate.com/backup.init_configs" | string | `"[{}]"` |  |
| clickhouse.podAnnotations."ad.stackstate.com/backup.instances" | string | `"[ { \"prometheus_url\": \"http://%%host%%:7171/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"clickhouse_backup_*\"] } ]"` |  |
| clickhouse.podAnnotations."ad.stackstate.com/clickhouse.check_names" | string | `"[\"openmetrics\"]"` |  |
| clickhouse.podAnnotations."ad.stackstate.com/clickhouse.init_configs" | string | `"[{}]"` |  |
| clickhouse.podAnnotations."ad.stackstate.com/clickhouse.instances" | string | `"[ { \"prometheus_url\": \"http://%%host%%:8001/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"ClickHouseAsyncMetrics_*\", \"ClickHouseMetrics_*\", \"ClickHouseProfileEvents_*\"] } ]"` |  |
| clickhouse.podAnnotations.checksum/stackstate-backup-config | string | `"{{ toJson .Values.backup | sha256sum }}"` |  |
| clickhouse.replicaCount | string | `nil` | Number of ClickHouse replicas per shard to deploy. When using global.suseObservability.sizing.profile, this value is determined by the sizing profile (1 for most profiles, 3 for 4000-ha). |
| clickhouse.resources.limits.cpu | string | `"1000m"` |  |
| clickhouse.resources.limits.memory | string | `"4Gi"` |  |
| clickhouse.resources.requests.cpu | string | `"500m"` |  |
| clickhouse.resources.requests.memory | string | `"4Gi"` |  |
| clickhouse.shards | int | `1` | Number of ClickHouse shards to deploy |
| clickhouse.sidecars | list | `[{"command":["/app/entrypoint.sh"],"env":[{"name":"BACKUP_CLICKHOUSE_ENABLED","valueFrom":{"configMapKeyRef":{"key":"backup_enabled","name":"suse-observability-clickhouse-backup"}}},{"name":"BACKUP_TABLES","value":"{{ .Values.backup.config.tables }}"},{"name":"CLICKHOUSE_REPLICA_ID","valueFrom":{"fieldRef":{"apiVersion":"v1","fieldPath":"metadata.name"}}}],"image":"{{ default .Values.backup.image.registry .Values.global.imageRegistry }}/{{ .Values.backup.image.repository }}:{{ .Values.backup.image.tag }}","imagePullPolicy":"IfNotPresent","name":"backup","ports":[{"containerPort":9746,"name":"supercronic"},{"containerPort":7171,"name":"backup-api"}],"resources":{"limits":{"cpu":"{{ .Values.backup.resources.limit.cpu }}","memory":"{{ .Values.backup.resources.limit.memory }}"},"requests":{"cpu":"{{ .Values.backup.resources.requests.cpu }}","memory":"{{ .Values.backup.resources.requests.memory }}"}},"securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"runAsNonRoot":true,"runAsUser":1001,"seccompProfile":{"type":"RuntimeDefault"}},"volumeMounts":[{"mountPath":"/bitnami/clickhouse","name":"data"},{"mountPath":"/bitnami/clickhouse/etc/conf.d/default","name":"config"},{"mountPath":"/bitnami/clickhouse/etc/conf.d/extra-configmap","name":"extra-config"},{"mountPath":"/bitnami/clickhouse/etc/users.d/users-extra-configmap","name":"users-extra-config"},{"mountPath":"/etc/clickhouse-backup.yaml","name":"clickhouse-backup-config","subPath":"config.yaml"},{"mountPath":"/app/entrypoint.sh","name":"clickhouse-backup-scripts","subPath":"entrypoint.sh"}]}]` | sidecar containers to run backups |
| clickhouse.usersExtraOverrides | string | `"<clickhouse>\n  <users>\n    <stackstate>\n        <no_password></no_password>\n        <grants>\n            <query>GRANT ALL ON *.*</query>\n        </grants>\n    </stackstate>\n  </users>\n</clickhouse>\n"` | Users extra configuration overrides. |
| clickhouse.volumePermissions.enabled | bool | `false` |  |
| clickhouse.zookeeper.enabled | bool | `false` | Disable Zookeeper from the clickhouse chart **Don't change unless otherwise specified**. |
| cluster-role.enabled | bool | `true` | Deploy the ClusterRole(s) and ClusterRoleBinding(s) together with the chart. Can be disabled if these need to be installed by an administrator of the Kubernetes cluster. |
| commonLabels | object | `{}` | Labels that will be added to all resources created by the stackstate chart (not the subcharts though) |
| deployment.compatibleWithArgoCD | bool | `false` | Whether to adjust the Chart to be compatible with ArgoCD. This feature is as of yet not deployed in the o11y-tenants and saas-tenants directories, so should be considered unfinished (see STAC-21445) |
| elasticsearch.clusterHealthCheckParams | string | `"wait_for_status=yellow&timeout=1s"` | The Elasticsearch cluster health status params that will be used by readinessProbe command |
| elasticsearch.clusterName | string | `"suse-observability-elasticsearch"` | Name override for Elasticsearch child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| elasticsearch.commonLabels | object | `{"app.kubernetes.io/part-of":"suse-observability"}` | Add additional labels to all resources created for elasticsearch |
| elasticsearch.enabled | bool | `true` | Enable / disable chart-based Elasticsearch. |
| elasticsearch.esJavaOpts | string | `"-Xmx3g -Xms3g -Des.allow_insecure_settings=true"` | JVM options |
| elasticsearch.extraEnvs | list | `[{"name":"action.auto_create_index","value":"true"},{"name":"indices.query.bool.max_clause_count","value":"10000"}]` | Extra settings that StackState uses for Elasticsearch. |
| elasticsearch.minimumMasterNodes | int | `2` | Minimum number of Elasticsearch master nodes. |
| elasticsearch.nodeGroup | string | `"master"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.enabled | bool | `true` |  |
| elasticsearch.prometheus-elasticsearch-exporter.es.uri | string | `"http://suse-observability-elasticsearch-master:9200"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.podAnnotations."ad.stackstate.com/exporter.check_names" | string | `"[\"openmetrics\"]"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.podAnnotations."ad.stackstate.com/exporter.init_configs" | string | `"[{}]"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.podAnnotations."ad.stackstate.com/exporter.instances" | string | `"[ { \"prometheus_url\": \"http://%%host%%:9108/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"elasticsearch_indices_store_*\", \"elasticsearch_cluster_health_*\"] } ]"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.limits.cpu | string | `"100m"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.limits.memory | string | `"100Mi"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.requests.cpu | string | `"100m"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.resources.requests.memory | string | `"100Mi"` |  |
| elasticsearch.prometheus-elasticsearch-exporter.serviceMonitor.enabled | bool | `false` |  |
| elasticsearch.prometheus-elasticsearch-exporter.serviceMonitor.labels | object | `{}` | Labels for the service monitor |
| elasticsearch.replicas | int | `3` | Number of Elasticsearch replicas. |
| elasticsearch.resources | object | `{"limits":{"cpu":"2000m","ephemeral-storage":"1Gi","memory":"4Gi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Mi","memory":"4Gi"}}` | Override Elasticsearch resources |
| elasticsearch.sysctlInitContainer | object | `{"enabled":true}` | Enable privileged init container to increase Elasticsearch virtual memory on underlying nodes. |
| elasticsearch.volumeClaimTemplate | object | `{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"250Gi"}}}` | PVC template defaulting to 250Gi default volumes |
| global.backup.enabled | bool | `false` |  |
| global.commonLabels | object | `{}` | Labels that will be added to all Deployments, StatefulSets, CronJobs, Jobs and their pods |
| global.features | object | `{"experimentalStackpacks":false}` | Feature switches for SUSE Observability. |
| global.features.experimentalStackpacks | bool | `false` | Enable StackPacks 2.0 to signal to all components that they should support the StackPacks 2.0 spec. This is a preproduction feature, usage may break your entire installation with upcoming releases. No backwards compatibility is guaranteed. |
| global.imagePullSecrets | list | `[]` | List of image pull secret names to be used by all images across all charts. |
| global.imageRegistry | string | `nil` | Image registry to be used by all images across all charts. When using global.suseObservability (global mode), set this to "registry.rancher.com" to match the default behavior of the suse-observability-values chart. |
| global.receiverApiKey | string | `""` | Deprecated. Use global.suseObservability.receiverApiKey instead. |
| global.storageClass | string | `nil` | StorageClass for all PVCs created by the chart. Can be overridden per PVC. |
| global.suseObservability | object | `{"adminPassword":"","adminPasswordBcrypt":"","affinity":{"nodeAffinity":null,"podAffinity":null,"podAntiAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":true,"topologyKey":"kubernetes.io/hostname"}},"baseUrl":"","license":"","pullSecret":{"password":"","username":""},"receiverApiKey":"","sizing":{"profile":""}}` | Simplified configuration section that allows users to specify high-level settings. When any values in this section are configured (license, baseUrl, sizing.profile, etc.), the chart will automatically use this configuration instead of the legacy stackstate.* values. This provides a single-chart installation experience without needing the separate suse-observability-values chart. NOTE: This section works in conjunction with existing global settings (imageRegistry, receiverApiKey, imagePullSecrets). IMPORTANT: When using this section, also set global.imageRegistry to "registry.rancher.com" for SUSE Observability images. |
| global.suseObservability.adminPassword | string | `""` | Admin password for the default 'admin' user (plain text). Mutually exclusive with adminPasswordBcrypt. Required (one of the two) when using global.suseObservability configuration unless other authentication methods (LDAP, OIDC, Keycloak) are configured. |
| global.suseObservability.adminPasswordBcrypt | string | `""` | Admin password as bcrypt hash. Mutually exclusive with adminPassword. |
| global.suseObservability.affinity | object | `{"nodeAffinity":null,"podAffinity":null,"podAntiAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":true,"topologyKey":"kubernetes.io/hostname"}}` | Affinity configuration for all SUSE Observability components including infrastructure. |
| global.suseObservability.affinity.nodeAffinity | string | `nil` | Node affinity configuration applied to all components (application and infrastructure). Standard Kubernetes nodeAffinity spec. |
| global.suseObservability.affinity.podAffinity | string | `nil` | Pod affinity configuration for application components only. Does NOT apply to infrastructure components. Standard Kubernetes podAffinity spec. |
| global.suseObservability.affinity.podAntiAffinity | object | `{"requiredDuringSchedulingIgnoredDuringExecution":true,"topologyKey":"kubernetes.io/hostname"}` | Simplified pod anti-affinity configuration for HA profiles. Applied to all infrastructure components (kafka, clickhouse, zookeeper, elasticsearch, hbase, victoria-metrics) for HA profiles. |
| global.suseObservability.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution | bool | `true` | Enable required (hard) pod anti-affinity. When true, pods must be scheduled on different nodes. When false, soft anti-affinity is used (preferred but not required). |
| global.suseObservability.affinity.podAntiAffinity.topologyKey | string | `"kubernetes.io/hostname"` | Topology key for pod anti-affinity. Determines the domain for spreading pods (e.g., kubernetes.io/hostname for node-level, topology.kubernetes.io/zone for zone-level). |
| global.suseObservability.baseUrl | string | `""` | Base URL for SUSE Observability (required when using global.suseObservability). |
| global.suseObservability.license | string | `""` | SUSE Observability license key (required when using global.suseObservability). |
| global.suseObservability.pullSecret | object | `{"password":"","username":""}` | Image pull secret configuration. |
| global.suseObservability.pullSecret.password | string | `""` | Password for image pull secret. |
| global.suseObservability.pullSecret.username | string | `""` | Username for image pull secret. |
| global.suseObservability.receiverApiKey | string | `""` | Optional, prefer to use Service Tokens instead for more control and better security. Will use stackstate.apiKey.key if not provided, stackstate.apiKey can also be used to provide the apiKey via an external secret. Please see the documentation on how to use Service Tokens instead. |
| global.suseObservability.sizing | object | `{"profile":""}` | Sizing profile configuration. |
| global.suseObservability.sizing.profile | string | `""` | Sizing profile name. Must match one of the available profiles: 10-nonha, 20-nonha, 50-nonha, 100-nonha, 150-ha, 250-ha, 500-ha, 4000-ha, trial. The chart will automatically apply resource limits, replica counts, and affinity configurations based on the selected profile. These act as intelligent defaults that can be overridden by component-specific values. |
| global.wait.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for wait containers. |
| global.wait.image.registry | string | `"quay.io"` | Base container image registry for wait containers. |
| global.wait.image.repository | string | `"stackstate/wait"` | Base container image repository for wait containers. |
| global.wait.image.tag | string | `"1.0.11-04b49abf"` | Container image tag for wait containers. |
| hbase.all.metrics.agentAnnotationsEnabled | bool | `true` |  |
| hbase.all.metrics.enabled | bool | `true` |  |
| hbase.commonLabels | object | `{"app.kubernetes.io/part-of":"suse-observability"}` | Add additional labels to all resources created for all hbase resources |
| hbase.console.enabled | bool | `true` | Enabled by default for debugging, but with 0 replicas. Manually scale up to 1 replica and open a shell in the container to access the stackgraph console. |
| hbase.console.integrity.enabled | bool | `false` | Enable / disable periodic integrity check to run though a cronjob. |
| hbase.console.integrity.schedule | string | `"*/30 * * * *"` | Schedule at which the integrity check runs |
| hbase.enabled | bool | `true` | Enable / disable chart-based HBase. |
| hbase.hbase.master.experimental.execLivenessProbe.enabled | bool | `true` |  |
| hbase.hbase.master.extraEnv | object | `{"open":{},"secret":{}}` | Extra environment variables for HBase master pods. |
| hbase.hbase.master.extraEnv.open | object | `{}` | Extra open environment variables to inject into HBase master pods. |
| hbase.hbase.master.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into HBase master pods via a Secret object. |
| hbase.hbase.master.replicaCount | string | `nil` | Number of HBase master node replicas. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| hbase.hbase.regionserver.extraEnv | object | `{"open":{},"secret":{}}` | Extra environment variables for HBase regionserver pods. |
| hbase.hbase.regionserver.extraEnv.open | object | `{}` | Extra open environment variables to inject into HBase regionserver pods. |
| hbase.hbase.regionserver.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into HBase regionserver pods via a Secret object. |
| hbase.hbase.regionserver.replicaCount | string | `nil` | Number of HBase regionserver node replicas. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| hbase.hdfs.datanode.extraEnv | object | `{"open":{},"secret":{}}` | Extra environment variables for HDFS datanode pods. |
| hbase.hdfs.datanode.extraEnv.open | object | `{}` | Extra open environment variables to inject into HDFS datanode pods. |
| hbase.hdfs.datanode.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into HDFS datanode pods via a Secret object. |
| hbase.hdfs.datanode.replicaCount | string | `nil` | Number of HDFS datanode replicas. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| hbase.hdfs.minReplication | int | `2` | Min number of copies we create from any data block. (If the hbase.hdfs.datanode.replicaCount is set to a lower value than this, we will use the replicaCount instead) |
| hbase.hdfs.namenode.extraEnv | object | `{"open":{},"secret":{}}` | Extra environment variables for HDFS namenode pods. |
| hbase.hdfs.namenode.extraEnv.open | object | `{}` | Extra open environment variables to inject into HDFS namenode pods. |
| hbase.hdfs.namenode.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into HDFS namenode pods via a Secret object. |
| hbase.hdfs.secondarynamenode.enabled | bool | `true` |  |
| hbase.hdfs.secondarynamenode.extraEnv | object | `{"open":{},"secret":{}}` | Extra environment variables for HDFS secondary namenode pods. |
| hbase.hdfs.secondarynamenode.extraEnv.open | object | `{}` | Extra open environment variables to inject into HDFS secondary namenode pods. |
| hbase.hdfs.secondarynamenode.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into HDFS secondary namenode pods via a Secret object. |
| hbase.stackgraph.version | string | `"7.13.18"` | The StackGraph server version, must be compatible with the StackState version |
| hbase.tephra.extraEnv | object | `{"open":{},"secret":{}}` | Extra environment variables for Tephra pods. |
| hbase.tephra.extraEnv.open | object | `{}` | Extra open environment variables to inject into Tephra pods. |
| hbase.tephra.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into Tephra pods via a Secret object. |
| hbase.tephra.replicaCount | string | `nil` | Number of Tephra replicas. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| hbase.version | string | `"2.5"` | Version of hbase to use |
| hbase.zookeeper.externalServers | string | `"suse-observability-zookeeper-headless"` | External Zookeeper if not used bundled Zookeeper chart **Don't change unless otherwise specified**. |
| ingress.annotations | object | `{}` | Annotations for ingress objects. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| ingress.hosts | list | `[]` | List of ingress hostnames; the paths are fixed to StackState backend services |
| ingress.ingressClassName | string | `""` |  |
| ingress.path | string | `"/"` |  |
| ingress.tls | list | `[]` | List of ingress TLS certificates to use. |
| kafka.command | list | `["/scripts/custom-setup.sh"]` | Override kafka container command. |
| kafka.commonLabels | object | `{"app.kubernetes.io/part-of":"suse-observability"}` | Add additional labels to all resources created for kafka |
| kafka.containerSecurityContext.allowPrivilegeEscalation | bool | `false` |  |
| kafka.containerSecurityContext.capabilities.drop[0] | string | `"ALL"` |  |
| kafka.containerSecurityContext.enabled | bool | `true` |  |
| kafka.containerSecurityContext.runAsNonRoot | bool | `true` |  |
| kafka.containerSecurityContext.runAsUser | int | `1001` |  |
| kafka.containerSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| kafka.defaultReplicationFactor | int | `2` |  |
| kafka.enabled | bool | `true` | Enable / disable chart-based Kafka. |
| kafka.externalZookeeper.servers | string | `"suse-observability-zookeeper-headless"` | External Zookeeper if not used bundled Zookeeper chart **Don't change unless otherwise specified**. |
| kafka.extraDeploy | list | `[{"apiVersion":"v1","data":{"custom-setup.sh":"#!/bin/bash\n\nID=\"${MY_POD_NAME#\"{{ include \"common.names.fullname\" . }}-\"}\"\n\nKAFKA_META_PROPERTIES=/bitnami/kafka/data/meta.properties\nif [[ -f ${KAFKA_META_PROPERTIES} ]]; then\n  ID=`grep -e ^broker.id= ${KAFKA_META_PROPERTIES} | sed 's/^broker.id=//'`\n  if [[ \"${ID}\" != \"\" ]] && [[ \"${ID}\" -gt 1000 ]]; then\n    echo \"Using broker ID ${ID} from ${KAFKA_META_PROPERTIES} for compatibility (STAC-9614)\"\n  fi\nfi\n\nexport KAFKA_CFG_BROKER_ID=\"$ID\"\n\nexec /entrypoint.sh /run.sh"},"kind":"ConfigMap","metadata":{"name":"kafka-custom-scripts"}}]` | Array of extra objects to deploy with the release |
| kafka.extraEnvVars | list | `[{"name":"KAFKA_CFG_RESERVED_BROKER_MAX_ID","value":"2000"},{"name":"KAFKA_CFG_TRANSACTIONAL_ID_EXPIRATION_MS","value":"2147483647"}]` | Extra environment variables to add to kafka pods. |
| kafka.extraVolumeMounts | list | `[{"mountPath":"/scripts/custom-setup.sh","name":"kafka-custom-scripts","subPath":"custom-setup.sh"}]` | Extra volumeMount(s) to add to Kafka containers. |
| kafka.extraVolumes | list | `[{"configMap":{"defaultMode":493,"name":"kafka-custom-scripts"},"name":"kafka-custom-scripts"}]` | Extra volume(s) to add to Kafka statefulset. |
| kafka.fullnameOverride | string | `"suse-observability-kafka"` | Name override for Kafka child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| kafka.image.registry | string | `"quay.io"` | Kafka image registry |
| kafka.image.repository | string | `"stackstate/kafka"` | Kafka image repository |
| kafka.image.tag | string | `"3.9.1-4e2ea587-242"` | Kafka image tag. **Since StackState relies on this specific version, it's advised NOT to change this.** When changing this version, be sure to change the pod annotation stackstate.com/kafkaup-operator.kafka_version aswell, in order for the kafkaup operator to upgrade the inter broker protocol version |
| kafka.initContainers | list | `[{"args":["-c","trap 'exit 1' INT TERM; while [ -z \"${KAFKA_CFG_INTER_BROKER_PROTOCOL_VERSION}\" ]; do echo \"KAFKA_CFG_INTER_BROKER_PROTOCOL_VERSION should be set by operator\"; sleep 1; done"],"command":["/bin/bash"],"image":"{{ include \"kafka.image\" . }}","imagePullPolicy":"","name":"check-inter-broker-protocol-version","resources":{"limits":{},"requests":{}},"securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"runAsNonRoot":true,"seccompProfile":{"type":"RuntimeDefault"}}}]` | required to make the kafka versionup operator work |
| kafka.logRetentionHours | int | `24` | The minimum age of a log file to be eligible for deletion due to age. |
| kafka.metrics.jmx.containerSecurityContext.allowPrivilegeEscalation | bool | `false` |  |
| kafka.metrics.jmx.containerSecurityContext.capabilities.drop[0] | string | `"ALL"` |  |
| kafka.metrics.jmx.containerSecurityContext.enabled | bool | `true` |  |
| kafka.metrics.jmx.containerSecurityContext.runAsNonRoot | bool | `true` |  |
| kafka.metrics.jmx.containerSecurityContext.runAsUser | int | `1001` |  |
| kafka.metrics.jmx.containerSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| kafka.metrics.jmx.enabled | bool | `true` | Whether or not to expose JMX metrics to Prometheus. |
| kafka.metrics.jmx.heapSizeMB | int | `256` |  |
| kafka.metrics.jmx.image.registry | string | `"quay.io"` | Kafka JMX exporter image registry |
| kafka.metrics.jmx.image.repository | string | `"stackstate/jmx-exporter"` | Kafka JMX exporter image repository |
| kafka.metrics.jmx.image.tag | string | `"0.20.0-d19cf0ff-222"` | Kafka JMX exporter image tag |
| kafka.metrics.jmx.resources.limits.cpu | string | `"1"` |  |
| kafka.metrics.jmx.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| kafka.metrics.jmx.resources.limits.memory | string | `"300Mi"` |  |
| kafka.metrics.jmx.resources.requests.cpu | string | `"200m"` |  |
| kafka.metrics.jmx.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| kafka.metrics.jmx.resources.requests.memory | string | `"300Mi"` |  |
| kafka.metrics.jmx.service.annotations."monitor.kubernetes-v2.stackstate.io/http-response-time" | string | `"{ \"deviatingThreshold\": 10.0, \"criticalThreshold\": 10.0 }"` |  |
| kafka.metrics.kafka.enabled | bool | `false` | Whether or not to create a standalone Kafka exporter to expose Kafka metrics. |
| kafka.metrics.serviceMonitor.enabled | bool | `false` | If `true`, creates a Prometheus Operator `ServiceMonitor` (also requires `kafka.metrics.kafka.enabled` or `kafka.metrics.jmx.enabled` to be `true`). |
| kafka.metrics.serviceMonitor.interval | string | `"20s"` | How frequently to scrape metrics. |
| kafka.metrics.serviceMonitor.labels | object | `{}` | Add extra labels to target a specific prometheus instance |
| kafka.offsetsTopicReplicationFactor | int | `2` |  |
| kafka.pdb.create | bool | `true` |  |
| kafka.pdb.maxUnavailable | int | `1` |  |
| kafka.pdb.minAvailable | string | `""` |  |
| kafka.persistence.size | string | `"100Gi"` | Size of persistent volume for each Kafka pod |
| kafka.podAnnotations | object | `{"ad.stackstate.com/jmx-exporter.check_names":"[\"openmetrics\"]","ad.stackstate.com/jmx-exporter.init_configs":"[{}]","ad.stackstate.com/jmx-exporter.instances":"[ { \"prometheus_url\": \"http://%%host%%:5556/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"*\"], \"type_overrides\": {\"kafka_server_replicamanager_total_underreplicatedpartitions_value\":\"gauge\", \"kafka_controller_kafkacontroller_offlinepartitionscount_value\":\"gauge\", \"kafka_controller_kafkacontroller_activecontrollercount_value\": \"gauge\"}}]","stackstate.com/kafkaup-operator.kafka_version":"3.9.1"}` | Kafka Pod annotations. |
| kafka.podLabels."app.kubernetes.io/part-of" | string | `"suse-observability"` |  |
| kafka.readinessProbe.initialDelaySeconds | int | `45` | Delay before readiness probe is initiated. |
| kafka.replicaCount | string | `nil` | Number of Kafka replicas. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| kafka.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2Gi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"2Gi"}}` | Kafka resources per pods. |
| kafka.service.annotations."monitor.kubernetes-v2.stackstate.io/http-response-time" | string | `"{ \"deviatingThreshold\": 10.0, \"criticalThreshold\": 10.0 }"` |  |
| kafka.service.headless.annotations."monitor.kubernetes-v2.stackstate.io/http-response-time" | string | `"{ \"deviatingThreshold\": 10.0, \"criticalThreshold\": 10.0 }"` |  |
| kafka.topic.stsMetricsV2.partitionCount | int | `10` |  |
| kafka.topicRetention | string | `"86400000"` | Max time in milliseconds to retain data in each topic. |
| kafka.transactionStateLogReplicationFactor | int | `2` |  |
| kafka.volumePermissions.enabled | bool | `false` |  |
| kafka.zookeeper.enabled | bool | `false` | Disable Zookeeper from the Kafka chart **Don't change unless otherwise specified**. |
| kafkaup-operator.enabled | bool | `true` |  |
| kafkaup-operator.image.pullPolicy | string | `""` |  |
| kafkaup-operator.image.registry | string | `"quay.io"` |  |
| kafkaup-operator.image.repository | string | `"stackstate/kafkaup-operator"` |  |
| kafkaup-operator.image.tag | string | `"0.0.6"` |  |
| kafkaup-operator.kafkaSelectors.podLabel.key | string | `"app.kubernetes.io/component"` |  |
| kafkaup-operator.kafkaSelectors.podLabel.value | string | `"kafka"` |  |
| kafkaup-operator.kafkaSelectors.statefulSetName | string | `"suse-observability-kafka"` |  |
| kafkaup-operator.startVersion | string | `"2.3.1"` |  |
| kubernetes-rbac-agent.clusterName.value | string | `"{{ .Release.Name }}"` |  |
| kubernetes-rbac-agent.containers.rbacAgent.affinity | object | `{}` | Set affinity |
| kubernetes-rbac-agent.containers.rbacAgent.env | object | `{}` | Additional environment variables |
| kubernetes-rbac-agent.containers.rbacAgent.image.repository | string | `"stackstate/kubernetes-rbac-agent"` |  |
| kubernetes-rbac-agent.containers.rbacAgent.nodeSelector | object | `{}` | Set a nodeSelector |
| kubernetes-rbac-agent.containers.rbacAgent.podAnnotations | object | `{"ad.stackstate.com/kubernetes-rbac-agent.check_names":"[\"openmetrics\"]","ad.stackstate.com/kubernetes-rbac-agent.init_configs":"[{}]","ad.stackstate.com/kubernetes-rbac-agent.instances":"[ { \"prometheus_url\": \"http://%%host%%:8080/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"*\"] } ]"}` | Additional annotations on the pod |
| kubernetes-rbac-agent.containers.rbacAgent.podLabels | object | `{}` | Additional labels on the pod |
| kubernetes-rbac-agent.containers.rbacAgent.priorityClassName | string | `""` | Set priorityClassName |
| kubernetes-rbac-agent.containers.rbacAgent.resources.limits.memory | string | `"40Mi"` | Memory resource limits. |
| kubernetes-rbac-agent.containers.rbacAgent.resources.requests.memory | string | `"25Mi"` | Memory resource requests. |
| kubernetes-rbac-agent.containers.rbacAgent.tolerations | list | `[]` | Set tolerations |
| kubernetes-rbac-agent.url.value | string | `"{{ include \"stackstate.rbacAgent.url\" . }}"` |  |
| minio.accessKey | string | `"setme"` | Access key for MinIO. Default is set to an invalid value that will cause MinIO to not start up to ensure users of this Helm chart set an explicit value. |
| minio.azuregateway.replicas | int | `1` |  |
| minio.fullnameOverride | string | `"suse-observability-minio"` | **N.B.: Do not change this value!** The fullname override for MinIO subchart is hardcoded so that the stackstate chart can refer to its components. |
| minio.image.registry | string | `"quay.io"` | MinIO image registry |
| minio.image.repository | string | `"stackstate/minio"` | MinIO image repository |
| minio.persistence.enabled | bool | `false` | Enables MinIO persistence. Must be enabled when MinIO is not configured as a gateway to AWS S3 or Azure Blob Storage. |
| minio.replicas | int | `1` | Number of MinIO replicas. |
| minio.s3gateway.replicas | int | `1` |  |
| minio.secretKey | string | `"setme"` | Secret key for MinIO. Default is set to an invalid value that will cause MinIO to not start up to ensure users of this Helm chart set an explicit value. |
| networkPolicy.enabled | bool | `false` | Enable creating of `NetworkPolicy` object and associated rules for StackState. |
| networkPolicy.spec | object | `{"ingress":[{"from":[{"podSelector":{}}]}],"podSelector":{"matchLabels":{}},"policyTypes":["Ingress"]}` | `NetworkPolicy` rules for StackState. |
| opentelemetry-collector.extraEnvs | list | `[{"name":"API_URL","valueFrom":{"configMapKeyRef":{"key":"api.url","name":"suse-observability-otel-collector"}}},{"name":"INTAKE_URL","valueFrom":{"configMapKeyRef":{"key":"intake.url","name":"suse-observability-otel-collector"}}}]` | Collector configuration, see: [doc](https://opentelemetry.io/docs/collector/configuration/). Contains API_URL with path to api server used to authorize requests |
| opentelemetry-collector.fullnameOverride | string | `"suse-observability-otel-collector"` | Name override for OTEL collector child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| opentelemetry-collector.image.registry | string | `"quay.io"` |  |
| opentelemetry-collector.image.repository | string | `"stackstate/sts-opentelemetry-collector"` | Repository where to get the image from. |
| opentelemetry-collector.image.tag | string | `"v0.0.25"` | Container image tag for 'opentelemetry-collector' containers. |
| opentelemetry-collector.initContainers[0].command[0] | string | `"sh"` |  |
| opentelemetry-collector.initContainers[0].command[1] | string | `"-c"` |  |
| opentelemetry-collector.initContainers[0].command[2] | string | `"/entrypoint -c suse-observability-clickhouse:9000,suse-observability-vmagent:8429,suse-observability-kafka-headless:9092 -t 300\n"` |  |
| opentelemetry-collector.initContainers[0].image | string | `"{{ include \"opentelemetry-collector.waitImageRegistry\" . }}/{{ .Values.global.wait.image.repository }}:{{ .Values.global.wait.image.tag }}"` |  |
| opentelemetry-collector.initContainers[0].imagePullPolicy | string | `"IfNotPresent"` |  |
| opentelemetry-collector.initContainers[0].name | string | `"otel-collector-init"` |  |
| opentelemetry-collector.mode | string | `"statefulset"` | deployment mode of OTEL collector. Valid values are "daemonset", "deployment", and "statefulset". |
| opentelemetry-collector.podAnnotations."ad.stackstate.com/opentelemetry-collector.check_names" | string | `"[\"openmetrics\"]"` |  |
| opentelemetry-collector.podAnnotations."ad.stackstate.com/opentelemetry-collector.init_configs" | string | `"[{}]"` |  |
| opentelemetry-collector.podAnnotations."ad.stackstate.com/opentelemetry-collector.instances" | string | `"[ { \"prometheus_url\": \"http://%%host%%:8888/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"*\"] } ]"` |  |
| opentelemetry-collector.replicaCount | int | `1` | only used with deployment mode |
| opentelemetry-collector.resources.limits.cpu | string | `"500m"` |  |
| opentelemetry-collector.resources.limits.memory | string | `"512Mi"` |  |
| opentelemetry-collector.resources.requests.cpu | string | `"250m"` |  |
| opentelemetry-collector.resources.requests.memory | string | `"512Mi"` |  |
| opentelemetry.enabled | bool | `true` | Enable / disable chart-based OTEL. |
| pull-secret.credentials | list | `[]` | Registry and associated credentials (username, password) that will be stored in the pull-secret |
| pull-secret.enabled | bool | `false` | Deploy the ImagePullSecret for the chart. |
| pull-secret.fullNameOverride | string | `""` | Name of the ImagePullSecret that will be created. This can be referenced by setting the `global.imagePullSecrets[0].name` value in the chart. |
| scc.enabled | bool | `false` | Create `SecurityContextConstraints` resource to manage Openshift security constraints for Stackstate. Has to be enabled when installing to Openshift >= 4.12 The resource is deployed as a Helm pre-install hook to avoid any warning for the first deployment. Because `helm uninstall` does not consider Helm hooks, the resource must be manually deleted after the Helm release is removed. |
| stackstate.allowedOrigins | list | `[]` | Third-party web domains allowed to make cross-origin requests |
| stackstate.apiKey | object | `{"fromExternalSecret":null,"key":null}` | Optional, API key configuration for StackState. prefer to use Service Tokens instead for more control and better security. |
| stackstate.apiKey.fromExternalSecret | string | `nil` | Use an external secret for the api key. This suppresses secret creation by StackState and gets the data from the secret with the provided name. |
| stackstate.apiKey.key | string | `nil` | API key to be used by the Receiver. |
| stackstate.authentication | object | `{"adminPassword":null,"file":{},"fromExternalSecret":null,"keycloak":{},"ldap":{},"oidc":{},"rancher":{},"roles":{"admin":[],"custom":{},"guest":[],"k8sTroubleshooter":[],"powerUser":[]},"serviceToken":{"bootstrap":{"dedicatedSubject":"","roles":[],"token":"","ttl":"24h"}},"sessionLifetime":""}` | Configure the authentication settings for StackState here. Only one of the authentication providers can be used, configuring multiple will result in an error. |
| stackstate.authentication.adminPassword | string | `nil` | Password for the 'admin' user that StackState creates by default |
| stackstate.authentication.file | object | `{}` | Configure users, their passwords and roles from (config) file |
| stackstate.authentication.fromExternalSecret | string | `nil` | Use an external secret for the authenticated secrets. This suppresses secret creation by StackState and gets the data from the secret with the provided name. |
| stackstate.authentication.keycloak | object | `{}` | Use Keycloak as authentication provider. See [Configuring Keycloak](#configuring-keycloak). |
| stackstate.authentication.ldap | object | `{}` | LDAP settings for StackState. See [Configuring LDAP](#configuring-ldap). |
| stackstate.authentication.oidc | object | `{}` | Use an OpenId Connect provider for authentication. See [Configuring OpenId Connect](#configuring-openid-connect). |
| stackstate.authentication.rancher | object | `{}` | Use Rancher as an OpenId Connect provider for authentication. See [Configuring Rancher authentication](#configuring-rancher-authentication). |
| stackstate.authentication.roles | object | `{"admin":[],"custom":{},"guest":[],"k8sTroubleshooter":[],"powerUser":[]}` | Extend the default role names in StackState |
| stackstate.authentication.roles.admin | list | `[]` | Extend the role names that have admin permissions (default: 'stackstate-admin') |
| stackstate.authentication.roles.custom | object | `{}` | Extend the authorization with custom roles {roleName: {systemPermissions: [], resourcePermissions: {}, viewPermissions: [], topologyScope: ""}} |
| stackstate.authentication.roles.guest | list | `[]` | Extend the role names that have guest permissions (default: 'stackstate-guest') |
| stackstate.authentication.roles.k8sTroubleshooter | list | `[]` | Extend the role names that have troubleshooter permissions (default: 'stackstate-k8s-troubleshooter') |
| stackstate.authentication.roles.powerUser | list | `[]` | Extend the role names that have power user permissions (default: 'stackstate-power-user') |
| stackstate.authentication.serviceToken.bootstrap.dedicatedSubject | string | `""` | Subject solely created for usage with this token and cleaned up with the token |
| stackstate.authentication.serviceToken.bootstrap.roles | list | `[]` | The roles the service token assumes when it’s used for authentication |
| stackstate.authentication.serviceToken.bootstrap.token | string | `""` | The service token to set as bootstrap token |
| stackstate.authentication.serviceToken.bootstrap.ttl | string | `"24h"` | The amount of time the service token is valid for |
| stackstate.authentication.sessionLifetime | string | `""` | Amount of time to keep a session when a user does not log in |
| stackstate.baseUrl | string | `nil` | **PROVIDE YOUR BASE URL HERE** Externally visible baseUrl of StackState. |
| stackstate.components.all.affinity | object | `{}` | Affinity settings for pod assignment on all components. |
| stackstate.components.all.clickHouse.database | string | `"otel"` |  |
| stackstate.components.all.clickHouse.hostnames | string | `"suse-observability-clickhouse-headless"` |  |
| stackstate.components.all.clickHouse.parameters.health_check_interval | string | `"5000"` |  |
| stackstate.components.all.clickHouse.password | string | `""` |  |
| stackstate.components.all.clickHouse.port | int | `8123` |  |
| stackstate.components.all.clickHouse.protocol | string | `"http"` |  |
| stackstate.components.all.clickHouse.username | string | `"stackstate"` |  |
| stackstate.components.all.deploymentStrategy.type | string | `"RecreateSingletonsOnly"` | Deployment strategy for StackState components. Possible values: `RollingUpdate`, `Recreate` and `RecreateSingletonsOnly`. `RecreateSingletonsOnly` uses `Recreate` for the singleton Deployments and `RollingUpdate` for the other Deployments. |
| stackstate.components.all.elasticsearchEndpoint | string | `""` | **Required if `elasticsearch.enabled` is `false`** Endpoint for shared Elasticsearch cluster. |
| stackstate.components.all.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.all.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods for all components. |
| stackstate.components.all.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object for all components. |
| stackstate.components.all.image.pullPolicy | string | `"IfNotPresent"` | The default pullPolicy used for all stateless components of StackState; individual service `pullPolicy`s can be overridden (see below). |
| stackstate.components.all.image.registry | string | `"quay.io"` | Base container image registry for all StackState containers, except for the wait container and the container-tools container |
| stackstate.components.all.image.repositorySuffix | string | `""` |  |
| stackstate.components.all.image.tag | string | `"7.0.0-snapshot.20260212100200-master-2b9246f"` | The default tag used for all stateless components of StackState; individual service `tag`s can be overridden (see below). |
| stackstate.components.all.kafkaEndpoint | string | `""` | **Required if `elasticsearch.enabled` is `false`** Endpoint for shared Kafka broker. |
| stackstate.components.all.metricStore.remoteWritePath | string | `"/api/v1/write"` | Remote write path used to ingest metrics, /api/v1/write is most common |
| stackstate.components.all.metrics.agentAnnotationsEnabled | bool | `true` | Put annotations on each pod to instruct the stackstate agent to scrape the metrics |
| stackstate.components.all.metrics.defaultAgentMetricsFilter | string | `"[\"kafka_consumer_consumer_fetch_manager_metrics*\", \"kafka_producer_producer_topic_metrics*\", \"jvm*\", \"akka_http_requests_active\", \"stackstate*\", \"receiver*\", \"stackgraph*\", \"caffeine*\"]"` |  |
| stackstate.components.all.metrics.enabled | bool | `true` | Enable metrics port. |
| stackstate.components.all.metrics.servicemonitor.additionalLabels | object | `{}` | Additional labels for targeting Prometheus operator instances. |
| stackstate.components.all.metrics.servicemonitor.enabled | bool | `false` | Enable `ServiceMonitor` object; `all.metrics.enabled` *must* be enabled. |
| stackstate.components.all.nodeSelector | object | `{}` | Node labels for pod assignment on all components. |
| stackstate.components.all.otelInstrumentation.enabled | bool | `false` |  |
| stackstate.components.all.otelInstrumentation.otlpExporterEndpoint | string | `""` |  |
| stackstate.components.all.otelInstrumentation.otlpExporterProtocol | string | `"grpc"` |  |
| stackstate.components.all.otelInstrumentation.serviceNamespace | string | `"{{ printf \"%s-%s\" .Chart.Name .Release.Namespace }}"` |  |
| stackstate.components.all.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.all.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| stackstate.components.all.securityContext.fsGroup | int | `65534` | The GID (group ID) used to mount volumes |
| stackstate.components.all.securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| stackstate.components.all.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| stackstate.components.all.securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| stackstate.components.all.tolerations | list | `[]` | Toleration labels for pod assignment on all components. |
| stackstate.components.all.zookeeperEndpoint | string | `""` | **Required if `zookeeper.enabled` is `false`** Endpoint for shared Zookeeper nodes. |
| stackstate.components.api.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.api.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.api.config | string | `""` | Configuration file contents to customize the default StackState api configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.api.docslink | string | `""` | Documentation URL root to use in the product help page & tooltips. |
| stackstate.components.api.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.api.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.api.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.api.image.imageRegistry | string | `""` | `imageRegistry` used for the `api` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.api.image.pullPolicy | string | `""` | `pullPolicy` used for the `api` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.api.image.repository | string | `"stackstate/stackstate-server"` | Repository of the api component Docker image. |
| stackstate.components.api.image.tag | string | `""` | Tag used for the `api` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.api.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.api.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.api.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `api` pods. |
| stackstate.components.api.replicaCount | int | `1` | Number of `api` replicas. |
| stackstate.components.api.resources | object | `{"limits":{"cpu":"2000m","ephemeral-storage":"2Gi","memory":"2Gi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Mi","memory":"2Gi"}}` | Resource allocation for `api` pods. |
| stackstate.components.api.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.api.sizing.javaHeapMemoryFraction | string | `"45"` |  |
| stackstate.components.api.supportMode | string | `""` | Mode of support, either Documentation or ContactStackstate |
| stackstate.components.api.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.api.yaml | object | `{}` |  |
| stackstate.components.authorizationSync.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.authorizationSync.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.authorizationSync.config | string | `""` | Configuration file contents to customize the default StackState notification configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.authorizationSync.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.authorizationSync.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.authorizationSync.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.authorizationSync.image.imageRegistry | string | `""` | `imageRegistry` used for the `notification` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.authorizationSync.image.pullPolicy | string | `""` | `pullPolicy` used for the `notification` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.authorizationSync.image.repository | string | `"stackstate/stackstate-server"` |  |
| stackstate.components.authorizationSync.image.tag | string | `""` | Tag used for the `notification` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.authorizationSync.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.authorizationSync.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.authorizationSync.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `notification` pods. |
| stackstate.components.authorizationSync.replicaCount | int | `1` | Number of `notification` replicas. |
| stackstate.components.authorizationSync.resources | object | `{"limits":{"cpu":"1500m","ephemeral-storage":"1Gi","memory":"1Gi"},"requests":{"cpu":"250m","ephemeral-storage":"1Mi","memory":"512Mi"}}` | Resource allocation for `notification` pods. |
| stackstate.components.authorizationSync.sizing.baseMemoryConsumption | string | `"25Mi"` |  |
| stackstate.components.authorizationSync.sizing.javaHeapMemoryFraction | string | `"70"` |  |
| stackstate.components.authorizationSync.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.backup.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.backup.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.backup.podAnnotations | object | `{}` | Extra annotations for backup pods. |
| stackstate.components.backup.podLabels | object | `{}` | Extra labels for backup pods. |
| stackstate.components.backup.resources.limits.cpu | string | `"3000m"` |  |
| stackstate.components.backup.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| stackstate.components.backup.resources.limits.memory | string | `"4000Mi"` |  |
| stackstate.components.backup.resources.requests.cpu | string | `"1000m"` |  |
| stackstate.components.backup.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| stackstate.components.backup.resources.requests.memory | string | `"4000Mi"` |  |
| stackstate.components.backup.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.checks.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.checks.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.checks.config | string | `""` | Configuration file contents to customize the default StackState state configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.checks.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.checks.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.checks.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.checks.image.imageRegistry | string | `""` | `imageRegistry` used for the `checks` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.checks.image.pullPolicy | string | `""` | `pullPolicy` used for the `state` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.checks.image.repository | string | `"stackstate/stackstate-server"` | Repository of the sync component Docker image. |
| stackstate.components.checks.image.tag | string | `""` | Tag used for the `state` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.checks.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.checks.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.checks.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `checks` pods. |
| stackstate.components.checks.replicaCount | int | `1` | Number of `checks` replicas. |
| stackstate.components.checks.resources | object | `{"limits":{"cpu":"2000m","ephemeral-storage":"1Gi","memory":"4000Mi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Mi","memory":"4000Mi"}}` | Resource allocation for `state` pods. |
| stackstate.components.checks.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.checks.sizing.javaHeapMemoryFraction | string | `"60"` |  |
| stackstate.components.checks.tmpToPVC | object | `{"storageClass":null,"volumeSize":"2Gi"}` | Whether to use PersistentVolume to store temporary files (/tmp) instead of pod ephemeral storage, empty - use pod ephemeral storage. |
| stackstate.components.checks.tmpToPVC.storageClass | string | `nil` | Storage class name of PersistentVolume used by /tmp directory. It stores temporary files/caches, so it should be the fastest possible. |
| stackstate.components.checks.tmpToPVC.volumeSize | string | `"2Gi"` | The size of the PersistentVolume for "/tmp" directory. |
| stackstate.components.checks.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.clickhouseCleanup.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.clickhouseCleanup.extraEnv.open | object | `{}` | Add additional environment variables to the pod |
| stackstate.components.clickhouseCleanup.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy `clickhouseCleanup` containers. |
| stackstate.components.clickhouseCleanup.image.registry | string | `"quay.io"` | Registry where to get the image from |
| stackstate.components.clickhouseCleanup.image.repository | string | `"stackstate/clickhouse"` | Repository where to get the image from. |
| stackstate.components.clickhouseCleanup.image.tag | string | `"25.9.5-85835861-137"` | Container image tag for 'clickhouseCleanup' containers. |
| stackstate.components.clickhouseCleanup.jobAnnotations | object | `{}` | Annotations for clickhouseCleanup job. |
| stackstate.components.clickhouseCleanup.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.clickhouseCleanup.podAnnotations | object | `{}` | Extra annotations for clickhouse cleanup job pods. |
| stackstate.components.clickhouseCleanup.podLabels | object | `{}` | Extra labels for clickhouse cleanup job pods. |
| stackstate.components.clickhouseCleanup.resources | object | `{"limits":{"cpu":"100m","ephemeral-storage":"1Gi","memory":"200Mi"},"requests":{"cpu":"100m","ephemeral-storage":"1Mi","memory":"200Mi"}}` | Resource allocation for `clickhouseCleanup` pods. |
| stackstate.components.clickhouseCleanup.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| stackstate.components.clickhouseCleanup.securityContext.fsGroup | int | `1001` | The GID (group ID) used to mount volumes |
| stackstate.components.clickhouseCleanup.securityContext.runAsGroup | int | `1001` | The GID (group ID) of the owning user of the process |
| stackstate.components.clickhouseCleanup.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| stackstate.components.clickhouseCleanup.securityContext.runAsUser | int | `1001` | The UID (user ID) of the owning user of the process |
| stackstate.components.clickhouseCleanup.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.configurationBackup.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.configurationBackup.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.configurationBackup.podAnnotations | object | `{}` | Extra annotations for configuration backup pods. |
| stackstate.components.configurationBackup.podLabels | object | `{}` | Extra labels for configuration backup pods. |
| stackstate.components.configurationBackup.resources.limits.cpu | string | `"1000m"` |  |
| stackstate.components.configurationBackup.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| stackstate.components.configurationBackup.resources.limits.memory | string | `"1000Mi"` |  |
| stackstate.components.configurationBackup.resources.requests.cpu | string | `"1000m"` |  |
| stackstate.components.configurationBackup.resources.requests.ephemeral-storage | string | `"100Mi"` |  |
| stackstate.components.configurationBackup.resources.requests.memory | string | `"1000Mi"` |  |
| stackstate.components.configurationBackup.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.containerTools.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for container-tools containers. |
| stackstate.components.containerTools.image.registry | string | `"quay.io"` | Base container image registry for container-tools containers. |
| stackstate.components.containerTools.image.repository | string | `"stackstate/container-tools"` | Base container image repository for container-tools containers. |
| stackstate.components.containerTools.image.tag | string | `"1.8.2-bci-517"` | Container image tag for container-tools containers. |
| stackstate.components.containerTools.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"2000Mi"}}` | Resource allocation for `kafkaTopicCreate` pods. |
| stackstate.components.correlate.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.correlate.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.correlate.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.correlate.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.correlate.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.correlate.image.imageRegistry | string | `""` | `imageRegistry` used for the `correlate` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.correlate.image.pullPolicy | string | `""` | `pullPolicy` used for the `correlate` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.correlate.image.repository | string | `"stackstate/stackstate-correlate"` | Repository of the correlate component Docker image. |
| stackstate.components.correlate.image.tag | string | `""` | Tag used for the `correlate` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.correlate.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.correlate.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.correlate.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `correlate` pods. |
| stackstate.components.correlate.replicaCount | string | `nil` | Number of `correlate` replicas. |
| stackstate.components.correlate.resources | object | `{"limits":{"cpu":"2000m","ephemeral-storage":"1Gi","memory":"2800Mi"},"requests":{"cpu":"600m","ephemeral-storage":"1Mi","memory":"2800Mi"}}` | Resource allocation for `correlate` pods. |
| stackstate.components.correlate.sizing.baseMemoryConsumption | string | `"400Mi"` |  |
| stackstate.components.correlate.sizing.javaHeapMemoryFraction | string | `"65"` |  |
| stackstate.components.correlate.split.aggregator.affinity | object | `{}` | Additional affinity settings for pod assignment. |
| stackstate.components.correlate.split.aggregator.extraEnv.open | object | `{"CONFIG_FORCE_stackstate_correlate_aggregation_workers":"3","CONFIG_FORCE_stackstate_correlate_correlateConnections_workers":"0","CONFIG_FORCE_stackstate_correlate_correlateHTTPTraces_workers":"0"}` | Extra open environment variables to inject into pods. Will merge with stackstate.components.correlate.extraEnv.open |
| stackstate.components.correlate.split.aggregator.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.correlate.extraEnv.secret |
| stackstate.components.correlate.split.aggregator.nodeSelector | object | `{}` | Additional node labels for pod assignment. |
| stackstate.components.correlate.split.aggregator.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.correlate.split.aggregator.replicaCount | int | `1` | Number of `aggregator correlate` replicas. |
| stackstate.components.correlate.split.aggregator.resources | object | `{"limits":{"cpu":null,"ephemeral-storage":null,"memory":null},"requests":{"cpu":null,"ephemeral-storage":null,"memory":null}}` | Resource allocation for pods. If not defined, will take from stackstate.components.correlate.resources |
| stackstate.components.correlate.split.aggregator.sizing.javaHeapMemoryFraction | string | `nil` |  |
| stackstate.components.correlate.split.aggregator.sizing.logsMemoryConsumption | string | `nil` |  |
| stackstate.components.correlate.split.aggregator.tolerations | list | `[]` | Additional toleration labels for pod assignment. |
| stackstate.components.correlate.split.connection.affinity | object | `{}` | Additional affinity settings for pod assignment. |
| stackstate.components.correlate.split.connection.extraEnv.open | object | `{"CONFIG_FORCE_stackstate_correlate_aggregation_workers":"0","CONFIG_FORCE_stackstate_correlate_correlateConnections_workers":"3","CONFIG_FORCE_stackstate_correlate_correlateHTTPTraces_workers":"0"}` | Extra open environment variables to inject into pods. Will merge with stackstate.components.correlate.extraEnv.open |
| stackstate.components.correlate.split.connection.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.correlate.extraEnv.secret |
| stackstate.components.correlate.split.connection.nodeSelector | object | `{}` | Additional node labels for pod assignment. |
| stackstate.components.correlate.split.connection.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.correlate.split.connection.replicaCount | int | `1` | Number of `connection correlate` replicas. |
| stackstate.components.correlate.split.connection.resources | object | `{"limits":{"cpu":null,"ephemeral-storage":null,"memory":null},"requests":{"cpu":null,"ephemeral-storage":null,"memory":null}}` | Resource allocation for pods. If not defined, will take from stackstate.components.correlate.resources |
| stackstate.components.correlate.split.connection.sizing.baseMemoryConsumption | string | `nil` |  |
| stackstate.components.correlate.split.connection.sizing.javaHeapMemoryFraction | string | `nil` |  |
| stackstate.components.correlate.split.connection.tolerations | list | `[]` | Additional toleration labels for pod assignment. |
| stackstate.components.correlate.split.enabled | bool | `false` | Split the correlate into the connection connection correlator, http correlator and aggregator |
| stackstate.components.correlate.split.httpTracing.affinity | object | `{}` | Additional affinity settings for pod assignment. |
| stackstate.components.correlate.split.httpTracing.extraEnv.open | object | `{"CONFIG_FORCE_stackstate_correlate_aggregation_workers":"0","CONFIG_FORCE_stackstate_correlate_correlateConnections_workers":"0","CONFIG_FORCE_stackstate_correlate_correlateHTTPTraces_workers":"3"}` | Extra open environment variables to inject into pods. Will merge with stackstate.components.correlate.extraEnv.open |
| stackstate.components.correlate.split.httpTracing.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.correlate.extraEnv.secret |
| stackstate.components.correlate.split.httpTracing.nodeSelector | object | `{}` | Additional node labels for pod assignment. |
| stackstate.components.correlate.split.httpTracing.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.correlate.split.httpTracing.replicaCount | int | `1` | Number of `httpTracing correlate` replicas. |
| stackstate.components.correlate.split.httpTracing.resources | object | `{"limits":{"cpu":null,"ephemeral-storage":null,"memory":null},"requests":{"cpu":null,"ephemeral-storage":null,"memory":null}}` | Resource allocation for pods. If not defined, will take from stackstate.components.correlate.resources |
| stackstate.components.correlate.split.httpTracing.sizing.javaHeapMemoryFraction | string | `nil` |  |
| stackstate.components.correlate.split.httpTracing.sizing.processAgentMemoryConsumption | string | `nil` |  |
| stackstate.components.correlate.split.httpTracing.tolerations | list | `[]` | Additional toleration labels for pod assignment. |
| stackstate.components.correlate.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.e2es.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.e2es.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.e2es.esDiskSpaceShare | string | `"10"` | How much disk space from ElasticSearch can use for k8s events ingestion |
| stackstate.components.e2es.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.e2es.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.e2es.image.pullPolicy | string | `""` | `pullPolicy` used for the `e2es` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.e2es.image.repository | string | `"stackstate/stackstate-kafka-to-es"` | Repository of the e2es component Docker image. |
| stackstate.components.e2es.image.tag | string | `""` | Tag used for the `e2es` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.e2es.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.e2es.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.e2es.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `e2es` pods. |
| stackstate.components.e2es.replicaCount | int | `1` | Number of `e2es` replicas. |
| stackstate.components.e2es.resources | object | `{"limits":{"cpu":"500m","ephemeral-storage":"1Gi","memory":"1500Mi"},"requests":{"cpu":"250m","ephemeral-storage":"1Mi","memory":"768Mi"}}` | Resource allocation for `e2es` pods. |
| stackstate.components.e2es.retention | int | `30` | Number of days to keep the events data on Es |
| stackstate.components.e2es.sizing.baseMemoryConsumption | string | `"300Mi"` |  |
| stackstate.components.e2es.sizing.javaHeapMemoryFraction | string | `"50"` |  |
| stackstate.components.e2es.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.healthSync.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.healthSync.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.healthSync.cache.backend | string | `"mapdb"` | Type of cache backend used by the service, possible values are mapdb, rocksdb and inmemory |
| stackstate.components.healthSync.config | string | `""` | Configuration file contents to customize the default StackState healthSync configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.healthSync.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.healthSync.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.healthSync.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.healthSync.image.imageRegistry | string | `""` | `imageRegistry` used for the `healthSync` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.healthSync.image.pullPolicy | string | `""` | `pullPolicy` used for the `healthSync` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.healthSync.image.repository | string | `"stackstate/stackstate-server"` | Repository of the healthSync component Docker image. |
| stackstate.components.healthSync.image.tag | string | `""` | Tag used for the `healthSync` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.healthSync.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.healthSync.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.healthSync.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `healthSync` pods. |
| stackstate.components.healthSync.replicaCount | int | `1` | Number of `healthSync` replicas. |
| stackstate.components.healthSync.resources | object | `{"limits":{"cpu":"1500m","ephemeral-storage":"1Gi","memory":"3500Mi"},"requests":{"cpu":"400m","ephemeral-storage":"1Mi","memory":"3500Mi"}}` | Resource allocation for `healthSync` pods. |
| stackstate.components.healthSync.sizing.baseMemoryConsumption | string | `"450Mi"` |  |
| stackstate.components.healthSync.sizing.javaHeapMemoryFraction | string | `"45"` |  |
| stackstate.components.healthSync.tmpToPVC | object | `{"storageClass":null,"volumeSize":"2Gi"}` | Whether to use PersistentVolume to store temporary files (/tmp) instead of pod ephemeral storage, empty - use pod ephemeral storage. |
| stackstate.components.healthSync.tmpToPVC.storageClass | string | `nil` | Storage class name of PersistentVolume used by /tmp directory. It stores temporary files/caches, so it should be the fastest possible. |
| stackstate.components.healthSync.tmpToPVC.volumeSize | string | `"2Gi"` | The size of the PersistentVolume for "/tmp" directory. |
| stackstate.components.healthSync.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.initializer.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.initializer.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.initializer.config | string | `""` | Configuration file contents to customize the default StackState initializer configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.initializer.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.initializer.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.initializer.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.initializer.image.imageRegistry | string | `""` | `imageRegistry` used for the `initializer` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.initializer.image.pullPolicy | string | `""` | `pullPolicy` used for the `initializer` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.initializer.image.repository | string | `"stackstate/stackstate-server"` | Repository of the initializer component Docker image. |
| stackstate.components.initializer.image.tag | string | `""` | Tag used for the `initializer` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.initializer.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.initializer.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.initializer.resources | object | `{"limits":{"cpu":"1500m","ephemeral-storage":"1Gi","memory":"1500Mi"},"requests":{"cpu":"250m","ephemeral-storage":"1Mi","memory":"512Mi"}}` | Resource allocation for `initializer` pods. |
| stackstate.components.initializer.sizing.baseMemoryConsumption | string | `"350Mi"` |  |
| stackstate.components.initializer.sizing.javaHeapMemoryFraction | string | `"65"` |  |
| stackstate.components.initializer.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.kafkaTopicCreate.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.kafkaTopicCreate.extraEnv.open | object | `{}` | Add additional environment variables to the pod |
| stackstate.components.kafkaTopicCreate.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.registry | string | `"quay.io"` | Base container image registry for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.repository | string | `"stackstate/kafka"` | Base container image repository for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.tag | string | `"3.9.1-4e2ea587-242"` | Container image tag for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.jobAnnotations | object | `{}` | Annotations for KafkaTopicCreate job. |
| stackstate.components.kafkaTopicCreate.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.kafkaTopicCreate.podAnnotations | object | `{}` | Extra annotations for kafka topic create job pods. |
| stackstate.components.kafkaTopicCreate.podLabels | object | `{}` | Extra labels for kafka topic create job pods. |
| stackstate.components.kafkaTopicCreate.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"2000Mi"}}` | Resource allocation for `kafkaTopicCreate` pods. |
| stackstate.components.kafkaTopicCreate.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| stackstate.components.kafkaTopicCreate.securityContext.fsGroup | int | `1001` | The GID (group ID) used to mount volumes |
| stackstate.components.kafkaTopicCreate.securityContext.runAsGroup | int | `1001` | The GID (group ID) of the owning user of the process |
| stackstate.components.kafkaTopicCreate.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| stackstate.components.kafkaTopicCreate.securityContext.runAsUser | int | `1001` | The UID (user ID) of the owning user of the process |
| stackstate.components.kafkaTopicCreate.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.nginxPrometheusExporter.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.registry | string | `"quay.io"` | Base container image registry for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.repository | string | `"stackstate/nginx-prometheus-exporter"` | Base container image repository for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.tag | string | `"1.5.1-f5b5d433-31"` | Container image tag for nginx-prometheus-exporter containers. |
| stackstate.components.notification.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.notification.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.notification.config | string | `""` | Configuration file contents to customize the default StackState notification configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.notification.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.notification.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.notification.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.notification.image.imageRegistry | string | `""` | `imageRegistry` used for the `notification` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.notification.image.pullPolicy | string | `""` | `pullPolicy` used for the `notification` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.notification.image.repository | string | `"stackstate/stackstate-server"` | Repository of the notification component Docker image. |
| stackstate.components.notification.image.tag | string | `""` | Tag used for the `notification` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.notification.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.notification.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.notification.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `notification` pods. |
| stackstate.components.notification.replicaCount | int | `1` | Number of `notification` replicas. |
| stackstate.components.notification.resources | object | `{"limits":{"cpu":"750m","ephemeral-storage":"1Gi","memory":"1500Mi"},"requests":{"cpu":"250m","ephemeral-storage":"1Mi","memory":"1500Mi"}}` | Resource allocation for `notification` pods. |
| stackstate.components.notification.sizing.baseMemoryConsumption | string | `"350Mi"` |  |
| stackstate.components.notification.sizing.javaHeapMemoryFraction | string | `"55"` |  |
| stackstate.components.notification.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.receiver.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.receiver.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.receiver.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.receiver.esDiskSpaceShare | string | `"90"` | How much disk space from ElasticSearch can use for k8s log ingestion |
| stackstate.components.receiver.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.receiver.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.receiver.image.imageRegistry | string | `""` | `imageRegistry` used for the `receiver` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.receiver.image.pullPolicy | string | `""` | `pullPolicy` used for the `receiver` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.receiver.image.repository | string | `"stackstate/stackstate-receiver"` | Repository of the receiver component Docker image. |
| stackstate.components.receiver.image.tag | string | `""` | Tag used for the `receiver` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.receiver.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.receiver.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.receiver.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `receiver` pods. |
| stackstate.components.receiver.replicaCount | int | `1` | Number of `receiver` replicas. |
| stackstate.components.receiver.resources | object | `{"limits":{"cpu":"3000m","ephemeral-storage":"1Gi","memory":"4Gi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Mi","memory":"4Gi"}}` | Resource allocation for `receiver` pods. |
| stackstate.components.receiver.retention | int | `7` | Number of days to keep the logs data on Es |
| stackstate.components.receiver.sizing.baseMemoryConsumption | string | `"300Mi"` |  |
| stackstate.components.receiver.sizing.javaHeapMemoryFraction | string | `"65"` |  |
| stackstate.components.receiver.split.base.affinity | object | `{}` | Additional affinity settings for pod assignment. |
| stackstate.components.receiver.split.base.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. Will merge with stackstate.components.receiver.extraEnv.open |
| stackstate.components.receiver.split.base.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.receiver.extraEnv.secret |
| stackstate.components.receiver.split.base.nodeSelector | object | `{}` | Additional node labels for pod assignment. |
| stackstate.components.receiver.split.base.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.receiver.split.base.replicaCount | int | `1` | Number of `base receiver` replicas. |
| stackstate.components.receiver.split.base.resources | object | `{"limits":{"cpu":null,"ephemeral-storage":"1Gi","memory":null},"requests":{"cpu":null,"ephemeral-storage":"1Mi","memory":null}}` | Resource allocation for pods. If not defined, will take from stackstate.components.receiver.resources |
| stackstate.components.receiver.split.base.sizing.baseMemoryConsumption | string | `nil` |  |
| stackstate.components.receiver.split.base.sizing.javaHeapMemoryFraction | string | `nil` |  |
| stackstate.components.receiver.split.base.tolerations | list | `[]` | Additional toleration labels for pod assignment. |
| stackstate.components.receiver.split.enabled | bool | `false` | Split the receiver into functional units for logs, intake and agent |
| stackstate.components.receiver.split.logs.affinity | object | `{}` | Additional affinity settings for pod assignment. |
| stackstate.components.receiver.split.logs.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. Will merge with stackstate.components.receiver.extraEnv.open |
| stackstate.components.receiver.split.logs.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.receiver.extraEnv.secret |
| stackstate.components.receiver.split.logs.nodeSelector | object | `{}` | Additional node labels for pod assignment. |
| stackstate.components.receiver.split.logs.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.receiver.split.logs.replicaCount | int | `1` | Number of `logs receiver` replicas. |
| stackstate.components.receiver.split.logs.resources | object | `{"limits":{"cpu":null,"ephemeral-storage":"1Gi","memory":null},"requests":{"cpu":null,"ephemeral-storage":"1Mi","memory":null}}` | Resource allocation for pods. If not defined, will take from stackstate.components.receiver.resources |
| stackstate.components.receiver.split.logs.sizing.javaHeapMemoryFraction | string | `nil` |  |
| stackstate.components.receiver.split.logs.sizing.logsMemoryConsumption | string | `nil` |  |
| stackstate.components.receiver.split.logs.tolerations | list | `[]` | Additional toleration labels for pod assignment. |
| stackstate.components.receiver.split.processAgent.affinity | object | `{}` | Additional affinity settings for pod assignment. |
| stackstate.components.receiver.split.processAgent.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. Will merge with stackstate.components.receiver.extraEnv.open |
| stackstate.components.receiver.split.processAgent.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.receiver.extraEnv.secret |
| stackstate.components.receiver.split.processAgent.nodeSelector | object | `{}` | Additional node labels for pod assignment. |
| stackstate.components.receiver.split.processAgent.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.receiver.split.processAgent.replicaCount | int | `1` | Number of `processAgent receiver` replicas. |
| stackstate.components.receiver.split.processAgent.resources | object | `{"limits":{"cpu":null,"ephemeral-storage":"1Gi","memory":null},"requests":{"cpu":null,"ephemeral-storage":"1Mi","memory":null}}` | Resource allocation for pods. If not defined, will take from stackstate.components.receiver.resources |
| stackstate.components.receiver.split.processAgent.sizing.javaHeapMemoryFraction | string | `nil` |  |
| stackstate.components.receiver.split.processAgent.sizing.processAgentMemoryConsumption | string | `nil` |  |
| stackstate.components.receiver.split.processAgent.tolerations | list | `[]` | Additional toleration labels for pod assignment. |
| stackstate.components.receiver.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.router.accesslog.enabled | bool | `false` | Enable access logging on the router |
| stackstate.components.router.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.router.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.router.errorlog.enabled | bool | `true` | Enable error logging on the router |
| stackstate.components.router.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.router.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.router.image.pullPolicy | string | `""` | `pullPolicy` used for the `router` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.router.image.registry | string | `"quay.io"` | Registry of the router component Docker image. |
| stackstate.components.router.image.repository | string | `"stackstate/envoy"` | Repository of the router component Docker image. |
| stackstate.components.router.image.tag | string | `"v1.31.1-92f410cd"` | Tag used for the `router` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.router.mode.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.router.mode.extraEnv.open | object | `{}` | Add additional environment variables to the pod |
| stackstate.components.router.mode.image.pullPolicy | string | `nil` | Image pull policy for router mode containers. |
| stackstate.components.router.mode.image.registry | string | `"quay.io"` | Base container image registry for router mode containers. |
| stackstate.components.router.mode.image.repository | string | `"stackstate/container-tools"` | Base container image repository for router mode containers. |
| stackstate.components.router.mode.image.tag | string | `"1.8.2-bci-517"` | Container image tag for router mode containers. |
| stackstate.components.router.mode.jobAnnotations | object | `{}` | Annotations for the router mode jobs. |
| stackstate.components.router.mode.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.router.mode.podAnnotations | object | `{}` | Extra annotations for router mode job pods. |
| stackstate.components.router.mode.podLabels | object | `{}` | Extra labels for router mode job pods. |
| stackstate.components.router.mode.resources | object | `{"limits":{"cpu":"200m","memory":"400Mi"},"requests":{"cpu":"100m","memory":"400Mi"}}` | Resource allocation for `router.mode` pods. |
| stackstate.components.router.mode.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| stackstate.components.router.mode.securityContext.fsGroup | int | `1001` | The GID (group ID) used to mount volumes |
| stackstate.components.router.mode.securityContext.runAsGroup | int | `1001` | The GID (group ID) of the owning user of the process |
| stackstate.components.router.mode.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| stackstate.components.router.mode.securityContext.runAsUser | int | `1001` | The UID (user ID) of the owning user of the process |
| stackstate.components.router.mode.status | string | `"active"` | Determines the mode being deployed. Possible values:  "maintenance": puts the system to maintenance mode, not serving the API  "active": puts the system in active mode, serving API requests  "automatic": puts the system in maintenance during helm upgrade (use a pre-hook) and activates the system using a post-hook               if deployment.compatibleWithArgoCD is set,doest the same but using PreSync PostSync from argocd               !!! This feature is considered experimental and should not be used in mission-critical setup by end-users !!! |
| stackstate.components.router.mode.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.router.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.router.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.router.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `router` pods. |
| stackstate.components.router.replicaCount | int | `1` | Number of `router` replicas. |
| stackstate.components.router.resources | object | `{"limits":{"cpu":"100m","ephemeral-storage":"1Gi","memory":"128Mi"},"requests":{"cpu":"100m","ephemeral-storage":"1Mi","memory":"128Mi"}}` | Resource allocation for `router` pods. |
| stackstate.components.router.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.server.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.server.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.server.config | string | `""` | Configuration file contents to customize the default StackState configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.server.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.server.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.server.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.server.image.imageRegistry | string | `""` | `imageRegistry` used for the `server` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.server.image.pullPolicy | string | `""` | `pullPolicy` used for the `server` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.server.image.repository | string | `"stackstate/stackstate-server"` | Repository of the server component Docker image. |
| stackstate.components.server.image.tag | string | `""` | Tag used for the `server` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.server.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.server.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.server.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `server` pods. |
| stackstate.components.server.replicaCount | int | `1` | Number of `server` replicas. |
| stackstate.components.server.resources | object | `{"limits":{"cpu":"3600m","ephemeral-storage":"1Gi","memory":"8Gi"},"requests":{"cpu":"3600m","ephemeral-storage":"1Mi","memory":"8Gi"}}` | Resource allocation for `server` pods. |
| stackstate.components.server.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.server.sizing.javaHeapMemoryFraction | string | `"45"` |  |
| stackstate.components.server.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.slicing.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.slicing.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.slicing.config | string | `""` | Configuration file contents to customize the default StackState slicing configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.slicing.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.slicing.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.slicing.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.slicing.image.imageRegistry | string | `""` | `imageRegistry` used for the `slicing` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.slicing.image.pullPolicy | string | `""` | `pullPolicy` used for the `slicing` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.slicing.image.repository | string | `"stackstate/stackstate-server"` | Repository of the slicing component Docker image. |
| stackstate.components.slicing.image.tag | string | `""` | Tag used for the `slicing` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.slicing.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.slicing.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.slicing.replicaCount | int | `1` | Number of `slicing` replicas. |
| stackstate.components.slicing.resources | object | `{"limits":{"cpu":"1500m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"250m","ephemeral-storage":"1Mi","memory":"1800Mi"}}` | Resource allocation for `slicing` pods. |
| stackstate.components.slicing.sizing.baseMemoryConsumption | string | `"500Mi"` |  |
| stackstate.components.slicing.sizing.javaHeapMemoryFraction | string | `"50"` |  |
| stackstate.components.slicing.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.state.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.state.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.state.config | string | `""` | Configuration file contents to customize the default StackState state configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.state.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.state.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.state.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.state.image.imageRegistry | string | `""` | `imageRegistry` used for the `state` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.state.image.pullPolicy | string | `""` | `pullPolicy` used for the `state` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.state.image.repository | string | `"stackstate/stackstate-server"` | Repository of the sync component Docker image. |
| stackstate.components.state.image.tag | string | `""` | Tag used for the `state` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.state.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.state.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.state.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `state` pods. |
| stackstate.components.state.replicaCount | int | `1` | Number of `state` replicas. |
| stackstate.components.state.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"1536Mi"}}` | Resource allocation for `state` pods. |
| stackstate.components.state.sizing.baseMemoryConsumption | string | `"300Mi"` |  |
| stackstate.components.state.sizing.javaHeapMemoryFraction | string | `"65"` |  |
| stackstate.components.state.tmpToPVC | object | `{"storageClass":null,"volumeSize":"2Gi"}` | Whether to use PersistentVolume to store temporary files (/tmp) instead of pod ephemeral storage, empty - use pod ephemeral storage. |
| stackstate.components.state.tmpToPVC.storageClass | string | `nil` | Storage class name of PersistentVolume used by /tmp directory. It stores temporary files/caches, so it should be the fastest possible. |
| stackstate.components.state.tmpToPVC.volumeSize | string | `"2Gi"` | The size of the PersistentVolume for "/tmp" directory. |
| stackstate.components.state.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.sync.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.sync.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.sync.cache.backend | string | `"mapdb"` | Type of cache backend used by the service, possible values are mapdb, rocksdb and inmemory |
| stackstate.components.sync.config | string | `""` | Configuration file contents to customize the default StackState sync configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. |
| stackstate.components.sync.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.sync.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.sync.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.sync.image.imageRegistry | string | `""` | `imageRegistry` used for the `sync` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.sync.image.pullPolicy | string | `""` | `pullPolicy` used for the `sync` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.sync.image.repository | string | `"stackstate/stackstate-server"` | Repository of the sync component Docker image. |
| stackstate.components.sync.image.tag | string | `""` | Tag used for the `sync` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.sync.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.sync.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.sync.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `sync` pods. |
| stackstate.components.sync.replicaCount | int | `1` | Number of `sync` replicas. |
| stackstate.components.sync.resources | object | `{"limits":{"cpu":"3000m","ephemeral-storage":"1Gi","memory":"4Gi"},"requests":{"cpu":"750m","ephemeral-storage":"1Mi","memory":"3Gi"}}` | Resource allocation for `sync` pods. |
| stackstate.components.sync.sizing.baseMemoryConsumption | string | `"400Mi"` |  |
| stackstate.components.sync.sizing.javaHeapMemoryFraction | string | `"60"` |  |
| stackstate.components.sync.tmpToPVC | object | `{"storageClass":null,"volumeSize":"2Gi"}` | Whether to use PersistentVolume to store temporary files (/tmp) instead of pod ephemeral storage, empty - use pod ephemeral storage. |
| stackstate.components.sync.tmpToPVC.storageClass | string | `nil` | Storage class name of PersistentVolume used by /tmp directory. It stores temporary files/caches, so it should be the fastest possible. |
| stackstate.components.sync.tmpToPVC.volumeSize | string | `"2Gi"` | The size of the PersistentVolume for "/tmp" directory. |
| stackstate.components.sync.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.ui.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.ui.agentMetricsFilter | string | `"[\"nginx*\"]"` |  |
| stackstate.components.ui.debug | bool | `false` |  |
| stackstate.components.ui.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.ui.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.ui.image.imageRegistry | string | `""` | `imageRegistry` used for the `ui` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.ui.image.pullPolicy | string | `""` | `pullPolicy` used for the `ui` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.ui.image.repository | string | `"stackstate/stackstate-ui"` | Repository of the ui component Docker image. |
| stackstate.components.ui.image.tag | string | `""` | Tag used for the `ui` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.ui.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.ui.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.ui.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `ui` pods. |
| stackstate.components.ui.replicaCount | string | `nil` | Number of `ui` replicas. |
| stackstate.components.ui.resources | object | `{"limits":{"cpu":"50m","ephemeral-storage":"1Gi","memory":"64Mi"},"requests":{"cpu":"50m","ephemeral-storage":"1Mi","memory":"64Mi"}}` | Resource allocation for `ui` pods. |
| stackstate.components.ui.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| stackstate.components.ui.securityContext.fsGroup | int | `101` | The GID (group ID) used to mount volumes |
| stackstate.components.ui.securityContext.runAsGroup | int | `101` | The GID (group ID) of the owning user of the process |
| stackstate.components.ui.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| stackstate.components.ui.securityContext.runAsUser | int | `101` | The UID (user ID) of the owning user of the process |
| stackstate.components.ui.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.vmagent.affinity | object | `{"podAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app.kubernetes.io/component","operator":"In","values":["receiver"]},{"key":"app.kubernetes.io/instance","operator":"In","values":["stackstate"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":80}]}}` | Affinity settings for vmagent pod. |
| stackstate.components.vmagent.agentMetricsFilter | string | `"[\"vm*\", \"go*\"]"` |  |
| stackstate.components.vmagent.extraArgs | object | `{}` |  |
| stackstate.components.vmagent.fullNameOverride | string | `"suse-observability-vmagent"` | Name for the service |
| stackstate.components.vmagent.image.repository | string | `"stackstate/vmagent"` |  |
| stackstate.components.vmagent.image.tag | string | `"v1.109.0-e6dc53fb"` |  |
| stackstate.components.vmagent.persistence.size | string | `"10Gi"` |  |
| stackstate.components.vmagent.persistence.storageClass | string | `nil` |  |
| stackstate.components.vmagent.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `vmagent` pods. |
| stackstate.components.vmagent.resources | object | `{"limits":{"cpu":"200m","ephemeral-storage":"1Gi","memory":"512Mi"},"requests":{"cpu":"200m","ephemeral-storage":"1Mi","memory":"256Mi"}}` | Resource allocation for vmagent pod. |
| stackstate.components.workloadObserver.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.workloadObserver.enabled | bool | `true` | Enable/disable the workload observer |
| stackstate.components.workloadObserver.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.workloadObserver.image.imageRegistry | string | `""` | `imageRegistry` used for the `workloadObserver` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.workloadObserver.image.pullPolicy | string | `""` | `pullPolicy` used for the `workloadObserver` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.workloadObserver.image.repository | string | `"stackstate/workload-observer"` | Repository of the workloadObserver component Docker image. |
| stackstate.components.workloadObserver.image.tag | string | `"e7f95a08-63"` | Tag used for the `workloadObserver` component Docker image.. |
| stackstate.components.workloadObserver.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.workloadObserver.persistence.size | string | `"1Gi"` |  |
| stackstate.components.workloadObserver.persistence.storageClass | string | `nil` |  |
| stackstate.components.workloadObserver.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.workloadObserver.poddisruptionbudget | object | `{"maxUnavailable":1}` | Number of `workloadObserver` replicas. |
| stackstate.components.workloadObserver.resources | object | `{"limits":{"cpu":"50m","ephemeral-storage":"1Gi","memory":"128Mi"},"requests":{"cpu":"20m","ephemeral-storage":"1Mi","memory":"24Mi"}}` | Resource allocation for `workloadObserver` pods. |
| stackstate.components.workloadObserver.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.deployment.edition | string | `"Prime"` | StackState edition, one of 'Community' or 'Prime' |
| stackstate.deployment.mode | string | `"SelfHosted"` | Deployment mode of StackState, possible values are 'Saas' and 'SelfHosted' |
| stackstate.email | object | `{"additionalProperties":{"mail.smtp.auth":"true","mail.smtp.starttls.enable":"true"},"enabled":false,"sender":"","server":{"auth":{"fromExternalSecret":null,"password":"","username":""},"host":"","port":587,"protocol":"smtp"}}` | Email configuration for StackState |
| stackstate.email.enabled | bool | `false` | Enable email notifications |
| stackstate.email.sender | string | `""` | Email sender mail address |
| stackstate.email.server.auth.password | string | `""` | Email server password |
| stackstate.email.server.auth.username | string | `""` | Email server username |
| stackstate.email.server.host | string | `""` | Email server host |
| stackstate.email.server.port | int | `587` | Email server port |
| stackstate.email.server.protocol | string | `"smtp"` | Email server protocol |
| stackstate.experimental | object | `{}` | Enable experimental features in StackState. Deprecated, use `stackstate.features` instead. |
| stackstate.features.role-k8s-authz | boolean | `true` | Deploy the Role(s) to populate permissions on Suse Observability |
| stackstate.features.server.split | boolean | `true` | Run a single service server or split in multiple sub services as api, state .... |
| stackstate.features.storeTransactionLogsToPVC.enabled | boolean | `false` | Whether the transaction logs for some services, API, Checks, HealthSync,State and Sync have to be stored to PVCs instead of pod ephemeral storage. |
| stackstate.features.storeTransactionLogsToPVC.storageClass | string | `nil` | Storage class name of PersistentVolume used by transaction logs. |
| stackstate.features.storeTransactionLogsToPVC.volumeSize | string | `"600Mi"` | The size of the persistent volume for the transaction logs. |
| stackstate.features.traces | boolean | `true` | Enable new traces UI and API |
| stackstate.instanceDebugApi.enabled | bool | `false` |  |
| stackstate.java | object | `{"trustStore":null,"trustStoreBase64Encoded":null,"trustStorePassword":null}` | Extra Java configuration for StackState |
| stackstate.java.trustStore | string | `nil` | Java TrustStore (cacerts) file to use |
| stackstate.java.trustStoreBase64Encoded | string | `nil` | Base64 encoded Java TrustStore (cacerts) file to use. Ignored if stackstate.java.trustStore is set. |
| stackstate.java.trustStorePassword | string | `nil` | Password to access the Java TrustStore (cacerts) file |
| stackstate.k8sAuthorization.enabled | bool | `true` |  |
| stackstate.license.fromExternalSecret | string | `nil` | Use an external secret for the license key. This suppresses secret creation by StackState and gets the data from the secret with the provided name. |
| stackstate.license.key | string | `nil` | **PROVIDE YOUR LICENSE KEY HERE** The StackState license key needed to start the server. |
| stackstate.receiver.baseUrl | string | `nil` | **DEPRECATED** Use stackstate.baseUrl instead |
| stackstate.stackpacks.image.deploymentModeOverride | string | `""` | Use the stackpacks from another deployment mode than StackState is running in |
| stackstate.stackpacks.image.pullPolicy | string | `""` | `pullPolicy` used for the `stackpacks` Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.stackpacks.image.registry | string | `"quay.io"` | `registry` used for the `stackpacks` Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.stackpacks.image.repository | string | `"stackstate/stackpacks"` | Repository of the `stackpacks` Docker image. |
| stackstate.stackpacks.image.version | string | `"20260211160822-master-32fe626"` | Version used for the `stackpacks` Docker image, the tag is build from the version and the stackstate edition + deployment mode |
| stackstate.stackpacks.installed | list | `[]` | Specify a list of stackpacks to be always installed including their configuration, for an example see [Auto-installing StackPacks](#auto-installing-stackpacks) |
| stackstate.stackpacks.localpvc.size | string | `"1Gi"` | Size of the Persistent Volume Claim (PVC) used to persist stackpacks when there's no HDFS |
| stackstate.stackpacks.localpvc.storageClass | string | `nil` |  |
| stackstate.stackpacks.pvc.size | string | `"1Gi"` | Size of the Persistent Volume Claim (PVC) used to copy stackpacks from the Docker image. |
| stackstate.stackpacks.pvc.storageClass | string | `nil` |  |
| stackstate.stackpacks.source | string | `"docker-image"` | Source of the stackpacks, for now just the docker-image. |
| stackstate.stackpacks.updateInterval | string | `"5 minutes"` |  |
| stackstate.stackpacks.upgradeOnStartup | list | `[]` | Specify additional stackpacks that will, on startup only, be upgraded to the latest version available. Note: The following StackPacks are automatically upgraded with SUSE Observability: kubernetes-v2, open-telemetry, stackstate-k8s-agent-v2, aad-v2, stackstate, plus either prime-kubernetes or community-kubernetes (based on stackstate.deployment.edition). Additional StackPacks declared in upgradeOnStartup will be merged with these defaults. |
| stackstate.topology.retentionHours | integer | `nil` | Number of hours topology will be retained. |
| stackstate.ui.defaultTimeRange | string | `nil` | Default time range  in the UI. One of LAST_5_MINUTES, LAST_15_MINUTES, LAST_30_MINUTES, LAST_1_HOUR, LAST_3_HOURS, LAST_6_HOURS, LAST_12_HOURS, LAST_24_HOURS, LAST_2_DAYS. No value or an unsupported value will automatically fall-back to LAST_1_HOUR. |
| victoria-metrics-0.backup.bucketName | string | `"sts-victoria-metrics-backup"` | Name of the MinIO bucket where Victoria Metrics backups are stored. |
| victoria-metrics-0.backup.s3Prefix | string | `"victoria-metrics-0"` |  |
| victoria-metrics-0.backup.scheduled.schedule | string | `"25 * * * *"` | Cron schedule for automatic backups of Victoria Metrics |
| victoria-metrics-0.enabled | bool | `true` |  |
| victoria-metrics-0.rbac.namespaced | bool | `true` |  |
| victoria-metrics-0.rbac.pspEnabled | bool | `false` |  |
| victoria-metrics-0.server.affinity | object | `{}` | Affinity settings for Victoria Metrics pod |
| victoria-metrics-0.server.extraArgs | object | `{"dedup.minScrapeInterval":"1ms","maxLabelsPerTimeseries":60,"search.cacheTimestampOffset":"10m"}` | Extra arguments for Victoria Metrics |
| victoria-metrics-0.server.extraLabels | object | `{"app.kubernetes.io/part-of":"suse-observability"}` | Extra labels for Victoria Metrics StatefulSet |
| victoria-metrics-0.server.fullnameOverride | string | `"suse-observability-victoria-metrics-0"` | Full name override |
| victoria-metrics-0.server.persistentVolume.size | string | `"250Gi"` | Size of storage for Victoria Metrics, ideally 20% of free space remains available at all times |
| victoria-metrics-0.server.podAnnotations | object | `{"ad.stackstate.com/victoria-metrics-0-server.check_names":"[\"openmetrics\"]","ad.stackstate.com/victoria-metrics-0-server.init_configs":"[{}]","ad.stackstate.com/victoria-metrics-0-server.instances":"[ { \"prometheus_url\": \"http://%%host%%:8428/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"vm*\", \"go*\"] } ]","ad.stackstate.com/vmbackup.check_names":"[\"openmetrics\"]","ad.stackstate.com/vmbackup.init_configs":"[{}]","ad.stackstate.com/vmbackup.instances":"[ { \"prometheus_url\": \"http://%%host%%:9746/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"supercronic_*\"] } ]"}` | Annotations for Victoria Metrics server pod |
| victoria-metrics-0.server.podLabels | object | `{"app.kubernetes.io/part-of":"suse-observability","stackstate-service":"victoriametrics"}` | Extra labels for Victoria Metrics pod |
| victoria-metrics-0.server.resources.limits.cpu | int | `1` |  |
| victoria-metrics-0.server.resources.limits.memory | string | `"4Gi"` |  |
| victoria-metrics-0.server.resources.requests.cpu | string | `"300m"` |  |
| victoria-metrics-0.server.resources.requests.memory | string | `"3584Mi"` |  |
| victoria-metrics-0.server.retentionPeriod | int | `1` | How long is data retained, when changing also consider updating the persistentVolume.size to match. The following optional suffixes are supported: h (hour), d (day), w (week), y (year). If suffix isn't set, then the duration is counted in months (default 1) |
| victoria-metrics-0.server.scrape.enabled | bool | `false` | StackState doesn't use the scraping of VictoriaMetrics |
| victoria-metrics-0.server.securityContext | object | `{"fsGroup":65534,"runAsGroup":65534,"runAsUser":65534}` | Custom security context settings for running as non-root |
| victoria-metrics-0.server.serviceMonitor.enabled | bool | `false` | If `true`, creates a Prometheus Operator `ServiceMonitor` |
| victoria-metrics-0.server.serviceMonitor.extraLabels | object | `{}` | Add extra labels to target a specific prometheus instance |
| victoria-metrics-0.server.serviceMonitor.interval | string | `"15s"` | Scrape interval for service monitor |
| victoria-metrics-1.backup.bucketName | string | `"sts-victoria-metrics-backup"` | Name of the MinIO bucket where Victoria Metrics backups are stored. |
| victoria-metrics-1.backup.s3Prefix | string | `"victoria-metrics-1"` | Prefix (dir name) used to store backup files, we may have multiple instances of Victoria Metrics, each of them should be stored into their own directory. |
| victoria-metrics-1.backup.scheduled.schedule | string | `"35 * * * *"` | Cron schedule for automatic backups of Victoria Metrics |
| victoria-metrics-1.enabled | bool | `true` |  |
| victoria-metrics-1.rbac.namespaced | bool | `true` |  |
| victoria-metrics-1.rbac.pspEnabled | bool | `false` |  |
| victoria-metrics-1.server.affinity | object | `{}` | Affinity settings for Victoria Metrics pod |
| victoria-metrics-1.server.extraArgs | object | `{"dedup.minScrapeInterval":"1ms","maxLabelsPerTimeseries":60}` | Extra arguments for Victoria Metrics |
| victoria-metrics-1.server.extraLabels."app.kubernetes.io/part-of" | string | `"suse-observability"` |  |
| victoria-metrics-1.server.fullnameOverride | string | `"suse-observability-victoria-metrics-1"` | Full name override |
| victoria-metrics-1.server.persistentVolume.size | string | `"250Gi"` | Size of storage for Victoria Metrics, ideally 20% of free space remains available at all times |
| victoria-metrics-1.server.podAnnotations | object | `{"ad.stackstate.com/victoria-metrics-0-server.check_names":"[\"openmetrics\"]","ad.stackstate.com/victoria-metrics-0-server.init_configs":"[{}]","ad.stackstate.com/victoria-metrics-0-server.instances":"[ { \"prometheus_url\": \"http://%%host%%:8428/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"*\"] } ]","ad.stackstate.com/vmbackup.check_names":"[\"openmetrics\"]","ad.stackstate.com/vmbackup.init_configs":"[{}]","ad.stackstate.com/vmbackup.instances":"[ { \"prometheus_url\": \"http://%%host%%:9746/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"supercronic_*\"] } ]"}` | Annotations for Victoria Metrics server pod |
| victoria-metrics-1.server.podLabels | object | `{"app.kubernetes.io/part-of":"suse-observability","stackstate-service":"victoriametrics"}` | Extra arguments for Victoria Metrics pod |
| victoria-metrics-1.server.resources.limits.cpu | int | `1` |  |
| victoria-metrics-1.server.resources.limits.memory | string | `"4Gi"` |  |
| victoria-metrics-1.server.resources.requests.cpu | string | `"300m"` |  |
| victoria-metrics-1.server.resources.requests.memory | string | `"3584Mi"` |  |
| victoria-metrics-1.server.retentionPeriod | int | `1` | How long is data retained, when changing also consider updating the persistentVolume.size to match. The following optional suffixes are supported: h (hour), d (day), w (week), y (year). If suffix isn't set, then the duration is counted in months (default 1) |
| victoria-metrics-1.server.scrape.enabled | bool | `false` | StackState doesn't use the scraping of VictoriaMetrics |
| victoria-metrics-1.server.securityContext | object | `{"fsGroup":65534,"runAsGroup":65534,"runAsUser":65534}` | Custom security context settings for running as non-root |
| victoria-metrics-1.server.serviceMonitor.enabled | bool | `false` | If `true`, creates a Prometheus Operator `ServiceMonitor` |
| victoria-metrics-1.server.serviceMonitor.extraLabels | object | `{}` | Add extra labels to target a specific prometheus instance |
| victoria-metrics-1.server.serviceMonitor.interval | string | `"15s"` | Scrape interval for service monitor |
| victoria-metrics.restore.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for `vmrestore` containers. |
| victoria-metrics.restore.image.registry | string | `"quay.io"` | Base container image registry for 'vmrestore' containers. |
| victoria-metrics.restore.image.repository | string | `"stackstate/vmrestore"` | Base container image repository for 'vmrestore' containers. |
| victoria-metrics.restore.image.tag | string | `"v1.109.0-4af7dbf5"` | Container image tag for 'vmrestore' containers. |
| victoria-metrics.restore.securityContext.enabled | bool | `true` |  |
| victoria-metrics.restore.securityContext.fsGroup | int | `65534` |  |
| victoria-metrics.restore.securityContext.runAsGroup | int | `65534` |  |
| victoria-metrics.restore.securityContext.runAsNonRoot | bool | `true` |  |
| victoria-metrics.restore.securityContext.runAsUser | int | `65534` |  |
| zookeeper.autopurge | object | `{"purgeInterval":3,"snapRetainCount":5}` | configurations of ZooKeeper auto purge, it deletes old snapshot and log files. ClickHouse creates a lot of operation and it should be purged to avoud out of disk space. |
| zookeeper.commonLabels."app.kubernetes.io/part-of" | string | `"suse-observability"` |  |
| zookeeper.containerSecurityContext.allowPrivilegeEscalation | bool | `false` |  |
| zookeeper.containerSecurityContext.capabilities.drop[0] | string | `"ALL"` |  |
| zookeeper.containerSecurityContext.enabled | bool | `true` |  |
| zookeeper.containerSecurityContext.runAsNonRoot | bool | `true` |  |
| zookeeper.containerSecurityContext.runAsUser | int | `1001` |  |
| zookeeper.containerSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| zookeeper.customLivenessProbe.exec.command[0] | string | `"/bin/bash"` |  |
| zookeeper.customLivenessProbe.exec.command[1] | string | `"-c"` |  |
| zookeeper.customLivenessProbe.exec.command[2] | string | `"echo \"ruok\" | timeout 2 nc -w 2 -q 1 localhost 2181 | grep imok"` |  |
| zookeeper.customLivenessProbe.failureThreshold | int | `6` |  |
| zookeeper.customLivenessProbe.initialDelaySeconds | int | `30` |  |
| zookeeper.customLivenessProbe.periodSeconds | int | `10` |  |
| zookeeper.customLivenessProbe.successThreshold | int | `1` |  |
| zookeeper.customLivenessProbe.timeoutSeconds | int | `5` |  |
| zookeeper.customReadinessProbe.exec.command[0] | string | `"/bin/bash"` |  |
| zookeeper.customReadinessProbe.exec.command[1] | string | `"-c"` |  |
| zookeeper.customReadinessProbe.exec.command[2] | string | `"echo \"ruok\" | timeout 2 nc -w 2 -q 1 localhost 2181 | grep imok"` |  |
| zookeeper.customReadinessProbe.failureThreshold | int | `6` |  |
| zookeeper.customReadinessProbe.initialDelaySeconds | int | `5` |  |
| zookeeper.customReadinessProbe.periodSeconds | int | `10` |  |
| zookeeper.customReadinessProbe.successThreshold | int | `1` |  |
| zookeeper.customReadinessProbe.timeoutSeconds | int | `5` |  |
| zookeeper.enabled | bool | `true` | Enable / disable chart-based Zookeeper. |
| zookeeper.externalServers | string | `""` | If `zookeeper.enabled` is set to `false`, use this list of external Zookeeper servers instead. |
| zookeeper.fourlwCommandsWhitelist | string | `"mntr, ruok, stat, srvr"` | Zookeeper four-letter-word (FLW) commands that are enabled. |
| zookeeper.fullnameOverride | string | `"suse-observability-zookeeper"` | Name override for Zookeeper child chart. **Don't change unless otherwise specified; this is a Helm v2 limitation, and will be addressed in a later Helm v3 chart.** |
| zookeeper.heapSize | int | `400` | HeapSize Size (in MB) for the Java Heap options (Xmx and Xms) |
| zookeeper.image.registry | string | `"quay.io"` | ZooKeeper image registry |
| zookeeper.image.repository | string | `"stackstate/zookeeper"` | ZooKeeper image repository |
| zookeeper.image.tag | string | `"3.9.3-eda4e09e-244"` | ZooKeeper image tag |
| zookeeper.jvmFlags | string | `"-Djute.maxbuffer=2097150"` |  |
| zookeeper.livenessProbe.enabled | bool | `false` | it must be disabled to apply the custom probe, the probe adds "-q" option to nc to wait 1sec until close the connection, it fixes problem of failing the probed |
| zookeeper.metrics.enabled | bool | `true` | Enable / disable Zookeeper Prometheus metrics. |
| zookeeper.metrics.serviceMonitor | object | `{"enabled":false,"selector":{}}` |  |
| zookeeper.metrics.serviceMonitor.enabled | bool | `false` | Enable creation of `ServiceMonitor` objects for Prometheus operator. |
| zookeeper.metrics.serviceMonitor.selector | object | `{}` | Default selector to use to target a certain Prometheus instance. |
| zookeeper.pdb.create | bool | `true` |  |
| zookeeper.pdb.maxUnavailable | int | `1` |  |
| zookeeper.pdb.minAvailable | string | `""` |  |
| zookeeper.persistence.size | string | `"8Gi"` | Size of the PVC for Zookeeper data. Default is 8Gi, will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| zookeeper.podAnnotations | object | `{"ad.stackstate.com/zookeeper.check_names":"[\"openmetrics\"]","ad.stackstate.com/zookeeper.init_configs":"[{}]","ad.stackstate.com/zookeeper.instances":"[ { \"prometheus_url\": \"http://%%host%%:9141/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"*\"] } ]"}` | Annotations for ZooKeeper pod. |
| zookeeper.podLabels."app.kubernetes.io/part-of" | string | `"suse-observability"` |  |
| zookeeper.readinessProbe.enabled | bool | `false` | it must be disabled to apply the custom probe, the probe adds "-q" option to nc to wait 1sec until close the connection, it fixes problem of failing the probed |
| zookeeper.replicaCount | string | `nil` | Default amount of Zookeeper replicas to provision. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| zookeeper.resources.limits.cpu | string | `"250m"` |  |
| zookeeper.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| zookeeper.resources.limits.memory | string | `"640Mi"` | Allocated memory should be bigger than JVM Heap Size (env var ZOO_HEAP_SIZE) and space used by Off-Heap Memory (e.g. Metaspace) |
| zookeeper.resources.requests.cpu | string | `"100m"` |  |
| zookeeper.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| zookeeper.resources.requests.memory | string | `"640Mi"` | Allocated memory should be bigger than JVM Heap Size (env var ZOO_HEAP_SIZE) and space used by Off-Heap Memory (e.g. Metaspace) |

## Authentication

For more details on configuring authentication please refer to the [StackState documentation](https://docs.stackstate.com).

### Configuring OpenId connect

Create a `oidc_values.yaml similar to the example below and add it as an additional argument to the installation:

```
helm install \
  --values oidc_values.yaml \
  ... \
stackstate/stackstate
```

Example:

```yaml
stackstate:
  authentication:
    oidc:
      clientId: stackstate-client-id
      secret: "some-secret"
      discoveryUri: http://oidc-provider/.well-known/openid-configuration
      authenticationMethod: client_secret_basic
      jwsAlgorithm: RS256
      scope: ["email", "fullname"]
      jwtClaims:
        usernameField: email
        groupsField: groups
      customParameters:
        access_type: offline # Custom request parameter
```

### Configuring Rancher authentication

```
NOTE: for internal use - SUSE ECM development - only. This capability is still under development and not for any other use yet.
```

Rancher authentication uses OpenId Connect (OIDC).

This authentication mechanism in SUSE Observability abstracts away certain OIDC config specific to Rancher to simplify the config required to set up the authentication method between Rancher and SUSE Observability.

To use it create a `rancher_auth_values.yaml` similar to the example below and add it as an additional argument to the installation:

```
helm install \
  --values rancher_auth_values.yaml \
  ... \
stackstate/stackstate
```

Example that sets up Rancher authentication.
```yaml
stackstate:
  authentication:
    rancher:
      clientId: "oidc-client"
      secret: "oidc-client-secret"
      baseUrl: "https://rancher-host"
    roles:
      admin: [ "Default Admin" ] # for now, we map here the display names in Rancher to the SO role
```
You can override and extend some of the OIDC config for Rancher with the following fields:
- `discoveryUri`
- `redirectUri`
- `customParams`

If you need to disable TLS verification due to a setup not using verifiable SSL certificates, you can disable SSL checks with some application config (don't use in production):
```yaml
stackstate:
  components:
    server:
      extraEnv:
        open:
          CONFIG_FORCE_stackstate_misc_sslCertificateChecking: false
```

### Configuring Keycloak

Create a `keycloak_values.yaml similar to the example below and add it as an additional argument to the installation:

```
helm install \
  --values keycloak_values.yaml \
  ... \
stackstate/stackstate
```

Example:
```yaml
stackstate:
  authentication:
    keycloak:
      url: http://keycloak-server/auth
      realm: test
      clientId: stackstate-client-id
      secret: "some-secret"
      authenticationMethod: client_secret_basic
      jwsAlgorithm: RS256
```

If the defaults of Keycloak don't suit your needs you can extend the yaml to select a different field as username and add groups/roles from fields other than the `roles` in keycloak:
```
stackstate:
  authentication:
    keycloak:
      jwtClaims:
        usernameField: email
        groupsField: groups
```

### Configuring LDAP

To use LDAP create a ldap_values.yaml similar to the example below (update for your LDAP configuration of course). Next to these keys there are 2 optional values that can be set but usually need to be read from a file so you'd specify them on the helm command line (see below):
* `stackstate.authentication.ldap.ssl.trustStore`: The Certificate Truststore to verify server certificates against
* `stackstate.authentication.ldap.ssl.trustCertificates`: The client Certificate trusted by the server (supports PEM, DER and PKCS7 formats)

There are also Base64 Encoded analogues of these values. They are ignored if `trustCertificates` and/or `trustStore` are set:
* `stackstate.authentication.ldap.ssl.trustStoreBase64Encoded`
* `stackstate.authentication.ldap.ssl.trustCertificatesBase64Encoded`
**Note: The reason for `*Base64Encoded` values is this is the only way to upload binary files via KOTS Config**

Only one of `trustCertificates`/`trustCertificatesBase64Encoded` or `trustStore`/`trustStoreBase64Encoded` will be used, `trustCertificates` takes precedence over `trustStore`, and `trustStore`/`trustCertificates` takes precedence over `trustStoreBase64Encoded`/`trustCertificatesBase64Encoded` respectively.

In order to search the groups that a user belongs to and from those get the roles the user can have in StackState we need to config values `rolesKey` and `groupMemberKey`.

Those values in the end will be used to form a filter (and extract the relevant attribute) that looks like:
```({groupMemberKey}=email=admin@sts.com,ou=employees,dc=stackstate,dc=com)```

This returns an entry similar to this:
```dn: {rolesKey}=stackstate-admin,ou=Group,ou=employee,o=stackstate,cn=people,dc=stackstate,dc=com```

Via the {rolesKey} we will get `stackstate-admin` as the role.

Note that the order of the parameters is of importance.

```yaml
stackstate:
  authentication:
    ldap:
      host: ldap-server
      port: 10636 # Standard LDAP SSL port, 10398 for plain LDAP
      bind:
        dn: ou=acme,dc=com # The bind DN to use to authenticate to LDAP
        password: foobar   # The bind password to use to authenticate to LDAP
      ssl:
        type: ssl          # Optional: The SSL Connection type to use to connect to LDAP (Either `ssl` or `starttls`)
      userQuery:
        emailKey: email
        usernameKey: cn
        parameters:
          - ou: employees
          - dc: stackstate
          - dc: com
      groupQuery:
        groupMemberKey: member
        rolesKey: cn
        parameters:
          - ou: groups
          - dc: stackstate
          - dc: com
```

The `trustStore` and `trustCertificates` values need to be set from the command line, as they typically contain binary data. A sample command for this looks like:

```shell
helm install \
--set-file stackstate.authentication.ldap.ssl.trustStore=./ldap-cacerts \
--set-file stackstate.authentication.ldap.ssl.trustCertificates=./ldap-certificate.pem \
--values ldap_values.yaml \
... \
stackstate/stackstate
```

### Configuring file based authentication

If you don't have an external identity provider you can configure users and there roles directly in StackState via a configuration file (a change will result in a restart of the API).

To use this create a `file_auth_values.yaml similar to the example below and add it as an additional argument to the installation:

```
helm install \
  --values file_auth_values.yaml \
  ... \
stackstate/stackstate
```

Example that creates 4 different users with the standard roles provided by Stackstate (see our [docs](https://docs.stackstate.com)):
```
stackstate:
  authentication:
    file:
      logins:
        - username: administrator
          passwordHash: 098f6bcd4621d373cade4e832627b4f6
          roles: [stackstate-admin]
        - username: guest1
          passwordHash: 098f6bcd4621d373cade4e832627b4f6
          roles: [stackstate-guest]
        - username: guest2
          passwordHash: 098f6bcd4621d373cade4e832627b4f6
          roles: [ stackstate-guest ]
        - username: maintainer
          passwordHash: 098f6bcd4621d373cade4e832627b4f6
          roles: [stackstate-power-user, stackstate-guest]
```

## Auto-installing StackPacks
It can be useful to have some StackPacks always installed in StackState. For example the Kuberentes StackPack configured for the cluster running StackState. For that purpose the value `stackstate.stackpacks.installed` can be used to specify the StackPacks that should be installed (by name) and their (required) configuration. As an example here the Kubernetes StackPack will be pre-installed:
```
stackstate:
  stackpacks:
    installed:
      - name: kubernetes
        configuration:
          kubernetes_cluster_name: test
```
