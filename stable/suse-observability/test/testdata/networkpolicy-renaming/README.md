# Main NetworkPolicy naming compatibility

These snapshots were generated from the pre-rename revision
`a5abc996ac5ed4b61dff408a8a93ef4ba6b617c8` with Helm 3.19.0. The revision's
complete local dependency graph was freshly rebuilt using only `file://`
repositories, without downloading charts.

Only the main-chart NetworkPolicy metadata name changes:

| Previous name | Canonical name |
| --- | --- |
| `<common.fullname>` | `suse-observability-networkpolicy` |

The old producer used `common.fullname`, which repeats the chart name even when
the release is named `suse-observability`; its default name was
`suse-observability-suse-observability`. Other release/fullname settings followed
that helper's lowercase, affix, common-value merging and 54-character truncation
rules. The new identity is defined in `templates/_names.tpl` and ignores those
naming settings.

The API version, labels, annotations, selectors, policy types, ingress/egress
rules and enabled setting are preserved. All other platform resources and their
specifications, including Pod configuration/checksums, stored credential names,
PVCs, StatefulSets, governing Services and subchart policies, remain unchanged.

## Frozen snapshots

Both fixtures use release `nightly`, namespace `observability`,
`test/values/full.yaml`, and `networkPolicy.enabled=true`. Only
`helm.sh/chart` and `app.kubernetes.io/version` labels are removed to allow
unrelated automatic chart/application version updates.

| Fixture | Additional inputs |
| --- | --- |
| `default.json` | Default rules and metadata |
| `custom.json` | `test/values/networkpolicy-custom.yaml` |

The custom case includes Pod match expressions, combined namespace/Pod peers,
IP blocks and exclusions, named/numeric ports, TCP/UDP, ingress and egress,
custom labels and annotations. Freeze these objects before changing the
implementation; do not regenerate them from the renamed chart.

To independently reproduce the baseline, export the documented Git revision
into a temporary directory and rebuild its local dependencies. Run:

```shell
NETWORK_POLICY_BASELINE_CHART=/path/to/export/stable/suse-observability \
  go test ./stable/suse-observability/test/... \
  -run '^TestNetworkPolicyNamingBaselineReproduction$' -count=1
```

This compares the old source against the snapshots without modifying them.
Ordinary tests compare the snapshots against the new policy, allowing only the
metadata name to change. Whole-platform comparisons cover release names,
namespaces, fullname overrides/affixes, common-value overrides, long names,
already-canonical names, custom rules, HA/non-HA sizing, split workers, Argo CD,
disabled policies and an independently configured collector policy.
Install and upgrade rendering must produce identical workload specifications;
the rename does not require a workload rollout or data migration.

The dedicated-helper check changes the helper only in a temporary chart copy
to verify that the real declaration consumes it. No fixture chart is installed.

These checks use offline rendering and manifest comparisons. Network-plugin
enforcement, actual Helm/Argo CD replacement, repeated upgrade, destination
ownership conflicts and rollback remain integration checks. No connectivity
or recovery-time guarantee is asserted.

Old and new standard Kubernetes NetworkPolicies with identical rules allow the
same traffic because policies are additive. Obsolete policies must still be
pruned: retained rules can preserve access when later updates tighten the
intended policy. Follow the chart README's ownership-based cleanup guidance.
