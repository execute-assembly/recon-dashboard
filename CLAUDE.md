# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a bug bounty / security reconnaissance automation toolkit. It implements a five-stage pipeline: passive subdomain enumeration → active DNS resolution → port scanning + CDN detection → HTTP probing → dashboard reporting.

**All reconnaissance must only be run against domains with explicit written authorization.**

## Pipeline Execution

```bash
# Full automated pipeline
./recon.sh <domain>

# Or run stages individually:

# Stage 1: Passive subdomain enumeration
./recon-files/subdomain2.sh [-a] <domain>
# -a = append mode; outputs subdomains/all_subs.txt

# Stage 2: Active DNS resolution + permutation
./recon-files/subdomains_active.sh [-b] <domain>
# -b = skip bruteforce; outputs subdomains/final_subs.txt

# Stage 3: Port scanning + CDN detection
./recon-files/port_scan.sh
# reads subdomains/final_subs.txt; outputs probe/port-scan/

# Stage 4: HTTP probing + sensitive path discovery
./recon-files/alive_httpx_probe.sh
# reads probe/port-scan/ports.json + domain_ips.json; outputs probe/httpx/

# Stage 5: Dashboard (serves on localhost:8000)
python3 server/app.py

# Optional: Web crawling (supplementary)
./recon-files/crawl.sh <domain_file> [--headless]
```

## Directory Structure

```
recon/
├── recon.sh                        # Main orchestration script
├── recon-files/
│   ├── subdomain2.sh               # Stage 1: Passive enumeration
│   ├── subdomains_active.sh        # Stage 2: Active DNS + permutation
│   ├── port_scan.sh                # Stage 3: Port scan + CDN detection
│   ├── alive_httpx_probe.sh        # Stage 4: HTTP probing + path discovery
│   └── crawl.sh                    # Optional: Web crawling
├── server/
│   ├── app.py                      # Flask REST API (port 8000)
│   └── static/index.html           # Dark-themed dashboard frontend
├── subdomains/
│   ├── all_subs.txt                # Stage 1 output (merged/deduped)
│   ├── final_subs.txt              # Stage 2 output (DNS-verified)
│   └── passive/                    # Per-source subdomain lists
│       ├── subfinder.txt
│       ├── crt_subs.txt
│       └── gitsubs.txt
├── probe/
│   ├── httpx/
│   │   ├── targets.txt             # Built domain:port pairs
│   │   ├── httpx_enriched.json     # JSONL enriched HTTP results
│   │   ├── path_hits_raw.json      # Raw path discovery results
│   │   └── path_hits.txt           # Pipe-delimited path hits
│   └── port-scan/
│       ├── scan_targets.txt        # Non-CDN IPs for masscan
│       ├── domain_ips.json         # Domain-to-IP mappings + CDN flags
│       └── ports.json              # Masscan results (IP → open ports)
└── temp/
    └── cdn_ranges.txt              # Cached CDN CIDR blocks (refreshed every 7 days)
```

## Architecture

**Stage 1 — Passive Recon** (`recon-files/subdomain2.sh`): Queries subfinder (with Shodan API), crt.sh certificate transparency logs, and GitHub API in parallel. Merges and deduplicates into `subdomains/all_subs.txt`.

**Stage 2 — Active Resolution** (`recon-files/subdomains_active.sh`): Resolves discovered subdomains via `puredns` (10k req/s general, 2k req/s trusted). Generates permutations with `alterx` (uses SecLists wordlist if <50 domains found), then re-resolves. Resolver lists downloaded from trickest/resolvers or fall back to SecLists. Outputs verified domains to `subdomains/final_subs.txt`.

**Stage 3 — Port Scanning + CDN Detection** (`recon-files/port_scan.sh`): Resolves IPs via `dnsx`, checks against CDN CIDR ranges (Cloudflare, Fastly, AWS CloudFront, GCP, Akamai, Incapsula, Sucuri — 8,194 blocks). Runs `masscan` (ports 0–10000 at 20k pps, requires sudo) only on non-CDN IPs. Outputs `probe/port-scan/domain_ips.json` (with `cdn` flag) and `probe/port-scan/ports.json`.

**Stage 4 — HTTP Probing** (`recon-files/alive_httpx_probe.sh`): Builds domain:port targets from port scan results (CDN domains default to ports 80, 443, 8080, 8443, 8000, 8888, 3000, 9090, 4443, 8081, 9443, 2087, 2083, 9200). Runs `httpx` with full enrichment (status, title, tech, content-length, web-server, ip, cname). Merges open port data into JSONL output. Probes 20 sensitive paths:
`.git/config`, `.env`, `robots.txt`, `sitemap.xml`, `crossdomain.xml`,
`clientaccesspolicy.xml`, `.well-known/security.txt`,
`api/swagger`, `api/swagger.json`, `api/openapi.json`, `v1/swagger`,
`actuator`, `actuator/env`, `actuator/mappings`,
`phpinfo.php`, `server-status`, `server-info`,
`wp-admin`, `wp-config.php.bak`, `.DS_Store`
Hits (status 200/201/301/302/403) go to `probe/httpx/path_hits.txt` (pipe-delimited: `URL|STATUS|SIZE`).

**Stage 5 — Dashboard** (`server/app.py` + `server/static/index.html`): Flask REST API on `0.0.0.0:8000`. Three tabs: **Hosts** (sortable/filterable table with status, title, tech, open ports, IP, CNAME), **Juicy Hits** (path hits with HIGH/MEDIUM/INFO severity), **JS Analysis** (reads from `probe/js-analysis/<domain>/` — endpoints and secrets). Severity classification: HIGH = `.env`, `.git`, `wp-config`, `phpinfo`, `server-status`, `actuator/env`; MEDIUM = `actuator`, `swagger`, `openapi`, `server-info`.

**Optional — Web Crawling** (`recon-files/crawl.sh`): Active crawling via `katana` (depth 3, JS parsing, 20 concurrent) and passive via `gau`/`waybackurls`. Passive functions currently commented out.

## Dependencies

```
# Recon tools
subfinder, github-subdomains, puredns, alterx, dnsx, httpx
katana, gau, waybackurls

# Scanning
masscan  (requires sudo)

# Utilities
curl, jq, whois, grepcidr, wget, python3

# Python
flask
```

**External data:** SecLists at `/usr/share/seclists/`, resolver lists from trickest/resolvers GitHub repo.

## Data Formats

- `subdomains/`: plain text domain lists (one per line)
- `probe/httpx/httpx_enriched.json`: JSONL — one JSON object per host (`url`, `status_code`, `title`, `tech`, `ip`, `cname`, `cdn`, `open_ports`)
- `probe/httpx/path_hits.txt`: pipe-delimited `URL|HTTP_STATUS|RESPONSE_SIZE`
- `probe/port-scan/domain_ips.json`: JSON array — `{domain, ip, cdn: bool}`
- `probe/port-scan/ports.json`: JSON — `{ip: {domains: [], ports: []}}`
- `probe/js-analysis/<domain>/`: JS analysis results (endpoints + secrets)

## Security Notes

- **CRITICAL:** GitHub token is hardcoded in `recon-files/subdomain2.sh` line ~46. Rotate it and move to an environment variable (`GITHUB_TOKEN`).
- The Flask dashboard binds to `0.0.0.0:8000` (all interfaces) — run only on trusted networks or bind to `127.0.0.1`.
- `masscan` requires sudo — ensure passwordless sudo is configured or run interactively.
- Do not commit live target data, scan results, or API tokens.
- Scripts hardcode SecLists paths at `/usr/share/seclists/` — adjust if installed elsewhere.
