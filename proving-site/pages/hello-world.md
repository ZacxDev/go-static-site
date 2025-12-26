---
title: Hello World - My First Blog Post
description: A demonstration of Markdown templates in go-static-site
author: Developer
date: 2024-01-15
tags:
  - markdown
  - demo
  - tutorial
---

# Hello World!

Welcome to this **Markdown** blog post. This demonstrates the Markdown template type in go-static-site.

## Features of Markdown Templates

Markdown templates support:

- **YAML Frontmatter** - Define metadata like title, description, author
- **Full Markdown Syntax** - Headers, lists, code blocks, links, images
- **Automatic HTML Conversion** - Using goldmark under the hood

### Code Examples

Here's some Go code:

```go
func main() {
    fmt.Println("Hello from go-static-site!")
}
```

And some JavaScript:

```javascript
const greet = (name) => `Hello, ${name}!`;
console.log(greet('World'));
```

## Links to Other Pages

Check out other template types:

- [Home (Plush)](/) - The main home page using Plush templates
- [About (Plush)](/about) - About page with Plush
- [Gomponents Home](/gom) - Home page using Gomponents
- [Gomponents Features](/gom/features) - Features page with data

## Blockquotes

> Markdown is a lightweight markup language that you can use to add formatting elements to plaintext text documents.
>
> — John Gruber

## Tables

| Template Type | Syntax | Best For |
|---------------|--------|----------|
| Plush | `<%= expr %>` | Designers, simple pages |
| Markdown | Standard MD | Blog posts, docs |
| Gomponents | Go code | Complex UIs, type safety |

---

*This post was rendered using the Markdown template engine.*
