# suse-observability

Helm chart for SUSE Observability

Current chart version is `2.9.1-pre.128`

**Homepage:** <https://gitlab.com/stackvista/stackstate.git>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../anomaly-detection/ | anomaly-detection | 5.2.0-snapshot.183 |
| file://../clickhouse/ | clickhouse | 3.6.9-suse-observability.28 |
| file://../common/ | common | * |
| file://../elasticsearch/ | elasticsearch | 8.19.4-stackstate.22 |
| file://../hbase/ | hbase | 0.2.138 |
| file://../kafka/ | kafka | 19.1.3-suse-observability.26 |
| file://../kafkaup-operator/ | kafkaup-operator | 0.1.29 |
| file://../kubernetes-rbac-agent/ | kubernetes-rbac-agent | 0.0.27 |
| file://../opentelemetry-collector | opentelemetry-collector | 0.108.0-stackstate.30 |
| file://../pull-secret/ | pull-secret | * |
| file://../suse-observability-sizing/ | suse-observability-sizing | 0.1.16 |
| file://../victoria-metrics-single/ | victoria-metrics-0(victoria-metrics-single) | 0.8.53-stackstate.51 |
| file://../victoria-metrics-single/ | victoria-metrics-1(victoria-metrics-single) | 0.8.53-stackstate.51 |
| file://../zookeeper/ | zookeeper | 8.1.2-suse-observability.25 |

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
| ai | object | `{"assistant":{"anthropic":{"apiKey":"","fromExternalSecret":{"key":"ANTHROPIC_API_KEY","name":""}},"bedrock":{"awsRegion":"eu-west-1"},"enabled":false,"provider":"bedrock"},"platformOptimization":{"enabled":false}}` | Note: These features are General Availability (GA) but disabled by default because you need to provide an LLM to be able to use them. |
| ai.assistant | object | `{"anthropic":{"apiKey":"","fromExternalSecret":{"key":"ANTHROPIC_API_KEY","name":""}},"bedrock":{"awsRegion":"eu-west-1"},"enabled":false,"provider":"bedrock"}` | AI Assistant & MCP Server. Entitlement: Included with SUSE Observability (Rancher Prime) |
| ai.assistant.anthropic.apiKey | string | `""` | Anthropic API key used when `ai.assistant.provider=anthropic`. |
| ai.assistant.anthropic.fromExternalSecret.key | string | `"ANTHROPIC_API_KEY"` | Key in the external secret that contains the Anthropic API key. |
| ai.assistant.anthropic.fromExternalSecret.name | string | `""` | Existing secret name that contains the Anthropic API key. When set, the chart will not create an internal secret for `ANTHROPIC_API_KEY`. |
| ai.assistant.bedrock.awsRegion | string | `"eu-west-1"` | AWS region used when `ai.assistant.provider=bedrock`. |
| ai.assistant.enabled | bool | `false` | Enables the AI Assistant UI, the dedicated backend process, and the MCP (Model Context Protocol) Server for Observability. |
| ai.assistant.provider | string | `"bedrock"` | LLM provider for AI Assistant. Possible values: `bedrock` or `anthropic`. |
| ai.platformOptimization | object | `{"enabled":false}` | Automatic Troubleshooting & Remediation. Entitlement: Requires 'SUSE Platform Optimization' add-on (part of Rancher Suite). Enabling this flag indicates you have the appropriate license for Platform Optimization. |
| ai.platformOptimization.enabled | bool | `false` | Enables advanced AI-driven automatic troubleshooting extensions. This builds upon the AI Assistant framework to provide proactive issue resolution. |
| anomaly-detection.enabled | bool | `false` | Enables anomaly detection chart |
| anomaly-detection.image.registry | string | `"quay.io"` | Base container image registry for all containers, except for the wait container |
| anomaly-detection.image.spotlightRepository | string | `"stackstate/spotlight"` | Repository of the spotlight Docker image. |
| anomaly-detection.image.tag | string | `"5.2.0-snapshot.192"` | the chart image tag, e.g. 4.1.3-latest |
| anomaly-detection.stackstate.instance | string | `"{{ include \"stackstate.router.endpoint\" . }}"` | **Required Stackstate instance URL. |
| backup.additionalLogging | string | `""` | Additional logback config for backup components |
| backup.configuration.bucketName | string | `"sts-configuration-backup"` | Name of the storage bucket to store configuration backups. |
| backup.configuration.maxLocalFiles | int | `10` | The maximum number of configuration backup files stored on the PVC for the configuration backup (which is only of limited size, see backup.configuration.scheduled.pvc.size. |
| backup.configuration.s3Prefix | string | `""` | Prefix (dir name) used to store backup files. |
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
| backup.elasticsearch.bucketName | string | `"sts-elasticsearch-backup"` | Name of the storage bucket where ElasticSearch snapshots are stored. |
| backup.elasticsearch.s3Prefix | string | `""` | Prefix (dir name) used to store backup files. |
| backup.elasticsearch.scheduled.schedule | string | `"0 0 3 * * ?"` | Cron schedule for automatic ElasticSearch snaphosts in [ElasticSearch cron schedule syntax](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/cron-expressions.html). |
| backup.elasticsearch.scheduled.snapshotRetentionExpireAfter | string | `"30d"` | Amount of time to keep ElasticSearch snapshots in [ElasticSearch time units](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/common-options.html#time-units). *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). |
| backup.elasticsearch.scheduled.snapshotRetentionMaxCount | string | `"30"` | Minimum number of ElasticSearch snapshots to keep. *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). |
| backup.elasticsearch.scheduled.snapshotRetentionMinCount | string | `"5"` | Minimum number of ElasticSearch snapshots to keep. *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). |
| backup.elasticsearch.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| backup.elasticsearch.securityContext.fsGroup | int | `65534` | The GID (group ID) of all files on all mounted volumes |
| backup.elasticsearch.securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| backup.elasticsearch.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| backup.elasticsearch.securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| backup.enabled | bool | `false` | Enables backup/restore for all data |
| backup.initJobAnnotations | object | `{}` | Annotations for Backup-init Job. |
| backup.poddisruptionbudget.maxUnavailable | int | `0` | Maximum number of pods that can be unavailable during the backup. |
| backup.stackGraph.bucketName | string | `"sts-stackgraph-backup"` | Name of the storage bucket to store StackGraph backups. |
| backup.stackGraph.restore.tempData.accessModes[0] | string | `"ReadWriteOnce"` |  |
| backup.stackGraph.restore.tempData.size | string | `nil` |  |
| backup.stackGraph.restore.tempData.storageClass | string | `nil` |  |
| backup.stackGraph.s3Prefix | string | `""` | Prefix (dir name) used to store backup files. |
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
| backup.storage.backend.azure | object | `{"accountKey":"","accountName":"","enabled":false,"endpoint":"","fromExternalSecret":""}` | Use Azure Blob Storage |
| backup.storage.backend.azure.accountKey | string | `""` | Azure storage account key (optional, falls back to managed identity) |
| backup.storage.backend.azure.accountName | string | `""` | Azure storage account name |
| backup.storage.backend.azure.enabled | bool | `false` | Enable Azure backend |
| backup.storage.backend.azure.endpoint | string | `""` | Azure blob endpoint (auto-derived from accountName if not set) |
| backup.storage.backend.azure.fromExternalSecret | string | `""` | Use an externally-managed secret for Azure backend credentials (keys: azureAccountName, azureAccountKey). When set, accountName/accountKey values are not required. |
| backup.storage.backend.pvc | object | `{"accessModes":["ReadWriteOnce"],"enabled":false,"size":"500Gi","storageClass":""}` | Use local PVC storage (default when no other backend configured) |
| backup.storage.backend.pvc.accessModes | list | `["ReadWriteOnce"]` | Access modes for the PVC |
| backup.storage.backend.pvc.enabled | bool | `false` | Enable PVC backend |
| backup.storage.backend.pvc.size | string | `"500Gi"` | Size of the PVC for S3Proxy data |
| backup.storage.backend.pvc.storageClass | string | `""` | Storage class for the PVC |
| backup.storage.backend.s3 | object | `{"accessKey":"","enabled":false,"endpoint":"","fromExternalSecret":"","region":null,"secretKey":""}` | Use external S3-compatible storage |
| backup.storage.backend.s3.accessKey | string | `""` | AWS access key (optional, falls back to instance profile/IRSA) |
| backup.storage.backend.s3.enabled | bool | `false` | Enable S3 backend |
| backup.storage.backend.s3.endpoint | string | `""` | S3 endpoint URL (optional, defaults to AWS) |
| backup.storage.backend.s3.fromExternalSecret | string | `""` | Use an externally-managed secret for S3 backend credentials (keys: backendAccessKey, backendSecretKey). When set, accessKey/secretKey values are not required. |
| backup.storage.backend.s3.region | string | `nil` | AWS region (defaults to us-east-1) |
| backup.storage.backend.s3.secretKey | string | `""` | AWS secret key (optional) |
| backup.storage.settingsPvc | object | `{"accessModes":["ReadWriteOnce"],"size":"2Gi","storageClass":""}` | PVC for local settings backup (always present) |
| backup.storage.settingsPvc.accessModes | list | `["ReadWriteOnce"]` | Access modes for the settings PVC |
| backup.storage.settingsPvc.size | string | `"2Gi"` | Size of the settings backup PVC |
| backup.storage.settingsPvc.storageClass | string | `""` | Storage class for the settings PVC |
| clickhouse.backup.image.registry | string | `"quay.io"` | Registry where to get the image from. |
| clickhouse.backup.image.repository | string | `"stackstate/clickhouse-backup"` | Repository where to get the image from. |
| clickhouse.backup.image.tag | string | `"2.6.43-47a710b2-173-release"` | Container image tag for 'clickhouse' backup containers. |
| clickhouse.backup.s3.endpoint | string | `"{{ include \"stackstate.s3proxy.endpoint\" . }}"` | S3-compatible endpoint for backup storage (resolved from s3proxy). |
| clickhouse.backup.s3.secretName | string | `"{{ include \"stackstate.s3proxy.secretName\" . }}"` | Name of the secret containing S3 credentials. |
| clickhouse.image.registry | string | `"quay.io"` | Registry where to get the image from |
| clickhouse.image.repository | string | `"stackstate/clickhouse"` | Repository where to get the image from. |
| clickhouse.image.tag | string | `"25.9.5-c65c568c-release-166"` | Container image tag for 'clickhouse' containers. |
| clickhouse.persistence.size | string | `nil` | Size of persistent volume for each clickhouse pod |
| clickhouse.replicaCount | string | `nil` | Number of ClickHouse replicas per shard to deploy. When using global.suseObservability.sizing.profile, this value is determined by the sizing profile (1 for most profiles, 3 for 4000-ha). |
| clickhouse.resources | object | `{}` |  |
| cluster-role.enabled | bool | `true` | Deploy the ClusterRole(s) and ClusterRoleBinding(s) together with the chart. Can be disabled if these need to be installed by an administrator of the Kubernetes cluster. |
| commonLabels | object | `{}` | Labels that will be added to all resources created by the stackstate chart (not the subcharts though) |
| deployment.compatibleWithArgoCD | bool | `false` | Whether to adjust the Chart to be compatible with ArgoCD. This feature is as of yet not deployed in the o11y-tenants and saas-tenants directories, so should be considered unfinished (see STAC-21445) |
| elasticsearch.esJavaOpts | string | `nil` | JVM options |
| elasticsearch.imageTag | string | `"8.19.4-c4527168-release-213"` | Elasticsearch image tag. Updated by updatecli. |
| elasticsearch.prometheus-elasticsearch-exporter.image.tag | string | `"v1.10.0-6159e1ba-release-174"` | Elasticsearch Prometheus exporter image tag. Updated by updatecli. |
| elasticsearch.prometheus-elasticsearch-exporter.resources | object | `{}` |  |
| elasticsearch.replicas | string | `nil` | Number of Elasticsearch replicas. |
| elasticsearch.resources | object | `{}` | Override Elasticsearch resources |
| elasticsearch.volumeClaimTemplate | object | `{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":null}}}` | PVC template defaulting to 250Gi default volumes |
| gateway | object | `{"annotations":{},"enabled":false,"filters":[],"hostnames":[],"parentRefs":[],"path":"","timeouts":{}}` | Gateway API configuration for exposing SUSE Observability via HTTPRoute. |
| gateway.annotations | object | `{}` | Annotations for HTTPRoute objects. |
| gateway.enabled | bool | `false` | Enable Gateway API HTTPRoute for SUSE Observability. |
| gateway.filters | list | `[]` | Optional filters for the HTTPRoute rule. |
| gateway.hostnames | list | `[]` | List of hostnames for the HTTPRoute. If not specified, derived from baseUrl. |
| gateway.parentRefs | list | `[]` | List of parent Gateway references (required when gateway.enabled is true). |
| gateway.path | string | `""` | Path prefix for the HTTPRoute rule. If not specified, derived from baseUrl or defaults to "/". |
| gateway.timeouts | object | `{}` | Optional timeouts for the HTTPRoute rule. |
| global.backup.enabled | bool | `false` |  |
| global.commonLabels | object | `{}` | Labels that will be added to all Deployments, StatefulSets, CronJobs, Jobs and their pods |
| global.features | object | `{"experimentalStackpacks":false}` | Feature switches for SUSE Observability. |
| global.features.experimentalStackpacks | bool | `false` | Enable StackPacks 2.0 to signal to all components that they should support the StackPacks 2.0 spec. This is a preproduction feature, usage may break your entire installation with upcoming releases. No backwards compatibility is guaranteed. |
| global.imagePullSecrets | list | `[]` | List of image pull secret names to be used by all images across all charts. |
| global.imageRegistry | string | `nil` | Image registry to be used by all images across all charts. When using global.suseObservability (global mode), set this to "registry.rancher.com" to match the default behavior of the suse-observability-values chart. |
| global.receiverApiKey | string | `""` | Deprecated. Use global.suseObservability.receiverApiKey instead. |
| global.s3proxy.credentials.accessKey | string | `"default-for-settings-only"` | Access key for S3Proxy authentication (override for production usage with global.backup.enabled) |
| global.s3proxy.credentials.fromExternalSecret | string | `""` | Use an externally-managed secret for credentials (keys: accesskey, secretkey). When set, the chart will not create a secret and accessKey/secretKey values are not required. |
| global.s3proxy.credentials.secretKey | string | `"default-secret-for-settings-only"` | Secret key for S3Proxy authentication (override for production usage with global.backup.enabled) |
| global.storageClass | string | `nil` | StorageClass for all PVCs created by the chart. Can be overridden per PVC. |
| global.suseObservability | object | `{"adminPassword":"","adminPasswordBcrypt":"","affinity":{"nodeAffinity":null,"podAffinity":null,"podAntiAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":true,"topologyKey":"kubernetes.io/hostname"}},"applicationDomains":["Observability"],"baseUrl":"","license":"","pullSecret":{"password":"","username":""},"receiverApiKey":"","sizing":{"profile":""}}` | Simplified configuration section that allows users to specify high-level settings. When any values in this section are configured (license, baseUrl, sizing.profile, etc.), the chart will automatically use this configuration instead of the legacy stackstate.* values. This provides a single-chart installation experience without needing the separate suse-observability-values chart. NOTE: This section works in conjunction with existing global settings (imageRegistry, receiverApiKey, imagePullSecrets). IMPORTANT: When using this section, also set global.imageRegistry to "registry.rancher.com" for SUSE Observability images. |
| global.suseObservability.adminPassword | string | `""` | Admin password for the default 'admin' user (plain text). Mutually exclusive with adminPasswordBcrypt. Required (one of the two) when using global.suseObservability configuration unless other authentication methods (LDAP, OIDC, Keycloak) are configured. |
| global.suseObservability.adminPasswordBcrypt | string | `""` | Admin password as bcrypt hash. Mutually exclusive with adminPassword. |
| global.suseObservability.affinity | object | `{"nodeAffinity":null,"podAffinity":null,"podAntiAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":true,"topologyKey":"kubernetes.io/hostname"}}` | Affinity configuration for all SUSE Observability components including infrastructure. |
| global.suseObservability.affinity.nodeAffinity | string | `nil` | Node affinity configuration applied to all components (application and infrastructure). Standard Kubernetes nodeAffinity spec. |
| global.suseObservability.affinity.podAffinity | string | `nil` | Pod affinity configuration for application components only. Does NOT apply to infrastructure components. Standard Kubernetes podAffinity spec. |
| global.suseObservability.affinity.podAntiAffinity | object | `{"requiredDuringSchedulingIgnoredDuringExecution":true,"topologyKey":"kubernetes.io/hostname"}` | Simplified pod anti-affinity configuration for HA profiles. Applied to all infrastructure components (kafka, clickhouse, zookeeper, elasticsearch, hbase, victoria-metrics) for HA profiles. |
| global.suseObservability.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution | bool | `true` | Enable required (hard) pod anti-affinity. When true, pods must be scheduled on different nodes. When false, soft anti-affinity is used (preferred but not required). |
| global.suseObservability.affinity.podAntiAffinity.topologyKey | string | `"kubernetes.io/hostname"` | Topology key for pod anti-affinity. Determines the domain for spreading pods (e.g., kubernetes.io/hostname for node-level, topology.kubernetes.io/zone for zone-level). |
| global.suseObservability.applicationDomains | list | `["Observability"]` | Enabled application domains. Default is ["Observability"]. Only "Observability" is currently supported. |
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
| global.wait.image.tag | string | `"1.0.12-9657337c-release-192"` | Container image tag for wait containers. |
| hbase.hbase.master.replicaCount | string | `nil` | Number of HBase master node replicas. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| hbase.hbase.regionserver.replicaCount | string | `nil` | Number of HBase regionserver node replicas. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| hbase.hdfs.datanode.replicaCount | string | `nil` | Number of HDFS datanode replicas. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| hbase.hdfs.version | string | `"java21-8-aef270ee-release-366"` | HDFS image build version (e.g. java21-8-27156f06-353). Derived from hadoop docker tag with semver prefix stripped. Updated by updatecli. |
| hbase.stackgraph.version | string | `"7.14.1"` |  |
| hbase.tephra.replicaCount | string | `nil` | Number of Tephra replicas. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| ingress.annotations | object | `{}` | Annotations for ingress objects. |
| ingress.enabled | bool | `false` | Enable use of ingress controllers. |
| ingress.hosts | list | `[]` | List of ingress hostnames; the paths are fixed to StackState backend services |
| ingress.ingressClassName | string | `""` |  |
| ingress.path | string | `"/"` |  |
| ingress.tls | list | `[]` | List of ingress TLS certificates to use. |
| kafka.image.registry | string | `"quay.io"` | Kafka image registry |
| kafka.image.repository | string | `"stackstate/kafka"` | Kafka image repository |
| kafka.image.tag | string | `"3.9.1-9.2-release"` | Kafka image tag. **Since StackState relies on this specific version, it's advised NOT to change this.** When changing this version, be sure to change the pod annotation stackstate.com/kafkaup-operator.kafka_version aswell, in order for the kafkaup operator to upgrade the inter broker protocol version |
| kafka.metrics.jmx.image.registry | string | `"quay.io"` | Kafka JMX exporter image registry |
| kafka.metrics.jmx.image.repository | string | `"stackstate/jmx-exporter"` | Kafka JMX exporter image repository |
| kafka.metrics.jmx.image.tag | string | `"0.20.0-58a72255-307-release"` | Kafka JMX exporter image tag |
| kafka.metrics.jmx.resources.limits.cpu | string | `"1"` |  |
| kafka.metrics.jmx.resources.limits.ephemeral-storage | string | `"1Gi"` |  |
| kafka.metrics.jmx.resources.limits.memory | string | `"300Mi"` |  |
| kafka.metrics.jmx.resources.requests.cpu | string | `"200m"` |  |
| kafka.metrics.jmx.resources.requests.ephemeral-storage | string | `"1Mi"` |  |
| kafka.metrics.jmx.resources.requests.memory | string | `"300Mi"` |  |
| kafka.persistence.size | string | `nil` | Size of persistent volume for each Kafka pod |
| kafka.replicaCount | string | `nil` | Number of Kafka replicas. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| kafka.resources | object | `{}` | Kafka resources per pods. |
| kafkaup-operator.enabled | bool | `true` |  |
| kafkaup-operator.image.pullPolicy | string | `""` |  |
| kafkaup-operator.image.registry | string | `"quay.io"` |  |
| kafkaup-operator.image.repository | string | `"stackstate/kafkaup-operator"` |  |
| kafkaup-operator.image.tag | string | `"0.0.7"` |  |
| kafkaup-operator.kafkaSelectors.podLabel.key | string | `"app.kubernetes.io/component"` |  |
| kafkaup-operator.kafkaSelectors.podLabel.value | string | `"kafka"` |  |
| kafkaup-operator.kafkaSelectors.statefulSetName | string | `"suse-observability-kafka"` |  |
| kafkaup-operator.startVersion | string | `"2.3.1"` |  |
| kubernetes-rbac-agent.clusterName.value | string | `"{{ .Release.Name }}"` |  |
| kubernetes-rbac-agent.containers.rbacAgent.affinity | object | `{}` | Set affinity |
| kubernetes-rbac-agent.containers.rbacAgent.env | object | `{}` | Additional environment variables |
| kubernetes-rbac-agent.containers.rbacAgent.image.repository | string | `"stackstate/kubernetes-rbac-agent"` |  |
| kubernetes-rbac-agent.containers.rbacAgent.image.tag | string | `"c96a7ba1-699-release"` |  |
| kubernetes-rbac-agent.containers.rbacAgent.nodeSelector | object | `{}` | Set a nodeSelector |
| kubernetes-rbac-agent.containers.rbacAgent.podAnnotations | object | `{"ad.stackstate.com/kubernetes-rbac-agent.check_names":"[\"openmetrics\"]","ad.stackstate.com/kubernetes-rbac-agent.init_configs":"[{}]","ad.stackstate.com/kubernetes-rbac-agent.instances":"[ { \"prometheus_url\": \"http://%%host%%:8080/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"*\"] } ]"}` | Additional annotations on the pod |
| kubernetes-rbac-agent.containers.rbacAgent.podLabels | object | `{}` | Additional labels on the pod |
| kubernetes-rbac-agent.containers.rbacAgent.priorityClassName | string | `""` | Set priorityClassName |
| kubernetes-rbac-agent.containers.rbacAgent.resources.limits.memory | string | `"40Mi"` | Memory resource limits. |
| kubernetes-rbac-agent.containers.rbacAgent.resources.requests.memory | string | `"25Mi"` | Memory resource requests. |
| kubernetes-rbac-agent.containers.rbacAgent.tolerations | list | `[]` | Set tolerations |
| kubernetes-rbac-agent.url.value | string | `"{{ include \"stackstate.rbacAgent.url\" . }}"` |  |
| minio | object | `{"accessKey":"","azuregateway":{"enabled":false},"fullnameOverride":"","persistence":{"enabled":false},"s3gateway":{"accessKey":"","enabled":false,"secretKey":"","serviceEndpoint":""},"secretKey":"","serviceAccount":{"annotations":{},"create":true,"name":""}}` | DEPRECATED: MinIO subchart has been replaced by S3Proxy. These values are kept for backward compatibility only. Please migrate to backup.storage.* values. Legacy minio.* values will be removed in a future release. |
| minio.accessKey | string | `""` | DEPRECATED: Use global.s3proxy.credentials.accessKey instead. If set (not empty and not "setme"), will be used as fallback for S3Proxy credentials. |
| minio.azuregateway | object | `{"enabled":false}` | DEPRECATED: Use backup.storage.backend.azure instead. When minio.azuregateway.enabled is true, S3Proxy will be configured to use Azure Blob as the backend. |
| minio.azuregateway.enabled | bool | `false` | DEPRECATED: Use backup.storage.backend.azure.enabled instead. |
| minio.fullnameOverride | string | `""` | DEPRECATED: Used for backward compatibility with existing deployments. |
| minio.persistence | object | `{"enabled":false}` | DEPRECATED: Use backup.storage.pvc |
| minio.persistence.enabled | bool | `false` | DEPRECATED: Use backup.storage.pvc.enabled instead. |
| minio.s3gateway | object | `{"accessKey":"","enabled":false,"secretKey":"","serviceEndpoint":""}` | DEPRECATED: Use backup.storage.backend.s3 instead. When minio.s3gateway.enabled is true, S3Proxy will be configured to use S3 as the backend. |
| minio.s3gateway.accessKey | string | `""` | DEPRECATED: Use backup.storage.backend.s3.accessKey instead. |
| minio.s3gateway.enabled | bool | `false` | DEPRECATED: Use backup.storage.backend.s3.enabled instead. |
| minio.s3gateway.secretKey | string | `""` | DEPRECATED: Use backup.storage.backend.s3.secretKey instead. |
| minio.s3gateway.serviceEndpoint | string | `""` | DEPRECATED: Use backup.storage.backend.s3.endpoint instead. |
| minio.secretKey | string | `""` | DEPRECATED: Use global.s3proxy.credentials.secretKey instead. If set (not empty and not "setme"), will be used as fallback for S3Proxy credentials. |
| minio.serviceAccount | object | `{"annotations":{},"create":true,"name":""}` | DEPRECATED: Use s3proxy.serviceAccount instead. |
| minio.serviceAccount.annotations | object | `{}` | DEPRECATED: Use s3proxy.serviceAccount.annotations instead. If set, will be merged into the S3Proxy service account annotations (s3proxy values take precedence). |
| minio.serviceAccount.create | bool | `true` | DEPRECATED: Use s3proxy.serviceAccount.create instead. |
| minio.serviceAccount.name | string | `""` | DEPRECATED: Use s3proxy.serviceAccount.name instead. If set, will be used as fallback for the S3Proxy service account name. |
| networkPolicy.enabled | bool | `false` | Enable creating of `NetworkPolicy` object and associated rules for StackState. |
| networkPolicy.spec | object | `{"ingress":[{"from":[{"podSelector":{}}]}],"podSelector":{"matchLabels":{}},"policyTypes":["Ingress"]}` | `NetworkPolicy` rules for StackState. |
| opentelemetry-collector.image.registry | string | `"quay.io"` | Base container image registry. |
| opentelemetry-collector.image.repository | string | `"stackstate/sts-opentelemetry-collector"` | Repository where to get the image from. |
| opentelemetry-collector.image.tag | string | `"v0.0.31"` | Container image tag for 'opentelemetry-collector' containers. |
| opentelemetry-collector.resources.limits.cpu | string | `"500m"` |  |
| opentelemetry-collector.resources.limits.memory | string | `"512Mi"` |  |
| opentelemetry-collector.resources.requests.cpu | string | `"250m"` |  |
| opentelemetry-collector.resources.requests.memory | string | `"512Mi"` |  |
| pull-secret.credentials | list | `[]` | Registry and associated credentials (username, password) that will be stored in the pull-secret |
| pull-secret.enabled | bool | `false` | Deploy the ImagePullSecret for the chart. |
| pull-secret.fullNameOverride | string | `""` | Name of the ImagePullSecret that will be created. This can be referenced by setting the `global.imagePullSecrets[0].name` value in the chart. |
| s3proxy.affinity | object | `{}` | Affinity settings for S3Proxy pod (merged with stackstate.components.all.affinity) |
| s3proxy.extraEnv.open | object | `{}` | Extra open environment variables to inject into the S3Proxy pod. |
| s3proxy.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into the S3Proxy pod via a `Secret` object. |
| s3proxy.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for S3Proxy |
| s3proxy.image.registry | string | `"quay.io"` | imageRegistry used for the S3Proxy Docker image |
| s3proxy.image.repository | string | `"stackstate/s3proxy"` | Image repository for S3Proxy |
| s3proxy.image.tag | string | `"3.1.0-c7312e71-release-22"` | Image tag for S3Proxy |
| s3proxy.logLevel | string | `"info"` |  |
| s3proxy.metrics.agentAnnotationsEnabled | bool | `true` | Put annotations on each pod to instruct the stackstate agent to scrape the metrics |
| s3proxy.metrics.defaultAgentMetricsFilter | string | `"[\"*\"]"` |  |
| s3proxy.metrics.enabled | bool | `true` | Enable / disable S3Proxy Prometheus metrics. |
| s3proxy.metrics.servicemonitor.additionalLabels | object | `{}` | Additional labels for targeting Prometheus operator instances. |
| s3proxy.metrics.servicemonitor.enabled | bool | `false` | Enable `ServiceMonitor` object; `all.metrics.enabled` *must* be enabled. |
| s3proxy.nodeSelector | object | `{}` | Node selector for S3Proxy pod (merged with stackstate.components.all.nodeSelector) |
| s3proxy.podAnnotations | object | `{}` | Annotations for S3Proxy pod |
| s3proxy.resources | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"700Mi"},"requests":{"cpu":"200m","ephemeral-storage":"1Mi","memory":"700Mi"}}` | Resource limits and requests for S3Proxy container |
| s3proxy.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| s3proxy.securityContext.fsGroup | int | `65534` | The GID (group ID) used to mount volumes |
| s3proxy.securityContext.runAsGroup | int | `65534` | The GID (group ID) of the owning user of the process |
| s3proxy.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| s3proxy.securityContext.runAsUser | int | `65534` | The UID (user ID) of the owning user of the process |
| s3proxy.serviceAccount.annotations | object | `{}` | Annotations for the S3Proxy service account (e.g. for IAM roles). |
| s3proxy.serviceAccount.create | bool | `true` | Whether to create the service account for S3Proxy. |
| s3proxy.serviceAccount.name | string | `""` | Override the service account name. Defaults to "suse-observability-minio" for backward compatibility with the old Minio subchart (e.g. IAM role bindings). |
| s3proxy.sizing.baseMemoryConsumption | string | `"250Mi"` | Memory reserved for OS and non-heap JVM usage (metaspace, threads, etc) |
| s3proxy.sizing.javaHeapMemoryFraction | string | `"70"` | Percentage of remaining memory (after baseMemoryConsumption) allocated to Java heap |
| s3proxy.tolerations | list | `[]` | Tolerations for S3Proxy pod (appended to stackstate.components.all.tolerations) |
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
| stackstate.components.aiAssistant.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.aiAssistant.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: secret-key - name: MY_OTHER_SECRET_ENV_VAR   secretName: my-other-k8s-secret   secretKey: another-secret-key |
| stackstate.components.aiAssistant.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.aiAssistant.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.aiAssistant.image.imageRegistry | string | `""` | `imageRegistry` used for the `ai-assistant` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.aiAssistant.image.pullPolicy | string | `""` | `pullPolicy` used for the `ai-assistant` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.aiAssistant.image.repository | string | `"stackstate/suse-observability-borg"` | Repository of the ai-assistant component Docker image. |
| stackstate.components.aiAssistant.image.tag | string | `"20260318165448-8cc501ea"` | Tag used for the `ai-assistant` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.aiAssistant.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.aiAssistant.persistence | object | `{"size":"1Gi","storageClass":""}` | Persistence settings for the SQLite database. |
| stackstate.components.aiAssistant.persistence.size | string | `"1Gi"` | Size of the PVC for the SQLite database. |
| stackstate.components.aiAssistant.persistence.storageClass | string | `""` | Storage class for the SQLite database PVC. |
| stackstate.components.aiAssistant.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.aiAssistant.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `ai-assistant` pods. |
| stackstate.components.aiAssistant.resources | object | `{}` | Resource allocation for `ai-assistant` pods. |
| stackstate.components.aiAssistant.serviceAccount.annotations | object | `{}` | Annotations for the ai-assistant `ServiceAccount`. Useful for e.g. AWS IAM Role binding via IRSA: `eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/my-ai-assistant-role`. |
| stackstate.components.aiAssistant.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.all.affinity | object | `{}` | Affinity settings for pod assignment on all components. |
| stackstate.components.all.clickHouse.hostnames | string | `"suse-observability-clickhouse-headless"` |  |
| stackstate.components.all.deploymentStrategy.type | string | `"RecreateSingletonsOnly"` | Deployment strategy for StackState components. Possible values: `RollingUpdate`, `Recreate` and `RecreateSingletonsOnly`. `RecreateSingletonsOnly` uses `Recreate` for the singleton Deployments and `RollingUpdate` for the other Deployments. |
| stackstate.components.all.elasticsearchEndpoint | string | `""` | **Required if `elasticsearch.enabled` is `false`** Endpoint for shared Elasticsearch cluster. |
| stackstate.components.all.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key |
| stackstate.components.all.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods for all components. |
| stackstate.components.all.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object for all components. |
| stackstate.components.all.image.pullPolicy | string | `"IfNotPresent"` | The default pullPolicy used for all stateless components of StackState; individual service `pullPolicy`s can be overridden (see below). |
| stackstate.components.all.image.registry | string | `"quay.io"` | Base container image registry for all StackState containers, except for the wait container and the container-tools container |
| stackstate.components.all.image.repositorySuffix | string | `""` |  |
| stackstate.components.all.image.tag | string | `"7.0.0-snapshot.20260429062734-master-4631d9b"` | The default tag used for all stateless components of StackState; individual service `tag`s can be overridden (see below). |
| stackstate.components.all.kafkaEndpoint | string | `""` | **Required if `elasticsearch.enabled` is `false`** Endpoint for shared Kafka broker. |
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
| stackstate.components.api.resources | object | `{}` | Resource allocation for `api` pods. |
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
| stackstate.components.authorizationSync.resources | object | `{}` | Resource allocation for `notification` pods. |
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
| stackstate.components.checks.resources | object | `{}` | Resource allocation for `state` pods. |
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
| stackstate.components.clickhouseCleanup.image.tag | string | `"25.9.5-c65c568c-release-166"` | Container image tag for 'clickhouseCleanup' containers. |
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
| stackstate.components.containerTools.image.tag | string | `"1.8.5-653"` | Container image tag for container-tools containers. |
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
| stackstate.components.correlate.resources | object | `{}` | Resource allocation for `correlate` pods. |
| stackstate.components.correlate.sizing.baseMemoryConsumption | string | `"400Mi"` |  |
| stackstate.components.correlate.sizing.javaHeapMemoryFraction | string | `"65"` |  |
| stackstate.components.correlate.split.aggregator.affinity | object | `{}` | Additional affinity settings for pod assignment. |
| stackstate.components.correlate.split.aggregator.extraEnv.open | object | `{"CONFIG_FORCE_stackstate_correlate_aggregation_workers":"3","CONFIG_FORCE_stackstate_correlate_correlateConnections_workers":"0","CONFIG_FORCE_stackstate_correlate_correlateHTTPTraces_workers":"0"}` | Extra open environment variables to inject into pods. Will merge with stackstate.components.correlate.extraEnv.open |
| stackstate.components.correlate.split.aggregator.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.correlate.extraEnv.secret |
| stackstate.components.correlate.split.aggregator.nodeSelector | object | `{}` | Additional node labels for pod assignment. |
| stackstate.components.correlate.split.aggregator.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.correlate.split.aggregator.replicaCount | int | `1` | Number of `aggregator correlate` replicas. |
| stackstate.components.correlate.split.aggregator.resources | object | `{}` | Resource allocation for pods. If not defined, will take from sizing profile defaults. |
| stackstate.components.correlate.split.aggregator.sizing.javaHeapMemoryFraction | string | `nil` |  |
| stackstate.components.correlate.split.aggregator.sizing.logsMemoryConsumption | string | `nil` |  |
| stackstate.components.correlate.split.aggregator.tolerations | list | `[]` | Additional toleration labels for pod assignment. |
| stackstate.components.correlate.split.connection.affinity | object | `{}` | Additional affinity settings for pod assignment. |
| stackstate.components.correlate.split.connection.extraEnv.open | object | `{"CONFIG_FORCE_stackstate_correlate_aggregation_workers":"0","CONFIG_FORCE_stackstate_correlate_correlateConnections_workers":"3","CONFIG_FORCE_stackstate_correlate_correlateHTTPTraces_workers":"0"}` | Extra open environment variables to inject into pods. Will merge with stackstate.components.correlate.extraEnv.open |
| stackstate.components.correlate.split.connection.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.correlate.extraEnv.secret |
| stackstate.components.correlate.split.connection.nodeSelector | object | `{}` | Additional node labels for pod assignment. |
| stackstate.components.correlate.split.connection.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.correlate.split.connection.replicaCount | int | `1` | Number of `connection correlate` replicas. |
| stackstate.components.correlate.split.connection.resources | object | `{}` | Resource allocation for pods. If not defined, will take from sizing profile defaults. |
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
| stackstate.components.correlate.split.httpTracing.resources | object | `{}` | Resource allocation for pods. If not defined, will take from sizing profile defaults. |
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
| stackstate.components.e2es.resources | object | `{}` | Resource allocation for `e2es` pods. |
| stackstate.components.e2es.retention | int | `30` | Number of days to keep the events data on Es |
| stackstate.components.e2es.sizing.baseMemoryConsumption | string | `"300Mi"` |  |
| stackstate.components.e2es.sizing.javaHeapMemoryFraction | string | `"50"` |  |
| stackstate.components.e2es.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.healthSync.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.healthSync.affinity | object | `{}` | Affinity settings for pod assignment. |
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
| stackstate.components.healthSync.resources | object | `{}` | Resource allocation for `healthSync` pods. |
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
| stackstate.components.initializer.resources | object | `{}` | Resource allocation for `initializer` pods. |
| stackstate.components.initializer.sizing.baseMemoryConsumption | string | `"350Mi"` |  |
| stackstate.components.initializer.sizing.javaHeapMemoryFraction | string | `"65"` |  |
| stackstate.components.initializer.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.kafkaTopicCreate.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.kafkaTopicCreate.extraEnv.open | object | `{}` | Add additional environment variables to the pod |
| stackstate.components.kafkaTopicCreate.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.registry | string | `"quay.io"` | Base container image registry for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.repository | string | `"stackstate/kafka"` | Base container image repository for kafka-topic-create containers. |
| stackstate.components.kafkaTopicCreate.image.tag | string | `"3.9.1-9.2-release"` | Container image tag for kafka-topic-create containers. |
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
| stackstate.components.mcp.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.mcp.envsFromExistingSecrets | list | `[]` | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: secret-key - name: MY_OTHER_SECRET_ENV_VAR   secretName: my-other-k8s-secret   secretKey: another-secret-key |
| stackstate.components.mcp.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.mcp.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. |
| stackstate.components.mcp.image.imageRegistry | string | `""` | `imageRegistry` used for the `mcp` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.mcp.image.pullPolicy | string | `""` | `pullPolicy` used for the `mcp` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.mcp.image.repository | string | `"stackstate/suse-observability-mcp"` | Repository of the mcp component Docker image. |
| stackstate.components.mcp.image.tag | string | `"20260326141755-fe7a6469"` | Tag used for the `mcp` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.mcp.nodeSelector | object | `{}` | Node labels for pod assignment. |
| stackstate.components.mcp.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.mcp.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `mcp` pods. |
| stackstate.components.mcp.replicaCount | int | `1` | Number of `mcp` replicas. |
| stackstate.components.mcp.resources | object | `{}` | Resource allocation for `mcp` pods. |
| stackstate.components.mcp.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.nginxPrometheusExporter.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.registry | string | `"quay.io"` | Base container image registry for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.repository | string | `"stackstate/nginx-prometheus-exporter"` | Base container image repository for nginx-prometheus-exporter containers. |
| stackstate.components.nginxPrometheusExporter.image.tag | string | `"1.5.1-fdbee6c2-111-release"` | Container image tag for nginx-prometheus-exporter containers. |
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
| stackstate.components.notification.resources | object | `{}` | Resource allocation for `notification` pods. |
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
| stackstate.components.receiver.resources | object | `{}` | Resource allocation for `receiver` pods. |
| stackstate.components.receiver.retention | int | `7` | Number of days to keep the logs data on Es |
| stackstate.components.receiver.sizing.baseMemoryConsumption | string | `"300Mi"` |  |
| stackstate.components.receiver.sizing.javaHeapMemoryFraction | string | `"65"` |  |
| stackstate.components.receiver.split.base.affinity | object | `{}` | Additional affinity settings for pod assignment. |
| stackstate.components.receiver.split.base.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. Will merge with stackstate.components.receiver.extraEnv.open |
| stackstate.components.receiver.split.base.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.receiver.extraEnv.secret |
| stackstate.components.receiver.split.base.nodeSelector | object | `{}` | Additional node labels for pod assignment. |
| stackstate.components.receiver.split.base.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.receiver.split.base.replicaCount | int | `1` | Number of `base receiver` replicas. |
| stackstate.components.receiver.split.base.resources | object | `{}` | Resource allocation for pods. If not defined, will take from sizing profile defaults. |
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
| stackstate.components.receiver.split.logs.resources | object | `{}` | Resource allocation for pods. If not defined, will take from sizing profile defaults. |
| stackstate.components.receiver.split.logs.sizing.javaHeapMemoryFraction | string | `nil` |  |
| stackstate.components.receiver.split.logs.sizing.logsMemoryConsumption | string | `nil` |  |
| stackstate.components.receiver.split.logs.tolerations | list | `[]` | Additional toleration labels for pod assignment. |
| stackstate.components.receiver.split.processAgent.affinity | object | `{}` | Additional affinity settings for pod assignment. |
| stackstate.components.receiver.split.processAgent.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. Will merge with stackstate.components.receiver.extraEnv.open |
| stackstate.components.receiver.split.processAgent.extraEnv.secret | object | `{}` | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.receiver.extraEnv.secret |
| stackstate.components.receiver.split.processAgent.nodeSelector | object | `{}` | Additional node labels for pod assignment. |
| stackstate.components.receiver.split.processAgent.podAnnotations | object | `{}` | Extra annotations |
| stackstate.components.receiver.split.processAgent.replicaCount | int | `1` | Number of `processAgent receiver` replicas. |
| stackstate.components.receiver.split.processAgent.resources | object | `{}` | Resource allocation for pods. If not defined, will take from sizing profile defaults. |
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
| stackstate.components.router.image.tag | string | `"1.37.1-1.13-release"` | Tag used for the `router` component Docker image; this will override `stackstate.components.all.image.tag` on a per-service basis. |
| stackstate.components.router.mode.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.router.mode.extraEnv.open | object | `{}` | Add additional environment variables to the pod |
| stackstate.components.router.mode.image.pullPolicy | string | `nil` | Image pull policy for router mode containers. |
| stackstate.components.router.mode.image.registry | string | `"quay.io"` | Base container image registry for router mode containers. |
| stackstate.components.router.mode.image.repository | string | `"stackstate/container-tools"` | Base container image repository for router mode containers. |
| stackstate.components.router.mode.image.tag | string | `"1.8.5-653"` | Container image tag for router mode containers. |
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
| stackstate.components.router.resources | object | `{}` | Resource allocation for `router` pods. |
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
| stackstate.components.server.resources | object | `{}` | Resource allocation for `server` pods. |
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
| stackstate.components.slicing.resources | object | `{}` | Resource allocation for `slicing` pods. |
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
| stackstate.components.state.resources | object | `{}` | Resource allocation for `state` pods. |
| stackstate.components.state.sizing.baseMemoryConsumption | string | `"300Mi"` |  |
| stackstate.components.state.sizing.javaHeapMemoryFraction | string | `"65"` |  |
| stackstate.components.state.tmpToPVC | object | `{"storageClass":null,"volumeSize":"2Gi"}` | Whether to use PersistentVolume to store temporary files (/tmp) instead of pod ephemeral storage, empty - use pod ephemeral storage. |
| stackstate.components.state.tmpToPVC.storageClass | string | `nil` | Storage class name of PersistentVolume used by /tmp directory. It stores temporary files/caches, so it should be the fastest possible. |
| stackstate.components.state.tmpToPVC.volumeSize | string | `"2Gi"` | The size of the PersistentVolume for "/tmp" directory. |
| stackstate.components.state.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.sync.additionalLogging | string | `""` | Additional logback config |
| stackstate.components.sync.affinity | object | `{}` | Affinity settings for pod assignment. |
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
| stackstate.components.sync.resources | object | `{}` | Resource allocation for `sync` pods. |
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
| stackstate.components.ui.resources | object | `{}` | Resource allocation for `ui` pods. |
| stackstate.components.ui.securityContext.enabled | bool | `true` | Whether or not to enable the securityContext |
| stackstate.components.ui.securityContext.fsGroup | int | `101` | The GID (group ID) used to mount volumes |
| stackstate.components.ui.securityContext.runAsGroup | int | `101` | The GID (group ID) of the owning user of the process |
| stackstate.components.ui.securityContext.runAsNonRoot | bool | `true` | Ensure that the user is not root (!= 0) |
| stackstate.components.ui.securityContext.runAsUser | int | `101` | The UID (user ID) of the owning user of the process |
| stackstate.components.ui.tolerations | list | `[]` | Toleration labels for pod assignment. |
| stackstate.components.vmagent.affinity | object | `{"podAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app.kubernetes.io/component","operator":"In","values":["receiver"]},{"key":"app.kubernetes.io/instance","operator":"In","values":["stackstate"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":80}]}}` | Affinity settings for vmagent pod. |
| stackstate.components.vmagent.extraArgs | object | `{}` |  |
| stackstate.components.vmagent.image.repository | string | `"stackstate/vmagent"` |  |
| stackstate.components.vmagent.image.tag | string | `"v1.136.0-068f508b-release-150"` |  |
| stackstate.components.vmagent.persistence.size | string | `"10Gi"` |  |
| stackstate.components.vmagent.persistence.storageClass | string | `nil` |  |
| stackstate.components.vmagent.poddisruptionbudget | object | `{"maxUnavailable":1}` | PodDisruptionBudget settings for `vmagent` pods. |
| stackstate.components.vmagent.resources | object | `{}` | Resource allocation for vmagent pod. |
| stackstate.components.workloadObserver.affinity | object | `{}` | Affinity settings for pod assignment. |
| stackstate.components.workloadObserver.enabled | bool | `true` | Enable/disable the workload observer |
| stackstate.components.workloadObserver.extraEnv.open | object | `{}` | Extra open environment variables to inject into pods. |
| stackstate.components.workloadObserver.image.imageRegistry | string | `""` | `imageRegistry` used for the `workloadObserver` component Docker image; this will override `global.imageRegistry` on a per-service basis. |
| stackstate.components.workloadObserver.image.pullPolicy | string | `""` | `pullPolicy` used for the `workloadObserver` component Docker image; this will override `stackstate.components.all.image.pullPolicy` on a per-service basis. |
| stackstate.components.workloadObserver.image.repository | string | `"stackstate/workload-observer"` | Repository of the workloadObserver component Docker image. |
| stackstate.components.workloadObserver.image.tag | string | `"864f55e0-164-release"` | Tag used for the `workloadObserver` component Docker image.. |
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
| stackstate.stackpacks.image.version | string | `"20260421152824-master-f02a31b"` | Version used for the `stackpacks` Docker image, the tag is build from the version and the stackstate edition + deployment mode |
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
| victoria-metrics-0.backup.awsSecrets | string | `"{{ include \"stackstate.s3proxy.secretName\" . }}"` | Name of the secret containing S3 credentials (resolved from s3proxy). |
| victoria-metrics-0.backup.bucketName | string | `"sts-victoria-metrics-backup"` | Name of the storage bucket where Victoria Metrics backups are stored. |
| victoria-metrics-0.backup.keepLastDaily | int | `7` | Number of daily backups to retain |
| victoria-metrics-0.backup.keepLastWeekly | int | `4` | Number of weekly backups to retain |
| victoria-metrics-0.backup.overrideS3Endpoint | string | `"http://{{ include \"stackstate.s3proxy.endpoint\" . }}"` | S3-compatible endpoint for backup storage (resolved from s3proxy). **Do not change this value!** |
| victoria-metrics-0.backup.s3Prefix | string | `"victoria-metrics-0"` |  |
| victoria-metrics-0.backup.scheduled.daily | string | `"55 0 * * *"` | Cron schedule for daily snapshot backups of Victoria Metrics |
| victoria-metrics-0.backup.scheduled.hourly | string | `"25 * * * *"` | Cron schedule for hourly incremental backups of Victoria Metrics |
| victoria-metrics-0.backup.setupCron.image.tag | string | `"1.8.5-653"` | Container-tools image for cron setup. Updated by updatecli. |
| victoria-metrics-0.backup.vmbackup.image.tag | string | `"v1.136.0-47ff639a-release-69"` | VM backup image tag. Updated by updatecli. |
| victoria-metrics-0.enabled | bool | `true` |  |
| victoria-metrics-0.server.fullnameOverride | string | `"suse-observability-victoria-metrics-0"` | Full name override |
| victoria-metrics-0.server.image.tag | string | `"v1.136.0-621be04a-release-156"` | Victoria Metrics server image tag. Updated by updatecli. |
| victoria-metrics-0.server.persistentVolume.size | string | `nil` | Size of storage for Victoria Metrics, ideally 20% of free space remains available at all times |
| victoria-metrics-0.server.resources | object | `{}` |  |
| victoria-metrics-1.backup.awsSecrets | string | `"{{ include \"stackstate.s3proxy.secretName\" . }}"` | Name of the secret containing S3 credentials (resolved from s3proxy). |
| victoria-metrics-1.backup.bucketName | string | `"sts-victoria-metrics-backup"` | Name of the storage bucket where Victoria Metrics backups are stored. |
| victoria-metrics-1.backup.keepLastDaily | int | `7` | Number of daily backups to retain |
| victoria-metrics-1.backup.keepLastWeekly | int | `4` | Number of weekly backups to retain |
| victoria-metrics-1.backup.overrideS3Endpoint | string | `"http://{{ include \"stackstate.s3proxy.endpoint\" . }}"` | S3-compatible endpoint for backup storage (resolved from s3proxy). **Do not change this value!** |
| victoria-metrics-1.backup.s3Prefix | string | `"victoria-metrics-1"` | Prefix (dir name) used to store backup files, we may have multiple instances of Victoria Metrics, each of them should be stored into their own directory. |
| victoria-metrics-1.backup.scheduled.daily | string | `"5 1 * * *"` | Cron schedule for daily snapshot backups of Victoria Metrics |
| victoria-metrics-1.backup.scheduled.hourly | string | `"35 * * * *"` | Cron schedule for hourly incremental backups of Victoria Metrics |
| victoria-metrics-1.backup.setupCron.image.tag | string | `"1.8.5-653"` | Container-tools image for cron setup. Updated by updatecli. |
| victoria-metrics-1.backup.vmbackup.image.tag | string | `"v1.136.0-47ff639a-release-69"` | VM backup image tag. Updated by updatecli. |
| victoria-metrics-1.enabled | bool | `true` |  |
| victoria-metrics-1.server.fullnameOverride | string | `"suse-observability-victoria-metrics-1"` | Full name override |
| victoria-metrics-1.server.image.tag | string | `"v1.136.0-621be04a-release-156"` | Victoria Metrics server image tag. Updated by updatecli. |
| victoria-metrics-1.server.persistentVolume.size | string | `nil` | Size of storage for Victoria Metrics, ideally 20% of free space remains available at all times |
| victoria-metrics-1.server.resources | object | `{}` |  |
| victoria-metrics.restore.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for `vmrestore` containers. |
| victoria-metrics.restore.image.registry | string | `"quay.io"` | Base container image registry for 'vmrestore' containers. |
| victoria-metrics.restore.image.repository | string | `"stackstate/vmrestore"` | Base container image repository for 'vmrestore' containers. |
| victoria-metrics.restore.image.tag | string | `"v1.136.0-2ca23936-release-52"` | Container image tag for 'vmrestore' containers. |
| victoria-metrics.restore.securityContext.enabled | bool | `true` |  |
| victoria-metrics.restore.securityContext.fsGroup | int | `65534` |  |
| victoria-metrics.restore.securityContext.runAsGroup | int | `65534` |  |
| victoria-metrics.restore.securityContext.runAsNonRoot | bool | `true` |  |
| victoria-metrics.restore.securityContext.runAsUser | int | `65534` |  |
| zookeeper.image.registry | string | `"quay.io"` | ZooKeeper image registry |
| zookeeper.image.repository | string | `"stackstate/zookeeper"` | ZooKeeper image repository |
| zookeeper.image.tag | string | `"3.9.5-3.2-release"` | ZooKeeper image tag |
| zookeeper.persistence.size | string | `nil` | Size of the PVC for Zookeeper data. Default is 8Gi, will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| zookeeper.replicaCount | string | `nil` | Default amount of Zookeeper replicas to provision. Will be overridden by sizing profile if using global.suseObservability.sizing.profile. |
| zookeeper.resources | object | `{}` |  |

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

## Backup Storage Configuration

SUSE Observability uses S3Proxy to provide an S3-compatible API for all backup operations. This enables consistent backup workflows regardless of whether you're using local storage or cloud providers like AWS S3 or Azure Blob Storage.

### Architecture Overview

S3Proxy manages two types of storage:

1. **Settings backup** (always local) - A small PVC (~2Gi) that stores configuration backups locally, ensuring they're always available for restore operations
2. **Main backup storage** (configurable) - Stores StackGraph, Elasticsearch, ClickHouse, and VictoriaMetrics backups

```
┌─────────────────────────────────────────────────────────┐
│                      S3Proxy                            │
│                                                         │
│   settings-local-backup    →  Local PVC (always)       │
│   sts-stackgraph-backup    →  PVC / S3 / Azure         │
│   sts-elasticsearch-backup →  PVC / S3 / Azure         │
│   sts-clickhouse-backup    →  PVC / S3 / Azure         │
│   sts-victoria-*           →  PVC / S3 / Azure         │
└─────────────────────────────────────────────────────────┘
```

### Default Configuration (Local PVC)

By default, only configuration backups are enabled and stored on local PVCs.

### Enable all backups using PVC

To store all backups on PVCs:

```yaml
global:
  backup:
    enabled: true

backup:
  storage:
    backend:
      pvc:
        enabled: true
        size: 500Gi  # Size for main backup storage, size depends on the expected combined backups size
s3proxy:
  credentials:
    # Credentials used internally in the cluster to access the backup storage (S3Proxy). Override with your own values for production use.
    accessKey: "my-access-key"
    secretKey: "my-secret-key"
```

### Enable all backups using AWS S3

To store backups in AWS S3:

```yaml
global:
  backup:
    enabled: true

backup:
  storage:
    backend:
      s3:
        enabled: true
        # Optional: Specify AWS region (required for AWS S3, can also be set via environment variable if needed, optional for S3-compatible storage)
        region: "eu-west-1"
        # Option 1: Use IAM role / IRSA (recommended)
        # Leave accessKey and secretKey empty

        # Option 2: Use explicit credentials
        accessKey: "aws-access-key"
        secretKey: "aws-secret-key"

        # Option 3: Use credentials from external secret:
        fromExternalSecret: "my-aws-credentials" # Kubernetes secret containing backendAccessKey and backendSecretKey
        # Optional: Custom S3-compatible endpoint (e.g., MinIO)
        # endpoint: "https://minio.example.com"

s3proxy:
  credentials:
    # Credentials used internally in the cluster to access the backup storage (S3Proxy). Override with your own values for production use.
    accessKey: "my-access-key"
    secretKey: "my-secret-key"
```

### Enable all backups using Azure Blob Storage

To store backups in Azure Blob Storage:

```yaml
global:
  backup:
    enabled: true

backup:
  storage:
    backend:
      azure:
        enabled: true
        accountName: "mystorageaccount"
        # Option 1: Use managed identity (recommended)
        # Leave accountKey empty

        # Option 2: Use explicit credentials
        accountKey: "your-storage-account-key"

        # Option 3: Use credentials from external secret:
        fromExternalSecret: "my-azure-credentials" # Kubernetes secret containing azureAccountName and azureAccountKey

s3proxy:
  credentials:
    # Credentials used internally in the cluster to access the backup storage (S3Proxy). Override with your own values for production use.
    accessKey: "my-access-key"
    secretKey: "my-secret-key"
```

### S3Proxy Credentials

S3Proxy requires credentials for clients (backup jobs) to authenticate. There are defaults pre-defined but for production usage it is recommended to set your own credentials.

```yaml
s3proxy:
  credentials:
    # Option 1: Set credentials directly
    accessKey: "my-access-key"
    secretKey: "my-secret-key"

    # Or use an existing secret
    fromExternalSecret: "my-s3proxy-credentials"
```

### Migration from MinIO

If you're upgrading from a previous version that used MinIO, a new PVC and bucket are created for the local settings backup. The existing settings-backup-pvc will not be removed automatically. Backups from that PVC can still be restored with the sts-backup cli by providing --from-pvc.

After a while (these backups only have 10 days retention), the old backup PVC can be removed. It is recommended to first check that new settings backups have been made (using the sts-backup-cli `settings list` command) and then manually delete the old PVC.

**Legacy MinIO values are deprecated and will be removed in a future release. Please migrate to the new backup storage configuration as soon as possible.**

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
