package sprintf_yaml

import "fmt"

// Sprintf with apiVersion marker: should be flagged.
var yamlStr = fmt.Sprintf("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: %s\n", "my-deploy") // want `fmt\.Sprintf with YAML marker "apiVersion:" suggests string-interpolated Kubernetes manifest`

// Sprintf with both markers: should be flagged (reports first matched marker).
var kindStr = fmt.Sprintf("kind: Pod\napiVersion: v1\n") // want `fmt\.Sprintf with YAML marker "apiVersion:" suggests string-interpolated Kubernetes manifest`

// Fprintf with both markers: should be flagged.
func writeManifest() {
	fmt.Fprintf(nil, "kind: Pod\napiVersion: v1\n") // want `fmt\.Fprintf with YAML marker "apiVersion:" suggests string-interpolated Kubernetes manifest`
}

// Sprintf with only kind marker: should be flagged.
var kindOnly = fmt.Sprintf("kind: Service\nname: foo\n") // want `fmt\.Sprintf with YAML marker "kind:" suggests string-interpolated Kubernetes manifest`

// Sprintf without markers: should NOT be flagged.
var safe = fmt.Sprintf("hello %s", "world")

// Sprintf with unrelated content: should NOT be flagged.
var unrelated = fmt.Sprintf("name: %s, age: %d", "alice", 30)

// Const format string: should be flagged.
const manifestTemplate = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: %s\n"

var fromConst = fmt.Sprintf(manifestTemplate, "my-pod") // want `fmt\.Sprintf with YAML marker "apiVersion:" suggests string-interpolated Kubernetes manifest`

// Non-const variable format string: should NOT be flagged.
var dynamicTemplate = "apiVersion: v1\nkind: Pod\n"

var fromVar = fmt.Sprintf(dynamicTemplate, "my-pod") //nolint:govet // intentional extra arg for test
