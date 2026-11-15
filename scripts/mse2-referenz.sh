#!/usr/bin/env bash
# Lässt **MSE2 selbst** die kanonischen Sets rastern — die Referenz mit
# bekannter Herkunft (#407).
#
# Eingabe *und* Ausgabe stammen damit von MSE2; die 27 Bilder aus der
# Gemeinschaft taten das nicht (siehe docs/entscheidungen/0026).
#
# Voraussetzung: MSE2 ist gebaut. Wie, steht in
# docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md — samt der acht
# Zeilen aus docs/mse2-macos.patch, ohne die es auf arm64 nicht linkt.
#
# **Und seit QA-03 (#1656)
# scheitert es, statt eine falsche Referenz zu schreiben** — bei einem
# fehlenden Schriftverzeichnis, bei Schriften, die MSE2 nicht so hat wie wir,
# und es legt je Satz ein Manifest, damit kein altes Bild unbemerkt liegen
# bleibt. Geprüft in `scripts/mse2-referenz.test.sh`.
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
quelle="${MSE2_QUELLE:-$HOME/mse2-quelle}"
programm="${MSE2_BINAER:-$quelle/build/magicseteditor}"
pakete="${MSE_FREMDPAKETE:-/Users/david/full-magic-pack/data}"
unterstuetzung="$HOME/Library/Application Support/MagicSetEditor4Mac"
sets="${MSE_KANON:-$unterstuetzung/kanonische-sets}"
ziel="${MSE_MSE2AUSGABE:-$unterstuetzung/mse2-ausgabe}"
installiert="${MSE2_SCHRIFTEN:-$HOME/Library/Fonts/mse2-referenz}"

# **Die Schriftprobe ist austauschbar** — für den Test dieses Skripts; im
# Betrieb ist sie der Lauf, der dieselbe Frage für den Beleg beantwortet.
schriftprobe="${MSE2_SCHRIFTPROBE:-go test -count=1 -tags fremdpakete ./internal/app -run ^TestDieReferenzschriftenReichen\$}"

if [[ ! -x "$programm" ]]; then
	echo "MSE2 ist nicht gebaut: $programm" >&2
	echo "Anleitung: docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md" >&2
	exit 1
fi

# **MSE2 sucht seine Pakete neben dem Programm**, unter `data`
# (`src/util/io/package_manager.cpp`, Zeilen 205–222). Ein Verweis genügt.
ln -sfn "$pakete" "$(dirname "$programm")/data"

# **Die Schriften des Packs müssen installiert sein.**
#
# MSE2 fragt wxWidgets nach dem **Namen** einer Schrift. Ist sie auf dem Rechner
# nicht installiert, nimmt wx eine Ersatzschrift — und die Referenz zeigt dann
# eine Karte, die MSE2 unter Windows nie zeichnen würde. Nachgemessen an
# `Archenemy`: der Abstand fiel von 25,190 auf **20,018**, nachdem die Schriften
# lagen. Ein Fünftel des gemessenen Abstands war der Fehler der Referenz.
#
# **Ohne Schriftverzeichnis keine Referenz.** Ein `if [[ -d "$schriften" ]]`
# übersprünge bei fehlendem Verzeichnis die Prüfung ganz.
schriften="$(dirname "$pakete")/Magic - Fonts"

if [[ ! -d "$schriften" ]]; then
	echo "Das Schriftverzeichnis des Packs fehlt: $schriften" >&2
	exit 1
fi

if [[ ! -d "$installiert" ]]; then
	echo "Die Schriften für MSE2 fehlen: $installiert" >&2
	echo "Legen mit:" >&2
	echo "  mkdir -p \"$installiert\"" >&2
	echo "  cp \"$schriften\"/*.[tT][tT][fF] \"$installiert\"/" >&2
	exit 1
fi

# **Und jede Schrift, die die Vorlagen der Sätze anfordern, liegt für MSE2 so
# da wie für uns** — dieselbe Datei. Gefragt wird der Lauf, der das für den
# Beleg beantwortet (`TestDieReferenzschriftenReichen`); er nennt, was fehlt.
if ! (cd "$wurzel" && MSE_FREMDPAKETE="$pakete" MSE_SCHRIFTWURZEL="$(dirname "$pakete")" \
	MSE_KANON="$sets" MSE2_SCHRIFTEN="$installiert" $schriftprobe >&2); then
	echo "Die Schriften für MSE2 reichen nicht — die Probe oben nennt, welche fehlen." >&2
	exit 1
fi

# hash gibt den SHA-256 einer Datei.
hash() { shasum -a 256 "$1" | cut -d' ' -f1; }

# schrifthash gibt einen Hash über Namen und Inhalt jeder installierten Schrift.
schrifthash() {
	(cd "$installiert" && find . -type f ! -name '.*' -print0 | sort -z |
		xargs -0 shasum -a 256) | shasum -a 256 | cut -d' ' -f1
}

mkdir -p "$ziel"
programmhash="$(hash "$programm")"
schriftenhash="$(schrifthash)"

# **MSE2 bekommt ein eigenes Zuhause.** Es lädt ein Paket zuerst aus seinem
# Benutzerverzeichnis und erst dann aus dem neben dem Programm
# (`PackageManager::openAny`, `src/util/io/package_manager.cpp`, Zeilen 47–54;
# das Verzeichnis ist `GetUserDataDir() + "/data"`, Zeile 205, Fassung f9d9356
# vom 2021-07-06). Auf diesem Rechner lag dort `M15-Magic-Pack` — 130 der 325
# Pakete, die beide Packs führen, sind verschieden, darunter `magic.mse-game`
# und drei Stile des Kanons. Ohne das Verzeichnis gilt nur `$pakete`, und
# `mse.config` fehlt auch: MSE2 nimmt dann seine Vorgaben, darunter
# `locale: en` und `card_anti_alias` (`src/data/settings.cpp`, Zeilen 126 und
# 165).
#
# **`HOME` allein isoliert MSE2 nicht**
# (#2006).
# wxWidgets fragt auf macOS Core Foundation nach dem Benutzerverzeichnis, und
# Core Foundation liest `HOME` nicht — erst `CFFIXED_USER_HOME` lenkt es um.
# Gemessen am 2026-10-02 (`docs/LEISTUNGSVERGLEICH.md` §2): mit leerem `HOME`
# lud MSE2 `magic.mse-game` aus `~/Library/Application Support/magicseteditor/data`.
# **Die Messung vom 2026-09-23** („es änderte hier nichts") verglich damit
# zweimal denselben Zustand. **Mit echter Isolation wiederholt am
# 2026-10-03**: `Archenemy`, `Black` und `Multicolor` — 6.866 von 6.866
# Bildern bytegleich mit der Referenz von davor, bei gleichem Programm, gleichen
# Schriften und gleichem Set (`docs/LEISTUNGSVERGLEICH.md` §6). Isoliert wird,
# damit es so bleibt, wenn sich eines der beiden Packs ändert.
#
# **Und darin muss `Library/Application Support/magicseteditor` liegen** —
# ohne den Ordner hängt MSE2 beim Beenden (gemessen am 2026-10-02).
zuhause="$(mktemp -d)"
trap 'rm -rf "$zuhause"' EXIT
mkdir -p "$zuhause/Library/Application Support/magicseteditor"

for satz in "$@"; do
	datei="$sets/$satz.mse-set"
	if [[ ! -e "$datei" ]]; then
		echo "kein Set: $datei" >&2
		exit 1
	fi

	echo "== $satz"

	# **Alte Bilder zuerst weg.** MSE2 nummeriert gleichnamige Karten eines
	# Laufs durch und überschreibt, was es neu schreibt — was aus einem
	# früheren Lauf übrig ist, rührt es nicht an (`CONFLICT_NUMBER_OVERWRITE`,
	# `src/main.cpp`, Zeile 254, und `src/util/file_utils.cpp`, Zeilen 88–95,
	# Fassung f9d9356 vom 2021-07-06). Es bliebe unbemerkt liegen.
	ausgabe="$ziel/$satz"
	rm -rf "$ausgabe"
	mkdir -p "$ausgabe"

	HOME="$zuhause" CFFIXED_USER_HOME="$zuhause" "$programm" --export-images "$datei" "$ausgabe/{card.name}.png"

	karten="$(unzip -p "$datei" set | grep -c '^card:' || true)"
	bilder="$(find "$ausgabe" -name '*.png' | wc -l | tr -d ' ')"

	# **Mehr Bilder als Karten gibt es nicht** — außer, jemand schrieb daneben.
	# Weniger sind möglich: eine Karte ohne Dateinamen legt MSE2 nicht ab
	# (`src/data/format/image.cpp`, Zeile 82).
	if ((bilder > karten)); then
		echo "$satz: $bilder Bilder bei $karten Karten — da liegt etwas, das nicht dazugehört" >&2
		exit 1
	fi

	# **Das Manifest** — Set, Programm, Schriften und jedes Bild mit Hash.
	# `saetzeMitReferenz` hält es gegen das heutige Set und gegen das
	# Verzeichnis (`internal/app/referenzmanifest_intern_test.go`).
	{
		printf '{\n  "format": 1,\n'
		printf '  "set": "sha256:%s",\n' "$(hash "$datei")"
		printf '  "programm": "sha256:%s",\n' "$programmhash"
		printf '  "schriften": "sha256:%s",\n' "$schriftenhash"
		printf '  "karten": %s,\n' "$karten"
		printf '  "erzeugt": "%s",\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
		printf '  "bilder": {'
		trenner=''
		while IFS= read -r bild; do
			printf '%s\n    "%s": "sha256:%s"' "$trenner" "$bild" "$(hash "$ausgabe/$bild")"
			trenner=','
		done < <(cd "$ausgabe" && find . -name '*.png' | sed 's|^\./||' | LC_ALL=C sort)
		printf '\n  }\n}\n'
	} > "$ausgabe/manifest.json"

	echo "   $bilder Bilder, $karten Karten, Manifest geschrieben"
done
