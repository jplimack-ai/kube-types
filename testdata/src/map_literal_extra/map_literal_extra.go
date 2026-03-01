package map_literal_extra

// Extra known GVK from user config: should be flagged with typed struct suggestion.
var widget = map[string]any{ // want `use \*v1\.Widget \(import "example\.io/api/v1"\) instead of map\[string\]any for example\.io/v1/Widget`
	"apiVersion": "example.io/v1",
	"kind":       "Widget",
}
