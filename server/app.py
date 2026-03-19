#!/usr/bin/env python3

from flask import Flask, jsonify
from pathlib import Path
import json
import re
from collections import defaultdict
from urllib.parse import urlparse

app = Flask(__name__, static_folder="static", static_url_path="")

BASE_DIR    = Path(__file__).parent.parent
PROBE_DIR   = BASE_DIR / "probe"
HTTPX_DIR   = PROBE_DIR / "httpx"
ENRICHED    = HTTPX_DIR / "httpx_enriched.json"
PATH_HITS   = HTTPX_DIR / "path_hits.txt"
JS_DIR      = PROBE_DIR / "js-analysis"

PORT_SERVICES = {
    21: "FTP", 22: "SSH", 23: "Telnet", 25: "SMTP", 53: "DNS",
    80: "HTTP", 110: "POP3", 111: "RPC", 135: "MSRPC", 139: "NetBIOS",
    143: "IMAP", 161: "SNMP", 389: "LDAP", 443: "HTTPS", 445: "SMB",
    465: "SMTPS", 587: "SMTP", 636: "LDAPS", 993: "IMAPS", 995: "POP3S",
    1433: "MSSQL", 1521: "Oracle", 2049: "NFS", 2083: "cPanel SSL",
    2087: "WHM SSL", 3000: "Dev/Node", 3306: "MySQL", 3389: "RDP",
    4443: "HTTPS-alt", 5432: "PostgreSQL", 5900: "VNC", 5984: "CouchDB",
    6379: "Redis", 7443: "HTTPS-alt", 8000: "HTTP-alt", 8080: "HTTP-alt",
    8081: "HTTP-alt", 8443: "HTTPS-alt", 8888: "HTTP-alt", 9090: "HTTP-alt",
    9200: "Elasticsearch", 9300: "Elasticsearch", 9443: "HTTPS-alt",
    27017: "MongoDB",
}

INTERESTING_RE = re.compile(
    r'login|admin|dashboard|portal|jenkins|grafana|kibana|gitlab|jira|confluence|phpmyadmin|cpanel|wp-admin',
    re.IGNORECASE
)
API_RE   = re.compile(r'api|swagger|openapi|graphql', re.IGNORECASE)
HIGH_RE  = re.compile(r'\.env|\.git|wp-config|phpinfo|server-status|actuator/env', re.IGNORECASE)
MED_RE   = re.compile(r'actuator|swagger|openapi|server-info', re.IGNORECASE)


def port_of(h):
    p = urlparse(h.get("url", ""))
    if p.port:
        return p.port
    return 443 if p.scheme == "https" else 80

def pick_primary(group):
    for p in [443, 80]:
        for h in group:
            if port_of(h) == p:
                return h
    return sorted(group, key=port_of)[0]

def host_badges(url, title):
    badges = []
    combined = f"{url} {title or ''}"
    if INTERESTING_RE.search(combined):
        badges.append("interesting")
    if API_RE.search(combined):
        badges.append("api")
    return badges

def hit_severity(url):
    if HIGH_RE.search(url):   return "high"
    if MED_RE.search(url):    return "medium"
    return "info"

def status_class(status):
    s = str(status)
    if s.startswith("2"):  return f"s{s}"
    if s.startswith("3"):  return f"s{s[:3]}"
    if s == "403":         return "s403"
    if s.startswith("4"):  return "s400"
    return ""


# ── API routes ────────────────────────────────

@app.route("/api/hosts")
def api_hosts():
    if not ENRICHED.exists():
        return jsonify({"error": "httpx_enriched.json not found"}), 404

    raw = []
    with open(ENRICHED) as f:
        for line in f:
            line = line.strip()
            if line:
                try:
                    raw.append(json.loads(line))
                except json.JSONDecodeError:
                    pass

    # group by hostname
    groups = defaultdict(list)
    for h in raw:
        hostname = urlparse(h.get("url", "")).hostname or h.get("url", "-")
        groups[hostname].append(h)

    hosts = []
    for hostname, group in sorted(groups.items()):
        primary    = pick_primary(group)
        url        = primary.get("url", "-")
        status     = primary.get("status_code", "-")
        title      = primary.get("title") or "-"
        server     = primary.get("webserver") or "-"
        tech       = primary.get("tech") or []
        ips        = primary.get("a") or []
        cname      = primary.get("cname") or []
        ctype      = primary.get("content_type") or "-"
        open_ports = primary.get("open_ports") or []

        ports_detail = [{"port": p, "service": PORT_SERVICES.get(p, "Unknown")} for p in open_ports]
        sc           = status_class(status)
        badges       = host_badges(url, title)

        entry = {
            "url":      url,
            "status":   status,
            "sc":       sc,
            "title":    title,
            "server":   server,
            "tech":     tech,
            "ips":      ips,
            "cname":    cname,
            "ctype":    ctype,
            "ports":    ports_detail,
            "badges":   badges,
        }

        # child ports
        if len(group) > 1:
            children = []
            for child in sorted(group, key=port_of):
                if child is primary:
                    continue
                c_ports = [{"port": p, "service": PORT_SERVICES.get(p, "Unknown")}
                           for p in (child.get("open_ports") or [])]
                children.append({
                    "url":    child.get("url", "-"),
                    "status": child.get("status_code", "-"),
                    "sc":     status_class(child.get("status_code", "-")),
                    "title":  child.get("title") or "-",
                    "server": child.get("webserver") or "-",
                    "tech":   child.get("tech") or [],
                    "ips":    child.get("a") or [],
                    "cname":  child.get("cname") or [],
                    "ctype":  child.get("content_type") or "-",
                    "ports":  c_ports,
                    "badges": [],
                })
            entry["children"] = children

        hosts.append(entry)

    total = len(hosts)
    s200  = sum(1 for h in raw if h.get("status_code") == 200)
    s403  = sum(1 for h in raw if h.get("status_code") == 403)
    s500  = sum(1 for h in raw if isinstance(h.get("status_code"), int) and h["status_code"] >= 500)

    return jsonify({
        "stats": {"total": total, "s200": s200, "s403": s403, "s500": s500},
        "hosts": hosts,
    })


@app.route("/api/hits")
def api_hits():
    if not PATH_HITS.exists():
        return jsonify({"hits": []})

    hits = []
    with open(PATH_HITS) as f:
        for line in f:
            parts = line.strip().split("|")
            if len(parts) == 3:
                hits.append({
                    "url":      parts[0],
                    "status":   parts[1],
                    "sc":       status_class(parts[1]),
                    "size":     parts[2],
                    "severity": hit_severity(parts[0]),
                })

    return jsonify({"hits": hits})


@app.route("/api/js")
def api_js():
    if not JS_DIR.exists():
        return jsonify({"domains": []})

    domains = []
    for entry in sorted(JS_DIR.iterdir()):
        if not entry.is_dir():
            continue

        endpoints_file = entry / "endpoints.json"
        secrets_file   = entry / "secrets.json"

        endpoint_count = 0
        secret_count   = 0
        severity       = "none"

        if endpoints_file.exists():
            try:
                endpoint_count = len(json.loads(endpoints_file.read_text()))
            except (json.JSONDecodeError, OSError):
                pass

        if secrets_file.exists():
            try:
                secrets = json.loads(secrets_file.read_text())
                secret_count = len(secrets)
                if any(s.get("severity") == "high" for s in secrets):
                    severity = "high"
                elif any(s.get("severity") == "medium" for s in secrets):
                    severity = "medium"
                elif secret_count > 0:
                    severity = "low"
            except (json.JSONDecodeError, OSError):
                pass

        domains.append({
            "domain":         entry.name,
            "endpoint_count": endpoint_count,
            "secret_count":   secret_count,
            "severity":       severity,
        })

    return jsonify({"domains": domains})


@app.route("/api/js/<domain>")
def api_js_domain(domain):
    domain_dir = JS_DIR / domain
    if not domain_dir.exists():
        return jsonify({"error": "domain not found"}), 404

    endpoints = []
    secrets   = []

    endpoints_file = domain_dir / "endpoints.json"
    secrets_file   = domain_dir / "secrets.json"

    if endpoints_file.exists():
        try:
            endpoints = json.loads(endpoints_file.read_text())
        except (json.JSONDecodeError, OSError):
            pass

    if secrets_file.exists():
        try:
            secrets = json.loads(secrets_file.read_text())
        except (json.JSONDecodeError, OSError):
            pass

    return jsonify({"endpoints": endpoints, "secrets": secrets})


@app.route("/")
def index():
    return app.send_static_file("index.html")


if __name__ == "__main__":
    print("\033[1m\033[34m[+]\033[0m Starting recon dashboard at \033[1mhttp://localhost:8000\033[0m")
    app.run(host="0.0.0.0", port=8000, debug=False)
