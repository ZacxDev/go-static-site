package pages

import (
	"fmt"

	"github.com/ZacxDev/go-static-site/components"
	"github.com/ZacxDev/go-static-site/components/layouts"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func init() {
	components.Register("proving_data", ProvingData)
}

// ProvingData demonstrates complex data handling with gomponents
func ProvingData(ctx *components.PageContext) g.Node {
	return layouts.BaseLayout(ctx,
		provingNav(ctx),

		h.Section(h.Class("page-header"),
			h.H1(g.Text(ctx.Title)),
			h.P(h.Class("lead"), g.Text("Demonstrating data handling in gomponents")),
		),

		// Stats dashboard
		h.Section(h.Class("stats-section"),
			h.H2(g.Text("Dashboard Stats")),
			statsCards(ctx),
		),

		// Users table
		h.Section(h.Class("users-section"),
			h.H2(g.Text("Users Table")),
			usersTable(ctx),
		),

		// Raw data display
		h.Section(h.Class("raw-data"),
			h.H2(g.Text("Raw Context Data")),
			h.P(g.Text("Data available in ctx.Data:")),
			h.Ul(
				h.Li(g.Textf("Title: %s", ctx.Title)),
				h.Li(g.Textf("Lang: %s", ctx.Lang)),
				h.Li(g.Textf("Path: %s", ctx.CurrentPath)),
				h.Li(g.Textf("Canonical: %s", ctx.Canonical)),
				h.Li(g.Textf("Production: %v", ctx.IsProductionEnvironment)),
			),
		),

		// PageContext methods demo
		h.Section(h.Class("methods-demo"),
			h.H2(g.Text("PageContext Methods")),
			h.P(g.Text("Gomponents pages have access to typed getter methods:")),
			h.Pre(
				h.Code(g.Text(`// Get data with type safety
users := ctx.GetSlice("users")
stats := ctx.GetMap("stats")
name := ctx.GetString("site_name")

// Translation helper
title := ctx.Text("home.title")

// URL parameters
slug := ctx.Param("slug")
`)),
			),
		),

		provingFooter(ctx),
	)
}

// statsCards renders dashboard stat cards
func statsCards(ctx *components.PageContext) g.Node {
	stats := ctx.GetMap("stats")
	if stats == nil {
		return h.P(g.Text("No stats available"))
	}

	return h.Div(h.Class("stats-grid"),
		statCard("Total Users", fmt.Sprintf("%v", stats["total_users"])),
		statCard("Active Today", fmt.Sprintf("%v", stats["active_today"])),
	)
}

// statCard renders a single stat card
func statCard(label, value string) g.Node {
	return h.Div(h.Class("stat-card"),
		h.Div(h.Class("stat-value"), g.Text(value)),
		h.Div(h.Class("stat-label"), g.Text(label)),
	)
}

// usersTable renders users in a table
func usersTable(ctx *components.PageContext) g.Node {
	users := ctx.GetSlice("users")
	if users == nil {
		return h.P(g.Text("No users data available"))
	}

	return h.Table(h.Class("data-table"),
		h.THead(
			h.Tr(
				h.Th(g.Text("ID")),
				h.Th(g.Text("Name")),
				h.Th(g.Text("Role")),
			),
		),
		h.TBody(
			g.Group(g.Map(users, func(u any) g.Node {
				user, ok := u.(map[string]any)
				if !ok {
					return nil
				}
				return h.Tr(
					h.Td(g.Textf("%v", user["id"])),
					h.Td(g.Text(user["name"].(string))),
					h.Td(
						h.Span(
							h.Class("role-badge"),
							g.If(user["role"] == "Admin", h.Class("role-admin")),
							g.Text(user["role"].(string)),
						),
					),
				)
			})),
		),
	)
}
