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
│   ├── port_scan.sh <domain>         # Stage 3: CDN detection + masscan
│   ├── alive_httpx_probe.sh <domain> # Stage 4: HTTP probing + path discovery
│   └── crawl.sh                      # Optional: web crawling (katana/gau/wayback)
├── subdomains/
│   ├── all_subs.txt                  # Stage 1 output
│   ├── final_subs.txt                # Stage 2 output (DNS-verified)
│   └── passive/                      # Per-source raw results
├── probe/
│   ├── httpx/
│   │   ├── <domain>_httpx_enriched.json   # JSONL — one host per line
│   │   ├── <domain>_path_hits.txt         # Pipe-delimited: URL|STATUS|SIZE
│   │   ├── <domain>_path_hits_raw.json    # Raw httpx output for path hits
│   │   └── <domain>_targets.txt           # Built domain:port pairs for httpx
│   └── port-scan/
│       ├── <domain>_domain_ips.json  # Domain → IP mappings + CDN flag
│       └── <domain>_ports.json       # Masscan results (IP → open ports)
├── temp/
│   └── cdn_ranges.txt                # Cached CDN CIDR blocks (7-day TTL)
└── server2/                          # Go dashboard server
    ├── cmd/main.go                   # Entry point → server.Run()
    ├── go.mod                        # module: github.com/execute-assembly/recon-dashboard
    ├── databases/
    │   └── <domain>_db.sql           # Per-target SQLite databases
    ├── internal/
    │   ├── database/
    │   │   ├── types.go              # All structs + PortServices map
    │   │   ├── db_ops.go             # DB open/cache, CreateNewTarget, ImportData, migrateDB
    │   │   ├── db_reads.go           # ReadHosts, ReadHits, transformHost, statusClass, splitTrim
    │   │   └── importHelper.go       # ImportHttpx, ImportPathHits, computeBadges, joinInts
    │   └── server/
    │       └── server.go             # Chi router + all HTTP handlers
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

Stage 3 — port_scan.sh <domain>
  Tools: dnsx, grepcidr, masscan (sudo, ports 0-10000 at 20k pps)
  CDN ranges: Cloudflare, Fastly, AWS CloudFront, GCP, Akamai, Incapsula, Sucuri
  Output: probe/port-scan/<domain>_domain_ips.json, probe/port-scan/<domain>_ports.json

Stage 4 — alive_httpx_probe.sh <domain>
  Tools: httpx
  CDN domains default ports: 80,443,8080,8443,8000,8888,3000,9090,4443,8081,9443,2087,2083,9200
  Sensitive paths probed: .git/config, .env, robots.txt, sitemap.xml,
    crossdomain.xml, clientaccesspolicy.xml, .well-known/security.txt,
    api/swagger, api/swagger.json, api/openapi.json, v1/swagger,
    actuator, actuator/env, actuator/mappings, phpinfo.php,
    server-status, server-info, wp-admin, wp-config.php.bak, .DS_Store
  Match codes: 200, 201, 301, 302, 403
  Output: probe/httpx/<domain>_httpx_enriched.json (JSONL)
          probe/httpx/<domain>_path_hits.txt (pipe-delimited: URL|STATUS|SIZE)
```

Run everything: `./recon.sh <domain>` from the recon root.

---

## Go Server

**Module:** `github.com/execute-assembly/recon-dashboard`
**Dependencies:** `github.com/go-chi/chi/v5`, `github.com/mattn/go-sqlite3`
**Runs on:** `http://127.0.0.1:8080`
**Run from:** recon root — all paths (`./databases/`, `../probe/httpx/`) are relative to cwd

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
| POST | `/api/import/{domain}` | `ImportHandler` | ✅ Done | Reads probe files → upserts into DB |
| GET | `/api/{domain}/hosts` | `HostHandler` | ✅ Done | Queries domains table → HostsResult JSON |
| GET | `/api/{domain}/hits` | `JuicyHandler` | ✅ Done | Queries juicy_hits → `{"hits":[...]}` |
| PATCH | `/api/{domain}/triage` | — | 📋 Planned | Updates triage_status for a host row |
| PATCH | `/api/{domain}/notes` | — | 📋 Planned | Updates notes for a host row |

### Request / Response shapes

**POST /api/targets/new**
```json
// request
{ "domain": "powercor.com.au" }
// response 200
{ "domain": "powercor.com.au" }
// response 400 (already exists or empty)
domain already exists
```

**GET /api/targets**
```json
{ "targets": ["powercor.com.au", "test.com.au"] }
```

**POST /api/import/{domain}** — body empty, domain in URL
```json
// response 200
{ "domain": "good" }
```

**GET /api/{domain}/hosts**
```json
{
  "stats": { "total": 120, "s200": 80, "s403": 30, "s500": 5 },
  "hosts": [
    {
      "id": 1,
      "url": "https://admin.powercor.com.au",
      "status": "200",
      "sc": "s200",
      "title": "Admin Portal",
      "server": "nginx",
      "tech": ["PHP", "WordPress"],
      "ips": ["1.2.3.4"],
      "cname": [],
      "ctype": "text/html",
      "ports": [{"port": "443", "service": "HTTPS"}],
      "badges": ["interesting"],
      "triage_status": "to-test",
      "notes": "login panel, worth testing"
    }
  ]
}
```

**GET /api/{domain}/hits**
```json
{
  "hits": [
    {
      "url": "https://admin.powercor.com.au/.env",
      "status": "200",
      "sc": "s200",
      "size": "1024",
      "severity": "high"
    }
  ]
}
```

**PATCH /api/{domain}/triage** *(planned)*
```json
// request body: { "domain_name": "https://admin.powercor.com.au", "status": "to-test" }
// triage_status values: none | to-test | dead-end | tested
```

**PATCH /api/{domain}/notes** *(planned)*
```json
// request body: { "domain_name": "https://admin.powercor.com.au", "notes": "SQL injection at /search" }
```

---

## Database Schema

One SQLite file per target: `server2/databases/<domain>_db.sql`

```sql
CREATE TABLE IF NOT EXISTS domains (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_name   TEXT UNIQUE,      -- full URL e.g. https://admin.example.com:8443
    status_code   TEXT,
    open_ports    TEXT,             -- comma-separated: "80, 443, 8080"
    title         TEXT,
    tech_stack    TEXT,             -- comma-separated: "nginx, PHP"
    content_type  TEXT,
    server        TEXT,
    ips           TEXT,             -- comma-separated: "1.2.3.4, 5.6.7.8"
    cname         TEXT,
    badges        TEXT,             -- comma-separated: "interesting,api"
    triage_status TEXT NOT NULL DEFAULT '',  -- none | to-test | dead-end | tested
    notes         TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS juicy_hits (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    url         TEXT UNIQUE,
    status_code TEXT,
    size        TEXT,
    severity    TEXT               -- high | medium | low
);
```

### Migration

`migrateDB()` runs on every `getDB()` call (i.e. first use of any domain). It runs:
```sql
ALTER TABLE domains ADD COLUMN triage_status TEXT NOT NULL DEFAULT ''
ALTER TABLE domains ADD COLUMN notes TEXT NOT NULL DEFAULT ''
```
SQLite returns an error if the column already exists — this error is silently ignored. Safe to run repeatedly.

### Severity classification (computed at import from path_hits.txt)

| Severity | HTTP Status |
|----------|-------------|
| `high` | 2xx |
| `medium` | 5xx |
| `low` | everything else (403 etc.) |

### Badge classification (computed at import from httpx_enriched.json)

| Badge | Trigger words in URL or title |
|-------|-------------------------------|
| `interesting` | login, admin, dashboard, portal, jenkins, grafana, kibana, gitlab, jira, confluence, phpmyadmin, cpanel, wp-admin |
| `api` | api, swagger, openapi, graphql |

---

## Data Formats (probe files on disk)

**`probe/httpx/<domain>_httpx_enriched.json`** — JSONL, one JSON object per line
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
**Important:** IP field is `"a"` (DNS A record), not `"ip"`. This maps to `HttpxEntry.IPs []string \`json:"a"\``.

**`probe/httpx/<domain>_path_hits.txt`** — pipe-delimited, one hit per line
```
https://admin.powercor.com.au/.env|200|1024
https://api.powercor.com.au/actuator|200|512
```

---

## Go Package Structure

### `internal/database/types.go`
All shared types:
- `HttpxEntry` — JSON struct for parsing httpx_enriched.json lines
- `Host` — raw DB row (all string fields, comma-sep arrays)
- `HostResponse` — JSON response shape (arrays split out, ports mapped to services)
- `Port` — `{port string, service string}`
- `HitResponse` — JSON response for juicy_hits rows
- `Stats` — `{total, s200, s403, s500}`
- `HostsResult` — `{stats Stats, hosts []HostResponse}` — top-level /hosts response
- `PortServices` — `map[int]string` mapping port numbers to service names

### `internal/database/db_ops.go`
- `getDB(domain)` — opens/caches SQLite connection, runs `migrateDB`
- `migrateDB(db)` — adds triage_status + notes columns if missing (errors ignored)
- `CreateNewTarget(name)` — checks existence, creates DB file + both tables
- `ImportData(domain)` — calls `ImportHttpx` then `ImportPathHits`

### `internal/database/importHelper.go`
- `ImportHttpx(domain)` — reads `../probe/httpx/<domain>_httpx_enriched.json`, bulk upserts into `domains` table; ON CONFLICT preserves triage_status/notes
- `ImportPathHits(domain)` — reads `../probe/httpx/<domain>_path_hits.txt`, bulk upserts into `juicy_hits`
- `computeBadges(url, title)` — returns comma-sep badge string
- `joinInts([]int)` — converts port int slice to comma-sep string
- `severityFromStatus(status)` — 2xx→high, 5xx→medium, else→low

### `internal/database/db_reads.go`
- `ReadHosts(domain)` — returns `HostsResult` with inline stats
- `ReadHits(domain)` — returns `[]HitResponse` (caller wraps in `{"hits":[...]}`)
- `transformHost(h Host)` — splits comma-sep strings, maps ports to service names, computes SC class
- `statusClass(code)` — `"200"→"s200"`, `"403"→"s403"`, `"4xx"→"s400"`, else `""`
- `splitTrim(s)` — splits comma-separated string, trims whitespace, drops empties

### `internal/server/server.go`
- Chi v5 router
- `serveHTML(path)` — uses `os.ReadFile` (NOT `http.ServeFile` — that redirects on dirs)
- All handlers: `HostHandler`, `JuicyHandler`, `ImportHandler`, `TargetHandler`, `NewTargetHandler`

---

## Frontend Pages

### target.html (/)
- Fetches `GET /api/targets` on load, renders a card per domain
- **Select** → `localStorage.setItem('recon_target', domain)` → navigate to `/index.html`
- **New Target** button → modal → `POST /api/targets/new` → refresh list
- Same dark theme CSS variables as index.html

### index.html (/index.html)
- Reads `localStorage.getItem('recon_target')` for the active domain
- Three tabs: **Host Enumeration**, **Juicy Hits**, **JS Analysis**
- **Import button** (Hosts + Hits tabs) → spinner → `POST /api/import/${domain}` → toast notification → reloads table
- Toast: top-right, green/red, slides down, 3.5s auto-dismiss
- **Detail panel** (click any host row) → slides in from right, shows full host info
  - Triage buttons: none / to-test / dead-end / tested → `PATCH /api/{domain}/triage` *(handler pending)*
  - Notes textarea + Save button → `PATCH /api/{domain}/notes` *(handler pending)*

---

## Remaining Work

1. **PATCH `/api/{domain}/triage`** handler + `UpdateTriage(domain, domainName, status string)` DB function
2. **PATCH `/api/{domain}/notes`** handler + `UpdateNotes(domain, domainName, notes string)` DB function
3. Wire triage/notes PATCH calls in index.html detail panel JS

---

## Security Notes

- GitHub token hardcoded in `recon-files/subdomain2.sh` line ~46 — **rotate and move to `$GITHUB_TOKEN` env var**
- Server binds to `127.0.0.1:8080` only — do not expose externally
- `masscan` requires sudo
- All recon must only target domains with explicit written authorisation
