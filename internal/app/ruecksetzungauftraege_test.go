package app

// **Die Auftragsschau sieht, was dieses Fenster tut** — Werkstück O4 von
// Phase 5.2 (#1428).
//
// **Ein interner Test**, weil die Schau die Naht zwischen zwei Feldern
// derselben Anwendung ist: sie hat keine Bindung, und von außen wäre nur zu
// sehen, dass eine Löschung sperrt — nicht, dass sie es aus diesem Grund tut.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieAuftragsschauSiehtDasBuchDiesesFensters(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("/nirgends/bestand.db")

	if laufend := anwendung.laufendeAuftraege(); len(laufend) != 0 {
		t.Fatalf("ein frisches Fenster hat Aufträge: %+v", laufend)
	}

	fertig := anwendung.auftragBeginnen("ausfuhr", "die Ausfuhr")

	laufend := anwendung.laufendeAuftraege()
	if len(laufend) != 1 || laufend[0].Stand != domain.AuftragLaeuft {
		t.Fatalf("das Buch sagt %+v", laufend)
	}

	// **Und nach dem Ende ist er nicht mehr laufend** — sonst sperrte ein
	// beendeter Auftrag jede Löschung für immer.
	fertig(nil)

	danach := anwendung.laufendeAuftraege()
	if len(danach) != 1 || danach[0].Stand == domain.AuftragLaeuft {
		t.Errorf("nach dem Ende sagt das Buch %+v", danach)
	}
}
