package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// headData is passed to tmplHead for rendering a project's <head> block.
type headData struct {
	Title             string
	Description       string
	Keywords          string
	Author            string
	AuthorURL         string
	URL               string
	SiteName          string
	Locale            string
	OGImage           string
	ProjectName       string
	JSONLDDescription string
	Languages         []string
	Category          string
	OperatingSystem   string
	CSS               string
	PageStyle         string
	FAQs              []faqItem
}

// faqItem holds a single FAQ entry for template rendering.
type faqItem struct {
	Question string
	Answer   string
}

// jsonEscape escapes a string for safe embedding inside a JSON string literal.
// It handles the minimal set required for well-formed JSON.
func jsonEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

// processProjectHTML regenerates the <head> block of a single project's index.html,
// preserving everything from <body onward verbatim.
func processProjectHTML(docsDir string, p Project, cfg *Config) error {
	htmlPath := filepath.Join(docsDir, p.Slug, "index.html")

	raw, err := os.ReadFile(htmlPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", htmlPath, err)
	}

	original := string(raw)

	// Locate <body using a case-insensitive search on a lowercased copy, then
	// slice the original string so capitalisation is preserved in the output.
	lower := strings.ToLower(original)
	bodyIdx := strings.Index(lower, "<body")
	if bodyIdx < 0 {
		return fmt.Errorf("%s: could not locate <body tag", htmlPath)
	}
	bodyOnward := original[bodyIdx:]

	// Build the template data.
	projectURL := cfg.Site.URL + "/" + p.Slug + "/"

	faqs := make([]faqItem, len(p.FAQ))
	for i, f := range p.FAQ {
		faqs[i] = faqItem{
			Question: jsonEscape(f.Question),
			Answer:   jsonEscape(f.Answer),
		}
	}

	ogImage := ""
	if p.Image != "" {
		ogImage = cfg.Site.URL + "/" + p.Image
	}

	// The JSON-LD description uses the first FAQ answer as a compact description
	// when no dedicated field exists — fall back to the page description.
	jsonLDDesc := p.Description
	if len(p.FAQ) > 0 && p.FAQ[0].Answer != "" {
		jsonLDDesc = p.FAQ[0].Answer
	}

	data := headData{
		Title:             p.Title,
		Description:       p.Description,
		Keywords:          p.Keywords,
		Author:            cfg.Site.Author,
		AuthorURL:         cfg.Site.AuthorURL,
		URL:               projectURL,
		SiteName:          cfg.Site.Title,
		Locale:            cfg.Site.Locale,
		OGImage:           ogImage,
		ProjectName:       p.Name,
		JSONLDDescription: jsonEscape(jsonLDDesc),
		Languages:         []string(p.Language),
		Category:          p.Category,
		OperatingSystem:   p.OperatingSystem,
		CSS:               p.CSS,
		PageStyle:         p.PageStyle,
		FAQs:              faqs,
	}

	var headBuf bytes.Buffer
	if err := tmplHead.Execute(&headBuf, data); err != nil {
		return fmt.Errorf("render head template for %s: %w", p.Slug, err)
	}

	newContent := "<!DOCTYPE html>\n<html lang=\"en\">\n" + headBuf.String() + "\n" + bodyOnward

	if err := atomicWrite(htmlPath, []byte(newContent)); err != nil {
		return fmt.Errorf("write %s: %w", htmlPath, err)
	}

	return nil
}

// atomicWrite writes data to path via a temporary file and an os.Rename,
// ensuring a partial write never leaves a corrupt file in place.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".seogen-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
