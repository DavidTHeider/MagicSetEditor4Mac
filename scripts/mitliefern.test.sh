#!/usr/bin/env bash
#
# Prüfungen des Beigabenschritts aus scripts/mitliefern.sh
# (#702).
#
# Sie laufen gegen **erfundene** Bündel und ein **erfundenes** Pack in einem
# Wegwerfverzeichnis — derselbe Grund wie bei `saat.test.sh`: das echte Pack
# wiegt 805 MB, und ein Lauf dagegen prüfte die Platte dieses Rechners statt
# des Skripts.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKRIPT="$WURZEL/scripts/mitliefern.sh"

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

buendel() { # buendel <verzeichnis> — legt ein signierbares .app an
  local ort="$1/Probe.app/Contents"
  mkdir -p "$ort/MacOS"

  cp /bin/echo "$ort/MacOS/Probe"

  cat > "$ort/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>CFBundleName</key><string>Probe</string>
  <key>CFBundleExecutable</key><string>Probe</string>
  <key>CFBundleIdentifier</key><string>probe.mitliefern.test</string>
</dict></plist>
PLIST

  codesign --force --deep --sign - "$1/Probe.app" >/dev/null 2>&1

  printf '%s\n' "$ort/MacOS/Probe"
}

pack() { # pack <verzeichnis> — ein erfundenes Pack mit allen erwarteten Paketen
  local ort="$1"

  for paket in magic.mse-game magic-m15.mse-style magic-m15-vanguard.mse-style \
    magic-modules.mse-include magic-mana-large.mse-symbol-font \
    magic-mana-small.mse-symbol-font magic-blends.mse-include \
    magic-default-image.mse-include; do
    mkdir -p "$ort/$paket"
    printf 'mse version: 2.0.0\n' > "$ort/$paket/inhalt"
  done
  # **Die Schriften liegen neben `data/`, nicht darin** — wie im
  # `M15-Magic-Pack` (#2071).
  local schriften
  schriften="$(dirname "$ort")/Magic - Fonts"
  mkdir -p "$schriften/template specific fonts"
  printf 'ttf\n' > "$schriften/mplantin.ttf"
  printf 'ttf\n' > "$schriften/template specific fonts/eigen.ttf"
}

siegel_gilt() { # siegel_gilt <app>
  codesign --verify --deep "$1" >/dev/null 2>&1
}

da() { # da <pfad>
  [ -e "$1" ]
}

ARBEIT="$(mktemp -d)"
trap 'rm -rf "$ARBEIT"' EXIT

printf '\n── Vorlagen und Beispielset ins Bündel\n'

pack "$ARBEIT/pack"
mkdir -p "$ARBEIT/kanon"
# **Ein echtes Archiv**: das Skript packt aus, und ein Textzettel ließe den
# Fall, um den es geht, nicht nachstellen.
mkdir -p "$ARBEIT/roh/Vanguard.mse-set"
printf 'mse version: 2.0.0\n' > "$ARBEIT/roh/Vanguard.mse-set/set"
printf 'bild\n' > "$ARBEIT/roh/Vanguard.mse-set/image0"
( cd "$ARBEIT/roh/Vanguard.mse-set" && zip -qr "$ARBEIT/kanon/Vanguard.mse-set" . )

BIN="$(buendel "$ARBEIT/gut")"
export MSE_PACK="$ARBEIT/pack" MSE_KANON="$ARBEIT/kanon"

behauptung "das erfundene Bündel ist zu Anfang gültig signiert" 0 \
  siegel_gilt "$ARBEIT/gut/Probe.app"

behauptung "die Beigaben landen im Bündel" 0 bash "$SKRIPT" "$BIN"
behauptung "und das Bündel ist danach wieder gültig signiert" 0 \
  siegel_gilt "$ARBEIT/gut/Probe.app"
behauptung "das Spielpaket liegt unter Resources/vorlagen" 0 \
  da "$ARBEIT/gut/Probe.app/Contents/Resources/vorlagen/magic.mse-game/inhalt"
behauptung "die Rahmenbilder liegen daneben" 0 \
  da "$ARBEIT/gut/Probe.app/Contents/Resources/vorlagen/magic-modules.mse-include/inhalt"
behauptung "das Beispielset liegt unter Resources/beispielsets" 0 \
  da "$ARBEIT/gut/Probe.app/Contents/Resources/beispielsets/Vanguard.mse-set"

# **Ausgepackt und nicht als Archiv**: ein Archiv wird nicht zurückgeschrieben
# ([0018](../docs/entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md)),
# und ein mitgeliefertes Set, das sich nicht sichern lässt, wäre eine Einladung
# in eine Sackgasse.
behauptung "und es ist ausgepackt" 0 \
  da "$ARBEIT/gut/Probe.app/Contents/Resources/beispielsets/Vanguard.mse-set/set"
# **Die Schriften gehen mit** (#2071): ohne sie trägt eine Karte Rahmen und
# Bild, aber keinen Text. Gesucht wird unter `Resources/vorlagen` über den
# ganzen Baum — dorthin also.
behauptung "die Schriften des Packs liegen unter Resources/vorlagen" 0 \
  da "$ARBEIT/gut/Probe.app/Contents/Resources/vorlagen/Magic - Fonts/mplantin.ttf"
behauptung "auch die vorlagenspezifischen im Unterordner" 0 \
  da "$ARBEIT/gut/Probe.app/Contents/Resources/vorlagen/Magic - Fonts/template specific fonts/eigen.ttf"

# **Ein zweiter Lauf räumt zuerst** — aus demselben Grund wie bei der Saat: ein
# Paket, das aus dem Pack verschwunden ist, bliebe sonst im Bündel liegen.
mkdir -p "$ARBEIT/gut/Probe.app/Contents/Resources/vorlagen/veraltet.mse-style"
printf 'alt\n' > "$ARBEIT/gut/Probe.app/Contents/Resources/vorlagen/veraltet.mse-style/style"
behauptung "ein zweiter Lauf geht durch" 0 bash "$SKRIPT" "$BIN"
behauptung "eine alte Beigabe wird geräumt" 1 \
  da "$ARBEIT/gut/Probe.app/Contents/Resources/vorlagen/veraltet.mse-style/style"

# ---------------------------------------------------- das Set ist die Kür

# **Ohne Beispielset geht die Freigabe weiter.** Vorlagen ohne Set sind
# unangenehm, aber brauchbar; ein Abbruch wäre die härtere Antwort, als der
# Fall verdient.
BIN2="$(buendel "$ARBEIT/ohneset")"
MSE_KANON="$ARBEIT/leer" behauptung "ohne Beispielset läuft es durch" 0 \
  bash "$SKRIPT" "$BIN2"
behauptung "und legt keines an" 1 \
  da "$ARBEIT/ohneset/Probe.app/Contents/Resources/beispielsets/Vanguard.mse-set"

# ------------------------------------------------------------ die Absagen

behauptung "ohne Pfad zum Programm bricht es ab" 1 bash "$SKRIPT"

mkdir -p "$ARBEIT/nackt"
printf 'kein Bündel\n' > "$ARBEIT/nackt/programm"
behauptung "außerhalb eines Bündels bricht es ab" 1 \
  bash "$SKRIPT" "$ARBEIT/nackt/programm"

behauptung "auf ein Programm, das es nicht gibt, bricht es ab" 1 \
  bash "$SKRIPT" "$ARBEIT/gibtsnicht/Contents/MacOS/X"

# **Ohne Pack keine Freigabe.** Ein Bündel ohne Vorlagen sieht aus wie eines mit
# und zeichnet nichts — das fiele erst dem auf, der es lädt.
BIN3="$(buendel "$ARBEIT/ohnepack")"
MSE_PACK="$ARBEIT/gibtsnicht" behauptung "ohne bezogenes Pack bricht es ab" 1 \
  bash "$SKRIPT" "$BIN3"

# **Und ein unvollständiges Pack ebenso.** Ein fehlendes Paket zeichnet eine
# leere Fläche, und das sähe aus wie ein Fehler der Anwendung.
rm -rf "$ARBEIT/pack/magic-modules.mse-include"
BIN4="$(buendel "$ARBEIT/luecke")"
behauptung "ein Pack ohne die Rahmenbilder bricht ab" 1 bash "$SKRIPT" "$BIN4"
# **Und eines ohne Schriften auch** (#2071) — eine Karte ohne Text sähe aus
# wie ein Fehler der Anwendung, und so stand es in 0.5.0 und im ersten
# Abbild von 0.6.0.
pack "$ARBEIT/pack2"
BIN5="$(buendel "$ARBEIT/ohneschriften")"
MSE_PACK="$ARBEIT/pack2" MSE_SCHRIFTEN="$ARBEIT/gibtsnicht" \
  behauptung "ein Pack ohne Schriften bricht ab" 1 bash "$SKRIPT" "$BIN5"

exit "$FEHLER"
