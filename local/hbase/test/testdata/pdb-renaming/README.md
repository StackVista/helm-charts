# HBase PDB naming compatibility

The snapshots were generated before this change from revision
`3ea779a7170a4603240376c7ba2c18c344db53df`, an ancestor of the rename, using
Helm 3.19.0 and freshly rebuilt local dependencies.

Only the following PDB metadata names change. The names are defined in
`templates/_names.tpl`; release/fullname settings no longer affect them.

| Previous name | Canonical name |
| --- | --- |
| `<hbase-fullname>-hbase-master` | `suse-observability-hbase-master` |
| `<hbase-fullname>-hbase-rs` | `suse-observability-hbase-rs` |
| `<hbase-fullname>-hdfs-nn` | `suse-observability-hdfs-nn` |
| `<hbase-fullname>-hdfs-snn` | `suse-observability-hdfs-snn` |
| `<hbase-fullname>-hdfs-dn` | `suse-observability-hdfs-dn` |
| `<hbase-fullname>-tephra` | `suse-observability-tephra` |

Selectors, `maxUnavailable: 1`, labels, annotations and creation conditions stay
unchanged. StatefulSet/PVC identities, governing Services and claim templates,
Pod specifications and checksums, configuration, credentials, RBAC, endpoints
and monitors must remain identical.

## Frozen snapshots

| Fixture | Mode | Secondary NameNode enabled | PDBs |
| --- | --- | --- | --- |
| `mono.json` | Mono | true (ignored in this mode) | 1 |
| `distributed.json` | Distributed | false | 5 |
| `distributed-snn.json` | Distributed | true | 6 |

Inputs are release `nightly`, namespace `observability`,
`zookeeper.externalServers=test-zookeeper`, the mode/secondary-node settings
above, and API capability `policy/v1/PodDisruptionBudget`. Each fixture contains
complete objects. Chart version and appVersion are pinned to `0.0.0` and
`naming-fixture` in a temporary copy so unrelated version changes do not
invalidate the baseline.

The original archives were:

- `common-0.4.28.tgz`, SHA-256
  `4365ebc9663fc85f56472be9a12b34d104a29eec682c7f1ada6e5a91d685584c`
- `suse-observability-sizing-0.1.16.tgz`, SHA-256
  `56f626a64dc4ecd3c1c564f29b14fa7d3d735c6956e5e23f894a2c686ffa8fbf`

Rebuilding the same source is sufficient; packaging timestamps can change
archive hashes without changing rendered objects.

Export the documented Git revision, build its HBase local dependencies, and
independently verify the baseline with:

```shell
HBASE_PDB_BASELINE_CHART=/path/to/export/local/hbase \
  go test ./local/hbase/test/... \
  -run '^TestHBasePDBNamingBaselineReproduction$' -count=1
```

This check compares old source to the snapshots without rewriting them.
Additional whole-chart comparisons cover namespaces, fullname settings,
already-canonical names, disabled metrics, shared monitor namespaces, custom
storage/credentials/metadata, both PDB API versions and all three inventories.
StatefulSet three-way patches are compared with an ordinary Helm 3 upgrade
using synthetic live status, annotations and scaling drift; claim templates
and governing Services must remain unchanged.

These checks render locally and exercise patch generation without a cluster.
Real evictions, controller status reconciliation, repeated upgrades and rollback
remain integration checks. Obsolete PDBs must be pruned; overlapping budgets
that select the same Pods can block evictions with HTTP 500 until cleanup.
No workload or data migration is introduced.

ServiceMonitor names remain unchanged: the common monitor template supports
placement in a shared monitoring namespace, so changing those names requires
a separate collision review.
