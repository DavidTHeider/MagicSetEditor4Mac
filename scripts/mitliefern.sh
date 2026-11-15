#!/usr/bin/env bash
#
# Vorlagen und ein Beispielset ins Bündel
# (#702).
#
#     ./scripts/mitliefern.sh build/bin/MagicSetEditor4Mac.app/Contents/MacOS/MagicSetEditor4Mac
#
# ## Warum nicht im postBuildHook wie die Saat
#
# **Weil es 167 MB sind.** `scripts/saat.sh` läuft bei jedem `wails dev`; wer
# dort zusätzlich das Vorlagenpack kopierte, machte aus jedem Neustart eine
# Kaffeepause. Dieser Schritt gehört in den Freigabeweg und sonst nirgendwohin.
#
# ## Woher die Dateien kommen
#
# Aus dem **bezogenen** Pack dieses Rechners und aus den kanonischen Sets — also
# von dort, wo §11 sie hinlegt. Ins Repositorium kommen sie nicht: der
# rechtliche Einwand ist mit
# [0025](../docs/entscheidungen/0025-die-vorlagen-duerfen-mitgeliefert-werden.md)
# gefallen, der zweite steht dort daneben und bleibt — 158 MB Binärkunst wären
# in jedem Klon für immer drin.
#
# ## Warum ganz und nicht ausgeschnitten
#
# Gemessen über eine Sonde im Paketleser: ein Renderlauf über **alle** 107
# Karten von Vanguard liest 60 Dateien, zusammen 1 MB. Ein Ausschnitt wäre also
# hundertmal kleiner — und seine Vollständigkeit ließe sich nicht belegen: die
# Stilskripte setzen ihre Pfade zusammen. Ein fehlendes Rahmenbild fiele erst
# bei jemandem auf, der eine Karte in einen Rahmen setzt, den die Stichprobe
# nicht hatte: sichtbar für ihn, unsichtbar für uns (AGENTS.md §10).
#
# ## Was er nicht tut
#
# Er prüft die Pakete nicht. Ob eines lesbar ist, beantwortet `msepaket` beim
# Öffnen; hier wird kopiert.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# PACK ist das bezogene Vorlagenverzeichnis dieses Rechners.
PACK="${MSE_PACK:-$HOME/Library/Application Support/MagicSetEditor4Mac/pakete/M15-Magic-Pack/data}"

# SCHRIFTEN sind die Schriften des Packs — **neben** `data/` und nicht darin,
# wie im `M15-Magic-Pack` (#2071). Ohne sie trägt eine Karte Rahmen und Bild,
# aber keinen Text: so ging 0.5.0 hinaus, und so stand das erste Abbild von
# 0.6.0 da. Die Einwilligung gilt dem vollen Pack samt `Magic - Fonts/`
# ([0025](../docs/entscheidungen/0025-die-vorlagen-duerfen-mitgeliefert-werden.md)).
SCHRIFTEN="${MSE_SCHRIFTEN:-$(dirname "$PACK")/Magic - Fonts}"

# KANON sind die kanonischen Sets.
KANON="${MSE_KANON:-$HOME/Library/Application Support/MagicSetEditor4Mac/kanonische-sets}"

# BEISPIEL ist das Set, das mitgeht.
#
# **Vanguard und nicht Archenemy**: es ist der einzige der zehn, der über
# **alle** seine Karten befundfrei zeichnet — Archenemy verlangt ein
# `archenemy.mse-style`, das im Pack nicht liegt. Und nicht eines der sechs
# Farbsets: die wiegen je über 217 MB.
BEISPIEL="${MSE_BEISPIELSET:-Vanguard.mse-set}"

# PAKETE ist der Schluss dessen, was eine M15-Karte braucht.
#
# **`magic-modules.mse-include` ist der Brocken** — 158 MB, 3.300 Dateien. Es
# trägt die Rahmen, die Textkästen und die Trims; ohne es zeichnet eine Karte
# eine leere Fläche.
PAKETE=(
  magic.mse-game
  magic-m15.mse-style
  magic-m15-vanguard.mse-style
  magic-modules.mse-include
  magic-mana-large.mse-symbol-font
  magic-mana-small.mse-symbol-font
  magic-blends.mse-include
  magic-default-image.mse-include
)

ROT=$'\033[31m'; GRUEN=$'\033[32m'; GELB=$'\033[33m'; AUS=$'\033[0m'
[ -t 1 ] || { ROT=''; GRUEN=''; GELB=''; AUS=''; }

ende() { printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"; exit 1; }
hin()  { printf '%s  ⚠ %s%s\n' "$GELB" "$1" "$AUS"; }
ok()   { printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"; }

BIN="${1:-}"
[ -n "$BIN" ] || ende "ohne Pfad zum gebauten Programm ist nicht zu sagen, in welches Bündel die Vorlagen gehören"

MACOS="$(cd "$(dirname "$BIN")" 2>/dev/null && pwd)" \
  || ende "das Programm $BIN gibt es nicht"
CONTENTS="$(dirname "$MACOS")"

[ "$(basename "$MACOS")" = "MacOS" ] && [ "$(basename "$CONTENTS")" = "Contents" ] \
  || ende "$BIN liegt in keinem .app-Bündel — erwartet ist …/X.app/Contents/MacOS/X"

# **Ohne Pack keine Freigabe.** Ein Bündel ohne Vorlagen sieht aus wie eines mit
# und zeichnet nichts; das fiele erst dem auf, der es lädt.
[ -d "$PACK" ] || ende "das bezogene Pack liegt nicht unter $PACK — erst beziehen, dann freigeben"

ZIELVORLAGEN="$CONTENTS/Resources/vorlagen"
ZIELSETS="$CONTENTS/Resources/beispielsets"

rm -rf "$ZIELVORLAGEN" "$ZIELSETS" || ende "die alten Beigaben ließen sich nicht räumen"
mkdir -p "$ZIELVORLAGEN" "$ZIELSETS" || ende "Resources anlegen"

for paket in "${PAKETE[@]}"; do
  [ -d "$PACK/$paket" ] || ende "$paket fehlt im Pack — ohne es zeichnet eine Karte nicht"

  cp -R "$PACK/$paket" "$ZIELVORLAGEN/" || ende "$paket ließ sich nicht kopieren"
done

ok "$(du -sh "$ZIELVORLAGEN" | cut -f1) Vorlagen: ${#PAKETE[@]} Pakete"

# **Unter die Vorlagen und nicht daneben**: die Anwendung sucht Schriften über
# den ganzen Baum jeder Vorlagenwurzel (`internal/schriften`, `MitPaketen`),
# und `Resources/vorlagen` ist eine davon.
[ -d "$SCHRIFTEN" ] \
  || ende "die Schriften des Packs liegen nicht unter $SCHRIFTEN — ohne sie trägt eine Karte keinen Text"
cp -R "$SCHRIFTEN" "$ZIELVORLAGEN/" || ende "die Schriften ließen sich nicht kopieren"
ok "$(du -sh "$ZIELVORLAGEN/$(basename "$SCHRIFTEN")" | cut -f1) Schriften: $(find "$ZIELVORLAGEN/$(basename "$SCHRIFTEN")" -type f \( -iname '*.ttf' -o -iname '*.otf' \) | wc -l | tr -d ' ') Dateien"

# **Das Beispielset ist die Kür, nicht die Pflicht.** Ohne es startet die
# Anwendung mit Vorlagen und ohne Set — unangenehm, aber brauchbar; ein
# Abbruch der Freigabe wäre die härtere Antwort als der Fall verdient.
if [ -f "$KANON/$BEISPIEL" ]; then
  # **Ausgepackt und nicht als Archiv**
  # ([0018](../docs/entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md),
  # #704).
  #
  # Ein `.mse-set` darf beides sein, und ein **Archiv** wird nicht
  # zurückgeschrieben: ⌘S darauf wird abgewiesen. Ein mitgeliefertes Set, das
  # sich nicht sichern lässt, wäre eine Einladung in eine Sackgasse — und das
  # ist der Fall, um den 0018 ausdrücklich gebeten hat („dann aber mit einem
  # Fall, nicht mit einem Verdacht").
  #
  # **Es kostet nichts**: gemessen 8,6 MB gepackt gegen 9,0 MB ausgepackt. Die
  # Kartenbilder sind bereits komprimiert.
  mkdir -p "$ZIELSETS/$BEISPIEL" || ende "$BEISPIEL anlegen"
  unzip -q "$KANON/$BEISPIEL" -d "$ZIELSETS/$BEISPIEL" \
    || ende "$BEISPIEL ließ sich nicht auspacken"
  ok "$(du -sh "$ZIELSETS/$BEISPIEL" | cut -f1) Beispielset: $BEISPIEL (ausgepackt)"
elif [ -d "$KANON/$BEISPIEL" ]; then
  cp -R "$KANON/$BEISPIEL" "$ZIELSETS/" || ende "$BEISPIEL ließ sich nicht kopieren"
  ok "$(du -sh "$ZIELSETS/$BEISPIEL" | cut -f1) Beispielset: $BEISPIEL"
else
  hin "$BEISPIEL liegt nicht unter $KANON — das Bündel geht ohne Beispielset"
fi

# **Und neu signieren** — aus demselben Grund wie in `scripts/saat.sh`: jede
# Datei, die nach Wails' Signatur unter `Resources` landet, bricht das Siegel,
# und auf arm64 ist die Ad-hoc-Signatur Voraussetzung
# ([0003](../docs/entscheidungen/0003-keine-notarisierung.md)).
BUENDEL="$(dirname "$CONTENTS")"
codesign --force --deep --options runtime --sign - "$BUENDEL" >/dev/null 2>&1 \
  || ende "das Bündel ließ sich nach dem Nachlegen der Vorlagen nicht neu signieren"

ok "Bündel neu signiert"
