#!/usr/bin/env bash
#
# Prüfungen des geteilten Racelaufs aus scripts/teilrennen.sh — ENG-03
# (#1728).
#
# Gegen ein **eigenes kleines Modul** in einem Zwischenverzeichnis: fünf Tests,
# einer davon nebenläufig, einer, der nur im eigenen Verzeichnis grün ist —
# denn `go test` fährt ein Paket dort, und der geteilte Lauf muss es auch.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# shellcheck source=scripts/teilrennen.sh
source "$WURZEL/scripts/teilrennen.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0
RAUM="$(mktemp -d)"
trap 'rm -rf "$RAUM"' EXIT

gleich() { # gleich <beschreibung> <erwartet> <bekommen>
  if [ "$2" = "$3" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"
  else
    printf '%s  ✗ %s — erwartet „%s“, bekommen „%s“%s\n' "$ROT" "$1" "$2" "$3" "$AUS"
    FEHLER=1
  fi
}

# ------------------------------------------------------------ das Probemodul

mkdir -p "$RAUM/modul/probe"
cat > "$RAUM/modul/go.mod" <<'MOD'
module probe.example

go 1.27
MOD
cat > "$RAUM/modul/probe/probe.go" <<'GO'
package probe

func Eins() int { return 1 }

func Zwei() int { return 2 }

func Nie() int { return 0 }

// Zerlegen scheitert an einer leeren Liste — so wie ein Fehler, der nur
// unter Last und in einer fremden Goroutine auftritt.
func Zerlegen(liste []int) []int { return liste[1:] }
GO
cat > "$RAUM/modul/probe/probe_test.go" <<'GO'
package probe

import (
	"os"
	"testing"
	"time"
)

func TestA(t *testing.T) { if Eins() != 1 { t.Fatal() } }
func TestB(t *testing.T) { if Zwei() != 2 { t.Fatal() } }
func TestC(t *testing.T) { t.Parallel(); if Eins() != 1 { t.Fatal() } }
func TestD(t *testing.T) { if os.Getenv("PROBE_ROT") != "" { t.Fatal("rot wie verlangt") } }

// TestF stirbt in einer eigenen Goroutine, wenn man es verlangt — ohne
// `--- FAIL`, denn das Testprogramm endet, bevor der Test fertig ist.
func TestF(t *testing.T) {
	if os.Getenv("PROBE_PANIK") == "" { return }
	fertig := make(chan struct{})
	go func() { defer close(fertig); Zerlegen(nil) }()
	<-fertig
}

// TestG hängt, wenn man es verlangt — und TestH fängt danach an, damit der
// zuletzt begonnene Test nicht der hängende ist.
func TestG(t *testing.T) {
	t.Parallel()
	if os.Getenv("PROBE_HAENGT") == "" { return }
	time.Sleep(time.Hour)
}
func TestH(t *testing.T) { t.Parallel() }

// TestE liest eine Datei neben sich — grün nur im Verzeichnis des Pakets.
func TestE(t *testing.T) {
	if _, err := os.Stat("probe.go"); err != nil { t.Fatal(err) }
}
GO

cd "$RAUM/modul" || exit 1

# ------------------------------------------------------------- die Teilung

GETEILT="./probe"
TEILE=2

teile_starten "$RAUM/teile" > /dev/null
teile_abwarten "$RAUM/teile" "$RAUM/geteilt.out" > "$RAUM/bericht" 2>&1
rueckgabe=$?

gleich "sechs grüne Tests in zwei Teilen sind grün" "0" "$rueckgabe"

gelaufen="$(cat "$RAUM/teile"/probe-*.log | grep -oE '^(--- PASS|=== RUN) +Test[A-F]' | awk '{print $NF}' | sort -u | paste -sd' ' -)"
gleich "jeder Test läuft, und keiner fehlt" "TestA TestB TestC TestD TestE TestF" "$gelaufen"

zweimal="$(cat "$RAUM/teile"/probe-*.log | grep -oE '^=== RUN +Test[A-F]$' | sort | uniq -d | paste -sd' ' -)"
gleich "kein Test läuft in zwei Teilen" "" "$zweimal"

# **Das Profil zählt jeden Block einmal** — zwei Teile, die denselben Block
# melden, dürfen die Anweisungen nicht verdoppeln, sonst sänke die Abdeckung
# mit jedem Teil.
bloecke="$(grep -c 'probe.go:' "$RAUM/geteilt.out")"
gleich "jeder Block steht einmal im vereinten Profil" "4" "$bloecke"
gleich "das Profil beginnt mit seiner Art" "mode: atomic" "$(head -1 "$RAUM/geteilt.out")"

# `Eins` rufen zwei Tests in verschiedenen Teilen — die Zählungen addieren sich.
eins="$(grep 'probe.go:3\.' "$RAUM/geteilt.out" | awk '{print $3}')"
gleich "die Zählungen der Teile addieren sich" "2" "$eins"
nie="$(grep 'probe.go:7\.' "$RAUM/geteilt.out" | awk '{print $3}')"
gleich "ein nie gelaufener Block bleibt bei null" "0" "$nie"

# ------------------------------------------------ der Dauerbeleg (T08)

# **Ein grüner Lauf hinterlässt seine Zeiten** — T08
# (#2101).
# Bis zum 2026-10-06 löschte die Pipeline die Protokolle der Teile nach
# einem grünen Lauf, und mit ihnen Übersetzungszeit, Wandzeit und die Dauer
# jedes Tests; wer die Teststufe schneller machen wollte, maß von vorn.
teile_belegen "$RAUM/teile" "$RAUM/dauer.tsv"
gleich "der Dauerbeleg nennt die Übersetzung" "1" "$(grep -c "^bau	probe	" "$RAUM/dauer.tsv")"
gleich "und beide Teile mit ihrer Wandzeit" "2" "$(grep -cE "^teil	probe	[12]	-	[0-9]+$" "$RAUM/dauer.tsv")"
gleich "und jeden Test mit seiner Dauer" "TestA TestB TestC TestD TestE TestF TestG TestH" \
  "$(awk -F'\t' '$1 == "test" && $5 ~ /^[0-9.]+$/ { print $4 }' "$RAUM/dauer.tsv" | sort -u | paste -sd' ' -)"

# **Und er bleibt, aber nicht unbegrenzt** — die jüngsten stehen.
mkdir -p "$RAUM/dauer"
for alt in 20260101T000000Z-a 20260102T000000Z-b 20260103T000000Z-c; do : > "$RAUM/dauer/$alt.tsv"; done
TEILBELEGE=2 teile_aufbewahren "$RAUM/teile" "$RAUM/dauer" > /dev/null
gleich "vom Dauerbeleg bleiben die jüngsten zwei, darunter der neue" "20260103T000000Z-c.tsv 2 1" \
  "$(find "$RAUM/dauer" -name '*.tsv' | sort | sed 's|.*/||' | head -1) $(find "$RAUM/dauer" -name '*.tsv' | wc -l | tr -d ' ') $(grep -l '^bau' "$RAUM/dauer"/*.tsv | wc -l | tr -d ' ')"

# ------------------------------------------------------------ der rote Teil

PROBE_ROT=1 teile_starten "$RAUM/rot" > /dev/null
teile_abwarten "$RAUM/rot" "$RAUM/rot.out" > "$RAUM/rotbericht" 2>&1
rueckgabe=$?

gleich "ein roter Test macht den Lauf rot" "1" "$rueckgabe"

if grep -q -- '--- FAIL: TestD' "$RAUM/rotbericht" && grep -q 'rot wie verlangt' "$RAUM/rotbericht"; then
  gleich "der Bericht nennt den roten Test und seine Meldung" "ja" "ja"
else
  gleich "der Bericht nennt den roten Test und seine Meldung" "ja" "nein"
fi

# -------------------------------------------------------- die Panik

# **Eine Panik in einer fremden Goroutine nennt ihren Ort** — am 2026-09-24
# zeigte der Bericht nur „panic: runtime error: slice bounds out of range
# [1:0]": kein Test, keine Datei, keine Zeile, und die Protokolle der Teile
# waren schon gelöscht. Ein roter Lauf, der seinen Fund nicht benennt,
# kostet einen zweiten vollen Lauf (AGENTS.md §4).
PROBE_PANIK=1 teile_starten "$RAUM/panik" > /dev/null
teile_abwarten "$RAUM/panik" "$RAUM/panik.out" > "$RAUM/panikbericht" 2>&1
gleich "eine Panik macht den Lauf rot" "1" "$?"
gleich "der Bericht nennt die Funktion, in der sie geschah" "1" \
  "$(grep -c 'probe.example/probe.Zerlegen' "$RAUM/panikbericht")"
gleich "und ihre Zeile" "1" "$(grep -cE 'probe/probe.go:[0-9]+' "$RAUM/panikbericht")"
gleich "und den Test, der zuletzt anfing" "1" "$(grep -c 'zuletzt angefangen: TestF' "$RAUM/panikbericht")"

# ------------------------------------------------------ der Hänger

# **Ein Teil, der in seine Frist läuft, nennt, was noch lief** — Zweitblick
# über Phase 6.11 (#1672): am 2026-10-04 hing ein Teil von `internal/app`
# dreißig Minuten, und der Bericht nannte nur den Test, der zuletzt anfing —
# einen, der allein in vier Sekunden grün ist. Go schreibt die laufenden
# Tests selbst unter „running tests:", und genau das fiel weg.
PROBE_HAENGT=1 TEILFRIST=3s teile_starten "$RAUM/haengt" > /dev/null
teile_abwarten "$RAUM/haengt" "$RAUM/haengt.out" > "$RAUM/haengtbericht" 2>&1
gleich "ein hängender Teil macht den Lauf rot" "1" "$?"
gleich "der Bericht nennt den Test, der noch lief" "1" \
  "$(grep -c 'noch lief: TestG' "$RAUM/haengtbericht")"
gleich "und nicht den, der nur zuletzt anfing" "0" \
  "$(grep -c 'noch lief: TestH' "$RAUM/haengtbericht")"

# ------------------------------------------------------ nichts gelaufen

# **Kein Teil ist kein grüner Lauf**: ein Paket, das sich nicht finden lässt,
# darf nicht als geprüft durchgehen.
GETEILT="./gibtsnicht"
teile_starten "$RAUM/leer" > /dev/null 2>&1
teile_abwarten "$RAUM/leer" "$RAUM/leer.out" > "$RAUM/leerbericht" 2>&1
gleich "ohne einen einzigen Teil ist der Lauf rot" "1" "$?"
gleich "und der Bericht sagt, warum" "1" "$(grep -c 'kein Teil ist gelaufen' "$RAUM/leerbericht")"

printf '\n'
if [ "$FEHLER" -eq 0 ]; then
  printf '%s✓ geteilter Racelauf in Ordnung%s\n' "$GRUEN" "$AUS"
else
  printf '%s✗ geteilter Racelauf fehlerhaft%s\n' "$ROT" "$AUS"
fi

exit "$FEHLER"
