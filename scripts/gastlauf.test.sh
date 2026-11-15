#!/usr/bin/env bash
#
# Prüfungen der Gastläufer — Befund 1 des Zweitblicks über Phase I: ein
# leerer VM-Name umging die Sperre »nie am Wirt«. Sie starten keine VM.
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/gastlauf.test.mjs"
