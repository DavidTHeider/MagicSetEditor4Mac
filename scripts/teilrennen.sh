# shellcheck shell=bash
#
# **Der Racelauf eines großen Pakets in Teilen** — ENG-03
# (#1728).
#
# Eingebunden von `scripts/ci.sh`; geprüft in `scripts/teilrennen.test.sh`.
#
# # Warum
#
# **`internal/app` war die Teststufe.** Gemessen an drei Läufen vom
# 2026-09-24: 1647, 1566 und 1489 Sekunden unter `-race` — bei einer Frist von
# 1800 —, während alle anderen Pakete nach höchstens 517 fertig waren. Danach
# arbeitete ein Kern von zehn. **In sechs Teilen: 511 Sekunden**, allein
# gemessen, mit denselben 19 ungelaufenen Anweisungen wie ungeteilt. Der Grund: `go test` fährt ein Paket in **einem** Prozess,
# und seine Tests ohne `t.Parallel` laufen darin nacheinander.
#
# **Die Teilung ändert nicht, was geprüft wird**, nur wie viele Prozesse es
# prüfen: dasselbe Testprogramm, einmal übersetzt, mit Race-Detektor und
# Abdeckung `atomic`, im Verzeichnis des Pakets wie bei `go test` — und jeder
# Test in genau einem Teil.
#
# # Was sie kostet
#
# **Ein Wettlauf zwischen zwei Tests verschiedener Teile wird nicht gesehen** —
# sie laufen in zwei Prozessen. Zwischen zwei Tests ohne `t.Parallel` sieht der
# Detektor ohnehin keinen, denn sie laufen nacheinander; bei den nebenläufigen
# fällt er für die Paare weg, die in verschiedenen Teilen landen.

# teile_starten übersetzt das Testprogramm jedes Pakets aus GETEILT **einmal**
# und startet es in TEILE Prozessen im Hintergrund — Teil *i* bekommt jeden
# TEILE-ten Test der geordneten Liste. `teile_abwarten` sammelt ein.
teile_starten() { # teile_starten <raum>
  local raum="$1" paket name verzeichnis teil muster
  mkdir -p "$raum"

  # shellcheck disable=SC2086 # GETEILT ist eine Liste von Mustern
  for paket in $(go list $GETEILT); do
    name="$(basename "$paket")"
    verzeichnis="$(go list -f '{{.Dir}}' "$paket")"

    local bau_anfang
    bau_anfang="$(date +%s)"
    if ! go test -c -race -cover -covermode=atomic -o "$raum/$name.test" "$paket" > "$raum/$name-bau.log" 2>&1; then
      printf '1\n' > "$raum/$name-bau.rc"

      continue
    fi
    printf '%s\n' "$(($(date +%s) - bau_anfang))" > "$raum/$name-bau.dauer"

    # **Die Liste in eine eigene Datei**, auch ihre Meldungen: ein Programm
    # mit Abdeckung klagt beim bloßen Auflisten über ein fehlendes
    # `GOCOVERDIR`, und scheitert es, gehört die Meldung zum Bau.
    if ! (cd "$verzeichnis" && "$raum/$name.test" -test.list '.*') > "$raum/$name.namen" 2> "$raum/$name-liste.log"; then
      cp "$raum/$name-liste.log" "$raum/$name-bau.log"
      printf '1\n' > "$raum/$name-bau.rc"

      continue
    fi

    for teil in $(seq 1 "$TEILE"); do
      muster="$(awk -v teil="$teil" -v teile="$TEILE" '(NR - 1) % teile == teil - 1' "$raum/$name.namen" | paste -sd'|' -)"

      # **`-test.v` in die Datei**: gezeigt wird sie nur, wenn der Teil rot
      # ist — dann aber mit jedem Lauf, der dazu gehört.
      (
        rueckgabe=0
        anfang="$(date +%s)"
        cd "$verzeichnis" &&
          "$raum/$name.test" -test.run "^(${muster})\$" -test.v -test.timeout "${TEILFRIST:-30m}" \
            -test.coverprofile="$raum/$name-$teil.out" > "$raum/$name-$teil.log" 2>&1 ||
          rueckgabe=$?
        printf '%s\n' "$(($(date +%s) - anfang))" > "$raum/$name-$teil.dauer"
        printf '%s\n' "$rueckgabe" > "$raum/$name-$teil.rc"
      ) &
    done
  done
}

# panikort nennt, wo eine Panik geschah: den Test, der zuletzt anfing, und
# die Rahmen der Goroutine, die starb.
#
# **Weil `grep` über die Zeilen allein es verschluckt.** Eine Panik in einer
# fremden Goroutine hat kein `--- FAIL` — das Testprogramm endet, bevor der
# Test fertig ist —, und ihr Stapel steht in Zeilen ohne Muster. Am
# 2026-09-24 sagte der Bericht deshalb nur „slice bounds out of range
# [1:0]", und die Protokolle der Teile waren schon gelöscht.
panikort() { # panikort <protokoll>
  grep -q '^panic' "$1" || return 0

  printf '        zuletzt angefangen: %s\n' \
    "$(grep -E '^=== (RUN|CONT) ' "$1" | tail -1 | awk '{print $NF}')"

  # **Bei einer Zeitüberschreitung, was noch lief** — Go schreibt es selbst
  # unter „running tests:". Der zuletzt begonnene Test ist dann oft ein
  # unschuldiger: am 2026-10-04 einer, der allein in vier Sekunden grün ist,
  # während ein anderer seit einer halben Stunde wartete.
  awk '/^[[:space:]]+running tests:/ { an = 1; next }
       an && /^[[:space:]][[:space:]]+[A-Za-z]/ { sub(/^[[:space:]]+/, ""); print "        noch lief: " $0; next }
       an { exit }' "$1"

  # Die ersten Rahmen nach der Kopfzeile der sterbenden Goroutine: je eine
  # Funktion und ihre Zeile. `testing.` und `runtime.` sagen nichts über den
  # Fund und fallen weg.
  awk '/^goroutine [0-9]+ \[running\]:/ { an = 1; next } an && NF == 0 { exit } an' "$1" |
    grep -vE '^(runtime|testing)\.|^[[:space:]]+/.*/src/(runtime|testing)/' |
    head -16 | sed 's/^/        /'
}

# teile_abwarten wartet auf jeden Teil, berichtet und legt die Profile zu
# einem zusammen. Rückgabe 0 nur, wenn jeder Teil grün war.
#
# **Jeder Block einmal**: dasselbe Paket in mehreren Teilen meldet jeden
# Block mehrmals. Aneinandergehängt zählte jede Anweisung so oft, wie es Teile
# gibt, und die Abdeckung wäre das Mittel der Teile — gemessen am 2026-09-24
# 64 bis 77 Prozent je Teil, wo das Ganze 99,9 hat. Die Zahlen eines Blocks
# addieren sich, so wie `atomic` es in einem Prozess täte.
teile_abwarten() { # teile_abwarten <raum> <profil>
  local raum="$1" profil="$2" rot=0 datei teil
  wait

  for datei in "$raum"/*-bau.rc; do
    [ -f "$datei" ] || continue
    printf '      %s ließ sich nicht übersetzen:\n' "$(basename "$datei" -bau.rc)"
    sed 's/^/        /' "${datei%.rc}.log"
    rot=1
  done

  local gelaufen=0

  for datei in "$raum"/*-[0-9]*.rc; do
    [ -f "$datei" ] || continue
    teil="${datei%.rc}"
    gelaufen=$((gelaufen + 1))

    if [ "$(cat "$datei")" = "0" ]; then
      printf '      ok  %s — %s\n' "$(basename "$teil")" "$(grep -E '^coverage:' "$teil.log" | tail -1)"
    else
      printf '      FAIL  %s\n' "$(basename "$teil")"
      grep -E -- '--- FAIL|^panic|^FAIL|WARNING: DATA RACE|^[[:space:]]+[a-z_]+_test\.go:[0-9]+:' "$teil.log" |
        head -60 | sed 's/^/        /'
      panikort "$teil.log"
      rot=1
    fi
  done

  # **Kein Teil ist kein grüner Lauf** — dann hat nichts geprüft.
  if [ "$gelaufen" -eq 0 ]; then
    printf '      kein Teil ist gelaufen\n'

    return 1
  fi

  awk 'FNR == 1 && /^mode:/ { modus = $0; next }
       { anweisungen[$1] = $2; zahl[$1] += $3 }
       END { print modus; for (block in zahl) print block, anweisungen[block], zahl[block] }' \
    "$raum"/*-[0-9]*.out > "$profil" 2>/dev/null || rot=1

  return "$rot"
}

# teile_belegen schreibt die Zeiten eines Laufs in eine Tabelle — T08
# (#2101):
# je Paket die Übersetzung (`bau`), je Teil die Wandzeit (`teil`) und je Test
# seine Dauer aus `-test.v` (`test`), in Sekunden.
#
# **Auch nach einem grünen Lauf.** Bis zum 2026-10-06 gingen die Protokolle
# der Teile danach weg, und mit ihnen jede Zahl, an der sich eine schnellere
# Teststufe messen ließe. **Was sie nicht hält**: die Pakete außerhalb der
# Teilung — für sie steht die Gesamtzeit im Protokoll (`ok … 12.3s`).
teile_belegen() { # teile_belegen <raum> <ziel>
  local raum="$1" ziel="$2" datei name teil
  {
    for datei in "$raum"/*-bau.dauer; do
      [ -f "$datei" ] || continue
      printf 'bau\t%s\t-\t-\t%s\n' "$(basename "$datei" -bau.dauer)" "$(cat "$datei")"
    done
    for datei in "$raum"/*-[0-9]*.dauer; do
      [ -f "$datei" ] || continue
      teil="$(basename "$datei" .dauer)"
      name="${teil%-*}"
      printf 'teil\t%s\t%s\t-\t%s\n' "$name" "${teil##*-}" "$(cat "$datei")"
      sed -nE "s/^[[:space:]]*--- (PASS|FAIL|SKIP): ([^ ]+) \(([0-9.]+)s\)$/test\t$name\t${teil##*-}\t\2\t\3/p" \
        "${datei%.dauer}.log"
    done
  } > "$ziel"
}

# teile_aufbewahren legt den Dauerbeleg eines Laufs unter `<ort>` ab — benannt
# nach Zeit und Commit, damit die Namen sich wie die Zeit ordnen — und lässt
# die jüngsten `TEILBELEGE` (20) stehen.
teile_aufbewahren() { # teile_aufbewahren <raum> <ort>
  local raum="$1" ort="$2" ziel
  mkdir -p "$ort" || return 1
  ziel="$ort/$(date -u +%Y%m%dT%H%M%SZ)-$(git rev-parse --short=12 HEAD 2>/dev/null || echo ohne-commit).tsv"
  teile_belegen "$raum" "$ziel" || return 1
  find "$ort" -maxdepth 1 -name '*.tsv' | sort |
    awk -v behalten="${TEILBELEGE:-20}" '{ zeile[NR] = $0 } END { for (i = 1; i <= NR - behalten; i++) print zeile[i] }' |
    while IFS= read -r alt; do rm -f "$alt"; done
  printf '      Dauerbeleg: %s\n' "${ziel#"$PWD"/}"
}
