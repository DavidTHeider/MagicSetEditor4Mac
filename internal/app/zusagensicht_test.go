package app_test

// **Was eine Zusage in der Deckzuordnung sichtbar macht** — Werkstück `T2`
// von Phase 5.6
// (#1467).
//
// # Warum der Lauf hier steht
//
// **Weil die Naht hier verdrahtet wird.** `service.Deckzuordnungspflege` hat
// seit U9 (#1420) ein Loch für die Zusagen und rechnete bis heute mit **null**
// — der Satz stand als Vertrag in ihrem Kopf. Ob es gefüllt ist, entscheiden
// die beiden Wurzeln in `internal/app`, und ein Dienst-Test sähe es nicht: er
// reicht die Naht selbst herein.
//
// **Beide Wurzeln, nicht eine.** `deckstart.go` baut das Fenster,
// `decknahtflaechen.go` die Flächen ohne Fenster; eine verdrahtete und eine
// vergessene Wurzel sind zwei Anwendungen (§16).

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// probedruckkennung ist der Druck, um den es in diesem Lauf geht.
const probedruckkennung = "a1b2c3d4-0000-4000-8000-00000000000a"

// zugesagtesExemplar legt drei Stück an und sagt zwei davon einem Handel zu.
func zugesagtesExemplar(t *testing.T, db *sql.DB, ctx context.Context) {
	t.Helper()

	id, err := storage.NeuerSammlungsspeicher(db).Anlegen(ctx, &domain.Exemplar{
		Printing: probedruckkennung, Menge: 3,
		Zustand: domain.ZustandNearMint, Sprache: "en",
	})
	if err != nil {
		t.Fatalf("Exemplar: %v", err)
	}

	if fehler := storage.NeueReservierungsablage(db).Festhalten(ctx,
		domain.Reservierung{
			Vorgang: "h1", Revision: 1, ExemplarID: id, Menge: 2, Fassung: 1,
		}, domain.Zusagevertrag{{
			Printing: probedruckkennung, Menge: 99, Guete: "nearMint", Sprache: "en",
		}}); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}
}

// **Eine offene Zusage steht in der Zuordnung — an der Fläche ohne Fenster.**
//
// **Zwei zugesagt, drei besessen, keines zugeordnet — also eines frei.**
func TestEineZusageStehtInDerZuordnungDerFlaeche(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(app.Scryfallumgebung, stilleGegenstelle(t))

	db := frischerBestand(t)
	ctx := probekontext(t)

	zugesagtesExemplar(t, db, ctx)

	flaeche, fehler := app.DeckflaecheUeber(db)
	if fehler != nil {
		t.Fatalf("DeckflaecheUeber: %v", fehler)
	}

	deck := flaeche.DeckAnlegen(ctx, "Aggro", "magic", "")
	if deck.Fehler != nil {
		t.Fatalf("Deck: %+v", deck.Fehler)
	}

	if platz := flaeche.DeckplatzLegen(
		ctx, deck.Deck.ID, "", probedruckkennung, 4, "hauptdeck",
	); platz.Fehler != nil {
		t.Fatalf("Platz: %+v", platz.Fehler)
	}

	stand := flaeche.DeckZuordnungen(ctx, deck.Deck.ID)
	if stand.Fehler != nil || len(stand.Zeilen) != 1 {
		t.Fatalf("Stand: %+v", stand)
	}

	if zeile := stand.Zeilen[0]; zeile.Reserviert != 2 || zeile.Frei != 1 {
		t.Errorf("zugesagt %d, frei %d — erwartet waren 2 und 1 (%+v)",
			zeile.Reserviert, zeile.Frei, zeile)
	}
}

// **Und dasselbe am Fenster** — die zweite Wurzel.
//
// **Nicht dieselbe Zeile zweimal.** `deckstart.go` verdrahtet die Naht
// eigenständig; wer nur die Fläche prüfte, hätte eine Anwendung, in der das
// Fenster null zeigt und die Kommandozeile zwei — und würde es erst im
// Betrieb sehen.
func TestEineZusageStehtInDerZuordnungAmFenster(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	t.Setenv(app.Scryfallumgebung, stilleGegenstelle(t))

	bestand, err := app.Datenbankpfad()
	if err != nil {
		t.Fatalf("Datenbankpfad: %v", err)
	}

	if fehler := os.MkdirAll(filepath.Dir(bestand), 0o700); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	anwendung := app.NeueApp(bestand)
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	// **Über denselben Bestand, den die Anwendung geöffnet hat** — ein
	// zweiter Griff an dieselbe Datei, wie ihn auch ein Handelsvorgang täte.
	db, oeffnungsfehler := storage.Oeffnen(probekontext(t), bestand)
	if oeffnungsfehler != nil {
		t.Fatalf("öffnen: %v", oeffnungsfehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	zugesagtesExemplar(t, db, probekontext(t))

	deck := anwendung.DeckAnlegen(t.Context(), "Aggro", "magic", "")
	if deck.Fehler != nil {
		t.Fatalf("Deck: %+v", deck.Fehler)
	}

	if platz := anwendung.DeckplatzLegen(t.Context(), deck.Deck.ID, "",
		probedruckkennung, 4, "hauptdeck"); platz.Fehler != nil {
		t.Fatalf("Platz: %+v", platz.Fehler)
	}

	stand := anwendung.DeckZuordnungen(t.Context(), deck.Deck.ID)
	if stand.Fehler != nil || len(stand.Zeilen) != 1 {
		t.Fatalf("Stand: %+v", stand)
	}

	if zeile := stand.Zeilen[0]; zeile.Reserviert != 2 || zeile.Frei != 1 {
		t.Errorf("zugesagt %d, frei %d — erwartet waren 2 und 1 (%+v)",
			zeile.Reserviert, zeile.Frei, zeile)
	}
}
