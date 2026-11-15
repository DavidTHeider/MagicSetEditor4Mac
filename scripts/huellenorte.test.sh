#!/usr/bin/env bash
#
# Fährt die Prüfungen des Ortsabdrucks der Hüllenprobe — T01
# (#2095).
# Die Gestalt wie `lastlauf.test.sh`: `ci.sh` fährt jede `*.test.sh`.

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/huellenorte.test.mjs"
