package components

import (
	"github.com/ZacxDev/go-static-site/config"
)

// LoadedMarkdownRoute represents a markdown route with its frontmatter
type LoadedMarkdownRoute struct {
	Path        string         `json:"path"`
	Frontmatter map[string]any `json:"frontmatter,omitempty"`
}

// PageContext provides all data available to gomponents pages
type PageContext struct {
	// Route parameters from URL (e.g., :slug, :id)
	Params map[string]string

	// Language and i18n
	Lang           string
	SupportedLangs []string
	Translations   map[string]string // Pre-filtered for current lang

	// Page metadata
	Title       string
	Description string
	Canonical   string
	CurrentPath string
	PathNoLang  string // Path without language prefix, for building lang switcher links

	// JavaScript/CSS bundles from esbuild
	EsbuildBundlePaths []string

	// Environment
	IsProductionEnvironment bool
	AppOrigin               string // Deprecated but kept for compatibility

	// Custom data from StaticRenderData + GlobalRenderContext
	Data map[string]any

	// Markdown routes (for blog listings, navigation, etc.)
	MarkdownRoutes []LoadedMarkdownRoute

	// Registered routes (for navigation generation)
	RegisteredRoutes []string

	// Internal references for advanced use cases
	Manifest *config.SiteManifest
	Route    config.Route
}

// Text returns a translation for the given key.
// Returns the key itself if no translation is found.
func (ctx *PageContext) Text(key string) string {
	if ctx.Translations == nil {
		return key
	}
	if t, ok := ctx.Translations[key]; ok {
		return t
	}
	return key
}

// Get retrieves a value from Data by key
func (ctx *PageContext) Get(key string) any {
	if ctx.Data == nil {
		return nil
	}
	return ctx.Data[key]
}

// GetString retrieves a string value from Data by key
func (ctx *PageContext) GetString(key string) string {
	if v, ok := ctx.Data[key].(string); ok {
		return v
	}
	return ""
}

// GetInt retrieves an int value from Data by key
func (ctx *PageContext) GetInt(key string) int {
	if v, ok := ctx.Data[key].(int); ok {
		return v
	}
	return 0
}

// GetBool retrieves a bool value from Data by key
func (ctx *PageContext) GetBool(key string) bool {
	if v, ok := ctx.Data[key].(bool); ok {
		return v
	}
	return false
}

// GetSlice retrieves a slice value from Data by key
func (ctx *PageContext) GetSlice(key string) []any {
	if v, ok := ctx.Data[key].([]any); ok {
		return v
	}
	return nil
}

// GetMap retrieves a map value from Data by key
func (ctx *PageContext) GetMap(key string) map[string]any {
	if v, ok := ctx.Data[key].(map[string]any); ok {
		return v
	}
	return nil
}

// HasParam checks if a URL parameter exists
func (ctx *PageContext) HasParam(key string) bool {
	if ctx.Params == nil {
		return false
	}
	_, ok := ctx.Params[key]
	return ok
}

// Param retrieves a URL parameter by key
func (ctx *PageContext) Param(key string) string {
	if ctx.Params == nil {
		return ""
	}
	return ctx.Params[key]
}
