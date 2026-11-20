Run chart tests from the repository root:

```sh
KUBECONFIG=/dev/null OTEL_AGENT_BINARY=/absolute/path/to/collector \
  OTEL_AGENT_IMAGE=collector-image@sha256:IMAGE_DIGEST \
  go test -count=1 -timeout 30m ./stable/suse-observability-agent/test/...
```

Extract `OTEL_AGENT_BINARY` from the same image. Without these variables, binary
validation and container transport tests skip.

CollectionDeclaration CI reads the source commit from the chart's collector image
and compares the bundled CRD with that commit's canonical export. It also
validates disabled, enabled, and declarations-only configurations with the
binary extracted from that image. Missing source labels or a different CRD fail
the check. Updatecli fetches the CRD and image tag from the same collector release.
The check gates the agent chart's internal publication in CI.

The existing Kubeconform validation also renders the collection-enabled,
declarations-only and consumption-disabled profiles with restricted resource
permissions, in addition to the default chart values.

For local validation before those artifacts are published:

```sh
python3 stable/suse-observability-agent/test/collection_declaration_ci.py \
  --canonical-crd /absolute/path/to/collector/collectiondeclaration-crd.yaml \
  --agent-binary /absolute/path/to/candidate-agent
```

The local overrides validate the candidate; CI uses the committed source and
chart image without overrides.

CollectionDeclaration Helm lifecycle, two-Pod takeover and backend acceptance
are deferred to BEEST qualification. This chart suite covers rendering,
canonical CRD equality and generated collector configuration.

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
