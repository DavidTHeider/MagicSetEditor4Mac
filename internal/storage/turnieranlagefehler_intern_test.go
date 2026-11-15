package storage

// **Die Fehlerwege des Anlegens samt Teilnehmenden** — DOM-04.AK-01
// (#1699).
//
// **Ein interner Lauf**, weil er die Stolpertreiber braucht: gegen eine echte
// SQLite-Datei beginnt jeder Zug, jede Kennung lässt sich lesen und jeder
// Abschluss gelingt.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDasAnlegenSamtTeilnehmendenReichtSeineFehlerDurch(t *testing.T) {
	mitNamen := domain.Turnier{
		Name: "T", Spiel: "magic", Teilnehmer: []domain.Teilnehmer{{Name: "Ada"}},
	}

	faelle := []struct {
		name    string
		treiber string
	}{
		{"der Zug beginnt nicht", "stolper-turnieranlage-beginn"},
		{"die Kennung des Turniers ist unlesbar", "stolper-turnier-kennung"},
		{"ein Name lässt sich nicht anmelden", "stolper-teilnehmer-kennung"},
		{"der Abschluss scheitert", "stolper-commit"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			_, fehler := NeueTurnierablage(db).Anlegen(probekontext(t), mitNamen)
			pruefeStolperfehler(t, fehler)
		})
	}
}
