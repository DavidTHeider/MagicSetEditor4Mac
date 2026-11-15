#!/usr/bin/env bash
#
# Fährt die Prüfungen der Fristen der Hüllenprobe — T02
# (#2096).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test --test-timeout=20000 "$wurzel/scripts/huellenfristen.test.mjs"
