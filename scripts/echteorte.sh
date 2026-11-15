# shellcheck shell=bash
#
# **Die Ortswache um die Go-Läufe** — A01
# (#2215).
#
# Von `scripts/ci.sh` und `scripts/vorpruefung.sh` gemeinsam geladen, damit
# beide Läufe **dasselbe** Zuhause bekommen und **dieselben** Orte ansehen.
# Bis A01 gab nur die Pipeline den Go-Läufen ein leeres Zuhause; die
# Vorprüfung schrieb mit dem echten — gemessen vom Audit zu 0.8.0 an zwei
# Drosseldateien und einem Scryfall-Eintrag im echten Anwendungsverzeichnis.
#
# Was der Abdruck sieht und was nicht, steht in `scripts/huellenorte.mjs`.

# ECHTEORTE_SKRIPTE ist das Verzeichnis dieser Datei — beim Laden bestimmt,
# auch wenn eine andere Muschel als bash sie lädt.
ECHTEORTE_SKRIPTE="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"

# echtes_zuhause gibt das Zuhause aus dem Benutzereintrag — **nicht `HOME`**,
# denn das ist gerade, was die Läufe austauschen. `ECHTES_ZUHAUSE` setzt nur
# `scripts/echteorte.test.sh`, damit seine Probe kein echtes anfasst.
echtes_zuhause() {
  if [ -n "${ECHTES_ZUHAUSE:-}" ]; then
    printf '%s\n' "$ECHTES_ZUHAUSE"
  else
    node -p "require('node:os').userInfo().homedir"
  fi
}

# heim_fuer_laeufe legt ein leeres Zuhause für die Go-Läufe an und gibt
# seinen Pfad — aufräumen muss, wer es bekommt.
heim_fuer_laeufe() { mktemp -d -t mse-laufzuhause; }

# echte_orte_merken <datei> schreibt den Abdruck der echten Orte.
echte_orte_merken() {
  node "$ECHTEORTE_SKRIPTE/huellenorte.mjs" abdruck "$(echtes_zuhause)" >"$1"
}

# echte_orte_pruefen <datei> vergleicht mit dem Abdruck: 0 unberührt, 1
# verändert (die Pfade stehen auf der Ausgabe), 3 nicht prüfbar.
echte_orte_pruefen() {
  node "$ECHTEORTE_SKRIPTE/huellenorte.mjs" vergleich "$(echtes_zuhause)" "$1"
}
