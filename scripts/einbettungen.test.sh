#!/usr/bin/env bash
#
# **Jede eingebettete Datei eines fremden Pakets ist zugeordnet** — #2393,
# Punkt 3: ein Wächter, der eine neue erzeugte Form bemerkt.
#
# Gelesen wird, was die vier ausgelieferten Programme mit den Schaltern der
# Freigabe einbinden (`go list -deps`, `EmbedFiles`): das Programm mit
# `production` und den Weglassschaltern, die drei Werkzeuge mit den
# Weglassschaltern allein (`releaseschalter.sh`). Jedes fremde Paket mit
# eingebetteten Dateien steht in `EINBETTUNGEN` (`releasequelltext.sh`) —
# **erzeugt**, dann liegen seine Bauquellen über `ERZEUGTE_MODULE` bei, oder
# **Quelle**, dann ist die Datei selbst die Form zum Bearbeiten, mit Grund.
#
# **Beide Richtungen**: ein neues Paket ohne Zeile ist rot, und eine Zeile
# ohne Paket auch — sonst stünde dort eine Zusage über etwas, das nicht mehr
# eingebunden ist.
#
# # Was er nicht prüft
#
# **Ob die Zuordnung stimmt.** Ob eine Datei handgeschrieben oder gebaut ist,
# entscheidet ein Mensch beim Lesen; der Wächter erzwingt nur, dass jemand
# es entschieden hat. **Und was nicht eingebettet ist**: übersetzter Go-Code
# aus Generatoren (Tabellen in `golang.org/x/text`) ist Go und liegt in
# `vendor/` bei.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0
gut() { printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"; }
rot() { printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"; FEHLER=1; }

ende() { printf '%s  ✗ abgebrochen: %s%s\n' "$ROT" "$1" "$AUS"; exit 1; }

# shellcheck source=scripts/releaseschalter.sh
source "$WURZEL/scripts/releaseschalter.sh"
# shellcheck source=scripts/releasequelltext.sh
source "$WURZEL/scripts/releasequelltext.sh"

# fremde_pakete <schalter> <paket…> nennt jedes fremde Paket, das die
# Programme mit diesen Schaltern einbinden, mit `+` davor, wenn es Dateien
# einbettet. **Ohne Netz und ohne go.sum zu schreiben** (`-mod=readonly`).
fremde_pakete() {
  local schalter="$1"
  shift
  (cd "$WURZEL" && GOFLAGS=-mod=readonly GOPROXY=off go list -tags "$schalter" -deps \
    -f '{{if not .Standard}}{{if .EmbedFiles}}+{{end}}{{.ImportPath}}{{end}}' "$@") |
    grep -v '^+\?github.com/DavidTHeider/MagicSetEditor4Mac' | grep -v '^$' | sort -u
}

PROGRAMM="$(fremde_pakete "production,$WEGLASSSCHALTER" .)" || ende "go list über das Programm scheiterte"
WERKZEUGE="$(fremde_pakete "$WEGLASSSCHALTER" ./cmd/mse ./cmd/mse-mcp ./cmd/mse-rest)" || ende "go list über die Werkzeuge scheiterte"
ALLE="$(printf '%s\n%s\n' "$PROGRAMM" "$WERKZEUGE" | sort -u)"
[ -n "$ALLE" ] || ende "kein fremdes Paket gefunden — der Wächter griffe ins Leere"

EINBETTEND="$(grep '^+' <<<"$ALLE" | cut -c2- | sort -u)"
ZUGEORDNET="$(printf '%s\n' "${EINBETTUNGEN[@]}" | cut -d'|' -f1 | sort -u)"

while IFS= read -r PAKET; do
  [ -n "$PAKET" ] || continue
  if grep -qxF "$PAKET" <<<"$ZUGEORDNET"; then
    gut "$PAKET ist zugeordnet"
  else
    rot "$PAKET bettet Dateien ein und steht nicht in EINBETTUNGEN — erzeugt oder Quelle?"
  fi
done <<<"$EINBETTEND"

while IFS= read -r PAKET; do
  [ -n "$PAKET" ] || continue
  grep -qxF "$PAKET" <<<"$EINBETTEND" ||
    rot "EINBETTUNGEN nennt $PAKET, das kein Programm mehr einbettet"
done <<<"$ZUGEORDNET"

# modul_von nennt das Modul aus `ERZEUGTE_MODULE`, zu dem ein Paket gehört —
# leer, wenn keines. **Eine Funktion und kein `case` in `$(…)`**: die Bash
# von macOS (3.2) liest dort die schließende Klammer eines Musters als Ende.
modul_von() { # modul_von <paket>
  local EINTRAG M
  for EINTRAG in "${ERZEUGTE_MODULE[@]}"; do
    M="${EINTRAG%%:*}"
    if [ "$1" = "$M" ] || [ "${1#"$M"/}" != "$1" ]; then
      printf '%s\n' "$M"
      return
    fi
  done
}

# **Eine erzeugte Einbettung hat ihre Bauquellen** — ihr Modul steht in
# `ERZEUGTE_MODULE`, und `bauquellen_beilegen` legt sie bei.
for EINTRAG in "${EINBETTUNGEN[@]}"; do
  IFS='|' read -r PAKET ART GRUND <<<"$EINTRAG"
  case "$ART" in
    erzeugt)
      MODUL="$(modul_von "$PAKET")"
      if [ -n "$MODUL" ]; then gut "$PAKET ist erzeugt, die Bauquellen von $MODUL liegen bei"
      else rot "$PAKET ist erzeugt, sein Modul steht nicht in ERZEUGTE_MODULE"; fi
      ;;
    quelle)
      if [ -n "$GRUND" ]; then gut "$PAKET ist Quelle: $GRUND"
      else rot "$PAKET ist Quelle, aber ohne Grund"; fi
      ;;
    *) rot "$PAKET hat die Art »$ART« — erzeugt oder quelle" ;;
  esac
done

# **Was die Weglassschalter nehmen**, ist wirklich weg — und ohne sie wäre es
# da. Die Gegenprobe läuft in jedem Lauf mit: ein Wächter, dessen Schalter
# nichts verschöben, wäre grün aus dem falschen Grund.
if grep -q 'web-client-prebuilt' <<<"$ALLE"; then
  rot "Tailscales vorgebaute Weboberfläche steckt trotz ts_omit_webclient im Programm"
else
  gut "Tailscales vorgebaute Weboberfläche steckt in keinem Programm"
fi
OHNE="$(fremde_pakete production .)" || ende "go list ohne Weglassschalter scheiterte"
if grep -q 'web-client-prebuilt' <<<"$OHNE" && grep -qx '+tailscale.com/util/eventbus' <<<"$OHNE"; then
  gut "ohne die Weglassschalter wären Weboberfläche und Ereignisbusseite da — die Schalter verschieben etwas"
else
  rot "ohne die Weglassschalter fehlt trotzdem etwas — die Gegenprobe misst nichts"
fi
if grep -qx '+tailscale.com/util/eventbus' <<<"$ALLE"; then
  rot "die Diagnoseseite des Ereignisbusses steckt trotz ts_omit_debugeventbus im Programm"
else
  gut "die Diagnoseseite des Ereignisbusses steckt in keinem Programm"
fi

printf '\n'
if [ "$FEHLER" -eq 0 ]; then
  printf '%s✓ Einbettungen zugeordnet%s\n' "$GRUEN" "$AUS"
else
  printf '%s✗ Einbettungen nicht zugeordnet%s\n' "$ROT" "$AUS"
fi

exit "$FEHLER"
