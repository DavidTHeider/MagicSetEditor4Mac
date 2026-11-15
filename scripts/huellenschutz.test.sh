#!/usr/bin/env bash
#
# Fährt die Prüfungen der Schutzrichtlinienprüfung der Hüllenprobe — SEC-17
# (#2164).
# Die Gestalt wie `lastlauf.test.sh`: `ci.sh` fährt jede `*.test.sh`.

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/huellenschutz.test.mjs"
