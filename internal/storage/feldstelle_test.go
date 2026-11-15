package storage_test

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// felderVon liest die Feldnamen einer Karte in ihrer Reihenfolge.
func felderVon(t *testing.T, db *sql.DB, karteID int64) []string {
	t.Helper()

	zeilen, err := db.QueryContext(probekontext(t),
		`SELECT feld FROM card_values WHERE card_id = ? ORDER BY reihenfolge, feld`, karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen ist gelesen

	var felder []string

	for zeilen.Next() {
		var feld string
		if fehler := zeilen.Scan(&feld); fehler != nil {
			t.Fatalf("scannen: %v", fehler)
		}

		felder = append(felder, feld)
	}

	if fehler := zeilen.Err(); fehler != nil {
		t.Fatalf("Zeilen: %v", fehler)
	}

	return felder
}

// dreifeldrigeKarte legt eine Karte mit drei Feldern in fester Reihenfolge an.
func dreifeldrigeKarte(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	return eineKarte(t, db,
		domain.Value{Field: "name", Text: "Feuerball"},
		domain.Value{Field: "text", Text: "Drei Schaden"},
		domain.Value{Field: "farbe", Text: "rot"})
}

// **Eine zurückgenommene Feldentfernung setzt das Feld an seine Stelle zurück.**
//
// # Warum das mehr ist als Schönheit
//
// Seit #159
// führt der Bestand die Quelldatei mit, und ein unverändertes Set kommt Byte
// für Byte zurück (#48). Der Bestand liest `card_values` **nach ihrer
// Reihenfolge**, und `feldSetzen` vergibt sie als `max + 1`.
//
// Wer also ein Feld entfernt und die Änderung zurücknimmt, bekam es zurück —
// **am Ende des Kartenblocks**. Aus einer zurückgenommenen Änderung wurde damit
// eine bleibende, sichtbar erst beim nächsten Speichern
// (#160).
func TestEineZurueckgenommeneEntfernungStelltDieStelleWiederHer(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := dreifeldrigeKarte(t, db)

	vorher := felderVon(t, db, karteID)
	if len(vorher) != 3 || vorher[1] != "text" {
		t.Fatalf("Ausgangslage %v", vorher)
	}

	// **Das mittlere Feld**, nicht das letzte: ein Feld, das ohnehin am Ende
	// stand, käme auch mit `max + 1` an die richtige Stelle zurück — und der
	// Test wäre grün, ohne seinen Fall auszuführen.
	if _, _, _, err := storage.NeuerKartenschreiber(db).
		FeldEntfernen(probekontext(t), karteID, "text"); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	setID := setVon(t, db, karteID)

	if _, gab, err := storage.NeuerVerlaufsleser(db).
		Zuruecknehmen(probekontext(t), setID); err != nil || !gab {
		t.Fatalf("zurücknehmen: %v (gab=%v)", err, gab)
	}

	if nachher := felderVon(t, db, karteID); !gleich(nachher, vorher) {
		t.Errorf("Reihenfolge %v, erwartet %v", nachher, vorher)
	}
}

// **Ein Feld, das es vorher nicht gab, kommt beim Zurücknehmen ans Ende.**
//
// Das ist kein Nachlassen, sondern die richtige Antwort: es hatte keine Stelle,
// und eine zu erfinden hieße, sich eine Ordnung auszudenken, die nie dastand.
func TestEinNeuesFeldKommtBeimZurueknehmenAnSeinEnde(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := dreifeldrigeKarte(t, db)

	if _, _, err := storage.NeuerKartenschreiber(db).
		FeldSetzenMitVorwert(probekontext(t), karteID, "neu", "Wert"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	setID := setVon(t, db, karteID)

	// Zurücknehmen entfernt es, Wiederherstellen setzt es erneut — und zwar
	// wieder ans Ende, wo es hergekommen ist.
	verlauf := storage.NeuerVerlaufsleser(db)

	if _, _, err := verlauf.Zuruecknehmen(probekontext(t), setID); err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	if _, _, err := verlauf.Wiederherstellen(probekontext(t), setID); err != nil {
		t.Fatalf("wiederherstellen: %v", err)
	}

	felder := felderVon(t, db, karteID)
	if len(felder) != 4 || felder[3] != "neu" {
		t.Errorf("Reihenfolge %v, erwartet »neu« am Ende", felder)
	}
}

// **Auch ein geänderter Wert behält seine Stelle.**
//
// `feldSetzen` schreibt über `INSERT … ON CONFLICT`, und die Reihenfolge bleibt
// dabei stehen. Geprüft wird es trotzdem: die Zusage lautet „eine Änderung
// verschiebt nichts", und sie gilt für den häufigsten Weg überhaupt.
func TestEineWertaenderungVerschiebtNichts(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := dreifeldrigeKarte(t, db)

	vorher := felderVon(t, db, karteID)

	if _, _, err := storage.NeuerKartenschreiber(db).
		FeldSetzenMitVorwert(probekontext(t), karteID, "text", "Fünf Schaden"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if nachher := felderVon(t, db, karteID); !gleich(nachher, vorher) {
		t.Errorf("Reihenfolge %v, erwartet %v", nachher, vorher)
	}
}

// setVon sucht das Set einer Karte.
func setVon(t *testing.T, db *sql.DB, karteID int64) int64 {
	t.Helper()

	var setID int64
	if err := db.QueryRowContext(probekontext(t),
		`SELECT set_id FROM cards WHERE id = ?`, karteID).Scan(&setID); err != nil {
		t.Fatalf("Set lesen: %v", err)
	}

	return setID
}

// gleich sagt, ob zwei Namenslisten übereinstimmen.
func gleich(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// **Ein Verlaufseintrag aus einer Sitzung vor #160 trägt keine Stelle.**
//
// Er wird trotzdem gegangen, und das Feld kommt ans Ende — dieselbe Antwort wie
// für ein Feld, das es nie gab. Eine Stelle zu erfinden hieße, sich eine
// Ordnung auszudenken, die nie dastand.
//
// **Kein Migrationsschritt nötig**, und hier steht, warum das trägt: `hinweg`
// und `rueckweg` sind JSON, und ein fehlender Schlüssel kommt als `nil` an.
// Der Test stellt eine solche Zeile her, indem er den Schlüssel wieder
// herausnimmt — näher kommt man einer alten Datenbank nicht, ohne eine
// mitzuliefern.
func TestEinAlterVerlaufseintragOhneStelleWirdTrotzdemGegangen(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := dreifeldrigeKarte(t, db)

	if _, _, _, err := storage.NeuerKartenschreiber(db).
		FeldEntfernen(probekontext(t), karteID, "text"); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	// Die Stelle wieder herausnehmen — so sah die Zeile vor #160 aus.
	if _, err := db.ExecContext(probekontext(t),
		`UPDATE history SET rueckweg = json_remove(rueckweg, '$.reihenfolge')`); err != nil {
		t.Fatalf("umschreiben: %v", err)
	}

	if _, gab, err := storage.NeuerVerlaufsleser(db).
		Zuruecknehmen(probekontext(t), setVon(t, db, karteID)); err != nil || !gab {
		t.Fatalf("zurücknehmen: %v (gab=%v)", err, gab)
	}

	// Da, aber hinten — und **nicht** verschwunden: eine Rücknahme, die an der
	// fehlenden Stelle scheiterte, wäre der schlechtere Ausgang.
	felder := felderVon(t, db, karteID)
	if len(felder) != 3 || felder[2] != "text" {
		t.Errorf("Reihenfolge %v, erwartet »text« am Ende", felder)
	}
}

// **`FeldSetzenAnStelle` setzt ein Feld dorthin, wo es stand** (#160).
//
// Das ist der zweite Teil einer vollständigen Rücknahme: `FeldEntfernen` nennt
// die Stelle, dieser Weg gibt sie zurück. Ohne ihn käme das Feld ans Ende, und
// aus einer zurückgenommenen Änderung würde eine bleibende.
func TestFeldSetzenAnStelleSetztDorthin(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := dreifeldrigeKarte(t, db)

	vorher := felderVon(t, db, karteID)
	schreiber := storage.NeuerKartenschreiber(db)

	alt, hatte, stelle, err := schreiber.FeldEntfernen(probekontext(t), karteID, "text")
	if err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	if !hatte || stelle == nil {
		t.Fatalf("vorher %q, hatte %v, Stelle %v", alt, hatte, stelle)
	}

	if _, _, fehler := schreiber.
		FeldSetzenAnStelle(probekontext(t), karteID, "text", alt, *stelle); fehler != nil {
		t.Fatalf("zurücksetzen: %v", fehler)
	}

	if nachher := felderVon(t, db, karteID); !gleich(nachher, vorher) {
		t.Errorf("Reihenfolge %v, erwartet %v", nachher, vorher)
	}
}
