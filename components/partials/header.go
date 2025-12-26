// Package partials provides reusable UI component partials.
package partials

import (
	"github.com/ZacxDev/go-static-site/components"
	c "maragu.dev/gomponents/components"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// SiteHeader renders a basic site header with navigation.
func SiteHeader(ctx *components.PageContext) g.Node {
	return h.Header(
		h.Nav(h.Class("main-nav"),
			h.A(h.Href("/"), g.Text(ctx.Text("nav.home"))),
		),
	)
}

// SiteHeaderWithNav renders a header with custom navigation links.
func SiteHeaderWithNav(ctx *components.PageContext, links []NavLink) g.Node {
	return h.Header(
		h.Nav(h.Class("main-nav"),
			g.Group(g.Map(links, func(link NavLink) g.Node {
				return h.A(
					h.Href(link.Href),
					c.Classes{"active": link.Href == ctx.CurrentPath},
					g.Text(link.Text),
				)
			})),
			g.If(len(ctx.SupportedLangs) > 1,
				LanguageSwitcher(ctx),
			),
		),
	)
}

// NavLink represents a navigation link.
type NavLink struct {
	Href string
	Text string
}

// LanguageSwitcher renders a language selection component.
func LanguageSwitcher(ctx *components.PageContext) g.Node {
	return h.Div(h.Class("lang-switcher"),
		g.Group(g.Map(ctx.SupportedLangs, func(lang string) g.Node {
			// Build the localized path using PathNoLang (path without language prefix)
			href := "/" + lang + ctx.PathNoLang
			return h.A(
				h.Href(href),
				c.Classes{"active": lang == ctx.Lang},
				g.Text(lang),
			)
		})),
	)
}
