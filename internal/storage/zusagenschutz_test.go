package storage_test

// **Die Zusage hält gegen den gewöhnlichen Schreibweg** — Werkstück `T2` von
// Phase 5.6
// (#1467).
//
// # Warum das im Speicher steht und nicht im Dienst
//
// **Weil dazwischen ein Fenster liegt.** Ein Dienst, der erst die Zusagen
// liest und dann schreibt, hat zwischen beidem einen Augenblick, in dem eine
// zweite Zusage dazukommen kann — und die wäre dann gegen einen Besitz
// gerechnet, den der erste Zug gerade wegschreibt. Dieselbe Überlegung wie bei
// [storage.Reservierungsablage.Festhalten]: die **Regel** steht in der Domäne,
// der **Augenblick** im Bestand.
//
// # Was dieser Lauf nicht hält
//
// **Dass zwei Prozesse es auch tun.** Gemessen wird hier ein Bestand in einer
// Datei mit einer Verbindung; dass `_txlock=immediate` zwei Schreiber
// serialisiert, steht in `stolperreservierungen_test.go` und hat dort seinen
// eigenen Grund.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// mitZusage legt ein Exemplar mit drei Stück und sagt zwei davon zu.
func mitZusage(t *testing.T) (
	speicher *storage.Sammlungsspeicher, exemplarID int64,
) {
	t.Helper()

	db := frischeDatenbank(t)
	speicher = storage.NeuerSammlungsspeicher(db)

	exemplarID, err := speicher.Anlegen(probekontext(t), probeexemplar())
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := storage.NeueReservierungsablage(db).Festhalten(probekontext(t),
		domain.Reservierung{
			Vorgang: "h1", Revision: 1, ExemplarID: exemplarID, Menge: 2, Fassung: 1,
		}, ueber(probeexemplar(), 99)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	return speicher, exemplarID
}

// **Eine Änderung unter die Zusage wird abgewiesen — und schreibt nichts.**
func TestEinSchreibwegNimmtKeineZugesagtenStuecke(t *testing.T) {
	speicher, id := mitZusage(t)
	ctx := probekontext(t)

	knapp := probeexemplar()
	knapp.ID, knapp.Fassung, knapp.Menge = id, 1, 1

	if fehler := speicher.Aendern(ctx, knapp); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	// **Und der Bestand steht, wo er stand.** Ein abgewiesener Zug, der die
	// Hälfte geschrieben hat, wäre schlimmer als einer, der durchgeht (§7,
	// Regel 3).
	stand, err := speicher.Einzeln(ctx, id)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if stand.Menge != 3 || stand.Fassung != 1 {
		t.Errorf("im Bestand stehen %d Stück in Fassung %d", stand.Menge, stand.Fassung)
	}
}

// **Über der Zusage geht dieselbe Änderung durch.**
//
// **Die Gegenprobe**: ohne sie hielte auch ein Speicher, der jede Änderung an
// einer zugesagten Zeile abweist — und eine Sammlung mit einem offenen Handel
// wäre nicht mehr zu pflegen.
func TestUeberDerZusageGehtDieAenderungDurch(t *testing.T) {
	speicher, id := mitZusage(t)
	ctx := probekontext(t)

	reicht := probeexemplar()
	reicht.ID, reicht.Fassung, reicht.Menge = id, 1, 2

	if fehler := speicher.Aendern(ctx, reicht); fehler != nil {
		t.Fatalf("es kam %v", fehler)
	}

	stand, err := speicher.Einzeln(ctx, id)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if stand.Menge != 2 {
		t.Errorf("im Bestand stehen %d Stück", stand.Menge)
	}
}

// **Und gelöscht wird eine zugesagte Zeile gar nicht.**
//
// **Sonst nähme die Kaskade die Zusage mit** (Wanderung 0069) — der Handel
// wüsste nichts mehr davon, und der Partner wartete auf eine Karte, von der
// dieses Programm nichts mehr weiß.
func TestEineZugesagteZeileWirdNichtGeloescht(t *testing.T) {
	speicher, id := mitZusage(t)
	ctx := probekontext(t)

	if fehler := speicher.Loeschen(ctx, id, 1); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	if _, err := speicher.Einzeln(ctx, id); err != nil {
		t.Errorf("die Zeile ist fort: %v", err)
	}
}
