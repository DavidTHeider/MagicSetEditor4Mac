#!/usr/bin/env bash
#
# **Der Freigabeweg ab der Sicherheitslage — ausgeführt, nicht gelesen**
# (#1830,
# Befund 4 des Nachlaufs über Phase 6.2).
#
# # Der Befund
#
# `sicherheitslage.test.sh` hielt, dass `release.sh` das Urteil fragt — und
# las dafür den **Text**. Ein `|| true` hinter dem Aufruf machte das Urteil
# wirkungslos, und der Wächter bliebe grün.
#
# # Was dieser Lauf tut
#
# Er fährt `freigabeschluss` — den Teil von `release.sh` ab der Lage — mit
# einer gescheiterten Lage und sieht ihn **vor „Fertig"** enden. Und er
# macht die Gegenprobe selbst: dieselbe Funktion mit einem wirkungslos
# gemachten Urteil muss „Fertig" erreichen, sonst sähe dieser Lauf den
# Unterschied gar nicht.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# shellcheck source=scripts/releasebeleg.sh
source "$WURZEL/scripts/releasebeleg.sh"
# shellcheck source=scripts/sicherheitslage.sh
source "$WURZEL/scripts/sicherheitslage.sh"

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

# **Was `release.sh` dem Schluss bereitstellt** — die Ausgaben und `ende`,
# das wie dort mit `exit 1` abbricht.
# shellcheck disable=SC2034 # GELB liest der Schluss
GELB=''
kopf() { printf '── %s\n' "$1"; }
ok() { printf '  ✓ %s\n' "$1"; }
ende() { printf '  ✗ %s\n' "$1"; exit 1; }

# lage legt fest, was die Erhebung ergibt — statt der Scanner.
LAGE=''
sicherheitslage() { printf '%s' "$LAGE"; }

# fahre fährt den Schluss in einer Unterschale über einem Ausgabeort und
# gibt Rückgabe und Ausgabe als eine Zeile: `<rückgabe>|<fertig?>`.
fahre() { # fahre <schlussdatei> <lage>
  local ort="$RAUM/ausgabe-$RANDOM" ausgabe rueckgabe=0
  mkdir -p "$ort"
  printf 'abc  MagicSetEditor4Mac.dmg\n' > "$ort/pruefsummen.txt"
  LAGE="$2"

  ausgabe="$(
    # shellcheck source=/dev/null
    source "$1"
    freigabeschluss "$ort" "9.9.9" ""
  )" || rueckgabe=$?

  if printf '%s' "$ausgabe" | grep -q '── Fertig'; then
    printf '%s|fertig' "$rueckgabe"
  else
    printf '%s|angehalten' "$rueckgabe"
  fi
}

GESCHEITERT=$'govulncheck: ohne-befund\nnpm audit: gescheitert — npm error code ELOCKVERIFY\n'
GRUENE=$'govulncheck: ohne-befund\nnpm audit: ohne-befund — kritisch 0 · hoch 0 · mittel 0 · niedrig 0\n'

SCHLUSS="$WURZEL/scripts/releaseschluss.sh"

# **Der eigentliche Fall**: eine gescheiterte Lage hält vor „Fertig" an.
gleich "eine gescheiterte Lage hält den Freigabeweg vor „Fertig“ an" \
  "1|angehalten" "$(fahre "$SCHLUSS" "$GESCHEITERT")"

# **Und die Gegenseite**: eine grüne Lage erreicht „Fertig" — sonst hieße
# „angehalten" oben nur, dass der Schluss nie durchläuft.
gleich "eine grüne Lage erreicht „Fertig“" \
  "0|fertig" "$(fahre "$SCHLUSS" "$GRUENE")"

# **Die Gegenprobe im Lauf** (§10): dasselbe mit einem wirkungslos gemachten
# Urteil muss „Fertig" erreichen. **Und der Eingriff muss etwas ändern** — ein
# `sed`, das nichts trifft, bewiese nur, dass die Kopie gleich ist.
sed 's/|| ende "\$meldung"/|| true/' "$SCHLUSS" > "$RAUM/wirkungslos.sh"
if cmp -s "$SCHLUSS" "$RAUM/wirkungslos.sh"; then
  gleich "die Gegenprobe greift in den Schluss ein" "geändert" "unverändert"
else
  gleich "die Gegenprobe greift in den Schluss ein" "geändert" "geändert"
fi
gleich "ein wirkungslos gemachtes Urteil ließe die gescheiterte Lage durch" \
  "0|fertig" "$(fahre "$RAUM/wirkungslos.sh" "$GESCHEITERT")"

# **Und `release.sh` fährt genau diesen Schluss** — sonst prüfte dieser Lauf
# eine Funktion, die niemand ruft.
# shellcheck disable=SC2016 # die Variablen sind Text im Skript
if grep -Eq '^[^#]*freigabeschluss "\$AUSGABE" "\$VERSION" "\$FINGERABDRUCK"' "$WURZEL/scripts/release.sh" \
  && ! grep -Eq '^[^#]*freigabeurteil ' "$WURZEL/scripts/release.sh"; then
  gleich "release.sh fährt den Schluss und fragt kein Urteil daneben" "ja" "ja"
else
  gleich "release.sh fährt den Schluss und fragt kein Urteil daneben" "ja" "nein"
fi

printf '\n'
if [ "$FEHLER" -eq 0 ]; then
  printf '%s✓ Freigabeschluss in Ordnung%s\n' "$GRUEN" "$AUS"
else
  printf '%s✗ Freigabeschluss fehlerhaft%s\n' "$ROT" "$AUS"
fi

exit "$FEHLER"
