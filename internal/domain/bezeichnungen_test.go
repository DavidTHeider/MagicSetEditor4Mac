package domain_test

// **Bezeichnungen und Sonderzustände** — Werkstück M20 der Regelmaschine
// (#1621).
//
// # Was eine Bezeichnung ist
//
// **Etwas, das ein Spieler *hat*, ohne dass es ein Objekt wäre.** CR 725.1
// sagt es für den Monarchen, CR 726.1 für die Initiative: es gibt sie in
// einer Partie erst, wenn ein Effekt sie vergibt — und bis dahin hat sie
// niemand.
//
// # Und was das Spiel selbst haben kann
//
// **Tag und Nacht** (CR 731.1) gehören **der Partie** und keinem Spieler.
// Sie fängt mit keinem von beidem an, und der Wechsel geschieht im
// Enttappschritt (CR 502.2) — nach der Zahl der Zauber, die der vorige
// aktive Spieler gewirkt hat (CR 731.2a, CR 731.2b).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Partie fängt ohne Bezeichnung an** — CR 725.1, CR 726.1, CR 731.1.
func TestEinePartieFaengtOhneBezeichnungAn(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)

	switch {
	case partie.Monarch != "":
		t.Errorf("Monarch ist %q", partie.Monarch)
	case partie.Initiative != "":
		t.Errorf("die Initiative hat %q", partie.Initiative)
	case partie.Tageszeit != "":
		t.Errorf("die Tageszeit ist %q", partie.Tageszeit)
	}
}

// **Nur einer ist Monarch** — CR 725.3.
func TestNurEinerIstMonarch(t *testing.T) {
	t.Parallel()

	erst, fehler := eineDreierpartie(t).Monarchwerden("sitz-1")
	if fehler != nil {
		t.Fatalf("monarch: %v", fehler)
	}

	if erst.Monarch != "sitz-1" {
		t.Fatalf("Monarch ist %q", erst.Monarch)
	}

	zweit, fehler := erst.Monarchwerden("sitz-2")
	if fehler != nil {
		t.Fatalf("zweiter monarch: %v", fehler)
	}

	if zweit.Monarch != "sitz-2" {
		t.Errorf("Monarch ist %q — CR 725.3 lässt nur einen zu", zweit.Monarch)
	}
}

// **Und nur einer hat die Initiative** — CR 726.3.
func TestNurEinerHatDieInitiative(t *testing.T) {
	t.Parallel()

	erst, fehler := eineDreierpartie(t).Initiativenehmen("sitz-1")
	if fehler != nil {
		t.Fatalf("initiative: %v", fehler)
	}

	zweit, fehler := erst.Initiativenehmen("sitz-3")
	if fehler != nil {
		t.Fatalf("zweite initiative: %v", fehler)
	}

	if zweit.Initiative != "sitz-3" {
		t.Errorf("die Initiative hat %q", zweit.Initiative)
	}
}

// **Was keine Bezeichnung bekommt.**
func TestWasKeineBezeichnungBekommt(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t).Scheidenaus([]domain.Ausgeschieden{
		{Sitz: "sitz-3", Grund: domain.GrundAufgabe},
	})

	faelle := map[string]string{
		"ein Sitz, den es nicht gibt": "sitz-9",
		"ein Gegangener":              "sitz-3",
	}

	for name, wer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := partie.Monarchwerden(wer); fehler == nil {
				t.Errorf("%s wurde Monarch", name)
			}

			if _, fehler := partie.Initiativenehmen(wer); fehler == nil {
				t.Errorf("%s nahm die Initiative", name)
			}
		})
	}
}

// **Geht der Monarch, wird es der aktive Spieler** — CR 725.4.
func TestGehtDerMonarchWirdEsDerAktiveSpieler(t *testing.T) {
	t.Parallel()

	partie, fehler := eineDreierpartie(t).Monarchwerden("sitz-2")
	if fehler != nil {
		t.Fatalf("monarch: %v", fehler)
	}

	danach := partie.Scheidenaus([]domain.Ausgeschieden{
		{Sitz: "sitz-2", Grund: domain.GrundAufgabe},
	})

	if danach.Monarch != "sitz-1" {
		t.Errorf("Monarch ist %q — sitz-1 ist der aktive Spieler",
			danach.Monarch)
	}
}

// **Und geht der aktive Spieler selbst, nimmt es der Nächste** — CR 725.4,
// zweiter Satz.
func TestGehtDerAktiveMonarchNimmtEsDerNaechste(t *testing.T) {
	t.Parallel()

	partie, fehler := eineDreierpartie(t).Monarchwerden("sitz-1")
	if fehler != nil {
		t.Fatalf("monarch: %v", fehler)
	}

	danach := partie.Scheidenaus([]domain.Ausgeschieden{
		{Sitz: "sitz-1", Grund: domain.GrundAufgabe},
	})

	if danach.Monarch != "sitz-2" {
		t.Errorf("Monarch ist %q", danach.Monarch)
	}
}

// **Die Initiative geht denselben Weg** — CR 726.4.
func TestDieInitiativeGehtDenselbenWeg(t *testing.T) {
	t.Parallel()

	partie, fehler := eineDreierpartie(t).Initiativenehmen("sitz-2")
	if fehler != nil {
		t.Fatalf("initiative: %v", fehler)
	}

	danach := partie.Scheidenaus([]domain.Ausgeschieden{
		{Sitz: "sitz-2", Grund: domain.GrundAufgabe},
	})

	if danach.Initiative != "sitz-1" {
		t.Errorf("die Initiative hat %q", danach.Initiative)
	}
}

// **Radioaktivitätsmarken liegen am Spieler** — CR 728.1 mit CR 122.
func TestRadioaktivitaetsmarkenLiegenAmSpieler(t *testing.T) {
	t.Parallel()

	danach, fehler := eineDreierpartie(t).Strahlung("sitz-2", 3)
	if fehler != nil {
		t.Fatalf("strahlung: %v", fehler)
	}

	sitz := danach.SitzVon("sitz-2")
	if sitz == nil || sitz.Strahlung != 3 {
		t.Fatalf("die Marken stehen bei %+v", sitz)
	}

	// **Und unter null gibt es keine** — dieselbe Regel wie bei jedem
	// anderen Zähler (CR 122.1a).
	weniger, fehler := danach.Strahlung("sitz-2", -9)
	if fehler != nil {
		t.Fatalf("abnehmen: %v", fehler)
	}

	if sitz := weniger.SitzVon("sitz-2"); sitz == nil || sitz.Strahlung != 0 {
		t.Errorf("die Marken stehen bei %+v", sitz)
	}
}

// **Ein Sitz, den es nicht gibt, bekommt keine Marken** — §8.
func TestEinUnbekannterSitzBekommtKeineStrahlung(t *testing.T) {
	t.Parallel()

	_, fehler := eineDreierpartie(t).Strahlung("sitz-9", 1)
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("ein unbekannter Sitz: %v", fehler)
	}
}

// **Tag wird Nacht, wenn im vorigen Zug kein Zauber kam** — CR 731.2a mit
// CR 502.2.
func TestTagWirdNachtOhneZauber(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)
	partie.Tageszeit = domain.TageszeitTag

	danach := partie.ZugBeenden().Weiter()

	if danach.Tageszeit != domain.TageszeitNacht {
		t.Errorf("die Tageszeit ist %q", danach.Tageszeit)
	}
}

// **Und Nacht wird Tag, wenn zwei kamen** — CR 731.2b.
func TestNachtWirdTagBeiZweiZaubern(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)
	partie.Tageszeit = domain.TageszeitNacht
	partie.Zug.Zauber = 2

	danach := partie.ZugBeenden().Weiter()

	if danach.Tageszeit != domain.TageszeitTag {
		t.Errorf("die Tageszeit ist %q", danach.Tageszeit)
	}
}

// **Bei einem bleibt es Nacht** — CR 731.2b verlangt zwei.
//
// **Die Gegenprobe**: ohne sie sagte der Lauf darüber nur, dass es
// **irgendwann** Tag wird.
func TestBeiEinemZauberBleibtEsNacht(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)
	partie.Tageszeit = domain.TageszeitNacht
	partie.Zug.Zauber = 1

	danach := partie.ZugBeenden().Weiter()

	if danach.Tageszeit != domain.TageszeitNacht {
		t.Errorf("die Tageszeit ist %q", danach.Tageszeit)
	}
}

// **Und ohne Tageszeit wird gar nicht geprüft** — CR 731.2c.
func TestOhneTageszeitWirdNichtGeprueft(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)

	danach := partie.ZugBeenden().Weiter()

	if danach.Tageszeit != "" {
		t.Errorf("aus nichts wurde %q", danach.Tageszeit)
	}
}

// **Der Zauberzähler geht mit dem Zug auf null** — CR 731.2a fragt nach dem
// **vorigen** Zug.
func TestDerZauberzaehlerGehtMitDemZugAufNull(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)
	partie.Tageszeit = domain.TageszeitNacht
	partie.Zug.Zauber = 2

	erst := partie.ZugBeenden().Weiter()
	if erst.Zug.Zauber != 0 {
		t.Fatalf("der Zähler steht bei %d", erst.Zug.Zauber)
	}

	// **Und im nächsten Zug wird es wieder Nacht** — niemand hat gewirkt.
	zweit := erst.ZugBeenden().Weiter()
	if zweit.Tageszeit != domain.TageszeitNacht {
		t.Errorf("die Tageszeit ist %q", zweit.Tageszeit)
	}
}

// **Gehen alle, hat niemand mehr die Bezeichnung** — CR 725.4 bis zu Ende
// gedacht.
//
// **Den Fall gibt es im Spiel nicht** (CR 104.2a beendet die Partie, sobald
// einer übrig ist), und er steht trotzdem hier: ein Rückfall, den kein Lauf
// betritt, ist ein Versprechen, das niemand hält (§10).
func TestGehenAlleHatNiemandMehrDieBezeichnung(t *testing.T) {
	t.Parallel()

	partie, fehler := eineDreierpartie(t).Monarchwerden("sitz-1")
	if fehler != nil {
		t.Fatalf("monarch: %v", fehler)
	}

	partie.Zug.Aktiv = ""
	partie.Ausgang.Ausgeschieden = []domain.Ausgeschieden{
		{Sitz: "sitz-1", Grund: domain.GrundAufgabe},
		{Sitz: "sitz-2", Grund: domain.GrundAufgabe},
		{Sitz: "sitz-3", Grund: domain.GrundAufgabe},
	}

	if danach := partie.Scheidenaus(nil); danach.Monarch != "" {
		t.Errorf("Monarch ist %q", danach.Monarch)
	}
}
