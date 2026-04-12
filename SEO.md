# SEO & AEO Strategy — t0lkim.dev

This document covers the SEO (Search Engine Optimisation) and AEO (Answer Engine Optimisation) infrastructure for t0lkim.dev, including the generation tool, configuration schema, and quarterly audit schedule.

## Architecture Overview

```
projects.yaml          <-- Single source of truth for all project metadata
    |
    v
tools/seogen/seogen    <-- Go CLI that generates SEO files from config
    |
    +-- docs/{slug}/index.html   <-- Regenerates <head> section (OG, Twitter, canonical, JSON-LD)
    +-- docs/sitemap.xml         <-- Generated sitemap with all published pages
    +-- docs/llms.txt            <-- AEO: compact site summary for LLMs
    +-- docs/llms-full.txt       <-- AEO: detailed Q&A content for answer engines
```

The Jekyll-processed pages (index.md, privacy-policy.md) use `_includes/head.html` for their SEO meta. The standalone HTML project pages are managed by the `seogen` tool.

## Generation Tool: seogen

A Go CLI that reads `projects.yaml` and generates all SEO/AEO metadata.

### Usage

```bash
cd /path/to/t0lkim.github.io
./tools/seogen/seogen                      # uses ./projects.yaml by default
./tools/seogen/seogen -config ./projects.yaml  # explicit config path
```

The tool:
1. Reads all project definitions from the YAML config
2. For each project with an existing `docs/{slug}/index.html`, regenerates the entire `<head>` section
3. Generates `docs/sitemap.xml`, `docs/llms.txt`, and `docs/llms-full.txt`

Pages without an existing `index.html` in `docs/` are skipped (head injection only runs on existing files) but still appear in llms.txt and sitemap.

### Building

```bash
cd tools/seogen
go build -o seogen .
```

Requires Go 1.26+ and one external dependency: `gopkg.in/yaml.v3`.

## projects.yaml Schema

### Top-level structure

```yaml
site:
  url: https://t0lkim.dev        # Base URL (no trailing slash)
  title: t0lkim.dev              # Site title for OG meta
  author: Mike Lott              # Author name
  author_url: https://t0lkim.dev # Author URL for JSON-LD
  author_handle: t0lkim          # Used for LinkedIn, GitHub, email URLs
  locale: en_SG                  # OG locale
  theme_color: "#00C9A7"         # theme-color meta (not currently used by seogen)
  description: "..."             # Site description
  bio: "..."                     # Bio paragraph for llms.txt

projects:
  - name: TeenTidal              # Display name
    slug: teen-tidal             # URL slug (no slashes, no ..)
    title: "TeenTidal — Case Study"  # <title> and OG title
    description: "..."           # meta description and OG description
    keywords: "keyword1, ..."    # meta keywords (comma-separated string)
    language: Swift              # string OR list of strings
    category: Parental Controls  # JSON-LD applicationCategory
    operating_system: iOS        # Optional: JSON-LD operatingSystem
    image: "..."                 # Optional: OG image URL
    css: teentidal.css           # CSS filename (in assets/css/)
    page_style: "/* overrides */" # Optional: inline <style> content
    sitemap_priority: "0.8"      # Sitemap priority
    faq:                         # FAQ entries for JSON-LD FAQPage schema
      - q: "What is TeenTidal?"
        a: "TeenTidal is..."
    llms_summary: "..."          # One-line summary for llms.txt
    llms_full: |                 # Multi-paragraph content for llms-full.txt
      **What is X?** ...

extra_pages:                     # Non-project sitemap entries
  - url: "https://t0lkim.dev/teen-tidal/privacy-policy"
    lastmod: "2026-04-05"
    changefreq: yearly
    priority: "0.3"
```

### Language field

The `language` field accepts either a single string or a list:

```yaml
# Single language
language: Rust

# Multiple languages
language:
  - Python
  - Bash
```

In JSON-LD output, single languages render as `"Rust"` and multiple as `["Python", "Bash"]`.

## Adding a New Project

1. Add a new entry to `projects.yaml` with all required fields
2. Create the project page HTML at `docs/{slug}/index.html` (body content)
3. Run `./tools/seogen/seogen` to generate the `<head>` section and update sitemap/llms files
4. Commit and push

If the HTML page doesn't exist yet, the project will appear in sitemap.xml and llms.txt but no head injection occurs.

## SEO Infrastructure

### Per-page metadata (standalone HTML)
- OpenGraph meta (title, description, url, site_name, type, locale, image)
- Twitter card meta (card, title, description)
- Canonical URL
- Keywords meta
- Author meta
- JSON-LD `@graph` with `SoftwareSourceCode` + `FAQPage` schemas

### Per-page metadata (Jekyll pages)
- `_includes/head.html` — shared head include with OG, Twitter, canonical, robots, theme-color
- `_layouts/profile.html` — adds JSON-LD `Person` schema for the profile page
- Frontmatter-driven: `title`, `description`, `keywords`, `image` fields

### Discoverability files
- `robots.txt` — allows all crawlers, explicitly welcomes AI crawlers (GPTBot, OAI-SearchBot, PerplexityBot, ClaudeBot, Google-Extended), blocks Bytespider
- `sitemap.xml` — all published page URLs with lastmod and priority
- `llms.txt` — compact site and project summary following the llms.txt spec
- `llms-full.txt` — detailed Q&A content per project for answer engine extraction

## Quarterly AEO Audit

Three remote scheduled triggers audit the site's SEO/AEO posture quarterly:

| Quarter | Date | Trigger ID |
|---------|------|------------|
| Q3 2026 | July 1 2026, 10am SGT | `trig_015D7bCxJ4xZAEnVTkvLh3XX` |
| Q4 2026 | Oct 1 2026, 10am SGT | `trig_01F5etWFgziLAk2nkgnUopQZ` |
| Q1 2027 | Jan 2 2027, 10am SGT | `trig_01GZruAoLgR4ZEcVcBQXABcv` |

Each audit:
1. Researches current algorithm changes and AEO best practices
2. Checks if llms.txt spec has evolved
3. Checks for new AI crawlers to add to robots.txt
4. Reviews JSON-LD schema recommendations
5. Compares current site structured data against best practices
6. Writes a report to `docs/SEO-AUDIT-{quarter}.md`

Audits do NOT make changes — they produce findings for review. Each audit references prior reports for continuity.

Manage triggers at: https://claude.ai/code/scheduled

## Known Limitations

- `seogen` uses `text/template` (not `html/template`) because `html/template` applies JS escaping inside `<script>` tags, which would break JSON-LD output. Since the YAML config is owner-controlled, this is acceptable. Do not use untrusted input in `projects.yaml`.
- Three project directories (mmm, course-vault, patron-vault) are empty stubs — `seogen` skips them for head injection but includes them in llms.txt and sitemap.
