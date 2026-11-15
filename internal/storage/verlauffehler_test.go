// Die Fehlerwege des Verlaufs.
//
// **Über eine geschlossene Datenbank.** Jede Abfrage scheitert dann sofort —
// das ist die kürzeste Art, einen Weg zu betreten, den eine kaputte Platte
// sonst herbeiführen müsste.
package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestVerlaufsfehler(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	// Ein Eintrag, damit es etwas zu bewegen gäbe.
	if _, _, err := storage.NeuerKartenschreiber(db).
		FeldSetzenMitVorwert(probekontext(t), karteID, "name", "Eisball"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	leser := storage.NeuerVerlaufsleser(db)

	t.Run("lesen", func(t *testing.T) {
		if _, err := leser.Verlauf(probekontext(t), setID); err == nil {
			t.Error("kein Fehler")
		}
	})

	t.Run("zurücknehmen", func(t *testing.T) {
		if _, _, err := leser.Zuruecknehmen(probekontext(t), setID); err == nil {
			t.Error("kein Fehler")
		}
	})

	t.Run("wiederherstellen", func(t *testing.T) {
		if _, _, err := leser.Wiederherstellen(probekontext(t), setID); err == nil {
			t.Error("kein Fehler")
		}
	})

	t.Run("schreiben", func(t *testing.T) {
		if _, _, err := storage.NeuerKartenschreiber(db).
			FeldSetzenMitVorwert(probekontext(t), karteID, "name", "Steinwand"); err == nil {
			t.Error("kein Fehler")
		}
	})
}

// **Ein leerer Verlauf ist kein Fehler.**
//
// ⌘Z am Anfang ist eine gewöhnliche Lage: es gibt nichts zurückzunehmen, und
// niemand will dafür eine Meldung.
func TestEinLeererVerlaufIstKeinFehler(t *testing.T) {
	db, setID, _ := bestandMitKarte(t)

	leser := storage.NeuerVerlaufsleser(db)

	_, genommen, err := leser.Zuruecknehmen(probekontext(t), setID)
	if err != nil || genommen {
		t.Errorf("zurückgenommen: %v / %v", genommen, err)
	}

	_, gebracht, err := leser.Wiederherstellen(probekontext(t), setID)
	if err != nil || gebracht {
		t.Errorf("wiederhergestellt: %v / %v", gebracht, err)
	}
}

// **Ein Verlaufseintrag mit kaputtem Weg wird gemeldet, nicht geraten.**
//
// Er kann aus einer älteren Fassung stammen oder aus einer Datenbank, an der
// jemand von Hand war. Wer ihn überginge, führte eine Rücknahme aus, die auf
// einen Stand zeigt, den es nie gab (§8).
func TestEinKaputterVerlaufswegWirdGemeldet(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	if _, err := db.ExecContext(probekontext(t),
		`INSERT INTO history (set_id, karte_id, zeitpunkt, handlung, hinweg, rueckweg)
		 VALUES (?, ?, '2026-08-15', 'feld', '{kein json', '{kein json')`,
		setID, karteID); err != nil {
		t.Fatalf("einfügen: %v", err)
	}

	leser := storage.NeuerVerlaufsleser(db)

	if _, err := leser.Verlauf(probekontext(t), setID); err == nil {
		t.Error("kein Fehler beim Lesen")
	}

	if _, _, err := leser.Zuruecknehmen(probekontext(t), setID); err == nil {
		t.Error("kein Fehler beim Zurücknehmen")
	}
}

// **Ein Weg, dessen Hinweg kaputt ist, fällt beim Wiederherstellen auf.**
func TestEinKaputterHinwegFaelltBeimWiederherstellenAuf(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	if _, err := db.ExecContext(probekontext(t),
		`INSERT INTO history (set_id, karte_id, zeitpunkt, handlung, hinweg, rueckweg, zurueckgenommen)
		 VALUES (?, ?, '2026-08-15', 'feld', '{kein json', '{"feld":"name"}', 1)`,
		setID, karteID); err != nil {
		t.Fatalf("einfügen: %v", err)
	}

	if _, _, err := storage.NeuerVerlaufsleser(db).
		Wiederherstellen(probekontext(t), setID); err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein kaputter Rückweg bei gültigem Hinweg.**
//
// Die beiden Wege werden nacheinander zerlegt; wer nur den ersten kaputt macht,
// betritt den zweiten Fehlerpfad nie.
func TestEinKaputterRueckwegWirdGemeldet(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	if _, err := db.ExecContext(probekontext(t),
		`INSERT INTO history (set_id, karte_id, zeitpunkt, handlung, hinweg, rueckweg)
		 VALUES (?, ?, '2026-08-15', 'feld', '{"feld":"name"}', '{kein json')`,
		setID, karteID); err != nil {
		t.Fatalf("einfügen: %v", err)
	}

	if _, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID); err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein leerer Weg ist kein kaputter.**
//
// Die Tabelle steht seit Phase 2 da; ein Eintrag aus einer Fassung, die noch
// keinen Hinweg kannte, trägt dort eine leere Zeichenkette. Sie ist kein JSON
// und darf trotzdem kein Fehler sein.
func TestEinLeererWegIstKeinFehler(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	if _, err := db.ExecContext(probekontext(t),
		`INSERT INTO history (set_id, karte_id, zeitpunkt, handlung, hinweg, rueckweg)
		 VALUES (?, ?, '2026-08-15', 'feld', '', '')`,
		setID, karteID); err != nil {
		t.Fatalf("einfügen: %v", err)
	}

	verlauf, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if len(verlauf) != 1 {
		t.Errorf("%d Einträge", len(verlauf))
	}
}
