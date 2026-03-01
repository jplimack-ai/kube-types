package sprintf_markers

import "fmt"

// Custom marker "metadata:" should be flagged with additional_markers config.
var metadataStr = fmt.Sprintf("metadata:\n  name: %s\n", "my-deploy") // want `fmt\.Sprintf with YAML marker "metadata:" suggests string-interpolated Kubernetes manifest`

// Default markers still work.
var apiVersionStr = fmt.Sprintf("apiVersion: v1\nkind: Pod\n") // want `fmt\.Sprintf with YAML marker "apiVersion:" suggests string-interpolated Kubernetes manifest`

// No markers: should NOT be flagged.
var safeStr = fmt.Sprintf("hello %s", "world")
