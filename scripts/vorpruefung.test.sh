#!/usr/bin/env bash
#
# Prüfungen des Tagblocks in scripts/vorpruefung.sh
# (#1452).
#
# **Sie fährt die Vorprüfung nicht.** Geprüft wird die **Ableitung** der
# Paketliste — der Satz, für den es diesen Block gibt: *die Paketliste wird
# berechnet und nicht eingetragen.* Eine Aufzählung im Skript veraltete beim
# nächsten Paket, das einen Tag bekommt, und wäre still zu kurz.
#
# # Was er hält
#
#  1. Der Block steht überhaupt da — er ist beim nächsten Umbau der erste,
#     der still verschwindet, und dann ist die Vorprüfung wieder blind für
#     alles hinter einem Build-Tag.
#  2. Die Ableitung findet, was dasteht: jedes Paket mit einer so markierten
#     Datei taucht in der Liste auf.
#  3. **Die Gegenprobe**: ein Paket, das heute keinen Tag trägt, kommt dazu,
#     sobald eine markierte Datei darin liegt. Ohne sie sagte Punkt 2 nur,
#     dass die Ableitung heute zufällig stimmt.
#
# # Was er nicht hält
#
# **Dass der Lauf den richtigen Linter ruft** oder dass sein Ergebnis stimmt —
# das tut die Vorprüfung selbst, und dahinter steht `make ci` mit dem vollen
# Lauf über **alle** Pakete. Dieser Test prüft die Auswahl, nicht die Prüfung.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKRIPT="$WURZEL/scripts/vorpruefung.sh"

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

# ableiten ist **dieselbe** Zeile wie im Skript.
#
# **Herausgeschnitten und nicht nachgebaut**: zwei Fassungen einer Regel laufen
# beim ersten Sonderfall auseinander, und dann prüft dieser Lauf eine, die es
# nicht gibt. Der `grep` unten hält, dass sie im Skript wirklich so steht.
ableiten() { # ableiten <tag>
  grep -rl "^//go:build $1" "$WURZEL/internal" "$WURZEL/cmd" 2>/dev/null \
    | xargs -n1 dirname 2>/dev/null | sort -u
}

# ---------------------------------------------------------------- 1. Der Block

if grep -q 'golangci-lint run --build-tags' "$SKRIPT"; then
  behauptung "vorpruefung.sh lintet über Build-Tags" "ja"
else
  behauptung "vorpruefung.sh lintet über Build-Tags" "nein"
  printf '%s  → ohne ihn ist jede Datei hinter einem Tag hier ungeprüft%s\n' \
    "$ROT" "$AUS"
  exit 1
fi

if grep -q 'grep -rl "\^//go:build \$tag"' "$SKRIPT"; then
  behauptung "die Paketliste wird abgeleitet, nicht aufgezählt" "ja"
else
  behauptung "die Paketliste wird abgeleitet, nicht aufgezählt" "nein"
fi

# ------------------------------------------------------------ 2. Sie findet was

for tag in fremdpakete netzlauf; do
  gefunden="$(ableiten "$tag" | wc -l | tr -d ' ')"

  if [ "$gefunden" -gt 0 ]; then
    behauptung "$tag: $gefunden Pakete abgeleitet" "ja"
  else
    behauptung "$tag: kein Paket abgeleitet — trägt ihn niemand mehr?" "nein"
  fi
done

# ------------------------------------------------------------- 3. Die Gegenprobe

# **Ein Paket, das heute keinen der beiden Tags trägt.** `internal/pdf` ist
# bewusst gewählt: es importiert nichts aus dem Projekt (§2) und wird deshalb
# von keiner Umstellung mitgezogen.
PROBEPAKET="internal/pdf"
PROBE="$WURZEL/$PROBEPAKET/probetag_test.go"

if ableiten fremdpakete | grep -q "/$PROBEPAKET\$"; then
  behauptung "$PROBEPAKET trägt den Tag heute nicht" "nein"
else
  behauptung "$PROBEPAKET trägt den Tag heute nicht" "ja"
fi

aufraeumen() { rm -f "$PROBE"; }
trap aufraeumen EXIT

cat > "$PROBE" <<'GOEOF'
//go:build fremdpakete

package pdf
GOEOF

if ableiten fremdpakete | grep -q "/$PROBEPAKET\$"; then
  behauptung "eine neue markierte Datei kommt in die Liste" "ja"
else
  behauptung "eine neue markierte Datei kommt in die Liste" "nein"
fi

aufraeumen
trap - EXIT

if ableiten fremdpakete | grep -q "/$PROBEPAKET\$"; then
  behauptung "und verschwindet wieder" "nein"
else
  behauptung "und verschwindet wieder" "ja"
fi

exit "$FEHLER"
