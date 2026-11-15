#!/usr/bin/env bash
#
# **Der schnelle Riegel vor der Pipeline** — fünf Minuten statt fünfzig.
#
# # Wofür er da ist
#
# `make ci` ist der Nachweis, ohne den nichts gepusht wird (§4). Er ist auch
# teuer: am 2026-09-11 gemessen rund fünfzig Minuten, und **davon** 42 für die
# Teststufe. Ihn nach jeder Änderung zu fahren heißt, in einer Nacht vier
# Stunden zu warten.
#
# **Was hier läuft, hat jeden Fund dieser Nacht gefunden**: eslint, die
# Zeilengrenze, die Abdeckung und die Dokuzahlen — vier rote Pipelines, kein
# einziger Fund, der die volle gebraucht hätte.
#
# # Was er nicht ist
#
# **Kein Ersatz.** Was hier fehlt und dort steht:
#
#   * `go test -race` über **alle** Pakete (das leere Zuhause hat er seit
#     A01 auch),
#   * die Übersetzung für Windows und Linux,
#   * die **Läufe** hinter `fremdpakete` und `netzlauf` — gelintet werden
#     sie hier seit dem 2026-09-13, ausgeführt nicht,
#   * `govulncheck`, `npm audit`, die Stückliste,
#   * die Skriptprüfungen unter `scripts/`.
#
# **Vor dem Push läuft `make ci`**, und der Hook prüft es nach. Dieser Riegel
# verkürzt den Weg dorthin, er ersetzt ihn nicht.

set -euo pipefail

cd "$(dirname "$0")/.."

# shellcheck source=/dev/null
source scripts/activate.sh >/dev/null 2>&1

# shellcheck source=scripts/linuxpakete.sh
source scripts/linuxpakete.sh

ROT=$'\033[31m'; GRUEN=$'\033[32m'; BLAU=$'\033[34m'; AUS=$'\033[0m'
[ -t 1 ] || { ROT=''; GRUEN=''; BLAU=''; AUS=''; }

FEHLER=0

kopf() { printf '\n%s── %s %s\n' "$BLAU" "$1" "$AUS"; }
ok()   { printf '  %s✓%s %s\n' "$GRUEN" "$AUS" "$1"; }
fehl() { printf '  %s✗%s %s\n' "$ROT" "$AUS" "$1"; FEHLER=1; }

kopf "Go"

if go vet ./internal/... ./cmd/... . 2>&1 | sed 's/^/      /'; then
  ok "go vet"
else
  fehl "go vet"
fi

# **Und für die zwei Ziele, für die die Pipeline mit übersetzt** —
# nachgetragen am 2026-09-23. Ein Test, der einen Helfer aus einer Datei mit
# `//go:build darwin` nahm, war hier grün und in `make ci` rot: dort sah ihn
# `golangci-lint (windows)`, hier übersetzte alles nur für diesen Rechner.
# `go vet` prüft die Typen samt Tests und kostet mit warmem Zwischenspeicher
# zwei Sekunden je Ziel.
#
# **Linux ohne die Hülle** (ENG-02): Wails 3 baut dort nur mit cgo (0132).
for ziel in windows linux; do
  pakete="./internal/... ./cmd/... ."
  [ "$ziel" = linux ] && pakete="$(pakete_fuer_linux)"
  # shellcheck disable=SC2086
  if GOOS="$ziel" go vet $pakete 2>&1 | sed 's/^/      /'; then
    ok "go vet ($ziel)"
  else
    fehl "go vet ($ziel)"
  fi
done

if GOOS=darwin golangci-lint run ./internal/... ./cmd/... . 2>&1 | sed 's/^/      /'; then
  ok "golangci-lint"
else
  fehl "golangci-lint"
fi

# **Und über die Build-Tags** — nachgetragen am 2026-09-13.
#
# # Warum das hier fehlte
#
# **Weil es eine Nacht gekostet hat.** Der Lauf darüber liest nur den
# Normalbau; zwei Befunde in `internal/mtgjson/netzlauf_test.go` und
# `internal/mtgjsonpreise/netzlauf_test.go` waren hier grün und haben `make
# ci` rot gemacht — zehn Minuten später und ohne Zusammenhang zu dem, woran
# gerade jemand saß. Genau das soll dieser Riegel verhindern.
#
# **Derselbe Fund gab es schon einmal** (#526):
# bis zum 2026-08-19 fehlten die Tags in `make ci` selbst, und neun Befunde
# lagen darin.
#
# # Warum nur über die Pakete, die den Tag tragen
#
# **Weil die Aussage nicht weiter reicht** (§10) — und weil es sonst teuer
# wäre. Gemessen am 2026-09-13 bei warmem Zwischenspeicher, eine Datei in
# `internal/app` angefasst:
#
#   | Lauf | Zeit |
#   | --- | ---: |
#   | ohne Tag, alle 45 Pakete | 1,25 s |
#   | `fremdpakete` über die 7 Pakete, die ihn tragen | 1,03 s |
#   | `netzlauf` über die 6 Pakete, die ihn tragen | 0,94 s |
#   | `fremdpakete` über alle 45, wie `make ci` | 4,42 s |
#   | `netzlauf` über alle 45 | 6,62 s |
#
# Zwei Sekunden statt elf. **Kalt sind es je rund zehn**, und das bleibt so —
# ein frischer Klon zahlt es einmal.
#
# **Tragen darf das nur, weil jede der 96 Dateien hinter den Tags eine
# Testdatei ist**: ein Tag ändert damit die Übersetzung seines **eigenen**
# Pakets und keines anderen. Legt jemand eine Produktivdatei hinter einen Tag,
# stimmt diese Überlegung nicht mehr — dann gilt wieder der volle Lauf aus
# `make ci`, und der findet es.
#
# **Die Paketliste wird berechnet und nicht eingetragen.** Eine Aufzählung
# hier veraltete beim nächsten Paket, das einen Tag bekommt, und wäre still
# zu kurz — dieselbe Gestalt wie der Fund, den dieser Block verhindert.
#
# **Und `abnahme`** — nachgetragen am 2026-09-26, als erst `make ci` einen
# G703 in der Abnahmeleitung fand. **Seine Dateien sind keine Testdateien**,
# aber sie ändern nur die Übersetzung ihres eigenen Pakets (`internal/fenster`)
# — die Überlegung darüber trägt auch hier.
#
# **Und `server`** (QA-09, #1740) — aus demselben Grund: die Wache vor dem
# Servermodus ändert nur die Übersetzung von `internal/fenster`.
#
# **Und `fmlauf`** (AI-05, #2042) — er ändert `internal/applefm` und seit
# AI-02 (#1743) die Testdateien von `internal/app`. **Und `messlauf`**, der
# Retrievallauf in `internal/flaechen` — bis AI-02 in keiner Schleife. **Und
# `sensorlauf`** (SEC-19, #2153) — die echte Bindung am Touch-ID-Sensor. **Und
# `vmlauf`** (SEC-22, #2163) — die Läufe zwischen zwei Gästen. **Und
# `probeanker`** (SEC03, #2356) — der Anker der gebauten Programme.
for tag in fremdpakete netzlauf abnahme server fmlauf messlauf sensorlauf vmlauf probeanker; do
  # `grep -rl` über die Tagzeile, daraus die Verzeichnisse, daraus Pakete.
  pakete=$(grep -rl "^//go:build $tag" internal cmd 2>/dev/null \
    | xargs -n1 dirname 2>/dev/null | sort -u | sed 's|^|./|;s|$|/...|')

  if [ -z "$pakete" ]; then
    ok "golangci-lint ($tag) — keine Datei trägt ihn"

    continue
  fi

  # Die Paketliste ist absichtlich mehrere Wörter.
  # shellcheck disable=SC2086
  if GOOS=darwin golangci-lint run --build-tags "$tag" $pakete 2>&1 \
    | sed "s/^/      [$tag] /"; then
    ok "golangci-lint ($tag)"
  else
    fehl "golangci-lint ($tag)"
  fi
done

# **Und `gosec`** — seit dem 2026-09-14
# (#1455).
#
# **Weil er drei Sekunden kostet und einundzwanzig Minuten spart.** Er stand
# bis dahin unter „was hier fehlt", und genau dort ist ein Fund
# hindurchgefallen: `//nolint:gosec` ist eine Anweisung an **golangci-lint**;
# `gosec` läuft eigenständig und liest sie nicht. Der Riegel war grün, die
# Stufe `security` rot — zwanzig Minuten später und ohne Zusammenhang zu dem,
# woran gerade jemand saß. Derselbe Fall wie bei den Build-Tags (§4).
#
# **Derselbe Umfang wie in der Pipeline**, und das ist keine Kosmetik: ein
# `./...` vom Wurzelverzeichnis läuft durch `frontend/node_modules` und kommt
# nicht wieder — gemessen am selben Tag, über zehn Minuten ohne Ergebnis.
if command -v gosec > /dev/null 2>&1; then
  if gosec -quiet ./internal/... ./cmd/... . 2>&1 | sed 's/^/      /'; then
    ok "gosec"
  else
    fehl "gosec"
  fi
else
  # **`command -v` und nicht `werkzeug_da`**: jene Hilfe steht in
  # `scripts/ci.sh`, und dieser Riegel lädt sie nicht — er soll allein
  # laufen können.
  fehl "gosec nicht gefunden — 'make tools' ausführen"
fi

kopf "Die Zeilengrenze (§2)"

# **Dieselbe Prüfung wie in der Pipeline — nicht eine zweite Suche daneben.**
# Eine zweite liefe auseinander.
weg() { printf '  – %s\n' "$1"; }
# shellcheck source=scripts/ci-lint.sh
source "$(dirname "$0")/ci-lint.sh"
# shellcheck source=scripts/vorpruefungsbericht.sh
source "$(dirname "$0")/vorpruefungsbericht.sh"
groesse_pruefen

kopf "Die Pakete mit harter Schwelle"

# **Ohne Detektor und ohne leeres Zuhause** — beides kostet, und beides ist in
# `make ci` da. Was hier gesucht wird, sind **unbetretene Anweisungen**, und
# die hängen an keinem von beidem.
# **Über alles und nicht nur über die acht Pakete mit Schwelle.**
#
# Eine Anweisung in `service` kann von einem Lauf in `app` oder in `cmd`
# betreten werden; wer nur die acht führe, meldete sie als unbetreten —
# nachgemessen am 2026-09-11 an `haushalt.go:56`, einer leeren Funktion, die
# ein anderes Paket ruft. **Ein Riegel, der falsch meldet, wird nicht mehr
# gelesen.**
#
# **Und es kostet fast nichts**: ohne Race-Detektor sind es 86 Sekunden für
# `internal/` und 33 für `cmd/` — gegen 42 Minuten mit.
KERNE="./internal/... ./cmd/..."

# **Der Rückgabewert von `go test` und nicht der von `grep`.**
#
# Der erste Anlauf hing ihn an eine Rohrleitung: `grep -v '^ok'` gibt 1 zurück,
# wenn **alles** grün war — und der Riegel meldete rot, gerade weil nichts zu
# melden war.
#
# **Die beiden Dateien tragen die Prozesskennung.**
#
# Vorher hießen sie fest `.vorpruefung.out` und `.vorpruefung.log` im
# Wurzelverzeichnis. Wer zwei Riegel nebeneinander laufen ließ, bekam vom
# ersten ein `rm -f` auf die Dateien des zweiten — und der zweite meldete
# **grün**: `grep` fand nichts, weil es nichts zu lesen gab, und `awk` nannte
# keine unbetretene Anweisung, weil es kein Profil las. Gemessen am
# 2026-09-13 an genau diesem Riegel: zwei Läufe, zwei Häkchen, **null**
# Messungen.
PROFIL="$(mktemp -t vorpruefung-profil)"
PROTOKOLL="$(mktemp -t vorpruefung-protokoll)"

# **Mit leerem Zuhause, wie in `scripts/ci.sh`** — A01 (#2215). Bis dahin
# fuhr der Riegel die Go-Läufe im echten: der Audit zu 0.8.0 sah dabei zwei
# Drosseldateien und einen Scryfall-Eintrag im echten Anwendungsverzeichnis
# wachsen. **Und mit Abdruck davor und danach**: dass das Zuhause gewechselt
# ist, ist eine Absicht; dass die echten Orte unberührt blieben, ein Befund.
# shellcheck source=scripts/echteorte.sh
source scripts/echteorte.sh
HEIM_FUER_LAUF="$(heim_fuer_laeufe)"
ORTE_VORHER="$(mktemp -t vorpruefung-orte)"
trap 'rm -rf "$PROFIL" "$PROTOKOLL" "$HEIM_FUER_LAUF" "$ORTE_VORHER"' EXIT
echte_orte_merken "$ORTE_VORHER"

if HOME="$HEIM_FUER_LAUF" go test -timeout 20m -coverprofile="$PROFIL" -covermode=count $KERNE \
    > "$PROTOKOLL" 2>&1; then
  grep -vE '^ok' "$PROTOKOLL" | sed 's/^/      /' || true
  ok "go test über die Kernpakete"
  GRUEN_GELAUFEN=ja
else
  fehlschlaege_nennen "$PROTOKOLL"
  fehl "go test über die Kernpakete"
  GRUEN_GELAUFEN=nein
fi

# **Gezählt wird in jedem Fall** — der Grund steht in
# `scripts/vorpruefungsbericht.sh`.
abdeckung_pruefen "$PROFIL" "$GRUEN_GELAUFEN"

ORTSBEFUND=0
ORTSAUSGABE="$(echte_orte_pruefen "$ORTE_VORHER")" || ORTSBEFUND=$?
case "$ORTSBEFUND" in
  0) ok "die echten Orte dieses Rechners blieben unberührt (mit Positivkontrolle)" ;;
  3) printf '  – %s\n' "die echten Orte: ${ORTSAUSGABE##*$'\n'}" ;;
  *) printf '%s\n' "$ORTSAUSGABE" | sed 's/^/      /'
     fehl "die Go-Läufe veränderten echte Orte dieses Rechners" ;;
esac

kopf "Frontend"

# **`npx` installiert nichts** (Phase 5 von 0.9.0, #2326). Ohne
# `node_modules` holte `npx eslint` ESLint 10 aus der Registry und `npx tsc`
# das fremde Paket `tsc` — am 2026-09-23, 2026-10-15 und 2026-10-16, jedes
# Mal in einem frischen Arbeitsbaum. Jetzt wird gesagt, was fehlt.
if [ ! -d frontend/node_modules ]; then
  fehl "frontend/node_modules fehlt — 'npm --prefix frontend ci --prefer-offline --offline' hilft"
fi

if (cd frontend && npx --no-install tsc --noEmit -p tsconfig.json 2>&1 | sed 's/^/      /'); then
  ok "tsc --noEmit"
else
  fehl "tsc"
fi

if (cd frontend && npx --no-install eslint src 2>&1 | sed 's/^/      /'); then
  ok "eslint"
else
  fehl "eslint"
fi

if (cd frontend && npx --no-install prettier --check src >/dev/null 2>&1); then
  ok "prettier"
else
  # **Der Hinweis steht in einfachen Anführungszeichen.**
  #
  # Rückwärtsstriche wären in der Shell **Befehlsersetzung**: die Meldung
  # führte `npx prettier --write src` aus, und zwar im Wurzelverzeichnis, wo es
  # kein `src` gibt.
  fehl "prettier — 'npx --prefix frontend prettier --write src' hilft"
fi

# **Die ganze Ausgabe nach `.vitest-protokoll.log`, und rot nennt den Lauf.**
#
# # Was hier vorher schiefging
#
# **`| tail -5` schnitt genau die Zeile weg, für die es die Stufe gibt.**
# Gemessen am 2026-09-15: ein Lauf von 3.533 war rot, die Stufe sagte
# „1 failed | 3532 passed" — und **welcher**, stand nirgends. Drei
# Wiederholungen von Hand waren grün, weil der Fehler unter Last auftrat.
#
# **Derselbe Fehler steht in AGENTS.md §4 schon aufgeschrieben**, für
# `scripts/ci.sh`: „Zwanzig Minuten für die Zeile `✗ Pipeline rot`, und die
# einzige Auskunft war ein zweiter voller Lauf." Er stand dort über den
# **Aufrufer** — und hier im Skript selbst.
#
# **Gezeigt werden im grünen Fall die letzten fünf Zeilen** (die Bilanz) und
# im roten die Zeilen, die den Lauf benennen: `×`, `FAIL`, `AssertionError`.
#
# **Unter `mktemp` und nicht unter einem festen Namen** — dieselbe Lehre wie
# bei den zwei Protokollen weiter oben: am 2026-09-13 räumten zwei
# gleichzeitige Läufe einander die Datei weg und meldeten **beide grün**.
VITESTPROTOKOLL="$(mktemp -t vorpruefung-vitest)"

# **Mit Abdeckung, und das ist seit dem 2026-09-16 so.**
#
# # Was der Riegel vorher durchließ
#
# **Eine neue `.tsx`-Datei mit einem unbetretenen Zweig.** Der Riegel fuhr
# `vitest run` ohne `--coverage`; die Schwelle von 100 % stand nur in
# `make ci`. Zweimal an einem Vormittag fiel dort ein Zweig um, den dieser
# Lauf nicht gemessen hatte — einmal `Partietisch.tsx:87`, einmal Zeile 156,
# beide aus Werkstück G3 (Issue 1460). Jedes Mal einundzwanzig Minuten für
# eine Auskunft, die hier in zwei zu haben ist.
#
# **Gemessen am 2026-09-16**: `vitest run` 78 s, `vitest run --coverage`
# 122 s. Vierundvierzig Sekunden gegen einundzwanzig Minuten.
#
# **Die Go-Seite bleibt getrennt** — dort misst die Stufe darüber schon die
# acht Kernpakete samt unbetretener Anweisungen. Hier fehlte die Hälfte.
if (cd frontend && npx --no-install vitest run --coverage >"$VITESTPROTOKOLL" 2>&1); then
  tail -8 "$VITESTPROTOKOLL" | sed 's/^/      /'
  ok "vitest mit Abdeckung"
else
  grep -E '^ *×|FAIL |AssertionError|Tests  |^ERROR|Uncovered|% Stmts|^ [a-zA-Z].*\|' \
    "$VITESTPROTOKOLL" | head -24 | sed 's/^/      /'
  fehl "vitest — die ganze Ausgabe steht in $VITESTPROTOKOLL"
fi

printf '\n'

if [ "$FEHLER" -eq 0 ]; then
  printf '%s✓ Vorprüfung grün%s — jetzt lohnt `make ci`\n' "$GRUEN" "$AUS"
else
  printf '%s✗ Vorprüfung rot%s — das findet `make ci` auch, nur langsamer\n' "$ROT" "$AUS"
fi

exit "$FEHLER"
