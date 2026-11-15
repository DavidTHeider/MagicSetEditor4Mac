package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Farbfeldstil trägt einen Radius und vier Randbreiten.**
//
//	IMPLEMENT_REFLECTION(ColorStyle) {
//	  REFLECT_BASE(Style);
//	  REFLECT(radius);
//	  REFLECT(left_width);
//	  REFLECT(right_width);
//	  REFLECT(top_width);
//	  REFLECT(bottom_width);
//	  REFLECT(combine);
//	}
//
// `src/data/field/color.cpp`, Zeilen 66–74, Fassung f9d9356 vom 2021-07-06.
//
// **Daran hängt der Kartenrand.** `magic-m15.mse-style` gibt `border color` die
// Breiten 17/17/17/18 und einen Radius von 18 — ohne sie füllte das Feld die
// ganze Karte schwarz
// (#270).
func TestEinFarbfeldstilTraegtRadiusUndRandbreiten(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
card style:
	border color:
		left: 0
		top: 0
		width: 375
		height: 523
		radius: 18
		left width:   17
		right width:  17
		top width:    17
		bottom width: 18
	ohne rand:
		left: 0
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

	stil, _ := paket.Stilmodell()
	if len(stil.Felder) != 2 {
		t.Fatalf("%d Feldstile", len(stil.Felder))
	}

	rand := stil.Felder[0].Rand
	gelesen := []string{
		strings.TrimSpace(stil.Felder[0].Ecke), strings.TrimSpace(rand.Links),
		strings.TrimSpace(rand.Rechts), strings.TrimSpace(rand.Oben),
		strings.TrimSpace(rand.Unten),
	}

	for stelle, erwartet := range []string{"18", "17", "17", "17", "18"} {
		if gelesen[stelle] != erwartet {
			t.Errorf("Angabe %d ist %q, erwartet %q", stelle, gelesen[stelle], erwartet)
		}
	}

	// **Ohne Angabe bleibt es leer.** Der Planbauer nimmt dann MSEs Vorgabe —
	// Radius 0 und keine Rahmung.
	if leer := stil.Felder[1]; leer.Ecke != "" || leer.Rand != (msepaket.Randangabe{}) {
		t.Errorf("ohne Angabe: %q, %+v", leer.Ecke, leer.Rand)
	}
}

// **Eine Wahl eines Farbfeldes trägt einen Namen und eine Farbe.**
//
//	IMPLEMENT_REFLECTION(ColorField::Choice) {
//	  REFLECT_IF_READING_SINGLE_VALUE {
//	    REFLECT_NAMELESS(name);
//	    auto col = parse_color(name);
//	    if (col) color = *col;
//	  } else {
//	    REFLECT(name);
//	    REFLECT(color);
//	  }
//	}
//
// `src/data/field/color.cpp`, Zeilen 41–54, Fassung f9d9356 vom 2021-07-06.
// **Beide Formen**: `choice: white` und ein Block mit `name:` und `color:`.
//
// Die **erste** Wahl ist die Vorgabe des Feldes, und daran hing der schwarze
// Kartenrand: `set field: border color` schreibt sie als Block
// (#270).
func TestEineWahlEinesFarbfeldesTraegtEineFarbe(t *testing.T) {
	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: color
	name: border color
	choice:
		name: black
		color: rgb(0,0,0)
	choice:
		name: gold
		color: rgb(200,180,0)
set field:
	type: color
	name: copyright text color
	choice: white
	choice: black
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

	spiel, _ := paket.Spielmodell()

	if len(spiel.Fields) != 1 || len(spiel.SetFields) != 1 {
		t.Fatalf("%d Kartenfelder, %d Setfelder", len(spiel.Fields), len(spiel.SetFields))
	}

	if haben := spiel.Fields[0].Wahlfarben; len(haben) != 2 || haben[0] != "rgb(0,0,0)" {
		t.Errorf("die Wahlfarben sind %v", haben)
	}

	// **Die einzeilige Form trägt die Farbe im Namen** — `choice: white` ist
	// bei MSE `name` *und* `color`, weil `parse_color` auf den Namen angewandt
	// wird.
	if haben := spiel.SetFields[0].Wahlfarben; len(haben) != 2 || haben[0] != "white" {
		t.Errorf("die Wahlfarben des Setfeldes sind %v", haben)
	}
}

// **Auch eine Stiloption vom Typ `color` führt ihre Wahlfarben.**
func TestEineStiloptionFuehrtIhreWahlfarben(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
styling field:
	type: color
	name: text color
	choice:
		name: black
		color: rgb(0,0,0)
	choice: white
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

	stil, _ := paket.Stilmodell()
	if len(stil.Stiloptionen) != 1 {
		t.Fatalf("%d Stiloptionen", len(stil.Stiloptionen))
	}

	haben := stil.Stiloptionen[0].Wahlfarben
	if len(haben) != 2 || haben[0] != "rgb(0,0,0)" || haben[1] != "white" {
		t.Errorf("die Wahlfarben sind %v", haben)
	}
}

// **Ein Verlauf ohne `center x:` läuft von der Mitte in die Ecke.**
//
//	GradientSymbolFilter::GradientSymbolFilter()
//	  : center_x(0.5), center_y(0.5)
//	  , end_x(1), end_y(1)
//	{}
//
// `src/render/symbol/filter.cpp`, Zeilen 138–142, Fassung f9d9356 vom
// 2021-07-06.
//
// # Was das gekostet hat
//
// **Keine** der Symbolfassungen von `magic.mse-game` nennt `center x:` oder
// `end x:`. Mit null als Vorgabe wird MSEs `len` eins, `t` überall null — und
// jeder Verlauf war eine **einfarbige** Fläche in `fill color 1`. Nachgemessen
// am Seltenheitssymbol von *Firebrand Incinerator*: gleichmäßig
// `rgb(214,196,94)`, wo MSE einen Verlauf zeichnet
// (#270).
func TestEinVerlaufOhneAngabeLaeuftVonDerMitte(t *testing.T) {
	pfad := probespielpaket(t, `mse version: 2.0.0
default set style:
	symbol:
		variation:
			name: rare
			fill type: linear gradient
			fill color 1: rgb(214,196,94)
			fill color 2: rgb(95,84,40)
		variation:
			name: eigen
			fill type: linear gradient
			center x: 0
			center y: 0.25
			end x: 0.5
			end y: 0.75
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
		t.Fatalf("%d Fassungen", len(fassungen))
	}

	if ohne := fassungen[0]; ohne.MitteX != 0.5 || ohne.MitteY != 0.5 ||
		ohne.EndeX != 1 || ohne.EndeY != 1 {
		t.Errorf("ohne Angabe: Mitte (%v,%v), Ende (%v,%v)",
			ohne.MitteX, ohne.MitteY, ohne.EndeX, ohne.EndeY)
	}

	// **Eine eigene Angabe gilt, auch wenn sie null ist.** `center x: 0` ist
	// eine Aussage und nicht das Fehlen einer.
	if mit := fassungen[1]; mit.MitteX != 0 || mit.MitteY != 0.25 ||
		mit.EndeX != 0.5 || mit.EndeY != 0.75 {
		t.Errorf("mit Angabe: Mitte (%v,%v), Ende (%v,%v)",
			mit.MitteX, mit.MitteY, mit.EndeX, mit.EndeY)
	}
}
