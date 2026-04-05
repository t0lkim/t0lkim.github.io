---
title: "CourseVault — Case Study"
description: "A Go CLI for downloading and archiving course content from Teachable platforms."
hero_title: "Course<span class=\"accent\">Vault</span>"
hero_sub: "A Go CLI that authenticates via browser cookies, enumerates curriculum, and downloads course videos with full metadata."
---

## The Problem

### Paid courses disappear

Online courses hosted on Teachable have no built-in download or export feature. You pay for the content, but you never own it. If your subscription lapses, the instructor removes the course, or the platform changes its terms, you lose access to material you paid for.

There is no offline viewing, no archival option, and no API-level export. The only way to keep a copy is to manually screen-record each lecture one at a time.

> No tool existed to authenticate against Teachable, enumerate a full course curriculum, download every video at original quality, and preserve the course structure with metadata — all from a single command.

- No download button on Teachable
- Subscriptions lapse, content disappears
- Platform changes break access without notice
- Manual screen recording loses quality and structure
- No machine-readable manifest of course contents

---

## The Solution

### One command, full course archive

CourseVault is a Go CLI that authenticates using cookies extracted directly from the Arc browser, walks the Teachable JSON API to enumerate every section and lecture, downloads videos at full CDN quality via yt-dlp, embeds MP4 metadata via ffmpeg, and writes a structured manifest file.

The output is a directory tree that mirrors the course structure, with numbered sections and lectures, tagged video files, and a `course-info.json` manifest containing the full course metadata and attachment inventory.

```
$ coursevault https://school.teachable.com/courses/enrolled/12345 ~/Videos/MyCourse

🎓 Teachable Downloader
   School: school.teachable.com
   Output: ~/Videos/MyCourse

🔑 Extracting cookies from Arc browser...
   Found 12 cookies for school.teachable.com

📋 Fetching course curriculum...
   Found 8 sections, 47 lectures

📹 [1/47] Introduction to the Course
   ✅ Saved
📹 [2/47] Setting Up Your Environment
   ⏭️  Already exists, skipping
...
📄 Wrote ~/Videos/MyCourse/course-info.json

🎉 Done! Downloaded 47/47 lectures to ~/Videos/MyCourse
```

---

## How It Works

### Five-stage pipeline

**1. Cookie Extraction**

CourseVault reads the Arc browser's SQLite cookie database from disk. Arc (Chromium-based) encrypts cookie values using AES-128-CBC with a key derived from the macOS Keychain. The tool retrieves the encryption key via the `security` CLI, derives the AES key using PBKDF2 with Chromium's standard parameters (1003 iterations, SHA-1, `saltysalt`), and decrypts each cookie's v10-format ciphertext: 3-byte prefix, 16-byte nonce, 16-byte IV, then CBC-encrypted payload with PKCS#7 padding.

The cookie database is copied to a temp file before reading to avoid SQLite lock conflicts with the running browser.

**2. API Enumeration**

Using the extracted session cookies, CourseVault calls Teachable's JSON API to retrieve the course structure:

- `/api/v1/courses/{id}/lecture_sections` for section names and ordering
- `/api/v1/courses/{id}/lectures` for lecture titles, positions, and section assignments
- `/api/v1/courses/{id}` for course-level metadata (title, description, author, image)

Sections and lectures are sorted by position. Unpublished lectures are skipped. Empty sections are pruned from the output.

**3. Video Download**

For each lecture, CourseVault fetches `/api/v1/courses/{id}/lectures/{id}/attachments` to find video attachments. The CDN URL (or fallback URL) is passed to yt-dlp with the session cookies written as a Netscape-format cookie file. yt-dlp handles HLS streams, Wistia embeds, and direct MP4 downloads, merging to MP4 output format.

**4. Metadata Embedding**

After each video downloads, ffmpeg writes MP4 metadata tags into the container: `title` (lecture name), `artist` (course author), `album` (course title), and `track` (lecture number within its section). The original file is replaced atomically via a temp file rename.

**5. Manifest Generation**

After all lectures are processed, CourseVault writes `course-info.json` to the output directory. This manifest contains the full course structure — title, description, author, image URL, course URL, course ID — with nested sections, lectures, and per-lecture attachment metadata (name, kind, content type, URL, CDN URL, file size).

---

## Design Decisions

### Why it works this way

- **Go over Rust for rapid prototyping** — This started as a weekend tool to solve a personal problem. Go's `database/sql` with the `go-sqlite3` CGo binding gave immediate SQLite access for cookie extraction. The compile-deploy cycle was fast enough that the entire tool was functional in a single session.

- **Arc cookie extraction over manual login** — Teachable uses standard session cookies for authentication. Extracting them from the browser avoids implementing OAuth flows, CSRF token handling, or credential storage. You log into Teachable once in Arc, then CourseVault reuses that session. The decryption handles Chromium's v10 format with the 16-byte nonce prepended before the IV — a detail not documented in most Chromium cookie extraction guides.

- **yt-dlp for video download over raw HTTP** — Teachable courses may host videos on Wistia, Vimeo, or their own CDN, using HLS, DASH, or direct MP4 URLs. yt-dlp handles all of these transparently. Writing a custom downloader for each hosting provider would be fragile and unnecessary when yt-dlp already does it correctly.

- **ffmpeg for metadata over custom MP4 parsing** — MP4 metadata (atoms/boxes) is a non-trivial binary format. ffmpeg's `-metadata` flag writes standard tags with a single command, using stream copy (`-c copy`) so no transcoding occurs. The operation adds milliseconds per file.

- **JSON manifest for resume and inventory** — The `course-info.json` file serves two purposes: it enables resume support (CourseVault checks for existing files on disk before downloading), and it provides a machine-readable inventory of the entire course including non-video attachments like PDFs and slides that could be downloaded in a future version.

- **Netscape cookie file for yt-dlp** — yt-dlp expects cookies in the legacy Netscape format. CourseVault writes a temp file with the decrypted cookies in this format and passes it via `--cookies`, then cleans up the file after the download completes.

---

## Current Status

**Project Status:** Complete — v1.0.0

| | |
|---|---|
| **Language** | Go |
| **License** | MIT |

CourseVault v1.0.0 downloads full Teachable courses with video, metadata, and structured manifests. It handles resume on re-run, verbose debugging output, and works with any Teachable-based platform where you have an active session in Arc.

---

[Source on GitHub](https://github.com/t0lkim/course-vault)

(c) 2026 t0lkim
