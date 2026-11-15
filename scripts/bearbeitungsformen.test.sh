#!/usr/bin/env bash
# Prüfungen von scripts/bearbeitungsformen.tsv — #2393. Die Fälle stehen in
# scripts/bearbeitungsformen.test.mjs; hier nur, damit die Pipeline sie fährt.
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
node --test "$wurzel/scripts/bearbeitungsformen.test.mjs"
