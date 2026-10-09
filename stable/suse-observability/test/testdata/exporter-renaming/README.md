# Elasticsearch exporter Deployment naming

The frozen Deployment snapshots were rendered from `27d40b33e`, before the
Deployment rename, with its packaged dependencies and Helm 3.19.0.
They use release `nightly`, namespace `observability` and `values/full.yaml`.
The exporter image tag is pinned to `naming-fixture` to avoid unrelated image
updates changing the baseline. The chart version label is excluded.

`default.json` captures the default Pod specification. `tls.json` adds a
chart-managed ServiceAccount, supplied certificate contents, an external TLS
Secret mount and external environment Secrets. Its settings are recorded in
`exporterNamingTLSValues`.

| Kind | Previous name | New name |
| --- | --- | --- |
| Deployment | `nightly-prometheus-elasticsearch-exporter` or the exporter fullname override | `suse-observability-prometheus-elasticsearch-exporter` |

Only the Deployment metadata name changes. Keep its selector and Pod labels,
Service DNS and selectors, accounts, RBAC, certificate Secret names and contents,
environment references and mount specifications. The default exporter reads
Elasticsearch metrics and has no PVC, exclusive writer or background operator.
Temporary exporter overlap or missing scrapes during replacement is accepted.
Custom containers and volumes keep their specifications; this test does not
establish lifecycle safety for arbitrary customer-provided code.

The main Ingress identity is retained in this step because a certificate controller
can use the Ingress UID as the owner of a Certificate. Deleting that owner can
garbage-collect the Certificate and, where certificate Secret owner references
are enabled, its TLS Secret. Recreating the Ingress could then reissue credentials.
The main HTTPRoute already has a canonical name, so it needs no rename.

Do not regenerate these snapshots from the implementation being tested.
