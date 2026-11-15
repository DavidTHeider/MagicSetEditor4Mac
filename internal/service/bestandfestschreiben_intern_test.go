package service

// **Ein Einsetzen, dessen Name nicht festgeschrieben ist, gilt als
// gescheitert** — BK-06b
// (#2131).
// Der Rückweg legt dann den alten Bestand zurück.
//
// **Nicht parallel**: er setzt eine Naht des Pakets.

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestEinNichtFestgeschriebenesEinsetzenLegtDenAltenBestandZurueck(t *testing.T) {
	voll := vollFestschreiben
	t.Cleanup(func() { vollFestschreiben = voll })

	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")
	sicherung := filepath.Join(ort, "sicherung.db")

	legeAn(t, bestand, "alt")
	legeAn(t, sicherung, "neu")

	// **Erst das Verzeichnis des Bestands scheitert** — das der
	// eingehenden Kopie ist dasselbe, also zählt der zweite Ruf.
	rufe := 0
	vollFestschreiben = func(d nachbardatei) error {
		if art(d) == "Verzeichnis" {
			rufe++
			if rufe == 2 {
				return errEingespielt
			}
		}

		return voll(d)
	}

	eingehend, err := eingehendeKopie(sicherung, bestand)
	if err != nil {
		t.Fatalf("die eingehende Kopie: %v", err)
	}

	if _, err := bestandErsetzen(eingehend, bestand); !errors.Is(err, errEingespielt) {
		t.Fatalf("erwartet war der eingespielte Fehler, gekommen ist %v", err)
	}

	if got := inhalt(t, bestand); got != "alt" {
		t.Errorf("am Ort des Bestands steht %q, erwartet war der alte", got)
	}
}
