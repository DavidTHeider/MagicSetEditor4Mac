#!/usr/bin/env bash
#
# Beweis und Alleinbetrieb — der Kopf jedes Pipelinelaufs
# (#1455).
#
# **Eine eigene Datei**, weil `ci.sh` mit ihr über die Zeilengrenze aus §2
# ginge. Der Schnitt folgt der Frage: dort steht, **was geprüft wird** — hier,
# **dass der Lauf allein läuft und einen Beweis hinterlässt**.
#
# Wird von `scripts/ci.sh` eingebunden, unmittelbar nach dem Wechsel in die
# Wurzel und **vor** allem anderen: die Umleitung muss vor der ersten Ausgabe
# stehen, sonst fehlt der Anfang im Protokoll.

# ------------------------------------------------- Beweis und Alleinbetrieb
#
# **Ein Lauf dieser Pipeline dauert über zwanzig Minuten** (gemessen am
# 2026-09-14: 1.278 s). Zwei Dinge dürfen deshalb nicht davon abhängen, dass
# der Aufrufer aufpasst.
#
# ## Das Protokoll schreibt der Lauf selbst
#
# Am 2026-09-14 ist ein voller Lauf rot geworden, und die rote Stufe war
# **weg**: der Aufrufer hatte die Ausgabe durch `tail -45` geschickt. Übrig
# blieb `✗ Pipeline rot` ohne Fundstelle — zwanzig Minuten für eine Zeile,
# die nichts sagt.
#
# **Deshalb schreibt der Lauf sein Protokoll, bevor die Ausgabe irgendwohin
# geht.** Wer sie abschneidet, schneidet nur seine Sicht ab.
#
# **Über einen Kindprozess und nicht über `exec > >(tee …)`**: bei der
# Prozessersetzung endet das Skript, bevor `tee` geleert hat, und die letzten
# Zeilen — also gerade die Schlusszeile — fehlen mal und stehen mal da. Eine
# Rohrleitung wartet auf beide Seiten.
#
# ## Und er läuft allein
#
# **Zwei Läufe nebeneinander sind nicht zweimal derselbe Lauf.** Der
# Race-Detektor teilt sich die Maschine, und `internal/app` und
# `internal/storage` laufen dann in ihre Frist statt durch — gemessen in der
# Nacht zum 2026-09-13. Der zweite Lauf kostet nicht nur seine eigenen zwanzig
# Minuten, er verdirbt die des ersten.
#
# `MSE_CI_TROTZDEM=1` hebt die Sperre auf; wer sie aufhebt, weiß warum.

PROTOKOLL="${MSE_CI_PROTOKOLL:-$WURZEL/.ci-protokoll.log}"
SPERRE="${MSE_CI_SPERRE:-$WURZEL/.ci-sperre}"

# **Die Sperre nimmt der äußere Lauf, bevor er das Protokoll öffnet** — T04
# (#2098).
# Bis zum 2026-10-06 stand sie im inneren: `tee` hatte die Datei des ersten
# Laufs schon geleert, wenn der zweite abgewiesen wurde, und vom ersten blieb
# die Ablehnung des zweiten. **Die Ablehnung hat jetzt ihren eigenen Beleg**
# neben dem Protokoll, und wer die Sperre mit `MSE_CI_TROTZDEM=1` aufhebt,
# schreibt in eine eigene Datei und lässt die fremde Sperre stehen.
#
# **Eine Sperre aus einem Verzeichnis**, weil `mkdir` unteilbar ist: zwei
# Prozesse, die es gleichzeitig versuchen, bekommen nicht beide ein Ja. **Und
# eine tote Sperre hält niemanden auf** — ein Lauf, den jemand mit Strg-C
# beendet, ließe sonst ein Verzeichnis zurück.
#
# **Sie gilt je Arbeitsbaum, nicht je Maschine**: `.ci-sperre` liegt in der
# Wurzel. Zwei Arbeitsbäume fahren zwei Läufe zugleich, und dann gilt für die
# Race-Tests, was oben steht — die Sperre verhindert es nicht. **Und wer den
# Lauf über die Kennung in `.ci-sperre/pid` beendet, beendet den äußeren**:
# die Sperre geht, der innere läuft zu Ende. Strg-C trifft beide.
if [ -z "${MSE_CI_IM_PROTOKOLL:-}" ]; then
  hinweis=""
  eigene_sperre=1

  if ! mkdir "$SPERRE" 2>/dev/null; then
    alter="$(cat "$SPERRE/pid" 2>/dev/null || echo "")"

    if [ -n "$alter" ] && kill -0 "$alter" 2>/dev/null; then
      if [ -z "${MSE_CI_TROTZDEM:-}" ]; then
        {
          printf 'Es läuft schon eine Pipeline (PID %s).\n' "$alter"
          printf 'Zwei Läufe nebeneinander treiben die Race-Tests in ihre Frist.\n'
          printf 'Wenn das trotzdem gewollt ist: MSE_CI_TROTZDEM=1 davorsetzen.\n'
        } >&2
        printf '%s abgewiesen: %s — Es läuft schon eine Pipeline (PID %s).\n' \
          "$(date '+%Y-%m-%dT%H:%M:%S')" "${*:-all}" "$alter" \
          >> "${PROTOKOLL%.log}-abgewiesen.log"

        exit 3
      fi

      eigene_sperre=0
      PROTOKOLL="${PROTOKOLL%.log}-$$.log"
      hinweis="Neben PID $alter, ausdrücklich erlaubt — dieses Protokoll steht in ${PROTOKOLL#"$WURZEL"/}."
    else
      hinweis='Eine Sperre von einem Lauf, den es nicht mehr gibt — weggeräumt.'
      rm -rf "$SPERRE"
      if ! mkdir "$SPERRE" 2>/dev/null; then
        printf 'Die Sperre hat gerade ein anderer Lauf genommen.\n' >&2
        exit 3
      fi
    fi
  fi

  if [ "$eigene_sperre" -eq 1 ]; then
    printf '%s' "$$" > "$SPERRE/pid" 2>/dev/null || true
    trap 'rm -rf "$SPERRE"' EXIT
  fi

  export MSE_CI_IM_PROTOKOLL=1
  export MSE_CI_PROTOKOLL="$PROTOKOLL"
  export MSE_CI_SPERRE="$SPERRE"

  : > "$PROTOKOLL"
  if [ -n "$hinweis" ]; then printf '%s\n' "$hinweis" | tee -a "$PROTOKOLL" >&2; fi

  set +e
  "$0" "$@" 2>&1 | tee -a "$PROTOKOLL"
  rueckgabe="${PIPESTATUS[0]}"
  set -e

  exit "$rueckgabe"
fi

# Projektlokale Werkzeuge verfügbar machen, ohne die Shell des Aufrufers zu ändern.

# STUFE_JETZT ist die Stufe, die gerade läuft — `fehl` schreibt sie mit.
STUFE_JETZT=""
ROTE_STUFEN=""

# **Und wie lange sie dauert** — QA-07.AK-01
# (#1738).
# Bis zum 2026-10-01 sagte ein Lauf nur, **dass** er fertig war: zwei grüne
# Läufe derselben Nacht brauchten 22 und 32 Minuten, und das Protokoll
# nannte keine Stufe, an der der Unterschied lag.
STUFE_BEGINN=$SECONDS
DAUERN=""

# stufe_beenden schreibt die Dauer der laufenden Stufe fort — einmal.
stufe_beenden() { [ -n "$STUFE_JETZT" ] || return 0; DAUERN="${DAUERN}${DAUERN:+ · }${STUFE_JETZT} $((SECONDS - STUFE_BEGINN)) s"; }

kopf()  { stufe_beenden; STUFE_JETZT="$1"; STUFE_BEGINN=$SECONDS; printf '\n%s── %s %s\n' "$BLAU" "$1" "$AUS"; }

# schlusszeile sagt, wie der Lauf ausgegangen ist — und **wo es stand**.
#
# **Hier und nicht in `ci.sh`**, weil sie dieselbe Frage beantwortet wie die
# Umleitung darüber: was bleibt von einem Lauf übrig, wenn niemand hinsah.
schlusszeile() {
  stufe_beenden
  printf '\n  Dauer je Stufe: %s · zusammen %d s\n' "${DAUERN:-—}" "$SECONDS"

  # **Und ob der Lauf sein Ziel hielt** — PERF03
  # (#2353),
  # Befund A-28: zwanzig Minuten auf diesem Rechner (`MSE_CI_ZIEL` in
  # Sekunden). **Eine Warnung und kein Rot**: ein langsamer Lauf ist kein
  # falscher, und wer ihn weiß, sucht die Stufe in der Zeile darüber.
  if [ "$SECONDS" -gt "${MSE_CI_ZIEL:-1200}" ]; then
    printf '%s  ◐ über dem Ziel von %d min — dieser Lauf brauchte %d min%s\n' \
      "$GELB" $(( ${MSE_CI_ZIEL:-1200} / 60 )) $(( SECONDS / 60 )) "$AUS"
  fi
  printf '\n'
  if [ "$FEHLER" -eq 0 ]; then
    if [ "$UEBERSPRUNGEN" -gt 0 ]; then
      # Als `if`, nicht als `[ … ] && …`: die Kurzform gäbe bei mehr als einer
      # übersprungenen Prüfung 1 zurück, und das ist hier — außerhalb jeder
      # Funktion, unter `set -e` — der stille Abbruch kurz vor der Schlusszeile.
      if [ "$UEBERSPRUNGEN" -eq 1 ]; then wort="Prüfung"; else wort="Prüfungen"; fi

      # **Ein Lauf mit Lücke ist kein Pflichtbeleg** — QA-01.AK-02
      # (#1733).
      # Der Stempel ist der Beleg, nach dem der Hook pushen lässt. Bis zum
      # 2026-09-22 stempelte ein Lauf, dem `node_modules` fehlte oder dessen
      # Registry nicht antwortete, und sagte „grün". **Wer bewusst so pushen
      # will, sagt es** — dann steht es in der Schlusszeile.
      if [ -z "${MSE_CI_UEBERSPRUNGEN_ERLAUBT:-}" ]; then
        printf '%s◐ Pipeline unvollständig — %d %s übersprungen, weil der Gegenstand fehlt; kein Beleg für den Push%s\n' \
          "$GELB" "$UEBERSPRUNGEN" "$wort" "$AUS"
        printf '  Wer bewusst so pushen will: MSE_CI_UEBERSPRUNGEN_ERLAUBT=1 make ci\n'
        printf '  Das vollständige Protokoll steht in %s\n' "${PROTOKOLL#"$WURZEL"/}"
        FEHLER=1

        return 0
      fi

      printf '%s✓ Pipeline grün — %d %s übersprungen, weil der Gegenstand fehlt (ausdrücklich erlaubt)%s\n' \
        "$GELB" "$UEBERSPRUNGEN" "$wort" "$AUS"
    else
      printf '%s✓ Pipeline grün%s\n' "$GRUEN" "$AUS"
    fi

    # **Nur nach einem grünen Lauf.** Bei einem roten steht die Reichweite im
    # Weg: dort will jemand wissen, was kaputt ist, und nicht, was ohnehin
    # ungeprüft bleibt.
    reichweite
    stempeln
  else
    # **Die rote Stufe steht in der Schlusszeile.**
    #
    # Ein bloßes `✗ Pipeline rot` sagte weniger, als der Rumpf weiß (§10):
    # welche Stufe gefallen ist, weiß dieses Skript in dem Augenblick, in dem
    # `fehl` läuft.
    if [ -n "${ROTE_STUFEN:-}" ]; then
      printf '%s✗ Pipeline rot — rot war: %s%s\n' \
        "$ROT" "$(printf '%s' "$ROTE_STUFEN" | sed 's/ *$//; s/ /, /g')" "$AUS"
    else
      printf '%s✗ Pipeline rot%s\n' "$ROT" "$AUS"
    fi

    printf '  Das vollständige Protokoll steht in %s\n' "${PROTOKOLL#"$WURZEL"/}"
  fi
}
