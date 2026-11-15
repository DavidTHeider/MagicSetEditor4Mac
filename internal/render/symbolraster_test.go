package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// viereck ist eine Form aus vier Ecken, in Anteilen der Zeichnung.
func viereck(links, oben, rechts, unten float64) render.Symbolteil {
	return render.Symbolteil{
		Art: render.TeilForm,
		Punkte: []render.Symbolpunkt{
			{Stelle: render.Punkt{X: links, Y: oben}},
			{Stelle: render.Punkt{X: rechts, Y: oben}},
			{Stelle: render.Punkt{X: rechts, Y: unten}},
			{Stelle: render.Punkt{X: links, Y: unten}},
		},
	}
}

// **Ein gerastertes Symbol kennt drei Bereiche: außen, Rand, innen.**
//
//	SymbolSet point = data[1] ? (data[0] ? SYMBOL_BORDER : SYMBOL_OUTSIDE)
//	                          : SYMBOL_INSIDE;
//
// `filter_symbol`, `src/render/symbol/filter.cpp`, Zeilen 32–36, Fassung
// f9d9356 vom 2021-07-06. Der Rasterer gibt keine Farbe her — die Farbe
// entscheidet erst der Filter der Variante
// (#323).
func TestEinSymbolHatDreiBereiche(t *testing.T) {
	maske := render.RastereSymbol(
		[]render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)}, 100, 100, 0.05, false)

	faelle := []struct {
		name     string
		x, y     int
		erwartet render.Symbolbereich
	}{
		{name: "in der Mitte", x: 50, y: 50, erwartet: render.SymbolInnen},
		{name: "in der Ecke", x: 2, y: 2, erwartet: render.SymbolAussen},

		// **Der Rand liegt außerhalb der Form.** Sein Strich sitzt mittig auf
		// der Kante, und die innere Hälfte überschreibt das Innere.
		{name: "gerade außerhalb der Kante", x: 23, y: 50, erwartet: render.SymbolRand},
		{name: "gerade innerhalb der Kante", x: 27, y: 50, erwartet: render.SymbolInnen},
	}

	for _, fall := range faelle {
		if haben := maske.Bei(fall.x, fall.y); haben != fall.erwartet {
			t.Errorf("%s (%d,%d): %v, erwartet %v",
				fall.name, fall.x, fall.y, haben, fall.erwartet)
		}
	}
}

// **`combine:` verrechnet eine Form mit dem, was schon dasteht.**
//
// Die sechs Arten stehen in `combineSymbolShape`
// (`src/render/symbol/viewer.cpp`, Zeilen 213–251). Geprüft wird die Stelle,
// an der sich zwei Formen überlappen, und eine, an der nur die zweite liegt.
//
// Fassung f9d9356 vom 2021-07-06.
func TestDieKombinationsartenEinerForm(t *testing.T) {
	faelle := []struct {
		name             string
		art              render.Kombinationsart
		ueberlappt, nur2 render.Symbolbereich
	}{
		{
			name: "überdecken", art: render.KombinationUeberdecken,
			ueberlappt: render.SymbolInnen, nur2: render.SymbolInnen,
		},
		{
			name: "vereinen", art: render.KombinationVereinen,
			ueberlappt: render.SymbolInnen, nur2: render.SymbolInnen,
		},
		{
			// **Abziehen nimmt weg**, wo die zweite Form liegt.
			name: "abziehen", art: render.KombinationAbziehen,
			ueberlappt: render.SymbolAussen, nur2: render.SymbolAussen,
		},
		{
			// **Schneiden behält nur das Gemeinsame.**
			name: "schneiden", art: render.KombinationSchneiden,
			ueberlappt: render.SymbolInnen, nur2: render.SymbolAussen,
		},
		{
			// **Unterschied nimmt das Gemeinsame weg und lässt den Rest.**
			name: "unterschied", art: render.KombinationUnterschied,
			ueberlappt: render.SymbolAussen, nur2: render.SymbolInnen,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			zweite := viereck(0.4, 0.4, 0.9, 0.9)
			zweite.Kombination = fall.art

			// **Die obere Form steht zuerst.** MSE zeichnet die Teile von
			// hinten nach vorn (`FOR_EACH_CONST_REVERSE`, Zeilen 156 und 206)
			// — was in der Datei oben steht, wird zuletzt gezeichnet und
			// verrechnet sich mit dem darunter.
			maske := render.RastereSymbol([]render.Symbolteil{
				zweite, viereck(0.1, 0.1, 0.6, 0.6),
			}, 100, 100, 0, false)

			if haben := maske.Bei(50, 50); haben != fall.ueberlappt {
				t.Errorf("wo sich beide überlappen: %v, erwartet %v", haben, fall.ueberlappt)
			}

			if haben := maske.Bei(80, 80); haben != fall.nur2 {
				t.Errorf("wo nur die zweite liegt: %v, erwartet %v", haben, fall.nur2)
			}
		})
	}
}

// **Eine Drehsymmetrie vervielfältigt ihre Kinder.**
//
//	double a = i * 2 * M_PI / copies;
//	Matrix2D rot(cos(a),-sin(a), sin(a),cos(a));
//	multiply = rot * old_m;
//	origin = old_o + (s->center - s->center * rot) * old_m;
//
// `combineSymbolPart`, `src/render/symbol/viewer.cpp`, Zeilen 163–179.
//
// Fassung f9d9356 vom 2021-07-06.
func TestEineDrehsymmetrieVervielfaeltigtIhreKinder(t *testing.T) {
	// Ein Klotz oben links, viermal um die Mitte gedreht — er muss danach in
	// allen vier Ecken liegen.
	symmetrie := render.Symbolteil{
		Art:        render.TeilSymmetrie,
		Spiegelart: render.SpiegelungDrehung,
		Kopien:     4,
		Mitte:      render.Punkt{X: 0.5, Y: 0.5},
		Griff:      render.Punkt{X: 0.5, Y: 0},
		Teile:      []render.Symbolteil{viereck(0.1, 0.1, 0.3, 0.3)},
	}

	maske := render.RastereSymbol([]render.Symbolteil{symmetrie}, 100, 100, 0, false)

	ecken := [][2]int{{20, 20}, {80, 20}, {80, 80}, {20, 80}}
	for _, ecke := range ecken {
		if haben := maske.Bei(ecke[0], ecke[1]); haben != render.SymbolInnen {
			t.Errorf("die Ecke (%d,%d) ist %v, erwartet innen", ecke[0], ecke[1], haben)
		}
	}

	if haben := maske.Bei(50, 50); haben != render.SymbolAussen {
		t.Errorf("die Mitte ist %v, erwartet außen", haben)
	}
}

// **Eine Gruppe ändert nichts — sie fasst nur zusammen.**
func TestEineGruppeAendertNichts(t *testing.T) {
	gruppe := render.Symbolteil{
		Art:   render.TeilGruppe,
		Teile: []render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)},
	}

	maske := render.RastereSymbol([]render.Symbolteil{gruppe}, 100, 100, 0, false)
	if haben := maske.Bei(50, 50); haben != render.SymbolInnen {
		t.Errorf("die Mitte ist %v, erwartet innen", haben)
	}
}

// **`border` malt die ganze Form in der Randfarbe.**
//
//	// draw border as interior
//	drawSymbolShape(shape, nullptr, &border, 0, 255, false, false);
//
// `combineSymbolShape`, `src/render/symbol/viewer.cpp`, Zeile 247, Fassung
// f9d9356 vom 2021-07-06.
//
// **Der Name führt in die Irre**: gezeichnet wird nicht ein Umriss, sondern die
// Fläche — nur eben in den Randpuffer, und damit in der Randfarbe. Wer daraus
// einen Ring machen will, zieht die Innenform mit `subtract` wieder ab.
func TestUmrandenMachtDieGanzeFormZumRand(t *testing.T) {
	form := viereck(0.25, 0.25, 0.75, 0.75)
	form.Kombination = render.KombinationUmranden

	maske := render.RastereSymbol([]render.Symbolteil{form}, 100, 100, 0.05, false)

	if mitte := maske.Bei(50, 50); mitte != render.SymbolRand {
		t.Errorf("die Mitte ist %v, erwartet Rand", mitte)
	}

	if aussen := maske.Bei(5, 5); aussen != render.SymbolAussen {
		t.Errorf("außerhalb ist %v, erwartet außen", aussen)
	}
}

// **Eine Spiegelsymmetrie spiegelt jede zweite Kopie.**
//
//	int copies = s->kind == SYMMETRY_REFLECTION ? s->copies / 2 * 2 : s->copies;
//	…
//	Matrix2D rot(cos(a+b),sin(a+b), sin(a+b),-cos(a+b));
//
// Zeilen 155 und 191. **Die Zahl der Kopien wird gerade** — eine Spiegelung
// braucht immer ein Gegenüber.
func TestEineSpiegelsymmetrieSpiegeltJedeZweiteKopie(t *testing.T) {
	symmetrie := render.Symbolteil{
		Art:        render.TeilSymmetrie,
		Spiegelart: render.SpiegelungAchse,
		Kopien:     2,
		Mitte:      render.Punkt{X: 0.5, Y: 0.5},

		// Der Griff steht senkrecht: gespiegelt wird an der Waagerechten.
		Griff: render.Punkt{X: 0, Y: 0.5},
		Teile: []render.Symbolteil{viereck(0.1, 0.1, 0.3, 0.3)},
	}

	maske := render.RastereSymbol([]render.Symbolteil{symmetrie}, 100, 100, 0, false)

	if oben := maske.Bei(20, 20); oben != render.SymbolInnen {
		t.Errorf("die Vorlage oben links ist %v, erwartet innen", oben)
	}

	if unten := maske.Bei(20, 80); unten != render.SymbolInnen {
		t.Errorf("das Spiegelbild unten links ist %v, erwartet innen", unten)
	}
}

// **Eine Symmetrie ohne Kopien zeichnet nichts.**
//
// `copies / 2 * 2` macht aus einer einzelnen Kopie bei einer Spiegelung null,
// und eine Schleife von null bis unter null läuft bei MSE gar nicht erst.
func TestEineSymmetrieOhneKopienZeichnetNichts(t *testing.T) {
	symmetrie := render.Symbolteil{
		Art:        render.TeilSymmetrie,
		Spiegelart: render.SpiegelungAchse,
		Kopien:     1,
		Mitte:      render.Punkt{X: 0.5, Y: 0.5},
		Teile:      []render.Symbolteil{viereck(0.1, 0.1, 0.9, 0.9)},
	}

	maske := render.RastereSymbol([]render.Symbolteil{symmetrie}, 20, 20, 0, false)

	for y := range 20 {
		for x := range 20 {
			if maske.Bei(x, y) != render.SymbolAussen {
				t.Fatalf("(%d,%d) ist %v, erwartet außen", x, y, maske.Bei(x, y))
			}
		}
	}
}

// **Ein hohes Bild rückt das Symbol in die Mitte.**
//
//	viewer.setZoom(height);
//	viewer.setOrigin(Vector2D(-(height-width) * 0.5, 0));
//
// `render_symbol`, Zeilen 30–34. Die Zeichnung wird über die **längere** Seite
// vergrößert und quer dazu mittig gesetzt.
func TestEinHohesBildRuecktDasSymbolInDieMitte(t *testing.T) {
	maske := render.RastereSymbol(
		[]render.Symbolteil{viereck(0.4, 0.4, 0.6, 0.6)}, 40, 100, 0, false)

	// Die Mitte der Fläche ist die Mitte des Symbols.
	if mitte := maske.Bei(20, 50); mitte != render.SymbolInnen {
		t.Errorf("die Mitte ist %v, erwartet innen", mitte)
	}

	// **Oben und unten bleibt es leer** — dort steht nichts von der Zeichnung.
	if oben := maske.Bei(20, 5); oben != render.SymbolAussen {
		t.Errorf("oben ist %v, erwartet außen", oben)
	}

	// **Und quer dazu dasselbe** — ein breites Bild rückt das Symbol nach unten.
	breit := render.RastereSymbol(
		[]render.Symbolteil{viereck(0.4, 0.4, 0.6, 0.6)}, 100, 40, 0, false)

	if mitte := breit.Bei(50, 20); mitte != render.SymbolInnen {
		t.Errorf("die Mitte des breiten Bildes ist %v, erwartet innen", mitte)
	}

	if links := breit.Bei(5, 20); links != render.SymbolAussen {
		t.Errorf("links ist %v, erwartet außen", links)
	}
}

// **Beim Unterschied wird der Rand hinterher gelöscht.**
//
//	if (clear) {
//	  border->SetPen(*wxTRANSPARENT_PEN);
//	  border->SetBrush(Color(0, …, 0));
//	  border->SetLogicalFunction(wxCOPY);
//	  border->DrawPolygon(…);
//	}
//
// `drawSymbolShape`, `src/render/symbol/viewer.cpp`, Zeilen 271–279, Fassung
// f9d9356 vom 2021-07-06. Innerhalb der Form bleibt vom Randstrich nichts —
// sonst zöge ein Unterschied eine Linie mitten durch das, was er weggenommen
// hat.
func TestBeimUnterschiedWirdDerRandGeloescht(t *testing.T) {
	zweite := viereck(0.4, 0.4, 0.9, 0.9)
	zweite.Kombination = render.KombinationUnterschied

	maske := render.RastereSymbol([]render.Symbolteil{
		zweite, viereck(0.1, 0.1, 0.6, 0.6),
	}, 100, 100, 0.05, false)

	// Mitten im weggenommenen Stück steht nichts — kein Rand.
	if mitte := maske.Bei(50, 50); mitte != render.SymbolAussen {
		t.Errorf("die gemeinsame Fläche ist %v, erwartet außen", mitte)
	}
}
