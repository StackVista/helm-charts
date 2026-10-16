# Pod-reader RoleBinding naming compatibility

The fixtures were generated before this rename from revision
`c6030933f0d8971cb1b8c4ea5079d487089f0e84` with Helm 3.19.0. The revision's entire
local dependency graph was rebuilt using only `file://` repositories.

The Pod-reader Role and RoleBinding retain the matching legacy name
`<common.fullname.short>-get-pods`, including release and fullname settings.
The existing dedicated helpers remain independently wired to their declarations.
A fixed binding name would require patching immutable `roleRef` when a fullname
override changes the referenced Role. Retaining both identities makes Helm
replace both objects instead, on upgrade and rollback, without migration machinery.
Customer-managed bindings can continue referencing the legacy Role.

Permissions, subjects, namespace scope, labels, annotations, workload and storage
identities remain unchanged by this fix.

## Frozen snapshots

Each fixture contains the complete old Role and binding. Only `helm.sh/chart`
and `app.kubernetes.io/version` labels are removed so automatic version updates
do not invalidate the baseline.

Inputs are release `nightly`, namespace `observability`, `test/values/full.yaml`,
API capabilities `policy/v1/PodDisruptionBudget` and `monitoring.coreos.com/v1`,
and `stackstate.features.server.split` as specified below.

| Fixture | Server split | Binding subjects |
| --- | --- | --- |
| `split.json` | true | Nine existing component ServiceAccounts |
| `monolithic.json` | false | Existing server ServiceAccount |

Keep these snapshots unchanged; do not regenerate them from the renamed chart.
To independently verify the baseline, export the documented Git revision,
rebuild its local dependencies, and run:

```shell
POD_READER_BASELINE_CHART=/path/to/export/stable/suse-observability \
  go test ./stable/suse-observability/test/... \
  -run '^TestPodReaderNamingBaselineReproduction$' -count=1
```

This compares old source to the fixtures without modifying them. Ordinary
tests compare new output with the snapshots, with no binding-name allowance. Whole-platform comparisons exercise default/custom releases, two
namespaces, fullname settings, long names, already-canonical names,
split/monolithic services, HA/non-HA sizing, Argo CD, disabled authorization
and cluster RBAC, custom metadata and split workers.

For bindings retaining their names, Helm 3 three-way patch checks preserve
immutable `roleRef` and produce no changes with synthetic live annotations.
Override changes replace the binding alongside its Role. Existing helper
tests independently change Role and binding names in temporary chart copies
to verify producer/consumer wiring; no test chart is installed.

These are offline render and patch-generation checks. Real authorization-cache
reconciliation, effective API access, Helm/Argo CD replacement, destination
ownership conflicts, repeated upgrades and rollback remain integration checks.
No availability or recovery-time guarantee is asserted.

`TestBindingNamingOverrideUpgradeRollback` renders install, changed-override
upgrade, repeated upgrade and rollback, in split/monolithic modes and two
namespaces, with and without affixes. It checks replacement inventories and
Helm three-way patches for every retained binding: `roleRef` must never change.
These tests use synthetic offline objects; real lifecycle reconciliation remains
an integration check.
