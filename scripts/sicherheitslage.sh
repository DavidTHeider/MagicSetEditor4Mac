# shellcheck shell=bash
#
# **Die Sicherheitslage in vier Wörtern** — ENG-04.AK-03
# (#1729).
#
# Eingebunden von `scripts/ci.sh` (Stufe `security`) und `scripts/release.sh`
# (Freigabebeleg); geprüft in `scripts/sicherheitslage.test.sh`.
#
# # Warum es mehr als grün und rot braucht
#
# **Ein Scanner, der nicht antwortet, hat nichts geprüft** — und das ist
# etwas anderes als „ohne Befund". AK-03 verlangt es wörtlich: nicht
# erreichbar, erfolgreich ohne Befund und bestätigte Schwachstelle sind
# verschiedene Zustände. Dazu kommt ein vierter, den die drei nicht fassen:
# ein Lauf, der aus einem anderen Grund scheitert (kaputte Sperrdatei, Code,
# der sich nicht laden lässt).
#
#   ohne-befund       der Scanner hat geantwortet, und nichts liegt über der
#                     Schwelle
#   befund            eine Schwachstelle — govulncheck: im erreichbaren Code,
#                     npm audit: ab `high`
#   nicht-erreichbar  die Datenbank oder die Registry hat nicht geantwortet
#   gescheitert       der Lauf ist aus einem anderen Grund gescheitert
#
# **Nachgelesen, nicht erschlossen**: govulncheck v1.6.0 endet mit 0 ohne
# Befund, 3 mit Befund und 1 bei einem Fehler (`internal/scan/errors.go`,
# `text.go`); npm 11.12.1 schreibt bei einem Ausfall des Audit-Endpunkts
# Meldung und Statuscode statt eines Berichts (`lib/utils/audit-error.js`).

# netzmuster erkennt eine Meldung, die von einem Ausfall des Netzes spricht.
netzmuster='dial tcp|no such host|i/o timeout|connection refused|TLS handshake|ENOTFOUND|EAI_AGAIN|ECONNREFUSED|ECONNRESET|ETIMEDOUT|audit endpoint returned an error|"statusCode":[[:space:]]*5[0-9][0-9]|Service Unavailable|fetching vulnerabilities'

# govuln_zustand ordnet einen Lauf von govulncheck ein.
govuln_zustand() { # govuln_zustand <rückgabe> <ausgabedatei>
  local rueckgabe="$1" ausgabe="$2"

  case "$rueckgabe" in
    0) printf 'ohne-befund\n' ;;
    3) printf 'befund\n' ;;
    *)
      if grep -Eq "$netzmuster" "$ausgabe" 2>/dev/null; then
        printf 'nicht-erreichbar\n'
      else
        printf 'gescheitert\n'
      fi
      ;;
  esac
}

# npmaudit_zustand ordnet einen Lauf von `npm audit --json` ein.
#
# **Die Schwelle ist `high`**, wie in der Stufe `security` (HAERTUNG.md): ein
# mittlerer Fund steht im Bericht, macht aber keinen Befund.
#
# **Der Rückgabewert allein entscheidet nicht**: npm endet bei einem Befund wie
# bei einem Ausfall mit 1, und ohne Bericht hat niemand etwas geprüft — auch
# nicht bei 0. **Widersprechen darf er dem Bericht aber nicht** (Befund 4 des
# Zweitblicks über Phase 6.2): npm 11.12.1 endet mit 1 genau dann, wenn eine
# Zahl ab der Schwelle darin steht (`npm-audit-report` 7.0.0,
# `lib/exit-code.js`). Eine 1 zu einem Bericht ohne solche Zahl heißt, dass
# vorher etwas anderes scheiterte.
npmaudit_zustand() { # npmaudit_zustand <rückgabe> <ausgabedatei>
  local rueckgabe="$1" ausgabe="$2" hoch

  # **Ein Bericht wird gelesen, nicht nach Wörtern durchsucht.** Nur wenn
  # keiner kam — oder einer mit Fehlerangabe —, entscheidet die Meldung.
  if hoch="$(node -e '
    const bericht = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8"))
    const zahlen = bericht?.metadata?.vulnerabilities
    if (!zahlen || bericht.error) process.exit(2)
    console.log((zahlen.high ?? 0) + (zahlen.critical ?? 0))
  ' "$ausgabe" 2>/dev/null)"; then
    # **Ein genannter Fund bleibt ein Befund**, auch wenn niemand rot wurde.
    if [ "$hoch" -gt 0 ]; then
      printf 'befund\n'
    elif [ "$rueckgabe" -ne 0 ]; then
      printf 'gescheitert\n'
    else
      printf 'ohne-befund\n'
    fi

    return
  fi

  if grep -Eq "$netzmuster" "$ausgabe" 2>/dev/null; then
    printf 'nicht-erreichbar\n'
  else
    printf 'gescheitert\n'
  fi
}

# govuln_kennungen gibt die Kennungen der Schwachstellen eines Laufs — jede
# einmal, geordnet. Der Beleg soll sagen, **welche**, nicht nur dass.
govuln_kennungen() { # govuln_kennungen <ausgabedatei>
  { grep -oE 'GO-[0-9]{4}-[0-9]+' "$1" 2>/dev/null || true; } | sort -u | paste -sd, - | sed 's/,/, /g'
}

# npmaudit_zahlen gibt die Zahlen je Schwere aus einem Bericht von
# `npm audit --json` — oder nichts, wenn keiner kam.
#
# **Auch unter der Schwelle**: ein mittlerer Fund macht keinen Befund, aber er
# gehört in die Auskunft. Die Stufe `security` ließ drei solche Funde durch,
# ohne dass sie irgendwo standen (#1729, 2026-09-21).
npmaudit_zahlen() { # npmaudit_zahlen <ausgabedatei>
  # shellcheck disable=SC2016 # `${…}` gehört JavaScript, nicht der Shell
  node -e '
    const bericht = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8"))
    const z = bericht?.metadata?.vulnerabilities
    if (!z || bericht.error) process.exit(2)
    console.log(`kritisch ${z.critical ?? 0} · hoch ${z.high ?? 0} · mittel ${z.moderate ?? 0} · niedrig ${z.low ?? 0}`)
  ' "$1" 2>/dev/null || true
}

# sicherheitslage fährt beide Scanner und schreibt je eine Zeile — den Zustand
# und dahinter, was ihn belegt: die Kennungen, die Zahlen oder den Grund.
sicherheitslage() { # sicherheitslage [go-pakete…]
  local raum rueckgabe zustand zusatz
  raum="$(mktemp -d)"

  rueckgabe=0
  govulncheck "$@" > "$raum/govuln" 2>&1 || rueckgabe=$?
  zustand="$(govuln_zustand "$rueckgabe" "$raum/govuln")"
  if [ "$zustand" = "befund" ]; then
    zusatz="$(govuln_kennungen "$raum/govuln")"
  else
    zusatz="$(grund "$zustand" "$raum/govuln")"
  fi
  printf 'govulncheck: %s%s\n' "$zustand" "${zusatz:+ — $zusatz}"

  rueckgabe=0
  (cd frontend && npm audit --json --audit-level=high) > "$raum/npm" 2>&1 || rueckgabe=$?
  zustand="$(npmaudit_zustand "$rueckgabe" "$raum/npm")"
  zusatz="$(npmaudit_zahlen "$raum/npm")"
  if [ "$zustand" = "gescheitert" ] && [ -n "$zusatz" ]; then
    zusatz="Rückgabe $rueckgabe zu einem Bericht ohne Fund ab high — $zusatz"
  fi
  [ -n "$zusatz" ] || zusatz="$(grund "$zustand" "$raum/npm")"
  printf 'npm audit: %s%s\n' "$zustand" "${zusatz:+ — $zusatz}"

  rm -rf "$raum"
}

# grund gibt die erste sprechende Zeile eines Laufs, der nicht ohne Befund war.
grund() { # grund <zustand> <ausgabedatei>
  [ "$1" = "ohne-befund" ] && return 0

  grep -Ev '^[[:space:]]*$' "$2" | head -1 | cut -c1-160

  return 0
}

# freigabeurteil sagt, ob eine Freigabe mit dieser Lage weitergeht — und
# wenn nicht, warum. Befund 1 des Zweitblicks über Phase 6.2.
#
# **Weiter geht es nur ohne Befund und mit Antwort.** Hier brach der
# Freigabeweg allein bei `befund` ab; ein erneuter Scan, der scheiterte, lief
# bis „Fertig" durch — und die grüne Pipeline davor belegt diesen späteren
# Lauf nicht. In der Stufe `security` ist `gescheitert` rot, hier jetzt auch.
#
# **Ein Scanner ohne Antwort hält sie ebenso an, es sei denn, jemand
# entscheidet es**: `MSE_FREIGABE_OHNE_SCANNER=1`, wie `MSE_CI_UEBERSPRUNGEN_ERLAUBT`
# in der Pipeline. Dann steht die Entscheidung als eigene Zeile in der Lage
# und damit im Beleg — eine hingenommene Lücke ist eine andere Aussage als
# eine unbemerkte. **Einen Befund übergeht sie nicht**: sie gilt der Lücke,
# nicht der Schwachstelle.
freigabeurteil() { # freigabeurteil <lagedatei>
  local lage="$1"

  # **Keine Antwort ist kein grüner Lauf.** `sicherheitslage` schreibt immer
  # je eine Zeile für beide Scanner; fehlt eine — oder die Datei —, ist der
  # Lauf unterwegs gestorben, und keine der Abfragen darunter fände etwas.
  if ! lage_vollstaendig "$lage"; then
    printf 'die Lage nennt nicht beide Scanner — sie ist nicht erhoben\n'

    return 1
  fi

  if grep -q ': befund' "$lage"; then
    printf 'eine bekannte Schwachstelle — sie gehört behoben oder begründet, bevor freigegeben wird\n'

    return 1
  fi

  if grep -q ': gescheitert' "$lage"; then
    printf 'ein Scanner ist gescheitert — ohne seine Antwort ist die Lage unbekannt\n'

    return 1
  fi

  if grep -q ': nicht-erreichbar' "$lage"; then
    if [ "${MSE_FREIGABE_OHNE_SCANNER:-}" != "1" ]; then
      printf 'ein Scanner war nicht erreichbar — später erneut, oder ausdrücklich: MSE_FREIGABE_OHNE_SCANNER=1\n'

      return 1
    fi

    printf 'Entscheidung: freigegeben, obwohl ein Scanner nicht antwortete (MSE_FREIGABE_OHNE_SCANNER=1)\n' |
      tee -a "$lage"
  fi

  return 0
}

# lage_vollstaendig sagt, ob eine Lage je Scanner **genau eine** Zeile mit
# einem der vier Zustände trägt — und sonst nichts außer der Entscheidung,
# die `freigabeurteil` selbst anhängt.
#
# **Der Anfang einer Zeile genügte nicht** (Befund 1 des Nachlaufs über Phase
# 6.2): `govulncheck: ` ohne Zustand, ein unbekanntes Wort oder eine zweite
# Zeile für denselben Scanner trafen keines der Sperrmuster, und die Freigabe
# ging weiter.
lage_vollstaendig() { # lage_vollstaendig <lagedatei>
  local zeile go=0 npm=0
  local muster='^(govulncheck|npm audit): (ohne-befund|befund|nicht-erreichbar|gescheitert)( — .*)?$'

  [ -r "$1" ] || return 1

  while IFS= read -r zeile || [ -n "$zeile" ]; do
    case "$zeile" in
      "Entscheidung: "*) continue ;;
    esac

    [[ "$zeile" =~ $muster ]] || return 1

    case "$zeile" in
      govulncheck:*) go=$((go + 1)) ;;
      *) npm=$((npm + 1)) ;;
    esac
  done < "$1"

  [ "$go" -eq 1 ] && [ "$npm" -eq 1 ]
}
