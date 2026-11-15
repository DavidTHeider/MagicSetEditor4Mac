#!/usr/bin/env bash
#
# **Jede Skriptprüfung läuft in der Pipeline** — C06
# (#2231).
#
# `skriptpruefungen` in `scripts/ci.sh` fährt die `*.test.sh` unter
# `scripts/`; eine `*.test.mjs` läuft nur, wenn eine davon sie nennt. Am
# 2026-10-10 nannte keine sieben von ihnen — `belegplan`, `dauerbetrieb`, die
# Zufallsmutation und vier weitere waren geschrieben, rot gesehen und grün
# gemacht und liefen danach nie wieder.
#
# **Was er nicht hält**: dass die nennende Hülle die Datei auch fährt. Er
# sucht den Pfad in Anführungszeichen, wie jede Hülle ihn `node --test`
# übergibt — eine Erwähnung im Kommentar trägt das Zeichen danach nicht.

set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

fehlt=0
for datei in "$wurzel"/scripts/*.test.mjs; do
	name="$(basename "$datei")"
	if ! grep -qF "scripts/$name\"" "$wurzel"/scripts/*.test.sh; then
		printf '  %s läuft in keiner Skriptprüfung\n' "$name"
		fehlt=1
	fi
done

exit "$fehlt"
