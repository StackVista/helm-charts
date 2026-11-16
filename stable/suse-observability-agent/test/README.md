Run chart tests from the repository root:

```sh
KUBECONFIG=/dev/null OTEL_AGENT_BINARY=/absolute/path/to/collector \
  OTEL_AGENT_IMAGE=collector-image@sha256:IMAGE_DIGEST \
  go test -count=1 -timeout 30m ./stable/suse-observability-agent/test/...
```

Extract `OTEL_AGENT_BINARY` from the same image. Without these variables, binary
validation and container transport tests skip.

ResourceCollection CI always checks the chart's pinned canonical collector
export and validates disabled, enabled, and declarations-only configurations
with the agent binary extracted from the chart's image. It fails if provenance
still names a working-tree export or the matching agent has not been adopted.
The check gates the agent chart's internal publication in CI.

The existing Kubeconform validation also renders the collection-enabled,
declarations-only and consumption-disabled profiles with restricted resource
permissions, in addition to the default chart values.

For local validation before those artifacts are published:

```sh
python3 stable/suse-observability-agent/test/resource_collection_ci.py \
  --canonical-crd /absolute/path/to/collector/resourcecollection-crd.yaml \
  --agent-binary /absolute/path/to/candidate-agent
```

The local overrides validate the candidate; CI uses the committed source and
chart image without overrides.

Lifecycle CI runs the collector's matching pinned harness on owned Kubernetes
clusters for native collection and topology compatibility. Both profiles check
declaration edits without rollout, two-Pod takeover and access recovery. The
native profile also checks upgrade, disable/re-enable, retained definitions and
reinstall. The topology profile checks native snapshots, deletion and sanitization
while custom collection is pending or blocked.

Both profiles use the immutable image digest that passed configuration validation.
The matrix gates internal publication in CI and uploads synthetic evidence.
Cluster tools use the collector harness's checksum-verified versions and
specialized Rancher K3s node image.

Run a local unpublished lifecycle candidate with:

```sh
python3 stable/suse-observability-agent/test/resource_collection_ci.py \
  --canonical-crd /absolute/path/to/collector/resourcecollection-crd.yaml \
  --collector-source /absolute/path/to/collector \
  --agent-image candidate-agent:local \
  --lifecycle-profile native \
  --node-image 'rancher/k3s@sha256:QUALIFIED_NODE_DIGEST' \
  --kubernetes-version EXPECTED_MINOR \
  --artifacts /tmp/resourcecollection-lifecycle-evidence
```

Use `--lifecycle-profile topology` for the compatibility profile and a separate
empty artifact directory. The binary comes from that image; an optional local
`--agent-binary` must match it. CI checks the harness revision against schema
provenance. These profiles do not establish the support floor, older chart/image
rollback qualification, other architectures or backend persistence/UI.

The container tests require a locally available image and Linux Docker host networking. They use synthetic log
files and ephemeral TLS servers, replacing Kubernetes metadata lookup. They
exercise discovery, Promtail-compatible export and native HTTP/gRPC with inline
PEM, multiple ConfigMap filenames, and certificates placed separately in the
system bundle and custom directory. Original system roots remain present.
Each route also exercises an explicit CONNECT proxy. The gRPC proxy fixture uses a synthetic
non-loopback destination that the proxy maps to its local TLS server, retaining
the chart's `NO_PROXY` exclusions.

Startup tests preserve the fixed A direct-export graph, exercise explicit B discovery and
SIGTERM, and reject queues on direct Promtail and both routed exporters. Use an image containing the
`stslogsagent` extension with discovery support; the A-only image cannot run B
configuration.

Containers run as the invoking user with SELinux labeling disabled for temporary
fixtures. These tests establish transport trust, not Kubernetes permissions,
DaemonSet rollout behavior or platform ingestion acceptance.
