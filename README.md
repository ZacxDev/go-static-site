# Go Static Site Generator

A declarative static site generator that puts configuration first. Define your entire site structure, routes, templates, JavaScript bundles, and translations in a single manifest file.

## What is Declarative Static Site Generation?

Unlike traditional static site generators that rely on magic directory layouts and implicit file naming conventions, Go Static Site Generator embraces **declarative configuration**. You explicitly define:

- **Routes**: What URLs your site responds to
- **Templates**: How content gets rendered (Plush templates or Markdown)
- **Assets**: JavaScript/CSS bundles and their dependencies
- **Translations**: Multi-language support with automatic route generation
- **Partials**: Reusable template components
- **Global Context**: Data available to all templates

This approach provides several benefits:
- **Explicit over implicit**: No guessing what files go where
- **Flexible structure**: Organize files however makes sense for your project
- **Type safety**: Starlark configuration provides better validation than YAML
- **Programmatic control**: Use functions and variables in your configuration
- **Single source of truth**: Everything about your site is defined in one place

## Configuration Formats

You can configure your site using either:
- **YAML** (`manifest.yaml`) - Simple and familiar
- **Starlark** (`manifest.star`) - Programmable configuration with functions and imports

## Starlark Configuration

Starlark is a Python-like configuration language that allows you to use functions, variables, loops, and imports in your site configuration. This makes it powerful for complex sites while remaining simple for basic use cases.

### Basic Starlark Example

```python
# manifest.star

# Site configuration
app_origin = "https://mysite.com"
default_layout_source = "templates/layouts/base.plush.html"
not_found_page_source = "pages/404.plush.html"
is_production_environment = False

# Load data from external files
blog_posts = read_yaml("data/blog-posts.yaml")
site_metadata = read_json("data/site.json")

# Define reusable partials
partials = {
    "header": partial(
        source="templates/partials/header.plush.html",
        template_type="PLUSH"
    ),
    "footer": partial(
        source="templates/partials/footer.plush.html",
        template_type="PLUSH"
    ),
    "blog_card": partial(
        source="templates/partials/blog-card.plush.html",
        template_type="PLUSH"
    )
}

# JavaScript/CSS targets
javascript_targets = {
    "main": js_target(
        source="src/main.ts",
        out_dir="static/js"
    ),
    "blog": js_target(
        source="src/blog.ts",
        out_dir="static/js"
    )
}

css_targets = {
    "styles": js_target(  # CSS uses same structure as JS
        source="src/styles.css",
        out_dir="static/css"
    )
}

# Translation configuration
translations = [
    translation(
        code="en",
        source="translations/en.yaml",
        source_type="YAML",
        is_default=True
    ),
    translation(
        code="es",
        source="translations/es.yaml",
        source_type="YAML"
    ),
    translation(
        code="fr",
        source="translations/fr.yaml",
        source_type="YAML"
    )
]

# Define routes
routes = [
    # Homepage
    route(
        path="/",
        source="pages/home.plush.html",
        template_type="PLUSH",
        page_title="Welcome to My Site",
        partial_deps=["header", "footer"],
        javascript_deps=["main"],
        static_render_data={
            "featured_posts": blog_posts["featured"],
            "site_name": site_metadata["name"]
        }
    ),

    # About page
    route(
        path="/about",
        source="pages/about.md",
        template_type="MARKDOWN",
        partial_deps=["header", "footer"],
        javascript_deps=["main"]
    ),

    # Dynamic blog post routes
    route(
        path="/blog/:slug",
        source="pages/blog/[slug]/[lang].md",
        template_type="MARKDOWN",
        partial_deps=["header", "footer", "blog_card"],
        javascript_deps=["blog"],
        sitemap_video_data={
            "title": "Blog Post Video",
            "description": "Video description",
            "thumbnail_loc": "/static/images/video-thumb.jpg",
            "content_loc": "/static/videos/blog-video.mp4",
            "duration": 180,
            "publication_date": "2024-01-15"
        }
    ),

    # Contact page with form handling
    route(
        path="/contact",
        source="pages/contact.plush.html",
        template_type="PLUSH",
        page_title="Contact Us",
        partial_deps=["header", "footer"],
        javascript_deps=["main"],
        static_render_data={
            "contact_email": site_metadata["contact"]["email"],
            "office_address": site_metadata["contact"]["address"]
        }
    )
]

# Global context available to all templates
global_render_context = {
    "site_name": site_metadata["name"],
    "site_description": site_metadata["description"],
    "social_links": site_metadata["social"],
    "current_year": 2024,
    "analytics_id": site_metadata["analytics_id"],
    "blog_posts_count": len(blog_posts["all"])
}

# Enable SPA mode for client-side routing
enable_spa_mode = False
```

### Advanced Starlark Features

#### Using External Data Files

```python
# manifest.star

# Load and process external data
products = read_yaml("data/products.yaml")
team_members = read_json("data/team.json")

# Process and transform data
featured_products = [p for p in products if p.get("featured", False)]

# Use in route configuration
routes = [
    route(
        path="/products",
        source="pages/products.plush.html",
        template_type="PLUSH",
        static_render_data={
            "all_products": products,
            "featured_products": featured_products,
            "product_categories": list(set([p["category"] for p in products]))
        }
    )
]
```

#### Modular Configuration with Imports

Create reusable configuration modules:

```python
# config/routes.star
def create_blog_routes(posts_data):
    blog_routes = []
    for post in posts_data:
        blog_routes.append(route(
            path="/blog/" + post["slug"],
            source="pages/blog/" + post["slug"] + "/[lang].md",
            template_type="MARKDOWN",
            page_title=post["title"],
            partial_deps=["header", "footer", "blog_card"],
            javascript_deps=["blog"]
        ))
    return blog_routes

def create_product_routes(products):
    return [
        route(
            path="/products/" + product["id"],
            source="pages/products/product.plush.html",
            template_type="PLUSH",
            static_render_data={"product": product}
        )
        for product in products
    ]
```

```python
# manifest.star
load("config/routes.star", "create_blog_routes", "create_product_routes")

blog_data = read_yaml("data/blog.yaml")
products = read_yaml("data/products.yaml")

routes = [
    # Static routes
    route(path="/", source="pages/home.plush.html", template_type="PLUSH"),
    route(path="/about", source="pages/about.md", template_type="MARKDOWN"),
] + create_blog_routes(blog_data["posts"]) + create_product_routes(products)
```

#### Conditional Configuration

```python
# manifest.star

# Environment-based configuration
is_production = read_json("package.json")["config"]["environment"] == "production"

app_origin = "https://mysite.com" if is_production else "http://localhost:3000"
is_production_environment = is_production

# Conditional JavaScript bundles
javascript_targets = {
    "main": js_target(source="src/main.ts", out_dir="static/js")
}

if is_production:
    javascript_targets["analytics"] = js_target(
        source="src/analytics.ts",
        out_dir="static/js"
    )

# Different routes for development vs production
routes = [
    route(path="/", source="pages/home.plush.html", template_type="PLUSH")
]

if not is_production:
    routes.append(
        route(
            path="/dev",
            source="pages/dev-tools.plush.html",
            template_type="PLUSH"
        )
    )
```

## Available Built-in Functions

### Site Configuration Functions

- `route(path, source, template_type, ...)` - Define a site route
- `translation(code, source, source_type, is_default)` - Configure translations
- `partial(source, template_type)` - Define reusable template components
- `js_target(source, out_dir)` - Configure JavaScript/CSS bundles

### Data Loading Functions

- `read_yaml(path)` - Load YAML file as Starlark data
- `read_json(path)` - Load JSON file as Starlark data

## Project Structure

```
your-project/
├── manifest.star              # Main configuration file
├── config/                    # Optional: modular config files
│   ├── routes.star
│   └── assets.star
├── data/                      # External data files
│   ├── blog-posts.yaml
│   ├── products.yaml
│   └── site.json
├── pages/                     # Page templates
│   ├── home.plush.html
│   ├── about.md
│   ├── blog/
│   │   └── my-post/
│   │       ├── en.md
│   │       └── es.md
│   └── 404.plush.html
├── templates/
│   ├── layouts/
│   │   └── base.plush.html
│   └── partials/
│       ├── header.plush.html
│       ├── footer.plush.html
│       └── blog-card.plush.html
├── src/                       # JavaScript/CSS source
│   ├── main.ts
│   ├── blog.ts
│   └── styles.css
├── static/                    # Static assets
│   ├── images/
│   └── fonts/
└── translations/
    ├── en.yaml
    ├── es.yaml
    └── fr.yaml
```

## Template Features

### Plush Templates

Plush templates provide powerful templating capabilities:

```html
<!-- pages/home.plush.html -->
<div class="hero">
    <h1><%= site_name %></h1>
    <p><%= text("welcome_message") %></p>

    <!-- Loop through data -->
    <% for (post) in featured_posts { %>
        <article>
            <h2><%= post.title %></h2>
            <p><%= truncate(post.description, 100) %></p>
        </article>
    <% } %>

    <!-- Include partials -->
    <%= partial("blog_card") %>
</div>
```

### Markdown with Frontmatter

```markdown
---
title: My Blog Post
description: A comprehensive guide to static sites
author: John Doe
date: 2024-01-15
tags: [web, development, static-sites]
---

# Welcome to My Blog

Your markdown content here with full support for:
- Code blocks
- Tables
- Links
- Images

You can also include partials in markdown:
<%= partial("call_to_action") %>
```

### Available Template Functions

- `text(key)` - Get translated text
- `partial(name)` - Include a partial template
- `markdown(text)` - Render markdown to HTML
- `truncate(text, length)` - Truncate text with ellipsis
- `startsWith(text, prefix)` - Check if text starts with prefix
- `endsWith(text, suffix)` - Check if text ends with suffix
- `contains(text, substring)` - Check if text contains substring
- `replace(text, old, new)` - Replace first occurrence
- `replaceAll(text, old, new)` - Replace all occurrences
- `urlEncode(text)` - URL encode text
- `stringify(data)` - Convert data to JSON string

## Commands

### Development Server

```bash
go-static-site serve [options]

Options:
  -p, --port int     Port to serve on (default 3000)
  -h, --host string  Host to serve on (default "localhost")
```

### Build Static Site

```bash
go-static-site build [options]

Options:
  -o, --output string  Output directory (default "public")
  --clean             Clean output directory before build
```

### Other Commands

```bash
# Validate configuration
go-static-site validate

# Generate sitemap
go-static-site sitemap

# Show routes
go-static-site routes
```

## Why Choose Go Static Site Generator?

**Explicit Configuration**: No magic directories or naming conventions. Your configuration file is the single source of truth.

**Flexible Organization**: Structure your project however makes sense - the manifest defines what goes where.

**Powerful Templating**: Plush templates with partials, helpers, and full programming capabilities.

**Built-in i18n**: Multi-language support with automatic route generation.

**Asset Pipeline**: Integrated JavaScript/CSS bundling with esbuild for fast builds.

**Type Safety**: Starlark configuration provides better validation and error messages than YAML.

**Programmatic Config**: Use functions, variables, and imports to reduce repetition and enable complex configurations.

**Performance**: Written in Go for fast builds and development server.

## Migration from YAML

If you have an existing `manifest.yaml`, you can easily convert it to Starlark:

```yaml
# manifest.yaml (old)
app_origin: https://mysite.com
routes:
  - path: /
    source: pages/home.plush.html
    template_type: PLUSH
```

```python
# manifest.star (new)
app_origin = "https://mysite.com"

routes = [
    route(
        path="/",
        source="pages/home.plush.html",
        template_type="PLUSH"
    )
]
```

The generator automatically detects and uses `.star` files when available, falling back to `.yaml` if not found.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT