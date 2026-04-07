---
name: server2 Go dashboard
description: Architecture, routes, features, and todo for the Vantage Go recon dashboard
type: project
---

Go-powered recon dashboard (Vantage) in server2/. Ignore old server/ directory.

**How to run:** `go run ./server2/cmd/main.go` → http://127.0.0.1:8080 (nginx proxies port 80)
**Build:** `./server2/build.sh [all|frontend|backend]`

---

## Key file structure

### Backend (Go)
| File | Purpose |
|------|---------|
| `server2/cmd/main.go` | Entry point, sets up logging (~/.recon/logs/recon.log), calls `server.Run()` |
| `server2/internal/server/server.go` | Chi v5 router, auth middleware, static file serving |
| `server2/internal/server/routes.go` | All HTTP handlers + auth (sessions, login, middleware) |
| `server2/internal/database/types.go` | Structs: HttpxEntry, Host, HostResponse, HitResponse, Stats, HostsResult, PortServices |
| `server2/internal/database/db_ops.go` | getDB, CreateNewTarget, ImportData, migrateDB, WriteNote, UpdateTriage, DeleteData, DbDir() |
| `server2/internal/database/importHelper.go` | ImportHttpx, ImportPathHits, computeBadges, severityFromStatus |
| `server2/internal/database/db_reads.go` | ReadHosts, ReadHits, GetStats, transformHost, statusClass, splitTrim |
| `server2/internal/tools/screenshot.go` | Async screenshot job system: job map, RWMutex, Screenshot(), SanitizeForFilename() |
| `server2/internal/tools/ai.go` | SendTelegram(), RunWorkFlow() — runs recon.sh, ingests data, sends Telegram summary |

### Frontend (React + TypeScript)
| File | Purpose |
|------|---------|
| `server2/frontend/src/main.tsx` | React entry point |
| `server2/frontend/src/App.tsx` | React Router — `/login` → LoginPage, `/` → TargetsPage, `/dashboard` → DashboardPage |
| `server2/frontend/src/lib/types.ts` | TypeScript interfaces + `fetchApi()` wrapper (redirects to /login on 401) |
| `server2/frontend/src/pages/LoginPage.tsx` | Login form — POST /api/login, redirects to /goaway on bad creds |
| `server2/frontend/src/pages/TargetsPage.tsx` | Target selection — table, stats, new target modal, delete |
| `server2/frontend/src/pages/DashboardPage.tsx` | Dashboard shell — tab switcher, fetches hosts + hits |
| `server2/frontend/src/pages/HostsTab.tsx` | Host enumeration table — virtual list, group/expand, filter, sort, hide, triage tags, import |
| `server2/frontend/src/pages/HitsTab.tsx` | Juicy hits table — severity badges, filter, sort by severity |
| `server2/frontend/src/pages/OverviewTab.tsx` | Overview — sidebar host list, screenshot viewer/capture, host info, triage, notes |
| `server2/frontend/src/pages/HostPanel.tsx` | Side panel — host detail, triage, notes, "Overview ↗" link |
| `server2/frontend/src/styles/globals.css` | Supabase-inspired dark theme, emerald accent #3ecf8e |
| `server2/frontend/vite.config.ts` | Vite build config — dev proxy `/api` + `/images` → http://127.0.0.1:8080 |

### Data paths (all under ~/.recon/)
| Path | Purpose |
|------|---------|
| `~/.recon/databases/<domain>_db.sql` | Per-target SQLite files |
| `~/.recon/logs/recon.log` | Server logs |
| `~/.recon/sessions.json` | Persisted login sessions (survives restarts) |
| `~/.recon/<domain>/probe/httpx/` | Recon script output consumed by ImportData |
| `server2/static/dist/` | Vite build output — served by Go |
| `server2/static/images/screenshots/` | Cached screenshot files |

---

## Auth
- Cookie-based session auth (`recon_session` cookie)
- Sessions persisted to `~/.recon/sessions.json` — survive server restarts
- TTL: 30 days
- Credentials: hardcoded in `routes.go` `Login_Handler` (user/pass vars)
- Wrong credentials → redirect to `/goaway` (returns HTML "Stop looking here")
- `authMiddleware` protects all `/api/*` routes except `/api/login`
- Frontend `fetchApi()` wrapper redirects to `/login` on any 401
- Login attempts logged via `slog` (success=INFO, fail=WARN) with username + IP

---

## Active API routes

### Auth
| Method | Route | Handler | Notes |
|--------|-------|---------|-------|
| POST | `/api/login` | `Login_Handler` | Body: `{ username, password }` → sets session cookie |
| GET | `/goaway` | `GoAway_Handler` | Honeypot page for failed logins |

### Target-level (all protected)
| Method | Route | Handler | Notes |
|--------|-------|---------|-------|
| GET | `/api/targets` | `Targets_Handler` | Lists targets from `~/.recon/databases/` |
| POST | `/api/targets/new` | `NewTargetHandler` | Creates new SQLite DB for domain |
| POST | `/api/import/{domain}` | `ImportHandler` | Reads probe JSON from disk, upserts into DB |
| DELETE | `/api/delete/{domain}` | `deleteTargetHandler` | Deletes DB file |
| GET | `/api/{domain}/hosts` | `Host_Handler` | Returns `{ stats, hosts[] }` |
| GET | `/api/{domain}/hits` | `Juicy_Handler` | Returns `{ hits[] }` |
| POST | `/api/run` | `RunRecon_Handler` | Triggers RunWorkFlow(domain) in goroutine |

### Host-level (all protected)
| Method | Route | Handler | Notes |
|--------|-------|---------|-------|
| PATCH | `/api/{domain}/host/{hostURL}/triage` | `Triage_Handler` | Body: `{ domain, status }` |
| PATCH | `/api/{domain}/host/{hostURL}/notes` | `Notes_Handler` | Body: `{ domain, notes }` |
| POST | `/api/{domain}/host/{hostURL}/screenshot` | `ScreenShot_Handler` | Starts gowitness job, returns `{ token }` |
| GET | `/api/{domain}/host/{hostURL}/screenshot/status` | `ScreenShotStatus_Handler` | Poll `?token=<uuid>` → pending/done/failed |
| GET | `/api/{domain}/host/{hostURL}/screenshot` | `ScreenShotServe_Handler` | Serves cached screenshot image |

---

## Automated recon pipeline (tools/ai.go)
- `POST /api/run {"domain":"example.com"}` → fires goroutine immediately, returns 200
- `RunWorkFlow(domain)`:
  1. `exec.Command("../recon.sh", domain)` with `cmd.Dir = ".."`
  2. On error → `SendTelegram("❌ recon failed...")`
  3. `database.CreateNewTarget(domain)` (ignores ErrDomainExists)
  4. `database.ImportData(domain)` — runs ImportHttpx + ImportPathHits
  5. `database.GetStats(domain)` — counts from DB
  6. `SendTelegram("✅ recon done...")` with host/hit counts
- Telegram env vars: `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`

---

## DB schema
```sql
CREATE TABLE domains (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  domain_name    TEXT UNIQUE,
  status_code    TEXT,
  open_ports     TEXT,  -- comma-separated
  title          TEXT,
  tech_stack     TEXT,  -- comma-separated
  content_type   TEXT,
  server         TEXT,
  ips            TEXT,  -- comma-separated
  cname          TEXT,  -- comma-separated
  badges         TEXT,  -- comma-separated: "interesting,api"
  triage_status  TEXT DEFAULT '',
  notes          TEXT DEFAULT '',
  tier_tag       TEXT DEFAULT '',
  tier_reason    TEXT DEFAULT ''
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
- Probe files read from `~/.recon/<domain>/probe/httpx/<domain>_httpx_enriched.json` (JSONL) and `<domain>_path_hits.txt`
- IP field in httpx JSON is `"a"` (DNS A record)
- ON CONFLICT upsert on `domain_name` preserves `triage_status`/`notes` on re-import
- Severity: 2xx→high, 5xx→medium, else→low
- Badges: "interesting" (login/admin/dashboard/portal/jenkins/kibana etc), "api" (api/swagger/openapi/graphql)
- `hostURL` Chi URL params must be `url.QueryUnescape`d

---

## TODO

1. **Automated recon tab** — frontend UI to trigger `POST /api/run`, show job status
2. **Port scanning** — implement `POST /api/{domain}/host/{hostURL}/portscan` (nmap/masscan, fire-and-forget like screenshots)
3. **JS secrets/routes tab** — replace stub, surface endpoints/secrets from JS files
4. **DB schema refactor** — remove comma-separated columns, add junction tables for ips/cnames/tech/ports
