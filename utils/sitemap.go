package utils

import (
	"encoding/xml"
	"fmt"
	"os"
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

	// Add routes to sitemap
	for _, route := range routes {
		if strings.Contains(route, "{lang:en|es}") {
			// Generate URLs for both languages
			for _, lang := range []string{"en", "es"} {
				langRoute := strings.Replace(route, "{lang:en|es}", lang, 1)
				url := Url{
					Loc:     fmt.Sprintf("%s%s", baseURL, langRoute),
					LastMod: time.Now().Format("2006-01-02"),
				}
				sitemap.Urls = append(sitemap.Urls, url)
			}
		} else {
			// Add non-language specific routes
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
