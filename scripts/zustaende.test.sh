#!/usr/bin/env bash
#
# Fährt die Prüfungen des Urteils der Zustandsprobe — G11
# (#2277).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/zustaende.test.mjs"
