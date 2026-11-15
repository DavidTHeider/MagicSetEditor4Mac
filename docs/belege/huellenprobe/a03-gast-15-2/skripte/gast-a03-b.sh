#!/bin/bash
# A03-Gegenprobe im Gast: Abnahmebau starten, über die Leitung den MCP-Schlüssel
# holen, dieselben Messungen wie das Drehbuch per curl fahren.
set -u
T="/Volumes/My Shared Files/gast-teilen"
A="$T/a03"; E="$T/a03-ergebnis-b"; mkdir -p "$E"
ORT=$(mktemp -d /tmp/huelle-XXXXXX); mkdir -p "$ORT/zuhause"
BUND="$ORT/bund.keychain-db"
security create-keychain -p leer "$BUND" && security unlock-keychain -p leer "$BUND" && security set-keychain-settings "$BUND"
APP="$ORT/Abnahme-B.app"; ditto "$T/Abnahme-B.app" "$APP"
SOCK="$ORT/leitung.sock"; PORT=9433
open -n "$APP" --env HOME="$ORT/zuhause" --env MSE_BESTAND="$ORT/bestand.db" --env MSE_SCHLUESSELBUND="$BUND" \
  --env MSE_ABNAHME="$SOCK" --env WAILS_MCP_PORT=$PORT --env WAILS_MCP_HIDE_CURSOR=1
for i in $(seq 1 60); do [ -S "$SOCK" ] && break; sleep 1; done
frage() { printf '%s\n' "$1" | nc -U -w 5 "$SOCK" | head -1; }
sleep 3
ROH=$(frage '{"was":"schluessel"}'); echo "$ROH" > "$E/schluessel-roh.txt"
SCHLUESSEL=$(printf '%s' "$ROH" | sed -E 's/.*"daten":"([^"]*)".*/\1/')
echo "schluessel-laenge ${#SCHLUESSEL}" >> "$E/schluessel-roh.txt"
ruf() { curl -s -w "\\nHTTP %{http_code}" -m 30 -H 'content-type: application/json' -H "authorization: Bearer $SCHLUESSEL" --data-binary @"$A/$1.json" "http://127.0.0.1:$PORT/mcp" > "$E/$2.json"; }
for i in $(seq 1 60); do ruf lage lage; grep -q '\\"einfuehrung\\":true' "$E/lage.json" && break; sleep 1; done
sw_vers > "$E/sw_vers.txt"
ruf einfuehrung-texte mit-einfuehrung-texte; ruf einfuehrung-kanten mit-einfuehrung-kanten; ruf einfuehrung-felder mit-einfuehrung-felder; ruf einfuehrung-symbole mit-einfuehrung-symbole
frage '{"was":"taste","taste":"Escape","mit":[]}' > "$E/taste1.txt"; sleep 1
frage '{"was":"taste","taste":"k","mit":["befehl"]}' > "$E/taste2.txt"; sleep 2
ruf lage lage-palette
ruf palette-texte mit-palette-texte; ruf palette-kanten mit-palette-kanten; ruf palette-felder mit-palette-felder; ruf palette-symbole mit-palette-symbole
ruf ohne-korrektur ohne-korrektur; sleep 1
ruf palette-texte ohne-palette-texte; ruf palette-kanten ohne-palette-kanten; ruf palette-felder ohne-palette-felder
screencapture -x "$E/bildschirm.png" 2>/dev/null
frage '{"was":"beenden"}' > "$E/ende.txt"
sleep 2; pkill -f "$APP/Contents/MacOS" 2>/dev/null
security delete-keychain "$BUND"
echo fertig
