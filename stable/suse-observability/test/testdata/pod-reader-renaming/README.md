# Pod-reader RoleBinding naming compatibility

The fixtures were generated before this rename from revision
`c6030933f0d8971cb1b8c4ea5079d487089f0e84` with Helm 3.19.0. The revision's entire
local dependency graph was rebuilt using only `file://` repositories.

Only the chart-owned Pod-reader RoleBinding's metadata name changes:

| Kind | Previous name | Canonical name |
| --- | --- | --- |
| RoleBinding | `<common.fullname.short>-get-pods` | `suse-observability-get-pods` |

The target is defined in the existing `stackstate.getPods.rolebinding.fullname`
helper in `_names.tpl`. It ignores release names and fullname overrides/affixes.
The default product release already has the canonical binding identity.

The Pod-reader Role keeps its existing name so customer-managed RoleBindings
can still reference it. The chart binding's `roleRef`, subjects, namespace scope,
labels, annotations and API version must remain unchanged. So must the Role's
`get`/`list` Pod permissions, ServiceAccounts, workload/Pod specifications and
checksums, application authorization mappings, stored credential identities,
PVCs, StatefulSets and Service DNS.

The RBAC agent Role and binding retain their identities: their Role name is also
used by the application's authorization subject mapping. Customer-facing
instance Roles/bindings and cluster-scoped RBAC are outside this rename.

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
tests compare new output with the snapshots, allowing only the binding name
to change. Whole-platform comparisons exercise default/custom releases, two
namespaces, fullname settings, long names, already-canonical names,
split/monolithic services, HA/non-HA sizing, Argo CD, disabled authorization
and cluster RBAC, custom metadata and split workers.

For bindings retaining their names, Helm 3 three-way patch checks preserve
immutable `roleRef` and produce no changes with synthetic live annotations.
Renamed bindings retain the same Role reference and subjects. Existing helper
tests independently change Role and binding names in temporary chart copies
to verify producer/consumer wiring; no test chart is installed.

These are offline render and patch-generation checks. Real authorization-cache
reconciliation, effective API access, Helm/Argo CD replacement, destination
ownership conflicts, repeated upgrades and rollback remain integration checks.
No availability or recovery-time guarantee is asserted.

Old and new bindings grant the same permissions during overlap. Prune obsolete
chart-owned bindings, since stale subjects can retain access after future
changes. Follow the README's ownership-based cleanup guidance and preserve the
Role, ServiceAccounts and customer-managed bindings.
