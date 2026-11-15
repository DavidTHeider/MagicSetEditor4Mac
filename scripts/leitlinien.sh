#!/usr/bin/env bash
#
# Schreibt CLAUDE.md aus AGENTS.md (AGENTS.md §9).
#
# Die beiden Dateien sind gleichlautend, weil Werkzeuge das eine oder das andere
# lesen und keines dabei eine andere Regel bekommen soll. AGENTS.md ist die
# Quelle; CLAUDE.md entsteht daraus. Dass es so bleibt, prüft die Stufe `format`.

set -euo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$WURZEL"

if [ ! -f AGENTS.md ]; then
  echo "leitlinien: AGENTS.md fehlt" >&2
  exit 1
fi

if cmp -s AGENTS.md CLAUDE.md 2>/dev/null; then
  echo "leitlinien: CLAUDE.md ist bereits gleichlautend"
  exit 0
fi

cp AGENTS.md CLAUDE.md
echo "leitlinien: CLAUDE.md aus AGENTS.md geschrieben"
