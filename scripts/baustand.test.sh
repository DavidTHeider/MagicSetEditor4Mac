#!/usr/bin/env bash
#
# Fährt die Prüfungen des Bauabdrucks — T06
# (#2099).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/baustand.test.mjs"
