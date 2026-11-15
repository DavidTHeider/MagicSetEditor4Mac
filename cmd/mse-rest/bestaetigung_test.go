// **Die Bestätigung beim Entfernen eines Pakets** —
// #1291,
// neu geschnitten in
// #1344.
//
// # Was hier steht und was im Dienst
//
// Hier stand der ganze Vergleich: die Wurzel las den Katalog, hielt den Pfad
// dagegen und rief danach `Entfernen`, das **erneut** las. Bestätigung und
// Wirkung bezogen sich nicht auf denselben Stand.
//
// Der Vergleich steht jetzt in `service.Paketverwalter.Entfernen`, hinter
// derselben einen Lesung, die auch löscht. **Hier bleibt die Pflicht dieser
// Fläche**: `DELETE /pakete/{art}/{name}` verlangt eine Bestätigung
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)) —
// eine Aussage über den **Aufruf**, nicht über den Katalog, und deshalb ohne
// jede Lesung.
package main

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne Pfad wird nichts entfernt — und der Katalog wird nicht angefasst.**
//
// Die zweite Hälfte ist der Kern: die Wache greift, **bevor** irgendetwas
// gelesen wird. Der Datenbankgriff ist `nil`; käme sie überhaupt bis zur
// Naht, wäre der Lauf kein Konflikt, sondern ein Absturz.
func TestOhnePfadWirdNichtsEntfernt(t *testing.T) {
	t.Parallel()

	var ohneBestand *sql.DB

	fehler := pakete{db: ohneBestand}.PaketEntfernen(
		probekontext(t), "mse-style", "magic-m15", "")
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("Fehler %v, erwartet war ErrConflict", fehler)
	}

	// **Die Meldung sagt, woher der Pfad kommt.** Eine Abweisung ohne Ausweg
	// ist eine Sackgasse, und ein Modell versucht es dann noch einmal genauso.
	if !strings.Contains(fehler.Error(), "»GET /pakete«") {
		t.Errorf("die Meldung nennt die Quelle nicht: %v", fehler)
	}
}
