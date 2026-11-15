#!/usr/bin/env bash
#
# Fährt die Prüfungen des Kopiervorgänge der Hüllenprobe — PERF01
# (#2351).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/huellenkopie.test.mjs"
