#!/usr/bin/env bash
#
# Der **Quelltext einer Freigabe** — D01
# (#2331,
# Vermerk [0190](../docs/entscheidungen/0190-das-repositorium-wird-oeffentlich.md)).
#
# # Wofür
#
# GPL-2.0 §3 verlangt zu jedem weitergegebenen Programm den **vollständigen**
# Quelltext — „all the source code for all modules it contains". Ein
# Go-Programm enthält die Module, die es statisch einbindet, und die
# Oberfläche die npm-Pakete in ihrem Bündel; das Archiv, das GitHub zu einem
# Tag bildet, enthält nur den eigenen Baum. Deshalb legt dieser Schritt ein
# eigenes Paket neben das Abbild — **an derselben Stelle**, was §3 im letzten
# Absatz als Weitergabe des Quelltexts gelten lässt.
#
# # Was darin liegt
#
# - der Baum am Commit (`git archive`, nicht der Arbeitsbaum),
# - `vendor/` — die Go-Module des Programms, aus dem Modulzwischenspeicher
#   und **ohne Netz** (`GOPROXY=off`),
# - `frontend/node_modules/` — **nur** die Pakete, die ins Bündel gehen
#   (`npm ls --omit=dev`), nicht die Werkzeuge des Baus,
# - `quellen/` — die Bauquellen dessen, was erzeugt eingebettet ist
#   (`ERZEUGTE_MODULE`),
# - `QUELLEN.md` — je Abhängigkeit Fassung, Lizenz und Ursprung, und die Tafel
#   der erzeugten Formen samt dem, was an ihnen **noch offen** ist (#2393),
# - `QUELLTEXT.md`, das sagt, was es ist und wie man baut.
#
# # Was nicht darin liegt
#
# **Die mitgelieferten Vorlagen** (`scripts/mitliefern.sh`): sie sind Daten
# und kein Modul des Programms — „mere aggregation" nach §2 — und stehen unter
# den Bedingungen ihrer Rechteinhaber (`DRITTANBIETER.md`).
#
# Es reist in `pruefsummen.txt` mit und ist damit signiert; deshalb läuft der
# Schritt vor dem `shasum`-Aufruf, wie `releasemanifest.sh`.

# produktionspakete nennt die npm-Pakete, die ins Bündel gehen — als Pfade
# relativ zum Verzeichnis der Oberfläche.
#
# **Der Vergleich geht über den aufgelösten Pfad** (`pwd -P`): `npm ls` nennt
# ihn so, und unter `/var` — einem Verweis auf `/private/var` — fand der
# erste Lauf der Prüfung deshalb kein einziges Paket.
produktionspakete() { # produktionspakete <frontend>
  local OBERFLAECHE
  OBERFLAECHE="$(cd "$1" && pwd -P)" || return 1

  (cd "$OBERFLAECHE" && npm ls --omit=dev --all --parseable 2>/dev/null) |
    while IFS= read -r PFAD; do
      case "$PFAD" in
        "$OBERFLAECHE"/node_modules/*) printf '%s\n' "${PFAD#"$OBERFLAECHE"/}" ;;
      esac
    done | sort -u
}

# ERZEUGTE_MODULE sind Go-Module, deren eingebundene Form **erzeugt** ist,
# je mit den Teilen, aus denen sie entsteht — `modul` für das ganze Modul,
# `modul:teil:teil` für Unterverzeichnisse. `go mod vendor` legt nur
# Go-Pakete bei, und keines davon trägt diese Quellen.
#
# - **SQLite** (Befund 1 des Zweitblicks über Phase 0): `go-sqlite3-wasm`
#   bringt es als übersetztes WebAssembly, die Form zum Bearbeiten sind die
#   C-Quellen samt Bauskripten — das ganze Modul.
# - **Die Laufzeit von Wails** (Befund 1 des Nachlaufs): das Fenster bettet
#   `runtime.js` ein, gebaut aus `internal/runtime` (TypeScript und
#   `Taskfile.yaml`).
#
# **Tailscales Weboberfläche stand hier bis #2393** — alle vier Programme
# trugen die vorgebauten Dateien aus `web-client-prebuilt` samt der Schrift
# Inter, und welcher Stand sie gebaut hatte, nannte das Modul nicht. Seither
# lässt der Bauschalter `ts_omit_webclient` sie weg (`releaseschalter.sh`),
# ebenso `ts_omit_debugeventbus` die Diagnoseseite des Ereignisbusses mit
# htmx. `go mod vendor` legt beide Pakete trotzdem bei — es fragt nach jedem
# Schalter —, aber kein Programm bindet sie ein.
ERZEUGTE_MODULE=(
  github.com/ncruces/go-sqlite3-wasm/v5
  github.com/wailsapp/wails/v3:internal/runtime
)

# EINBETTUNGEN ordnet jedes fremde Paket, das die ausgelieferten Programme
# mit Dateien einbinden (`//go:embed`), einer Art zu — `paket|erzeugt` oder
# `paket|quelle|grund`. **Erzeugt** heißt: die Form zum Bearbeiten ist eine
# andere, und ihr Modul steht in `ERZEUGTE_MODULE`. **Quelle** heißt: die
# Datei selbst ist die Form zum Bearbeiten. Gelesen am 2026-10-16, jede
# Datei geöffnet; `einbettungen.test.sh` hält die Liste vollständig, in
# beiden Richtungen.
EINBETTUNGEN=(
  'github.com/ncruces/go-sqlite3-wasm/v5|erzeugt'
  'github.com/ncruces/go-sqlite3-wasm/v5/fts5|erzeugt'
  'github.com/wailsapp/wails/v3/internal/assetserver/bundledassets|erzeugt'
  'github.com/wailsapp/wails/v3/pkg/application|quelle|`inline_event_shim.js` ist handgeschrieben (67 Zeilen, kommentiert), `assets/alpha/index.html` eine Seite mit einem PNG als base64'
  'github.com/wailsapp/wails/v3/pkg/updater|quelle|`assets/window.html` ist handgeschrieben (706 Zeilen)'
  'github.com/tailscale/tailcat|quelle|`README.md`'
  'tailscale.com|quelle|`VERSION.txt`, `ALPINE.txt` und zwei Zeilen mit Toolchain-Ständen'
  'tailscale.com/net/dnsfallback|quelle|`dns-fallback-servers.json`, Daten'
)

# bauquellen_beilegen kopiert je erzeugtem Modul sein Verzeichnis — oder die
# genannten Teile — aus dem Modulzwischenspeicher nach `quellen/`, an der
# Fassung, die `vendor/modules.txt` nennt, ohne Netz.
bauquellen_beilegen() { # bauquellen_beilegen <baum> <modulspeicher> <modul[:teil…]…>
  local BAUM="$1" SPEICHER="$2" EINTRAG MODUL TEILE TEIL FASSUNG ZIEL
  shift 2
  for EINTRAG in "$@"; do
    MODUL="${EINTRAG%%:*}"
    TEILE=""
    [ "$MODUL" = "$EINTRAG" ] || TEILE="${EINTRAG#*:}"
    FASSUNG="$(awk -v m="$MODUL" '$1 == "#" && $2 == m { print $3; exit }' "$BAUM/vendor/modules.txt")"
    [ -n "$FASSUNG" ] && [ -d "$SPEICHER/$MODUL@$FASSUNG" ] ||
      ende "die Bauquellen von $MODUL liegen nicht im Modulspeicher"
    ZIEL="$BAUM/quellen/$MODUL@$FASSUNG"
    if [ -z "$TEILE" ]; then
      mkdir -p "$(dirname "$ZIEL")"
      cp -R "$SPEICHER/$MODUL@$FASSUNG" "$(dirname "$ZIEL")/" ||
        ende "die Bauquellen von $MODUL ließen sich nicht beilegen"
    else
      for TEIL in ${TEILE//:/ }; do
        [ -d "$SPEICHER/$MODUL@$FASSUNG/$TEIL" ] ||
          ende "$MODUL hat an $FASSUNG kein Verzeichnis $TEIL"
        mkdir -p "$ZIEL/$(dirname "$TEIL")"
        cp -R "$SPEICHER/$MODUL@$FASSUNG/$TEIL" "$ZIEL/$(dirname "$TEIL")/" ||
          ende "die Bauquellen $TEIL von $MODUL ließen sich nicht beilegen"
      done
    fi
    chmod -R u+w "$BAUM/quellen"
  done
}

# summe ist die SHA-256-Summe einer Datei.
summe() { shasum -a 256 "$1" | cut -d' ' -f1; }

# bearbeitungsformen_beilegen legt die Ursprungsarchive aus
# `scripts/bearbeitungsformen.tsv` unter `quellen/bearbeitungsformen/` bei —
# #2393, Weg 1. **Geladen wird in einen Speicher** (`build/quellarchive/`),
# damit zwei Bauläufe nacheinander (`baugleich.sh`) nicht zweimal 77 MB
# holen; **beigelegt wird nur, was die Summe aus der Liste trägt** — ein
# Archiv, das sich auf dem Server geändert hat, hält die Freigabe an, und ein
# verdorbenes im Speicher wird neu geladen. Ein Ziel, das mehrere Pakete
# teilen (ein Monorepo), liegt einmal bei.
#
# **Keine leeren Felder in der Liste**: `read` mit Tabulator in `IFS` fasst
# zwei aufeinanderfolgende zusammen; der Erzeuger schreibt `-` statt leer.
bearbeitungsformen_beilegen() { # bearbeitungsformen_beilegen <baum> <liste> <speicher>
  local BAUM="$1" LISTE="$2" SPEICHER="$3" ART NAME FASSUNG HERKUNFT ADRESSE SUMME BYTES ZIEL VERMERK
  [ -s "$LISTE" ] || ende "die Liste der Bearbeitungsformen fehlt ($LISTE)"

  while IFS=$'\t' read -r ART NAME FASSUNG HERKUNFT ADRESSE SUMME BYTES ZIEL VERMERK; do
    case "$ART" in '' | '#'*) continue ;; esac
    [ "$ADRESSE" != - ] || continue
    [ ! -e "$BAUM/quellen/bearbeitungsformen/$ZIEL" ] || continue

    if [ ! -f "$SPEICHER/$ZIEL" ] || [ "$(summe "$SPEICHER/$ZIEL")" != "$SUMME" ]; then
      mkdir -p "$(dirname "$SPEICHER/$ZIEL")"
      curl -fsSL --retry 3 --max-time 900 -o "$SPEICHER/$ZIEL.teil" "$ADRESSE" ||
        ende "$ADRESSE ließ sich nicht laden ($NAME $FASSUNG)"
      if [ "$(summe "$SPEICHER/$ZIEL.teil")" != "$SUMME" ]; then
        rm -f "$SPEICHER/$ZIEL.teil"
        ende "$ZIEL trägt nicht die Summe aus der Liste — auf dem Server geändert? (#2393)"
      fi
      mv "$SPEICHER/$ZIEL.teil" "$SPEICHER/$ZIEL"
    fi

    mkdir -p "$(dirname "$BAUM/quellen/bearbeitungsformen/$ZIEL")"
    cp "$SPEICHER/$ZIEL" "$BAUM/quellen/bearbeitungsformen/$ZIEL" ||
      ende "$ZIEL ließ sich nicht beilegen"
  done < "$LISTE"
}

# lizenzdatei nennt die Lizenzdatei, die ein Go-Modul nach `vendor/`
# mitbringt — relativ zum Baum, oder »keine in vendor/«. Am 2026-10-13
# brachten alle 69 Module dieses Programms eine mit.
lizenzdatei() { # lizenzdatei <baum> <modul>
  local DATEI
  DATEI="$(find "$1/vendor/$2" -maxdepth 1 -type f \( -iname 'LICEN[CS]E*' -o -iname 'COPYING*' \) 2>/dev/null | sort | head -1)"
  if [ -n "$DATEI" ]; then printf '%s' "${DATEI#"$1"/}"; else printf 'keine in vendor/'; fi
}

# quellenverzeichnis schreibt `QUELLEN.md`: jedes Go-Modul und jedes Paket
# fürs Bündel mit Fassung, Lizenz und Ursprung.
# shellcheck disable=SC2016 # die Backticks sind Markdown, nicht die Shell
quellenverzeichnis() { # quellenverzeichnis <baum> <oberfläche> [liste]
  local BAUM="$1" OBERFLAECHE="$2" LISTE="${3:-}" PAKET MODUL FASSUNG
  {
    printf '# Ursprungsquellen der Abhängigkeiten\n\n'
    printf 'Die Abhängigkeiten liegen in der Form bei, in der ihre Pakete erscheinen.\n'
    printf '**Wo diese Form gebaut oder erzeugt ist**, ist die Form zum Bearbeiten\n'
    printf 'der Ursprung an derselben Fassung; erzeugte Go-Module bringen ihre\n'
    printf 'Bauquellen unter `quellen/` mit.\n\n'
    printf '## Go-Module (`vendor/`)\n\n| Modul | Fassung | Lizenzdatei | Ursprung |\n| --- | --- | --- | --- |\n'
    while read -r MODUL FASSUNG; do
      printf '| %s | %s | %s | https://%s |\n' "$MODUL" "$FASSUNG" "$(lizenzdatei "$BAUM" "$MODUL")" "$MODUL"
    done < <(awk '$1 == "#" && $2 !~ /^=>/ { print $2, $3 }' "$BAUM/vendor/modules.txt" 2>/dev/null)
    printf '\n## npm-Pakete (`frontend/node_modules/`)\n\n| Paket | Fassung | Lizenz | Ursprung |\n| --- | --- | --- | --- |\n'
    while IFS= read -r PAKET; do
      node -e '
        const p = require(process.argv[1])
        const r = typeof p.repository === "string" ? p.repository : (p.repository?.url ?? "")
        const u = r.replace(/^git\+/, "").replace(/\.git$/, "").replace(/^github:/, "https://github.com/")
        console.log(`| ${p.name} | ${p.version} | ${p.license ?? "—"} | ${u || "—"} |`)
      ' "$OBERFLAECHE/$PAKET/package.json"
    done < <(produktionspakete "$OBERFLAECHE")
    cat <<'ERZEUGT'

## Erzeugte Formen und ihre Bauquellen

| Form | Steckt in | Bauquelle unter `quellen/` | Offen |
| --- | --- | --- | --- |
| SQLite als WebAssembly (`github.com/ncruces/go-sqlite3-wasm/v5`) | allen vier Programmen | das ganze Modul mit `build/build.sh`, und die SQLite-Quelle (gemeinfrei) unter `quellen/bearbeitungsformen/sqlite/` — genau die Dateien, die `build/download.sh` lädt, mit eigener SHA-256-Summe (#2393) | — |
| `runtime.js` von Wails (`internal/assetserver/bundledassets`) | dem Fenster | `github.com/wailsapp/wails/v3@…/internal/runtime` — TypeScript und `Taskfile.yaml`, an derselben Fassung | — |
| Tailscales vorgebaute Weboberfläche (`github.com/tailscale/web-client-prebuilt`) samt der Schrift Inter, und die Diagnoseseite des Ereignisbusses mit htmx (`tailscale.com/util/eventbus`) | **keinem** — seit #2393 lassen die Bauschalter `ts_omit_webclient` und `ts_omit_debugeventbus` sie weg; `vendor/` trägt die Pakete trotzdem, weil `go mod vendor` nach jedem Schalter fragt | — | — |
| PixiJS-Transcoder `basis_transcoder.wasm` und `libktx.wasm` | **keinem** — sie liegen im npm-Paket `pixi.js` bei und gehen nicht ins Bündel der Oberfläche | ihre Quellen sind Basis Universal und KTX-Software; sie reisen nur als Teil des Pakets mit | — |
ERZEUGT
    if [ -n "$LISTE" ]; then
      printf '\n## Bearbeitungsformen (`quellen/bearbeitungsformen/`)\n\n'
      printf 'Je Paket im Bündel das Archiv seines Repositoriums an dem Commit, aus dem es\n'
      printf 'veröffentlicht ist, und die SQLite-Quelle — erhoben mit\n'
      printf '`scripts/bearbeitungsformen.mjs`, beim Bau gegen ihre Summe geprüft (#2393).\n\n'
      printf '| Paket oder Datei | Fassung | Herkunft | Ziel | Vermerk |\n| --- | --- | --- | --- | --- |\n'
      awk -F'\t' '!/^#/ && NF >= 8 {
        for (i = 1; i <= 9; i++) if ($i == "-" || $i == "") $i = "—"
        printf "| %s | %s | %s | %s | %s |\n", $2, $3, $4, $8, $9
      }' "$LISTE"
    fi
  } > "$BAUM/QUELLEN.md"
}

# quelltext_schreiben legt `MagicSetEditor4Mac-<fassung>-quelltext.tar.gz`
# neben die übrigen Artefakte.
quelltext_schreiben() { # quelltext_schreiben <ausgabe> <fassung> <wurzel>
  local AUSGABE="$1" FASSUNG="$2" WURZEL="$3"
  local NAME="MagicSetEditor4Mac-$FASSUNG-quelltext"
  local RAUM BAUM PAKET

  kopf "Quelltext"

  RAUM="$(mktemp -d)"
  BAUM="$RAUM/$NAME"

  git -C "$WURZEL" archive --format=tar --prefix="$NAME/" HEAD |
    tar -x -C "$RAUM" ||
    ende "der Baum am Commit ließ sich nicht auspacken"

  (cd "$BAUM" && GOFLAGS=-mod=mod GOPROXY=off go mod vendor) ||
    ende "die Go-Module ließen sich nicht ohne Netz beilegen"

  while IFS= read -r PAKET; do
    mkdir -p "$BAUM/frontend/$(dirname "$PAKET")"
    cp -R "$WURZEL/frontend/$PAKET" "$BAUM/frontend/$(dirname "$PAKET")/" ||
      ende "das npm-Paket $PAKET ließ sich nicht beilegen"
  done < <(produktionspakete "$WURZEL/frontend")

  # **Nur, wo das Modul eingebunden ist** — ein Baum ohne `vendor/modules.txt`
  # (ein Programm ohne fremde Module) braucht keine Bauquellen.
  if [ -s "$BAUM/vendor/modules.txt" ]; then
    bauquellen_beilegen "$BAUM" "$(cd "$BAUM" && go env GOMODCACHE)" "${ERZEUGTE_MODULE[@]}"
  fi
  bearbeitungsformen_beilegen "$BAUM" "$WURZEL/scripts/bearbeitungsformen.tsv" "$WURZEL/build/quellarchive"
  quellenverzeichnis "$BAUM" "$WURZEL/frontend" "$WURZEL/scripts/bearbeitungsformen.tsv"

  cat > "$BAUM/QUELLTEXT.md" <<QUELLTEXT
# MagicSetEditor4Mac $FASSUNG — Quelltext

Der Quelltext zur Freigabe \`v$FASSUNG\`, nach GPL-2.0-or-later
(\`LICENSE\`), Commit \`$(git -C "$WURZEL" rev-parse HEAD)\`.

- Der eigene Baum am Commit, wie er im Repositorium steht — vollständig.
- \`vendor/\`: die Go-Module, die das Programm einbindet, je mit der
  Lizenzdatei, die das Modul mitbringt.
- \`frontend/node_modules/\`: die npm-Pakete, die ins Bündel der Oberfläche
  gehen, wie sie veröffentlicht sind. Die Werkzeuge des Baus (Vite,
  TypeScript, …) nennt \`frontend/package-lock.json\`.
- \`QUELLEN.md\`: je Go-Modul Fassung, Lizenzdatei und Ursprung, je
  npm-Paket Fassung, Lizenzangabe und Ursprung, und **die erzeugten Formen
  mit ihren Bauquellen**. Die Abhängigkeiten liegen in der Form bei, in der
  ihre Pakete erscheinen; wo diese Form gebaut ist, ist der Ursprung an
  derselben Fassung die Form zum Bearbeiten, und was erzeugt eingebettet
  ist, bringt seine Bauquellen unter \`quellen/\` mit.
- \`quellen/bearbeitungsformen/\`: **der Ursprung jedes npm-Pakets im Bündel**
  — das Repositorium an dem Commit, aus dem es veröffentlicht ist — und die
  SQLite-Quelle. \`QUELLEN.md\` nennt je Paket Herkunft und Ziel, und wo ein
  Paket nicht an seinem eigenen Commit liegt, warum.

Bauen: \`README.md\` (Einrichtung) und \`docs/RELEASE.md\` (Freigabebau).
Fremde Inhalte und ihre Bedingungen: \`DRITTANBIETER.md\` und
\`docs/ABHAENGIGKEITEN.md\`. Die Vorlagen im Bündel gehören nicht zum
Programm und stehen nicht unter der GPL.
QUELLTEXT

  tar -czf "$AUSGABE/$NAME.tar.gz" -C "$RAUM" "$NAME" ||
    ende "das Quelltextpaket ließ sich nicht packen"
  rm -rf "$RAUM"

  ok "$NAME.tar.gz"
}
