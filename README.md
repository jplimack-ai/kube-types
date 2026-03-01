# kube-types

A [golangci-lint](https://golangci-lint.run/) v2 module plugin that detects untyped Kubernetes manifest construction and suggests typed Go structs.

## Why

Kubernetes resources constructed via `map[string]any`, `fmt.Sprintf` YAML templates, or `unstructured.Unstructured` bypass Go's type system entirely. This means:

- No compile-time field validation
- No IDE autocompletion
- Typos in field names are runtime errors (or silent misconfigurations)
- Schema changes break silently

`kube-types` catches these patterns and points you to the typed struct you should use instead.

## Checks

### `map_literal`

Detects `map[string]any` (or `map[string]interface{}`) composite literals containing both `"apiVersion"` and `"kind"` keys, including named type aliases like `type Manifest map[string]any`.

```go
// Flagged:
m := map[string]any{
    "apiVersion": "apps/v1",
    "kind":       "Deployment",
    "metadata":   map[string]any{"name": "my-deploy"},
}
// Diagnostic: use *appsv1.Deployment (import "k8s.io/api/apps/v1") instead of map[string]any for apps/v1/Deployment

// Also supports const values:
const ver = "apps/v1"
m := map[string]any{"apiVersion": ver, "kind": "Deployment"}
```

### `sprintf_yaml`

Detects `fmt.Sprintf` and `fmt.Fprintf` calls where the format string contains YAML markers like `apiVersion:` or `kind:`, suggesting string-interpolated Kubernetes manifest construction. Supports both literal and `const` format strings.

```go
// Flagged:
yaml := fmt.Sprintf("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: %s", name)

// Also flagged (const format string):
const tmpl = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: %s\n"
yaml := fmt.Sprintf(tmpl, name)
```

### `unstructured_gvk`

Detects `SetGroupVersionKind` calls on `*unstructured.Unstructured` with known GVK literals, and `SetAPIVersion` + `SetKind` call pairs on the same receiver.

```go
// Flagged:
u := &unstructured.Unstructured{}
u.SetGroupVersionKind(schema.GroupVersionKind{
    Group: "apps", Version: "v1", Kind: "Deployment",
})
// Diagnostic: SetGroupVersionKind(apiVersion="apps/v1", kind="Deployment") on unstructured.Unstructured:
//   use *appsv1.Deployment (import "k8s.io/api/apps/v1") instead

// Also flagged (SetAPIVersion + SetKind pair):
u.SetAPIVersion("apps/v1")
u.SetKind("Deployment")
```

## Built-in GVKs

The plugin ships with ~35 known GVKs covering the most common Kubernetes resources:

- **core/v1**: Pod, Service, ConfigMap, Secret, ServiceAccount, Namespace, PersistentVolume, PersistentVolumeClaim, Node, Endpoints, ResourceQuota, LimitRange
- **apps/v1**: Deployment, StatefulSet, DaemonSet, ReplicaSet
- **batch/v1**: Job, CronJob
- **networking.k8s.io/v1**: Ingress, NetworkPolicy, IngressClass
- **rbac.authorization.k8s.io/v1**: Role, ClusterRole, RoleBinding, ClusterRoleBinding
- **policy/v1**: PodDisruptionBudget
- **storage.k8s.io/v1**: StorageClass
- **autoscaling/v2**: HorizontalPodAutoscaler
- **admissionregistration.k8s.io/v1**: ValidatingWebhookConfiguration, MutatingWebhookConfiguration
- **scheduling.k8s.io/v1**: PriorityClass
- **discovery.k8s.io/v1**: EndpointSlice

Unknown GVKs still produce a diagnostic suggesting you generate typed structs.

## Installation

Use golangci-lint's [custom-gcl module plugin builder](https://golangci-lint.run/plugins/module-plugins/) to include this plugin.

In your `.custom-gcl.yml`:

```yaml
version: v2.1.0
plugins:
  - module: github.com/togethercomputer/kube-types
    import: github.com/togethercomputer/kube-types
    version: latest
```

Then build your custom `golangci-lint`:

```sh
custom-gcl
```

## Configuration

Add to your `.golangci.yml`:

```yaml
version: "2"

linters-settings:
  custom:
    kube-types:
      type: "module"
      description: "Detects untyped Kubernetes manifest construction and suggests typed structs"
      settings:
        # Analyze test files (default: false).
        include_test_files: false

        # Per-check configuration. Omit for all checks enabled with defaults.
        checks:
          map_literal:
            enabled: true
          sprintf_yaml:
            enabled: true
            additional_markers:
              - "metadata:"
          unstructured_gvk:
            enabled: true

        # Register additional GVKs beyond the built-in table.
        extra_known_gvks:
          - api_version: "example.io/v1"
            kind: "Widget"
            typed_package: "example.io/api/v1.Widget"

        # Skip diagnostics for specific GVKs (format: "apiVersion/kind").
        ignore_gvks:
          - "v1/ConfigMap"

linters:
  enable:
    - kube-types
```

### Configuration Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `include_test_files` | `bool` | `false` | Analyze `_test.go` files |
| `checks` | `map[string]CheckConfig` | all enabled | Per-check enable/disable and settings |
| `checks.<name>.enabled` | `*bool` | `true` | Enable or disable a specific check |
| `checks.sprintf_yaml.additional_markers` | `[]string` | `[]` | Extra YAML markers beyond `apiVersion:` and `kind:` |
| `extra_known_gvks` | `[]GVKEntry` | `[]` | Additional GVK-to-typed-struct mappings |
| `extra_known_gvks[].api_version` | `string` | required | API version (e.g. `"apps/v1"`) |
| `extra_known_gvks[].kind` | `string` | required | Kind (e.g. `"Deployment"`) |
| `extra_known_gvks[].typed_package` | `string` | required | Full typed package path (e.g. `"k8s.io/api/apps/v1.Deployment"`) |
| `ignore_gvks` | `[]string` | `[]` | GVK keys to suppress (`"apiVersion/kind"` format) |

### Suppressing Diagnostics

Use `//nolint:kube-types` to suppress a specific line:

```go
m := map[string]any{ //nolint:kube-types
    "apiVersion": "v1",
    "kind":       "ConfigMap",
}
```

## Development

```sh
# Run tests
make test

# Run linter
make lint

# Tidy deps
make tidy

# Build
make build
```

## Known Limitations

- **Two-step map construction** is not detected. `m := make(map[string]any); m["apiVersion"] = "v1"; m["kind"] = "Pod"` won't fire because the keys are set via statements, not in the composite literal.
- **Cross-function SetAPIVersion/SetKind pairs** are not tracked. Both calls must be on the same receiver variable within the same function body.
- **Non-const variable format strings** in `fmt.Sprintf` are not analyzed. Only string literals and `const` strings are resolved.
- The `sprintf_yaml` check uses substring matching for markers. A string like `"log kind: info"` would be flagged if `kind:` is a marker.
