#!/usr/bin/env bash
#
# Fährt die Prüfungen des Mutationsläufers — T05
# (#2108).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/mutationslauf.test.mjs"
