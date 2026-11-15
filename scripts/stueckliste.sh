#!/usr/bin/env bash
#
# Die Stücklisten außerhalb der Freigabe — `make sbom`.
#
# **Derselbe Weg wie in der Freigabe** (`scripts/releasestueckliste.sh`):
# CycloneDX in JSON, vor dem Ablegen geprüft. `make sbom` rief
# `cyclonedx-gomod` bis zum Zweitblick über Phase 5 (Befund 5) ohne `-json`
# und schrieb XML unter `stueckliste-go.json`.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$WURZEL" || exit 1

# shellcheck source=scripts/activate.sh
source "$WURZEL/scripts/activate.sh" >/dev/null 2>&1

ok() { printf '  ✓ %s\n' "$1"; }
hin() { printf '  ~ %s\n' "$1"; }
ende() { printf '  ✗ %s\n' "$1"; exit 1; }

# shellcheck source=scripts/releasestueckliste.sh
source "$WURZEL/scripts/releasestueckliste.sh"

stuecklisten_schreiben "$WURZEL"
