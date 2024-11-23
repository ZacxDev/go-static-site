package utils

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ZacxDev/go-static-site/config"
)

type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	VideoNs string   `xml:"xmlns:video,attr"`
	Urls    []Url    `xml:"url"`
}

type Video struct {
	XMLName         xml.Name `xml:"video:video"`
	Title           string   `xml:"video:title"`
	Description     string   `xml:"video:description"`
	ThumbnailLoc    string   `xml:"video:thumbnail_loc"`
	ContentLoc      string   `xml:"video:content_loc"`
	Duration        int      `xml:"video:duration,omitempty"`
	PublicationDate string   `xml:"video:publication_date"`
}

type Url struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
	Video      *Video `xml:"video:video,omitempty"`
}

func GenerateSitemaps(routes []string, origin string, manifestRoutes []config.Route) error {
	xmlOutput, err := GenerateSitemapContent(routes, origin, manifestRoutes)
	if err != nil {
		return err
	}

	xmlFile, err := os.Create("public/sitemap.xml")
	if err != nil {
		return err
	}
	defer xmlFile.Close()

	xmlFile.Write([]byte(xml.Header))
	xmlFile.Write([]byte(xmlOutput))

	return nil
}

func GenerateSitemapContent(routes []string, origin string, manifestRoutes []config.Route) (string, error) {
	baseURL := origin
	sitemap := Sitemap{
		Xmlns:   "http://www.sitemaps.org/schemas/sitemap/0.9",
		VideoNs: "http://www.google.com/schemas/sitemap-video/1.1",
	}

	// Regex to match language pattern in routes
	langPattern := regexp.MustCompile(`\/\{lang:([^}]+)\}\/`)

	// Process each route
	for _, route := range routes {
		matches := langPattern.FindStringSubmatch(route)
		if len(matches) > 1 {
			// Extract supported languages from the pattern
			langs := strings.Split(matches[1], "|")

			// Base route without the language pattern
			baseRoute := langPattern.ReplaceAllString(route, "/")

			// Generate URLs for each supported language
			for _, lang := range langs {
				url := Url{
					Loc:     fmt.Sprintf("%s/%s%s", baseURL, lang, baseRoute),
					LastMod: time.Now().Format("2006-01-02"),
				}

				// Add video data if available in the route configuration
				for _, r := range manifestRoutes {
					if r.Path == baseRoute && r.SitemapVideoData != nil && r.SitemapVideoData.Title != "" {
						url.Video = &Video{
							Title:           r.SitemapVideoData.Title,
							Description:     r.SitemapVideoData.Description,
							ThumbnailLoc:    r.SitemapVideoData.ThumbnailLoc,
							ContentLoc:      r.SitemapVideoData.ContentLoc,
							PublicationDate: r.SitemapVideoData.PublicationDate,
						}

						if r.SitemapVideoData.Duration > 0 {
							url.Video.Duration = r.SitemapVideoData.Duration
						}
						break
					}
				}

				sitemap.Urls = append(sitemap.Urls, url)
			}
		} else {
			// Handle non-language specific routes
			// Skip if the route is empty or just a slash
			if route == "" || route == "/" {
				continue
			}

			url := Url{
				Loc:     fmt.Sprintf("%s%s", baseURL, route),
				LastMod: time.Now().Format("2006-01-02"),
			}

			// Add video data if available in the route configuration
			for _, r := range manifestRoutes {
				if r.Path == route && r.SitemapVideoData != nil && r.SitemapVideoData.Title != "" {
					url.Video = &Video{
						Title:           r.SitemapVideoData.Title,
						Description:     r.SitemapVideoData.Description,
						ThumbnailLoc:    r.SitemapVideoData.ThumbnailLoc,
						ContentLoc:      r.SitemapVideoData.ContentLoc,
						PublicationDate: r.SitemapVideoData.PublicationDate,
					}

					if r.SitemapVideoData.Duration > 0 {
						url.Video.Duration = r.SitemapVideoData.Duration
					}
					break
				}
			}

			sitemap.Urls = append(sitemap.Urls, url)
		}
	}

	// Generate XML sitemap
	xmlOutput, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		return "", err
	}

	return string(xmlOutput), nil
}
