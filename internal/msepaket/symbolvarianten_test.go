package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Setsymbol hat benannte Fassungen — `variation:`.**
//
//	IMPLEMENT_REFLECTION_NO_SCRIPT(SymbolVariation) {
//	  REFLECT(name);
//	  REFLECT(border_radius);
//	  REFLECT_NAMELESS(filter);
//	}
//	SymbolVariation::SymbolVariation() : border_radius(0.05) {}
//
// `src/data/field/symbol.cpp`, Zeilen 31–46, Fassung f9d9356 vom 2021-07-06.
//
// **Sie stehen an zwei Orten**: das Spiel erklärt sie am `set field:`, die
// Vorlage ergänzt sie unter `set info style:`. `magic.mse-game` führt `common`,
// `uncommon`, `rare` und `mythic rare`; `magic-m15.mse-style` legt
// `invertedcommon` dazu.
//
// **Daran hängt das Seltenheitssymbol jeder Karte** —
// `symbol_variation(symbol: set.symbol, variation: "common")`, 180 Befunde über
// die 190 Stile von `M15-Magic-Pack`
// (#323).
func TestEinSetsymbolHatBenannteFassungen(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
set info style:
	symbol:
		variation:
			name: invertedcommon
			border radius: 0.10
			fill type: solid
			fill color:   rgb(255,255,255)
			border color: rgb(0,0,0)
		variation:
			name: uncommon
			fill type: linear gradient
			fill color 1: rgb(224,224,224)
			fill color 2: rgb(84,84,84)
			border color 1: rgb(0,0,0)
			border color 2: rgb(1,2,3)
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(stil.Symbolfassungen) != 2 {
		t.Fatalf("%d Fassungen, erwartet zwei: %+v", len(stil.Symbolfassungen), stil.Symbolfassungen)
	}

	voll := stil.Symbolfassungen[0]
	if voll.Name != "invertedcommon" || voll.Randbreite != 0.10 || voll.Fuellart != "solid" {
		t.Errorf("die erste Fassung: %+v", voll)
	}

	if voll.Fuellfarbe != "rgb(255,255,255)" || voll.Randfarbe != "rgb(0,0,0)" {
		t.Errorf("die Farben der ersten Fassung: %+v", voll)
	}

	verlauf := stil.Symbolfassungen[1]
	if verlauf.Fuellart != "linear gradient" || verlauf.Fuellfarbe2 != "rgb(84,84,84)" {
		t.Errorf("die zweite Fassung: %+v", verlauf)
	}

	// **Ohne Angabe gilt 0,05** — MSEs `border_radius(0.05)`.
	if verlauf.Randbreite != 0.05 {
		t.Errorf("die Vorgabe der Randbreite ist %g, erwartet 0,05", verlauf.Randbreite)
	}
}

// **Das Spiel erklärt die Grundfassungen — unter `default set style:`.**
//
// Nicht am `set field:`: das Feld erklärt, *was* ein Setsymbol ist; wie es
// aussieht, steht im Stil, und der des Spiels ist die Vorgabe für jede Vorlage
// (`magic.mse-game/set_fields`, Zeilen 535–571).
//
// **Wer sie am Feld suchte, fände keine** — und `symbol_variation` kennte dann
// die Fassung `common` nicht, die jede Karte braucht
// (#323).
func TestDasSpielErklaertDieGrundfassungen(t *testing.T) {
	pfad := probespielpaket(t, `mse version: 2.0.0
default set style:
	#### ein Kommentar ist kein Feld
	symbol:
		max aspect ratio: 2.5
		variation:
			name: common
			border radius: 0.10
			fill type: solid
			fill color: rgb(0,0,0)
		variation:
			name: rare
			fill type: linear gradient
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	fassungen := paket.Spielsymbolfassungen()
	if len(fassungen) != 2 {
		t.Fatalf("%d Fassungen, erwartet zwei: %+v", len(fassungen), fassungen)
	}

	if fassungen[0].Name != "common" || fassungen[1].Name != "rare" {
		t.Errorf("die Namen: %q und %q", fassungen[0].Name, fassungen[1].Name)
	}
}

// **Die Gestalt einer `.mse-symbol` kommt über die Bildquelle.**
//
// Der Planbauer öffnet keine Pakete; er fragt die Quelle, die er ohnehin hat.
// **Was nicht lesbar ist, ist leer** — ein Fehler hier bräche den ganzen Plan
// für eine Datei, die auf den meisten Karten gar nicht gebraucht wird (§8).
func TestDieGestaltKommtUeberDieBildquelle(t *testing.T) {
	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "probe.mse-set")

	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\npart:\n\ttype: shape\n\tpoint:\n\t\tposition: (0.2,0.2)\n"
	if err := os.WriteFile(filepath.Join(paket, "z.mse-symbol"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// **Eine Datei, die der Zerteiler abweist** — ein leerer Schlüssel ist
	// der letzte Grund, eine MSE-Datei zurückzuweisen (Vermerk 0014).
	kaputt := "mse version: 2.0.0\n: ohne Schlüssel\n"
	if err := os.WriteFile(filepath.Join(paket, "kaputt.mse-symbol"), []byte(kaputt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle := msepaket.NeueBildquelle(wurzel)

	if teile := quelle.Symbolgestalt("/probe.mse-set/z.mse-symbol").Teile; len(teile) != 1 {
		t.Errorf("%d Teile, erwartet einen", len(teile))
	}

	for _, verweis := range []string{
		"/probe.mse-set/gibtsnicht.mse-symbol",
		"/probe.mse-set/kaputt.mse-symbol",
		"kein Paketpfad",
	} {
		if teile := quelle.Symbolgestalt(verweis).Teile; len(teile) != 0 {
			t.Errorf("%q gibt %d Teile, erwartet keine", verweis, len(teile))
		}
	}
}
