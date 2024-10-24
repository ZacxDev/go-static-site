package utils

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	Urls    []Url    `xml:"url"`
}

type Url struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func GenerateSitemaps(routes []string) error {
	xmlOutput, err := GenerateSitemapContent(routes)
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

func GenerateSitemapContent(routes []string) (string, error) {
	baseURL := "https://mylinksprofile.com"
	sitemap := Sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
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
