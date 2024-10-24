package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ZacxDev/go-static-site/config"
	"github.com/ZacxDev/go-static-site/javascript"
	"github.com/ZacxDev/go-static-site/utils"
	"github.com/gobuffalo/plush"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var registeredRoutes []string

func SetupRouter() (*mux.Router, error) {
	router := mux.NewRouter()

	// Load manifest
	manifest, err := loadManifest("manifest.yaml")
	if err != nil {
		return nil, fmt.Errorf("error loading manifest: %v", err)
	}

	// Set up middleware
	router.NotFoundHandler = http.HandlerFunc(GetCustom404Handler(manifest.NotFoundPageSource))

	// Set up static file serving
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	translations, err := loadTranslations(manifest.Translations)
	if err != nil {
		return nil, fmt.Errorf("error loading translations: %v", err)
	}

	var langPathPattern string
	if len(manifest.Translations) > 0 {
		var langPathPatternB strings.Builder
		langPathPatternB.WriteString("/{lang:")
		for i, translation := range manifest.Translations {
			langPathPatternB.WriteString(translation.Code)

			if i+1 < len(manifest.Translations) {
				langPathPatternB.WriteRune('|')
			}
		}
		langPathPatternB.WriteString("}")
		langPathPattern = langPathPatternB.String()
	} else {
		langPathPattern = manifest.Translations[0].Code
	}

	emittedJS, err := javascript.CompileJSTarget(manifest.JavascriptTargets)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Set up routes from manifest
	for _, route := range manifest.Routes {
		re := regexp.MustCompile("\\/:\\w+")
		isDynParam := re.Match([]byte(route.Path))
		if isDynParam {
			// Handle dynamic blog post routes
			err := setupDynamicParamRoutes(router, route, emittedJS, translations, manifest)
			if err != nil {
				return nil, fmt.Errorf("error setting up blog routes: %v", err)
			}
		} else {
			// Set up route with optional language parameter
			router.HandleFunc(langPathPattern+route.Path, DynamicHandler(route, manifest, emittedJS, translations)).Methods("GET")
			router.HandleFunc(route.Path, DynamicHandler(route, manifest, emittedJS, translations)).Methods("GET")

			registeredRoutes = append(registeredRoutes, langPathPattern+route.Path, route.Path)
		}
	}

	sitemap, err := utils.GenerateSitemapContent(registeredRoutes)
	router.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(sitemap))
	}).Methods("GET")

	return router, nil
}

func setupDynamicParamRoutes(
	router *mux.Router,
	route config.Route,
	emittedJS map[string]string,
	translations map[string]map[string]string,
	manifest *config.SiteManifest,
) error {
	re := regexp.MustCompile(":\\w+")
	globRoute := re.ReplaceAllString(route.Path, "*")
	globDirPath := "pages" + globRoute
	blogPosts, err := filepath.Glob(globDirPath)
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

		isMDSource := route.TemplateType == "MARKDOWN"
		dynSourceRe := regexp.MustCompile("\\[\\w+\\]")
		isDynSource := dynSourceRe.Match([]byte(route.Source))

		for supportedLang := range translations {
			// Route with language parameter
			langPath := re.ReplaceAllString(route.Path, slug)
			var source string
			if isDynSource {
				if isMDSource {
					source = filepath.Join(postDir, supportedLang+".md")
				} else {
					source = filepath.Join(postDir, supportedLang+".plush.html")
				}
			} else {
				source = route.Source
			}

			router.HandleFunc("/"+supportedLang+langPath, DynamicHandler(config.Route{
				Path:           langPath,
				Source:         source,
				TemplateType:   route.TemplateType,
				JavascriptDeps: route.JavascriptDeps,
				PartialDeps:    route.PartialDeps,
			}, manifest, emittedJS, translations)).Methods("GET")
			registeredRoutes = append(registeredRoutes, "/"+supportedLang+langPath)
		}
	}

	return nil
}

func loadManifest(filename string) (*config.SiteManifest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var manifest config.SiteManifest
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func loadTranslations(trans []config.Translation) (map[string]map[string]string, error) {
	translations := make(map[string]map[string]string, 0)

	for _, tr := range trans {
		file := tr.Source
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		if tr.SourceType == "YAML" {
			var langTranslations map[string]string
			err = yaml.Unmarshal(data, &langTranslations)
			if err != nil {
				return nil, err
			}

			translations[tr.Code] = langTranslations
		} else {
			return nil, errors.New(fmt.Sprintf("unsupported translation source type: %s", tr.SourceType))
		}
	}

	return translations, nil
}

// PartialProcessingContext tracks partial inclusion to prevent circular dependencies
type PartialProcessingContext struct {
	ProcessedPartials map[string]bool
	CurrentDepth      int
	MaxDepth          int
}

// NewPartialProcessingContext creates a new context for partial processing
func NewPartialProcessingContext() *PartialProcessingContext {
	return &PartialProcessingContext{
		ProcessedPartials: make(map[string]bool),
		CurrentDepth:      0,
		MaxDepth:          10, // Maximum nesting depth to prevent infinite recursion
	}
}

// PreprocessTemplate handles partial injection before template rendering
func PreprocessTemplate(
	content string,
	route config.Route,
	manifest *config.SiteManifest,
	ctx *PartialProcessingContext,
) (string, error) {
	if ctx == nil {
		ctx = NewPartialProcessingContext()
	}

	if ctx.CurrentDepth >= ctx.MaxDepth {
		return "", fmt.Errorf("maximum partial nesting depth (%d) exceeded", ctx.MaxDepth)
	}

	// Regular expression to find partial tags: <%= partial("name") %>
	partialRegex := regexp.MustCompile(`<%=\s*partial\("([^"]+)"\)\s*%>`)

	// Find all partial references
	matches := partialRegex.FindAllStringSubmatch(content, -1)

	// Replace each partial reference with its content
	for _, match := range matches {
		fullMatch := match[0]
		partialName := match[1]

		// Check for circular dependencies
		if ctx.ProcessedPartials[partialName] {
			return "", fmt.Errorf("circular dependency detected in partial: %s", partialName)
		}

		// Verify partial is in dependencies
		found := false
		for _, dep := range route.PartialDeps {
			if dep == partialName {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("partial %s not declared in partial_deps", partialName)
		}

		// Get partial configuration
		partialConfig, exists := manifest.Partials[partialName]
		if !exists {
			return "", fmt.Errorf("partial %s not found in manifest", partialName)
		}

		// Load partial content
		partialContent, err := loadPartial(partialConfig)
		if err != nil {
			return "", errors.WithStack(err)
		}

		// Mark this partial as being processed
		ctx.ProcessedPartials[partialName] = true
		ctx.CurrentDepth++

		// Recursively process any nested partials
		processedContent, err := PreprocessTemplate(partialContent, route, manifest, ctx)
		if err != nil {
			return "", errors.WithStack(err)
		}

		// Unmark the partial after processing
		delete(ctx.ProcessedPartials, partialName)
		ctx.CurrentDepth--

		// Replace the partial tag with its processed content
		content = strings.Replace(content, fullMatch, processedContent, 1)
	}

	return content, nil
}

func PreprocessAllTemplates(route config.Route, manifest *config.SiteManifest) func(string) (string, error) {
	ctx := NewPartialProcessingContext()
	return func(content string) (string, error) {
		return PreprocessTemplate(content, route, manifest, ctx)
	}
}

func DynamicHandler(
	route config.Route,
	manifest *config.SiteManifest,
	emittedJS map[string]string,
	translations map[string]map[string]string,
) http.HandlerFunc {
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

		ctx.Set("lang", lang)

		var supportedLangs []string
		for lang := range translations {
			supportedLangs = append(supportedLangs, lang)
		}

		ctx.Set("supportedLangs", supportedLangs)
		ctx.Set("appOrigin", os.Getenv("APP_ORIGIN"))

		// Pass in javascript bundle paths
		for _, tsDepLabl := range route.JavascriptDeps {
			for label, publicPath := range emittedJS {
				if label == tsDepLabl {
					ctx.Set(tsDepLabl, publicPath)
				}
			}
		}

		// Add helper functions
		ctx.Set("startsWith", func(s string, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		})

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

		ctx.Set("replacePattern", func(s string, pat, n string) string {
			re := regexp.MustCompile(pat)
			return re.ReplaceAllString(s, n)
		})

		// Add canonical URL helper
		pathNoLang := strings.Replace(r.URL.Path, "/"+lang+"/", "/", 1)
		c := fmt.Sprintf("%s/%s%s", manifest.Origin, lang, pathNoLang)
		ctx.Set("canonical", c)

		ctx.Set("currentPath", r.URL.Path)

		var content string
		var err error

		switch route.TemplateType {
		case "PLUSH":
			content, err = renderPlushTemplate(route.Source, route, manifest, ctx)
		case "MARKDOWN":
			var title, desc string
			content, title, desc, err = renderMarkdownTemplate(route.Source, route, manifest)
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

		baseContentB, err := os.ReadFile("templates/layouts/base.plush.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing base layout: %v", err), http.StatusInternalServerError)
			return
		}

		// Preprocess base template for partials
		preprocess := PreprocessAllTemplates(route, manifest)
		baseContent, err := preprocess(string(baseContentB))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error preprocessing base layout: %v", err), http.StatusInternalServerError)
			return
		}

		baseLayout, err := plush.Parse(baseContent)
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

func renderPlushTemplate(source string, route config.Route, manifest *config.SiteManifest, ctx *plush.Context) (string, error) {
	content, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}

	// Preprocess template for partials
	preprocess := PreprocessAllTemplates(route, manifest)
	preprocessed, err := preprocess(string(content))
	if err != nil {
		return "", err
	}

	template, err := plush.Parse(preprocessed)
	if err != nil {
		return "", err
	}

	return template.Exec(ctx)
}

func renderMarkdownTemplate(source string, route config.Route, manifest *config.SiteManifest) (string, string, string, error) {
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

	// Preprocess markdown content for partials
	preprocess := PreprocessAllTemplates(route, manifest)
	preprocessed, err := preprocess(parts[1])
	if err != nil {
		return "", "", "", err
	}

	// Parse the Markdown content
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	md := []byte(preprocessed)
	htmlContent := markdown.ToHTML(md, p, nil)
	contentHtml := strings.Replace(`
  <article class="flex flex-col gap-4 blog-container">
  [content]
  </article>
  `, "[content]", string(htmlContent), 1)

	return contentHtml, metadata["title"], metadata["description"], nil
}

func loadPartial(partial config.Partial) (string, error) {
	content, err := os.ReadFile(partial.Source)
	if err != nil {
		return "", errors.WithStack(err)
	}

	switch partial.TemplateType {
	case "PLUSH":
		return string(content), nil
	case "MARKDOWN":
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		p := parser.NewWithExtensions(extensions)
		htmlContent := markdown.ToHTML(content, p, nil)
		return string(htmlContent), nil
	default:
		return "", fmt.Errorf("unsupported partial template type: %s", partial.TemplateType)
	}
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
