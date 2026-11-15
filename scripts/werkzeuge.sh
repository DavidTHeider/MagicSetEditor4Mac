#!/usr/bin/env bash
#
# Die drei Werkzeuge ins Bündel
# (#992).
#
#     ./scripts/werkzeuge.sh build/bin/MagicSetEditor4Mac.app
#
# ## Der Befund, der dazu geführt hat
#
# `scripts/release.sh` baute **ein** Programm — das Fenster. `mse`, `mse-mcp`
# und `mse-rest` waren gebaut, geprüft und über vier Flächen abgesichert, und
# **erreichten niemanden, der die Anwendung installiert**: wer sie benutzen
# wollte, brauchte das Repositorium und eine Go-Werkzeugkette.
#
# **Drei und nicht zwei.** #992 nennt `mse` und `mse-mcp`; `mse-rest` kam erst
# in Phase U dazu und hat dieselbe Lücke. Zwei von drei Flächen mitzuliefern
# wäre genau die Ungleichheit, gegen die §16 steht — und der Satz „es liegt
# diesem Bündel nicht bei" stand auf der Hilfeseite für alle drei.
#
# Das ist die erste Dimension aus AGENTS.md §10 — dieselbe Gestalt wie #852:
# volle Abdeckung, kein Weg zum Benutzenden.
#
# ## Warum `Contents/MacOS/` und nicht `Resources/`
#
# **Weil dort ausführbare Dateien hingehören.** Apples Bündelaufbau sieht
# `Contents/MacOS/` für das Programm und seine Helfer vor; `Resources/` ist für
# alles, was nicht läuft. Dort liegen in diesem Bündel die 167 MB Vorlagen.
#
# **Und weil `codesign --deep` sie dort mitnimmt.** Auf arm64 startet ein
# unsigniertes Binärprogramm nicht ([0003](../docs/entscheidungen/0003-keine-notarisierung.md));
# dieser Schritt läuft deshalb **vor** der Signatur in `release.sh`.
#
# ## Warum nichts in den `PATH` gelegt wird
#
# **Weil diese Anwendung keinen Installationsschritt hat und keinen bekommen
# soll.** Ein `mse` in `/usr/local/bin` verlangte Rechte, die ein Doppelklick
# nicht hat, und hinterließe eine Datei, die beim Löschen der Anwendung
# stehenbliebe. Wer die Werkzeuge im `PATH` will, legt den Verweis selbst —
# der Einzeiler steht im `README.md` und in `docs/RELEASE.md`.
#
# Das ist eine Entscheidung mit Preis: `mse` ist nach der Installation nicht
# einfach da. Der Preis gehört ausgesprochen und nicht verschwiegen.
#
# ## Was er nicht tut
#
# Er **signiert nicht**. Das tut `release.sh` für das ganze Bündel auf einmal;
# zweimal zu signieren hieße, die zweite Signatur zur ersten in Widerspruch zu
# setzen, sobald jemand die Reihenfolge ändert.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

ROT=$'\033[31m'; GRUEN=$'\033[32m'; AUS=$'\033[0m'
[ -t 1 ] || { ROT=''; GRUEN=''; AUS=''; }

ende() { printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS" >&2; exit 1; }
ok()   { printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"; }

APP="${1:-}"
[ -n "$APP" ] || ende "der Werkzeugschritt braucht den Pfad des Bündels"

MACOS="$APP/Contents/MacOS"
[ -d "$MACOS" ] || ende "$MACOS gibt es nicht — ist das ein .app-Bündel?"

# **Ausdrücklich für das Ziel aus §12**, nicht für den Rechner, der gerade
# baut. Ein Bündel für darwin/arm64 mit einem Werkzeug für etwas anderes wäre
# schwer zu bemerken und schwer zu erklären.
export GOOS=darwin GOARCH=arm64

# **Mit denselben Schaltern wie das Programm** (`bauen.sh freigabe`) —
# WART-02 (#2173).
# Gemessen am 2026-10-07: ohne sie wogen die drei zusammen 243 MB statt
# 160 MB und trugen je rund 3.300 Pfade dieses Rechners mit aus. Start,
# Ausgabe, Exitcodes, SIGTERM und der Abzug nach SIGQUIT blieben gleich; ein
# Abzug nennt seither `github.com/DavidTHeider/…` statt eines Pfades hier.
# **Und mit den Weglassschaltern** (#2393) — die Liste steht in
# `releaseschalter.sh`, `release.sh` prüft sie an jedem fertigen Werkzeug.
# shellcheck source=scripts/releaseschalter.sh
. "$WURZEL/scripts/releaseschalter.sh"
for werkzeug in mse mse-mcp mse-rest; do
  ( cd "$WURZEL" && go build -tags "$WEGLASSSCHALTER" -trimpath -ldflags "-s -w" -o "$MACOS/$werkzeug" "./cmd/$werkzeug" ) \
    || ende "$werkzeug ließ sich nicht bauen"
done

# **Sie müssen laufen, nicht nur dastehen** (§10). Ein Wächter, der eine Datei
# zählt, hielte die halbe Zusage — und die andere Hälfte wäre unsichtbar.
#
# **Und sie laufen in einem Wegwerf-Zuhause.** `mse-mcp` öffnet beim Start den
# Bestand unter `~/Library/Application Support` und hebt ihn auf seinen
# Schemastand. Hier stand der echte: gemessen am 2026-09-27 war der Bestand
# dieses Rechners an einem Vormittag über 0085, 0086 und 0087 gehoben, von
# niemandem als von Läufen dieses Schritts, jede Stufe mit einer Sicherung
# von 127 MB daneben — gebaut wird hier für die Freigabe, die Abnahme und
# jede Prüfung des Schritts, und keiner davon soll den Bestand anfassen.
PRUEFHEIM="$(mktemp -d)"
trap 'rm -rf "$PRUEFHEIM"' EXIT

HOME="$PRUEFHEIM" "$MACOS/mse" hilfe >/dev/null 2>&1 || ende "mse läuft im Bündel nicht"
HOME="$PRUEFHEIM" "$MACOS/mse-mcp" </dev/null >/dev/null 2>&1 || ende "mse-mcp läuft im Bündel nicht"

# **Der Dienst wird nicht gestartet, sondern abgewiesen.** Ihn wirklich lauschen
# zu lassen hieße, in der Freigabe einen Port zu belegen und ihn wieder
# einzufangen; mit einem unmöglichen Port läuft er an, liest seine Argumente und
# endet von selbst.
#
# **Was das nicht zeigt**: dass er bedient. Das zeigen die Flächenläufe in
# `internal/flaechen`, und zwar gegen dasselbe Programm. Hier geht es um die
# Frage, ob die Datei im Bündel ein lauffähiges Programm ist — und `126`
# (nicht ausführbar) wie `127` (nicht gefunden) sagen, dass sie es nicht ist.
HOME="$PRUEFHEIM" "$MACOS/mse-rest" -port 99999 </dev/null >/dev/null 2>&1
LAUF=$?
case "$LAUF" in
  126|127) ende "mse-rest läuft im Bündel nicht (Exit $LAUF)" ;;
esac

ok "mse, mse-mcp und mse-rest liegen in Contents/MacOS und laufen"
