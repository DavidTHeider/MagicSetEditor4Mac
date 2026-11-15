#!/usr/bin/env bash
#
# Einzige Quelle der Wahrheit für die Integrationspipeline (AGENTS.md §4).
#
# Lokal:          make ci          bzw.  ./scripts/ci.sh all
# GitHub Actions: ./scripts/ci.sh <stufe>
#
# Beide Wege rufen dieselben Kommandos auf — damit können lokale und entfernte
# Prüfung nicht auseinanderlaufen. Wer eine Prüfung ergänzt, ergänzt sie hier
# und nicht in der Workflow-Datei.
#
# Stufen: format · types · lint · test · security · all
#
# ## Was ein Übersprung bedeutet
#
# In Phase 1 gibt es Teile des Projekts noch nicht — kein go.mod, kein Frontend.
# Eine Stufe, die deshalb nichts prüft, meldet das **sichtbar** und wird am Ende
# gezählt. Eine grüne Pipeline, die nichts geprüft hat, wäre eine Behauptung;
# die Schlusszeile nennt deshalb immer, wie viel übersprungen wurde.
#
# Sobald der jeweilige Teil da ist, greift die Prüfung ohne weiteres Zutun.

set -euo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$WURZEL"

# **Beweis und Alleinbetrieb stehen in einer eigenen Datei** (§2) — und sie
# wird **hier** eingebunden, vor der ersten Ausgabe.
# shellcheck source=scripts/ci-beweis.sh
. "$(dirname "$0")/ci-beweis.sh"

export GOPATH="$WURZEL/.go"
export GOBIN="$WURZEL/.go/bin"
export GOMODCACHE="$WURZEL/.go/mod"
export GOCACHE="$WURZEL/.go/build"
export GOLANGCI_LINT_CACHE="$WURZEL/.go/golangci-lint"

# **Die gepinnte Toolchain gilt hier und nicht nur in `activate.sh`.**
#
# Die `go`-Direktive in `go.mod` nennt eine **Untergrenze**: ist die Toolchain
# des Rechners neuer, nimmt Go sie. Am 2026-09-01 hat Homebrew mitten in einer
# Sitzung von 1.26.6 auf 1.27.0 gehoben, und `lint` wurde rot — golangci-lint
# v2.12.2 kann die Standardbibliothek von 1.27 nicht lesen. Der Quelltext war
# unverändert.
#
# **Und sie steht hier, weil der `pre-push`-Hook `activate.sh` nicht liest.**
# Stand sie nur dort, wäre der Lauf von Hand grün und der beim Pushen rot —
# zwei Pipelines mit zwei Ergebnissen, und §4 sagt, dass es eine ist.
#
# Wer Go hebt, hebt es hier, in `go.mod`, in `scripts/activate.sh` und in der
# Werkzeugtabelle von AGENTS.md — und sieht nach, ob die gepinnten Werkzeuge
# die neue Standardbibliothek lesen können.
export GOTOOLCHAIN=go1.27.4

# **Und die macOS-Untergrenze für cgo** (ENG-02) — dieselbe Datei wie beim Bau,
# sonst prüfte die Pipeline ein anderes Programm als das, das ausgeliefert wird.
# shellcheck source=scripts/macos.sh
. "$WURZEL/scripts/macos.sh"

export PATH="$GOBIN:$PATH"

# Bewusst nicht "./...": unter frontend/node_modules liegen Go-Dateien fremder
# npm-Pakete, die zum Modul gehörten und in Prüfung wie Abdeckung einflössen.
#
# `./cmd/...` steht seit #143 dabei: der MCP-Server ist ein zweites Programm
# (docs/entscheidungen/0010), und ein Composition Root, den die Pipeline nicht
# anfasst, ist genau die Stelle, an der ein Fehler unbemerkt bleibt.
GO_PAKETE="./internal/... ./cmd/... ."

# **Für Linux ohne die Hülle** (ENG-02) — warum, steht in der Datei.
# shellcheck source=scripts/linuxpakete.sh
. "$WURZEL/scripts/linuxpakete.sh"

# OHNE_RENNEN sind die Pakete, die **ohne** Race-Detektor laufen — die
# Begründung steht bei `stufe_test`, der Wächter in
# `internal/werkzeuge/rennliste_test.go`.
OHNE_RENNEN="./internal/storage/... ./internal/werkzeuge/..."

# GETEILT sind die Pakete, deren Racelauf in TEILE Prozesse geht — warum, und
# was es kostet, steht in `scripts/teilrennen.sh` (ENG-03).
#
# **`cmd/mse` kam mit QA-07 dazu** (#1738):
# am 2026-10-01 war es mit 1.168 Sekunden auf **einem** Kern der lange Pol der
# Teststufe — 259 Tests, 7 davon parallel, weil die übrigen `HOME` mit
# `t.Setenv` setzen und das nur je Prozess geht. Gemessen in `PIPELINE.md`.
GETEILT="./internal/app ./cmd/mse"
TEILE=6

# shellcheck source=scripts/teilrennen.sh
. "$WURZEL/scripts/teilrennen.sh"

# shellcheck source=scripts/testlaeufe.sh
source "$(dirname "$0")/testlaeufe.sh"

ROT=$'\033[31m'; GRUEN=$'\033[32m'; GELB=$'\033[33m'; BLAU=$'\033[34m'; AUS=$'\033[0m'
[ -t 1 ] || { ROT=''; GRUEN=''; GELB=''; BLAU=''; AUS=''; }

FEHLER=0
UEBERSPRUNGEN=0

stufe_format() {
  kopf "Format"

  leitlinien_pruefen
  npm_pinnung_pruefen
  if hat_go; then werkzeug_pinnung_pruefen; fi

  if hat_go && werkzeug_da gofumpt; then
    local dateien unformatiert
    dateien="$(git ls-files '*.go')"
    unformatiert=""
    if [ -n "$dateien" ]; then
      unformatiert="$(printf '%s\n' "$dateien" | xargs gofumpt -l)"
    fi
    if [ -n "$unformatiert" ]; then
      fehl "nicht gofumpt-konform:"
      printf '      %s\n' $unformatiert
    else
      ok "gofumpt"
    fi
  elif hat_go; then
    fehl "gofumpt nicht gefunden — 'make tools' ausführen"
  else
    weg "kein go.mod"
  fi

  if hat_frontend; then
    if (cd frontend && npm run --silent format:check >/dev/null 2>&1); then
      ok "prettier"
    else
      fehl "prettier — 'npm run format' im Verzeichnis frontend ausführen"
    fi
  else
    weg "kein Frontend (bzw. 'npm ci' fehlt)"
  fi
}

# ---------------------------------------------------------- Typisierung

stufe_types() {
  kopf "Typisierung"

  if ! hat_go; then
    weg "kein go.mod"
  else
    if go vet $GO_PAKETE 2>&1 | sed 's/^/      /'; then ok "go vet"; else fehl "go vet"; fi
    if go build $GO_PAKETE 2>&1 | sed 's/^/      /'; then ok "go build"; else fehl "go build"; fi

    # Auch für die beschriebenen, nicht gebauten Plattformen übersetzen
    # (docs/PORTIERUNG.md). `go build` allein prüft die eigene Plattform; Code
    # hinter `//go:build windows` bräche sonst unbemerkt. Ein Bau ist keine
    # Erprobung — er sagt nur, dass der Code dort überhaupt entsteht.
    for ziel in windows/amd64 windows/arm64 linux/amd64; do
      local pakete="$GO_PAKETE"
      [ "${ziel%%/*}" = linux ] && pakete="$(pakete_fuer_linux)"
      # shellcheck disable=SC2086
      if GOOS="${ziel%%/*}" GOARCH="${ziel##*/}" go build $pakete 2>&1 | sed 's/^/      /'; then
        ok "go build $ziel"
      else
        fehl "go build $ziel"
      fi
    done
  fi

  if hat_frontend; then
    if (cd frontend && npx --no-install tsc --noEmit 2>&1 | sed 's/^/      /'); then
      ok "tsc --noEmit"
    else
      fehl "tsc --noEmit"
    fi
  else
    weg "kein Frontend"
  fi
}

# --------------------------------------------------------------- Linting

# shellcheck source=scripts/ci-lint.sh
source "$(dirname "$0")/ci-lint.sh"

# ------------------------------------------------------------------ Tests

# shellcheck source=scripts/ci-abdeckung.sh
source "$(dirname "$0")/ci-abdeckung.sh"

stufe_test() {
  kopf "Tests"

  if ! hat_go; then
    weg "kein go.mod"
  else
    # **Die Go-Läufe bekommen ein leeres Zuhause**
    # (#891).
    #
    # `Paketverzeichnis()` und `Bilderverzeichnis()` zeigen unter
    # `os.UserConfigDir()`, und das ist auf darwin `$HOME/Library/…`. Ein Lauf
    # ohne eigenes Zuhause liest damit den Stand **dieses** Rechners mit — und
    # ist grün oder rot, je nachdem, was dort liegt.
    #
    # Gefunden am 2026-08-25: `TestDieZeichennahtStehtVollstaendig` war grün,
    # **weil** hier `M15-Magic-Pack` installiert ist. Auf einem frischen Klon
    # wäre er rot gewesen — und niemand hätte je davon erfahren, weil es
    # keinen entfernten Lauf gibt (Vermerk 0007).
    #
    # **Hier und nicht in jedem Test.** Ein `t.Setenv` je Datei ist eine Regel,
    # an die sich jemand erinnern muss; dies ist eine, die gilt.
    #
    # **Und die echten Orte werden davor und danach abgedruckt** — A01
    # (#2215): das Zuhause zu wechseln ist eine Absicht, dass es gewirkt hat,
    # ein Befund. Dieselbe Wache fährt `scripts/vorpruefung.sh`.
    # shellcheck source=scripts/echteorte.sh
    source "$(dirname "$0")/echteorte.sh"
    HEIM_FUER_LAUF="$(heim_fuer_laeufe)"
    local orte_vorher
    orte_vorher="$(mktemp -t ci-orte)"
    trap 'rm -rf "$HEIM_FUER_LAUF" "$orte_vorher"' RETURN
    echte_orte_merken "$orte_vorher"

    # **Die geteilten Pakete starten zuerst**, im Hintergrund — sie sind der
    # längste Lauf, und die übrigen Pakete laufen daneben.
    local teile anfang
    teile="$(mktemp -d)"
    anfang="$(date +%s)"
    HOME="$HEIM_FUER_LAUF" teile_starten "$teile"

    # **30 Minuten je Paket statt Gos zehn.** Am 2026-09-06 gemessen: `app`
    # und `storage` brauchten 601 s, und der Ausgang war kein Fehlschlag,
    # sondern ein `panic: test timed out` — der einen Lauf nannte, der zwei
    # Sekunden gedauert hatte. Die Zeit geht über das ganze Paket, die Meldung
    # zeigt auf den, der gerade dran war.
    # **Der Race-Detektor läuft, wo es Nebenläufigkeit gibt** — und dort
    # nicht, wo er einen WebAssembly-Interpreter vermisst.
    #
    # Am 2026-09-11 gemessen: `internal/storage` kostete unter `-race` **1357
    # Sekunden** gegen 28,5 ohne — das 48-fache, bei **null** Goroutinen im
    # Produktivcode. Der Grund ist nachlesbar: SQLite läuft hier als
    # WebAssembly (`ncruces/go-sqlite3` über wazero), und der Detektor
    # instrumentiert jeden Speicherzugriff des Interpreters.
    #
    # **Was das kostet und was es hält**, steht bei dem Wächter, der die Liste
    # ehrlich hält: `internal/werkzeuge/rennliste_test.go`.
    if HOME="$HEIM_FUER_LAUF" \
        go test -race -timeout 30m \
        -coverprofile=.abdeckung-race.out -covermode=atomic $(rennpakete) \
        2>&1 | sed 's/^/      /'; then
      ok "go test -race (mit leerem Zuhause)"
    else
      fehl "go test -race"
    fi

    if teile_abwarten "$teile" .abdeckung-geteilt.out; then
      ok "go test -race, geteilt: ${GETEILT} in ${TEILE} Teilen ($(($(date +%s) - anfang)) s)"
    else
      fehl "go test -race, geteilt: ${GETEILT}"
    fi
    teile_aufbewahren "$teile" "$WURZEL/.ci-dauer"
    rm -rf "$teile"

    if HOME="$HEIM_FUER_LAUF" \
        go test -timeout 30m \
        -coverprofile=.abdeckung-ruhig.out -covermode=atomic $(ruhigepakete) \
        2>&1 | sed 's/^/      /'; then
      ok "go test (ohne Detektor, wo keine Goroutine läuft)"
    else
      fehl "go test"
    fi

    if profile_vereinen; then
      abdeckung_pruefen
    else
      fehl "die Abdeckungsprofile vereinen"
    fi

    local ortsbefund=0 ortsausgabe
    ortsausgabe="$(echte_orte_pruefen "$orte_vorher")" || ortsbefund=$?
    case "$ortsbefund" in
      0) ok "die echten Orte dieses Rechners blieben unberührt (mit Positivkontrolle)" ;;
      3) weg "die echten Orte: ${ortsausgabe##*$'\n'}" ;;
      *) printf '%s\n' "$ortsausgabe" | sed 's/^/      /'
         fehl "die Go-Läufe veränderten echte Orte dieses Rechners" ;;
    esac
  fi

  if hat_frontend; then
    if (cd frontend && npm run --silent test:coverage 2>&1 | sed 's/^/      /'); then
      ok "vitest mit Abdeckung"
    else
      fehl "vitest"
    fi
  else
    weg "kein Frontend"
  fi

  geraetemodell_pruefen
  skriptpruefungen
}

# **`fm serve` steht in einer eigenen Datei** (§2) — es läuft nur, wo es da
# ist, und warum das Fehlen kein `weg` ist, steht dort.
# shellcheck source=scripts/ci-geraetemodell.sh
. "$(dirname "$0")/ci-geraetemodell.sh"

# skriptpruefungen fährt die *.test.sh unter scripts/.
#
# Shell-Code ist Code (AGENTS.md §3). Die Freigabe entscheidet, ob eine Version
# überhaupt hinausgeht — dass ihre Prüfung stimmt, gehört nicht der Umsicht
# überlassen, sondern in die Pipeline.
skriptpruefungen() {
  local gefunden=0 datei

  for datei in scripts/*.test.sh; do
    [ -f "$datei" ] || continue
    gefunden=1

    if bash "$datei" >/dev/null 2>&1; then
      ok "$(basename "$datei")"
    else
      bash "$datei" 2>&1 | sed 's/^/      /'
      fehl "$(basename "$datei")"
    fi
  done

  [ "$gefunden" -eq 1 ] || weg "keine Skriptprüfungen vorhanden"
}

# -------------------------------------------------------------- Sicherheit

# **Die Stufe `security` steht in einer eigenen Datei** (§2) — seit ENG-04 mit
# den vier Zuständen aus `scripts/sicherheitslage.sh`.
# shellcheck source=scripts/ci-sicherheit.sh
source "$(dirname "$0")/ci-sicherheit.sh"

# stempeln schreibt den Beleg eines grünen Laufs.
#
# **Die Arbeit steht in `scripts/stempel.sh`**, und das ist kein Zufall: sie
# hat einen Fehlerweg, der sich nur mit einem echten Repositorium zeigt — ein
# Stand, der sich **während** des Laufs ändert. Hier drin wäre er erst nach
# zehn Minuten Pipeline zu betreten, und was so lange braucht, prüft niemand
# (#1389,
# Befund 4 des Zweitblicks über die Phasen 4.0 bis 4.9). `hook.test.sh` fährt
# das Skript in einem Wegwerfrepositorium und stellt den Fall her.
#
# **Nur nach `all`.** Eine einzelne Stufe ist kein grüner Lauf, und ein
# Stempel dafür wäre die Behauptung, es sei einer gewesen.
stempeln() {
  [ "${STUFE:-}" = "all" ] || return 0

  ./scripts/stempel.sh "$STAND_ANFANG"
}

# **Die Reichweite steht in einer eigenen Datei** (§2): dieses Skript ist über
# 400 Zeilen gewachsen, und der Schnitt folgt der Frage — hier steht, **was
# geprüft wird**, dort, **was ein grüner Lauf nicht sagt**.
# shellcheck source=scripts/ci-reichweite.sh
. "$(dirname "$0")/ci-reichweite.sh"

# noch_aktuell hält einen überholten Lauf zwischen zwei Stufen an — QA-07.AK-03
# (#1968).
#
# **Hat sich der Stand seit dem Anfang geändert, gibt es keinen Beleg mehr**
# (`stempel.sh`), und jede weitere Stufe kostete Minuten für einen Lauf, den
# niemand mehr braucht. Die Sperre geht mit dem Ende des Prozesses.
noch_aktuell() {
  ./scripts/stempel.sh --aktuell "$STAND_ANFANG" && return 0

  printf '\n%s  ✗ überholt — der Stand hat sich während des Laufs geändert; hier ist Schluss%s\n' \
    "$ROT" "$AUS"
  FEHLER=1
  schlusszeile
  exit 1
}

# ----------------------------------------------------------------- Ablauf

STUFE="${1:-all}"

# **Der Stand vor der ersten Stufe.** Er entscheidet am Ende, ob es einen
# Beleg gibt — siehe `stempeln` oben und `scripts/stempel.sh`.
STAND_ANFANG="$(./scripts/stempel.sh --stand)"

case "$STUFE" in
  format)   stufe_format ;;
  types)    stufe_types ;;
  lint)     stufe_lint ;;
  test)     stufe_test ;;
  security) stufe_security ;;
  all)      stufe_format; noch_aktuell; stufe_types; noch_aktuell; stufe_lint
            noch_aktuell; stufe_test; noch_aktuell; stufe_security ;;
  *)        printf 'Unbekannte Stufe: %s\n' "$1" >&2
            printf 'Erlaubt: format types lint test security all\n' >&2
            exit 2 ;;
esac

schlusszeile
exit "$FEHLER"
