package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die erste Fassung gewinnt, nicht die letzte**
// (#548).
//
//	FOR_EACH(v, style->variations) {
//	  if (v->name == variation) {
//	    return make_intrusive<SymbolToImage>(value, filename, value->last_update, v);
//	  }
//	}
//
// `src/script/functions/image.cpp`, Zeilen 164–170, Fassung f9d9356 vom
// 2021-07-06 — **`return` in der Schleife.**
//
// # Warum die Reihenfolge so ist
//
// Ein Stilpaket **erbt** die Fassungen des Spiels und hängt seine eigenen
// dahinter: `set_info_style.cloneFrom(game->default_set_style)` steht **vor**
// `REFLECT(set_info_style)` (`src/data/stylesheet.cpp`, Zeilen 106–113), und
// `Reader::handle(name, vector<T>&)` macht `emplace_back` ohne zu leeren
// (`src/util/io/reader.hpp`, Zeilen 218–226).
//
// **Bei gleichem Namen gewinnt damit die Fassung des Spiels**, und die des
// Stilpakets ist ein Duplikat dahinter, das nie zum Zug kommt.
//
// # Am Artefakt belegt
//
// Die elf `special`-Karten von `White` erscheinen in MSE2s eigener Ausgabe
// **weiß** — die Randfarbe der **Spiel**-Fassung `rgb(255,255,255)`. Die
// Fassung des Stilpakets hätte `rgb(0,0,0)` ergeben.
func TestDieErsteSymbolfassungGewinnt(t *testing.T) {
	t.Parallel()

	// Die Reihenfolge ist die des Lesers: Spiel zuerst, Stilpaket dahinter.
	fassungen := []msepaket.Symbolfassung{
		{Name: "special", Feld: "symbol", Randfarbe: "rgb(255,255,255)"},
		{Name: "special", Feld: "symbol", Randfarbe: "rgb(0,0,0)"},
	}

	gewaehlt, gefunden := service.ErsteFassung(fassungen, "symbol", "special")
	if !gefunden {
		t.Fatal("keine Fassung gefunden")
	}

	if gewaehlt.Randfarbe != "rgb(255,255,255)" {
		t.Errorf("Randfarbe %q, erwartet \"rgb(255,255,255)\" — die spätere "+
			"Fassung hat gewonnen", gewaehlt.Randfarbe)
	}
}

// TestEineFassungGehoertZuIhremFeld ist der zweite Befund aus #548.
//
// **MSE sucht die Fassung am Stil des konkreten Felds** (`value->fieldP`,
// `image.cpp:161`), wir verflachten sie über alle Felder.
//
// `magic.mse-game/set_fields` erklärt `symbol:` **und** `masterpiece symbol:`,
// und sie unterscheiden sich an genau einer Stelle: `watermark` ist unter
// `symbol:` `fill color: rgb(0,0,0)`, unter `masterpiece symbol:`
// `rgb(255,255,255)`.
func TestEineFassungGehoertZuIhremFeld(t *testing.T) {
	t.Parallel()

	fassungen := []msepaket.Symbolfassung{
		{Name: "watermark", Feld: "symbol", Fuellfarbe: "rgb(0,0,0)"},
		{Name: "watermark", Feld: "masterpiece symbol", Fuellfarbe: "rgb(255,255,255)"},
	}

	faelle := []struct {
		feld     string
		erwartet string
	}{
		{"symbol", "rgb(0,0,0)"},
		{"masterpiece symbol", "rgb(255,255,255)"},
	}

	for _, fall := range faelle {
		t.Run(fall.feld, func(t *testing.T) {
			t.Parallel()

			gewaehlt, gefunden := service.ErsteFassung(fassungen, fall.feld, "watermark")
			if !gefunden {
				t.Fatalf("keine Fassung für %q", fall.feld)
			}

			if gewaehlt.Fuellfarbe != fall.erwartet {
				t.Errorf("Füllfarbe %q, erwartet %q — die Fassung eines anderen "+
					"Feldes hat gewonnen", gewaehlt.Fuellfarbe, fall.erwartet)
			}
		})
	}
}

// TestEineUnbekannteFassungWirdNichtGefunden prüft den leeren Fall.
//
// **MSE wirft dort** — `Variation of symbol not found`
// (`src/script/functions/image.cpp`, Zeile 171). Bei uns wird daraus ein
// Befund; eine erfundene Fassung zeichnete ein Symbol in falschen Farben, und
// niemand sähe warum (§8).
func TestEineUnbekannteFassungWirdNichtGefunden(t *testing.T) {
	t.Parallel()

	fassungen := []msepaket.Symbolfassung{
		{Name: "common", Feld: "symbol"},
	}

	faelle := []struct {
		name       string
		feld, wahl string
	}{
		{"die Fassung gibt es nicht", "", "gibtesnicht"},
		{"das Feld passt nicht", "masterpiece symbol", "common"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if _, gefunden := service.ErsteFassung(fassungen, fall.feld, fall.wahl); gefunden {
				t.Error("eine Fassung gefunden, erwartet keine")
			}
		})
	}
}
