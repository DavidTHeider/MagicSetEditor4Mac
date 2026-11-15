#!/usr/bin/env bash
#
# Die Abdeckungsprüfung der Stufe `test` (AGENTS.md §3).
#
# Ausgelagert aus scripts/ci.sh wegen der Zeilengrenze aus §2.
# Wird von scripts/ci.sh eingebunden und nutzt dessen ok/fehl/weg.
#
# ## Warum .covignore hier gelesen wird
#
# Die Datei stand von Anfang an da und wurde von niemandem gelesen — sie war
# eine Beschreibung, die wie eine Regel aussah. Wer eine Ausnahme eintrug, hatte
# keine, und wer keine eintrug, bekam trotzdem eine: `main.go` fiel in die
# Gesamtabdeckung ein und drückte sie unter die Schwelle, obwohl die Begründung
# dafür längst dastand.
#
# Jetzt gilt: was hier ausgenommen wird, steht in .covignore, und was in
# .covignore steht, braucht eine Begründung. Beides wird nachgezählt.

# Kernpakete mit harter 100-Prozent-Schwelle.
#
# **`aktualisierung` kam am 2026-08-25 dazu** (#1100). Es ist ein Blatt wie
# `scryfall` und wird wie eines gehalten: ein Klient, der nur zum Teil geprüft
# ist, prüft nichts.
#
# **`sicherung` und `vorschauweg` kamen am 2026-08-28 dazu** (#1270), und der
# erste von beiden war schon seit dem 2026-08-26 da und stand nicht hier — das
# ist die Lücke, die diese Liste hat und die niemand sieht. Sie ist jetzt
# nachgezählt: `internal/werkzeuge/kernpakete_test.go` hält sie gegen die
# Pakete unter `internal/`, und wer eines auslässt, begründet es dort.
#
# **`laufbeleg` kam am 2026-09-22 dazu** (QA-01, #1733). Es ist kein Teil der
# Anwendung, sondern das Werkzeug, das Messläufe belegt — und ein Beleger,
# der nur zum Teil geprüft ist, belegt nichts.
#
# **`fenster` kam am 2026-09-24 dazu** (ENG-02, #1727): die Hülle um Wails 3.
# Was eine laufende Anwendung braucht, steht allein in `wails.go` und ist in
# `.covignore` begründet; alles, was darüber entscheidet, wird gemessen.
KERNPAKETE="laufbeleg austausch entdeckung fernleitung leitung aktualisierung anmeldung anthropic applefm openai xai bezug biometrie prozessspeicher cardtrader kennwort ollama peer totp domain service storage app fenster mcp msepaket msescript mtgapi mtggraphql mtgjson mtgjsonpreise netz pdf render rest sammelliste schluesselbund scrydex scryfall sicherung svg vcard vorschauweg"
SCHWELLE_GESAMT=90

# covignore_pruefen weist Ausnahmen ohne Begründung ab.
#
# Eine Ausnahme ohne Grund ist eine stille Absenkung der Schwelle: sie sieht aus
# wie eine Entscheidung und ist keine. Der Kopf von .covignore sagt das zu —
# hier wird es eingelöst.
#
# **Getrennt von covignore_pfade, und das ist der Punkt.** Dort liefe `fehl`
# in einer Kommandosubstitution — also in einer Subshell: die Meldung
# erschiene, `FEHLER=1` bliebe darin zurück, und die Pipeline meldete grün.
# Ein Wächter, der meldet und nicht hält, ist schlimmer als keiner: er sieht
# aus wie eine Zusage.
covignore_pruefen() {
  [ -f .covignore ] || return 0

  local zeile pfad begruendung
  while IFS= read -r zeile; do
    case "$(printf '%s' "$zeile" | tr -d '[:space:]')" in
      '' | '#'*) continue ;;
    esac

    pfad="$(printf '%s' "$zeile" | sed 's/#.*//' | xargs)"
    begruendung="$(printf '%s' "$zeile" | sed -n 's/.*#[[:space:]]*//p' | xargs)"

    if [ -z "$begruendung" ]; then
      fehl ".covignore: $pfad ist ohne Begründung ausgenommen (AGENTS.md §3)"
    fi
  done < .covignore
}

# covignore_pfade gibt die ausgenommenen Pfade aus — nur die mit Begründung.
# Läuft in einer Subshell; hier darf deshalb nichts stehen, was FEHLER setzt.
covignore_pfade() {
  [ -f .covignore ] || return 0

  local zeile pfad begruendung
  while IFS= read -r zeile; do
    case "$(printf '%s' "$zeile" | tr -d '[:space:]')" in
      '' | '#'*) continue ;;
    esac

    pfad="$(printf '%s' "$zeile" | sed 's/#.*//' | xargs)"
    begruendung="$(printf '%s' "$zeile" | sed -n 's/.*#[[:space:]]*//p' | xargs)"

    [ -n "$begruendung" ] && printf '%s\n' "$pfad"
  done < .covignore

  return 0
}

# profil_ohne_ausnahmen schreibt das Abdeckungsprofil ohne die ausgenommenen
# Pfade nach stdout.
profil_ohne_ausnahmen() {
  local pfade muster
  pfade="$(covignore_pfade)"

  if [ -z "$pfade" ]; then
    cat .abdeckung.out

    return 0
  fi

  # Die Pfade im Profil tragen den Modulpfad als Präfix; verglichen wird
  # deshalb auf Vorkommen, nicht auf Gleichheit. Frontend-Einträge stehen in
  # derselben Datei, treffen hier aber auf nichts — ihre Schwelle setzt vitest
  # durch. Das ist Absicht: eine Ausnahmeliste, zwei Durchsetzungen.
  muster="$(printf '%s\n' "$pfade" | sed 's/[].[^$\\*/]/\\&/g' | paste -sd'|' -)"

  grep -Ev "($muster)" .abdeckung.out
}

abdeckung_pruefen() {
  [ -f .abdeckung.out ] || { fehl "keine Abdeckungsdatei"; return 0; }

  covignore_pruefen

  local gefiltert
  gefiltert="$(mktemp)"
  profil_ohne_ausnahmen > "$gefiltert"

  # Ein Profil ohne eine einzige messbare Anweisung ist weder rot noch grün.
  #
  # Das als Fehler zu werten hielte die Pipeline ohne Befund an; als grün
  # durchzuwinken behauptete eine Messung, die nicht stattgefunden hat. Also
  # sichtbar übersprungen — sobald es Produktivcode gibt, greift die Schwelle
  # ohne weiteres Zutun.
  local anweisungen
  anweisungen="$(awk '!/^mode:/ { summe += $2 } END { print summe + 0 }' "$gefiltert")"
  if [ "$anweisungen" -eq 0 ]; then
    weg "keine messbaren Anweisungen außerhalb der Ausnahmen"
    rm -f "$gefiltert"

    return 0
  fi

  # Aus **Anweisungszahlen**, nicht aus gemittelten Prozenten.
  #
  # `go tool cover -func` gibt je Funktion einen auf eine Nachkommastelle
  # gerundeten Prozentwert. Wer die mittelt und mit 100 vergleicht, lässt 99,95
  # als „100.0" durch — und mittelt außerdem eine einzeilige Funktion gegen eine
  # mit vierzig Anweisungen. Das Profil selbst nennt je Block die Zahl der
  # Anweisungen und wie oft er gelaufen ist; damit ist die Rechnung exakt.
  local gelaufen alle anzeige
  read -r gelaufen alle <<< "$(zaehlung "$gefiltert" '')"
  anzeige="$(prozent "$gelaufen" "$alle")"

  # Kreuzmultiplikation statt Division: `gelaufen/alle >= schwelle/100`.
  if [ "$((gelaufen * 100))" -ge "$((alle * SCHWELLE_GESAMT))" ]; then
    ok "Gesamtabdeckung ${anzeige} % — ${gelaufen} von ${alle} Anweisungen (Schwelle ${SCHWELLE_GESAMT} %, ohne .covignore)"
  else
    fehl "Gesamtabdeckung ${anzeige} % (${gelaufen} von ${alle}) unter der Schwelle von ${SCHWELLE_GESAMT} %"
    go tool cover -func="$gefiltert" | awk '$3 != "100.0%" { print "      " $0 }' | head -15
  fi

  kernpakete_pruefen "$gefiltert"
  rm -f "$gefiltert"
}

# zaehlung gibt „<gelaufen> <gesamt>" der Anweisungen aus.
#
# $2 ist ein Pfadteil zum Filtern; leer heißt: alles. Zwei ganze Zahlen und
# kein Prozentwert — **entschieden wird ganzzahlig**, gerundet wird nur für die
# Ausgabe. Der Nachlauf des Zweitblicks hat gezeigt, warum das nicht dasselbe
# ist: 19.999 von 20.000 Anweisungen ergeben formatiert „100.00", und ein
# Vergleich gegen diese Zeichenkette nähme sie an.
zaehlung() {
  awk -v teil="$2" '
    !/^mode:/ && (teil == "" || index($1, teil) > 0) {
      gesamt += $2
      if ($3 > 0) { gelaufen += $2 }
    }
    END { printf "%d %d", gelaufen + 0, gesamt + 0 }
  ' "$1"
}

# prozent formatiert eine Zählung für die Ausgabe — und nur dafür.
prozent() {
  awk -v gelaufen="$1" -v gesamt="$2" \
    'BEGIN { if (gesamt == 0) { print "-" } else { printf "%.2f", 100 * gelaufen / gesamt } }'
}

# kernpakete_pruefen verlangt vollständige Abdeckung — ganzzahlig entschieden.
#
# Ein vorhandenes Kernpaket **ohne** Profildaten ist ein Fehler, kein
# Übersprung: es hieße, dass seine Tests nicht gelaufen sind, und genau dann
# wäre ein stilles Weitergehen die falsche Antwort.
kernpakete_pruefen() {
  local profil="$1" paket wert


  for paket in $KERNPAKETE; do
    [ -d "internal/$paket" ] || continue

    local gelaufen alle
    read -r gelaufen alle <<< "$(zaehlung "$profil" "internal/$paket/")"

    if [ "$alle" -eq 0 ]; then
      fehl "internal/$paket ist vorhanden, steht aber nicht im Abdeckungsprofil — sind seine Tests gelaufen?"

      continue
    fi

    # Ganzzahlig und ohne Prozentumweg: gleich viele gelaufene wie vorhandene.
    if [ "$gelaufen" -eq "$alle" ]; then
      ok "internal/$paket 100 % (${alle} Anweisungen)"
    else
      wert="$(prozent "$gelaufen" "$alle")"
      fehl "internal/$paket bei ${wert} % — ${gelaufen} von ${alle} Anweisungen, verlangt sind alle (AGENTS.md §3)"
      go tool cover -func="$profil" \
        | awk -v p="internal/$paket/" '$0 ~ p && $3 != "100.0%" { print "      " $0 }' | head -10
    fi
  done
}
