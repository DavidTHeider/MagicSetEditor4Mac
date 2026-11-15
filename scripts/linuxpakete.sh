# shellcheck shell=bash
#
# **Was für Linux übersetzt wird** — eingebunden von `scripts/ci.sh` und
# `scripts/vorpruefung.sh` (ENG-02, #1727).
#
# Unter Linux baut Wails 3 nur mit cgo, und ohne Linux-Maschine gibt es kein
# cgo für Linux (Vermerk 0132). Übersetzt und gelintet wird dort deshalb alles
# **außer** dem Programm selbst und `internal/fenster`; was das kostet, steht im
# Vermerk.
#
# # Warum Muster je Verzeichnis und keine Liste von Paketen
#
# Drei Pakete bestehen nur aus Tests (`flaechen`, `schichten`, `werkzeuge`).
# `go build` nimmt sie in einem Muster mit `...` stillschweigend hin und weist
# sie als ausdrücklich genanntes Paket ab. Die Muster bleiben deshalb Muster —
# je Verzeichnis unter `internal/` eines, ohne das der Hülle.

pakete_fuer_linux() {
  local verzeichnis
  for verzeichnis in internal/*/; do
    verzeichnis="${verzeichnis%/}"
    [ "$verzeichnis" = internal/fenster ] || printf './%s/... ' "$verzeichnis"
  done
  printf './cmd/...\n'
}
