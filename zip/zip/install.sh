#!/usr/bin/env bash
# install.sh — sets up all recon toolkit dependencies on Fedora
set -euo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; BOLD='\033[1m'; NC='\033[0m'
info()  { echo -e "${GREEN}[+]${NC} $*"; }
warn()  { echo -e "${YELLOW}[!]${NC} $*"; }
die()   { echo -e "${RED}[✗]${NC} $*" >&2; exit 1; }
title() { echo -e "\n${BOLD}── $* ──${NC}"; }

[[ $EUID -eq 0 ]] || die "Run as root: sudo bash install.sh"

# ─────────────────────────────────────────
# Go version to install
# ─────────────────────────────────────────
GO_VERSION="1.22.4"
GOROOT="/usr/local/go"
GOPATH="$HOME/go"

# ─────────────────────────────────────────
# System packages
# ─────────────────────────────────────────
title "System packages"
dnf install -y \
  git curl wget jq whois unzip zip tar \
  gcc gcc-c++ make \
  sqlite sqlite-devel \
  libpcap libpcap-devel \
  nmap \
  grepcidr \
  python3 \
  chromium \
  2>/dev/null || warn "Some packages may have failed — check above"

# ─────────────────────────────────────────
# Go
# ─────────────────────────────────────────
title "Go ${GO_VERSION}"
INSTALLED_GO=""
command -v go &>/dev/null && INSTALLED_GO=$(go version 2>/dev/null | awk '{print $3}' | tr -d 'go')

if [[ "$INSTALLED_GO" == "$GO_VERSION" ]]; then
  info "Go ${GO_VERSION} already installed — skipping"
else
  info "Downloading Go ${GO_VERSION}..."
  wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
  rm -rf "$GOROOT"
  tar -C /usr/local -xzf /tmp/go.tar.gz
  rm /tmp/go.tar.gz

  # Profile entry (idempotent)
  PROFILE_LINE='export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin'
  if ! grep -qF "$PROFILE_LINE" /etc/profile.d/go.sh 2>/dev/null; then
    echo "$PROFILE_LINE" > /etc/profile.d/go.sh
    chmod 644 /etc/profile.d/go.sh
  fi
  info "Go installed — profile written to /etc/profile.d/go.sh"
fi

export PATH=$PATH:${GOROOT}/bin:${GOPATH}/bin

# ─────────────────────────────────────────
# masscan (build from source — often not in Fedora repos)
# ─────────────────────────────────────────
title "masscan"
if command -v masscan &>/dev/null; then
  info "masscan already installed — skipping"
else
  info "Building masscan from source..."
  git clone --depth 1 https://github.com/robertdavidgraham/masscan /tmp/masscan-src
  make -C /tmp/masscan-src -j"$(nproc)"
  install -m 755 /tmp/masscan-src/bin/masscan /usr/local/bin/masscan
  rm -rf /tmp/masscan-src
  info "masscan installed to /usr/local/bin/masscan"
fi

# ─────────────────────────────────────────
# Go-based recon tools
# ─────────────────────────────────────────
title "Go recon tools"

declare -A GO_TOOLS=(
  [subfinder]="github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest"
  [httpx]="github.com/projectdiscovery/httpx/cmd/httpx@latest"
  [dnsx]="github.com/projectdiscovery/dnsx/cmd/dnsx@latest"
  [alterx]="github.com/projectdiscovery/alterx/cmd/alterx@latest"
  [puredns]="github.com/d3mondev/puredns/v2@latest"
  [katana]="github.com/projectdiscovery/katana/cmd/katana@latest"
  [github-subdomains]="github.com/gwen001/github-subdomains@latest"
  [gau]="github.com/lc/gau/v2/cmd/gau@latest"
  [waybackurls]="github.com/tomnomnom/waybackurls@latest"
  [gowitness]="github.com/sensepost/gowitness@latest"
)

for name in "${!GO_TOOLS[@]}"; do
  if command -v "$name" &>/dev/null; then
    info "  $name — already installed, skipping"
  else
    info "  installing $name..."
    go install "${GO_TOOLS[$name]}" 2>/dev/null \
      && info "  $name — ok" \
      || warn "  $name — FAILED (check manually)"
  fi
done

# ─────────────────────────────────────────
# SecLists
# ─────────────────────────────────────────
title "SecLists"
SECLISTS_DIR="/usr/share/seclists"
if [[ -d "$SECLISTS_DIR" ]]; then
  info "SecLists already present at $SECLISTS_DIR — skipping"
else
  info "Cloning SecLists (this is large, ~1GB)..."
  git clone --depth 1 https://github.com/danielmiessler/SecLists "$SECLISTS_DIR"
fi

# ─────────────────────────────────────────
# Done
# ─────────────────────────────────────────
title "Done"
echo ""
echo -e "  ${GREEN}All tools installed.${NC}"
echo -e "  Run the following or open a new shell:"
echo -e "    ${BOLD}source /etc/profile.d/go.sh${NC}"
echo ""
echo -e "  Remember to set your GitHub token before running recon:"
echo -e "    ${BOLD}export GITHUB_TOKEN=your_token_here${NC}"
echo ""
