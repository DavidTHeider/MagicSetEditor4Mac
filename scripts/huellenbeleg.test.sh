#!/usr/bin/env bash
#
# Fährt die Prüfungen des Belegs einer Hüllenprobe — B06
# (#2225) — und
# die der Ausphasen des Bildschirms, C03 (#2228).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test \
	"$wurzel/scripts/huellenbeleg.test.mjs" \
	"$wurzel/scripts/bildschirmzeit.test.mjs"
