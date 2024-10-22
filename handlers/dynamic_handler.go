package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gobuffalo/plush"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"github.com/ZacxDev/go-static-site/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type RouteManifest struct {
	Routes []Route `yaml:"routes"`
}

type Route struct {
	Path         string `yaml:"path"`
	Source       string `yaml:"source"`
	TemplateType string `yaml:"template_type"`
}

var registeredRoutes []string

var translations map[string]map[string]string

func SetupRouter() (*mux.Router, error) {
	router := mux.NewRouter()

	// Set up middleware
	router.NotFoundHandler = http.HandlerFunc(Custom404Handler)

	// Set up static file serving
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Load manifest
	manifest, err := loadManifest("manifest.yaml")
	if err != nil {
		return nil, fmt.Errorf("error loading manifest: %v", err)
	}

	// Load translations
	translations, err = loadTranslations("translations")
	if err != nil {
		return nil, fmt.Errorf("error loading translations: %v", err)
	}

	// Set up routes from manifest
	for _, route := range manifest.Routes {
		if strings.Contains(route.Path, ":slug") {
			// Handle dynamic blog post routes
			err := setupBlogRoutes(router, route)
			if err != nil {
				return nil, fmt.Errorf("error setting up blog routes: %v", err)
			}
		} else {
			// Set up route with optional language parameter
			router.HandleFunc("/{lang:en|es}"+route.Path, DynamicHandler(route, manifest)).Methods("GET")
			router.HandleFunc(route.Path, DynamicHandler(route, manifest)).Methods("GET")

			registeredRoutes = append(registeredRoutes, "/{lang:en|es}"+route.Path, route.Path)
		}
	}

	sitemap, err := utils.GenerateSitemapContent(registeredRoutes)
	router.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(sitemap))
	}).Methods("GET")

	return router, nil
}

func setupBlogRoutes(router *mux.Router, route Route) error {
	blogPosts, err := filepath.Glob("pages/blog/*")
	if err != nil {
		return errors.WithStack(err)
	}

	for _, postDir := range blogPosts {
		isDir, err := isDirectory(postDir)
		if err != nil {
			return errors.WithStack(err)
		}

		if !isDir {
			continue
		}

		slug := filepath.Base(postDir)
		if slug == "" {
			continue
		}

		for supportedLang := range translations {
			// Route with language parameter
			langPath := strings.Replace(route.Path, ":slug", slug, 1)
			router.HandleFunc("/"+supportedLang+langPath, DynamicHandler(Route{
				Path:         langPath,
				Source:       filepath.Join(postDir, supportedLang+".md"),
				TemplateType: route.TemplateType,
			}, nil)).Methods("GET")
			registeredRoutes = append(registeredRoutes, "/"+supportedLang+langPath)
		}
	}

	return nil
}

func loadManifest(filename string) (*RouteManifest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var manifest RouteManifest
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func loadTranslations(dir string) (map[string]map[string]string, error) {
	translations := make(map[string]map[string]string)

	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		lang := strings.TrimSuffix(filepath.Base(file), ".yaml")
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var langTranslations map[string]string
		err = yaml.Unmarshal(data, &langTranslations)
		if err != nil {
			return nil, err
		}

		translations[lang] = langTranslations
	}

	return translations, nil
}

func DynamicHandler(route Route, manifest *RouteManifest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := plush.NewContext()
		vars := mux.Vars(r)
		ctx.Set("params", vars)
		ctx.Set("registeredRoutes", registeredRoutes)

		// Get language from URL parameter, default to "en"
		lang := vars["lang"]
		if lang == "" {
			lang = "en"
		}

		// Add translation helper
		ctx.Set("text", func(key string) string {
			if t, ok := translations[lang][key]; ok {
				return t
			}
			return key
		})

		// Add language helper
		ctx.Set("lang", lang)

		var supportedLangs []string
		for lang := range translations {
			supportedLangs = append(supportedLangs, lang)
		}

		ctx.Set("supportedLangs", supportedLangs)
		ctx.Set("appOrigin", os.Getenv("APP_ORIGIN"))

		// Add startsWith helper
		ctx.Set("startsWith", func(s string, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		})

		// Add matches helper
		ctx.Set("matches", func(s string, pat string) bool {
			re := regexp.MustCompile(pat)
			return re.Match([]byte(s))
		})

		ctx.Set("replace", func(s string, old string, n string) string {
			return strings.Replace(s, old, n, 1)
		})

		ctx.Set("replaceAll", func(s string, old string, n string) string {
			return strings.ReplaceAll(s, old, n)
		})

		// Add matches helper
		ctx.Set("replacePattern", func(s string, pat, n string) string {
			re := regexp.MustCompile(pat)
			return re.ReplaceAllString(s, n)
		})

		// Add canonical URL helper
		pathNoLang := strings.Replace(r.URL.Path, "/"+lang+"/", "/", 1)
		c := fmt.Sprintf("https://mylinksprofile.com/%s%s", lang, pathNoLang)
		ctx.Set("canonical", c)

		ctx.Set("currentPath", r.URL.Path)

		var content string
		var err error

		switch route.TemplateType {
		case "PLUSH":
			content, err = renderPlushTemplate(route.Source, ctx)
		case "MARKDOWN":
			var title, desc string
			content, title, desc, err = renderMarkdownTemplate(route.Source)
			ctx.Set("title", title)
			ctx.Set("description", desc)
		default:
			http.Error(w, "Unsupported template type", http.StatusInternalServerError)
			return
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
			return
		}

		ctx.Set("yield", template.HTML(content))

		baseContent, err := os.ReadFile("templates/layouts/base.plush.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing base layout: %v", err), http.StatusInternalServerError)
			return
		}

		baseLayout, err := plush.Parse(string(baseContent))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing base layout: %v", err), http.StatusInternalServerError)
			return
		}

		pageHtml, err := baseLayout.Exec(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error executing base layout: %v", err), http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(pageHtml))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func renderPlushTemplate(source string, ctx *plush.Context) (string, error) {
	content, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}

	template, err := plush.Parse(string(content))
	if err != nil {
		return "", err
	}

	return template.Exec(ctx)
}

func renderMarkdownTemplate(source string) (string, string, string, error) {
	content, err := os.ReadFile(source)
	if err != nil {
		return "", "", "", err
	}

	// Split the content into frontmatter and Markdown
	parts := strings.SplitN(string(content), "\n---\n", 3)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid Markdown file format: %s", source)
	}

	// Parse the frontmatter
	var metadata map[string]string
	err = yaml.Unmarshal([]byte(parts[0]), &metadata)
	if err != nil {
		return "", "", "", fmt.Errorf("error parsing frontmatter: %v", err)
	}

	// Parse the Markdown content
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	md := []byte(parts[1])
	htmlContent := markdown.ToHTML(md, p, nil)
	contentHtml := strings.Replace(`
  <article class="flex flex-col gap-4 blog-container">
  [content]
  </article>
  `, "[content]", string(htmlContent), 1)

	return contentHtml, metadata["title"], metadata["description"], nil
}

func GetRegisteredRoutes() []string {
	return registeredRoutes
}

func isDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
