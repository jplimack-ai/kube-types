package unstructured_gvk

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SetGroupVersionKind with known GVK literal: should be flagged.
func knownGVK() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{ // want `SetGroupVersionKind\(apiVersion="apps/v1", kind="Deployment"\) on unstructured.Unstructured: use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead`
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
}

// SetGroupVersionKind with core API (no group): should be flagged.
func coreGVK() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{ // want `SetGroupVersionKind\(apiVersion="v1", kind="Pod"\) on unstructured.Unstructured: use \*corev1\.Pod \(import "k8s\.io/api/core/v1"\) instead`
		Version: "v1",
		Kind:    "Pod",
	})
}

// SetGroupVersionKind with unknown GVK: should NOT be flagged.
func unknownGVK() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "example.io",
		Version: "v1",
		Kind:    "Widget",
	})
}

// SetGroupVersionKind with variable arg: should NOT be flagged.
func variableGVK() {
	u := &unstructured.Unstructured{}
	gvk := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}
	u.SetGroupVersionKind(gvk)
}

// SetAPIVersion + SetKind pair with known GVK: should be flagged.
func setPair() {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("apps/v1") // want `SetAPIVersion\("apps/v1"\) \+ SetKind\("Deployment"\) on unstructured\.Unstructured: use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead`
	u.SetKind("Deployment")
}

// Only SetAPIVersion, no SetKind: should NOT be flagged.
func setAPIVersionOnly() {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("apps/v1")
}

// Only SetKind, no SetAPIVersion: should NOT be flagged.
func setKindOnly() {
	u := &unstructured.Unstructured{}
	u.SetKind("Deployment")
}

// SetAPIVersion + SetKind with unknown GVK: should NOT be flagged.
func setPairUnknown() {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("example.io/v1")
	u.SetKind("Widget")
}
