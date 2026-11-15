#!/usr/bin/env bash
# **`fm serve` in der Pipeline — nur, wenn es da ist** (AI-02,
# #1743).
#
# Die Vorgabe des Auftraggebers vom 2026-10-04: im Rahmen von `make ci` wird
# **nur** `fm serve` echt geprüft, und **nur, falls es vorhanden ist** —
# sonst wird der Lauf übersprungen. Die kleinen Ollama-Modelle laufen nicht
# hier; sie werden in der KI-Phase gemessen und danach wieder gelöscht.
#
# # Warum das Fehlen kein `weg` ist
#
# `weg` zählt eine übersprungene **Prüfung des Quelltexts**, und der Lauf
# stempelt dann nicht. Ein Rechner ohne macOS 27 hat kein `/usr/bin/fm` —
# das ist eine Eigenschaft der Maschine, nicht des Stands, und sie würde
# dort jeden Push anhalten. Gemeldet wird es trotzdem, gelb und mit Grund,
# und `reichweite` nennt es nach dem Lauf noch einmal.
#
# # Warum das Vorhandensein streng ist
#
# **Ist `fm` da, aber das Modell nicht bereit, ist der Lauf rot** —
# `TestDasEchteGeraetemodell` meldet dann BLOCKED. Ein abgeschaltetes Apple
# Intelligence ist ein Zustand, den der Mensch an diesem Rechner kennen soll.
#
# **Eingebunden und nicht ausgeführt**, wie `ci-reichweite.sh`.

# GERAETEMODELL_GEPRUEFT sagt `reichweite`, ob der Lauf `fm serve` erreicht hat.
GERAETEMODELL_GEPRUEFT=""

# geraetemodell_lauf fährt den Lauf gegen das echte Gerätemodell — ohne Beleg
# im Baum: die Pipeline prüft einen Stand und schreibt keinen.
geraetemodell_lauf() {
  MSE_FMLAUF_OHNE_BELEG=1 go test -count=1 -tags fmlauf -timeout 5m \
    -run TestDasEchteGeraetemodell ./internal/applefm/ 2>&1 | sed 's/^/      /'
  return "${PIPESTATUS[0]}"
}

# geraetemodell_pruefen prüft `fm serve`, wenn das Programm unter `$1` liegt
# (Vorgabe `/usr/bin/fm`, dieselbe Stelle wie `applefm.Programm`).
geraetemodell_pruefen() {
  local programm="${1:-/usr/bin/fm}"

  if [ ! -x "$programm" ]; then
    warn "fm serve: nicht geprüft — $programm fehlt auf diesem Rechner (erst ab macOS 27)"
    GERAETEMODELL_GEPRUEFT="nein"
    return 0
  fi

  if geraetemodell_lauf; then
    ok "fm serve: das echte Gerätemodell antwortet, lauscht auf keinem TCP-Port"
    GERAETEMODELL_GEPRUEFT="ja"
  else
    fehl "fm serve: der Lauf gegen das echte Gerätemodell"
    GERAETEMODELL_GEPRUEFT="rot"
  fi
}
