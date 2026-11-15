package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Hinter der Frage nach dem Besitz hat jeder Weg seinen eigenen Fehler** —
// #1815.
//
// # Warum es diesen Lauf gibt
//
// Bis #1815 erreichte ein geschlossener Bestand jeden dieser Fehlerwege.
// Seitdem scheitert er schon an der Frage, ob die Kennung dem Profil gehört —
// und der Weg dahinter blieb unbetreten, während der alte Lauf grün blieb,
// aus dem falschen Grund.
//
// **Hier fehlt deshalb genau die eine Tafel, die der Ruf nach der Frage
// liest**: die Frage geht durch, das Lesen oder Schreiben nicht. Dieselbe
// Vorrichtung wie in `setpackarten_test.go`. Wo die Frage selbst die Tafel
// braucht — `partien` —, bricht ein Auslöser das Schreiben ab.
func TestHinterDerBesitzfrageHatJederWegSeinenFehler(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		// vorbereitung nimmt dem Ruf den Weg hinter der Frage — meist eine
		// fehlende Tafel (Kinder vor ihren Eltern verworfen), wo die Tafel
		// selbst gebraucht wird, ein Auslöser, der das Schreiben abbricht.
		vorbereitung []string

		// satz ist, was die Meldung nennen muss: **wobei** es scheiterte.
		satz string

		ruf func(context.Context, *sql.DB, fremdeKennungen) error
	}{
		"ein Feld schreiben": {
			[]string{`DROP TABLE card_values`},
			`Feld "name" schreiben`,
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				return NeuerKartenschreiber(db).FeldSetzen(c, k.karte, "name", "X")
			},
		},
		"eine Marke entfernen": {
			[]string{`DROP TABLE karten_marken`},
			"eine Marke entfernen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				return NeueMarken(db).Entfernen(c, k.karte, "fremd")
			},
		},
		"die Marken des Sets": {
			[]string{`DROP TABLE karten_marken`},
			"die Marken des Sets lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeueMarken(db).JeKarte(c, k.set)
				return f
			},
		},
		"die Marken einer Karte": {
			[]string{`DROP TABLE karten_marken`},
			"Marken lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeueMarken(db).AnKarte(c, k.karte)
				return f
			},
		},
		"einen Verweis setzen": {
			[]string{`DROP TABLE karten_verweis`},
			"den Verweis der Karte 200 setzen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				return NeueKartenverweise(db).Setzen(c, k.karte, domain.Kartenverweis{Printing: "druck-fremd", Name: "X"})
			},
		},
		"einen Verweis lesen": {
			[]string{`DROP TABLE karten_verweis`},
			"den Verweis der Karte 200 lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeueKartenverweise(db).Fuer(c, k.karte)
				return f
			},
		},
		"einen Verweis lösen": {
			[]string{`DROP TABLE karten_verweis`},
			"den Verweis der Karte 200 lösen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeueKartenverweise(db).Loesen(c, k.karte)
				return f
			},
		},
		"die Verweise des Sets": {
			[]string{`DROP TABLE karten_verweis`},
			"die Verweise des Sets 20 lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeueKartenverweise(db).JeSet(c, k.set)
				return f
			},
		},
		"die Karten": {
			[]string{`DROP TABLE card_values`},
			"Karten lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuerSetLeser(db).Karten(c, k.set)
				return f
			},
		},
		"die Rohblöcke": {
			[]string{`DROP TABLE raw_blocks`},
			"Rohblöcke lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuerSetLeser(db).Rohbloecke(c, k.set)
				return f
			},
		},
		"die Kartenstile": {
			[]string{`DROP TABLE card_values`},
			"die Kartenstile lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuerSetLeser(db).Kartenstile(c, k.set)
				return f
			},
		},
		// **Eine Sicht mit NULL** statt der Tafel — der Scan in eine
		// Zeichenkette scheitert (dieselbe Gestalt wie im Arbeitsverlauf).
		"ein Kartenstil": {
			[]string{
				`DROP TABLE card_values`,
				`CREATE VIEW card_values AS SELECT id AS card_id, 'stylesheet' AS feld, NULL AS wert, 0 AS reihenfolge FROM cards`,
			},
			"einen Kartenstil lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuerSetLeser(db).Kartenstile(c, k.set)
				return f
			},
		},
		"die Setangaben": {
			[]string{`DROP TABLE set_values`},
			"Setangaben lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuerSetLeser(db).Setwerte(c, k.set)
				return f
			},
		},
		"die Belegung": {
			[]string{`DROP TABLE card_values`},
			"die Belegung lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuerSetLeser(db).Belegung(c, k.set)
				return f
			},
		},
		"die Verteilung": {
			[]string{`DROP TABLE card_values`},
			"die Verteilung lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuerSetLeser(db).Verteilung(c, k.set, "name")
				return f
			},
		},
		"einen Deckplatz löschen": {
			[]string{`DROP TABLE deckzuordnungen`, `DROP TABLE deckplaetze`},
			"den Deckplatz löschen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				return NeueDeckablage(db).PlatzLoeschen(c, k.deck, k.platz)
			},
		},
		"die Zuordnungen": {
			[]string{`DROP TABLE deckzuordnungen`},
			"die Zuordnungen lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeueDeckzuordnungen(db).JeDeck(c, k.deck)
				return f
			},
		},
		"die Herkunft": {
			[]string{`DROP TABLE deck_herkunft`},
			"die Herkunft lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeueHerkunftsablage(db).Deckherkunft(c, k.deck)
				return f
			},
		},
		"eine Bindung lesen": {
			[]string{`DROP TABLE matchmeldungen`, `DROP TABLE matchbindungen`},
			"die Bindung lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeueMatchablage(db).Bindung(c, k.paarung)
				return f
			},
		},
		"eine Meldung schreiben": {
			[]string{`DROP TABLE matchmeldungen`},
			"die Meldung schreiben",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				return NeueMatchablage(db).MeldungSchreiben(c, domain.Matchmeldung{
					PaarungID: k.paarung, Von: "sitz2", Fassung: 1, PartieID: k.partie, Kette: "k",
				})
			},
		},
		"die Meldungen": {
			[]string{`DROP TABLE matchmeldungen`},
			"die Meldungen lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeueMatchablage(db).Meldungen(c, k.paarung)
				return f
			},
		},
		"die Freigaben einer Partie": {
			[]string{`DROP TABLE partiefreigaben`},
			"die Freigaben lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuePartieablage(db).Freigaben(c, k.partie)
				return f
			},
		},
		"ein Befehl": {
			[]string{`DROP TABLE partiebefehle`},
			"den Befehl nachschlagen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, _, f := NeuePartieablage(db).Befehlsergebnis(c, k.partie, "befehl-fremd")
				return f
			},
		},
		"der Strom": {
			[]string{`DROP TABLE partieereignisse`},
			"den Strom lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuePartieablage(db).Ereignisse(c, k.partie, 0, 0)
				return f
			},
		},
		"der letzte Befehl": {
			[]string{`DROP TABLE partiebefehle`},
			"den letzten Befehl nachschlagen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuePartieablage(db).LetzterBefehl(c, k.partie)
				return f
			},
		},
		"die Kette": {
			[]string{`DROP TABLE partieereignisse`},
			"die Kette lesen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, _, f := NeuePartieablage(db).LetzteKette(c, k.partie)
				return f
			},
		},
		"eine Gruppe schreiben": {
			[]string{`DROP TABLE card_values`},
			`den vorigen Wert von "name" lesen`,
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				return NeuerKartenschreiber(db).GruppeSetzenMitHerkunft(c, []domain.Kartenschritt{{
					KarteID: k.karte, Werte: []domain.Value{{Field: "name", Text: "X"}},
					Herkunft: "ki", Herkuenfte: []domain.KIFeldherkunft{probeherkunft()},
				}})
			},
		},
		"eine Partie fortschreiben": {
			[]string{`CREATE TRIGGER stolpern BEFORE UPDATE ON partien
			          BEGIN SELECT RAISE(ABORT, 'gestolpert'); END`},
			"die Partie fortschreiben",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				return NeuePartieablage(db).Anwenden(c, 1, "befehl-neu", domain.Partie{ID: k.partie, Fassung: 2}, nil)
			},
		},
		"die Offenlegungen zählen": {
			[]string{`DROP TABLE partiefreigaben`},
			"die Freigaben zählen",
			func(c context.Context, db *sql.DB, k fremdeKennungen) error {
				_, f := NeuePartieablage(db).Offenlegungszahl(c, k.partie)
				return f
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, kennungen := besitzbestand(t)

			for _, anweisung := range fall.vorbereitung {
				if _, fehler := db.ExecContext(t.Context(), anweisung); fehler != nil {
					t.Fatalf("%s: %v", anweisung, fehler)
				}
			}

			fehler := fall.ruf(MitProfil(t.Context(), 2), db, kennungen)

			switch {
			case fehler == nil:
				t.Fatal("ohne die Tafel kam der Ruf durch")
			case errors.Is(fehler, domain.ErrNotFound):
				t.Fatalf("gescheitert an der Frage nach dem Besitz statt am Weg dahinter: %v", fehler)
			case !strings.Contains(fehler.Error(), fall.satz):
				t.Errorf("die Meldung %q nennt nicht %q", fehler, fall.satz)
			}
		})
	}
}

// probeherkunft ist eine vollständige Herkunft für das Feld »name«.
func probeherkunft() domain.KIFeldherkunft {
	return domain.KIFeldherkunft{
		Feld: "name", Anbieter: "ollama", Modell: "gemma4:8b-q4",
		Aufgabe: "Schreibe den Namen.", Auftrag: "ki-1815",
		Zeitpunkt: time.Date(2026, 9, 23, 9, 0, 0, 0, time.UTC),
		Bezug:     domain.KIInhaltsbezug([]string{"X"}),
	}
}
