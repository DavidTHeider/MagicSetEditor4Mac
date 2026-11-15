#!/usr/bin/env bash
#
# Prüfungen des Lastlaufs — QA-08
# (#1739),
# Z3 und Z4 des Zweitblicks über Phase 6.11.
#
# **Sie fahren weder `mse-rest` noch `mse-mcp`.** Geprüft werden
#
#   - die Wirkung (`scripts/lastlaufwirkung.mjs`): ein Dienst-Double, das 200
#     sagt und nichts schreibt, zählt als Fehler, ein ehrliches nicht;
#   - die Prozesse (`scripts/lastlaufdienste.mjs`): ein Dienst, der nicht
#     anläuft, ein Prozess, der mitten im Ruf stirbt, eine Antwort, die
#     ausbleibt — an kleinen Node-Programmen, die sich so verhalten.
#
# Der Lauf selbst, `node scripts/lastlauf.mjs`, baut die Programme und braucht
# Zeit und steht in `docs/RELEASE.md` und nicht in der Pipeline.
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test \
	"$wurzel/scripts/lastlaufwirkung.test.mjs" \
	"$wurzel/scripts/lastlaufdienste.test.mjs" \
	"$wurzel/scripts/lastlaufprofile.test.mjs"
