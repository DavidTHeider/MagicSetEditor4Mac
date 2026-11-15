package storage

// **Die zwei Wege, die eine echte Datenbank nicht hergibt** — Befund 9 des
// Nachlaufs zum Zweitblick über Phase 5.2 (#1435).
//
// Die Umfangsbindung liest **in** der Löschtransaktion, welche Profile
// betroffen sind. Eine Zeile, die sich nicht lesen lässt, und ein Abbruch
// **nach** der Schleife kommen im Betrieb vor — eine abgebrochene Verbindung,
// eine volle Platte — und niemand sieht sie, weil sie selten sind.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieUmfangsbindungMeldetIhreLesefehler(t *testing.T) {
	nichts := func([]domain.Profil, [][]domain.Betroffenes) error { return nil }

	for name, fall := range map[string]struct {
		treiber string
		fahre   func(*Nutzerablage) error
	}{
		"die Profilliste lässt sich nicht lesen": {
			"stolper-umfangsprofil-scan", func(a *Nutzerablage) error {
				return a.ProfilLoeschen(probekontext(t), 1, "Mein Profil", 0, nichts)
			},
		},
		"die Profilliste bricht ab": {
			"stolper-umfangsprofil-ende", func(a *Nutzerablage) error {
				return a.ProfilLoeschen(probekontext(t), 1, "Mein Profil", 0, nichts)
			},
		},
		// **Beim Menschen und nicht nur beim Profil**: es sind zwei Aufrufer
		// derselben Bindung, und ein Lauf über einen ließe den anderen
		// unbetreten (§3).
		"die Profilliste des Menschen lässt sich nicht lesen": {
			"stolper-umfangsprofil-scan", func(a *Nutzerablage) error {
				return a.NutzerLoeschen(probekontext(t), 1, "Ich", nichts)
			},
		},

		// **Die fremden Zugänge des Menschen** (#2287) — gezählt im selben
		// Geschäft, vor der Bindung.
		"die fremden Zugänge lassen sich nicht zählen": {
			"stolper-fremdzugang", func(a *Nutzerablage) error {
				return a.NutzerLoeschen(probekontext(t), 1, "Ich", nichts)
			},
		},

		// **Die Abfrage selbst** — `stolper-fragen` träfe die Namensprüfung
		// davor, und der Lauf endete, bevor die Bindung liefe.
		"die Profilliste lässt sich gar nicht erst abfragen": {
			"stolper-umfangsprofil-abfrage", func(a *Nutzerablage) error {
				return a.ProfilLoeschen(probekontext(t), 1, "Mein Profil", 0, nichts)
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			db, fehler := sql.Open(fall.treiber, ":memory:")
			if fehler != nil {
				t.Fatalf("öffnen: %v", fehler)
			}

			t.Cleanup(func() { _ = db.Close() })

			pruefeStolperfehler(t, fall.fahre(NeueNutzerablage(db)))
		})
	}
}
