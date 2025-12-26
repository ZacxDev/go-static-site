package pages

import (
	"github.com/ZacxDev/go-static-site/components"
	"github.com/ZacxDev/go-static-site/components/layouts"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func init() {
	components.Register("proving_home", ProvingHome)
}

// ProvingHome is the gomponents version of the home page
func ProvingHome(ctx *components.PageContext) g.Node {
	return layouts.BaseLayout(ctx,
		// Navigation
		provingNav(ctx),

		// Hero section
		h.Section(h.Class("hero"),
			h.H1(g.Text(ctx.Text("home.title"))),
			h.P(h.Class("lead"), g.Text(ctx.Text("home.subtitle"))),
		),

		// Template info
		h.Section(h.Class("template-demo"),
			h.H2(g.Text("Gomponents Template")),
			h.P(
				g.Text("This page is rendered using "),
				h.Strong(g.Text("Gomponents")),
				g.Text(" - type-safe Go components."),
			),
			h.P(
				g.Text("Current path: "),
				h.Code(g.Text(ctx.CurrentPath)),
			),
			h.P(
				g.Text("Language: "),
				h.Code(g.Text(ctx.Lang)),
			),
		),

		// Links to other pages
		h.Section(h.Class("links"),
			h.H3(g.Text("Explore Template Types:")),
			h.Ul(
				h.Li(h.A(h.Href("/"), g.Text("Plush Home"))),
				h.Li(h.A(h.Href("/about"), g.Text("Plush About"))),
				h.Li(h.A(h.Href("/blog/hello-world"), g.Text("Markdown Blog Post"))),
				h.Li(h.A(h.Href("/gom/features"), g.Text("Gomponents Features"))),
				h.Li(h.A(h.Href("/gom/data"), g.Text("Gomponents Data Demo"))),
			),
		),

		// Footer
		provingFooter(ctx),
	)
}

// provingNav renders the navigation for proving site gomponents pages
func provingNav(ctx *components.PageContext) g.Node {
	return h.Header(
		h.Nav(h.Class("main-nav"),
			h.A(h.Href("/"), g.Text(ctx.Text("nav.home"))),
			h.A(h.Href("/about"), g.Text(ctx.Text("nav.about"))),
			h.A(h.Href("/blog/hello-world"), g.Text(ctx.Text("nav.blog"))),
			h.Span(h.Class("separator"), g.Text("|")),
			navLink(ctx, "/gom", "Gomponents"),
			navLink(ctx, "/gom/features", "Features"),
			navLink(ctx, "/gom/data", "Data"),
		),
		h.Div(h.Class("lang-switcher"),
			g.Group(g.Map(ctx.SupportedLangs, func(lang string) g.Node {
				href := "/" + lang + ctx.PathNoLang
				return h.A(
					h.Href(href),
					g.If(lang == ctx.Lang, h.Class("active")),
					g.Text(lang),
				)
			})),
		),
	)
}

// navLink creates a navigation link with active state
func navLink(ctx *components.PageContext, href, text string) g.Node {
	isActive := ctx.CurrentPath == href
	return h.A(
		h.Href(href),
		g.If(isActive, h.Class("active")),
		g.Text(text),
	)
}

// provingFooter renders the footer for proving site
func provingFooter(ctx *components.PageContext) g.Node {
	siteName := ctx.GetString("site_name")
	if siteName == "" {
		siteName = "Proving Site"
	}
	return h.Footer(
		h.P(g.Textf("© %d %s. Built with go-static-site.", 2024, siteName)),
		h.P(g.Textf("Template: Gomponents | Lang: %s", ctx.Lang)),
	)
}
