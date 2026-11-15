package storage

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Was hinter der Frage nach dem Besitz noch „gibt es nicht" sagt** —
// #1815.
//
// Die Frage nach dem Besitz beantwortet für die meisten Wege auch die, ob es
// eine Kennung gibt. Hier stehen die Stellen, an denen das nicht reicht — und
// die Helfer, die ihre eigene Zusage behalten, obwohl ihr heutiger Rufer sie
// schon gefragt hat.

// **Dieselbe Karte zweimal gelöscht** ist „nicht gefunden", und keine geht:
// nach dem ersten Löschen gibt es sie im selben Zug nicht mehr. So verhielt
// es sich vor #1815 auch; die Frage nach dem Besitz fragt davor, als es sie
// noch gab.
func TestDieselbeKarteZweimalGeloeschtIstNichtGefunden(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)
	eigner := MitProfil(t.Context(), 2)

	fehler := NeuerKartenschreiber(db).KartenLoeschen(eigner, []int64{fremd.karte, fremd.karte})
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("zweimal dieselbe Karte: %v, erwartet nicht gefunden", fehler)
	}

	if name, lesefehler := NeuerSetLeser(db).NameZurKarte(eigner, fremd.karte); lesefehler != nil || name != "Fremde Karte" {
		t.Errorf("die Karte ist fort: %q (%v)", name, lesefehler)
	}
}

// **Die Helfer halten ihre Zusage auch allein.** Ihr Rufer fragt heute
// vorher nach dem Besitz; fiele die Frage dort einmal weg, sagten sie immer
// noch „gibt es nicht" statt mit einer leeren Stelle weiterzurechnen.
func TestDieHelferSagenAuchAllein(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)

	geschaeft, fehler := db.BeginTx(t.Context(), nil)
	if fehler != nil {
		t.Fatalf("beginnen: %v", fehler)
	}

	t.Cleanup(func() { _ = geschaeft.Rollback() })

	if _, fehler := stellenLesen(t.Context(), geschaeft, fremd.karte, 999); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("die Stellen mit einer Karte, die es nicht gibt: %v", fehler)
	}

	if _, fehler := turnierZurPaarung(t.Context(), geschaeft, 999); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("das Turnier einer Paarung, die es nicht gibt: %v", fehler)
	}

	if _, _, _, fehler := NeuerKartenschreiber(db).imLauf(MitProfil(t.Context(), 2), geschaeft, 999, "name",
		func(*sql.Tx) error { return nil }); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("eine Feldänderung an einer Karte, die es nicht gibt: %v", fehler)
	}
}

// **Auch im eigenen Turnier zeigt nichts auf Fremdes**: eine Bindung an eine
// Partie, eine Paarung mit einem Teilnehmer und ein Teilnehmer mit einem Deck,
// die es im Turnier oder im Profil nicht gibt, sind „nicht gefunden".
func TestEinEigenesTurnierZeigtAufNichtsFremdes(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)
	eigner := MitProfil(t.Context(), 2)

	ausFremdemProfil(t, db, map[string]func() error{
		"Bindung an eine fremde Partie": func() error {
			return NeueMatchablage(db).BindungSchreiben(eigner, domain.Matchbindung{
				TurnierID: fremd.turnier, PaarungID: fremd.paarung, Fassung: 1, PartieID: "gibt-es-nicht",
			})
		},
		"Runde mit einem Teilnehmer, der nicht im Turnier steht": func() error {
			_, f := NeueTurnierablage(db).RundeLegen(eigner, fremd.turnier, 2,
				[]domain.Paarung{{A: fremd.teilnehmer, B: 999}})
			return f
		},
		"Teilnehmer mit fremdem Deck": func() error {
			_, f := NeueTurnierablage(db).Anmelden(eigner, fremd.turnier, domain.Teilnehmer{Name: "Cleo", DeckID: 999})
			return f
		},
	})
}
