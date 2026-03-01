package map_literal

// Known GVK: should be flagged with typed struct suggestion.
var deployment = map[string]any{ // want `use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead of map\[string\]any for apps/v1/Deployment`
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
	"metadata": map[string]any{
		"name": "my-deploy",
	},
}

// Known GVK with interface{}: should also be flagged.
var pod = map[string]interface{}{ // want `use \*corev1\.Pod \(import "k8s\.io/api/core/v1"\) instead of map\[string\]any for v1/Pod`
	"apiVersion": "v1",
	"kind":       "Pod",
}

// Unknown GVK: should be flagged with generic message.
var crd = map[string]any{ // want `map\[string\]any with apiVersion "example\.io/v1" and kind "Widget" constructs a Kubernetes manifest without type safety`
	"apiVersion": "example.io/v1",
	"kind":       "Widget",
}
