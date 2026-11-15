package storage_test

// **Mehrere Exemplare in einem Zug** — DOM-03.AK-02
// (#1698):
// die Fehlkarten eines Decks kommen **zusammen** auf die Wunschliste oder gar
// nicht. Eine halbe Liste sähe aus wie eine vollständige, und der nächste
// Versuch fände die erste Hälfte schon da.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

const (
	wunschdruckA = "a1b2c3d4-0000-4000-8000-0000000000a1"
	wunschdruckB = "a1b2c3d4-0000-4000-8000-0000000000b2"
)

func wunschzeilen() []*domain.Exemplar {
	return []*domain.Exemplar{
		{Printing: wunschdruckA, Menge: 3, Zustand: domain.ZustandNearMint, Sprache: "en", Wunsch: true, Name: "Blitz"},
		{Printing: wunschdruckB, Menge: 1, Zustand: domain.ZustandNearMint, Sprache: "en", Wunsch: true, Name: "Glut"},
	}
}

func TestMehrereExemplareEntstehenInEinemZug(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	kennungen, fehler := speicher.AlleAnlegen(probekontext(t), wunschzeilen())
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if len(kennungen) != 2 || kennungen[0] == kennungen[1] {
		t.Fatalf("Kennungen %v, erwartet zwei verschiedene", kennungen)
	}

	for stelle, kennung := range kennungen {
		exemplar, lesefehler := speicher.Einzeln(probekontext(t), kennung)
		if lesefehler != nil {
			t.Fatalf("lesen: %v", lesefehler)
		}

		erwartet := wunschzeilen()[stelle]
		if exemplar.Printing != erwartet.Printing || exemplar.Menge != erwartet.Menge ||
			!exemplar.Wunsch || exemplar.Name != erwartet.Name {
			t.Errorf("Zeile %d ist %+v", stelle, exemplar)
		}
	}
}

// **Scheitert eine, entsteht keine** — die zweite stößt an einen Wunsch, den
// es schon gibt, und die erste geht mit zurück.
func TestEinScheiterndesExemplarNimmtDieAnderenMit(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	if _, fehler := speicher.Anlegen(probekontext(t), wunschzeilen()[1]); fehler != nil {
		t.Fatalf("der vorhandene Wunsch: %v", fehler)
	}

	_, fehler := speicher.AlleAnlegen(probekontext(t), wunschzeilen())
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("Fehler = %v, erwartet ErrConflict", fehler)
	}

	if gezaehlt := zeilen(t, db, "sammlung"); gezaehlt != 1 {
		t.Errorf("%d Zeilen, erwartet nur den vorhandenen Wunsch", gezaehlt)
	}
}

// **Eine leere Liste schreibt nichts und ist kein Fehler.**
func TestEineLeereSammelanlageSchreibtNichts(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)

	kennungen, fehler := storage.NeuerSammlungsspeicher(db).AlleAnlegen(probekontext(t), nil)
	if fehler != nil || len(kennungen) != 0 {
		t.Errorf("Kennungen %v, Fehler %v", kennungen, fehler)
	}
}

// **Ohne Profil gibt es keinen Ort** — Vermerk 0066.
func TestEineSammelanlageOhneProfilSchreibtNichts(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)

	if _, fehler := storage.NeuerSammlungsspeicher(db).AlleAnlegen(
		context.Background(), wunschzeilen()); fehler == nil {
		t.Fatal("ohne Profil ging die Anlage durch")
	}

	if gezaehlt := zeilen(t, db, "sammlung"); gezaehlt != 0 {
		t.Errorf("%d Zeilen, erwartet keine", gezaehlt)
	}
}
