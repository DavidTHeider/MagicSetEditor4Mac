package storage

// **Was die Übersicht und die Entwürfe abweisen** — TRADE-01
// (#1715).
//
// Ohne Profil, an einer geschlossenen Datenbank, an einer Zeile, die kein
// Angebot ist, und — über die Stolpertreiber — mitten im Lesen.

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// entwurfsangebot ist ein kleinster Entwurf.
func entwurfsangebot(vorgang string) domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung, Vorgang: vorgang, Revision: 1,
		Art: domain.HandelTausch, Initiator: "anna", Parteien: []string{"anna", "bert"},
		Zustand: domain.HandelEntwurf,
	}
}

// alleWege ruft jeden Weg dieser Datei und gibt je Weg den Fehler.
func alleWege(ctx context.Context, ablage *Handelsablage) map[string]error {
	_, uebersicht := ablage.Uebersicht(ctx)

	return map[string]error{
		"Übersicht":   uebersicht,
		"ersetzen":    ablage.EntwurfErsetzen(ctx, entwurfsangebot("h1")),
		"löschen":     ablage.EntwurfLoeschen(ctx, "h1"),
		"archivieren": ablage.Archivieren(ctx, "h1", true),
	}
}

func TestDieUebersichtBrauchtEinProfilUndEineOffeneDatenbank(t *testing.T) {
	t.Parallel()

	for weg, fehler := range alleWege(t.Context(), NeueHandelsablage(migrierteDatenbank(t))) {
		if fehler == nil {
			t.Errorf("%s ohne Profil ging durch", weg)
		}
	}

	db := migrierteDatenbank(t)
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	for weg, fehler := range alleWege(probekontext(t), NeueHandelsablage(db)) {
		if fehler == nil {
			t.Errorf("%s an einer geschlossenen Datenbank ging durch", weg)
		}
	}
}

// **Mitten im Lesen**: eine krumme Zeile, ein Abbruch — und ein Ändern, das
// keine Zeile traf, dessen Rückfrage dann scheitert.
func TestDieUebersichtMeldetIhreLesefehler(t *testing.T) {
	for _, treiber := range []string{"stolper-umwandeln", "stolper-lesen"} {
		t.Run(treiber, func(t *testing.T) {
			db, fehler := sql.Open(treiber, "")
			if fehler != nil {
				t.Fatalf("Stolpertreiber: %v", fehler)
			}

			t.Cleanup(func() { _ = db.Close() })

			for weg, fehler := range alleWege(probekontext(t), NeueHandelsablage(db)) {
				if fehler == nil {
					t.Errorf("%s: der Lesefehler ging verloren", weg)
				}
			}
		})
	}
}

// **Was kein Angebot ist, ist ein Befund** — und ein Archiv, dessen
// Bedingung eine von Hand geschriebene Zeile verletzt, meldet sich.
func TestDieUebersichtAnVerbogenenZeilen(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	ablage, ctx := NeueHandelsablage(db), probekontext(t)

	if fehler := ablage.Schreiben(ctx, entwurfsangebot("h1")); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	// **Für SQLite ein JSON, für Go kein Angebot** — sonst scheiterte schon
	// `json_each` in der Abfrage und nicht das Entpacken.
	if _, fehler := db.ExecContext(ctx,
		`UPDATE handelsrevisionen SET angebot = '{"Revision": "eins"}'`); fehler != nil {
		t.Fatalf("verbiegen: %v", fehler)
	}

	if _, fehler := ablage.Uebersicht(ctx); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Übersicht: %v", fehler)
	}

	if _, fehler := db.ExecContext(ctx, `UPDATE handelsrevisionen SET vorgang = ''`); fehler != nil {
		t.Fatalf("verbiegen: %v", fehler)
	}

	if fehler := ablage.Archivieren(ctx, "", true); fehler == nil {
		t.Error("ein Archiv ohne Vorgang ging durch")
	}
}
