# Generated ConfigMap naming compatibility

The pre-change source is commit
`c0971c0ee6756ccf7a4f23cc8655577fa3079e19`.
It contains the earlier application/logging ConfigMap and exporter-monitoring
renames. No frozen manifests are regenerated from the implementation under test.

| Previous name | New name |
| --- | --- |
| `<parent-fullname>-router` | `suse-observability-router` |
| `<parent-fullname>-stackpacks-scripts` | `suse-observability-stackpacks-scripts` |
| `<release>-rbac-agent-url` | `suse-observability-rbac-agent-url` |
| `<release>-rbac-agent-cluster-name` | `suse-observability-rbac-agent-cluster-name` |
| `<operator-fullname>-kafkaup-config` | `suse-observability-kafkaup-config` |
| `<global-backup-fullname>-sts-hbase-backup` | `suse-observability-hbase-backup` |

The previous fullname expressions retain their original truncation and overrides.
The normal compatibility test restores exactly those six expressions in a temporary
chart. The Git-baseline test reads an independently exported chart.
Each compares the complete platform for install and upgrade rendering, including
non-default releases, namespaces, overrides, affixes, long names, sizing,
Mono/Distributed HBase, split/monolithic application, router modes, Argo CD,
component/backup conditions, external Secrets and custom metadata.

Only those ConfigMap identities, Pod references and their existing Deployment
checksums may change. StatefulSet claim templates, PVCs, credentials, Services,
selectors, accounts, RBAC, hook scripts, dynamic mode state and immutable Jobs
remain unchanged. Local Kubernetes strategic merge patches cover upgrade, repeat
and rollback for Deployments and StatefulSets while preserving synthetic live
metadata. No API server is started or contacted.

To reproduce independently, export the revision with `git archive`, then build
its `file://` dependency graph from the leaves upward with Helm 3
`dependency build --skip-refresh`. Build the current dependency graph the same way.
Check that the source revision resolves before exporting:

```sh
git cat-file -e c0971c0ee6756ccf7a4f23cc8655577fa3079e19^{commit}
```

Use an offline kubeconfig (its server can be `http://127.0.0.1:1`) and run from
the current repository root:

```sh
PERIPHERAL_CONFIGMAP_BASELINE_CHART=/tmp/pre-change/stable/suse-observability \
KUBECONFIG=/tmp/offline-kubeconfig.yaml \
go test ./stable/suse-observability/test/... \
  -run '^TestPeripheralConfigMapNaming' -count=1
```

This verifies manifest and patch compatibility. It does not prove application
readiness, runtime coordination, pruning or a recovery-time bound.
Preserve router mode/state ConfigMaps and backup maps referenced by existing Jobs.
Prune obsolete renamed maps only after Pods and controller templates have stopped
referencing them, as described in the chart README.
