---
title: "TheStreamDeck — Case Study"
description: "A Raspberry Pi-based unified streaming audio appliance with touchscreen, physical VU meters, and instant source switching."
hero_title: "TheStreamDeck"
hero_sub: "A Raspberry Pi-based unified streaming audio appliance. One device, one amp input, instant source switching, album art on screen, and physical analogue VU meters."
---

## The Problem

### Too many sources, too much friction

An audiophile with multiple streaming subscriptions — Tidal HiFi Plus, Plexamp (local library), BBC 6 Music, internet radio — experiences frustrating, slow source switching when listening through a high-quality integrated amplifier. Each service runs on a different device, requiring manual amp input selection.

The fragmented workflow disrupts listening sessions, adds unnecessary physical interaction, and wastes the visual potential of album artwork during playback.

> Switching between audio sources requires changing physical devices and amp inputs — slow, disruptive, and frustrating. No unified device exists that consolidates all streaming sources into a single digital output with instant switching.

- Each source = different device
- Manual amp input switching
- No album art during playback
- Broken listening flow
- Underuse of paid subscriptions

---

## The Solution

### One device, one input, instant switching

TheStreamDeck consolidates BBC 6 Music, Tidal HiFi Plus, Plexamp, and internet radio into a single Raspberry Pi 4 appliance with a touchscreen, physical source buttons, real-time VU meter visualisation, and S/PDIF digital output to an integrated amplifier.

**4 Sources, 1 Output**

BBC 6 Music (MPD), Tidal HiFi Plus (Tidal Connect), Plexamp (headless), and internet radio — all routed through PipeWire to a single S/PDIF optical output.

**< 3 Second Switching**

Physical buttons and touchscreen for instant source selection. A Rust daemon manages source lifecycle — stop current, start new, update UI. No amp input changes needed.

---

## Audio Sources

### Four sources, unified

| Source | Quality | Details |
|---|---|---|
| BBC 6 Music | 320kbps AAC | MPD player, HLS stream. ICY metadata for now-playing info. |
| Tidal HiFi Plus | Lossless FLAC 24/96 | Tidal Connect via Podman. Cast from phone, Pi renders audio. |
| Plexamp | FLAC (local library) | Headless Node.js player. Cast-based control from any client. |
| Internet Radio | Variable | MPD with preset stations. Shoutcast, Icecast, HLS streams. |

---

## Hardware Architecture

### Purpose-built audio appliance

| Component | Name | Details |
|---|---|---|
| Compute | Raspberry Pi 4 Model B | Quad A72 @ 1.5GHz, VideoCore VI GPU. PiOS Lite 64-bit. Wired Ethernet for reliability. |
| Digital Transport | HiFiBerry Digi2 Pro | S/PDIF output (TOSLINK optical + coax). Dual oscillator, low jitter. 192kHz/24-bit. NOT a DAC — pure digital transport. |
| Display | 5" Touchscreen | IPS capacitive touch. HDMI + USB (no GPIO conflict). Fits 70mm case height to match amplifier form factor. |
| VU Meter DAC | MCP4922 | Dual 12-bit SPI DAC driving L/R analogue panel meters. 4096 steps, zero jitter. Powered from Pi's 3.3V rail. |
| VU Meters | Analogue Panel Meters | Moving-coil needle meters. Software VU ballistics (300ms IIR filter). Series resistor + trim pot per channel for calibration. |
| Amplifier | NAD C 3xx Series | Optical digital input, Cirrus Logic internal DAC. All D/A conversion happens in the amp — the Pi stays purely digital. |

### Signal Path

```
Streaming Source (Tidal / Plexamp / MPD)
    |
    v
PipeWire (audio server / multiplexer)
    |
    +---> HiFiBerry Digi2 Pro (I2S -> S/PDIF)
    |         |
    |         +---> TOSLINK optical cable
    |                   |
    |                   v
    |              Amplifier optical input
    |                   |
    |                   v
    |              Internal DAC (D/A conversion)
    |                   |
    |                   v
    |              Amplifier stage -> Speakers
    |
    +---> PipeWire monitor source (audio tap)
              |
              v
         VU Meter / Spectrum (cpal + RustFFT -> display + MCP4922)
```

**Why Digi HAT, not DAC HAT:** The Pi stays purely digital — no analogue stages on the Pi side. The amplifier's Cirrus Logic DAC is purpose-built with dedicated power regulation and shielding. Using a DAC HAT would mean unnecessary double conversion (digital -> HAT DAC -> analogue -> amp). TOSLINK optical provides galvanic isolation, eliminating ground loop hum from the Pi's switching PSU.

---

## User Interface

### Touchscreen + physical controls

The UI renders via Rust + Slint (GPU-accelerated, OpenGL ES on VideoCore VI). No X11 or Wayland — direct framebuffer rendering. The display shows album artwork, VU meters, now-playing metadata, source buttons, and transport controls.

```
+-----------------------------------------------------------------------+
|                                                                       |
|  +---------------------------+  +----------------------------------+  |
|  |                           |  |                                  |  |
|  |      Album Artwork        |  |      VU Meter / Spectrum         |  |
|  |       (600x600)           |  |       Visualisation              |  |
|  |                           |  |        (L + R)                   |  |
|  +---------------------------+  +----------------------------------+  |
|                                                                       |
|  +-----------------------------------------------------------------+  |
|  |  Track: Song Title                              Artist Name     |  |
|  |  Album: Album Title                             FLAC 24/96      |  |
|  +-----------------------------------------------------------------+  |
|                                                                       |
|     [ BBC ] [ TDAL ] [ PLEX ] [ RDIO ]      [<<] [  > / ||  ] [>>]    |
|                                                                       |
+-----------------------------------------------------------------------+
```

Physical buttons mirror the touchscreen sources (GPIO-connected). A rotary encoder provides volume control. All controls work with the display off.

---

## Design Decisions

### Why it works this way

- **Digital transport over DAC HAT** — The amplifier's internal Cirrus Logic DAC is objectively superior to any ~$30 Pi HAT DAC. Keeping the Pi purely digital avoids double D/A conversion and eliminates ground loop hum via TOSLINK galvanic isolation. The HiFiBerry Digi2 Pro provides bit-perfect S/PDIF with dual oscillators for low jitter.

- **MCP4922 SPI DAC for VU meters over PWM** — Hardware PWM on GPIO 18 conflicts with the HiFiBerry's PCM_CLK. Software PWM produces visible needle jitter (~13 usable steps). The MCP4922 gives 4096 steps with zero jitter via SPI — proven superior in head-to-head testing. Total cost: ~$4 for the IC.

- **PipeWire as the audio multiplexer** — PipeWire is the default audio server on PiOS Bookworm. It natively multiplexes all sources (MPD, Tidal Connect, Plexamp) into a single output. Monitor sources provide an audio tap for VU meter visualisation without affecting the main signal path.

- **Rust + Slint for the UI** — Purpose-built for embedded touchscreens. GPU-accelerated via OpenGL ES on VideoCore VI. Works with framebuffer directly — no X11/Wayland needed. Keeps the entire codebase in Rust. VU meter renders at 60fps with ~8-15% CPU overhead.

- **Cast-based control model for Tidal and Plexamp** — The Pi is the renderer, not the controller. Tidal Connect and Plexamp both support cast-based playback — control from your phone, audio plays through the Pi. This avoids building complex playback UIs and leverages the native apps' full feature sets.

- **435mm width form factor** — Designed to match the NAD C 3xx amplifier width. The case height is locked to 70mm (matching the NAD C 338 profile). This constrains the display to 5" max and the VU meters to 45-50mm — deliberate trade-offs for visual coherence in the listening position.

- **Software VU ballistics over mechanical damping** — A 300ms IIR filter in software implements VU-standard ballistics (1-1.5% overshoot). Budget analogue meters have poor mechanical damping, so software compensation is essential. Higher-quality meters can rely on their own inertia with only a gentle 50ms smoothing pass.

---

## Evolution

### How it took shape

**v0.1.0** — 4 April 2026 — Problem Statement & Technical Specification

- **Architecture:** Hardware architecture designed: Pi 4B + HiFiBerry Digi2 Pro + NAD optical input. First-principles analysis confirmed Digi HAT over DAC HAT. Software stack defined: PiOS Lite, PipeWire, MPD, Tidal Connect (Podman), Plexamp headless.
- **Added:** Source switching architecture with Rust daemon + Slint UI. VU visualisation pipeline: cpal + RustFFT (NEON-accelerated on aarch64). Architecture diagram Rev A. 758-line research file.

**v0.2.0** — 5 April 2026 — VU Meter Subsystem & Display Upgrade

- **Added:** MCP4922 dual SPI DAC circuit designed. KiCad schematic authored. VU meter driver wiring defined with series resistor + trim pot calibration. Printable screen templates for UI sketching.
- **Changed:** Display upgraded from 7" DSI (720p) to 13.3" HDMI (1080p). UI layout updated for full HD canvas. GPIO allocation revised for SPI DAC. BOM updated.

**v0.3.0** — 5 April 2026 — Form Factor & Industrial Design

- **Changed:** Display revised from 13.3" to 5" to fit 70mm case height matching amplifier form factor. VU meters resized from 85mm to 45-50mm. Case width locked to 435mm (standard hi-fi component width).
- **Added:** Front panel concept diagram. Blank wireframe PDFs for sketching. Case wireframe PDFs with four VU meter layout options (circular + rectangular).

---

## Build Plan

### Five phases to first sound

1. **Audio Foundation** (Week 1) — PiOS setup, PipeWire config, HiFiBerry Digi2 Pro I2S output. MPD playing BBC 6 Music through TOSLINK. First sound through the amplifier.
2. **All Sources Online** (Week 2) — Tidal Connect container, Plexamp headless, internet radio presets. Source Controller daemon with GPIO button switching. <3s switching verified.
3. **Display & UI** (Week 3) — Touchscreen integration. Slint UI with album art, now-playing, source buttons, transport controls. Metadata retrieval from all sources.
4. **VU Meters** (Week 4) — MCP4922 wiring and driver. cpal audio capture from PipeWire monitor. RustFFT analysis. Software ballistics. Physical meter calibration.
5. **Case & Polish** (Week 5) — 435mm x 70mm enclosure. Component mounting. Cable management. Boot optimisation. Final integration testing.

---

## Current Status

**Project Status:** Design Complete — v0.3.1

| | |
|---|---|
| Platform | Raspberry Pi 4 |
| Language | Rust + Slint |
| Phase | Pre-build |

Hardware architecture, VU meter circuit design (KiCad schematic), form factor, and software stack are fully specified. The project is ready to begin Phase 1 (audio foundation) once the hardware is assembled. Key components: Pi 4B (owned), HiFiBerry Digi2 Pro (to order), 5" touchscreen (to order), MCP4922 + VU meters (to order).

---

(c) 2026 t0lkim
