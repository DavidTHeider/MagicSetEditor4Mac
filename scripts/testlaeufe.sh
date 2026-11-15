#!/usr/bin/env bash
#
# **Welche Pakete mit und welche ohne Race-Detektor laufen** — und wie ihre
# zwei Abdeckungsprofile wieder eines werden.
#
# Aus `ci.sh` herausgelöst, als es über 400 Zeilen ging (§2). Der Schnitt folgt
# der Frage: dort stehen die **Stufen**, hier die **Aufteilung** eines Laufs.
#
# **Wird von `ci.sh` eingelesen**, nicht selbst aufgerufen.

# rennpakete sind alle außer denen aus OHNE_RENNEN und GETEILT — die
# geteilten laufen mit Detektor, aber in eigenen Prozessen
# (`scripts/teilrennen.sh`).
#
# **Über `go list` und nicht über eine zweite Aufzählung**: eine Liste, die
# jemand pflegen muss, vergisst das nächste Paket — und dann liefe es in
# **keinem** der beiden Läufe (§10).
rennpakete() {
  go list $GO_PAKETE | grep -v -F -f <(go list $OHNE_RENNEN) | grep -v -x -F -f <(go list $GETEILT)
}

# ruhigepakete sind die aus OHNE_RENNEN.
ruhigepakete() {
  go list $OHNE_RENNEN
}

# profile_vereinen legt die drei Abdeckungsprofile zu einem zusammen.
#
# **Drei Läufe, ein Profil**: die Schwellen gelten über das Ganze, und drei
# Dateien nebeneinander prüfte niemand zusammen.
profile_vereinen() {
  [ -f .abdeckung-race.out ] || return 1
  [ -f .abdeckung-ruhig.out ] || return 1
  [ -f .abdeckung-geteilt.out ] || return 1

  {
    head -n 1 .abdeckung-race.out
    tail -n +2 .abdeckung-race.out
    tail -n +2 .abdeckung-ruhig.out
    tail -n +2 .abdeckung-geteilt.out
  } > .abdeckung.out

  rm -f .abdeckung-race.out .abdeckung-ruhig.out .abdeckung-geteilt.out
}
ok()    { printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"; }
warn()  { printf '%s  ~ %s%s\n' "$GELB" "$1" "$AUS"; }
# fehl meldet einen Befund — **und wo er aufgetreten ist**.
#
# **Die Stufe steht in `STUFE_JETZT`**, die `kopf` setzt. Ohne sie sagte die
# Schlusszeile nur „rot", und wer die Ausgabe abgeschnitten hat, fährt den
# ganzen Lauf noch einmal, um zu erfahren, wo. Am 2026-09-14 ist genau das
# passiert — zwanzig Minuten für eine Zeile ohne Fundstelle.
#
# **Jede Stufe nur einmal**, sonst stünde bei zehn Befunden zehnmal `test`.
fehl()  {
  printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"
  FEHLER=1

  case " ${ROTE_STUFEN:-} " in
    *" ${STUFE_JETZT:-} "*) ;;
    *) ROTE_STUFEN="${ROTE_STUFEN:-}${STUFE_JETZT:-} " ;;
  esac
}
weg()   { printf '%s  – übersprungen: %s%s\n' "$GELB" "$1" "$AUS"; UEBERSPRUNGEN=$((UEBERSPRUNGEN + 1)); }

hat_go()       { [ -f go.mod ]; }
hat_frontend() { [ -f frontend/package.json ] && [ -d frontend/node_modules ]; }

werkzeug_da() { command -v "$1" >/dev/null 2>&1; }

# shellcheck source=scripts/ci-pinnung.sh
source "$(dirname "$0")/ci-pinnung.sh"

# --------------------------------------------------------------- Format
