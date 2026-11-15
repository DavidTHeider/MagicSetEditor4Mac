package service_test

// **Die Folgen einer Berichtigung** — DOM-04.AK-02
// (#1699).
//
// **Ein Lesen und kein Schreiben**: wer fragt, was eine Berichtigung
// verschöbe, hat noch nichts berichtigt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zweiGespielteRunden legt vier Namen an und spielt zwei Runden, in denen
// jeweils der Erste gewinnt. Es gibt die Kennung des Turniers und die der
// ersten Paarung.
func zweiGespielteRunden(t *testing.T, werk *turnierwerk) (turnierID, erste int64) {
	t.Helper()

	turnierID = werk.turnierAnlegen(t).ID
	werk.anmelden(t, turnierID, "Anna", "Bert", "Cleo", "Dora")

	for range 2 {
		runde, err := werk.leitung.RundeLegen(t.Context(), turnierID)
		if err != nil {
			t.Fatalf("die Runde: %v", err)
		}

		if erste == 0 {
			erste = runde.Paarungen[0].ID
		}

		for _, paarung := range runde.Paarungen {
			if err := werk.leitung.ErgebnisSetzen(t.Context(), turnierID,
				werk.fassung(t, turnierID), paarung.ID, &domain.Ergebnis{SpieleA: 2}); err != nil {
				t.Fatalf("das Ergebnis: %v", err)
			}
		}
	}

	return turnierID, erste
}

func TestEineBerichtigungSagtVorherWasSieVerschiebt(t *testing.T) {
	werk := neuesTurnierwerk(t)
	turnierID, erste := zweiGespielteRunden(t, werk)
	gesetzt := werk.turniere.gesetzt

	folgen, err := werk.leitung.Korrekturfolgen(t.Context(), turnierID, erste,
		&domain.Ergebnis{SpieleB: 2})
	if err != nil {
		t.Fatalf("die Folgen: %v", err)
	}

	if folgen.Runde != 1 || len(folgen.SpaetereRunden) != 1 || len(folgen.Verschoben) == 0 {
		t.Errorf("die Folgen: %+v", folgen)
	}

	// **Gefragt ist nicht geschrieben.**
	if werk.turniere.gesetzt != gesetzt {
		t.Errorf("die Frage hat %d Ergebnisse geschrieben", werk.turniere.gesetzt-gesetzt)
	}
}

// **Ein Stand, der kein Ergebnis ist, wird nicht durchgerechnet** — dieselbe
// Prüfung wie beim Eintragen.
func TestFolgenEinesUnmoeglichenStandes(t *testing.T) {
	werk := neuesTurnierwerk(t)
	turnierID, erste := zweiGespielteRunden(t, werk)

	if _, err := werk.leitung.Korrekturfolgen(t.Context(), turnierID, erste,
		&domain.Ergebnis{}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("0–0–0: %v, erwartet ErrValidation", err)
	}

	if _, err := werk.leitung.Korrekturfolgen(t.Context(), 999, erste, nil); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fehlendes Turnier: %v, erwartet ErrNotFound", err)
	}

	if _, err := werk.leitung.Korrekturfolgen(t.Context(), turnierID, 999, nil); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("eine fremde Paarung: %v, erwartet ErrNotFound", err)
	}
}
