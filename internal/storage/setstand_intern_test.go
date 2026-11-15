package storage

// **Ein Änderungszähler je Set** — PERF02
// (#2352):
// jeder Weg, der eine gezeichnete Karte verändern kann, hebt ihn; das
// Aufschlagen nicht, und ein anderes Set nicht.

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// standVon liest den Zähler eines Sets über den Leser.
func standVon(t *testing.T, leser *SetLeser, setID int64) int64 {
	t.Helper()

	stand, err := leser.Setstand(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Setstand %d: %v", setID, err)
	}

	return stand
}

func TestJederWegZurKarteHebtDenSetstand(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	leser := NeuerSetLeser(db)

	for _, befehl := range []string{
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id) VALUES (1, 'Probe', 'magic', '2026-10-15', 1)`,
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id) VALUES (2, 'Andere', 'magic', '2026-10-15', 1)`,
	} {
		if _, err := db.ExecContext(probekontext(t), befehl); err != nil {
			t.Fatal(err)
		}
	}

	if standVon(t, leser, 1) != 0 || standVon(t, leser, 2) != 0 {
		t.Fatal("ein neues Set beginnt nicht bei null")
	}

	for _, schritt := range []struct{ was, befehl string }{
		{"eine Karte kommt dazu", `INSERT INTO cards (id, set_id, reihenfolge) VALUES (10, 1, 0)`},
		{"ein Wert kommt dazu", `INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES (10, 'name', 'Ada', 0)`},
		{"ein Wert ändert sich gleich lang", `UPDATE card_values SET wert = 'Bob' WHERE card_id = 10 AND feld = 'name'`},
		{"ein Wert geht", `DELETE FROM card_values WHERE card_id = 10 AND feld = 'name'`},
		{"die Reihenfolge ändert sich", `UPDATE cards SET reihenfolge = 5 WHERE id = 10`},
		{"ein Wert des Sets kommt dazu", `INSERT INTO set_values (set_id, feld, wert, reihenfolge) VALUES (1, 'title', 'Probe', 0)`},
		{"ein Wert des Sets ändert sich", `UPDATE set_values SET wert = 'Prob2' WHERE set_id = 1`},
		{"ein Wert des Sets geht", `DELETE FROM set_values WHERE set_id = 1`},
		{"ein Schlüsselwort kommt dazu", `INSERT INTO set_keywords (id, set_id, reihenfolge, wort) VALUES (5, 1, 0, 'flug')`},
		{"ein Schlüsselwort ändert sich", `UPDATE set_keywords SET wort = 'flug2' WHERE id = 5`},
		{"ein Schlüsselwort geht", `DELETE FROM set_keywords WHERE id = 5`},
		{"der Stil wechselt", `UPDATE sets SET stil = 'magic-m15' WHERE id = 1`},
		{"die Fassung des Stils wechselt", `UPDATE sets SET stilfassung = '2' WHERE id = 1`},
		{"das Spiel wechselt", `UPDATE sets SET spiel = 'vs' WHERE id = 1`},
		{"die Fassung des Spiels wechselt", `UPDATE sets SET spielfassung = '3' WHERE id = 1`},
		{"der Name wechselt", `UPDATE sets SET name = 'Probe 2' WHERE id = 1`},

		// **Die Stiloptionen und der Ort der Bilder** — Befund 1 des
		// Zweitblicks über Phase 3 (#2352, Wanderung 0100): die
		// Vorschau liest beide, und keiner der vierzehn Auslöser sah sie.
		{"der Ort der Bilder wechselt", `UPDATE sets SET herkunft = '/Sets/Probe.mse-set' WHERE id = 1`},
		{"eine Stiloption kommt dazu", `INSERT INTO styling_values (set_id, feld, wert, reihenfolge) VALUES (1, 'frame', 'gold', 0)`},
		{"eine Stiloption ändert sich", `UPDATE styling_values SET wert = 'silber' WHERE set_id = 1`},
		{"eine Stiloption geht", `DELETE FROM styling_values WHERE set_id = 1`},
		{"die Karte geht", `DELETE FROM cards WHERE id = 10`},
	} {
		vorher := standVon(t, leser, 1)

		if _, err := db.ExecContext(probekontext(t), schritt.befehl); err != nil {
			t.Fatalf("%s: %v", schritt.was, err)
		}

		if standVon(t, leser, 1) <= vorher {
			t.Errorf("%s hob den Setstand nicht", schritt.was)
		}
	}

	if standVon(t, leser, 2) != 0 {
		t.Error("die Änderungen am einen Set hoben den Stand des anderen")
	}

	vorher := standVon(t, leser, 1)

	if _, err := db.ExecContext(probekontext(t), `UPDATE sets SET geoeffnet = '2026-10-16' WHERE id = 1`); err != nil {
		t.Fatal(err)
	}

	if standVon(t, leser, 1) != vorher {
		t.Error("das Aufschlagen hob den Setstand")
	}
}

func TestEinUnbekanntesSetHatKeinenStand(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)

	if _, err := NeuerSetLeser(db).Setstand(probekontext(t), 99); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unbekanntes Set: %v", err)
	}

	if _, err := db.ExecContext(probekontext(t), `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id) VALUES (3, 'Weg', 'magic', '2026-10-15', 1)`); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(probekontext(t), `DELETE FROM sets WHERE id = 3`); err != nil {
		t.Fatal(err)
	}

	var zeilen int
	if err := db.QueryRowContext(probekontext(t), `SELECT count(*) FROM set_aenderungen WHERE set_id = 3`).Scan(&zeilen); err != nil || zeilen != 0 {
		t.Errorf("ein gelöschtes Set hinterließ %d Zeilen (%v)", zeilen, err)
	}
}

// **Ein Set ohne Zähler ist ein Mangel des Bestands** und keine Null — die
// Auslöser legen ihn mit dem Set an; fehlt er, meldet der Leser es, statt
// einen Stand zu erfinden.
func TestEinSetOhneZaehlerHatKeinenStand(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)

	for _, befehl := range []string{
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id) VALUES (4, 'Ohne', 'magic', '2026-10-15', 1)`,
		`DELETE FROM set_aenderungen WHERE set_id = 4`,
	} {
		if _, err := db.ExecContext(probekontext(t), befehl); err != nil {
			t.Fatal(err)
		}
	}

	if stand, err := NeuerSetLeser(db).Setstand(probekontext(t), 4); err == nil {
		t.Errorf("ein Set ohne Zähler hat den Stand %d", stand)
	}
}

// **Wer schreibt, was schon dasteht, hebt nichts** — Befund 14 des
// Zweitblicks über Phase 3 (Wanderung 0100). Der Massenschreiber führt je
// Karte einen Upsert aus, und eine Massenänderung, die nichts änderte,
// machte eine teure Prüfung unnötig veraltet.
func TestUnveraendertesSchreibenHebtDenSetstandNicht(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	leser := NeuerSetLeser(db)

	for _, befehl := range []string{
		`INSERT INTO sets (id, name, spiel, stil, herkunft, geoeffnet, profil_id) VALUES (1, 'Probe', 'magic', 'm15', '/a', '2026-10-15', 1)`,
		`INSERT INTO cards (id, set_id, reihenfolge) VALUES (10, 1, 0)`,
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES (10, 'name', 'Ada', 0)`,
		`INSERT INTO set_values (set_id, feld, wert, reihenfolge) VALUES (1, 'title', 'Probe', 0)`,
		`INSERT INTO styling_values (set_id, feld, wert, reihenfolge) VALUES (1, 'frame', 'gold', 0)`,
		`INSERT INTO set_keywords (id, set_id, reihenfolge, wort) VALUES (5, 1, 0, 'flug')`,
	} {
		if _, err := db.ExecContext(probekontext(t), befehl); err != nil {
			t.Fatal(err)
		}
	}

	vorher := standVon(t, leser, 1)

	for _, schritt := range []struct{ was, befehl string }{
		{"ein Kartenwert, wie er ist", `INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES (10, 'name', 'Ada', 0)
			ON CONFLICT (card_id, feld) DO UPDATE SET wert = excluded.wert`},
		{"ein Wert des Sets, wie er ist", `UPDATE set_values SET wert = 'Probe' WHERE set_id = 1`},
		{"eine Stiloption, wie sie ist", `INSERT INTO styling_values (set_id, feld, wert, reihenfolge) VALUES (1, 'frame', 'gold', 0)
			ON CONFLICT (set_id, feld) DO UPDATE SET wert = excluded.wert`},
		{"ein Schlüsselwort, wie es ist", `UPDATE set_keywords SET wort = 'flug' WHERE id = 5`},
		{"die Karte an ihrem Platz", `UPDATE cards SET reihenfolge = 0 WHERE id = 10`},
		{"Spiel, Stil, Name und Ort, wie sie sind", `UPDATE sets SET spiel = 'magic', stil = 'm15', name = 'Probe', herkunft = '/a' WHERE id = 1`},
	} {
		if _, err := db.ExecContext(probekontext(t), schritt.befehl); err != nil {
			t.Fatalf("%s: %v", schritt.was, err)
		}

		if standVon(t, leser, 1) != vorher {
			t.Errorf("%s hob den Setstand", schritt.was)

			vorher = standVon(t, leser, 1)
		}
	}
}

// **Jeder Auslöser vergleicht jede Spalte seiner Tafel** — sonst fiele eine
// Spalte, die später dazukommt, still aus dem Setstand. Ausgenommen sind nur
// die Spalten von `sets`, die keine Karte anders zeichnen; jede mit Grund.
func TestJederAusloeserVergleichtJedeSpalte(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)

	ohne := map[string]string{
		"sets.id":                "der Schlüssel ändert sich nicht",
		"sets.geoeffnet":         "das Aufschlagen ändert keine Karte",
		"sets.mse_fassung":       "die Fassung von MSE, mit der die Datei geschrieben wurde",
		"sets.versionskontrolle": "ein Vermerk für die Weitergabe, nicht für das Zeichnen",
		"sets.apprentice_code":   "ein Kürzel für Apprentice, nicht für das Zeichnen",
		"sets.profil_id":         "wem es gehört, nicht wie es aussieht",
		"cards.id":               "der Schlüssel ändert sich nicht",
		"set_keywords.id":        "der Schlüssel ändert sich nicht",
	}

	for ausloeser, tafel := range map[string]string{
		"set_aenderungen_set": "sets", "set_aenderungen_karte": "cards", "set_aenderungen_wert": "card_values",
		"set_aenderungen_setwert": "set_values", "set_aenderungen_wort": "set_keywords", "set_aenderungen_stil": "styling_values",
	} {
		var quelltext string
		if err := db.QueryRowContext(probekontext(t),
			`SELECT sql FROM sqlite_master WHERE type = 'trigger' AND name = ?`, ausloeser).Scan(&quelltext); err != nil {
			t.Fatalf("%s: %v", ausloeser, err)
		}

		for _, spalte := range spaltennamen(t, db, tafel) {
			if _, ausgenommen := ohne[tafel+"."+spalte]; !ausgenommen && !vergleichtSpalte(quelltext, spalte) {
				t.Errorf("%s vergleicht %s.%s nicht", ausloeser, tafel, spalte)
			}
		}
	}
}

// vergleichtSpalte sagt, ob ein Auslöser `spalte` vergleicht — **und von ihr
// ausgelöst wird** (Nachlauf über Phase 3, Befund 7): steht vor `ON` eine
// Liste `UPDATE OF …`, feuert er nur für deren Spalten, und ein Vergleich
// ohne Eintrag dort bliebe stumm.
func vergleichtSpalte(quelltext, spalte string) bool {
	if !strings.Contains(quelltext, "OLD."+spalte+" IS NOT NEW."+spalte) {
		return false
	}

	liste := ausloeserliste.FindStringSubmatch(quelltext)
	if liste == nil {
		return true
	}

	for _, eintrag := range strings.Split(liste[1], ",") {
		if strings.TrimSpace(eintrag) == spalte {
			return true
		}
	}

	return false
}

// ausloeserliste liest die Spalten hinter `UPDATE OF` bis `ON`.
var ausloeserliste = regexp.MustCompile(`(?is)\bUPDATE\s+OF\s+(.+?)\s+ON\s`)

// **Der Wächter sieht die Auslöserliste** — eine Spalte, die verglichen wird
// und in `UPDATE OF` fehlt, löst nichts aus.
func TestDerWaechterSiehtDieAusloeserliste(t *testing.T) {
	t.Parallel()

	mitListe := "CREATE TRIGGER x AFTER UPDATE OF spiel, stil ON sets\n" +
		"WHEN OLD.spiel IS NOT NEW.spiel OR OLD.stil IS NOT NEW.stil OR OLD.neu IS NOT NEW.neu BEGIN END"
	ohneListe := "CREATE TRIGGER y AFTER UPDATE ON card_values WHEN OLD.wert IS NOT NEW.wert BEGIN END"

	for name, fall := range map[string]struct {
		quelltext, spalte string
		erwartet          bool
	}{
		"in Liste und Vergleich":     {mitListe, "stil", true},
		"verglichen, nicht in Liste": {mitListe, "neu", false},
		"in keinem":                  {mitListe, "name", false},
		"ohne Liste löst jede aus":   {ohneListe, "wert", true},
	} {
		if vergleichtSpalte(fall.quelltext, fall.spalte) != fall.erwartet {
			t.Errorf("%s: %s nicht %v", name, fall.spalte, fall.erwartet)
		}
	}
}

// spaltennamen gibt die Namen der Spalten einer Tafel.
func spaltennamen(t *testing.T, db *sql.DB, tafel string) []string {
	t.Helper()

	zeilen, err := db.QueryContext(probekontext(t), `SELECT name FROM pragma_table_info(?)`, tafel)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur gelesen

	var spalten []string

	for zeilen.Next() {
		var spalte string
		if err := zeilen.Scan(&spalte); err != nil {
			t.Fatal(err)
		}

		spalten = append(spalten, spalte)
	}

	if err := zeilen.Err(); err != nil {
		t.Fatal(err)
	}

	return spalten
}
