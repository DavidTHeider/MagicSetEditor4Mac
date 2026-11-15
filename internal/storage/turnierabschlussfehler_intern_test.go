package storage

// **Die Fehlerwege von Abschluss und Protokoll** — DOM-04.AK-04
// (#1699).
//
// **Ein interner Lauf**, weil er die Stolpertreiber braucht: gegen eine echte
// SQLite-Datei beginnt jeder Zug, jedes Zählen gelingt, und jede Zeile des
// Protokolls lässt sich schreiben.

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestAbschlussUndProtokollReichenIhreFehlerDurch(t *testing.T) {
	abschliessen := func(a *Turnierablage) error { return a.Abschliessen(probekontext(t), 1, 1) }
	protokoll := func(a *Turnierablage) error {
		_, fehler := a.Protokoll(probekontext(t), 1)

		return fehler
	}
	anmelden := func(a *Turnierablage) error {
		_, fehler := a.Anmelden(probekontext(t), 1, domain.Teilnehmer{Name: "Ada"})

		return fehler
	}

	// einlesen bringt eine Runde mit einem Ergebnis — DOM-04.AK-03.
	einlesen := func(a *Turnierablage) error {
		_, fehler := a.Einlesen(probekontext(t), domain.Turnier{
			Name: "T", Spiel: "magic",
			Teilnehmer: []domain.Teilnehmer{{ID: 1, Name: "Ada"}, {ID: 2, Name: "Bo"}},
			Runden: []domain.Runde{{Nummer: 1, Paarungen: []domain.Paarung{
				{A: 1, B: 2, Ergebnis: &domain.Ergebnis{SpieleA: 2}},
			}}},
		})

		return fehler
	}

	faelle := []struct {
		name    string
		treiber string
		ruf     func(*Turnierablage) error
	}{
		{"der Abschluss beginnt nicht", "stolper-turnieranlage-beginn", abschliessen},
		{"der Abschluss schreibt nicht", "stolper-schreiben", abschliessen},
		{"der Abschluss zählt nicht", "stolper-abschluss-zaehlen", abschliessen},
		{"der Abschluss protokolliert nicht", "stolper-protokoll-schreiben", abschliessen},
		{"der Abschluss schließt nicht ab", "stolper-commit", abschliessen},
		{"das Protokoll lässt sich nicht fragen", "stolper-protokoll-lesen", protokoll},
		{"eine Zeile des Protokolls ist unlesbar", "stolper-protokoll-scan", protokoll},
		{"das Anmelden zählt nicht", "stolper-fassung-zaehlen", anmelden},
		{"das Anmelden protokolliert nicht", "stolper-protokoll-schreiben", anmelden},
		{"das Anlegen protokolliert nicht", "stolper-protokoll-schreiben", func(a *Turnierablage) error {
			_, fehler := a.Anlegen(probekontext(t), domain.Turnier{Name: "T", Spiel: "magic"})

			return fehler
		}},
		{"ein Ergebnis lässt sich nicht einlesen", "stolper-paarung-schreiben", einlesen},
		{"das Ergebnis lässt sich nicht schreiben", "stolper-paarung-schreiben", func(a *Turnierablage) error {
			return a.ErgebnisSetzen(probekontext(t), 1, 1, &domain.Ergebnis{SpieleA: 2})
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			pruefeStolperfehler(t, fall.ruf(NeueTurnierablage(db)))
		})
	}
}

// **Ohne Profil wird nichts abgeschlossen** — jede Wurzelabfrage klemmt auf
// eines (#1351).
func TestOhneProfilSchliesstNichtsAb(t *testing.T) {
	db, err := sql.Open("stolper-schreiben", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if fehler := NeueTurnierablage(db).Abschliessen(context.Background(), 1, 1); fehler == nil {
		t.Error("ohne Profil abgeschlossen")
	}
}
