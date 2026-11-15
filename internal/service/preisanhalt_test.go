package service_test

// **Was die zwei Seiten eines Tauschs ungefähr wert sind** — TRADE-02b
// (#1951).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// preisleser gibt je Druck, was eingetragen ist, und zählt die Fragen.
type preisleser struct {
	staende map[string][]domain.Preisstand
	fehler  error
	fragen  []string
}

func (l *preisleser) Neueste(_ context.Context, printing string) ([]domain.Preisstand, error) {
	l.fragen = append(l.fragen, printing)

	return l.staende[printing], l.fehler
}

var anhaltsTag = time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)

// genannt ist ein Verkaufspreis für Papier, wie das Portfolio ihn zählt.
func genannt(printing, waehrung string, finish domain.Finish, wert int64, tage int) domain.Preisstand {
	return domain.Preisstand{
		Printing: printing, Quelle: domain.Preisquelle("scryfall"),
		Markt: domain.MarktPapier, Art: domain.PreisartVerkauf, Lage: domain.PreisGenannt,
		Waehrung: waehrung, Finish: finish, Wert: wert,
		Erhoben: anhaltsTag.AddDate(0, 0, -tage),
	}
}

func position(von, printing string, menge int, finish string) domain.Handelsposition {
	return domain.Handelsposition{Von: von, Printing: printing, Menge: menge, Finish: finish}
}

func TestDerAnhaltRechnetJeSeiteUndWaehrung(t *testing.T) {
	leser := &preisleser{staende: map[string][]domain.Preisstand{
		"bolt":  {genannt("bolt", "EUR", domain.FinishNonfoil, 150, 1), genannt("bolt", "USD", domain.FinishNonfoil, 180, 1)},
		"ring":  {genannt("ring", "EUR", domain.FinishNonfoil, 1790, 3), genannt("ring", "EUR", domain.FinishFoil, 4000, 3)},
		"goyf":  {genannt("goyf", "EUR", domain.FinishFoil, 2000, 2)},
		"leere": {},
	}}

	anhalt, fehler := service.Handelsanhalt(context.Background(), leser, []domain.Handelsposition{
		position("ich", "bolt", 2, "nonfoil"),
		position("du", "ring", 1, "nonfoil"),
		position("ich", "goyf", 1, "foil"),
		position("du", "leere", 3, ""),
	}, "ich", "")
	if fehler != nil {
		t.Fatal(fehler)
	}

	// **Je Zeile der Preis mal Menge** — und eine Zeile ohne Preis ist leer,
	// nicht null.
	zeilen := anhalt.Zeilen
	if len(zeilen) != 4 || len(zeilen[3].Preise) != 0 {
		t.Fatalf("Zeilen: %+v", zeilen)
	}

	if bolt := zeilen[0].Preise; len(bolt) != 2 || bolt[0].Waehrung != "EUR" || bolt[0].Wert != 300 ||
		bolt[1].Waehrung != "USD" || bolt[1].Wert != 360 {
		t.Errorf("bolt: %+v", bolt)
	}

	// **Veredelt zu veredelt**: die gewöhnliche Ringzeile nimmt 17,90 und
	// nicht den Foilpreis.
	if ring := zeilen[1].Preise; len(ring) != 1 || ring[0].Wert != 1790 {
		t.Errorf("ring: %+v", ring)
	}

	// **Die eigene Seite zuerst**, dann das Gegenüber.
	if len(anhalt.Seiten) != 2 || anhalt.Seiten[0].Von != "ich" || anhalt.Seiten[1].Von != "du" {
		t.Fatalf("Seiten: %+v", anhalt.Seiten)
	}

	ich, du := anhalt.Seiten[0], anhalt.Seiten[1]
	if ich.OhnePreis != 0 || du.OhnePreis != 1 {
		t.Errorf("ohne Preis: ich %d, du %d", ich.OhnePreis, du.OhnePreis)
	}

	if eur := summeIn(ich.Summen, "EUR"); eur.Wert != 2300 || !eur.Erhoben.Equal(anhaltsTag.AddDate(0, 0, -2)) {
		t.Errorf("meine Summe: %+v", ich.Summen)
	}

	// **Die Differenz ist aus meiner Sicht**: erhalten weniger gegeben — und
	// sie ist unvollständig, solange eine Zeile keinen Preis hat.
	eur := differenzIn(anhalt.Differenzen, "EUR")
	if eur.Wert != 1790-2300 || eur.Vollstaendig {
		t.Errorf("Differenz EUR: %+v", anhalt.Differenzen)
	}

	// **Die Hauptwährung ist die mit den meisten Zeilen**, ohne Vorgabe.
	if anhalt.Hauptwaehrung != "EUR" {
		t.Errorf("Hauptwährung: %q", anhalt.Hauptwaehrung)
	}
}

func TestDerAnhaltIstVollstaendigWennJedeZeileEinenPreisHat(t *testing.T) {
	leser := &preisleser{staende: map[string][]domain.Preisstand{
		"bolt": {genannt("bolt", "EUR", domain.FinishNonfoil, 150, 0), genannt("bolt", "USD", domain.FinishNonfoil, 180, 0)},
		"ring": {genannt("ring", "USD", domain.FinishNonfoil, 200, 0)},
	}}

	anhalt, fehler := service.Handelsanhalt(context.Background(), leser, []domain.Handelsposition{
		position("du", "ring", 1, "nonfoil"),
		position("ich", "bolt", 1, "nonfoil"),
		position("du", "bolt", 1, "nonfoil"),
	}, "ich", "EUR")
	if fehler != nil {
		t.Fatal(fehler)
	}

	// **Die eigene Seite steht vorn**, auch wenn das Gegenüber zuerst kam.
	if len(anhalt.Seiten) != 2 || anhalt.Seiten[0].Von != "ich" {
		t.Errorf("Seiten: %+v", anhalt.Seiten)
	}

	// **Vollständig nur in der Währung, die jede Zeile nennt** — hier USD.
	if usd := differenzIn(anhalt.Differenzen, "USD"); !usd.Vollstaendig || usd.Wert != 200+180-180 {
		t.Errorf("USD: %+v", anhalt.Differenzen)
	}

	if eur := differenzIn(anhalt.Differenzen, "EUR"); eur.Vollstaendig {
		t.Errorf("EUR: %+v", anhalt.Differenzen)
	}

	// **Die Vorgabe gewinnt**, wo sie vorkommt.
	if anhalt.Hauptwaehrung != "EUR" {
		t.Errorf("Hauptwährung: %q", anhalt.Hauptwaehrung)
	}

	// **Ein Druck, eine Frage** — `bolt` steht zweimal da.
	if len(leser.fragen) != 2 {
		t.Errorf("Fragen: %v", leser.fragen)
	}

	// **Und die Differenzen stehen nach Währung geordnet.**
	if len(anhalt.Differenzen) != 2 || anhalt.Differenzen[0].Waehrung != "EUR" {
		t.Errorf("Ordnung: %+v", anhalt.Differenzen)
	}
}

func TestEineVorgabeOhnePreiseWirdNichtHauptwaehrung(t *testing.T) {
	leser := &preisleser{staende: map[string][]domain.Preisstand{
		"bolt": {genannt("bolt", "USD", domain.FinishNonfoil, 180, 0)},
	}}

	anhalt, fehler := service.Handelsanhalt(context.Background(), leser,
		[]domain.Handelsposition{position("ich", "bolt", 1, "")}, "ich", "EUR")
	if fehler != nil {
		t.Fatal(fehler)
	}

	if anhalt.Hauptwaehrung != "USD" {
		t.Errorf("Hauptwährung: %q", anhalt.Hauptwaehrung)
	}

	// **Eine Seite ohne Zeilen fehlt nicht** — das Gegenüber gibt nichts, und
	// das ist eine Aussage.
	if len(anhalt.Seiten) != 1 {
		t.Errorf("Seiten: %+v", anhalt.Seiten)
	}
}

func TestOhnePositionenGibtEsKeinenAnhalt(t *testing.T) {
	anhalt, fehler := service.Handelsanhalt(context.Background(), &preisleser{}, nil, "ich", "EUR")
	if fehler != nil || len(anhalt.Zeilen) != 0 || anhalt.Hauptwaehrung != "" || len(anhalt.Differenzen) != 0 {
		t.Errorf("%+v (%v)", anhalt, fehler)
	}
}

func TestEinLesefehlerGehtAusDemAnhaltZurueck(t *testing.T) {
	kaputt := errors.New("Datei weg")

	_, fehler := service.Handelsanhalt(context.Background(), &preisleser{fehler: kaputt},
		[]domain.Handelsposition{position("ich", "bolt", 1, "")}, "ich", "")
	if !errors.Is(fehler, kaputt) {
		t.Errorf("Fehler: %v", fehler)
	}
}

func summeIn(summen []service.Anhaltspreis, waehrung string) service.Anhaltspreis {
	for _, summe := range summen {
		if summe.Waehrung == waehrung {
			return summe
		}
	}

	return service.Anhaltspreis{}
}

func differenzIn(differenzen []service.Anhaltsdifferenz, waehrung string) service.Anhaltsdifferenz {
	for _, differenz := range differenzen {
		if differenz.Waehrung == waehrung {
			return differenz
		}
	}

	return service.Anhaltsdifferenz{}
}
