package partials

import (
	"time"

	"github.com/ZacxDev/go-static-site/components"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// SiteFooter renders a basic site footer.
func SiteFooter(ctx *components.PageContext) g.Node {
	return h.Footer(
		h.P(g.Textf("© %d", time.Now().Year())),
	)
}

// SiteFooterWithContent renders a footer with custom content.
func SiteFooterWithContent(ctx *components.PageContext, content ...g.Node) g.Node {
	return h.Footer(
		g.Group(content),
	)
}

// SiteFooterWithLinks renders a footer with navigation links.
func SiteFooterWithLinks(ctx *components.PageContext, links []NavLink) g.Node {
	return h.Footer(
		h.Nav(h.Class("footer-nav"),
			g.Group(g.Map(links, func(link NavLink) g.Node {
				return h.A(h.Href(link.Href), g.Text(link.Text))
			})),
		),
		h.P(g.Textf("© %d", time.Now().Year())),
	)
}
