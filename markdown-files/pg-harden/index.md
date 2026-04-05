---
title: "pg-harden — Case Study"
description: "A Rust CLI tool for PostgreSQL security hardening — scan, report, and enforce deny-all security posture."
hero_title: "pg-harden"
hero_sub: "A Rust CLI for PostgreSQL security hardening. Scan, report, and enforce a deny-all security posture across your database fleet."
---

## The Problem

### PostgreSQL ships open by default

PostgreSQL's out-of-the-box configuration prioritises ease of setup over security. Default installs allow local trust authentication, MD5 password hashing, unencrypted connections, and unrestricted public schema access. Most hardening guides are checklists that require manual, repetitive work across every instance.

As infrastructure grows — especially in environments with multiple PostgreSQL instances across LXC containers and VMs — there's no practical way to verify that every database meets a consistent security baseline without a dedicated tool.

> No single-binary tool exists to scan, report on, and enforce a deny-all PostgreSQL security posture across local and remote instances, with rollback capability and CI/CD integration.

- Default config allows trust auth
- MD5 hashing still prevalent
- SSL often disabled
- Manual hardening doesn't scale
- No drift detection between instances

---

## The Solution

### Scan, report, harden, rollback

`pg-harden` is a single Rust binary with **no runtime dependencies** beyond OpenSSH for remote mode. It scans PostgreSQL instances against a comprehensive security checklist, reports findings with severity levels, applies hardening profiles, and tracks every change for rollback.

**1. Scan & Report**

30+ security checks across authentication, SSL/TLS, connections, logging, privileges, extensions, and network exposure. Output as coloured terminal, Markdown, JSON, or HTML.

**2. Harden & Rollback**

Apply hardening profiles (minimal, standard, strict) with dry-run preview. Every change is tracked and reversible. Baseline snapshots enable drift detection over time.

```
# pg-harden scan
$ pg-harden scan -H 192.0.2.10
pg-harden v0.2.1 — PostgreSQL Security Hardening
Scanning PostgreSQL 18.3 at 192.0.2.10:5432...

PASS  auth-scram     SCRAM-SHA-256 authentication enforced
PASS  ssl-enabled    SSL/TLS connection active

Summary: 2 passed, 0 failed
Exit code: 0
```

```
# multi-target scan
$ pg-harden scan -H 192.0.2.0/28 -f json
Expanding CIDR 192.0.2.0/28 → 14 hosts
Scanning sequentially...

── 192.0.2.1 ──
PASS  auth-scram | PASS  ssl-enabled

── 192.0.2.2 ──
FAIL  auth-scram | PASS  ssl-enabled

Overall: 1 host with findings (exit code: 2)
```

---

## Architecture

### Local and remote execution

`pg-harden` runs locally on the database server or remotely via SSH. In remote mode, a stateless agent binary is SCP'd to the target, executed, and returns JSON results. The agent is cached for future runs.

```
Local Machine                          Remote Host (LXC/VM)

┌──────────────┐     SSH + SCP        ┌──────────────────┐
│  pg-harden   │ ───────────────────▶ │ pg-harden-agent  │
│    (CLI)     │                      │   (stateless)    │
└──────────────┘                      └────────┬─────────┘
      ▲                                        │
      │         JSON response         ┌────────▼─────────┐
      │◀──────────────────────────────│   PostgreSQL     │
                                      │   + OS checks    │
                                      └──────────────────┘
```

**Design principles:** Single binary, no runtime dependencies beyond OpenSSH, non-destructive by default (scan without changes), all changes reversible, practical hardening over checkbox compliance.

---

## Security Checks

### Deny-all posture

The philosophy: **locked down by default, explicitly open what's needed** — like nftables drop-all inbound. Currently 3 checks implemented (v0.2.1), with 30+ planned across 7 categories.

| Severity | Check | Description |
|----------|-------|-------------|
| Critical | `auth-scram` | SCRAM-SHA-256 enforced, MD5 disabled |
| Critical | `ssl-enabled` | SSL/TLS connection enforcement |
| Critical | `auth-pghba` | pg_hba.conf audit — no trust, no md5 |
| High | `ssl-require` | hostssl enforced in pg_hba.conf |
| High | `conn-listen` | listen_addresses not wildcard '*' |
| High | `priv-public` | Revoke CREATE on public schema |
| High | `log-pgaudit` | pgaudit extension installed and configured |
| High | `net-allowed` | pg_hba.conf allowed CIDR review |
| Medium | `log-connections` | Connection logging enabled |
| Medium | `conn-timeout` | Idle transaction timeout set |
| Medium | `ext-risky` | Dangerous extensions audit (dblink, plpythonu) |
| Medium | `conf-checksums` | Data checksums enabled |

Optional OS-level checks (with `--os` flag) cover PGDATA permissions, sysctl tuning, systemd hardening, and firewall rules.

---

## Evolution

### How it was built

**v0.1.0** — 10 February 2026 — CLI Skeleton & First Checks

- **Added:** Clap-based CLI with `scan` subcommand. 3 checks: `auth-scram`, `ssl-enabled`, `auth-pghba`. Connection via TCP host (`-H`) or Unix socket (`-s`). Coloured text and JSON output formats.
- **Architecture:** Check filtering with include (`-c`) and exclude (`-x`). Offline mode (`--offline`) for file-based checks. Environment variable support for all PG connection params.

**v0.2.0** — 11 February 2026 — Multi-Target & Network Scanning

- **Added:** CIDR target support — `-H 192.168.1.0/24` expands to individual hosts via `ipnet` crate. Network and broadcast addresses excluded automatically. IPv4 and IPv6 supported.
- **Added:** Hostname DNS resolution — `-H db.example.com` resolves via `ToSocketAddrs`. Dual-stack hosts scan all IPs. Deduplication prevents scanning the same IP twice.
- **Added:** Repeatable `-H` flag — combine CIDR blocks, hostnames, and bare IPs freely. Per-host report grouping with individual summaries and aggregate overall line.
- **Changed:** JSON output restructured to `hosts[]` array for multi-target. Help text expanded to 9 usage examples. 428 insertions across 9 files.

**v0.2.1** — 11 February 2026 — Polish & Verification

- **Fixed:** Help text alignment — example descriptions aligned to a consistent column. Tested against PostgreSQL 18.3 on hardened Debian 13 LXC: 2/2 checks pass (SCRAM-SHA-256, SSL).
- **Spec:** Deny-all hardening specification authored — comprehensive lockdown checklist covering connection controls, authentication, privilege minimisation, audit logging, network security, and BCP.

---

## Design Decisions

### Why it works this way

- **Rust for a single static binary** — No Python, no Node, no runtime dependencies. The binary runs on any Linux system without installing anything. This is critical for scanning database servers where you don't want to install tooling.
- **SSH-based remote execution over agent frameworks** — Instead of requiring an agent daemon on every host, `pg-harden` SCP's a stateless agent binary on demand and executes it via SSH. The agent is cached for performance but can be removed with `--no-cache`. This piggybacks on existing SSH infrastructure — no new network services, no new attack surface.
- **Deny-all philosophy over compliance checkboxes** — Most hardening tools map to CIS benchmarks and report pass/fail against a checklist. `pg-harden` starts from the opposite direction: everything is denied by default, and the tool verifies that only explicitly permitted access exists. This catches misconfigurations that checkbox compliance misses.
- **Non-destructive by default** — `pg-harden scan` never modifies anything. `pg-harden harden` requires explicit invocation and supports `--dry-run`. Every change is logged with a rollback manifest. This makes it safe to run in production without risk of accidental lockout.
- **Exit codes for CI/CD** — Exit 0 = all pass, 1 = warnings, 2 = critical findings, 3 = connection error. This makes `pg-harden` a native CI/CD gate — add it to your pipeline and fail the build if security posture degrades.
- **CIDR expansion for fleet scanning** — Instead of listing every host, `-H 192.0.2.0/24` scans the entire subnet. Combined with per-host JSON output, this enables fleet-wide security posture reporting from a single command.

---

## Roadmap

### What's next

1. **CLI + Core Checks** [Complete] — Clap skeleton, PostgreSQL connection, 3 checks (auth-scram, ssl-enabled, auth-pghba), text and JSON output, multi-target CIDR scanning, hostname resolution.
2. **Full Check Suite** [Next] — Implement remaining 27+ checks from the deny-all spec: connection controls, privilege minimisation, pgaudit integration, extension auditing, network exposure review. Severity-based filtering and thresholds.
3. **Reporting** [Planned] — Markdown, HTML, and JSON report generation. CIS PostgreSQL Benchmark mapping. Executive summary with prioritised remediation steps.
4. **Hardening Engine** [Planned] — Apply hardening profiles (minimal, standard, strict). Dry-run preview. Configuration change tracking with rollback manifests. Feature toggles for granular control.
5. **Remote Mode** [Planned] — SSH agent execution model. SCP binary to target, execute, return JSON. Agent caching. Jump host support (`-J bastion@gateway`).
6. **OS-Level Checks** [Planned] — PGDATA permissions, config file permissions, sysctl tuning, systemd hardening, firewall rule audit. Enabled via `--os` flag.
7. **Baseline & Drift Detection** [Planned] — Save security state as baseline snapshots. Compare current state against baseline to detect configuration drift. Named baselines for pre/post-upgrade comparison.

---

## Current Status

**Project Status:** Active — v0.2.1

| | |
|---|---|
| **Language** | Rust (edition 2024) |
| **Checks** | 3 / 30+ planned |
| **License** | MIT |

`pg-harden` v0.2.1 is functional for SCRAM and SSL verification across single hosts and CIDR subnets. The deny-all hardening specification is complete and drives the Phase 2 check suite expansion. Tested against PostgreSQL 18.3 on custom hardened Debian 13 LXC containers.

---

[Source on GitHub](https://github.com/t0lkim/pg-harden)

(c) 2026 t0lkim
