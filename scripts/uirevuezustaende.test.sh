#!/usr/bin/env bash
#
# Fährt die Prüfungen des Messvertrags der befüllten Oberflächen — AP-13
# (#2339).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/uirevuezustaende.test.mjs"
