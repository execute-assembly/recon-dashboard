# Recon Dashboard — Project Overview

## What Is It

A bug bounty / penetration testing reconnaissance automation toolkit with a Go-powered web dashboard. The pipeline discovers subdomains, resolves them, port scans, HTTP probes for live services and sensitive paths, then stores all results in a per-target SQLite database served through a dark-themed interactive dashboard.

## Why Use It

- Automates the full passive → active recon pipeline in one command
- Per-target isolated databases — run recon on multiple targets without data mixing
- Dashboard lets you triage hosts, take notes, and filter findings without leaving the browser
- Import is on-demand — run recon, come back later, hit Import, data is in the DB

---

## Repository Layout

```
recon/
├── recon.sh                          # Orchestrates the full pipeline
├── recon-files/
│   ├── subdomain2.sh                 # Stage 1: passive subdomain enumeration
│   ├── subdomains_active.sh          # Stage 2: DNS resolution + permutation
│   ├── port_scan.sh                  # Stage 3: CDN detection + masscan
│   ├── alive_httpx_probe.sh          # Stage 4: HTTP probing + path discovery
│   └── crawl.sh                      # Optional: web crawling (katana/gau/wayback)
├── subdomains/
│   ├── all_subs.txt                  # Stage 1 output
│   ├── final_subs.txt                # Stage 2 output (DNS-verified)
│   └── passive/                      # Per-source raw results
├── probe/
│   ├── httpx/
│   │   ├── httpx_enriched.json       # JSONL — one host per line
│   │   └── path_hits.txt             # Pipe-delimited: URL|STATUS|SIZE
│   └── port-scan/
│       ├── domain_ips.json           # Domain → IP mappings + CDN flag
│       └── ports.json                # Masscan results
├── temp/
│   └── cdn_ranges.txt                # Cached CDN CIDR blocks (7-day TTL)
└── server2/                          # Go dashboard server
    ├── cmd/main.go                   # Entry point → server.Run()
    ├── go.mod                        # module: github.com/execute-assembly/recon-dashboard
    ├── databases/
    │   └── <domain>_db.sql           # Per-target SQLite databases
    ├── internal/
    │   ├── database/
    │   │   └── db_ops.go             # All DB logic (open, schema, queries)
    │   └── server/
    │       └── server.go             # Chi router, all HTTP handlers
    └── static/
        ├── target.html               # Target selection page (served at /)
        └── index.html                # Dashboard (served at /index.html)
```

---

## Recon Pipeline

```
Stage 1 — subdomain2.sh
  Tools: subfinder, crt.sh (curl), github-subdomains
  Output: subdomains/all_subs.txt

Stage 2 — subdomains_active.sh
  Tools: puredns (10k/2k rate limit), alterx (permutations)
  Resolvers: downloaded from trickest/resolvers, fallback to SecLists
  Output: subdomains/final_subs.txt

Stage 3 — port_scan.sh
  Tools: dnsx, grepcidr, masscan (sudo, ports 0-10000 at 20k pps)
  CDN ranges: Cloudflare, Fastly, AWS CloudFront, GCP, Akamai, Incapsula, Sucuri
  Output: probe/port-scan/domain_ips.json, probe/port-scan/ports.json

Stage 4 — alive_httpx_probe.sh
  Tools: httpx
  CDN domains default ports: 80,443,8080,8443,8000,8888,3000,9090,4443,8081,9443,2087,2083,9200
  Sensitive paths probed: .git/config, .env, robots.txt, sitemap.xml,
    crossdomain.xml, clientaccesspolicy.xml, .well-known/security.txt,
    api/swagger, api/swagger.json, api/openapi.json, v1/swagger,
    actuator, actuator/env, actuator/mappings, phpinfo.php,
    server-status, server-info, wp-admin, wp-config.php.bak, .DS_Store
  Match codes: 200, 201, 301, 302, 403
  Output: probe/httpx/httpx_enriched.json (JSONL), probe/httpx/path_hits.txt
```

Run everything: `./recon.sh <domain>` from the recon root.

---

## Go Server

**Module:** `github.com/execute-assembly/recon-dashboard`
**Go version:** 1.24.4
**Dependencies:** `github.com/go-chi/chi/v5`, `github.com/mattn/go-sqlite3`
**Runs on:** `http://127.0.0.1:8080`
**Run from:** recon root — paths like `./databases/` and `probe/httpx/` are relative to cwd

### Starting the server

```bash
cd /home/kali/recon
go run ./server2/cmd/main.go
```

---

## HTTP Routes

| Method | Path | Handler | Status | Description |
|--------|------|---------|--------|-------------|
| GET | `/` | `serveHTML` | ✅ Done | Serves `target.html` |
| GET | `/index.html` | `serveHTML` | ✅ Done | Serves `index.html` |
| GET | `/api/targets` | `TargetHandler` | ✅ Done | Lists `databases/*.sql` files |
| POST | `/api/targets/new` | `NewTargetHandler` | ✅ Done | Creates DB + schema for domain |
| POST | `/api/import/{domain}` | `ImportHandler` | 🔧 In progress | Reads probe files → upserts into DB |
| GET | `/api/hosts` | `HostHandler` | 🔧 Stub | Reads domains table → returns JSON |
| GET | `/api/hits` | `JuicyHandler` | 🔧 Stub | Reads juicy_hits table → returns JSON |
| PATCH | `/api/domains/{domain}/triage` | — | 📋 Planned | Updates triage_status for a host |
| PATCH | `/api/domains/{domain}/notes` | — | 📋 Planned | Updates notes for a host |

### Request / Response shapes

**POST /api/targets/new**
```json
// request
{ "domain": "powercor.com.au" }
// response 200
{ "domain": "powercor.com.au" }
// response 409
domain already exists
// response 400
domain cannot be empty
```

**GET /api/targets**
```json
{ "targets": ["powercor.com.au", "test.com.au"] }
```

**POST /api/import/{domain}**
```json
// response 200
{ "hosts_imported": 312, "hits_imported": 47 }
```

**GET /api/hosts** *(planned shape — must match index.html)*
```json
{
  "stats": { "total": 120, "s200": 80, "s403": 30, "s500": 5 },
  "hosts": [
    {
      "url": "https://admin.powercor.com.au",
      "status": 200, "sc": "s200",
      "title": "Admin Portal",
      "server": "nginx",
      "tech": ["PHP", "WordPress"],
      "ips": ["1.2.3.4"],
      "cname": [],
      "ctype": "text/html",
      "ports": [{"port": 443, "service": "HTTPS"}, {"port": 8080, "service": "HTTP-alt"}],
      "badges": ["interesting"],
      "triage_status": "to-test",
      "notes": "login panel, worth testing"
    }
  ]
}
```

**GET /api/hits** *(planned shape — must match index.html)*
```json
{
  "hits": [
    {
      "url": "https://admin.powercor.com.au/.env",
      "status": "200", "sc": "s200",
      "size": "1024",
      "severity": "HIGH"
    }
  ]
}
```

**PATCH /api/domains/{domain}/triage**
```json
// request — domain = target (e.g. powercor.com.au), domain_name = host URL
{ "domain": "powercor.com.au", "status": "to-test" }
```

**PATCH /api/domains/{domain}/notes**
```json
{ "domain": "powercor.com.au", "notes": "takes URL param at /search" }
```

---

## Database Schema

One SQLite file per target: `server2/databases/<domain>_db.sql`

```sql
CREATE TABLE IF NOT EXISTS domains (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_name  TEXT UNIQUE,   -- full URL e.g. https://admin.example.com:8443
    status_code  TEXT,
    open_ports   TEXT,          -- comma-separated: "80, 443, 8080"
    title        TEXT,
    tech_stack   TEXT,          -- comma-separated: "nginx, PHP"
    content_type TEXT,
    server       TEXT,
    ips          TEXT,          -- comma-separated: "1.2.3.4, 5.6.7.8"
    cname        TEXT,
    badges       TEXT,          -- comma-separated: "interesting,api"
    triage_status TEXT DEFAULT 'none',  -- none | to-test | dead-end | tested
    notes        TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS juicy_hits (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    url         TEXT UNIQUE,
    status_code TEXT,
    size        TEXT,
    severity    TEXT            -- HIGH | MEDIUM | INFO
);
```

### Severity classification (computed at import time)

| Severity | Patterns |
|----------|----------|
| HIGH | `.env`, `.git`, `wp-config`, `phpinfo`, `server-status`, `actuator/env` |
| MEDIUM | `actuator`, `swagger`, `openapi`, `server-info` |
| INFO | everything else |

### Badge classification (computed at import time)

| Badge | Trigger |
|-------|---------|
| `interesting` | URL or title matches: login, admin, dashboard, portal, jenkins, grafana, kibana, gitlab, jira, confluence, phpmyadmin, cpanel, wp-admin |
| `api` | URL or title matches: api, swagger, openapi, graphql |

---

## Data Formats (probe files on disk)

**`probe/httpx/httpx_enriched.json`** — JSONL, one JSON object per line
```json
{
  "url": "https://admin.powercor.com.au",
  "status_code": 200,
  "title": "Admin Portal",
  "tech": ["nginx", "PHP:7.4"],
  "content_type": "text/html",
  "webserver": "nginx/1.18",
  "a": ["1.2.3.4"],
  "cname": ["admin.cdn.example.com"],
  "open_ports": [80, 443, 8080]
}
```
Note: IP field is `"a"` (DNS A record), not `"ip"`.

**`probe/httpx/path_hits.txt`** — pipe-delimited, one hit per line
```
https://admin.powercor.com.au/.env|200|1024
https://api.powercor.com.au/actuator|200|512
```

---

## Frontend Pages

### target.html (/)
- Fetches `GET /api/targets` on load, renders a card per domain
- **Select** → `localStorage.setItem('recon_target', domain)` → navigate to `/index.html`
- **New Target** button → modal → `POST /api/targets/new` → refresh list

### index.html (/index.html)
- Reads `localStorage.getItem('recon_target')` for the active domain
- Three tabs: **Host Enumeration**, **Juicy Hits**, **JS Analysis**
- **Import button** (Hosts + Hits tabs) → `POST /api/import/{domain}` → reloads tables
- **Detail panel** (click any host row) → shows full host info, triage buttons, notes textarea
- Triage → `PATCH /api/domains/{domain}/triage`
- Notes save → `PATCH /api/domains/{domain}/notes`

---

## Build Order (remaining work)

1. `database.ImportData(domain)` — reads probe files, upserts into DB
2. `HostHandler` — queries domains table, crafts JSON matching index.html shape
3. `JuicyHandler` — queries juicy_hits table, crafts JSON matching index.html shape
4. `PATCH /api/domains/{domain}/triage` + handler
5. `PATCH /api/domains/{domain}/notes` + handler

---

## Security Notes

- GitHub token hardcoded in `recon-files/subdomain2.sh` line ~46 — **rotate and move to `$GITHUB_TOKEN` env var**
- Server binds to `127.0.0.1:8080` only — do not expose externally
- `masscan` requires sudo
- All recon must only target domains with explicit written authorisation
