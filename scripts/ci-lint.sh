#!/usr/bin/env bash
#
# Die Stufe `lint` aus scripts/ci.sh (AGENTS.md §4).
#
# Ausgelagert am 2026-08-25 wegen der Zeilengrenze aus §2 — genau wie
# `ci-abdeckung.sh` und `ci-pinnung.sh` davor. Wird von scripts/ci.sh
# eingebunden und nutzt dessen ok/fehl/weg.
#
# ## Warum die Stufe die größte ist
#
# Sie prüft dreierlei, das nur den Namen teilt: die Linter über **jede**
# beschriebene Plattform und jeden Build-Tag, die Zeilengrenze aus §2, und die
# Wächter über Dateien der Projektwurzel. Der Schnitt folgt der Frage: dort
# steht, **welche** Stufen es gibt, hier, **was** eine davon tut.

# --------------------------------------------------------------- Linting

stufe_lint() {
  kopf "Linting"

  if hat_go && werkzeug_da golangci-lint; then
    # Über **jede** beschriebene Plattform, nicht nur die des Rechners:
    # golangci-lint sieht nur die Dateien, die für das aktuelle GOOS übersetzt
    # werden. Ohne diese Schleife entzieht sich jede plattformabhängige Datei
    # der Prüfung, und lokal ist trotzdem alles grün.
    for ziel in darwin windows linux; do
      local pakete="$GO_PAKETE"
      [ "$ziel" = linux ] && pakete="$(pakete_fuer_linux)"
      # shellcheck disable=SC2086
      if GOOS="$ziel" golangci-lint run $pakete 2>&1 | sed "s/^/      [$ziel] /"; then
        ok "golangci-lint ($ziel)"
      else
        fehl "golangci-lint ($ziel)"
      fi
    done

    # **Und über den Build-Tag `fremdpakete`**
    # (#526).
    #
    # Die Schleife darüber deckt ihn **nicht** ab: sie wechselt das Zielsystem,
    # nicht die Tags. Bis zum 2026-08-19 waren damit genau die Läufe ungeprüft,
    # an denen die Parität gemessen wird — neun Befunde lagen darin, drei davon
    # `gosec`. Der Kommentar oben behauptete das Gegenteil („jede Datei hinter
    # einem Build-Tag"); das ist die halb gehaltene Zusage aus §10.
    #
    # **Nur für darwin**: diese Läufe brauchen die kanonischen Sets und ein
    # bezogenes Paket auf **diesem** Rechner (§3). Sie für Windows zu übersetzen
    # sagte nichts, was die Schleife oben nicht schon sagt.
    #
    # **Und `netzlauf` ebenso** (#888). Er befragt Scryfall und läuft deshalb
    # nicht mit; **übersetzt und geprüft wird er trotzdem** — sonst wäre er die
    # eine Datei, in der jeder Befund unbemerkt bliebe, und gerade sie fährt
    # gegen einen Rechner, den wir nicht kontrollieren.
    #
    # **Und `abnahme`** (ENG-02.AK-04): die Griffe in AppKit, mit denen die
    # Hüllenprobe das native Fenster bedient. Kein Lauf in der Pipeline
    # betritt sie — gelintet werden sie trotzdem, sonst verrottete genau der
    # Teil, den nur der Abnahmebau übersetzt.
    #
    # **Und `server`** (QA-09, #1740): die Wache vor dem Servermodus und der
    # Lauf, der ihn an Wails' echtem Lauscher angreift
    # (`internal/fenster/assetweiche_server*.go`). Übersetzt nur der Bau
    # `browser`, und den fährt keine Stufe.
    #
    # **Und `fmlauf`** (AI-05, #2042): der Lauf gegen Apples Gerätemodell und
    # der Schalter, der dafür den Riegel öffnet. Er braucht macOS 27 und
    # läuft deshalb nie in der Pipeline. Seit AI-02 (#1743) trägt ihn auch
    # der Vergleich der Textmodelle in `internal/app`.
    #
    # **Und `messlauf`**: der Retrievallauf gegen eine echte
    # Einbettungslaufzeit. Er stand bis AI-02 in keiner Schleife — die
    # Lücke, gegen die dieser Block steht.
    #
    # **Und `sensorlauf`** (SEC-19, #2153): die echte Bindung von
    # `LocalAuthentication` am Touch-ID-Sensor. Er öffnet den Dialog dieses
    # Rechners und läuft deshalb nie in der Pipeline.
    #
    # **Und `vmlauf`** (SEC-22, #2163): die Läufe zwischen zwei Gästen —
    # mDNS, TLS und der Fernweg —, gestartet über `lume ssh`.
    #
    # **Und `probeanker`** (SEC03, #2356): der Anker, den die gebauten
    # Programme der Flächenläufe aus `MSE_PROBEANKER` lesen. Ein Freigabeweg
    # baut nie damit (`echterschluessel_test.go`).
    for tag in fremdpakete netzlauf abnahme server fmlauf messlauf sensorlauf vmlauf probeanker; do
      if GOOS=darwin golangci-lint run --build-tags "$tag" $GO_PAKETE 2>&1 \
        | sed "s/^/      [$tag] /"; then
        ok "golangci-lint ($tag)"
      else
        fehl "golangci-lint ($tag)"
      fi
    done
  elif hat_go; then
    fehl "golangci-lint nicht gefunden — 'make tools' ausführen"
  else
    weg "kein go.mod"
  fi

  if hat_frontend; then
    if (cd frontend && npm run --silent lint 2>&1 | sed 's/^/      /'); then
      ok "eslint"
    else
      fehl "eslint"
    fi
  else
    weg "kein Frontend"
  fi

  groesse_pruefen
}

# ZEILENGRENZE ist die Schwelle aus AGENTS.md §2 — ohne „etwa".
#
# An einer Zahl, über die niemand abstimmt, wächst eine Datei jedes Mal um genau
# eine Zeile. Diese Prüfung nimmt die Zahl beim Wort. Testdateien sind nicht
# ausgenommen; sie sind der Anlass der Regel.
ZEILENGRENZE=400

groesse_pruefen() {
  # Nur vorhandene Verzeichnisse durchsuchen.
  #
  # `find` über einen fehlenden Pfad endet mit 1, und unter `set -o pipefail`
  # reißt das die Zuweisung und mit ihr das ganze Skript ab — lautlos und
  # ausgerechnet dann, wenn nichts zu beanstanden ist. In Phase 1 fehlen
  # `internal`, `frontend/src` und `saat` planmäßig.
  #
  # **`.swift` kam am 2026-09-24 dazu** (#1828), mit der WebKit-Probe unter
  # `scripts/webkitprobe/` — die erste Quelle in dieser Sprache, und §2 sagt
  # „keine Quelltextdatei", nicht „keine Go-Datei".
  #
  # **Perl, C und Python kamen am 2026-10-11 dazu** (#2213): die Treiber der
  # Gastproben sind Perl, E5 bringt eine C-Bibliothek mit, und
  # `abzuegeschrumpfen.py` lag schon länger da — keine davon zählte der
  # Wächter, bis die Gegenprobe mit 401 Zeilen grün blieb.
  #
  # **`cmd` kam am 2026-08-27 dazu** (#1165). Es fehlte, seit es die Grenze
  # gibt, und niemand sah es: `cmd/` trägt drei Programme mit echtem
  # Quelltext — die Verteilung der Kommandozeile, die Nähte des MCP-Servers,
  # den Aufbau des Dienstes. §2 sagt „keine Quelltextdatei", und ein Wächter,
  # der die halbe Zusage hält, macht die andere unsichtbar (§10).
  local suchpfade=()
  for verzeichnis in internal scripts frontend/src saat cmd; do
    [ -d "$verzeichnis" ] && suchpfade+=("$verzeichnis")
  done

  if [ ${#suchpfade[@]} -eq 0 ]; then
    weg "keine Quellverzeichnisse vorhanden"
    return 0
  fi

  local zu_gross
  zu_gross="$(
    find "${suchpfade[@]}" \
         \( -name '*.go' -o -name '*.ts' -o -name '*.tsx' -o -name '*.sh' \
            -o -name '*.mjs' -o -name '*.js' -o -name '*.css' -o -name '*.sql' \
            -o -name '*.swift' -o -name '*.pl' -o -name '*.pm' -o -name '*.c' \
            -o -name '*.py' \) \
         -type f 2>/dev/null \
      | sort \
      | while read -r datei; do
          # `wc -l` zählt Zeilenenden, nicht Zeilen: eine Datei ohne letzten
          # Umbruch käme um eins zu kurz davon. `awk` zählt Datensätze.
          zeilen="$(awk 'END{print NR}' "$datei")"
          # Als `if`, nicht als `[ … ] && printf`: die Kurzform gibt 1 zurück,
          # wenn die Datei klein genug ist — bei der letzten Datei wäre das
          # unter `set -e` der stille Abbruch, und zwar genau dann, wenn alles
          # in Ordnung ist.
          if [ "$zeilen" -gt "$ZEILENGRENZE" ]; then
            printf '%6d  %s\n' "$zeilen" "$datei"
          fi
        done
  )"

  if [ -z "$zu_gross" ]; then
    ok "keine Datei über $ZEILENGRENZE Zeilen"
    return 0
  fi

  printf '%s\n' "$zu_gross" | sed 's/^/      /'
  fehl "Dateien über $ZEILENGRENZE Zeilen — aufteilen (AGENTS.md §2)"
}
