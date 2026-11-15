#!/usr/bin/env bash
#
# Die Versionsprüfung der Freigabe (docs/RELEASE.md).
#
# Ausgelagert aus scripts/release.sh, damit sie sich prüfen lässt, ohne etwas zu
# bauen — `scripts/release.test.sh` fährt sie gegen erfundene Stände.
#
# Vier Orte müssen dieselbe Version nennen: wails.json, CHANGELOG.md,
# frontend/package.json — und der Tag darf es noch **nicht** geben. Läuft einer
# davon weg, entsteht ein Artefakt, dessen „Über"-Dialog etwas anderes sagt als
# die Release-Notiz, und niemand merkt es, weil beides für sich stimmt.

# version_aus_wails liest productVersion aus wails.json.
version_aus_wails() {
  local datei="${1:-wails.json}"

  sed -n 's/.*"productVersion"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$datei" | head -1
}

# version_aus_changelog liest die oberste **datierte** Fassung.
#
# Absichtlich nur datierte: `## [Unveröffentlicht]` ist keine Fassung, sondern
# der Ort, an dem die nächste entsteht. Und eine Fassung mit Nummer ohne Datum
# ist ein vergessenes Datum, kein Sammelplatz — sie wird nicht als gültig
# angenommen, sondern übergangen, damit die Prüfung darunter anschlägt.
version_aus_changelog() {
  local datei="${1:-CHANGELOG.md}"

  sed -n 's/^## \[\([0-9][^]]*\)\][[:space:]]*—[[:space:]]*[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9].*/\1/p' \
    "$datei" | head -1
}

# version_aus_paketbeschreibung liest "version" aus frontend/package.json.
#
# Sie steht dort nicht aus Ordnungsliebe: `npm sbom` weigert sich ohne sie
# („Purl can only be generated for version types"), und die Stückliste ist ein
# Releaseartefakt. Ein vierter Ort, an dem eine Version steht, ist ein vierter
# Ort, an dem sie wegläuft — also wird sie mitgeprüft.
version_aus_paketbeschreibung() {
  local datei="${1:-frontend/package.json}"

  sed -n 's/^[[:space:]]*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$datei" | head -1
}

# versionen_pruefen vergleicht die vier Orte.
#
# Gibt 0 zurück, wenn alle übereinstimmen; sonst schreibt sie den Grund nach
# stderr und gibt 1 zurück.
versionen_pruefen() {
  local gewuenscht="$1" wailsdatei="${2:-wails.json}" changelogdatei="${3:-CHANGELOG.md}"
  local paketdatei="${4:-frontend/package.json}"

  # **Das Repositorium, das geprüft wird, und nicht das, in dem wir stehen.**
  #
  # Bis zum 2026-08-21 fragte die Tagprüfung unten `git` ohne Ortsangabe — also
  # das umgebende Repositorium. Der Test daneben reicht Dateien aus einem
  # Probeverzeichnis herein und stand dabei im **echten** Baum: ein Tag `v0.1.0`
  # dort machte „drei übereinstimmende Angaben gehen durch" dauerhaft rot, ohne
  # dass sich an den geprüften Dateien etwas geändert hätte
  # (#734).
  #
  # **Der Ort folgt der ersten Datei.** Wer `wails.json` aus `/tmp/probe` prüft,
  # meint die Fassung dort — und deren Tags stehen im Repositorium **dort**,
  # nicht im aufrufenden.
  local repoort
  repoort="$(dirname "$wailsdatei")"

  if [ -z "$gewuenscht" ]; then
    printf 'release: keine Version angegeben — Aufruf: release.sh <version>\n' >&2

    return 1
  fi

  if ! printf '%s' "$gewuenscht" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    printf 'release: %s ist keine Version der Form <haupt>.<neben>.<fehlerbehebung>\n' \
      "$gewuenscht" >&2

    return 1
  fi

  local ausWails ausChangelog fehler=0
  ausWails="$(version_aus_wails "$wailsdatei")"
  ausChangelog="$(version_aus_changelog "$changelogdatei")"

  if [ "$ausWails" != "$gewuenscht" ]; then
    printf 'release: %s nennt %s, freigegeben werden soll %s\n' \
      "$wailsdatei" "${ausWails:-nichts}" "$gewuenscht" >&2
    fehler=1
  fi

  if [ "$ausChangelog" != "$gewuenscht" ]; then
    printf 'release: %s nennt als oberste datierte Fassung %s, freigegeben werden soll %s\n' \
      "$changelogdatei" "${ausChangelog:-nichts}" "$gewuenscht" >&2
    printf '        (eine Fassung mit Nummer und ohne Datum gilt nicht — sie ist ein\n' >&2
    printf '         vergessenes Datum, kein Sammelplatz)\n' >&2
    fehler=1
  fi

  if [ -f "$paketdatei" ]; then
    local ausPaket
    ausPaket="$(version_aus_paketbeschreibung "$paketdatei")"

    if [ "$ausPaket" != "$gewuenscht" ]; then
      printf 'release: %s nennt %s, freigegeben werden soll %s\n' \
        "$paketdatei" "${ausPaket:-nichts}" "$gewuenscht" >&2
      fehler=1
    fi
  fi

  if git -C "$repoort" rev-parse -q --verify "refs/tags/v$gewuenscht" >/dev/null 2>&1; then
    printf 'release: den Tag v%s gibt es bereits — diese Fassung ist schon heraus\n' \
      "$gewuenscht" >&2
    fehler=1
  fi

  return "$fehler"
}
