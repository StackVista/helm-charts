# HBase monitors and probe-script ConfigMaps

Pre-change source: `225d8e44c4dca0910154c47aa20481a920cf8352`.
Extraction baseline: `c851a02ae1f2f7eb64c5538a148194ea36eee4b4`.
Both revisions are resolved and exported from repository Git history.

| Kind | Previous name | Canonical name |
| --- | --- | --- |
| ServiceMonitor | `<hbase-common.fullname.short>-hbase-master` | `suse-observability-hbase-master` |
| ServiceMonitor | `<hbase-common.fullname.short>-hbase-rs` | `suse-observability-hbase-rs` |
| ServiceMonitor | `<hbase-common.fullname.short>-hdfs-nn` | `suse-observability-hdfs-nn` |
| ServiceMonitor | `<hbase-common.fullname.short>-hdfs-snn` | `suse-observability-hdfs-snn` |
| ServiceMonitor | `<hbase-common.fullname.short>-hdfs-dn` | `suse-observability-hdfs-dn` |
| ServiceMonitor | `<hbase-common.fullname.short>-tephra` | `suse-observability-tephra` |
| ConfigMap | `<hbase-common.fullname.short>-hbase-rs-scripts` | `suse-observability-hbase-rs-scripts` |
| ConfigMap | `<hbase-common.fullname.short>-hdfs-dn-scripts` | `suse-observability-hdfs-dn-scripts` |

Monitors outside the release namespace retain their previous names. Their common
template supports shared monitoring namespaces, where fixed names could collide.
The naming resolver preserves direct-over-common namespace precedence.
Only the two probe-script volume references change in existing StatefulSets.
Scripts, probes, file modes, mounts, controllers, Services, immutable selectors,
claim templates, PVCs, credentials, RBAC and all other resources remain equal.
Default releases also replace the script maps and roll their existing consumers.

Normal tests unpack the locally built HBase dependency into a disposable
platform chart and restore only the eight old naming expressions. Complete
manifests are compared with an exact metadata-name/reference allowlist.
The Git comparison uses actual pre-change source. The independent-step test
copies extraction, adds only the nine new helper definitions, and wires the
eight declarations and two consumers. It preserves that older revision's inline
PDB naming and all intervening monitoring, configuration and RBAC identities.
No future or previous naming batch is required for consumers to work.

Every affected StatefulSet passes Helm 3 three-way patch application for upgrade,
repeat and rollback. Patches must preserve immutable fields, complete target
specifications, live annotations, resource versions and status. No historical
fixtures are rewritten.

To reproduce, export `local/` and `stable/suse-observability/` from each revision
with `git archive`. Recursively build only `file://` dependencies from children
to parents using Helm 3 `dependency build --skip-refresh`; build current
dependencies the same way. From the current repository root:

```sh
KUBECONFIG=/tmp/offline-kubeconfig.yaml \
HBASE_PERIPHERAL_BASELINE_CHART=/tmp/pre-change/stable/suse-observability \
HBASE_PERIPHERAL_EXTRACTION_CHART=/tmp/extraction/stable/suse-observability \
go test ./stable/suse-observability/test/... \
  -run '^TestHBasePeripheralNaming' -count=1
```

Use synthetic offline kubeconfig credentials with server `http://127.0.0.1:1`.
Nineteen scenarios cover install/upgrade, releases/namespaces, root/subchart
overrides and affixes, long names, already-canonical names, direct/common/shared
monitoring namespaces, Mono/Distributed modes, secondary NameNode, disabled
metrics/monitors, Argo hooks, sizing, storage, backups and custom metadata.
All checks are local; no environment or installed sensitive data is accessed.
Actual Pod recovery, Prometheus reconciliation and Helm/Argo CD lifecycle remain
integration checks. Pruning and interrupted-operation recovery follow the chart
README; no recovery-time guarantee is claimed.
