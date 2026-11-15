package storage_test

// **Was Protokoll und bedingtes Schreiben melden, wenn der Bestand nicht
// mitspielt** — TRADE-03 (#1717).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDasProtokollOhneProfilUndOhneTafel(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueHandelsablage(db)

	if _, fehler := ablage.Protokoll(t.Context(), "v"); fehler == nil {
		t.Error("gelesen ohne Profil")
	}

	if _, fehler := db.ExecContext(t.Context(), `DROP TABLE handelsprotokoll`); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := ablage.Protokoll(probekontext(t), "v"); fehler == nil {
		t.Error("gelesen ohne Tafel")
	}
}

func TestEinKrummerZeitpunktImProtokoll(t *testing.T) {
	db := frischeDatenbank(t)

	// **Ein Einfügen von Hand geht** — gesperrt sind Ändern und Löschen.
	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO handelsprotokoll (profil_id, vorgang, revision, von, nach, zeit)
		 VALUES (1, 'v', 1, '', 'entwurf', 'gestern')`); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := storage.NeueHandelsablage(db).Protokoll(probekontext(t), "v"); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Protokoll: %v", fehler)
	}
}

func TestFortschreibenOhneTafel(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := db.ExecContext(t.Context(), `DROP TABLE handelsrevisionen`); fehler != nil {
		t.Fatal(fehler)
	}

	vorher := einAngebotZumFortschreiben()
	if fehler := storage.NeueHandelsablage(db).Fortschreiben(probekontext(t), vorher, vorher); fehler == nil ||
		errors.Is(fehler, domain.ErrUeberholt) {
		t.Fatalf("Fortschreiben ohne Tafel: %v", fehler)
	}
}
