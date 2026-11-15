#!/usr/bin/env bash
#
# Zwei Bauläufe, Eintrag für Eintrag verglichen — I06
# (#2264).
#
#     ./scripts/baugleich.sh zweimal <raum>      # zwei Bauten am HEAD, dann beide Vergleiche
#     ./scripts/baugleich.sh vergleiche <bündel-a> <bündel-b>
#     ./scripts/baugleich.sh abbilder <abbild-a> <abbild-b>
#
# ## Was »unabhängig« hier heißt
#
# `zweimal` legt **zwei Arbeitsbäume** am selben Commit an, jeden an einem
# anderen Pfad, jeden mit **eigenem Go-Bauzwischenspeicher** und eigenem
# `node_modules` aus `npm ci`, und baut in jedem mit `release.sh` unter
# `MSE_FREIGABE_NUR_BAU=1` — dem Freigabeweg ohne Pipeline und ohne Freigabe.
# Geteilt wird allein der Modulspeicher: er ist nach Prüfsumme adressiert, und
# ein zweiter hieße, jedes Modul noch einmal aus dem Netz zu holen. Ein
# Pfad, der ins Programm gelangte, ein Zwischenspeicher, der einen alten Stand
# lieferte, oder eine Uhrzeit im Ergebnis stünden damit als Unterschied da.
#
# ## Was er zusagt
#
# **Jeder Unterschied steht mit seinem Grund da, oder der Lauf endet rot.**
# Byte-Gleichheit behauptet er nicht: wo sie fehlt, nennt er die Datei, und der
# Grund ist einer von dreien, jeder an der Datei **nachgeprüft** und nicht aus
# ihrem Namen geschlossen:
#
# - `baustand.json` — **der Zeitpunkt des Baus**: ohne das Feld `gebaut` sind
#   beide gleich. Ein anderer `head` ist kein Zeitpunkt und bleibt ungeklärt.
# - ein Mach-O-Programm — **nur das Siegel**: nach `codesign
#   --remove-signature` an je einer Kopie sind beide gleich. Das Siegel des
#   Hauptprogramms hasht `Info.plist` und das Siegelverzeichnis mit; ändert
#   sich eines davon, ändert sich das Siegel, und der Code bleibt derselbe.
# - `_CodeSignature/CodeResources` — **folgt aus** den Dateien, deren Hash
#   darin anders steht, und nur dann, wenn **jede** davon selbst als
#   Unterschied dasteht.
#
# Alles andere — eine Datei nur auf einer Seite, ein anderes Linkziel, anderer
# Code — ist **ungeklärt**.
#
# ## Was er nicht prüft
#
# Ob das Bündel **läuft**. Gleich gebaut ist nicht gleich geprüft; dafür gibt
# es die Pipeline und die Läufe am Bündel. Und das Abbild selbst vergleicht er
# nicht Byte für Byte: `hdiutil` schreibt Zeitstempel und Kennungen, deshalb
# hängt `abbilder` beide **nur lesend** ein und vergleicht ihren Inhalt.

baugleich_vergleiche() { # baugleich_vergleiche <a> <b> — Tafel auf stdout, 1 bei Ungeklärtem
  python3 -I - "$1" "$2" <<'PY'
import hashlib, json, os, plistlib, shutil, subprocess, sys, tempfile

a, b = sys.argv[1], sys.argv[2]

def eintraege(wurzel):
    aus = {}
    for ort, ordner, dateien in os.walk(wurzel):
        for name in ordner + dateien:
            pfad = os.path.join(ort, name)
            rel = "./" + os.path.relpath(pfad, wurzel)
            if os.path.islink(pfad):
                aus[rel] = "link:" + os.readlink(pfad)
            elif os.path.isfile(pfad):
                h = hashlib.sha256()
                with open(pfad, "rb") as f:
                    for stueck in iter(lambda: f.read(1 << 20), b""):
                        h.update(stueck)
                aus[rel] = h.hexdigest()
    return aus

ea, eb = eintraege(a), eintraege(b)
alle = sorted(set(ea) | set(eb))
anders = [p for p in alle if ea.get(p) != eb.get(p)]

def stand_ohne_zeit(p):
    try:
        x = json.load(open(os.path.join(a, p)))
        y = json.load(open(os.path.join(b, p)))
    except (OSError, ValueError):
        return False
    x.pop("gebaut", None)
    y.pop("gebaut", None)
    return x == y

def ist_macho(pfad):
    with open(pfad, "rb") as f:
        kopf = f.read(4)
    return kopf in (b"\xcf\xfa\xed\xfe", b"\xce\xfa\xed\xfe", b"\xca\xfe\xba\xbe")

def nur_siegel(p):
    pa, pb = os.path.join(a, p), os.path.join(b, p)
    if not (ist_macho(pa) and ist_macho(pb)):
        return False
    with tempfile.TemporaryDirectory() as raum:
        ka, kb = os.path.join(raum, "a"), os.path.join(raum, "b")
        shutil.copyfile(pa, ka)
        shutil.copyfile(pb, kb)
        for k in (ka, kb):
            if subprocess.run(["codesign", "--remove-signature", k],
                              capture_output=True).returncode != 0:
                return False
        return open(ka, "rb").read() == open(kb, "rb").read()

def siegelfolge(p, erklaert):
    vorne = p[: -len("_CodeSignature/CodeResources")]
    try:
        x = plistlib.load(open(os.path.join(a, p), "rb"))
        y = plistlib.load(open(os.path.join(b, p), "rb"))
    except (OSError, plistlib.InvalidFileException):
        return None
    if {k: v for k, v in x.items() if k not in ("files", "files2")} != \
       {k: v for k, v in y.items() if k not in ("files", "files2")}:
        return None
    folgt = set()
    for teil in ("files", "files2"):
        xa, ya = x.get(teil, {}), y.get(teil, {})
        for k in set(xa) | set(ya):
            if xa.get(k) != ya.get(k):
                folgt.add(k)
    if not folgt or any(vorne + k not in erklaert for k in folgt):
        return None
    return sorted(folgt)

gruende, offen = {}, 0
for p in anders:
    if p not in ea:
        gruende[p] = "**nur im zweiten Bau**"
    elif p not in eb:
        gruende[p] = "**nur im ersten Bau**"
    elif ea[p].startswith("link:") or eb[p].startswith("link:"):
        gruende[p] = None
    elif p.endswith("/baustand.json") and stand_ohne_zeit(p):
        gruende[p] = "der Zeitpunkt des Baus (`gebaut`) — ohne ihn gleich"
    elif nur_siegel(p):
        gruende[p] = "nur das Siegel — ohne es gleich"
    else:
        gruende[p] = None

# Das Siegelverzeichnis zuletzt: es folgt aus Dateien, die vorher erklärt sind.
erklaert = {p for p, g in gruende.items() if g and not g.startswith("**")}
for p in anders:
    if gruende[p] is None and p.endswith("/_CodeSignature/CodeResources"):
        folgt = siegelfolge(p, erklaert)
        if folgt:
            gruende[p] = "folgt aus: " + ", ".join("`%s`" % k for k in folgt)

if anders:
    print("| Eintrag | Grund |")
    print("| --- | --- |")
for p in anders:
    g = gruende[p]
    if g is None:
        g = "**ungeklärt**"
    if g.startswith("**"):
        offen += 1
    print("| `%s` | %s |" % (p, g))
print()
print("%d Einträge verglichen, %d gleich, %d anders, davon %d ungeklärt"
      % (len(alle), len(alle) - len(anders), len(anders), offen))
sys.exit(1 if offen else 0)
PY
}

baugleich_abbilder() { # baugleich_abbilder <abbild-a> <abbild-b>
  local raum rc
  raum="$(mktemp -d)"
  mkdir -p "$raum/a" "$raum/b"
  if ! hdiutil attach -readonly -nobrowse -noverify -mountpoint "$raum/a" "$1" >/dev/null; then
    rm -rf "$raum"; printf 'das erste Abbild ließ sich nicht einhängen\n' >&2; return 2
  fi
  if ! hdiutil attach -readonly -nobrowse -noverify -mountpoint "$raum/b" "$2" >/dev/null; then
    hdiutil detach "$raum/a" >/dev/null
    rm -rf "$raum"; printf 'das zweite Abbild ließ sich nicht einhängen\n' >&2; return 2
  fi
  baugleich_vergleiche "$raum/a" "$raum/b"
  rc=$?
  hdiutil detach "$raum/a" >/dev/null
  hdiutil detach "$raum/b" >/dev/null
  rm -rf "$raum"
  return "$rc"
}

baugleich_zweimal() { # baugleich_zweimal <raum> — Bauprotokolle und Vergleiche unter <raum>
  local raum="$1" wurzel commit version seite rc_buendel rc_abbild
  wurzel="$(git rev-parse --show-toplevel)" || return 2
  commit="$(git -C "$wurzel" rev-parse HEAD)"
  # shellcheck source=scripts/releaseversion.sh
  version="$(source "$wurzel/scripts/releaseversion.sh" && version_aus_wails "$wurzel/wails.json")"
  mkdir -p "$raum"
  for seite in a b; do
    git -C "$wurzel" worktree add --detach "$raum/$seite" "$commit" >/dev/null 2>&1 \
      || { printf 'der Arbeitsbaum %s ließ sich nicht anlegen\n' "$seite" >&2; return 2; }
    mkdir -p "$raum/$seite/.go"
    ln -s "$wurzel/.go/mod" "$raum/$seite/.go/mod"
    ( cd "$raum/$seite" && MSE_FREIGABE_NUR_BAU=1 ./scripts/release.sh "$version" ) \
      > "$raum/bau-$seite.log" 2>&1
    printf 'bau-%s rc=%s\n' "$seite" "$?" >> "$raum/zweimal.txt"
  done
  baugleich_vergleiche "$raum/a/build/bin/MagicSetEditor4Mac.app" "$raum/b/build/bin/MagicSetEditor4Mac.app" \
    > "$raum/buendel.md"
  rc_buendel=$?
  baugleich_abbilder "$raum/a/build/freigabe/MagicSetEditor4Mac-$version.dmg" \
    "$raum/b/build/freigabe/MagicSetEditor4Mac-$version.dmg" > "$raum/abbild.md"
  rc_abbild=$?
  printf 'commit=%s fassung=%s buendel rc=%s abbild rc=%s\n' "$commit" "$version" "$rc_buendel" "$rc_abbild" \
    >> "$raum/zweimal.txt"
  [ "$rc_buendel" = 0 ] && [ "$rc_abbild" = 0 ]
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  set -uo pipefail
  case "${1:-}" in
    zweimal)    baugleich_zweimal "$2" ;;
    vergleiche) baugleich_vergleiche "$2" "$3" ;;
    abbilder)   baugleich_abbilder "$2" "$3" ;;
    *) printf 'Aufruf: %s zweimal <raum> | vergleiche <bündel-a> <bündel-b> | abbilder <abbild-a> <abbild-b>\n' "$0" >&2; exit 2 ;;
  esac
fi
