#!/usr/bin/env bash
#
# Vite und der Wails-Proxy sprechen dieselbe Adresse — DEV-01
# (#2090).
#
# # Wofür es diesen Wächter gibt
#
# **`make dev-browser` gab 502.** Unter `localhost` lauschte Vite nur auf
# `[::1]`, und der Proxy von Wails wählt für `localhost` und `127.0.0.1` fest
# `tcp4` (`internal/assetserver/build_dev.go`, v3.0.0-beta.23) — er klopfte an
# `127.0.0.1`, wo niemand war.
#
# # Was er hält
#
#  1. **Vite antwortet unter der Adresse, die das Programm bekommt**, über IPv4
#     wie der Proxy.
#  2. **Ein belegter Platz wird abgewiesen** (`--strictPort`) statt still
#     ausgewichen.
#  3. **Nach dem Beenden ist der Platz frei.**
#
# # Was er nicht hält
#
# Den Weg durch das Programm selbst — dafür braucht es einen Bau. Ihn belegt
# der Lauf von `make dev-browser` im PR.

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FEHLER=0

# shellcheck source=scripts/entwickelnvite.sh
source "$WURZEL/scripts/entwickelnvite.sh"

behauptung() { # behauptung <beschreibung> <ja|nein>
  if [ "$2" = ja ]; then printf '  ✓ %s\n' "$1"; else printf '  ✗ %s\n' "$1"; FEHLER=1; fi
}

freier_platz() { python3 -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()'; }

# antwortet_ipv4 <adresse> fragt so, wie der Proxy fragt: über IPv4, auch
# wenn der Name `localhost` heißt.
antwortet_ipv4() {
  for _ in $(seq 60); do
    [ "$(curl -4 -s -m 1 -o /dev/null -w '%{http_code}' "$1/" 2>/dev/null)" = 200 ] && return 0
    kill -0 "$VITE_PID" 2>/dev/null || return 1
    sleep 0.25
  done
  return 1
}

echo "entwickelnvite"

[ -x "$WURZEL/frontend/node_modules/.bin/vite" ] || { echo "  – übersprungen: frontend/node_modules fehlt"; exit 0; }

platz="$(freier_platz)"
vite_starten "$platz"
if antwortet_ipv4 "$VITE"; then
  behauptung "Vite antwortet über IPv4 unter $VITE" ja
else
  behauptung "Vite antwortet über IPv4 unter $VITE" nein
fi
kill "$VITE_PID" 2>/dev/null
wait "$VITE_PID" 2>/dev/null
if lsof -nP -iTCP:"$platz" -sTCP:LISTEN >/dev/null 2>&1; then
  behauptung "nach dem Beenden ist der Platz frei" nein
else
  behauptung "nach dem Beenden ist der Platz frei" ja
fi

# **Ein belegter Platz** — ein fremder Lauscher auf IPv4 und IPv6.
platz="$(freier_platz)"
python3 -c '
import socket, sys, time
p = int(sys.argv[1])
a = socket.socket(socket.AF_INET); a.bind(("127.0.0.1", p)); a.listen()
b = socket.socket(socket.AF_INET6); b.bind(("::1", p)); b.listen()
time.sleep(30)' "$platz" &
fremd=$!
sleep 0.5
vite_starten "$platz"
for _ in $(seq 40); do kill -0 "$VITE_PID" 2>/dev/null || break; sleep 0.25; done
if kill -0 "$VITE_PID" 2>/dev/null; then
  behauptung "ein belegter Platz wird abgewiesen" nein
  kill "$VITE_PID" 2>/dev/null
else
  behauptung "ein belegter Platz wird abgewiesen" ja
fi
kill "$fremd" 2>/dev/null
wait "$fremd" 2>/dev/null

exit "$FEHLER"
