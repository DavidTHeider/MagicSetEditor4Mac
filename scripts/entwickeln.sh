#!/usr/bin/env bash
#
# Die Anwendung zum Entwickeln — ENG-02 (#1727), der Ersatz für `wails dev`.
#
#     ./scripts/entwickeln.sh           # das Fenster, die Oberfläche von Vite (make dev)
#     ./scripts/entwickeln.sh browser   # dieselbe Anwendung im Browser (make dev-browser)
#
# ## Was `wails dev` tat, und wie es hier steht
#
# Unter Wails 2 startete `wails dev` Vite, legte ein Bündel samt Saat, öffnete
# das Fenster **und** stellte dieselbe Anwendung mit echten Bindungen unter
# `localhost:34115` bereit. Wails 3 trennt das: ein Programm hat ein Fenster
# **oder** ist ein HTTP-Server, je nach Bauschalter (`server`, Vermerk 0132).
# Die Sichtprüfung im Browser ist deshalb ein eigener Aufruf — unter
# **derselben** Adresse wie vorher, damit `docs/SICHTPRUEFUNG.md` und
# `.claude/launch.json` weiter stimmen.
#
# **Beide Arten nehmen die Oberfläche von Vite** (`FRONTEND_DEVSERVER_URL`):
# eine Änderung am Frontend ist nach dem Neuladen da, ohne neu zu bauen. Eine
# Änderung an Go braucht einen neuen Aufruf.
#
# ## Was gleich bleibt
#
# `MSE_BESTAND`, `MSE_SCHLUESSELBUND` und `HOME` gehen unverändert an das
# Programm — die Startkonfigurationen für einen eigenen Bestand tragen weiter.

set -euo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$WURZEL"

ART="${1:-fenster}"
case "$ART" in
  fenster) BAU="entwicklung" ;;
  browser) BAU="browser" ;;
  *) printf 'unbekannte Art: %s — fenster oder browser\n' "$ART" >&2; exit 2 ;;
esac

# **Vite auf einem eigenen Platz** und nicht auf 5173: dort läuft die
# Startkonfiguration `frontend`, und `--strictPort` soll
# einen belegten Platz melden statt still auszuweichen — sonst zeigte das
# Programm die Oberfläche eines anderen Laufs.
VITE_PLATZ="${MSE_VITE_PLATZ:-34116}"
# shellcheck source=scripts/entwickelnvite.sh
source "$WURZEL/scripts/entwickelnvite.sh"
vite_starten "$VITE_PLATZ"
PROGRAMM_PID=""

# **Beendet wird beides, und zwar sofort** — ein Messlauf, der dieses Skript
# beendet (`scripts/nahtlauf.mjs`), soll nicht ein Programm zurücklassen, das
# den Platz 34115 weiter belegt. Deshalb läuft auch das Programm im
# Hintergrund und das Skript wartet mit `wait`: auf einen Vordergrundprozess
# wartend, führte die Shell die Falle erst nach dessen Ende aus.
aufraeumen() {
  [ -n "$PROGRAMM_PID" ] && kill "$PROGRAMM_PID" 2>/dev/null
  kill "$VITE_PID" 2>/dev/null
  return 0
}
trap aufraeumen EXIT
trap 'exit 130' INT TERM

./scripts/bauen.sh "$BAU"

export FRONTEND_DEVSERVER_URL="$VITE"
if [ "$ART" = browser ]; then
  export WAILS_SERVER_PORT="${WAILS_SERVER_PORT:-34115}"
  printf '  → http://localhost:%s\n' "$WAILS_SERVER_PORT"
fi

build/bin/MagicSetEditor4Mac.app/Contents/MacOS/MagicSetEditor4Mac &
PROGRAMM_PID=$!
wait "$PROGRAMM_PID"
