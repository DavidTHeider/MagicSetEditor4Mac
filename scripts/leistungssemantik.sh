#!/usr/bin/env bash
#
# **Dieselbe Ausgabe vor und nach einer Optimierung** — PERF-01.AK-06
# (#2168).
#
#     ./scripts/leistungssemantik.sh <mse-vorher> <mse-nachher> <arbeitsort>
#
# Zwei gebaute `mse` fahren dieselben kanonischen Sets ein und schreiben je
# Set die Kartenliste, vier Ausfuhren und die ersten 25 Karten als Bild; am
# Ende wird Datei für Datei verglichen, die Befunde der Ausfuhren ohne ihren
# Pfad. **Eine Zeile `VERSCHIEDEN` ist ein Befund** gegen die Optimierung,
# und der Lauf endet dann mit 1.
#
# **Und eine Zeile `GESCHEITERT` auch** — Befund 3 des Zweitblicks über
# Phase 0.7.0-I: scheitert ein Schritt auf **beiden** Seiten gleich, gibt es
# nichts, worin sie sich unterscheiden, und ein Vergleich allein hieß dann
# „gleich". Jeder Schritt muss gelingen und seine Datei hinterlassen; erst
# dann wird verglichen (`scripts/leistungssemantik.test.sh`).
#
# Eingänge: das Pack unter `~/full-magic-pack` (als APFS-Klon in ein eigenes
# Zuhause unter dem Arbeitsort) und die kanonischen Sets unter
# `~/Library/Application Support/MagicSetEditor4Mac/kanonische-sets` — gelesen,
# nicht verändert. Bestand und Schlüsselbund sind eigene (§3).
#
# **Was er nicht belegt**: Karten jenseits der ersten 25 je Set, andere
# Vorlagen und den Weg des Fensters. Gefahren am 2026-10-07 für PERF-02 und
# PERF-03: 365 Dateien je Stand, alle gleich — und nach der Prüfung jedes
# Schritts noch einmal mit demselben `mse` auf beiden Seiten: 365 Dateien,
# kein Schritt gescheitert.

set -uo pipefail

[ $# -eq 3 ] || { echo "Aufruf: $0 <mse-vorher> <mse-nachher> <arbeitsort>" >&2; exit 2; }

VORHER="$1"; NACHHER="$2"; ORT="$3"
KANON="$HOME/Library/Application Support/MagicSetEditor4Mac/kanonische-sets"
PACK="$HOME/full-magic-pack"
HEIM="$ORT/heim"
PAKETE="$HEIM/Library/Application Support/MagicSetEditor4Mac/pakete/full-magic-pack"

mkdir -p "$ORT"
if [ ! -d "$PAKETE/data" ]; then
  mkdir -p "$PAKETE"
  cp -cR "$PACK/data" "$PAKETE/data"
  cp -cR "$PACK/Magic - Fonts" "$PAKETE/Magic - Fonts"
fi

GESCHEITERT=0

# gescheitert vermerkt einen Schritt, der nicht gelang.
gescheitert() {
  echo "GESCHEITERT: $*"
  GESCHEITERT=1
}

lauf() { # lauf <mse> <ziel>
  local bin="$1" erg="$2" stand satz id karte karten vorlage
  stand="$(basename "$erg")"
  rm -rf "$erg"
  mkdir -p "$erg"
  for satz in Archenemy Planechase Vanguard; do
    local umgebung=(env HOME="$HEIM" MSE_BESTAND="$erg/$satz.db" MSE_SCHLUESSELBUND="$HEIM/keiner.keychain")
    # **Der Exitcode zählt, nicht nur die Antwort** — `pipefail` gibt den
    # der ganzen Kette, und ein `mse`, das gültiges JSON schreibt und mit
    # einem Fehler endet, ist gescheitert (Befund 2 des Nachlaufs).
    if ! id="$("${umgebung[@]}" "$bin" oeffnen "$KANON/$satz.mse-set" --json |
      python3 -c 'import json,sys; d=json.load(sys.stdin); print(d.get("setID") or d.get("id") or "")' 2>/dev/null)" || [ -z "$id" ]; then
      gescheitert "$stand $satz ließ sich nicht einfahren"
      continue
    fi
    if ! "${umgebung[@]}" "$bin" kartenliste "$id" >"$erg/$satz.liste" 2>&1 || [ ! -s "$erg/$satz.liste" ]; then
      gescheitert "$stand $satz: die Kartenliste"
    fi
    for vorlage in magic-simple-exporter magic-namecheck magic-text magic-mws; do
      if ! "${umgebung[@]}" "$bin" vorlagen "$id" ausfuehren "$vorlage.mse-export-template" \
        "$erg/$satz.$vorlage.aus" >"$erg/$satz.$vorlage.log" 2>&1 || [ ! -s "$erg/$satz.$vorlage.aus" ]; then
        gescheitert "$stand $satz: die Ausfuhr $vorlage"
      fi
      sed -i '' "s#$erg#ORT#g" "$erg/$satz.$vorlage.log"
    done
    if ! karten="$("${umgebung[@]}" "$bin" karten "$id" --json |
      python3 -c 'import json,sys; d=json.load(sys.stdin); k=d if isinstance(d,list) else d["karten"]; print(" ".join(str(x.get("ID") or x.get("id")) for x in k[:25]))' 2>/dev/null)" || [ -z "$karten" ]; then
      gescheitert "$stand $satz nennt keine Karte"
    fi
    for karte in $karten; do
      if ! "${umgebung[@]}" "$bin" zeichnen "$karte" "$erg/$satz.$karte.png" >/dev/null 2>&1 || [ ! -s "$erg/$satz.$karte.png" ]; then
        gescheitert "$stand $satz $karte ließ sich nicht zeichnen"
      fi
    done
    rm -f "$erg/$satz.db" "$erg/$satz.db-wal" "$erg/$satz.db-shm" "$erg/$satz.db.sperre"
  done
}

lauf "$VORHER" "$ORT/vorher"
lauf "$NACHHER" "$ORT/nachher"

if [ "$GESCHEITERT" -ne 0 ]; then
  echo "GESCHEITERT — ohne vollständige Ausgabe gibt es nichts zu vergleichen"
  exit 1
fi

if diff -rq "$ORT/vorher" "$ORT/nachher"; then
  echo "gleich: $(find "$ORT/vorher" -type f | wc -l | tr -d ' ') Dateien"
else
  echo "VERSCHIEDEN — siehe oben"
  exit 1
fi
