#!/usr/bin/env bash
#
# Der Vite-Lauf der Entwicklung — eingebunden von `scripts/entwickeln.sh`
# und von seinem Wächter `scripts/entwickelnvite.test.sh` (DEV-01, #2090).
#
# vite_starten <platz> startet Vite im Hintergrund und setzt `VITE` (die
# Adresse, die das Programm als `FRONTEND_DEVSERVER_URL` bekommt) und
# `VITE_PID`.
vite_starten() {
  local platz="$1"

  # **IPv4 ausdrücklich, auf beiden Seiten** — DEV-01 (#2090). Unter
  # `localhost` lauschte Vite (Node 24) nur auf `[::1]`; der Proxy von Wails
  # wählt für `localhost` und `127.0.0.1` fest `tcp4`
  # (`internal/assetserver/build_dev.go`, v3.0.0-beta.23) und bekam 502.
  # Gemessen am 2026-10-06: `curl localhost` 200, `curl -4 127.0.0.1` keine
  # Verbindung. Kein Schalter am System, nur die eigene Adresse.
  # shellcheck disable=SC2034 # gelesen vom Aufrufer
  VITE="http://127.0.0.1:$platz"

  # **Vite unmittelbar und nicht über `npm run`**: so ist die gemerkte
  # Kennung die von Vite selbst, und das Beenden trifft den Prozess, der den
  # Platz hält — nicht einen Vermittler, der ihn weiterreichen müsste.
  ( cd "$WURZEL/frontend" && exec node_modules/.bin/vite --host 127.0.0.1 --port "$platz" --strictPort >/dev/null 2>&1 ) &
  # shellcheck disable=SC2034 # gelesen vom Aufrufer
  VITE_PID=$!
}
