package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
	manifest, err := LoadManifest("manifest.yaml")
	if err != nil {
		return nil, fmt.Errorf("error loading manifest: %v", err)
	}

	// Set up middleware
	router.NotFoundHandler = http.HandlerFunc(GetCustom404Handler(manifest.NotFoundPageSource, manifest.DefaultLayoutSource))

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
		if len(manifest.Translations) == 0 {
			return nil, errors.New("must have at least one translation")
		}

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

	sitemap, err := utils.GenerateSitemapContent(registeredRoutes, manifest.AppOrigin, manifest.Routes)
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

func LoadManifest(filename string) (*config.SiteManifest, error) {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Try alternative extension if yaml file doesn't exist
		if strings.HasSuffix(filename, ".yaml") {
			starFilename := strings.TrimSuffix(filename, ".yaml") + ".star"
			if _, err := os.Stat(starFilename); err == nil {
				return config.ParseStarlarkManifest(starFilename)
			}
		}
		return nil, fmt.Errorf("manifest file not found: %s", filename)
	}

	// Parse based on file extension
	if strings.HasSuffix(filename, ".star") {
		return config.ParseStarlarkManifest(filename)
	}

	// Default to YAML parsing
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
		ctx.Set("appOrigin", manifest.AppOrigin)
		ctx.Set("apiOrigin", manifest.APIOrigin)
		ctx.Set("isProductionEnvironment", manifest.IsProductionEnviroment)

		jsSrcs := make([]string, len(route.JavascriptDeps))
		// Pass in javascript bundle paths
		for i, tsDepLabl := range route.JavascriptDeps {
			for label, publicPath := range emittedJS {
				if label == tsDepLabl {
					jsSrcs[i] = publicPath
				}
			}
		}
		ctx.Set("javascript_sources", jsSrcs)

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
		c := fmt.Sprintf("%s/%s%s", manifest.AppOrigin, lang, pathNoLang)
		ctx.Set("canonical", c)

		ctx.Set("currentPath", r.URL.Path)

		for key, value := range route.StaticRenderData {
			ctx.Set(key, value)
		}

		ctx.Set("stringify", func(data map[string]any) string {
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				log.Fatalf("Error marshalling to JSON: %v", err)
			}

			// Convert JSON bytes to string
			jsonString := string(jsonBytes)

			return jsonString
		})

		ctx.Set("urlEncode", func(input string) string {
			return url.QueryEscape(input)
		})

		ctx.Set("unescapeString", func(input string) string {
			return html.UnescapeString(input)
		})

		ctx.Set("html", func(input string) template.HTML {
			return template.HTML(input)
		})

		ctx.Set("secondsToISO8601", func(durationSeconds uint64) string {
			return secondsToISO8601(durationSeconds)
		})

		for key, value := range manifest.GlobalRenderContext {
			ctx.Set(key, value)
		}

		layoutSource := manifest.DefaultLayoutSource
		if route.LayoutSource != "" {
			layoutSource = route.LayoutSource
		}

		if layoutSource == "" {
			http.Error(w, "No layout source specified", http.StatusInternalServerError)
			return
		}

		var content string
		var err error

		switch route.TemplateType {
		case "PLUSH":
			ctx.Set("title", route.PageTitle)
			content, err = renderPlushTemplate(route.Source, route, manifest, ctx)
		case "MARKDOWN":
			var title, desc string
			content, title, desc, err = renderMarkdownTemplate(route.Source, route, manifest)
			ctx.Set("title", title)
			ctx.Set("description", desc)
		default:
			fmt.Println("Unsupported template type")
			http.Error(w, "Unsupported template type", http.StatusInternalServerError)
			return
		}

		if err != nil {
			msg := fmt.Sprintf("Error rendering template: %v", err)
			fmt.Printf("%+v\n", msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		ctx.Set("yield", template.HTML(content))

		baseContentB, err := os.ReadFile(layoutSource)
		if err != nil {
			msg := fmt.Sprintf("Error parsing base layout: %v", err)
			fmt.Printf("%+v\n", msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		// Preprocess base template for partials
		preprocess := PreprocessAllTemplates(route, manifest)
		baseContent, err := preprocess(string(baseContentB))
		if err != nil {
			msg := fmt.Sprintf("Error preprocessing base layout: %v", err)
			fmt.Printf("%+v\n", msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		baseLayout, err := plush.Parse(baseContent)
		if err != nil {
			msg := fmt.Sprintf("Error parsing base layout: %v", err)
			fmt.Printf("%+v\n", msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		pageHtml, err := baseLayout.Exec(ctx)
		if err != nil {
			msg := fmt.Sprintf("Error executing base layout: %v", err)
			fmt.Printf("%+v\n", msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(pageHtml))
		if err != nil {
			msg := fmt.Sprintf("Error writing response: %v", err)
			fmt.Printf("%+v\n", msg)
			http.Error(w, msg, http.StatusInternalServerError)
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

	res, err := template.Exec(ctx)
	return res, err
}

func renderMarkdownTemplate(source string, route config.Route, manifest *config.SiteManifest) (string, string, string, error) {
	content, err := os.ReadFile(source)
	if err != nil {
		return "", "", "", err
	}

	// Check if content starts with a frontmatter section (---)
	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "---\n") {
		return "", "", "", fmt.Errorf("markdown file must start with frontmatter section: %s", source)
	}

	// Find the end of the frontmatter section
	endOfFrontmatter := strings.Index(contentStr[4:], "\n---\n")
	if endOfFrontmatter == -1 {
		return "", "", "", fmt.Errorf("invalid Markdown file format - no closing frontmatter delimiter found: %s", source)
	}

	// Extract frontmatter and markdown content
	frontmatter := contentStr[4 : endOfFrontmatter+4]  // Skip initial "---\n" and get until end
	markdownContent := contentStr[endOfFrontmatter+8:] // Skip both "---\n" delimiters

	// Parse the frontmatter
	var metadata map[string]string
	err = yaml.Unmarshal([]byte(frontmatter), &metadata)
	if err != nil {
		return "", "", "", fmt.Errorf("error parsing frontmatter: %v", err)
	}

	// Preprocess markdown content for partials
	preprocess := PreprocessAllTemplates(route, manifest)
	preprocessed, err := preprocess(markdownContent)
	if err != nil {
		return "", "", "", err
	}

	// Parse the Markdown content
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.HardLineBreak
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

func secondsToISO8601(durationSeconds uint64) string {
	d := time.Duration(durationSeconds) * time.Second

	// Extract hours, minutes, and seconds
	hours := int64(d / time.Hour)
	d %= time.Hour
	minutes := int64(d / time.Minute)
	d %= time.Minute
	seconds := int64(d / time.Second)

	// Construct the ISO 8601 string
	result := "PT"
	if hours > 0 {
		result += fmt.Sprintf("%dH", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dM", minutes)
	}
	if seconds > 0 || result == "PT" { // Include seconds if nothing else is present
		result += fmt.Sprintf("%dS", seconds)
	}

	return result
}
