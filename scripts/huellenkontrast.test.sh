#!/usr/bin/env bash
#
# Fährt die Prüfungen der Kontrastrechnung der Hüllenprobe — ABN-01.AK-03
# (#2175).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/huellenkontrast.test.mjs"
