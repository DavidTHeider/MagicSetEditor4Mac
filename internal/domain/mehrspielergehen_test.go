package domain_test

// **Wer mitten im Spiel geht** — CR 800.4a und CR 800.4j (Fassung
// 2026-09-25), #1971.
//
// # Was hier fehlte
//
// **M13 hat die Reihenfolge gebaut, die einen Gegangenen überspringt** —
// beim Passen und beim nächsten Zug ([TestDiePrioritaetUeberspringtDenGegangenen]).
// Nicht gebaut war, was in dem Augenblick geschieht, in dem jemand geht:
// hielt er die Priorität, behielt er sie, und die anderen konnten nicht
// passen. `FuzzZugfolge` hat es beim Bau von QA-06b mit drei Sitzen
// nachgestellt (#1967).
//
// **CR 800.4a**: die Priorität geht an den Nächsten in Zugreihenfolge, der
// noch im Spiel ist. **CR 800.4j**: geht der aktive Spieler in seinem Zug,
// läuft der Zug ohne aktiven Spieler zu Ende; wo er Priorität bekäme,
// bekommt sie der Nächste in Zugreihenfolge.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Wer mit der Priorität geht, gibt sie weiter** — CR 800.4a.
func TestWerMitDerPrioritaetGehtGibtSieWeiter(t *testing.T) {
	t.Parallel()

	partie, fehler := eineDreierpartie(t).Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	danach := partie.Scheidetaus("sitz-2", domain.GrundAufgabe)

	if danach.Stand == domain.StandBeendet {
		t.Fatal("die Partie ist beendet — zu dritt läuft sie weiter")
	}

	if danach.Zug.Prioritaet != "sitz-3" {
		t.Fatalf("nach dem Gehen von sitz-2 hat %q Priorität", danach.Zug.Prioritaet)
	}

	// **Und sie lässt sich weitergeben** — der Beleg, dass die Partie nicht
	// hängt: sitz-3 passt, und die Runde ist vollständig.
	vorher := danach.Zug.Abschnitt

	weiter, fehler := danach.Passen("sitz-3")
	if fehler != nil {
		t.Fatalf("sitz-3 passt nicht: %v", fehler)
	}

	if weiter.Zug.Abschnitt == vorher {
		t.Errorf("sitz-1 und sitz-3 haben gepasst, und der Abschnitt steht noch bei %+v", vorher)
	}
}

// **Wer gegangen ist, hat nicht mehr gepasst** — CR 117.4 mit CR 800.4a.
//
// „Alle passen nacheinander" meint die, die noch im Spiel sind. Ein Passen
// von vorhin zählte sonst für einen, der nicht mehr da ist, und der Abschnitt
// endete, bevor sitz-3 gefragt war.
func TestEinGegangenerHatNichtMehrGepasst(t *testing.T) {
	t.Parallel()

	partie, fehler := eineDreierpartie(t).Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	partie = partie.Scheidetaus("sitz-1", domain.GrundAufgabe)
	vorher := partie.Zug.Abschnitt

	danach, fehler := partie.Passen("sitz-2")
	if fehler != nil {
		t.Fatalf("sitz-2 passt nicht: %v", fehler)
	}

	if danach.Zug.Abschnitt != vorher || danach.Zug.Prioritaet != "sitz-3" {
		t.Errorf("nach dem Passen von sitz-2 steht %+v, Priorität %q — sitz-3 war nicht gefragt",
			danach.Zug.Abschnitt, danach.Zug.Prioritaet)
	}
}

// **Der Zug des Gegangenen läuft ohne ihn zu Ende** — CR 800.4j.
func TestDerZugDesGegangenenLaeuftOhneIhnZuEnde(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t).Scheidetaus("sitz-1", domain.GrundAufgabe)

	if partie.Zug.Prioritaet != "sitz-2" {
		t.Fatalf("nach dem Gehen des aktiven Spielers hat %q Priorität", partie.Zug.Prioritaet)
	}

	vorher := partie.Zug.Abschnitt

	for _, wer := range []string{"sitz-2", "sitz-3"} {
		gepasst, fehler := partie.Passen(wer)
		if fehler != nil {
			t.Fatalf("%s passt nicht: %v", wer, fehler)
		}

		partie = gepasst
	}

	// **Im nächsten Schritt bekäme der aktive Spieler die Priorität** —
	// statt seiner bekommt sie der Nächste in Zugreihenfolge.
	switch {
	case partie.Zug.Abschnitt == vorher:
		t.Fatalf("zwei Passen, und der Abschnitt steht noch bei %+v", vorher)
	case partie.Zug.Nummer != 1:
		t.Fatalf("der Zug ist schon %d — er sollte zu Ende laufen", partie.Zug.Nummer)
	case partie.Zug.Prioritaet != "sitz-2":
		t.Errorf("im Abschnitt %+v hat %q Priorität", partie.Zug.Abschnitt, partie.Zug.Prioritaet)
	}

	// **Und der nächste Zug ist der von sitz-2** — CR 800.4k.
	if naechster := partie.ZugBeenden().Weiter(); naechster.Zug.Aktiv != "sitz-2" {
		t.Errorf("nach dem Zug des Gegangenen ist %q dran", naechster.Zug.Aktiv)
	}
}

// **Eine Abwurfpflicht geht mit dem, der sie hatte** — CR 800.4j.
//
// Seine Hand verlässt mit ihm das Spiel (CR 800.4a), und abwerfen kann nur,
// wer am Zug ist. Bliebe die Pflicht, stünde der Aufräumschritt für immer.
func TestEineAbwurfpflichtGehtMitDemDerSieHatte(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)
	partie.Objekte = handvoll("sitz-1", domain.Handgrenze+2)
	partie.Zug.Abschnitt = domain.Abschnitt{Phase: domain.PhaseEnde, Schritt: domain.SchrittEnde}
	partie.Zug.Prioritaet = "sitz-1"

	for _, wer := range []string{"sitz-1", "sitz-2", "sitz-3"} {
		gepasst, fehler := partie.Passen(wer)
		if fehler != nil {
			t.Fatalf("%s passt nicht: %v", wer, fehler)
		}

		partie = gepasst
	}

	if partie.Zug.Abwurf != 2 {
		t.Fatalf("die Vorrichtung trägt nicht: abzuwerfen sind %d", partie.Zug.Abwurf)
	}

	danach := partie.Scheidetaus("sitz-1", domain.GrundAufgabe)

	switch {
	case danach.Zug.Abwurf != 0:
		t.Errorf("die Abwurfpflicht steht noch: %d", danach.Zug.Abwurf)
	case danach.Zug.Aktiv != "sitz-2" || danach.Zug.Nummer != 2:
		t.Errorf("der Zug lief nicht zu Ende: Zug %d, aktiv %q", danach.Zug.Nummer, danach.Zug.Aktiv)
	case danach.Zug.Prioritaet == "":
		t.Errorf("im Abschnitt %+v hat niemand Priorität", danach.Zug.Abschnitt)
	}
}
