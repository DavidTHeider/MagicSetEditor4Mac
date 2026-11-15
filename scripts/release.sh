#!/usr/bin/env bash
#
# Die Freigabe (docs/RELEASE.md).
#
#     ./scripts/release.sh 0.1.0
#
# Sie entsteht **lokal**, nicht in GitHub Actions (AGENTS.md §4). Ziel ist
# ausschließlich macOS auf Apple Silicon.
#
# ## Keine Notarisierung
#
# Es gibt keinen Apple-Developer-Account und es soll keinen geben
# (docs/entscheidungen/0003). Das `.app` wird **ad hoc** signiert — auf arm64
# ist das keine Wahl, sondern Voraussetzung, sonst startet es überhaupt nicht.
# `spctl --assess` wird die App danach ablehnen. Das ist der **erwartete**
# Zustand; dieses Skript sagt es an der Stelle, an der der Befund auftaucht,
# damit ihn niemand später „behebt" und damit die Entscheidung rückgängig macht,
# ohne sie zu ändern.

set -euo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$WURZEL"

# shellcheck source=scripts/releaseversion.sh
source "$WURZEL/scripts/releaseversion.sh"

# shellcheck source=scripts/releasebeleg.sh
source "$WURZEL/scripts/releasebeleg.sh"

# shellcheck source=scripts/releasesignatur.sh
source "$WURZEL/scripts/releasesignatur.sh"

# shellcheck source=scripts/releasesignieren.sh
source "$WURZEL/scripts/releasesignieren.sh"

# shellcheck source=scripts/releasezettel.sh
source "$WURZEL/scripts/releasezettel.sh"

# shellcheck source=scripts/sicherheitslage.sh
source "$WURZEL/scripts/sicherheitslage.sh"
# shellcheck source=scripts/releaseschluss.sh
source "$WURZEL/scripts/releaseschluss.sh"

# shellcheck source=scripts/releaseschalter.sh
source "$WURZEL/scripts/releaseschalter.sh"

export GOPATH="$WURZEL/.go"
export GOBIN="$WURZEL/.go/bin"
export GOMODCACHE="$WURZEL/.go/mod"
export GOCACHE="$WURZEL/.go/build"
export PATH="$GOBIN:$PATH"

# **Die gepinnte Toolchain und die macOS-Untergrenze gelten auch hier** (§1,
# ENG-02). Die Pipeline setzt sie für sich selbst; `werkzeuge.sh` baut danach
# drei Programme, und ohne diese Zeile baute sie der Go des Rechners.
# shellcheck source=/dev/null
source "$WURZEL/scripts/activate.sh" >/dev/null 2>&1

ROT=$'\033[31m'; GRUEN=$'\033[32m'; GELB=$'\033[33m'; BLAU=$'\033[34m'; AUS=$'\033[0m'
[ -t 1 ] || { ROT=''; GRUEN=''; GELB=''; BLAU=''; AUS=''; }

kopf() { printf '\n%s── %s %s\n' "$BLAU" "$1" "$AUS"; }
ok()   { printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"; }
hin()  { printf '%s  ~ %s%s\n' "$GELB" "$1" "$AUS"; }
ende() { printf '%s  ✗ %s%s\n' "$ROT" "$1" "$AUS"; exit 1; }

VERSION="${1:-}"
NAME="MagicSetEditor4Mac"
APP="build/bin/$NAME.app"
AUSGABE="build/freigabe"

# ------------------------------------------------------------------ Vorprüfung

kopf "Vorprüfung"

# **Ein Bau ohne Freigabe fragt nicht nach dem Tag** — er entsteht auch für
# eine Fassung, die schon heraus ist, und gibt sie nicht noch einmal heraus
# (`MSE_FREIGABE_NUR_BAU`, unten bei der Pipeline).
if [ "${MSE_FREIGABE_NUR_BAU:-}" = 1 ]; then
  [ "$VERSION" = "$(version_aus_wails)" ] || ende "wails.json nennt nicht $VERSION"
  hin "nur der Bau — Fassung $VERSION aus wails.json, ohne Abgleich mit CHANGELOG.md und Tags"
else
  versionen_pruefen "$VERSION" || ende "die Versionsangaben laufen auseinander"
  ok "Version $VERSION in wails.json und CHANGELOG.md"
fi

if [ -n "$(git status --porcelain)" ]; then
  git status --short | sed 's/^/      /'
  ende "der Arbeitsbaum ist nicht sauber"
fi
ok "Arbeitsbaum sauber"

if [ "$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then
  hin "die Freigabe entsteht nicht auf main, sondern auf $(git rev-parse --abbrev-ref HEAD)"
fi

# ---------------------------------------------------------------- Pipeline

kopf "Pipeline"

# **`MSE_FREIGABE_NUR_BAU=1` baut, ohne freizugeben** (I06,
# #2264) —
# für den Vergleich zweier Bauläufe in `scripts/baugleich.sh`. Ohne Pipeline
# entsteht deshalb keine Freigabe: der Lauf endet nach dem Abbild, vor
# Stücklisten, Prüfsummen, Signatur und Beleg, und `baugleich.test.sh` hält die
# Reihenfolge.
if [ "${MSE_FREIGABE_NUR_BAU:-}" = 1 ]; then
  hin "nur der Bau — ohne Pipeline, und deshalb ohne Freigabe"
else
  ./scripts/ci.sh all || ende "die Pipeline ist rot"
fi

# -------------------------------------------------------------------- Bau

kopf "Bau"
./scripts/bauen.sh freigabe || ende "der Bau ist gescheitert"
[ -d "$APP" ] || ende "$APP ist nicht entstanden"
ok "$APP gebaut"

# **Die Bauschalter am fertigen Programm** (ENG-02.AK-05) — warum dort und
# nicht am Aufruf, steht in `releaseschalter.sh`.
BEFUND="$(go version -m "$APP/Contents/MacOS/$NAME" | schalter_aus | schalter_befund)"
[ -z "$BEFUND" ] || ende "das Programm trägt die falschen Bauschalter: $BEFUND"
ok "Bauschalter: production, die Weglassschalter und sonst keiner"

# --------------------------------------------------------------- Signatur

# ---------------------------------------------------------------- Beigaben

kopf "Vorlagen und Beispielset"

# **Erst hier und nicht in `scripts/bauen.sh`.** Es sind 167 MB; wer sie bei
# jedem `make dev` kopierte, machte aus jedem Neustart eine Kaffeepause
# (#702).
"$WURZEL/scripts/mitliefern.sh" "$APP/Contents/MacOS/$NAME" \
  || ende "die Beigaben ließen sich nicht ins Bündel legen"

# ---------------------------------------------------------------- Werkzeuge

kopf "Kommandozeile, MCP-Server und Dienst"

# **Vor der Signatur**, damit `codesign --deep` sie mitnimmt: auf arm64 startet
# ein unsigniertes Binärprogramm nicht ([0003](../docs/entscheidungen/0003-keine-notarisierung.md)).
#
# Bis zum 2026-08-24 baute dieser Weg **ein** Programm, und `mse`, `mse-mcp`
# wie `mse-rest` erreichten niemanden, der die Anwendung installiert
# (#992).
"$WURZEL/scripts/werkzeuge.sh" "$APP" \
  || ende "die Werkzeuge ließen sich nicht ins Bündel legen"

# **Die Weglassschalter an jedem Werkzeug** (#2393) — wie oben am Programm,
# am fertigen Binärprogramm und nicht am Aufruf.
for werkzeug in mse mse-mcp mse-rest; do
  BEFUND="$(go version -m "$APP/Contents/MacOS/$werkzeug" | schalter_aus | schalter_befund werkzeug)"
  [ -z "$BEFUND" ] || ende "$werkzeug trägt die falschen Bauschalter: $BEFUND"
done
ok "Bauschalter der Werkzeuge: die Weglassschalter und sonst keiner"

kopf "Signatur"

# **Mit Hardened Runtime und ohne Entitlements** (SEC03, #2356, Vermerk
# 0195): gemessen am Abnahmebündel — WebKit, SQLite, purego, Touch ID,
# Teilen, Dateidialoge und die drei Werkzeuge laufen so. Die Laufzeit nimmt
# einem fremden Prozess `DYLD_INSERT_LIBRARIES` und das Anhängen eines
# Debuggers; was Apple notarisiert, ist eine andere Frage (0003).
codesign --force --deep --options runtime --sign - "$APP"
ok "ad hoc signiert, mit Hardened Runtime"

# Erst auffangen, dann durchsuchen — nicht in einer Pipeline.
#
# `codesign … | grep -q` sieht richtig aus und ist unter `set -o pipefail` eine
# Falle: `grep -q` endet beim ersten Treffer, `codesign` schreibt weiter in eine
# geschlossene Leitung, bekommt SIGPIPE und endet von null verschieden. Die
# Pipeline gilt damit als gescheitert — ausgerechnet dann, wenn die Signatur
# stimmt.
SIGNATUR="$(codesign -dv --verbose=2 "$APP" 2>&1)"
case "$SIGNATUR" in
  *Signature=adhoc*) ;;
  *) printf '%s\n' "$SIGNATUR" | sed 's/^/      /'
     ende "die Signatur ist nicht ad hoc — hat sich der Bauweg geändert?" ;;
esac
# **Und die Laufzeit trägt die Flagge** — `flags=0x10002(adhoc,runtime)`.
case "$SIGNATUR" in
  *"(adhoc,runtime)"*) ;;
  *) printf '%s\n' "$SIGNATUR" | sed 's/^/      /'
     ende "das Bündel trägt keine Hardened Runtime — hat sich der Bauweg geändert?" ;;
esac
# **Und jedes Programm im Bündel** (Zweitblick über Phase 5): `--deep` siegelt
# die drei Werkzeuge mit, aber ein Siegel nur am Bündel zu lesen, sagte über
# sie nichts — und **überall im Bündel** (Nachlauf, Befund 4,
# `releasehaertung.sh`).
# shellcheck source=scripts/releasehaertung.sh
source "$WURZEL/scripts/releasehaertung.sh"
PROGRAMME="$(programme_gehaertet "$APP")" || ende "$PROGRAMME trägt keine Hardened Runtime"
[ "$PROGRAMME" -ge 4 ] || ende "nur $PROGRAMME Programme im Bündel gefunden — erwartet die Anwendung und drei Werkzeuge"
ok "Signatur nachgesehen: adhoc, Hardened Runtime — an allen $PROGRAMME Programmen im Bündel"

# spctl **muss** hier ablehnen. Ein Erfolg wäre der auffällige Fall: er hieße,
# dass doch ein Developer-ID-Zertifikat im Spiel ist — und damit eine
# Entscheidung gefallen wäre, die niemand aufgeschrieben hat.
#
# Die Deutung des Rückgabewerts steht in `releasesignatur.sh` — dort, wo sie
# sich prüfen lässt, ohne zu bauen. Was sie soll und was daran vorher falsch
# war, steht im Kopf jener Datei.
set +e
spctl --assess --type execute "$APP" >/dev/null 2>&1
SPCTL=$?
set -e

if URTEIL="$(spctl_urteil "$SPCTL")"; then
  hin "$URTEIL"
else
  ende "$URTEIL"
fi

# ------------------------------------------------------------------- Bündel

kopf "Bündel"

rm -rf "$AUSGABE" && mkdir -p "$AUSGABE/dmg"
cp -R "$APP" "$AUSGABE/dmg/"
ln -s /Applications "$AUSGABE/dmg/Programme"
erststartnotiz > "$AUSGABE/dmg/ERSTER START — bitte lesen.txt"

hdiutil create -volname "$NAME $VERSION" -srcfolder "$AUSGABE/dmg" \
  -ov -format UDZO "$AUSGABE/$NAME-$VERSION.dmg" >/dev/null
ok "$AUSGABE/$NAME-$VERSION.dmg"

if [ "${MSE_FREIGABE_NUR_BAU:-}" = 1 ]; then
  hin "nur gebaut — keine Stücklisten, keine Prüfsummen, keine Signatur, kein Beleg"
  exit 0
fi

# --------------------------------------------------------------- Stückliste

kopf "Stückliste und Prüfsummen"

# shellcheck source=scripts/releasestueckliste.sh
source "$WURZEL/scripts/releasestueckliste.sh"

stuecklisten_schreiben "$AUSGABE"

# **Das Kompatibilitätsmanifest** (O6, #1430) — es reist in `pruefsummen.txt`
# mit und ist damit signiert; deshalb steht es **vor** dem `shasum`-Aufruf.
#
# **Eine eigene Datei**, weil `release.sh` mit ihm über 400 Zeilen ginge (§2) —
# dieselbe Trennung wie bei `releaseversion.sh` und den drei anderen.
# shellcheck source=scripts/releasemanifest.sh
source "$WURZEL/scripts/releasemanifest.sh"

manifest_schreiben "$AUSGABE" "$VERSION" "$WURZEL"

# **Und die Herkunft** (I06, #2264): Commit, Werkzeuge, die Summe jedes
# Programms im Bündel und beider Stücklisten — ebenfalls vor `shasum`, damit
# sie unter der Signatur steht.
# shellcheck source=scripts/releaseherkunft.sh
source "$WURZEL/scripts/releaseherkunft.sh"

herkunft_schreiben "$AUSGABE" "$APP" "$WURZEL"

# **Und der Quelltext** (D01, #2331, Vermerk 0190): GPL-2.0 §3 verlangt ihn
# vollständig und an derselben Stelle wie das Abbild — auch er vor `shasum`,
# damit er unter der Signatur steht.
# shellcheck source=scripts/releasequelltext.sh
source "$WURZEL/scripts/releasequelltext.sh"

quelltext_schreiben "$AUSGABE" "$VERSION" "$WURZEL"

(cd "$AUSGABE" && shasum -a 256 ./*.dmg ./*.json ./*.tar.gz > pruefsummen.txt)
ok "pruefsummen.txt"

# --------------------------------------------------------------- Signatur

kopf "Signatur der Freigabe"

# **Eine Signatur über die Prüfsummen, nicht drei über die Artefakte.**
#
# Die Prüfsummen decken das `.dmg` und beide Stücklisten; die Signatur deckt
# die Prüfsummen. Wer sie prüft und danach `shasum -c` fährt, hat die ganze
# Freigabe geprüft — und muss dafür genau **einen** Schlüssel kennen.
#
# **Sie ist etwas anderes als die Ad-hoc-Signatur oben.** Jene sagt dem
# Betriebssystem, dass das Bündel seit dem Bau unverändert ist; sie trägt
# keine Identität, und genau deshalb weist Gatekeeper sie ab
# ([0003](../docs/entscheidungen/0003-keine-notarisierung.md)). Diese sagt,
# **wer** gebaut hat — mit demselben Schlüssel, der unter jedem Commit steht.
#
# **Ein Fehler hier beendet die Freigabe.** Ein unsigniertes Abbild neben einer
# Prüfsummenliste, die jeder mitliefern kann, ist keine Freigabe, sondern eine
# Datei im Netz.
MELDUNG="$(signiere "$AUSGABE/pruefsummen.txt")" ||
  ende "die Signatur scheiterte: $MELDUNG"

ok "pruefsummen.txt.asc"

# **Und sie wird im selben Lauf geprüft.** Eine Signatur, die niemand prüft,
# ist eine Datei (§10).
MELDUNG="$(signaturpruefen "$AUSGABE/pruefsummen.txt")" ||
  ende "die eigene Signatur ließ sich nicht prüfen: $MELDUNG"

# **Und ein leerer Fingerabdruck beendet die Freigabe.**
#
# Ein Vorgabewert wie `${FINGERABDRUCK:-unbekannt}` ließe eine fehlende
# Messung wie ein Ergebnis aussehen — die Stelle, an der ein Wächter seine
# halbe Zusage verliert (§10). Der Beleg ist dafür da, **womit** signiert
# wurde — ohne die Antwort gibt es keinen Beleg.
FINGERABDRUCK="$(signaturfingerabdruck "$AUSGABE/pruefsummen.txt")"
[ -n "$FINGERABDRUCK" ] ||
  ende "die Signatur ließ keinen Fingerabdruck lesen — ohne ihn sagt der Beleg nicht, womit signiert wurde"

# **Und signiert wurde mit dem Schlüssel, gegen den geprüft wird.**
#
# Signiert wird mit dem, was `gpg` **auf diesem Rechner** unter der Kennung
# findet; die ausgelieferte Anwendung prüft gegen den Schlüssel, der in ihr
# eingebettet ist (`internal/aktualisierung/signatur.go`). Laufen sie
# auseinander, ist hier alles grün — und jede installierte Fassung weist die
# Freigabe ab.
MELDUNG="$(schluesselstimmtueberein "$AUSGABE/pruefsummen.txt")" ||
  ende "$MELDUNG"

ok "Signatur nachgesehen: $FINGERABDRUCK"

# ------------------------------------------------------------ Freigabebeleg

# **Ab der Lage in einer Funktion** (#1830): `releaseschluss.test.sh` fährt
# sie mit einer gescheiterten Lage und sieht sie vor „Fertig" enden — und
# macht die Gegenprobe mit einem wirkungslos gemachten Urteil.
freigabeschluss "$AUSGABE" "$VERSION" "$FINGERABDRUCK"

# **Zuerst die Checkliste, dann der Tag.**
#
# Ein Skript, das nach dem Bau als nächstes `gh release create` vorschlägt,
# sagt damit, der Rest sei erledigt — auch was `docs/RELEASE.md` „nicht
# optional" nennt: die Läufe gegen bezogene Daten, hell und dunkel, ein
# wirklich gedruckter Bogen, der Erststart-Weg.
printf '\n%sNoch von Hand — und in dieser Reihenfolge:%s\n' "$GELB" "$AUS"
printf '  1. docs/RELEASE.md, Abschnitt „Vor jeder Freigabe": **jede** Zeile abarbeiten.\n'
printf '     Sie ist nicht optional, und dieses Skript prüft keine einzige davon.\n'
printf '     Jede erledigte Prüfung wird eingetragen — sonst bleibt sie Prosa:\n'
printf '     scripts/releasebeleg.sh haken %s %s <1..7> "<wer>"\n' \
  "$AUSGABE/$(belegname "$VERSION")" "$BELEGHASH"
printf '  2. Das .dmg einhängen und die App wirklich öffnen (AGENTS.md §3).\n'
printf '  3. Erst wenn 1 und 2 stehen:\n'
printf '     git tag -s v%s -m "Fassung %s" && git push origin v%s\n' "$VERSION" "$VERSION" "$VERSION"
# **Nur die Dateien, nicht das Staging-Verzeichnis.** `%s/*` nimmt auch `dmg/`
# mit — den Ordner, aus dem `hdiutil` das Abbild gebaut hat —, und
# `gh release create` bricht an einem Verzeichnis ab. Der erste Anlauf hätte
# genau hier gestanden und wäre nicht gelaufen (Issue 878).
#
# **Und `pruefsummen.txt.asc` gehört dazu.** Sie fehlte hier bis zum
# 2026-08-28, und das war kein Schönheitsfehler: `aktualisierung.Laden` weist
# eine Freigabe **ohne** Signatur ab, bevor es das Abbild überhaupt lädt. Wer
# genau diesem Befehl folgte, veröffentlichte eine Fassung, die die eingebaute
# Aktualisierung nicht annimmt — und die niemand von Hand prüfen kann.
# Angestrichen vom ersten Zweitblick über die Freigabe 0.3.0
# (#1170).
#
# `docs/RELEASE.md` nannte sie die ganze Zeit richtig. Zwei Anleitungen für
# denselben Schritt laufen auseinander, sobald eine davon jemand ändert —
# und `release.test.sh` hält diesen Befehl jetzt gegen die Dateien, die
# danebenliegen.
printf '     gh release create v%s %s/*.dmg %s/*.json %s/*.tar.gz %s/pruefsummen.txt %s/pruefsummen.txt.asc %s/%s --notes-file <notiz>\n\n' \
  "$VERSION" "$AUSGABE" "$AUSGABE" "$AUSGABE" "$AUSGABE" "$AUSGABE" "$AUSGABE" "$(belegname "$VERSION")"
