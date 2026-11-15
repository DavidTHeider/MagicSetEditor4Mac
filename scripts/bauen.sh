#!/usr/bin/env bash
#
# Der Bau des `.app` — ENG-02 (#1727).
#
#     ./scripts/bauen.sh              # der Freigabebau: build/bin/MagicSetEditor4Mac.app
#     ./scripts/bauen.sh entwicklung  # dasselbe Bündel ohne `production` (`make dev`)
#     ./scripts/bauen.sh browser      # mit `server` statt eines Fensters (Sichtprüfung)
#     ./scripts/bauen.sh abnahme      # mit der Abnahmeleitung und Wails' MCP (Hüllenprobe)
#
# ## Warum ein eigenes Skript und nicht `wails3 build`
#
# Unter Wails 2 legte `wails build` das Bündel: Oberfläche bauen, Go bauen,
# `Info.plist` aus einer Vorlage, Symbole umrechnen, `postBuildHooks`. Wails 3
# tut das über eine Taskfile-Kette mit eigenem Projektaufbau (`build/config.yml`,
# `Taskfile.yml`), und wer sie übernähme, übernähme einen zweiten Bauweg samt
# Werkzeug. Die Schritte sind fünf, und jeder steht hier — sichtbar und ohne
# weiteres Werkzeug.
#
# ## Die Bauschalter
#
# **Der Freigabebau trägt `production` und sonst nichts** (ENG-02.AK-05): ohne
# ihn trägt der Bau die Entwicklerwerkzeuge der Webansicht, mit `mcp` einen
# MCP-Server von Wails selbst, mit `server` einen HTTP-Server statt eines
# Fensters. Das zweite und dritte sind **Zugänge**, die kein Freigabebau haben
# darf; `scripts/release.sh` prüft das am fertigen Programm (`go version -m`).
#
# **Die zwei anderen Arten sind Werkzeuge der Entwicklung** und legen dasselbe
# Bündel samt Saat — sonst startete die Anwendung dort ohne Vorlagen, und die
# Sichtprüfung sähe eine andere Anwendung als die ausgelieferte. Die Oberfläche
# bauen sie nicht: sie kommt von Vite (`scripts/entwickeln.sh`).

set -euo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$WURZEL"

# **Die gepinnte Umgebung, auch ohne `source` davor** (§1) — sonst baute der
# Go des Rechners, und das Bündel trüge eine andere Fassung als die Pipeline.
# shellcheck source=/dev/null
source scripts/activate.sh >/dev/null 2>&1

ROT=$'\033[31m'; GRUEN=$'\033[32m'; AUS=$'\033[0m'
[ -t 1 ] || { ROT=''; GRUEN=''; AUS=''; }
ok()   { printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"; }
ende() { printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"; exit 1; }

ART="${1:-freigabe}"
case "$ART" in
  freigabe)    SCHALTER="production" ;;
  entwicklung) SCHALTER="" ;;
  browser)     SCHALTER="server" ;;
  # **Die Abnahme am nativen Fenster** (ENG-02.AK-04): die Leitung aus
  # `internal/fenster/abnahme.go` und Wails' eigener MCP-Server für die Seite.
  # Beides sind Zugänge — `scripts/releaseschalter.sh` weist sie ab.
  abnahme)     SCHALTER="abnahme,mcp" ;;
  *)           ende "unbekannte Art: $ART — freigabe, entwicklung, browser oder abnahme" ;;
esac

NAME="MagicSetEditor4Mac"
APP="build/bin/$NAME.app"
INHALT="$APP/Contents"

# **Die Fassung steht in `wails.json`** — dort lesen sie `scripts/releaseversion.sh`
# und der Wächter über `app.Fassung`; ein zweiter Ort liefe weg.
FASSUNG="$(sed -n 's/.*"productVersion"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' wails.json | head -1)"
[ -n "$FASSUNG" ] || ende "wails.json nennt keine productVersion"

# **Der Stand, aus dem ein Abnahmebündel entsteht** — T06 (#2099). Vor dem
# Bau gerechnet: was sich währenddessen ändert, gehört nicht dazu. Die
# Hüllenprobe hält ihn gegen den Baum, bevor sie misst (`scripts/baustand.mjs`).
if [ "$ART" = abnahme ]; then
  ABDRUCK="$(node scripts/baustand.mjs abdruck)" || ende "der Bauabdruck ließ sich nicht rechnen"
fi

# ------------------------------------------------------------------ Oberfläche

# **Beim Entwickeln nur, wenn es sie noch nicht gibt**: `go:embed` verlangt
# ein `index.html`, aber gezeigt wird, was Vite liefert. Und **kein `npm ci`**
# außerhalb der Freigabe — es löscht `node_modules` und legt es neu an.
if [ "$ART" = freigabe ]; then
  ( cd frontend && npm ci --no-audit --no-fund >/dev/null && npm run build >/dev/null ) \
    || ende "die Oberfläche ließ sich nicht bauen"
  ok "Oberfläche gebaut (frontend/dist)"
elif [ "$ART" = abnahme ]; then
  # **Ein Abnahmebündel misst die Oberfläche, die im Baum steht** — ohne
  # Neubau trug es das `dist` vom ersten Bau, und eine Hüllenprobe maß am
  # 2026-10-02 die Oberfläche von vor dem Umbau (UXR-01, Messung nachher).
  ( cd frontend && npm run build >/dev/null ) || ende "die Oberfläche ließ sich nicht bauen"
  ok "Oberfläche gebaut (frontend/dist) — die Abnahme misst den Stand im Baum"
elif [ ! -f frontend/dist/index.html ]; then
  ( cd frontend && npm run build >/dev/null ) || ende "die Oberfläche ließ sich nicht bauen"
  ok "Oberfläche einmal gebaut — go:embed braucht sie"
fi

# ------------------------------------------------------------------------- Go

rm -rf "$APP"
mkdir -p "$INHALT/MacOS" "$INHALT/Resources"

# **Die macOS-Untergrenze** — dieselbe Datei wie in Pipeline und Umgebung —
# und das Ziel aus §12, ausdrücklich und nicht vom Rechner geerbt.
export CGO_ENABLED=1 GOOS=darwin GOARCH=arm64
# shellcheck source=scripts/macos.sh
. "$WURZEL/scripts/macos.sh"

# **Die Weglassschalter in jeder Art** (#2393), nicht nur in der Freigabe:
# was die Abnahme am nativen Fenster prüft, soll das Programm sein, das
# hinausgeht. Die Liste steht in `releaseschalter.sh`.
# shellcheck source=scripts/releaseschalter.sh
. "$WURZEL/scripts/releaseschalter.sh"

if [ "$ART" = freigabe ]; then
  go build -tags "${SCHALTER:+$SCHALTER,}$WEGLASSSCHALTER" -trimpath -ldflags "-s -w" -o "$INHALT/MacOS/$NAME" . \
    || ende "das Programm ließ sich nicht bauen"
else
  go build -tags "${SCHALTER:+$SCHALTER,}$WEGLASSSCHALTER" -o "$INHALT/MacOS/$NAME" . \
    || ende "das Programm ließ sich nicht bauen"
fi
ok "Programm gebaut (darwin/arm64, ab macOS $MACOSX_DEPLOYMENT_TARGET, Art $ART)"

# --------------------------------------------------------------------- Bündel

sed "s/@FASSUNG@/$FASSUNG/g" build/darwin/Info.plist > "$INHALT/Info.plist" \
  || ende "Info.plist ließ sich nicht schreiben"
# **Beim Entwickeln darf die Webansicht ins lokale Netz** — zu Vite und
# seinem Nachladen. Unter Wails 2 stand das in einer eigenen Vorlage
# (`Info.dev.plist`); der Freigabebau hat es nicht, und dort bleibt es so.
if [ "$ART" != freigabe ]; then
  plutil -insert NSAppTransportSecurity -json '{"NSAllowsLocalNetworking":true}' \
    "$INHALT/Info.plist" || ende "Info.plist ließ sich für die Entwicklung nicht ergänzen"
fi
# **Der Abnahmebau hat eine eigene Kennung** — mit der der Anwendung teilte er
# deren Vorgaben (die Erinnerung der Dateidialoge) und den Speicher ihrer
# Webansicht (`localStorage`); die Hüllenprobe öffnete einen Dialog dort, wo
# die Anwendung dieses Rechners zuletzt war. `scripts/huellenprobe.mjs` weist
# ein Bündel ohne sie ab.
if [ "$ART" = abnahme ]; then
  plutil -replace CFBundleIdentifier -string "io.github.davidtheider.magicseteditor4mac.abnahme" \
    "$INHALT/Info.plist" || ende "Info.plist ließ sich für die Abnahme nicht ergänzen"
fi
plutil -lint "$INHALT/Info.plist" >/dev/null || ende "Info.plist ist kein gültiges plist"

# **Die Sätze des Bündels auf Englisch** (#1996) — die Entwicklungsregion ist
# Deutsch, und ohne diese Datei sähe ein englisches System die Rückfrage
# „Lokales Netzwerk" auf Deutsch.
cp -R build/darwin/en.lproj "$INHALT/Resources/" || ende "en.lproj ließ sich nicht legen"
printf 'APPL????' > "$INHALT/PkgInfo"

# symbole rechnet ein PNG in ein .icns um — dieselben Größen, die `iconutil`
# erwartet.
symbole() {
  local quelle="$1" ziel="$2" satz
  satz="$(mktemp -d)/symbol.iconset"
  mkdir -p "$satz"

  for kante in 16 32 128 256 512; do
    sips -z "$kante" "$kante" "$quelle" --out "$satz/icon_${kante}x${kante}.png" >/dev/null
    sips -z "$((kante * 2))" "$((kante * 2))" "$quelle" \
      --out "$satz/icon_${kante}x${kante}@2x.png" >/dev/null
  done

  iconutil -c icns "$satz" -o "$ziel"
  rm -rf "$(dirname "$satz")"
}

symbole build/appicon.png "$INHALT/Resources/iconfile.icns" || ende "das Programmsymbol"
symbole build/dokument.png "$INHALT/Resources/dokument.icns" || ende "das Dokumentsymbol"
ok "Bündel gelegt: $APP (Fassung $FASSUNG)"

if [ "$ART" = abnahme ]; then
  printf '{"head":"%s","abdruck":"%s","art":"%s","schalter":"%s","gebaut":"%s"}\n' \
    "$(git rev-parse HEAD)" "$ABDRUCK" "$ART" "$SCHALTER" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    > "$INHALT/Resources/baustand.json" || ende "der Baustand ließ sich nicht legen"
  ok "Baustand gelegt: ${ABDRUCK:0:12}"
fi

# ----------------------------------------------------------------------- Saat

# **Der frühere `postBuildHook`** — die Saat gehört ins Bündel, sonst startet
# die Anwendung ins Leere (`scripts/saat.sh`).
"$WURZEL/scripts/saat.sh" "$INHALT/MacOS/$NAME" || ende "die Saat"

# ------------------------------------------------------------------ Werkzeuge

# **Im Abnahmebau auch die drei Werkzeuge** — `RESTStarten` und `MCPStarten`
# suchen `mse-rest` und `mse-mcp` neben dem Programm, und die
# Ein-Klick-Einrichtung (G08, #2272) braucht sie am nativen Fenster. Bis zum
# Vertragsrundgang (B01) lagen sie nur im Freigabebau, und beide Bindungen
# lehnten in der Probe fachlich ab.
if [ "$ART" = abnahme ]; then
  "$WURZEL/scripts/werkzeuge.sh" "$APP" >/dev/null || ende "die Werkzeuge ließen sich nicht ins Bündel legen"
  ok "Werkzeuge im Bündel: mse, mse-mcp, mse-rest"
fi

# ------------------------------------------------------------------- Siegel

# **Zuletzt versiegeln, wie `scripts/release.sh`** — gefunden bei I03 (#2261):
# die Werkzeuge kamen nach dem Siegel ins Bündel, `codesign --verify` sagte
# „a sealed resource is missing or invalid", und Gatekeeper prüfte im Gast
# ein beschädigtes Bündel statt eines ad hoc signierten. Jede Datei des
# Bündels liegt jetzt; danach ändert dieses Skript nichts mehr daran.
#
# **Mit Hardened Runtime und ohne Entitlements** (SEC03, #2356, Vermerk
# 0195) — wie `scripts/release.sh`, damit jede Hüllenprobe das Bündel so
# fährt, wie es ausgeliefert wird.
codesign --force --deep --options runtime --sign - "$APP" >/dev/null 2>&1 || ende "das Bündel ließ sich nicht versiegeln"
codesign --verify --deep --strict "$APP" >/dev/null 2>&1 || ende "das Siegel des Bündels hält nicht"
ok "Bündel versiegelt (ad hoc)"
