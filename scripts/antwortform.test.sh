#!/usr/bin/env bash
#
# Fährt die Prüfungen der Antwortformprüfung — C01
# (#2226).
# Braucht TypeScript aus `frontend/node_modules`.

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test \
	"$wurzel/scripts/antwortform.test.mjs"
