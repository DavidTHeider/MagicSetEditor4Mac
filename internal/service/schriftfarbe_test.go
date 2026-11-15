package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schriftfeld baut ein Textfeld mit einer Schriftfarbe.
func schriftfeld(farbe string) msepaket.Feldstil {
	feld := flaeche("name")
	feld.Schrift = msepaket.Schrift{Name: "Helvetica", Groesse: "12", Farbe: farbe}

	return feld
}

// textebene sucht die eine Textebene eines Plans.
func textebene(t *testing.T, plan *render.Plan) render.Ebene {
	t.Helper()

	gezeichnet := ebenen(t, plan)
	for i := range gezeichnet {
		if gezeichnet[i].Art == render.ArtText {
			return gezeichnet[i]
		}
	}

	t.Fatal("der Plan führt keine Textebene")

	return render.Ebene{}
}

// weiss ist die Farbe, die die Fälle unten erwarten.
var weiss = render.Farbe{R: 255, G: 255, B: 255, A: 255}

// **Eine Schriftfarbe darf ein Skript sein.**
//
// Über die 190 Stile von `M15-Magic-Pack` nennen 3.011 Feldstile eine
// `font: color:`, und **2.321** davon sind Skripte. Bis
// #382 wurde
// der rohe Text als Farbangabe gelesen — er ist keine, und jede dieser Ebenen
// wurde schwarz.
func TestEineSchriftfarbeDarfEinSkriptSein(t *testing.T) {
	faelle := map[string]string{
		"gerechnete Farbe":     "{ rgb(255,255,255) }",
		"gerechneter Farbname": `{ "white" }`,
		"Bedingung":            `{ if card.farbe == "rot" then rgb(255,255,255) else rgb(0,0,0) }`,
	}

	for name, angabe := range faelle {
		t.Run(name, func(t *testing.T) {
			plan, befunde := gedeutetMit(t, stilkarte("rot"), bildstil(schriftfeld(angabe)))

			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			if gesetzt := textebene(t, plan).Farbe; gesetzt != weiss {
				t.Errorf("Farbe %+v, erwartet Weiß", gesetzt)
			}
		})
	}
}

// **Ein Farbname bleibt ein Farbname.** Die häufigere Form daneben — 690 der
// 3.011 Angaben sind Literale — geht denselben Weg wie bisher.
func TestEineSchriftfarbeDarfEinLiteralBleiben(t *testing.T) {
	for _, angabe := range []string{"white", "rgb(255,255,255)", "#ffffff"} {
		t.Run(angabe, func(t *testing.T) {
			plan, befunde := gedeutetMit(t, stilkarte("rot"), bildstil(schriftfeld(angabe)))

			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			if gesetzt := textebene(t, plan).Farbe; gesetzt != weiss {
				t.Errorf("Farbe %+v, erwartet Weiß", gesetzt)
			}
		})
	}
}

// **Eine Angabe, die sich nicht ausrechnen lässt, wird benannt** (§8) — und die
// Ebene bleibt, in Schwarz. Ein Text ohne Farbe ist schlechter als einer in der
// falschen.
func TestEineUnausrechenbareSchriftfarbeWirdBenannt(t *testing.T) {
	plan, befunde := gedeutetMit(t, stilkarte("rot"),
		bildstil(schriftfeld("{ gibtesnicht() }")))

	if len(befunde) == 0 {
		t.Fatal("kein Befund, obwohl die Angabe nicht auszurechnen ist")
	}

	if !strings.Contains(befunde[0].Ort, "name") ||
		!strings.Contains(befunde[0].Meldung, "ausrechnen") {
		t.Errorf("Befund %+v", befunde[0])
	}

	if gesetzt := textebene(t, plan).Farbe; gesetzt != (render.Farbe{A: 255}) {
		t.Errorf("Farbe %+v, erwartet Schwarz", gesetzt)
	}
}

// **Eine Angabe, die keine Farbe ergibt, wird ebenfalls benannt.**
//
// Ausgerechnet hat sie sich — nur steht am Ende kein Farbwert und kein Name,
// den [gedeuteteFarbe] kennt. Schweigend Schwarz zu nehmen hieße, einen
// unlesbaren Text als Absicht auszugeben.
func TestEineSchriftfarbeOhneFarbeWirdBenannt(t *testing.T) {
	_, befunde := gedeutetMit(t, stilkarte("rot"), bildstil(schriftfeld(`{ "karminrot" }`)))

	if len(befunde) == 0 {
		t.Fatal("kein Befund, obwohl die Angabe keine Farbe ergibt")
	}
}

// **Und die Textfarbe eines Symbols geht denselben Weg.**
//
// `stilsymbole.go` las sie mit demselben Aufruf roh — dieselbe Klasse, dieselbe
// Stelle im Plan (§10: die Klasse treffen, nicht den Fall).
func TestDieTextfarbeEinesSymbolsDarfEinSkriptSein(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "10", Oben: "20", Breite: "60", Hoehe: "15",
		Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
		ImmerSymbole:  true,
	}

	vorlage := service.Vorlage{
		Stil: bildstil(feld),
		Symbolschriften: map[string]msepaket.Symbolschrift{
			"mana": {Name: "mana", Symbole: []msepaket.Symbol{
				{Code: "R", Bild: "r.png", Textgruppe: 0, Textschrift: msepaket.Symboltextschrift{
					Name: "Helvetica", Groesse: 0.6, Farbe: "{ rgb(255,255,255) }",
				}},
			}},
		},
	}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(),
		&domain.CardRow{ID: 1, Values: []domain.Value{{Field: "casting cost", Text: "R"}}},
		vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 1 || len(gezeichnet[0].Abschnitte) != 1 {
		t.Fatalf("Ebenen %+v", gezeichnet)
	}

	symbole := gezeichnet[0].Abschnitte[0].Symbole
	if len(symbole) != 1 {
		t.Fatalf("%d Symbole, erwartet eines", len(symbole))
	}

	if farbe := symbole[0].Textsatz.Farbe; farbe != weiss {
		t.Errorf("Textfarbe %+v, erwartet Weiß", farbe)
	}
}

// **Ein Fehlerwert nennt seinen eigenen Grund.**
//
// `card.gibtsnicht` ist in MSE kein Abbruch, sondern ein Wert, den die Vorlage
// mit `or else` auffangen darf. Fängt sie ihn nicht auf, steht er hier — und
// die Meldung soll sagen, **welcher** Zugriff ins Leere ging, nicht nur, dass
// etwas schieflief.
func TestEineSchriftfarbeAusEinemFehlerwertWirdBenannt(t *testing.T) {
	_, befunde := gedeutetMit(t, stilkarte("rot"), bildstil(schriftfeld("{ card.gibtsnicht }")))

	if len(befunde) == 0 {
		t.Fatal("kein Befund, obwohl der Zugriff ins Leere ging")
	}

	if !strings.Contains(befunde[0].Meldung, "gibtsnicht") {
		t.Errorf("Befund %+v — er nennt den fehlgeschlagenen Zugriff nicht", befunde[0])
	}
}

// **Nichts ist keine Farbe, aber auch kein Fehler.**
//
// `{ if … then rgb(…) }` ohne `else` ergibt bei MSE `nil`, und dann gilt die
// Vorgabe. Ein Befund daraus machte aus einer Vorlage, die schweigt, eine, die
// kaputt ist (#345).
func TestEineSchriftfarbeAusNichtsIstKeinBefund(t *testing.T) {
	plan, befunde := gedeutetMit(t, stilkarte("rot"),
		bildstil(schriftfeld(`{ if card.farbe == "blau" then rgb(255,255,255) }`)))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if gesetzt := textebene(t, plan).Farbe; gesetzt != (render.Farbe{A: 255}) {
		t.Errorf("Farbe %+v, erwartet Schwarz", gesetzt)
	}
}
