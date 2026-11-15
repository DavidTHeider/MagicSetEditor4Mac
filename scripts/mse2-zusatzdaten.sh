#!/usr/bin/env bash
#
# **Was MSE2 mit Zusatzdaten macht — gemessen und nicht angenommen** — FMT-05.AK-02
# (#1694).
#
#     scripts/mse2-zusatzdaten.sh
#
# Öffnet ein Set mit **eingebetteten** Zusatzfeldern (`mse4mac …` im Setkopf
# und in einer Karte) und einer **Datei daneben** (`probe.mse4mac.json`) im
# echten MSE2, speichert es über `write_set_file` wieder und vergleicht.
#
# **Ein eigener Arbeitsort**: MSE2 wird in ein Wegwerfverzeichnis kopiert und
# bekommt dort einen eigenen `data`-Ordner — Verweise auf das Paket-Pack plus
# eine kleine Rundlaufvorlage. Der MSE2-Datenordner dieses Rechners bleibt
# unberührt. **Nicht Teil von `make ci`**: er braucht das gebaute MSE2 (0026)
# und das Paket-Pack.

set -euo pipefail

quelle="${MSE2_QUELLE:-$HOME/mse2-quelle}"
programm="${MSE2_BINAER:-$quelle/build/magicseteditor}"
pakete="${MSE_FREMDPAKETE:-/Users/david/full-magic-pack/data}"

[[ -x "$programm" ]] || { echo "MSE2 ist nicht gebaut: $programm" >&2; exit 1; }
[[ -d "$pakete" ]] || { echo "Das Paket-Pack fehlt: $pakete" >&2; exit 1; }

ort="$(mktemp -d "${TMPDIR:-/tmp}/mse2-zusatzdaten.XXXXXX")"
trap 'rm -rf "$ort"' EXIT

cp "$programm" "$ort/magicseteditor"
mkdir -p "$ort/data/rundlauf.mse-export-template" "$ort/probe.mse-set"
for paket in "$pakete"/*; do ln -s "$paket" "$ort/data/"; done

cat > "$ort/data/rundlauf.mse-export-template/export-template" <<'VORLAGE'
mse version: 2.0.0
game: magic
short name: Rundlauf
full name: Rundlauf
file type: *.txt|*.txt
create directory: true
script:
	write_set_file(file:"zurueck.mse-set")
	"geschrieben"
VORLAGE

printf 'mse version: 2.0.0\ngame: magic\nstylesheet: m15\nset info:\n\ttitle: Rundlaufprobe\n\tmse4mac preisstand: 4.17 EUR\ncard:\n\tname: Eins\n\tnotes: eine Notiz\n\tmse4mac exemplar: nearMint\n' \
	> "$ort/probe.mse-set/set"
echo '{"fassung":1,"preise":{"Eins":"4.17 EUR"}}' > "$ort/probe.mse4mac.json"

cd "$ort"
echo "MSE2: $(git -C "$quelle" log -1 --format='%h vom %ad' --date=short)"
echo
echo "── Öffnen (Meldungen von MSE2)"
./magicseteditor --export rundlauf probe.mse-set ausgabe.txt 2>&1 | grep -E 'Unexpected key: .mse4mac' | sed 's/\x1b\[[0-9;]*m//g' || true

mkdir zurueck
(cd zurueck && unzip -q ../ausgabe-files/zurueck.mse-set)

echo
echo "── Nach dem Speichern durch MSE2"
for feld in 'mse4mac preisstand' 'mse4mac exemplar' 'notes: eine Notiz'; do
	if grep -rq "$feld" zurueck; then echo "  bleibt:   $feld"; else echo "  verloren: $feld"; fi
done
inhalt="$(find zurueck -type f -exec basename {} \; | sort | tr '\n' ' ')"
if [[ -f probe.mse4mac.json ]]; then echo "  die Datei daneben liegt unverändert da, und MSE2 nahm sie nicht mit: $inhalt"; fi
