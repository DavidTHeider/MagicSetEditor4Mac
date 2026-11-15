#!/usr/bin/env bash
#
# Der Beweis, den ein roter Lauf hinterlässt
# (#1455).
#
# # Wofür es diesen Wächter gibt
#
# **Am 2026-09-14 hat ein voller Lauf zwanzig Minuten gekostet und nichts
# hinterlassen.** Er war rot, und die rote Stufe war weg: der Aufrufer hatte
# die Ausgabe durch `tail -45` geschickt. Übrig blieb `✗ Pipeline rot` — eine
# Zeile ohne Fundstelle, und die einzige Auskunft war ein zweiter voller Lauf.
#
# **Zwei Dinge waren daran falsch, und nur eines davon war der Aufrufer.**
# Das Skript wusste, welche Stufe gefallen war — `fehl` lief in ihr —, und hat
# es weggeworfen. Ein Kopf, der weniger sagt, als sein Rumpf weiß (§10).
#
# # Was er hält
#
#  1. **Das Protokoll entsteht auch dann, wenn niemand hinsieht.** Geprüft
#     über einen echten Lauf, dessen Ausgabe nach `/dev/null` geht.
#  2. **`fehl` schreibt die Stufe mit** — behauptet nicht, sondern ausgeführt:
#     die Funktionen werden aus den Skripten herausgeschnitten und gefahren.
#  3. **Die Schlusszeile gibt sie aus**, und die Gegenprobe steht daneben: ohne
#     Befund steht dort keine Stufe.
#  4. **Zwei Läufe nebeneinander weist das Skript ab.** Ein zweiter kostet
#     nicht nur seine eigenen zwanzig Minuten — er treibt die Race-Tests des
#     ersten in ihre Frist.
#
# # Was er nicht hält
#
# **Dass eine Stufe das Richtige prüft.** Er prüft den Beleg, nicht die
# Prüfung. Und er fährt keine Stufe: jede von ihnen kostet Minuten, und dieser
# Wächter läuft in jedem `make ci` mit.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKRIPT="$WURZEL/scripts/ci.sh"

# **Beweis und Alleinbetrieb stehen in einer eigenen Datei** (§2); wer nur
# `ci.sh` durchsieht, findet die Zusagen nicht mehr.
BEWEIS="$WURZEL/scripts/ci-beweis.sh"
MELDER="$WURZEL/scripts/testlaeufe.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0

behauptung() { # behauptung <beschreibung> <ja|nein>
  if [ "$2" = "ja" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"
  else
    printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"
    FEHLER=1
  fi
}

ABLAGE="$(mktemp -d)"
trap 'rm -rf "$ABLAGE"' EXIT

# ------------------------------------------- 1. Das Protokoll ohne Zuschauer
#
# **Eine unbekannte Stufe kostet nichts** und geht trotzdem durch denselben
# Kopf: die Umleitung steht vor der Fallunterscheidung. Was danach in der
# Datei steht, stand nie auf einem Bildschirm.

protokoll="$ABLAGE/protokoll.log"

# **`env -u MSE_CI_IM_PROTOKOLL`, und das ist der Fund dieses Wächters an
# sich selbst.** Läuft er in `make ci` mit, hat der Elternlauf die Marke
# schon gesetzt und **vererbt sie**: das Kind hielte sich für den inneren
# Lauf, überspränge die Umleitung und schriebe kein Protokoll. Von Hand war
# er grün, im Lauf rot — gefunden am 2026-09-15, vom ersten `make ci` über
# ihn selbst.
env -u MSE_CI_IM_PROTOKOLL \
  MSE_CI_PROTOKOLL="$protokoll" MSE_CI_SPERRE="$ABLAGE/sperre" \
  "$SKRIPT" diese-stufe-gibt-es-nicht >/dev/null 2>&1

if [ -s "$protokoll" ] &&
  grep -q "diese-stufe-gibt-es-nicht" "$protokoll"; then
  behauptung "das Protokoll entsteht, auch wenn die Ausgabe verworfen wird" "ja"
else
  behauptung "das Protokoll entsteht, auch wenn die Ausgabe verworfen wird" "nein"
  printf '%s  → ohne es ist ein roter Lauf von zwanzig Minuten beweislos%s\n' \
    "$ROT" "$AUS"
fi

# ------------------------------------------------ 2. `fehl` merkt sich die Stufe
#
# **Herausgeschnitten und gefahren, nicht nachgebaut** — dieselbe Gestalt wie
# in `audit.test.sh`. Zwei Fassungen einer Regel laufen beim ersten Sonderfall
# auseinander, und dann prüft dieser Lauf eine, die es nicht gibt.

stufen_aus_lauf() { # stufen_aus_lauf → die gesammelten Stufen
  bash -c "
    GRUEN=''; ROT=''; GELB=''; BLAU=''; AUS=''
    FEHLER=0; UEBERSPRUNGEN=0
    $(sed -n '/^# STUFE_JETZT ist die Stufe/,/^kopf()/p' "$BEWEIS")
    $(sed -n '/^fehl()  {/,/^}/p' "$MELDER")
    kopf Format   >/dev/null
    kopf Tests    >/dev/null
    fehl 'etwas'  >/dev/null
    fehl 'noch etwas' >/dev/null
    printf '%s' \"\$ROTE_STUFEN\"
  "
}

gesammelt="$(stufen_aus_lauf)"

case "$gesammelt" in
  *Tests*) behauptung "fehl schreibt die laufende Stufe mit" "ja" ;;
  *)       behauptung "fehl schreibt die laufende Stufe mit" "nein"
           printf '%s  → gesammelt wurde: %q%s\n' "$ROT" "$gesammelt" "$AUS" ;;
esac

# **Die Gegenprobe**: die Stufe, in der nichts gefallen ist, steht nicht da.
# Ohne sie sagte die Zeile darüber nur, dass irgendetwas gesammelt wird.
case "$gesammelt" in
  *Format*) behauptung "eine grüne Stufe steht nicht in der Liste" "nein" ;;
  *)        behauptung "eine grüne Stufe steht nicht in der Liste" "ja" ;;
esac

# ---------------------------------------- 2a. Jede Stufe nennt ihre Dauer
#
# **QA-07.AK-01** (#1738):
# `kopf` schließt die vorige Stufe ab, und jede steht **einmal** in der Liste.

dauern_aus_lauf() { # dauern_aus_lauf → die gesammelten Dauern
  bash -c "
    GRUEN=''; ROT=''; GELB=''; BLAU=''; AUS=''
    $(sed -n '/^# STUFE_JETZT ist die Stufe/,/^kopf()/p' "$BEWEIS")
    kopf Format >/dev/null
    kopf Tests  >/dev/null
    stufe_beenden
    STUFE_JETZT=''
    stufe_beenden
    printf '%s' \"\$DAUERN\"
  "
}

case "$(dauern_aus_lauf)" in
  'Format 0 s · Tests 0 s') behauptung "jede Stufe steht einmal mit ihrer Dauer da" "ja" ;;
  *)                        behauptung "jede Stufe steht einmal mit ihrer Dauer da" "nein"
                            printf '%s  → gesammelt wurde: %q%s\n' "$ROT" "$(dauern_aus_lauf)" "$AUS" ;;
esac

# ------------------------------------------------- 3. Die Schlusszeile nennt sie

if grep -q 'Pipeline rot — rot war' "$BEWEIS" &&
  grep -q 'Das vollständige Protokoll steht in' "$BEWEIS"; then
  behauptung "die Schlusszeile nennt die Stufe und das Protokoll" "ja"
else
  behauptung "die Schlusszeile nennt die Stufe und das Protokoll" "nein"
fi

# ------------------------------------ 3a. Ein Lauf mit Lücke stempelt nicht
#
# **QA-01.AK-02** (#1733):
# „Fehlende Pflichtfixtures … können keinen gültigen grünen Pflichtbeleg
# erzeugen." Der Stempel **ist** der Pflichtbeleg dieses Laufs — der Hook
# lässt nach ihm pushen. Bis hierhin stempelte ein Lauf, dem `node_modules`
# fehlte oder dessen Registry nicht antwortete, und sagte „grün".
#
# **Herausgeschnitten und gefahren**, wie oben: `schlusszeile` aus dem Skript,
# `stempeln` und `reichweite` als Spione.

schluss_aus_lauf() { # schluss_aus_lauf <übersprungen> [erlaubt] → Ausgabe
  if [ -n "${2:-}" ]; then
    set -- "$1" MSE_CI_UEBERSPRUNGEN_ERLAUBT=1
  else
    set -- "$1" MSE_CI_NICHTS=1
  fi

  env -u MSE_CI_UEBERSPRUNGEN_ERLAUBT "$2" bash -c "
    GRUEN=''; ROT=''; GELB=''; BLAU=''; AUS=''
    FEHLER=0; UEBERSPRUNGEN=$1; ROTE_STUFEN=''
    PROTOKOLL=/protokoll; WURZEL=/
    reichweite() { :; }
    stufe_beenden() { :; }
    stempeln() { printf 'GESTEMPELT\n'; }
    $(sed -n '/^schlusszeile() {/,/^}/p' "$BEWEIS")
    schlusszeile
    printf 'FEHLER=%s\n' \"\$FEHLER\"
  "
}

luecke="$(schluss_aus_lauf 2)"

case "$luecke" in
  *GESTEMPELT*) behauptung "ein Lauf mit übersprungenen Stufen stempelt nicht" "nein" ;;
  *)            behauptung "ein Lauf mit übersprungenen Stufen stempelt nicht" "ja" ;;
esac

case "$luecke" in
  *unvollständig*FEHLER=1*) behauptung "er sagt »unvollständig« und endet mit 1" "ja" ;;
  *)                        behauptung "er sagt »unvollständig« und endet mit 1" "nein"
                            printf '%s  → gesagt wurde: %s%s\n' "$ROT" "$luecke" "$AUS" ;;
esac

# **Wer es bewusst will, bekommt den Stempel** — und die Zeile sagt, dass
# etwas fehlt.
case "$(schluss_aus_lauf 2 erlaubt)" in
  *übersprungen*GESTEMPELT*FEHLER=0*) behauptung "mit ausdrücklicher Erlaubnis wird gestempelt" "ja" ;;
  *)                                  behauptung "mit ausdrücklicher Erlaubnis wird gestempelt" "nein" ;;
esac

# **Die Gegenprobe**: ein vollständiger Lauf stempelt weiter. Ohne sie sagten
# die Zeilen darüber auch dann „ja", wenn gar nicht mehr gestempelt würde.
case "$(schluss_aus_lauf 0)" in
  *GESTEMPELT*FEHLER=0*) behauptung "ein vollständiger Lauf stempelt weiter" "ja" ;;
  *)                     behauptung "ein vollständiger Lauf stempelt weiter" "nein" ;;
esac

# ------------------------------------------- 3a. Ein Lauf über dem Ziel sagt es
#
# **PERF03** (#2353),
# Befund A-28: es gab kein Ziel für die Dauer der Pipeline. Jetzt sind es
# zwanzig Minuten auf diesem Rechner, und wer darüber liegt, hört es — **als
# Warnung**: ein langsamer Lauf ist kein roter, und der Stempel bleibt.

dauer_aus_lauf() { # dauer_aus_lauf <sekunden> → Ausgabe
  bash -c "
    GRUEN=''; ROT=''; GELB=''; BLAU=''; AUS=''
    FEHLER=0; UEBERSPRUNGEN=0; ROTE_STUFEN=''
    PROTOKOLL=/protokoll; WURZEL=/
    reichweite() { :; }
    stufe_beenden() { :; }
    stempeln() { printf 'GESTEMPELT\n'; }
    $(sed -n '/^schlusszeile() {/,/^}/p' "$BEWEIS")
    SECONDS=$1
    schlusszeile
    printf 'FEHLER=%s\n' \"\$FEHLER\"
  "
}

case "$(dauer_aus_lauf 1500)" in
  *"über dem Ziel"*GESTEMPELT*FEHLER=0*) behauptung "ein Lauf über zwanzig Minuten warnt und stempelt" "ja" ;;
  *)                                     behauptung "ein Lauf über zwanzig Minuten warnt und stempelt" "nein" ;;
esac

case "$(dauer_aus_lauf 600)" in
  *"über dem Ziel"*) behauptung "ein Lauf unter dem Ziel warnt nicht" "nein" ;;
  *GESTEMPELT*)      behauptung "ein Lauf unter dem Ziel warnt nicht" "ja" ;;
  *)                 behauptung "ein Lauf unter dem Ziel warnt nicht" "nein" ;;
esac

# ------------------------------------------------------- 4. Nur ein Lauf zugleich
#
# **Mit einer Sperre, die einem lebenden Prozess gehört.** Die eigene
# Prozesskennung tut es: sie lebt, solange dieser Wächter läuft.
#
# **Ohne die Erlaubnis des Aufrufers** — wer die Pipeline mit
# `MSE_CI_TROTZDEM=1` startete, vererbte sie bis zum 2026-09-30 an diesen
# Lauf, und die Sperre, die er prüfen soll, war aufgehoben.

sperre="$ABLAGE/sperre-belegt"
mkdir -p "$sperre"
printf '%s' "$$" > "$sperre/pid"

ausgabe="$(env -u MSE_CI_IM_PROTOKOLL -u MSE_CI_TROTZDEM \
  MSE_CI_PROTOKOLL="$ABLAGE/zweit.log" MSE_CI_SPERRE="$sperre" \
  "$SKRIPT" format 2>&1)"
rueckgabe=$?

if [ "$rueckgabe" -ne 0 ] && printf '%s' "$ausgabe" | grep -q "läuft schon"; then
  behauptung "ein zweiter Lauf wird abgewiesen" "ja"
else
  behauptung "ein zweiter Lauf wird abgewiesen" "nein"
  printf '%s  → Rückgabe %s, Ausgabe: %s%s\n' "$ROT" "$rueckgabe" "$ausgabe" "$AUS"
fi

# ------------------------- 4a. Ein abgewiesener Lauf lässt den ersten stehen
#
# **T04** (#2098):
# bis zum 2026-10-06 öffnete der zweite Lauf das Protokoll mit `tee`, **bevor**
# er an der Sperre abgewiesen wurde — die Datei des ersten war danach leer bis
# auf die Ablehnung. Der Fall darüber sah es nicht: er gab dem zweiten Lauf
# eine eigene Datei.

gemeinsam="$ABLAGE/gemeinsam.log"
printf 'erster Lauf, Zeile 1\nerster Lauf, Zeile 2\n' > "$gemeinsam"
vorher="$(shasum < "$gemeinsam")"

env -u MSE_CI_IM_PROTOKOLL -u MSE_CI_TROTZDEM \
  MSE_CI_PROTOKOLL="$gemeinsam" MSE_CI_SPERRE="$sperre" \
  "$SKRIPT" format >/dev/null 2>&1
rueckgabe=$?

if [ "$rueckgabe" -eq 3 ] && [ "$(shasum < "$gemeinsam")" = "$vorher" ]; then
  behauptung "ein abgewiesener Lauf verändert das Protokoll des ersten nicht" "ja"
else
  behauptung "ein abgewiesener Lauf verändert das Protokoll des ersten nicht" "nein"
  printf '%s  → Rückgabe %s, danach: %s%s\n' "$ROT" "$rueckgabe" "$(head -c 200 "$gemeinsam")" "$AUS"
fi

# **Und die Ablehnung hat ihren eigenen Beleg** — wer nicht hinsah, findet
# sonst nur einen Lauf, der nie stattgefunden hat.
abgewiesen="${gemeinsam%.log}-abgewiesen.log"
if grep -q "läuft schon" "$abgewiesen" 2>/dev/null &&
  grep -q "PID $$" "$abgewiesen" 2>/dev/null; then
  behauptung "die Ablehnung steht in einem eigenen Beleg" "ja"
else
  behauptung "die Ablehnung steht in einem eigenen Beleg" "nein"
fi

# **Die Sperre des ersten bleibt seine** — auch die Prozesskennung darin.
if [ -d "$sperre" ] && [ "$(cat "$sperre/pid" 2>/dev/null)" = "$$" ]; then
  behauptung "die Sperre des ersten bleibt unberührt" "ja"
else
  behauptung "die Sperre des ersten bleibt unberührt" "nein"
fi

# ------------------------------- 4b. Eine tote Sperre hält niemanden auf
#
# **Ein Lauf, den jemand abgebrochen hat**, hinterlässt sein Verzeichnis. Der
# nächste räumt es weg, **sagt es im eigenen Protokoll** und gibt die Sperre
# am Ende wieder frei.
tot="$ABLAGE/sperre-tot"
mkdir -p "$tot"
sh -c 'exit 0' &
totpid=$!
wait "$totpid"
printf '%s' "$totpid" > "$tot/pid"

env -u MSE_CI_IM_PROTOKOLL -u MSE_CI_TROTZDEM \
  MSE_CI_PROTOKOLL="$ABLAGE/nach-tot.log" MSE_CI_SPERRE="$tot" \
  "$SKRIPT" diese-stufe-gibt-es-nicht >/dev/null 2>&1

if grep -q "weggeräumt" "$ABLAGE/nach-tot.log" 2>/dev/null &&
  grep -q "Unbekannte Stufe" "$ABLAGE/nach-tot.log" 2>/dev/null &&
  [ ! -e "$tot" ]; then
  behauptung "eine tote Sperre wird geräumt, protokolliert und freigegeben" "ja"
else
  behauptung "eine tote Sperre wird geräumt, protokolliert und freigegeben" "nein"
fi

# **Und die Sperre lässt sich aufheben** — sonst wäre sie eine Mauer und keine
# Vorgabe, und jemand behülfe sich beim ersten Anlass mit `--no-verify`.
if grep -q 'MSE_CI_TROTZDEM' "$BEWEIS"; then
  behauptung "die Sperre lässt sich begründet aufheben" "ja"
else
  behauptung "die Sperre lässt sich begründet aufheben" "nein"
fi

exit "$FEHLER"
