# shellcheck shell=bash
#
# **Die macOS-Untergrenze für cgo** — eingebunden von `scripts/ci.sh`,
# `scripts/activate.sh` und `scripts/bauen.sh`, damit sie an **einer** Stelle
# steht (ENG-02, #1727).
#
# Dieselbe Zahl wie `LSMinimumSystemVersion` in `build/darwin/Info.plist`
# (§12); ein Wächter hält beide zusammen
# (`internal/werkzeuge/frontendpinnung_test.go`).
#
# # Warum es sie gibt
#
# Wails 3 übersetzt seine Fensterseite mit cgo. Ohne Angabe baut clang gegen
# den Rechner, auf dem gebaut wird, und der Linker warnt über jede Objektdatei,
# sie sei für ein neueres macOS gebaut, als gebunden wird — gemessen am
# 2026-09-23 unter macOS 27. Ein Programm, das seine Untergrenze nur im Bündel
# verspricht, hält die Zusage nicht. **15.2 seit dem 2026-10-11** — die
# Untergrenze, die ein Lauf belegt (Vermerk 0184).
#
# `-O2 -g` stehen dabei, weil `CGO_CFLAGS` und `CGO_LDFLAGS` die Vorgabe von Go
# **ersetzen** und nicht ergänzen.

export MACOSX_DEPLOYMENT_TARGET=15.2
export CGO_CFLAGS="-O2 -g -mmacosx-version-min=$MACOSX_DEPLOYMENT_TARGET"
export CGO_LDFLAGS="-O2 -g -mmacosx-version-min=$MACOSX_DEPLOYMENT_TARGET"
