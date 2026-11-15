#!/usr/bin/env bash
#
# Fährt die Prüfungen der Klickhelfer der Hüllenprobe — R04
# (#2094).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/huellenklick.test.mjs"
