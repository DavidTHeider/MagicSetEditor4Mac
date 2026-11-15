#!/usr/bin/env bash
#
# Fährt die Prüfungen der Auswertung von `dauerbetrieb` und `dauerteile` — B04
# (#2223).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test \
	"$wurzel/scripts/dauerbetrieb.test.mjs" \
	"$wurzel/scripts/dauerteile.test.mjs"
