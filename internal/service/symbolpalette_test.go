package service_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Palette eines Feldes** —
// #564.
//
// # Wofür
//
// Im Eingabefeld für die Manakosten steht `2RR`. Auf der Karte werden daraus
// Symbole; im Feld bleiben es Buchstaben, und wer nicht weiß, dass `T` das
// Tap-Symbol ist, findet es nicht. MSE bietet dafür ein Menü an, und es steht
// **im Paket** (`insert symbol menu:`).
//
// # Was `service` beisteuert und was nicht
//
// Es beantwortet die drei Fragen, die eine Umgebung brauchen: **welche**
// Symbolschrift dieses Feld benutzt (der Name darf ein Skript sein), ob der
// Code roh oder in `<sym>…</sym>` eingefügt wird (`always symbol:`), und
// **welche Fassung** eines Symbols gerade gilt (`enabled:`).
//
// Gerastert wird nicht hier: jeder Eintrag trägt einen Plan, und den malt, wer
// einen Rasterer hat (§2).
func TestDiePaletteEinesFeldes(t *testing.T) {
	vorlage := paletteprobe()

	eintraege, befunde, err := service.NeuerPlanbauer().
		Symbolpalette(t.Context(), &domain.CardRow{ID: 1}, "kosten", vorlage)
	if err != nil {
		t.Fatalf("Palette: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %v", befunde)
	}

	if len(eintraege) != 4 {
		t.Fatalf("%d Einträge: %+v", len(eintraege), eintraege)
	}

	faelle := []struct {
		art     msepaket.Menueart
		code    string
		mitPlan bool
	}{
		{msepaket.MenueCode, "W", true},
		{msepaket.MenueCode, "U", true},
		{msepaket.MenueTrenner, "", false},
		{msepaket.MenueUntermenue, "", false},
	}

	for stelle, fall := range faelle {
		eintrag := eintraege[stelle]
		if eintrag.Art != fall.art || eintrag.Code != fall.code ||
			(eintrag.Plan != nil) != fall.mitPlan {
			t.Errorf("Eintrag %d: %+v", stelle, eintrag)
		}
	}

	// **Ein Untermenü trägt seine Kinder, und die haben wieder Pläne.**
	if kinder := eintraege[3].Kinder; len(kinder) != 1 || kinder[0].Plan == nil {
		t.Errorf("das Untermenü ist %+v", eintraege[3])
	}
}

// **`always symbol:` entscheidet, was eingefügt wird.**
//
//	String code = style().symbol_font.font->insertSymbolCode(id);
//	if (!style().always_symbol) {
//	  code = _("<sym>") + code + _("</sym>");
//	}
//
// `TextValueEditor::onCommand`, `src/gui/value/text.cpp`, Zeilen 589–593,
// Fassung f9d9356 vom 2021-07-06. Bei den Manakosten ist der ganze Wert
// Symbole; im Regeltext steht `<sym>T</sym>` mitten im Fließtext.
func TestDiePaletteSagtObSieAuszeichnet(t *testing.T) {
	vorlage := paletteprobe()

	for _, fall := range []struct {
		feld  string
		roh   bool
		warum string
	}{
		{"kosten", true, "`always symbol: true` fügt roh ein"},
		{"regeltext", false, "ohne die Angabe wird in `<sym>` gefasst"},
	} {
		t.Run(fall.feld, func(t *testing.T) {
			_, _, err := service.NeuerPlanbauer().
				Symbolpalette(t.Context(), &domain.CardRow{ID: 1}, fall.feld, vorlage)
			if err != nil {
				t.Fatalf("Palette: %v", err)
			}

			if gab := service.FuegtRohEin(vorlage.Stil, fall.feld); gab != fall.roh {
				t.Errorf("%v statt %v — %s", gab, fall.roh, fall.warum)
			}
		})
	}
}

// **Ein Feld ohne Symbolschrift hat keine Palette** — und das ist kein Befund.
//
// Über `magic-m15.mse-style` führen die allerwenigsten Feldstile einen
// `symbol font:`-Block; eine Meldung je Feld ohne einen wäre eine Liste, die
// niemand liest (§8).
func TestEinFeldOhneSymbolschriftHatKeinePalette(t *testing.T) {
	eintraege, befunde, err := service.NeuerPlanbauer().
		Symbolpalette(t.Context(), &domain.CardRow{ID: 1}, "name", paletteprobe())
	if err != nil {
		t.Fatalf("Palette: %v", err)
	}

	if eintraege != nil || len(befunde) > 0 {
		t.Errorf("Einträge %+v, Befunde %v", eintraege, befunde)
	}
}

// paletteprobe baut einen Stil mit einem Symbolfeld, einem Textfeld mit
// Symbolschrift und einem ohne.
func paletteprobe() service.Vorlage {
	return service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 140,
			Felder: []msepaket.Feldstil{
				{
					Feld: "kosten", Links: "10", Oben: "10", Breite: "50", Hoehe: "20",
					Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "12"},
					ImmerSymbole:  true,
				},
				{
					Feld: "regeltext", Links: "10", Oben: "40", Breite: "50", Hoehe: "20",
					Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "12"},
				},
				{Feld: "name", Links: "10", Oben: "70", Breite: "50", Hoehe: "20"},
			},
		},
		Symbolschriften: map[string]msepaket.Symbolschrift{
			"probe-symbole": {
				Name:        "probe-symbole",
				Bildgroesse: 12,
				Symbole: []msepaket.Symbol{
					{Code: "W", Bild: "weiss.png"},
					{Code: "U", Bild: "blau.png"},
					{Code: "1/2", Bild: "halb.png"},
				},
				Einfuegemenue: []msepaket.Menueeintrag{
					{Art: msepaket.MenueCode, Name: "W"},
					{Art: msepaket.MenueCode, Name: "U"},
					{Art: msepaket.MenueTrenner},
					{Art: msepaket.MenueUntermenue, Name: "halb", Kinder: []msepaket.Menueeintrag{
						{Art: msepaket.MenueCode, Name: "1/2"},
					}},
				},
			},
		},
	}
}

// **Ein Code ohne Namen wird weggelassen.**
//
// Drei stehen so im Pack: `space-psi-xerent` schreibt ein nacktes `item:`
// dorthin, wo ein `type: line` hingehörte, `magic-embossedletters` und
// `magic-mana-large-extra-udelude` je ein `item:` mit einem Leerzeichen
// dahinter. MSE liest beides leer — `trim_left("  ")` ist `""`
// (`src/util/io/reader.cpp`, Zeile 224, Fassung f9d9356 vom 2021-07-06) — und
// baut daraus einen Menüeintrag ohne Bild, der nichts einfügt. Eine
// Schaltfläche ohne Aufschrift und ohne Wirkung ist keine.
func TestEinCodeOhneNamenFaelltWeg(t *testing.T) {
	vorlage := paletteprobe()
	schrift := vorlage.Symbolschriften["probe-symbole"]
	schrift.Einfuegemenue = []msepaket.Menueeintrag{
		{Art: msepaket.MenueCode, Name: "W"},
		{Art: msepaket.MenueCode, Name: "  "},
		{Art: msepaket.MenueCode, Name: ""},
	}
	vorlage.Symbolschriften["probe-symbole"] = schrift

	eintraege, _, err := service.NeuerPlanbauer().
		Symbolpalette(t.Context(), &domain.CardRow{ID: 1}, "kosten", vorlage)
	if err != nil {
		t.Fatalf("Palette: %v", err)
	}

	if len(eintraege) != 1 || eintraege[0].Code != "W" {
		t.Errorf("%+v", eintraege)
	}
}

// **Ein Menü, aus dem nichts übrig bleibt, ist keins.**
//
// `nil` und nicht die leere Liste — derselbe Grund wie beim Leser: MSE fragt
// `if (insert_symbol_menu)`, bevor es irgendetwas tut.
func TestEinMenueOhneBrauchbarenEintragIstKeins(t *testing.T) {
	vorlage := paletteprobe()
	schrift := vorlage.Symbolschriften["probe-symbole"]
	schrift.Einfuegemenue = []msepaket.Menueeintrag{{Art: msepaket.MenueCode, Name: ""}}
	vorlage.Symbolschriften["probe-symbole"] = schrift

	eintraege, _, err := service.NeuerPlanbauer().
		Symbolpalette(t.Context(), &domain.CardRow{ID: 1}, "kosten", vorlage)
	if err != nil {
		t.Fatalf("Palette: %v", err)
	}

	if eintraege != nil {
		t.Errorf("%+v", eintraege)
	}
}

// **Ein Code, den die Schrift nicht kennt, bekommt keinen Plan.**
//
// Ein Menü nennt Codes, die eine Bedingung gerade weggenommen hat — `enabled:`
// hat sie beim Auflösen der Fassungen entfernt. MSE zeigt dann einen Eintrag
// ohne Kachel (`InsertSymbolMenu::makeMenuItem`, `src/data/symbol_font.cpp`,
// Zeilen 508–518, Fassung f9d9356 vom 2021-07-06).
func TestEinUnbekannterCodeBekommtKeinenPlan(t *testing.T) {
	vorlage := paletteprobe()
	schrift := vorlage.Symbolschriften["probe-symbole"]
	schrift.Einfuegemenue = []msepaket.Menueeintrag{
		{Art: msepaket.MenueCode, Name: "kennt keiner"},
	}
	vorlage.Symbolschriften["probe-symbole"] = schrift

	eintraege, _, err := service.NeuerPlanbauer().
		Symbolpalette(t.Context(), &domain.CardRow{ID: 1}, "kosten", vorlage)
	if err != nil {
		t.Fatalf("Palette: %v", err)
	}

	if len(eintraege) != 1 || eintraege[0].Plan != nil {
		t.Errorf("%+v", eintraege)
	}
}

// **`label:` schlägt den Namen, und ohne beides wird großgeschrieben.**
//
//	if (m.label.empty())
//	  m.label.default_ = tr(*symbol_font_for_reading(), "menu_item", m.name, capitalize);
//
// `after_reading(InsertSymbolMenu&)`, `src/data/symbol_font.cpp`, Zeilen
// 545–549, Fassung f9d9356 vom 2021-07-06. Der Rückfall ist `capitalize` —
// **jedes Wort** groß, anders als `name_to_caption` für ein Feld.
func TestDieAufschriftEinesPaletteneintrags(t *testing.T) {
	vorlage := paletteprobe()
	schrift := vorlage.Symbolschriften["probe-symbole"]
	schrift.Einfuegemenue = []msepaket.Menueeintrag{
		{Art: msepaket.MenueCode, Name: "W", Beschriftung: "Weißes Mana"},
		{Art: msepaket.MenueCode, Name: "U"},
		{Art: msepaket.MenueEigen, Name: "beliebige zahl"},
	}
	vorlage.Symbolschriften["probe-symbole"] = schrift

	eintraege, _, err := service.NeuerPlanbauer().
		Symbolpalette(t.Context(), &domain.CardRow{ID: 1}, "kosten", vorlage)
	if err != nil {
		t.Fatalf("Palette: %v", err)
	}

	erwartet := []string{"Weißes Mana", "U", "Beliebige Zahl"}
	for stelle, soll := range erwartet {
		if gab := eintraege[stelle].Beschriftung; gab != soll {
			t.Errorf("%q statt %q", gab, soll)
		}
	}
}

// **Ein Feld, das der Stil nicht kennt, hat keine Palette** — und die
// Symbolschrift, die nicht vorliegt, ist ein Befund.
func TestEinePaletteOhneStilUndOhneSchrift(t *testing.T) {
	bauer := service.NeuerPlanbauer()

	eintraege, _, err := bauer.Symbolpalette(
		t.Context(), &domain.CardRow{ID: 1}, "gibtsnicht", paletteprobe())
	if err != nil {
		t.Fatalf("Palette: %v", err)
	}

	if eintraege != nil {
		t.Errorf("ein unbekanntes Feld ergab %+v", eintraege)
	}

	ohneSchrift := paletteprobe()
	ohneSchrift.Symbolschriften = nil

	_, befunde, err := bauer.Symbolpalette(
		t.Context(), &domain.CardRow{ID: 1}, "kosten", ohneSchrift)
	if err != nil {
		t.Fatalf("Palette: %v", err)
	}

	if len(befunde) != 1 {
		t.Errorf("Befunde: %v", befunde)
	}
}

// **Eine Schrift ohne Einfügemenü ergibt keine Palette** — und keinen Befund:
// eine Schrift, die nichts zum Einfügen anbietet, ist kein Mangel.
func TestEineSchriftOhneMenueErgibtKeinePalette(t *testing.T) {
	vorlage := paletteprobe()
	schrift := vorlage.Symbolschriften["probe-symbole"]
	schrift.Einfuegemenue = nil
	vorlage.Symbolschriften["probe-symbole"] = schrift

	eintraege, befunde, err := service.NeuerPlanbauer().
		Symbolpalette(t.Context(), &domain.CardRow{ID: 1}, "kosten", vorlage)
	if err != nil {
		t.Fatalf("Palette: %v", err)
	}

	if eintraege != nil || len(befunde) > 0 {
		t.Errorf("Einträge %+v, Befunde %v", eintraege, befunde)
	}
}

// **Ein abgebrochener Lauf gibt seinen Fehler weiter** (§7).
func TestEineAbgebrocheneePaletteMeldetDenAbbruch(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, _, err := service.NeuerPlanbauer().
		Symbolpalette(ctx, &domain.CardRow{ID: 1}, "kosten", paletteprobe()); err == nil {
		t.Error("der Abbruch kam nicht durch")
	}
}

// **Der Weg zur Palette geht über den Stil der Karte** — denselben, den die
// Vorschau nimmt.
//
// **Ohne Behelf.** Fällt der Stil aus, gibt es keine Symbolschrift und damit
// nichts anzubieten; der Behelfsplan zeichnet Feldwerte als Zeilen und kennt
// keine Symbole.
func TestDieKartenpaletteGehtUeberDenStilDerKarte(t *testing.T) {
	karte := &domain.CardRow{ID: 1}

	t.Run("ohne Stilverzeichnis gibt es nichts", func(t *testing.T) {
		eintraege, roh, befunde, err := service.NeuerPlanbauer().
			Kartenpalette(t.Context(), karte, "kosten")
		if err != nil {
			t.Fatalf("Palette: %v", err)
		}

		if eintraege != nil || roh || len(befunde) > 0 {
			t.Errorf("%+v / %v / %v", eintraege, roh, befunde)
		}
	})

	t.Run("ohne passenden Stil kommen nur seine Befunde", func(t *testing.T) {
		bauer := service.NeuerPlanbauer().MitStilen(festerStil{
			befunde: []domain.Befund{{Meldung: "kein Stilpaket"}},
		})

		eintraege, _, befunde, err := bauer.Kartenpalette(t.Context(), karte, "kosten")
		if err != nil {
			t.Fatalf("Palette: %v", err)
		}

		if eintraege != nil || len(befunde) != 1 {
			t.Errorf("%+v / %v", eintraege, befunde)
		}
	})

	t.Run("mit Stil kommt die Palette samt `always symbol:`", func(t *testing.T) {
		bauer := service.NeuerPlanbauer().
			MitStilen(festerStil{vorlage: paletteprobe(), da: true})

		eintraege, roh, _, err := bauer.Kartenpalette(t.Context(), karte, "kosten")
		if err != nil {
			t.Fatalf("Palette: %v", err)
		}

		if len(eintraege) != 4 || !roh {
			t.Errorf("%d Einträge, roh=%v", len(eintraege), roh)
		}

		// **Und `regeltext` zeichnet aus** — dasselbe Menü, andere Antwort.
		_, rohText, _, err := bauer.Kartenpalette(t.Context(), karte, "regeltext")
		if err != nil {
			t.Fatalf("Palette: %v", err)
		}

		if rohText {
			t.Error("ein Feld ohne `always symbol:` gilt als roh")
		}
	})
}
