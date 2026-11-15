#!/usr/bin/env bash
#
# Prüfungen der Fuzzkampagne — QA-06
# (#1737),
# Z5 des Zweitblicks über Phase 6.11.
#
# **Sie fuzzen nichts.** Geprüft wird die Auswertung der drei Teile einer
# Kampagne (`scripts/fuzzbericht.mjs`) an Läufen, die der Test vorgibt: was ein
# Fund ist, was ein gescheiterter Teil, und dass ein fehlender Teil Status und
# Exitcode ändert. Die Kampagne selbst, `node scripts/fuzzkampagne.mjs`,
# braucht Zeit und steht in `docs/RELEASE.md` und nicht in der Pipeline.
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/fuzzbericht.test.mjs"
