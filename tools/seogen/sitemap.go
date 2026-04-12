package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// sitemapData is passed to tmplSitemap.
type sitemapData struct {
	BaseURL    string
	LastMod    string
	Projects   []sitemapEntry
	ExtraPages []sitemapEntry
}

// sitemapEntry is a single <url> block inside sitemap.xml.
type sitemapEntry struct {
	URL        string
	LastMod    string
	ChangeFreq string
	Priority   string
}

// generateSitemap writes sitemap.xml into docsDir.
// Only projects whose index.html exists on disk are included.
func generateSitemap(docsDir string, cfg *Config) error {
	today := time.Now().UTC().Format("2006-01-02")

	var projectEntries []sitemapEntry
	for _, p := range cfg.Projects {
		htmlPath := filepath.Join(docsDir, p.Slug, "index.html")
		if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
			continue
		}
		priority := p.SitemapPriority
		if priority == "" {
			priority = "0.8"
		}
		projectEntries = append(projectEntries, sitemapEntry{
			URL:      cfg.Site.URL + "/" + p.Slug + "/",
			LastMod:  today,
			Priority: priority,
		})
	}

	var extraEntries []sitemapEntry
	for _, ep := range cfg.ExtraPages {
		lastmod := ep.LastMod
		if lastmod == "" {
			lastmod = today
		}
		changefreq := ep.ChangeFreq
		if changefreq == "" {
			changefreq = "yearly"
		}
		priority := ep.Priority
		if priority == "" {
			priority = "0.3"
		}
		extraEntries = append(extraEntries, sitemapEntry{
			URL:        ep.URL,
			LastMod:    lastmod,
			ChangeFreq: changefreq,
			Priority:   priority,
		})
	}

	data := sitemapData{
		BaseURL:    cfg.Site.URL,
		LastMod:    today,
		Projects:   projectEntries,
		ExtraPages: extraEntries,
	}

	var buf bytes.Buffer
	if err := tmplSitemap.Execute(&buf, data); err != nil {
		return fmt.Errorf("render sitemap template: %w", err)
	}

	outPath := filepath.Join(docsDir, "sitemap.xml")
	if err := atomicWrite(outPath, buf.Bytes()); err != nil {
		return fmt.Errorf("write sitemap.xml: %w", err)
	}

	return nil
}
