package service_test

// **Prüfen ohne Anlegen** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// **Eine eigene Datei**, seit `kiauftragswerk_test.go` an die Zeilengrenze
// aus §2 kam. Der Schnitt folgt der Frage: dort steht, was ein **Auftrag**
// ist — Revision, Schema, Frist, Übernahme —, hier, was gilt, wenn es keinen
// gibt: eine Frage ohne Zielkarte.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Prüfen legt keinen Auftrag an — und kostet nichts.**
//
// # Wofür es das gibt
//
// **Für eine Frage ohne Zielkarte** (A13.3, #1452). Eine Bilderkennung liest
// ein Foto; welche Karte darauf ist, ist ja gerade die Frage. [Anlegen]
// friert Revision und Schema **einer** Karte ein und hätte keine.
//
// **Was sie trotzdem braucht, sind dieselben zwei Regeln**: die Fähigkeit
// muss belegt sein, und das Modell muss hier rechnen dürfen. Eine zweite
// Fassung davon neben dem Werk liefe beim ersten neuen Ort auseinander.
func TestPruefenLegtKeinenAuftragAn(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}

	wunsch := einWunsch()

	// **Ohne Zielkarte** — das ist der Fall, für den es die Methode gibt.
	wunsch.Zielkarte = 0

	modell, fehler := einWerk(t, karten).Pruefen(wunsch)
	if fehler != nil {
		t.Fatalf("Pruefen: %v", fehler)
	}

	if modell.Alias != "schreibt:8b" {
		t.Errorf("geprüft wurde %q", modell.Alias)
	}

	// **Und die Karte wurde nicht einmal gelesen** — `Anlegen` läse sie, um
	// ihre Revision einzufrieren.
	if karten.schreibvorgaenge != 0 {
		t.Errorf("es wurde %d mal geschrieben", karten.schreibvorgaenge)
	}
}

// **Und sie weist dasselbe ab wie das Anlegen.**
//
// **Das ist die Gegenprobe zur Zeile darüber**: ohne sie sagte der Lauf nur,
// dass `Pruefen` durchlässt — nicht, dass es überhaupt prüft.
func TestPruefenWeistDasselbeAb(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(service.KIWunsch) service.KIWunsch{
		"eine unbelegte Fähigkeit": func(w service.KIWunsch) service.KIWunsch {
			w.Faehigkeit = domain.KIVision

			return w
		},
		"ein Modell, das es nicht gibt": func(w service.KIWunsch) service.KIWunsch {
			w.Alias = "gibtesnicht"

			return w
		},
	}

	for name, aendern := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wunsch := aendern(einWunsch())
			wunsch.Zielkarte = 0

			if _, fehler := einWerk(t, &kiKartenprobe{}).Pruefen(wunsch); fehler == nil {
				t.Error("es wurde durchgelassen")
			}
		})
	}
}
