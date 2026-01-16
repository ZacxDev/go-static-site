# Proving Site Manifest
# Demonstrates all template types: Plush, Markdown, and Gomponents

# Site configuration
app_origin = "http://localhost:8080"
default_layout_source = "layouts/base.plush.html"
not_found_page_source = "templates/404.plush.html"
is_production_environment = False
default_port = "8080"
output_dir = "dist"
static_dir = "static"

# Translations
translations = [
    translation(
        code = "en",
        source = "translations/en.json",
        source_type = "JSON",
        is_default = True,
    ),
    translation(
        code = "es",
        source = "translations/es.json",
        source_type = "JSON",
        is_default = False,
    ),
]

# Global context available to all templates
global_render_context = {
    "site_name": "Proving Site",
    "year": 2024,
}

# Routes demonstrating all template types
routes = [
    # Plush template route
    route(
        path = "/",
        source = "templates/home.plush.html",
        template_type = "PLUSH",
        page_title = "Home - Plush Template",
        static_render_data = {
            "features": [
                {"name": "Plush Templates", "desc": "Traditional template syntax"},
                {"name": "Markdown", "desc": "Write content in Markdown"},
                {"name": "Gomponents", "desc": "Type-safe Go components"},
            ],
        },
    ),

    # Plush about page
    route(
        path = "/about",
        source = "templates/about.plush.html",
        template_type = "PLUSH",
        page_title = "About - Plush Template",
    ),

    # Markdown blog post
    route(
        path = "/blog/hello-world",
        source = "pages/hello-world.md",
        template_type = "MARKDOWN",
        page_title = "Hello World",
    ),

    # Gomponents home page (alternative)
    route(
        path = "/gom",
        template_type = "GOMPONENTS",
        component_id = "proving_home",
        page_title = "Home - Gomponents",
    ),

    # Gomponents features page
    route(
        path = "/gom/features",
        template_type = "GOMPONENTS",
        component_id = "proving_features",
        page_title = "Features - Gomponents",
        static_render_data = {
            "features": [
                {"name": "Type Safety", "desc": "Compile-time error checking"},
                {"name": "IDE Support", "desc": "Full autocomplete and refactoring"},
                {"name": "Performance", "desc": "No template parsing overhead"},
                {"name": "Composition", "desc": "Build UIs with function composition"},
            ],
        },
    ),

    # Gomponents with dynamic data
    route(
        path = "/gom/data",
        template_type = "GOMPONENTS",
        component_id = "proving_data",
        page_title = "Data Demo - Gomponents",
        static_render_data = {
            "users": [
                {"id": 1, "name": "Alice", "role": "Admin"},
                {"id": 2, "name": "Bob", "role": "User"},
                {"id": 3, "name": "Charlie", "role": "User"},
            ],
            "stats": {
                "total_users": 3,
                "active_today": 2,
            },
        },
    ),
]
