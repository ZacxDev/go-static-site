package pages

import (
	"github.com/ZacxDev/go-static-site/components"
	"github.com/ZacxDev/go-static-site/components/layouts"
	"github.com/ZacxDev/go-static-site/components/partials"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func init() {
	components.Register("example", ExamplePage)
}

// ExamplePage demonstrates a simple gomponents page.
// Use it as a template for creating new pages.
//
// In manifest.star:
//
//	route(
//	    path = "/example",
//	    template_type = "GOMPONENTS",
//	    component_id = "example",
//	    page_title = "Example Page",
//	)
func ExamplePage(ctx *components.PageContext) g.Node {
	return layouts.WithHeader(ctx,
		partials.SiteHeader(ctx),
		partials.SiteFooter(ctx),
		// Page content
		h.Section(h.Class("hero"),
			h.H1(g.Text(ctx.Title)),
			h.P(g.Text("This is an example gomponents page.")),
		),
		h.Section(h.Class("content"),
			h.H2(g.Text("Features")),
			h.Ul(
				h.Li(g.Text("Type-safe HTML generation")),
				h.Li(g.Text("Compile-time error checking")),
				h.Li(g.Text("Component composition")),
				h.Li(g.Text("No template parsing overhead")),
			),
		),
		// Show some context data
		g.If(ctx.Lang != "",
			h.P(g.Textf("Current language: %s", ctx.Lang)),
		),
	)
}
