package storage_test

// **Die Liste der Handelsvorgänge und ihre Entwürfe** — TRADE-01
// (#1715).

import (
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// benannterDruck legt einen Druck mit Scryfall-Kennung an.
func benannterDruck(t *testing.T, db *sql.DB, name, scryfall string) {
	t.Helper()

	ergebnis, fehler := db.ExecContext(t.Context(),
		`INSERT INTO drucke (name, setcode, nummer, sprache) VALUES (?, 'm11', '1', 'en')`, name)
	if fehler != nil {
		t.Fatalf("Druck: %v", fehler)
	}

	id, _ := ergebnis.LastInsertId() //nolint:errcheck // SQLite gibt sie immer

	if _, fehler := db.ExecContext(t.Context(),
		`INSERT INTO druckkennungen (druck_id, system, semantik, wert) VALUES (?, 'scryfall', 'druck', ?)`,
		id, scryfall); fehler != nil {
		t.Fatalf("Kennung: %v", fehler)
	}
}

// **Je Vorgang die letzte Revision — mit Anfang, letzter Änderung, den
// Namen seiner Karten und dem Archiv.**
func TestDieUebersichtNenntZeitenKartenUndArchiv(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueHandelsablage(db)
	ctx := probekontext(t)

	benannterDruck(t, db, "Lightning Bolt", "abc-123")

	erste := einVorgangsangebot()
	zweite := einVorgangsangebot(func(a *domain.Handelsangebot) {
		a.Revision, a.Eltern, a.Zustand = 2, erste.Abdruck(), domain.HandelBeidseitigVerbucht
	})
	fremde := einVorgangsangebot(func(a *domain.Handelsangebot) {
		a.Vorgang = "h2"
		a.Positionen[0].Printing = "unbekannt"
	})

	for _, angebot := range []domain.Handelsangebot{erste, zweite, fremde} {
		if fehler := ablage.Schreiben(ctx, angebot); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	for _, befehl := range []string{
		`UPDATE handelsrevisionen SET angelegt = '2026-01-01 08:00:00' WHERE vorgang = 'h1' AND revision = 1`,
		`UPDATE handelsrevisionen SET geaendert = '2026-02-01 09:00:00' WHERE vorgang = 'h1' AND revision = 2`,
	} {
		if _, fehler := db.ExecContext(ctx, befehl); fehler != nil {
			t.Fatalf("datieren: %v", fehler)
		}
	}

	if fehler := ablage.Archivieren(ctx, "h1", true); fehler != nil {
		t.Fatalf("archivieren: %v", fehler)
	}

	zeilen, fehler := ablage.Uebersicht(ctx)
	if fehler != nil {
		t.Fatalf("Übersicht: %v", fehler)
	}

	if len(zeilen) != 2 || zeilen[0].Angebot.Vorgang != "h2" {
		t.Fatalf("Zeilen: %+v", zeilen)
	}

	h1 := zeilen[1]
	if h1.Angebot.Revision != 2 || !h1.Archiviert ||
		!h1.Angelegt.Equal(time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)) ||
		!h1.Geaendert.Equal(time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)) ||
		!slices.Equal(h1.Karten, []string{"Lightning Bolt"}) {
		t.Errorf("h1: %+v", h1)
	}

	// **Ein Druck, den dieser Bestand nicht kennt, hat keinen Namen** — und
	// keine leere Zeichenkette an seiner Stelle.
	if zeilen[0].Archiviert || len(zeilen[0].Karten) != 0 {
		t.Errorf("h2: %+v", zeilen[0])
	}

	if zurueck := ablage.Archivieren(ctx, "h1", false); zurueck != nil {
		t.Fatalf("zurückholen: %v", zurueck)
	}

	if zeilen, fehler = ablage.Uebersicht(ctx); fehler != nil || zeilen[1].Archiviert {
		t.Errorf("h1 blieb im Archiv (%v)", fehler)
	}
}

// **Ein Entwurf wird ersetzt und gelöscht — nichts sonst.**
func TestNurEinEntwurfLaesstSichErsetzenUndLoeschen(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueHandelsablage(db)
	ctx := probekontext(t)

	entwurf := einVorgangsangebot()
	angeboten := einVorgangsangebot(func(a *domain.Handelsangebot) {
		a.Vorgang, a.Zustand = "h2", domain.HandelAngeboten
	})

	for _, angebot := range []domain.Handelsangebot{entwurf, angeboten} {
		if fehler := ablage.Schreiben(ctx, angebot); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	geaendert := einVorgangsangebot(func(a *domain.Handelsangebot) { a.Positionen[0].Menge = 5 })
	if fehler := ablage.EntwurfErsetzen(ctx, geaendert); fehler != nil {
		t.Fatalf("ersetzen: %v", fehler)
	}

	if gelesen, fehler := ablage.Vorgang(ctx, "h1"); fehler != nil || gelesen.Positionen[0].Menge != 5 {
		t.Errorf("der Entwurf blieb: %+v (%v)", gelesen, fehler)
	}

	for _, fall := range []struct {
		name     string
		vorgang  string
		erwartet error
	}{
		{"ein angebotener", "h2", domain.ErrConflict},
		{"ein fehlender", "h9", domain.ErrNotFound},
	} {
		angebot := einVorgangsangebot(func(a *domain.Handelsangebot) { a.Vorgang = fall.vorgang })
		if fehler := ablage.EntwurfErsetzen(ctx, angebot); !errors.Is(fehler, fall.erwartet) {
			t.Errorf("ersetzen, %s: %v", fall.name, fehler)
		}

		if fehler := ablage.EntwurfLoeschen(ctx, fall.vorgang); !errors.Is(fehler, fall.erwartet) {
			t.Errorf("löschen, %s: %v", fall.name, fehler)
		}
	}

	marken := storage.NeueObjektmarkenablage(db)
	if fehler := marken.Setzen(ctx, domain.MarkenartHandel, "h1", "messe"); fehler != nil {
		t.Fatalf("Marke: %v", fehler)
	}

	if fehler := ablage.EntwurfLoeschen(ctx, "h1"); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if _, fehler := ablage.Vorgang(ctx, "h1"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("der Entwurf ist noch da: %v", fehler)
	}

	// **Seine Marke geht mit** — der Auslöser aus 0093.
	var uebrig int
	if fehler := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM objektmarken WHERE kennung = 'h1'`).Scan(&uebrig); fehler != nil || uebrig != 0 {
		t.Errorf("%d Marken blieben (%v)", uebrig, fehler)
	}
}

// **Archivieren braucht einen Vorgang.**
func TestArchivierenBrauchtEinenVorgang(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueHandelsablage(frischeDatenbank(t))
	if fehler := ablage.Archivieren(probekontext(t), "h9", true); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v", fehler)
	}
}
