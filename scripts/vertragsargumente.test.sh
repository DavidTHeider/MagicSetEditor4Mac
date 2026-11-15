#!/usr/bin/env bash
#
# Fährt die Prüfung der Argumentzahlen im Vertragsrundgang — #2528.
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/vertragsargumente.test.mjs"
