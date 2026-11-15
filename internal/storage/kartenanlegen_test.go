package storage_test

// **Mehrere Karten in einem Zug** — Werkstück B von Phase 4.11
// (#1337).
//
// # Was nachgelesen wurde (§15)
//
// `AddCardsScript::perform` (`src/data/add_cards_script.cpp`, Zeilen 37–46,
// Fassung f9d9356 vom 2021-07-06) hängt die Karten in **einer** Handlung ans
// Set — `AddCardAction(ADD, set, cards)`. Ein ⌘Z holt sie zusammen zurück.
//
// **Zwölf Grundlande als zwölf Schritte wären zwölfmal ⌘Z für einen
// Handgriff.** §15: Semantik konservativ.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Drei Karten, ein ⌘Z.**
func TestMehrereKartenSindEinVerlaufsschritt(t *testing.T) {
	db := frischeDatenbank(t)
	setID := leeresSet(t, db)

	schreiber := storage.NeuerKartenschreiber(db)

	kennungen, fehler := schreiber.KartenAnlegen(probekontext(t), setID,
		[][]domain.Feldstand{
			{{Feld: "name", Wert: "Plains"}},
			{{Feld: "name", Wert: "Island"}},
			{{Feld: "name", Wert: "Swamp"}},
		})
	if fehler != nil {
		t.Fatalf("KartenAnlegen: %v", fehler)
	}

	if len(kennungen) != 3 {
		t.Fatalf("%d Kennungen, erwartet waren drei", len(kennungen))
	}

	// **Alle drei liegen da, jede mit ihrem Namen.**
	leser := storage.NeuerSetLeser(db)

	karten, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Karten: %v", err)
	}

	if len(karten) != 3 {
		t.Fatalf("%d Karten im Set, erwartet waren drei", len(karten))
	}

	// **Und ein einziges Zurücknehmen räumt alle drei.** Das ist der Kern:
	// ohne die gemeinsame Gruppe bliebe nach dem ersten ⌘Z eine Karte weg und
	// zwei stünden da.
	verlauf := storage.NeuerVerlaufsleser(db)

	if _, ging, fehler := verlauf.Zuruecknehmen(probekontext(t), setID); fehler != nil ||
		!ging {
		t.Fatalf("Zuruecknehmen: ging=%v, %v", ging, fehler)
	}

	danach, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Karten nach der Rücknahme: %v", err)
	}

	if len(danach) != 0 {
		t.Errorf("%d Karten nach einem ⌘Z, erwartet waren null", len(danach))
	}
}

// **Eine einzelne Karte bleibt ein Schritt für sich.**
//
// Die Gegenprobe: ohne sie wäre auch ein Weg grün, der **jede** Anlage in eine
// Gruppe steckte — und dann nähme ein ⌘Z nach zwei einzelnen Anlagen beide
// zurück.
func TestEineEinzelneKarteBleibtEinSchritt(t *testing.T) {
	db := frischeDatenbank(t)
	setID := leeresSet(t, db)

	schreiber := storage.NeuerKartenschreiber(db)

	for range 2 {
		if _, fehler := schreiber.KarteAnlegen(probekontext(t), setID); fehler != nil {
			t.Fatalf("KarteAnlegen: %v", fehler)
		}
	}

	verlauf := storage.NeuerVerlaufsleser(db)

	if _, ging, fehler := verlauf.Zuruecknehmen(probekontext(t), setID); fehler != nil ||
		!ging {
		t.Fatalf("Zuruecknehmen: ging=%v, %v", ging, fehler)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Karten: %v", err)
	}

	if len(karten) != 1 {
		t.Errorf("%d Karten nach einem ⌘Z, erwartet war eine", len(karten))
	}
}

// **Ohne Karten geschieht nichts.**
func TestKartenAnlegenOhneKarten(t *testing.T) {
	db := frischeDatenbank(t)

	kennungen, fehler := storage.NeuerKartenschreiber(db).
		KartenAnlegen(probekontext(t), leeresSet(t, db), nil)
	if fehler != nil {
		t.Fatalf("KartenAnlegen: %v", fehler)
	}

	if len(kennungen) != 0 {
		t.Errorf("%d Kennungen", len(kennungen))
	}
}

// leeresSet legt ein Set ohne Karten an und gibt seine Kennung.
func leeresSet(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probeset", Game: "probe"}, nil)
	if fehler != nil {
		t.Fatalf("Set anlegen: %v", fehler)
	}

	return setID
}

// **Die beiden Fehlerwege des Zuges.**
//
// # Was sie halten, und was die Transaktion hält
//
// Sie treffen die Stellen, an denen ein Ruf **von außen** scheitern kann: eine
// Datenbank, die nicht mehr da ist, und ein Set, das es nicht gibt.
//
// **Die Zusage „ganz oder gar nicht" hält nicht dieser Lauf, sondern die
// Transaktion** — `BeginTx` … `Commit`, und ein `Rollback` im `defer`. Das
// steht hier, weil es sonst so aussähe, als prüfte der Lauf sie: ein Abbruch
// **mitten** in der Reihe ließe sich nur mit einem Feldwert erzwingen, den die
// Ablage abweist, und einen solchen gibt es nicht — `card_values` nimmt jede
// Zeichenkette. Ein Lauf, der das behauptete, hätte einen Zweig erfunden.
func TestKartenAnlegenScheitertGanz(t *testing.T) {
	t.Run("ohne Datenbank", func(t *testing.T) {
		db := frischeDatenbank(t)
		setID := leeresSet(t, db)

		schreiber := storage.NeuerKartenschreiber(db)

		if fehler := db.Close(); fehler != nil {
			t.Fatalf("schließen: %v", fehler)
		}

		if _, fehler := schreiber.KartenAnlegen(probekontext(t), setID,
			[][]domain.Feldstand{{{Feld: "name", Wert: "Plains"}}}); fehler == nil {
			t.Error("die geschlossene Datenbank nahm Karten an")
		}
	})

	t.Run("ein Set, das es nicht gibt", func(t *testing.T) {
		db := frischeDatenbank(t)

		// **Der Fremdschlüssel entscheidet es**, nicht eine Abfrage davor —
		// derselbe Grund wie bei [Kartenaufsicht.KarteAnlegen]: zwischen einem
		// Nachsehen und dem Schreiben läge ein Zeitfenster.
		_, fehler := storage.NeuerKartenschreiber(db).KartenAnlegen(probekontext(t), 4711,
			[][]domain.Feldstand{{{Feld: "name", Wert: "Plains"}}})
		if fehler == nil {
			t.Error("ein Set, das es nicht gibt, nahm Karten an")
		}
	})
}
