#!/usr/bin/env bash
#
# Fährt die Prüfungen des Belegplans — B06
# (#2225), das Freigabetor aus J04.

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test \
	"$wurzel/scripts/belegplan.test.mjs"
