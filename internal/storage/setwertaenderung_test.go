package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Setangabe lässt sich ändern** —
// #567.
//
// # Wofür
//
// `set_values` wurde seit Phase 2 geschrieben und seit Phase 9 gelesen —
// **geändert nie**. Ein Set hat 89 Angaben in `magic.mse-game`: Titel,
// Setcode, Copyright, Symbol; wer den Setcode korrigieren wollte, konnte es
// nicht.
//
// # Warum mit Vorwert und in einer Transaktion
//
// Derselbe Grund wie bei einem Kartenfeld
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)): wer
// erst liest und dann schreibt, kann die zugesagte Rücknahme auf einen Stand
// führen, den es nie gab.
func TestEineSetangabeLaesstSichAendern(t *testing.T) {
	db := frischeDatenbank(t)
	ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		VALUES (1, 'Probe', 'probe', '2026-08-19T10:00:00.000Z', 1)`)

	const setID = 1

	schreiber := storage.NeuerKartenschreiber(db)

	vorher, hatte, err := schreiber.SetwertSetzenMitVorwert(probekontext(t), setID, "set code", "PRB")
	if err != nil {
		t.Fatalf("setzen: %v", err)
	}

	// **„Gab es nicht" ist etwas anderes als „war leer".** Wer beides
	// zusammenwirft, kann die Rücknahme nicht mehr ausführen.
	if hatte {
		t.Errorf("das Feld galt als vorhanden, Vorwert %q", vorher)
	}

	werte, err := storage.NeuerSetLeser(db).Setwerte(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(werte) != 1 || werte[0].Field != "set code" || werte[0].Text != "PRB" {
		t.Fatalf("%+v", werte)
	}

	// **Und beim zweiten Mal steht der Vorwert da.**
	vorher, hatte, err = schreiber.SetwertSetzenMitVorwert(probekontext(t), setID, "set code", "XYZ")
	if err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if !hatte || vorher != "PRB" {
		t.Errorf("Vorwert %q, hatte %v", vorher, hatte)
	}
}

// **Ein Set, das es nicht gibt, ist ein Befund** — nicht eine Zeile, die
// niemandem gehört.
func TestEineSetangabeAnEinemFremdenSetWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)

	_, _, err := storage.NeuerKartenschreiber(db).
		SetwertSetzenMitVorwert(probekontext(t), 4711, "set code", "PRB")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("%v — erwartet war ErrNotFound", err)
	}
}

// **Ein leerer Feldname wird abgewiesen** — auf dem Schreibweg und nicht an
// der Fassade (#89): so gilt die Regel für jeden Schreiber.
func TestEineSetangabeOhneFeldnamenWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)
	ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		VALUES (1, 'Probe', 'probe', '2026-08-19T10:00:00.000Z', 1)`)

	const setID = 1

	_, _, err := storage.NeuerKartenschreiber(db).
		SetwertSetzenMitVorwert(probekontext(t), setID, "   ", "PRB")
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("%v — erwartet war ErrValidation", err)
	}
}

// **Und die Änderung steht im Verlauf** (#52), in derselben Transaktion: ein
// Eintrag, der danach geschrieben wird, fehlt genau dann, wenn jemand
// dazwischen abstürzt — und dann ist die Änderung da und die Rücknahme fort.
func TestEineSetangabeStehtImVerlauf(t *testing.T) {
	db := frischeDatenbank(t)
	ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		VALUES (1, 'Probe', 'probe', '2026-08-19T10:00:00.000Z', 1)`)

	const setID = 1

	schreiber := storage.NeuerKartenschreiber(db)
	if _, _, err := schreiber.SetwertSetzenMitVorwert(
		probekontext(t), setID, "set code", "PRB"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	eintraege, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if len(eintraege) != 1 {
		t.Fatalf("%d Einträge", len(eintraege))
	}

	eintrag := eintraege[0]

	// **Eine eigene Handlung.** Bis hierher stand dort immer `feld`; die
	// Rücknahme muss unterscheiden können, wohin sie schreibt — `card_values`
	// oder `set_values`.
	if eintrag.Handlung != "setwert" {
		t.Errorf("Handlung %q", eintrag.Handlung)
	}

	// **Und keine Kartenkennung**: eine Setangabe gehört keiner Karte.
	if eintrag.KarteID != 0 {
		t.Errorf("KarteID %d", eintrag.KarteID)
	}

	if eintrag.Feld != "set code" || eintrag.Nachher != "PRB" || eintrag.GabVorher {
		t.Errorf("%+v", eintrag)
	}
}

// **⌘Z nimmt eine Setänderung zurück** —
// #567.
//
// # Warum das nicht nebenbei ging
//
// Die Rücknahme schrieb **immer** nach `card_values` (`wegGehen`,
// `internal/storage/verlaufbewegen.go`). Eine Setänderung darüber
// zurückzunehmen hieße, eine Karte anzufassen, die es nicht gibt — die
// Kartenkennung ist null.
//
// **Ohne diesen Weg wäre eine Setänderung die einzige, die ⌘Z nicht kennt**,
// und eine Rücknahme, die manche Änderungen überspringt, ist schlimmer als
// keine: sie führt auf einen Stand, den es nie gab.
func TestEineSetangabeLaesstSichZurueckhemen(t *testing.T) {
	db := frischeDatenbank(t)
	ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		VALUES (1, 'Probe', 'probe', '2026-08-19T10:00:00.000Z', 1)`)

	const setID = 1

	schreiber := storage.NeuerKartenschreiber(db)
	leser := storage.NeuerSetLeser(db)
	verlauf := storage.NeuerVerlaufsleser(db)

	if _, _, err := schreiber.SetwertSetzenMitVorwert(
		probekontext(t), setID, "set code", "PRB"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if _, _, err := schreiber.SetwertSetzenMitVorwert(
		probekontext(t), setID, "set code", "XYZ"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	// **Einmal zurück: `PRB`.**
	if _, da, err := verlauf.Zuruecknehmen(probekontext(t), setID); err != nil || !da {
		t.Fatalf("zurücknehmen: %v / %v", err, da)
	}

	if wert := setwert(t, leser, setID, "set code"); wert != "PRB" {
		t.Errorf("nach dem Zurücknehmen steht %q", wert)
	}

	// **Und noch einmal: die Zeile war vorher nicht da.** „Gab es nicht" ist
	// etwas anderes als „war leer" — eine leere Zeile bliebe beim Speichern
	// als `set code:` in der Datei stehen.
	if _, da, err := verlauf.Zuruecknehmen(probekontext(t), setID); err != nil || !da {
		t.Fatalf("zurücknehmen: %v / %v", err, da)
	}

	werte, err := leser.Setwerte(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(werte) != 0 {
		t.Errorf("die Zeile steht noch: %+v", werte)
	}

	// **Und wieder vor.**
	if _, da, err := verlauf.Wiederherstellen(probekontext(t), setID); err != nil || !da {
		t.Fatalf("wiederherstellen: %v / %v", err, da)
	}

	if wert := setwert(t, leser, setID, "set code"); wert != "PRB" {
		t.Errorf("nach dem Wiederherstellen steht %q", wert)
	}
}

// **Eine zurückgenommene Entfernung kommt an ihre Stelle zurück** (#160) —
// dieselbe Zusage wie bei einem Kartenfeld, und aus demselben Grund: der
// Bestand liest `set_values` nach der Reihenfolge, und die Quelldatei wird
// danach geschrieben.
func TestEineZurueckgenommeneSetangabeKommtAnIhreStelle(t *testing.T) {
	db := frischeDatenbank(t)
	ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		VALUES (1, 'Probe', 'probe', '2026-08-19T10:00:00.000Z', 1)`)
	ausfuehren(t, db, `INSERT INTO set_values VALUES (1, 'title', 'Mein Set', 0)`)
	ausfuehren(t, db, `INSERT INTO set_values VALUES (1, 'set code', 'PRB', 1)`)
	ausfuehren(t, db, `INSERT INTO set_values VALUES (1, 'artist', 'Jemand', 2)`)

	const setID = 1

	if _, _, err := storage.NeuerKartenschreiber(db).SetwertSetzenMitVorwert(
		probekontext(t), setID, "set code", "XYZ"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if _, da, err := storage.NeuerVerlaufsleser(db).
		Zuruecknehmen(probekontext(t), setID); err != nil || !da {
		t.Fatalf("zurücknehmen: %v / %v", err, da)
	}

	werte, err := storage.NeuerSetLeser(db).Setwerte(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	namen := make([]string, 0, len(werte))
	for _, wert := range werte {
		namen = append(namen, wert.Field)
	}

	if len(namen) != 3 || namen[1] != "set code" {
		t.Errorf("die Reihenfolge ist %v", namen)
	}
}

// setwert liest eine einzelne Setangabe.
func setwert(t *testing.T, leser *storage.SetLeser, setID int64, feld string) string {
	t.Helper()

	werte, err := leser.Setwerte(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for _, wert := range werte {
		if wert.Field == feld {
			return wert.Text
		}
	}

	return ""
}

// **Ein alter Verlaufseintrag ohne Stelle wird trotzdem gegangen** — dieselbe
// Zusage wie bei einem Kartenfeld (#160).
//
// `hinweg` und `rueckweg` sind JSON, und ein fehlender Schlüssel kommt als
// `nil` an. Der Test stellt eine solche Zeile her, indem er den Schlüssel
// wieder herausnimmt — näher kommt man einer alten Datenbank nicht, ohne eine
// mitzuliefern.
func TestEinAlterSetverlaufOhneStelleWirdTrotzdemGegangen(t *testing.T) {
	db := frischeDatenbank(t)
	ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		VALUES (1, 'Probe', 'probe', '2026-08-19T10:00:00.000Z', 1)`)
	ausfuehren(t, db, `INSERT INTO set_values VALUES (1, 'title', 'Mein Set', 0)`)
	ausfuehren(t, db, `INSERT INTO set_values VALUES (1, 'set code', 'PRB', 1)`)

	const setID = 1

	if _, _, err := storage.NeuerKartenschreiber(db).SetwertSetzenMitVorwert(
		probekontext(t), setID, "set code", "XYZ"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	// Die Stelle wieder herausnehmen — so sah die Zeile vor #160 aus.
	if _, err := db.ExecContext(probekontext(t),
		`UPDATE history SET rueckweg = json_remove(rueckweg, '$.reihenfolge')`); err != nil {
		t.Fatalf("umschreiben: %v", err)
	}

	if _, gab, err := storage.NeuerVerlaufsleser(db).
		Zuruecknehmen(probekontext(t), setID); err != nil || !gab {
		t.Fatalf("zurücknehmen: %v (gab=%v)", err, gab)
	}

	// **Da, und mit dem alten Wert** — eine Rücknahme, die an der fehlenden
	// Stelle scheiterte, wäre der schlechtere Ausgang.
	if wert := setwert(t, storage.NeuerSetLeser(db), setID, "set code"); wert != "PRB" {
		t.Errorf("nach dem Zurücknehmen steht %q", wert)
	}
}
