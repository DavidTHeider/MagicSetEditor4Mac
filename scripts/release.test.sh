#!/usr/bin/env bash
#
# Prüfungen der Versionslogik aus scripts/releaseversion.sh.
#
# Sie laufen gegen **erfundene** Stände in einem Wegwerfverzeichnis. Der Grund
# ist derselbe wie beim Schichtenwächter: die Gegenprobe von Hand zu fahren
# hieße, die echten Dateien zu verfälschen, den roten Lauf anzusehen und alles
# zurückzunehmen — danach wäre der Beleg wieder fort.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# shellcheck source=scripts/releaseversion.sh
source "$WURZEL/scripts/releaseversion.sh"

# shellcheck source=scripts/releasesignatur.sh
source "$WURZEL/scripts/releasesignatur.sh"

# shellcheck source=scripts/releasebeleg.sh
source "$WURZEL/scripts/releasebeleg.sh"

# shellcheck source=scripts/releasesignieren.sh
source "$WURZEL/scripts/releasesignieren.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0

behauptung() { # behauptung <beschreibung> <erwarteter-exit> <kommando…>
  local beschreibung="$1" erwartet="$2"
  shift 2

  local ausgabe status
  ausgabe="$("$@" 2>&1)"
  status=$?

  if [ "$status" -eq "$erwartet" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$beschreibung" "$AUS"
  else
    printf '%s  ✗ %s — Exit %d, erwartet war %d%s\n' "$ROT" "$beschreibung" "$status" "$erwartet" "$AUS"
    printf '      %s\n' "$ausgabe"
    FEHLER=1
  fi
}

stand() { # stand <verzeichnis> <wails-version> <changelog-überschrift> [paket-version]
  mkdir -p "$1"
  printf '{\n  "productVersion": "%s"\n}\n' "$2" > "$1/wails.json"
  printf '# Änderungsprotokoll\n\n%s\n\n- etwas\n' "$3" > "$1/CHANGELOG.md"
  printf '{\n  "version": "%s"\n}\n' "${4:-$2}" > "$1/package.json"
}

pruefe_in() { # pruefe_in <verzeichnis> <version>
  versionen_pruefen "$2" "$1/wails.json" "$1/CHANGELOG.md" "$1/package.json"
}

ARBEIT="$(mktemp -d)"
trap 'rm -rf "$ARBEIT"' EXIT

# **Der Stand des aufrufenden Repositoriums, bevor irgendetwas läuft.**
#
# Am 2026-08-21 hat dieser Test aus einem `git worktree` heraus im **geteilten**
# Repositorium einen Commit angelegt, ihn `v0.1.0` getaggt und `core.bare` auf
# `true` gestellt. Danach ging im Hauptbaum kein `git status` mehr, und der Test
# selbst war dauerhaft rot — auch auf `main`, auch ohne jede Änderung
# (#734).
#
# **Ein Wächter, der nur das Ergebnis prüft, hätte das nicht gesehen.** Deshalb
# steht hier eine Aufnahme vorher und ein Vergleich nachher.
VORHER_TAGS="$(git tag -l 2>/dev/null | sort | tr '\n' ' ')"
VORHER_KOPF="$(git rev-parse -q --verify HEAD 2>/dev/null || printf 'keiner')"
VORHER_BARE="$(git config --get core.bare 2>/dev/null || printf 'ungesetzt')"

printf '\n── Versionsprüfung der Freigabe\n'

stand "$ARBEIT/gut" "0.1.0" "## [0.1.0] — 2026-08-08"
behauptung "drei übereinstimmende Angaben gehen durch" 0 pruefe_in "$ARBEIT/gut" "0.1.0"

behauptung "eine abweichende Wunschversion wird abgewiesen" 1 pruefe_in "$ARBEIT/gut" "0.2.0"

stand "$ARBEIT/wails-alt" "0.0.9" "## [0.1.0] — 2026-08-08"
behauptung "wails.json läuft weg" 1 pruefe_in "$ARBEIT/wails-alt" "0.1.0"

stand "$ARBEIT/ohne-datum" "0.1.0" "## [0.1.0]"
behauptung "eine Fassung ohne Datum gilt nicht" 1 pruefe_in "$ARBEIT/ohne-datum" "0.1.0"

stand "$ARBEIT/nur-sammel" "0.1.0" "## [Unveröffentlicht]"
behauptung "die Sammelstelle ist keine Fassung" 1 pruefe_in "$ARBEIT/nur-sammel" "0.1.0"

stand "$ARBEIT/gut2" "0.1.0" "## [0.1.0] — 2026-08-08"
behauptung "keine Version angegeben" 1 pruefe_in "$ARBEIT/gut2" ""
behauptung "keine Version der Form x.y.z" 1 pruefe_in "$ARBEIT/gut2" "v0.1"

stand "$ARBEIT/paket-alt" "0.1.0" "## [0.1.0] — 2026-08-08" "0.0.9"
behauptung "frontend/package.json läuft weg" 1 pruefe_in "$ARBEIT/paket-alt" "0.1.0"

# Der Tag darf es noch nicht geben. Geprüft in einem eigenen Repository, damit
# der Test nicht davon abhängt, welche Tags dieses hier gerade führt.
TAGPROBE="$ARBEIT/tagprobe"
stand "$TAGPROBE" "0.1.0" "## [0.1.0] — 2026-08-08"
(
  cd "$TAGPROBE" || exit 1
  git init -q . && git add -A && git -c commit.gpgsign=false commit -qm "Probe"
) >/dev/null 2>&1

# **Es muss ein eigenes sein.** Schlägt `git init` fehl, liefe jedes `git` hier
# gegen das aufrufende Repositorium — und genau so ist der Schaden entstanden.
if [ "$(git -C "$TAGPROBE" rev-parse --show-toplevel 2>/dev/null)" != "$(cd "$TAGPROBE" && pwd -P)" ]; then
  printf '  ✗ das Probeverzeichnis ist kein eigenes Repositorium — abgebrochen\n'
  exit 1
fi

mit_tag() { ( cd "$TAGPROBE" && versionen_pruefen "$1" wails.json CHANGELOG.md package.json ); }

behauptung "ohne vorhandenen Tag geht es durch" 0 mit_tag "0.1.0"
git -C "$TAGPROBE" tag v0.1.0 >/dev/null 2>&1
behauptung "ein bereits vorhandener Tag wird abgewiesen" 1 mit_tag "0.1.0"

printf '\n── Was der Lauf im aufrufenden Repositorium hinterlässt\n'

behauptung "die Tagliste ist unverändert" 0 \
  test "$VORHER_TAGS" = "$(git tag -l 2>/dev/null | sort | tr '\n' ' ')"
behauptung "der Kopf ist unverändert" 0 \
  test "$VORHER_KOPF" = "$(git rev-parse -q --verify HEAD 2>/dev/null || printf 'keiner')"
behauptung "core.bare ist unverändert" 0 \
  test "$VORHER_BARE" = "$(git config --get core.bare 2>/dev/null || printf 'ungesetzt')"

printf '\n── Das Urteil über spctl\n'

# **Die Gegenprobe zum fail-open**, das der Zweitblick über Phase 8 gefunden
# hat: vorher galt *jeder* Status ungleich null als erwartete Ablehnung. Diese
# vier Zeilen sind der Beleg, dass das nicht mehr gilt — 127 ist das fehlende
# Werkzeug, 1 der falsche Pfad, 130 der Abbruch.
behauptung "Status 3 ist die erwartete Ablehnung" 0 spctl_urteil 3
behauptung "Status 0 ist eine Annahme und damit ein Abbruch" 1 spctl_urteil 0
behauptung "Status 127 — spctl fehlt — gilt nicht als Ablehnung" 1 spctl_urteil 127
behauptung "Status 1 — falscher Pfad — gilt nicht als Ablehnung" 1 spctl_urteil 1

# Und der Text sagt, welcher der drei Fälle es war; sonst stünde im Protokoll
# dreimal dasselbe.
#
# **Erst auffangen, dann durchsuchen** — dieselbe Falle wie bei `codesign` in
# `release.sh`, nur andersherum: unter `set -o pipefail` trägt die Leitung den
# Rückgabewert von `spctl_urteil`, und der ist hier absichtlich 1. Ein
# `spctl_urteil 127 | grep -q …` prüfte damit nicht den Text, sondern meldete
# immer einen Fehlschlag.
URTEILSTEXT="$(spctl_urteil 127 || true)"

if ! printf '%s' "$URTEILSTEXT" | grep -q 'Werkzeugfehler'; then
  printf '%s  ✗ Status 127 wird nicht als Werkzeugfehler benannt%s\n' "$ROT" "$AUS"
  FEHLER=1
else
  printf '%s  ✓ Status 127 wird als Werkzeugfehler benannt%s\n' "$GRUEN" "$AUS"
fi

# ------------------------------------------------------ Der Freigabebeleg

printf '\n%sDer Freigabebeleg%s\n' "$GRUEN" "$AUS"

BELEGRAUM="$(mktemp -d)"
trap 'rm -rf "$BELEGRAUM"' EXIT

cat > "$BELEGRAUM/pruefsummen.txt" <<'PRUEF'
abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd  ./MagicSetEditor4Mac-9.9.9.dmg
ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00  ./stueckliste-go.json
PRUEF

# **Der Hash kommt aus der Datei, nicht aus einer zweiten Rechnung.**
GELESEN="$(artefakthash "$BELEGRAUM/pruefsummen.txt")"
if [ "$GELESEN" = "abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd" ]; then
  printf '%s  ✓ der Artefakthash kommt aus der Prüfsummendatei%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ der Artefakthash ist %s%s\n' "$ROT" "$GELESEN" "$AUS"
  FEHLER=1
fi

# **Eine Prüfsummendatei ohne Abbild ist kein Beleg**, sondern ein Fehler.
cat > "$BELEGRAUM/ohnedmg.txt" <<'PRUEF'
ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00ff00  ./stueckliste-go.json
PRUEF
behauptung "eine Prüfsummendatei ohne .dmg wird abgewiesen" 1 \
  artefakthash "$BELEGRAUM/ohnedmg.txt"

BELEG="$BELEGRAUM/$(belegname 9.9.9)"
belegschreiben "$BELEG" "9.9.9" "deadbeef" "$GELESEN" "$BELEGRAUM/pruefsummen.txt"

# **Die Handprüfungen stehen offen darin.** Ein Beleg, der sie vorab abhakt,
# ist schlimmer als keiner.
if [ "$(grep -c -- '- \[ \]' "$BELEG")" -eq 7 ]; then
  printf '%s  ✓ sieben Handprüfungen stehen offen im Beleg%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ %s offene Handprüfungen, erwartet sind sieben%s\n' \
    "$ROT" "$(grep -c -- '- \[ \]' "$BELEG")" "$AUS"
  FEHLER=1
fi

# **Der Beleg nennt das Artefakt und den Commit.**
for erwartet in "$GELESEN" "deadbeef"; do
  if grep -qF "$erwartet" "$BELEG"; then
    printf '%s  ✓ der Beleg nennt %.12s…%s\n' "$GRUEN" "$erwartet" "$AUS"
  else
    printf '%s  ✗ der Beleg nennt %.12s… nicht%s\n' "$ROT" "$erwartet" "$AUS"
    FEHLER=1
  fi
done

# **Ein Haken braucht den passenden Hash.** Das ist die ganze Zusage dieser
# Datei: ein Beleg lässt sich nicht auf ein anderes Abbild fortschreiben.
behauptung "ein Haken mit fremdem Artefakthash wird abgewiesen" 1 \
  belegzeile "$BELEG" "0000000000000000000000000000000000000000000000000000000000000000" 1 "wer" "2026-08-25"

behauptung "eine Prüfung, die es nicht gibt, wird abgewiesen" 1 \
  belegzeile "$BELEG" "$GELESEN" 9 "wer" "2026-08-25"

behauptung "ein Haken mit passendem Hash wird eingetragen" 0 \
  belegzeile "$BELEG" "$GELESEN" 5 "David" "2026-08-25"

if grep -q 'Prüfung 5 · David' "$BELEG"; then
  printf '%s  ✓ der eingetragene Haken steht im Beleg%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ der eingetragene Haken fehlt%s\n' "$ROT" "$AUS"
  FEHLER=1
fi

# **Und das Kästchen ist wirklich angekreuzt.**
#
# Bis zum 2026-08-26 war es das **nicht**: `belegzeile` hängte die Zeile unter
# „Eingetragen" an und ließ die Liste darüber stehen. Ein Beleg, dessen
# Aufzählung sieben offene Kästchen zeigt, während darunter drei erledigte
# Prüfungen stehen, sagt beides — und wer nur den Kopf liest, liest das
# Falsche. Aufgefallen beim Eintragen der ersten drei Prüfungen für 0.2.0.
if grep -q -- '- \[x\] \*\*5 Druckbogen' "$BELEG"; then
  printf '%s  ✓ das Kästchen der Prüfung 5 ist angekreuzt%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ das Kästchen der Prüfung 5 steht noch offen%s\n' "$ROT" "$AUS"
  FEHLER=1
fi

# **Und die anderen sechs stehen weiter offen.** Ein Ankreuzen, das alle
# trifft, wäre schlimmer als keines.
if [ "$(grep -c -- '- \[ \] \*\*' "$BELEG")" -eq 6 ]; then
  printf '%s  ✓ die übrigen sechs stehen weiter offen%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ es stehen %s statt sechs Kästchen offen%s\n' \
    "$ROT" "$(grep -c -- '- \[ \] \*\*' "$BELEG")" "$AUS"
  FEHLER=1
fi

# ------------------------------------------------------- Die GPG-Signatur
#
# **Sie wird wirklich gefahren**, nicht nachgebildet: der Schlüssel liegt auf
# diesem Rechner, und eine Signaturprüfung gegen eine erfundene Signatur
# prüfte den Aufbau der Zeichenkette und nicht die Signatur.
#
# **Ohne Schlüssel wird übersprungen und gesagt, dass übersprungen wurde.** Ein
# Lauf, der ohne gpg still grün ist, behauptet etwas über eine Prüfung, die
# nicht stattgefunden hat (§10).

printf '\n'

SIGNATURWERK="$(mktemp -d)"
trap 'rm -rf "$SIGNATURWERK"' EXIT

# fingerabdruckSteht gibt 0, wenn der Fingerabdruck vierzig Hexzeichen hat.
#
# **Als eigene Hülle**, weil `behauptung` einen Rückgabewert vergleicht und
# `signaturfingerabdruck` einen Text herausgibt: ohne sie ginge eine leere
# Ausgabe als Erfolg durch, und genau das ist der Fehler, den dieser Fall
# fangen soll.
fingerabdruckSteht() { # fingerabdruckSteht <datei>
  local gefunden
  gefunden="$(signaturfingerabdruck "$1")"

  if printf '%s' "$gefunden" | grep -qE '^[0-9A-F]{40}$'; then
    return 0
  fi

  printf 'der Fingerabdruck ist %s\n' "${gefunden:-leer}" >&2

  return 1
}

printf 'die Prüfsummen einer erfundenen Freigabe\n' > "$SIGNATURWERK/pruefsummen.txt"

behauptung "eine Datei, die es nicht gibt, wird abgewiesen" 1 \
  signiere "$SIGNATURWERK/gibtsnicht.txt"

behauptung "ohne Schlüssel wird abgewiesen" 1 \
  signiere "$SIGNATURWERK/pruefsummen.txt" ""

behauptung "eine fehlende Signatur wird gemeldet" 1 \
  signaturpruefen "$SIGNATURWERK/pruefsummen.txt"

if [ -n "$(signaturschluessel)" ] && command -v gpg >/dev/null 2>&1; then
  behauptung "die Prüfsummen lassen sich signieren" 0 \
    signiere "$SIGNATURWERK/pruefsummen.txt"

  behauptung "und die Signatur passt" 0 \
    signaturpruefen "$SIGNATURWERK/pruefsummen.txt"

  # **Und der Fingerabdruck steht wirklich da.**
  #
  # Er hat bis zum 2026-08-26 **gefehlt**, und niemand hat es gesehen: die
  # erste Fassung las ihn aus der Meldung von `gpg` — also aus einem Satz in
  # der Sprache des Rechners. Auf Deutsch heißt die Zeile „mittels
  # EDDSA-Schlüssel …", das Muster suchte „Fingerabdruck … = …", und der
  # Freigabelauf sagte brav „Signatur nachgesehen: unbekannt". Ein Beleg, der
  # „signiert" sagt und nicht **womit**, lässt die eine Frage offen, für die
  # es ihn gibt (§10: ein Wächter, der die halbe Zusage hält).
  behauptung "der Fingerabdruck kommt heraus" 0 \
    fingerabdruckSteht "$SIGNATURWERK/pruefsummen.txt"

  # **Und signiert wurde mit dem Schlüssel, gegen den die Anwendung prüft.**
  #
  # Signiert wird mit dem, was `gpg` auf **diesem Rechner** unter der Kennung
  # findet; die ausgelieferte Anwendung prüft gegen den eingebetteten
  # Schlüssel. Laufen sie auseinander, ist der Freigabelauf grün — und jede
  # installierte Fassung weist die Freigabe ab. Angestrichen vom Nachlauf des
  # Gesamtreviews.
  behauptung "der Schlüssel der Freigabe ist der der Anwendung" 0 \
    schluesselstimmtueberein "$SIGNATURWERK/pruefsummen.txt"

  # **Die Gegenprobe, und sie unterscheidet wirklich**: eine Quelle, die einen
  # **anderen** Schlüssel nennt, macht denselben Aufruf rot. Ohne sie stünde
  # nur fest, dass die Prüfung etwas zurückgibt.
  printf 'const Signaturschluessel = "0000000000000000"\n' \
    > "$SIGNATURWERK/signatur.go"

  behauptung "ein fremder Anwendungsschlüssel wird abgewiesen" 1 \
    schluesselstimmtueberein "$SIGNATURWERK/pruefsummen.txt" "$SIGNATURWERK/signatur.go"

  # **Und eine Quelle ohne Schlüssel ebenso.**
  printf 'kein Schlüssel hier\n' > "$SIGNATURWERK/ohne.go"

  behauptung "eine Quelle ohne Schlüssel wird abgewiesen" 1 \
    schluesselstimmtueberein "$SIGNATURWERK/pruefsummen.txt" "$SIGNATURWERK/ohne.go"

  # **Die Gegenprobe.** Eine Signatur, die auch zu einem geänderten Inhalt
  # passt, ist keine — und ohne diesen Fall stünde nur fest, dass `gpg`
  # zweimal dasselbe sagt.
  printf 'jemand hat die Zeile ausgetauscht\n' > "$SIGNATURWERK/pruefsummen.txt"

  behauptung "eine geänderte Datei fällt durch" 1 \
    signaturpruefen "$SIGNATURWERK/pruefsummen.txt"
else
  printf '  ○ übersprungen: kein Signaturschlüssel oder kein gpg auf diesem Rechner\n'
fi

# ------------------------------------- Der Veröffentlichungsbefehl ist voll
#
# **Der Befund, aus dem dieser Fall entstanden ist**
# (#1170):
# `release.sh` erzeugt und prüft `pruefsummen.txt.asc`, nannte sie aber im
# abschließend ausgegebenen `gh release create` **nicht**. Wer genau diesem
# Befehl folgte, veröffentlichte eine Fassung ohne Signatur — und
# `aktualisierung.Laden` weist die ab, bevor es das Abbild lädt.
#
# **Ein Wächter über den Text und nicht über den Lauf.** Der Befehl wird
# ausgegeben, nicht ausgeführt; was ihn hält, ist, dass er **jede** Datei
# nennt, die das Skript daneben legt.
#
# **Was er nicht hält**: ob die genannte Datei wirklich entsteht. Das prüft der
# Bau, und der läuft hier nicht.
#
# **Gelesen wird die `printf`-Zeile, nicht der Bereich darum.** Der erste
# Anlauf nahm alles von `gh release create` bis `notes-file` — und traf damit
# den **Kommentar**, der `pruefsummen.txt.asc` erklärt. Die Gegenprobe blieb
# grün: der Wächter maß den Text über dem Befehl statt den Befehl (§10).
veroeffentlichungsbefehl() {
  # **Die `printf`-Zeile und ihre Argumentzeile.** Der Belegname steht nicht
  # im Muster, sondern als `$(belegname …)` darunter — beide zusammen sind der
  # Befehl.
  grep -A2 -F "printf '     gh release create" "$WURZEL/scripts/release.sh"
}

for datei in "*.dmg" "*.json" "pruefsummen.txt" "pruefsummen.txt.asc" "belegname"; do
  if veroeffentlichungsbefehl | grep -qF -- "$datei"; then
    behauptung "der Veröffentlichungsbefehl nennt $datei" 0 true
  else
    behauptung "der Veröffentlichungsbefehl nennt $datei" 0 false
  fi
done

printf '\n'
if [ "$FEHLER" -eq 0 ]; then
  printf '%s✓ Versions-, Signatur-, Beleg- und GPG-Prüfung in Ordnung%s\n' "$GRUEN" "$AUS"
else
  printf '%s✗ Versions-, Signatur-, Beleg- oder GPG-Prüfung fehlerhaft%s\n' "$ROT" "$AUS"
fi

exit "$FEHLER"
