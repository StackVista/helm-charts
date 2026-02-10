# AGENTS.md - Guidelines for AI Coding Agents

This repository contains Helm charts for SUSE Observability and related components.
The main product chart is `stable/suse-observability/`. Charts with "stackstate" in
the name are deprecated and should be ignored.

## Repository Structure

```
helm-charts/
├── stable/                     # All Helm charts
│   ├── suse-observability/     # Main product chart
│   ├── common/                 # Shared chart library
│   ├── elasticsearch/          # Supporting charts
│   ├── hbase/
│   ├── kafka/
│   ├── clickhouse/
│   └── ...
├── helmtestutil/               # Go test utilities
├── .gitlab-ci.jsonnet          # CI source (generates .gitlab-ci.yml)
├── .pre-commit-config.yaml     # Pre-commit hooks
└── go.mod                      # Go module for tests
```

## Build & Test Commands

**IMPORTANT**: Always run tests after making changes to a Helm chart to validate
your changes. When adding new functionality, add corresponding tests. When modifying
existing functionality, update the relevant tests. When removing functionality,
remove or update tests that are no longer applicable.

### Running Tests

Tests use Go with Terratest. Run from the repository root:

```bash
# Run all tests for a specific chart
go test ./stable/suse-observability/test/...

# Run a single test by name
go test ./stable/suse-observability/test/... -run TestHelmBasicRender

# Run tests with verbose output
go test -v ./stable/suse-observability/test/... -run TestAuthenticationLdap

# Run all tests in the repository
go test ./...
```

### Helm Commands

```bash
# Update chart dependencies (run from chart directory)
helm dependencies build stable/suse-observability

# Or use the helper script
./stable/suse-observability/update-chart-dependencies.sh

# Lint a chart
helm lint stable/suse-observability

# Render templates locally (useful for debugging)
helm template my-release stable/suse-observability -f values.yaml
```

### Pre-commit Hooks

```bash
# Install pre-commit
pre-commit install

# Run all hooks manually
pre-commit run --all-files

# Run specific hook
pre-commit run helmlint --all-files
```

## Code Style Guidelines

### Helm Templates

- Use `.helmignore` to exclude test files from packaged charts
- Document values in `values.yaml` using helm-docs comment format:
  ```yaml
  # key.subkey -- Description of the value.
  key:
    subkey: "default-value"
  ```
- README.md files are auto-generated from `README.md.gotmpl` using helm-docs
- Do NOT manually edit README.md files in chart directories

### Go Test Files

Tests live in `stable/<chart>/test/` directories with test values in `test/values/`.

**Imports** - Use this order (stdlib, external, internal):
```go
package test

import (
    "testing"

    "github.com/gruntwork-io/terratest/modules/helm"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
    corev1 "k8s.io/api/core/v1"
)
```

**Test function naming**: `Test<Feature><Scenario>` (e.g., `TestAuthenticationLdapMissingValues`)

**Test patterns** - Use the helmtestutil package:

```go
// Basic render test - expects success
func TestFeatureWorks(t *testing.T) {
    output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/feature.yaml")
    resources := helmtestutil.NewKubernetesResources(t, output)

    // Access typed K8s resources
    deployment, ok := resources.Deployments["my-deployment"]
    require.True(t, ok, "Deployment should exist")
    assert.Equal(t, int32(3), *deployment.Spec.Replicas)
}

// Error validation test - expects failure
func TestFeatureValidation(t *testing.T) {
    err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/invalid.yaml")
    require.Contains(t, err.Error(), "expected error message")
}

// Test with helm.Options for SetValues
func TestWithOptions(t *testing.T) {
    output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
        ValuesFiles: []string{"values/full.yaml"},
        SetValues: map[string]string{
            "global.backup.enabled": "true",
        },
    })
    resources := helmtestutil.NewKubernetesResources(t, output)
    // assertions...
}
```

**KubernetesResources struct** provides typed access to:
- `Deployments`, `StatefulSets`, `DaemonSets`
- `ConfigMaps`, `Secrets`
- `Services`, `ServiceAccounts`
- `Ingresses`, `Jobs`, `CronJobs`
- `ClusterRoles`, `ClusterRoleBindings`, `Roles`, `RoleBindings`
- `PersistentVolumeClaims`, `Pdbs`
- `ServiceMonitors` (Prometheus)

### Advanced Test Example

For a comprehensive example of testing complex features with multiple scenarios, see:
**`stable/suse-observability/test/features_test.go`**

This test file demonstrates:

1. **Testing feature flags and their defaults**:
   ```go
   func TestFeaturesTracesDefault(t *testing.T) {
       output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
           ValuesFiles: []string{"values/full.yaml"},
       })
       resources := helmtestutil.NewKubernetesResources(t, output)

       // Verify deployment environment variables
       deployment, ok := resources.Deployments["suse-observability-api"]
       require.True(t, ok, "API deployment should exist")
       expected := corev1.EnvVar{Name: "CONFIG_FORCE_stackstate_webUIConfig_featureFlags_traces", Value: "true"}
       assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, expected)

       // Verify ConfigMap contents
       configMap, ok := resources.ConfigMaps["suse-observability-api"]
       require.True(t, ok, "API configmap should exist")
       assert.Contains(t, configMap.Data["application_stackstate.conf"], expectedClickhouseConfig)
   }
   ```

2. **Testing with SetValues to override configuration**:
   ```go
   func TestFeaturesTracesDisabled(t *testing.T) {
       output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
           ValuesFiles: []string{"values/full.yaml"},
           SetValues: map[string]string{
               "stackstate.features.traces": "false",
           },
       })
       // ... verify feature is disabled
   }
   ```

3. **Testing resource existence with helper functions**:
   ```go
   func assertSplitComponentResourcesExist(t *testing.T, resources *helmtestutil.KubernetesResources, shouldExist bool) {
       for _, component := range splitComponents {
           _, ok := resources.Deployments[component]
           assert.Equal(t, shouldExist, ok, "%s deployment existence", component)

           _, ok = resources.ConfigMaps[component]
           assert.Equal(t, shouldExist, ok, "%s configMap existence", component)

           // ... check other resources
       }
   }
   ```

4. **Testing RBAC resources with structured comparisons**:
   ```go
   func checkClusterRoleBinding(t *testing.T, expected, got rbacv1.ClusterRoleBinding) {
       assert.Equal(t, expected.Name, got.Name, "ClusterRoleBinding name should match expected")
       assert.Equal(t, expected.RoleRef, got.RoleRef, "ClusterRoleBinding role reference should match expected")
       assert.Equal(t, expected.Subjects, got.Subjects, "ClusterRoleBinding subjects should match expected")
   }
   ```

5. **Testing with multiple scenarios using table-driven patterns**:
   - Test default behavior (`TestFeaturesServerSplitDefault`)
   - Test explicitly enabled (`TestFeaturesServerSplitEnabled`)
   - Test explicitly disabled (`TestFeaturesServerSplitDisabled`)
   - Test override behavior (`TestFeaturesServerSplitExperimentalOverFeatures`)

**Key patterns from features_test.go**:
- Define expected values as constants or package-level variables for reusability
- Use helper functions to extract common assertion logic
- Test both positive (resource exists) and negative (resource does not exist) cases
- Verify multiple resource types affected by a single feature flag
- Check environment variables, ConfigMap contents, and resource specifications
- Use descriptive test names that indicate what scenario is being tested

### YAML Style

- Use 2-space indentation
- Use lowercase for keys with camelCase for multi-word keys
- Quote strings that could be interpreted as other types
- Files must end with a newline
- No trailing whitespace

### Shell Scripts

- Must have executable shebang: `#!/bin/bash` or `#!/usr/bin/env bash`
- Validated with shellcheck
- Use `gawk` instead of `awk` for portability

## CI/CD Notes

- `.gitlab-ci.yml` is auto-generated from `.gitlab-ci.jsonnet` - DO NOT EDIT MANUALLY
- Pre-commit hooks run: helm lint, shellcheck, helm-docs, jsonnet formatting
- CI validates with: yamale, yamllint, kubeconform, helm lint

## Chart Structure

Each chart follows this structure:
```
stable/<chart>/
├── Chart.yaml          # Chart metadata and dependencies
├── values.yaml         # Default values (annotated for helm-docs)
├── README.md           # Auto-generated (do not edit)
├── README.md.gotmpl    # Template for README generation
├── .helmignore         # Files to exclude from chart package
├── templates/          # Kubernetes manifest templates
├── test/               # Go unit tests
│   ├── *_test.go
│   └── values/         # Test value files
├── ci/                 # CI-specific values
└── linter_values.yaml  # Values for linting (optional)
```

## Common Validation Patterns

Test value files for validation scenarios use descriptive names:
- `feature_enabled.yaml` - Valid configuration
- `feature_missing_required.yaml` - Missing required field
- `feature_invalid_value.yaml` - Invalid value format

## Tool Versions

- Go: 1.24.0+
- Helm: latest
- Pre-commit: latest
