#!/usr/bin/env bash
#
# Fährt die Prüfungen des Auswerters der Bindungsreichweite — B01 (#2220).
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/bindungsreichweite.test.mjs"
