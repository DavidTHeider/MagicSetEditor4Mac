// **Die Verdrahtung der abgestuften Rücksetzung** — Werkstück O4 von Phase
// 5.2 (#1428).
//
// # Was dieser Lauf hält
//
// Dass die Wurzel den Weg **erreicht** — nicht, was er tut. Der Arbeitsablauf
// über alle Flächen steht in `internal/flaechen`; hier geht es um die
// Verdrahtung, die sonst niemand ausführt.
package main

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieDreiRuecksetzungswegeErreichenIhreNaht(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	ort := filepath.Join(zuhause, "bestand.db")

	db, fehler := storage.Oeffnen(probekontext(t), ort)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	auskunft := ruecksetzungsauskunft{db: db}

	arten, fehler := auskunft.Ruecksetzungsarten(probekontext(t))
	if fehler != nil {
		t.Fatalf("die Arten: %v", fehler)
	}

	reihe, richtig := arten.([]string)
	if !richtig || len(reihe) != 6 {
		t.Errorf("die Reihe ist %v", arten)
	}

	if _, err := auskunft.Ruecksetzungsvorschau(
		probekontext(t), "layout", "", ""); err != nil {
		t.Errorf("die Vorschau: %v", err)
	}

	if _, err := auskunft.Zuruecksetzen(
		probekontext(t), "layout", "", "", "", "", ""); err != nil {
		t.Errorf("zurücksetzen: %v", err)
	}

	// **Und eine erfundene Art kommt als Fehler zurück** — aus dem Fachkern
	// und nicht aus der Verdrahtung.
	if _, err := auskunft.Ruecksetzungsvorschau(
		probekontext(t), "alles", "", ""); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("eine erfundene Art: %v", err)
	}

	// **Und die ganze Anwendung geht hier gar nicht** (G09, #2273): der Dienst
	// tauscht seinen Bestand nicht, und das sagt das Hindernis — nicht der
	// fehlende Nachweis.
	if _, err := auskunft.Zuruecksetzen(
		probekontext(t), "anwendung", "", "", "", "", ""); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("die ganze Anwendung über den Dienst: %v — erwartet ein Konflikt", err)
	}
}
