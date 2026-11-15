#!/usr/bin/env bash
#
# Prüfungen des Quelltextpakets aus scripts/releasequelltext.sh — D01
# (#2331).
#
# Gegen einen **nachgestellten** Baum: ein Repositorium mit einem Go-Modul
# ohne fremde Abhängigkeit und einer Oberfläche mit einem Paket fürs Bündel
# und einem Werkzeug des Baus. Geprüft wird, dass das Paket den Baum **am
# Commit** trägt (nicht den Arbeitsbaum), das Paket fürs Bündel beilegt, das
# Werkzeug weglässt und sagt, was es ist. **Was er nicht prüft**: `vendor/`
# mit echten Modulen — dafür bräuchte er den Modulzwischenspeicher und ein
# Netz; das tut der Freigabebau selbst, und er bricht ab, wenn es scheitert.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0
RAUM="$(mktemp -d)"
trap 'rm -rf "$RAUM"' EXIT

kopf() { :; }
ok() { :; }
ende() { printf '%s  ✗ abgebrochen: %s%s\n' "$ROT" "$1" "$AUS"; exit 1; }

# shellcheck source=scripts/releasequelltext.sh
source "$WURZEL/scripts/releasequelltext.sh"

pruefe() { # pruefe <beschreibung> <bedingung…>
  local BESCHREIBUNG="$1"
  shift
  if "$@"; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$BESCHREIBUNG" "$AUS"
  else
    printf '%s  ✗ %s%s\n' "$ROT" "$BESCHREIBUNG" "$AUS"
    FEHLER=1
  fi
}

BAUM="$RAUM/baum"
mkdir -p "$BAUM/frontend/node_modules/buendelpaket" "$BAUM/frontend/node_modules/bauwerkzeug"
cd "$BAUM" || exit 1
git init -q && git config user.email probe@beispiel.invalid && git config user.name Probe && git config commit.gpgsign false
printf 'module beispiel.invalid/probe\n\ngo 1.27\n' > go.mod
printf 'package main\n\nfunc main() {}\n' > main.go
printf 'GPL\n' > LICENSE
printf '{"name":"probe","version":"1.0.0","dependencies":{"buendelpaket":"1.0.0"},"devDependencies":{"bauwerkzeug":"1.0.0"}}\n' > frontend/package.json
printf '{"name":"buendelpaket","version":"1.0.0","license":"MIT","repository":{"type":"git","url":"git+https://beispiel.invalid/buendelpaket.git"}}\n' > frontend/node_modules/buendelpaket/package.json
printf '{"name":"bauwerkzeug","version":"1.0.0"}\n' > frontend/node_modules/bauwerkzeug/package.json
printf 'node_modules/\n' > frontend/.gitignore
# **Eine Bearbeitungsform über `file://`** (#2393) — der Freigabebau lädt sie
# und prüft die Summe; hier ohne Netz.
printf 'ursprung\n' > "$RAUM/ursprung.tar.gz"
SUMME="$(shasum -a 256 "$RAUM/ursprung.tar.gz" | cut -d' ' -f1)"
mkdir -p scripts
printf '# Probe\nnpm\tbuendelpaket\t1.0.0\t%040d\tfile://%s\t%s\t9\tnpm/buendelpaket.tar.gz\t\nnpm-typen\tnurtypen\t1.0.0\t-\t-\t-\t-\t-\tnur Typen\n' 0 "$RAUM/ursprung.tar.gz" "$SUMME" > scripts/bearbeitungsformen.tsv
git add -A && git commit -q -m probe
printf 'ungesichert\n' > nur-im-arbeitsbaum.txt

AUSGABE="$RAUM/ausgabe"
mkdir -p "$AUSGABE"
quelltext_schreiben "$AUSGABE" 9.9.9 "$BAUM" >/dev/null

PAKET="$AUSGABE/MagicSetEditor4Mac-9.9.9-quelltext.tar.gz"
pruefe "das Paket liegt neben den Artefakten" test -s "$PAKET"
INHALT="$(tar -tzf "$PAKET" 2>/dev/null)"
enthaelt() { grep -qx "MagicSetEditor4Mac-9.9.9-quelltext/$1" <<<"$INHALT"; }
fehlt() { ! grep -q "$1" <<<"$INHALT"; }
fehlt_in() { ! grep -q "$2" <<<"$1"; }

pruefe "der Baum am Commit: main.go" enthaelt main.go
pruefe "der Baum am Commit: LICENSE" enthaelt LICENSE
pruefe "das Paket fürs Bündel liegt bei" enthaelt "frontend/node_modules/buendelpaket/package.json"
pruefe "das Werkzeug des Baus liegt nicht bei" fehlt "bauwerkzeug"
pruefe "der Arbeitsbaum geht nicht mit" fehlt "nur-im-arbeitsbaum"
pruefe "QUELLTEXT.md sagt, was es ist" enthaelt QUELLTEXT.md
pruefe "QUELLTEXT.md nennt den Commit" grep -q "$(git -C "$BAUM" rev-parse HEAD)" <(tar -xzOf "$PAKET" MagicSetEditor4Mac-9.9.9-quelltext/QUELLTEXT.md)

# **Die Ursprungsquellen** (Befund 1 des Zweitblicks über Phase 0): jedes
# Paket fürs Bündel mit Fassung, Lizenz und Ursprung.
QUELLEN="$(tar -xzOf "$PAKET" MagicSetEditor4Mac-9.9.9-quelltext/QUELLEN.md 2>/dev/null)"
pruefe "QUELLEN.md liegt bei" enthaelt QUELLEN.md
pruefe "QUELLEN.md nennt das Paket fürs Bündel mit Fassung, Lizenz und Ursprung" grep -q "buendelpaket | 1.0.0 | MIT | https://beispiel.invalid/buendelpaket" <<<"$QUELLEN"
pruefe "QUELLEN.md nennt das Werkzeug des Baus nicht" fehlt_in "$QUELLEN" bauwerkzeug
pruefe "die Bearbeitungsform liegt bei" enthaelt "quellen/bearbeitungsformen/npm/buendelpaket.tar.gz"
pruefe "QUELLEN.md nennt die Bearbeitungsform mit Herkunft und Ziel" grep -q "| buendelpaket | 1.0.0 | 0000000000000000000000000000000000000000 | npm/buendelpaket.tar.gz |" <<<"$QUELLEN"
pruefe "QUELLEN.md nennt ein Paket nur mit Typen samt Grund" grep -q "| nurtypen | 1.0.0 | — | — | nur Typen |" <<<"$QUELLEN"
pruefe "die Tafel der erzeugten Formen hat keinen offenen Punkt" fehlt_in "$(sed -n '/^## Erzeugte Formen/,/^## /p' <<<"$QUELLEN" | awk -F'|' 'NF > 5 && NR > 4 { print $5 }' | grep -v '^ *— *$')" "."

# **Die Bauquellen erzeugter Module** — aus dem Modulzwischenspeicher, ohne
# Netz, an der Fassung, die `vendor/modules.txt` nennt.
ZWEIG="$RAUM/zweig"
mkdir -p "$ZWEIG/vendor" "$RAUM/modcache/beispiel.invalid/erzeugt@v1.2.3/build"
printf '# beispiel.invalid/erzeugt v1.2.3\n## explicit\nbeispiel.invalid/erzeugt\n' > "$ZWEIG/vendor/modules.txt"
printf '#!/bin/sh\ncurl -O https://beispiel.invalid/quelle.tar.gz\n' > "$RAUM/modcache/beispiel.invalid/erzeugt@v1.2.3/build/download.sh"
bauquellen_beilegen "$ZWEIG" "$RAUM/modcache" beispiel.invalid/erzeugt >/dev/null
pruefe "die Bauquellen des erzeugten Moduls liegen bei" test -s "$ZWEIG/quellen/beispiel.invalid/erzeugt@v1.2.3/build/download.sh"
bricht_ab() {
  bash -c "source '$WURZEL/scripts/releasequelltext.sh'; ende() { exit 7; }; bauquellen_beilegen '$ZWEIG' '$RAUM/modcache' beispiel.invalid/unbekannt" >/dev/null 2>&1
  [ $? -eq 7 ]
}
pruefe "ein Modul, das vendor nicht kennt, bricht ab" bricht_ab

# **Nur die Teile, aus denen die Form erzeugt ist** (Befund 1 des Nachlaufs
# über Phase 0): Wails bettet `runtime.js` ein, Tailscale seine vorgebaute
# Weboberfläche — die Quellen liegen in Unterverzeichnissen großer Module,
# und das ganze Modul beizulegen hieße Dutzende Megabyte für ein paar.
mkdir -p "$RAUM/modcache/beispiel.invalid/gross@v2.0.0/innen/laufzeit/src" "$RAUM/modcache/beispiel.invalid/gross@v2.0.0/anderes"
printf 'export const x = 1\n' > "$RAUM/modcache/beispiel.invalid/gross@v2.0.0/innen/laufzeit/src/index.ts"
printf 'groß\n' > "$RAUM/modcache/beispiel.invalid/gross@v2.0.0/anderes/viel.bin"
printf '# beispiel.invalid/gross v2.0.0\n## explicit\nbeispiel.invalid/gross\n' >> "$ZWEIG/vendor/modules.txt"
bauquellen_beilegen "$ZWEIG" "$RAUM/modcache" beispiel.invalid/gross:innen/laufzeit >/dev/null
pruefe "von einem großen Modul liegt nur der genannte Teil bei" test -s "$ZWEIG/quellen/beispiel.invalid/gross@v2.0.0/innen/laufzeit/src/index.ts"
pruefe "der Rest des großen Moduls liegt nicht bei" test ! -e "$ZWEIG/quellen/beispiel.invalid/gross@v2.0.0/anderes"

# **Die Bearbeitungsformen** (#2393): geladen in den Speicher, geprüft und
# beigelegt — und eine falsche Summe hält an, statt still ein anderes Archiv
# beizulegen.
LISTE="$RAUM/liste.tsv"
printf 'npm\tx\t1.0.0\t%040d\tfile://%s\t%s\t9\tnpm/gemeinsam.tar.gz\t\nnpm\ty\t2.0.0\t%040d\tfile://%s\t%s\t9\tnpm/gemeinsam.tar.gz\t\n' 0 "$RAUM/ursprung.tar.gz" "$SUMME" 0 "$RAUM/ursprung.tar.gz" "$SUMME" > "$LISTE"
bearbeitungsformen_beilegen "$ZWEIG" "$LISTE" "$RAUM/speicher" >/dev/null
pruefe "ein Ziel zweier Pakete liegt einmal bei" cmp -s "$RAUM/ursprung.tar.gz" "$ZWEIG/quellen/bearbeitungsformen/npm/gemeinsam.tar.gz"
pruefe "der Speicher behält das geladene Archiv" test -s "$RAUM/speicher/npm/gemeinsam.tar.gz"
falsche_summe() {
  sed 's/\t[0-9a-f]\{64\}\t/\t0000000000000000000000000000000000000000000000000000000000000000\t/' "$LISTE" > "$RAUM/falsch.tsv"
  bash -c "source '$WURZEL/scripts/releasequelltext.sh'; ende() { exit 7; }; bearbeitungsformen_beilegen '$RAUM/zweig2' '$RAUM/falsch.tsv' '$RAUM/speicher2'" >/dev/null 2>&1
  [ $? -eq 7 ] && [ ! -e "$RAUM/zweig2/quellen/bearbeitungsformen/npm/gemeinsam.tar.gz" ]
}
pruefe "eine falsche Summe bricht ab und legt nichts bei" falsche_summe
verdorbener_speicher() {
  mkdir -p "$RAUM/speicher3/npm" && printf 'verdorben\n' > "$RAUM/speicher3/npm/gemeinsam.tar.gz"
  bearbeitungsformen_beilegen "$RAUM/zweig3" "$LISTE" "$RAUM/speicher3" >/dev/null 2>&1 &&
    cmp -s "$RAUM/ursprung.tar.gz" "$RAUM/zweig3/quellen/bearbeitungsformen/npm/gemeinsam.tar.gz"
}
pruefe "ein verdorbenes Archiv im Speicher wird neu geladen" verdorbener_speicher

# **Die Lizenz je Go-Modul** — `QUELLTEXT.md` sagt sie zu, und `QUELLEN.md`
# nennt die Datei, die das Modul in `vendor/` mitbringt.
mkdir -p "$ZWEIG/vendor/beispiel.invalid/erzeugt" "$ZWEIG/vendor/beispiel.invalid/gross" "$RAUM/leer"
printf 'MIT\n' > "$ZWEIG/vendor/beispiel.invalid/erzeugt/LICENSE.md"
quellenverzeichnis "$ZWEIG" "$RAUM/leer"
pruefe "QUELLEN.md nennt die Lizenzdatei eines Go-Moduls" grep -q "| beispiel.invalid/erzeugt | v1.2.3 | vendor/beispiel.invalid/erzeugt/LICENSE.md |" "$ZWEIG/QUELLEN.md"
pruefe "QUELLEN.md sagt, wo ein Go-Modul keine mitbringt" grep -q "| beispiel.invalid/gross | v2.0.0 | keine in vendor/ |" "$ZWEIG/QUELLEN.md"

exit "$FEHLER"
