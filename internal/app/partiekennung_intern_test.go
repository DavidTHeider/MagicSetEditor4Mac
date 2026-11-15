package app

// **Die Kartenkennung an der Naht** — Werkstück G2 von Phase 5.5
// (#1459).
//
// **Eine eigene Datei neben `partie_intern_test.go`**, weil jene über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht, **was**
// die Naht tut, hier, **wie** eine Kennung dabei aussieht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Kartenkennung wird zur Zeichenkette — und die Null zu nichts.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil `internal/app` die DTO-Abbildung prüft** (§3), und diese Abbildung
// trägt eine Entscheidung: die Null wird **leer** und nicht `"0"`. Das
// Fenster soll „gibt es ein Bild?" mit einem Blick beantworten können statt
// mit einem Vergleich gegen eine Zeichenkette, die wie eine Kennung aussieht.
//
// **Die Wege über den Bestand sehen das nicht**: dort liegt alles in der
// Bibliothek, und eine verdeckte Karte trägt gar keine Kennung (G2,
// #1459).
func TestDieKartenkennungWirdZurZeichenkette(t *testing.T) {
	t.Parallel()

	dto := partieAus(domain.Partiesicht{
		PartieID: "p1", Spiel: "magic", Fuer: "sitz-1",
		Objekte: []domain.Sichtobjekt{
			{Kennung: "offen", KarteID: 77, Zone: domain.ZoneHand},
			{Kennung: "verdeckt", Zone: domain.ZoneBibliothek},
		},
	})

	if len(dto.Objekte) != 2 {
		t.Fatalf("%d Objekte", len(dto.Objekte))
	}

	if dto.Objekte[0].KarteID != "77" {
		t.Errorf("die offene Karte trägt %q", dto.Objekte[0].KarteID)
	}

	if dto.Objekte[1].KarteID != "" {
		t.Errorf("die verdeckte trägt %q — erwartet nichts",
			dto.Objekte[1].KarteID)
	}
}
