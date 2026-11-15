#!/usr/bin/env bash
#
# Prüfungen des `pre-push`-Hooks
# (#1389).
#
# ## Warum der Hook prüft und nicht rechnet
#
# **Gemessen am 2026-09-02**, nicht überlegt: `git` öffnet die SSH-Verbindung
# zu `git-receive-pack` **bevor** es den Hook ruft — 1,8 Sekunden vorher, mit
# `GIT_TRACE=1` abgelesen. Wer im Hook zehn Minuten Pipeline fährt, hält diese
# Verbindung so lange offen, und GitHub schließt sie:
#
#     HOOK: warte 600 Sekunden ab 21:33:45
#     Connection to github.com closed by remote host.
#     HOOK: fertig 21:43:45, gebe frei
#     RÜCKGABE=141
#
# 141 ist 128 + 13, also **SIGPIPE**. Der Hook hat freigegeben, git wollte das
# Paket senden, und niemand hörte mehr zu — ohne Meldung, ohne `To github.com`.
# Bei 300 Sekunden ging derselbe Weg noch durch; die Grenze liegt dazwischen.
#
# Seither **rechnet der Hook nicht mehr**: er sieht nach, ob die Pipeline für
# genau diesen Stand schon grün war, und weist sonst ab. Zehn Minuten Arbeit
# gehören vor den Push, nicht hinein.
#
# ## Was diese Datei prüft — und was nicht
#
# Sie fährt den Hook gegen ein **erfundenes** Repositorium in einem
# Wegwerfverzeichnis und schiebt ihm die Refzeilen auf die Standardeingabe,
# wie `git` es tut. Was sie **nicht** prüft: dass `git push` danach wirklich
# überträgt — dafür braucht es eine Gegenstelle, und die gehört nicht in die
# Pipeline (Vermerk 0007).

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOOK="$WURZEL/.githooks/pre-push"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0

behauptung() { # behauptung <beschreibung> <erwarteter-exit> <refzeile> [umgebung…]
  local beschreibung="$1" erwartet="$2" refzeile="$3"
  shift 3

  local ausgabe status
  ausgabe="$(cd "$PROBE" && printf '%s\n' "$refzeile" | env "$@" "$HOOK" origin git@example.invalid 2>&1)"
  status=$?

  if [ "$status" -eq "$erwartet" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$beschreibung" "$AUS"
  else
    printf '%s  ✗ %s — Exit %d, erwartet war %d%s\n' "$ROT" "$beschreibung" "$status" "$erwartet" "$AUS"
    printf '      %s\n' "$ausgabe"
    FEHLER=1
  fi
}

# nennt prüft, dass die Ausgabe ein Wort enthält — eine Abweisung ohne Grund
# schickt jemanden auf die Suche (§8).
nennt() { # nennt <beschreibung> <wort> <refzeile>
  local beschreibung="$1" wort="$2" refzeile="$3"

  local ausgabe
  ausgabe="$(cd "$PROBE" && printf '%s\n' "$refzeile" | "$HOOK" origin git@example.invalid 2>&1)"

  if printf '%s' "$ausgabe" | grep -q -- "$wort"; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$beschreibung" "$AUS"
  else
    printf '%s  ✗ %s — »%s« kommt nicht vor%s\n' "$ROT" "$beschreibung" "$wort" "$AUS"
    printf '      %s\n' "$ausgabe"
    FEHLER=1
  fi
}

ORT="$(mktemp -d)"
trap 'rm -rf "$ORT"' EXIT

PROBE="$ORT/probe"
mkdir -p "$PROBE"
(
  cd "$PROBE"
  git init -q
  git config user.email probe@example.invalid
  git config user.name Probe
  git config commit.gpgsign false
  printf 'x\n' > eine-datei
  git add -A
  git commit -qm "probe"
) >/dev/null 2>&1

SHA="$(cd "$PROBE" && git rev-parse HEAD)"
STEMPEL="$PROBE/.git/mse-pipeline"

# **Ein direkter Push auf main wird weiter abgewiesen** (§5) — das ist die
# ältere Zusage, und sie darf durch die neue nicht verlorengehen.
behauptung "main wird abgewiesen" 1 "refs/heads/main $SHA refs/heads/main 0000"
nennt "und sagt, warum" "main" "refs/heads/main $SHA refs/heads/main 0000"

# **Ohne Stempel wird abgewiesen** — und zwar sofort, nicht nach zehn Minuten.
behauptung "ohne Stempel abgewiesen" 1 "refs/heads/zweig $SHA refs/heads/zweig 0000"
nennt "und nennt make ci" "make ci" "refs/heads/zweig $SHA refs/heads/zweig 0000"

# **Mit Stempel geht es durch.**
printf '%s sauber\n' "$SHA" > "$STEMPEL"
behauptung "mit Stempel freigegeben" 0 "refs/heads/zweig $SHA refs/heads/zweig 0000"


# **Ein Stempel für einen anderen Stand zählt nicht.** Sonst gälte ein grüner
# Lauf von gestern für den Commit von heute.
printf '%s sauber\n' "0000000000000000000000000000000000000000" > "$STEMPEL"
behauptung "fremder Stempel abgewiesen" 1 "refs/heads/zweig $SHA refs/heads/zweig 0000"

# **Ein Stempel aus einem schmutzigen Baum zählt nicht.** Was gelaufen ist,
# war dann nicht dieser Commit.
printf '%s schmutzig\n' "$SHA" > "$STEMPEL"
behauptung "schmutziger Stempel abgewiesen" 1 "refs/heads/zweig $SHA refs/heads/zweig 0000"

# **Und ein sauberer Stempel bei schmutzigem Baum auch nicht** — zwischen Lauf
# und Push kann jemand etwas geändert haben.
printf '%s sauber\n' "$SHA" > "$STEMPEL"
printf 'y\n' > "$PROBE/noch-eine"
behauptung "geänderter Baum abgewiesen" 1 "refs/heads/zweig $SHA refs/heads/zweig 0000"
rm -f "$PROBE/noch-eine"

# **Das Löschen eines Zweigs braucht keinen Stempel** — es überträgt nichts,
# was geprüft werden könnte.
behauptung "das Löschen eines Zweigs geht durch" 0 \
  "(delete) 0000000000000000000000000000000000000000 refs/heads/zweig $SHA"

# **Ein Stand, der sich während des Laufs ändert, bekommt keinen Beleg.**
#
# Befund 4 des Zweitblicks über die Phasen 4.0 bis 4.9: `ci.sh` las Commit und
# Sauberkeit erst **am Ende**. Wer während der zehn Minuten committet, bekam
# einen Beleg über einen Stand, den keine Stufe gesehen hat.
#
# **Der Fall wird hier hergestellt**, nicht beschrieben: der Anfangsstand wird
# genommen, dann wird committet, dann gestempelt.
STEMPELSKRIPT="$WURZEL/scripts/stempel.sh"

(
  cd "$PROBE" || exit 1

  anfang="$("$STEMPELSKRIPT" --stand)"

  printf 'noch etwas\n' > zweite-datei
  git add -A
  git commit -qm "waehrend des laufs"

  "$STEMPELSKRIPT" "$anfang" >/dev/null 2>&1
) >/dev/null 2>&1

if [ -f "$STEMPEL" ]; then
  printf '%s  ✗ ein Commit während des Laufs bekam einen Beleg%s\n' "$ROT" "$AUS"
  FEHLER=1
else
  printf '%s  ✓ ein Commit während des Laufs bekommt keinen Beleg%s\n' "$GRUEN" "$AUS"
fi

# **Und ein alter Beleg wird dabei weggeräumt.**
#
# Ohne das bliebe der Beleg von vorhin stehen, und der Hook ließe den neuen
# Commit durch — mit einem Beleg über einen anderen Stand. Ein Beleg, der
# nicht mehr gilt, ist schlimmer als keiner.
(
  cd "$PROBE" || exit 1

  printf '%s sauber\n' "$(git rev-parse HEAD)" > "$STEMPEL"

  anfang="$("$STEMPELSKRIPT" --stand)"

  printf 'und noch etwas\n' > dritte-datei
  git add -A
  git commit -qm "wieder waehrend des laufs"

  "$STEMPELSKRIPT" "$anfang" >/dev/null 2>&1
) >/dev/null 2>&1

if [ -f "$STEMPEL" ]; then
  printf '%s  ✗ der alte Beleg blieb liegen%s\n' "$ROT" "$AUS"
  FEHLER=1
else
  printf '%s  ✓ und der alte Beleg wird weggeräumt%s\n' "$GRUEN" "$AUS"
fi

# **Ein unveränderter Stand bekommt seinen Beleg.**
#
# Die Gegenprobe: ohne sie wäre der Lauf darüber auch grün, wenn `stempel.sh`
# **nie** stempelte.
(
  cd "$PROBE" || exit 1

  anfang="$("$STEMPELSKRIPT" --stand)"
  "$STEMPELSKRIPT" "$anfang" >/dev/null 2>&1
) >/dev/null 2>&1

if [ -f "$STEMPEL" ]; then
  printf '%s  ✓ ein unveränderter Stand bekommt seinen Beleg%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ ein unveränderter Stand bekam keinen Beleg%s\n' "$ROT" "$AUS"
  FEHLER=1
fi

# **Und beide Seiten meinen dieselbe Datei.**
#
# Der Hook liest den Stempel, `scripts/ci.sh` schreibt ihn — zwei Dateien, ein
# Name. Läuft er auseinander, weist der Hook **jeden** Push ab und niemand
# sieht, warum: die Meldung sagt „nicht gelaufen", und gelaufen ist sie.
if grep -q 'mse-pipeline' "$STEMPELSKRIPT"; then
  printf '%s  ✓ stempel.sh schreibt denselben Beleg%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ stempel.sh schreibt keinen Beleg namens mse-pipeline%s\n' "$ROT" "$AUS"
  FEHLER=1
fi

# **Und `ci.sh` ruft es mit dem Anfangsstand.**
#
# Ohne das Argument stempelte es über den Stand am Ende — genau der Befund,
# den die drei Läufe darüber schließen.
if grep -q 'stempel.sh "\$STAND_ANFANG"' "$WURZEL/scripts/ci.sh"; then
  printf '%s  ✓ ci.sh gibt den Anfangsstand hinein%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ ci.sh ruft stempel.sh ohne den Anfangsstand%s\n' "$ROT" "$AUS"
  FEHLER=1
fi

# **Und sie schreibt ihn nur nach einem vollen Lauf.** Eine einzelne Stufe ist
# kein grüner Lauf; ein Stempel dafür wäre die Behauptung, es sei einer gewesen.
if grep -q 'STUFE.*=.*"all"' "$WURZEL/scripts/ci.sh"; then
  printf '%s  ✓ und nur nach einem vollen Lauf%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ ci.sh stempelt ohne Prüfung auf die Stufe%s\n' "$ROT" "$AUS"
  FEHLER=1
fi

# **Und ein signierter Tag ebenso** — Befund 4 des Nachlaufs über die Phasen
# 4.0 bis 4.9.
#
# ## Warum das ein eigener Fall ist
#
# **Bei einem annotierten Tag ist die Kennung auf der Refzeile die des
# Tag-Objekts, nicht die des Commits.** `git rev-parse v0.3.0` gibt
# `d64a565a…`, `git rev-parse v0.3.0^{commit}` dagegen `c89375f6…` — zwei
# verschiedene Kennungen für denselben Stand.
#
# Der Stempel trägt die **Commit**-Kennung. Der Hook verglich sie unverändert
# und wies damit genau den Weg ab, den `docs/RELEASE.md` vorschreibt:
# `git tag -s … && git push origin …`. Die Pipeline war grün, der Stempel
# stimmte, und der Push ging trotzdem nicht durch.
#
# **Die Zusage ändert sich dabei nicht**: ein ungestempelter Stand wird
# weiterhin abgewiesen — der Fall darunter zeigt es.
(
  cd "$PROBE"
  git tag -a -m "probe" v-probe
) >/dev/null 2>&1

TAGSHA="$(cd "$PROBE" && git rev-parse v-probe)"

behauptung "ein Tag auf einem gestempelten Commit geht durch" 0 \
  "refs/tags/v-probe $TAGSHA refs/tags/v-probe 0000"

# **Und ein Tag auf einem ungestempelten Commit nicht.** Ohne diesen Fall
# hieße die Korrektur „Tags werden durchgelassen".
(
  cd "$PROBE"
  printf 'y\n' > noch-eine
  git add -A
  git commit -qm "zweiter"
  git tag -a -m "probe" v-ungestempelt
) >/dev/null 2>&1

UNGESTEMPELT="$(cd "$PROBE" && git rev-parse v-ungestempelt)"

behauptung "ein Tag ohne Stempel wird abgewiesen" 1 \
  "refs/tags/v-ungestempelt $UNGESTEMPELT refs/tags/v-ungestempelt 0000"

# **Ein überholter Lauf hört zwischen zwei Stufen auf** — QA-07.AK-03
# (#1968).
#
# Wer während der Pipeline committet oder den Baum ändert, hat einen Lauf vor
# sich, der keinen Beleg mehr geben kann (siehe oben). **`--aktuell` sagt es,
# bevor die nächste Stufe Minuten kostet**: Exit 0 heißt weiter, 1 heißt
# überholt.
(
  cd "$PROBE" || exit 1

  anfang="$("$STEMPELSKRIPT" --stand)"
  "$STEMPELSKRIPT" --aktuell "$anfang" >/dev/null 2>&1 || exit 2

  printf 'mitten im lauf\n' > vierte-datei
  "$STEMPELSKRIPT" --aktuell "$anfang" >/dev/null 2>&1 && exit 3

  rm vierte-datei
  "$STEMPELSKRIPT" --aktuell "$anfang" >/dev/null 2>&1 || exit 4
) >/dev/null 2>&1
ueberholt=$?

if [ "$ueberholt" -eq 0 ]; then
  printf '%s  ✓ --aktuell erkennt einen überholten Stand — und einen zurückgesetzten nicht%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ --aktuell irrt (Fall %d)%s\n' "$ROT" "$ueberholt" "$AUS"
  FEHLER=1
fi

# **Und `ci.sh` fragt zwischen jeder Stufe des vollen Laufs.**
if [ "$(grep -c '; noch_aktuell; ' "$WURZEL/scripts/ci.sh")" -ge 1 ] &&
  [ "$(grep -o 'noch_aktuell' "$WURZEL/scripts/ci.sh" | wc -l)" -ge 5 ]; then
  printf '%s  ✓ ci.sh hält einen überholten Lauf zwischen den Stufen an%s\n' "$GRUEN" "$AUS"
else
  printf '%s  ✗ ci.sh fragt nicht zwischen den Stufen, ob der Lauf überholt ist%s\n' "$ROT" "$AUS"
  FEHLER=1
fi

exit "$FEHLER"
