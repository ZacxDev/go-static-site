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
	Routes                 []Route                     `yaml:"routes"`
	JavascriptTargets      map[string]JavascriptTarget `yaml:"javascript"`
	Translations           []Translation               `yaml:"translations"`
	AppOrigin              string                      `yaml:"app_origin"`
	APIOrigin              string                      `yaml:"api_origin"`
	NotFoundPageSource     string                      `yaml:"not_found_page_source"`
	Partials               map[string]Partial          `yaml:"partials"`
	DefaultLayoutSource    string                      `yaml:"default_layout_source"`
	IsProductionEnviroment bool                        `yaml:"IsProductionEnviroment"`
	GlobalRenderContext    map[string]any              `yaml:"global_render_context"`
	EnableSpaMode          bool                        `yaml:"enable_spa_mode"`
}

type VideoData struct {
	Title           string `yaml:"title"`
	Description     string `yaml:"description"`
	ThumbnailLoc    string `yaml:"thumbnail_loc"`
	ContentLoc      string `yaml:"content_loc"`
	Duration        int    `yaml:"duration"`
	PublicationDate string `yaml:"publication_date"`
}

type Route struct {
	Path             string         `yaml:"path"`
	Source           string         `yaml:"source"`
	TemplateType     string         `yaml:"template_type"`
	JavascriptDeps   []string       `yaml:"javascript_deps"`
	PartialDeps      []string       `yaml:"partial_deps"`
	LayoutSource     string         `yaml:"layout_source,omitempty"`
	PageTitle        string         `yaml:"title,omitempty"`
	StaticRenderData map[string]any `yaml:"static_render_data"`
	SitemapVideoData *VideoData     `yaml:"sitemap_video_data,omitempty"`
}

type Translation struct {
	Code       string `yaml:"code"`
	Source     string `yaml:"source"`
	SourceType string `yaml:"source_type"`
	IsDefault  bool   `yaml:"is_default"`
}
