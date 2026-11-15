#!/usr/bin/env bash
#
# Prüfungen des Berichts der Vorprüfung — was sie im roten Fall zeigt.
#
# # Der Anlass
#
# **Am 2026-09-30 war die Vorprüfung rot und nannte zwei Funde** — die
# Schichtenregel und den Bindungsvertrag. Behoben, `make ci` gestartet, und
# zwanzig Minuten später war die Pipeline rot mit **vier anderen**: drei
# Wächter in `internal/werkzeuge` und eine unbetretene Anweisung in
# `internal/app`. Die Vorprüfung hatte sie gesehen und nicht gezeigt:
#
#  1. sie schnitt den Bericht nach **zwanzig Zeilen** ab, und die ersten zwei
#     Funde füllten sie;
#  2. sie zählte die unbetretenen Anweisungen **nur, wenn alles grün war**.
#
# Das ist der Satz aus §4: „Was eine Stufe im roten Fall zeigt, benennt den
# Fund — eine Bilanz ohne Namen kostet einen zweiten vollen Lauf."
#
# # Was er hält
#
#  1. Jeder rote Lauf steht mit Namen da, auch hinter vielen Zeilen eines
#     anderen.
#  2. Die rote Paketzeile steht da, und eine Panik auch.
#  3. Die Abdeckung wird gezählt, **wie auch immer** `go test` ausging.
#
# # Was er nicht hält
#
# Dass die Namen reichen, um den Fund zu verstehen — dafür steht darunter die
# erste Zeile seiner Begründung, mehr nicht.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0

behauptung() { # behauptung <beschreibung> <bedingung-erfüllt?>
  if [ "$2" = "ja" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"
  else
    printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"
    FEHLER=1
  fi
}

if [ ! -f "$WURZEL/scripts/vorpruefungsbericht.sh" ]; then
  behauptung "scripts/vorpruefungsbericht.sh steht da" "nein"
  exit 1
fi

# shellcheck source=scripts/vorpruefungsbericht.sh
source "$WURZEL/scripts/vorpruefungsbericht.sh"

PROTOKOLL="$(mktemp -t vorpruefungsbericht)"
trap 'rm -f "$PROTOKOLL"' EXIT

# **Ein Protokoll wie das vom 2026-09-30**: zuerst ein Paket mit langer
# Begründung, dann eines mit einem zweiten Fund und einer Panik.
{
  printf -- '--- FAIL: TestKeinPaketVerletztDieSchichtenregel (0.25s)\n'
  for zeile in $(seq 1 30); do
    printf '    schichten_test.go:50: Zeile %d der Begründung\n' "$zeile"
  done
  printf 'FAIL\n'
  printf 'FAIL\tgithub.com/x/internal/schichten\t1.768s\n'
  printf 'ok  \tgithub.com/x/internal/domain\t0.5s\n'
  printf -- '--- FAIL: TestJedeFreieEingabeStehtInDerFeldpruefung (0.06s)\n'
  printf '    feldpruefung_test.go:71: Diagnose.tsx steht nicht darin\n'
  printf -- '    --- FAIL: TestJederEinstellungswertHatEinenNamen/de (0.01s)\n'
  printf '        einstellungstexte_test.go:70: netz.schnittstelle fehlt\n'
  printf 'panic: kaputt\n'
  printf 'FAIL\tgithub.com/x/internal/werkzeuge\t21.887s\n'
} > "$PROTOKOLL"

bericht="$(fehlschlaege_nennen "$PROTOKOLL")"

for name in TestKeinPaketVerletztDieSchichtenregel \
  TestJedeFreieEingabeStehtInDerFeldpruefung \
  'TestJederEinstellungswertHatEinenNamen/de' \
  'internal/schichten' 'internal/werkzeuge' 'panic: kaputt'; do
  if grep -qF "$name" <<<"$bericht"; then
    behauptung "der Bericht nennt $name" "ja"
  else
    behauptung "der Bericht nennt $name" "nein"
  fi
done

# **Die erste Zeile der Begründung, und nicht alle dreißig.**
if grep -qF 'Zeile 1 der Begründung' <<<"$bericht" &&
  ! grep -qF 'Zeile 2 der Begründung' <<<"$bericht"; then
  behauptung "je Fund die erste Zeile der Begründung" "ja"
else
  behauptung "je Fund die erste Zeile der Begründung" "nein"
fi

# **Ein grünes Paket steht nicht darin** — sonst wäre es wieder eine Bilanz.
if grep -qF 'internal/domain' <<<"$bericht"; then
  behauptung "ein grünes Paket bleibt draußen" "nein"
else
  behauptung "ein grünes Paket bleibt draußen" "ja"
fi

# ------------------------------------------- 3. Gezählt wird in jedem Fall

# **Die Zählung steht außerhalb des Zweigs**, der vom Ausgang von `go test`
# abhängt — eine Zeile ohne Einrückung, nach dem `fi`.
# shellcheck disable=SC2016 # das Dollarzeichen ist gemeint, nicht erweitert
if grep -qE '^abdeckung_pruefen "\$PROFIL" ' "$WURZEL/scripts/vorpruefung.sh"; then
  behauptung "die Abdeckung wird auch nach einem roten Lauf gezählt" "ja"
else
  behauptung "die Abdeckung wird auch nach einem roten Lauf gezählt" "nein"
fi

exit "$FEHLER"
