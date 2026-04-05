---
title: TeenTidal — Case Study
description: How a parenting gap in Tidal's content filtering led to building a native iOS app with two-layer explicit content protection.
hero_title: Teen<span class="accent">Tidal</span>
hero_sub: How a gap in Tidal's parental controls led to building a native iOS app with two-layer explicit content filtering.
---

## The Problem

### A gap no setting could fix

A parent with a family Tidal subscription wants to give their child a safe music streaming experience. Tidal has an explicit content toggle — but any user can re-enable it from within the app. iOS Screen Time's "Clean" restriction only applies to Apple Music, not third-party apps.

The result: either the child has unrestricted access to explicit content, or the parent resorts to fragile workarounds like Guided Access and Screen Time layering that an older child could circumvent.

> **Problem Statement:** No structural mechanism exists to permanently filter explicit content in Tidal when a child uses their own iOS device unsupervised. This is a platform gap — not a configuration issue.

**Constraints:**
- Tidal's toggle is user-bypassable
- Screen Time ignores third-party apps
- No server-side parental control API
- Guided Access is brittle at best

---

## The Solution

### Two layers, zero settings

TeenTidal is a native iOS/iPadOS app that wraps Tidal's official SDK with **structural content filtering** — not a toggle the child can find and disable, but an architectural guarantee enforced at two independent layers.

**Layer 1 — FilteredCatalog Gate:** Every API response passes through a single gateway service that strips any item where `explicit == true` before it reaches the UI. The child never sees explicit content in search, browse, or discovery.

**Layer 2 — PlayerService Guard:** A belt-and-braces check at the playback layer. Before any track or video ID reaches the Tidal Player SDK, the guard verifies it against the explicit flag. If Layer 1 somehow misses something, Layer 2 catches it.

There is **no settings screen**. No toggle to disable. No deep link that bypasses the filter. The filtering is hardcoded into the architecture — it cannot be turned off without rebuilding the app.

---

## Screenshots

Real screenshots from the v0.3.0-alpha running on iPhone 17 Pro simulator.

- Home — `assets/home.png`
- Search — `assets/search.png`
- Album Detail — `assets/album-detail.png`
- Swipe Actions — `assets/swipe-action.png`

---

## Walkthrough

A quick walkthrough of the discovery flow — browsing, searching, and navigating between screens.

Video: `assets/walkthrough.mp4` (v0.3.0-alpha, iPhone 17 Pro)

---

## Evolution

### v0.1.0 — 2 April 2026 — Architecture & Foundation

- **Architecture:** Two-layer filtering architecture designed and implemented. `FilteredCatalog` wraps all API responses, stripping items where `explicit == true`. `PlayerService` guard adds a belt-and-braces check before any track reaches the Tidal Player SDK.
- **Added:** OAuth 2.1 PKCE authentication via `ASWebAuthenticationSession`. Tidal SDK integration for Auth, Player, and EventProducer. Token storage in iOS Keychain. Three-tab layout: Home, Search, Library.
- **Added:** Basic views: SearchView with debounced queries (500ms), AlbumDetailView with filtered track lists, ArtistDetailView with top tracks, MiniPlayerView and FullPlayerView with lock screen controls.
- **Testing:** 15 unit tests: 10 for FilteredCatalog (explicit tracks/albums/videos filtered, edge cases) and 5 for PlayerService (playback guard, queue re-filtering). Strict Swift 6.0 concurrency compliance.

### v0.2.0-alpha — 3 April 2026 — Full-Featured Discovery & QA Pass

- **Added:** Home tab with 4 curated discovery categories (Pop Hits, Acoustic, Soundtrack, Chill) rendered as horizontal album carousels with resolved cover art. Full search across tracks, albums, and artists with inline playback and swipe actions.
- **Added:** Playlist CRUD — create playlists, add tracks via sheet, delete with confirmation. Local persistence via `LocalPlaylistStore` using UserDefaults. Full track objects stored, surviving app restarts.
- **Added:** Design system established: `TeenTidalColors` (dark theme, teal #00C9A7 accent), `TeenTidalTypography` (rounded fonts), `TeenTidalSpacing` tokens. iPad `NavigationSplitView` sidebar layout. Favourites tab (4th tab) with category counts.
- **QA Pass:** Comprehensive audit: zero `.contextMenu` usage (all swipe actions), zero nested NavigationStacks, all navigation destinations registered via value-based enums. Environment propagation verified at app root. No force unwraps in app code. Strict Swift 6.0 Sendable compliance.

### v0.3.0-alpha — 4 April 2026 — Artist Profiles, Favourites & Downloads

- **Added:** Artist profiles with split content carousels (Albums, EPs, Singles filtered by `albumType`). Similar Artists horizontal carousel with circular images. Artist Radio linking to radio playlist. Artist favourite toggle on header and swipe rows.
- **Added:** Dedicated Favourites tab (moved from Library) with directory structure: Tracks, Artists, Albums, EPs, Singles — each folder only shown when non-empty. Downloads tab with offline queue, status badges (pending/complete/failed), and swipe-to-delete.
- **Changed:** Tab bar expanded from 4 to 5 tabs. Library stripped to playlists only. All context menus replaced with consistent swipe left/right actions. API compliance hardened — PKCE-only public client, client secret removed from binary.
- **Added:** Two-tier persistent disk cache (L1 memory + L2 disk) with tiered TTLs: discover 30min, detail 10min, search 5min. Survives app restarts, clears on logout. Navigation reworked to value-based `FavouritesDestination` enum eliminating closure-based loops.

---

## Design Decisions

- **No settings screen:** The filter is hardcoded into the architecture. There is no toggle, no preferences pane, no deep link, and no URL scheme that could expose unfiltered content. A child cannot disable the filter without rebuilding the app from source.
- **Custom API client instead of SDK decoders:** Tidal's official SDK has non-optional `key` and `keyScale` enum fields in its `TracksAttributes` model, but the API returns null for these on virtually every track. This causes `JSONDecoder` to fail on every response containing track data. Rather than fork the SDK, a custom `TidalAPIClient` was built using raw `URLSession` with lenient Codable models where null-prone fields are optional.
- **Swipe actions over context menus:** Context menus were initially used for track/album actions. They were replaced entirely with swipe left/right actions in v0.3.0. Swipe actions are more discoverable for younger users, feel more native on iOS, and are consistent with the app's parental constraint of not exposing hidden functionality.
- **Value-based navigation over closure-based:** Early versions used closure-based `NavigationLink` which caused navigation loops in the artist profile. Reworked to value-based navigation via a `FavouritesDestination` enum, with each tab owning exactly one `NavigationStack`.
- **Dark theme with child-friendly typography:** Dark background (#0D1117) with teal accent (#00C9A7) and rounded system fonts. Dark theme reduces visual fatigue. Rounded fonts feel approachable without looking infantile.
- **Local-first persistence:** Favourites, playlists, and download queues are stored in `UserDefaults` on the device. No backend, no cloud sync, no analytics. The privacy policy is one page long because there is nothing to disclose.
- **iPad NavigationSplitView layout:** Instead of scaling the iPhone tab bar to iPad, the app uses `NavigationSplitView` with a sidebar listing all tabs and a detail pane.
- **Hidden admin logout:** Logout is triggered by a hidden 5-tap gesture on the Home screen "Listen" header. There is no visible logout button, no account screen, and no settings path.

---

## API Etiquette & Hygiene

### Single-request resource inclusion

**Investigation:** Early versions made 20+ individual track fetches per search query — one per result row to resolve artist names and album artwork.

**Decision:** Tidal's JSON:API `include` parameter lets you request related resources in a single response. Search now uses `include=tracks,tracks.albums,tracks.albums.coverArt,tracks.artists,albums,artists,albums.coverArt,albums.artists,artists.profileArt` to get everything in one call.

**Result:** Search went from 20+ API calls to exactly 1.

### Tiered caching with staleness awareness

**Investigation:** Browsing the app repeatedly hits the same endpoints. Without caching, every screen transition is an API call.

**Decision:** A two-tier cache: L1 (in-memory) for instant access, L2 (disk cache) to survive app restarts. TTLs are tiered: discovery 30min, detail 10min, search 5min.

**Result:** Repeat navigation is instant. API calls only fire when cache entries expire.

### Sequential fetching for rate limit compliance

**Investigation:** Concurrent batch fetches via `TaskGroup` triggered HTTP 429 responses.

**Decision:** All batch fallback fetches converted to sequential execution. Video fetches removed from search (not displayed in UI).

**Result:** Zero 429 errors in testing.

### Debounced search

**Investigation:** Every keystroke in the search field would trigger an API call.

**Decision:** Search input debounced with 500ms delay via `Task.sleep`. Previous search task cancelled when new input arrives.

**Result:** A typical search triggers 1 API call instead of 10+.

### OAuth compliance — PKCE-only public client

**Investigation:** The initial implementation included the client secret in the app binary.

**Decision:** Client secret removed entirely. TeenTidal operates as a PKCE-only public client. Tokens stored in iOS Keychain. Correct `Accept: application/vnd.api+json` header set on all v2 API requests.

**Result:** No secrets in the binary.

### Server-side explicit filtering

**Investigation:** The client-side `FilteredCatalog` strips explicit content after it arrives, but explicit data still crosses the network.

**Decision:** Search requests include `explicitFilter=EXCLUDE` as a query parameter. The client-side filter remains as a second layer for endpoints where `explicitFilter` isn't available.

**Result:** Two independent filtering layers.

### Artwork resolution strategy

**Investigation:** Tidal's JSON:API uses a multi-hop relationship chain for artwork. Early versions showed placeholder images.

**Decision:** Three lookup tables built from each response's `included` array: artist ID → name, artwork ID → 750x750 image URL, album ID → image URL (resolved via coverArt relationship).

**Result:** Album art, artist profile images, and track thumbnails all resolve correctly from included resources. No additional API calls needed.

---

## Tidal Developer Platform Compliance

TeenTidal is built to comply with Tidal's [Developer Guidelines (v3.0)](https://developer.tidal.com/documentation/guidelines/guidelines-developer-guidelines) and [Developer Terms of Service (v3.0)](https://developer.tidal.com/documentation/guidelines/guidelines-developer-terms). This section documents how each requirement is met.

### Playback

Tidal requires that playback is **only available through an official, unmodified version of the TIDAL Player SDK module**. TeenTidal uses the official `tidal-sdk-ios` (v0.11.9) Player module for all playback operations. No custom playback implementation exists — the SDK handles all audio streaming, DRM, and content delivery.

### Authentication

Tidal mandates OAuth 2.1 with PKCE for user authentication. TeenTidal implements the Authorization Code Flow with PKCE via `ASWebAuthenticationSession`. The client secret was removed from the app binary in v0.3.0 — the app operates as a PKCE-only public client. Tokens are stored in the iOS Keychain via the SDK's credential provider, and refresh is handled automatically.

### Content Access

- **No content modification:** TeenTidal does not edit, remix, alter, or re-encode any Tidal content. The explicit content filter operates on metadata (`explicit == true` flag) — it controls which content is *presented* to the user, not the content itself.
- **No content storage:** No Tidal audio or video content is stored locally. The download queue in v0.3.0 tracks *intent* only — actual offline download requires the Tidal Offliner SDK and production credentials.
- **Temporary caching only:** API response metadata (album titles, artist names, artwork URLs) is cached with tiered TTLs (5–30 minutes) and cleared on logout. No indefinite storage of Tidal content.
- **No stream ripping:** TeenTidal cannot capture, record, or export audio streams. Playback is entirely within the Tidal Player SDK's controlled pipeline.

### API Usage

- **JSON:API v2 compliance:** All API requests use the correct `Accept: application/vnd.api+json` header and the `include` parameter for compound documents, per the JSON:API specification.
- **Rate limiting:** Batch requests are sequential (not concurrent) to avoid triggering HTTP 429 responses. Search is debounced at 500ms. No scraping, crawling, or automated mass data retrieval.
- **Server-side filtering:** Search requests use `explicitFilter=EXCLUDE` where supported by the API, reducing unnecessary data transfer.

### Prohibited Uses — Compliance

| Tidal Restriction | TeenTidal Compliance |
|---|---|
| No AI/ML use of content | No AI or machine learning features. Content metadata is used only for display and filtering. |
| No visual synchronisation | No video sync, slideshow, or visual media output from audio content. |
| No content mixing/remixing | Audio content is played unmodified through the official Player SDK. |
| No data scraping | API used only for catalogue browsing and search. No bulk data extraction. |
| No competing service | TeenTidal is a parental filter layer over Tidal, not a competing music service. Users must have their own Tidal subscription. |
| No commercial use | TeenTidal is a personal/family tool, not a commercial product. Source code is proprietary and not distributed. |
| Branding compliance | Tidal branding and attribution will be included per Design Guidelines before any public release. |

### Production Mode

TeenTidal is currently in development mode. Moving to production requires submitting the application for formal review by Tidal, including source code review for compliance. The app is parked pending this process. The development mode limitation (`PENotAllowed` for streaming) is the expected behaviour for unreviewed applications.

---

## Current Status

**Status:** Parked — Alpha Complete  
**Platform:** iOS / iPadOS 17+  
**Language:** Swift 6.0  
**Distribution:** Proprietary

TeenTidal is currently parked pending access to the Tidal developer programme for production credentials. The alpha is feature-complete for discovery, library management, and content filtering. Playback requires production mode — development credentials return `PENotAllowed` for streaming operations.
