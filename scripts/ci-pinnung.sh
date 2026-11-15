#!/usr/bin/env bash
#
# Die Pinnungsprüfungen der Stufe `format` (AGENTS.md §1 und §9).
#
# Ausgelagert aus scripts/ci.sh, weil die Datei über die Zeilengrenze aus §2
# gewachsen war. Der Schnitt folgt der Frage: hier steht, was prüft, *ob die
# Angaben über das Projekt stimmen* — nicht, ob der Code richtig ist.
#
# Wird von scripts/ci.sh eingebunden und nutzt dessen ok/fehl/weg.

# ------------------------------------------------------- Leitlinien (§9)

# Die beiden Leitliniendateien müssen gleichlauten.
#
# Sie sind absichtlich zwei Dateien und nicht eine mit Verweis: Werkzeuge lesen
# das eine oder das andere. Zwei Dateien laufen auseinander, sobald niemand
# nachzählt — also zählt hier jemand nach.
leitlinien_pruefen() {
  if [ ! -f AGENTS.md ] || [ ! -f CLAUDE.md ]; then
    fehl "AGENTS.md oder CLAUDE.md fehlt"
    return 0
  fi

  if cmp -s AGENTS.md CLAUDE.md; then
    ok "AGENTS.md und CLAUDE.md gleichlautend"
  else
    fehl "AGENTS.md und CLAUDE.md laufen auseinander — 'make leitlinien' ausführen"
  fi
}

# ------------------------------------------------------ npm-Pinnung (§1)

# npm_pinnung_pruefen hält das laufende npm gegen `packageManager`.
#
# Das Feld allein genügt nicht: ohne Corepack liest `npm ci` es gar nicht, und
# eine falsche Angabe liefe kommentarlos durch. Es ist die **Deklaration** —
# wirksam wird sie erst hier. `internal/werkzeuge/frontendpinnung_test.go` hält
# die andere Hälfte: dass Feld und Tabelle in AGENTS.md §1 übereinstimmen.
#
# Warum das zählt: `npm sbom` ist ein Bordmittel, und die Stückliste ist ein
# Releaseartefakt. Eine andere npm-Fassung erzeugt eine andere Stückliste.
npm_pinnung_pruefen() {
  [ -f frontend/package.json ] || { weg "keine frontend/package.json"; return 0; }

  local erwartet gefunden
  erwartet="$(sed -n 's/.*"packageManager"[[:space:]]*:[[:space:]]*"npm@\([^"]*\)".*/\1/p' \
    frontend/package.json)"

  if [ -z "$erwartet" ]; then
    fehl "frontend/package.json nennt kein packageManager: \"npm@<version>\""

    return 0
  fi

  gefunden="$(npm --version 2>/dev/null)"
  if [ "$gefunden" = "$erwartet" ]; then
    ok "npm $gefunden entspricht packageManager"
  else
    fehl "npm $gefunden läuft, gepinnt ist $erwartet (frontend/package.json, packageManager)"
  fi
}

# ------------------------------------------------ Werkzeugpinnung (§1)

# werkzeug_pinnung_pruefen hält jedes Werkzeug auf dem Pfad gegen seine
# Fassung in `tools/go.mod`.
#
# **Die Pinnung ist die Deklaration, wirksam wird sie hier** — dieselbe Form
# wie bei npm darüber. Bis QA-CI01 fragte `ci.sh` nur, **ob** gofumpt da ist;
# im Hauptarbeitsbaum lag v0.11.0, gepinnt war v0.12.0, und die Stufe nannte
# zehn Dateien, an denen nichts falsch war
# (#2106).
#
# Die Fassung kommt aus der `require`-Zeile: der `tool`-Block nennt dasselbe
# Modul ohne Zahl. Gelesen wird die laufende Fassung mit `go version -m`.
werkzeug_pinnung_pruefen() {
  [ -f tools/go.mod ] || { weg "keine tools/go.mod"; return 0; }

  local name modul erwartet pfad gefunden abweichend=""
  while read -r name modul; do
    erwartet="$(awk -v m="$modul" '$1 == m && $2 ~ /^v/ { print $2; exit }' tools/go.mod)"
    if ! pfad="$(command -v "$name" 2>/dev/null)"; then
      abweichend+="      $name fehlt, gepinnt ist $erwartet"$'\n'
      continue
    fi
    gefunden="$(go version -m "$pfad" 2>/dev/null | awk '$1 == "mod" { print $3; exit }')"
    if [ "$gefunden" != "$erwartet" ]; then
      abweichend+="      $name ${gefunden:-ohne Fassung} läuft, gepinnt ist $erwartet"$'\n'
    fi
  done <<'LISTE'
gofumpt mvdan.cc/gofumpt
golangci-lint github.com/golangci/golangci-lint/v2
gosec github.com/securego/gosec/v2
govulncheck golang.org/x/vuln
cyclonedx-gomod github.com/CycloneDX/cyclonedx-gomod
LISTE

  if [ -n "$abweichend" ]; then
    fehl "Werkzeuge weichen von tools/go.mod ab — 'make tools' ausführen:"$'\n'"${abweichend%$'\n'}"
  else
    ok "Werkzeuge entsprechen tools/go.mod"
  fi
}
