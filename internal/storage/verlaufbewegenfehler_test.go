package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Jeder Schritt der Bewegung kann scheitern, und jeder wird gemeldet** —
// Werkstück U6 von Phase 5.1 (#1417). Ein Trigger lässt genau die eine
// Anweisung abbrechen, um die es geht: das Entfernen eines Wertes, sein
// Zurückschreiben, das Fortschreiben des Zeigers, der Abschluss.
func TestJederSchrittDerBewegungMeldetSeinenFehler(t *testing.T) {
	faelle := []struct {
		name    string
		feld    string
		vorwert string
		stelle  bool
		trigger string
	}{
		{
			// **Ein Feld, das die Karte vorher nicht hatte** — sein Rückweg
			// ist ein DELETE, und genau der ist gesperrt.
			name: "das Entfernen eines Wertes scheitert",
			feld: "typ",
			trigger: `CREATE TRIGGER sperre BEFORE DELETE ON card_values
			          BEGIN SELECT RAISE(ABORT, 'gesperrt'); END`,
		},
		{
			name:    "das Zurückschreiben eines Wertes scheitert",
			feld:    "name",
			vorwert: "Eisball",
			trigger: `CREATE TRIGGER sperre BEFORE UPDATE ON card_values
			          BEGIN SELECT RAISE(ABORT, 'gesperrt'); END`,
		},
		{
			// **Ein Wert mit Stelle** — der Rückweg setzt ihn an seine Stelle,
			// und genau dieses Schreiben ist gesperrt.
			name:    "das Zurückschreiben eines Wertes an seine Stelle scheitert",
			feld:    "name",
			vorwert: "Eisball",
			stelle:  true,
			trigger: `CREATE TRIGGER sperre BEFORE UPDATE ON card_values
			          BEGIN SELECT RAISE(ABORT, 'gesperrt'); END`,
		},
		{
			name: "das Fortschreiben des Zeigers scheitert",
			feld: "name",
			trigger: `CREATE TRIGGER sperre BEFORE UPDATE OF zurueckgenommen ON history
			          BEGIN SELECT RAISE(ABORT, 'gesperrt'); END`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, setID, karteID := bestandMitKarte(t)
			schreiber := storage.NeuerKartenschreiber(db)

			if fall.vorwert != "" {
				if _, _, err := schreiber.FeldSetzenMitVorwert(probekontext(t), karteID, fall.feld, fall.vorwert); err != nil {
					t.Fatalf("Vorwert setzen: %v", err)
				}
			}

			if fall.stelle {
				if _, err := db.ExecContext(probekontext(t),
					`UPDATE card_values SET reihenfolge = 0 WHERE card_id = ? AND feld = ?`,
					karteID, fall.feld); err != nil {
					t.Fatalf("Stelle setzen: %v", err)
				}
			}

			if _, _, err := schreiber.FeldSetzenMitVorwert(probekontext(t), karteID, fall.feld, "Steinwand"); err != nil {
				t.Fatalf("setzen: %v", err)
			}

			if _, err := db.ExecContext(probekontext(t), fall.trigger); err != nil {
				t.Fatalf("Trigger: %v", err)
			}

			if _, _, err := storage.NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), setID); err == nil {
				t.Error("die Bewegung ging durch, obwohl ein Schritt gesperrt war")
			}
		})
	}
}
