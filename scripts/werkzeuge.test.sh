#!/usr/bin/env bash
#
# Prüfungen des Werkzeugschritts aus scripts/werkzeuge.sh
# (#992).
#
# ## Warum dieser Wächter wirklich baut
#
# Die Abnahme von #992 verlangt ausdrücklich, dass die beiden Programme im
# Bündel **laufen** und nicht nur dastehen. Ein Wächter, der eine Datei zählt,
# hielte die halbe Zusage — und die andere Hälfte wäre unsichtbar (§10).
#
# Er läuft gegen ein **erfundenes** Bündel in einem Wegwerfverzeichnis; gebaut
# werden die echten Programme, denn genau die sollen laufen.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKRIPT="$WURZEL/scripts/werkzeuge.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0

behauptung() { # behauptung <beschreibung> <erwarteter-exit> <kommando…>
  local beschreibung="$1" erwartet="$2"
  shift 2

  local ausgabe status
  ausgabe="$("$@" 2>&1)"
  status=$?

  if [ "$status" -eq "$erwartet" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$beschreibung" "$AUS"
  else
    printf '%s  ✗ %s — Exit %d, erwartet war %d%s\n' "$ROT" "$beschreibung" "$status" "$erwartet" "$AUS"
    printf '      %s\n' "$ausgabe"
    FEHLER=1
  fi
}

buendel() { # buendel <verzeichnis> — legt ein Bündelgerüst an
  mkdir -p "$1/Probe.app/Contents/MacOS"

  printf '%s\n' "$1/Probe.app"
}

ORT="$(mktemp -d)"
trap 'rm -rf "$ORT"' EXIT

APP="$(buendel "$ORT")"

behauptung "der Werkzeugschritt läuft" 0 "$SKRIPT" "$APP"

# **Mit eigenem Zuhause, und zwar für jedes Programm aus dem Bündel.** Ohne
# es öffnet `mse-mcp` den Bestand **dieses Rechners** und hebt ihn auf den
# Schemastand des Zweigs, der gerade geprüft wird. Gemessen am 2026-09-27:
# der echte Bestand stand an einem Vormittag über 0085, 0086 und 0087, je
# mit einer Sicherung von 127 MB daneben — und eine Pipeline wurde rot, weil
# ihr Zweig 0087 noch nicht kannte. Gebaut wird weiter mit dem echten
# Zuhause (Go braucht seine Zwischenspeicher); nur die Läufe nicht. Dieselbe
# Umlenkung steht im Schritt selbst (`werkzeuge.sh`).
HEIM="$ORT/heim"
mkdir -p "$HEIM"

# **Beide liegen da und laufen** — die Zusage aus #992.
behauptung "mse liegt im Bündel und antwortet" 0 env HOME="$HEIM" "$APP/Contents/MacOS/mse" hilfe

# **Mit geschlossener Eingabe.** Der MCP-Server liest stdio und endet, wenn
# sein Gegenüber auflegt; ohne die Umleitung wartete er auf ein Terminal, das
# in der Pipeline niemandem gehört — gesehen, nicht überlegt: der erste Anlauf
# lief zehn Minuten und wurde abgebrochen.
mcpLaeuft() { HOME="$HEIM" "$APP/Contents/MacOS/mse-mcp" </dev/null; }

behauptung "mse-mcp liegt im Bündel und endet sauber" 0 mcpLaeuft

# **Und der Beleg, dass das Umlenken greift**: der Bestand entsteht im
# eigenen Zuhause. Stünde er dort nicht, hätte der Lauf einen anderen
# geöffnet — und das wäre wieder der des Rechners.
behauptung "mse-mcp legt seinen Bestand im eigenen Zuhause an" 0 \
  test -f "$HEIM/Library/Application Support/MagicSetEditor4Mac/bestand.db"

# **Der Dienst wird nicht gestartet**, sondern nach seinen Argumenten gefragt:
# ein unmöglicher Port lässt ihn anlaufen und von selbst enden. Geprüft ist
# damit, dass die Datei ein lauffähiges Programm ist — nicht, dass sie bedient.
restLaeuft() {
  HOME="$HEIM" "$APP/Contents/MacOS/mse-rest" -port 99999 </dev/null >/dev/null 2>&1
  local lauf=$?
  case "$lauf" in
    126|127) return 1 ;;
    *) return 0 ;;
  esac
}

behauptung "mse-rest liegt im Bündel und läuft an" 0 restLaeuft

# **Mit denselben Schaltern wie das Programm daneben** — WART-02
# (#2173).
# Gemessen am 2026-10-07 (`go1.27.1`, darwin/arm64): ohne `-trimpath -ldflags
# "-s -w"` wog jeder Helfer rund 80 MB statt 53 MB, und jeder trug rund 3.300
# Pfade dieses Rechners mit aus — samt dem Namen dessen, der baut.
#
# **Geprüft am Ergebnis, nicht an der Bauinfo**: mit `-trimpath` schreibt Go
# die `-ldflags` gar nicht hinein. `-s` zeigt sich daran, dass `go tool nm`
# kein `main.main` mehr findet, `-w` am Fehlen der DWARF-Abschnitte.
gebautWieDasProgramm() { # gebautWieDasProgramm <programm>
  go version -m "$1" | grep -q $'build\t-trimpath=true' || return 1
  ! go tool nm "$1" 2>/dev/null | grep -q ' T main\.main$' || return 1
  ! otool -l "$1" | grep -qE '__z?debug_' || return 1
  ! strings "$1" | grep -qF "$WURZEL"
}

for werkzeug in mse mse-mcp mse-rest; do
  behauptung "$werkzeug ist ohne Symboltafel und ohne Pfade dieses Rechners gebaut" 0 \
    gebautWieDasProgramm "$APP/Contents/MacOS/$werkzeug"
done

# **Ohne Bündel wird abgewiesen**, statt irgendwohin zu bauen.
behauptung "ohne Argument wird abgewiesen" 1 "$SKRIPT"
behauptung "ein Verzeichnis ohne Contents/MacOS wird abgewiesen" 1 "$SKRIPT" "$ORT/gibtsnicht.app"

if [ "$FEHLER" -ne 0 ]; then
  exit 1
fi
