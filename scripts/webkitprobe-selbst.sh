#!/usr/bin/env bash
#
# **Die Selbstprobe der WebKit-Probe** —
# #1828.
#
# Fährt drei Drehbücher gegen eine eigene Probeseite
# (`scripts/webkitprobe/selbst/`): eines über den Motor `wk2`, eines über
# `wk1`, und eines, das rot sein muss. **Ein Prüfwerkzeug, das nie rot wird,
# prüft nichts** (§10, die Gegenprobe) — deshalb gehört das rote Drehbuch
# dazu, und ein kaputtes, das gar nicht erst anfangen darf.
#
# **Nicht Teil von `make ci`** — und deshalb heißt die Datei nicht
# `*.test.sh`, denn die fährt `scripts/ci.sh` alle: die Probe braucht eine
# Anmeldesitzung mit Fenstersystem, und eine Pipeline über SSH hätte keine.
# Aufruf: `make webkitprobe`.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SELBST="$WURZEL/scripts/webkitprobe/selbst"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0
RAUM="$(mktemp -d)"
DIENER=""

trap '[ -n "$DIENER" ] && kill "$DIENER" 2>/dev/null; rm -rf "$RAUM"' EXIT

gleich() { # gleich <beschreibung> <erwartet> <bekommen>
  if [ "$2" = "$3" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"
  else
    printf '%s  ✗ %s — erwartet „%s“, bekommen „%s“%s\n' "$ROT" "$1" "$2" "$3" "$AUS"
    FEHLER=1
  fi
}

enthaelt() { # enthaelt <beschreibung> <datei> <text>
  if grep -qF -- "$3" "$2"; then
    gleich "$1" "ja" "ja"
  else
    gleich "$1" "ja" "nein"
  fi
}

# **Ein freier Platz, den das Betriebssystem wählt** (Port 0) — ein fester
# wäre beim zweiten Lauf nebenher besetzt.
python3 -u -m http.server 0 --bind 127.0.0.1 --directory "$SELBST" > "$RAUM/diener.log" 2>&1 &
DIENER=$!
# **Abgekoppelt**, damit die Shell beim Aufräumen nicht „Terminated" über
# einen Diener ruft, dessen Ende gewollt ist.
disown "$DIENER"
for _ in $(seq 50); do
  PLATZ="$(sed -n 's/.*port \([0-9]*\).*/\1/p' "$RAUM/diener.log" | head -1)"
  [ -n "$PLATZ" ] && break
  sleep 0.1
done
[ -n "${PLATZ:-}" ] || { printf '%s  ✗ die Probeseite ließ sich nicht ausliefern%s\n' "$ROT" "$AUS"; exit 1; }
BASIS="http://127.0.0.1:$PLATZ/"

printf '\n── Motor wk2: Ereignisse, WebGL, Vorschlagsliste, Erscheinungsbild\n'
"$WURZEL/scripts/webkitprobe.sh" "$SELBST/gruen-wk2.probe" --basis "$BASIS" --ausgabe "$RAUM/wk2" > /dev/null 2>&1
gleich "das grüne Drehbuch über wk2 besteht" "0" "$?"
enthaelt "die Vorschlagsliste ist WebKits eigenes Fenster und nennt ihre Zeile" "$RAUM/wk2/bericht.md" "Zeile 1: ph — Phyrexianisch"
enthaelt "WebGL 2 steht zur Verfügung" "$RAUM/wk2/bericht.md" "✓ Z6 \`pruefe window.probeWebGL.startsWith('WebGL 2.0')\`"
[ -s "$RAUM/wk2/seite-dunkel.png" ]
gleich "der Abzug im dunklen Erscheinungsbild liegt vor" "0" "$?"

printf '\n── Motor wk1: was ein Vorleser sieht\n'
"$WURZEL/scripts/webkitprobe.sh" "$SELBST/gruen-wk1.probe" --basis "$BASIS" --ausgabe "$RAUM/wk1" > /dev/null 2>&1
gleich "das grüne Drehbuch über wk1 besteht" "0" "$?"
enthaelt "der Dialog steht mit Namen und Rolle im Baum" "$RAUM/wk1/bericht.md" "AXGroup «Befehle» · web dialog"
enthaelt "die Ansage nennt das fokussierte Feld" "$RAUM/wk1/bericht.md" "Ansage: AXTextField «Suchen» · text field · fokussiert"

printf '\n── Die Gegenproben\n'
"$WURZEL/scripts/webkitprobe.sh" "$SELBST/rot.probe" --basis "$BASIS" --ausgabe "$RAUM/rot" > /dev/null 2>&1
gleich "ein roter Schritt macht den Lauf rot" "1" "$?"
enthaelt "und der Bericht nennt ihn mit seiner Zeile" "$RAUM/rot/bericht.md" "✗ Z4 \`warte #gibtsnicht\`"
enthaelt "und hält dort an" "$RAUM/rot/bericht.md" "**Ergebnis: rot.**"

printf 'lade seite.html\nschwebe #oeffnen\n' > "$RAUM/kaputt.probe"
"$WURZEL/scripts/webkitprobe.sh" "$RAUM/kaputt.probe" --basis "$BASIS" --ausgabe "$RAUM/kaputt" > "$RAUM/kaputt.log" 2>&1
gleich "ein unbekannter Schritt hält das Drehbuch vor dem ersten Schritt an" "2" "$?"
enthaelt "und nennt Zeile und Wort" "$RAUM/kaputt.log" "Zeile 2: unbekannter Schritt »schwebe«"

printf '\n'
if [ "$FEHLER" -eq 0 ]; then
  printf '%s✓ WebKit-Probe in Ordnung%s\n' "$GRUEN" "$AUS"
else
  printf '%s✗ WebKit-Probe fehlerhaft%s\n' "$ROT" "$AUS"
fi

exit "$FEHLER"
