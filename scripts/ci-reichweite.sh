#!/usr/bin/env bash
# **Was ein grüner Lauf nicht sagt.**
#
# **Eine eigene Datei**, weil `ci.sh` mit ihr über 400 Zeilen ginge (§2). Der
# Schnitt folgt der Frage: dort steht, was geprüft wird; hier, was dabei
# **ungeprüft bleibt**.
#
# **Eingebunden und nicht ausgeführt**: sie erklärt eine Funktion und ruft
# nichts.

# reichweite sagt nach einem grünen Lauf, **was er nicht sagt**.
#
# Ohne sie steht dort ein Haken, und ein Haken sieht aus wie ein Beweis.
# `docs/entscheidungen/0007` nennt den Preis des fehlenden entfernten Laufs,
# und diese Ausgabe nennt ihn nach jedem grünen Lauf (Issue 870).
#
# **Die Zahlen werden gelesen, nicht geschrieben.** Eine neue Zeile in
# `.covignore` ändert die Ausgabe; eine Konstante daneben liefe auseinander.
reichweite() {
  ausnahmen=$(grep -vc -e '^#' -e '^$' .covignore || true)
  fremdlaeufe=$(grep -rl 'go:build fremdpakete' internal cmd 2>/dev/null | wc -l | tr -d ' ')
  netzlaeufe=$(grep -rl 'go:build netzlauf' internal cmd 2>/dev/null | wc -l | tr -d ' ')

  printf '%s  Was dieser Lauf nicht belegt:%s\n' "$GELB" "$AUS"
  printf '%s    · eine Maschine, keine Plattform — kein zweiter Rechner, kein frischer Klon (Vermerk 0007)%s\n' "$GELB" "$AUS"
  printf '%s    · %s Einträge in .covignore sind von der Abdeckung ausgenommen%s\n' "$GELB" "$ausnahmen" "$AUS"
  printf '%s    · %s Dateien hinter dem Tag fremdpakete sind nicht gelaufen — darunter der Kanonlauf gegen MSE2%s\n' "$GELB" "$fremdlaeufe" "$AUS"
  printf '%s    · %s Dateien hinter dem Tag netzlauf sind nicht gelaufen — kein Aufruf ging an Scryfall oder magicthegathering.io%s\n' "$GELB" "$netzlaeufe" "$AUS"

  # **Und ob `fm serve` erreicht wurde** (AI-02, #1743): es läuft nur, wo es
  # da ist, und ein grüner Lauf ohne es sieht aus wie einer mit.
  if [ "${GERAETEMODELL_GEPRUEFT:-}" = "nein" ]; then
    printf '%s    · fm serve wurde nicht geprüft — /usr/bin/fm fehlt auf diesem Rechner%s\n' "$GELB" "$AUS"
  fi
  printf '%s    · kein Ollama-Modell lief — die fünf kleinen Modelle sind gemessen, nicht geprüft (docs/belege/modellmessung.md)%s\n' "$GELB" "$AUS"

  # **Und was der letzte Messlauf sagt** — QA-01.AK-04
  # (#1733):
  # „Überfällige oder zum Commit unpassende Nachweise sind sichtbar." Die
  # Zeile darüber sagt, dass die Läufe hier nicht liefen; diese, wann und an
  # welchem Stand sie zuletzt liefen. **Die Reichweite berichtet und sperrt
  # nicht** — eine Prüfung, die mit der Uhr rot würde, hielte jeden Push an,
  # ohne dass sich im Baum etwas geändert hätte.
  for plan in docs/belege/*.plan.json; do
    [ -f "$plan" ] || continue

    # **Die Zeile von `go` selbst gehört nicht dazu**: endet die Prüfung mit
    # 1 — bei jedem Beleg, der nicht PASS und aktuell ist —, hängt `go run`
    # »exit status 1« an, und die letzte Zeile war dann diese statt des
    # Standes (#1801).
    #
    # **Ohne die Eingaben, und die Zeile sagt es** (»ohne Eingaben«): die
    # Pipeline kennt sie nicht, und ob sie noch dieselben sind, sagt
    # `make belege` (Zweitblick über Phase 6.0).
    stand="$(go run ./internal/laufbeleg/lauf -plan "$plan" -pruefen -ohne-eingaben 2>&1 || true)"
    printf '%s    · Beleg %s%s\n' "$GELB" \
      "$(printf '%s\n' "$stand" | grep -v '^exit status [0-9]*$' | tail -1)" "$AUS"
  done
}
