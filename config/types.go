package config


type JavascriptTarget struct {
	Source string
	OutDir string
}

type SiteManifest struct {
	Routes                 []Route
	JavascriptTargets      map[string]JavascriptTarget
	Translations           []Translation
	AppOrigin              string
	NotFoundPageSource     string
	DefaultLayoutSource    string
	IsProductionEnviroment bool
	GlobalRenderContext    map[string]any
	EnableSpaMode          bool
}

type VideoData struct {
	Title           string
	Description     string
	ThumbnailLoc    string
	ContentLoc      string
	Duration        int
	PublicationDate string
}

type Route struct {
	Path             string
	Source           string
	TemplateType     string
	JavascriptDeps   []string
	LayoutSource     string
	PageTitle        string
	StaticRenderData map[string]any
	SitemapVideoData *VideoData
}

type Translation struct {
	Code       string
	Source     string
	SourceType string
	IsDefault  bool
}
