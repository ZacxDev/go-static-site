# Gomponents: The Recommended Template Engine

Gomponents is the **recommended way** to build pages with go-static-site. It provides compile-time type safety, full IDE support, and superior performance compared to traditional string-based templates.

## Why Gomponents is the New Standard

### The Problem with String Templates

Traditional template engines (Plush, Jinja, Handlebars) have fundamental limitations:

```html
<!-- Plush template - errors discovered at runtime -->
<%= user.nmae %>           <!-- Typo: silent failure or runtime error -->
<%= nonExistentVar %>      <!-- No compile-time validation -->
<%= partial("headr") %>    <!-- Typo in partial name -->
```

These issues manifest as:
- **Runtime errors** in production
- **Silent failures** producing incorrect HTML
- **No IDE autocomplete** for variables or helpers
- **Refactoring blindness** - rename a variable, break templates

### The Gomponents Solution

```go
// Gomponents - errors caught at compile time
h.P(g.Text(user.Nmae))     // ❌ Compile error: user.Nmae undefined
h.P(g.Text(nonExistentVar)) // ❌ Compile error: undefined
partials.Headr(ctx)         // ❌ Compile error: undefined
```

**Benefits:**
| Aspect | Plush | Gomponents |
|--------|-------|------------|
| Error detection | Runtime | Compile-time |
| IDE support | Limited syntax highlighting | Full autocomplete, go-to-definition, refactoring |
| Performance | Parse → Execute → Render | Direct function calls |
| Testability | Difficult (requires HTTP context) | Unit test any component |
| Type safety | None (all values are `any`) | Full Go type system |
| Debugging | Template stack traces | Standard Go debugging |

### When to Use Each Template Type

| Template Type | Best For | Avoid For |
|---------------|----------|-----------|
| **GOMPONENTS** | Application pages, complex UIs, data-driven content | — |
| **MARKDOWN** | Blog posts, documentation, content-heavy pages | Logic-heavy pages |
| **PLUSH** | Legacy templates, rapid prototyping | New development |

**Recommendation:** Use Gomponents for all new pages. Use Markdown for content. Migrate Plush templates opportunistically.

---

## Quick Start

### 1. Create a Page Component

```go
// components/pages/home.go
package pages

import (
    "github.com/ZacxDev/go-static-site/components"
    "github.com/ZacxDev/go-static-site/components/layouts"
    g "maragu.dev/gomponents"
    h "maragu.dev/gomponents/html"
)

func init() {
    components.Register("home", HomePage)
}

func HomePage(ctx *components.PageContext) g.Node {
    return layouts.BaseLayout(ctx,
        h.H1(g.Text(ctx.Title)),
        h.P(g.Text("Welcome to the site")),
    )
}
```

### 2. Add Route in manifest.star

```python
route(
    path = "/",
    template_type = "GOMPONENTS",
    component_id = "home",
    page_title = "Home",
)
```

### 3. Import Pages Package

Ensure `main.go` imports the pages package:

```go
package main

import (
    "github.com/ZacxDev/go-static-site/cmd"
    _ "github.com/ZacxDev/go-static-site/components/pages"  // Register components
)

func main() {
    cmd.Execute()
}
```

That's it! The page renders using your type-safe Go component.

---

## Architecture Overview

### How Gomponents Integrates

```
manifest.star                    Go Components                 Output
┌─────────────────┐             ┌─────────────────┐           ┌──────────┐
│ route(          │             │ func HomePage() │           │          │
│   path="/",     │────────────▶│   g.Node        │──────────▶│  HTML    │
│   component_id= │  lookup     │ {               │  render   │          │
│     "home"      │             │   return ...    │           │          │
│ )               │             │ }               │           │          │
└─────────────────┘             └─────────────────┘           └──────────┘
```

**Request Flow:**

1. **Route Match** → `handlers/dynamic_handler.go` matches URL to route
2. **Component Lookup** → `components.Get(route.ComponentID)` retrieves registered function
3. **Context Build** → `PageContext` populated with all template data
4. **Render** → Component function called, returns `g.Node`
5. **Output** → Node rendered to HTML string, sent to client

**Key difference from Plush:** Gomponents routes bypass Plush layout wrapping entirely. The component handles its own complete HTML structure via layouts.

### Component Registration

Components self-register in `init()`:

```go
func init() {
    components.Register("my-page", MyPage)  // ID → Function mapping
}
```

The registry (`components/registry.go`) is a simple map:

```go
var registry = make(map[string]PageFunc)

func Register(id string, fn PageFunc) { registry[id] = fn }
func Get(id string) (PageFunc, bool)  { return registry[id] }
```

**Import trigger:** The blank import in `main.go` (`_ ".../components/pages"`) causes Go to execute all `init()` functions in that package, populating the registry before any routes are processed.

---

## PageContext: Your Data Interface

Every component receives `*components.PageContext` containing all data needed for rendering:

### Available Fields

```go
type PageContext struct {
    // URL & Routing
    Params       map[string]string  // URL parameters (:slug, :id)
    CurrentPath  string             // "/blog/hello-world"
    PathNoLang   string             // Path without lang prefix (for lang switcher)

    // Internationalization
    Lang           string              // "en", "es", etc.
    SupportedLangs []string            // ["en", "es"]
    Translations   map[string]string   // Pre-filtered for current lang

    // Page Metadata
    Title       string  // From route.page_title
    Description string  // Page description
    Canonical   string  // Full canonical URL

    // Assets
    EsbuildBundlePaths []string  // JS/CSS bundle paths from javascript_deps

    // Environment
    IsProductionEnvironment bool
    AppOrigin               string  // "https://mysite.com"

    // Custom Data
    Data map[string]any  // Merged static_render_data + global_render_context

    // Navigation Helpers
    MarkdownRoutes   []LoadedMarkdownRoute  // For blog listings
    RegisteredRoutes []string               // All site routes

    // Internal (advanced use)
    Manifest *config.SiteManifest
    Route    config.Route
}
```

### Typed Getter Methods

Avoid manual type assertions with convenience methods:

```go
// Basic getters
ctx.Get("key")           // any (nil if missing)
ctx.GetString("key")     // string ("" if missing/wrong type)
ctx.GetInt("key")        // int (0 if missing/wrong type)
ctx.GetBool("key")       // bool (false if missing/wrong type)
ctx.GetSlice("key")      // []any (nil if missing/wrong type)
ctx.GetMap("key")        // map[string]any (nil if missing/wrong type)

// Translation
ctx.Text("nav.home")     // Returns translation or key itself

// URL Parameters
ctx.Param("slug")        // Get URL parameter
ctx.HasParam("slug")     // Check if parameter exists
```

### Data Flow from Manifest

Data reaches `ctx.Data` from two sources:

```python
# manifest.star

# 1. Global data (available to ALL routes)
global_render_context = {
    "site_name": "My Site",
    "year": 2024,
}

# 2. Route-specific data
route(
    path = "/products",
    template_type = "GOMPONENTS",
    component_id = "products",
    static_render_data = {
        "products": read_json("data/products.json"),
        "featured": True,
    },
)
```

In the component, both are merged into `ctx.Data`:

```go
func ProductsPage(ctx *components.PageContext) g.Node {
    siteName := ctx.GetString("site_name")  // From global_render_context
    products := ctx.GetSlice("products")     // From static_render_data
    featured := ctx.GetBool("featured")      // From static_render_data
    // ...
}
```

---

## Layouts

Layouts provide the HTML5 document structure and handle common concerns like meta tags and asset loading.

### BaseLayout (Recommended)

Full HTML5 structure with automatic CSS/JS bundle inclusion:

```go
func MyPage(ctx *components.PageContext) g.Node {
    return layouts.BaseLayout(ctx,
        // Your content here - becomes the <body> content
        h.Main(
            h.H1(g.Text(ctx.Title)),
            h.P(g.Text("Page content")),
        ),
    )
}
```

**What BaseLayout provides:**
- `<!doctype html>` and `<html lang="...">`
- `<head>` with title, viewport meta, description, canonical link
- CSS `<link>` tags for all `.css` files in `EsbuildBundlePaths`
- JS `<script defer>` tags for all `.js` files (at end of body)

### WithHeader (Full Page Structure)

For pages with header/footer:

```go
func MyPage(ctx *components.PageContext) g.Node {
    return layouts.WithHeader(ctx,
        myHeader(ctx),   // Header component
        myFooter(ctx),   // Footer component (can be nil)
        // Content wrapped in <main>
        h.Section(h.Class("hero"),
            h.H1(g.Text(ctx.Title)),
        ),
        h.Section(h.Class("content"),
            h.P(g.Text("More content")),
        ),
    )
}
```

### Minimal (Bare Structure)

For custom HTML structures:

```go
func CustomPage(ctx *components.PageContext) g.Node {
    return layouts.Minimal(ctx,
        // Just your content, minimal head
        h.Div(g.Text("Custom structure")),
    )
}
```

### Custom Layouts

Create project-specific layouts in `components/layouts/`:

```go
// components/layouts/marketing.go
package layouts

func MarketingLayout(ctx *components.PageContext, content ...g.Node) g.Node {
    return c.HTML5(c.HTML5Props{
        Title:    ctx.Title + " | My Brand",
        Language: ctx.Lang,
        Head: []g.Node{
            // Custom meta tags, fonts, etc.
            h.Link(h.Rel("stylesheet"), h.Href("/marketing.css")),
        },
        Body: append([]g.Node{
            marketingHeader(ctx),
        }, append(content, marketingFooter(ctx))...),
    })
}
```

---

## Building Components

### Standard Import Aliases

```go
import (
    g "maragu.dev/gomponents"            // Core: g.Text, g.If, g.Map, g.Raw
    h "maragu.dev/gomponents/html"       // Elements: h.Div, h.H1, h.A, h.Class
    c "maragu.dev/gomponents/components" // Utilities: c.Classes, c.HTML5
)
```

### Element Composition

Gomponents uses function composition - elements contain other elements:

```go
h.Div(                          // <div>
    h.Class("card"),            //   class="card"
    h.H2(                       //   <h2>
        g.Text("Title"),        //     Title
    ),                          //   </h2>
    h.P(                        //   <p>
        g.Text("Content"),      //     Content
    ),                          //   </p>
)                               // </div>
```

### Attributes

```go
// Class attribute
h.Div(h.Class("container"))

// Multiple classes
h.Div(h.Class("btn btn-primary"))

// Other attributes
h.A(h.Href("/about"), h.Title("About page"), g.Text("About"))
h.Img(h.Src("/logo.png"), h.Alt("Logo"))
h.Input(h.Type("email"), h.Name("email"), h.Placeholder("you@example.com"))

// Data attributes
h.Div(g.Attr("data-id", "123"))

// Boolean attributes
h.Input(h.Type("checkbox"), h.Checked())
h.Button(h.Disabled())
```

### Dynamic Classes

```go
import c "maragu.dev/gomponents/components"

h.Div(
    c.Classes{
        "nav-link": true,
        "active":   isCurrentPage,
        "disabled": !hasPermission,
    },
    g.Text("Link"),
)
```

### Conditional Rendering

```go
// Simple condition
g.If(ctx.IsProductionEnvironment,
    h.Script(h.Src("/analytics.js")),
)

// If-else pattern
g.Group([]g.Node{
    g.If(isLoggedIn, h.A(h.Href("/profile"), g.Text("Profile"))),
    g.If(!isLoggedIn, h.A(h.Href("/login"), g.Text("Login"))),
})

// Lazy evaluation (for expensive operations)
helpers.Iff(shouldLoad, func() g.Node {
    return expensiveComponent()
})
```

### Loops and Lists

```go
// Map slice to nodes
h.Ul(
    g.Map(items, func(item Item) g.Node {
        return h.Li(g.Text(item.Name))
    })...,  // Note: spread operator
)

// With index
h.Ol(
    g.Map(items, func(item Item) g.Node {
        return h.Li(g.Textf("%s", item.Name))
    })...,
)

// From ctx.Data (requires type assertion)
products := ctx.GetSlice("products")
h.Div(
    g.Map(products, func(p any) g.Node {
        product := p.(map[string]any)
        return h.Div(
            h.H3(g.Text(product["name"].(string))),
            h.P(g.Text(product["description"].(string))),
        )
    })...,
)
```

### Grouping Nodes

```go
// Combine multiple nodes without a wrapper element
g.Group([]g.Node{
    h.H1(g.Text("Title")),
    h.P(g.Text("Paragraph 1")),
    h.P(g.Text("Paragraph 2")),
})
```

### Raw HTML

```go
// When you need unescaped HTML
g.Raw("<strong>Bold</strong>")

// Markdown to HTML
helpers.MarkdownNode(markdownContent)
```

---

## Partials (Reusable Components)

Create reusable components as functions in `components/partials/`:

### Simple Partial

```go
// components/partials/header.go
package partials

func SiteHeader(ctx *components.PageContext) g.Node {
    return h.Header(
        h.Nav(h.Class("main-nav"),
            h.A(h.Href("/"), g.Text(ctx.Text("nav.home"))),
            h.A(h.Href("/about"), g.Text(ctx.Text("nav.about"))),
        ),
    )
}
```

### Partial with Parameters

```go
// components/partials/card.go
package partials

type CardProps struct {
    Title       string
    Description string
    ImageURL    string
    LinkURL     string
}

func Card(props CardProps) g.Node {
    return h.Article(h.Class("card"),
        g.If(props.ImageURL != "",
            h.Img(h.Src(props.ImageURL), h.Alt(props.Title)),
        ),
        h.H3(g.Text(props.Title)),
        h.P(g.Text(props.Description)),
        g.If(props.LinkURL != "",
            h.A(h.Href(props.LinkURL), g.Text("Learn more")),
        ),
    )
}
```

Usage:

```go
partials.Card(partials.CardProps{
    Title:       "My Product",
    Description: "A great product",
    ImageURL:    "/images/product.jpg",
    LinkURL:     "/products/my-product",
})
```

### Navigation with Active State

```go
// components/partials/nav.go
package partials

type NavLink struct {
    Href string
    Text string
}

func Navigation(ctx *components.PageContext, links []NavLink) g.Node {
    return h.Nav(h.Class("main-nav"),
        g.Map(links, func(link NavLink) g.Node {
            isActive := ctx.CurrentPath == link.Href
            return h.A(
                h.Href(link.Href),
                c.Classes{"active": isActive},
                g.Text(link.Text),
            )
        })...,
    )
}
```

### Language Switcher

```go
func LanguageSwitcher(ctx *components.PageContext) g.Node {
    return h.Div(h.Class("lang-switcher"),
        g.Map(ctx.SupportedLangs, func(lang string) g.Node {
            href := "/" + lang + ctx.PathNoLang
            return h.A(
                h.Href(href),
                c.Classes{"active": lang == ctx.Lang},
                g.Text(lang),
            )
        })...,
    )
}
```

---

## Helper Functions

Available in `components/helpers` - mirrors Plush helpers for easy migration:

### String Helpers

```go
helpers.StartsWith(s, prefix)      // strings.HasPrefix
helpers.EndsWith(s, suffix)        // strings.HasSuffix
helpers.Contains(s, sub)           // strings.Contains
helpers.Matches(s, pattern)        // regex match
helpers.Replace(s, old, new)       // first occurrence
helpers.ReplaceAll(s, old, new)    // all occurrences
helpers.ReplacePattern(s, pat, r)  // regex replace
helpers.Upper(s)                   // uppercase
helpers.Lower(s)                   // lowercase
helpers.Truncate(s, max)           // truncate with "..."
```

### HTML/Content Helpers

```go
helpers.Markdown(input)        // markdown → template.HTML
helpers.MarkdownNode(input)    // markdown → g.Node
helpers.HTML(input)            // raw HTML (unescaped)
helpers.Raw(html)              // raw HTML as g.Node
helpers.EscapeString(input)    // escape HTML
helpers.UnescapeString(input)  // unescape HTML entities
```

### Encoding Helpers

```go
helpers.URLEncode(input)       // percent-encode
helpers.URLDecode(input)       // decode percent-encoded
helpers.Stringify(data)        // JSON marshal
helpers.StringifyPretty(data)  // JSON marshal (formatted)
```

### Gomponents Helpers

```go
helpers.Text(s)                // g.Text (HTML-escaped)
helpers.Textf(format, args...) // formatted text
helpers.If(condition, node)    // conditional render
helpers.Iff(condition, fn)     // lazy conditional
helpers.Map(items, fn)         // slice → []g.Node
helpers.Group(nodes)           // combine nodes
```

---

## Complete Example: Blog Post Page

```go
// components/pages/blog_post.go
package pages

import (
    "github.com/ZacxDev/go-static-site/components"
    "github.com/ZacxDev/go-static-site/components/helpers"
    "github.com/ZacxDev/go-static-site/components/layouts"
    "github.com/ZacxDev/go-static-site/components/partials"
    c "maragu.dev/gomponents/components"
    g "maragu.dev/gomponents"
    h "maragu.dev/gomponents/html"
)

func init() {
    components.Register("blog_post", BlogPostPage)
}

func BlogPostPage(ctx *components.PageContext) g.Node {
    post := ctx.GetMap("post")
    if post == nil {
        return layouts.BaseLayout(ctx, notFound(ctx))
    }

    title := post["title"].(string)
    content := post["content"].(string)
    author := post["author"].(string)
    date := post["date"].(string)
    tags := post["tags"].([]any)

    return layouts.WithHeader(ctx,
        partials.SiteHeader(ctx),
        partials.SiteFooter(ctx),

        h.Article(h.Class("blog-post"),
            // Header
            h.Header(h.Class("post-header"),
                h.H1(g.Text(title)),
                h.Div(h.Class("post-meta"),
                    h.Span(g.Textf("By %s", author)),
                    h.Time(g.Text(date)),
                ),
                tagList(tags),
            ),

            // Content (markdown rendered to HTML)
            h.Div(h.Class("post-content"),
                helpers.MarkdownNode(content),
            ),

            // Footer
            h.Footer(h.Class("post-footer"),
                shareButtons(ctx, title),
                relatedPosts(ctx),
            ),
        ),
    )
}

func tagList(tags []any) g.Node {
    if len(tags) == 0 {
        return nil
    }
    return h.Div(h.Class("tags"),
        g.Map(tags, func(t any) g.Node {
            tag := t.(string)
            return h.A(
                h.Href("/blog/tag/"+tag),
                h.Class("tag"),
                g.Text(tag),
            )
        })...,
    )
}

func shareButtons(ctx *components.PageContext, title string) g.Node {
    url := ctx.Canonical
    return h.Div(h.Class("share-buttons"),
        h.A(
            h.Href("https://twitter.com/share?url="+helpers.URLEncode(url)),
            h.Target("_blank"),
            g.Text("Share on Twitter"),
        ),
    )
}

func relatedPosts(ctx *components.PageContext) g.Node {
    related := ctx.GetSlice("related_posts")
    if related == nil || len(related) == 0 {
        return nil
    }
    return h.Section(h.Class("related-posts"),
        h.H3(g.Text("Related Posts")),
        h.Ul(
            g.Map(related, func(p any) g.Node {
                post := p.(map[string]any)
                return h.Li(
                    h.A(
                        h.Href(post["url"].(string)),
                        g.Text(post["title"].(string)),
                    ),
                )
            })...,
        ),
    )
}

func notFound(ctx *components.PageContext) g.Node {
    return h.Div(h.Class("not-found"),
        h.H1(g.Text("Post Not Found")),
        h.P(g.Text("The requested blog post could not be found.")),
        h.A(h.Href("/blog"), g.Text("Back to Blog")),
    )
}
```

---

## Project Structure

```
your-project/
├── main.go                         # Imports components/pages
├── manifest.star                   # Routes with component_id references
├── components/
│   ├── registry.go                 # Register/Get functions
│   ├── context.go                  # PageContext definition
│   ├── helpers/
│   │   └── helpers.go              # Utility functions
│   ├── layouts/
│   │   └── base.go                 # BaseLayout, WithHeader, Minimal
│   ├── partials/
│   │   ├── header.go               # Site header components
│   │   ├── footer.go               # Site footer components
│   │   └── [your-partials].go      # Reusable UI components
│   └── pages/
│       ├── init.go                 # Package documentation
│       ├── home.go                 # HomePage component
│       ├── about.go                # AboutPage component
│       └── [your-pages].go         # Your page components
├── data/                           # JSON data files
├── static/                         # Static assets
└── translations/                   # i18n files
```

---

## Migration from Plush

### Step-by-Step Migration

1. **Create component file** in `components/pages/`
2. **Register with same route path** using `component_id`
3. **Translate template syntax:**

| Plush | Gomponents |
|-------|------------|
| `<%= title %>` | `g.Text(ctx.Title)` |
| `<%= text("key") %>` | `g.Text(ctx.Text("key"))` |
| `<% if (x) { %> ... <% } %>` | `g.If(x, ...)` |
| `<% for (item) in items { %>` | `g.Map(items, func(item) g.Node { ... })` |
| `<%= partial("path") %>` | `partials.MyPartial(ctx)` |
| `<%= yield %>` | N/A (use layout composition) |
| `<%== rawHtml %>` | `g.Raw(rawHtml)` |

4. **Update manifest route:**

```python
# Before (Plush)
route(
    path = "/about",
    source = "pages/about.plush.html",
    template_type = "PLUSH",
    page_title = "About",
)

# After (Gomponents)
route(
    path = "/about",
    template_type = "GOMPONENTS",
    component_id = "about",
    page_title = "About",
)
```

5. **Remove old template file** once component works

### Gradual Migration Strategy

You can run Plush and Gomponents side-by-side:

```python
routes = [
    # Migrated to Gomponents
    route(path="/", template_type="GOMPONENTS", component_id="home"),
    route(path="/about", template_type="GOMPONENTS", component_id="about"),

    # Still using Plush (migrate later)
    route(path="/legacy", source="pages/legacy.plush.html", template_type="PLUSH"),

    # Content stays in Markdown
    route(path="/blog/:slug", source="pages/blog/[slug].md", template_type="MARKDOWN"),
]
```

---

## Testing Components

Gomponents are just functions - test them like any Go code:

```go
// components/pages/home_test.go
package pages

import (
    "strings"
    "testing"

    "github.com/ZacxDev/go-static-site/components"
)

func TestHomePage(t *testing.T) {
    ctx := &components.PageContext{
        Title: "Test Home",
        Lang:  "en",
        Data: map[string]any{
            "site_name": "Test Site",
        },
    }

    node := HomePage(ctx)

    var buf strings.Builder
    err := node.Render(&buf)
    if err != nil {
        t.Fatalf("render failed: %v", err)
    }

    html := buf.String()

    if !strings.Contains(html, "Test Home") {
        t.Error("expected title in output")
    }
    if !strings.Contains(html, "<h1>") {
        t.Error("expected h1 element")
    }
}
```

---

## Performance Considerations

### Gomponents Performance Advantages

1. **No template parsing** - Components compile to direct function calls
2. **No reflection** - Strong typing eliminates runtime type checks
3. **Efficient string building** - `io.Writer` interface avoids allocations
4. **Parallel rendering** - Build process renders pages concurrently

### Best Practices

```go
// ✅ Good: Define static content once
var navLinks = []NavLink{
    {Href: "/", Text: "Home"},
    {Href: "/about", Text: "About"},
}

func Header(ctx *components.PageContext) g.Node {
    return Navigation(ctx, navLinks)
}

// ❌ Avoid: Recreating slices on every render
func Header(ctx *components.PageContext) g.Node {
    links := []NavLink{...}  // Allocates every call
    return Navigation(ctx, links)
}
```

---

## Common Patterns

### SEO Meta Tags

```go
func seoHead(ctx *components.PageContext, ogImage string) []g.Node {
    return []g.Node{
        h.Meta(h.Name("description"), h.Content(ctx.Description)),
        h.Meta(h.Name("robots"), h.Content("index, follow")),

        // Open Graph
        h.Meta(g.Attr("property", "og:title"), h.Content(ctx.Title)),
        h.Meta(g.Attr("property", "og:description"), h.Content(ctx.Description)),
        h.Meta(g.Attr("property", "og:url"), h.Content(ctx.Canonical)),
        g.If(ogImage != "",
            h.Meta(g.Attr("property", "og:image"), h.Content(ogImage)),
        ),

        // Twitter Card
        h.Meta(h.Name("twitter:card"), h.Content("summary_large_image")),
    }
}
```

### Forms

```go
func contactForm(ctx *components.PageContext) g.Node {
    return h.Form(
        h.Action("/api/contact"),
        h.Method("POST"),
        h.Class("contact-form"),

        formField("name", "Name", "text", true),
        formField("email", "Email", "email", true),

        h.Div(h.Class("field"),
            h.Label(h.For("message"), g.Text("Message")),
            h.Textarea(
                h.Name("message"),
                h.ID("message"),
                h.Required(),
                h.Rows("5"),
            ),
        ),

        h.Button(h.Type("submit"), g.Text(ctx.Text("form.submit"))),
    )
}

func formField(name, label, inputType string, required bool) g.Node {
    return h.Div(h.Class("field"),
        h.Label(h.For(name), g.Text(label)),
        h.Input(
            h.Type(inputType),
            h.Name(name),
            h.ID(name),
            g.If(required, h.Required()),
        ),
    )
}
```

### Responsive Images

```go
func responsiveImage(src, alt string, widths []int) g.Node {
    srcset := make([]string, len(widths))
    for i, w := range widths {
        srcset[i] = fmt.Sprintf("%s?w=%d %dw", src, w, w)
    }

    return h.Img(
        h.Src(src),
        h.Alt(alt),
        g.Attr("srcset", strings.Join(srcset, ", ")),
        g.Attr("sizes", "(max-width: 768px) 100vw, 50vw"),
        h.Loading("lazy"),
    )
}
```

---

## Troubleshooting

### "component not found: xyz"

1. Ensure component is registered in `init()`:
   ```go
   func init() {
       components.Register("xyz", XyzPage)
   }
   ```
2. Ensure `main.go` imports the pages package:
   ```go
   import _ "github.com/ZacxDev/go-static-site/components/pages"
   ```
3. Check `component_id` in manifest matches registration ID exactly

### Type assertion panics

Data from `ctx.Data` is `map[string]any`. Always handle missing/wrong types:

```go
// ❌ Panics if key missing or wrong type
name := ctx.Data["name"].(string)

// ✅ Safe access with typed getters
name := ctx.GetString("name")  // Returns "" if missing

// ✅ Or explicit check
if name, ok := ctx.Data["name"].(string); ok {
    // use name
}
```

### Empty output

1. Ensure layout returns complete HTML structure
2. Check component returns non-nil `g.Node`
3. Verify route has `template_type = "GOMPONENTS"`

---

## Resources

- [Gomponents Documentation](https://www.gomponents.com/)
- [Gomponents GitHub](https://github.com/maragudk/gomponents)
- [Proving Site Example](proving-site/) - Working examples in this repo
