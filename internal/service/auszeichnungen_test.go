package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Auszeichnungen stehen nicht roh auf der Karte.**
//
// # Der Befund
//
// Ein MSE-Kartentext trägt Auszeichnungen in spitzen Klammern:
//
//	Alliance Subversive — <sym>T</sym>: Target creature is lawbound…
//	<i-auto>(You have alliance if you control another subversive creature.)</i-auto>
//	<i-flavor>"What do you mean why? Burn! Burn!"</i-flavor>
//
// Bis zum 2026-08-16 standen sie **wörtlich** im Textkasten — auf jeder der vier
// Referenzkarten sichtbar. Gezählt in `magic.mse-game`: 2.165 `<atom-param>`,
// 108 `<sym>`, 16 `<i-auto>`, dazu `<atom-sep>` und `<atom-autorem>`.
//
// # Was hier zugesagt wird, und was nicht
//
// **Die Klammern verschwinden.** Was sie bedeuten, wird noch nicht gezeichnet:
// `render.Satz` führt `Fett` und `Kursiv`, und **niemand liest sie** — der
// Rasterer kennt keine kursive Schnittvariante, und die Schriftsuche nimmt nur
// einen Namen entgegen.
//
// Deshalb bekommt jede weggelassene Auszeichnung einen **Befund**: sichtbar
// falsch ist besser als unsichtbar fehlend, und wer die Befundzahl als Maß
// nimmt, soll nicht dadurch besser dastehen, dass etwas halb getan wurde (§8).
func TestAuszeichnungenStehenNichtRohAufDerKarte(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "375", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
		}},
	}

	// **`<atom-param>` ist nicht gedeutet, `<i-auto>` schon.** Der Befund
	// gehört dem ersten; würde er auch für kursiv gemeldet, stünde er auf jeder
	// Karte mit Flavortext (#326).
	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{{
		Field: "text",
		Text:  `Fliegend. <i-auto>(Das heißt: <atom-param>hoch</atom-param>.)</i-auto>`,
	}}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText {
			continue
		}

		if strings.Contains(ebene.Klartext(), "<") {
			t.Errorf("der Text lautet %q und trägt noch spitze Klammern", ebene.Klartext())
		}

		if ebene.Klartext() != "Fliegend. (Das heißt: hoch.)" {
			t.Errorf("der Text lautet %q, erwartet ohne Auszeichnungen", ebene.Klartext())
		}

		break
	}

	// **Und der Befund reist mit**, mit dem Namen der Auszeichnung, die
	// **nicht** gedeutet ist.
	var genannt bool

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "atom-param") {
			genannt = true
		}
	}

	if !genannt {
		t.Errorf("keine Meldung nennt die weggelassene Auszeichnung: %v", befunde)
	}
}

// **Ein Text ohne Auszeichnungen bekommt keinen Befund.**
//
// Sonst trüge jede Karte eine Meldung, und eine Meldung, die immer dasteht,
// liest niemand mehr.
func TestEinTextOhneAuszeichnungenBleibtOhneBefund(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "375", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "Fliegend."},
	}}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde ohne Auszeichnungen: %v", befunde)
	}
}

// **Ein `<` ohne schließende Klammer bleibt stehen.**
//
// Ein Kartentext darf ein Kleinerzeichen enthalten — `1 < 2` ist keine
// Auszeichnung. Wer alles ab `<` verwürfe, fräße den halben Satz.
func TestEinEinzelnesKleinerzeichenBleibtStehen(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "375", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "1 < 2 und 3 > 2"},
	}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText && ebene.Klartext() != "1 < 2 und 3 > 2" {
			t.Errorf("der Text lautet %q, erwartet unverändert", ebene.Klartext())
		}
	}
}

// **Die Randfälle einer spitzen Klammer**, jeder einzeln.
//
// Sie stehen als Tabelle, weil sie eine Familie sind: was zwischen `<` und `>`
// steht, entscheidet, ob es Text ist oder eine Auszeichnung — und jede
// Entscheidung hat ihren eigenen Grund.
func TestDieRandfaelleEinerSpitzenKlammer(t *testing.T) {
	faelle := []struct {
		name     string
		text     string
		erwartet string
	}{
		{"eine offene Klammer ohne Ende", "1 < 2", "1 < 2"},
		{"leere Klammern", "a <> b", "a <> b"},
		{"eine Klammer mit Leerzeichen", "a < b > c", "a < b > c"},
		{"eine schließende Auszeichnung", "a</i>b", "ab"},
		{"eine Auszeichnung mit Ziffer", "a<h1>b", "ab"},
		{"zwei nacheinander", "<i><b>x</b></i>", "x"},
		{"ein Größerzeichen allein", "3 > 2", "3 > 2"},

		// **Ohne Leerzeichen**, und genau hier fiel der erste Anlauf um: `2`
		// ging als Auszeichnungsname durch, und aus dem Vergleich wurde `10`.
		// Angestrichen vom Zweitblick — der Prüffall darüber trägt Leerzeichen
		// und betritt diesen Weg nie.
		{"ein Vergleich ohne Leerzeichen", "1<2>0", "1<2>0"},
		{"eine Auszeichnung beginnt mit einem Buchstaben", "a<-b>c", "a<-b>c"},
	}

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "375", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan, _, err := service.NeuerPlanbauer().Stilplan(
				t.Context(),
				&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
					{Field: "text", Text: fall.text},
				}},
				service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, ebene := range plan.Ebenen {
				if ebene.Art == render.ArtText {
					if ebene.Klartext() != fall.erwartet {
						t.Errorf("%q wird zu %q, erwartet %q",
							fall.text, ebene.Klartext(), fall.erwartet)
					}

					return
				}
			}

			t.Fatal("der Plan führt keine Textebene")
		})
	}
}

// **Ein Auszeichnungsname mit einem Sonderzeichen ist keiner.**
//
// `<a.b>` und `<a b>` sind Text; die Grenze läuft über die erlaubten Zeichen,
// und sie wird hier einzeln geprüft statt nur an einem Beispiel.
func TestEinNameMitSonderzeichenIstKeineAuszeichnung(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "375", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
		}},
	}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(),
		&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "text", Text: "x<a.b>y"},
		}},
		service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText && ebene.Klartext() != "x<a.b>y" {
			t.Errorf("der Text lautet %q, erwartet unverändert", ebene.Klartext())
		}
	}
}

// **Eine Auszeichnung darf einen Parameter tragen.**
//
// # Der Befund des Nachlaufs
//
// MSE2 deutet fünf mit Parameter — `<color:black>`, `<font:Beleren Bold>`,
// `<size:…>`, `<margin:0:0:5>`, `<align:…>` (`element.cpp:98–130`) —, und sie
// stehen auch im Pack. Der erste Anlauf ließ weder Doppelpunkt noch Leerzeichen
// zu und **ließ sie roh auf der Karte stehen**.
//
// **Der Fall `1<2>0` war geschlossen, die Klasse nicht.**
func TestEineAuszeichnungDarfEinenParameterTragen(t *testing.T) {
	faelle := []struct {
		name     string
		text     string
		erwartet string
	}{
		{"eine Farbe", `a<color:black>b</color>c`, "abc"},
		{"eine Schrift mit Leerzeichen", `a<font:Beleren Bold>b</font>c`, "abc"},
		{"drei Zahlen", `a<margin:0:0:5>b</margin>c`, "abc"},

		// **Und der Vergleich bleibt ein Vergleich.** Vor dem Doppelpunkt darf
		// keine Ziffer stehen.
		{"ein Vergleich mit Doppelpunkt", "1<2:3>0", "1<2:3>0"},
	}

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "375", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan, _, err := service.NeuerPlanbauer().Stilplan(
				t.Context(),
				&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
					{Field: "text", Text: fall.text},
				}},
				service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, ebene := range plan.Ebenen {
				if ebene.Art == render.ArtText {
					if ebene.Klartext() != fall.erwartet {
						t.Errorf("%q wird zu %q, erwartet %q",
							fall.text, ebene.Klartext(), fall.erwartet)
					}

					return
				}
			}

			t.Fatal("der Plan führt keine Textebene")
		})
	}
}

// **Ein Doppelpunkt ohne Namen davor ist keine Auszeichnung.**
func TestEinDoppelpunktOhneNamenIstKeineAuszeichnung(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "375", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
		}},
	}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(),
		&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "text", Text: "a<:b>c"},
		}},
		service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText && ebene.Klartext() != "a<:b>c" {
			t.Errorf("der Text lautet %q, erwartet unverändert", ebene.Klartext())
		}
	}
}

// **Ein Name aus lauter Doppelpunkten hat keinen Namen.**
func TestEinNameAusDoppelpunktenIstKeineAuszeichnung(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "375", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
		}},
	}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(),
		&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "text", Text: "a<1:b>c"},
		}},
		service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText && ebene.Klartext() != "a<1:b>c" {
			t.Errorf("der Text lautet %q, erwartet unverändert — eine Ziffer ist kein Name",
				ebene.Klartext())
		}
	}
}
