package service_test

// **Was die Doppel über Gruppen sagen** — abgetrennt am 2026-09-13, weil
// `kartenaufsicht_test.go` mit ihnen über die Zeilengrenze aus §2 ging (A13.1b, #1452).
//
// # Der Schnitt folgt der Frage
//
// Dort stehen die Doppel für den **einzelnen** Weg; hier die Antwort auf die
// eine Frage, die A13.1b hinzugefügt hat: was geschieht, wenn jemand eine
// **Gruppe** schreiben will. **Sie scheitern laut** (§3) — ein Doppel, das
// hier still gutginge, ließe einen Lauf über Gruppen grün werden, der gar
// keine geschrieben hat.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// GruppeSetzenMitHerkunft gibt es an diesem Doppel **nur der Schnittstelle
// wegen** (A13.1b, #1452) — und sie scheitert laut (§3).
func (*feldprobe) GruppeSetzenMitHerkunft(
	context.Context, []domain.Kartenschritt,
) error {
	return errors.New("dieses Doppel schreibt keine Gruppen")
}

// **Die Aufsicht hält auch die Gruppe** — Werkstück A13.1b (#1452).
//
// # Warum das ein eigener Lauf ist
//
// **Weil eine Gruppe die Prüfung an einer anderen Stelle braucht**: nicht vor
// dem Schreiben **dieser** Karte, sondern vor dem Schreiben der **ersten**.
// Eine Aufsicht, die mitten in der Schleife prüfte, hätte den Teilbestand, den
// A13.1b ausschließt.
func TestDieAufsichtHaeltAuchDieGruppe(t *testing.T) {
	protokoll := &lebenszyklusprotokoll{}
	felder := &feldprobe{}
	sicht := &kartensicht{karten: map[int64]int64{1: 10, 2: 10, 3: 20}}
	auge := service.NeueKartenaufsicht(sicht, protokoll).MitFeldweg(felder, felder)

	eineGruppe := func(karten ...int64) []domain.Kartenschritt {
		schritte := make([]domain.Kartenschritt, 0, len(karten))
		for _, karte := range karten {
			schritte = append(schritte, domain.Kartenschritt{
				KarteID: karte, SetID: 10,
				Werte:      []domain.Value{{Field: "name", Text: "x"}},
				Herkunft:   "ki",
				Herkuenfte: []domain.KIFeldherkunft{{Feld: "name"}},
			})
		}

		return schritte
	}

	t.Run("zwei eigene Karten kommen durch", func(t *testing.T) {
		// **Das Doppel scheitert laut** — geprüft wird deshalb, dass die
		// Aufsicht **bis dorthin** kommt und nicht vorher abweist.
		fehler := auge.GruppeVonAussenMitHerkunft(t.Context(), eineGruppe(1, 2))
		if errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("die Aufsicht wies eigene Karten ab: %v", fehler)
		}
	})

	t.Run("eine fremde Karte bricht die ganze Gruppe", func(t *testing.T) {
		// **Die fremde steht hinten** — und trotzdem wird die erste nicht
		// geschrieben: geprüft wird **vor** dem ersten Schreiben.
		fehler := auge.GruppeVonAussenMitHerkunft(t.Context(), eineGruppe(1, 3))
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler %v, erwartet war ErrNotFound", fehler)
		}
	})
}
