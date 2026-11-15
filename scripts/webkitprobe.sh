#!/usr/bin/env bash
#
# **Die WebKit-Probe bauen und fahren** —
# #1828.
#
#     scripts/webkitprobe.sh <drehbuch> [--ausgabe <verzeichnis>] \
#                            [--basis <adresse>] [--motor wk1|wk2] \
#                            [--breite 1440] [--hoehe 900] [--dunkel] [--sichtbar]
#
# Baut das Programm aus `scripts/webkitprobe/*.swift` nach
# `.go/bin/webkitprobe` — neu nur, wenn eine Quelle jünger ist als das
# Programm — und reicht die Argumente durch. **Der Stand geht mit**: der
# Commit und ein `+`, wenn der Baum nicht sauber ist. Ein Bericht ohne Stand
# belegte einen Lauf, von dem niemand weiß, worüber er ging (AGENTS.md §3).
#
# **Was die Probe kann und was nicht**, steht im Kopf von
# `scripts/webkitprobe/Motoren.swift`; die Sprache der Drehbücher in
# `scripts/webkitprobe/Drehbuch.swift`.

set -euo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ZIEL="$WURZEL/.go/bin/webkitprobe"
QUELLEN=("$WURZEL"/scripts/webkitprobe/*.swift)

# **`xcrun`, nicht ein `swiftc` aus dem Pfad**: welches Swift gilt, sagt die
# gewählte Xcode-Fassung — dieselbe, mit der das Programm gebaut wird.
if [ ! -x "$ZIEL" ] || [ -n "$(find "${QUELLEN[@]}" -newer "$ZIEL" 2>/dev/null)" ]; then
  mkdir -p "$(dirname "$ZIEL")"
  # `-suppress-warnings`: WebKit 1 ist seit macOS 10.14 als veraltet markiert,
  # und genau diesen Weg braucht der Motor `wk1` — die Warnung sagte bei jedem
  # Bau, was hier absichtlich so ist.
  xcrun swiftc -swift-version 5 -O -suppress-warnings -o "$ZIEL" "${QUELLEN[@]}"
fi

# **Die Belege selbst zählen nicht** — sonst stünde ab dem zweiten Drehbuch
# einer Abnahme ein `+` im Bericht, weil der erste Bericht schon daliegt.
STAND="$(git -C "$WURZEL" rev-parse --short HEAD)"
if [ -n "$(git -C "$WURZEL" status --porcelain -- . ':(exclude)docs/belege/webkitprobe')" ]; then
  STAND="$STAND+"
fi

exec "$ZIEL" "$@" --stand "$STAND"
