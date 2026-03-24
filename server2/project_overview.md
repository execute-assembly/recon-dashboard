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

### Backend (Go)
| File | Purpose |
|------|---------|
| `server2/cmd/main.go` | Entry point, sets up logging, calls `server.Run()` |
| `server2/internal/server/server.go` | Chi v5 router, middleware, static file serving |
| `server2/internal/server/routes.go` | All HTTP handler functions |
| `server2/internal/database/types.go` | Structs: HttpxEntry, Host, HostResponse, HitResponse, Stats, HostsResult, PortServices |
| `server2/internal/database/db_ops.go` | getDB (cached connections), CreateNewTarget, ImportData, migrateDB, WriteNote, UpdateTriage, DeleteData |
| `server2/internal/database/importHelper.go` | ImportHttpx, ImportPathHits, computeBadges, severityFromStatus |
| `server2/internal/database/db_reads.go` | ReadHosts, ReadHits, ReadHostsForAI, transformHost, statusClass, splitTrim, DomainForAI struct |
| `server2/internal/tools/screenshot.go` | Async screenshot job system: job map, RWMutex, GetJob/SetJob, Screenshot(), SanitizeForFilename() |
| `server2/internal/tools/ai.go` | AI triage: systemPrompt, AnalyiseDomains(), AnlyiseBatch() — calls Haiku, strips markdown, prints raw JSON |

### Frontend (React + TypeScript)
| File | Purpose |
|------|---------|
| `server2/frontend/src/main.tsx` | React entry point |
| `server2/frontend/src/App.tsx` | React Router setup — `/` → TargetsPage, `/dashboard` → DashboardPage |
| `server2/frontend/src/lib/types.ts` | TypeScript interfaces: Host, Hit, HostStats |
| `server2/frontend/src/pages/TargetsPage.tsx` | Target selection — table, stats, new target modal, delete |
| `server2/frontend/src/pages/DashboardPage.tsx` | Dashboard shell — tab switcher, fetches hosts + hits, passes to tabs |
| `server2/frontend/src/pages/HostsTab.tsx` | Host enumeration table — virtual list, group/expand, filter, sort, hide, triage tags, import |
| `server2/frontend/src/pages/HitsTab.tsx` | Juicy hits table — severity badges, filter, sort by severity |
| `server2/frontend/src/pages/OverviewTab.tsx` | Overview — sidebar host list, screenshot viewer/capture, host info, triage, notes |
| `server2/frontend/src/pages/HostPanel.tsx` | Side panel — host detail, triage, notes, "Overview ↗" link |
| `server2/frontend/src/styles/globals.css` | Dark theme (Uncodixify spec, steel blue #4d9fff accent) |
| `server2/frontend/vite.config.ts` | Vite build config — dev proxy `/api` + `/images` → http://127.0.0.1:8080 |

### Built output + data
| Path | Purpose |
|------|---------|
| `server2/static/dist/` | Vite build output — served by Go FileServer |
| `server2/static/images/screenshots/` | Cached screenshot files (served at `/images/screenshots/`) |
| `server2/databases/<domain>_db.sql` | Per-target SQLite files (gitignored) |
| `server2/backups/` | Old CSS/HTML backups (orange theme, originals) — not in use |

> **Note:** `server2/static/js/` and `server2/static/css/` are the old vanilla JS/CSS — no longer served. Everything is React now via `static/dist/`.

---

## Active API routes

### Target-level
| Method | Route | Handler | Notes |
|--------|-------|---------|-------|
| GET | `/api/targets` | `Targets_Handler` | Lists targets from `./databases/*_db.sql` |
| POST | `/api/targets/new` | `NewTargetHandler` | Creates new SQLite DB for domain |
| POST | `/api/import/{domain}` | `ImportHandler` | Reads probe JSON from disk, upserts into DB |
| DELETE | `/api/delete/{domain}` | `deleteTargetHandler` | Deletes DB file + all data |
| GET | `/api/{domain}/hosts` | `Host_Handler` | Returns `{ stats, hosts[] }` |
| GET | `/api/{domain}/hits` | `Juicy_Handler` | Returns `{ hits[] }` |
| POST | `/api/{domain}/ai/domains` | `AiDomain_Handler` | Runs Haiku triage on all hosts, batched 50 at a time |

### Host-level
| Method | Route | Handler | Notes |
|--------|-------|---------|-------|
| PATCH | `/api/{domain}/host/{hostURL}/triage` | `Triage_Handler` | Body: `{ domain, status }` |
| PATCH | `/api/{domain}/host/{hostURL}/notes` | `Notes_Handler` | Body: `{ domain, notes }` |
| POST | `/api/{domain}/host/{hostURL}/screenshot` | `ScreenShot_Handler` | Starts gowitness job, returns `{ token }` |
| GET | `/api/{domain}/host/{hostURL}/screenshot/status` | `ScreenShotStatus_Handler` | Poll with `?token=<uuid>` → pending/done/failed |
| GET | `/api/{domain}/host/{hostURL}/screenshot` | `ScreenShotServe_Handler` | Serves cached screenshot image |

### Static file serving
- `/dist/*` — Vite-built React app from `static/dist/`
- `/images/*` — screenshots from `static/images/`
- `/*` (catch-all) → `static/dist/index.html` (React SPA)

---

## Frontend features

### TargetsPage (`/`)
- Table of targets — domain, total hosts, 200 OK, 403, 5xx loaded async per-target
- Active target highlighted with accent left-border + "active" badge
- Filter input, target count, summary bar across all targets
- "+ New Target" modal → `POST /api/targets/new`
- Delete button per row → `DELETE /api/delete/{domain}`
- Click row → sets `localStorage.recon_target`, navigates to `/dashboard`

### DashboardPage (`/dashboard`)
- Tab switcher: Host Enumeration, Juicy Hits, JS Analysis, Overview
- Fetches hosts + hits on mount, passes down to tabs
- Tab counts shown in header (total hosts, hits count)
- "← domain" link back to targets

### Host Enumeration tab
- Virtualised list (TanStack React Virtual) for performance
- Hosts grouped by hostname — expand/collapse child rows (alt ports)
- Filter by URL, title, server
- Sort by any column
- Column visibility toggles: IPs, CNAME, Content-Type
- Triage tags shown inline on collapsed rows (hidden when expanded to prevent overlap)
- **Hide/unhide rows** — hover a row to reveal a `hide` button; hides the entire hostname group. Hidden state persisted to `localStorage` as `recon_hidden_<domain>`. Toolbar shows "○ N hidden" button to reveal hidden rows as dimmed entries with `unhide` button.
- Import button → `POST /api/import/{domain}`
- Click row → opens HostPanel side panel
- "Overview ↗" from panel → switches to Overview tab with host selected

### Juicy Hits tab
- Table: severity badge (high/medium/low, colour-coded red/orange/yellow), status code (colour-coded), size, URL (clickable, opens in new tab)
- Filter by URL, severity, status code
- Sort by severity toggle (asc/desc)
- Hit counts in header: high / medium / low / total
- Empty state messages

### JS Analysis tab
- Stub — not yet implemented

### Overview tab
- Collapsible sidebar (300px) — host list with filter, click to select
- Screenshot viewer — auto-loads existing screenshot on host select
- "Take Screenshot" button → `POST /api/.../screenshot`, polls status every 1.5s, displays on done
- Scan Ports button — stubbed
- Host info grid: URL, status, title, server, tech, ports, IPs, CNAME, content-type
- Related path hits for selected host
- Triage buttons + notes textarea with Save

### HostPanel (side panel)
- Opens on hosts table row click (ESC to close)
- All host fields in structured layout
- Triage buttons with live PATCH update
- Notes textarea + Save
- "Overview ↗" link

---

## DB schema (current — flat, comma-separated)

```sql
CREATE TABLE domains (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  domain_name    TEXT UNIQUE,
  status_code    TEXT,
  open_ports     TEXT,  -- comma-separated: "80, 443"
  title          TEXT,
  tech_stack     TEXT,  -- comma-separated
  content_type   TEXT,
  server         TEXT,
  ips            TEXT,  -- comma-separated
  cname          TEXT,  -- comma-separated
  badges         TEXT,  -- comma-separated: "interesting,api"
  triage_status  TEXT DEFAULT '',
  notes          TEXT DEFAULT '',
  tier_tag       TEXT DEFAULT '',  -- tier1, tier2, tier3 (set by AI)
  tier_reason    TEXT DEFAULT ''   -- short reason from AI (≤5 words)
);

CREATE TABLE juicy_hits (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  url          TEXT UNIQUE,
  status_code  TEXT,
  size         TEXT,
  severity     TEXT  -- high, medium, low
);
```

---

## Important data notes
- Probe files read from `../probe/httpx/<domain>_httpx_enriched.json` (JSONL) and `../probe/httpx/<domain>_path_hits.txt`
- IP field in httpx JSON is `"a"` (DNS A record) → `HttpxEntry.IPs []string \`json:"a"\``
- ON CONFLICT upsert on `domain_name` preserves `triage_status`/`notes` on re-import
- Severity: 2xx→high, 5xx→medium, else→low
- Badges: "interesting" (login/admin/dashboard/portal/jenkins/kibana etc), "api" (api/swagger/openapi/graphql)
- `hostURL` Chi URL params must be `url.QueryUnescape`d — stored decoded in DB

## Color theme
- Steel blue accent: `--accent: #4d9fff`, `--accent-dim: #0a1a30`
- Status colours: green (2xx), orange (3xx), red (403/4xx/5xx)
- Orange backup available in `server2/backups/` — not in use
- Follows Uncodixify spec: no uppercase label overload, no transform animations, shadow max 8px, opacity transitions only

---

## TODO

1. **AI domain triage — finish wiring** — backend calls Haiku and gets JSON back, next steps:
   - Add `tier_tag` and `tier_reason` columns to `domains` table via `migrateDB` version bump
   - Unmarshal Haiku JSON response into `TierResult` structs (strip markdown fences first)
   - `UPDATE domains SET tier_tag=$1, tier_reason=$2 WHERE domain_name=$3` for each result
   - Add `tier_tag`/`tier_reason` to `HostResponse` and return from `ReadHosts`
   - Frontend: add tier1/tier2/tier3 options to the tag filter dropdown in HostsTab
   - Frontend: show tier badge on host rows alongside existing triage/badge tags
   - Frontend: "AI Triage" button in toolbar → `POST /api/{domain}/ai/domains`

2. **Finish port scanning** — implement `POST /api/{domain}/host/{hostURL}/portscan`
   - Fire-and-forget pattern matching screenshot system: POST returns token immediately, poll GET for status + results
   - Run `nmap` or `masscan` against resolved IPs
   - Store results in DB, display in Overview tab (wire up "Scan Ports" button)

3. **Finish JS secrets/routes tab** — replace stub in DashboardPage
   - Parse JS files found during httpx scan
   - Surface endpoints, secrets, API keys
   - Table per domain: JS file URL, finding type, value, severity

4. **DB schema refactor** — remove comma-separated columns, add junction tables
   - `ips`, `cnames`, `tech`, `open_ports` tables keyed by domain_id
   - `ReadHosts` does Go-side join to assemble `Host` struct
   - Migration via `migrateDB` version bump + backfill from comma-separated data

5. ~~**Screenshot route**~~ — ✅ done (async job system, gowitness, polling, serve image)
6. ~~**Delete target route**~~ — ✅ done (`deleteTargetHandler` + `database.DeleteData()`)
7. ~~**Juicy Hits tab**~~ — ✅ done (React migration complete)
8. ~~**AI route + Haiku integration**~~ — ✅ done (endpoint live, batching works, JSON response printing correctly)
