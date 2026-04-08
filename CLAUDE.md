# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

See `AGENTS.md` for detailed code style guidelines, test patterns, and advanced examples.

## Project Overview

Helm charts repository for SUSE Observability. The main product chart is `stable/suse-observability/`. Charts with "stackstate" in the name are deprecated and should be ignored.

## Common Commands

### Running Tests (Go + Terratest, from repo root)

```bash
go test ./stable/suse-observability/test/...                          # All tests for main chart
go test ./stable/suse-observability/test/... -run TestFeatureName      # Single test
go test -v ./stable/suse-observability/test/... -run TestFeatureName   # Verbose
go test ./stable/<chart>/itest/...                                     # Integration tests
go test ./...                                                          # All tests
```

### Helm

```bash
helm dependencies build stable/suse-observability                     # Update dependencies
./stable/suse-observability/update-chart-dependencies.sh              # Helper for subchart deps
helm lint stable/suse-observability                                   # Lint chart
helm template my-release stable/suse-observability -f values.yaml     # Render templates
```

### Pre-commit

```bash
pre-commit install && pre-commit install-hooks    # Setup
pre-commit run --all-files                        # Run all hooks
pre-commit run helmlint --all-files               # Run specific hook
```

## Updatecli (Docker Image Updates)

For tasks involving updatecli pipelines — adding/removing tracked images, changing tag filters, modifying the update pipeline, or debugging image version bumps — see **[UPDATECLI.md](UPDATECLI.md)** for full documentation of the pipeline architecture, target chain, tag format patterns, and local testing workflow.

## Architecture

- **`stable/`** - All Helm charts. Key charts: `suse-observability` (main), `common` (shared library), plus infrastructure charts (elasticsearch, hbase, kafka, clickhouse, etc.)
- **`helmtestutil/`** - Custom Go test utilities wrapping Terratest. Provides `RenderHelmTemplate()`, `RenderHelmTemplateError()`, `RenderHelmTemplateOptsNoError()`, and `NewKubernetesResources()` for parsing rendered output into typed K8s objects (Deployments, StatefulSets, ConfigMaps, Services, etc.)
- **`.gitlab-ci.jsonnet`** - CI pipeline source. Generates `.gitlab-ci.yml` (auto-generated, never edit manually)
- **`scripts/`** - Build/publish/version-bump scripts

## Test Structure

Tests live in `stable/<chart>/test/` with test value files in `test/values/`. Test values use descriptive naming: `feature_enabled.yaml`, `feature_missing_required.yaml`, `feature_invalid_value.yaml`.

### Test Patterns

```go
// Successful render
output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/feature.yaml")
resources := helmtestutil.NewKubernetesResources(t, output)

// Expected error
err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/invalid.yaml")

// With helm.Options for SetValues overrides
output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
    ValuesFiles: []string{"values/full.yaml"},
    SetValues: map[string]string{"global.backup.enabled": "true"},
})
```

## Key Rules

- **README.md files in chart directories are auto-generated** from `README.md.gotmpl` via helm-docs. Never edit them manually.
- **`.gitlab-ci.yml` is auto-generated** from `.gitlab-ci.jsonnet`. Never edit manually.
- **Document values** in `values.yaml` using helm-docs comment format: `# key.subkey -- Description.`
- **Go imports order**: stdlib, external, internal (`gitlab.com/StackVista/DevOps/helm-charts/helmtestutil`).
- **Test naming**: `Test<Feature><Scenario>` (e.g., `TestAuthenticationLdapMissingValues`).
- **YAML**: 2-space indent, camelCase keys, quote ambiguous strings, newline at EOF.
- **Shell scripts**: Must pass shellcheck. Use `gawk` instead of `awk`.
- **Go version**: 1.24.0+
