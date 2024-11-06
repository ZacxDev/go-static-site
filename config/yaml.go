package config

type Partial struct {
	Source       string `yaml:"source"`
	TemplateType string `yaml:"template_type"`
}

type JavascriptTarget struct {
	Source string `yaml:"source"`
	OutDir string `yaml:"out_dir"`
}

type SiteManifest struct {
	Routes              []Route                     `yaml:"routes"`
	JavascriptTargets   map[string]JavascriptTarget `yaml:"javascript"`
	Translations        []Translation               `yaml:"translations"`
	Origin              string                      `yaml:"origin"`
	NotFoundPageSource  string                      `yaml:"not_found_page_source"`
	Partials            map[string]Partial          `yaml:"partials"`
	DefaultLayoutSource string                      `yaml:"default_layout_source"`
}

type Route struct {
	Path           string   `yaml:"path"`
	Source         string   `yaml:"source"`
	TemplateType   string   `yaml:"template_type"`
	JavascriptDeps []string `yaml:"javascript_deps"`
	PartialDeps    []string `yaml:"partial_deps"`
	LayoutSource   string   `yaml:"layout_source,omitempty"`
	PageTitle      string   `yaml:"title,omitempty"`
}

type Translation struct {
	Code       string `yaml:"code"`
	Source     string `yaml:"source"`
	SourceType string `yaml:"source_type"`
}
