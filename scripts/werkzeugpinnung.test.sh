#!/usr/bin/env bash
#
# Die Werkzeuge in `.go/bin` sind die gepinnten — QA-CI01
# (#2106).
#
# # Wofür es diesen Wächter gibt
#
# **Der Audit für 0.7.0 fand zehn Formatabweichungen, die keine waren.** Im
# Hauptarbeitsbaum lag gofumpt v0.11.0 vom 2026-09-14; gepinnt war seit #1784
# v0.12.0, und mit dem stehen dieselben zehn Dateien in `internal/mcp` sauber
# da (gemessen am 2026-10-06). `ci.sh` fragte nur, **ob** gofumpt da ist —
# nicht, **welches**. Eine Pinnung, die niemand gegen das laufende Werkzeug
# hält, ist eine Deklaration (AGENTS.md §1).
#
# # Was er hält
#
#  1. **Ein Werkzeug in der gepinnten Fassung ist grün.**
#  2. **Eine andere Fassung ist rot, und die Meldung nennt beide Zahlen.**
#  3. **Ein fehlendes Werkzeug ist rot**, nicht übersprungen.
#  4. **Die Fassung kommt aus der `require`-Zeile**, nicht aus dem `tool`-Block,
#     der dasselbe Modul ohne Zahl nennt.
#
# # Was er nicht hält
#
# **Ob die Pinnung selbst stimmt** — das hält `internal/werkzeuge/pinnung_test.go`
# gegen AGENTS.md. Und er fährt kein echtes `go version -m`: ein Ersatz-`go`
# gibt die Ausgabe in derselben Gestalt aus.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FEHLER=0
ARBEIT="$(mktemp -d)"
trap 'rm -rf "$ARBEIT"' EXIT

# Ersatz für ok/fehl/weg aus ci.sh: sie schreiben nur mit.
MELDUNGEN=""
ok()   { MELDUNGEN+="ok:$1"$'\n'; }
fehl() { MELDUNGEN+="fehl:$1"$'\n'; }
weg()  { MELDUNGEN+="weg:$1"$'\n'; }

# shellcheck source=scripts/ci-pinnung.sh
source "$WURZEL/scripts/ci-pinnung.sh"

behauptung() { # behauptung <beschreibung> <ja|nein>
  if [ "$2" = ja ]; then
    printf '  ✓ %s\n' "$1"
  else
    printf '  ✗ %s\n%s\n' "$1" "$MELDUNGEN" | sed 's/^/    /'
    FEHLER=1
  fi
}

# aufbauen <gofumpt-fassung|-> baut ein Projekt mit tools/go.mod, fünf
# Werkzeugen und einem Ersatz-go, das deren Fassung in der Gestalt von
# `go version -m` ausgibt. `-` lässt gofumpt weg.
aufbauen() {
  rm -rf "${ARBEIT:?}/projekt" "${ARBEIT:?}/bin"
  mkdir -p "$ARBEIT/projekt/tools" "$ARBEIT/bin"
  cat >"$ARBEIT/projekt/tools/go.mod" <<'MOD'
module beispiel/tools

go 1.27.1

tool (
	github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod
	github.com/golangci/golangci-lint/v2/cmd/golangci-lint
	github.com/securego/gosec/v2/cmd/gosec
	golang.org/x/vuln/cmd/govulncheck
	mvdan.cc/gofumpt
)

require (
	github.com/CycloneDX/cyclonedx-gomod v1.12.0 // indirect
	github.com/golangci/golangci-lint/v2 v2.14.0 // indirect
	github.com/securego/gosec/v2 v2.29.0 // indirect
	golang.org/x/vuln v1.8.0 // indirect
	mvdan.cc/gofumpt v0.12.0 // indirect
)
MOD
  local name modul fassung
  while read -r name modul fassung; do
    if [ "$name" = gofumpt ]; then
      [ "$1" = - ] && continue
      fassung="$1"
    fi
    printf '#!/bin/sh\n# %s %s\n' "$modul" "$fassung" >"$ARBEIT/bin/$name"
    chmod +x "$ARBEIT/bin/$name"
  done <<'LISTE'
gofumpt mvdan.cc/gofumpt v0.12.0
golangci-lint github.com/golangci/golangci-lint/v2 v2.14.0
gosec github.com/securego/gosec/v2 v2.29.0
govulncheck golang.org/x/vuln v1.8.0
cyclonedx-gomod github.com/CycloneDX/cyclonedx-gomod v1.12.0
LISTE
  # Ersatz-go: `go version -m <pfad>` liest die Kommentarzeile des Werkzeugs.
  cat >"$ARBEIT/bin/go" <<'GO'
#!/bin/sh
[ "$1" = version ] && [ "$2" = -m ] || exit 2
set -- $(sed -n '2s/^# //p' "$3")
printf '%s: go1.27.1\n\tpath\t%s\n\tmod\t%s\t%s\th1:abc=\n' "$3" "$1" "$1" "$2"
GO
  chmod +x "$ARBEIT/bin/go"
}

echo "werkzeugpinnung"

aufbauen v0.12.0
MELDUNGEN="$(cd "$ARBEIT/projekt" && PATH="$ARBEIT/bin:/usr/bin:/bin" werkzeug_pinnung_pruefen; printf '%s' "$MELDUNGEN")"
case "$MELDUNGEN" in
  *fehl:*) behauptung "gepinnte Fassungen sind grün" nein ;;
  *ok:*) behauptung "gepinnte Fassungen sind grün" ja ;;
  *) behauptung "gepinnte Fassungen sind grün" nein ;;
esac

aufbauen v0.11.0
MELDUNGEN="$(cd "$ARBEIT/projekt" && PATH="$ARBEIT/bin:/usr/bin:/bin" werkzeug_pinnung_pruefen; printf '%s' "$MELDUNGEN")"
case "$MELDUNGEN" in
  *fehl:*gofumpt*v0.11.0*v0.12.0*) behauptung "eine andere Fassung ist rot und nennt beide" ja ;;
  *) behauptung "eine andere Fassung ist rot und nennt beide" nein ;;
esac
case "$MELDUNGEN" in
  *golangci-lint*) behauptung "nur das abweichende Werkzeug wird genannt" nein ;;
  *) behauptung "nur das abweichende Werkzeug wird genannt" ja ;;
esac

aufbauen -
MELDUNGEN="$(cd "$ARBEIT/projekt" && PATH="$ARBEIT/bin:/usr/bin:/bin" werkzeug_pinnung_pruefen; printf '%s' "$MELDUNGEN")"
case "$MELDUNGEN" in
  *fehl:*gofumpt*fehlt*) behauptung "ein fehlendes Werkzeug ist rot" ja ;;
  *) behauptung "ein fehlendes Werkzeug ist rot" nein ;;
esac

exit "$FEHLER"
