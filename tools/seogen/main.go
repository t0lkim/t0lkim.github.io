package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	configPath := flag.String("config", "./projects.yaml", "path to projects.yaml config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "seogen: load config: %v\n", err)
		os.Exit(1)
	}

	// The docs/ directory is relative to the directory containing the config file.
	configDir := filepath.Dir(*configPath)
	docsDir := filepath.Join(configDir, "docs")

	// Regenerate HTML <head> for every project that has an index.html on disk.
	for _, p := range cfg.Projects {
		htmlPath := filepath.Join(docsDir, p.Slug, "index.html")
		if _, statErr := os.Stat(htmlPath); os.IsNotExist(statErr) {
			fmt.Printf("seogen: skip %s (no index.html found)\n", p.Slug)
			continue
		}
		if err := processProjectHTML(docsDir, p, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "seogen: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("seogen: updated %s/index.html\n", p.Slug)
	}

	// Generate sitemap.xml.
	if err := generateSitemap(docsDir, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "seogen: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("seogen: updated sitemap.xml")

	// Generate compact llms.txt.
	if err := generateLLMSTxt(docsDir, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "seogen: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("seogen: updated llms.txt")

	// Generate verbose llms-full.txt.
	if err := generateLLMSFullTxt(docsDir, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "seogen: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("seogen: updated llms-full.txt")
}
