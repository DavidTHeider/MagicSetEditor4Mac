#!/usr/bin/env bash
#
# Prüfungen der Audit-Stufe von scripts/ci.sh
# (#1350).
#
# # Der Anlass
#
# Am 2026-09-04 gab `registry.npmjs.org` für den Audit-Endpunkt zweimal **503**,
# und die Pipeline schrieb beide Male „npm audit — Befunde ab 'high'". Das ist
# eine Meldung, die mehr behauptet, als sie weiß: sie nennt einen Befund, wo gar
# keine Antwort kam.
#
# # Was er hält
#
#  1. Ein **Ausfall** wird als übergangen geführt — nicht als grün und nicht
#     als Befund.
#  2. Ein **echter Befund** bleibt rot.
#  3. Ein **sauberer Lauf** bleibt grün.
#
# **Er fährt `npm` nicht.** Geprüft wird die Funktion selbst, herausgeschnitten
# und mit einem erfundenen `npm` im Pfad — ein Lauf gegen die echte Registry
# prüfte die Leitung (§3), und die war ja gerade der Anlass.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# **Seit ENG-04 steht die Stufe in einer eigenen Datei** und liest den Bericht
# als JSON (`scripts/ci-sicherheit.sh`, `scripts/sicherheitslage.sh`).
SKRIPT="$WURZEL/scripts/ci-sicherheit.sh"

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

# **Die Funktion muss es geben und sie muss gerufen werden.** Eine, die
# dasteht und die niemand ruft, ist dieselbe Lücke wie ein Paket ohne Aufrufer.
for zeile in '^npm_audit_pruefen()' '^    npm_audit_pruefen$'; do
  if ! grep -q "$zeile" "$SKRIPT"; then
    printf '%s  ✗ ci-sicherheit.sh führt %s nicht%s\n' "$ROT" "$zeile" "$AUS"
    exit 1
  fi
done

printf '%s  ✓ die Funktion steht da und wird gerufen%s\n' "$GRUEN" "$AUS"

PROBE="$(mktemp -d)"
trap 'rm -rf "$PROBE"' EXIT

mkdir -p "$PROBE/frontend" "$PROBE/bin"

# fahre ruft die Funktion mit einem erfundenen `npm`.
#
# **Ausgabe und Rückgabewert kommen aus der Probe** — genau die beiden Dinge,
# an denen die Funktion unterscheidet.
#
# **Die Ausgabe steht in einer Datei und nicht im Skript**: ein Bericht als JSON
# trägt Anführungszeichen, und in eine Zeile des erfundenen `npm` eingesetzt
# verlöre er sie — der saubere Lauf wurde dann zu einem gescheiterten, und der
# Befund war nur zufällig rot (ENG-04, beim Umstellen auf `--json` gesehen).
fahre() { # fahre <ausgabe> <rückgabewert>
  printf '%s\n' "$1" > "$PROBE/ausgabe"
  cat > "$PROBE/bin/npm" <<NPM
#!/usr/bin/env bash
cat "$PROBE/ausgabe"
exit $2
NPM

  chmod +x "$PROBE/bin/npm"

  # **Mit `set -euo pipefail`, wie `ci.sh` sie fährt.** Ohne diese Zeile war
  # die Vorrichtung großzügiger als der Betrieb (§3): der erste Anlauf der
  # Funktion brach unter `set -e` still ab, und dieser Test war grün.
  (cd "$PROBE" && PATH="$PROBE/bin:$PATH" bash -c "
    set -euo pipefail
    ok()   { printf 'OK %s\n'   \"\\\$1\"; }
    fehl() { printf 'FEHL %s\n' \"\\\$1\"; }
    weg()  { printf 'WEG %s\n'  \"\\\$1\"; }
    source \"$SKRIPT\"
    npm_audit_pruefen
  " | grep -oE '^(OK|FEHL|WEG)')
}

behauptung "ein sauberer Lauf ist grün" "OK" \
  "$(fahre '{"metadata":{"vulnerabilities":{"info":0,"low":0,"moderate":0,"high":0,"critical":0,"total":0}}}' 0)"

behauptung "ein Ausfall der Registry wird übergangen" "WEG" \
  "$(fahre 'npm error audit endpoint returned an error' 1)"

behauptung "ein echter Befund bleibt rot" "FEHL" \
  "$(fahre '{"metadata":{"vulnerabilities":{"info":0,"low":0,"moderate":0,"high":3,"critical":0,"total":3}}}' 1)"

# **Und ein mittlerer Fund allein ist grün** — die Schwelle ist `high`. Ohne
# diesen Fall bewiese der saubere Lauf nicht, dass die Zahlen gelesen werden.
behauptung "ein mittlerer Fund allein bleibt grün" "OK" \
  "$(fahre '{"metadata":{"vulnerabilities":{"info":0,"low":2,"moderate":1,"high":0,"critical":0,"total":3}}}' 0)"

if [ "$FEHLER" -eq 0 ]; then
  printf '%s  ✓ audit.test.sh%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ audit.test.sh%s\n' "$ROT" "$AUS"
fi

exit "$FEHLER"
