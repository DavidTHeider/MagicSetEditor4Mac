#!/usr/bin/env bash
#
# Prüfungen der Schlusszeile von scripts/ci.sh
# (#870).
#
# **Sie fährt die Pipeline nicht.** Geprüft wird die Funktion `reichweite`
# selbst, herausgeschnitten und in einer Muschel ausgeführt: ein Testlauf, der
# `ci.sh` ruft, riefe sich selbst.
#
# # Was er hält
#
#  1. Die Zeile steht überhaupt da — sie ist beim Umbau der Schlussmeldung die
#     erste, die still verschwindet.
#  2. Ihre Zahlen sind **gelesen**: eine erfundene `.covignore` mit drei
#     Zeilen ergibt drei.
#  3. Ein **roter** Lauf sagt sie nicht — dort ist sie im Weg.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# **Zwei Dateien, seit `ci.sh` über 400 Zeilen ging** (§2): dort steht der
# Ablauf, hier die Funktion. **Beide werden geprüft** — die eine muss sie
# erklären, die andere sie rufen; wer nur eine ansähe, ließe die Hälfte offen.
# **Die Schlusszeile steht in `ci-beweis.sh`**, seit `ci.sh` mit dem Riegel
# über die Zeilengrenze aus §2 ging. Hier stand `ci.sh`, und der Wächter wurde
# rot — richtig: er prüft, **wo** `reichweite` gerufen wird, und das hatte sich
# geändert. Ein Wächter, der beim Umbau schweigt, ist der schlechtere.
SKRIPT="$WURZEL/scripts/ci-beweis.sh"
QUELLE="$WURZEL/scripts/ci-reichweite.sh"

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

# **Die Funktion muss es geben.** Ohne diese Prüfung wäre der ganze Lauf grün,
# sobald jemand sie entfernt: `grep` fände nichts, und die Zählungen daneben
# liefen gegen leere Ausgaben.
if ! grep -q '^reichweite()' "$QUELLE"; then
  printf '%s  ✗ ci-reichweite.sh hat keine Funktion reichweite()%s\n' "$ROT" "$AUS"
  exit 1
fi

# **Und sie muss gerufen werden.** Eine Funktion, die dasteht und die niemand
# ruft, ist dieselbe Lücke wie ein Paket ohne Aufrufer.
if ! grep -q '^    reichweite$' "$SKRIPT"; then
  printf '%s  ✗ ci.sh ruft reichweite nicht%s\n' "$ROT" "$AUS"
  exit 1
fi

# **Sie steht im grünen Zweig.** Bei einem roten Lauf will jemand wissen, was
# kaputt ist, und nicht, was ohnehin ungeprüft bleibt.
if grep -A 2 "Pipeline rot" "$SKRIPT" | grep -q 'reichweite'; then
  printf '%s  ✗ reichweite steht auch im roten Zweig%s\n' "$ROT" "$AUS"
  FEHLER=1
else
  printf '%s  ✓ die Reichweite steht nur im grünen Zweig%s\n' "$GRUEN" "$AUS"
fi

# Die Funktion in einem Wegwerfverzeichnis mit erfundenen Gegenständen fahren.
PROBE="$(mktemp -d)"
trap 'rm -rf "$PROBE"' EXIT

mkdir -p "$PROBE/internal" "$PROBE/cmd"
printf '# Kommentar\n\na.go  # eins\nb.go  # zwei\nc.go  # drei\n' > "$PROBE/.covignore"
printf '//go:build fremdpakete\npackage x\n' > "$PROBE/internal/eins_test.go"
printf '//go:build fremdpakete\npackage x\n' > "$PROBE/cmd/zwei_test.go"
printf '//go:build netzlauf\npackage x\n' > "$PROBE/internal/drei_test.go"

# **Ein Plan und ein nachgebautes `go`** — QA-01.AK-04 (#1733): die
# Reichweite nennt, was der letzte Messlauf sagt. Gefahren wird nicht das
# Werkzeug, sondern ob die Reichweite es fragt und seine Antwort zeigt.
mkdir -p "$PROBE/docs/belege" "$PROBE/bin"
printf '{}\n' > "$PROBE/docs/belege/fremdpakete.plan.json"
# **Und wie das echte `go run` endet**: endet das Programm mit 1, hängt `go`
# selbst »exit status 1« an die Fehlerausgabe. Die Vorrichtung ließ das weg,
# und die Reichweite zeigte im Betrieb genau diese Zeile statt des Standes —
# gefunden am ersten grünen Lauf mit einem roten Beleg (#1801).
#
# **Und es merkt sich, womit es gerufen wurde** — die Reichweite rechnet die
# Eingaben nicht (`-ohne-eingaben`): sie kennt sie nicht, und zwölf Gigabyte
# je Pipelinelauf wären ein Preis ohne Ertrag (Zweitblick über Phase 6.0).
cat > "$PROBE/bin/go" <<SKRIPT
#!/bin/sh
printf '%s\n' "\$*" >> "$PROBE/aufrufe"
printf "fremdpakete: FAIL · anderer Stand — seit abc1234 hat sich geändert\n"
printf "exit status 1\n" >&2
exit 1
SKRIPT
chmod +x "$PROBE/bin/go"

AUSGABE="$(cd "$PROBE" && PATH="$PROBE/bin:$PATH" GELB='' AUS='' bash -c "
  $(sed -n '/^reichweite()/,/^}/p' "$QUELLE")
  reichweite
")"

behauptung "die Ausnahmen werden gezählt" "3" \
  "$(printf '%s' "$AUSGABE" | sed -n 's/.*· \([0-9]*\) Einträge in .covignore.*/\1/p')"

# **Je Tag eine eigene Zeile, und je Zeile eine eigene Zusicherung.**
#
# Hier stand ein Muster auf „Dateien hinter dem Tag", und mit der zweiten Zeile
# traf es beide — die Zusicherung bekam `2\n0` und wurde rot. Ein Muster, das
# mehr trifft als gemeint, prüft nicht mehr, was es sagt.
behauptung "die Läufe hinter fremdpakete werden gezählt" "2" \
  "$(printf '%s' "$AUSGABE" | sed -n 's/.*· \([0-9]*\) Dateien hinter dem Tag fremdpakete.*/\1/p')"

behauptung "die Läufe hinter netzlauf werden gezählt" "1" \
  "$(printf '%s' "$AUSGABE" | sed -n 's/.*· \([0-9]*\) Dateien hinter dem Tag netzlauf.*/\1/p')"

behauptung "die eine Maschine wird benannt" "1" \
  "$(printf '%s' "$AUSGABE" | grep -c 'eine Maschine, keine Plattform')"

# **Und der Stand des Belegs steht da** — auch wenn die Prüfung mit 1 endet:
# die Reichweite berichtet, sie sperrt nicht.
behauptung "der Stand des Belegs wird genannt" "1" \
  "$(printf '%s' "$AUSGABE" | grep -c 'Beleg fremdpakete: FAIL · anderer Stand')"

behauptung "die Reichweite rechnet die Eingaben nicht" "1" \
  "$(grep -c -e '-pruefen -ohne-eingaben$' "$PROBE/aufrufe")"

exit "$FEHLER"
