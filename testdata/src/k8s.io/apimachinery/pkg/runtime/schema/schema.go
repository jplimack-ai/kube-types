package schema

// GroupVersionKind is a minimal mock of k8s.io/apimachinery's GroupVersionKind for testing.
type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}
