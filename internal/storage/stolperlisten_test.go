// **Die Fehlerwege der Listenablage und der Ersetzung** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// **Eine dritte Datei neben `stolperabfragen_test.go`**, weil jene mit ihr
// über 400 Zeilen ginge (§2). Der Schnitt folgt einer Frage: dort stehen die
// Wege, die es seit Phase 2 gibt, hier die aus Phase 3.7.
package storage

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Fehlerwege der Listenablage** (#1168).
//
// Dieselbe Zusage wie überall: ein Bestand, der nicht antwortet, meldet sich
// — sonst stünde die Kartenliste ohne Spalten da und niemand wüsste, warum.
func TestDieListenablageMeldetIhreFehlerpfade(t *testing.T) {
	faelle := []struct {
		name    string
		treiber string
		lauf    func(*Listenablage) error
	}{
		{
			"Spalten lesen: die Abfrage scheitert", "stolper-fragen",
			func(a *Listenablage) error { _, err := a.Spalten(probekontext(t), "magic"); return err },
		},
		{
			"Spalten lesen: eine Zeile ist unlesbar", "stolper-umwandeln",
			func(a *Listenablage) error { _, err := a.Spalten(probekontext(t), "magic"); return err },
		},
		{
			"Spalten lesen: ein Lesefehler danach", "stolper-lesen",
			func(a *Listenablage) error { _, err := a.Spalten(probekontext(t), "magic"); return err },
		},
		{
			"Spalten schreiben: das Räumen scheitert", "stolper-schreiben",
			func(a *Listenablage) error {
				return a.SpaltenSetzen(probekontext(t), "magic",
					[]domain.Spaltenwahl{{Feld: "name", Sichtbar: true}})
			},
		},
		{
			"Spalten schreiben: die Spalte selbst scheitert", "stolper-zweites-schreiben",
			func(a *Listenablage) error {
				return a.SpaltenSetzen(probekontext(t), "magic",
					[]domain.Spaltenwahl{{Feld: "name", Sichtbar: true}})
			},
		},
		{
			"Spalten schreiben: das Commit scheitert", "stolper-commit",
			func(a *Listenablage) error {
				return a.SpaltenSetzen(probekontext(t), "magic",
					[]domain.Spaltenwahl{{Feld: "name", Sichtbar: true}})
			},
		},
		{
			"die Sortierung lesen", "stolper-fragen",
			func(a *Listenablage) error { _, err := a.Sortierung(probekontext(t), "magic"); return err },
		},
		{
			"die Sortierung schreiben", "stolper-schreiben",
			func(a *Listenablage) error {
				return a.SortierungSetzen(probekontext(t), "magic", domain.Kartenordnung{})
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			if fehler := fall.lauf(NeueListenablage(db)); fehler == nil {
				t.Error("der Lauf kam ohne Fehler zurück")
			}
		})
	}
}

// **Und auf einer geschlossenen Datenbank beginnt keine Transaktion.**
func TestDieSpaltenwahlAufEinerGeschlossenenDatenbank(t *testing.T) {
	fehler := NeueListenablage(geschlosseneDatenbank(t)).SpaltenSetzen(
		probekontext(t), "magic", nil)
	if fehler == nil {
		t.Error("die Spaltenwahl begann auf einer geschlossenen Datenbank")
	}
}

// **Und auf einer geschlossenen Datenbank beginnt keine Transaktion.**
func TestWerteSetzenAufEinerGeschlossenenDatenbank(t *testing.T) {
	fehler := NeuerKartenschreiber(geschlosseneDatenbank(t)).WerteSetzen(probekontext(t),
		[]domain.Feldwechsel{{KarteID: 1, Feld: "typ", Wert: "Artefakt"}})
	if fehler == nil {
		t.Error("die Ersetzung begann auf einer geschlossenen Datenbank")
	}
}
