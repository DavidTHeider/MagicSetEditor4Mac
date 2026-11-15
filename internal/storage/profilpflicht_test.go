package storage_test

import (
	"strings"
	"testing"
)

// TestEineVergesseneProfilzuordnungWirdLautAbgewiesen hält die Kehrseite von
// `0035_profiltrennung.sql` fest.
//
// # Der Vorgabewert ist eine Falle, und sie muss zuschnappen
//
// SQLite kann eine `NOT NULL`-Spalte nur mit Vorgabewert nachrüsten
// (`ALTER TABLE … ADD COLUMN`), und `sets`, `decks`, `turniere` und `formate`
// haben sie deshalb mit `DEFAULT 0` bekommen. Eine Zeile ohne Zuordnung landet
// damit bei Profil 0 — **das es nicht gibt**, denn `profile.id` ist
// `AUTOINCREMENT` und beginnt bei 1.
//
// Ohne Wächter wäre das der leiseste aller Fehler: das Schreiben glückt, die
// Zeile steht in der Datei, und **niemand** sieht sie je wieder, weil jede
// Leseabfrage auf ein echtes Profil klemmt. Ein Set, das nach dem Anlegen
// verschwindet, sieht aus wie ein Lesefehler und ist ein Schreibfehler.
//
// **Die sieben neu gebauten Tabellen haben das Problem nicht** — sie sind neu
// gebaut, ihre `profil_id` trägt keinen Vorgabewert, und ein `INSERT` ohne sie
// scheitert an `NOT NULL`. Geprüft wird deshalb genau die andere Hälfte.
func TestEineVergesseneProfilzuordnungWirdLautAbgewiesen(t *testing.T) {
	faelle := []struct {
		tabelle   string
		einfuegen string
	}{
		{"sets", `INSERT INTO sets (name, spiel, geoeffnet)
		          VALUES ('probe', 'magic', '2026-01-01T00:00:00.000Z')`},
		{"decks", `INSERT INTO decks (name, spiel) VALUES ('probe', 'magic')`},
		{"turniere", `INSERT INTO turniere (name, spiel) VALUES ('probe', 'magic')`},
		{"formate", `INSERT INTO formate (name, spiel) VALUES ('probe', 'magic')`},
	}

	for _, fall := range faelle {
		t.Run(fall.tabelle, func(t *testing.T) {
			db := frischeDatenbank(t)

			_, err := db.ExecContext(t.Context(), fall.einfuegen)
			if err == nil {
				t.Fatalf("%s nahm eine Zeile ohne profil_id an", fall.tabelle)
			}

			if !strings.Contains(err.Error(), "profil_id") {
				t.Errorf("die Meldung nennt die Spalte nicht: %v", err)
			}
		})
	}
}
