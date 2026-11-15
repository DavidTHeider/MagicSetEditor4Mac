#!/usr/bin/env bash
#
# Prüfungen der Ortswache um die Go-Läufe — A01
# (#2215).
#
# **Der Anlass**: der Audit zu 0.8.0 sah während `make vorpruefung` zwei
# Drosseldateien und einen Scryfall-Eintrag im **echten** Anwendungsverzeichnis
# wachsen. `scripts/ci.sh` gab den Go-Läufen seit #891 ein leeres Zuhause,
# die Vorprüfung nicht — zwei Pipelines (§4). `scripts/echteorte.sh` ist die
# Wache, die beide jetzt teilen; hier wird sie an einem **falschen** echten
# Zuhause gefahren, damit kein Lauf das richtige anfasst.
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$wurzel"

# shellcheck source=scripts/echteorte.sh
source scripts/echteorte.sh

raum="$(mktemp -d)"
trap 'rm -rf "$raum"' EXIT

export ECHTES_ZUHAUSE="$raum/echt"
mkdir -p "$ECHTES_ZUHAUSE/Library/Application Support/MagicSetEditor4Mac/drosseln"
printf 'x' >"$ECHTES_ZUHAUSE/Library/Application Support/MagicSetEditor4Mac/drosseln/a.drossel"

fehler=0
pruefe() {
  if [ "$2" = "$3" ]; then
    printf 'ok   %s\n' "$1"
  else
    printf 'FEHL %s — erwartet %s, bekommen %s\n' "$1" "$3" "$2"
    fehler=1
  fi
}

# **Ohne Änderung ist der Vergleich grün.**
echte_orte_merken "$raum/vorher.json"
status=0
echte_orte_pruefen "$raum/vorher.json" >"$raum/ausgabe" || status=$?
pruefe "unberührte Orte geben 0" "$status" 0

# **Eine Datei tief im Anwendungsverzeichnis macht ihn rot** — und er nennt sie.
sleep 0.01
printf 'xx' >"$ECHTES_ZUHAUSE/Library/Application Support/MagicSetEditor4Mac/drosseln/a.drossel"
status=0
echte_orte_pruefen "$raum/vorher.json" >"$raum/ausgabe" || status=$?
pruefe "eine veränderte Drosseldatei gibt 1" "$status" 1
pruefe "die Ausgabe nennt sie" "$(grep -c 'a.drossel' "$raum/ausgabe")" 1

# **Und die Vorgaben unter der Kennung des Bündels genauso** — sie folgen
# nicht `HOME` (Befund F-02 des Audits).
echte_orte_merken "$raum/vorher.json"
mkdir -p "$ECHTES_ZUHAUSE/Library/Preferences"
printf 'p' >"$ECHTES_ZUHAUSE/Library/Preferences/com.wails.magicseteditor4mac.plist"
status=0
echte_orte_pruefen "$raum/vorher.json" >"$raum/ausgabe" || status=$?
pruefe "eine neue Vorgabendatei gibt 1" "$status" 1

# **Das Zuhause der Läufe ist ein anderes als das echte.**
heim="$(heim_fuer_laeufe)"
lage=aussen
if [[ "$heim" == "$ECHTES_ZUHAUSE"* ]]; then lage=innen; fi
pruefe "das Laufzuhause liegt nicht im echten" "$lage" aussen
pruefe "das Laufzuhause ist ein Verzeichnis" "$([ -d "$heim" ] && echo ja || echo nein)" ja
rm -rf "$heim"

exit "$fehler"
