package peer

// **Die Nonceliste wächst nicht ohne Ende** — Befund Z3 des Zweitblicks über
// Phase 6.7 (NET-Teil): jedes gültige Angebot vermerkte seine Zufallszahl für
// eine Stunde, ohne Obergrenze.

import (
	"fmt"
	"testing"
	"time"
)

func TestDieNoncelisteHaeltIhreGrenze(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	liste := NeueNonceliste(func() time.Time { return jetzt }, time.Hour)
	liste.hoechstens = 3

	for nummer := range 4 {
		if fehler := liste.Merke(fmt.Sprintf("n%d", nummer)); fehler != nil {
			t.Fatalf("n%d: %v", nummer, fehler)
		}
	}

	if len(liste.gesehen) != 3 {
		t.Errorf("die Liste hält %d", len(liste.gesehen))
	}

	// **Die älteste ist gegangen** — die jüngste ist noch eine Wiederholung.
	if fehler := liste.Merke("n3"); fehler == nil {
		t.Error("die jüngste wurde ein zweites Mal genommen")
	}

	if fehler := liste.Merke("n0"); fehler != nil {
		t.Errorf("die älteste ist noch da: %v", fehler)
	}
}

// **Und die Frist räumt weiter** — von vorn, ohne die ganze Liste zu lesen.
func TestDieFristRaeumtVonVorn(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	liste := NeueNonceliste(func() time.Time { return jetzt }, time.Hour)

	if fehler := liste.Merke("alt"); fehler != nil {
		t.Fatal(fehler)
	}

	jetzt = jetzt.Add(2 * time.Hour)

	if fehler := liste.Merke("neu"); fehler != nil {
		t.Fatal(fehler)
	}

	if _, da := liste.gesehen["alt"]; da || len(liste.reihe) != 1 {
		t.Errorf("geblieben: %v, Reihe %d", liste.gesehen, len(liste.reihe))
	}
}
