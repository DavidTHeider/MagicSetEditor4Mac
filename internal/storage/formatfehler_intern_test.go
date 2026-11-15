package storage

// **Die Fehlerwege des Formatspeichers** — Werkstück 4 von Phase 4.6
// (#1293).
//
// **Ein interner Lauf**, weil er den Stolpertreiber braucht: `RowsAffected`
// scheitert beim SQLite-Treiber nie, und im Ernstfall entscheidet es, ob die
// Anwendung „das Format stand nicht da" von „das Zählen ging schief"
// unterscheidet.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Zählfehler an Änderung und Löschung.**
func TestDieFormatablageMeldetIhreZaehlfehler(t *testing.T) {
	for _, fall := range []struct {
		name string
		ruf  func(*Formatablage) error
	}{
		{"ändern", func(a *Formatablage) error {
			_, err := a.Aendern(probekontext(t), domain.Format{
				ID: 1, Fassung: 1, Name: "F", Spiel: "magic",
			})

			return err
		}},
		{"löschen", func(a *Formatablage) error {
			return a.Loeschen(probekontext(t), 1)
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open("stolper-zaehlen", "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			if fehler := fall.ruf(NeueFormatablage(db)); fehler == nil {
				t.Error("der Zählfehler ging verloren")
			}
		})
	}
}

// **Und die Nachfrage danach meldet einen geschlossenen Bestand.**
func TestDieFormatnachfrageMeldetEinenGeschlossenenBestand(t *testing.T) {
	if err := NeueFormatablage(geschlosseneDatenbank(t)).warumNicht(
		probekontext(t), 1, 1); err == nil {
		t.Error("die geschlossene Datenbank nahm die Nachfrage an")
	}
}

// **Und die zwei Wege, die nur ein Treiber herbeiführt** — eine Zeile, die
// sich nicht umwandeln lässt, und ein Zeilenende mit nachgereichtem Fehler.
func TestDieFormatablageMeldetIhreTreiberfehler(t *testing.T) {
	for _, treiber := range []string{"stolper-umwandeln", "stolper-lesen"} {
		t.Run(treiber, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			ablage := NeueFormatablage(db)

			if _, fehler := ablage.Formate(probekontext(t), "magic"); fehler == nil {
				t.Error("die Formatliste kam durch")
			}

			if _, fehler := ablage.Format(probekontext(t), 1); fehler == nil {
				t.Error("das einzelne Format kam durch")
			}
		})
	}
}
