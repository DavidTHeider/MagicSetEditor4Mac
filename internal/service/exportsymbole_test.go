package service_test

import (
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`symbols_to_html` legt Manasymbole neben die Ausgabe** —
// #658.
//
//	FOR_EACH(sym, symbols) {
//	  String filename = symbol_font.name() + _("-") + clean_filename(sym.text) + _(".png");
//	  map<String,wxSize>::iterator it = ei.exported_images.find(filename);
//	  if (it == ei.exported_images.end()) {
//	    Image img = symbol_font.getImage(size, sym);
//	    … img.SaveFile(fn.GetFullPath());
//	    it = ei.exported_images.insert(make_pair(filename, wxSize(…))).first;
//	  }
//	  html += _("<img src='") + filename + _("' alt='") + html_escape(sym.text)
//	       +  _("' width='")  + (String() << it->second.x)
//	       +  _("' height='") + (String() << it->second.y) + _("'>");
//	}
//
// `src/script/functions/export.cpp`, Zeilen 183–205, Fassung f9d9356 vom
// 2021-07-06.
//
// `magic-planesculptors` ruft es zweimal, `magic-spoiler` ebenso — dort
// allerdings nur, wenn die Wahlmöglichkeit `mana_symbols` gesetzt ist; sonst
// überschreibt die Vorlage den Namen mit `to_html`.

// symbolschriftprobe ist eine Symbolschrift mit drei Zeichen.
type symbolschriftprobe struct {
	schriften map[string]msepaket.Symbolschrift
}

func (s symbolschriftprobe) Symbolschrift(name string) (msepaket.Symbolschrift, bool) {
	schrift, da := s.schriften[name]

	return schrift, da
}

// symbolrastererprobe gibt jedem Symbol ein Bild fester Größe.
type symbolrastererprobe struct {
	// gerastert zählt, wie oft wirklich gerastert wurde — für die Merkliste.
	gerastert []string

	// masse hält die Maße des letzten Aufrufs. **Sie stehen hier, weil eine
	// Attrappe, die ihr Argument nicht ansieht, jede Rechnung darüber
	// durchwinkt** — die Gegenprobe „Bildgröße weggelassen" war grün (§10).
	masse render.Symbolmasse

	fehlt string
}

func (r *symbolrastererprobe) SymbolAlsBild(
	symbol render.Symbol, masse render.Symbolmasse,
) (bild *image.NRGBA, befund string) {
	if r.fehlt != "" {
		return nil, r.fehlt
	}

	r.gerastert = append(r.gerastert, symbol.Text)
	r.masse = masse

	return image.NewNRGBA(image.Rect(0, 0, 11, 13)), ""
}

func manaschrift() symbolschriftprobe {
	return symbolschriftprobe{schriften: map[string]msepaket.Symbolschrift{
		"mana": {
			Name:              "mana",
			Bildgroesse:       135,
			AbstandWaagerecht: 3,
			AbstandSenkrecht:  2,
			Symbole: []msepaket.Symbol{
				{Code: "W", Bild: "w.png"},
				{Code: "2", Bild: "2.png"},
				{Code: "T/2", Bild: "t2.png"},
				{Code: "é", Bild: "e.png"},

				// **Codes, die im `alt` maskiert werden müssen.** Kein Pack
				// schreibt sie so; geprüft wird hier MSEs `html_escape`, und
				// dessen fünf Zweige erreicht nur, wer sie herbeiführt.
				{Code: "<", Bild: "lt.png"},
				{Code: ">", Bild: "gt.png"},
				{Code: "&", Bild: "amp.png"},
				{Code: "'", Bild: "apos.png"},
				{Code: "\"", Bild: "quot.png"},
				{Code: "\u0001", Bild: "esc.png"},

				// **Und zwei, die den Dateinamen fordern.**
				{Code: " X", Bild: "sp.png"},
				{Code: "/", Bild: "slash.png"},
			},
		},
	}}
}

func mitSymbolen(mitschrift *beigabenmitschrift, rasterer *symbolrastererprobe) ausfuhrprobe {
	ausfuehrer := service.NeuerVorlagenausfuehrer()
	ausfuehrer.MitSymbolen(manaschrift(), rasterer)

	return ausfuhrprobe{
		ausfuehrer: ausfuehrer,
		mittel: service.Ausfuhrmittel{
			Beigaben:    mitschrift,
			Verzeichnis: "ausfuhr-files",
		},
	}
}

func TestSymbolsToHTML(t *testing.T) {
	t.Parallel()

	mitschrift := &beigabenmitschrift{}
	rasterer := &symbolrastererprobe{}

	ausgabe, _, err := mitSymbolen(mitschrift, rasterer).
		Fuehre(t.Context(),
			vorlageMit(`symbols_to_html("2WW", symbol_font: "mana")`),
			eineKarte(), nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	erwartet := `<img src='mana-2.png' alt='2' width='11' height='13'>` +
		`<img src='mana-W.png' alt='W' width='11' height='13'>` +
		`<img src='mana-W.png' alt='W' width='11' height='13'>`

	if ausgabe != erwartet {
		t.Errorf("Ausgabe:\n  %s\nerwartet:\n  %s", ausgabe, erwartet)
	}

	// **Jedes Symbol wird einmal gerastert und einmal geschrieben** — das ist
	// die Merkliste `ei.exported_images`. Ohne sie ginge `W` zweimal über die
	// Platte, und über 300 Karten wäre das dasselbe Bild sechshundertmal.
	if strings.Join(rasterer.gerastert, "|") != "2|W" {
		t.Errorf("gerastert wurde %q", rasterer.gerastert)
	}

	if len(mitschrift.gelegt) != 2 {
		t.Errorf("abgelegt wurden %d Dateien: %v", len(mitschrift.gelegt), mitschrift.gelegt)
	}

	// **Die Maße kommen aus zwei Quellen** — `symbol_font_size` aus dem
	// Aufruf, der Rest aus dem Paket. Ohne diese Zusicherung prüfte kein Test,
	// in welcher Größe gerastert wird, und das ist die Zahl, die am Ende im
	// `width`-Attribut steht.
	erwarteteMasse := render.Symbolmasse{
		Groesse: msescript.HTMLVorgabegroesse, Bildgroesse: 135,
		AbstandWaagerecht: 3, AbstandSenkrecht: 2,
	}

	if rasterer.masse != erwarteteMasse {
		t.Errorf("Maße %+v, erwartet %+v", rasterer.masse, erwarteteMasse)
	}
}

// **`symbol_font_size` schlägt die Vorgabe durch.**
func TestDieSymbolgroesseKommtAusDemAufruf(t *testing.T) {
	t.Parallel()

	rasterer := &symbolrastererprobe{}

	if _, _, err := mitSymbolen(&beigabenmitschrift{}, rasterer).
		Fuehre(t.Context(),
			vorlageMit(`symbols_to_html("W", symbol_font: "mana", symbol_font_size: 40)`),
			eineKarte(), nil, nil); err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if rasterer.masse.Groesse != 40 {
		t.Errorf("Größe %v, erwartet 40", rasterer.masse.Groesse)
	}
}

// **Der Dateiname wird gesäubert, der `alt`-Text maskiert.**
//
//	bool is_filename_char(Char c) {
//	  return isAlnum(c) || c == _(' ') || c == _('_') || c == _('-') || c == _('.');
//	}
//
// `src/util/file_utils.cpp`, Zeilen 47–72, Fassung f9d9356 vom 2021-07-06.
// Der Schrägstrich in `T/2` fällt weg — sonst schriebe die Vorlage in ein
// Unterverzeichnis, das niemand gemeint hat.
func TestDerSymboldateinameWirdGesaeubert(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, zug, erwartet string }{
		{
			name:     "ein Schrägstrich fällt weg",
			zug:      "T/2",
			erwartet: `<img src='mana-T2.png' alt='T/2' width='11' height='13'>`,
		},
		{
			// **Über 0x7f wird im `alt` zur Zahlenmarke**, im Dateinamen
			// bleibt es stehen: `isAlnum` ist bei wxWidgets unicodefähig.
			name:     "ein Umlaut bleibt im Namen und wird im alt maskiert",
			zug:      "é",
			erwartet: `<img src='mana-é.png' alt='&#233;' width='11' height='13'>`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ausgabe, _, err := mitSymbolen(&beigabenmitschrift{}, &symbolrastererprobe{}).
				Fuehre(t.Context(),
					vorlageMit(`symbols_to_html("`+fall.zug+`", symbol_font: "mana")`),
					eineKarte(), nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v", err)
			}

			if ausgabe != fall.erwartet {
				t.Errorf("%s, erwartet %s", ausgabe, fall.erwartet)
			}
		})
	}
}

// **`to_html` benutzt denselben Weg** — das ist der Grund, warum der Rückruf
// aus #723
// existiert.
//
//	} else if (is_substr(str, i, _("/sym"))) {
//	  … ret += symbols_to_html(symbols, *symbol_font, symbol_size);
//
// `export.cpp`, Zeilen 230–236.
func TestToHTMLSetztDieselbenBilder(t *testing.T) {
	t.Parallel()

	mitschrift := &beigabenmitschrift{}

	ausgabe, _, err := mitSymbolen(mitschrift, &symbolrastererprobe{}).
		Fuehre(t.Context(),
			vorlageMit(`to_html("Kosten: <sym>2W</sym>", symbol_font: "mana")`),
			eineKarte(), nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	erwartet := `Kosten: <span class="symbol">` +
		`<img src='mana-2.png' alt='2' width='11' height='13'>` +
		`<img src='mana-W.png' alt='W' width='11' height='13'>` +
		`</span>`

	if ausgabe != erwartet {
		t.Errorf("Ausgabe:\n  %s\nerwartet:\n  %s", ausgabe, erwartet)
	}

	if len(mitschrift.gelegt) != 2 {
		t.Errorf("abgelegt wurden %d Dateien", len(mitschrift.gelegt))
	}
}

// **Ohne die Fähigkeit gibt es den Namen nicht** — dieselbe Regel wie bei
// `write_image_file` (#216) und den drei Geschwistern (#657).
func TestOhneSymbolquelleGibtEsDenNamenNicht(t *testing.T) {
	t.Parallel()

	_, _, err := probelauf(service.Ausfuhrmittel{
		Beigaben: &beigabenmitschrift{}, Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(), vorlageMit(`symbols_to_html("W", symbol_font: "mana")`),
			eineKarte(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "nicht bekannt") {
		t.Fatalf("Fehler = %v", err)
	}
}

// **Was schiefgehen kann, wird benannt.**
//
// Eine halbe Ausfuhr ist eine, die jemand für vollständig hält (§8) — anders
// als bei einer Karte, wo eine Ebene weniger brauchbar bleibt.
func TestWasBeimSymbolschreibenSchiefgeht(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, skript, teil string
		rasterer           *symbolrastererprobe
	}{
		{
			name:     "eine unbekannte Symbolschrift",
			skript:   `symbols_to_html("W", symbol_font: "gibtsnicht")`,
			teil:     "gibtsnicht",
			rasterer: &symbolrastererprobe{},
		},
		{
			name:     "ohne symbol_font",
			skript:   `symbols_to_html("W")`,
			teil:     "symbol_font",
			rasterer: &symbolrastererprobe{},
		},
		{
			name:     "ein Symbolbild, das nicht daliegt",
			skript:   `symbols_to_html("W", symbol_font: "mana")`,
			teil:     "liegt nicht vor",
			rasterer: &symbolrastererprobe{fehlt: "w.png liegt nicht vor"},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := mitSymbolen(&beigabenmitschrift{}, fall.rasterer).
				Fuehre(t.Context(), vorlageMit(fall.skript), eineKarte(), nil, nil)
			if err == nil {
				t.Fatal("der Aufruf ging durch")
			}

			if !strings.Contains(err.Error(), fall.teil) {
				t.Errorf("Meldung %q nennt %q nicht", err.Error(), fall.teil)
			}
		})
	}
}

// **Ein Zeichen, das die Schrift nicht führt, wird übersprungen** — so wie
// beim Zeichnen einer Karte.
//
//	// unknown code, draw single character as text
//	//out.push_back(DrawableSymbol(text.substr(pos, 1), _(""), defaultSymbol()));
//	pos += 1;
//
// `SymbolFont::split`, `src/data/symbol_font.cpp`, Zeilen 192–231, Fassung
// f9d9356 vom 2021-07-06 — der `push_back` ist auskommentiert.
func TestEinUnbekanntesZeichenWirdUebersprungen(t *testing.T) {
	t.Parallel()

	ausgabe, _, err := mitSymbolen(&beigabenmitschrift{}, &symbolrastererprobe{}).
		Fuehre(t.Context(), vorlageMit(`symbols_to_html("WxW", symbol_font: "mana")`),
			eineKarte(), nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if strings.Count(ausgabe, "<img") != 2 {
		t.Errorf("%s — erwartet zwei Bilder", ausgabe)
	}
}
