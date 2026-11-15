// **Die Identität eines Auftrags** — Befund 8 des Zweitblicks über Phase 5.2
// (#1435).
//
// **Eine eigene Datei**, weil `kostenjournal_test.go` an der Zeilengrenze
// steht (§2). Die Frage hier ist eine andere als dort: nicht „bucht das
// Journal richtig", sondern „**meint** dieselbe Kennung denselben Auftrag".
package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Dieselbe Kennung für einen anderen Auftrag ist ein Konflikt.**
//
// Befund 8 des Zweitblicks über Phase 5.2: der vorhandene Posten kam zurück,
// ohne dass irgendetwas verglichen wurde. Eine wiederverwendete Kennung ließ
// damit eine **fremde** Reservierung wie die Freigabe des neuen Rufs aussehen
// — und der Höchstbetrag, den jemand gesehen hatte, war nicht der, der galt.
func TestDieselbeKennungFuerEinenAnderenAuftragIstEinKonflikt(t *testing.T) {
	journal := journalMitBudget(t, 5000)

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 300), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("der erste: %v", fehler)
	}

	for name, anderer := range map[string]domain.Kostenposten{
		"anderer Höchstbetrag": auftrag("eins", 900),
		"anderes Konto": {
			Kennung: "eins", ProfilID: 1, Konto: "woanders",
			Aufgabe: "karte.text", Hoechstbetrag: euro(300),
		},
		"andere Aufgabe": {
			Kennung: "eins", ProfilID: 1, Konto: "anbieter",
			Aufgabe: "bild.entwurf", Hoechstbetrag: euro(300),
		},
		// **Ein anderes Profil fehlt hier mit Absicht**: dieselbe Kennung in
		// einem anderen Profil ist seit Wanderung 0082 ein **anderer** Auftrag
		// und kein Konflikt (Nachlauf über Phase 6.1, Befund 4) —
		// `TestDieselbeKennungInZweiProfilenSindZweiAuftraege`.
	} {
		t.Run(name, func(t *testing.T) {
			_, fehler := journal.Reservieren(t.Context(), anderer,
				domain.ZeitraumTag, probezeit)

			if !errors.Is(fehler, domain.ErrConflict) {
				t.Errorf("abgewiesen wurde mit %v", fehler)
			}
		})
	}

	// **Und der ursprüngliche Posten steht unverändert da.**
	stand, fehler := journal.Stand(t.Context(), 1, "anbieter", domain.ZeitraumTag, probezeit)
	if fehler != nil {
		t.Fatalf("der Stand: %v", fehler)
	}

	if stand.Reserviert.Wert != 300 {
		t.Errorf("reserviert sind %d statt 300", stand.Reserviert.Wert)
	}
}
