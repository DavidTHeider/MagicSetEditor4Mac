package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zusatzpflege ist die Pflege der Vorrichtung: sie gibt, was sie bekommt, und
// schreibt mit, was gesetzt wurde.
type zusatzpflege struct {
	felder  []service.Zusatzfeldwert
	fehler  error
	gesetzt []string
}

func (p *zusatzpflege) Felder(context.Context, int64, int64) ([]service.Zusatzfeldwert, error) {
	return p.felder, p.fehler
}

func (p *zusatzpflege) Setzen(
	_ context.Context, _, _ int64, feld, wert string,
) (service.Zusatzvorwert, error) {
	p.gesetzt = append(p.gesetzt, feld+"="+wert)

	return service.Zusatzvorwert{}, p.fehler
}

func mitZusatzpflege(pflege *zusatzpflege) *app.App {
	anwendung := &app.App{}
	anwendung.MitZusatzwertpflege(pflege)

	return anwendung
}

// **Die Zusatzfelder einer Karte kommen mit Beschreibung und Wert an**
// (#2298) —
// in derselben Gestalt wie ein Kartenfeld, damit das Formular sie mit
// denselben Bauteilen zeichnet.
func TestDieZusatzfelderEinerKarteKommenAlsBeschreibungUndWert(t *testing.T) {
	t.Parallel()

	pflege := &zusatzpflege{felder: []service.Zusatzfeldwert{
		{
			Feld: domain.Field{
				Name: "keyword icons", Kind: domain.KindMultipleChoice,
				Choices: []string{"flying", "haste"}, Beschreibung: "Icons for keywords",
			},
			Wert: "flying", Gespeichert: true,
		},
		{Feld: domain.Field{Name: "promo label", Kind: domain.KindText}},
	}}

	antwort := mitZusatzpflege(pflege).Kartenzusatzfelder(probekontext(t), "7", "1")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(antwort.Felder) != 2 {
		t.Fatalf("%d Felder", len(antwort.Felder))
	}

	erstes := antwort.Felder[0]
	if erstes.Form.Name != "keyword icons" || erstes.Form.Art != "multiple choice" ||
		len(erstes.Form.Wahlen) != 2 || erstes.Wert != "flying" || !erstes.Gespeichert {
		t.Errorf("das erste Feld: %+v", erstes)
	}

	if antwort.Felder[1].Gespeichert || antwort.Felder[1].Wert != "" {
		t.Errorf("das zweite Feld: %+v", antwort.Felder[1])
	}
}

// **Fehler kommen übersetzt an, und eine leere Liste bleibt eine Liste.**
func TestDieZusatzfelderMeldenIhreFehler(t *testing.T) {
	t.Parallel()

	t.Run("eine unbrauchbare Kennung", func(t *testing.T) {
		antwort := mitZusatzpflege(&zusatzpflege{}).Kartenzusatzfelder(probekontext(t), "7", "x")
		if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeValidierung || antwort.Felder == nil {
			t.Errorf("Antwort %+v", antwort)
		}
	})

	t.Run("ein unbrauchbares Set", func(t *testing.T) {
		antwort := mitZusatzpflege(&zusatzpflege{}).Kartenzusatzfelder(probekontext(t), "", "1")
		if antwort.Fehler == nil || antwort.Fehler.Feld != "setID" {
			t.Errorf("Antwort %+v", antwort)
		}
	})

	t.Run("eine Karte, die es nicht gibt", func(t *testing.T) {
		antwort := mitZusatzpflege(&zusatzpflege{fehler: domain.ErrNotFound}).
			Kartenzusatzfelder(probekontext(t), "7", "1")
		if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNichtGefunden || antwort.Felder == nil {
			t.Errorf("Antwort %+v", antwort)
		}
	})
}

// **Gesetzt wird über die Pflege, und ihr Fehler kommt am Feld an.**
func TestEinZusatzwertWirdUeberDieBindungGesetzt(t *testing.T) {
	t.Parallel()

	pflege := &zusatzpflege{}

	if antwort := mitZusatzpflege(pflege).
		ZusatzwertSetzen(probekontext(t), "7", "1", "promo label", "Prerelease"); antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(pflege.gesetzt) != 1 || pflege.gesetzt[0] != "promo label=Prerelease" {
		t.Errorf("gesetzt: %v", pflege.gesetzt)
	}

	abgewiesen := mitZusatzpflege(&zusatzpflege{fehler: domain.NewFieldError(
		"feld", domain.SchluesselValidierung, "kein Zusatzfeld")}).
		ZusatzwertSetzen(probekontext(t), "7", "1", "gibtsnicht", "x")
	if abgewiesen.Fehler == nil || abgewiesen.Fehler.Feld != "feld" {
		t.Errorf("Antwort %+v", abgewiesen)
	}

	for _, kennungen := range [][2]string{{"", "1"}, {"7", "x"}} {
		if antwort := mitZusatzpflege(&zusatzpflege{}).
			ZusatzwertSetzen(probekontext(t), kennungen[0], kennungen[1], "promo label", "x"); antwort.Fehler == nil {
			t.Errorf("%v: kein Befund", kennungen)
		}
	}
}
