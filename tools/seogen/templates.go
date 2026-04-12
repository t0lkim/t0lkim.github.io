package main

import "text/template"

// tmplHead is executed once per project to produce the <head>…</head> block.
// It receives a headData value.
var tmplHead = template.Must(template.New("head").Funcs(template.FuncMap{
	// langValue returns the programmingLanguage value for JSON-LD:
	// a bare JSON string for one language, a JSON array for multiple.
	"langValue": func(langs []string) string {
		if len(langs) == 1 {
			return `"` + jsonEscape(langs[0]) + `"`
		}
		out := "["
		for i, l := range langs {
			if i > 0 {
				out += ", "
			}
			out += `"` + jsonEscape(l) + `"`
		}
		out += "]"
		return out
	},
}).Parse(`<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}}</title>
  <meta name="description" content="{{.Description}}">
  <meta name="keywords" content="{{.Keywords}}">
  <meta name="author" content="{{.Author}}">
  <link rel="canonical" href="{{.URL}}">
  <!-- Open Graph -->
  <meta property="og:title" content="{{.Title}}">
  <meta property="og:description" content="{{.Description}}">
  <meta property="og:url" content="{{.URL}}">
  <meta property="og:site_name" content="{{.SiteName}}">
  <meta property="og:type" content="article">
  <meta property="og:locale" content="{{.Locale}}">{{if .OGImage}}
  <meta property="og:image" content="{{.OGImage}}">{{end}}
  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary">
  <meta name="twitter:title" content="{{.Title}}">
  <meta name="twitter:description" content="{{.Description}}">
  <!-- JSON-LD -->
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "SoftwareSourceCode",
        "name": "{{.ProjectName}}",
        "description": "{{.JSONLDDescription}}",
        "programmingLanguage": {{langValue .Languages}},
        "author": { "@type": "Person", "name": "{{.Author}}", "url": "{{.AuthorURL}}" },
        "url": "{{.URL}}",
        "applicationCategory": "{{.Category}}"{{if .OperatingSystem}},
        "operatingSystem": "{{.OperatingSystem}}"{{end}}
      },
      {
        "@type": "FAQPage",
        "mainEntity": [
          {{range $i, $faq := .FAQs}}{{if $i}},
          {{end}}{
            "@type": "Question",
            "name": "{{$faq.Question}}",
            "acceptedAnswer": {
              "@type": "Answer",
              "text": "{{$faq.Answer}}"
            }
          }{{end}}
        ]
      }
    ]
  }
  </script>
  <link rel="stylesheet" href="../assets/css/base.css">
  <link rel="stylesheet" href="../assets/css/{{.CSS}}">
  <style>{{.PageStyle}}</style>
</head>`))

// tmplSitemap is executed once to produce sitemap.xml.
// It receives a sitemapData value.
var tmplSitemap = template.Must(template.New("sitemap").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>{{.BaseURL}}/</loc>
    <lastmod>{{.LastMod}}</lastmod>
    <changefreq>monthly</changefreq>
    <priority>1.0</priority>
  </url>
{{range .Projects}}  <url>
    <loc>{{.URL}}</loc>
    <lastmod>{{.LastMod}}</lastmod>
    <changefreq>monthly</changefreq>
    <priority>{{.Priority}}</priority>
  </url>
{{end}}{{range .ExtraPages}}  <url>
    <loc>{{.URL}}</loc>
    <lastmod>{{.LastMod}}</lastmod>
    <changefreq>{{.ChangeFreq}}</changefreq>
    <priority>{{.Priority}}</priority>
  </url>
{{end}}</urlset>`))

// tmplLlms is executed once to produce the compact llms.txt.
// It receives an llmsData value.
var tmplLlms = template.Must(template.New("llms").Parse(`# {{.SiteTitle}}

> {{.SiteDescription}}

## About the Author

{{.Bio}}

## Projects

{{range .Projects}}- [{{.Name}}]({{.URL}}): {{.Summary}}
{{end}}
## Links

- Website: {{.BaseURL}}
- LinkedIn: https://www.linkedin.com/in/{{.AuthorHandle}}/
- GitHub: https://github.com/{{.AuthorHandle}}
- Email: me@{{.AuthorHandle}}.dev

## Optional

- Full content: {{.BaseURL}}/llms-full.txt
`))

// tmplLlmsFull is executed once to produce the verbose llms-full.txt.
// It receives an llmsData value.
var tmplLlmsFull = template.Must(template.New("llms-full").Parse(`# {{.SiteTitle}} — Full Content

> {{.SiteDescription}}

## About {{.Author}}

{{.Bio}}

---

{{range .Projects}}## Project: {{.Name}}

{{.Full}}
---

{{end}}`))
