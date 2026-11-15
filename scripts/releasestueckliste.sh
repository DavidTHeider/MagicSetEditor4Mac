#!/usr/bin/env bash
#
# Die Stücklisten einer Freigabe — eine für Go, eine für npm.
#
# Wird von scripts/release.sh eingebunden und erwartet dessen `ok`, `hin` und
# `ende`.

# stuecklisten_schreiben <ausgabe> — legt `stueckliste-go.json` und
# `stueckliste-npm.json` in <ausgabe> an; ein Fehler beendet die Freigabe.
stuecklisten_schreiben() {
  local ausgabe="$1"

  # Ein Fehler hier beendet die Freigabe.
  #
  # Vorher war es ein Hinweis, und das Skript meldete danach trotzdem „Fertig" —
  # während docs/CHRONIK.md zusagte, beide Stücklisten entstünden. Eine Zusage,
  # die bei Bedarf zum Hinweis wird, ist keine.
  #
  # **Mit `-json`** (SEC04, #2357): ohne den Schalter schreibt
  # `cyclonedx-gomod` XML, auch unter einem Namen auf `.json` — gemessen am
  # 2026-10-16. Bis 0.8.0 lag so XML unter `stueckliste-go.json`, und kein
  # Werkzeug, das JSON erwartet, konnte sie lesen.
  if cyclonedx-gomod mod -licenses -json -output "$ausgabe/.go-sbom" >/dev/null 2>&1 \
    && stueckliste_ist_json "$ausgabe/.go-sbom"; then
    mv "$ausgabe/.go-sbom" "$ausgabe/stueckliste-go.json"
    ok "stueckliste-go.json"
  else
    rm -f "$ausgabe/.go-sbom"
    ende "cyclonedx-gomod fehlgeschlagen oder ohne CycloneDX in JSON — 'make tools' ausführen; die Stückliste gehört zum Release"
  fi

  # Die Go-Stückliste trägt eine Pseudoversion, keine Releaseversion.
  #
  # `cyclonedx-gomod` liest die Modulversion aus Git, und ein Modul ohne
  # passenden Tag hat keine. Das ist kein Fehler des Bauwegs, sondern die Folge
  # der Reihenfolge: erst bauen, dann taggen. Es steht hier, damit niemand die
  # Zahl für die Releaseversion hält.
  hin "die Go-Stückliste nennt eine Pseudoversion — der Tag entsteht erst danach"

  # Über eine Zwischendatei, und mit einer Prüfung auf Inhalt.
  #
  # Eine Umleitung legt die Datei **vor** dem Kommando an: scheitert `npm sbom`,
  # bleibt eine Datei mit null Byte liegen — und die sieht im Verzeichnis aus wie
  # ein Artefakt. Genau so ist es beim ersten Probelauf passiert.
  if (cd frontend && npm sbom --sbom-format cyclonedx) > "$ausgabe/.npm-sbom" 2>/dev/null \
    && stueckliste_ist_json "$ausgabe/.npm-sbom"; then
    mv "$ausgabe/.npm-sbom" "$ausgabe/stueckliste-npm.json"
    ok "stueckliste-npm.json"
  else
    rm -f "$ausgabe/.npm-sbom"
    ende "npm sbom fehlgeschlagen oder ohne CycloneDX in JSON — die Stückliste gehört zum Release"
  fi
}

# stueckliste_ist_json <datei> — wahr, wenn die Datei eine CycloneDX-Stückliste
# in JSON ist. **Eine leere Datei ist keine**, und eine in XML auch nicht.
stueckliste_ist_json() {
  python3 -I -c 'import json,sys; b=json.load(open(sys.argv[1])); sys.exit(0 if b.get("bomFormat")=="CycloneDX" else 1)' "$1" 2>/dev/null
}

