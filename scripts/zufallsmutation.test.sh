#!/usr/bin/env bash
#
# Fährt die Prüfungen der Zufallsmutation und ihres Berichts — B05
# (#2224).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test \
	"$wurzel/scripts/zufallsmutationkern.test.mjs" \
	"$wurzel/scripts/zufallsmutationbericht.test.mjs"
