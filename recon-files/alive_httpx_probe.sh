#!/bin/bash

# ─────────────────────────────────────────────
#  alive_httpx_probe.sh
#  Input:  probe/port-scan/ports.json
#  Output: probe/httpx/
# ─────────────────────────────────────────────

set -euo pipefail

subs_dir="subdomains"
probe_dir="probe"
httpx_dir="probe/httpx"
portscan_dir="probe/port-scan"

RED='\e[31m'; GREEN='\e[32m'; YELLOW='\e[33m'; BLUE='\e[34m'
BOLD="\e[1m"; ENDCOLOR='\e[0m'

mkdir -p "$httpx_dir"

# ─────────────────────────────────────────────
check_tools() {
    for tool in httpx curl jq python3; do
        if ! command -v "$tool" &>/dev/null; then
            echo -e "${BOLD}${RED}[!]${ENDCOLOR} $tool is not installed!"
            exit 1
        fi
    done
}

# ─────────────────────────────────────────────
# Build domain:port target list from ports.json + CDN domains fallback
build_targets() {
    echo -e "${BOLD}${BLUE}[+]${ENDCOLOR} Building targets from port scan results..."

    python3 - "$portscan_dir/ports.json" "$portscan_dir/domain_ips.json" <<'EOF' > "$httpx_dir/targets.txt"
import sys, json

ports_file      = sys.argv[1]
domain_ips_file = sys.argv[2]

CDN_DEFAULT_PORTS = [80, 443, 8080, 8443, 8000, 8888, 3000, 9090, 4443, 8081, 9443, 2087, 2083, 9200]

# domains with known open ports from scan
targets = set()
scanned_domains = set()
for entry in json.load(open(ports_file)):
    for domain in entry.get("domains", []):
        for port in entry.get("ports", []):
            targets.add(f"{domain}:{port}")
        scanned_domains.add(domain)

# CDN domains — not scanned, probe default ports
for entry in json.load(open(domain_ips_file)):
    if entry.get("cdn"):
        domain = entry["domain"]
        if domain not in scanned_domains:
            for port in CDN_DEFAULT_PORTS:
                targets.add(f"{domain}:{port}")

for t in sorted(targets):
    print(t)
EOF

    echo -e "${BOLD}${GREEN}[*]${ENDCOLOR} Targets crafted: ${BOLD}$(wc -l < "$httpx_dir/targets.txt")${ENDCOLOR} domain:port pairs\n"
}

# ─────────────────────────────────────────────
httpx_enrich() {
    echo -e "${BOLD}${BLUE}[+]${ENDCOLOR} Probing and enriching targets..."
    httpx -silent -follow-redirects \
        -l "$httpx_dir/targets.txt" \
        -status-code \
        -title \
        -tech-detect \
        -content-length \
        -web-server \
        -ip \
        -cname \
        -json \
        -o "$httpx_dir/httpx_enriched.json" > /dev/null 2>&1
    echo -e "${BOLD}${GREEN}[*]${ENDCOLOR} Enrichment complete: ${BOLD}$(wc -l < "$httpx_dir/httpx_enriched.json") hosts${ENDCOLOR}\n"
}

# ─────────────────────────────────────────────
# Merge port data into enriched JSON
merge_ports() {
    echo -e "${BOLD}${BLUE}[+]${ENDCOLOR} Merging port data into enriched results..."

    python3 - "$httpx_dir/httpx_enriched.json" "$portscan_dir/ports.json" <<'EOF' > "$httpx_dir/httpx_enriched_merged.json"
import sys, json, re
from urllib.parse import urlparse

enriched_file = sys.argv[1]
ports_file    = sys.argv[2]

# build domain -> ports mapping from ports.json
domain_ports = {}
for entry in json.load(open(ports_file)):
    for domain in entry.get("domains", []):
        domain_ports.setdefault(domain, set()).update(entry.get("ports", []))

# enrich each httpx result with port list
results = []
for line in open(enriched_file):
    line = line.strip()
    if not line:
        continue
    try:
        host = json.loads(line)
    except json.JSONDecodeError:
        continue

    url = host.get("url", "")
    parsed = urlparse(url)
    domain = parsed.hostname or ""

    ports = sorted(domain_ports.get(domain, []))
    if ports:
        host["open_ports"] = ports

    results.append(host)

# write as JSONL (one per line, same as httpx output)
for r in results:
    print(json.dumps(r))
EOF

    # replace original with merged
    mv "$httpx_dir/httpx_enriched_merged.json" "$httpx_dir/httpx_enriched.json"
    echo -e "${BOLD}${GREEN}[*]${ENDCOLOR} Port data merged into enriched JSON\n"
}

# ─────────────────────────────────────────────
path_probe() {
    echo -e "${BOLD}${BLUE}[+]${ENDCOLOR} Probing juicy paths..."

    local paths=(
        /.git/config /.env /robots.txt /sitemap.xml /crossdomain.xml
        /clientaccesspolicy.xml /.well-known/security.txt
        /api/swagger /api/swagger.json /api/openapi.json /v1/swagger
        /actuator /actuator/env /actuator/mappings
        /phpinfo.php /server-status /server-info
        /wp-admin /wp-config.php.bak /.DS_Store
    )

    # generate full url+path list
    while IFS= read -r base_url; do
        for path in "${paths[@]}"; do
            echo "${base_url}${path}"
        done
    done < <(jq -r '.url' "$httpx_dir/httpx_enriched.json") > "$httpx_dir/path_targets.txt"

    # probe with httpx, output json
    httpx -silent \
        -l "$httpx_dir/path_targets.txt" \
        -mc 200,201,301,302,403 \
        -status-code \
        -content-length \
        -json \
        -o "$httpx_dir/path_hits_raw.json" > /dev/null 2>&1

    # convert json to pipe-delimited url|status|size
    python3 - "$httpx_dir/path_hits_raw.json" <<'EOF' > "$httpx_dir/path_hits.txt"
import sys, json
for line in open(sys.argv[1]):
    line = line.strip()
    if not line:
        continue
    try:
        h = json.loads(line)
        size = h.get("content_length") or 0
        if size > 0:
            print(f"{h['url']}|{h['status_code']}|{size}")
    except (json.JSONDecodeError, KeyError):
        pass
EOF

    echo -e "${BOLD}${GREEN}[*]${ENDCOLOR} Path hits found: ${BOLD}$(wc -l < "$httpx_dir/path_hits.txt")${ENDCOLOR}\n"
}

# ─────────────────────────────────────────────
main() {
    check_tools
    build_targets
    httpx_enrich
    merge_ports
    path_probe
    echo -e "${BOLD}${GREEN}[*]${ENDCOLOR} Enriched JSON: ${BOLD}$httpx_dir/httpx_enriched.json${ENDCOLOR}"
    echo -e "${BOLD}${GREEN}[*]${ENDCOLOR} Path hits:     ${BOLD}$httpx_dir/path_hits.txt${ENDCOLOR}\n"
}

main
