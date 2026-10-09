# Anomaly-detection PDB and ServiceMonitor compatibility

The two frozen snapshots were generated before changing the naming helpers from
revision `9ed0a1cb4837895c16e7e1c1b23ed2dfa8e69c13`, an ancestor of this change,
using Helm 3.19.0. Both contain complete PDB and manager ServiceMonitor objects.

Only these two metadata names are allowed to change:

| Kind | Previous name | Canonical name |
| --- | --- | --- |
| PodDisruptionBudget | `<legacy-fullname>-anomaly-detection` | `suse-observability-anomaly-detection` |
| ServiceMonitor | `<legacy-fullname>-spotlight-manager` | `suse-observability-spotlight-manager` |

Workloads, their immutable selectors and Pod specifications, Service DNS,
configuration checksums, accounts, RBAC, synthetic credentials, image pull
references and the artifact PVC must remain identical. Both renamed objects
belong to the release namespace. Their labels and selection behavior are
preserved, including the monitor's existing namespace-based instance label.
The existing PDB selects component `anomaly-detection`; correcting its mismatch
with the default manager/worker Pod components is separate work.

Snapshot inputs are release `nightly`, namespace `observability`,
`stackstate.instance=https://analysis.example`, `global.receiverApiKey=test-key`,
`metrics.serviceMonitor.enabled=true`, and API capability
`policy/v1/PodDisruptionBudget`. `custom-budget.json` additionally sets
`pdb.maxUnavailable=25%`, `commonLabels.fixture-scope=preserved`, and
`poddisruptionbudget.annotations.fixture-note=preserved`.

Chart version and appVersion are pinned to `0.0.0` and `naming-fixture` in a
temporary copy so unrelated version updates do not invalidate the baseline.
The original dependency archive is `common-0.4.28.tgz`, SHA-256
`e33afc39a77537dc960b45fea93c9a28124d80b89c9d441732a25380b5d4c9b6`.
Rebuilding from the same source is sufficient; archive timestamps can change
that hash without changing the rendered objects.

To independently reproduce the snapshots, export the documented revision and
build its local anomaly-detection dependencies. Then run:

```shell
ANOMALY_PERIPHERAL_BASELINE_CHART=/path/to/export/local/anomaly-detection \
  go test ./local/anomaly-detection/test/... \
  -run '^TestAnomalyPeripheralNamingBaselineReproduction$' -count=1
```

This compares the old source to the frozen objects without rewriting them.
Other tests compare all rendered resources for release/namespace variations,
fullname overrides and affixes, long names, already-canonical names, disabled
monitoring, custom limits, credentials, storage and external references.

These are offline rendering checks. No existing cluster or application data is
accessed. Actual Helm upgrade/rollback, eviction and Prometheus reconciliation
remain integration checks. Replacement requires deletion of obsolete objects:
Argo CD must prune them, and failed operations can require ownership-based
cleanup of stale PDBs and monitors.
