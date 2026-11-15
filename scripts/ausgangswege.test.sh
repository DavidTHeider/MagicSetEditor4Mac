#!/usr/bin/env bash
#
# Fährt die Prüfungen des Messvertrags der Ausgangswerte — P0
# (#2336).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/ausgangswege.test.mjs"
