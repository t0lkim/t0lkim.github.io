---
title: "TidalPlaylistBuilder — Case Study"
description: "Build Tidal playlists from BBC programmes, CSV files, and curated text track listings. Rust CLI with OAuth device flow and anthology support."
hero_title: "TidalPlaylistBuilder"
hero_sub: "Build Tidal playlists from BBC programmes, CSV files, and curated text track listings. Create multi-volume artist anthologies from simple text files."
---

## The Problem

### Manual playlist creation doesn't scale

Tidal has no way to import playlists from external sources. If you hear a track listing on a BBC programme, find a curated list in a magazine, or want to build a comprehensive artist anthology spanning decades of work, the only option is manually searching and adding tracks one by one.

For a 10-volume David Bowie anthology covering 60 years, that's hundreds of tracks. Doing it manually would take hours of repetitive searching, and any error means starting a playlist over.

> No tooling exists to programmatically create Tidal playlists from external track listings -- BBC programme tracklists, CSV exports, or curated text files require manual track-by-track searching and adding.

- No Tidal playlist import
- Manual search per track
- No batch operations in the app
- Anthology curation is hours of work

---

## The Solution

### Text files in, Tidal playlists out

`tpb` (TidalPlaylistBuilder) is a Rust CLI that reads track listings from simple text files, searches Tidal's API for each track, and creates playlists with matched results. It supports three input formats and includes an anthology system for building multi-volume artist collections.

**Text Files**
One line per track: `Artist - Title`

**CSV Files**
Columns for artist, title, album. Exported from spreadsheets or databases.

**BBC Programmes**
Tracklists from BBC 6 Music, Radio 2, and other BBC programme pages.

```
$ tpb auth
Open this URL to authorise:
https://link.tidal.com/XXXXX
Waiting for authorisation...
Authenticated. Token saved to Keychain.

$ tpb status
Authenticated as Mike Lott (HiFi Plus)
Token expires: 2026-04-12T08:00:00Z
```

---

## Anthologies

### Multi-volume artist collections

The anthology system builds comprehensive artist collections from a directory of text files and a `manifest.json`. Each volume becomes a separate Tidal playlist, organised in a named folder. The manifest defines volume titles, descriptions, and artist aliases for search matching.

| Artist | Volumes | Description |
|---|---|---|
| **David Bowie -- An Anthology** | 10 volumes | From the bewildered boy of 1964 through Ziggy Stardust, the Berlin Trilogy, and the final act of Blackstar. 60 years of reinvention in curated chronological volumes. |
| **Iron Maiden** | Anthology | The complete Iron Maiden catalogue organised by era. |
| **New Model Army** | Anthology | Four decades of post-punk, from Vengeance through to the present. |
| **Ryuichi Sakamoto** | Anthology | YMO, film scores, ambient works, and solo piano across a lifetime of composition. |

```
$ python scripts/build_anthology.py anthologies/david-bowie/ --dry-run
Loading manifest: David Bowie - An Anthology (10 volumes)
Vol. I: The Bewildered Boy (1964-1969) — 24 tracks
Vol. II: The Man Who Sold the World (1970-1971) — 18 tracks
Vol. III: Ziggy Stardust & the Spiders (1971-1973) — 22 tracks
...
Vol. X: The Final Act (2013-2016) — 15 tracks
Dry run complete. 198 tracks across 10 volumes.
```

---

## Design Decisions

### Why it works this way

- **Rust CLI with Keychain token storage** -- Single binary, no runtime dependencies. OAuth tokens are stored in the macOS Keychain via the system's security framework, not in plaintext config files. The device auth flow means no client secret in the binary.
- **Text files as the universal input format** -- Every track listing can be reduced to `Artist - Title` lines in a text file. BBC tracklists, magazine lists, Discogs exports, hand-curated selections -- all become the same simple format. No complex parsers, no brittle scrapers.
- **Anthology manifest system** -- A `manifest.json` defines the folder name, volume titles, descriptions, and artist aliases. This separates curation (the creative work of choosing tracks and writing descriptions) from execution (searching Tidal and creating playlists). The manifest is the human-authored artefact; the tooling handles the mechanical work.
- **Python scripts alongside Rust CLI** -- The core CLI (`tpb`) handles auth, status, and single-playlist creation in Rust. The anthology builder and batch operations use Python scripts that call the Tidal API directly. This reflects the project's evolution -- the Python scripts came first as prototypes, the Rust CLI followed for the permanent tooling.
- **Dry-run by default for anthologies** -- Building a 10-volume anthology creates 10 playlists and hundreds of API calls. `--dry-run` shows what would be created without touching Tidal. This prevents accidental duplicate playlists and lets you verify track matching before committing.

---

## Evolution

### How it was built

**v0.1.0** -- 6 March 2026 -- Foundation -- CLI, Auth, Keychain

Added: Clap CLI skeleton with `auth`, `status` subcommands. Tidal OAuth device flow authentication. macOS Keychain token storage. 4.6 MB release binary (LTO + strip).

Code Review: 6 findings from full code review: `expect()` panic replaced with `Result` propagation, clippy warnings resolved, unused error types removed, redundant builder pattern simplified.

**Scripts** -- Ongoing -- Anthology Builder & Batch Operations

Added: Python anthology builder: reads `manifest.json` + volume text files, searches Tidal, creates folder + playlists. Dry-run support. Artist alias matching for side projects and collaborations.

Added: 4 curated anthologies: David Bowie (10 volumes), Iron Maiden, New Model Army, Ryuichi Sakamoto. Track reordering script. Batch track addition script.

---

## Current Status

**Project Status:** Active -- v0.1.0

| | |
|---|---|
| **Language** | Rust + Python |
| **Anthologies** | 4 artists |
| **License** | MIT |

The Rust CLI handles authentication and will expand to cover playlist creation natively. The Python anthology scripts are functional and have been used to build all 4 artist anthologies on a live Tidal account.

---

[Source on GitHub](https://github.com/t0lkim/tidal-playlist-builder)

(c) 2026 t0lkim
