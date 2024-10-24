# Go Static Site

The one-file declarative static site generator. Define your site routes/sources in a single manifest file.

## Introduction

Go Static Site is designed with simplicity in mind - we believe configuring your static site should be as straightforward as describing it. By using a single YAML manifest file, you can define your entire site structure, including routes, templates, JavaScript bundles, and translations. No more memorizing magic directory layouts or file naming patterns.

## Features

- 🗺️ Single file configuration
- 🌐 i18n enabled by default
- 📦 JavaScript transpilation and injection using checksums to enable heavy caching without risk of serving stale code
- 🖌️ Template rendering with Plush
- Partial injection for simple modular and reusable component templates
- ✍️ Markdown support with frontmatter
- 🗺️ Automatic sitemap generation
- 🔄 Development server with hot reloading (coming soon)
- 📱 Static site generation for production

## Stack

- **Go**: Core runtime and static site generation
- **Plush**: Template engine for HTML templates
- **esbuild**: JavaScript bundling and optimization
- **Markdown**: Content authoring with frontmatter support
- **Gorilla Mux**: Routing
- **YAML**: Configuration format

## Quick Start

1. Install the generator:
```bash
go install github.com/yourusername/go-static-site@latest
```

2. Create a `manifest.yaml` file in your project root:
```yaml
origin: https://mysite.com

partials:
  header:
    source: "templates/partials/header.plush.html"
    template_type: "PLUSH"

routes:
  - path: /
    source: pages/home.plush.html
    template_type: PLUSH
    partial_deps:
      - header
    javascript_deps:
      - main

  - path: /blog/:slug
    source: pages/blog/[slug]/[lang].md
    template_type: MARKDOWN
    partial_deps:
      - header
    javascript_deps:
      - blog

javascript:
  main:
    source: src/main.ts
    out_dir: static/js
  blog:
    source: src/blog.ts
    out_dir: static/js

translations:
  - code: en
    source: translations/en.yaml
    source_type: YAML
  - code: es
    source: translations/es.yaml
    source_type: YAML

not_found_page_source: pages/404.plush.html
```

3. Start the development server:
```bash
go-static-site serve
```

4. Build for production:
```bash
go-static-site build
```

## Project Structure

```
your-project/
├── manifest.yaml
├── pages/
│   ├── home.plush.html
│   ├── blog/
│   │   └── my-post/
│   │       ├── en.md
│   │       └── es.md
│   └── 404.plush.html
├── src/
│   ├── main.ts
│   └── blog.ts
├── static/
│   └── css/
│       └── styles.css
├── templates/
│   └── layouts/
│       └── base.plush.html
└── translations/
    ├── en.yaml
    └── es.yaml
```

## Configuration

### Routes
Routes can be static or dynamic:
- Static routes: `/about`, `/contact`
- Dynamic routes: `/blog/:slug`, `/products/:id`
- Language-specific routes are automatically generated based on your translations

### JavaScript Bundling
- Uses esbuild for blazing fast bundling
- Automatic file hashing for cache busting
- Bundles are specified in the manifest and can be referenced in routes

### Templates
Two template types are supported:
- `PLUSH`: HTML templates with Go's Plush templating engine
- `MARKDOWN`: Markdown files with YAML frontmatter

### Translations
- YAML-based translation files
- Automatic language route generation
- Translation helper available in templates: `<%= text("key") %>`

## Markdown Frontmatter

```markdown
title: My Blog Post
description: A great post about things
---
# Content starts here

Your markdown content...
```

## Development

```bash
# Start development server
go-static-site serve -p 9010

# Build static site
go-static-site build
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
