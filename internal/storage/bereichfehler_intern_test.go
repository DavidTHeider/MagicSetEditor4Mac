package storage

// **Die drei Fehlerwege des Bereichswechsels hinter der ersten Anweisung** —
// Befund 7 des Zweitblicks über Phase 5.1
// (#1405).
//
// **Ein interner Test**, weil er den Stolpertreiber aus demselben Paket
// braucht — wie `deckfehler_intern_test.go`.

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jede Anweisung des Zuges meldet, wenn sie scheitert — und zwar ihren.**
//
// Eine echte Datenbank gibt nur die **erste** her: ein `DROP TABLE` trifft
// das Lesen des Platzes, und was danach käme, wird nie erreicht.
//
// # Warum der Fall im Fehlertext steht und nicht nur „irgendein Fehler"
//
// **Weil der erste Anlauf grün war, ohne einen dieser Wege zu betreten** —
// Befund 1 des Nachlaufs zum Zweitblick über Phase 5.1. Er verlangte nur
// `fehler != nil`; der Stolpertreiber gibt für jedes ungestolperte Schreiben
// `driver.RowsAffected(0)` zurück, und das kennt kein `LastInsertId`. Alle
// drei Läufe scheiterten deshalb an derselben Stelle — „die Kennung des
// Deckplatzes lesen" — und keiner an der, um die es ihm ging. Wer die drei
// Stolperstellen entfernte, sah trotzdem Grün.
//
// **Der Text ist die Zusage.** Er steht in `deckbereich.go` bei
// `uebersetze(…)`, und wer ihn dort ändert, ändert ihn hier.
func TestJedeAnweisungDesBereichswechselsMeldetIhrenFehler(t *testing.T) {
	for _, fall := range []struct {
		treiber  string
		erwartet string
	}{
		{"stolper-bereich-menge", "die Menge im Zielbereich lesen"},
		{"stolper-bereich-loeschen", "den alten Deckplatz löschen"},
		{"stolper-bereich-setzen", "den Deckplatz legen"},
	} {
		t.Run(fall.treiber, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			_, fehler := NeueDeckablage(db).PlatzVerschieben(
				probekontext(t), 1, 1, domain.BereichSideboard)

			switch {
			case fehler == nil:
				t.Fatal("der Fehler ging verloren")
			case !strings.Contains(fehler.Error(), fall.erwartet):
				t.Errorf("Fehler %q nennt nicht %q — der Lauf ist an einer "+
					"anderen Stelle gescheitert als der, die er prüft",
					fehler.Error(), fall.erwartet)
			}
		})
	}
}
