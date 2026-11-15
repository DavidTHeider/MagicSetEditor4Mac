#!/usr/bin/env bash
#
# Der **Freigabebeleg** — Werkstück 5 von Phase Z
# (#892).
#
# # Wofür
#
# `docs/RELEASE.md` nennt Läufe, die „nicht optional" sind: der Lauf gegen
# bezogene Daten, die Sichtprüfung in hell und dunkel, ein wirklich gedruckter
# Bogen, der Erststart-Weg. `release.sh` **druckt** sie als Liste und prüft
# keine einzige.
#
# Damit hängt jede dieser Zusagen an Prosa und an nichts sonst. Wer eine Woche
# später fragt, ob der Druckbogen für **diese** Fassung wirklich gedruckt
# wurde, findet keine Antwort — nur den Satz, dass es zu tun war.
#
# # Warum am Artefakthash und nicht am Tag
#
# **Ein Tag lässt sich verschieben, ein Hash nicht.** `git tag -f` zeigt danach
# auf einen anderen Commit, und ein Beleg, der am Tag hinge, hätte plötzlich
# eine andere Fassung bestätigt. Der Beleg nennt deshalb die SHA-256 des
# `.dmg`; wer neu baut, bekommt einen neuen Hash und damit einen neuen Beleg.
#
# **Ein neuer Bau ist ein neues Artefakt**, auch bei gleichem Quelltext: das
# `.dmg` trägt Zeitstempel. Das ist der Preis dieser Bindung und zugleich ihr
# Sinn — geprüft wird, was ausgeliefert wird.
#
# # Was er nicht kann
#
# **Er glaubt, was jemand einträgt.** Ein Haken sagt, dass eine Person zu einem
# Zeitpunkt behauptet hat, die Prüfung gemacht zu haben. Mehr kann ein Beleg
# über eine Handprüfung nicht sagen — und weniger wäre gar nichts.

# belegname gibt den Dateinamen des Belegs zu einer Fassung.
belegname() { # belegname <fassung>
  printf 'freigabebeleg-%s.md' "$1"
}

# artefakthash liest die SHA-256 des Abbilds aus einer Prüfsummendatei.
#
# **Aus der Datei und nicht neu gerechnet**: der Beleg soll dieselbe Zahl
# nennen, die ausgeliefert wird. Rechnete er selbst, könnte er eine andere
# Datei erwischen als die, die neben ihm liegt.
artefakthash() { # artefakthash <pruefsummendatei>
  local datei="$1"

  [ -r "$datei" ] || { printf 'die Prüfsummendatei %s ist nicht lesbar\n' "$datei" >&2; return 1; }

  local zeile
  zeile="$(grep -E '\.dmg$' "$datei" | head -1)"

  [ -n "$zeile" ] || { printf 'in %s steht kein .dmg\n' "$datei" >&2; return 1; }

  printf '%s' "${zeile%% *}"
}

# belegschreiben legt den Beleg an — mit dem, was die Maschine weiß.
#
# **Die Handprüfungen stehen als offene Haken darin**, nicht als erledigt. Ein
# Beleg, der sie vorab abhakt, ist schlimmer als keiner.
belegschreiben() { # belegschreiben <ziel> <fassung> <commit> <hash> <pruefsummen> [fingerabdruck] [lage]
  local ziel="$1" fassung="$2" commit="$3" hash="$4" pruefsummen="$5"
  local fingerabdruck="${6-}" lage="${7-}"

  {
    printf '# Freigabebeleg %s\n\n' "$fassung"
    printf '**Artefakt:** `%s`\n\n' "$hash"
    printf '**Commit:** `%s`\n\n' "$commit"

    # **Der Fingerabdruck gehört in den Beleg.** Ein Beleg, der „signiert"
    # sagt und nicht **womit**, lässt die eine Frage offen, für die es ihn
    # gibt (#892).
    if [ -n "$fingerabdruck" ]; then
      printf '**Signiert mit:** `%s`\n\n' "$fingerabdruck"
      printf 'Die Signatur steht in `pruefsummen.txt.asc` und deckt die\n'
      printf 'Prüfsummen; die Prüfsummen decken das Abbild, beide\n'
      printf 'Stücklisten und `herkunft.json`. Prüfen:\n\n'
      printf '```bash\ngpg --verify pruefsummen.txt.asc pruefsummen.txt \\\n'
      printf '  && shasum -c pruefsummen.txt\n```\n\n'
    fi

    printf 'Dieser Beleg gilt für **dieses** Abbild. Wer neu baut, bekommt ein\n'
    printf 'neues und braucht einen neuen Beleg — auch bei gleichem Quelltext.\n\n'
    printf -- '---\n\n## Prüfsummen\n\n```\n'
    cat "$pruefsummen"
    printf '```\n\n'
    printf -- '---\n\n## Was die Maschine geprüft hat\n\n'
    printf -- '- [x] `make ci` — alle Stufen grün (siehe unten)\n'
    printf -- '- [x] Bündel gebaut, ad hoc signiert, mit Hardened Runtime\n'
    printf -- '- [x] Stücklisten und Prüfsummen erzeugt\n\n'

    # **Die Sicherheitslage steht mit ihrem Wort da** — ENG-04.AK-03 (#1729).
    # Ein Scanner, der nicht antwortete, ist **nicht** grün; der Beleg sagt,
    # welcher der vier Zustände es war (`scripts/sicherheitslage.sh`).
    printf -- '---\n\n## Sicherheitslage\n\n'
    if [ -n "$lage" ] && [ -s "$lage" ]; then
      sed 's/^/- /' "$lage"
      printf '\n'
    else
      printf '**Sicherheitslage: nicht erhoben** — dieser Beleg sagt nichts über\n'
      printf 'bekannte Schwachstellen.\n\n'
    fi
    printf -- '---\n\n## Was von Hand geprüft wird\n\n'
    printf 'Jede Zeile bekommt Datum und Namen, sobald sie gemacht ist —\n'
    printf 'über `scripts/releasebeleg.sh haken <beleg> <hash> <nummer> <wer>`.\n'
    printf 'Ohne den passenden Artefakthash weist der Befehl ab.\n\n'
    printf -- '- [ ] **1 Fremdpaketlauf** — `go test -tags fremdpakete ./...` gegen die bezogenen Pakete\n'
    printf -- '- [ ] **2 Kanonlauf** — der Abstand zu MSE2 über die Stichprobe, mit der Zahl daneben\n'
    printf -- '- [ ] **3 Sichtprüfung hell**\n'
    printf -- '- [ ] **4 Sichtprüfung dunkel**\n'
    printf -- '- [ ] **5 Druckbogen** — wirklich gedruckt, nicht nur erzeugt\n'
    printf -- '- [ ] **6 Erststart** — `.dmg` eingehängt, App aus dem Abbild geöffnet, Gatekeeper-Weg gegangen\n'
    printf -- '- [ ] **7 Bestand** — nach dem Start stehen Sets und Karten da\n\n'
    printf -- '---\n\n## Eingetragen\n\n'
  } > "$ziel"
}

# belegzeile hängt eine erledigte Handprüfung an — nach Prüfung des Hashes.
#
# **Der Hash steht im Aufruf**, damit niemand einen Beleg für ein anderes
# Abbild fortschreibt. Passt er nicht, wird abgewiesen.
belegzeile() { # belegzeile <beleg> <hash> <nummer> <wer> <datum>
  local beleg="$1" hash="$2" nummer="$3" wer="$4" datum="$5"

  [ -w "$beleg" ] || { printf 'der Beleg %s ist nicht beschreibbar\n' "$beleg" >&2; return 1; }

  if ! grep -qF "$hash" "$beleg"; then
    printf 'dieser Beleg gilt nicht für %s — er nennt ein anderes Artefakt\n' "$hash" >&2
    return 1
  fi

  case "$nummer" in
    [1-7]) ;;
    *) printf 'die Prüfung %s gibt es nicht — es sind 1 bis 7\n' "$nummer" >&2; return 1 ;;
  esac

  # **Das Kästchen wird angekreuzt, nicht nur die Zeile angehängt.**
  #
  # Bis zum 2026-08-26 geschah nur das zweite, und der Beleg sagte beides: oben
  # sieben offene Kästchen, unten die erledigten Prüfungen. Wer nur den Kopf
  # las, las das Falsche. Aufgefallen beim Eintragen der ersten drei Prüfungen
  # für 0.2.0.
  #
  # **Nur die genannte Zeile**, erkannt an ihrer Nummer hinter `**`. Ein
  # `s/\[ \]/[x]/` ohne Anker träfe alle sieben — und ein Beleg, der alles
  # abhakt, ist schlimmer als einer, der nichts abhakt.
  local muster="s/^- \\[ \\] \\*\\*${nummer} /- [x] **${nummer} /"

  # **In eine Nachbardatei und dann umbenannt** — dieselbe Überlegung wie beim
  # atomaren Schreiben: ein abgebrochenes `sed -i` ließe einen halben Beleg
  # zurück.
  local nachbar
  nachbar="$(mktemp "${beleg%.md}.XXXXXX")" || return 1

  sed "$muster" "$beleg" > "$nachbar" && mv "$nachbar" "$beleg" || {
    rm -f "$nachbar"
    return 1
  }

  printf -- '- **%s** · Prüfung %s · %s\n' "$datum" "$nummer" "$wer" >> "$beleg"
}

# **Als Kommando aufrufbar** — `releasebeleg.sh haken <beleg> <hash> <nr> <wer>`.
#
# Das Datum setzt das Skript, nicht der Aufrufer: ein Datum, das jemand
# eintippt, ist eine zweite Behauptung neben der ersten.
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
  case "${1:-}" in
    haken)
      shift
      [ $# -eq 4 ] || {
        printf 'Aufruf: releasebeleg.sh haken <beleg> <artefakthash> <1..7> "<wer>"\n' >&2
        exit 2
      }
      belegzeile "$1" "$2" "$3" "$4" "$(date +%Y-%m-%d)"
      ;;
    *)
      printf 'Aufruf: releasebeleg.sh haken <beleg> <artefakthash> <1..7> "<wer>"\n' >&2
      exit 2
      ;;
  esac
fi
