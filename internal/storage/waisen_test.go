// **Die Migration räumt verwaiste Zeilen weg** —
// #1220.
//
// # Woher sie kommen
//
// Das Schema trägt `ON DELETE CASCADE` seit `0002_setbestand.sql`. **SQLite
// beachtet Fremdschlüssel aber nur, wenn `foreign_keys` an ist**, und das kam
// erst am 2026-08-15 mit
// #152 dazu.
//
// Der erste Lauf hier ist der **Beleg** dafür: dieselbe Folge aus Anlegen und
// Löschen lässt mit dem Pragma null Zeilen zurück und ohne es eine.
package storage

import (
	"database/sql"
	"strings"
	"testing"
)

// ohneFremdschluessel ist der Bestand von vor dem 2026-08-15 — der, in dem
// Waisen überhaupt entstehen konnten.
//
// **Das `foreign_keys(0)` steht seit dem 2026-09-06 ausdrücklich da**, und der
// Grund ist ein Treiberwechsel (Werkstück 14 von Phase 4.12): `modernc` ließ
// Fremdschlüssel aus, solange niemand sie einschaltete — das ist SQLites
// eigene Voreinstellung. **`ncruces` schaltet sie an.** Gemessen an diesem
// Tag: dieselbe Adresse ohne jedes Pragma gibt dort `PRAGMA foreign_keys = 1`.
//
// Am Produktivweg ändert das nichts, denn `Oeffnen` setzt das Pragma seit
// #152
// ohnehin. Hier ändert es alles: **ohne diese Zeile lässt sich der Schaden
// nicht mehr herstellen, den die Migration wegräumt** — die drei Läufe wären
// grün, weil sie nichts mehr prüfen.
const ohneFremdschluessel = "file::memory:?_pragma=busy_timeout(5000)" +
	"&_pragma=foreign_keys(0)"

// waisen zählt die Kartenwerte ohne Karte.
func waisen(t *testing.T, db *sql.DB) int {
	t.Helper()

	var zahl int
	if err := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM card_values WHERE card_id NOT IN (SELECT id FROM cards)`).
		Scan(&zahl); err != nil {
		t.Fatalf("zählen: %v", err)
	}

	return zahl
}

// karteMitWertUndWeg legt Set, Karte und Wert an und löscht die Karte wieder.
func karteMitWertUndWeg(t *testing.T, db *sql.DB) {
	t.Helper()

	for _, anweisung := range []string{
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id) VALUES (1,'S','probe','2026-01-01', 1)`,
		`INSERT INTO cards (id, set_id, reihenfolge) VALUES (1,1,0)`,
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES (1,'name','Probekarte',0)`,
		`DELETE FROM cards WHERE id = 1`,
	} {
		if _, err := db.ExecContext(probekontext(t), anweisung); err != nil {
			t.Fatalf("%s: %v", anweisung, err)
		}
	}
}

// **Der Beleg: ohne `foreign_keys` bleibt die Waise, mit dem Pragma nicht.**
func TestOhneFremdschluesselBleibtEineWaiseZurueck(t *testing.T) {
	mitPragma, err := Oeffnen(probekontext(t), ImArbeitsspeicher)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = mitPragma.Close() }) //nolint:errcheck // Test

	if fehler := Migriere(probekontext(t), mitPragma, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	karteMitWertUndWeg(t, mitPragma)

	if zahl := waisen(t, mitPragma); zahl != 0 {
		t.Errorf("mit foreign_keys blieben %d Waisen", zahl)
	}

	// **Und ohne das Pragma** — der Stand vor dem 2026-08-15.
	ohnePragma, err := FuerLaeufe(ohneFremdschluessel)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = ohnePragma.Close() }) //nolint:errcheck // Test

	if fehler := Migriere(probekontext(t), ohnePragma, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	karteMitWertUndWeg(t, ohnePragma)

	if zahl := waisen(t, ohnePragma); zahl != 1 {
		t.Errorf("ohne foreign_keys blieben %d Waisen, erwartet 1", zahl)
	}
}

// **Der Zähler nennt jede der drei Stellen** —
// #1220.
//
// Er ist der Wächter darüber, dass der Bestand sauber **bleibt**: die
// Migration räumt einmal weg, und ohne diese Zeile merkte niemand, wenn wieder
// welche entstünden.
func TestDerWaisenzaehlerNenntJedeStelle(t *testing.T) {
	db, err := FuerLaeufe(ohneFremdschluessel)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Test

	if fehler := Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	karteMitWertUndWeg(t, db)

	for _, anweisung := range []string{
		`INSERT INTO set_values (set_id, feld, wert, reihenfolge) VALUES (99,'titel','X',0)`,
	} {
		if _, fehler := db.ExecContext(probekontext(t), anweisung); fehler != nil {
			t.Fatalf("%s: %v", anweisung, fehler)
		}
	}

	zahlen, err := Waisen(probekontext(t), db)
	if err != nil {
		t.Fatalf("zählen: %v", err)
	}

	if zahlen.Kartenwerte != 1 || zahlen.Setangaben != 1 {
		t.Errorf("Zahlen %+v, erwartet je 1 bei Kartenwerten und Setangaben", zahlen)
	}

	if text := WaisenText(zahlen); !strings.Contains(text, "Kartenwerte 1") {
		t.Errorf("der Text lautet %q", text)
	}
}

// **Und ohne Waisen sagt er das.**
func TestOhneWaisenSagtDerZaehlerKeine(t *testing.T) {
	if text := WaisenText(Waisenzahlen{}); text != "keine" {
		t.Errorf("der Text lautet %q, erwartet »keine«", text)
	}
}

// **Ein Fehler beim Zählen kommt durch.**
func TestEinFehlerBeimZaehlenKommtDurch(t *testing.T) {
	db := offeneDatenbank(t)

	if _, err := Waisen(probekontext(t), db); err == nil {
		t.Error("ohne Schema wurde gezählt")
	}
}
