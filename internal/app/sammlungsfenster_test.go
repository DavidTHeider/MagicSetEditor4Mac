package app_test

// **Nur holen, was gezeichnet wird** — Befund aus U8, Werkstück
// #1500.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zehnExemplare gibt zehn, von A bis J.
func zehnExemplare() []*domain.Exemplar {
	alle := make([]*domain.Exemplar, 0, 10)

	for nummer := range 10 {
		name := string(rune('A' + nummer))
		alle = append(alle, &domain.Exemplar{
			ID: int64(nummer + 1), Printing: "scryfall:" + name,
			Name: name, Setcode: "m11", Menge: 10 - nummer,
			Finish: domain.FinishAus(nummer%2 == 0),
		})
	}

	return alle
}

func fensterapp(doppel *sammlungsnahtdoppel) *app.App {
	anwendung := app.NeueApp("/nirgends/bestand.db")
	anwendung.MitSammlung(doppel)

	// **Über den echten Dienst** (Befund 1 des Zweitblicks über Phase 5.2):
	// die Bindung rechnet nichts mehr selbst, und ein Doppel an ihrer Stelle
	// prüfte die Übersetzung statt des Weges.
	anwendung.MitSammlungsfenster(service.NeuesSammlungsfenster(fensterablage{doppel}))

	return anwendung
}

// fensterablage macht aus dem Sammlungsdoppel eine Ablage für den Dienst.
//
// **Sie liest einmal und wählt daraus** — so, wie der Bestand es im
// Lesegeschäft tut.
type fensterablage struct{ doppel *sammlungsnahtdoppel }

func (f fensterablage) Fenster(
	ctx context.Context, waehle func([]*domain.Exemplar) []int64,
) ([]*domain.Exemplar, error) {
	alle, fehler := f.doppel.Liste(ctx)
	if fehler != nil {
		return nil, fehler
	}

	nach := make(map[int64]*domain.Exemplar, len(alle))
	for _, eines := range alle {
		nach[eines.ID] = eines
	}

	gewaehlt := []*domain.Exemplar{}

	for _, kennung := range waehle(alle) {
		if eines, da := nach[kennung]; da {
			gewaehlt = append(gewaehlt, eines)
		}
	}

	return gewaehlt, nil
}

// **Das Fenster gibt den Ausschnitt und beide Zahlen.**
func TestDasSammlungsfensterGibtAusschnittUndZahlen(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsnahtdoppel{exemplare: zehnExemplare()}

	antwort := fensterapp(doppel).Sammlungsfenster(t.Context(), "", "", "", 0, 3)
	if antwort.Fehler != nil {
		t.Fatalf("das Fenster: %+v", antwort.Fehler)
	}

	if len(antwort.Exemplare) != 3 {
		t.Fatalf("es sind %d Zeilen", len(antwort.Exemplare))
	}

	if antwort.Gesamt != 10 || antwort.Insgesamt != 10 {
		t.Errorf("die Zahlen sind %d und %d", antwort.Gesamt, antwort.Insgesamt)
	}

	// **Nach Bezeichnung geordnet** — die Vorgabe.
	if antwort.Exemplare[0].Name != "A" {
		t.Errorf("die erste Zeile ist %q", antwort.Exemplare[0].Name)
	}
}

// **Die Sicht wirkt vor dem Ausschnitt** — sonst wären es die falschen Zeilen.
func TestDieSichtWirktVorDemAusschnitt(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsnahtdoppel{exemplare: zehnExemplare()}

	// **Nach Menge absteigend**: A hat zehn, J hat eine.
	antwort := fensterapp(doppel).Sammlungsfenster(t.Context(), "", "", "menge", 0, 2)
	if antwort.Fehler != nil {
		t.Fatalf("nach Menge: %+v", antwort.Fehler)
	}

	if antwort.Exemplare[0].Name != "A" || antwort.Exemplare[1].Name != "B" {
		t.Errorf("nach Menge stehen %q und %q vorn",
			antwort.Exemplare[0].Name, antwort.Exemplare[1].Name)
	}

	// **Ein Filter senkt die Gesamtzahl, nicht die Zahl insgesamt.**
	gefiltert := fensterapp(doppel).Sammlungsfenster(t.Context(), "", "foil", "", 0, 0)
	if gefiltert.Fehler != nil {
		t.Fatalf("gefiltert: %+v", gefiltert.Fehler)
	}

	if gefiltert.Gesamt != 5 || gefiltert.Insgesamt != 10 {
		t.Errorf("gefiltert sind es %d von %d", gefiltert.Gesamt, gefiltert.Insgesamt)
	}

	// **Und die Suche ebenso.**
	gesucht := fensterapp(doppel).Sammlungsfenster(t.Context(), "scryfall:C", "", "", 0, 0)
	if gesucht.Gesamt != 1 || gesucht.Exemplare[0].Name != "C" {
		t.Errorf("gesucht wurden %d: %+v", gesucht.Gesamt, gesucht.Exemplare)
	}
}

// **Null heißt alles** — die Gruppenansicht braucht die ganze Liste.
func TestNullHeisstAlles(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsnahtdoppel{exemplare: zehnExemplare()}

	antwort := fensterapp(doppel).Sammlungsfenster(t.Context(), "", "", "", 0, 0)
	if len(antwort.Exemplare) != 10 {
		t.Errorf("es sind %d Zeilen", len(antwort.Exemplare))
	}
}

// **Eine erfundene Sicht wird abgewiesen** — statt still zur Vorgabe zu werden.
func TestEineErfundeneSammlungssichtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsnahtdoppel{exemplare: zehnExemplare()}

	for name, fall := range map[string][2]string{
		"Ausführung": {"glitzernd", ""},
		"Ordnung":    {"", "nach Farbe"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := fensterapp(doppel).Sammlungsfenster(
				t.Context(), "", fall[0], fall[1], 0, 10)

			if antwort.Fehler == nil {
				t.Error("die erfundene Sicht kam durch")
			}

			if antwort.Exemplare == nil {
				t.Error("die Liste ist null statt leer")
			}
		})
	}
}

// **Ohne Naht und mit Fehler kommt eine leere Liste und ein Fehler** — nicht
// null, und nicht stillschweigend nichts.
func TestDasSammlungsfensterMeldetSeineFehler(t *testing.T) {
	t.Parallel()

	ohne := app.NeueApp("/nirgends/bestand.db").
		Sammlungsfenster(t.Context(), "", "", "", 0, 10)
	if ohne.Fehler == nil || ohne.Exemplare == nil {
		t.Errorf("ohne Naht: %+v", ohne)
	}

	kaputt := fensterapp(&sammlungsnahtdoppel{fehler: errors.New("der Bestand ist zu")}).
		Sammlungsfenster(t.Context(), "", "", "", 0, 10)
	if kaputt.Fehler == nil || kaputt.Exemplare == nil {
		t.Errorf("mit Fehler: %+v", kaputt)
	}
}
