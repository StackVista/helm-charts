Run chart tests from the repository root:

```sh
KUBECONFIG=/dev/null go test -count=1 -timeout 30m ./stable/suse-observability-agent/test/...
```

Optionally set `OTEL_AGENT_BINARY` to the collector binary extracted from the
chart's selected image. The binary test runs `validate` on the rendered config and
container arguments, checking collector components, fields and feature gates
without starting collection. It skips when the variable is unset.
