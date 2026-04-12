package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// llmsData is passed to both tmplLlms and tmplLlmsFull.
type llmsData struct {
	SiteTitle       string
	SiteDescription string
	Bio             string
	Author          string
	AuthorHandle    string
	BaseURL         string
	Projects        []llmsProject
}

// llmsProject holds per-project data used in LLM files.
type llmsProject struct {
	Name    string
	URL     string
	Summary string
	Full    string
}

// buildLLMSData constructs the shared data structure used by both LLM file templates.
// Only projects whose index.html exists on disk are included.
func buildLLMSData(docsDir string, cfg *Config) llmsData {
	var projects []llmsProject
	for _, p := range cfg.Projects {
		htmlPath := filepath.Join(docsDir, p.Slug, "index.html")
		if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
			continue
		}
		projects = append(projects, llmsProject{
			Name:    p.Name,
			URL:     cfg.Site.URL + "/" + p.Slug + "/",
			Summary: p.LLMSSummary,
			Full:    strings.TrimRight(p.LLMSFull, "\n"),
		})
	}

	return llmsData{
		SiteTitle:       cfg.Site.Title,
		SiteDescription: cfg.Site.Description,
		Bio:             cfg.Site.Bio,
		Author:          cfg.Site.Author,
		AuthorHandle:    cfg.Site.AuthorHandle,
		BaseURL:         cfg.Site.URL,
		Projects:        projects,
	}
}

// generateLLMSTxt writes the compact llms.txt to docsDir.
func generateLLMSTxt(docsDir string, cfg *Config) error {
	data := buildLLMSData(docsDir, cfg)

	var buf bytes.Buffer
	if err := tmplLlms.Execute(&buf, data); err != nil {
		return fmt.Errorf("render llms template: %w", err)
	}

	outPath := filepath.Join(docsDir, "llms.txt")
	if err := atomicWrite(outPath, buf.Bytes()); err != nil {
		return fmt.Errorf("write llms.txt: %w", err)
	}

	return nil
}

// generateLLMSFullTxt writes the verbose llms-full.txt to docsDir.
func generateLLMSFullTxt(docsDir string, cfg *Config) error {
	data := buildLLMSData(docsDir, cfg)

	var buf bytes.Buffer
	if err := tmplLlmsFull.Execute(&buf, data); err != nil {
		return fmt.Errorf("render llms-full template: %w", err)
	}

	outPath := filepath.Join(docsDir, "llms-full.txt")
	if err := atomicWrite(outPath, buf.Bytes()); err != nil {
		return fmt.Errorf("write llms-full.txt: %w", err)
	}

	return nil
}
