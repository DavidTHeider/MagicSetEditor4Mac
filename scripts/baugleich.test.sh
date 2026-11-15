#!/usr/bin/env bash
#
# Prüfungen des Bauvergleichs aus scripts/baugleich.sh — I06
# (#2264).
#
# Gegen **nachgestellte** Bündel: zwei echte Freigabebauten dauern zwanzig
# Minuten und gehören in den Beleg `docs/belege/baugleich.md`, nicht in jede
# Pipeline. Geprüft wird, dass jeder Unterschied entweder seinen Grund trägt
# oder den Lauf rot macht — und dass ein Grund nur dort steht, wo er stimmt.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# shellcheck source=scripts/baugleich.sh
source "$WURZEL/scripts/baugleich.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0
RAUM="$(mktemp -d)"
trap 'rm -rf "$RAUM"' EXIT

gleich() { # gleich <beschreibung> <erwartet> <bekommen>
  if [ "$2" = "$3" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"
  else
    printf '%s  ✗ %s — erwartet „%s“, bekommen „%s“%s\n' "$ROT" "$1" "$2" "$3" "$AUS"
    FEHLER=1
  fi
}

enthaelt() { # enthaelt <beschreibung> <teil> <text>
  case "$3" in
    *"$2"*) printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS" ;;
    *) printf '%s  ✗ %s — „%s“ fehlt in:\n%s%s\n' "$ROT" "$1" "$2" "$3" "$AUS"; FEHLER=1 ;;
  esac
}

# buendel <ort> <gebaut> — ein nachgestelltes Bündel mit Stand, Text und Link.
buendel() {
  mkdir -p "$1/Contents/MacOS" "$1/Contents/Resources/saat"
  printf '{"head":"abc","art":"freigabe","gebaut":"%s"}' "$2" > "$1/Contents/Resources/baustand.json"
  printf 'vorlage' > "$1/Contents/Resources/saat/eine vorlage.txt"
  ln -s /Applications "$1/Programme"
}

# --------------------------------------------------------------- gleich

buendel "$RAUM/a1" "2026-10-11T20:00:00Z"
buendel "$RAUM/b1" "2026-10-11T20:00:00Z"
AUS1="$(baugleich_vergleiche "$RAUM/a1" "$RAUM/b1")"; RC1=$?
gleich "zwei gleiche Bündel sind gleich" "0" "$RC1"
enthaelt "die Zahl der verglichenen Dateien steht da" "3 Einträge verglichen, 3 gleich" "$AUS1"

# ------------------------------------------------- der Zeitpunkt des Baus

buendel "$RAUM/a2" "2026-10-11T20:00:00Z"
buendel "$RAUM/b2" "2026-10-11T20:09:00Z"
AUS2="$(baugleich_vergleiche "$RAUM/a2" "$RAUM/b2")"; RC2=$?
gleich "ein anderer Bauzeitpunkt ist erklärt" "0" "$RC2"
enthaelt "und er steht mit Grund da" "| \`./Contents/Resources/baustand.json\` | der Zeitpunkt des Baus" "$AUS2"

# ------------------------------------------- aber nicht ein anderer Stand

buendel "$RAUM/a3" "2026-10-11T20:00:00Z"
buendel "$RAUM/b3" "2026-10-11T20:00:00Z"
printf '{"head":"def","art":"freigabe","gebaut":"2026-10-11T20:00:00Z"}' > "$RAUM/b3/Contents/Resources/baustand.json"
AUS3="$(baugleich_vergleiche "$RAUM/a3" "$RAUM/b3")"; RC3=$?
gleich "ein anderer Commit im Stand ist nicht erklärt" "1" "$RC3"
enthaelt "er steht als ungeklärt da" "| \`./Contents/Resources/baustand.json\` | **ungeklärt**" "$AUS3"

# ------------------------------------------------- eine Datei mehr, ein Link

buendel "$RAUM/a4" "2026-10-11T20:00:00Z"
buendel "$RAUM/b4" "2026-10-11T20:00:00Z"
printf 'neu' > "$RAUM/b4/Contents/Resources/saat/zweite.txt"
rm "$RAUM/b4/Programme" && ln -s /tmp "$RAUM/b4/Programme"
AUS4="$(baugleich_vergleiche "$RAUM/a4" "$RAUM/b4")"; RC4=$?
gleich "eine Datei nur in einem Bündel macht rot" "1" "$RC4"
enthaelt "sie steht mit ihrer Seite da" "| \`./Contents/Resources/saat/zweite.txt\` | **nur im zweiten Bau**" "$AUS4"
enthaelt "ein anderes Linkziel ist ungeklärt" "| \`./Programme\` | **ungeklärt**" "$AUS4"

# ------------------------------------------------------ nur das Siegel

if command -v clang >/dev/null 2>&1 && command -v codesign >/dev/null 2>&1; then
  buendel "$RAUM/a5" "2026-10-11T20:00:00Z"
  buendel "$RAUM/b5" "2026-10-11T20:00:00Z"
  printf 'int main(void){return 0;}\n' > "$RAUM/p.c"
  clang -o "$RAUM/p" "$RAUM/p.c" 2>/dev/null
  cp "$RAUM/p" "$RAUM/a5/Contents/MacOS/Probe"
  cp "$RAUM/p" "$RAUM/b5/Contents/MacOS/Probe"
  codesign --force --sign - -i probe.eins "$RAUM/a5/Contents/MacOS/Probe" 2>/dev/null
  codesign --force --sign - -i probe.zwei "$RAUM/b5/Contents/MacOS/Probe" 2>/dev/null
  AUS5="$(baugleich_vergleiche "$RAUM/a5" "$RAUM/b5")"; RC5=$?
  gleich "ein Programm, das sich nur im Siegel unterscheidet, ist erklärt" "0" "$RC5"
  enthaelt "es steht als Siegel da" "| \`./Contents/MacOS/Probe\` | nur das Siegel" "$AUS5"

  # **Die Gegenprobe**: ein Programm mit anderem Code ist nicht „nur das Siegel“.
  printf 'int main(void){return 1;}\n' > "$RAUM/q.c"
  clang -o "$RAUM/b5/Contents/MacOS/Probe" "$RAUM/q.c" 2>/dev/null
  codesign --force --sign - -i probe.zwei "$RAUM/b5/Contents/MacOS/Probe" 2>/dev/null
  AUS5B="$(baugleich_vergleiche "$RAUM/a5" "$RAUM/b5")"; RC5B=$?
  gleich "anderer Code ist ungeklärt" "1" "$RC5B"
  enthaelt "und steht so da" "| \`./Contents/MacOS/Probe\` | **ungeklärt**" "$AUS5B"
else
  printf '  ~ clang oder codesign fehlen — das Siegel ist nicht geprüft\n'
fi

# ------------------------------------------------- das Siegelverzeichnis

siegel() { # siegel <ort> <hash des Stands>
  mkdir -p "$1/Contents/_CodeSignature"
  python3 -I - "$1/Contents/_CodeSignature/CodeResources" "$2" <<'PY'
import plistlib, sys
d = {"files": {"Resources/baustand.json": sys.argv[2].encode()},
     "files2": {"Resources/baustand.json": {"hash2": sys.argv[2].encode()},
                "Resources/saat/eine vorlage.txt": {"hash2": b"gleich"}},
     "rules": {}, "rules2": {}}
plistlib.dump(d, open(sys.argv[1], "wb"))
PY
}

buendel "$RAUM/a6" "2026-10-11T20:00:00Z"
buendel "$RAUM/b6" "2026-10-11T20:09:00Z"
siegel "$RAUM/a6" "eins"
siegel "$RAUM/b6" "zwei"
AUS6="$(baugleich_vergleiche "$RAUM/a6" "$RAUM/b6")"; RC6=$?
gleich "ein Siegelverzeichnis, das nur erklärte Dateien anders hasht, ist erklärt" "0" "$RC6"
enthaelt "es nennt die Datei, aus der es folgt" "| \`./Contents/_CodeSignature/CodeResources\` | folgt aus: \`Resources/baustand.json\`" "$AUS6"

# **Die Gegenprobe**: hasht es eine Datei anders, die gleich ist, folgt es aus nichts.
buendel "$RAUM/a7" "2026-10-11T20:00:00Z"
buendel "$RAUM/b7" "2026-10-11T20:00:00Z"
siegel "$RAUM/a7" "eins"
siegel "$RAUM/b7" "zwei"
AUS7="$(baugleich_vergleiche "$RAUM/a7" "$RAUM/b7")"; RC7=$?
gleich "ein Siegelverzeichnis über gleiche Dateien ist ungeklärt" "1" "$RC7"

# ------------------------------------------ der Bau ohne Pipeline legt nichts aus

# **Was `MSE_FREIGABE_NUR_BAU` darf**: bauen und das Abbild legen. Eine Freigabe
# ohne Pipeline darf daraus nicht werden — der Ausstieg steht vor den
# Prüfsummen, und damit vor Signatur und Beleg.
AUSSTIEG="$(grep -n 'if \[ "${MSE_FREIGABE_NUR_BAU' "$WURZEL/scripts/release.sh" | tail -1 | cut -d: -f1)"
ABBILD="$(grep -n 'ok "$AUSGABE/$NAME-$VERSION.dmg"' "$WURZEL/scripts/release.sh" | cut -d: -f1)"
PRUEFSUMMEN="$(grep -n 'kopf "Stückliste und Prüfsummen"' "$WURZEL/scripts/release.sh" | cut -d: -f1)"
PIPELINE="$(grep -n './scripts/ci.sh all' "$WURZEL/scripts/release.sh" | cut -d: -f1)"
gleich "der Bau ohne Pipeline steigt nach dem Abbild und vor den Prüfsummen aus" "ja" \
  "$([ -n "$AUSSTIEG" ] && [ -n "$ABBILD" ] && [ -n "$PRUEFSUMMEN" ] \
     && [ "$ABBILD" -lt "$AUSSTIEG" ] && [ "$AUSSTIEG" -lt "$PRUEFSUMMEN" ] && echo ja || echo nein)"
gleich "und er steigt wirklich aus" "ja" \
  "$(sed -n "${AUSSTIEG},$((AUSSTIEG + 3))p" "$WURZEL/scripts/release.sh" | grep -q '^  exit 0$' && echo ja || echo nein)"
gleich "die Pipeline überspringt nur er" "ja" \
  "$(sed -n "$((PIPELINE - 3)),${PIPELINE}p" "$WURZEL/scripts/release.sh" | grep -q 'MSE_FREIGABE_NUR_BAU' && echo ja || echo nein)"

exit "$FEHLER"
