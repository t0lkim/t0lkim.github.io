---
title: "PatronVault — Case Study"
description: "A Rust CLI for downloading and archiving media content from Patreon creators you subscribe to."
hero_title: "Patron<span class=\"accent\">Vault</span>"
hero_sub: "A Rust CLI that archives Patreon media you've paid for — because the platform won't let you."
---

## The Problem

### Patreon has no export feature for patrons

You pay creators monthly for access to exclusive content — images, videos, audio, archives. But Patreon provides no way to download or export that content in bulk. There is no "Download All" button. There is no archive feature. If a creator deletes a post, changes tiers, or leaves the platform, the content you paid for is gone.

The official Patreon API v2 does not help. It is creator-focused — managing campaigns, members, and post metadata — and provides no subscriber-facing endpoints for downloading patron-only media. This is a known limitation that has persisted for years.

> You pay for content but cannot archive it. The official API cannot access media. If the creator or platform disappears, so does everything you paid for.

- No bulk download or export feature exists
- Official API v2 has no media access endpoints
- Content disappears when creators delete posts or leave
- Tier changes can revoke access to already-paid content
- No first-party archival solution

---

## The Solution

### Cookie-authenticated CLI with manifest-based resume

PatronVault is a Rust CLI that uses Patreon's internal JSON:API v1 — the same API the web application calls when you browse posts in your browser. Authentication is a single `session_id` cookie extracted from a logged-in browser session.

The pipeline is straightforward: resolve the creator's campaign ID, paginate through all accessible posts, extract media URLs from JSON:API compound documents, download everything to an organised directory tree, and write a manifest for future resume runs.

```
$ patron-vault https://www.patreon.com/CreatorName --cookie "session_id=abc123"

Resolved campaign: CreatorName (ID: 67890)
Fetching posts... 150 posts across 8 pages
Downloading media...
  [1/312] 2026-03-01_New-Track/audio.mp3 (4.2 MB)
  [2/312] 2026-03-01_New-Track/cover.jpg (1.1 MB)
  ...
  [312/312] 2025-01-15_First-Post/image.png (856 KB)

Complete: 312 files, 2.8 GB total
Manifest saved to output/CreatorName/manifest.json
```

---

## Architecture

### Data flow

```
1. User provides creator URL + session_id cookie
                    |
2. resolve_campaign()
   +-- Fetch creator page HTML
   +-- Extract campaign_id from embedded JSON
   +-- Fallback: query /api/campaigns?filter[vanity]=NAME
                    |
3. all_posts(campaign_id)
   +-- GET /api/posts?filter[campaign_id]=ID&include=attachments,images,media,audio
   +-- Parse JSON:API response (data[] + included[])
   +-- Resolve relationships: post -> media items via included array
   +-- Filter: skip posts where current_user_can_view = false
   +-- Paginate via cursor until no more pages
   +-- Return Vec<(PostAttributes, Vec<MediaInfo>)>
                    |
4. download_all()
   +-- Create output/CreatorName/ directory
   +-- For each post: create YYYY-MM-DD_Post-Title/ subdirectory
   +-- For each media item: HTTP GET download_url -> write to file
   +-- Track progress, skip existing files if --skip-existing
   +-- Build manifest data
                    |
5. save_manifest()
   +-- Write manifest.json to output/CreatorName/
```

### Source structure

```
patron-vault
+-- src/
|   +-- main.rs          CLI entry point (clap argument parsing, orchestration)
|   +-- api.rs           Patreon internal API client (auth, pagination, media extraction)
|   +-- models.rs        Serde types for JSON:API v1 responses
|   +-- download.rs      File download, directory organisation, manifest management
+-- Cargo.toml
```

### Output structure

```
output/
+-- CreatorName/
    +-- 2026-03-01_Post-Title/
    |   +-- image1.jpg
    |   +-- image2.png
    |   +-- video.mp4
    +-- 2026-02-28_Another-Post/
    |   +-- audio.mp3
    +-- manifest.json
```

---

## Design Decisions

### Why it works this way

- **Internal API over official API v2** — The official Patreon API v2 does not support media or attachment access. This is not a workaround for convenience — it is the only viable path. Every community tool that downloads Patreon media (gallery-dl, patreon-dl-node, PatreonDownloader) uses the internal API for the same reason.

- **Cookie authentication over OAuth** — The internal API authenticates via the `session_id` cookie from a browser session. No API keys, no OAuth flow, no app registration. The user logs into Patreon in their browser, extracts the cookie, and passes it to the CLI. This avoids the official API's scope limitations entirely.

- **Synchronous ureq over async** — Downloads are intentionally sequential with deliberate delays between requests (500ms between API calls, 200ms between file downloads). An async runtime adds complexity for no benefit when you are rate-limiting yourself by design. ureq provides a clean synchronous HTTP client with no runtime overhead.

- **Manifest-based resume** — Every run writes a `manifest.json` tracking each post and file with its download status. Combined with `--skip-existing`, this enables resumable downloads across sessions. If a run is interrupted, the next run skips already-downloaded files and picks up where it left off.

- **Conservative rate limiting** — Patreon's internal API has no documented rate limits, but aggressive requests risk triggering Cloudflare protection or account flags. The tool imposes its own delays: 500ms between paginated API requests and 200ms between file downloads. Slow and steady over fast and banned.

- **Campaign resolution with fallback** — Creator pages embed the campaign ID in their HTML as JSON. The primary method parses this from the page source, checking two known patterns. If both fail, the tool falls back to querying the campaigns API endpoint by vanity name. Two methods means resilience against Patreon changing their page structure.

- **Date-prefixed directories** — Posts are stored in `YYYY-MM-DD_Post-Title/` subdirectories. This sorts chronologically in any file manager without needing metadata, makes the archive browsable without any tooling, and avoids filename collisions across posts.

---

## API Integration

### JSON:API v1 compound documents

Patreon's internal API returns JSON:API v1 compound documents — a format where related resources are not nested inside parent objects but placed in a separate top-level `included` array and linked via `relationships`.

A post listing request includes multiple resource types:

```
GET https://www.patreon.com/api/posts
  ?include=campaign,attachments,audio,images,media,native_video_insights,user
  &filter[campaign_id]={campaign_id}
  &filter[is_draft]=false
  &sort=-published_at
  &json-api-version=1.0
  &page[count]=20
  &page[cursor]={cursor}
```

The response separates posts from their media:

- `data[]` contains post resources with relationship references
- `included[]` contains the actual media objects, attachments, and other related resources
- `meta.pagination` provides cursor-based pagination

### Media resolution from the included array

Media objects are resolved through a three-step lookup:

1. A post's `relationships.images.data` contains references like `[{"id": "111", "type": "media"}]`
2. The matching resource is found in `included` where `id == "111"` and `type == "media"`
3. The `download_url` or `image_urls.original` is extracted from that resource's attributes

This is standard JSON:API compound document resolution — the same pattern regardless of whether the media is an image, video, audio file, or attachment.

### CDN download URLs

Media download URLs point directly to Patreon's CDN. They contain an embedded authentication token in the URL itself, are valid for approximately 24 hours, do not require the session cookie for the actual download, and support standard HTTP range requests.

### Pagination

Cursor-based pagination via `meta.pagination.cursors.next`. The cursor is an opaque string passed as `page[cursor]` on subsequent requests. When the cursor is `null` or absent, all pages have been fetched.

---

## Current Status

**Project Status:** v0.1.0

| | |
|---|---|
| **Language** | Rust |
| **Dependencies** | clap, serde, ureq, anyhow, chrono |
| **License** | MIT |

PatronVault v0.1.0 handles the complete pipeline: campaign resolution, paginated post fetching, JSON:API media extraction, file download with progress tracking, and manifest generation for resumable runs. Known limitations include manual session cookie refresh (no auto-renewal), no handling of externally-hosted embedded video (YouTube, Vimeo), and full metadata re-fetch on each run (only file downloads are skipped, not post enumeration).

---

[Source on GitHub](https://github.com/t0lkim/patron-vault)

(c) 2026 t0lkim
