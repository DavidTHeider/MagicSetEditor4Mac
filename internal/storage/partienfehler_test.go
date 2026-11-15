package storage_test

// **Jeder Fehlerweg des Partiespeichers ist betreten** — Werkstück G1 von
// Phase 5.5
// (#1458).
//
// # Warum das ein eigener Lauf ist
//
// **Weil ein Fehlerweg, den niemand betritt, eine Vermutung ist.** Ein
// Speicher meldet, was SQLite hergibt; ob er es **weitergibt** oder still
// eine leere Antwort zurückgibt, sieht man nur, wenn SQLite wirklich klagt —
// und dafür muss man ihm etwas wegnehmen.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDerPartiespeicherMeldetWasSQLiteHergibt(t *testing.T) {
	t.Parallel()

	t.Run("ohne Profil geht kein Weg", func(t *testing.T) {
		t.Parallel()

		ablage := partieablage(t)
		ctx := t.Context()

		if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler == nil {
			t.Error("Anlegen ohne Profil ging durch")
		}

		if fehler := ablage.Anwenden(
			ctx, 1, "b1", einePartie(), nil); fehler == nil {
			t.Error("Anwenden ohne Profil ging durch")
		}

		if _, fehler := ablage.Standsetzen(
			ctx, "p1", domain.StandBeendet); fehler == nil {
			t.Error("Standsetzen ohne Profil ging durch")
		}
	})

	t.Run("ohne Tabellen scheitert jeder Weg", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		ablage := storage.NeuePartieablage(db)
		ctx := probekontext(t)

		verlaufBefehl(t, db, `DROP TABLE partiebefehle`)
		verlaufBefehl(t, db, `DROP TABLE partieereignisse`)
		verlaufBefehl(t, db, `DROP TABLE partien`)

		if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler == nil {
			t.Error("Anlegen ging durch")
		}

		if _, fehler := ablage.Partie(ctx, "p1"); fehler == nil {
			t.Error("Partie ging durch")
		}

		if _, fehler := ablage.Partien(ctx); fehler == nil {
			t.Error("Partien ging durch")
		}

		if fehler := ablage.Anwenden(
			ctx, 1, "b1", einePartie(), nil); fehler == nil {
			t.Error("Anwenden ging durch")
		}

		if _, _, fehler := ablage.Befehlsergebnis(
			ctx, "p1", "b1"); fehler == nil {
			t.Error("Befehlsergebnis ging durch")
		}

		if _, fehler := ablage.Ereignisse(ctx, "p1", 1, 9); fehler == nil {
			t.Error("Ereignisse ging durch")
		}

		if _, _, fehler := ablage.LetzteKette(ctx, "p1"); fehler == nil {
			t.Error("LetzteKette ging durch")
		}
	})

	// **Ohne die Ereignistabelle scheitert das Anwenden mitten darin** —
	// und der Zustand bleibt, wie er war. Das ist die Atomarität von der
	// anderen Seite: nicht „die Fassung stimmte nicht", sondern „das
	// Schreiben ging schief".
	t.Run("ein halbes Schreiben hinterlässt nichts", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		ablage := storage.NeuePartieablage(db)
		ctx := probekontext(t)

		if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		verlaufBefehl(t, db, `DROP TABLE partieereignisse`)

		danach := einePartie()
		danach.Fassung = 2
		danach.Objekte[0].Getappt = true

		if fehler := ablage.Anwenden(ctx, 1, "b1", danach,
			[]domain.Ereignis{{
				PartieID: "p1", Folge: 1, Art: "a",
				Oeffentlich: map[string]any{"sitz": "sitz-1"},
				Kette:       "k1",
			}}); fehler == nil {
			t.Fatal("das halbe Schreiben ging durch")
		}

		gelesen, fehler := ablage.Partie(ctx, "p1")
		if fehler != nil {
			t.Fatalf("lesen: %v", fehler)
		}

		// **Die Gegenprobe zur Transaktion**: der Zustand ist zurückgerollt,
		// obwohl das `UPDATE` selbst gelungen war.
		if gelesen.Fassung != 1 || gelesen.Objekte[0].Getappt {
			t.Errorf("der halbe Stand steht da: %+v", gelesen)
		}
	})

	// **Und ohne die Befehlstabelle ebenso** — das ist der zweite Schritt
	// derselben Transaktion, und er hat seinen eigenen Weg.
	t.Run("ein ungemerkter Befehl rollt den Strom zurück", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		ablage := storage.NeuePartieablage(db)
		ctx := probekontext(t)

		if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		verlaufBefehl(t, db, `DROP TABLE partiebefehle`)

		danach := einePartie()
		danach.Fassung = 2

		if fehler := ablage.Anwenden(ctx, 1, "b1", danach,
			[]domain.Ereignis{{
				PartieID: "p1", Folge: 1, Art: "a",
				Oeffentlich: map[string]any{"sitz": "sitz-1"},
				Kette:       "k1",
			}}); fehler == nil {
			t.Fatal("der ungemerkte Befehl ging durch")
		}

		strom, fehler := ablage.Ereignisse(ctx, "p1", 1, 9)
		if fehler != nil {
			t.Fatalf("Ereignisse: %v", fehler)
		}

		if len(strom) != 0 {
			t.Errorf("es stehen %d Ereignisse da", len(strom))
		}
	})

	// **Ein Befehl ohne Ereignisse schreibt nur den Zustand.**
	//
	// **Ein echter Fall**: ein Regelausleger darf zu dem Schluss kommen,
	// dass nichts geschieht. Wer dann einen Befehl mit den Folgen `0..0`
	// merkte, hätte einen Verweis auf ein Ereignis, das es nicht gibt.
	t.Run("ein Befehl ohne Ereignisse", func(t *testing.T) {
		t.Parallel()

		ablage := partieablage(t)
		ctx := probekontext(t)

		if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		danach := einePartie()
		danach.Fassung = 2

		if fehler := ablage.Anwenden(ctx, 1, "b1", danach, nil); fehler != nil {
			t.Fatalf("anwenden: %v", fehler)
		}

		gelesen, fehler := ablage.Partie(ctx, "p1")
		if fehler != nil {
			t.Fatalf("lesen: %v", fehler)
		}

		if gelesen.Fassung != 2 {
			t.Errorf("die Fassung ist %d", gelesen.Fassung)
		}
	})

	// **Was kein JSON ist, kommt nicht als Partie zurück.**
	t.Run("ein beschädigter Zustand wird nicht still repariert", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		ablage := storage.NeuePartieablage(db)
		ctx := probekontext(t)

		if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		verlaufBefehl(t, db, `UPDATE partien SET zustand = 'kein json'`)

		if _, fehler := ablage.Partie(
			ctx, "p1"); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("der beschädigte Zustand gab %v", fehler)
		}

		if _, fehler := ablage.Partien(
			ctx); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("die Liste gab %v", fehler)
		}

		if _, fehler := ablage.Standsetzen(
			ctx, "p1", domain.StandBeendet); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Errorf("Standsetzen gab %v", fehler)
		}
	})

	// **Und ein beschädigtes Ereignis ebenso.**
	t.Run("ein beschädigtes Ereignis wird nicht still repariert", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		ablage := storage.NeuePartieablage(db)
		ctx := probekontext(t)

		if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		danach := einePartie()
		danach.Fassung = 2

		if fehler := ablage.Anwenden(ctx, 1, "b1", danach,
			[]domain.Ereignis{{
				PartieID: "p1", Folge: 1, Art: "a",
				Oeffentlich: map[string]any{"sitz": "sitz-1"},
				Kette:       "k1",
			}}); fehler != nil {
			t.Fatalf("anwenden: %v", fehler)
		}

		verlaufBefehl(t, db,
			`UPDATE partieereignisse SET oeffentlich = 'kein json'`)

		if _, fehler := ablage.Ereignisse(
			ctx, "p1", 1, 9); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("das beschädigte Ereignis gab %v", fehler)
		}

		verlaufBefehl(t, db,
			`UPDATE partieereignisse SET oeffentlich = '{}', privat = 'nein'`)

		if _, fehler := ablage.Ereignisse(
			ctx, "p1", 1, 9); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("die beschädigte Ergänzung gab %v", fehler)
		}
	})
}

// **Ein Ereignis, das sich nicht kodieren lässt, wird nicht geschrieben.**
//
// **Ein echter Fall**: der öffentliche Teil eines Ereignisses ist ein
// `map[string]any`, und `any` darf alles sein. Wer einen Kanal hineinlegt,
// hat etwas hineingelegt, das kein JSON wird.
//
// **Beide Teile getrennt**, denn sie sind zwei Spalten und zwei Wege: ein
// Fehler im öffentlichen sagt nichts über die private Ergänzung.
func TestEinUnkodierbaresEreignisWirdNichtGeschrieben(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Ereignis{
		"im öffentlichen Teil": {
			PartieID: "p1", Folge: 1, Art: "a",
			Oeffentlich: map[string]any{"kanal": make(chan int)},
			Kette:       "k1",
		},
		"in der Ergänzung": {
			PartieID: "p1", Folge: 1, Art: "a",
			Oeffentlich: map[string]any{"sitz": "sitz-1"},
			Privat: []domain.Ergaenzung{{
				Empfaenger: "sitz-1",
				Inhalt:     map[string]any{"kanal": make(chan int)},
			}},
			Kette: "k1",
		},
	}

	for name, ereignis := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ablage := partieablage(t)
			ctx := probekontext(t)

			if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
				t.Fatalf("anlegen: %v", fehler)
			}

			danach := einePartie()
			danach.Fassung = 2

			if fehler := ablage.Anwenden(ctx, 1, "b1", danach,
				[]domain.Ereignis{ereignis}); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("es kam %v", fehler)
			}

			// **Und der Zustand ist zurückgerollt** — das Kodieren scheitert
			// **innerhalb** der Transaktion, nach dem `UPDATE`.
			gelesen, fehler := ablage.Partie(ctx, "p1")
			if fehler != nil {
				t.Fatalf("lesen: %v", fehler)
			}

			if gelesen.Fassung != 1 {
				t.Errorf("die Fassung ist %d", gelesen.Fassung)
			}
		})
	}
}
