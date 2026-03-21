#!/usr/bin/env bash
# deploy_memory.sh — copies Claude memory files to the correct location for this project.
#
# Usage:
#   bash deploy_memory.sh                        # assumes recon dir is ~/recon
#   bash deploy_memory.sh /path/to/your/recon    # explicit path
#
set -euo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info() { echo -e "${GREEN}[+]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
die()  { echo -e "${RED}[✗]${NC} $*" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MEMORY_SRC="$SCRIPT_DIR/memory"

[[ -d "$MEMORY_SRC" ]] || die "memory/ directory not found next to this script"

# ── Determine recon project path ──
RECON_DIR="${1:-$HOME/recon}"
RECON_DIR="$(realpath "$RECON_DIR")" 2>/dev/null || die "Path not found: $RECON_DIR"
[[ -d "$RECON_DIR" ]] || die "Recon directory does not exist: $RECON_DIR"

# ── Compute Claude's encoded project path ──
# Claude encodes the project dir by replacing every / with -
# e.g. /home/user/recon → -home-user-recon
ENCODED="${RECON_DIR//\//-}"
MEMORY_DEST="$HOME/.claude/projects/${ENCODED}/memory"

info "Recon dir:   $RECON_DIR"
info "Memory dest: $MEMORY_DEST"

mkdir -p "$MEMORY_DEST"
cp "$MEMORY_SRC"/*.md "$MEMORY_DEST/"

info "Deployed $(ls "$MEMORY_SRC"/*.md | wc -l) memory file(s):"
for f in "$MEMORY_SRC"/*.md; do
  echo "    $(basename "$f")"
done

echo ""
echo -e "${GREEN}Done.${NC} Claude will pick these up automatically on next session in $RECON_DIR"
