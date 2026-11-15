#!/usr/bin/env bash
#
# Prüfungen des Semantikvergleichs — Befund 3 des Zweitblicks über Phase
# 0.7.0-I zu PERF-01 (#2168).
#
# **Ein Vergleich ist kein Beleg, wenn beide Seiten scheitern.** Bis dahin
# endete `leistungssemantik.sh` mit 0, wenn jede Ausfuhr und jede Zeichnung
# auf beiden Seiten gleich fehlschlug: es gab nichts, worin sie sich
# unterschieden. Geprüft wird hier an einem nachgebauten `mse`, das
# einfährt, Karten nennt und — je nach Schalter — ausführt und zeichnet oder
# scheitert. Das echte `mse`, das Pack und die kanonischen Sets braucht der
# Test nicht.
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ort="$(mktemp -d)"
trap 'rm -rf "$ort"' EXIT

# Ein Zuhause mit leeren kanonischen Sets und einem leeren Pack.
heim="$ort/heim"
mkdir -p "$heim/Library/Application Support/MagicSetEditor4Mac/kanonische-sets" \
	"$heim/full-magic-pack/data" "$heim/full-magic-pack/Magic - Fonts"

# doppel schreibt ein `mse`, das je nach $1 gelingt, scheitert oder anders
# zeichnet.
doppel() {
	local datei="$ort/mse-$1"
	cat >"$datei" <<SKRIPT
#!/usr/bin/env bash
art="$1"
case "\$1" in
oeffnen)
	echo '{"setID":"7"}'
	if [ "\$art" = oeffnenscheitert ]; then exit 19; fi ;;
kartenliste)
	[ "\$art" = listeleer ] || echo "Karte 1" ;;
karten)
	echo '{"karten":[{"id":"11"},{"id":"12"}]}'
	if [ "\$art" = kartenscheitert ]; then exit 23; fi ;;
vorlagen)
	[ "\$art" = scheitert ] && { echo "Vorlage fehlt" >&2; exit 1; }
	echo "ausgabe" >"\$5" ;;
zeichnen)
	[ "\$art" = scheitert ] && exit 1
	echo "bild-\$art" >"\$3" ;;
esac
SKRIPT
	chmod +x "$datei"
	echo "$datei"
}

fall() { # fall <name> <erwarteter Exitcode> <vorher> <nachher>
	local name="$1" erwartet="$2" code=0
	HOME="$heim" "$wurzel/scripts/leistungssemantik.sh" "$(doppel "$3")" "$(doppel "$4")" "$ort/$name" >"$ort/$name.log" 2>&1 || code=$?
	if [ "$code" -ne "$erwartet" ]; then
		echo "✗ $name: Exitcode $code, erwartet $erwartet"
		cat "$ort/$name.log"
		exit 1
	fi
	echo "✓ $name"
}

fall "beide gelingen gleich" 0 gelingt gelingt
fall "beide scheitern gleich" 1 scheitert scheitert
fall "die Bilder unterscheiden sich" 1 gelingt anders

# **Auch Einfahren und Kartennennung zählen** — Befund 2 des Nachlaufs: ein
# `mse`, das gültiges JSON schreibt und trotzdem mit einem Fehler endet, und
# eine leere Kartenliste galten als gelungen.
fall "das Einfahren endet mit einem Fehler" 1 oeffnenscheitert oeffnenscheitert
fall "die Kartennennung endet mit einem Fehler" 1 kartenscheitert kartenscheitert
fall "die Kartenliste ist leer" 1 listeleer listeleer
