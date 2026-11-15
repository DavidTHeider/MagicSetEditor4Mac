#!/usr/bin/env bash
#
# Prüfungen des Schritts `fm serve` in der Pipeline (AI-02,
# #1743).
#
# **Er fährt das Gerätemodell nicht.** Geprüft wird die Entscheidung davor:
# die Datei wird eingelesen und der Lauf durch einen Stellvertreter ersetzt.
#
# # Was er hält
#
#  1. **Fehlt `fm`, wird nicht gefahren** — gelb gemeldet, ohne Befund und
#     ohne als übersprungene Prüfung zu zählen; `reichweite` sagt es danach.
#  2. **Ist `fm` da, wird gefahren**, und ein roter Lauf ist ein Befund.
#  3. `ci.sh` ruft den Schritt in der Teststufe — eine Funktion, die niemand
#     ruft, ist dieselbe Lücke wie ein Paket ohne Aufrufer.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
QUELLE="$WURZEL/scripts/ci-geraetemodell.sh"

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

# fahre liest die Datei in einer frischen Muschel, ersetzt den Lauf durch
# `$2` (`true` oder `false`) und prüft das Programm `$1`. Ausgabe: die Zeile
# des Schritts, dann FEHLER, UEBERSPRUNGEN, GERAETEMODELL_GEPRUEFT und ob der
# Lauf gefahren wurde.
fahre() {
  bash -c '
    ok()   { printf "OK %s\n" "$1"; }
    warn() { printf "WARN %s\n" "$1"; }
    weg()  { printf "WEG %s\n" "$1"; UEBERSPRUNGEN=$((UEBERSPRUNGEN + 1)); }
    fehl() { printf "FEHL %s\n" "$1"; FEHLER=1; }
    FEHLER=0; UEBERSPRUNGEN=0; GEFAHREN=nein; LAUF="$2"
    . "$1"
    geraetemodell_lauf() { GEFAHREN=ja; "$LAUF"; }
    geraetemodell_pruefen "$3"
    printf "%s %s %s %s\n" "$FEHLER" "$UEBERSPRUNGEN" "$GERAETEMODELL_GEPRUEFT" "$GEFAHREN"
  ' _ "$QUELLE" "$2" "$1"
}

ORT="$(mktemp -d)"
trap 'rm -rf "$ORT"' EXIT
printf '#!/bin/sh\n' > "$ORT/fm"
chmod +x "$ORT/fm"

AUSGABE="$(fahre "$ORT/fehlt" true)"
behauptung "ohne fm: gelb gemeldet" "WARN" "$(printf '%s\n' "$AUSGABE" | head -1 | cut -d' ' -f1)"
behauptung "ohne fm: nicht gefahren, kein Befund, keine übersprungene Prüfung" "0 0 nein nein" \
  "$(printf '%s\n' "$AUSGABE" | tail -1)"

AUSGABE="$(fahre "$ORT/fm" true)"
behauptung "mit fm und grünem Lauf: gefahren, kein Befund" "0 0 ja ja" "$(printf '%s\n' "$AUSGABE" | tail -1)"

AUSGABE="$(fahre "$ORT/fm" false)"
behauptung "mit fm und rotem Lauf: ein Befund" "1 0 rot ja" "$(printf '%s\n' "$AUSGABE" | tail -1)"

behauptung "ci.sh ruft den Schritt in der Teststufe" "1" \
  "$(sed -n '/^stufe_test()/,/^}/p' "$WURZEL/scripts/ci.sh" | grep -c '^  geraetemodell_pruefen$')"

behauptung "reichweite nennt einen Lauf ohne fm" "1" \
  "$(grep -c 'GERAETEMODELL_GEPRUEFT:-}" = "nein"' "$WURZEL/scripts/ci-reichweite.sh")"

exit "$FEHLER"
