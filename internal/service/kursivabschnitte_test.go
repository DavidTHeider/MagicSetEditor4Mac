package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Flavortext wird kursiv, nicht entfernt.**
//
// # Der Befund
//
// Seit dem 2026-08-16 werden Auszeichnungen **entfernt** und benannt. Der
// Regeltext einer Karte aus `magic-m15` sieht so aus:
//
//	Fliegend\n<i-flavor>Er kam mit dem Sturm.</i-flavor>
//
// Bis hierher stand der Flavortext aufrecht neben dem Regeltext. MSE2 erkennt
// die Auszeichnung über ein **Präfix** — `is_tag(text, pos, "<i")`,
// `src/render/text/element.cpp`, Zeile 80, Fassung f9d9356 vom 2021-07-06 —,
// weshalb `<i>`, `<i-flavor>` und `<i-auto>` dieselbe Wirkung haben
// (#312).
func TestDerFlavortextWirdKursiv(t *testing.T) {
	plan := planMitText(t, "Fliegend. <i-flavor>Er kam mit dem Sturm.</i-flavor>")

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText {
			continue
		}

		erwartet := []render.Abschnitt{
			{Text: "Fliegend. "},
			{Text: "Er kam mit dem Sturm.", Kursiv: true},
		}

		if len(ebene.Abschnitte) != len(erwartet) {
			t.Fatalf("%d Abschnitte, erwartet %d: %+v",
				len(ebene.Abschnitte), len(erwartet), ebene.Abschnitte)
		}

		for nummer, abschnitt := range ebene.Abschnitte {
			if abschnitt.Text != erwartet[nummer].Text ||
				abschnitt.Kursiv != erwartet[nummer].Kursiv {
				t.Errorf("Abschnitt %d ist %+v, erwartet %+v",
					nummer, abschnitt, erwartet[nummer])
			}
		}

		// **Und der Klartext ist derselbe wie zuvor** — was nur misst, merkt
		// von den Schnitten nichts.
		if klar := ebene.Klartext(); klar != "Fliegend. Er kam mit dem Sturm." {
			t.Errorf("der Klartext lautet %q", klar)
		}

		// **Die Kursivschrift steht im Satz**, sonst hätte der Rasterer keinen
		// Namen, unter dem er sie suchen könnte.
		if ebene.Satz.Kursivschrift != "Helvetica-Oblique" {
			t.Errorf("die Kursivschrift ist %q, erwartet Helvetica-Oblique",
				ebene.Satz.Kursivschrift)
		}

		return
	}

	t.Fatal("der Plan führt keine Textebene")
}

// **Eine verschachtelte Auszeichnung endet erst mit der äußeren.**
//
// MSE zählt sie mit — `italic += 1` beim Öffnen, `-= 1` beim Schließen
// (`element.cpp`, Zeilen 80–81). Wer beim ersten `</i>` aufhörte, ließe den
// Rest aufrecht stehen.
func TestEineVerschachtelteKursivstelleEndetMitDerAeusseren(t *testing.T) {
	plan := planMitText(t, "<i>a<i>b</i>c</i>d")

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText {
			continue
		}

		erwartet := []render.Abschnitt{{Text: "abc", Kursiv: true}, {Text: "d"}}

		if len(ebene.Abschnitte) != len(erwartet) {
			t.Fatalf("%d Abschnitte, erwartet %d: %+v",
				len(ebene.Abschnitte), len(erwartet), ebene.Abschnitte)
		}

		for nummer, abschnitt := range ebene.Abschnitte {
			if abschnitt.Text != erwartet[nummer].Text ||
				abschnitt.Kursiv != erwartet[nummer].Kursiv {
				t.Errorf("Abschnitt %d ist %+v, erwartet %+v",
					nummer, abschnitt, erwartet[nummer])
			}
		}

		return
	}

	t.Fatal("der Plan führt keine Textebene")
}

// planMitText baut einen Plan über ein einziges Textfeld.
func planMitText(t *testing.T, text string) *render.Plan {
	t.Helper()

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "10", Oben: "10", Breite: "300", Hoehe: "100",
			Schrift: msepaket.Schrift{
				Name: "Helvetica", KursivName: "Helvetica-Oblique", Groesse: "12",
			},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: text},
	}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte,
		service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	return plan
}

// **Ein Text aus lauter Auszeichnungen ergibt einen leeren Abschnitt.**
//
// `<i></i>` kommt vor — ein Flavortext, den jemand gelöscht und dessen
// Auszeichnung er stehen gelassen hat. Ohne Abschnitt hätte die Ebene nichts
// zu setzen, und [render.SetzeAbschnitte] gäbe eine Höhe von null; das Feld
// verschwände samt seiner senkrechten Ausrichtung. Ein leerer Abschnitt ist
// eine leere Zeile, und das ist, was dort steht.
func TestEinTextAusLauterAuszeichnungenErgibtEinenLeerenAbschnitt(t *testing.T) {
	plan := planMitText(t, "<i></i>")

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText {
			continue
		}

		if len(ebene.Abschnitte) != 1 || ebene.Abschnitte[0].Text != "" {
			t.Errorf("die Abschnitte sind %+v, erwartet ein leerer", ebene.Abschnitte)
		}

		return
	}

	t.Fatal("der Plan führt keine Textebene")
}

// **`<sym>` mitten im Regeltext wird zu Symbolbildern.**
//
// # Der Befund
//
// Ein Feld war **entweder** ganz Text **oder** ganz Symbole; das trifft die
// Manakosten und ließ jedes `<sym>T</sym>` im Regeltext als Buchstaben stehen.
// MSE zerlegt den Zug in Elemente, und ein Element innerhalb von `<sym>` wird
// ein `SymbolTextElement` (`src/render/text/element.cpp`, Zeilen 82–83 und
// 230–232, Fassung f9d9356 vom 2021-07-06)
// (#312).
func TestEinSymbolMittenImRegeltextWirdZuBildern(t *testing.T) {
	plan := planMitSymbolen(t, "Tappe: <sym>WU</sym> hinzufügen.")

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText {
			continue
		}

		if len(ebene.Abschnitte) != 3 {
			t.Fatalf("%d Abschnitte, erwartet drei: %+v",
				len(ebene.Abschnitte), ebene.Abschnitte)
		}

		mitte := ebene.Abschnitte[1]
		if len(mitte.Symbole) != 2 {
			t.Fatalf("der mittlere Abschnitt trägt %d Symbole, erwartet zwei: %+v",
				len(mitte.Symbole), mitte)
		}

		if mitte.Symbole[0].Bild.Beschreibung() != "/probe-symbole.mse-symbol-font/w.png" {
			t.Errorf("das erste Symbol zeigt auf %q", mitte.Symbole[0].Bild.Beschreibung())
		}

		// **Die Größe kommt aus `symbol font: size:`**, nicht aus der
		// Schriftgröße des Textes daneben.
		if mitte.Symbolmasse.Groesse != 15 {
			t.Errorf("die Symbolgröße ist %.0f, erwartet 15", mitte.Symbolmasse.Groesse)
		}

		// **Und der Text ringsum bleibt Text.**
		if ebene.Abschnitte[0].Text != "Tappe: " || len(ebene.Abschnitte[0].Symbole) != 0 {
			t.Errorf("der erste Abschnitt ist %+v", ebene.Abschnitte[0])
		}

		return
	}

	t.Fatal("der Plan führt keine Textebene")
}

// **Ein unbekannter Code wird übersprungen und benannt.**
//
//	// unknown code, draw single character as text
//	//out.push_back(DrawableSymbol(text.substr(pos, 1), _(""), defaultSymbol()));
//	pos += 1;
//
// `SymbolFont::split`, `src/data/symbol_font.cpp`, Zeilen 192–231, Fassung
// f9d9356 vom 2021-07-06 — der `push_back` ist auskommentiert.
//
// **Hier stand bis zum 2026-08-19 „lässt den Abschnitt Text bleiben"**, mit
// dem Satz „sichtbar falsch statt unsichtbar fehlend" daneben. Der Satz gilt
// weiter, nur zeigt er in die andere Richtung: aus **allen** Manasymbolen
// Buchstaben zu machen, weil eines fehlt, ist die größere Entstellung
// (#517).
func TestEinUnbekannterSymbolcodeWirdUebersprungen(t *testing.T) {
	stil := symbolstil()

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "<sym>WZ</sym>"},
	}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte,
		service.Vorlage{Stil: stil, Symbolschriften: probesymbole()})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	// **Das bekannte Symbol steht da, das unbekannte Zeichen nicht.**
	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText && len(ebene.Abschnitte[0].Symbole) != 1 {
			t.Errorf("%d Symbole, erwartet eines: %+v",
				len(ebene.Abschnitte[0].Symbole), ebene.Abschnitte[0])
		}
	}

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "Z") {
			return
		}
	}

	t.Errorf("kein Befund nennt das unbekannte Zeichen: %v", befunde)
}

// **Ohne Symbolschrift am Feld bleibt `<sym>` Text, und der Grund steht da.**
func TestOhneSymbolschriftBleibtSymText(t *testing.T) {
	stil := symbolstil()
	stil.Felder[0].Symbolschrift = msepaket.Symbolangabe{}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "<sym>W</sym>"},
	}}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte,
		service.Vorlage{Stil: stil, Symbolschriften: probesymbole()})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "keine Symbolschrift") {
			return
		}
	}

	t.Errorf("kein Befund nennt die fehlende Symbolschrift: %v", befunde)
}

// **Eine benannte Symbolschrift, die nicht vorliegt, ebenso.**
func TestEineFehlendeSymbolschriftLaesstSymText(t *testing.T) {
	stil := symbolstil()
	stil.Felder[0].Symbolschrift.Name = "gibtsnicht"

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "<sym>W</sym>"},
	}}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte,
		service.Vorlage{Stil: stil, Symbolschriften: probesymbole()})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "gibtsnicht") {
			return
		}
	}

	t.Errorf("kein Befund nennt die fehlende Schrift: %v", befunde)
}

// symbolstil ist ein Stil mit einem Textfeld, das eine Symbolschrift nennt.
func symbolstil() msepaket.Stil {
	return msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "10", Oben: "10", Breite: "300", Hoehe: "100",
			Schrift:       msepaket.Schrift{Name: "Helvetica", Groesse: "12"},
			Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
		}},
	}
}

// probesymbole ist eine Symbolschrift mit zwei Zeichen.
func probesymbole() map[string]msepaket.Symbolschrift {
	return map[string]msepaket.Symbolschrift{
		"probe-symbole": {Name: "probe-symbole", Symbole: []msepaket.Symbol{
			{Code: "W", Bild: "w.png"},
			{Code: "U", Bild: "u.png"},
		}},
	}
}

// planMitSymbolen baut einen Plan über ein Textfeld mit Symbolschrift.
func planMitSymbolen(t *testing.T, text string) *render.Plan {
	t.Helper()

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: text},
	}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte,
		service.Vorlage{Stil: symbolstil(), Symbolschriften: probesymbole()})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	return plan
}
