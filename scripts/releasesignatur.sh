#!/usr/bin/env bash
#
# Das Urteil über `spctl` (docs/RELEASE.md).
#
# Ausgelagert aus scripts/release.sh aus demselben Grund wie die
# Versionsprüfung: so lässt es sich fahren, ohne etwas zu bauen —
# `scripts/release.test.sh` schickt erfundene Rückgabewerte hinein.
#
# ## Warum es das überhaupt gibt
#
# Der naheliegende Weg ist fail-open:
#
#     if spctl --assess --type execute "$APP" >/dev/null 2>&1; then
#       ende "spctl nimmt die App an"
#     fi
#     hin "spctl lehnt ab — der erwartete Zustand"
#
# `spctl` gibt bei einer Ablehnung 3 zurück, aber ein fehlendes Werkzeug gibt
# 127, ein falscher Pfad 1, ein Abbruch 130 — und jeder davon liefe in dieselbe
# Zeile: „lehnt ab, alles wie erwartet". Ein Wächter, der Abwesenheit für
# Bestätigung nimmt, bestätigt nichts.

# spctl_urteil deutet den Rückgabewert von `spctl --assess`.
#
# Schreibt die Deutung nach stdout und gibt zurück:
#   0 — die erwartete Ablehnung
#   1 — alles andere, und der Text sagt, was
spctl_urteil() {
  local status="$1"

  case "$status" in
    0)
      printf 'spctl nimmt die App an — steht ein Zertifikat im Schlüsselbund, das hier nicht hingehört? (docs/entscheidungen/0003)\n'

      return 1
      ;;
    3)
      printf 'spctl lehnt ab (Status 3) — der erwartete Zustand, siehe docs/RELEASE.md\n'

      return 0
      ;;
    *)
      printf 'spctl endete mit Status %s — das ist weder Annahme noch Ablehnung, sondern ein Werkzeugfehler\n' \
        "$status"

      return 1
      ;;
  esac
}
