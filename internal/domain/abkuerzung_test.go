package domain_test

// **Abkürzungen** — Werkstück M1 der Regelmaschine
// (#1603).
//
// # Wogegen hier geprüft wird
//
// **CR 732.1**: am Tisch benutzt niemand jede einzelne Entscheidung, sondern
// verabredete Abkürzungen. **CR 732.2a**: wer Priorität hat, schlägt eine
// Folge vor, die **vorhersehbar** ist.
//
// **Die Zahl, um die es geht, steht seit G9 in der Regelmatrix**: zehn der
// zwölf Abschnitte vergeben Priorität, bei zwei Sitzen sind das **zwanzig
// Passen je Zug**, in denen nichts geschieht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Abkürzung reicht bis zum Ende des Zuges durch.**
func TestEineAbkuerzungReichtBisZumZugende(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie().ZugBeginnen("sitz-1")
	partie.Objekte = nil

	vorher := partie.Zug.Nummer

	danach, fehler := partie.Abkuerzen(partie.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("abkürzen: %v", fehler)
	}

	switch {
	case danach.Zug.Nummer != vorher+1:
		t.Errorf("der Zug steht bei %d statt %d", danach.Zug.Nummer, vorher+1)
	case danach.Zug.Aktiv != "sitz-2":
		t.Errorf("am Zug ist %q", danach.Zug.Aktiv)
	}
}

// **Sie hält an, wo etwas auf dem Stapel liegt** — CR 732.2a: was dort
// wartet, ist keine vorhersehbare Folge, sondern eine Handlung.
func TestEineAbkuerzungHaeltAmStapelAn(t *testing.T) {
	t.Parallel()

	partie := eineStapelpartie(t)
	vorher := partie.Zug.Abschnitt

	danach, fehler := partie.Abkuerzen(partie.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("abkürzen: %v", fehler)
	}

	if danach.Zug.Abschnitt != vorher {
		t.Fatalf("der Abschnitt ist %+v statt %+v",
			danach.Zug.Abschnitt, vorher)
	}
}

// **Und sie hält an, wo jemand abwerfen muss** — CR 514.1: das ist eine
// Wahl, und eine Wahl ist nicht vorhersehbar.
func TestEineAbkuerzungHaeltVorDerAbwurfpflichtAn(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie().ZugBeginnen("sitz-1")
	partie.Objekte = handvoll("sitz-1", domain.Handgrenze+2)

	danach, fehler := partie.Abkuerzen(partie.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("abkürzen: %v", fehler)
	}

	if danach.Zug.Abwurf != 2 {
		t.Fatalf("abzuwerfen sind %d", danach.Zug.Abwurf)
	}
}

// **Wer nicht an der Reihe ist, kürzt nicht ab** — CR 732.2a: vorschlagen
// darf, wer Priorität hat.
func TestWerNichtAnDerReiheIstKuerztNichtAb(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie().ZugBeginnen("sitz-1")

	if _, fehler := partie.Abkuerzen("sitz-2"); fehler == nil {
		t.Fatal("der Sitz ohne Priorität hat abgekürzt")
	}
}

// **Eine beendete Partie kürzt nichts ab.**
func TestEineBeendetePartieKuerztNichtAb(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie().ZugBeginnen("sitz-1")
	wer := partie.Zug.Prioritaet
	partie = partie.Scheidetaus("sitz-2", domain.GrundAufgabe)

	if _, fehler := partie.Abkuerzen(wer); fehler == nil {
		t.Fatal("die beendete Partie hat abgekürzt")
	}
}

// handvoll gibt so viele Karten auf der Hand eines Sitzes.
func handvoll(wer string, wie int) []domain.Spielobjekt {
	objekte := make([]domain.Spielobjekt, 0, wie)

	for i := range wie {
		objekte = append(objekte, domain.Spielobjekt{
			Kennung: wer + "/h" + string(rune('a'+i)), Art: domain.ObjektKarte,
			Beherrscher: wer, Eigner: wer, Zone: domain.ZoneHand,
		})
	}

	return objekte
}
