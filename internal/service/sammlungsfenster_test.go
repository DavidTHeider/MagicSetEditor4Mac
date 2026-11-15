// **Der Ausschnitt einer Sammlung** — Werkstück
// #1500.
//
// Geprüft wird der Weg über **zwei** Abfragen: was geordnet wird, ist die
// schmale Liste; was herauskommt, sind die vollen Zeilen des Ausschnitts. Ein
// Doppel, das beides gleich beantwortete, prüfte die Naht und nicht die Regel.
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// fensterdoppel hält eine Sammlung im Arbeitsspeicher — schmal und voll.
type fensterdoppel struct {
	schmal []*domain.Exemplar
	voll   map[int64]*domain.Exemplar

	gefragt     []int64
	lesefehler  error
	holefehler  error
	ordnungsruf int
}

func (f *fensterdoppel) Fenster(
	_ context.Context, waehle func([]*domain.Exemplar) []int64,
) ([]*domain.Exemplar, error) {
	f.ordnungsruf++

	if f.lesefehler != nil {
		return nil, f.lesefehler
	}

	kennungen := waehle(f.schmal)
	f.gefragt = kennungen

	if f.holefehler != nil {
		return nil, f.holefehler
	}

	volle := make([]*domain.Exemplar, 0, len(kennungen))
	for _, kennung := range kennungen {
		if eines, da := f.voll[kennung]; da {
			volle = append(volle, eines)
		}
	}

	return volle, nil
}

// dreiExemplare gibt ein Doppel mit drei Zeilen — zwei Nonfoil, eine Foil.
func dreiExemplare() *fensterdoppel {
	schmal := []*domain.Exemplar{
		{ID: 1, Name: "Blitzschlag", Menge: 1, Notiz: "Zauber"},
		{ID: 2, Name: "Ätherfluss", Menge: 4, Notiz: "Zauber"},
		{
			ID: 3, Name: "Zauberpult", Menge: 2,
			Finish: domain.FinishFoil, Notiz: "Zauber",
		},
	}

	voll := map[int64]*domain.Exemplar{}
	for _, eines := range schmal {
		voll[eines.ID] = &domain.Exemplar{
			ID: eines.ID, Name: eines.Name, Menge: eines.Menge,
			Zustand: domain.ZustandNearMint, Sprache: "de",
		}
	}

	return &fensterdoppel{schmal: schmal, voll: voll}
}

// **Die Sicht ordnet die schmale Liste, geholt wird nur der Ausschnitt.**
func TestDasSammlungsfensterHoltNurDenAusschnitt(t *testing.T) {
	t.Parallel()

	doppel := dreiExemplare()

	fenster, zahlen, fehler := service.NeuesSammlungsfenster(doppel).Sicht(
		t.Context(), domain.Sammlungssicht{
			Ausfuehrung: domain.AusfuehrungNonfoil,
			Sortierung:  domain.SortierungMenge,
		}, 0, 1)
	if fehler != nil {
		t.Fatalf("Sicht: %v", fehler)
	}

	if zahlen.Gesamt != 2 || zahlen.Insgesamt != 3 {
		t.Errorf("gesehen %d, insgesamt %d — erwartet waren 2 und 3", zahlen.Gesamt, zahlen.Insgesamt)
	}

	// **Nur eine Kennung wurde gefragt** — das ist der Sinn der zwei Abfragen.
	if len(doppel.gefragt) != 1 || doppel.gefragt[0] != 2 {
		t.Errorf("gefragt wurde nach %v", doppel.gefragt)
	}

	// **Und was herauskommt, ist die volle Zeile**, nicht die schmale.
	if len(fenster) != 1 || fenster[0].Zustand != domain.ZustandNearMint {
		t.Errorf("das Fenster ist %+v", fenster)
	}
}

// **Eine Anzahl von null heißt alles** — die Gruppenansicht braucht es so.
func TestDasSammlungsfensterOhneAnzahlGibtAlles(t *testing.T) {
	t.Parallel()

	doppel := dreiExemplare()

	fenster, zahlen, fehler := service.NeuesSammlungsfenster(doppel).Sicht(
		t.Context(), domain.Sammlungssicht{}, 0, 0)
	if fehler != nil {
		t.Fatalf("Sicht: %v", fehler)
	}

	if len(fenster) != 3 || zahlen.Gesamt != 3 {
		t.Errorf("es sind %d Zeilen von %d", len(fenster), zahlen.Gesamt)
	}
}

// **Eine Sicht, die es nicht gibt, wird abgewiesen — bevor gelesen wird.**
func TestDasSammlungsfensterWeistEineUnbekannteSichtAb(t *testing.T) {
	t.Parallel()

	doppel := dreiExemplare()

	_, _, fehler := service.NeuesSammlungsfenster(doppel).Sicht(
		t.Context(), domain.Sammlungssicht{Sortierung: "farbe"}, 0, 10)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("abgewiesen wurde mit %v", fehler)
	}

	if doppel.ordnungsruf != 0 {
		t.Error("gelesen wurde trotzdem")
	}
}

// **Beide Fehlerwege kommen heraus.**
func TestDasSammlungsfensterMeldetBeideFehler(t *testing.T) {
	t.Parallel()

	for name, bauen := range map[string]func() *fensterdoppel{
		"beim Lesen der Ordnung": func() *fensterdoppel {
			doppel := dreiExemplare()
			doppel.lesefehler = errors.New("die Ablage schweigt")

			return doppel
		},
		"beim Holen der Zeilen": func() *fensterdoppel {
			doppel := dreiExemplare()
			doppel.holefehler = errors.New("die Ablage schweigt")

			return doppel
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := service.NeuesSammlungsfenster(bauen()).Sicht(
				t.Context(), domain.Sammlungssicht{}, 0, 10)
			if fehler == nil {
				t.Error("der Fehler kam nicht heraus")
			}
		})
	}
}
