#!/usr/bin/env bash
#
# Prüfungen von scripts/mse2-referenz.sh — QA-03.AK-03
# (#1656).
#
# **Sie fährt MSE2 nicht.** Ein nachgebautes Programm legt Bilder ab, ein
# Probeset ist ein ZIP mit zwei Karten, und die Schriftprobe ist ein Befehl,
# den der Test vorgibt.
#
# # Was er hält
#
#  1. Ein fehlendes Schriftverzeichnis bricht ab — hier stand ein `if`, das die
#     Prüfung dann übersprang.
#  2. Eine rote Schriftprobe bricht ab.
#  3. Ein Bild aus einem früheren Lauf liegt danach nicht mehr da, und das
#     Manifest nennt Set, Kartenzahl und jedes Bild.
#  4. Mehr Bilder als Karten brechen ab.
#  5. MSE2 läuft unter einem eigenen `HOME` — sonst lädt es Pakete zuerst aus
#     seinem Benutzerverzeichnis, und die Referenz zeigt eine andere Vorlage als
#     die, die wir zeichnen. **Und unter einem eigenen `CFFIXED_USER_HOME`**
#     (#2006):
#     `HOME` allein lenkt wxWidgets auf macOS nicht um. Darin liegt
#     `Library/Application Support/magicseteditor`, sonst hängt MSE2 beim
#     Beenden.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKRIPT="$WURZEL/scripts/mse2-referenz.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0

behauptung() { # behauptung <beschreibung> <erwartet> <gegeben>
  if [ "$3" = "$2" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"
  else
    printf '%s  ✗ %s — %q, erwartet war %q%s\n' "$ROT" "$1" "$3" "$2" "$AUS"
    FEHLER=1
  fi
}

PROBE="$(mktemp -d)"
trap 'rm -rf "$PROBE"' EXIT

# aufbauen legt eine frische Umgebung an; $1 ist die Zahl der Bilder, die das
# nachgebaute MSE2 je Aufruf schreibt.
aufbauen() {
  rm -rf "${PROBE:?}"/*
  mkdir -p "$PROBE/mse2" "$PROBE/pack/data" "$PROBE/pack/Magic - Fonts" \
    "$PROBE/schriften" "$PROBE/sets" "$PROBE/ziel/Probe" "$PROBE/zuhause"
  printf 'eine Schrift' > "$PROBE/pack/Magic - Fonts/a.ttf"
  cp "$PROBE/pack/Magic - Fonts/a.ttf" "$PROBE/schriften/"
  printf 'alt' > "$PROBE/ziel/Probe/Alt.png"

  printf 'mse version: 2.0.0\ngame: magic\ncard:\n\tname: Karte A\ncard:\n\tname: Karte B\n' \
    > "$PROBE/set"
  (cd "$PROBE" && zip -q sets/Probe.mse-set set)

  cat > "$PROBE/mse2/magicseteditor" <<SKRIPT
#!/bin/sh
ziel="\$(dirname "\$3")"
printf '%s' "\$HOME" > "$PROBE/mse2-zuhause"
printf '%s' "\${CFFIXED_USER_HOME:-}" > "$PROBE/mse2-cfzuhause"
if [ -d "\${CFFIXED_USER_HOME:-/fehlt}/Library/Application Support/magicseteditor" ]; then
  printf 'da' > "$PROBE/mse2-cfordner"
fi
i=0
while [ \$i -lt $1 ]; do
  i=\$((i + 1))
  printf 'bild' > "\$ziel/Karte \$i.png"
done
SKRIPT
  chmod +x "$PROBE/mse2/magicseteditor"
}

# fahre ruft das Skript für den Satz `Probe`; $1 ist die Schriftprobe.
fahre() {
  HOME="$PROBE/zuhause" MSE2_BINAER="$PROBE/mse2/magicseteditor" \
    MSE_FREMDPAKETE="$PROBE/pack/data" MSE_KANON="$PROBE/sets" \
    MSE_MSE2AUSGABE="$PROBE/ziel" MSE2_SCHRIFTEN="$PROBE/schriften" \
    MSE2_SCHRIFTPROBE="$1" bash "$SKRIPT" Probe 2>&1
}

# 1. Ohne Schriftverzeichnis
aufbauen 2
rm -rf "$PROBE/pack/Magic - Fonts"
AUSGABE="$(fahre true)"; CODE=$?
behauptung "ein fehlendes Schriftverzeichnis bricht ab" "1" "$CODE"
behauptung "und sagt es" "1" "$(printf '%s' "$AUSGABE" | grep -c 'Schriftverzeichnis des Packs fehlt')"

# 2. Eine rote Schriftprobe
aufbauen 2
AUSGABE="$(fahre false)"; CODE=$?
behauptung "eine rote Schriftprobe bricht ab" "1" "$CODE"
behauptung "und das alte Bild bleibt, denn gezeichnet wurde nicht" "alt" "$(cat "$PROBE/ziel/Probe/Alt.png")"

# 3. Ein Lauf, der trägt
aufbauen 2
AUSGABE="$(fahre true)"; CODE=$?
behauptung "ein Lauf mit Schriften endet grün" "0" "$CODE"
behauptung "das Bild des früheren Laufs ist fort" "nein" "$([ -e "$PROBE/ziel/Probe/Alt.png" ] && echo ja || echo nein)"
MANIFEST="$PROBE/ziel/Probe/manifest.json"
behauptung "das Manifest ist JSON" "0" "$(python3 -m json.tool "$MANIFEST" >/dev/null 2>&1; echo $?)"
behauptung "es nennt zwei Karten" "2" "$(python3 -c "import json,sys; print(json.load(open(sys.argv[1]))['karten'])" "$MANIFEST")"
behauptung "und zwei Bilder" "Karte 1.png,Karte 2.png" \
  "$(python3 -c "import json,sys; print(','.join(sorted(json.load(open(sys.argv[1]))['bilder'])))" "$MANIFEST")"
behauptung "und den Hash des Sets" "sha256:$(shasum -a 256 "$PROBE/sets/Probe.mse-set" | cut -d' ' -f1)" \
  "$(python3 -c "import json,sys; print(json.load(open(sys.argv[1]))['set'])" "$MANIFEST")"

# 5. Ein eigenes Zuhause für MSE2
aufbauen 2
mkdir -p "$PROBE/zuhause/Library/Application Support/magicseteditor/data"
AUSGABE="$(fahre true)"; CODE=$?
ZUHAUSE="$(cat "$PROBE/mse2-zuhause" 2>/dev/null)"
behauptung "MSE2 läuft nicht unter dem Zuhause des Aufrufers" "anderes" \
  "$([ -n "$ZUHAUSE" ] && [ "$ZUHAUSE" != "$PROBE/zuhause" ] && echo anderes || echo dasselbe)"
behauptung "und findet dort kein Paketverzeichnis" "keines" \
  "$([ -e "$ZUHAUSE/Library/Application Support/magicseteditor/data" ] && echo eines || echo keines)"
behauptung "und es bleibt nichts davon liegen" "fort" "$([ -e "$ZUHAUSE" ] && echo da || echo fort)"
behauptung "Core Foundation bekommt dasselbe Zuhause" "$ZUHAUSE" "$(cat "$PROBE/mse2-cfzuhause" 2>/dev/null)"
behauptung "und darin den Ordner, ohne den MSE2 beim Beenden hängt" "da" "$(cat "$PROBE/mse2-cfordner" 2>/dev/null)"

# 4. Mehr Bilder als Karten
aufbauen 3
AUSGABE="$(fahre true)"; CODE=$?
behauptung "mehr Bilder als Karten brechen ab" "1" "$CODE"

exit "$FEHLER"
