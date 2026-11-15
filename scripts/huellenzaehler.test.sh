#!/usr/bin/env bash
#
# Fährt die Prüfungen des Handlungszählers der Hüllenprobe — G01
# (#2248).
# Die Gestalt wie `huellenorte.test.sh`: `ci.sh` fährt jede `*.test.sh`.

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/huellenzaehler.test.mjs"
