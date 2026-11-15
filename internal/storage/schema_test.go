package storage_test

import (
	"database/sql"
	"strings"
	"testing"
)

// Jede Tabelle ist STRICT.
//
// Ohne STRICT nimmt SQLite in einer TEXT-Spalte auch eine Zahl entgegen und
// gibt sie als Zahl zurück. Ein Feldwert wäre dann je nach Schreibweg ein
// anderer Typ, und zwei gleiche Werte verglichen sich als ungleich — ein
// Fehler, der erst in einer Auswertung auffällt, wo niemand ihn sucht.
func TestJedeTabelleIstStrict(t *testing.T) {
	db := frischeDatenbank(t)

	for _, tabelle := range tabellen(t, db) {
		t.Run(tabelle, func(t *testing.T) {
			if !strings.Contains(strings.ToUpper(schemaVon(t, db, tabelle)), "STRICT") {
				t.Errorf("die Tabelle %s ist nicht STRICT", tabelle)
			}
		})
	}
}

// Jede Löschung kehrt hinter sich.
//
// **Der Wächter steht an der Stelle, an der die Zusage eingelöst wird**
// (AGENTS.md §10): nicht „gibt es eine Fremdschlüsselspalte", sondern „trägt
// jede Spalte, die auf sets oder cards zeigt, auch ON DELETE CASCADE". Ohne das
// bleiben Werte liegen — in der Datei und in jeder Sicherung.
func TestJedeAbhaengigeTabelleKehrtHinterSich(t *testing.T) {
	db := frischeDatenbank(t)

	traeger := map[string]bool{"sets": true, "cards": true, "packages": true}

	for _, tabelle := range tabellen(t, db) {
		schema := schemaVon(t, db, tabelle)

		for eltern := range traeger {
			// `Cut` und nicht `Index`: der zweite kann -1 liefern, und ein
			// Schnitt an -1 wäre ein Absturz in einem Wächter — also genau da,
			// wo niemand ihn erwartet.
			_, dahinter, verweist := strings.Cut(schema, "REFERENCES "+eltern+"(")
			if !verweist {
				continue
			}

			t.Run(tabelle+" → "+eltern, func(t *testing.T) {
				// Nur der Rest der Spaltendefinition zählt, nicht die ganze
				// Tabelle: sonst genügte ein CASCADE irgendwo weiter unten.
				zeile, _, _ := strings.Cut(dahinter, "\n")

				if !strings.Contains(zeile, "ON DELETE CASCADE") {
					t.Errorf("%s verweist auf %s ohne ON DELETE CASCADE — beim Löschen "+
						"bleiben Werte liegen", tabelle, eltern)
				}
			})
		}
	}
}

// Und die Gegenprobe zum Wächter darüber: es wird wirklich gelöscht.
//
// Ein Wächter über das Schema prüft die Absicht. Dass die Absicht wirkt, zeigt
// nur eine echte Löschung — `PRAGMA foreign_keys` muss dafür an sein, und ob
// sie es ist, sagt kein Schema.
func TestEinGeloeschtesSetLaesstNichtsLiegen(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)

	ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		VALUES (1, 'Probe', 'probe', '2026-08-08T10:00:00.000Z', 1)`)
	ausfuehren(t, db, `INSERT INTO cards (id, set_id, reihenfolge) VALUES (1, 1, 0)`)
	ausfuehren(t, db, `INSERT INTO card_values VALUES (1, 'name', 'Erste', 0)`)
	ausfuehren(t, db, `INSERT INTO set_values VALUES (1, 'title', 'Mein Set', 0)`)
	ausfuehren(t, db, `INSERT INTO styling_values VALUES (1, 'foil', 'ja', 0)`)
	ausfuehren(t, db, `INSERT INTO raw_blocks VALUES (1, 'set', 12, 'unbekannt: wert', 0)`)
	ausfuehren(t, db, `INSERT INTO history (set_id, zeitpunkt, handlung, rueckweg)
		VALUES (1, '2026-08-08T10:00:00.000Z', 'angelegt', '')`)

	ausfuehren(t, db, `INSERT INTO set_quellen (set_id, datei, inhalt)
		VALUES (1, 'set', x'6162')`)

	ausfuehren(t, db, `DELETE FROM sets WHERE id = 1`)

	for _, tabelle := range []string{
		"cards", "card_values", "set_values", "styling_values", "raw_blocks", "history",
		"set_quellen",
	} {
		t.Run(tabelle, func(t *testing.T) {
			var uebrig int
			if err := db.QueryRowContext(ctx,
				"SELECT count(*) FROM "+tabelle).Scan(&uebrig); err != nil { //nolint:gosec // fester Tabellenname aus der Liste
				t.Fatalf("%s nicht lesbar: %v", tabelle, err)
			}
			if uebrig != 0 {
				t.Errorf("%d Zeilen in %s liegen nach der Löschung noch da", uebrig, tabelle)
			}
		})
	}
}

// Die Volltextsuche bleibt beim Ändern und Löschen synchron.
//
// Ein Index, der beim Anlegen stimmt und beim Löschen nicht, liefert Treffer
// auf Zeilen, die es nicht mehr gibt — und das fällt erst auf, wenn jemand
// darauf klickt.
func TestDieVolltextsucheBleibtSynchron(t *testing.T) {
	db := frischeDatenbank(t)

	ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		VALUES (1, 'Probe', 'probe', '2026-08-08T10:00:00.000Z', 1)`)
	ausfuehren(t, db, `INSERT INTO cards (id, set_id, reihenfolge) VALUES (1, 1, 0)`)

	ausfuehren(t, db, `INSERT INTO card_values VALUES (1, 'name', 'Feuerball', 0)`)
	if treffer := suchen(t, db, "Feuerball"); treffer != 1 {
		t.Errorf("nach dem Anlegen %d Treffer, erwartet war 1", treffer)
	}

	ausfuehren(t, db, `UPDATE card_values SET wert = 'Eisball' WHERE card_id = 1 AND feld = 'name'`)
	if treffer := suchen(t, db, "Feuerball"); treffer != 0 {
		t.Errorf("nach dem Ändern %d Treffer auf den alten Wert, erwartet waren 0", treffer)
	}
	if treffer := suchen(t, db, "Eisball"); treffer != 1 {
		t.Errorf("nach dem Ändern %d Treffer auf den neuen Wert, erwartet war 1", treffer)
	}

	ausfuehren(t, db, `DELETE FROM card_values WHERE card_id = 1`)
	if treffer := suchen(t, db, "Eisball"); treffer != 0 {
		t.Errorf("nach dem Löschen %d Treffer, erwartet waren 0", treffer)
	}
}

// Der Index trägt die Zeilennummer des Feldwerts — auch nach einem Ändern.
//
// **Gemessen am 2026-09-09 in der Sichtprüfung von U5** (#1416): in einem
// Bestand mit 481.236 Feldwerten dauerte das **zweite** Schreiben desselben
// Feldes 1,2 bis 1,5 Sekunden, das erste 30 ms. Der Auslöser löschte über
// `card_id` und `feld`, und `card_id` ist in der FTS-Tabelle unindiziert —
// jede Änderung las den ganzen Index. Seit `0049_volltext_zeilennummer.sql`
// heißt die Indexzeile wie die Wertzeile, und gelöscht wird über sie.
func TestDerVolltextindexFolgtDerZeilennummer(t *testing.T) {
	db := frischeDatenbank(t)

	ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		VALUES (1, 'Probe', 'probe', '2026-08-08T10:00:00.000Z', 1)`)
	ausfuehren(t, db, `INSERT INTO cards (id, set_id, reihenfolge) VALUES (1, 1, 0)`)
	ausfuehren(t, db, `INSERT INTO card_values VALUES (1, 'name', 'Feuerball', 0)`)
	ausfuehren(t, db, `INSERT INTO card_values VALUES (1, 'text', 'Drei Schaden', 1)`)

	// Das Ändern ist der Fall, der die Nummern auseinanderlaufen ließ: der
	// alte Auslöser löschte die Indexzeile und legte eine **neue** an — mit
	// der nächsten freien Nummer, während die Wertzeile ihre behielt.
	ausfuehren(t, db, `UPDATE card_values SET wert = 'Eisball' WHERE card_id = 1 AND feld = 'name'`)

	var passend, werte int
	if err := db.QueryRowContext(probekontext(t), `
		SELECT (SELECT count(*) FROM card_suche s
		          JOIN card_values v ON v.rowid = s.rowid
		         WHERE v.card_id = s.card_id AND v.feld = s.feld AND v.wert = s.wert),
		       (SELECT count(*) FROM card_values)`).Scan(&passend, &werte); err != nil {
		t.Fatalf("Zeilennummern nicht lesbar: %v", err)
	}

	if passend != werte {
		t.Errorf("%d von %d Indexzeilen tragen die Nummer ihrer Wertzeile", passend, werte)
	}
}

// raw_blocks steht von Anfang an — nicht nachgerüstet.
//
// Wer unbekannte Felder erst wegwirft und sie später mitführen will, hat sie in
// jeder bereits importierten Datenbank verloren, und ein zurückgeschriebenes Set
// öffnet in MSE2 nicht mehr (docs/entscheidungen/0002).
func TestRawBlocksStehtImSchema(t *testing.T) {
	db := frischeDatenbank(t)

	ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		VALUES (1, 'Probe', 'probe', '2026-08-08T10:00:00.000Z', 1)`)
	ausfuehren(t, db, `INSERT INTO raw_blocks VALUES (1, 'set', 12, 'unverstanden: wert', 0)`)

	var inhalt string
	if err := db.QueryRowContext(probekontext(t),
		`SELECT inhalt FROM raw_blocks WHERE set_id = 1`).Scan(&inhalt); err != nil {
		t.Fatalf("raw_blocks nicht lesbar: %v", err)
	}
	if inhalt != "unverstanden: wert" {
		t.Errorf("Inhalt %q", inhalt)
	}
}

// ---------------------------------------------------------------- Hilfsmittel

func tabellen(t *testing.T, db *sql.DB) []string {
	t.Helper()

	// **Ohne die Schattentabellen von FTS5.**
	//
	// Hier stand `name NOT LIKE 'card_suche%'` — der Name **einer**
	// Volltexttabelle. Mit der zweiten (`retrieval_volltext`, A15, #1514)
	// wurde daraus eine Liste, die jemand pflegen muss, und die Frage ist
	// gar keine nach Namen: **eine virtuelle Tabelle und ihre Schatten
	// können kein STRICT tragen** — SQLite legt sie selbst an, und ihre
	// Gestalt gehört dem Modul.
	//
	// Gefragt wird deshalb nach der Eigenschaft: was `CREATE VIRTUAL TABLE`
	// ist, und was zu einer solchen gehört, bleibt draußen. Der Rest muss
	// STRICT sein, auch jede Tafel, die morgen dazukommt.
	zeilen, err := db.QueryContext(probekontext(t),
		`SELECT name FROM sqlite_master WHERE type = 'table'
		   AND name NOT LIKE 'sqlite_%'
		   AND sql NOT LIKE 'CREATE VIRTUAL TABLE%'
		   AND NOT EXISTS (
		     SELECT 1 FROM sqlite_master AS virtuell
		     WHERE virtuell.sql LIKE 'CREATE VIRTUAL TABLE%'
		       AND sqlite_master.name LIKE virtuell.name || '\_%' ESCAPE '\')
		 ORDER BY name`)
	if err != nil {
		t.Fatalf("Tabellen nicht lesbar: %v", err)
	}
	defer func() {
		if fehler := zeilen.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	}()

	var namen []string

	for zeilen.Next() {
		var name string
		if fehler := zeilen.Scan(&name); fehler != nil {
			t.Fatalf("Tabellen nicht lesbar: %v", fehler)
		}

		namen = append(namen, name)
	}

	if fehler := zeilen.Err(); fehler != nil {
		t.Fatalf("Tabellen nicht lesbar: %v", fehler)
	}

	if len(namen) == 0 {
		t.Fatal("keine Tabellen gefunden")
	}

	return namen
}

func schemaVon(t *testing.T, db *sql.DB, tabelle string) string {
	t.Helper()

	var schema string
	if err := db.QueryRowContext(probekontext(t),
		`SELECT sql FROM sqlite_master WHERE name = ?`, tabelle).Scan(&schema); err != nil {
		t.Fatalf("Schema von %s nicht lesbar: %v", tabelle, err)
	}

	return schema
}

func ausfuehren(t *testing.T, db *sql.DB, anweisung string) {
	t.Helper()

	if _, err := db.ExecContext(probekontext(t), anweisung); err != nil {
		t.Fatalf("%s: %v", anweisung, err)
	}
}

func suchen(t *testing.T, db *sql.DB, wort string) int {
	t.Helper()

	var treffer int
	if err := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM card_suche WHERE card_suche MATCH ?`, wort).Scan(&treffer); err != nil {
		t.Fatalf("Suche nach %q: %v", wort, err)
	}

	return treffer
}
