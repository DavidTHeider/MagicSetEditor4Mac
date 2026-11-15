package storage

// **Die Zählfehler der Deckablage** — Werkstück 2 von Phase 4.6
// (#1293).
//
// **Ein interner Lauf**, weil er den Stolpertreiber braucht: `RowsAffected`
// scheitert beim SQLite-Treiber nie, und im Ernstfall entscheidet es, ob die
// Anwendung „das Deck stand nicht da" von „das Zählen ging schief"
// unterscheidet.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieDeckablageMeldetIhreZaehlfehler(t *testing.T) {
	for _, fall := range []struct {
		name string
		ruf  func(*Deckablage) error
	}{
		{"ein Deck löschen", func(a *Deckablage) error {
			return a.Loeschen(probekontext(t), 1)
		}},
		{"einen Platz löschen", func(a *Deckablage) error {
			return a.PlatzLoeschen(probekontext(t), 1, 1)
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open("stolper-zaehlen", "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			if fehler := fall.ruf(NeueDeckablage(db)); fehler == nil {
				t.Error("der Zählfehler ging verloren")
			}
		})
	}
}

// **Und die zwei Wege, die nur ein Treiber herbeiführt.**
//
// Eine Zeile, die sich nicht umwandeln lässt, und ein Zeilenende, das einen
// Fehler nachreicht: keines kann der SQLite-Treiber, und jedes entscheidet im
// Ernstfall, ob die Anwendung es meldet.
func TestDieDeckablageMeldetIhreTreiberfehler(t *testing.T) {
	for _, treiber := range []string{"stolper-umwandeln", "stolper-lesen"} {
		t.Run(treiber, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			ablage := NeueDeckablage(db)

			if _, fehler := ablage.Decks(probekontext(t), "magic"); fehler == nil {
				t.Error("die Deckliste kam durch")
			}

			if _, fehler := ablage.Deck(probekontext(t), 1); fehler == nil {
				t.Error("das einzelne Deck kam durch")
			}

			// **Und die Frage nach den Decks einer Karte** (#1369) — sie
			// steht vor einem Löschen, und eine halbe Liste dort hieße
			// „liegt in keinem weiteren Deck", wo niemand nachgesehen hat.
			if _, fehler := ablage.DecksMitKarte(probekontext(t), 1); fehler == nil {
				t.Error("die Decks zur Karte kamen durch")
			}

			// **Und das Suchen eines vorhandenen Platzes** — der Weg, den
			// `PlatzSetzen` vor dem Fortschreiben geht. Ein Fehler dort ist
			// weder „steht nicht da" noch „geschrieben".
			if _, fehler := ablage.PlatzSetzen(probekontext(t), 1, domain.Deckplatz{
				Printing: "a1b2c3d4-0000-0000-0000-000000000000",
				Menge:    1, Bereich: domain.BereichHauptdeck,
			}); fehler == nil {
				t.Error("das Suchen kam durch")
			}

			// **Und derselbe Weg über eine eigene Karte** — die andere Spalte.
			if _, fehler := ablage.PlatzSetzen(probekontext(t), 1, domain.Deckplatz{
				KarteID: 5, Menge: 1, Bereich: domain.BereichHauptdeck,
			}); fehler == nil {
				t.Error("das Suchen über die Karte kam durch")
			}
		})
	}
}

// **Die Änderung meldet ihren Zählfehler** — Werkstück 3.
//
// `Aendern` fragt bei null getroffenen Zeilen nach, **warum**. Zählt der
// Treiber nicht, ist das weder „gibt es nicht" noch „überholt", und beides zu
// melden wäre falsch.
func TestDieDeckaenderungMeldetIhrenZaehlfehler(t *testing.T) {
	db, err := sql.Open("stolper-zaehlen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	zweiterSchreibzaehler = 0

	if _, fehler := NeueDeckablage(db).Aendern(probekontext(t), domain.Deck{
		ID: 1, Fassung: 1, Name: "Deck", Spiel: "magic",
	}); fehler == nil {
		t.Error("der Zählfehler ging verloren")
	}
}

// **Und die Nachfrage danach meldet einen geschlossenen Bestand.**
//
// Die Unterscheidung zwischen „gibt es nicht" und „jemand anderes hat
// geschrieben" hängt an einer **zweiten** Abfrage, und sie hat ihren eigenen
// Fehlerweg — dieselbe Form wie in `sammlung.go`.
func TestDieDecknachfrageMeldetEinenGeschlossenenBestand(t *testing.T) {
	if err := NeueDeckablage(geschlosseneDatenbank(t)).warumNicht(
		probekontext(t), 1, 1); err == nil {
		t.Error("die geschlossene Datenbank nahm die Nachfrage an")
	}
}
