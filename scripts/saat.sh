#!/usr/bin/env bash
#
# Die Saat ins Bündel (#58).
#
#     ./scripts/saat.sh build/bin/MagicSetEditor4Mac.app/Contents/MacOS/MagicSetEditor4Mac
#
# ## Warum aus dem Bauskript und nicht aus dem Makefile
#
# Der Aufruf steht in `scripts/bauen.sh`, und über das bauen Freigabe,
# `make bau` und `make dev` gleichermaßen (ENG-02 — unter Wails 2 war es ein
# `postBuildHook` in `wails.json`). Zwei Kopien dieses Schrittes wären genau
# der Fall aus Phase 1: dieselbe Zusage an zwei Orten läuft auseinander, und
# jede Datei stimmt für sich. `internal/werkzeuge/saatbuendel_test.go` hält,
# dass es bei einer bleibt.
#
# ## Was er nicht tut
#
# Er prüft die Saat nicht. Ob ein Paket lesbar ist, beantwortet `msepaket` beim
# Öffnen; hier wird kopiert. Und er weiß nichts über bezogene Pakete — die
# liegen unter `~/Library/Application Support/`, nicht im Bündel (§11).

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
QUELLE="$WURZEL/saat"

ROT=$'\033[31m'; GRUEN=$'\033[32m'; AUS=$'\033[0m'
[ -t 1 ] || { ROT=''; GRUEN=''; AUS=''; }

ende() { printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"; exit 1; }

BIN="${1:-}"
[ -n "$BIN" ] || ende "ohne Pfad zum gebauten Programm ist nicht zu sagen, in welches Bündel die Saat gehört"

# **Fail closed statt still danebenlegen.** Ohne diese Prüfung schriebe ein
# Aufruf mit einem Pfad außerhalb eines Bündels die Saat irgendwohin — und die
# gebaute Anwendung fände sie nicht, ohne dass jemand einen Fehler sähe.
MACOS="$(cd "$(dirname "$BIN")" 2>/dev/null && pwd)" \
  || ende "das Programm $BIN gibt es nicht"
CONTENTS="$(dirname "$MACOS")"

[ "$(basename "$MACOS")" = "MacOS" ] && [ "$(basename "$CONTENTS")" = "Contents" ] \
  || ende "$BIN liegt in keinem .app-Bündel — erwartet ist …/X.app/Contents/MacOS/X"

[ -d "$QUELLE" ] || ende "$QUELLE gibt es nicht — ohne Saat startet die Anwendung ins Leere"

ZIEL="$CONTENTS/Resources/saat"

# **Erst räumen, dann kopieren.** Ein Paket, das aus dem Repository verschwunden
# ist, bliebe sonst im Bündel liegen, und die Anwendung böte einen Stil an, den
# es nicht mehr gibt.
rm -rf "$ZIEL" || ende "die alte Saat im Bündel ließ sich nicht räumen"
mkdir -p "$(dirname "$ZIEL")" || ende "Resources anlegen"
cp -R "$QUELLE" "$ZIEL" || ende "die Saat ließ sich nicht ins Bündel legen"

# **Und neu signieren — hier und nicht im Aufrufer.**
#
# Wails signiert **vor** dem postBuildHook („Self-signing application", dann
# „Executing post build hook"). Jede Datei, die danach unter `Resources` landet,
# bricht das Siegel: `codesign --verify` meldet „a sealed resource is missing or
# invalid". Auf arm64 ist die Ad-hoc-Signatur keine Wahl, sondern Voraussetzung
# ([0003](../docs/entscheidungen/0003-keine-notarisierung.md)), und der
# Erststartweg über Gatekeeper setzt sie voraus.
#
# Wer das dem Aufrufer überließe, hätte die halbe Zusage (§10): `make bau`
# signiert danach ohnehin noch einmal, `make dev` nicht.
BUENDEL="$(dirname "$CONTENTS")"
codesign --force --deep --options runtime --sign - "$BUENDEL" >/dev/null 2>&1 \
  || ende "das Bündel ließ sich nach dem Nachlegen der Saat nicht neu signieren"

printf '%s  ✓ Saat im Bündel: %s%s\n' "$GRUEN" "$ZIEL" "$AUS"
