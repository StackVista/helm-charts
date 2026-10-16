# Main ClusterRoleBinding naming compatibility

The snapshots were generated before this rename from
`91d4b4b016a40d233d05f6285d2dc932fb094bd9` with Helm 3.19.0. The revision's local
dependency graph was freshly rebuilt using only `file://` repositories.

Only the authentication ClusterRoleBinding uses canonical naming:
`suse-observability-authentication` in the product namespace and
`suse-observability-<namespace>-authentication` elsewhere. It includes the full
namespace and ignores release names and fullname overrides/affixes. Its immutable
Role reference is always `system:auth-delegator`.

Authorization retains `<common.fullname.cluster.unique>-authorization` alongside
its legacy ClusterRole. Changing fullname settings replaces both identities,
avoiding an immutable `roleRef` patch. The legacy namespace/truncation behavior
remains unchanged. The canonical namespace-distinction test applies only to
authentication; it does not claim new collision guarantees for retained Roles.

Role references, subjects, labels, annotations and creation conditions remain
unchanged for a given set of naming values. Customer-managed Role references,
ServiceAccounts, application configuration, Pods, storage and credentials are
preserved by this fix.

## Frozen snapshots

Each fixture contains both old bindings and the old authorization ClusterRole.
Only `helm.sh/chart` and `app.kubernetes.io/version` labels are removed so
automatic version updates do not invalidate the baseline.

Inputs are release `nightly`, namespace `observability`, `test/values/full.yaml`,
API capabilities `policy/v1/PodDisruptionBudget` and `monitoring.coreos.com/v1`,
and `stackstate.features.server.split` as specified below.

| Fixture | Server split | Existing subject |
| --- | --- | --- |
| `split.json` | true | API ServiceAccount in `observability` |
| `monolithic.json` | false | Server ServiceAccount in `observability` |

Keep the snapshots unchanged; do not regenerate them from the renamed chart.
Export the documented revision, build its local dependencies, and verify the
baseline independently with:

```shell
CLUSTER_BINDING_BASELINE_CHART=/path/to/export/stable/suse-observability \
  go test ./stable/suse-observability/test/... \
  -run '^TestClusterBindingNamingBaselineReproduction$' -count=1
```

This compares old source to the snapshots without rewriting them. Ordinary tests
allow only the authentication binding name to change. Full-platform comparisons cover
default/custom releases, product/custom namespaces, fullname settings,
long names, already-canonical bindings, split/monolithic services, HA/non-HA,
Argo CD, disabled cluster RBAC and custom metadata. A separate check confirms
that the same release name produces distinct authentication binding names in five namespaces,
including maximum-length namespaces sharing a prefix.

Helm 3 patch checks produce no changes for bindings retaining their names,
using synthetic live annotations and resource versions. Both replaced and
retained bindings keep their immutable Role references and subjects.
Existing helper tests independently change Role and binding helpers in temporary
chart copies to verify producer/consumer wiring; no fixture chart is installed.

These checks use offline rendering and patch generation. Actual token review,
subject access review, authorization-cache reconciliation, Helm/Argo CD
replacement, destination ownership conflicts, repeated upgrades and rollback
remain integration checks. No availability or recovery-time guarantee is made.

Prune obsolete chart-owned bindings because stale subjects can retain access
after later updates. Follow the README's cluster-wide, ownership-based cleanup
guidance; preserve ClusterRoles, ServiceAccounts and other installations'
bindings. No Pod rollout or data/token migration is introduced.

`TestBindingNamingOverrideUpgradeRollback` checks override-changing upgrades,
repeated upgrades and rollback using offline rendering and Helm three-way patch
generation. Authorization and Pod-reader bindings must be replaced when their
Role names change; retained binding names must never patch `roleRef`.
Authentication stays canonical and can update subjects without changing its Role.
