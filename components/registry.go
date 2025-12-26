package components

import (
	g "maragu.dev/gomponents"
)

// PageFunc is the signature for all page components
type PageFunc func(ctx *PageContext) g.Node

// registry holds all registered page components
var registry = make(map[string]PageFunc)

// Register adds a page component to the registry
func Register(id string, fn PageFunc) {
	registry[id] = fn
}

// Get retrieves a page component by ID
func Get(id string) (PageFunc, bool) {
	fn, ok := registry[id]
	return fn, ok
}

// MustGet retrieves a page component or panics
func MustGet(id string) PageFunc {
	fn, ok := Get(id)
	if !ok {
		panic("component not found: " + id)
	}
	return fn
}

// List returns all registered component IDs
func List() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	return ids
}
