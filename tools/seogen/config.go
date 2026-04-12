package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top-level structure parsed from projects.yaml.
type Config struct {
	Site       SiteMeta    `yaml:"site"`
	Projects   []Project   `yaml:"projects"`
	ExtraPages []ExtraPage `yaml:"extra_pages"`
}

// SiteMeta holds site-wide metadata used across all generated files.
type SiteMeta struct {
	URL          string `yaml:"url"`
	Title        string `yaml:"title"`
	Author       string `yaml:"author"`
	AuthorURL    string `yaml:"author_url"`
	AuthorHandle string `yaml:"author_handle"`
	Locale       string `yaml:"locale"`
	ThemeColor   string `yaml:"theme_color"`
	Description  string `yaml:"description"`
	Bio          string `yaml:"bio"`
}

// Project holds per-project metadata used for HTML head injection, sitemap, and LLM files.
type Project struct {
	Name            string    `yaml:"name"`
	Slug            string    `yaml:"slug"`
	Title           string    `yaml:"title"`
	Description     string    `yaml:"description"`
	Keywords        string    `yaml:"keywords"`
	Language        LangField `yaml:"language"`
	Category        string    `yaml:"category"`
	OperatingSystem string    `yaml:"operating_system"`
	Image           string    `yaml:"image"`
	CSS             string    `yaml:"css"`
	PageStyle       string    `yaml:"page_style"`
	SitemapPriority string    `yaml:"sitemap_priority"`
	FAQ             []FAQ     `yaml:"faq"`
	LLMSSummary     string    `yaml:"llms_summary"`
	LLMSFull        string    `yaml:"llms_full"`
}

// FAQ is a single question-and-answer pair for a project page.
type FAQ struct {
	Question string `yaml:"q"`
	Answer   string `yaml:"a"`
}

// ExtraPage is a sitemap entry for pages that are not generated project pages.
type ExtraPage struct {
	URL        string `yaml:"url"`
	LastMod    string `yaml:"lastmod"`
	ChangeFreq string `yaml:"changefreq"`
	Priority   string `yaml:"priority"`
}

// LangField accepts either a single YAML scalar or a YAML sequence for the
// "language" field, normalising both forms to []string.
type LangField []string

// UnmarshalYAML implements yaml.Unmarshaler for LangField, accepting both
// a bare string ("Swift") and a sequence (["Python", "Bash"]).
func (l *LangField) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		*l = []string{value.Value}
		return nil
	}
	var list []string
	if err := value.Decode(&list); err != nil {
		return err
	}
	*l = list
	return nil
}

// loadConfig reads and unmarshals the YAML config at the given path.
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// Validate slugs to prevent path traversal.
	for _, p := range cfg.Projects {
		if strings.Contains(p.Slug, "..") || strings.Contains(p.Slug, "/") {
			return nil, fmt.Errorf("invalid slug %q: must not contain '..' or '/'", p.Slug)
		}
	}
	return &cfg, nil
}
