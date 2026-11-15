#!/usr/bin/env bash
#
# **Jedes Programm im Bündel trägt die Hardened Runtime** — SEC03
# (#2356),
# Vermerk [0195](../docs/entscheidungen/0195-hardened-runtime-ohne-entitlements.md).
#
# **Überall im Bündel, nicht nur in `Contents/MacOS`** (Nachlauf über Phase 5
# von 0.9.0, Befund 4): verschachtelter Code darf auch unter `Helpers`,
# `XPCServices` oder `Frameworks` liegen, und ein Programm dort zählte sonst
# nicht. Gezählt wird, was `file` ein **ausführbares Mach-O** nennt;
# Bibliotheken tragen die Flagge nicht.
#
# Eigene Datei, damit `releasehaertung.test.sh` sie an einem nachgestellten
# Bündel prüfen kann, ohne eine Freigabe zu bauen.

# programme_gehaertet <app> — gibt die Zahl der Programme aus; endet mit 1 und
# nennt das erste Programm ohne Laufzeit, wenn eines fehlt.
#
# **Erst auffangen, dann vergleichen**, nicht `codesign … | grep -q`: unter
# `set -o pipefail` endet `codesign` sonst mit SIGPIPE (siehe `release.sh`).
programme_gehaertet() {
  local app="$1" zahl=0 datei siegel
  while IFS= read -r -d '' datei; do
    case "$(file -b "$datei")" in
      *Mach-O*executable*) ;;
      *) continue ;;
    esac
    zahl=$((zahl + 1))
    siegel="$(codesign -dv "$datei" 2>&1)"
    case "$siegel" in
      *"(adhoc,runtime)"*) ;;
      *) printf '%s\n' "${datei#"$app"/}"; return 1 ;;
    esac
  done < <(find "$app/Contents" -type f -perm -u+x -print0)
  printf '%s\n' "$zahl"
}
