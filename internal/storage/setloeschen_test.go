// **Ein Set wieder loswerden** —
// #1218.

package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// TestSetLoeschenRaeumtDenGanzenAnhang prüft, dass nichts liegen bleibt.
//
// **Gezählt wird in jeder Tabelle, die auf `sets` zeigt** — nicht nur in
// `sets`. Die Kaskaden stehen im Schema und sind damit eine Zusage des
// Schemas; ob sie greifen, sagt erst eine Zählung. Ein Rest in `card_values`
// wäre unsichtbar: er hängt an einer Karte, die es nicht mehr gibt.
func TestSetLoeschenRaeumtDenGanzenAnhang(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)

	set := probeSet()
	set.Keywords = []*domain.Keyword{{Word: "Fliegend", Match: "Fliegend"}}

	kennung, err := speicher.SetAnlegen(probekontext(t), set,
		[]domain.RawBlock{{File: "set", Line: 12, Text: "unverstanden: wert"}})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Ein zweites Set steht daneben**, und es bleibt. Ohne es prüfte der
	// Test eine Datenbank, die nach dem Löschen ohnehin leer wäre — und ein
	// `DELETE` ohne `WHERE` bestünde ihn.
	fremd := probeSet()
	fremd.Name = "Bleibt stehen"
	fremdkennung, fremdfehler := speicher.SetAnlegen(probekontext(t), fremd, nil)
	if fremdfehler != nil {
		t.Fatalf("das zweite Set anlegen: %v", fremdfehler)
	}

	if loeschfehler := speicher.SetLoeschen(probekontext(t), kennung, "Mein Probeset"); loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	for tabelle, erwartet := range map[string]int{
		"sets": 1, "cards": 2, "card_values": 2, "set_values": 1,
		"raw_blocks": 0, "set_keywords": 0, "card_suche": 2,
	} {
		t.Run(tabelle, func(t *testing.T) {
			if gezaehlt := zaehle(t, db, tabelle); gezaehlt != erwartet {
				t.Errorf("%d Zeilen in %s, erwartet waren %d", gezaehlt, tabelle, erwartet)
			}
		})
	}

	t.Run("das andere Set steht noch", func(t *testing.T) {
		var name string
		if abfrage := db.QueryRowContext(probekontext(t),
			`SELECT name FROM sets WHERE id = ?`, fremdkennung).Scan(&name); abfrage != nil {
			t.Fatalf("das zweite Set lesen: %v", abfrage)
		}
		if name != "Bleibt stehen" {
			t.Errorf("das zweite Set heißt %q", name)
		}
	})
}

// TestSetLoeschenRaeumtDenVolltextindex ist die Zählung, die man vergisst.
//
// `card_suche` hängt an `card_values` über einen Auslöser, nicht über einen
// Fremdschlüssel. Ob SQLite den Auslöser bei einer **kaskadierten** Löschung
// feuert, hängt an `recursive_triggers` — das ist nichts, was man annimmt.
func TestSetLoeschenRaeumtDenVolltextindex(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)

	kennung, err := speicher.SetAnlegen(probekontext(t), probeSet(), nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if vorher := zaehle(t, db, "card_suche"); vorher == 0 {
		t.Fatal("der Index war schon vor dem Löschen leer — die Zählung sagt dann nichts")
	}

	if loeschfehler := speicher.SetLoeschen(probekontext(t), kennung, "Mein Probeset"); loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	if nachher := zaehle(t, db, "card_suche"); nachher != 0 {
		t.Errorf("%d Zeilen stehen noch im Volltextindex", nachher)
	}
}

// TestSetLoeschenLaesstDieSammlungStehen ist die Zusage aus Vermerk 0052.
//
// **Ein Exemplar hängt an einem Druck, nicht an einem Set.** Wer sein Set
// löscht, hat die Karten nicht verkauft — und ein Bestand, der die Sammlung
// mitnimmt, verlöre Kaufpreise, die niemand wiederherstellen kann.
func TestSetLoeschenLaesstDieSammlungStehen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)

	kennung, err := speicher.SetAnlegen(probekontext(t), probeSet(), nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	sammlung := storage.NeuerSammlungsspeicher(db)
	if _, aufnahme := sammlung.Anlegen(probekontext(t), &domain.Exemplar{
		Printing: "xln-1", Menge: 3, Zustand: domain.ZustandNearMint, Sprache: "de",
	}); aufnahme != nil {
		t.Fatalf("ein Exemplar aufnehmen: %v", aufnahme)
	}

	if loeschfehler := speicher.SetLoeschen(probekontext(t), kennung, "Mein Probeset"); loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	if gezaehlt := zaehle(t, db, "sammlung"); gezaehlt != 1 {
		t.Errorf("%d Exemplare in der Sammlung, erwartet war 1", gezaehlt)
	}
}

// TestSetLoeschenVerlangtDenRichtigenNamen ist die Bestätigung.
func TestSetLoeschenVerlangtDenRichtigenNamen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)

	kennung, err := speicher.SetAnlegen(probekontext(t), probeSet(), nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	for name, fall := range map[string]struct {
		name     string
		erwartet error
	}{
		"ein anderer Name":                {"Mein Probeset ", domain.ErrConflict},
		"ein leerer Name":                 {"", domain.ErrConflict},
		"die Groß-/Kleinschreibung zählt": {"mein probeset", domain.ErrConflict},
	} {
		t.Run(name, func(t *testing.T) {
			loeschfehler := speicher.SetLoeschen(probekontext(t), kennung, fall.name)
			if !errors.Is(loeschfehler, fall.erwartet) {
				t.Fatalf("Fehler war %v, erwartet war %v", loeschfehler, fall.erwartet)
			}
			if gezaehlt := zaehle(t, db, "sets"); gezaehlt != 1 {
				t.Errorf("%d Sets, das Set ist trotz falschem Namen weg", gezaehlt)
			}
		})
	}
}

// TestSetLoeschenKenntDasSetNicht meldet ein Set, das es nicht gibt.
func TestSetLoeschenKenntDasSetNicht(t *testing.T) {
	db := frischeDatenbank(t)

	fehler := storage.NeuerSetSpeicher(db).SetLoeschen(probekontext(t), 4711, "Irgendwas")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("Fehler war %v, erwartet war ErrNotFound", fehler)
	}
}

// **Eine gelöschte Kennung kehrt nicht wieder** (#2306). Das Fenster merkt
// sich gelöschte Sets und verwirft eine Finder-Meldung für sie — das trägt
// nur, solange `sets.id` `AUTOINCREMENT` hat. Ohne nähme SQLite nach dem
// Löschen der **größten** Kennung dieselbe wieder, und die Meldung meinte
// ein anderes Set.
func TestEineGeloeschteSetkennungKehrtNichtWieder(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)

	alt, err := speicher.SetAnlegen(probekontext(t), probeSet(), nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if loeschfehler := speicher.SetLoeschen(probekontext(t), alt, "Mein Probeset"); loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	neu, err := speicher.SetAnlegen(probekontext(t), probeSet(), nil)
	if err != nil {
		t.Fatalf("wieder anlegen: %v", err)
	}

	if neu == alt {
		t.Errorf("die gelöschte Kennung %d kam wieder", alt)
	}
}
