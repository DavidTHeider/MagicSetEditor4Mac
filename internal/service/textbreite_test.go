package service_test

// **Die Inhaltsbreite eines Textzugs** — der Teil des zweistufigen Satzes, der
// Schriftmaße braucht.
//
// Getrennt von `inhaltsbreite_test.go`, das den Symbolzug hält: der eine misst
// über eine Multiplikation, der andere über eine Schrift von außen.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schriftprobe gibt für jeden Namen dieselbe gleichbreite Schrift.
type schriftprobe struct {
	breite float64
	fehler error
}

// **Der fette Schnitt ist doppelt so breit.** Eine Vorrichtung, die das
// Gewicht wegwürfe, wäre großzügiger als der Betrieb (§3).
func (s schriftprobe) Suche(_ string, fett bool) (render.Schrift, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	if fett {
		return gleichbreiteSchrift{breite: s.breite * 2}, nil
	}

	return gleichbreiteSchrift{breite: s.breite}, nil
}

type gleichbreiteSchrift struct{ breite float64 }

func (g gleichbreiteSchrift) Breite(rune) float64 { return g.breite }
func (gleichbreiteSchrift) Zeilenhoehe() float64  { return 1.2 }

// **Auch ein Textzug wird gemessen** — die zweite Hälfte des zweistufigen
// Satzes.
//
// Anders als beim Symbolzug braucht es dafür die Schrift, und die kennt
// `internal/service` nach §2 nicht: sie kommt als [render.Schriftquelle] von
// außen, verdrahtet im Composition Root — dieselbe, die auch der Rasterer
// bekommt.
func TestDieInhaltsbreiteEinesTextfeldsWirdGemessen(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
			{
				Feld: "type", Links: "10", Oben: "40", Hoehe: "23",
				Rechts:  "{ card_style.name.content_width }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "abcd"},
		{Field: "type", Text: "x"},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	// `abcd` à 0,5 bei Größe 10 ergibt 20; die Typzeile reicht also von 10 bis
	// 20 und ist 10 breit. Ohne Messung wäre `content_width` 100 und die
	// Typzeile 90 breit.
	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText && ebene.Klartext() == "x" {
			if ebene.Rahmen.Breite != 10 {
				t.Errorf("die Typzeile ist %.0f breit, erwartet 10", ebene.Rahmen.Breite)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Typzeile")
}

// **Wo die Schrift nicht zu finden ist, wird nicht gemessen.**
//
// Eine fehlende Schrift ist ein Befund des Rasterers, kein Grund, hier eine
// Zahl zu erfinden. Es bleibt bei der deklarierten Breite.
func TestOhneAuffindbareSchriftBleibtDieDeklarierteBreite(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
			Schrift: msepaket.Schrift{Name: "Fehlt", Groesse: "10"},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "abcd"},
	}}

	for _, fall := range []struct {
		name   string
		quelle render.Schriftquelle
	}{
		{name: "keine Schriftquelle", quelle: nil},
		{name: "die Schrift fehlt", quelle: schriftprobe{fehler: errKeineSchrift}},
		{name: "die Schrift ist null breit", quelle: schriftprobe{breite: 0}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(fall.quelle).
				Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if len(plan.Ebenen) < 2 {
				t.Fatal("der Plan führt keine Textebene")
			}
		})
	}
}

var errKeineSchrift = errors.New("keine Schrift zu diesem Namen")

// **Gemessen wird, was gezeichnet wird — nicht die rohe Karte.**
//
// # Der Befund des Zweitblicks
//
// [inhaltsmasse] las die Karte, bevor die Vorgabewerte ergänzt und die
// Auszeichnungen entfernt waren. Ein Feld, das seinen Wert erst aus `default:`
// bekommt, blieb ungemessen; ein Text mit `<i-auto>` wurde **mit** den spitzen
// Klammern gemessen und **ohne** sie gezeichnet.
//
// **Messung und Darstellung arbeiteten auf verschiedenen Stufen derselben
// Datenaufbereitung** — und die Anordnung folgte der falschen.
func TestGemessenWirdWasGezeichnetWird(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{
			{
				Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
			{
				Feld: "type", Links: "10", Oben: "40", Hoehe: "23",
				Rechts:  "{ card_style.name.content_width }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	// **Der Name kommt aus dem Vorgabewert und trägt eine Auszeichnung.**
	// Gezeichnet werden vier Zeichen (`abcd`), gemessen wurden vorher elf
	// (`<i>abcd</i>`) — oder gar keine, weil die Karte das Feld nicht führt.
	vorlage := service.Vorlage{Stil: stil, Kartenfelder: []domain.Field{{
		Name: "name", Default: `"<i>abcd</i>"`,
	}}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(),
			&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{{Field: "type", Text: "x"}}},
			vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText || ebene.Klartext() != "x" {
			continue
		}

		// Vier Zeichen à 0,5 bei Größe 10 ergeben 20; ab 10 also 10 breit.
		if ebene.Rahmen.Breite != 10 {
			t.Errorf("die Typzeile ist %.0f breit, erwartet 10 — gemessen wurde nicht, "+
				"was gezeichnet wird", ebene.Rahmen.Breite)
		}

		return
	}

	t.Fatal("der Plan führt keine Typzeile")
}
