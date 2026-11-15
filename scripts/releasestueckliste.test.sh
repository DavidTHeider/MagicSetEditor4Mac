#!/usr/bin/env bash
#
# Prüfungen der Stücklisten aus scripts/releasestueckliste.sh — SEC04
# (#2357).
#
# Gegen **Attrappen** von `cyclonedx-gomod` und `npm`: der echte Bau gehört
# in den Freigabelauf. Die Attrappe von `cyclonedx-gomod` tut, was v1.12.0
# tut — **gemessen am 2026-10-16**: ohne `-json` schreibt es XML, auch unter
# einem Namen auf `.json`. Die `stueckliste-go.json` der Freigabe 0.8.0 ist
# 87.828 Byte groß, genau so groß wie die XML-Ausgabe desselben Stands.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

ok() { :; }
hin() { :; }
ende() { printf '%s\n' "$*" >&2; exit 1; }

# shellcheck source=scripts/releasestueckliste.sh
source "$WURZEL/scripts/releasestueckliste.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0
RAUM="$(mktemp -d)"
trap 'rm -rf "$RAUM"' EXIT

behauptung() { # behauptung <beschreibung> <wahr?>
  if [ "$2" = ja ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"
  else
    printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"
    FEHLER=1
  fi
}

# attrappen <art der npm-ausgabe> — `cyclonedx-gomod` wie v1.12.0 und ein
# `npm`, dessen `sbom` JSON, XML oder nichts schreibt.
attrappen() {
  mkdir -p "$RAUM/bin" "$RAUM/frontend"
  cat > "$RAUM/bin/cyclonedx-gomod" <<'ATTRAPPE'
#!/usr/bin/env bash
json=nein; ziel=""
while [ $# -gt 0 ]; do
  case "$1" in
    -json) json=ja ;;
    -output) shift; ziel="$1" ;;
  esac
  shift
done
if [ "$json" = ja ]; then
  printf '{"bomFormat":"CycloneDX","specVersion":"1.6","components":[]}\n' > "$ziel"
else
  printf '<?xml version="1.0" encoding="UTF-8"?>\n<bom xmlns="http://cyclonedx.org/schema/bom/1.6"/>\n' > "$ziel"
fi
ATTRAPPE
  case "$1" in
    json) printf '#!/usr/bin/env bash\nprintf %s\n' "'{\"bomFormat\":\"CycloneDX\",\"specVersion\":\"1.5\"}'" > "$RAUM/bin/npm" ;;
    xml) printf '#!/usr/bin/env bash\nprintf %s\n' "'<bom/>'" > "$RAUM/bin/npm" ;;
    leer) printf '#!/usr/bin/env bash\nexit 1\n' > "$RAUM/bin/npm" ;;
  esac
  chmod +x "$RAUM/bin/cyclonedx-gomod" "$RAUM/bin/npm"
}

# lauf — fährt die Stücklisten im Raum; gibt den Ausgang.
lauf() {
  rm -rf "$RAUM/aus"
  mkdir -p "$RAUM/aus"
  (cd "$RAUM" && PATH="$RAUM/bin:$PATH" stuecklisten_schreiben "$RAUM/aus") >/dev/null 2>&1
}

istCycloneDXJSON() {
  python3 -I -c 'import json,sys; b=json.load(open(sys.argv[1])); sys.exit(0 if b.get("bomFormat")=="CycloneDX" else 1)' "$1" 2>/dev/null
}

attrappen json
lauf
status=$?
behauptung "der Lauf gelingt" "$([ $status -eq 0 ] && echo ja || echo nein)"
behauptung "stueckliste-go.json ist CycloneDX in JSON" "$(istCycloneDXJSON "$RAUM/aus/stueckliste-go.json" && echo ja || echo nein)"
behauptung "stueckliste-npm.json ist CycloneDX in JSON" "$(istCycloneDXJSON "$RAUM/aus/stueckliste-npm.json" && echo ja || echo nein)"

attrappen xml
lauf
status=$?
behauptung "eine npm-Stückliste in XML beendet die Freigabe" "$([ $status -ne 0 ] && echo ja || echo nein)"
behauptung "und bleibt nicht als .json liegen" "$([ ! -e "$RAUM/aus/stueckliste-npm.json" ] && echo ja || echo nein)"

attrappen leer
lauf
status=$?
behauptung "ein gescheitertes npm sbom beendet die Freigabe" "$([ $status -ne 0 ] && echo ja || echo nein)"
behauptung "und hinterlässt keine Zwischendatei" "$([ ! -e "$RAUM/aus/.npm-sbom" ] && echo ja || echo nein)"

exit "$FEHLER"
