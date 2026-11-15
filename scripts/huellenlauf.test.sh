#!/usr/bin/env bash
#
# Fährt die Prüfungen des Laufs über die Drehbücher — PERF01
# (#2351).

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test "$wurzel/scripts/huellenlauf.test.mjs"
