# suse-observability

Helm chart for SUSE Observability

Current chart version is `2.13.2-pre.41`

**Homepage:** <https://gitlab.com/stackvista/stackstate.git>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../../local/anomaly-detection/ | anomaly-detection | * |
| file://../../local/clickhouse/ | clickhouse | * |
| file://../../local/common/ | common | * |
| file://../../local/elasticsearch/ | elasticsearch | * |
| file://../../local/hbase/ | hbase | * |
| file://../../local/kafka/ | kafka | * |
| file://../../local/kafkaup-operator/ | kafkaup-operator | * |
| file://../../local/kubernetes-rbac-agent/ | kubernetes-rbac-agent | * |
| file://../../local/opentelemetry-collector/ | opentelemetry-collector | * |
| file://../../local/pull-secret/ | pull-secret | * |
| file://../../local/suse-observability-sizing/ | suse-observability-sizing | * |
| file://../../local/victoria-metrics-single/ | victoria-metrics-0(victoria-metrics-single) | * |
| file://../../local/victoria-metrics-single/ | victoria-metrics-1(victoria-metrics-single) | * |
| file://../../local/zookeeper/ | zookeeper | * |

## Node draining

Application PodDisruptionBudgets default to
`unhealthyPodEvictionPolicy: AlwaysAllow`. This allows node drains to evict
running, unready pods while ready pods remain subject to the existing budget.
Set `stackstate.components.<component>.poddisruptionbudget.unhealthyPodEvictionPolicy`
to `IfHealthyBudget` to retain recovery protection, or `null` to omit the field.

Database and storage PDBs use Kubernetes' default `IfHealthyBudget` policy to
keep eviction conservative during recovery. This applies to Elasticsearch,
Kafka, ZooKeeper, VictoriaMetrics (including the combined budget), HBase, HDFS
and Tephra. The AI assistant and anomaly detection use the same policy because
the assistant stores conversations in SQLite and the anomaly detection manager
owns persistent model state. ClickHouse has no PDB template.

## Pod metadata

Pod annotations for HBase workloads can be set through
`hbase.all.podAnnotations` or each HBase component's `podAnnotations` value.
The corresponding `podLabels` values configure their pod labels.
Vmagent accepts `stackstate.components.vmagent.podAnnotations` and inherits
`stackstate.components.all.podAnnotations`. The KafkaUp operator accepts
`kafkaup-operator.podAnnotations` and `kafkaup-operator.podLabels`.
Anomaly detection accepts `anomaly-detection.podAnnotations` for both its
manager and worker pods.

These settings are optional. HBase and KafkaUp preserve their pod selector
labels when custom labels conflict. Their pod metadata and anomaly detection
pod annotations are literal values, not Helm templates. Numeric and boolean
annotation values are converted to strings, including vmagent annotations.
Generated configuration checksums on KafkaUp, anomaly detection and vmagent
pods are preserved so configuration changes still restart the pods.

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

## Bundled RBAC agent configuration

The platform configures its bundled Kubernetes RBAC agent through the
`kubernetes-rbac-agent.connection` template helper. By default, the agent connects to
this installation's router and uses the Helm release name as its cluster name.
The subchart's `kubernetes-rbac-agent.url` and `kubernetes-rbac-agent.clusterName`
settings are ignored in platform installations, including inline values and
external ConfigMap references retained from previous releases. Upgrades replace
any previously customized nested connection with the platform's connection.
Explicit `global.url.fromSecret` and `global.clusterName.fromSecret` settings retain
precedence: each selects its value from the specified Secret and suppresses the
corresponding internal ConfigMap.
The standalone subchart and the SUSE Observability agent chart retain their own
connection configuration.

## Bundled backup and anomaly-detection connections

The platform supplies VictoriaMetrics and ClickHouse backup endpoints and
credential Secret names through `victoria-metrics.backup.connection` and
`clickhouse.backup.connection`. Configure backup storage through
`backup.storage.backend` and S3Proxy credentials through `global.s3proxy.credentials`,
including `fromExternalSecret` for an existing Secret.

The nested `victoria-metrics-{0,1}.backup.awsSecrets`,
`victoria-metrics-{0,1}.backup.overrideS3Endpoint`, and
`clickhouse.backup.s3.endpoint` / `secretName` settings are ignored in platform
installations, including saved values from earlier releases. Backup buckets,
prefixes, schedules and retention remain configurable.

Bundled anomaly detection always uses this installation's router through
`anomaly-detection.stackstate.instance`. A saved `anomaly-detection.stackstate.instance`
value is ignored; authentication settings are unchanged. Standalone subcharts
retain their configurable connections.

Backup checksums now include the resolved connection rather than template text.
The first upgrade changes the checksum and triggers a rollout of VictoriaMetrics
instances with backups enabled and ClickHouse when global backups are enabled.
Allow for interruptions while these pods restart. This refactor preserves
StatefulSet/PVC identities and default backup destinations; previously customized
nested connections switch to the platform connection.

The OpenTelemetry instrumentation namespace defaults to
`<chart-name>-<release-namespace>` in a helper. Explicit shared and per-component
`otelInstrumentation.serviceNamespace` settings, including templates, still work.

## Resource naming

Generated collector relay ConfigMaps use these platform names:

| Mode | ConfigMap |
| --- | --- |
| Deployment | `suse-observability-otel-collector-config` |
| DaemonSet | `suse-observability-otel-collector-agent` |
| StatefulSet | `suse-observability-otel-collector-statefulset` |

These names ignore release and fullname settings. Explicit
`opentelemetry-collector.configMap.existingName` selections and disabled generation
remain unchanged. Standalone collector charts retain their original defaults.

The Deployment relay map has its own `-config` suffix. Older charts using
Deployment mode and the default collector fullname emitted it under the same
name as the platform endpoint map, `suse-observability-otel-collector`. The new
names separate relay configuration from `api.url` and `intake.url`. Endpoint map
identity, data and all `API_URL`/`INTAKE_URL` references remain unchanged.

Default platform installations use StatefulSet mode and keep their ConfigMap
identity and checksum. For eligible custom names, existing controllers update
only the relay volume reference and its configuration checksum, rolling Pods
through the platform's default `RollingUpdate` strategy. Relay contents, commands,
probes, mounts, labels, accounts, RBAC, Services, autoscalers, StatefulSet names, governing
Services, claim templates, PVCs and checkpoint locations are preserved.
This batch works independently after extraction and requires no data migration,
account replacement or later naming change. Rollouts can interrupt telemetry
collection and forwarding; local checks establish no recovery-time bound.

Repeated upgrades retain the intended names. Rollback restores the previous maps
and references for previously working configurations. Rolling back to the old
default-fullname Deployment configuration recreates its pre-existing duplicate
ConfigMap names; that configuration is not a healthy rollback baseline.

Argo CD must prune obsolete generated relay maps. `PruneLast=true` with pruning
enabled can defer deletion until updated workloads are healthy. After an
interrupted upgrade, rollback or sync without pruning, finish or restore the
intended revision and inspect its owned ConfigMap inventory and consumer
references. Keep maps still referenced by Pods or controllers until consumers
converge, and preserve explicit external selections.
Remove only obsolete relay maps owned by this installation; preserve the endpoint
map, storage and credentials. Verify collector readiness and telemetry forwarding
afterwards. Actual Helm/Argo CD lifecycle and runtime recovery remain integration
checks.

Two internal Role/RoleBinding pairs use canonical names from `_names.tpl`:

| Component | Role | RoleBinding |
| --- | --- | --- |
| Router automatic-mode hooks | `suse-observability-router-mode` | `suse-observability-router-mode` |
| Bundled Kafka-upgrade operator | `suse-observability-kafkaup` | `suse-observability-kafkaup-binding` |

These names ignore release names and fullname settings. The router pair is
created only in automatic mode; the operator pair is created only when the
bundled operator is enabled. Each pair remains in the release namespace, so
installations in different namespaces can use the same names.
Standalone Kafka-upgrade operator charts retain their existing naming defaults.

Roles and bindings change together because a RoleBinding's `roleRef` is immutable.
Helm replaces a renamed binding rather than changing its Role reference in place.
Later fullname changes keep the canonical Role reference and update only mutable
account subjects where necessary. Existing accounts, permissions, hook phases,
weights and deletion policies, workload specifications, Kafka selection, router
mode configuration and script targets are preserved. The router pair already
has these names with default product naming; the operator pair changes even for
the default release. The RBAC-agent Role/binding and its application authorization
subject retain their existing identities.

This batch can be installed or upgraded independently of the other naming
changes after helper extraction. It introduces no Pod rollout, account/token
replacement or data migration. Authorization can be interrupted while old/new
objects are replaced and caches reconcile; local render and patch checks do not
establish recovery time or actual controller behavior.

Argo CD must prune obsolete operator RBAC. Successful router hooks retain their
existing automatic cleanup; a failed hook from an earlier revision can leave
RBAC under its previous name because `before-hook-creation` matches the new
name only. After an interrupted upgrade, rollback or sync without pruning,
finish or restore the intended revision, inspect the rendered Role/binding
inventory and remove only obsolete objects owned by that installation.
Check release ownership, hook annotations, namespaces, subjects and Role
references before cleanup. Preserve accounts and customer-managed bindings.
Update external automation that selects the old names; any custom binding to
one of these internal Roles must reference its new name. Verify router hook
completion and Kafka-operator access after the operation.

The bundled collector's peripheral resources use fixed names:

| Kind | Name |
| --- | --- |
| PodDisruptionBudget | `suse-observability-otel-collector` |
| ServiceMonitor | `suse-observability-otel-collector` |
| PodMonitor | `suse-observability-otel-collector-agent` |
| PrometheusRule | `suse-observability-otel-collector` |
| NetworkPolicy | `suse-observability-otel-collector` |

These names ignore release names, parent naming settings and collector
`fullnameOverride`/`nameOverride` settings. Default collector names are unchanged.
Only metadata names change: selectors, release labels, namespaces, disruption
limits, scrape configuration, alert rules and network rules stay unchanged.
Creation conditions are preserved, including the Deployment-only budget and
DaemonSet-only PodMonitor. Each installation needs its own resource namespace,
including when `opentelemetry-collector.namespaceOverride` is configured.
Standalone collector charts retain their existing naming defaults.

This collector peripheral change can be deployed independently of other resource renames.
Workload names and Pod templates, Services and router endpoints, ConfigMaps,
Secrets, accounts, RBAC, autoscalers and StatefulSet claim templates remain
unchanged. No later naming step, data migration or collector rollout is required
to make these peripheral objects select the existing collector.

For custom collector names, an ordinary Helm upgrade creates replacements and
deletes the previous objects. Rollback restores the previous names. During
replacement, overlapping PDBs can block eviction with HTTP 500, and overlapping
monitors or rules can duplicate scraping or alert evaluation. Identical old/new
NetworkPolicies retain the same additive permissions while they overlap.
Argo CD must prune obsolete objects; `PruneLast=true` with pruning enabled can
defer deletion until the intended resources are healthy.

After an interrupted non-atomic upgrade, rollback or Argo CD sync without
pruning, stale peripheral objects can remain indefinitely. Finish or restore the
intended revision, compare its five-kind inventory in the collector namespace,
verify release ownership and remove only obsolete objects belonging to that
installation. Do not delete by shared labels. Retain collector workloads,
configuration, credentials, accounts, Services and storage throughout cleanup.
Verify eviction, Prometheus targets/rules and network enforcement afterwards.
Local render checks do not establish controller reconciliation or recovery time.

The UI and replication checker now use the fixed names `suse-observability-ui` and
`suse-observability-replication-checker`, including their component-specific supporting
resources.
The vmagent ConfigMap and S3Proxy ServiceMonitor also use the fixed `suse-observability` prefix.

Main application and logging ConfigMaps use `suse-observability-<component>` and
`suse-observability-<component>-log`. This applies to `api`, `checks`,
`notification`, `health-sync`, `authorization-sync`, `state`, `sync`, `slicing`,
`server`, `receiver`, `correlate`, `initializer` and `e2es`, including the shared
configuration used by split receiver and correlate workers. Release names and
root/global fullname settings no longer affect these ConfigMap names. Creation
conditions, configuration contents, labels and annotations are unchanged.
Additional generated ConfigMaps use these fixed names:

| Configuration | Name |
| --- | --- |
| Router bootstrap | `suse-observability-router` |
| Stackpack copy scripts | `suse-observability-stackpacks-scripts` |
| Bundled RBAC-agent URL | `suse-observability-rbac-agent-url` |
| Bundled RBAC-agent cluster name | `suse-observability-rbac-agent-cluster-name` |
| Bundled Kafka-upgrade operator | `suse-observability-kafkaup-config` |
| HBase backup environment | `suse-observability-hbase-backup` |

These names ignore the release name and fullname settings. Their contents and
creation conditions are unchanged. The RBAC agent still identifies its cluster
by the Helm release name. External global connection Secrets retain precedence.
The bundled resource names are supplied by the platform; standalone subcharts
and the SUSE Observability agent chart retain their existing naming defaults.
Other resources with fixed default names are unchanged.

Router active/maintenance/automatic and mode-script ConfigMaps retain their names:
pre-upgrade hooks need the existing configuration, and automatic mode stores
runtime state. Backup log, restore-script and backup-CLI ConfigMaps also retain
their names because existing Jobs and embedded manual-job manifests can reference
them. Renaming those maps could strand a pending Job. `kube-root-ca.crt` is
Kubernetes-managed and is outside chart resource naming.

Existing Deployments update their ConfigMap volume/environment references and
existing configuration checksums through an ordinary Helm upgrade.
The HBase backup environment rename updates its StatefulSet ConfigMap reference
in both Mono and Distributed modes. Its credential Secret retains
its old name and contents. The Kafka-upgrade operator retains its Deployment,
replica count, update strategy and Kafka selection; no second controller is created.
These changes roll affected Pods and can interrupt application processing and
availability while replacements start. A completed rollout references the
replacement ConfigMaps; repeated upgrades retain those names. Rollback restores
the previous revision's ConfigMaps and references and rolls Pods again.
Default product naming retains application/logging, router-bootstrap, stackpack
and RBAC-agent ConfigMap identities. The Kafka-operator and HBase-backup ConfigMap
names also change for the default release, causing their consumers to roll.
Environment and credential Secrets, workload identities and strategies,
StatefulSet identities and claim templates, PVCs, mount paths, Service DNS and
external resource choices are preserved. No storage or credential migration is
required.

Argo CD must prune obsolete ConfigMaps. `PruneLast=true`, together with enabled
pruning, can defer deletion until the updated workloads are healthy. After an
interrupted upgrade or rollback, finish or restore the intended revision first.
Compare its ConfigMap inventory with live objects and verify release ownership.
Remove only obsolete chart-owned ConfigMaps once Pods and controller templates
no longer reference them; old Pods can still need the previous configuration
during rollout. Do not delete by shared labels, or remove router state,
externally supplied configuration or credential objects. Verify application
readiness after the rollout; its duration depends on the installation.

Main-chart PodDisruptionBudgets always use `suse-observability-<component>`,
independently of the release name or root/global fullname overrides, prefixes
and suffixes. These naming settings no longer affect main-chart PDB names.
Selectors, disruption limits, labels and creation conditions are unchanged;
other subchart-owned PDBs retain their names, except for anomaly detection and HBase
as described below.
When a PDB name changes, Helm creates its replacement before deleting the old
object. While both budgets select the same Pod, Kubernetes rejects eviction
requests with HTTP 500, so node drains fail until the obsolete budget is removed.
After cleanup, eviction requests follow the remaining budget's disruption limits.
Update external automation that selects the previous PDB names.

Main-chart ServiceMonitors also always use `suse-observability-<component>`,
including the existing receiver and correlate worker suffixes. Root/global
fullname overrides, prefixes and suffixes no longer affect their names.
UI and S3Proxy monitor names are already canonical. Service names, selectors,
monitor labels, namespace selection and scrape endpoints stay unchanged.
During replacement, Prometheus may discover both monitor identities or interrupt
scraping while its operator reconciles the configuration. The original monitor
count and target configuration are restored once obsolete monitors are removed
and the operator reconciles. Stale monitors selected by Prometheus can continue
causing duplicate scraping until cleanup.
Update external automation that references the previous ServiceMonitor names.
Other subchart-owned ServiceMonitors retain their names, except for the collector,
anomaly detection, HBase and the Elasticsearch exporter as described here.

Anomaly detection's PDB and manager ServiceMonitor now use the fixed names
`suse-observability-anomaly-detection` and `suse-observability-spotlight-manager`.
Release names and root, subchart or global fullname overrides, prefixes and
suffixes do not change these two names. Their existing labels, selectors,
disruption limits, namespace selection, scrape settings and creation conditions
are preserved. The manager and worker Deployments, Service DNS, artifact PVC,
configuration, credentials and RBAC keep their existing identities.
Both renamed objects remain in the release namespace, so releases with the
same name in different namespaces have separate PDBs and ServiceMonitors.
The same eviction, duplicate-scraping and pruning considerations apply to these
anomaly-detection objects. Update automation referencing their previous names.

HBase's PDBs also use canonical names:
`suse-observability-hbase-master`, `suse-observability-hbase-rs`,
`suse-observability-hdfs-nn`, `suse-observability-hdfs-snn`,
`suse-observability-hdfs-dn`, and `suse-observability-tephra`.
Root, HBase and global fullname settings no longer affect these budget names.
Their labels, selectors, annotations, `maxUnavailable: 1`, namespace scope and
Mono/Distributed creation conditions are unchanged. HBase StatefulSet identities,
claim templates, PVCs, governing Services, credentials and RBAC are preserved.
The same eviction and
stale-budget recovery guidance applies; update automation using old PDB names.

HBase's six ServiceMonitors use the same canonical component names when kept
in the release namespace, preserving selectors, namespace selection, scrape
settings, labels and creation conditions. Monitors placed outside the release
namespace by `hbase.servicemonitor.namespace` or
`hbase.common.servicemonitor.namespace` retain their previous names to avoid
collisions between installations in a shared monitoring namespace.

The two HBase probe-script ConfigMaps use
`suse-observability-hbase-rs-scripts` and `suse-observability-hdfs-dn-scripts`.
Their contents, file modes, probe commands and mount paths stay unchanged.
Their existing StatefulSets update only the corresponding script volume
reference, rolling region-server and DataNode Pods in Distributed mode,
including with default product naming. This can temporarily interrupt storage
availability. StatefulSet/PVC identities, claim templates, data mounts,
governing Services, selectors, replicas, images, credentials and accounts are
preserved. No data migration or follow-up rename is required.
This batch also works independently on the extraction baseline. Rollback restores
the previous script ConfigMaps and references; repeated upgrades keep the new
names. Local render and patch checks do not establish runtime recovery time.

Prune obsolete HBase monitors and script maps in Argo CD; `PruneLast=true` with
pruning enabled can wait for healthy updated workloads. After an interrupted
operation, compare the intended revision's owned inventory and preserve any old
script map still referenced by Pods or controllers until consumers converge.
Remove only obsolete objects owned by this installation, retaining storage and
credentials, and verify probe readiness and scrape targets after recovery.

These overlaps are temporary only when obsolete resources are deleted. No data
or storage migration is required, but failed deployments or disabled pruning can
leave stale PDBs and ServiceMonitors requiring operational cleanup.

For Argo CD, enable automatic pruning (`spec.syncPolicy.automated.prune: true`)
or perform a manual sync with pruning (`argocd app sync <application> --prune`).
Syncing without pruning leaves the old objects behind. `PruneLast=true` only
controls when pruning happens; it does not enable pruning.

Helm creates or updates target resources before deleting obsolete resources.
A non-atomic upgrade can create replacements and fail before reaching deletion.
For Helm 3, consider `--atomic` for rollback on upgrade failure or
`--cleanup-on-fail` to request removal of newly created resources on failure.
Recovery and deletion can also fail, so verify the resulting inventory even when
using these options.

To recover stale resources after an interrupted upgrade, rollback or Argo CD sync:

1. Choose the intended chart revision and values, whether finishing the upgrade
   or restoring the previous revision. Render it with the installation's release
   name, namespace and values. Compare its PDB and ServiceMonitor inventory with
   live resources in their actual namespaces; ServiceMonitors may use a separate
   monitoring namespace.
2. Confirm the intended PDBs have the expected selectors and disruption limits,
   and the intended ServiceMonitors have the expected selectors and endpoints.
   Identify obsolete objects using Helm release ownership or Argo CD tracking
   metadata, rather than names alone.
3. Delete only obsolete PDBs and ServiceMonitors belonging to this installation
   and absent from the intended manifests. Old and new objects share labels and
   selectors, so do not delete them in bulk by a shared label selector.
4. Complete or retry the chosen upgrade, rollback or pruning sync if needed.
   Verify that overlapping PDBs no longer block evictions and Prometheus selects
   only the intended monitors.

The optional main-chart NetworkPolicy now uses the fixed name
`suse-observability-networkpolicy`, defined in `_names.tpl`. Release names and
root/global/common fullname settings no longer affect this name. Its API version,
labels, annotations, Pod selector, ingress/egress rules, policy types and enabled
setting are unchanged. It belongs to the release namespace, so releases in
different namespaces have separate policies. Subchart policies keep their names.
This rename does not change Pod specifications, trigger Pod rollouts or migrate
storage or credentials; update automation referencing the old policy name.

Standard Kubernetes NetworkPolicy permissions are additive. Old and new copies
with identical rules allow the same traffic, but an obsolete policy left behind
can preserve permissions when later updates tighten the intended rules. Enable
Argo CD pruning and verify cleanup after failed Helm upgrades or rollbacks.
Compare the intended revision's policy inventory and rules with actual objects
and release ownership; remove only obsolete policies from this installation.
Do not delete policies in bulk using shared labels or selectors. Network-plugin
reconciliation may temporarily affect connectivity during replacement; verify
traffic enforcement after a completed upgrade or rollback.

The internal Pod-reader Role and RoleBinding retain the matching legacy name
`<common.fullname.short>-get-pods`. The authorization ClusterRole and
ClusterRoleBinding likewise retain `<common.fullname.cluster.unique>-authorization`,
including existing release, namespace and fullname settings. Customer-managed
bindings can continue referencing these Roles. When fullname overrides or affixes
change, their chart-owned bindings change names alongside the referenced Roles.
Helm replaces these bindings on upgrade or rollback instead of patching the
immutable `roleRef` field. Keeping canonical binding names while their Role names
change would cause that patch to fail. These two bindings are retained exceptions
to canonical naming; no replacement machinery or migration flags are introduced.

The authentication ClusterRoleBinding uses the canonical name
`suse-observability-authentication` in namespace `suse-observability`, and
`suse-observability-<namespace>-authentication` elsewhere. It ignores release names
and fullname overrides/affixes and includes the full namespace without truncation.
Its Role reference is always Kubernetes' `system:auth-delegator`, so subsequent
fullname changes can update its subjects without changing immutable `roleRef`.

Permissions, ServiceAccount subject selection, labels, annotations and creation
conditions remain unchanged. No account, workload, storage or credential identity
is renamed by these binding helpers. Changing fullname settings can still rename
other legacy resources; this binding fix does not make every such change safe.

Prune obsolete chart-owned bindings after replacement, including in Argo CD.
For interrupted upgrades or rollbacks, compare the intended revision's binding
inventory with actual objects, verify release ownership and subjects, and remove
only obsolete bindings belonging to this installation. Preserve Roles,
ServiceAccounts, customer-managed bindings and other installations' resources.
Verify Pod-reader access, token review and authorization after the operation;
authorization caches may take time to reconcile. Update external automation that
selects the previous authentication binding name.

The Elasticsearch exporter Deployment uses the fixed name
`suse-observability-prometheus-elasticsearch-exporter`, including when exporter
fullname or name overrides are configured. Its existing Service DNS name,
selectors, Pod configuration, accounts and certificate Secret identities are
preserved. Replacement may briefly interrupt exporter metrics or run both
exporters; update automation that references the old Deployment name.

The exporter's optional ServiceMonitor, PodMonitor and PrometheusRule also use
`suse-observability-prometheus-elasticsearch-exporter` when placed in the release
namespace, independently of release names and exporter name/fullname overrides.
Objects placed elsewhere retain their legacy exporter fullname, including release
and name/fullname overrides, so previously distinct installations can share a
monitoring namespace. Each object's own `serviceMonitor.namespace`,
`podMonitor.namespace` or `prometheusRule.namespace` setting controls this choice.
An unset namespace or one equal to the release namespace uses the canonical name.
Their labels, namespace settings,
creation conditions, selectors, scrape configuration, alert expressions and
rule-group names are unchanged. Service DNS, Pod specifications, RBAC, accounts
and certificate Secrets remain unchanged. These names belong to different
resource kinds; the existing choice between ServiceMonitor and PodMonitor is
preserved.
Installations in different namespaces can use the same canonical names.
Explicit monitoring namespaces remain supported without introducing collisions
between installations whose legacy exporter names are distinct.

Prune obsolete exporter monitors and rules in Argo CD and verify their inventory
after failed Helm upgrades or rollbacks. Stale monitors can duplicate scraping;
stale rules can continue evaluating previous alerts. Compare the intended
revision's objects in their configured namespaces, verify ownership, and remove
only obsolete objects belonging to this installation. Never delete them using
shared labels. Verify targets and rules after Prometheus reconciles.

The router Deployment uses the fixed name `suse-observability-router`, ignoring
fullname overrides, prefixes and suffixes. Its Service DNS, selectors, Pod
configuration contents, accounts, credentials, dynamic-mode ConfigMaps, and hook
identities stay unchanged. Its bootstrap ConfigMap uses the canonical name listed
above. Automatic-mode scripts select router Deployments by
their component and Helm release labels within the release namespace. The
pre-upgrade hook can therefore restart the old router into maintenance mode,
and the post-upgrade hook restores active mode on the replacement. If both
Deployments are present, the scripts restart and wait for both.
The hook Role grants Deployment watch permission before the jobs run. Both
scripts use watched rollout status with a 120-second timeout; this is below the
maintenance job's 180-second deadline. A timeout or watch error stops the hook
instead of reporting a completed rollout.
Replacement can interrupt connections or briefly run both proxies. Update
automation that references the old Deployment name.

For renamed exporter and router Deployments, Argo CD must prune obsolete
Deployments to finish replacement. After an interrupted upgrade or rollback,
compare the intended Deployment inventory and release ownership before deleting
obsolete controllers individually; old and new controllers share Pod labels.
Their retained Services, configuration, accounts, Secrets and storage resources
must not be removed as part of this cleanup.

The main Ingress retains its existing naming settings to preserve ownership of
controller-managed Certificates and TLS Secrets. The main HTTPRoute is already canonical.

When upgrading a release with a different name, Helm replaces the renamed resources.
This also applies to affected resources when using root `fullnameOverride`,
`fullnamePrefix`, or `fullnameSuffix`, or `global.fullnamePrefix` / `global.fullnameSuffix`,
even with the default release name. Update external manifests and automation that reference
the previous UI or replication-checker resource names.
Allow for a UI interruption and restart any active replication check after the upgrade.
When the vmagent ConfigMap name changes, its existing StatefulSet rolls the pod to reference
the new name. With the default single vmagent replica, including HA sizing profiles, this
temporarily interrupts metrics ingestion. The persistent queue retains the same PVC.
StatefulSet names, PVC names, HBase Services, and storage bindings are unchanged.
Other components retain their existing naming until a later migration.

## Replication checks without a local CLI

The stateless `replication-checker` Deployment lets you run `sts-backup replication check` through `kubectl`. It is enabled by default for HA sizing profiles and disabled for non-HA profiles and installations without a profile. Generated HA sizing files also enable it; regenerate older files or enable it explicitly. An explicit `true` or `false` overrides the built-in profile default:

```yaml
stackstate:
  components:
    replicationChecker:
      enabled: false
```

The checker uses `stackstate.components.containerTools.image`, which must contain `sts-backup` v0.10.0 or later. Configure its resources under `stackstate.components.replicationChecker.resources`. It runs checks only when invoked and uses no persistent storage.

```shell
kubectl -n observability get deployments -l app.kubernetes.io/component=replication-checker
kubectl -n observability exec deployment/<replication-checker-deployment> -c replication-checker -- \
  sts-backup replication check --namespace observability --wait
```

Replace the namespace and Deployment name with those of your installation. Add `--output json` for automation and stop maintenance if the command returns nonzero. If the Pod reschedules or the exec session is interrupted, rerun the complete check successfully before disrupting another node. A successful report does not replace Longhorn health, capacity, backup, or node-removal checks.

The checker has a dedicated service account with namespace-scoped permissions to list Pods and StatefulSets and create `pods/exec` requests. Although the checker runs read-only queries, `pods/exec` permits arbitrary commands in namespace Pods. Disabling the checker removes its Deployment, service account, Role, and RoleBinding. Workload-observer retains its original permissions.

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

The global affinity applies with and without a sizing profile.

### Scheduling Precedence (Affinity and Tolerations)

Affinity and tolerations resolve through layers, from most to least specific. The most specific layer that sets a value wins:

| Workload | Layers (most specific first) |
|----------|------------------------------|
| Application components | split type (`stackstate.components.<component>.split.<type>`), `stackstate.components.<component>`, `stackstate.components.all`, `global.suseObservability` |
| Backup and restore jobs | `stackstate.components.backup` (`stackstate.components.configurationBackup` for configuration backups), `stackstate.components.all`, `global.suseObservability` |
| HBase | `hbase.<component>`, `hbase.all`, `global.suseObservability` |
| Other infrastructure subcharts | the subchart's own value (for example `kafka.affinity`, `clickhouse.backup.tolerations`, `victoria-metrics-0.server.affinity`), `global.suseObservability` |

**Tolerations** are resolved as a whole list. The first layer that sets `tolerations` replaces every less specific layer; lists are not concatenated. Set `tolerations: []` to clear inherited tolerations for a component.

```yaml
global:
  suseObservability:
    tolerations:
      - key: dedicated
        operator: Equal
        value: observability
        effect: NoSchedule
stackstate:
  components:
    api:
      tolerations:
        - key: dedicated
          operator: Equal
          value: observability
          effect: NoSchedule
        - key: api-only
          operator: Exists
    ui:
      tolerations: []
```

**Affinity** is resolved per field: `nodeAffinity`, `podAffinity` and `podAntiAffinity` are picked independently from the most specific layer that sets them. A component that only sets `podAffinity` keeps the inherited `nodeAffinity` and the sizing profile's HA anti-affinity. A field is replaced as a whole, never merged with less specific layers. Set a field to `{}` to clear it:

```yaml
kafka:
  affinity:
    podAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 10
          podAffinityTerm:
            topologyKey: kubernetes.io/hostname
            labelSelector:
              matchLabels:
                app.kubernetes.io/name: zookeeper
clickhouse:
  affinity:
    podAntiAffinity: {}
```

Elasticsearch keeps its own values: `elasticsearch.nodeAffinity` overrides the global node affinity, and `elasticsearch.antiAffinity` (`hard`/`soft`) overrides the sizing profile's anti-affinity mode. Leaving them unset inherits the global and sizing values.

#### Breaking changes

- Tolerations are no longer concatenated. Previously `stackstate.components.all.tolerations` was appended to each component's tolerations, and `hbase.all.tolerations` to each HBase component's. A component that sets its own tolerations must now repeat any shared ones it still needs.
- An explicit `tolerations: []` now clears. It used to be the default everywhere, so a values file copied from `helm show values` with a leftover `stackstate.components.api.tolerations: []` gave api the `stackstate.components.all` tolerations; now it gets none. Likewise a leftover `elasticsearch.nodeAffinity: {}` (the old default) now drops the global node affinity on Elasticsearch. On tainted or dedicated node pools these pods stay Pending. Remove such leftovers.
- A local `affinity` on Kafka, ClickHouse, ZooKeeper, VictoriaMetrics, HBase and the other infrastructure subcharts no longer drops the global node affinity or the sizing profile's anti-affinity. Only the fields it sets are replaced. On HA profiles a `nodeAffinity`-only override now gets the sizing profile's required pod anti-affinity back, so pods can stay Pending on clusters with fewer nodes than replicas. Set `podAntiAffinity: {}` to keep the old behaviour.
- Elasticsearch follows the global anti-affinity mode. It used to be `hard` regardless; with an HA profile and `global.suseObservability.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution: false` it is now `soft`. Set `elasticsearch.antiAffinity: hard` to keep the old behaviour.
- `global.suseObservability.affinity` now also applies without a sizing profile.
- The ClickHouse `checksum/backup-config` annotation no longer includes backup scheduling values, so ClickHouse restarts once on upgrade.

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

Values are documented by support level.

**Public values**, listed below, are safe to set. They keep working across upgrades. When one is renamed, changes meaning or is removed, the release notes say so first and the old form keeps working for a deprecation period.

**Internal values** are everything in `values.yaml` that is not listed here. They cover chart wiring, per-image settings (the global registry and pull secrets are public), most replica counts, the JVM memory split (`sizing` blocks), deprecated values and debug settings, and may change or disappear in any release without notice. Overriding one is not supported. To size an installation, use `global.suseObservability.sizing.profile` and the public `resources` values.

### Public

| Key | Description | Type | Default |
|-----|-------------|------|---------|
| ai.assistant.anthropic.apiKey | Anthropic API key used when `ai.assistant.provider=anthropic`. | string | `""` |
| ai.assistant.anthropic.fromExternalSecret.key | Key in the external secret that contains the Anthropic API key. | string | `"ANTHROPIC_API_KEY"` |
| ai.assistant.anthropic.fromExternalSecret.name | Existing secret name that contains the Anthropic API key. When set, the chart will not create an internal secret for `ANTHROPIC_API_KEY`. | string | `""` |
| ai.assistant.bedrock.awsRegion | AWS region used when `ai.assistant.provider=bedrock`. | string | `"eu-west-1"` |
| ai.assistant.enabled | Enables the AI Assistant UI, the dedicated backend process, and the MCP (Model Context Protocol) Server for Observability. Enabling the AI Assistant always enables the MCP Server. | bool | `false` |
| ai.assistant.provider | LLM provider for AI Assistant. Possible values: `bedrock` or `anthropic`. | string | `"bedrock"` |
| ai.mcp.enabled | Enables the MCP Server for use with external code assistants, without enabling the AI Assistant or configuring an LLM provider. The MCP Server is also enabled automatically when `ai.assistant.enabled=true`. | bool | `false` |
| ai.platformOptimization.enabled | Enables advanced AI-driven automatic troubleshooting extensions. This builds upon the AI Assistant framework to provide proactive issue resolution. | bool | `false` |
| anomaly-detection.enabled | Enables anomaly detection chart | bool | `false` |
| anomaly-detection.pdb | PodDisruptionBudget of the anomaly detection pods. | object | `{"maxUnavailable":0}` |
| anomaly-detection.podAnnotations | Extra annotations for the anomaly detection manager and worker pods. Configuration checksum annotations are preserved. | object | `{}` |
| backup.additionalLogging | Additional logback config for backup components | string | `""` |
| backup.configuration.bucketName | Name of the storage bucket to store configuration backups. | string | `"sts-configuration-backup"` |
| backup.configuration.enabled | Enable scheduled configuration/settings backups. Set to false to disable (do NOT use an invalid cron schedule). Runs independently of global.backup.enabled (writes a local settings snapshot); the remote upload step is separately gated by global.backup.enabled. | bool | `true` |
| backup.configuration.maxLocalFiles | The maximum number of configuration backups stored on the settings PVC (see backup.storage.settingsPvc.size). Each backup consists of a settings file and a StackPacks archive; StackPacks archives without a matching settings file are removed. | int | `10` |
| backup.configuration.s3Prefix | Prefix (dir name) used to store backup files. | string | `""` |
| backup.configuration.scheduled.backupRetentionTimeDelta | Time to keep configuration backups. The value is passed to GNU date tool to determine a specific date, and files older than this date will be deleted. | string | `"365 days ago"` |
| backup.configuration.scheduled.pvc.accessModes | Access mode for settings backup data. | list | `["ReadWriteOnce"]` |
| backup.configuration.scheduled.pvc.size | Size of volume for settings backup | string | `"1Gi"` |
| backup.configuration.scheduled.pvc.storageClass | Storage class of the volume for settings backup data. | string | `nil` |
| backup.configuration.scheduled.schedule | Cron schedule for automatic configuration backups in [Kubernetes cron schedule syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#cron-schedule-syntax). | string | `"0 4 * * *"` |
| backup.configuration.securityContext.enabled | Whether or not to enable the securityContext | bool | `true` |
| backup.configuration.securityContext.fsGroup | The GID (group ID) of all files on all mounted volumes | int | `65534` |
| backup.configuration.securityContext.runAsGroup | The GID (group ID) of the owning user of the process | int | `65534` |
| backup.configuration.securityContext.runAsNonRoot | Ensure that the user is not root (!= 0) | bool | `true` |
| backup.configuration.securityContext.runAsUser | The UID (user ID) of the owning user of the process | int | `65534` |
| backup.configuration.yaml.maxSizeLimit | Max size of the settings backup or installed via a stackpack | string | `"100Mi"` |
| backup.elasticsearch.bucketName | Name of the storage bucket where ElasticSearch snapshots are stored. | string | `"sts-elasticsearch-backup"` |
| backup.elasticsearch.enabled | Enable scheduled ElasticSearch snapshots (creation of the SLM policy). Set to false to disable (do NOT use an invalid cron schedule). Only takes effect when global.backup.enabled is true. | bool | `true` |
| backup.elasticsearch.s3Prefix | Prefix (dir name) used to store backup files. | string | `""` |
| backup.elasticsearch.scheduled.schedule | Cron schedule for automatic ElasticSearch snaphosts in [ElasticSearch cron schedule syntax](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/cron-expressions.html). | string | `"0 0 3 * * ?"` |
| backup.elasticsearch.scheduled.snapshotRetentionExpireAfter | Amount of time to keep ElasticSearch snapshots in [ElasticSearch time units](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/common-options.html#time-units). *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). | string | `"30d"` |
| backup.elasticsearch.scheduled.snapshotRetentionMaxCount | Minimum number of ElasticSearch snapshots to keep. *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). | string | `"30"` |
| backup.elasticsearch.scheduled.snapshotRetentionMinCount | Minimum number of ElasticSearch snapshots to keep. *Note:* By default, the retention task itself [runs daily at 1:30 AM UTC](https://www.elastic.co/guide/en/elasticsearch/reference/7.6/slm-settings.html#slm-retention-schedule). | string | `"5"` |
| backup.elasticsearch.securityContext.enabled | Whether or not to enable the securityContext | bool | `true` |
| backup.elasticsearch.securityContext.fsGroup | The GID (group ID) of all files on all mounted volumes | int | `65534` |
| backup.elasticsearch.securityContext.runAsGroup | The GID (group ID) of the owning user of the process | int | `65534` |
| backup.elasticsearch.securityContext.runAsNonRoot | Ensure that the user is not root (!= 0) | bool | `true` |
| backup.elasticsearch.securityContext.runAsUser | The UID (user ID) of the owning user of the process | int | `65534` |
| backup.enabled | Enables backup/restore for all data | bool | `false` |
| backup.initJobAnnotations | Annotations for Backup-init Job. | object | `{}` |
| backup.manualJobs.containerSecurityContext | Container security overrides for all manual Elasticsearch, StackGraph, configuration and VictoriaMetrics backup/restore Jobs embedded in the backup scripts ConfigMap, including their init containers. Merged over common.container.securityContext; explicit false and zero override inherited values. Does not affect scheduled backups, the settings-PVC initializer, or other workloads. Pod security remains controlled by the component's securityContext. | object | `{}` |
| backup.poddisruptionbudget.maxUnavailable | Maximum number of pods that can be unavailable during the backup. | int | `0` |
| backup.stackGraph.bucketName | Name of the storage bucket to store StackGraph backups. | string | `"sts-stackgraph-backup"` |
| backup.stackGraph.enabled | Enable scheduled StackGraph backups. Set to false to disable (do NOT use an invalid cron schedule). Only takes effect when global.backup.enabled is true. | bool | `true` |
| backup.stackGraph.restore.tempData.accessModes | Access mode for temporary backup data. | list | `["ReadWriteOnce"]` |
| backup.stackGraph.restore.tempData.size | Size of volume for temporary backup data; must be large enough to store a single backup. It defaults to a datanode size on Distributed Mode and to the stackgraph size on Mono mode. | string | `nil` |
| backup.stackGraph.restore.tempData.storageClass | Storage class of the volume for temporary backup data. | string | `nil` |
| backup.stackGraph.s3Prefix | Prefix (dir name) used to store backup files. | string | `""` |
| backup.stackGraph.scheduled.backupRetentionTimeDelta | Time to keep StackGraph backups in. The value is passed to GNU date tool  to determine a specific date, and files older than this date will be deleted. | string | `"30 days ago"` |
| backup.stackGraph.scheduled.implementation | Allows choosing the backup implementation to run on a schedule, can be "v1", "v2" or "all". For now the default is "v1", due to the v2 system not declared 'stable' yet. Not-schedule jobs will be suspended. | string | `"all"` |
| backup.stackGraph.scheduled.schedule | Cron schedule for automatic StackGraph backups in [Kubernetes cron schedule syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#cron-schedule-syntax). | string | `"0 3 * * *"` |
| backup.stackGraph.scheduled.tempData.accessModes | Access mode for temporary backup data. | list | `["ReadWriteOnce"]` |
| backup.stackGraph.scheduled.tempData.size | Size of volume for temporary backup data; must be large enough to store a single backup. It defaults to a datanode size on Distributed Mode and to the stackgraph size on Mono mode. | string | `nil` |
| backup.stackGraph.scheduled.tempData.storageClass | Storage class of the volume for temporary backup data. | string | `nil` |
| backup.stackGraph.securityContext.enabled | Whether or not to enable the securityContext | bool | `true` |
| backup.stackGraph.securityContext.fsGroup | The GID (group ID) of all files on all mounted volumes | int | `65534` |
| backup.stackGraph.securityContext.runAsGroup | The GID (group ID) of the owning user of the process | int | `65534` |
| backup.stackGraph.securityContext.runAsNonRoot | Ensure that the user is not root (!= 0) | bool | `true` |
| backup.stackGraph.securityContext.runAsUser | The UID (user ID) of the owning user of the process | int | `65534` |
| backup.stackGraph.v2.tempData.accessModes | Access mode for temporary backup data. | list | `["ReadWriteOnce"]` |
| backup.stackGraph.v2.tempData.size | Size of volume for temporary backup data; only used as a buffer for sending data to s3, so does not need to fit a complete backup. | string | `"20Gi"` |
| backup.stackGraph.v2.tempData.storageClass | Storage class of the volume for temporary backup data. | string | `nil` |
| backup.storage.backend.azure.accountKey | Azure storage account key (optional, falls back to managed identity) | string | `""` |
| backup.storage.backend.azure.accountName | Azure storage account name | string | `""` |
| backup.storage.backend.azure.enabled | Enable Azure backend | bool | `false` |
| backup.storage.backend.azure.endpoint | Azure blob endpoint (auto-derived from accountName if not set) | string | `""` |
| backup.storage.backend.azure.fromExternalSecret | Use an externally-managed secret for Azure backend credentials (keys: azureAccountName, azureAccountKey). When set, accountName/accountKey values are not required. | string | `""` |
| backup.storage.backend.pvc.accessModes | Access modes for the PVC | list | `["ReadWriteOnce"]` |
| backup.storage.backend.pvc.enabled | Enable PVC backend | bool | `false` |
| backup.storage.backend.pvc.size | Size of the PVC for S3Proxy data | string | `"500Gi"` |
| backup.storage.backend.pvc.storageClass | Storage class for the PVC | string | `""` |
| backup.storage.backend.s3.accessKey | AWS access key (optional, falls back to instance profile/IRSA) | string | `""` |
| backup.storage.backend.s3.enabled | Enable S3 backend | bool | `false` |
| backup.storage.backend.s3.endpoint | S3 endpoint URL (optional, defaults to AWS) | string | `""` |
| backup.storage.backend.s3.fromExternalSecret | Use an externally-managed secret for S3 backend credentials (keys: backendAccessKey, backendSecretKey). When set, accessKey/secretKey values are not required. | string | `""` |
| backup.storage.backend.s3.region | AWS region (defaults to us-east-1) | string | `nil` |
| backup.storage.backend.s3.secretKey | AWS secret key (optional) | string | `""` |
| backup.storage.settingsPvc.accessModes | Access modes for the settings PVC | list | `["ReadWriteOnce"]` |
| backup.storage.settingsPvc.size | Size of the settings backup PVC | string | `"2Gi"` |
| backup.storage.settingsPvc.storageClass | Storage class for the settings PVC | string | `""` |
| clickhouse.affinity | Affinity for the ClickHouse pods. Each field set here replaces the same field from the global affinity. | object | `{}` |
| clickhouse.backup.affinity | Affinity for the ClickHouse backup CronJob. Each field set here replaces the same field from the global affinity. | object | `{}` |
| clickhouse.backup.bucketName | Name of the storage bucket where ClickHouse backups are stored. | string | `"sts-clickhouse-backup"` |
| clickhouse.backup.config.keep_remote | Number of ClickHouse backups, full and incremental, to keep in remote storage. `0` keeps all. | int | `308` |
| clickhouse.backup.enabled | Enable scheduled ClickHouse backups. Set to false to disable (do NOT use an invalid cron schedule). Only takes effect when global.backup.enabled is true. | bool | `true` |
| clickhouse.backup.nodeSelector | Node labels for pod assignment of the ClickHouse backup CronJob. `stackstate.components.backup` and `stackstate.components.all` do not apply to it. | object | `{}` |
| clickhouse.backup.podAnnotations | Extra annotations on the ClickHouse backup CronJob pods. | object | `{}` |
| clickhouse.backup.podLabels | Extra labels on the ClickHouse backup CronJob pods. | object | `{}` |
| clickhouse.backup.resources | Resources of the ClickHouse backup container. Limits use the key `limit`, not `limits`. | object | `{"limit":{"cpu":"100m","memory":"250Mi"},"requests":{"cpu":"50m","memory":"250Mi"}}` |
| clickhouse.backup.s3Prefix | Prefix (directory) for ClickHouse backup files in the bucket. | string | `""` |
| clickhouse.backup.scheduled.full_schedule | Cron schedule for full ClickHouse backups. | string | `"45 0 * * *"` |
| clickhouse.backup.scheduled.incremental_schedule | Cron schedule for incremental ClickHouse backups. | string | `"45 3-23 * * *"` |
| clickhouse.backup.tolerations | Tolerations for the ClickHouse backup CronJob. Falls back to `global.suseObservability` when unset. | string | `nil` |
| clickhouse.logLevel | ClickHouse server log level, for example `information`, `debug` or `trace`. | string | `"information"` |
| clickhouse.nodeSelector | Node labels for pod assignment of the ClickHouse pods. | object | `{}` |
| clickhouse.persistence.size | Size of persistent volume for each clickhouse pod | string | `nil` |
| clickhouse.persistence.storageClass | Storage class of the ClickHouse volumes. Falls back to `global.storageClass`. | string | `""` |
| clickhouse.podAnnotations | Extra annotations on the ClickHouse pods. Merged with the chart's own annotations. | object | `{}` |
| clickhouse.podLabels | Extra labels on the ClickHouse pods. Merged with the chart's own labels. | object | `{}` |
| clickhouse.priorityClassName | PriorityClass of the ClickHouse pods. | string | `""` |
| clickhouse.resources | Resource allocation for ClickHouse pods. Fields left unset are taken from the sizing profile. | object | `{}` |
| clickhouse.tolerations | Tolerations for the ClickHouse pods. Falls back to `global.suseObservability` when unset. | string | `nil` |
| clickhouse.volumePermissions.enabled | Run a root init container that fixes the ownership of the ClickHouse volumes. Not compatible with the restricted Pod Security Standard. | bool | `false` |
| cluster-role.enabled | Deploy the ClusterRole(s) and ClusterRoleBinding(s) together with the chart. Can be disabled if these need to be installed by an administrator of the Kubernetes cluster. | bool | `true` |
| commonLabels | Labels that will be added to all resources created by the stackstate chart (not the subcharts though) | object | `{}` |
| deployment.compatibleWithArgoCD | Whether to adjust the Chart to be compatible with ArgoCD. This feature is as of yet not deployed in the o11y-tenants and saas-tenants directories, so should be considered unfinished (see STAC-21445) | bool | `false` |
| elasticsearch.antiAffinity | Pod anti-affinity of the Elasticsearch pods: `hard` or `soft`. When unset, follows `global.suseObservability.affinity.podAntiAffinity` on HA profiles, and is `hard` otherwise. | string | `""` |
| elasticsearch.maxUnavailable | `maxUnavailable` of the Elasticsearch PodDisruptionBudget. Set to `0` to remove the PodDisruptionBudget. | int | `1` |
| elasticsearch.nodeAffinity | Node affinity for the Elasticsearch pods. When unset, the global node affinity is used. | string | `nil` |
| elasticsearch.nodeSelector | Node labels for pod assignment of the Elasticsearch pods. | object | `{}` |
| elasticsearch.podAnnotations | Extra annotations on the Elasticsearch pods. Merged with the chart's own annotations. | object | `{}` |
| elasticsearch.podLabels | Extra labels on the Elasticsearch pods. Merged with the chart's own labels. | object | `{}` |
| elasticsearch.priorityClassName | PriorityClass of the Elasticsearch pods. | string | `""` |
| elasticsearch.prometheus-elasticsearch-exporter.podAnnotations | Extra annotations on the Elasticsearch Prometheus exporter pods. Merged with the chart's own annotations. | object | `{}` |
| elasticsearch.prometheus-elasticsearch-exporter.podLabels | Extra labels on the Elasticsearch Prometheus exporter pods. Merged with the chart's own labels. | object | `{}` |
| elasticsearch.prometheus-elasticsearch-exporter.podSecurityContext | Pod security context of the Elasticsearch Prometheus exporter. On OpenShift set it to `""` so the SCC assigns the user. | object | `{"runAsNonRoot":true,"runAsUser":1000,"seccompProfile":{"type":"RuntimeDefault"}}` |
| elasticsearch.prometheus-elasticsearch-exporter.priorityClassName | PriorityClass of the Elasticsearch Prometheus exporter pods. | string | `""` |
| elasticsearch.prometheus-elasticsearch-exporter.resources | Resource allocation for the Elasticsearch Prometheus exporter. | object | `{}` |
| elasticsearch.resources | Override Elasticsearch resources | object | `{}` |
| elasticsearch.sysctlInitContainer.enabled | Run a privileged init container that sets `vm.max_map_count` on the node. Disable on OpenShift or under the restricted Pod Security Standard, and set the sysctl on the nodes instead. | bool | `true` |
| elasticsearch.tolerations | Tolerations for the Elasticsearch pods. Falls back to `global.suseObservability` when unset. | string | `nil` |
| elasticsearch.volumeClaimTemplate | PVC template defaulting to 250Gi default volumes | object | `{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":null}}}` |
| gateway.annotations | Annotations for HTTPRoute objects. | object | `{}` |
| gateway.backendWeight | Optional weight for traffic splitting. | string | `nil` |
| gateway.enabled | Enable Gateway API HTTPRoute for SUSE Observability. | bool | `false` |
| gateway.filters | Optional filters for the HTTPRoute rule. | list | `[]` |
| gateway.hostnames | List of hostnames for the HTTPRoute. If not specified, derived from baseUrl. | list | `[]` |
| gateway.parentRefs | List of parent Gateway references (required when gateway.enabled is true). | list | `[]` |
| gateway.path | Path prefix for the HTTPRoute rule. If not specified, derived from baseUrl or defaults to "/". | string | `""` |
| gateway.timeouts | Optional timeouts for the HTTPRoute rule. | object | `{}` |
| global.backup.enabled | Enables backup/restore for all data | bool | `false` |
| global.commonLabels | Labels that will be added to all Deployments, StatefulSets, CronJobs, Jobs and their pods | object | `{}` |
| global.imagePullSecrets | List of image pull secret names to be used by all images across all charts. | list | `[]` |
| global.imageRegistry | Image registry to be used by all images across all charts. When using global.suseObservability (global mode), set this to "registry.rancher.com" to match the default behavior of the suse-observability-values chart. | string | `nil` |
| global.s3proxy.credentials.accessKey | Access key for S3Proxy authentication (override for production usage with global.backup.enabled) | string | `"default-for-settings-only"` |
| global.s3proxy.credentials.fromExternalSecret | Use an externally-managed secret for credentials (keys: accesskey, secretkey). When set, the chart will not create a secret and accessKey/secretKey values are not required. | string | `""` |
| global.s3proxy.credentials.secretKey | Secret key for S3Proxy authentication (override for production usage with global.backup.enabled) | string | `"default-secret-for-settings-only"` |
| global.storageClass | StorageClass for all PVCs created by the chart. Can be overridden per PVC. | string | `nil` |
| global.suseObservability.adminPassword | Admin password for the default 'admin' user (plain text). Mutually exclusive with adminPasswordBcrypt. Required (one of the two) when using global.suseObservability configuration unless other authentication methods (LDAP, OIDC, Keycloak) are configured. | string | `""` |
| global.suseObservability.adminPasswordBcrypt | Admin password as bcrypt hash. Mutually exclusive with adminPassword. | string | `""` |
| global.suseObservability.affinity.nodeAffinity | Node affinity configuration applied to all components (application and infrastructure). Standard Kubernetes nodeAffinity spec. | string | `nil` |
| global.suseObservability.affinity.podAffinity | Pod affinity configuration for application components only. Does NOT apply to infrastructure components. Standard Kubernetes podAffinity spec. | string | `nil` |
| global.suseObservability.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution | Enable required (hard) pod anti-affinity. When true, pods must be scheduled on different nodes. When false, soft anti-affinity is used (preferred but not required). | bool | `true` |
| global.suseObservability.affinity.podAntiAffinity.topologyKey | Topology key for pod anti-affinity. Determines the domain for spreading pods (e.g., kubernetes.io/hostname for node-level, topology.kubernetes.io/zone for zone-level). | string | `"kubernetes.io/hostname"` |
| global.suseObservability.applicationDomains | Enabled application domains. Default is ["Observability"]. Only "Observability" is currently supported. | list | `["Observability"]` |
| global.suseObservability.baseUrl | Base URL for SUSE Observability (required when using global.suseObservability). | string | `""` |
| global.suseObservability.license | SUSE Observability license key (required when using global.suseObservability). | string | `""` |
| global.suseObservability.pullSecret.password | Password for image pull secret. | string | `""` |
| global.suseObservability.pullSecret.username | Username for image pull secret. | string | `""` |
| global.suseObservability.receiverApiKey | Optional, prefer to use Service Tokens instead for more control and better security. Will use stackstate.apiKey.key if not provided, stackstate.apiKey can also be used to provide the apiKey via an external secret. Please see the documentation on how to use Service Tokens instead. | string | `""` |
| global.suseObservability.sizing.profile | Sizing profile name. Must match one of the available profiles: 10-nonha, 20-nonha, 50-nonha, 100-nonha, 150-ha, 250-ha, 500-ha, 4000-ha, trial. The chart will automatically apply resource limits, replica counts, and affinity configurations based on the selected profile. These act as intelligent defaults that can be overridden by component-specific values. | string | `""` |
| global.suseObservability.tolerations | Tolerations applied to every SUSE Observability workload, including subcharts, hooks, backup and restore jobs and CronJobs. A component's own tolerations (or `stackstate.components.all.tolerations`) replace this list rather than extending it; set them to `[]` to clear. | string | `nil` |
| hbase.all.affinity | Affinity for all HBase components. Each field set on a component replaces the same field here. Falls back to `global.suseObservability` when unset. | object | `{}` |
| hbase.all.nodeSelector | Node labels for pod assignment on all HBase, HDFS, Tephra, StackGraph and console pods. | object | `{}` |
| hbase.all.podAnnotations | Extra annotations on all HBase component pods. | object | `{}` |
| hbase.all.podLabels | Extra labels on all HBase component pods. | object | `{}` |
| hbase.all.tolerations | Tolerations for all HBase components. A component's own `tolerations` replace this list. Falls back to `global.suseObservability` when unset. | string | `nil` |
| hbase.console.enabled | Deploy the StackGraph console Deployment. Its pods only run when `hbase.console.replicaCount` is above 0. | bool | `true` |
| hbase.console.podAnnotations | Extra annotations on the HBase console and integrity CronJob pods. Override `hbase.all.podAnnotations`. | object | `{}` |
| hbase.console.podLabels | Extra labels on the HBase console and integrity CronJob pods. Override `hbase.all.podLabels`; selector labels are preserved. | object | `{}` |
| hbase.console.replicaCount | Number of StackGraph console replicas. Set to 1 to run maintenance commands such as an immediate retention cleanup. | int | `0` |
| hbase.console.resources | Resource allocation for the StackGraph console pod (HA profiles). Fields left unset are taken from the sizing profile. | object | `{}` |
| hbase.hbase.master.podAnnotations | Extra annotations on the HBase master pods. Override `hbase.all.podAnnotations`. | object | `{}` |
| hbase.hbase.master.podLabels | Extra labels on the HBase master pods. Override `hbase.all.podLabels`; selector labels are preserved. | object | `{}` |
| hbase.hbase.master.resources | Resource allocation for the HBase master pods (HA profiles). Fields left unset are taken from the sizing profile. | object | `{}` |
| hbase.hbase.regionserver.podAnnotations | Extra annotations on the HBase region server pods. Override `hbase.all.podAnnotations`. | object | `{}` |
| hbase.hbase.regionserver.podLabels | Extra labels on the HBase region server pods. Override `hbase.all.podLabels`; selector labels are preserved. | object | `{}` |
| hbase.hbase.regionserver.resources | Resource allocation for the HBase region server pods (HA profiles). Fields left unset are taken from the sizing profile. | object | `{}` |
| hbase.hdfs.datanode.extraEnv.open | Extra environment variables for the HDFS datanode pods, for example `HDFS_CONF_*` settings. | object | `{}` |
| hbase.hdfs.datanode.persistence.size | Size of each HDFS datanode volume (HA profiles). When unset, the sizing profile sets it. | string | `nil` |
| hbase.hdfs.datanode.persistence.storageClass | Storage class of the HDFS datanode volumes. Falls back to `global.storageClass`. | string | `nil` |
| hbase.hdfs.datanode.podAnnotations | Extra annotations on the HDFS datanode pods. Override `hbase.all.podAnnotations`. | object | `{}` |
| hbase.hdfs.datanode.podLabels | Extra labels on the HDFS datanode pods. Override `hbase.all.podLabels`; selector labels are preserved. | object | `{}` |
| hbase.hdfs.datanode.replicaCount | Number of HDFS datanode replicas. When unset, the sizing profile sets it. | string | `nil` |
| hbase.hdfs.datanode.resources | Resource allocation for the HDFS datanode pods (HA profiles). Fields left unset are taken from the sizing profile. | object | `{}` |
| hbase.hdfs.minReplication | Minimum number of HDFS block replicas a write needs to succeed (HA profiles). Capped at the number of datanodes. | int | `2` |
| hbase.hdfs.namenode.persistence.size | Size of the HDFS namenode volume (HA profiles). | string | `"20Gi"` |
| hbase.hdfs.namenode.podAnnotations | Extra annotations on the HDFS namenode pods. Override `hbase.all.podAnnotations`. | object | `{}` |
| hbase.hdfs.namenode.podLabels | Extra labels on the HDFS namenode pods. Override `hbase.all.podLabels`; selector labels are preserved. | object | `{}` |
| hbase.hdfs.namenode.resources | Resource allocation for the HDFS namenode pod (HA profiles). Fields left unset are taken from the sizing profile. | object | `{}` |
| hbase.hdfs.replication | HDFS block replication factor (HA profiles). When unset, the sizing profile sets it. Must be lower than `hbase.hdfs.datanode.replicaCount` and at least `hbase.hdfs.minReplication`. | string | `nil` |
| hbase.hdfs.secondarynamenode.podAnnotations | Extra annotations on the HDFS secondary namenode pods. Override `hbase.all.podAnnotations`. | object | `{}` |
| hbase.hdfs.secondarynamenode.podLabels | Extra labels on the HDFS secondary namenode pods. Override `hbase.all.podLabels`; selector labels are preserved. | object | `{}` |
| hbase.hdfs.secondarynamenode.resources | Resource allocation for the HDFS secondary namenode pod (HA profiles). Fields left unset are taken from the sizing profile. | object | `{}` |
| hbase.hdfs.volumePermissions.enabled | Run a root init container that fixes the ownership of the HDFS data volumes. Not compatible with the restricted Pod Security Standard. | bool | `false` |
| hbase.stackgraph.persistence.size | Size of the StackGraph volume (non-HA profiles). When unset, the sizing profile sets it. | string | `nil` |
| hbase.stackgraph.podAnnotations | Extra annotations on the StackGraph pods. Override `hbase.all.podAnnotations`. | object | `{}` |
| hbase.stackgraph.podLabels | Extra labels on the StackGraph pods. Override `hbase.all.podLabels`; selector labels are preserved. | object | `{}` |
| hbase.stackgraph.resources | Resource allocation for the StackGraph pod (non-HA profiles). Fields left unset are taken from the sizing profile. | object | `{}` |
| hbase.tephra.persistence.size | Size of the Tephra snapshot volume (non-HA profiles). | string | `"1Gi"` |
| hbase.tephra.podAnnotations | Extra annotations on the Tephra pods. Override `hbase.all.podAnnotations`. | object | `{}` |
| hbase.tephra.podLabels | Extra labels on the Tephra pods. Override `hbase.all.podLabels`; selector labels are preserved. | object | `{}` |
| hbase.tephra.resources | Resource allocation for the Tephra pods. Fields left unset are taken from the sizing profile. | object | `{}` |
| ingress.annotations | Annotations for ingress objects. | object | `{}` |
| ingress.enabled | Enable use of ingress controllers. | bool | `false` |
| ingress.hosts | List of ingress hostnames; the paths are fixed to StackState backend services | list | `[]` |
| ingress.ingressClassName | Ingress class name for the ingress object. | string | `""` |
| ingress.path | Path of the ingress rule for each host. | string | `"/"` |
| ingress.tls | List of ingress TLS certificates to use. | list | `[]` |
| kafka.affinity | Affinity for the Kafka pods. Each field set here replaces the same field from the global affinity. | object | `{}` |
| kafka.metrics.jmx.resources | Resource allocation for the Kafka JMX exporter container. | object | `{"limits":{"cpu":"1","ephemeral-storage":"1Gi","memory":"300Mi"},"requests":{"cpu":"200m","ephemeral-storage":"1Mi","memory":"300Mi"}}` |
| kafka.nodeSelector | Node labels for pod assignment of the Kafka pods. | object | `{}` |
| kafka.pdb | PodDisruptionBudget of the Kafka pods. Only created on HA profiles. | object | `{"create":true,"maxUnavailable":1,"minAvailable":""}` |
| kafka.persistence.size | Size of persistent volume for each Kafka pod | string | `nil` |
| kafka.persistence.storageClass | Storage class of the Kafka volumes. Falls back to `global.storageClass`. | string | `""` |
| kafka.podAnnotations | Extra annotations on the Kafka pods. Merged with the chart's own annotations. | object | `{}` |
| kafka.podLabels | Extra labels on the Kafka pods. Merged with the chart's own labels. | object | `{}` |
| kafka.priorityClassName | PriorityClass of the Kafka pods. | string | `""` |
| kafka.resources | Kafka resources per pods. | object | `{}` |
| kafka.tolerations | Tolerations for the Kafka pods. Falls back to `global.suseObservability` when unset. | string | `nil` |
| kafka.volumePermissions.enabled | Run a root init container that fixes the ownership of the Kafka volumes. Not compatible with the restricted Pod Security Standard. | bool | `false` |
| kafkaup-operator.podAnnotations | Extra annotations for the KafkaUp operator pod. The configuration checksum annotation is preserved. | object | `{}` |
| kafkaup-operator.podLabels | Extra labels for the KafkaUp operator pod. Selector labels are preserved. | object | `{}` |
| kafkaup-operator.resources | Resources of the KafkaUp operator pod. Fields left unset are taken from the sizing profile. | object | `{"limits":{"cpu":"50m","memory":"64Mi"},"requests":{"cpu":"25m","memory":"32Mi"}}` |
| kubernetes-rbac-agent.containers.rbacAgent.affinity | Set affinity | object | `{}` |
| kubernetes-rbac-agent.containers.rbacAgent.nodeSelector | Set a nodeSelector | object | `{}` |
| kubernetes-rbac-agent.containers.rbacAgent.podAnnotations | Additional annotations on the pod | object | `{"ad.stackstate.com/kubernetes-rbac-agent.check_names":"[\"openmetrics\"]","ad.stackstate.com/kubernetes-rbac-agent.init_configs":"[{}]","ad.stackstate.com/kubernetes-rbac-agent.instances":"[ { \"prometheus_url\": \"http://%%host%%:8080/metrics\", \"namespace\": \"stackstate\", \"metrics\": [\"*\"] } ]"}` |
| kubernetes-rbac-agent.containers.rbacAgent.podLabels | Additional labels on the pod | object | `{}` |
| kubernetes-rbac-agent.containers.rbacAgent.priorityClassName | PriorityClass of the RBAC agent pod. | string | `""` |
| kubernetes-rbac-agent.containers.rbacAgent.resources.limits.memory | Memory resource limits. | string | `"40Mi"` |
| kubernetes-rbac-agent.containers.rbacAgent.resources.requests.memory | Memory resource requests. | string | `"25Mi"` |
| kubernetes-rbac-agent.containers.rbacAgent.tolerations | Set tolerations | string | `nil` |
| networkPolicy.enabled | Enable creating of `NetworkPolicy` object and associated rules for StackState. | bool | `false` |
| networkPolicy.spec | `NetworkPolicy` rules for StackState. | object | `{"ingress":[{"from":[{"podSelector":{}}]}],"podSelector":{"matchLabels":{}},"policyTypes":["Ingress"]}` |
| opentelemetry-collector.affinity | Affinity for the OpenTelemetry collector pods. Each field set here replaces the same field from the global affinity. | object | `{}` |
| opentelemetry-collector.extraEnvs | Extra environment variables for the OpenTelemetry collector, as a list of Kubernetes `EnvVar`. Cannot override `MY_POD_IP`, `K8S_NODE_NAME`, `API_URL`, `INTAKE_URL`, or a chart-generated `GOMEMLIMIT`. | list | `[]` |
| opentelemetry-collector.gateway | Gateway API route for the OpenTelemetry collector: `enabled`, `kind` (`HTTPRoute` or `GRPCRoute`), `parentRefs` (required when enabled), `hostnames`, `annotations`, `path`, `pathType`, `filters`, `timeouts`, `port`, and `additionalGateways`. | object | `{"additionalGateways":[],"annotations":{},"enabled":false,"filters":[],"hostnames":[],"kind":"HTTPRoute","parentRefs":[],"path":"/","pathType":"PathPrefix","timeouts":{}}` |
| opentelemetry-collector.ingress | Ingress for the OpenTelemetry collector OTLP endpoints: `enabled`, `annotations`, `ingressClassName`, `hosts` (each path takes `path`, `pathType`, `port` and an optional `serviceName`), `tls`, and `additionalIngresses` (each with a unique `name`). Cannot be enabled together with `opentelemetry-collector.gateway`. | object | `{"additionalIngresses":[],"enabled":false}` |
| opentelemetry-collector.nodeSelector | Node labels for pod assignment of the OpenTelemetry collector pods. | object | `{}` |
| opentelemetry-collector.podAnnotations | Extra annotations on the OpenTelemetry collector pods. Merged with the chart's own annotations. | object | `{}` |
| opentelemetry-collector.podLabels | Extra labels on the OpenTelemetry collector pods. Merged with the chart's own labels. | object | `{}` |
| opentelemetry-collector.priorityClassName | PriorityClass of the OpenTelemetry collector pods. | string | `""` |
| opentelemetry-collector.resources | Resource allocation for the OpenTelemetry collector pods. | object | `{"limits":{"cpu":"500m","memory":"512Mi"},"requests":{"cpu":"250m","memory":"512Mi"}}` |
| opentelemetry-collector.tolerations | Tolerations for the OpenTelemetry collector pods. Falls back to `global.suseObservability` when unset. | string | `nil` |
| pull-secret.credentials | Registry and associated credentials (username, password) that will be stored in the pull-secret | list | `[]` |
| pull-secret.enabled | Deploy the ImagePullSecret for the chart. | bool | `false` |
| s3proxy.affinity | Affinity settings for the S3Proxy pod. Each field set here replaces the same field from `stackstate.components.all.affinity` and global affinity. | object | `{}` |
| s3proxy.extraEnv.open | Extra open environment variables to inject into the S3Proxy pod. | object | `{}` |
| s3proxy.extraEnv.secret | Extra secret environment variables to inject into the S3Proxy pod via a `Secret` object. | object | `{}` |
| s3proxy.logLevel | Log level of S3Proxy. | string | `"info"` |
| s3proxy.metrics.agentAnnotationsEnabled | Put annotations on each pod to instruct the stackstate agent to scrape the metrics | bool | `true` |
| s3proxy.metrics.enabled | Enable / disable S3Proxy Prometheus metrics. | bool | `true` |
| s3proxy.metrics.servicemonitor.additionalLabels | Additional labels for targeting Prometheus operator instances. | object | `{}` |
| s3proxy.metrics.servicemonitor.enabled | Enable `ServiceMonitor` object; `all.metrics.enabled` *must* be enabled. | bool | `false` |
| s3proxy.nodeSelector | Node selector for S3Proxy pod (merged with stackstate.components.all.nodeSelector) | object | `{}` |
| s3proxy.podAnnotations | Annotations for S3Proxy pod | object | `{}` |
| s3proxy.resources | Resource limits and requests for S3Proxy container | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"700Mi"},"requests":{"cpu":"200m","ephemeral-storage":"1Mi","memory":"700Mi"}}` |
| s3proxy.serviceAccount.annotations | Annotations for the S3Proxy service account (e.g. for IAM roles). | object | `{}` |
| s3proxy.tolerations | Tolerations for the S3Proxy pod. Replaces `stackstate.components.all.tolerations` and global tolerations when set. | string | `nil` |
| scc.enabled | Create `SecurityContextConstraints` resource to manage Openshift security constraints for Stackstate. Has to be enabled when installing to Openshift >= 4.12 The resource is deployed as a Helm pre-install hook to avoid any warning for the first deployment. Because `helm uninstall` does not consider Helm hooks, the resource must be manually deleted after the Helm release is removed. | bool | `false` |
| stackstate.allowedOrigins | Third-party web domains allowed to make cross-origin requests | list | `[]` |
| stackstate.apiKey.fromExternalSecret | Use an external secret for the api key. This suppresses secret creation by StackState and gets the data from the secret with the provided name. | string | `nil` |
| stackstate.apiKey.key | API key to be used by the Receiver. | string | `nil` |
| stackstate.authentication.adminPassword | Password for the 'admin' user that StackState creates by default | string | `nil` |
| stackstate.authentication.file | Configure users, their passwords and roles from (config) file | object | `{}` |
| stackstate.authentication.fromExternalSecret | Use an external secret for the authenticated secrets. This suppresses secret creation by StackState and gets the data from the secret with the provided name. | string | `nil` |
| stackstate.authentication.keycloak | Use Keycloak as authentication provider. See [Configuring Keycloak](#configuring-keycloak). Set `oidcLogout: true` inside this block to enable OIDC RP-initiated logout and `skipLoginPage: true` to redirect straight to Keycloak (both default false). | object | `{}` |
| stackstate.authentication.ldap | LDAP settings for StackState. See [Configuring LDAP](#configuring-ldap). | object | `{}` |
| stackstate.authentication.oidc | Use an OpenId Connect provider for authentication. See [Configuring OpenId Connect](#configuring-openid-connect). Set `oidcLogout: true` inside this block to enable OIDC RP-initiated logout and `skipLoginPage: true` to redirect straight to the provider (both default false). | object | `{}` |
| stackstate.authentication.rancher | Use Rancher as an OpenId Connect provider for authentication. See [Configuring Rancher authentication](#configuring-rancher-authentication). Optional `scope` array defaults to `[openid, profile, offline_access]`; add `groups` for Rancher 2.14 or later after allowing it on the OIDCClient. Set `oidcLogout: true` inside this block to enable OIDC RP-initiated logout and `skipLoginPage: true` to redirect straight to Rancher (both default false). | object | `{}` |
| stackstate.authentication.roles.admin | Extend the role names that have admin permissions (default: 'stackstate-admin') | list | `[]` |
| stackstate.authentication.roles.custom | Extend the authorization with custom roles {roleName: {systemPermissions: [], resourcePermissions: {}, viewPermissions: [], topologyScope: ""}} | object | `{}` |
| stackstate.authentication.roles.guest | Extend the role names that have guest permissions (default: 'stackstate-guest') | list | `[]` |
| stackstate.authentication.roles.k8sTroubleshooter | Extend the role names that have troubleshooter permissions (default: 'stackstate-k8s-troubleshooter') | list | `[]` |
| stackstate.authentication.roles.powerUser | Extend the role names that have power user permissions (default: 'stackstate-power-user') | list | `[]` |
| stackstate.authentication.serviceToken.bootstrap.dedicatedSubject | Subject solely created for usage with this token and cleaned up with the token | string | `""` |
| stackstate.authentication.serviceToken.bootstrap.roles | The roles the service token assumes when it’s used for authentication | list | `[]` |
| stackstate.authentication.serviceToken.bootstrap.token | The service token to set as bootstrap token | string | `""` |
| stackstate.authentication.serviceToken.bootstrap.ttl | The amount of time the service token is valid for | string | `"24h"` |
| stackstate.authentication.sessionLifetime | Amount of time to keep a session when a user does not log in | string | `""` |
| stackstate.baseUrl | **PROVIDE YOUR BASE URL HERE** Externally visible baseUrl of StackState. | string | `nil` |
| stackstate.components.aiAssistant.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.aiAssistant.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: secret-key - name: MY_OTHER_SECRET_ENV_VAR   secretName: my-other-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.aiAssistant.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.aiAssistant.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.aiAssistant.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.aiAssistant.persistence.size | Size of the PVC for the SQLite database. | string | `"1Gi"` |
| stackstate.components.aiAssistant.persistence.storageClass | Storage class for the SQLite database PVC. | string | `""` |
| stackstate.components.aiAssistant.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.aiAssistant.poddisruptionbudget | PodDisruptionBudget settings for `ai-assistant` pods. | object | `{"maxUnavailable":1}` |
| stackstate.components.aiAssistant.resources | Resource allocation for `ai-assistant` pods. | object | `{}` |
| stackstate.components.aiAssistant.serviceAccount.annotations | Annotations for the ai-assistant `ServiceAccount`. Useful for e.g. AWS IAM Role binding via IRSA: `eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/my-ai-assistant-role`. | object | `{}` |
| stackstate.components.aiAssistant.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.all.affinity | Affinity settings for pod assignment on all components. | object | `{}` |
| stackstate.components.all.deploymentStrategy.type | Deployment strategy for StackState components. Possible values: `RollingUpdate`, `Recreate` and `RecreateSingletonsOnly`. `RecreateSingletonsOnly` uses `Recreate` for the singleton Deployments and `RollingUpdate` for the other Deployments. | string | `"RecreateSingletonsOnly"` |
| stackstate.components.all.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.all.extraEnv.open | Extra open environment variables to inject into pods for all components. | object | `{}` |
| stackstate.components.all.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object for all components. | object | `{}` |
| stackstate.components.all.metrics.agentAnnotationsEnabled | Put annotations on each pod to instruct the stackstate agent to scrape the metrics | bool | `true` |
| stackstate.components.all.metrics.enabled | Enable metrics port. | bool | `true` |
| stackstate.components.all.metrics.servicemonitor.additionalLabels | Additional labels for targeting Prometheus operator instances. | object | `{}` |
| stackstate.components.all.metrics.servicemonitor.enabled | Enable `ServiceMonitor` object; `all.metrics.enabled` *must* be enabled. | bool | `false` |
| stackstate.components.all.nodeSelector | Node labels for pod assignment on all components. | object | `{}` |
| stackstate.components.all.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.all.tolerations | Toleration labels for pod assignment on all components. | string | `nil` |
| stackstate.components.api.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.api.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.api.config | Configuration file contents to customize the default StackState api configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. | string | `""` |
| stackstate.components.api.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.api.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.api.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.api.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.api.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.api.poddisruptionbudget | PodDisruptionBudget settings for `api` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.api.resources | Resource allocation for `api` pods. | object | `{}` |
| stackstate.components.api.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.api.yaml.maxSizeLimit | Max size of the sty file that will be imported or installed via a stackpack, e.g. "1Mi". | string | `nil` |
| stackstate.components.authorizationSync.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.authorizationSync.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.authorizationSync.config | Configuration file contents to customize the default StackState notification configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. | string | `""` |
| stackstate.components.authorizationSync.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.authorizationSync.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.authorizationSync.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.authorizationSync.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.authorizationSync.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.authorizationSync.poddisruptionbudget | PodDisruptionBudget settings for `authorizationSync` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.authorizationSync.resources | Resource allocation for `notification` pods. | object | `{}` |
| stackstate.components.authorizationSync.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.backup.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.backup.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.backup.podAnnotations | Extra annotations for backup pods. | object | `{}` |
| stackstate.components.backup.podLabels | Extra labels for backup pods. | object | `{}` |
| stackstate.components.backup.resources | Resource allocation for `backup` pods. | object | `{"limits":{"cpu":"3000m","ephemeral-storage":"1Gi","memory":"4000Mi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Mi","memory":"4000Mi"}}` |
| stackstate.components.backup.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.checks.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.checks.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.checks.config | Configuration file contents to customize the default StackState state configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. | string | `""` |
| stackstate.components.checks.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.checks.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.checks.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.checks.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.checks.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.checks.poddisruptionbudget | PodDisruptionBudget settings for `checks` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.checks.resources | Resource allocation for `state` pods. | object | `{}` |
| stackstate.components.checks.tmpToPVC.storageClass | Storage class name of PersistentVolume used by /tmp directory. It stores temporary files/caches, so it should be the fastest possible. | string | `nil` |
| stackstate.components.checks.tmpToPVC.volumeSize | The size of the PersistentVolume for "/tmp" directory. | string | `"2Gi"` |
| stackstate.components.checks.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.clickhouseCleanup.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.clickhouseCleanup.extraEnv.open | Add additional environment variables to the pod | object | `{}` |
| stackstate.components.clickhouseCleanup.jobAnnotations | Annotations for clickhouseCleanup job. | object | `{}` |
| stackstate.components.clickhouseCleanup.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.clickhouseCleanup.podAnnotations | Extra annotations for clickhouse cleanup job pods. | object | `{}` |
| stackstate.components.clickhouseCleanup.podLabels | Extra labels for clickhouse cleanup job pods. | object | `{}` |
| stackstate.components.clickhouseCleanup.resources | Resource allocation for `clickhouseCleanup` pods. | object | `{"limits":{"cpu":"100m","ephemeral-storage":"1Gi","memory":"200Mi"},"requests":{"cpu":"100m","ephemeral-storage":"1Mi","memory":"200Mi"}}` |
| stackstate.components.clickhouseCleanup.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.configurationBackup.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.configurationBackup.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.configurationBackup.podAnnotations | Extra annotations for configuration backup pods. | object | `{}` |
| stackstate.components.configurationBackup.podLabels | Extra labels for configuration backup pods. | object | `{}` |
| stackstate.components.configurationBackup.resources | Resource allocation for `configurationBackup` pods. | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"1000Mi"},"requests":{"cpu":"1000m","ephemeral-storage":"100Mi","memory":"1000Mi"}}` |
| stackstate.components.configurationBackup.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.containerTools.resources | Resource allocation for `kafkaTopicCreate` pods. | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"2000Mi"}}` |
| stackstate.components.correlate.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.correlate.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.correlate.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.correlate.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.correlate.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.correlate.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.correlate.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.correlate.poddisruptionbudget | PodDisruptionBudget settings for `correlate` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.correlate.resources | Resource allocation for `correlate` pods. | object | `{}` |
| stackstate.components.correlate.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.e2es.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.e2es.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.e2es.esDiskSpaceShare | How much disk space from ElasticSearch can use for k8s events ingestion | string | `"10"` |
| stackstate.components.e2es.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.e2es.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.e2es.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.e2es.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.e2es.poddisruptionbudget | PodDisruptionBudget settings for `e2es` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.e2es.resources | Resource allocation for `e2es` pods. | object | `{}` |
| stackstate.components.e2es.retention | Number of days to keep the events data on Es | int | `30` |
| stackstate.components.e2es.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.healthSync.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.healthSync.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.healthSync.config | Configuration file contents to customize the default StackState healthSync configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. | string | `""` |
| stackstate.components.healthSync.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.healthSync.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.healthSync.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.healthSync.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.healthSync.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.healthSync.poddisruptionbudget | PodDisruptionBudget settings for `healthSync` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.healthSync.resources | Resource allocation for `healthSync` pods. | object | `{}` |
| stackstate.components.healthSync.tmpToPVC.storageClass | Storage class name of PersistentVolume used by /tmp directory. It stores temporary files/caches, so it should be the fastest possible. | string | `nil` |
| stackstate.components.healthSync.tmpToPVC.volumeSize | The size of the PersistentVolume for "/tmp" directory. | string | `"2Gi"` |
| stackstate.components.healthSync.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.initializer.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.initializer.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.initializer.config | Configuration file contents to customize the default StackState initializer configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. | string | `""` |
| stackstate.components.initializer.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.initializer.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.initializer.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.initializer.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.initializer.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.initializer.resources | Resource allocation for `initializer` pods. | object | `{}` |
| stackstate.components.initializer.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.kafkaTopicCreate.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.kafkaTopicCreate.extraEnv.open | Add additional environment variables to the pod | object | `{}` |
| stackstate.components.kafkaTopicCreate.jobAnnotations | Annotations for KafkaTopicCreate job. | object | `{}` |
| stackstate.components.kafkaTopicCreate.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.kafkaTopicCreate.podAnnotations | Extra annotations for kafka topic create job pods. | object | `{}` |
| stackstate.components.kafkaTopicCreate.podLabels | Extra labels for kafka topic create job pods. | object | `{}` |
| stackstate.components.kafkaTopicCreate.resources | Resource allocation for `kafkaTopicCreate` pods. | object | `{"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"2000Mi"},"requests":{"cpu":"500m","ephemeral-storage":"1Mi","memory":"2000Mi"}}` |
| stackstate.components.kafkaTopicCreate.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.mcp.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.mcp.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: secret-key - name: MY_OTHER_SECRET_ENV_VAR   secretName: my-other-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.mcp.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.mcp.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.mcp.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.mcp.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.mcp.poddisruptionbudget | PodDisruptionBudget settings for `mcp` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.mcp.replicaCount | Number of `mcp` replicas. | int | `1` |
| stackstate.components.mcp.resources | Resource allocation for `mcp` pods. | object | `{}` |
| stackstate.components.mcp.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.notification.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.notification.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.notification.config | Configuration file contents to customize the default StackState notification configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. | string | `""` |
| stackstate.components.notification.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.notification.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.notification.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.notification.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.notification.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.notification.poddisruptionbudget | PodDisruptionBudget settings for `notification` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.notification.resources | Resource allocation for `notification` pods. | object | `{}` |
| stackstate.components.notification.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.receiver.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.receiver.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.receiver.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.receiver.esDiskSpaceShare | How much disk space from ElasticSearch can use for k8s log ingestion | string | `"90"` |
| stackstate.components.receiver.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.receiver.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.receiver.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.receiver.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.receiver.poddisruptionbudget | PodDisruptionBudget settings for `receiver` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.receiver.resources | Resource allocation for `receiver` pods. | object | `{}` |
| stackstate.components.receiver.retention | Number of days to keep the logs data on Es | int | `7` |
| stackstate.components.receiver.split.base.affinity | Affinity settings for pod assignment. Each field set here replaces the same field from the component, `all` and global affinity. | object | `{}` |
| stackstate.components.receiver.split.base.extraEnv.open | Extra open environment variables to inject into pods. Will merge with stackstate.components.receiver.extraEnv.open | object | `{}` |
| stackstate.components.receiver.split.base.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.receiver.extraEnv.secret | object | `{}` |
| stackstate.components.receiver.split.base.nodeSelector | Additional node labels for pod assignment. | object | `{}` |
| stackstate.components.receiver.split.base.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.receiver.split.base.resources | Resource allocation for pods. If not defined, will take from sizing profile defaults. | object | `{}` |
| stackstate.components.receiver.split.base.tolerations | Toleration labels for pod assignment. Replaces the component, `all` and global tolerations when set. | string | `nil` |
| stackstate.components.receiver.split.logs.affinity | Affinity settings for pod assignment. Each field set here replaces the same field from the component, `all` and global affinity. | object | `{}` |
| stackstate.components.receiver.split.logs.extraEnv.open | Extra open environment variables to inject into pods. Will merge with stackstate.components.receiver.extraEnv.open | object | `{}` |
| stackstate.components.receiver.split.logs.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.receiver.extraEnv.secret | object | `{}` |
| stackstate.components.receiver.split.logs.nodeSelector | Additional node labels for pod assignment. | object | `{}` |
| stackstate.components.receiver.split.logs.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.receiver.split.logs.resources | Resource allocation for pods. If not defined, will take from sizing profile defaults. | object | `{}` |
| stackstate.components.receiver.split.logs.tolerations | Toleration labels for pod assignment. Replaces the component, `all` and global tolerations when set. | string | `nil` |
| stackstate.components.receiver.split.processAgent.affinity | Affinity settings for pod assignment. Each field set here replaces the same field from the component, `all` and global affinity. | object | `{}` |
| stackstate.components.receiver.split.processAgent.extraEnv.open | Extra open environment variables to inject into pods. Will merge with stackstate.components.receiver.extraEnv.open | object | `{}` |
| stackstate.components.receiver.split.processAgent.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. Will merge with stackstate.components.receiver.extraEnv.secret | object | `{}` |
| stackstate.components.receiver.split.processAgent.nodeSelector | Additional node labels for pod assignment. | object | `{}` |
| stackstate.components.receiver.split.processAgent.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.receiver.split.processAgent.resources | Resource allocation for pods. If not defined, will take from sizing profile defaults. | object | `{}` |
| stackstate.components.receiver.split.processAgent.tolerations | Toleration labels for pod assignment. Replaces the component, `all` and global tolerations when set. | string | `nil` |
| stackstate.components.receiver.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.replicationChecker.affinity | Affinity settings for replication checker Pod assignment. | object | `{}` |
| stackstate.components.replicationChecker.enabled | Enable the stateless replication checker. Null enables it only for built-in HA sizing profiles; true or false overrides the profile default. | bool | `nil` |
| stackstate.components.replicationChecker.nodeSelector | Node labels for replication checker Pod assignment. | object | `{}` |
| stackstate.components.replicationChecker.podAnnotations | Extra annotations for replication checker Pods. | object | `{}` |
| stackstate.components.replicationChecker.resources.limits.cpu | CPU limit for the replication checker. | string | `"500m"` |
| stackstate.components.replicationChecker.resources.limits.memory | Memory limit for the replication checker. | string | `"256Mi"` |
| stackstate.components.replicationChecker.resources.requests.cpu | Requested CPU for the replication checker. | string | `"10m"` |
| stackstate.components.replicationChecker.resources.requests.memory | Requested memory for the replication checker. | string | `"64Mi"` |
| stackstate.components.replicationChecker.tolerations | Tolerations for replication checker Pod assignment. | string | `nil` |
| stackstate.components.router.accesslog.enabled | Enable access logging on the router | bool | `false` |
| stackstate.components.router.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.router.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.router.errorlog.enabled | Enable error logging on the router | bool | `true` |
| stackstate.components.router.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.router.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.router.mode.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.router.mode.extraEnv.open | Add additional environment variables to the pod | object | `{}` |
| stackstate.components.router.mode.jobAnnotations | Annotations for the router mode jobs. | object | `{}` |
| stackstate.components.router.mode.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.router.mode.podAnnotations | Extra annotations for router mode job pods. | object | `{}` |
| stackstate.components.router.mode.podLabels | Extra labels for router mode job pods. | object | `{}` |
| stackstate.components.router.mode.resources | Resource allocation for `router.mode` pods. | object | `{"limits":{"cpu":"200m","memory":"400Mi"},"requests":{"cpu":"100m","memory":"400Mi"}}` |
| stackstate.components.router.mode.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.router.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.router.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.router.poddisruptionbudget | PodDisruptionBudget settings for `router` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.router.resources | Resource allocation for `router` pods. | object | `{}` |
| stackstate.components.router.secureCookies.enabled | When enabled, the router appends the `Secure` attribute to `Set-Cookie` response headers for requests that reached the edge over HTTPS (detected via the `X-Forwarded-Proto: https` header set by a TLS-terminating proxy such as Traefik or an ingress controller). Plain-HTTP deployments are left untouched. | bool | `false` |
| stackstate.components.router.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.server.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.server.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.server.config | Configuration file contents to customize the default StackState configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. | string | `""` |
| stackstate.components.server.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.server.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.server.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.server.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.server.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.server.poddisruptionbudget | PodDisruptionBudget settings for `server` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.server.resources | Resource allocation for `server` pods. | object | `{}` |
| stackstate.components.server.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.slicing.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.slicing.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.slicing.config | Configuration file contents to customize the default StackState slicing configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. | string | `""` |
| stackstate.components.slicing.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.slicing.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.slicing.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.slicing.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.slicing.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.slicing.resources | Resource allocation for `slicing` pods. | object | `{}` |
| stackstate.components.slicing.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.state.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.state.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.state.config | Configuration file contents to customize the default StackState state configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. | string | `""` |
| stackstate.components.state.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.state.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.state.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.state.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.state.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.state.poddisruptionbudget | PodDisruptionBudget settings for `state` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.state.resources | Resource allocation for `state` pods. | object | `{}` |
| stackstate.components.state.tmpToPVC.storageClass | Storage class name of PersistentVolume used by /tmp directory. It stores temporary files/caches, so it should be the fastest possible. | string | `nil` |
| stackstate.components.state.tmpToPVC.volumeSize | The size of the PersistentVolume for "/tmp" directory. | string | `"2Gi"` |
| stackstate.components.state.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.sync.additionalLogging | Additional logback config | string | `""` |
| stackstate.components.sync.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.sync.config | Configuration file contents to customize the default StackState sync configuration, environment variables have higher precedence and can be used as overrides. StackState configuration is in the [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) format, see [StackState documentation](https://docs.stackstate.com/setup/installation/kubernetes/) for examples. | string | `""` |
| stackstate.components.sync.envsFromExistingSecrets | Configure environment variables from existing secrets. envsFromExistingSecret - name: MY_SECRET_ENV_VAR   secretName: my-k8s-secret   secretKey: my-secret-key - name: ANOTHER_ENV_VAR   secretName: another-k8s-secret   secretKey: another-secret-key | list | `[]` |
| stackstate.components.sync.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.sync.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.sync.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.sync.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.sync.poddisruptionbudget | PodDisruptionBudget settings for `sync` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.sync.resources | Resource allocation for `sync` pods. | object | `{}` |
| stackstate.components.sync.tmpToPVC.storageClass | Storage class name of PersistentVolume used by /tmp directory. It stores temporary files/caches, so it should be the fastest possible. | string | `nil` |
| stackstate.components.sync.tmpToPVC.volumeSize | The size of the PersistentVolume for "/tmp" directory. | string | `"2Gi"` |
| stackstate.components.sync.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.ui.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.ui.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.ui.extraEnv.secret | Extra secret environment variables to inject into pods via a `Secret` object. | object | `{}` |
| stackstate.components.ui.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.ui.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.ui.poddisruptionbudget | PodDisruptionBudget settings for `ui` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.ui.resources | Resource allocation for `ui` pods. | object | `{}` |
| stackstate.components.ui.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.components.vmagent.affinity | Affinity settings for vmagent pod. | object | `{"podAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app.kubernetes.io/component","operator":"In","values":["receiver"]},{"key":"app.kubernetes.io/instance","operator":"In","values":["stackstate"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":80}]}}` |
| stackstate.components.vmagent.persistence.size | Size of the `vmagent` buffer volume. | string | `"10Gi"` |
| stackstate.components.vmagent.persistence.storageClass | Storage class of the `vmagent` buffer volume. | string | `nil` |
| stackstate.components.vmagent.podAnnotations | Extra annotations for vmagent pods. Override stackstate.components.all.podAnnotations and metrics annotations; the generated configuration checksum is preserved. | object | `{}` |
| stackstate.components.vmagent.poddisruptionbudget | PodDisruptionBudget settings for `vmagent` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.vmagent.resources | Resource allocation for vmagent pod. | object | `{}` |
| stackstate.components.workloadObserver.affinity | Affinity settings for pod assignment. | object | `{}` |
| stackstate.components.workloadObserver.enabled | Enable/disable the workload observer | bool | `true` |
| stackstate.components.workloadObserver.extraEnv.open | Extra open environment variables to inject into pods. | object | `{}` |
| stackstate.components.workloadObserver.nodeSelector | Node labels for pod assignment. | object | `{}` |
| stackstate.components.workloadObserver.persistence.size | Size of the `workloadObserver` volume. | string | `"1Gi"` |
| stackstate.components.workloadObserver.persistence.storageClass | Storage class of the `workloadObserver` volume. | string | `nil` |
| stackstate.components.workloadObserver.podAnnotations | Extra annotations | object | `{}` |
| stackstate.components.workloadObserver.poddisruptionbudget | PodDisruptionBudget settings for `workloadObserver` pods. | object | `{"maxUnavailable":1,"unhealthyPodEvictionPolicy":"AlwaysAllow"}` |
| stackstate.components.workloadObserver.resources | Resource allocation for `workloadObserver` pods. | object | `{"limits":{"cpu":"50m","ephemeral-storage":"1Gi","memory":"128Mi"},"requests":{"cpu":"20m","ephemeral-storage":"1Mi","memory":"24Mi"}}` |
| stackstate.components.workloadObserver.tolerations | Toleration labels for pod assignment. | string | `nil` |
| stackstate.email.additionalProperties | Additional JavaMail properties for the SMTP connection. | object | `{"mail.smtp.auth":"true","mail.smtp.starttls.enable":"true"}` |
| stackstate.email.enabled | Enable email notifications | bool | `false` |
| stackstate.email.sender | Email sender mail address | string | `""` |
| stackstate.email.server.auth.fromExternalSecret | Use an external secret for the email configuration. This suppresses secret creation by StackState and gets the data from the secret with the provided name. | string | `nil` |
| stackstate.email.server.auth.password | Email server password | string | `""` |
| stackstate.email.server.auth.username | Email server username | string | `""` |
| stackstate.email.server.host | Email server host | string | `""` |
| stackstate.email.server.port | Email server port | int | `587` |
| stackstate.email.server.protocol | Email server protocol | string | `"smtp"` |
| stackstate.features.role-k8s-authz | Deploy the Role(s) to populate permissions on Suse Observability | boolean | `true` |
| stackstate.features.server.split | Run a single service server or split in multiple sub services as api, state .... | boolean | `true` |
| stackstate.features.storeTransactionLogsToPVC.enabled | Whether the transaction logs for some services, API, Checks, HealthSync,State and Sync have to be stored to PVCs instead of pod ephemeral storage. | boolean | `false` |
| stackstate.features.storeTransactionLogsToPVC.storageClass | Storage class name of PersistentVolume used by transaction logs. | string | `nil` |
| stackstate.features.storeTransactionLogsToPVC.volumeSize | The size of the persistent volume for the transaction logs. | string | `"600Mi"` |
| stackstate.java.trustStore | Java TrustStore (cacerts) file to use. Prefer `stackstate.java.trustStoreFromExternalSecret`, which keeps the trust store out of the size-limited Helm release secret. | string | `nil` |
| stackstate.java.trustStoreBase64Encoded | Base64 encoded Java TrustStore (cacerts) file to use. Ignored if stackstate.java.trustStore is set. Prefer `stackstate.java.trustStoreFromExternalSecret`, which keeps the trust store out of the size-limited Helm release secret. | string | `nil` |
| stackstate.java.trustStoreFromExternalSecret.key | Key in the external secret holding the Java TrustStore (cacerts) file. | string | `"java-cacerts"` |
| stackstate.java.trustStoreFromExternalSecret.name | Name of an existing secret holding the Java TrustStore (cacerts) file. This is the preferred way to configure the trust store: it keeps the trust store out of the Helm release secret, which has a 1MB size limit. Takes precedence over `stackstate.java.trustStore` and `stackstate.java.trustStoreBase64Encoded`. | string | `""` |
| stackstate.java.trustStoreFromExternalSecret.passwordKey | Optional key in the external secret holding the trust store password. When empty, `stackstate.java.trustStorePassword` is used instead. | string | `""` |
| stackstate.java.trustStorePassword | Password to access the Java TrustStore (cacerts) file | string | `nil` |
| stackstate.k8sAuthorization.enabled | Whether we use k8s RBAC as authorization authority | boolean | `true` |
| stackstate.license.fromExternalSecret | Use an external secret for the license key. This suppresses secret creation by StackState and gets the data from the secret with the provided name. | string | `nil` |
| stackstate.license.key | **PROVIDE YOUR LICENSE KEY HERE** The StackState license key needed to start the server. | string | `nil` |
| stackstate.silencing.maxSilenceTime | The maximum duration a monitor check state can be silenced for. Accepts any HOCON duration, for example "24 hours" or "30 minutes". | string | `"24 hours"` |
| stackstate.stackpacks.extraImages | Additional stackpack Docker images to mount after the built-in stackpack images. | list | `[]` |
| stackstate.stackpacks.installed | Specify a list of stackpacks to be always installed including their configuration, for an example see [Auto-installing StackPacks](#auto-installing-stackpacks) | list | `[]` |
| stackstate.stackpacks.localpvc.size | Size of the Persistent Volume Claim (PVC) used to persist stackpacks when there's no HDFS | string | `"1Gi"` |
| stackstate.stackpacks.localpvc.storageClass | Storage class of the PVC used to persist stackpacks when there is no HDFS. | string | `nil` |
| stackstate.stackpacks.pvc.size | Size of the Persistent Volume Claim (PVC) used to copy stackpacks from the Docker image. | string | `"1Gi"` |
| stackstate.stackpacks.pvc.storageClass | Storage class of the PVC used to copy stackpacks from the Docker image. | string | `nil` |
| stackstate.stackpacks.updateInterval | Interval at which StackState checks for StackPack updates. | string | `"5 minutes"` |
| stackstate.stackpacks.upgradeOnStartup | Specify additional stackpacks that will, on startup only, be upgraded to the latest version available. The following StackPacks are always upgraded on startup: kubernetes-v2, open-telemetry, stackstate-k8s-agent-v2, aad-v2, otel-k8s-crd, suse-observability, plus prime-kubernetes or community-kubernetes (based on stackstate.deployment.edition). The suse-observability StackPack replaces the stackstate StackPack and migrates an existing stackstate installation; do not add stackstate to this list. StackPacks listed here are added to these defaults. | list | `[]` |
| stackstate.topology.retentionHours | Number of hours topology will be retained. | integer | `nil` |
| stackstate.ui.defaultTimeRange | Default time range  in the UI. One of LAST_5_MINUTES, LAST_15_MINUTES, LAST_30_MINUTES, LAST_1_HOUR, LAST_3_HOURS, LAST_6_HOURS, LAST_12_HOURS, LAST_24_HOURS, LAST_2_DAYS. No value or an unsupported value will automatically fall-back to LAST_1_HOUR. | string | `nil` |
| victoria-metrics-0.backup.bucketName | Name of the storage bucket where Victoria Metrics backups are stored. | string | `"sts-victoria-metrics-backup"` |
| victoria-metrics-0.backup.enabled | Enable scheduled backups of this Victoria Metrics instance. Set to false to disable (do NOT use an invalid cron schedule). Only takes effect when global.backup.enabled is true. | bool | `true` |
| victoria-metrics-0.backup.keepLastDaily | Number of daily backups to retain | int | `7` |
| victoria-metrics-0.backup.keepLastWeekly | Number of weekly backups to retain | int | `4` |
| victoria-metrics-0.backup.minHistorySeconds | Whole number of seconds; refuse the hourly backup when local storage holds no data older than this while the destination already holds a completed backup, so a rebuilt instance cannot overwrite a backup it has not restored from. `0` disables the check. Must stay below this instance's retention (above it, no data can ever be that old and backups are refused permanently, not only after a redeploy) and at or under 40 days (above that VictoriaMetrics ignores the query time range and the check silently passes). Durations such as `1d` are not accepted. | int | `86400` |
| victoria-metrics-0.backup.s3Prefix | Prefix (dir name) used to store backup files, we may have multiple instances of Victoria Metrics, each of them should be stored into their own directory. | string | `"victoria-metrics-0"` |
| victoria-metrics-0.backup.scheduled.daily | Cron schedule for daily snapshot backups of Victoria Metrics | string | `"55 0 * * *"` |
| victoria-metrics-0.backup.scheduled.hourly | Cron schedule for hourly incremental backups of Victoria Metrics | string | `"25 * * * *"` |
| victoria-metrics-0.backup.setupCron.resources | Resources of the container that sets up the backup cron schedule. | object | `{"limits":{"cpu":"20m","memory":"64Mi"},"requests":{"cpu":"20m","memory":"64Mi"}}` |
| victoria-metrics-0.backup.vmbackup.resources | Resources of the Victoria Metrics backup container. | object | `{"limits":{"cpu":"100m","memory":"512Mi"},"requests":{"cpu":"10m","memory":"512Mi"}}` |
| victoria-metrics-0.server.affinity | Affinity for this Victoria Metrics instance. Each field set here replaces the same field from the global affinity. | object | `{}` |
| victoria-metrics-0.server.nodeSelector | Node labels for pod assignment of this Victoria Metrics instance. | object | `{}` |
| victoria-metrics-0.server.persistentVolume.size | Size of storage for Victoria Metrics, ideally 20% of free space remains available at all times | string | `nil` |
| victoria-metrics-0.server.persistentVolume.storageClass | Storage class of this Victoria Metrics volume. Falls back to `global.storageClass`. | string | `""` |
| victoria-metrics-0.server.podAnnotations | Extra annotations on the first Victoria Metrics instance pods. Merged with the chart's own annotations. | object | `{}` |
| victoria-metrics-0.server.podLabels | Extra labels on the first Victoria Metrics instance pods. Merged with the chart's own labels. | object | `{}` |
| victoria-metrics-0.server.priorityClassName | PriorityClass of the first Victoria Metrics instance pods. | string | `""` |
| victoria-metrics-0.server.resources | CPU/Memory requests and limits | object | `{}` |
| victoria-metrics-0.server.retentionPeriod | Metrics retention of this instance. A number without suffix is in months; `h`, `d`, `w` and `y` suffixes are supported. Set the same value on both instances. Ignored with the `trial` sizing profile, which sets 3d. | int | `1` |
| victoria-metrics-0.server.tolerations | Tolerations for this Victoria Metrics instance. Falls back to `global.suseObservability` when unset. | string | `nil` |
| victoria-metrics-1.backup.bucketName | Name of the storage bucket where Victoria Metrics backups are stored. | string | `"sts-victoria-metrics-backup"` |
| victoria-metrics-1.backup.enabled | Enable scheduled backups of this Victoria Metrics instance. Set to false to disable (do NOT use an invalid cron schedule). Only takes effect when global.backup.enabled is true. | bool | `true` |
| victoria-metrics-1.backup.keepLastDaily | Number of daily backups to retain | int | `7` |
| victoria-metrics-1.backup.keepLastWeekly | Number of weekly backups to retain | int | `4` |
| victoria-metrics-1.backup.minHistorySeconds | Whole number of seconds; refuse the hourly backup when local storage holds no data older than this while the destination already holds a completed backup, so a rebuilt instance cannot overwrite a backup it has not restored from. `0` disables the check. Must stay below this instance's retention (above it, no data can ever be that old and backups are refused permanently, not only after a redeploy) and at or under 40 days (above that VictoriaMetrics ignores the query time range and the check silently passes). Durations such as `1d` are not accepted. | int | `86400` |
| victoria-metrics-1.backup.s3Prefix | Prefix (dir name) used to store backup files, we may have multiple instances of Victoria Metrics, each of them should be stored into their own directory. | string | `"victoria-metrics-1"` |
| victoria-metrics-1.backup.scheduled.daily | Cron schedule for daily snapshot backups of Victoria Metrics | string | `"5 1 * * *"` |
| victoria-metrics-1.backup.scheduled.hourly | Cron schedule for hourly incremental backups of Victoria Metrics | string | `"35 * * * *"` |
| victoria-metrics-1.backup.setupCron.resources | Resources of the container that sets up the backup cron schedule. | object | `{"limits":{"cpu":"20m","memory":"64Mi"},"requests":{"cpu":"20m","memory":"64Mi"}}` |
| victoria-metrics-1.backup.vmbackup.resources | Resources of the Victoria Metrics backup container. | object | `{"limits":{"cpu":"100m","memory":"512Mi"},"requests":{"cpu":"10m","memory":"512Mi"}}` |
| victoria-metrics-1.server.affinity | Affinity for this Victoria Metrics instance. Each field set here replaces the same field from the global affinity. | object | `{}` |
| victoria-metrics-1.server.nodeSelector | Node labels for pod assignment of this Victoria Metrics instance. | object | `{}` |
| victoria-metrics-1.server.persistentVolume.size | Size of storage for Victoria Metrics, ideally 20% of free space remains available at all times | string | `nil` |
| victoria-metrics-1.server.persistentVolume.storageClass | Storage class of this Victoria Metrics volume. Falls back to `global.storageClass`. | string | `""` |
| victoria-metrics-1.server.podAnnotations | Extra annotations on the second Victoria Metrics instance pods. Merged with the chart's own annotations. | object | `{}` |
| victoria-metrics-1.server.podLabels | Extra labels on the second Victoria Metrics instance pods. Merged with the chart's own labels. | object | `{}` |
| victoria-metrics-1.server.priorityClassName | PriorityClass of the second Victoria Metrics instance pods. | string | `""` |
| victoria-metrics-1.server.resources | CPU/Memory requests and limits | object | `{}` |
| victoria-metrics-1.server.retentionPeriod | Metrics retention of this instance. A number without suffix is in months; `h`, `d`, `w` and `y` suffixes are supported. Set the same value on both instances. Ignored with the `trial` sizing profile, which sets 3d. | int | `1` |
| victoria-metrics-1.server.tolerations | Tolerations for this Victoria Metrics instance. Falls back to `global.suseObservability` when unset. | string | `nil` |
| zookeeper.affinity | Affinity for the ZooKeeper pods. Each field set here replaces the same field from the global affinity. | object | `{}` |
| zookeeper.nodeSelector | Node labels for pod assignment of the ZooKeeper pods. | object | `{}` |
| zookeeper.pdb | PodDisruptionBudget of the ZooKeeper pods. Only created on HA profiles. | object | `{"create":true,"maxUnavailable":1,"minAvailable":""}` |
| zookeeper.persistence.size | Size of the PVC for ZooKeeper data. When unset, the sizing profile sets it. | string | `nil` |
| zookeeper.persistence.storageClass | Storage class of the ZooKeeper volumes. Falls back to `global.storageClass`. | string | `""` |
| zookeeper.podAnnotations | Extra annotations on the ZooKeeper pods. Merged with the chart's own annotations. | object | `{}` |
| zookeeper.podLabels | Extra labels on the ZooKeeper pods. Merged with the chart's own labels. | object | `{}` |
| zookeeper.priorityClassName | PriorityClass of the ZooKeeper pods. | string | `""` |
| zookeeper.resources | Resource allocation for ZooKeeper pods. Fields left unset are taken from the sizing profile. | object | `{}` |
| zookeeper.tolerations | Tolerations for the ZooKeeper pods. Falls back to `global.suseObservability` when unset. | string | `nil` |
| zookeeper.volumePermissions.enabled | Run an init container that fixes the ownership of the ZooKeeper volumes. It sets no security context, so it runs as the image's default user. Not compatible with the restricted Pod Security Standard. | bool | `false` |

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

By default SUSE Observability appends `?client_name=StsOidcClient` to the redirect URI. Some providers, such as AD FS, require an exact redirect URI match but cannot register a URI with a query string. For those, set `omitClientNameParameter: true` inside the `oidc` (or `rancher`) block to send the `redirectUri` exactly as configured, and register `https://<base-url>/loginCallback` at the provider.

The OIDC-based authentication providers (`oidc`, `keycloak` and `rancher`) support two additional options that map directly to the SUSE Observability backend authentication configuration. Set them **inside the active provider block** &mdash; they have no effect for the `ldap` or file-based providers, and are ignored when placed anywhere else under `stackstate.authentication`:

- `oidcLogout` (default `false`): when enabled, logging out of SUSE Observability (via the `/logout` URL) also performs an OIDC RP-initiated logout, ending the user's session at the identity provider. When disabled, logging out only clears the local SUSE Observability session and the provider session stays active.
- `skipLoginPage` (default `false`): when enabled, the SUSE Observability login page is skipped and the user is redirected straight to the configured OIDC provider.

> **Note:** Enabling `skipLoginPage: true` while leaving `oidcLogout: false` is usually undesirable. The login page is skipped, but logging out only clears the local session, so the still-active provider session immediately re-authenticates the user and they appear unable to log out. When you enable `skipLoginPage`, enable `oidcLogout` as well unless you have a specific reason not to.

Example (shown for `oidc`; nest the same two keys under `keycloak` or `rancher` when using those providers):

```yaml
stackstate:
  authentication:
    oidc:
      oidcLogout: true
      skipLoginPage: true
      clientId: stackstate-client-id
      # ... remaining oidc configuration
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
- `omitClientNameParameter`

Set `stackstate.authentication.rancher.scope` to an array containing `openid` and any of `profile`, `offline_access`, and `groups`.
When omitted, the scope remains `[openid, profile, offline_access]`.
For Rancher versions before 2.14, omit this option because those versions do not support `groups`.

For Rancher 2.14 or later, first allow `openid`, `profile`, `offline_access`, and `groups` in the OIDCClient's `spec.scopes`.
Preserve any other scopes that the client requires.
An empty `spec.scopes` does not allow `groups`; requesting it can cause `invalid_scope` and prevent login.
Then configure SUSE Observability:

```yaml
stackstate:
  authentication:
    rancher:
      scope: [openid, profile, offline_access, groups]
```

Use the same sequence after upgrading Rancher to 2.14 or later.
After the Helm rollout, log out and log in again.
Check that the new session contains the expected groups and grants access through an existing group binding.
The Rancher preset keeps `usernameField = "sub"` and `groupsField = "groups"`.

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

Setting `stackstate.authentication.ldap.ssl.skipCertificateValidation: true` disables validation of the LDAP server certificate and hostname, and any configured trust store or certificates are ignored. The connection is still encrypted but is open to man-in-the-middle attacks, so use it only temporarily or for testing.

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

The recommended way to provide the trust store and the certificates is to create a secret for them up
front and refer to it from the values, so they never become part of the Helm release secret:

```shell
kubectl create secret generic ldap-truststore \
--from-file=cacerts=./ldap-cacerts \
--from-file=certificates.pem=./ldap-certificate.pem
```

```yaml
stackstate:
  authentication:
    ldap:
      ssl:
        type: ssl
        trustStoreFromExternalSecret:
          name: ldap-truststore
          key: cacerts
        trustCertificatesFromExternalSecret:
          name: ldap-truststore
          key: certificates.pem
```

Only the blocks you set are used, so a secret holding just a trust store or just the certificates
works too, and the two may live in different secrets. See
[Configuring a Java trust store](#configuring-a-java-trust-store) for why an external secret is
preferable.

Alternatively the `trustStore` and `trustCertificates` values can be set from the command line, as
they typically contain binary data:

```shell
helm install \
--set-file stackstate.authentication.ldap.ssl.trustStore=./ldap-cacerts \
--set-file stackstate.authentication.ldap.ssl.trustCertificates=./ldap-certificate.pem \
--values ldap_values.yaml \
... \
stackstate/stackstate
```

### Configuring a Java trust store

To make SUSE Observability trust certificates that are not in the default Java trust store, provide
your own trust store (`cacerts`) file. Create a secret for it and refer to that secret from the
values:

```shell
kubectl create secret generic java-truststore \
--from-file=java-cacerts=./cacerts \
--from-literal=java-cacerts-password=changeit
```

```yaml
stackstate:
  java:
    trustStoreFromExternalSecret:
      name: java-truststore
      key: java-cacerts
      passwordKey: java-cacerts-password
```

Use an external secret rather than the `stackstate.java.trustStore` and
`stackstate.java.trustStoreBase64Encoded` values. Helm stores every value you supply in the release
secret, and that secret is subject to the 1MB size limit of the underlying etcd object. A trust store
of a few hundred KB is counted twice against that limit (once for the value, once for the secret
rendered from it) and does not compress, which is enough on its own to make an upgrade fail.

The `stackstate.java.trustStore` and `stackstate.java.trustStoreBase64Encoded` values still take a
trust store inline and remain supported, but they are subject to the size limit described above:

```shell
helm install \
--set-file stackstate.java.trustStore=./cacerts \
--set stackstate.java.trustStorePassword=changeit \
... \
suse-observability/suse-observability
```

Note that when the trust store comes from an external secret, changing that secret does not restart
the pods that mount it. Rotating a trust store therefore needs a manual restart of the SUSE
Observability workloads.

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

### Upgrading backup job settings

Backup job rendering now applies scheduling, metadata and security settings consistently.
Before upgrading, check the following changes against your custom values and cluster policies:

* **Manual job container security.** Newly created manual Elasticsearch, StackGraph,
  configuration and VictoriaMetrics backup/restore Jobs inherit
  `common.container.securityContext`, including init containers. This includes non-root
  execution, dropped capabilities, seccomp and any read-only-root override. Existing Jobs
  are not modified. A component's `securityContext.enabled: false` disables only its pod
  security context; it does not disable container restrictions.

  Prefer images that support non-root execution. If a custom image requires root, use
  `backup.manualJobs.containerSecurityContext` to override the required fields for manual
  Jobs without changing scheduled backups, the settings-PVC initializer or application
  workloads:

  ```yaml
  backup:
    manualJobs:
      containerSecurityContext:
        runAsNonRoot: false
        runAsUser: 0
        runAsGroup: 0
  ```

  These overrides apply to all embedded manual Jobs and their init containers. Unspecified
  fields retain the common defaults; explicit `false` and `0` take precedence. Set
  `readOnlyRootFilesystem: false` here only if your custom image requires a writable root
  filesystem. Root execution also requires a namespace admission policy that permits it.
  Configuration restore/download Jobs have writable `/tmp` storage. The configuration
  upload Job mounts the existing settings PVC at `/settings-backup-data`, so uploads can
  be written with a read-only root filesystem and survive deletion of the upload Job.

* **Configuration pod security source.** Configuration backup CronJobs, manual Jobs and
  the legacy settings-PVC initializer now consistently use
  `backup.configuration.securityContext`. Copy configuration-specific overrides previously
  supplied through `backup.stackGraph.securityContext`, including `enabled`, `runAsUser`,
  `runAsGroup`, `runAsNonRoot` and `fsGroup`. Verify that the selected UID/GID can access
  existing backup files. The settings apply to future CronJob runs and newly created
  manual/initializer Jobs; existing Jobs and pods retain their original settings.

* **Explicit empty component overrides.** Empty strings in
  `stackstate.components.backup` and `stackstate.components.configurationBackup` node
  selectors, pod annotations and pod labels now override inherited values. Remove a
  component key to inherit its shared value instead of setting it to `""`. An empty
  node-selector value requires a node with that empty label value; otherwise the pod
  remains unschedulable.

* **Reserved component labels.** New backup pods use
  `app.kubernetes.io/component: backup`, except scheduled StackGraph v2 pods, which use
  `backup-v2`. Custom labels can no longer replace these values. Update NetworkPolicy,
  monitoring and operational selectors before upgrading; during the transition, account
  for existing pods that retain their previous labels.

* **Settings-PVC initializer resources.** The initializer always requests and limits
  `10m` CPU and `100Mi` memory and receives the common container security context, even
  when `stackstate.components.containerTools.resources` is `null`. Check namespace quotas,
  LimitRanges and admission policies before upgrading. The fixed initializer resources
  do not have a dedicated override.

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

## Adding StackPack images

Use `stackstate.stackpacks.extraImages` to make additional StackPacks available without replacing the built-in images. Each image must contain `.sts` packages in `/stackpacks` and provide `/bin/sh` and `cp`. The chart copies these packages into the platform after the built-in packages.

```yaml
stackstate:
  stackpacks:
    extraImages:
      - name: my-stackpack
        registry: ghcr.io
        repository: my-organization/my-stackpack
        tag: "0.0.1"
        pullPolicy: IfNotPresent
```

Use a unique `name` for each image, including the built-in images. Set `registry` explicitly to pull from a registry other than `global.imageRegistry`. Choose a published image tag that supports your Kubernetes node architecture.

Helm merges lists by replacement. Keep all additional images in the `extraImages` list when you update it. Leave `stackstate.stackpacks.images` unchanged to retain the built-in packages.

For a private registry, create an image pull secret in the installation namespace. Add its name to `global.imagePullSecrets` alongside the secrets your installation already uses:

```yaml
global:
  imagePullSecrets:
    - my-stackpack-registry
```

Pass the additional values file to your existing Helm installation or upgrade command. Adding an image makes its packages available in the StackPacks catalog. Install a StackPack through the UI or CLI, or configure automatic installation below.

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
