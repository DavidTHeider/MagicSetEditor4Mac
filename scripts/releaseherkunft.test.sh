#!/usr/bin/env bash
#
# Prüfungen der Herkunftsdatei aus scripts/releaseherkunft.sh — I06
# (#2264).
#
# Gegen ein **nachgestelltes** Bündel und nachgestellte Stücklisten: ein
# echter Freigabebau dauert Minuten und gehört in den Freigabelauf, nicht in
# jede Pipeline. Geprüft wird, dass jede Zahl in der Datei aus der Datei
# stammt, die sie nennt.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

kopf() { :; }
ok() { :; }
ende() { printf '%s\n' "$*" >&2; exit 1; }

# shellcheck source=scripts/releaseherkunft.sh
source "$WURZEL/scripts/releaseherkunft.sh"

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

feld() { # feld <ausdruck in Python über h>
  python3 -I -c "import json,sys; h=json.load(open(sys.argv[1])); print($1)" "$RAUM/aus/herkunft.json"
}

mkdir -p "$RAUM/Probe.app/Contents/MacOS" "$RAUM/aus"
printf 'fenster' > "$RAUM/Probe.app/Contents/MacOS/Probe"
printf 'kommandozeile' > "$RAUM/Probe.app/Contents/MacOS/mse"
printf '{}' > "$RAUM/aus/stueckliste-go.json"
printf '[]' > "$RAUM/aus/stueckliste-npm.json"
printf 'abbild' > "$RAUM/aus/Probe-1.0.0.dmg"

summe() { shasum -a 256 "$1" | cut -d' ' -f1; }

herkunft_schreiben "$RAUM/aus" "$RAUM/Probe.app" "$WURZEL"

gleich "die Datei ist JSON" "0" "$(python3 -I -m json.tool "$RAUM/aus/herkunft.json" >/dev/null 2>&1; echo $?)"
gleich "der Commit ist der des Baums" "$(git -C "$WURZEL" rev-parse HEAD)" "$(feld "h['commit']")"
gleich "ob der Baum sauber war, steht da" "True" "$(feld "isinstance(h['sauber'], bool)")"
gleich "die Toolchain ist die gepinnte" "$(go env GOVERSION)" "$(feld "h['werkzeuge']['go']")"
gleich "die Stücklisten nennen ihr Werkzeug mit Fassung" "$(awk '$1=="github.com/CycloneDX/cyclonedx-gomod" && $2 ~ /^v/ {print $2}' "$WURZEL/tools/go.mod")" "$(feld "h['werkzeuge']['cyclonedx-gomod']")"
gleich "Wails steht mit seiner Fassung da" "$(awk '$1=="github.com/wailsapp/wails/v3"{print $2}' "$WURZEL/go.mod")" "$(feld "h['werkzeuge']['wails']")"
gleich "jedes Programm im Bündel, mit seiner Summe" \
  "Probe=$(summe "$RAUM/Probe.app/Contents/MacOS/Probe") mse=$(summe "$RAUM/Probe.app/Contents/MacOS/mse")" \
  "$(feld "' '.join(p['name']+'='+p['sha256'] for p in sorted(h['programme'], key=lambda p: p['name']))")"
gleich "beide Stücklisten, mit ihrer Summe" \
  "stueckliste-go.json=$(summe "$RAUM/aus/stueckliste-go.json") stueckliste-npm.json=$(summe "$RAUM/aus/stueckliste-npm.json")" \
  "$(feld "' '.join(s['name']+'='+s['sha256'] for s in sorted(h['stuecklisten'], key=lambda s: s['name']))")"
gleich "das Abbild, mit seiner Summe" "$(summe "$RAUM/aus/Probe-1.0.0.dmg")" "$(feld "h['abbild']['sha256']")"

# **Ohne Bündel keine Herkunft** — eine Datei, die Programme verschwiege,
# sähe aus wie eine vollständige.
gleich "ohne Bündel bricht der Schritt ab" "1" \
  "$( (herkunft_schreiben "$RAUM/aus" "$RAUM/fehlt.app" "$WURZEL") >/dev/null 2>&1; echo $?)"

exit "$FEHLER"
