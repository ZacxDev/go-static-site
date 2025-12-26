package pages

import (
	"github.com/ZacxDev/go-static-site/components"
	"github.com/ZacxDev/go-static-site/components/layouts"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func init() {
	components.Register("proving_features", ProvingFeatures)
}

// ProvingFeatures demonstrates gomponents with static_render_data
func ProvingFeatures(ctx *components.PageContext) g.Node {
	return layouts.BaseLayout(ctx,
		provingNav(ctx),

		h.Section(h.Class("page-header"),
			h.H1(g.Text(ctx.Title)),
			h.P(h.Class("lead"), g.Text("Demonstrating gomponents features with data from manifest.star")),
		),

		// Features from static_render_data
		h.Section(h.Class("features"),
			h.H2(g.Text("Features")),
			featuresGrid(ctx),
		),

		// Code example
		h.Section(h.Class("code-example"),
			h.H2(g.Text("How This Page Works")),
			h.P(g.Text("This page receives data from manifest.star via static_render_data:")),
			h.Pre(
				h.Code(g.Text(`route(
    path = "/gom/features",
    template_type = "GOMPONENTS",
    component_id = "proving_features",
    static_render_data = {
        "features": [
            {"name": "Type Safety", "desc": "..."},
            ...
        ],
    },
)`)),
			),
			h.P(g.Text("The component accesses this data via ctx.GetSlice(\"features\").")),
		),

		// Gomponents benefits
		h.Section(h.Class("benefits"),
			h.H2(g.Text("Why Gomponents?")),
			h.Dl(
				definitionItem("Type Safety", "Errors are caught at compile time, not runtime."),
				definitionItem("IDE Support", "Full autocomplete, go-to-definition, and refactoring."),
				definitionItem("Performance", "No template parsing - components compile to direct function calls."),
				definitionItem("Testability", "Components are just functions - easy to unit test."),
				definitionItem("Composition", "Build complex UIs by composing simple functions."),
			),
		),

		provingFooter(ctx),
	)
}

// featuresGrid renders features from static_render_data
func featuresGrid(ctx *components.PageContext) g.Node {
	features := ctx.GetSlice("features")
	if features == nil {
		return h.P(g.Text("No features data available"))
	}

	return h.Div(h.Class("feature-grid"),
		g.Group(g.Map(features, func(f any) g.Node {
			feature, ok := f.(map[string]any)
			if !ok {
				return nil
			}
			name, _ := feature["name"].(string)
			desc, _ := feature["desc"].(string)
			return featureCard(name, desc)
		})),
	)
}

// featureCard renders a single feature card
func featureCard(name, desc string) g.Node {
	return h.Div(h.Class("feature-card"),
		h.H3(g.Text(name)),
		h.P(g.Text(desc)),
	)
}

// definitionItem creates a dt/dd pair
func definitionItem(term, definition string) g.Node {
	return g.Group([]g.Node{
		h.Dt(g.Text(term)),
		h.Dd(g.Text(definition)),
	})
}
