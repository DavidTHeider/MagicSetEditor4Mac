package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein `include file:` bringt seine Datei mit — und behält, was unter ihm
// steht** (#240).
//
// # Der Befund
//
// `magic-blends.mse-include/new-blends` beginnt so:
//
//	include file: masks
//
//	mask_hybrid_with_land :=	{ false }
//	…
//	trim_colors :=
//	    replace@(match:"…", replace:"")
//
// Alle diese Zeilen sind eingerückt, und im Baum hängen sie damit **unter** der
// `include file:`-Zeile. Der Einbinder ersetzte den Knoten durch den Inhalt der
// geladenen Datei und ließ seine Kinder fallen.
//
// Nachgemessen am zusammengesetzten Spielskript von `magic.mse-game`, 19.125
// Zeilen: `trim_colors` wird **achtmal gerufen und nirgends definiert**.
// Dasselbe für `mask_hybrid_with_land`, `hybrid_previews`, `colored_lands` und
// `id_mask`. Die *unein*gerückten Vereinbarungen derselben Dateien —
// `color_background`, `watermark_colors`, `ident_include` — waren da.
//
// Das war der Grund, warum `card color → image` in `magic-m15.mse-style` nicht
// zeichnete: die Angabe ruft über drei Ecken `trim_colors`.
func TestEinIncludeBehaeltSeineKinder(t *testing.T) {
	quelle := dateien{"m.mse-include/geladen": "geladen: ja\n"}

	baum, befunde := einbinden(t, quelle,
		"include file: /m.mse-include/geladen\n\tdarunter: ja\n\tnochmal: ja\n")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	schluessel := make([]string, 0, 3)
	sammle(&schluessel, baum.Wurzel)

	if gebaut := strings.Join(schluessel, ","); gebaut != "geladen,darunter,nochmal" {
		t.Errorf("%q — die Kinder des Include sind weg", gebaut)
	}
}

// **Auch ein Verweis im Wert einer Zeile behält, was unter ihm steht.**
//
// `magic.mse-game` schreibt `init script: include file: script`, und darunter
// darf weiterer Text stehen. Diese Form ging schon richtig — der Test hält sie
// fest, damit die Behebung der einen Form die andere nicht mitreißt.
func TestEinVerweisImWertBehaeltEbenfalls(t *testing.T) {
	quelle := dateien{"m.mse-include/geladen": "geladen: ja\n"}

	baum, befunde := einbinden(t, quelle,
		"init script: include file: /m.mse-include/geladen\n\tdarunter: ja\n")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	schluessel := make([]string, 0, 3)
	sammle(&schluessel, baum.Wurzel)

	if gebaut := strings.Join(schluessel, ","); gebaut != "init script,geladen,darunter" {
		t.Errorf("%q", gebaut)
	}
}

// **Und die Kinder werden selbst aufgelöst.**
//
// Steht unter einem `include file:` noch eines, muss auch das seine Datei
// bekommen. Ohne diesen Fall wäre die Behebung eine halbe: die Kinder wären
// zurück, aber roh.
func TestDieKinderEinesIncludeWerdenSelbstAufgeloest(t *testing.T) {
	quelle := dateien{
		"m.mse-include/eins": "eins: ja\n",
		"m.mse-include/zwei": "zwei: ja\n",
	}

	baum, befunde := einbinden(t, quelle,
		"include file: /m.mse-include/eins\n\tinclude file: /m.mse-include/zwei\n")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	schluessel := make([]string, 0, 2)
	sammle(&schluessel, baum.Wurzel)

	if gebaut := strings.Join(schluessel, ","); gebaut != "eins,zwei" {
		t.Errorf("%q", gebaut)
	}
}

// sammle geht über den ganzen Baum, weil die Einrückung Kinder erzeugt.
func sammle(hinein *[]string, knoten []*msepaket.Knoten) {
	for _, kind := range knoten {
		*hinein = append(*hinein, kind.Schluessel)
		sammle(hinein, kind.Kinder)
	}
}

// **`include_file:` ist derselbe Verweis wie `include file:`** — QA-06b
// (#1967).
//
// MSE2 macht jeden Schlüssel kanonisch, bevor es ihn vergleicht
// (`canonical_name_form(trim(key))`, `src/util/io/reader.cpp` Zeile 217), und
// fragt dann nach `include_file` (`src/util/io/reader.hpp` Zeile 175, Fassung
// f9d9356 vom 2021-07-06). Beide Schreibweisen binden dort ein; hier blieb
// die mit Unterstrich als gewöhnlicher Eintrag stehen.
func TestEinVerweisInKanonischerSchreibweiseWirdEingebunden(t *testing.T) {
	t.Parallel()

	quelle := dateien{
		"m.mse-include/scripts": "mana_t: eins\n",
		"m.mse-include/keys_en": "schluessel: ja\n",
	}

	baum, befunde := einbinden(t, quelle,
		"include_file: /m.mse-include/scripts\ninclude_localized_file: /m.mse-include/keys\n")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if baum.Ersten("mana_t") == nil || baum.Ersten("schluessel") == nil {
		t.Errorf("nicht eingebunden: %d Knoten, der erste %q", len(baum.Wurzel), baum.Wurzel[0].Schluessel)
	}
}
