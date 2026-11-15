# shellcheck shell=bash
#
# **Welche Bauschalter im Programm stecken** — ENG-02.AK-05 (#1727),
# Vermerk 0132.
#
# Wails 3 öffnet mit Bauschaltern Zugänge: ohne `production` die
# Entwicklerwerkzeuge der Webansicht, mit `mcp` einen MCP-Server von Wails
# selbst, mit `server` einen HTTP-Server statt eines Fensters. **Ein
# Freigabebau trägt `production` und keinen anderen Zugang** — dazu die
# Weglassschalter unten, die nichts öffnen, sondern etwas weglassen.
#
# # Warum am fertigen Programm und nicht am Aufruf
#
# Der Aufruf steht in `scripts/bauen.sh`, und ein Wächter über ihn sähe den
# Text, nicht das Programm. Go schreibt die Schalter ins Programm, und
# `go version -m` liest sie wieder (`build -tags=…`).
#
# # Und was ein Freigabebau weglässt
#
# **Die Weglassschalter** (`WEGLASSSCHALTER`, #2393) nehmen aus fremden
# Modulen, was dieses Programm nicht bedient — und dessen Bearbeitungsform
# jede Freigabe sonst beilegen müsste (GPL-2.0 §3, Vermerk 0190):
#
# - `ts_omit_webclient` — Tailscales vorgebaute Weboberfläche samt der
#   Schrift Inter (`github.com/tailscale/web-client-prebuilt`). Ihr Stand
#   steht im vorgebauten Modul nicht, und diese Anwendung zeigt sie nie.
# - `ts_omit_debugeventbus` — die Diagnoseseite des Ereignisbusses von
#   tailscale.com mit htmx, minifiziert (`util/eventbus/assets`); gefunden am
#   2026-10-16 beim Zählen der Einbettungen, in der Inventur vom 2026-10-13
#   stand sie nicht.
#
# Beide Schalter kennt die gebundene Fassung von tailscale.com
# (`ipn/ipnlocal/web_client_stub.go`, `util/eventbus/debughttp_off.go`).
# **Das Programm trägt sie zu `production`, die drei Werkzeuge allein** —
# sie binden Wails nicht ein. Gebaut wird in `bauen.sh` und `werkzeuge.sh`
# mit dieser Liste; `einbettungen.test.sh` hält, dass mit ihr keine
# Einbettung ohne Zuordnung im Programm steckt.
#
# Geprüft in `scripts/releaseschalter.test.sh`.

WEGLASSSCHALTER="ts_omit_webclient,ts_omit_debugeventbus"

# schalter_aus liest die Bauschalter aus der Ausgabe von `go version -m` —
# kommagetrennt, leer ohne Schalter.
schalter_aus() {
  sed -n 's/^[[:space:]]*build[[:space:]]*-tags=//p' | head -1
}

# schalter_befund liest die Schalter von der Standardeingabe und nennt, was
# daran falsch ist — nichts, wenn es ein Freigabebau ist. Die Rolle ist
# `programm` (Vorgabe: `production` und die Weglassschalter) oder `werkzeug`
# (die Weglassschalter allein).
schalter_befund() { # schalter_befund [programm|werkzeug]
  local rolle="${1:-programm}" schalter verlangt name weitere
  schalter="$(cat)"

  case "$rolle" in
    programm) verlangt="production,$WEGLASSSCHALTER" ;;
    werkzeug) verlangt="$WEGLASSSCHALTER" ;;
    *) printf 'unbekannte Rolle %s — programm oder werkzeug\n' "$rolle"; return ;;
  esac

  if [ "$rolle" = programm ] && ! printf '%s\n' "$schalter" | tr ',' '\n' | grep -qx production; then
    printf 'ohne production — die Entwicklerwerkzeuge der Webansicht wären offen\n'
    return
  fi

  for name in ${WEGLASSSCHALTER//,/ }; do
    if ! printf '%s\n' "$schalter" | tr ',' '\n' | grep -qx "$name"; then
      printf 'ohne %s — eingebunden bliebe, wessen Bearbeitungsform nicht beiliegt (#2393)\n' "$name"
      return
    fi
  done

  weitere="$(printf '%s\n' "$schalter" | tr ',' '\n' | grep -vxF -f <(printf '%s\n' "$verlangt" | tr ',' '\n') | grep -v '^$' | paste -sd, - || true)"
  if [ -n "$weitere" ]; then
    printf 'mit %s — ein Zugang, den kein Freigabebau haben darf\n' "$weitere"
  fi
}
