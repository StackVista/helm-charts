# SUSE Observability Sizing Chart

This library chart provides named templates for sizing SUSE Observability deployments based on profiles.

## How Sizing Templates Are Wired

### 1. Profile Detection & Activation

```
User Sets Profile
│
│   values.yaml:
│   global:
│     suseObservability:
│       sizing:
│         profile: "500-ha"
│
▼
┌─────────────────────────────────────────────────────┐
│ suse-observability/templates/_helper-global.tpl     │
│                                                      │
│ {{- define "suse-observability.global.enabled" -}}  │
│   Checks if ANY of:                                 │
│   - global.suseObservability.sizing.profile         │
│   - global.suseObservability.license                │
│   - global.suseObservability.baseUrl                │
│   → Returns "true" if sizing mode active            │
│ {{- end -}}                                         │
└─────────────────────────────────────────────────────┘
```

### 2. Main Chart → Sizing Templates (StackState Components)

```
┌──────────────────────────────────────────────────────────────┐
│ suse-observability/templates/deployment-api.yaml             │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ Step 1: Call sizing template for resources                   │
│ {{- $profileResources :=                                     │
│       include "common.sizing.stackstate.api.resources" .     │
│       | trim -}}                                             │
│                                                               │
│ Step 2: Use profile value OR default                         │
│ {{- $evaluatedResources :=                                   │
│       .Values.stackstate.components.api.resources }}         │
│ {{- if $profileResources }}                                  │
│   {{- $evaluatedResources = fromYaml $profileResources }}    │
│ {{- end }}                                                   │
│                                                               │
│ Step 3: Render in deployment                                 │
│ resources:                                                    │
│   {{- toYaml $evaluatedResources | nindent 2 }}             │
│                                                               │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│ suse-observability-sizing/templates/_sizing-stackstate.tpl   │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ {{- define "common.sizing.stackstate.api.resources" -}}      │
│ {{- if and .Values.global                                    │
│            .Values.global.suseObservability                  │
│            .Values.global.suseObservability.sizing           │
│            .Values.global.suseObservability.sizing.profile -}}│
│   {{- $profile := ... -}}                                    │
│                                                               │
│   {{- if eq $profile "500-ha" }}                             │
│   requests:                                                   │
│     cpu: "3000m"                                             │
│     memory: "6Gi"                                            │
│     ephemeral-storage: "1Mi"                                 │
│   limits:                                                     │
│     cpu: "6000m"                                             │
│     memory: "6Gi"                                            │
│     ephemeral-storage: "2Gi"                                 │
│   {{- else if eq $profile "150-ha" }}                        │
│   requests:                                                   │
│     ...                                                       │
│   {{- end }}                                                 │
│ {{- end }}                                                   │
│ {{- end }}                                                   │
│                                                               │
│ Returns: YAML structure (or empty if no profile match)       │
└──────────────────────────────────────────────────────────────┘
```

### 3. Environment Variables (Multi-Level Merge)

```
┌──────────────────────────────────────────────────────────────┐
│ suse-observability/templates/_helper-service-config.tpl      │
│                                                               │
│ Merge order (LAST WINS):                                     │
│                                                               │
│ 1. Default all-level env                                     │
│    stackstate.components.all.extraEnv.open                   │
│                                                               │
│ 2. ┌────────────────────────────────────────────┐            │
│    │ Profile all-level env                      │            │
│    │ common.sizing.stackstate.all.extraEnv.open │            │
│    └────────────────────────────────────────────┘            │
│                                                               │
│ 3. Default component-level env                               │
│    stackstate.components.<component>.extraEnv.open           │
│                                                               │
│ 4. ┌────────────────────────────────────────────┐            │
│    │ Profile component-level env                │            │
│    │ common.sizing.stackstate.<component>.env   │            │
│    └────────────────────────────────────────────┘            │
│                                                               │
│ Result: Component pod gets merged env vars                   │
└──────────────────────────────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────────────────────────────┐
│ suse-observability-sizing/templates/                         │
│   _sizing-stackstate-env.tpl                                 │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ All-level (applies to ALL components):                       │
│ {{- define "common.sizing.stackstate.all.extraEnv.open" -}}  │
│   {{- if eq $profile "500-ha" }}                             │
│   CONFIG_FORCE_stackstate_agents_agentLimit: "500"           │
│   {{- end }}                                                 │
│ {{- end }}                                                   │
│                                                               │
│ Component-level (applies to ONE component):                  │
│ {{- define "common.sizing.stackstate.server.env" -}}         │
│   {{- if eq $profile "150-ha" }}                             │
│   CONFIG_FORCE_stackstate_sync_initializationBatchParallelism: "4"│
│   {{- end }}                                                 │
│ {{- end }}                                                   │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

### 4. Java Heap Memory Configuration (Special Merge)

```
┌──────────────────────────────────────────────────────────────┐
│ suse-observability/templates/deployment-healthSync.yaml      │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ Step 1: Get profile base memory                              │
│ {{- $profileBaseMemory :=                                    │
│       include "common.sizing.stackstate.healthSync           │
│                .baseMemoryConsumption" . | trim -}}          │
│                                                               │
│ Step 2: Get default sizing dict                              │
│ {{- $evaluatedSizing :=                                      │
│       .Values.stackstate.components.healthSync.sizing }}     │
│                                                               │
│ Step 3: Merge (profile value FIRST to override defaults)     │
│ {{- if $profileBaseMemory }}                                 │
│   {{- $evaluatedSizing = merge (dict)                        │
│         (dict "baseMemoryConsumption" $profileBaseMemory)    │
│         $evaluatedSizing }}                                  │
│ {{- end }}                                                   │
│                                                               │
│ Note: Helm merge gives precedence to FIRST dict with key!    │
│                                                               │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│ suse-observability-sizing/templates/_sizing-stackstate.tpl   │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ {{- define "common.sizing.stackstate.healthSync              │
│             .baseMemoryConsumption" -}}                       │
│   {{- if eq $profile "500-ha" }}1200Mi                       │
│   {{- else if eq $profile "4000-ha" }}1800Mi                 │
│   {{- end }}                                                 │
│ {{- end }}                                                   │
│                                                               │
│ IMPORTANT: Returns value WITHOUT quotes!                     │
│ ✅ Correct: 1200Mi                                           │
│ ❌ Wrong:   "1200Mi" (breaks arithmetic in JAVA_OPTS calc)   │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

### 5. Subchart Integration Patterns

There are **two different patterns** used across subcharts for integrating with the sizing library:

#### Pattern A: Direct Calls (kafka, kafkaup-operator, elasticsearch, clickhouse, victoria-metrics-single, zookeeper)

These subcharts **call sizing templates directly** from their resource templates:

```
┌──────────────────────────────────────────────────────────────┐
│ kafka/templates/statefulset.yaml                             │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ {{- $sizingReplicaCount :=                                   │
│       include "common.sizing.kafka.replicaCount" . | trim -}}│
│                                                               │
│ replicas: {{ .Values.replicaCount }}                         │
│ {{- if $sizingReplicaCount }}                                │
│ replicas: {{ $sizingReplicaCount }}                          │
│ {{- end }}                                                   │
│                                                               │
│ resources:                                                    │
│   {{- $profileResources :=                                   │
│         include "common.sizing.kafka.resources" . | trim -}} │
│   {{- if $profileResources }}                                │
│   {{- fromYaml $profileResources | toYaml | nindent 2 }}     │
│   {{- else }}                                                │
│   {{- toYaml .Values.resources | nindent 2 }}                │
│   {{- end }}                                                 │
│                                                               │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│ suse-observability-sizing/templates/_sizing-kafka.tpl        │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ {{- define "common.sizing.kafka.replicaCount" -}}            │
│   {{- if eq $profile "150-ha" }}3                            │
│   {{- else if eq $profile "500-ha" }}5                       │
│   {{- end }}                                                 │
│ {{- end }}                                                   │
│                                                               │
│ Returns: Value or empty (subchart handles fallback)          │
└──────────────────────────────────────────────────────────────┘
```

**Benefits:**

- Simpler - no intermediate wrapper layer
- Fallback logic lives in the subchart template where it's used
- Direct visibility into what's being evaluated

**Used by:** kafka, kafkaup-operator, elasticsearch, clickhouse, victoria-metrics-single, zookeeper

#### Pattern B: Local Wrapper Helpers (hbase)

The hbase subchart has **`_sizing.tpl`** with wrapper functions that encapsulate sizing detection logic:

```
┌──────────────────────────────────────────────────────────────┐
│ hbase/templates/stackgraph-statefulset.yaml                  │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ storage: {{ include "hbase.stackgraph.persistence.size" . }} │
│                                                               │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│ hbase/templates/_sizing.tpl (LOCAL WRAPPER)                  │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ {{- define "hbase.stackgraph.persistence.size" -}}           │
│                                                               │
│ Step 1: Check if sizing profile is set                       │
│ {{- if and .Values.global .Values.global.suseObservability   │
│           .Values.global.suseObservability.sizing            │
│           .Values.global.suseObservability.sizing.profile -}}│
│                                                               │
│ Step 2: Use profile value OR default                         │
│   {{- $profileSize :=                                        │
│         include "common.sizing.hbase.stackgraph              │
│                   .persistence.size" . | trim -}}            │
│   {{- if $profileSize -}}                                    │
│     {{- $profileSize -}}          ← FROM SIZING CHART        │
│   {{- else -}}                                               │
│     {{- .Values.stackgraph.persistence.size -}}              │
│   {{- end -}}                     ← FALLBACK TO DEFAULT      │
│ {{- else -}}                                                 │
│   {{- .Values.stackgraph.persistence.size -}}                │
│ {{- end -}}                       ← STANDALONE MODE          │
│ {{- end }}                                                   │
│                                                               │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│ suse-observability-sizing/templates/_sizing-hbase.tpl        │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ {{- define "common.sizing.hbase.stackgraph                   │
│             .persistence.size" -}}                            │
│   {{- if eq $profile "150-ha" }}50Gi                         │
│   {{- else if eq $profile "500-ha" }}100Gi                   │
│   {{- end }}                                                 │
│ {{- end }}                                                   │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

**Benefits:**

- Encapsulates mode detection logic (DRY principle)
- Provides chart-specific helper namespace (`hbase.*` instead of `common.sizing.*`)
- Easier to read in resource templates

**Used by:** hbase only

**Why the difference?**

- HBase was the first subchart to be migrated and established the wrapper pattern
- Other subcharts adopted the simpler direct-call pattern
- Both patterns work correctly; the difference is stylistic

#### Pattern C: Main Chart Integration (`suse-observability`)

The main `suse-observability` chart has **two** helper files for different purposes:

**`_helper-global.tpl`** - Integration orchestration and global logic:

- Mode detection (`suse-observability.global.enabled`)
- Value extraction with fallback (`suse-observability.global.license`)
- Validation (`suse-observability.global.validate`)
- Affinity multi-layer merging (`suse-observability.global.affinity`)
- Not named `_sizing.tpl` because it contains **orchestration logic**, not just sizing

**`_helper-sizing-overrides.tpl`** - Subchart value wrappers:

- Convenience wrappers for subchart values
- Example: `suse-observability.sizing.kafka.persistence.size`
- Wraps `common.sizing.X.*` templates for use in values overrides
- Used in `suse-observability/values.yaml` to set subchart defaults from profiles

### 6. Complete Data Flow Example (API Component)

```
User YAML
─────────
global:
  suseObservability:
    sizing:
      profile: "500-ha"
    license: "..."
    baseUrl: "..."

          │
          ▼

┌─────────────────────────────────────────┐
│ Profile Detection                        │
│ suse-observability.global.enabled        │
│ → Returns: "true"                        │
└────────────┬────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│ deployment-api.yaml                      │
│                                          │
│ Calls 3 sizing templates:               │
│ 1. api.resources                         │
│ 2. api.baseMemoryConsumption             │
│ 3. api.env (if exists)                   │
└────────────┬────────────────────────────┘
             │
             ├─────────────────────────┐
             │                         │
             ▼                         ▼
┌──────────────────────┐   ┌──────────────────────┐
│ api.resources        │   │ api.baseMemory...    │
│                      │   │                      │
│ Returns YAML:        │   │ Returns: 500Mi       │
│ requests:            │   │ (no quotes!)         │
│   cpu: "3000m"       │   │                      │
│   memory: "6Gi"      │   │ Used in:             │
│   ephemeral-storage  │   │ - JAVA_OPTS calc     │
│ limits:              │   │ - Direct memory      │
│   cpu: "6000m"       │   │                      │
│   memory: "6Gi"      │   │                      │
│   ephemeral-storage  │   │                      │
└──────────┬───────────┘   └──────────┬───────────┘
           │                          │
           │                          │
           └──────────┬───────────────┘
                      │
                      ▼
            ┌─────────────────────┐
            │ Kubernetes Pod Spec │
            │                     │
            │ resources:          │
            │   requests:         │
            │     cpu: 3000m      │
            │     memory: 6Gi     │
            │     ephemeral-...   │
            │   limits:           │
            │     cpu: 6000m      │
            │     memory: 6Gi     │
            │     ephemeral-...   │
            │                     │
            │ env:                │
            │ - JAVA_OPTS=        │
            │   -Xmx...           │
            │   -XX:MaxDirect...  │
            └─────────────────────┘
```

## Template Naming Convention

All sizing templates follow this pattern:

```
common.sizing.<component-type>.<component-name>.<property>
```

### Examples by Component Type

**StackState Components:**

- `common.sizing.stackstate.api.resources`
- `common.sizing.stackstate.healthSync.baseMemoryConsumption`
- `common.sizing.stackstate.sync.env`
- `common.sizing.stackstate.all.extraEnv.open`

**Infrastructure Components:**

- `common.sizing.kafka.storage.size`
- `common.sizing.kafka.replicaCount`
- `common.sizing.clickhouse.storage.size`
- `common.sizing.elasticsearch.master.resources`
- `common.sizing.hbase.regionserver.replicaCount`

**Global Configuration:**

- `common.sizing.stackstate.server.split` (mono vs split mode)
- `common.sizing.affinity.ha` (whether HA affinity enabled)

## Key Design Principles

### 1. Complete Replacement (Not Merge)

When a sizing template returns resources, it **completely replaces** the default:

```yaml
# Profile template returns complete structure
requests:
  cpu: "3000m"
  memory: "6Gi"
  ephemeral-storage: "1Mi"  ← Must include ALL resources
limits:
  cpu: "6000m"
  memory: "6Gi"
  ephemeral-storage: "2Gi"  ← Missing this breaks functionality
```

### 2. No Quotes for Arithmetic Values

Templates used in calculations must return **unquoted** values:

```yaml
# ✅ Correct - no quotes
{{- if eq $profile "500-ha" }}1200Mi{{- end }}

# ❌ Wrong - breaks arithmetic
{{- if eq $profile "500-ha" }}"1200Mi"{{- end }}
```

### 3. Helm Merge Order

First dict wins when merging:

```yaml
# ✅ Correct - profile overrides default
{{- $result = merge (dict) (dict "key" $profileValue) $defaults }}

# ❌ Wrong - default overrides profile
{{- $result = merge (dict) $defaults (dict "key" $profileValue) }}
```

### 4. Empty Return = Use Default

If template returns empty, caller uses default from values.yaml:

```yaml
{{- define "common.sizing.stackstate.router.resources" -}}
{{- if eq $profile "150-ha" }}
# ... return resources ...
{{- end }}
# No else case = returns empty for other profiles
{{- end }}

# Caller uses default:
{{- if $profileResources }}
  {{- use profile }}
{{- else }}
  {{- use .Values default }}  ← Falls back here
{{- end }}
```
