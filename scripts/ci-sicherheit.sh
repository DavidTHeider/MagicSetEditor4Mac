#!/usr/bin/env bash
#
# Die Stufe `security` aus scripts/ci.sh (AGENTS.md §4) — ausgelagert mit
# ENG-04 (#1729),
# als sie die vier Zustände aus `scripts/sicherheitslage.sh` bekam und `ci.sh`
# über die Zeilengrenze aus §2 gegangen wäre. Wird von scripts/ci.sh
# eingebunden und nutzt dessen ok/fehl/weg.
#
# # Was ein Zustand hier bedeutet
#
#   ohne-befund       grün
#   befund            rot — eine bekannte Schwachstelle
#   nicht-erreichbar  **übergangen**, nicht grün: der Lauf hat nichts geprüft,
#                     und ein übergangener Lauf stempelt nicht (§4)
#   gescheitert       rot — der Scanner lief nicht durch

# shellcheck source=scripts/sicherheitslage.sh
source "$(dirname "${BASH_SOURCE[0]}")/sicherheitslage.sh"

stufe_security() {
  kopf "Sicherheit"

  if hat_go && werkzeug_da gosec; then
    # shellcheck disable=SC2086
    if gosec -quiet $GO_PAKETE 2>&1 | sed 's/^/      /'; then ok "gosec"; else fehl "gosec"; fi
  elif hat_go; then
    fehl "gosec nicht gefunden — 'make tools' ausführen"
  else
    weg "kein go.mod"
  fi

  if hat_go && werkzeug_da govulncheck; then
    govulncheck_pruefen
  elif hat_go; then
    fehl "govulncheck nicht gefunden — 'make tools' ausführen"
  fi

  if hat_frontend; then
    npm_audit_pruefen
  else
    weg "kein Frontend"
  fi
}

# zustand_melden setzt einen der vier Zustände in eine Zeile der Stufe um.
zustand_melden() { # zustand_melden <scanner> <zustand>
  case "$2" in
    ohne-befund)      ok "$1" ;;
    befund)           fehl "$1 — eine bekannte Schwachstelle" ;;
    nicht-erreichbar) weg "$1 — nicht erreichbar (kein Befund, keine Prüfung)" ;;
    *)                fehl "$1 — gescheitert" ;;
  esac
}

# govulncheck_pruefen fährt govulncheck und ordnet ein, was herauskam.
#
# **Der Rückgabewert zählt** — 3 ist ein Befund, 1 ein Fehler — und wird mit
# `|| rueckgabe=$?` gehalten: so bricht `set -e` nicht ab, und der Wert bleibt
# der des Scanners.
govulncheck_pruefen() {
  local raum rueckgabe=0
  raum="$(mktemp -d)"

  # shellcheck disable=SC2086
  govulncheck $GO_PAKETE > "$raum/govuln" 2>&1 || rueckgabe=$?

  sed 's/^/      /' "$raum/govuln"
  zustand_melden "govulncheck" "$(govuln_zustand "$rueckgabe" "$raum/govuln")"
  rm -rf "$raum"
}

# npm_audit_pruefen trennt einen Befund von einem Ausfall der Registry.
#
# **Weil die Stufe beides gleich meldete.** Am 2026-09-04 gab
# `registry.npmjs.org` für den Audit-Endpunkt **503**, und die Pipeline schrieb
# „npm audit — Befunde ab 'high'". Das ist eine Meldung, die mehr behauptet, als
# sie weiß: sie nennt einen Befund, wo gar keine Antwort kam — genau die
# Gestalt, vor der §10 warnt.
#
# **Seit ENG-04 liest sie den Bericht** (`--json`) statt nach Wörtern zu
# suchen: ein Befund ist eine Zahl über `high`, ein Ausfall eine Meldung ohne
# Bericht (`npmaudit_zustand`). Gezeigt wird die Bilanz, nicht der ganze
# Bericht.
#
# **Rot bleibt rot.** Ein Ausfall wird als **übergangen** geführt und nicht als
# grün: der Lauf hat nichts geprüft, und das gehört gesagt.
npm_audit_pruefen() {
  local raum rueckgabe=0 bilanz
  raum="$(mktemp -d)"

  # **Nicht als Zuweisung aus einer Ersetzung.** Unter `set -e` bricht eine
  # Ersetzung, die nicht null zurückgibt, den ganzen Lauf ab — still, mitten
  # in der Stufe. Genau das ist beim ersten Anlauf 2026-09-04 passiert: die
  # Ausgabe endete nach `govulncheck`, und `make` gab Error 1 ohne eine
  # einzige rote Zeile.
  (cd frontend && npm audit --json --audit-level=high) > "$raum/npm" 2>&1 || rueckgabe=$?

  bilanz="$(npmaudit_zahlen "$raum/npm")"
  if [ -n "$bilanz" ]; then
    printf '      %s\n' "$bilanz"
  else
    # **`head` zuerst**: umgekehrt bekäme `sed` ein SIGPIPE, und unter
    # `pipefail` endete die Stufe mitten im Satz.
    head -5 "$raum/npm" | sed 's/^/      /'
  fi

  zustand_melden "npm audit" "$(npmaudit_zustand "$rueckgabe" "$raum/npm")"
  rm -rf "$raum"
}
