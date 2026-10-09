# Bundled collector relay ConfigMap renaming

Pre-change source: `ecb7ef9770d3685df9bc033ece48ce5577c35290`.
Extraction baseline: `c851a02ae1f2f7eb64c5538a148194ea36eee4b4`.
Both revisions are resolved and exported from repository Git history.

| Mode | Previous name | New name |
| --- | --- | --- |
| Deployment | `<collector-fullname>` | `suse-observability-otel-collector-config` |
| DaemonSet | `<collector-fullname>-agent` | `suse-observability-otel-collector-agent` |
| StatefulSet | `<collector-fullname>-statefulset` | `suse-observability-otel-collector-statefulset` |

External ConfigMap choices and disabled generation are preserved.
Standalone subcharts are unchanged. Generated names are unconditional.
The platform uses the collector's default RollingUpdate strategy.
All declarations, volume resolution and checksums already consume the three
extracted helper interfaces.

The platform endpoint map retains `suse-observability-otel-collector`, its data
and every environment reference. The Deployment relay name is distinct:
the old default-fullname Deployment mode emitted two ConfigMaps with the same
effective namespace/name. It is not a healthy install/rollback baseline.
Rollback rendering into that older configuration restores its existing
collision; rollback of previously working configurations restores working names
and references. Default StatefulSet and DaemonSet names remain unchanged.

Normal tests remove only the three new parent definitions in a disposable chart.
The Git comparison uses the actual pre-change source. The independent-step
check adds only these definitions to extraction, without intervening renames.
Complete manifests allow changes only to generated relay metadata names,
their exact volume references and existing checksum annotations. Relay contents,
endpoint data/environment references, external choices, controllers, selectors,
labels, storage, checkpoint paths, accounts, RBAC and every other field stay equal.

Tests retain all rendered documents and resolve omitted namespaces when checking
identity uniqueness. They rename only the relay document, preserving the
endpoint document even when older manifests give them the same effective name.
Helm 3 three-way upgrade, repeat and rollback patches preserve complete controller
objects, current metadata and immutable StatefulSet/selector fields.
No historical fixtures are regenerated.

To reproduce, export `local/` and `stable/suse-observability/` from both revisions
with `git archive`. Build their `file://` graphs from children to parents using
Helm 3 `dependency build --skip-refresh`; build current dependencies likewise.
From the current repository root:

```sh
KUBECONFIG=/tmp/offline-kubeconfig.yaml \
COLLECTOR_CONFIGMAP_BASELINE_CHART=/tmp/pre-change/stable/suse-observability \
COLLECTOR_CONFIGMAP_EXTRACTION_CHART=/tmp/extraction/stable/suse-observability \
go test ./stable/suse-observability/test/... \
  -run '^TestCollectorConfigMapNaming' -count=1
```

Use a synthetic kubeconfig with server `http://127.0.0.1:1`.
Twenty-two scenarios cover install/upgrade, modes, default/custom releases and
names, naming settings, namespaces, external/disabled configuration,
custom relay contents, persistent claims, checkpoints, autoscaling, explicit
checksum templates, sizing, Argo hooks and Recreate.
All verification is local, with no installed sensitive data or cluster access.
Runtime readiness, forwarding, pruning and actual Helm/Argo CD lifecycle remain
integration checks. Recovery follows the chart README; no time bound is claimed.
