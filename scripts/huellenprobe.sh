#!/usr/bin/env bash
#
# **Die Hüllenprobe fahren und belegen** — die Abnahmen am nativen Fenster:
# ENG-02.AK-04 (#1727),
# die Menüleiste aus #255 und FIX-F02 bis FIX-F04 (#1746–#1748).
#
#     scripts/huellenprobe.sh [<drehbuch> …]
#
# Baut die Oberfläche, das Bündel mit der Abnahmeleitung (`scripts/bauen.sh
# abnahme`) und `mse`, dann fährt es die Drehbücher unter
# `scripts/huellenprobe/` — ohne Namen alle. **Der Beleg** landet je Drehbuch
# in `docs/belege/huellenprobe/`, die Bilder unter `$TMPDIR/huellenprobe/`.
#
# **Nicht Teil von `make ci`**: es braucht ein Fenstersystem, startet die
# Anwendung über LaunchServices und öffnet echte Dateidialoge. Aufruf:
# `make huellenprobe`.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$WURZEL" || exit 1

ende() { printf '  ✗ %s\n' "$1"; exit 1; }

# shellcheck source=scripts/activate.sh
source scripts/activate.sh >/dev/null 2>&1

# **Die Oberfläche immer neu**: `bauen.sh` baut sie außerhalb der Freigabe
# nur, wenn es noch keine gibt — und eine alte gäbe einen Beleg über einen
# Stand, der nicht der geprüfte ist.
( cd frontend && npm run build >/dev/null 2>&1 ) || ende "die Oberfläche ließ sich nicht bauen"
./scripts/bauen.sh abnahme >/dev/null 2>&1 || ende "das Bündel ließ sich nicht bauen"
go build -o .go/bin/mse ./cmd/mse || ende "mse ließ sich nicht bauen"

# **Der Bildschirm bleibt an, solange die Probe läuft** — C03 (#2228): eine
# Webansicht, die niemand sieht, zeichnet nicht, und am 2026-10-10 wurden so
# zwanzig Drehbücher an Fristen rot. Schaltet ihn trotzdem jemand aus, sagt es
# der Lauf (`scripts/bildschirmzeit.mjs`).
caffeinate -dimsu node scripts/huellenprobe.mjs "$@"
