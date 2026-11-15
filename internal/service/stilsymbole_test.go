package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Feld mit `symbol font:` zeichnet Bilder, keinen Text**
// (#173,
// #240).
//
// # Der Befund
//
// `magic-m15.mse-style` legt die Manakosten so an:
//
//	casting cost:
//	    font:        { name: MPlantin, size: 15 }
//	    symbol font: { name: magic-mana-large, size: 15 }
//
// Beide Blöcke stehen da, und **der zweite gewinnt**: `2R` ist keine
// Zeichenfolge, sondern zwei Bilder. Bis hierher zeichnete dieses Programm den
// Text `2R` in MPlantin — lesbar, aber keine Manakosten.
//
// # Wie ein Text in Symbole zerfällt
//
// **Der längste Code zuerst.** `magic-mana-large` führt neben `R` auch `2/R`
// und `V/W`; wer zeichenweise ginge, machte aus `2/R` drei Symbole statt einem.
func TestEinSymbolfeldWirdZuEinemSymbolzug(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "10", Oben: "20", Breite: "60", Hoehe: "15",
		Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
		ImmerSymbole:  true,
	}

	vorlage := symbolvorlage(feld)

	plan, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "2R"), vorlage)
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gebaut := ebenen(t, plan)

	if len(gebaut) != 1 || gebaut[0].Art != render.ArtSymbolzug {
		t.Fatalf("%d Ebenen, erwartet ein Symbolzug: %+v", len(gebaut), gebaut)
	}

	// **Der Rahmen ist der des Feldes, nicht der eines Symbols.** Wo die
	// Symbole darin sitzen, rechnet der Rasterer aus.
	if gebaut[0].Rahmen != (render.Rechteck{Links: 10, Oben: 20, Breite: 60, Hoehe: 15}) {
		t.Errorf("der Rahmen ist %+v", gebaut[0].Rahmen)
	}

	if len(gebaut[0].Abschnitte) != 1 {
		t.Fatalf("%d Abschnitte, erwartet einer", len(gebaut[0].Abschnitte))
	}

	symbole := gebaut[0].Abschnitte[0].Symbole
	if len(symbole) != 2 {
		t.Fatalf("%d Symbole, erwartet zwei: %+v", len(symbole), symbole)
	}

	if symbole[0].Bild.Beschreibung() != "/mana.mse-symbol-font/2.png" {
		t.Errorf("erstes Symbol %q", symbole[0].Bild.Beschreibung())
	}

	if symbole[1].Bild.Beschreibung() != "/mana.mse-symbol-font/r.png" {
		t.Errorf("zweites Symbol %q", symbole[1].Bild.Beschreibung())
	}

	// **Die Größe der Symbolschrift, nicht die des Feldes.**
	if groesse := gebaut[0].Abschnitte[0].Symbolmasse.Groesse; groesse != 15 {
		t.Errorf("die Symbolgröße ist %g, erwartet 15", groesse)
	}
}

// **Der Code, der in der Schrift zuerst steht, gewinnt.** `2/R` ist damit ein
// Symbol und nicht drei — weil die Schrift ihn vor `2` und `R` führt.
func TestDerErsteCodeDerSchriftGewinnt(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
		Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
		ImmerSymbole:  true,
	}

	plan, befunde := gedeutetMitVorlage(t,
		kartemit("casting cost", "2/RR"), symbolvorlage(feld))
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gebaut := ebenen(t, plan)

	if len(gebaut) != 1 {
		t.Fatalf("%d Ebenen, erwartet eine: %+v", len(gebaut), gebaut)
	}

	symbole := gebaut[0].Abschnitte[0].Symbole
	if len(symbole) != 2 || symbole[0].Bild.Beschreibung() != "/mana.mse-symbol-font/hybrid.png" {
		t.Errorf("die Symbole sind %+v", symbole)
	}
}

// **Die Ausrichtung des Feldes reist mit, ausgerechnet wird sie nicht hier.**
//
// `casting cost` in `magic-m15.mse-style` trägt `alignment: middle right`, und
// der Kasten ist breiter als die Kosten. Bis zum 2026-08-17 rechnete der
// Planbauer den Versatz selbst — waagerecht, und senkrecht gar nicht. Wo der
// Zug landet, prüft jetzt `TestEinSymbolzugSitztDortWoDasFeldIhnHinstellt` in
// `internal/render`: erst dort ist bekannt, wie breit die Symbole sind (0022).
func TestEinSymbolzugTraegtDieAusrichtungSeinesFeldes(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
		Ausrichtung:   "middle right",
		Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
		ImmerSymbole:  true,
	}

	plan, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "2R"), symbolvorlage(feld))
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gebaut := ebenen(t, plan)

	if gebaut[0].Satz.Ausrichtung != render.RechtsMitte {
		t.Errorf("die Ausrichtung ist %q, erwartet %q",
			gebaut[0].Satz.Ausrichtung, render.RechtsMitte)
	}
}

// **Ein Zeichen, das die Schrift nicht führt, wird benannt.**
//
// Nicht übersprungen: eine Manakostenangabe, der ein Symbol fehlt, sieht aus
// wie eine billigere Karte (§8).
func TestEinUnbekanntesSymbolWirdBenannt(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
		Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
		ImmerSymbole:  true,
	}

	_, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "RX"), symbolvorlage(feld))

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet war einer: %v", len(befunde), befunde)
	}
}

// **Ohne die Schrift zur Hand bleibt es ein Befund, keine leere Ebene.**
func TestEineFehlendeSymbolschriftWirdBenannt(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
		Symbolschrift: msepaket.Symbolangabe{Name: "gibtsnicht", Groesse: "15"},
		ImmerSymbole:  true,
	}

	_, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "R"),
		planvorlage(bildstil(feld), nil))

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde: %v", len(befunde), befunde)
	}
}

// symbolvorlage baut eine Vorlage mit einer kleinen Symbolschrift.
//
// **Codes verschiedener Länge**, denn genau daran entscheidet sich, ob der
// Zerleger den längsten zuerst nimmt.
func symbolvorlage(feld msepaket.Feldstil) service.Vorlage {
	return planvorlage(bildstil(feld), map[string]msepaket.Symbolschrift{
		// **Die zusammengesetzten Codes stehen vorn**, so wie in
		// `magic-mana-large`: `V/W` in Zeile 21, `R` in Zeile 788. Die
		// Reihenfolge ist die Entscheidung des Schriftautors, und sie
		// entscheidet die Zerlegung (#293).
		"mana": {Name: "mana", Symbole: []msepaket.Symbol{
			{Code: "2/R", Bild: "hybrid.png"},
			{Code: "R", Bild: "r.png"},
			{Code: "2", Bild: "2.png"},
		}},
	})
}

func planvorlage(
	stil msepaket.Stil, schriften map[string]msepaket.Symbolschrift,
) service.Vorlage {
	return service.Vorlage{Stil: stil, Symbolschriften: schriften}
}

// gedeutetMitVorlage baut den Plan aus einer eigens gefüllten Vorlage.
func gedeutetMitVorlage(
	t *testing.T, karte *domain.CardRow, vorlage service.Vorlage,
) (*render.Plan, []domain.Befund) {
	t.Helper()

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	return plan, befunde
}

// **Ein Symbol schreibt die Gruppe auf sein Bild, die `draw text:` nennt.**
//
//	if (sym->draw_text >= 0 && sym->draw_text < (int)results.size()) {
//	  out.push_back(DrawableSymbol(results.str(), results.str(sym->draw_text), *sym));
//	}
//
// `SymbolFont::split`, `src/data/symbol_font.cpp`, Zeilen 205–221, Fassung
// f9d9356 vom 2021-07-06.
//
// **Daran hängt jede Zahl in den Manakosten** — `code: ([XYZ0-9])/W` mit
// `draw text: 1` zeichnet den Kreis und schreibt die Ziffer hinein
// (#336).
func TestEinSymbolSchreibtSeineGruppe(t *testing.T) {
	faelle := []struct {
		name      string
		symbol    msepaket.Symbol
		text      string
		erwartet  string
		erwartetA render.Ausrichtung
	}{
		{
			name: "die erste Gruppe eines Musters",
			symbol: msepaket.Symbol{
				Code: `([0-9])/W`, Bild: "kreis.png", Ausdruck: true, Textgruppe: 1,
				Textausrichtung: "bottom right",
			},
			text: "8/W", erwartet: "8", erwartetA: render.RechtsUnten,
		},
		{
			// **Ohne `draw text:` steht nichts auf dem Bild.**
			name:   "ohne draw text",
			symbol: msepaket.Symbol{Code: "W", Bild: "w.png", Textgruppe: -1},
			text:   "W", erwartet: "",
		},
		{
			// **Ohne Muster ist die einzige Gruppe der Code selbst** —
			// `sym->draw_text >= 0 ? sym->code : ""` (Zeile 221).
			name:   "ohne Muster gilt der Code",
			symbol: msepaket.Symbol{Code: "W", Bild: "w.png", Textgruppe: 0},
			text:   "W", erwartet: "W",
		},
		{
			// **Eine Nummer außerhalb der Gruppen zeichnet nichts** — MSE
			// prüft sie gegen `results.size()`.
			name: "eine Nummer, die es nicht gibt",
			symbol: msepaket.Symbol{
				Code: `([0-9])/W`, Bild: "kreis.png", Ausdruck: true, Textgruppe: 7,
			},
			text: "8/W", erwartet: "",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			schrift := msepaket.Symbolschrift{
				Name: "probe", Symbole: []msepaket.Symbol{fall.symbol},
			}

			symbole := service.SymbolbilderFuerTest(t.Context(), fall.text, schrift, nil)
			if len(symbole) != 1 {
				t.Fatalf("%d Symbole, erwartet eines", len(symbole))
			}

			if symbole[0].Text != fall.erwartet {
				t.Errorf("Text %q, erwartet %q", symbole[0].Text, fall.erwartet)
			}

			if symbole[0].Textsatz.Ausrichtung != fall.erwartetA {
				t.Errorf("Ausrichtung %q, erwartet %q",
					symbole[0].Textsatz.Ausrichtung, fall.erwartetA)
			}
		})
	}
}

// **Eine Gruppe, die nicht getroffen hat, ist leer und kein Fehler.**
//
// MSE liest sie über `results.str(i)`, und das gibt dort eine leere
// Zeichenkette. `([0-9])?([XY])` trifft mit `X` nur die zweite Gruppe.
func TestEineNichtGetroffeneSymbolgruppeIstLeer(t *testing.T) {
	schrift := msepaket.Symbolschrift{Name: "probe", Symbole: []msepaket.Symbol{{
		Code: `([0-9])?([XY])`, Bild: "kreis.png", Ausdruck: true, Textgruppe: 1,
	}}}

	symbole := service.SymbolbilderFuerTest(t.Context(), "X", schrift, nil)
	if len(symbole) != 1 {
		t.Fatalf("%d Symbole, erwartet eines", len(symbole))
	}

	if symbole[0].Text != "" {
		t.Errorf("Text %q, erwartet leer", symbole[0].Text)
	}
}
