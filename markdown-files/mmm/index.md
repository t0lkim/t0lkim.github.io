---
title: "MMM (MyMediaManager) — Case Study"
description: "A Rust CLI for organising images and videos with deduplication, EXIF-based renaming, and date-based directory structure."
hero_title: "<span class=\"accent\">MMM</span>"
hero_sub: "A Rust CLI for organising images and videos with deduplication, EXIF-based renaming, and date-based directory structure."
---

## The Problem

### Years of media, zero organisation

Photos and videos accumulate across phones, cameras, SD cards, cloud downloads, and desktop folders. The same image exists in three places under three different names. Files are named `IMG_0042.jpg` or `DSC_1893.CR2` with no indication of when or where they were taken. Manual sorting is tedious for hundreds of files and completely impractical at thousands.

The core issues compound over time:

- Duplicate files waste storage and create confusion about which copy is the "real" one
- Camera-assigned filenames carry no meaningful information
- No consistent directory structure makes finding a specific photo a needle-in-a-haystack exercise
- File modification dates are unreliable — copying, syncing, and cloud downloads all reset them
- Photos with GPS data carry useful location context that never makes it into the filename or folder structure

> No single tool scans multiple directories, deduplicates by content, renames from EXIF metadata, sorts into a date hierarchy, and reverse-geocodes GPS coordinates — all offline, with a dry-run mode and chunked processing for safety.

---

## The Solution

### Scan, deduplicate, rename, organise

`media-organiser` is a Rust CLI that takes one or more input directories, finds every image and video, eliminates duplicates by content hash, extracts the original capture date from EXIF metadata, and sorts files into a `YYYY/MM/DD` directory structure with date-and-location-based filenames.

A companion binary, `dedup-verifier`, independently verifies that flagged duplicates are genuine before you delete them — using a deliberately different hashing approach for safety.

```
# Preview what would happen (nothing is modified)
media-organiser ~/Photos --dry-run
```

```
# Organise files from multiple sources into a single directory
media-organiser ~/Photos ~/Camera/DCIM -o ~/Organised
```

```
# Verify duplicates independently, then clean up
dedup-verifier ~/Organised/duplicates/
```

**Output structure:**

```
~/Organised/
├── 2024/
│   ├── 01/
│   │   └── 15/
│   │       ├── 2024-01-15-143022-London-GB.jpg
│   │       └── 2024-01-15-143025-London-GB.jpg
│   └── 03/
│       └── 20/
│           └── 2024-03-20-091500.mp4
├── unsorted/
│   └── unknown.bmp
└── duplicates/
    ├── 000/
    │   ├── manifest.txt
    │   └── IMG_0042.jpg
    └── 001/
        ├── manifest.txt
        └── clip.mov
```

---

## How It Works

### Two-phase pipeline

The system uses a two-pass architecture. Phase A is entirely read-only — it scans, hashes, deduplicates, extracts metadata, and plans every file operation. In dry-run mode, execution stops here. Phase B executes the plan: moving duplicates, then renaming and sorting unique files in chunks with user confirmation between batches.

**Phase A: Scan**

1. **Scan** — Recursively walk all input directories, filtering for 33 supported image and video extensions (JPEG, PNG, HEIC, RAW variants, MOV, MP4, MKV, and more).
2. **Hash** — Three-phase dedup cascade: group by file size (free — metadata only), then partial BLAKE3 hash (first 64KB + last 64KB), then full BLAKE3 hash only for files that survived both filters. Typically less than 1% of files reach the full-hash phase.
3. **Deduplicate** — Files with matching full hashes are grouped. The first file in each group is kept as the original; the rest are flagged as duplicates.
4. **Extract metadata** — Read EXIF data (images) or container atoms (video) for creation date and GPS coordinates. Fall back to filesystem creation date when metadata is absent. Files with no date go to `unsorted/`.
5. **Reverse geocode** — When GPS coordinates are present, look up the nearest city and country using an offline GeoNames dataset (bundled k-d tree, no network requests). The location is sanitised and appended to the filename.
6. **Plan** — Compute the target path for every unique file: `YYYY/MM/DD/YYYY-MM-DD-HHMMSS[-location].ext`. Resolve filename collisions with numeric suffixes.

**Phase B: Process**

7. **Move duplicates** — Each duplicate group is moved to a numbered subdirectory under `duplicates/` with a `manifest.txt` recording the BLAKE3 hash and original file path.
8. **Organise** — Unique files are renamed and moved into the date hierarchy in configurable chunks (default 100), pausing for confirmation between batches.
9. **Report** — Print a summary of files scanned, organised, duplicated, and any errors.

---

## Design Decisions

### Why it works this way

- **Rust for performance and safety** — Media libraries can contain tens of thousands of large files. Rust's zero-cost abstractions, efficient I/O, and memory safety make it practical to hash and move files at scale without garbage collection pauses or runtime overhead. The release binary is LTO-optimised and stripped.

- **Content hashing over filename comparison** — Two files named `IMG_0042.jpg` might be different photos from different cameras, and two files with completely different names might be byte-identical copies. BLAKE3 content hashing is the only reliable way to determine whether two files are actually the same. The three-phase cascade (size, partial hash, full hash) minimises I/O by proving uniqueness as cheaply as possible.

- **EXIF date over file modification date** — File modification timestamps are fragile. Copying to a USB drive, syncing via cloud, or extracting from a ZIP archive all reset the modification date. EXIF DateTimeOriginal records the moment the shutter fired, which is the date that actually matters for organisation. Filesystem dates are only used as a fallback when EXIF is absent.

- **Chunked processing with confirmation** — Moving thousands of files in one uninterruptible operation is risky. Chunked processing (default: 100 files per batch) with user confirmation between batches provides a natural checkpoint. If something looks wrong, you can stop immediately — files already moved stay moved, nothing else is touched.

- **Dry-run by default philosophy** — The tool is designed so the recommended first step is always `--dry-run`. The dry-run report shows every planned operation with metadata source tags (`[EXIF]`, `[FS]`, `[NO DATE]`) so you can review before committing. No files are created, moved, or deleted during a dry run.

- **Independent verification binary** — The `dedup-verifier` uses BLAKE3 keyed mode (a different algorithm from the main binary's unkeyed mode) and always hashes the full file (no partial-hash shortcut). A bug in one binary cannot produce a false positive in both — the same independent-channel principle used in safety-critical systems.

- **Offline reverse geocoding** — GPS-to-location lookups use the `reverse_geocoder` crate with a bundled GeoNames dataset loaded into a k-d tree. No network requests, no API keys, no rate limits. Location data is sanitised for filename safety before being appended.

---

## Current Status

**Project Status:** Active — v0.1.0

| | |
|---|---|
| **Language** | Rust (edition 2021) |
| **Supported formats** | 22 image + 11 video extensions |
| **License** | MIT |

`media-organiser` v0.1.0 is fully functional with recursive multi-directory scanning, three-phase BLAKE3 deduplication, EXIF and video metadata extraction, offline reverse geocoding, date-based directory organisation, dry-run mode, chunked processing, and independent duplicate verification.

---

[Source on GitHub](https://github.com/t0lkim/media-organiser)

(c) 2026 t0lkim
