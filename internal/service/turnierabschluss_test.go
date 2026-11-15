package service_test

// **Abschluss eines Turniers und was danach gilt** — DOM-04.AK-04
// (#1699),
// Vermerk 0139.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// abgeschlossen spielt eine Runde zu Ende und schließt das Turnier ab.
func abgeschlossen(t *testing.T, werk *turnierwerk) (turnierID int64, teilnehmer []int64, paarung int64) {
	t.Helper()

	turnierID = werk.turnierAnlegen(t).ID
	teilnehmer = werk.anmelden(t, turnierID, "Anna", "Bert")

	runde, err := werk.leitung.RundeLegen(t.Context(), turnierID)
	if err != nil {
		t.Fatalf("die Runde: %v", err)
	}

	paarung = runde.Paarungen[0].ID

	// **Mit offenem Tisch schließt nichts ab.**
	var feldfehler *domain.FieldError
	if err := werk.leitung.Abschliessen(t.Context(), turnierID, werk.fassung(t, turnierID)); !errors.As(err, &feldfehler) ||
		feldfehler.Key != domain.SchluesselTurnierNichtFertig {
		t.Fatalf("mit offenem Tisch: %v, erwartet turnierNichtFertig", err)
	}

	if err := werk.leitung.ErgebnisSetzen(t.Context(), turnierID, werk.fassung(t, turnierID),
		paarung, &domain.Ergebnis{SpieleA: 2}); err != nil {
		t.Fatalf("das Ergebnis: %v", err)
	}

	if err := werk.leitung.Abschliessen(t.Context(), turnierID, werk.fassung(t, turnierID)); err != nil {
		t.Fatalf("abschließen: %v", err)
	}

	return turnierID, teilnehmer, paarung
}

func TestEinAbgeschlossenesTurnierWeistJedeAenderungAb(t *testing.T) {
	werk := neuesTurnierwerk(t)
	turnierID, teilnehmer, paarung := abgeschlossen(t, werk)

	faelle := map[string]func() error{
		"anmelden": func() error {
			_, err := werk.leitung.Anmelden(t.Context(), turnierID, domain.Teilnehmer{Name: "Cleo"})

			return err
		},
		"aussteigen": func() error {
			return werk.leitung.Aussteigen(t.Context(), turnierID, teilnehmer[1], true)
		},
		"runde legen": func() error {
			_, err := werk.leitung.RundeLegen(t.Context(), turnierID)

			return err
		},
		"Ergebnis setzen": func() error {
			return werk.leitung.ErgebnisSetzen(t.Context(), turnierID, werk.fassung(t, turnierID), paarung, nil)
		},
	}

	for name, aendern := range faelle {
		var feldfehler *domain.FieldError

		err := aendern()
		if !errors.Is(err, domain.ErrConflict) || !errors.As(err, &feldfehler) ||
			feldfehler.Key != domain.SchluesselTurnierAbgeschlossen {
			t.Errorf("%s: %v, erwartet turnierAbgeschlossen", name, err)
		}
	}

	// **Kopieren geht, und die Kopie ist offen** (Vermerk 0139).
	kopie, err := werk.leitung.Kopieren(t.Context(), turnierID, "Freitag II", true)
	if err != nil || kopie.IstAbgeschlossen() {
		t.Errorf("die Kopie: %+v, %v", kopie, err)
	}

	// **Wiedereröffnet nimmt es wieder an.**
	if err := werk.leitung.Wiedereroeffnen(t.Context(), turnierID, werk.fassung(t, turnierID)); err != nil {
		t.Fatalf("wiedereröffnen: %v", err)
	}

	if _, err := werk.leitung.Anmelden(t.Context(), turnierID, domain.Teilnehmer{Name: "Cleo"}); err != nil {
		t.Errorf("anmelden nach dem Wiedereröffnen: %v", err)
	}
}

// **Abschluss, Wiedereröffnen und Protokoll an einem fehlenden Turnier sind
// ein Nein** — und das Protokoll kommt vom Speicher.
func TestAbschlussUndProtokollBrauchenEinTurnier(t *testing.T) {
	werk := neuesTurnierwerk(t)

	for name, ruf := range map[string]func() error{
		"abschließen":    func() error { return werk.leitung.Abschliessen(t.Context(), 99, 1) },
		"wiedereröffnen": func() error { return werk.leitung.Wiedereroeffnen(t.Context(), 99, 1) },
		"Protokoll": func() error {
			_, err := werk.leitung.Protokoll(t.Context(), 99)

			return err
		},
	} {
		if err := ruf(); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s: %v, erwartet ErrNotFound", name, err)
		}
	}

	turnierID := werk.turnierAnlegen(t).ID
	werk.turniere.protokoll = []domain.Protokolleintrag{{Vorgang: domain.TurnierAngelegt, Angaben: "Küchentisch"}}

	protokoll, err := werk.leitung.Protokoll(t.Context(), turnierID)
	if err != nil || len(protokoll) != 1 || protokoll[0].Vorgang != domain.TurnierAngelegt {
		t.Errorf("das Protokoll: %+v, %v", protokoll, err)
	}
}
