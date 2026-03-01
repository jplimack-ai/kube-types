package ignore_gvk

// Deployment is normally flagged, but is in IgnoreGVKs — should NOT be flagged.
var deployment = map[string]any{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
	"metadata": map[string]any{
		"name": "my-deploy",
	},
}

// Pod is NOT in IgnoreGVKs — should still be flagged.
var pod = map[string]any{ // want `use \*corev1\.Pod \(import "k8s\.io/api/core/v1"\) instead of map\[string\]any for v1/Pod`
	"apiVersion": "v1",
	"kind":       "Pod",
}
