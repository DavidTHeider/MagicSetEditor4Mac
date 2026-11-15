#!/usr/bin/env bash
#
# Fährt die Prüfungen des Rufbuchs der Hüllenprobe — B01 (#2220).
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/huellenrufbuch.test.mjs"
