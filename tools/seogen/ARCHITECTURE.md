# seogen — Architecture Specification

## Overview

`seogen` is a single-binary Go CLI tool that regenerates SEO metadata for a static Jekyll site. One external dependency: `gopkg.in/yaml.v3`.

---

## Package / File Structure

```
seogen/
├── main.go          — flag parsing, orchestration, exits
├── config.go        — Config, Project, FAQ types; YAML unmarshalling
├── html.go          — HTML head regeneration logic
├── sitemap.go       — sitemap.xml generation
├── llms.go          — llms.txt and llms-full.txt generation
├── templates.go     — all text/template definitions (head, sitemap, llms)
├── go.mod
├── go.sum
└── ARCHITECTURE.md
```

Flat package (`package main`). No sub-packages — the tool is small enough that separation by file is sufficient.

---

## Key Types

```go
// config.go

type Config struct {
    Site     SiteMeta  `yaml:"site"`
    Projects []Project `yaml:"projects"`
}

type SiteMeta struct {
    BaseURL     string `yaml:"base_url"`
    Author      string `yaml:"author"`
    Title       string `yaml:"title"`
    Description string `yaml:"description"`
    TwitterHandle string `yaml:"twitter_handle"`
}

type Project struct {
    Slug        string      `yaml:"slug"`
    Title       string      `yaml:"title"`
    Description string      `yaml:"description"`
    Keywords    []string    `yaml:"keywords"`
    Language    LangField   `yaml:"language"`
    CSS         string      `yaml:"css"`        // e.g. "teentidal.css"
    OGImage     string      `yaml:"og_image"`
    FAQs        []FAQ       `yaml:"faqs"`
    DatePublished string    `yaml:"date_published"`
    DateModified  string    `yaml:"date_modified"`
}

type FAQ struct {
    Question string `yaml:"question"`
    Answer   string `yaml:"answer"`
}
```

---

## Language Field: String vs []string

`LangField` is a custom type that implements `yaml.Unmarshaler`:

```go
type LangField []string

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
```

Downstream code always operates on `[]string` — no special-casing required elsewhere.

---

## HTML Injection Strategy

`html.go` performs a purely textual split — no HTML parser dependency.

1. Read the entire file into a `string`.
2. Find `<body` (case-insensitive, using `strings.Index` on a lowercased copy for the offset, then slicing the original). Everything from that offset onward is `bodyOnward`.
3. Generate a fresh `<head>…</head>` block by executing the head template (see below).
4. Write `<!DOCTYPE html>\n<html lang="en">\n` + newHead + `\n` + bodyOnward back to the file (atomic write via `os.WriteFile` to a temp file then `os.Rename`).

This avoids any dependency on `golang.org/x/net/html` and is robust because Jekyll's generated HTML is consistent and machine-produced.

---

## Template Approach

All templates live in `templates.go` as package-level `*template.Template` values, initialised via `template.Must(template.New(...).Parse(...))` with raw string literals.

| Template       | Output               | Notes                                          |
|----------------|----------------------|------------------------------------------------|
| `tmplHead`     | `<head>…</head>`     | Receives `headData` struct; includes JSON-LD `@graph` with `SoftwareSourceCode` + `FAQPage` nodes |
| `tmplSitemap`  | `sitemap.xml`        | Iterates over projects whose `index.html` exists; uses `lastmod` from `DateModified` |
| `tmplLlms`     | `llms.txt`           | Compact: one entry per project, title + URL + one-line description |
| `tmplLlmsFull` | `llms-full.txt`      | Full: title, URL, description, keywords, FAQs  |

The `headData` struct is assembled in `html.go` from `Config` + `Project` before template execution. Stylesheet hrefs (`../assets/css/base.css`, `../assets/css/{project.css}`) are rendered last inside `<head>`, before the closing tag, preserving the per-page `<style>` block which is extracted from the original `<head>` via a simple `<style>…</style>` substring search before the split.
