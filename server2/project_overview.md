---
name: server2 Go dashboard
description: Architecture, routes, features, and todo for the server2/ Go recon dashboard
type: project
---

Go-powered recon dashboard in server2/. Replaces old Python server/ (ignore server/, use server2/).

**Why:** Per-target SQLite DBs, triage, notes, and a dark operator-style dashboard for bug bounty recon.

**How to apply:** All server work targets server2/. Run from recon root: `go run ./server2/cmd/main.go` → http://127.0.0.1:8080

---

## Key file structure

| File | Purpose |
|------|---------|
| `server2/cmd/main.go` | Entry point, calls `server.Run()` |
| `server2/internal/server/server.go` | Chi v5 router + all HTTP handlers |
| `server2/internal/database/types.go` | Structs: HttpxEntry, Host, HostResponse, HitResponse, Stats, HostsResult, PortServices |
| `server2/internal/database/db_ops.go` | getDB (cached connections), CreateNewTarget, ImportData, migrateDB, WriteNote, UpdateTriage |
| `server2/internal/database/importHelper.go` | ImportHttpx, ImportPathHits, computeBadges, severityFromStatus |
| `server2/internal/database/db_reads.go` | ReadHosts, ReadHits, transformHost, statusClass, splitTrim |
| `server2/static/target.html` | Target selection page (/) — table layout, async per-target stats, filter, delete |
| `server2/static/index.html` | Main dashboard (/index.html) — tabs: Host Enumeration, Juicy Hits, JS Analysis, Overview |
| `server2/static/css/dashboard.css` | Single stylesheet for index.html (Uncodixify spec, steel blue accent #4d9fff) |
| `server2/static/js/panel.js` | Detail side panel, triage buttons, notes save |
| `server2/static/js/hosts.js` | Hosts table render, group by hostname, expand/collapse child rows, triage tags |
| `server2/static/js/hits.js` | Juicy hits table render |
| `server2/static/js/overview.js` | Overview tab: sidebar list, host detail, screenshot, port scan, triage, notes |
| `server2/static/js/utils.js` | escHtml, showToast, filterTable, sortTable, renderBadges, showTab, collapseAll |
| `server2/databases/<domain>_db.sql` | Per-target SQLite files (gitignored) |
| `server2/backups/` | CSS/HTML backups: orange_dashboard.css, orange_target.html, dashboard_original.css, etc. |

---

## Active API routes

### Target-level
| Method | Route | Handler | Notes |
|--------|-------|---------|-------|
| GET | `/api/targets` | `TargetHandler` | Lists targets from `./databases/*_db.sql` |
| POST | `/api/targets/new` | `NewTargetHandler` | Creates new SQLite DB for domain |
| POST | `/api/import/{domain}` | `ImportHandler` | Reads probe JSON from disk, upserts into DB |
| GET | `/api/{domain}/hosts` | `HostHandler` | Returns `{ stats, hosts[] }` |
| GET | `/api/{domain}/hits` | `JuicyHandler` | Returns `{ hits[] }` |

### Host-level
| Method | Route | Handler | Notes |
|--------|-------|---------|-------|
| PATCH | `/api/{domain}/host/{hostURL}/triage` | `TriageHandler` | Body: `{ domain, status }` — updates triage_status |
| PATCH | `/api/{domain}/host/{hostURL}/notes` | `NotesHandler` | Body: `{ domain, notes }` — updates notes |

### Stubbed (commented out, not yet implemented)
| Method | Route | Notes |
|--------|-------|-------|
| POST | `/api/{domain}/host/{hostURL}/screenshot` | Frontend calls this; handler doesn't exist yet |
| POST | `/api/{domain}/host/{hostURL}/portscan` | Frontend calls this; handler doesn't exist yet |
| DELETE | `/api/delete/{domain}` | Frontend delete button calls this; handler doesn't exist yet |

### Static file serving
- `/css/*`, `/js/*` — served from `static/` via `http.FileServer` middleware wrapper (runs before Chi routing)
- `*.png`, `*.jpg`, `*.jpeg`, `*.gif`, `*.webp`, `/images/*` — also served via FileServer (added for screenshot display)
- `/index.html` → `static/index.html`
- `/*` (catch-all) → `static/target.html`

---

## Frontend features (implemented)

### target.html
- Table of targets (not cards) — domain, total hosts, 200 OK, 403, 5xx loaded async per-target from `/api/{domain}/hosts`
- Active target highlighted with accent left-border + "active" badge
- Filter input, target count, summary bar (totals across all targets)
- "+ New Target" modal → `POST /api/targets/new`
- Delete button per row → `DELETE /api/delete/{domain}` (frontend wired, backend stub missing)
- Clicking row or Select → sets `localStorage.recon_target`, redirects to `/index.html`

### index.html tabs
- **Host Enumeration** — sortable table, grouped by hostname with expand/collapse child rows, triage tags on rows, column toggles (IPs/CNAME/Content-Type), Import button
- **Juicy Hits** — path hits table grouped by domain, severity badges
- **JS Analysis** — JS endpoint/secret analysis table (expand per domain)
- **Overview** — collapsible sidebar (300px, ◀/▶ toggle) with host list + filter; main panel shows screenshot, port scan, host info grid, path hits, triage buttons, notes

### Detail panel (side panel on hosts tab)
- Opens on row click, shows all host fields
- Triage buttons (none/to-test/dead-end/tested) → `PATCH /api/{domain}/host/{hostURL}/triage`
- Notes textarea + Save → `PATCH /api/{domain}/host/{hostURL}/notes`
- "Overview ↗" button jumps to Overview tab and selects host

### Overview tab actions
- **Take Screenshot** button → `POST /api/{domain}/host/{hostURL}/screenshot` (backend stub missing)
- **Scan Ports** button → `POST /api/{domain}/host/{hostURL}/portscan` (backend stub missing)
- Screenshot auto-loads on host select via `GET /api/{domain}/host/{hostURL}/screenshot`

---

## DB schema (current — flat, comma-separated)

```sql
CREATE TABLE domains (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  domain_name    TEXT UNIQUE,
  status         TEXT,
  title          TEXT,
  server         TEXT,
  tech           TEXT,  -- comma-separated
  ports          TEXT,  -- JSON: [{port, service}]
  ips            TEXT,  -- comma-separated
  cname          TEXT,  -- comma-separated
  ctype          TEXT,
  triage_status  TEXT DEFAULT 'none',
  notes          TEXT DEFAULT '',
  badges         TEXT   -- JSON array
);

CREATE TABLE path_hits (
  id       INTEGER PRIMARY KEY AUTOINCREMENT,
  domain   TEXT,
  url      TEXT,
  status   INTEGER,
  size     INTEGER,
  severity TEXT
);
```

---

## Important data notes
- Probe files read from `../probe/httpx/<domain>_httpx_enriched.json` (JSONL) and `../probe/httpx/<domain>_path_hits.txt`
- IP field in httpx JSON is `"a"` (DNS A record) → `HttpxEntry.IPs []string \`json:"a"\``
- ON CONFLICT upsert on `domain_name` preserves `triage_status`/`notes` on re-import
- Severity: 2xx→high, 5xx→medium, else→low
- Badges: "interesting" (login/admin/dashboard etc), "api" (api/swagger/openapi/graphql)
- `hostURL` Chi URL params must be `url.QueryUnescape`d — stored decoded in DB

## Color theme
- Steel blue accent: `--accent: #4d9fff`, `--accent-dim: #0a1a30`
- Orange backup available: `server2/backups/orange_dashboard.css` + `orange_target.html`
- Follows Uncodixify spec (skills/Deign.md): no uppercase label overload, no transform animations, shadow max 8px, opacity transitions only

---

## TODO next

1. **Screenshot route** — implement `POST /api/{domain}/host/{hostURL}/screenshot`
   - Run `gowitness` or `chromium --headless` against the host URL, save PNG to `./screenshots/{domain}/{encoded_url}.png`
   - Implement `GET /api/{domain}/host/{hostURL}/screenshot` to serve the PNG file
   - Remove hardcoded `/test-screenshot.png` from `overview.js`

2. **Port scan route** — implement `POST /api/{domain}/host/{hostURL}/portscan`
   - Fire-and-forget pattern: POST starts scan (returns immediately), poll `GET /api/{domain}/host/{hostURL}/portscan` for status + results
   - Run `nmap` or `masscan` against resolved IPs for the host
   - Store results in DB (see TODO #3 for schema)

3. **DB schema refactor** — remove comma-separated columns, add junction tables
   - `ips` table: `(id, domain_id, ip TEXT)`
   - `cnames` table: `(id, domain_id, cname TEXT)`
   - `tech` table: `(id, domain_id, tech TEXT)`
   - `open_ports` table: `(id, ip TEXT, port INT, service TEXT)` — keyed by IP not domain, since IPs can be shared across domains
   - `ReadHosts` does Go-side join to assemble `Host` struct (avoids complex SQL joins)
   - Migration: write `migrateDB` version bump to create new tables, backfill from existing comma-separated data

4. **Delete target route** — implement `DELETE /api/delete/{domain}`
   - Remove `./databases/{domain}_db.sql`
   - Frontend already wired: sends `DELETE /api/delete/{domain}`, removes row on 2xx
