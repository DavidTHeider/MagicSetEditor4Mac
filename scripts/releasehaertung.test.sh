#!/usr/bin/env bash
#
# Prüfungen von scripts/releasehaertung.sh — Nachlauf über Phase 5 von 0.9.0,
# Befund 4 (SEC03, #2356).
#
# **An einem nachgestellten Bündel aus Kopien von `/usr/bin/true`**, ad hoc
# gesiegelt mit und ohne `--options runtime` — gemessen am 2026-10-16: die
# Kopie trägt danach `flags=0x10002(adhoc,runtime)` beziehungsweise
# `flags=0x2(adhoc)`. Ein Programm unter `Contents/Helpers` ohne Laufzeit
# muss die Prüfung anhalten; bis zum Nachlauf las sie nur `Contents/MacOS`.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# shellcheck source=scripts/releasehaertung.sh
source "$WURZEL/scripts/releasehaertung.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0
RAUM="$(mktemp -d)"
trap 'rm -rf "$RAUM"' EXIT

behauptung() { # behauptung <beschreibung> <wahr?>
  if [ "$2" = ja ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"
  else
    printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"
    FEHLER=1
  fi
}

# programm <pfad> <mit|ohne> — eine gesiegelte Kopie von /usr/bin/true.
programm() {
  mkdir -p "$(dirname "$1")"
  cp /usr/bin/true "$1"
  if [ "$2" = mit ]; then
    codesign --force --options runtime --sign - "$1" 2>/dev/null
  else
    codesign --force --sign - "$1" 2>/dev/null
  fi
}

APP="$RAUM/Probe.app"
for name in Probe mse mse-mcp mse-rest; do programm "$APP/Contents/MacOS/$name" mit; done
printf 'kein Programm\n' > "$APP/Contents/MacOS/lies-mich.txt"
chmod +x "$APP/Contents/MacOS/lies-mich.txt"

ZAHL="$(programme_gehaertet "$APP")"; CODE=$?
behauptung "vier gehärtete Programme gehen durch — und eine ausführbare Textdatei zählt nicht" \
  "$([ "$CODE" -eq 0 ] && [ "$ZAHL" = 4 ] && echo ja || echo nein)"

programm "$APP/Contents/Helpers/helfer" ohne
FUND="$(programme_gehaertet "$APP")"; CODE=$?
behauptung "ein Helfer ohne Laufzeit unter Contents/Helpers hält an und wird genannt" \
  "$([ "$CODE" -eq 1 ] && [ "$FUND" = Contents/Helpers/helfer ] && echo ja || echo nein)"

programm "$APP/Contents/Helpers/helfer" mit
ZAHL="$(programme_gehaertet "$APP")"; CODE=$?
behauptung "gehärtet zählt er mit" \
  "$([ "$CODE" -eq 0 ] && [ "$ZAHL" = 5 ] && echo ja || echo nein)"

exit "$FEHLER"
