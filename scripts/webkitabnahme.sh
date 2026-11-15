#!/usr/bin/env bash
#
# **Die WebKit-Abnahmen fahren und belegen** —
# #1828.
#
#     scripts/webkitabnahme.sh                   # alle Drehbücher, mit Übersicht
#     scripts/webkitabnahme.sh ux05-assistent …  # nur diese, ohne Übersicht
#
# Baut die Oberfläche und die Anwendung im Servermodus (`scripts/bauen.sh
# browser`, die Oberfläche **eingebettet** wie im Bündel und nicht von Vite),
# legt einen frischen Prüfbestand an, startet die Anwendung darüber auf einem
# freien Platz und fährt jedes Drehbuch unter `scripts/webkitprobe/abnahme/`
# mit der WebKit-Probe. **Der Beleg** — je Drehbuch der Bericht, dazu eine
# Übersicht mit Stand und Ergebnis — landet in `docs/belege/webkitprobe/`, die
# Bilder in `.webkitabnahme/` (nicht eingecheckt).
#
# **Nicht Teil von `make ci`**: es braucht ein Fenstersystem und baut die
# Anwendung. Aufruf: `make webkitabnahme`.
#
# **Was der Prüfbestand trägt, und warum**: das Set aus der Saat (fünf
# Karten, die Karteneditor und Schubladen brauchen), drei Exemplare (dann
# steht die Sammlung als Liste da, nicht als Leerzustand), ein Deck
# und eine laufende Partie zwischen Ada und Bo (der Tisch).

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$WURZEL" || exit 1

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

ok()   { printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"; }
ende() { printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"; exit 1; }

# shellcheck source=scripts/activate.sh
source scripts/activate.sh >/dev/null 2>&1

RAUM="$(mktemp -d)"
DIENER=""
trap '[ -n "$DIENER" ] && kill "$DIENER" 2>/dev/null; rm -rf "$RAUM"' EXIT

printf '\n── Bauen\n'
( cd frontend && npm run build >/dev/null 2>&1 ) || ende "die Oberfläche ließ sich nicht bauen"
./scripts/bauen.sh browser >/dev/null 2>&1 || ende "die Anwendung ließ sich nicht bauen"
go build -o .go/bin/mse ./cmd/mse || ende "mse ließ sich nicht bauen"
ok "Oberfläche, Anwendung im Servermodus und mse gebaut"

printf '\n── Prüfbestand\n'
# **Ein eigenes Zuhause, erst nach dem Bau** — T01 (#2095). Bis zum
# 2026-10-06 trennte diese Abnahme Bestand und Bund, aber nicht `HOME`: `mse`
# und die Anwendung lasen Pakete, Einstellungen und Zwischenspeicher dieses
# Rechners. Der Bau braucht das echte (npm, Swift), die Probe der WebKit-Seite
# auch — sie ist der Gast, nicht die Anwendung. Ob die echten Orte unberührt
# blieben, sagt der Abdruck am Ende.
ECHT="$HOME"
node scripts/huellenorte.mjs abdruck "$ECHT" > "$RAUM/orte-vorher.json" || ende "die echten Orte ließen sich nicht abdrucken"
mkdir -p "$RAUM/zuhause"
export HOME="$RAUM/zuhause"
# **Ein zweites Spiel, selbst geschrieben** — WK04 (#2105): mit der Saat allein
# sieht der Setassistent ein Spiel und zeigt zwei Schritte statt drei.
PAKETE="$HOME/Library/Application Support/MagicSetEditor4Mac/pakete"
mkdir -p "$PAKETE"
cp -R testdaten/assistent/*.mse-* "$PAKETE/" || ende "die Vorrichtung des Assistenten ließ sich nicht legen"
# **Und die Vorgaben und der Dienst dieses Rechners bleiben unberührt** (QA01,
# #2382): `defaults` folgt keinem umgelenkten `HOME`, und am 2026-10-20 zog der
# Kennungsumzug hier die echten Vorgaben in eine neue Domäne.
export MSE_BESTAND="$RAUM/bestand.db" MSE_SCHLUESSELBUND="$RAUM/bund.keychain-db" MSE_DIENST_GESPERRT=an
security create-keychain -p leer "$MSE_SCHLUESSELBUND" 2>/dev/null
security unlock-keychain -p leer "$MSE_SCHLUESSELBUND" || ende "der Prüfbund lässt sich nicht öffnen"

mse() { .go/bin/mse "$@" >/dev/null || ende "mse $*"; }

mse oeffnen saat/probe-werkstatt.mse-set
mse exemplar scryfall:aaa-1 3
mse exemplar scryfall:bbb-2 1 nearMint de
mse exemplar scryfall:ccc-3 4 nearMint ja nonfoil 2026-09-01 199 EUR
# **Zwei mit einer Scryfall-Kennung** (UX-03, #1675): ein Deck nimmt nur
# einen Druck mit UUID auf — die drei darüber zeigen, was es sagt, wenn nicht.
mse exemplar 4c1a8f6e-2b3d-4e5f-8a9b-0c1d2e3f4a5b 2
mse exemplar 7d2e9a1b-3c4d-4e5f-9a0b-1c2d3e4f5a6b 1
mse deck neu Probe-Deck probe
for karte in 1 2 3 4 5; do mse deck legen 1 "--karte=$karte" 8; done
mse partie-anlegen probe ada:Ada:1 bo:Bo:1

# **Ein Gegenüber für den Handelsassistenten** (UX-07, #1679): ohne einen
# Partner zeigt „Handeln" nur seinen leeren Zustand. Die Kennung stammt aus
# einem zweiten Bestand mit eigenem Bund — eine erfundene hielte die Prüfung
# der Kennung nicht aus, und ein geteilter Bund legte beide Schlüssel unter
# denselben Namen.
security create-keychain -p leer "$RAUM/gegenueber.keychain-db" 2>/dev/null
security unlock-keychain -p leer "$RAUM/gegenueber.keychain-db" || ende "der zweite Prüfbund lässt sich nicht öffnen"
AUSWEIS="$(MSE_BESTAND="$RAUM/gegenueber.db" MSE_SCHLUESSELBUND="$RAUM/gegenueber.keychain-db" \
  .go/bin/mse peer-anlegen)"
GEGENUEBER="$(printf '%s\n' "$AUSWEIS" | awk -F'\t' '$1 == "kennung" { print $2 }')"
ABDRUCK="$(printf '%s\n' "$AUSWEIS" | awk -F'\t' '$1 == "fingerabdruck" { print $2 }')"
[ -n "$GEGENUEBER" ] && [ -n "$ABDRUCK" ] || ende "das Gegenüber bekam keinen Ausweis"
mse peer-anlegen
mse kontakt-aufnehmen "$GEGENUEBER" Bo
# **Bestätigt**, denn gehandelt wird nur mit einem Gegenüber, dessen
# Fingerabdruck verglichen ist (N1, #1454).
mse kontakt-bestaetigen "$GEGENUEBER" "$ABDRUCK"
# **Und mitspielen darf Bo auch** (UX-08.AK-04, #1680): an einem fremden Tisch
# steht nur zur Wahl, wem „Eine Partie spielen" freigegeben ist.
mse zweck-erlauben "$GEGENUEBER" partie

# **Ein zweites Profil mit einem Deck und einer Freigabe** (UX-08.AK-04,
# #1680): Cleo legt ihr Leihdeck an und lässt den ersten Nutzer hinein —
# dann steht ihr Profil in „Ein Deck aus einem fremden Profil holen" zur
# Wahl. Gewechselt wird der Mensch, hin und zurück (`nutzer-wechseln`;
# `profil` allein ließe den ersten nur lesen); die Anwendung startet beim
# ersten.
mse nutzer-neu Cleo
mse nutzer-wechseln 2
mse deck neu Leihdeck probe
mse freigabe-setzen 2 1 leser
mse nutzer-wechseln 1

# **Ein Fensterbild für `ux09-gestaltung`** (UX-09, #1681) — dorthin, wo
# die Anwendung es ablegt (`beiwerk/nutzer/` neben dem Bestand). Über den
# Dateidialog käme es nicht herein: der gehört dem Wails-Prozess.
mkdir -p "$RAUM/beiwerk/nutzer"
python3 - "$RAUM/beiwerk/nutzer/fensterbild-1.png" <<'BILD' || ende "das Fensterbild ließ sich nicht anlegen"
import struct, sys, zlib
breite, hoehe = 96, 64
zeilen = b''.join(
    b'\x00' + b''.join(
        bytes((40 + 2 * x, 90 + 2 * y, 160 if (x // 12 + y // 12) % 2 else 60))
        for x in range(breite))
    for y in range(hoehe))
def stueck(art, daten):
    return struct.pack('>I', len(daten)) + art + daten + struct.pack('>I', zlib.crc32(art + daten))
with open(sys.argv[1], 'wb') as datei:
    datei.write(b'\x89PNG\r\n\x1a\n'
                + stueck(b'IHDR', struct.pack('>IIBBBBB', breite, hoehe, 8, 2, 0, 0, 0))
                + stueck(b'IDAT', zlib.compress(zeilen))
                + stueck(b'IEND', b''))
BILD
ok "ein Set, fünf Exemplare, ein Deck, eine Partie, ein Gegenüber, ein zweites Profil mit Freigabe, ein Fensterbild"

printf '\n── Anwendung\n'
PLATZ="$(python3 -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')"
env -u FRONTEND_DEVSERVER_URL WAILS_SERVER_PORT="$PLATZ" \
  build/bin/MagicSetEditor4Mac.app/Contents/MacOS/MagicSetEditor4Mac > "$RAUM/anwendung.log" 2>&1 &
DIENER=$!
disown "$DIENER"

for _ in $(seq 60); do
  curl -s -m 1 -o /dev/null "http://127.0.0.1:$PLATZ/" && break
  sleep 0.5
done
curl -s -m 1 -o /dev/null "http://127.0.0.1:$PLATZ/" || ende "die Anwendung antwortet nicht auf $PLATZ"
ok "läuft auf $PLATZ"

printf '\n── Drehbücher\n'
BELEGE="docs/belege/webkitprobe"
BILDER=".webkitabnahme"
mkdir -p "$BELEGE" "$BILDER"
# **Derselbe Stand wie in jedem Bericht**: die Belege zählen nicht mit, sonst
# trüge ein zweiter Lauf ein `+`, nur weil der erste sie geschrieben hat.
STAND="$(git rev-parse --short HEAD)"
[ -z "$(git status --porcelain -- . ':(exclude)docs/belege/webkitprobe')" ] || STAND="$STAND+"

zeilen=()
rot=0
# **Genannte Drehbücher allein** — `scripts/webkitabnahme.sh ux05-assistent …`.
# Ein Teillauf schreibt ihre Berichte, aber nicht die Übersicht: die sagt,
# wie **alle** am Stand liefen.
GENANNT=" $* "
for drehbuch in scripts/webkitprobe/abnahme/*.probe; do
  name="$(basename "$drehbuch" .probe)"
  if [ "$#" -gt 0 ] && [[ "$GENANNT" != *" $name "* ]]; then continue; fi
  rm -rf "${BILDER:?}/$name"
  if HOME="$ECHT" ./scripts/webkitprobe.sh "$drehbuch" --basis "http://127.0.0.1:$PLATZ/" \
       --ausgabe "$BILDER/$name" > /dev/null 2>&1; then
    ergebnis="bestanden"
    ok "$name"
  else
    ergebnis="**rot**"
    rot=1
    printf '%s  ✗ %s%s\n' "$ROT" "$name" "$AUS"
  fi
  cp "$BILDER/$name/bericht.md" "$BELEGE/$name.md" 2>/dev/null ||
    printf '(kein Bericht)\n' > "$BELEGE/$name.md"
  schritte="$(grep -cE '^- (✓|✗) Z' "$BELEGE/$name.md")"
  zeilen+=("| [\`$name\`]($name.md) | $ergebnis | $schritte |")
done

kill "$DIENER" 2>/dev/null
DIENER=""
if abweichend="$(node scripts/huellenorte.mjs vergleich "$ECHT" "$RAUM/orte-vorher.json")"; then
  ok "die echten Orte dieses Rechners blieben unberührt"
  orte="unberührt"
else
  printf '%s  ✗ echte Orte verändert:%s\n%s\n' "$ROT" "$AUS" "$abweichend"
  orte="**verändert**: $(printf '%s' "$abweichend" | tr '\n' ';')"
  rot=1
fi

if [ "$#" -gt 0 ]; then
  printf '\n%s Teillauf über %s — die Übersicht bleibt, wie sie ist\n' "$([ "$rot" -eq 0 ] && echo ✓ || echo ✗)" "$*"
  exit "$rot"
fi

# shellcheck disable=SC2016 # Backticks sind Markdown, keine Shell
{
  printf '# WebKit-Abnahmen — Übersicht\n\n'
  printf 'Geschrieben von `scripts/webkitabnahme.sh` (#1828). **Stand `%s`**, am %s, auf %s, macOS %s.\n\n' \
    "$STAND" "$(date +%F)" "$(sysctl -n hw.model)" "$(sw_vers -productVersion)"
  printf 'Die Anwendung lief im Servermodus mit **eingebetteter** Oberfläche über einem frischen Prüfbestand; jedes Drehbuch fuhr in der WebKit-Probe. **Was das nicht belegt**: Menüleiste, Fensterrahmen, native Dateidialoge und das Ablegen aus dem Finder — und VoiceOvers eigene Satzbildung (`scripts/webkitprobe/Motoren.swift`).\n\n'
  printf '| Drehbuch | Ergebnis | Schritte |\n| --- | --- | --- |\n'
  printf '%s\n' "${zeilen[@]}"
  printf '\nDie echten Orte dieses Rechners (`huellenorte.mjs`, bis zur zweiten Ebene): %s.\n' "$orte"
} > "$BELEGE/uebersicht.md"

printf '\n'
if [ "$rot" -eq 0 ]; then
  printf '%s✓ alle WebKit-Abnahmen bestanden — Beleg in %s%s\n' "$GRUEN" "$BELEGE" "$AUS"
else
  printf '%s✗ WebKit-Abnahmen rot — Beleg in %s%s\n' "$ROT" "$BELEGE" "$AUS"
fi

exit "$rot"
