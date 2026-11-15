#!/usr/bin/env bash
#
# **Was die Vorprüfung über ihre Läufe berichtet** — eingebunden von
# `scripts/vorpruefung.sh`, geprüft von `scripts/vorpruefungsbericht.test.sh`.
# Der Anlass steht dort im Kopf.
#
# `ok` und `fehl` kommen aus dem Skript, das diese Datei einbindet.

# fehlschlaege_nennen nennt jeden roten Lauf mit Namen und der ersten Zeile
# seiner Begründung, dazu jede rote Paketzeile und jede Panik.
#
# **Nicht die ersten zwanzig Zeilen.** Die füllte am 2026-09-30 ein einziger
# Fund, und vier andere standen erst zwanzig Minuten später in der Pipeline.
fehlschlaege_nennen() { # fehlschlaege_nennen <protokoll>
  awk '
    /^[[:space:]]*--- FAIL/ { print; begruendung = 1; next }
    begruendung && /^[[:space:]]+[^[:space:]-]/ { print; begruendung = 0; next }
    { begruendung = 0 }
    /^FAIL[[:space:]]/ || /^panic:/ { print }
  ' "$1" | sed 's/^/      /'
}

# abdeckung_pruefen zählt die unbetretenen Anweisungen der Pakete mit harter
# Schwelle — **gleich, wie `go test` ausging.**
#
# Bis zum 2026-09-30 stand sie im grünen Zweig, und ein roter Lauf verschwieg
# sie; die Pipeline fand sie dann eine Stufe später. **Nach einem roten Lauf
# kann ein Fund daher kommen**, dass ein `t.Fatal` den Rest des Laufs
# übersprang — das sagt die Zeile darüber.
#
# **Gezählt werden Anweisungen, nicht gerundete Prozente.** Am 2026-09-11
# stand `internal/service` bei „100.0 %" und meinte 9180 von 9183.
#
# **Nur die Pakete mit harter Schwelle**, und die Liste kommt aus
# `ci-abdeckung.sh` — eine zweite daneben liefe auseinander (§10). Ohne diesen
# Filter meldete der Riegel `internal/schriften`, das in `ohneHarteSchwelle`
# steht; ein Riegel, der falsch meldet, wird nicht mehr gelesen.
abdeckung_pruefen() { # abdeckung_pruefen <profil> <tests-gruen: ja|nein>
  local profil="$1" gruen="$2" mit_schwelle offen

  mit_schwelle=$(grep -oE '^KERNPAKETE="[^"]+"' scripts/ci-abdeckung.sh |
    sed 's/^KERNPAKETE="//; s/"$//' | tr ' ' '\n' |
    sed 's#^#MagicSetEditor4Mac/internal/#; s#$#/#')

  # **Ein leeres Profil ist keine Messung.** Ohne diese Zeile stünde „jede
  # Anweisung betreten" da, weil keine gelesen wurde (§10).
  if [ ! -s "$profil" ]; then
    fehl "das Abdeckungsprofil ist leer — hier wurde nichts gemessen"

    return
  fi

  offen=$(awk 'NR > 1 && $NF == 0 { print $1 }' "$profil" |
    grep -F -f <(printf '%s\n' "$mit_schwelle") |
    grep -v -F -f <(grep -oE '^[^ #]+' .covignore | grep -v '^$') || true)

  if [ -z "$offen" ]; then
    ok "jede Anweisung der Kernpakete betreten"
  else
    printf '%s\n' "$offen" | head -20 | sed 's/^/      /'

    if [ "$gruen" = "nein" ]; then
      printf '      (nach einem roten Lauf — ein abgebrochener Test kann sie übersprungen haben)\n'
    fi

    fehl "unbetretene Anweisungen in den Kernpaketen"
  fi
}
