#!/usr/bin/env bash
#
# Prüfungen des Bündelschritts aus scripts/saat.sh.
#
# Sie laufen gegen **erfundene** Bündel in einem Wegwerfverzeichnis — derselbe
# Grund wie bei release.test.sh: ein echtes `.app` zu bauen, die Zeile
# herauszunehmen und alles zurückzunehmen ließe den Beleg danach wieder fort.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKRIPT="$WURZEL/scripts/saat.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0

behauptung() { # behauptung <beschreibung> <erwarteter-exit> <kommando…>
  local beschreibung="$1" erwartet="$2"
  shift 2

  local ausgabe status
  ausgabe="$("$@" 2>&1)"
  status=$?

  if [ "$status" -eq "$erwartet" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$beschreibung" "$AUS"
  else
    printf '%s  ✗ %s — Exit %d, erwartet war %d%s\n' "$ROT" "$beschreibung" "$status" "$erwartet" "$AUS"
    printf '      %s\n' "$ausgabe"
    FEHLER=1
  fi
}

buendel() { # buendel <verzeichnis> — legt ein signierbares .app an und gibt den Binärpfad
  local ort="$1/Probe.app/Contents"
  mkdir -p "$ort/MacOS"

  # **Ein echtes Mach-O und kein Textzettel.** `codesign` signiert keine
  # Zeichenkette, und ohne Signatur ließe sich der Fall, um den es hier geht —
  # das Siegel bricht beim Nachlegen von Dateien —, nicht nachstellen.
  cp /bin/echo "$ort/MacOS/Probe"

  cat > "$ort/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>CFBundleName</key><string>Probe</string>
  <key>CFBundleExecutable</key><string>Probe</string>
  <key>CFBundleIdentifier</key><string>probe.saat.test</string>
</dict></plist>
PLIST

  codesign --force --deep --sign - "$1/Probe.app" >/dev/null 2>&1

  printf '%s\n' "$ort/MacOS/Probe"
}

siegel_gilt() { # siegel_gilt <app> — 0, wenn die Signatur die Ressourcen deckt
  codesign --verify --deep "$1" >/dev/null 2>&1
}

da() { # da <pfad> — 0, wenn es die Datei gibt
  [ -e "$1" ]
}

ARBEIT="$(mktemp -d)"
trap 'rm -rf "$ARBEIT"' EXIT

printf '\n── Die Saat ins Bündel\n'

# ---------------------------------------------------------------- der Weg

BIN="$(buendel "$ARBEIT/gut")"
behauptung "das erfundene Bündel ist zu Anfang gültig signiert" 0 \
  siegel_gilt "$ARBEIT/gut/Probe.app"

# **Die Gegenprobe zum Nachsignieren, und sie steht vor der Zusage.** Wails
# signiert *vor* dem postBuildHook; wer danach eine Datei in `Resources` legt,
# bricht das Siegel. Ohne diesen Fall stünde unten eine gültige Signatur, von
# der niemand wüsste, ob sie je in Gefahr war.
mkdir -p "$ARBEIT/gut/Probe.app/Contents/Resources"
printf 'nachgelegt\n' > "$ARBEIT/gut/Probe.app/Contents/Resources/dazwischen"
behauptung "eine nachgelegte Datei bricht das Siegel" 1 \
  siegel_gilt "$ARBEIT/gut/Probe.app"
rm "$ARBEIT/gut/Probe.app/Contents/Resources/dazwischen"

behauptung "die Saat wird ins Bündel gelegt" 0 bash "$SKRIPT" "$BIN"
behauptung "und das Bündel ist danach wieder gültig signiert" 0 \
  siegel_gilt "$ARBEIT/gut/Probe.app"
behauptung "das Spielpaket liegt unter Resources/saat" 0 \
  da "$ARBEIT/gut/Probe.app/Contents/Resources/saat/probe.mse-game/game"
behauptung "die selbst gezeichnete Grafik kommt mit" 0 \
  da "$ARBEIT/gut/Probe.app/Contents/Resources/saat/probe-bild.mse-style/rahmen.png"

# **Ein zweiter Lauf räumt zuerst.** Ein Paket, das aus dem Repository
# verschwunden ist, bliebe sonst im Bündel liegen — und die gebaute Anwendung
# zeigte einen Stil, den es nicht mehr gibt.
mkdir -p "$ARBEIT/gut/Probe.app/Contents/Resources/saat/veraltet.mse-style"
printf 'alt\n' > "$ARBEIT/gut/Probe.app/Contents/Resources/saat/veraltet.mse-style/style"
behauptung "ein zweiter Lauf geht durch" 0 bash "$SKRIPT" "$BIN"
behauptung "eine alte Saat im Bündel wird geräumt" 1 \
  da "$ARBEIT/gut/Probe.app/Contents/Resources/saat/veraltet.mse-style/style"

# ------------------------------------------------------------ die Absagen

behauptung "ohne Pfad zum Programm bricht es ab" 1 bash "$SKRIPT"

mkdir -p "$ARBEIT/nackt"
printf 'kein Bündel\n' > "$ARBEIT/nackt/Probe"
behauptung "ein Programm außerhalb eines Bündels wird abgewiesen" 1 \
  bash "$SKRIPT" "$ARBEIT/nackt/Probe"

# **Ohne Quelle keine stille Null.** Ein Repository ohne `saat/` soll rot
# werden und nicht ein Bündel mit leerem Resources-Verzeichnis erzeugen.
mkdir -p "$ARBEIT/ohnesaat/scripts"
cp "$SKRIPT" "$ARBEIT/ohnesaat/scripts/saat.sh"
LEER_BIN="$(buendel "$ARBEIT/ohnesaat-ziel")"
behauptung "ohne saat/ im Repository bricht es ab" 1 \
  bash "$ARBEIT/ohnesaat/scripts/saat.sh" "$LEER_BIN"
behauptung "und legt kein leeres Verzeichnis an" 1 \
  da "$ARBEIT/ohnesaat-ziel/Probe.app/Contents/Resources/saat"

printf '\n'

if [ "$FEHLER" -ne 0 ]; then
  printf '%sSkriptprüfung rot%s\n' "$ROT" "$AUS"
  exit 1
fi

printf '%sSkriptprüfung grün%s\n' "$GRUEN" "$AUS"
