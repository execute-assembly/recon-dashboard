# Uncodixify — Pentesting Recon Triage Dashboard

This document exists to teach you how to act as non-Codex as possible when building UI for a pentesting recon triage dashboard.

Codex UI is the default AI aesthetic: soft gradients, floating panels, eyebrow labels, decorative copy, hero sections in dashboards, oversized rounded corners, transform animations, dramatic shadows, and layouts that try too hard to look premium. It's the visual language that screams "an AI made this" because it follows the path of least resistance.

The second trap is "hacker cosplay" UI: neon green terminals, matrix rain backgrounds, glitch text, faux-CLI decorations, animated scanlines, pulsing threat rings, and gratuitous skull icons. That aesthetic screams "a non-pentester designed this for a movie prop." Real operators need density, clarity, and speed — not theatrics.

This file is your guide to break both patterns. Everything listed below is what Codex UI and fake-hacker UI do by default. Your job is to recognize these patterns, avoid them completely, and build an interface that feels like a real operator tool — functional, dense, and honest.

When you read this document, you're learning what NOT to do. The banned patterns are your red flags. The normal implementations are your blueprint. Follow them strictly, and you'll create UI that feels like Burp Suite, Nuclei's output, Shodan, or a well-built internal red team tool — not another generic AI dashboard or Hollywood hacking screen.

This is how you Uncodixify a pentesting recon triage dashboard.

---

## Keep It Normal (Uncodexy-UI Standard — Recon Triage)

- Sidebars: normal (240–260px fixed width, solid background, simple border-right, no floating shells, no rounded outer corners. Contains scope targets, scan profiles, and saved filters — not decorative nav.)
- Headers: normal (simple text, no eyebrows, no uppercase labels, no gradient text, just h1/h2 with proper hierarchy. Page title is the current target or scan session — nothing else.)
- Sections: normal (standard padding 20–30px, no hero blocks inside the dashboard, no decorative copy. Sections are functional: asset inventory, findings table, service map, port list.)
- Navigation: normal (simple links, subtle hover states, no transform animations, no badges unless functional. Tabs for switching between recon phases: discovery, enumeration, vulns, notes.)
- Buttons: normal (solid fills or simple borders, 8–10px radius max, no pill shapes, no gradient backgrounds. Actions are "Run Scan", "Export", "Mark Triaged", "Add to Scope" — plain labels.)
- Cards: normal (simple containers, 8–12px radius max, subtle borders, no shadows over 8px blur, no floating effect. Used for grouping a host's open ports, services, and findings — not for KPI vanity.)
- Forms: normal (standard inputs, clear labels above fields, no fancy floating labels, simple focus states. Inputs for target CIDR, wordlist path, thread count, timeout — functional fields.)
- Inputs: normal (solid borders, simple focus ring, no animated underlines, no morphing shapes. Monospace font for IP addresses, domains, and hash values.)
- Modals: normal (centered overlay, simple backdrop, no slide-in animations, straightforward close button. Used for finding detail view and raw request/response inspection.)
- Dropdowns: normal (simple list, subtle shadow, no fancy animations, clear selected state. For selecting scan profiles, severity filters, tool output source.)
- Tables: normal (clean rows, simple borders, subtle hover, no zebra stripes unless needed, left-aligned text. This is the core component — findings, hosts, ports, URLs all live in tables. Dense rows, small font, high information per pixel.)
- Lists: normal (simple items, consistent spacing, no decorative bullets, clear hierarchy. Used for subdomain lists, directory bruteforce results, DNS records.)
- Tabs: normal (simple underline or border indicator, no pill backgrounds, no sliding animations. Tabs separate recon phases or tool outputs.)
- Badges: normal (small text, simple border or background, 6–8px radius, no glows, only when needed. Severity badges: critical/high/medium/low/info — flat colors, no gradients, no pulses.)
- Avatars: not used. This is a tool, not a social app. If team attribution is needed, use plain text initials.
- Icons: normal (simple shapes, consistent size 16–20px, no decorative icon backgrounds, monochrome or subtle color. Icons for copy-to-clipboard, external link, expand row — utility only.)
- Typography: normal (monospace for technical data like IPs, ports, hashes, URLs, headers. Sans-serif for labels and UI chrome. Clear hierarchy, readable sizes 13–15px body. No mixed serif/sans combos.)
- Spacing: normal (consistent scale 4/8/12/16/24/32px, no random gaps, no excessive padding. Tight spacing preferred — operators want density, not whitespace.)
- Borders: normal (1px solid, subtle colors, no thick decorative borders, no gradient borders.)
- Shadows: normal (subtle 0 2px 8px rgba(0,0,0,0.1) max, no dramatic drop shadows, no colored shadows.)
- Transitions: normal (100–200ms ease, no bouncy animations, no transform effects, simple opacity/color changes.)
- Layouts: normal (standard grid/flex, no creative asymmetry, predictable structure, clear content hierarchy. Sidebar + main content area + optional detail panel. That's it.)
- Grids: normal (consistent columns, standard gaps, no creative overlaps, responsive breakpoints.)
- Flexbox: normal (simple alignment, standard gaps, no creative justify tricks.)
- Containers: normal (max-width 1400px or full-width for data-heavy views, standard padding, no creative widths.)
- Wrappers: normal (simple containing divs, no decorative purposes, functional only.)
- Panels: normal (simple background differentiation, subtle borders, no floating detached panels, no glass effects. Detail panels show raw HTTP, certificate info, screenshot previews — plain containers.)
- Toolbars: normal (simple horizontal layout, standard height 48–56px, clear actions, no decorative elements. Contains scan controls, filter bar, export button, bulk actions.)
- Footers: minimal or absent. Scan status bar at the bottom if needed — plain text showing "Scanning 142/500 hosts • 23 findings • 4m elapsed".
- Breadcrumbs: normal (simple text with separators, no fancy styling. Shows: Scope > target.com > 192.168.1.0/24 > host detail.)
- Status indicators: simple colored dot (4–6px) next to text or a flat background color on the table row. No animated pulses, no glowing rings, no ::before pseudo-element tricks.
- Severity colors: flat and muted. Critical: muted red. High: muted orange. Medium: muted yellow. Low: muted blue. Info: muted gray. No neon, no glow, no saturation above 60%.
- Code/output blocks: monospace, dark background, no syntax highlighting unless genuinely useful. Simple pre/code blocks for raw output, HTTP responses, headers, payloads.

Think Burp Suite. Think Shodan. Think Nuclei terminal output. Think a senior pentester's internal tool. They don't decorate. They show data. The UI gets out of the way and lets the operator triage fast.

---

## Hard No

- Everything you are used to doing and is a basic "YES" to you.
- No oversized rounded corners.
- No pill overload.
- No floating glassmorphism shells as the default visual language.
- No soft corporate gradients used to fake taste.
- No generic dark SaaS UI composition.
- No decorative sidebar blobs.
- No "control room" cosplay unless explicitly requested.
- No serif headline + system sans fallback combo as a shortcut to "premium."
- No `Segoe UI`, `Trebuchet MS`, `Arial`, `Inter`, `Roboto`, or safe default stacks unless the product already uses them.
- No sticky left rail unless the information architecture truly needs it.
- No metric-card grid as the first instinct.
- No fake charts that exist only to fill space.
- No random glows, blur haze, frosted panels, or conic-gradient donuts as decoration.
- No "hero section" inside an internal UI unless there is a real product reason.
- No alignment that creates dead space just to look expensive.
- No overpadded layouts.
- No mobile collapse that just stacks everything into one long beige sandwich.
- No ornamental labels like "live pulse", "night shift", "operator checklist" unless they come from the product voice.
- No generic startup copy.
- No style decisions made because they are easy to generate.

### Pentesting-Specific Hard No

- No neon green text on black backgrounds. This is not The Matrix.
- No terminal-emulator CSS as the main UI wrapper. If you need a terminal, embed one — don't skin the whole dashboard as a fake CLI.
- No animated scanlines, CRT effects, or glitch text.
- No "THREAT DETECTED" banners with pulsing red borders.
- No world map with animated attack lines. If geolocation data is relevant, show it in a plain table column or a simple static map.
- No skull icons, crosshair cursors, or biohazard symbols as decoration.
- No radar/sonar sweep animations for scan progress. Use a plain progress bar or text percentage.
- No "hacking in progress" loading screens. Show a spinner or progress text.
- No gratuitous use of red. Red is for critical severity — nowhere else.
- No fake "decrypting" or "accessing mainframe" text animations.
- No hexagonal grid layouts pretending to be a "cyber operations center."
- No excessive use of the word "cyber" in any label.

- No Headlines of any sort

```html
<div class="headline">
  <small>Recon Command</small>
  <h2>Your attack surface at a glance.</h2>
  <p>
    Triage findings, track assets, and prioritize
    what matters across your engagements.
  </p>
</div>
```

This is not allowed.

- `<small>` headers are NOT allowed
- Big no to rounded `span`s
- Colors going towards blue — **NOPE, bad.** Dark muted colors are best.

- Anything in the structure of this card is a **BIG no**.

```html
<div class="team-note">
  <small>Focus</small>
  <strong>
    Stay on top of critical findings and unverified assets.
  </strong>
</div>
```

This one is **THE BIGGEST NO**.

---

## Specifically Banned (Based on Mistakes)

- Border radii in the 20px to 32px range across everything (uses 12px everywhere — too much)
- Repeating the same rounded rectangle on sidebar, cards, buttons, and panels
- Sidebar width around 280px with a brand block on top and nav links below (: 248px with brand block)
- Floating detached sidebar with rounded outer shell
- Canvas chart placed in a glass card with no product-specific reason
- Donut chart paired with hand-wavy percentages (if you show a vuln severity breakdown, use a plain stacked bar or just numbers in a table)
- UI cards using glows instead of hierarchy
- Mixed alignment logic where some content hugs the left edge and some content floats in center-ish blocks
- Overuse of muted gray-blue text that weakens contrast and clarity
- "Premium dark mode" that really means blue-black gradients plus cyan accents (has radial gradients in background)
- UI typography that feels like a template instead of a tool
- Eyebrow labels (: "MARCH SNAPSHOT" uppercase with letter-spacing)
- Hero sections inside dashboards (has full hero-strip with decorative copy)
- Decorative copy like "Your attack surface, simplified" as page headers
- Section notes and mini-notes everywhere explaining what the UI does
- Transform animations on hover (: translateX(2px) on nav links)
- Dramatic box shadows (: 0 24px 60px rgba(0,0,0,0.35))
- Status indicators with ::before pseudo-elements creating colored dots with animation
- Muted labels with uppercase + letter-spacing (overuses this pattern)
- Pipeline bars with gradient fills (: linear-gradient(90deg, var(--primary), #4fe0c0))
- KPI cards in a grid as the default dashboard layout (has 3-column kpi-grid)
- "Team focus" or "Recent activity" panels with decorative internal copy
- Tables with tag badges for every status (overuses .tag class)
- Workspace blocks in sidebar with call-to-action buttons
- Brand marks with gradient backgrounds (: linear-gradient(135deg, #2a2a2a, #171717))
- Nav badges showing counts or "Live" status (has nav-badge class)
- Quota/usage panels with progress bars (has three quota sections)
- Footer lines with meta information (: "Recon dashboard • dark mode • single-file HTML")
- Trend indicators with colored text (: trend-up, trend-flat classes)
- Rail panels on the right side with "Today" schedule (has full right rail)
- Multiple nested panel types (panel, panel-2, rail-panel, table-panel)

### Pentesting-Specific Bans

- Severity donut charts as the primary summary — use a simple count row: `Critical: 3 | High: 12 | Medium: 28 | Low: 45`
- Animated "scanning" indicators that loop forever — show real progress or hide when idle
- Fake live-updating numbers that tick up for visual effect
- World/network topology maps used as decoration with no interactive function
- "Dashboard score" or "security grade" widgets (A+ / B- / etc.) unless the product genuinely calculates one
- Port lists displayed as colorful tag clouds instead of sorted tables
- Screenshot galleries with lightbox animations — use inline thumbnails that expand on click, plain overlay
- Vulnerability cards with threat-level gradients (red-to-orange fade backgrounds)
- "Attack path" visualizations that are just decorative node graphs with no real data
- Hexagonal or circular stat displays for simple numbers

---

## Rule

If a UI choice feels like a default AI UI move, ban it and pick the harder, cleaner option.
If a UI choice feels like "hacker movie" decoration, ban it and pick the boring, functional option.

- Colors should stay calm, not fight.
- Data density matters more than visual impact. A pentester wants to see 50 rows of findings, not 6 cards with icons.
- Monospace where data is technical. Sans-serif where it's UI chrome. Never decorative fonts.
- Every pixel should serve triage speed. If a UI element doesn't help the operator decide "is this worth investigating?" — remove it.

- You are bad at picking colors follow this priority order when selecting colors:

1. **Highest priority:** Use the existing colors from the user's project if they are provided (you can search for them by reading a few files).
2. If the project does not provide a palette, **get inspired from one of the predefined palettes below**.
3. Do **not invent random color combinations** unless explicitly requested.

You do not have to always choose the first palette. Select one randomly when drawing inspiration.

---

## Severity Color Tokens

Use these flat, muted severity colors across the entire dashboard. No gradients, no glows, no neon.

| Severity | Background | Text | Border |
|----------|-----------|------|--------|
| Critical | `#3b1219` | `#f87171` | `#7f1d1d` |
| High | `#3b1e08` | `#fb923c` | `#7c2d12` |
| Medium | `#3b3508` | `#facc15` | `#713f12` |
| Low | `#0c2340` | `#60a5fa` | `#1e3a5f` |
| Info | `#1a1a2e` | `#94a3b8` | `#334155` |

---

## Dark Color Schemes

| Palette | Background | Surface | Primary | Secondary | Accent | Text |
|---------|-----------|---------|---------|-----------|--------|------|
| Void Space | `#0d1117` | `#161b22` | `#58a6ff` | `#79c0ff` | `#f78166` | `#c9d1d9` |
| Graphite Pro | `#18181b` | `#27272a` | `#a855f7` | `#ec4899` | `#14b8a6` | `#fafafa` |
| Obsidian Depth | `#0f0f0f` | `#1a1a1a` | `#00d4aa` | `#00a3cc` | `#ff6b9d` | `#f5f5f5` |
| Carbon Elegance | `#121212` | `#1e1e1e` | `#bb86fc` | `#03dac6` | `#cf6679` | `#e1e1e1` |
| Charcoal Studio | `#1c1c1e` | `#2c2c2e` | `#0a84ff` | `#5e5ce6` | `#ff375f` | `#f2f2f7` |
| Onyx Matrix | `#0e0e10` | `#1c1c21` | `#00ff9f` | `#00e0ff` | `#ff0080` | `#f0f0f0` |
| Twilight Mist | `#1a1625` | `#2d2438` | `#9d7cd8` | `#7aa2f7` | `#ff9e64` | `#dcd7e8` |
| Midnight Canvas | `#0a0e27` | `#151b3d` | `#6c8eff` | `#a78bfa` | `#f472b6` | `#e2e8f0` |
| Slate Noir | `#0f172a` | `#1e293b` | `#38bdf8` | `#818cf8` | `#fb923c` | `#f1f5f9` |
| Deep Ocean | `#001e3c` | `#0a2744` | `#4fc3f7` | `#29b6f6` | `#ffa726` | `#eceff1` |

**Preferred for recon dashboards:** Void Space, Graphite Pro, Obsidian Depth, Carbon Elegance. These are dark without being theatrical. They don't scream "hacker" — they just reduce eye strain during long triage sessions.

---

## Light Color Schemes

| Palette | Background | Surface | Primary | Secondary | Accent | Text |
|---------|-----------|---------|---------|-----------|--------|------|
| Cloud Canvas | `#fafafa` | `#ffffff` | `#2563eb` | `#7c3aed` | `#dc2626` | `#0f172a` |
| Pearl Minimal | `#f8f9fa` | `#ffffff` | `#0066cc` | `#6610f2` | `#ff6b35` | `#212529` |
| Ivory Studio | `#f5f5f4` | `#fafaf9` | `#0891b2` | `#06b6d4` | `#f59e0b` | `#1c1917` |
| Porcelain Clean | `#f9fafb` | `#ffffff` | `#4f46e5` | `#8b5cf6` | `#ec4899` | `#111827` |
| Alabaster Pure | `#fcfcfc` | `#ffffff` | `#1d4ed8` | `#2563eb` | `#dc2626` | `#1e293b` |
| Frost Bright | `#f1f5f9` | `#f8fafc` | `#0f766e` | `#14b8a6` | `#e11d48` | `#0f172a` |
| Linen Soft | `#fef7f0` | `#fffbf5` | `#d97706` | `#ea580c` | `#0284c7` | `#292524` |
| Cream Elegance | `#fefce8` | `#fefce8` | `#65a30d` | `#84cc16` | `#f97316` | `#365314` |
| Arctic Breeze | `#f0f9ff` | `#f8fafc` | `#0284c7` | `#0ea5e9` | `#f43f5e` | `#0c4a6e` |
| Sand Warm | `#faf8f5` | `#ffffff` | `#b45309` | `#d97706` | `#059669` | `#451a03` |

Light themes are acceptable for report-generation views or exported triage summaries. The active recon workspace should default to dark.

---

## Component Guidance — Recon Triage Specifics

### Findings Table (Core Component)
- Dense rows, 13–14px monospace for technical fields (IP, port, URL, hash)
- Columns: Severity | Host | Port | Service | Finding Title | Tool Source | Status (New/Triaged/FP/Verified)
- Row click expands inline detail — no modal unless showing raw HTTP request/response
- Sortable columns, filterable by severity/status/tool
- Bulk select with checkbox column for mass-triage actions
- No colored row backgrounds per severity — use a small colored dot or left-border accent only

### Asset Inventory
- Plain table: IP | Hostname | Ports (comma-separated) | Services | OS Guess | Last Seen
- No cards. No grid layout. Just rows.
- Click a row to see full port/service detail in a side panel (slide-in from right, plain background, no animation beyond simple appear)

### Scan Controls
- Toolbar at top: target input (monospace), profile dropdown, "Run" button, progress text
- No decorative scan animations. Show: "Running subfinder... 142 found" as plain text that updates
- Progress: simple bar or percentage text. No circular progress, no radar sweep

### Notes / Collaboration
- Plain textarea per finding or per host
- Markdown support is fine, rendered simply
- No rich text editor with toolbar unless explicitly requested
- Timestamp + operator name as plain text, not a styled "comment card"

### Scope Management
- Sidebar section or dedicated tab
- List of in-scope targets: plain text rows with delete/edit actions
- Add target: simple input + "Add" button
- No drag-and-drop, no kanban, no decorative grouping

---

## In Your Internal Reasoning

Act as if you don't see this document. List all the stuff you would normally do for a pentesting dashboard UI (neon accents, threat rings, animated maps, glowing severity badges, hexagonal layouts, terminal-skin wrappers, dramatic shadows, gradient severity bars, hero sections with "Attack Surface Overview" copy, oversized KPI cards with trend arrows). Acknowledge those instincts — AND DON'T DO ANY OF THEM. Make it follow Uncodixify.