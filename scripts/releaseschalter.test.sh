#!/usr/bin/env bash
#
# Prüfungen der Bauschalter aus scripts/releaseschalter.sh — ENG-02.AK-05
# (#1727).
#
# Gegen **erfundene** Ausgaben von `go version -m`: ein Programm mit `server`
# zu bauen, um es hier abzuweisen, kostete eine Minute und bewiese nicht mehr.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# shellcheck source=scripts/releaseschalter.sh
source "$WURZEL/scripts/releaseschalter.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0

gut()  { printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"; }
rot()  { printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"; FEHLER=1; }

# ausgabe <schalter> baut die Zeilen, die `go version -m` für ein Programm
# schreibt — mit Tabulatoren, wie Go sie setzt (am 2026-09-24 nachgelesen).
ausgabe() {
  printf 'build/bin/X.app/Contents/MacOS/X: go1.27.1\n'
  printf '\tpath\tgithub.com/DavidTHeider/MagicSetEditor4Mac\n'
  printf '\tbuild\t-buildmode=exe\n'
  printf '\tbuild\t-compiler=gc\n'
  [ -n "$1" ] && printf '\tbuild\t-tags=%s\n' "$1"
  printf '\tbuild\tCGO_ENABLED=1\n'
  printf '\tbuild\tGOOS=darwin\n'
}

erwarte_gut() { # erwarte_gut <beschreibung> <schalter> [rolle]
  local befund
  befund="$(ausgabe "$2" | schalter_aus | schalter_befund "${3:-programm}")"
  if [ -z "$befund" ]; then gut "$1"; else rot "$1 — Befund: $befund"; fi
}

erwarte_befund() { # erwarte_befund <beschreibung> <schalter> <wort im befund> [rolle]
  local befund
  befund="$(ausgabe "$2" | schalter_aus | schalter_befund "${4:-programm}")"
  case "$befund" in
    *"$3"*) gut "$1" ;;
    *) rot "$1 — erwartet war ein Befund mit „$3“, kam: „$befund“" ;;
  esac
}

# **Die Weglassschalter** (#2393) nehmen aus fremden Modulen, was dieses
# Programm nicht bedient und dessen Bearbeitungsform sonst beiliegen müsste.
# Ein Freigabebau trägt sie — das Programm zu `production` dazu, die drei
# Werkzeuge allein.
W="$WEGLASSSCHALTER"

erwarte_gut    "production mit den Weglassschaltern geht durch" "production,$W"
erwarte_gut    "die Reihenfolge zählt nicht" "${W#*,},production,${W%%,*}"
erwarte_befund "production allein lässt Tailscales Weboberfläche drin" "production" "ts_omit_webclient"
erwarte_befund "ohne den zweiten Weglassschalter bleibt der Ereignisbus" "production,ts_omit_webclient" "ts_omit_debugeventbus"
erwarte_befund "ohne Schalter fehlt production" "" "production"
erwarte_befund "server ist ein Zugang" "production,$W,server" "server"
erwarte_befund "mcp ist ein Zugang" "production,$W,mcp" "mcp"
erwarte_befund "beide werden genannt" "mcp,production,$W,server" "mcp,server"
erwarte_befund "server ohne production" "server,$W" "production"
erwarte_befund "ein Name, der production nur enthält, zählt nicht" "noproduction,$W" "production"

erwarte_gut    "ein Werkzeug trägt die Weglassschalter allein" "$W" werkzeug
erwarte_befund "ein Werkzeug ohne sie lässt die Weboberfläche drin" "" "ts_omit_webclient" werkzeug
erwarte_befund "ein Werkzeug mit production trägt einen fremden Schalter" "production,$W" "production" werkzeug
erwarte_befund "eine unbekannte Rolle ist ein Befund" "$W" "Rolle" fenster

# **Und der Freigabeweg ruft die Prüfung** — sonst hielte sie nur ihre eigenen
# Fälle. Gelesen wird die Zeile, die das gebaute Programm prüft, nicht ein
# Kommentar darüber.
if grep -Eq '^[^#]*go version -m "\$APP/Contents/MacOS/\$NAME"[^#]*\| *schalter_aus' \
  "$WURZEL/scripts/release.sh"; then
  gut "release.sh prüft die Schalter am gebauten Programm"
else
  rot "release.sh prüft die Schalter am gebauten Programm nicht"
fi

# **Und an jedem Werkzeug** — `werkzeuge.sh` baut sie getrennt, und ein
# Werkzeug ohne die Weglassschalter trüge die Weboberfläche mit.
if grep -Eq '^[^#]*go version -m "\$APP/Contents/MacOS/\$werkzeug"[^#]*\| *schalter_aus *\| *schalter_befund werkzeug' \
  "$WURZEL/scripts/release.sh"; then
  gut "release.sh prüft die Schalter an jedem Werkzeug"
else
  rot "release.sh prüft die Schalter an den Werkzeugen nicht"
fi

# **Gebaut wird mit derselben Liste** — `bauen.sh` und `werkzeuge.sh` lesen
# sie aus diesem Blatt und schreiben sie nicht ab.
for skript in bauen.sh werkzeuge.sh; do
  if grep -Eq '^[^#]*-tags "[^"]*\$WEGLASSSCHALTER' "$WURZEL/scripts/$skript"; then
    gut "$skript baut mit \$WEGLASSSCHALTER"
  else
    rot "$skript baut nicht mit \$WEGLASSSCHALTER"
  fi
done

printf '\n'
if [ "$FEHLER" -eq 0 ]; then
  printf '%s✓ Bauschalter in Ordnung%s\n' "$GRUEN" "$AUS"
else
  printf '%s✗ Bauschalter fehlerhaft%s\n' "$ROT" "$AUS"
fi

exit "$FEHLER"
