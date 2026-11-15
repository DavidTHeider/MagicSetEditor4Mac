#!/usr/bin/env bash
#
# Fährt die Prüfungen des Scanners nach eingeschmuggelten Geheimnissen — I05
# (#2263).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test \
	"$wurzel/scripts/geheimnisscan.test.mjs"
