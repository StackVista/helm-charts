# Main ClusterRoleBinding naming compatibility

The snapshots were generated before this rename from
`91d4b4b016a40d233d05f6285d2dc932fb094bd9` with Helm 3.19.0. The revision's local
dependency graph was freshly rebuilt using only `file://` repositories.

Only two main-chart ClusterRoleBinding metadata names change. Their existing
helpers in `_names.tpl` now ignore release names and fullname overrides/affixes.

| Purpose | Previous name | Canonical name outside the product namespace |
| --- | --- | --- |
| Authentication | `<common.fullname.cluster.unique>-authentication` | `suse-observability-<namespace>-authentication` |
| Authorization | `<common.fullname.cluster.unique>-authorization` | `suse-observability-<namespace>-authorization` |

In namespace `suse-observability`, the names are
`suse-observability-authentication` and `suse-observability-authorization`,
preserving the default product release's existing identities.
Other namespaces are included in full to distinguish cluster-scoped bindings.
Tests cover two maximum-length namespaces differing only in their last
character; truncating the namespace would make their binding names collide.
This check concerns the renamed bindings; retained ClusterRole naming rules
are unchanged.

`roleRef`, subjects and subject namespaces, labels, annotations, API version and
creation conditions remain unchanged. The authorization ClusterRole retains
its name and permissions so customer-managed bindings can still reference it.
Authentication retains `system:auth-delegator`. All other RBAC resources,
ServiceAccounts, application authorization mappings, configuration/checksums,
Pod specifications, stored credential identities, PVCs, StatefulSets and Service
DNS must stay unchanged.

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
allow only the two binding names to change. Full-platform comparisons cover
default/custom releases, product/custom namespaces, fullname settings,
long names, already-canonical bindings, split/monolithic services, HA/non-HA,
Argo CD, disabled cluster RBAC and custom metadata. A separate check confirms
that the same release name produces distinct binding names in five namespaces,
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
