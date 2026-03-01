package map_literal_named

// Named type alias for map[string]any: should be flagged.
type Manifest map[string]any

var deployment = Manifest{ // want `use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead of map\[string\]any for apps/v1/Deployment`
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
}

// Named type alias with unknown GVK: should be flagged with generic message.
var crd = Manifest{ // want `map\[string\]any with apiVersion "custom\.io/v1" and kind "Foo" constructs a Kubernetes manifest without type safety`
	"apiVersion": "custom.io/v1",
	"kind":       "Foo",
}

// Named type alias for map[string]string: should NOT be flagged.
type Labels map[string]string

var labels = Labels{
	"app":     "my-app",
	"version": "v1",
}
