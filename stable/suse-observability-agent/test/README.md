Run chart tests from the repository root:

```sh
KUBECONFIG=/dev/null OTEL_AGENT_BINARY=/absolute/path/to/collector \
  OTEL_AGENT_IMAGE=collector-image@sha256:IMAGE_DIGEST \
  go test -count=1 -timeout 30m ./stable/suse-observability-agent/test/...
```

Extract `OTEL_AGENT_BINARY` from the same image. Without these variables, binary
validation and container transport tests skip.

`TestOtelCustomCertificatesExport` uses the binary without Docker. It preserves
each OTel workload's rendered platform exporter, authentication and trust
environment, replacing Kubernetes receivers with a local Prometheus fixture.
HTTP and gRPC export must trust inline PEM and external ConfigMap files, reject
an unrelated CA or incorrect hostname, and retain system-bundle trust with custom
certificates enabled or disabled. Certificate mount paths are relocated into temporary directories
for this process fixture.

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
