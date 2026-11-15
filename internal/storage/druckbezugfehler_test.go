package storage_test

// **Jede Naht meldet einen Druck, den sie nicht anlegen kann** — Werkstück A2
// von Phase 5.3 (#1441).
//
// `druckZeile` hat drei Wege, auf denen SQLite selbst scheitert: die Kennung
// ist nicht nachzuschlagen, der Druck nicht anzulegen, die Kennung nicht
// anzuhängen. Jede der fünf Nähte führt dorthin — und jede muss den Fehler
// **durchreichen**, statt ihre Zeile ohne Druck zu schreiben.
//
// **Kein `t.Parallel()`**, aus dem Grund, der in `sammlung_test.go` steht.

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// druckstoerungen sind die drei Eingriffe, nach denen kein Druck mehr entsteht.
var druckstoerungen = []struct{ name, satz string }{
	{"die Kennungen sind nicht nachzuschlagen", `DROP TABLE druckkennungen`},
	{
		"der Druck ist nicht anzulegen",
		`CREATE TRIGGER sperre BEFORE INSERT ON drucke BEGIN SELECT RAISE(ABORT, 'gesperrt'); END`,
	},
	{
		"die Kennung ist nicht anzuhängen",
		`CREATE TRIGGER sperre BEFORE INSERT ON druckkennungen BEGIN SELECT RAISE(ABORT, 'gesperrt'); END`,
	},
}

// druckNaht ist eine Naht mit dem, was sie vor der Störung braucht.
type druckNaht struct {
	name string
	// vorbereiten legt an, was der Lauf braucht — **vor** der Störung.
	vorbereiten func(t *testing.T, db *sql.DB) any
	lauf        func(t *testing.T, db *sql.DB, vorbereitet any) error
}

// vorbereitetes holt, was `vorbereiten` abgelegt hat — in der Gestalt, die der
// Lauf erwartet.
func vorbereitetes[T any](t *testing.T, wert any) T {
	t.Helper()

	gestalt, passt := wert.(T)
	if !passt {
		t.Fatalf("vorbereitet ist %T", wert)
	}

	return gestalt
}

// neuerAlias ist ein Druck, den noch keine Zeile kennt — so führt jeder Lauf
// bis zum Anlegen.
const neuerAlias = "feedface-0000-0000-0000-000000000001"

func druckNaehte() []druckNaht {
	return []druckNaht{
		{
			name:        "Sammlung anlegen",
			vorbereiten: func(*testing.T, *sql.DB) any { return nil },
			lauf: func(t *testing.T, db *sql.DB, _ any) error {
				exemplar := probeexemplar()
				exemplar.Printing = neuerAlias

				_, err := storage.NeuerSammlungsspeicher(db).Anlegen(probekontext(t), exemplar)

				return err
			},
		},
		{
			name: "Sammlung ändern",
			vorbereiten: func(t *testing.T, db *sql.DB) any {
				speicher := storage.NeuerSammlungsspeicher(db)
				if _, err := speicher.Anlegen(probekontext(t), probeexemplar()); err != nil {
					t.Fatalf("Anlegen: %v", err)
				}

				liste, err := speicher.Liste(probekontext(t))
				if err != nil {
					t.Fatalf("Liste: %v", err)
				}

				return liste[0]
			},
			lauf: func(t *testing.T, db *sql.DB, vorbereitet any) error {
				exemplar := vorbereitetes[*domain.Exemplar](t, vorbereitet)
				exemplar.Printing = neuerAlias

				return storage.NeuerSammlungsspeicher(db).Aendern(probekontext(t), exemplar)
			},
		},
		{
			name: "Verweis setzen",
			vorbereiten: func(t *testing.T, db *sql.DB) any {
				setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
					Name: "Probeset", Game: "magic",
					Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Erste"}}}},
				}, nil)
				if err != nil {
					t.Fatalf("Set anlegen: %v", err)
				}

				karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
				if err != nil {
					t.Fatalf("Karten lesen: %v", err)
				}

				return karten[0].ID
			},
			lauf: func(t *testing.T, db *sql.DB, vorbereitet any) error {
				karteID := vorbereitetes[int64](t, vorbereitet)
				verweis := probeverweis()
				verweis.Printing = neuerAlias

				return storage.NeueKartenverweise(db).Setzen(probekontext(t), karteID, verweis)
			},
		},
		{
			name: "Deckplatz legen",
			vorbereiten: func(t *testing.T, db *sql.DB) any {
				return deckAnlegen(t, storage.NeueDeckablage(db)).ID
			},
			lauf: func(t *testing.T, db *sql.DB, vorbereitet any) error {
				deckID := vorbereitetes[int64](t, vorbereitet)
				_, err := storage.NeueDeckablage(db).PlatzSetzen(probekontext(t), deckID, domain.Deckplatz{
					Printing: neuerAlias, Menge: 1, Bereich: domain.BereichHauptdeck,
				})

				return err
			},
		},
		{
			name:        "Preisstand eintragen",
			vorbereiten: func(*testing.T, *sql.DB) any { return nil },
			lauf: func(t *testing.T, db *sql.DB, _ any) error {
				stand := probestand(20, 417)
				stand.Printing = neuerAlias

				return storage.NeuerPreisspeicher(db).Eintragen(probekontext(t), stand)
			},
		},
		{
			name: "Deckabzug anmelden",
			vorbereiten: func(t *testing.T, db *sql.DB) any {
				turnier, err := storage.NeueTurnierablage(db).Anlegen(probekontext(t),
					domain.Turnier{Name: "Küchentisch", Spiel: "magic"})
				if err != nil {
					t.Fatalf("Turnier anlegen: %v", err)
				}

				return turnier.ID
			},
			lauf: func(t *testing.T, db *sql.DB, vorbereitet any) error {
				turnierID := vorbereitetes[int64](t, vorbereitet)
				_, err := storage.NeueTurnierablage(db).Anmelden(probekontext(t), turnierID, domain.Teilnehmer{
					Name: "Ada",
					Deckabzug: []domain.Abzugsplatz{{
						Bezeichnung: "Bolt", Printing: neuerAlias, Menge: 1,
						Bereich: domain.BereichHauptdeck, Stelle: 1,
					}},
				})

				return err
			},
		},
	}
}

func TestJedeNahtMeldetEinenDruckDenSieNichtAnlegenKann(t *testing.T) {
	for _, naht := range druckNaehte() {
		for _, stoerung := range druckstoerungen {
			t.Run(naht.name+": "+stoerung.name, func(t *testing.T) {
				db := frischeDatenbank(t)
				vorbereitet := naht.vorbereiten(t, db)

				mussAusfuehren(t, db, stoerung.satz)

				if err := naht.lauf(t, db, vorbereitet); err == nil {
					t.Error("die Zeile wurde ohne Druck geschrieben")
				}
			})
		}
	}
}

// **Und die Wege der Sammlung, die nicht am Druck scheitern.**
func TestDieSammlungMeldetIhreZugfehlerUmDenDruck(t *testing.T) {
	t.Run("das Ändern scheitert nach dem Druck an der Tabelle", func(t *testing.T) {
		db := frischeDatenbank(t)
		speicher := storage.NeuerSammlungsspeicher(db)

		if _, err := speicher.Anlegen(probekontext(t), probeexemplar()); err != nil {
			t.Fatalf("Anlegen: %v", err)
		}

		liste, err := speicher.Liste(probekontext(t))
		if err != nil {
			t.Fatalf("Liste: %v", err)
		}

		mussAusfuehren(t, db, `DROP TABLE sammlung`)

		if err := speicher.Aendern(probekontext(t), liste[0]); err == nil {
			t.Error("das Ändern ohne Tabelle kam durch")
		}
	})

	t.Run("der Abschluss des Änderns scheitert", func(t *testing.T) {
		db, err := sql.Open("stolper-druck-commit", "")
		if err != nil {
			t.Fatalf("Treiber: %v", err)
		}

		t.Cleanup(func() { _ = db.Close() })

		if err := storage.NeuerSammlungsspeicher(db).Aendern(probekontext(t), probeexemplar()); err == nil {
			t.Error("der gescheiterte Abschluss kam durch")
		}
	})

	t.Run("NachDruck verlangt ein Profil", func(t *testing.T) {
		if _, err := storage.NeuerSammlungsspeicher(frischeDatenbank(t)).
			NachDruck(context.Background(), 1); err == nil {
			t.Error("NachDruck ohne Profil kam durch")
		}
	})

	t.Run("der Verweis meldet einen geschlossenen Bestand", func(t *testing.T) {
		db, err := storage.FuerLaeufe("file:" + t.TempDir() + "/zu.db")
		if err != nil {
			t.Fatalf("öffnen: %v", err)
		}

		if err := db.Close(); err != nil {
			t.Fatalf("schließen: %v", err)
		}

		if err := storage.NeueKartenverweise(db).Setzen(probekontext(t), 1, probeverweis()); err == nil {
			t.Error("das Setzen auf dem geschlossenen Bestand kam durch")
		}
	})
}
