# Application ConfigMap and exporter monitoring names

The pre-change source is `6871f337366782b3757de14d244a928bef06ec96`.
It was exported from Git and its complete local dependency graph rebuilt with
Helm 3.19.0, using only `file://` sources and `helm dependency build --skip-refresh`.
Baseline verification renders that source independently; it never rewrites
historical fixtures.

## Reviewed inventory

For each component below, only its generated application and logging ConfigMap
names change. The legacy prefix is the exact output of `common.fullname.short`,
including release names and local/global fullname settings.

| Component | Application ConfigMap | Logging ConfigMap |
| --- | --- | --- |
| API | `suse-observability-api` | `suse-observability-api-log` |
| Checks | `suse-observability-checks` | `suse-observability-checks-log` |
| Notification | `suse-observability-notification` | `suse-observability-notification-log` |
| Health sync | `suse-observability-health-sync` | `suse-observability-health-sync-log` |
| Authorization sync | `suse-observability-authorization-sync` | `suse-observability-authorization-sync-log` |
| State | `suse-observability-state` | `suse-observability-state-log` |
| Sync | `suse-observability-sync` | `suse-observability-sync-log` |
| Slicing | `suse-observability-slicing` | `suse-observability-slicing-log` |
| Server | `suse-observability-server` | `suse-observability-server-log` |
| Receiver | `suse-observability-receiver` | `suse-observability-receiver-log` |
| Correlate | `suse-observability-correlate` | `suse-observability-correlate-log` |
| Initializer | `suse-observability-initializer` | `suse-observability-initializer-log` |
| E2ES | `suse-observability-e2es` | `suse-observability-e2es-log` |

Creation conditions are unchanged, so each rendering includes only the objects
needed by its selected features and sizing. Split receiver and correlate workers
continue sharing their component ConfigMaps.

The exporter's optional `ServiceMonitor`, `PodMonitor` and `PrometheusRule` each
use `suse-observability-prometheus-elasticsearch-exporter`. Their legacy names
come from `elasticsearch-exporter.fullname`. Monitoring namespaces, selectors,
labels, scrape configuration, rule expressions and rule-group names are retained.
Different resource kinds may use the same name. The existing mutually exclusive
ServiceMonitor/PodMonitor choice stays unchanged.
Installations in different namespaces use the same canonical names. Explicit
monitoring namespace settings remain supported for one installation per target
monitoring namespace; names are not qualified with the installation namespace.

## Compatibility checks

`TestApplicationConfigMapNamingCompatibility` restores only these 29 helper
bodies in a disposable copy of the current chart. It compares complete platform
manifests, allowing only the reviewed metadata names, ConfigMap volume names and
the existing application-configuration checksum annotations. Configuration data
and all other Pod fields are compared unchanged.

The independent check uses the actual pre-change Git revision:

```shell
KUBECONFIG=/path/to/offline-kubeconfig \
CONFIGMAP_MONITOR_BASELINE_CHART=/path/to/export/stable/suse-observability \
  go test ./stable/suse-observability/test/... \
  -run '^TestApplicationConfigMapNamingGitBaselineCompatibility$' -count=1
```

Export both `local/` and `stable/suse-observability/` from the source revision and
build dependencies recursively, from children to parents. Use matching Helm
versions and an offline kubeconfig when running the comparison. Neither check
deploys a chart, contacts Kubernetes, or accesses installed credentials.

Twenty scenarios cover install/upgrade rendering, default and custom releases,
namespaces, fullname overrides/affixes, long names, split/monolithic servers,
split workers, HA/non-HA sizing, Argo CD, external Secret references, disabled
authorization, exporter overrides, certificate/account settings, monitor
namespaces, custom metadata and disabled monitoring. Helm 3 three-way patches
are applied locally for upgrade, repeated upgrade and rollback, preserving
synthetic live metadata and checking the complete resulting Deployments.
Separate exporter tests render the same release in five namespaces and confirm
distinct namespace/name identities, preserved target namespaces, and valid names.
Two installation namespaces share their first 62 characters. Additional checks
verify that an explicit monitoring namespace preserves the selected targets.

Default product naming retains ConfigMap names and checksums. Non-default naming
updates existing Deployments' volume references and triggers normal Pod rollouts.
Credential and environment Secrets, PVCs, StatefulSets, Service DNS, controller
identities/strategies, RBAC and external resource choices stay unchanged.
Router state and subcharts with fixed default identities are retained.

Actual readiness, mount behavior and Helm/Argo CD replacement, pruning,
destination ownership conflicts and rollback remain integration checks.
Offline renders and patches do not establish a recovery-time guarantee.
Follow the main README's pruning and stale-object recovery guidance; old Pods
can still need their previous ConfigMaps until their rollout completes.
