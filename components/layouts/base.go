// Package layouts provides layout components for page composition.
package layouts

import (
	"strings"

	"github.com/ZacxDev/go-static-site/components"
	c "maragu.dev/gomponents/components"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// BaseLayout is the default HTML5 layout that wraps page content.
// It includes standard head elements, CSS/JS bundle loading, and a consistent page structure.
func BaseLayout(ctx *components.PageContext, content ...g.Node) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:    ctx.Title,
		Language: ctx.Lang,
		Head: []g.Node{
			// Meta tags
			h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
			g.If(ctx.Description != "",
				h.Meta(h.Name("description"), h.Content(ctx.Description)),
			),
			h.Link(h.Rel("canonical"), h.Href(ctx.Canonical)),

			// CSS from esbuild bundles
			g.Group(g.Map(ctx.EsbuildBundlePaths, func(path string) g.Node {
				if strings.HasSuffix(path, ".css") {
					return h.Link(h.Rel("stylesheet"), h.Href(path))
				}
				return nil
			})),
		},
		Body: []g.Node{
			// Main content
			g.Group(content),

			// JavaScript from esbuild bundles (at end of body)
			g.Group(g.Map(ctx.EsbuildBundlePaths, func(path string) g.Node {
				if strings.HasSuffix(path, ".js") {
					return h.Script(h.Src(path), h.Defer())
				}
				return nil
			})),
		},
	})
}

// WithHeader wraps content with a header and optional footer.
func WithHeader(ctx *components.PageContext, header, footer g.Node, content ...g.Node) g.Node {
	bodyContent := []g.Node{header}
	bodyContent = append(bodyContent, h.Main(g.Group(content)))
	if footer != nil {
		bodyContent = append(bodyContent, footer)
	}

	return c.HTML5(c.HTML5Props{
		Title:    ctx.Title,
		Language: ctx.Lang,
		Head: []g.Node{
			h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
			g.If(ctx.Description != "",
				h.Meta(h.Name("description"), h.Content(ctx.Description)),
			),
			h.Link(h.Rel("canonical"), h.Href(ctx.Canonical)),
			g.Group(g.Map(ctx.EsbuildBundlePaths, func(path string) g.Node {
				if strings.HasSuffix(path, ".css") {
					return h.Link(h.Rel("stylesheet"), h.Href(path))
				}
				return nil
			})),
		},
		Body: append(bodyContent,
			g.Group(g.Map(ctx.EsbuildBundlePaths, func(path string) g.Node {
				if strings.HasSuffix(path, ".js") {
					return h.Script(h.Src(path), h.Defer())
				}
				return nil
			})),
		),
	})
}

// Minimal is a minimal HTML5 layout without extra structure.
// Useful for simple pages or custom layouts.
func Minimal(ctx *components.PageContext, content ...g.Node) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:    ctx.Title,
		Language: ctx.Lang,
		Head: []g.Node{
			h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
		},
		Body: content,
	})
}
