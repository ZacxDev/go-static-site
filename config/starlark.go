package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.starlark.net/starlark"
)

// ModuleLoader handles loading Starlark modules
type ModuleLoader struct {
	cache map[string]starlark.StringDict
}

// NewModuleLoader creates a new ModuleLoader instance
func NewModuleLoader() *ModuleLoader {
	return &ModuleLoader{
		cache: make(map[string]starlark.StringDict),
	}
}

// LoadModule implements the starlark.Thread.Load interface
func (l *ModuleLoader) LoadModule(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	// Check cache first
	if cached, ok := l.cache[module]; ok {
		return cached, nil
	}

	// Handle both absolute and relative paths
	var modulePath string
	if filepath.IsAbs(module) {
		modulePath = module
	} else {
		// Get the directory of the current file being executed
		currentFile := thread.Name
		dir := filepath.Dir(currentFile)
		modulePath = filepath.Join(dir, module)
	}

	// Read and execute the module
	content, err := os.ReadFile(modulePath)
	if err != nil {
		return nil, fmt.Errorf("error reading module %s: %v", module, err)
	}

	// Create a new thread for module execution
	moduleThread := &starlark.Thread{
		Name:  modulePath,
		Load:  l.LoadModule,
		Print: thread.Print,
	}

	// Execute the module
	globals, err := starlark.ExecFile(moduleThread, modulePath, content, builtins())
	if err != nil {
		return nil, fmt.Errorf("error executing module %s: %v", module, err)
	}

	// Cache the results
	l.cache[module] = globals
	return globals, nil
}

// ParseStarlarkManifest reads and parses a manifest.star file into a SiteManifest
func ParseStarlarkManifest(filename string) (*SiteManifest, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading starlark manifest: %v", err)
	}

	// Create module loader
	loader := NewModuleLoader()

	// Create thread with module loading capability
	thread := &starlark.Thread{
		Name:  filename,
		Load:  loader.LoadModule,
		Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) },
	}

	// Execute the manifest file
	globals, err := starlark.ExecFile(thread, filename, content, builtins())
	if err != nil {
		return nil, fmt.Errorf("error executing starlark manifest: %v", err)
	}

	manifest := &SiteManifest{}

	// Parse basic string fields
	if v, ok := globals["app_origin"]; ok {
		manifest.AppOrigin = v.(starlark.String).GoString()
	}
	if v, ok := globals["api_origin"]; ok {
		manifest.APIOrigin = v.(starlark.String).GoString()
	}
	if v, ok := globals["default_layout_source"]; ok {
		manifest.DefaultLayoutSource = v.(starlark.String).GoString()
	}
	if v, ok := globals["not_found_page_source"]; ok {
		manifest.NotFoundPageSource = v.(starlark.String).GoString()
	}
	if v, ok := globals["is_production_environment"]; ok {
		manifest.IsProductionEnviroment = v.(starlark.Bool).Truth().String() == "True"
	}

	// Parse routes
	if v, ok := globals["routes"]; ok {
		routes, err := parseRoutes(v)
		if err != nil {
			return nil, fmt.Errorf("error parsing routes: %v", err)
		}
		manifest.Routes = routes
	}

	// Parse translations
	if v, ok := globals["translations"]; ok {
		translations, err := parseTranslations(v)
		if err != nil {
			return nil, fmt.Errorf("error parsing translations: %v", err)
		}
		manifest.Translations = translations
	}

	// Parse partials
	if v, ok := globals["partials"]; ok {
		partials, err := parsePartials(v)
		if err != nil {
			return nil, fmt.Errorf("error parsing partials: %v", err)
		}
		manifest.Partials = partials
	}

	// Parse javascript_targets
	if v, ok := globals["javascript_targets"]; ok {
		targets, err := parseJavascriptTargets(v)
		if err != nil {
			return nil, fmt.Errorf("error parsing javascript targets: %v", err)
		}
		manifest.JavascriptTargets = targets
	}

	// Parse css_targets
	if v, ok := globals["css_targets"]; ok {
		targets, err := parseJavascriptTargets(v) // Reuse the same parser since structure is identical
		if err != nil {
			return nil, fmt.Errorf("error parsing css targets: %v", err)
		}
		// Add CSS targets to JavaScript targets map since they share the same structure
		for k, v := range targets {
			manifest.JavascriptTargets[k] = v
		}
	}

	if v, ok := globals["global_render_context"]; ok {
		globalContext, err := parseGlobalContext(v)
		if err != nil {
			return nil, fmt.Errorf("error parsing global_render_context: %v", err)
		}
		manifest.GlobalRenderContext = globalContext
	}

	return manifest, nil
}

func parseGlobalContext(v starlark.Value) (map[string]any, error) {
	dict, ok := v.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("global_render_context must be a dict")
	}

	globalContext := make(map[string]any)
	for _, item := range dict.Items() {
		key := item[0].(starlark.String).GoString()
		globalContext[key] = convertStarlarkToGo(item[1])
	}
	return globalContext, nil
}

// builtins returns a list of built-in functions and values for Starlark execution
func builtins() starlark.StringDict {
	return starlark.StringDict{
		"route":     starlark.NewBuiltin("route", routeBuiltin),
		"trans":     starlark.NewBuiltin("trans", translationBuiltin),
		"partial":   starlark.NewBuiltin("partial", partialBuiltin),
		"js_target": starlark.NewBuiltin("js_target", jsTargetBuiltin),
	}
}

func routeBuiltin(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path, source, templateType, layoutSource, pageTitle string
	var javascriptDeps, partialDeps *starlark.List
	var staticData, videoData *starlark.Dict

	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"path", &path,
		"source", &source,
		"template_type?", &templateType,
		"layout_source?", &layoutSource,
		"javascript_deps?", &javascriptDeps,
		"partial_deps?", &partialDeps,
		"page_title?", &pageTitle,
		"static_render_data?", &staticData,
		"sitemap_video_data?", &videoData,
	); err != nil {
		return nil, err
	}

	// Convert deps lists to string slices
	jsDepsList := make([]string, 0)
	if javascriptDeps != nil {
		for i := 0; i < javascriptDeps.Len(); i++ {
			jsDepsList = append(jsDepsList, javascriptDeps.Index(i).(starlark.String).GoString())
		}
	}

	partialDepsList := make([]string, 0)
	if partialDeps != nil {
		for i := 0; i < partialDeps.Len(); i++ {
			partialDepsList = append(partialDepsList, partialDeps.Index(i).(starlark.String).GoString())
		}
	}

	staticDataMap := make(map[string]any)
	if staticData != nil {
		for _, item := range staticData.Items() {
			key := item[0].(starlark.String).GoString()
			value := convertStarlarkToGo(item[1])
			staticDataMap[key] = value
		}
	}

	var videoDataStruct *starlarkVideoData
	if videoData != nil {
		videoDataStruct = &starlarkVideoData{}
		for _, item := range videoData.Items() {
			key := item[0].(starlark.String).GoString()
			switch key {
			case "title":
				videoDataStruct.title = item[1].(starlark.String).GoString()
			case "description":
				videoDataStruct.description = item[1].(starlark.String).GoString()
			case "thumbnail_loc":
				videoDataStruct.thumbnailLoc = item[1].(starlark.String).GoString()
			case "content_loc":
				videoDataStruct.contentLoc = item[1].(starlark.String).GoString()
			case "duration":
				if num, ok := item[1].(starlark.Int); ok {
					val, _ := num.Int64()
					videoDataStruct.duration = val
				}
			case "publication_date":
				videoDataStruct.publicationDate = item[1].(starlark.String).GoString()
			}
		}
	}

	return &starlarkRoute{
		path:             path,
		source:           source,
		templateType:     templateType,
		layoutSource:     layoutSource,
		javascriptDeps:   jsDepsList,
		partialDeps:      partialDepsList,
		pageTitle:        pageTitle,
		staticRenderData: staticDataMap,
		sitemapVideoData: videoDataStruct,
	}, nil
}

func translationBuiltin(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var code, source, sourceType string

	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"code", &code,
		"source", &source,
		"source_type?", &sourceType,
	); err != nil {
		return nil, err
	}

	if sourceType == "" {
		sourceType = "YAML"
	}

	return &starlarkTranslation{
		code:       code,
		source:     source,
		sourceType: sourceType,
	}, nil
}

func partialBuiltin(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var source, templateType string

	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"source", &source,
		"template_type", &templateType,
	); err != nil {
		return nil, err
	}

	return &starlarkPartial{
		source:       source,
		templateType: templateType,
	}, nil
}

func jsTargetBuiltin(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var source, outDir string

	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"source", &source,
		"out_dir", &outDir,
	); err != nil {
		return nil, err
	}

	return &starlarkJSTarget{
		source: source,
		outDir: outDir,
	}, nil
}

// Custom Starlark types to represent manifest components
type starlarkVideoData struct {
	title           string
	description     string
	thumbnailLoc    string
	contentLoc      string
	duration        int64
	publicationDate string
}

type starlarkRoute struct {
	path             string
	source           string
	templateType     string
	layoutSource     string
	javascriptDeps   []string
	partialDeps      []string
	pageTitle        string
	staticRenderData map[string]any
	sitemapVideoData *starlarkVideoData
}

func (r *starlarkRoute) String() string        { return fmt.Sprintf("route(%q)", r.path) }
func (r *starlarkRoute) Type() string          { return "route" }
func (r *starlarkRoute) Freeze()               {} // immutable
func (r *starlarkRoute) Truth() starlark.Bool  { return true }
func (r *starlarkRoute) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: route") }

type starlarkTranslation struct {
	code       string
	source     string
	sourceType string
}

func (t *starlarkTranslation) String() string       { return fmt.Sprintf("translation(%q)", t.code) }
func (t *starlarkTranslation) Type() string         { return "translation" }
func (t *starlarkTranslation) Freeze()              {} // immutable
func (t *starlarkTranslation) Truth() starlark.Bool { return true }
func (t *starlarkTranslation) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: translation")
}

type starlarkPartial struct {
	source       string
	templateType string
}

func (p *starlarkPartial) String() string        { return fmt.Sprintf("partial(%q)", p.source) }
func (p *starlarkPartial) Type() string          { return "partial" }
func (p *starlarkPartial) Freeze()               {} // immutable
func (p *starlarkPartial) Truth() starlark.Bool  { return true }
func (p *starlarkPartial) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: partial") }

type starlarkJSTarget struct {
	source string
	outDir string
}

func (j *starlarkJSTarget) String() string        { return fmt.Sprintf("js_target(%q)", j.source) }
func (j *starlarkJSTarget) Type() string          { return "js_target" }
func (j *starlarkJSTarget) Freeze()               {} // immutable
func (j *starlarkJSTarget) Truth() starlark.Bool  { return true }
func (j *starlarkJSTarget) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: js_target") }

// Helper functions to parse Starlark values into Go types
func parseRoutes(v starlark.Value) ([]Route, error) {
	list, ok := v.(*starlark.List)
	if !ok {
		return nil, fmt.Errorf("routes must be a list")
	}

	var routes []Route
	for i := 0; i < list.Len(); i++ {
		route, ok := list.Index(i).(*starlarkRoute)
		if !ok {
			return nil, fmt.Errorf("invalid route at index %d", i)
		}

		var videoData *VideoData
		if route.sitemapVideoData != nil {
			videoData = &VideoData{
				Title:           route.sitemapVideoData.title,
				Description:     route.sitemapVideoData.description,
				ThumbnailLoc:    route.sitemapVideoData.thumbnailLoc,
				ContentLoc:      route.sitemapVideoData.contentLoc,
				Duration:        int(route.sitemapVideoData.duration),
				PublicationDate: route.sitemapVideoData.publicationDate,
			}
		}

		routes = append(routes, Route{
			Path:             route.path,
			Source:           route.source,
			TemplateType:     route.templateType,
			LayoutSource:     route.layoutSource,
			JavascriptDeps:   route.javascriptDeps,
			PartialDeps:      route.partialDeps,
			PageTitle:        route.pageTitle,
			StaticRenderData: route.staticRenderData,
			SitemapVideoData: videoData,
		})
	}
	return routes, nil
}

func parseTranslations(v starlark.Value) ([]Translation, error) {
	list, ok := v.(*starlark.List)
	if !ok {
		return nil, fmt.Errorf("translations must be a list")
	}

	var translations []Translation
	for i := 0; i < list.Len(); i++ {
		trans, ok := list.Index(i).(*starlarkTranslation)
		if !ok {
			return nil, fmt.Errorf("invalid translation at index %d", i)
		}
		translations = append(translations, Translation{
			Code:       trans.code,
			Source:     trans.source,
			SourceType: trans.sourceType,
		})
	}
	return translations, nil
}

func parsePartials(v starlark.Value) (map[string]Partial, error) {
	dict, ok := v.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("partials must be a dict")
	}

	partials := make(map[string]Partial)
	for _, item := range dict.Items() {
		key := item[0].(starlark.String).GoString()
		partial, ok := item[1].(*starlarkPartial)
		if !ok {
			return nil, fmt.Errorf("invalid partial: %s", key)
		}
		partials[key] = Partial{
			Source:       partial.source,
			TemplateType: partial.templateType,
		}
	}
	return partials, nil
}

func parseJavascriptTargets(v starlark.Value) (map[string]JavascriptTarget, error) {
	dict, ok := v.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("javascript_targets must be a dict")
	}

	targets := make(map[string]JavascriptTarget)
	for _, item := range dict.Items() {
		key := item[0].(starlark.String).GoString()
		target, ok := item[1].(*starlarkJSTarget)
		if !ok {
			return nil, fmt.Errorf("invalid javascript target: %s", key)
		}
		targets[key] = JavascriptTarget{
			Source: target.source,
			OutDir: target.outDir,
		}
	}
	return targets, nil
}

func convertStarlarkToGo(v starlark.Value) any {
	switch v := v.(type) {
	case starlark.String:
		return v.GoString()
	case starlark.Int:
		val, _ := v.Int64()
		return val
	case starlark.Float:
		return float64(v)
	case starlark.Bool:
		return bool(v)
	case *starlark.List:
		result := make([]any, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			result = append(result, convertStarlarkToGo(v.Index(i)))
		}
		return result
	case *starlark.Dict:
		result := make(map[string]any)
		for _, item := range v.Items() {
			key := item[0].(starlark.String).GoString()
			result[key] = convertStarlarkToGo(item[1])
		}
		return result
	default:
		return nil
	}
}
