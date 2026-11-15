package storage_test

// **Zwei Arten, einen Bestand kaputtzumachen** — Hinweis 10 des Nachlaufs zum
// Zweitblick über Phase 5.2 (#1435).

import (
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// verbiegen dreht zweihundert Byte in der Mitte der Datei um.
//
// **Die Mitte und nicht der Anfang**: dort stehen Kopf und Schema, und eine
// Datei, die sich gar nicht mehr öffnen lässt, prüfte etwas anderes — nämlich
// `Oeffnen`, nicht die Integritätsabfrage.
//
// **Das ist der schwere Schaden**: SQLite meldet ihn, indem die Abfrage
// scheitert.
// dieSeitengroesse ist, was SQLite seit 3.12 als Vorgabe nimmt.
const dieSeitengroesse = 4096

// seitenanfang gibt den Beginn der **vorletzten** Seite.
//
// **Nicht der letzten**: die ist oft halb leer, und ein Schaden in einem
// ungenutzten Bereich fällt niemandem auf. Nicht der ersten: dort stehen Kopf
// und Schema. Die vorletzte trägt verlässlich Tabellendaten.
func seitenanfang(roh []byte) int {
	seiten := len(roh) / dieSeitengroesse
	if seiten < 3 {
		return dieSeitengroesse
	}

	return (seiten - 2) * dieSeitengroesse
}

func verbiegen(t *testing.T, ort string) {
	t.Helper()

	eingearbeitet(t, ort)

	roh, err := os.ReadFile(ort) // #nosec G304 -- der Pfad kommt aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(roh) < 4096 {
		t.Fatalf("die Datei ist nur %d Byte groß — zu klein zum Verbiegen", len(roh))
	}

	// **Eine ganze Seite und nicht 200 Byte in der Mitte.**
	//
	// Hier standen 200 Byte ab `len(roh)/2`, und das war eine Wette auf die
	// **Dateiaufteilung**: mit Wanderung 0057 (A9, #1448) rückte die Mitte
	// auf eine andere Seite, `integrity_check` gab den Schaden als **Zeile**
	// zurück statt zu scheitern — und der Zweig `klageIstBefund` blieb
	// unbetreten, ohne dass ein Lauf rot wurde. Angestrichen hat es die
	// Abdeckung, nicht ein roter Test; seither sagt der Lauf daneben
	// ausdrücklich, dass er den **schweren** Fall meint.
	//
	// **Die erste Seite bleibt heil**: dort stehen Kopf und Schema, und wer
	// sie mitverbiegt, bekommt den Schaden schon beim **Öffnen**.
	verbogen := seitenanfang(roh)

	for stelle := verbogen; stelle < verbogen+dieSeitengroesse; stelle++ {
		roh[stelle] ^= 0xFF
	}

	// #nosec G703 -- der Pfad kommt aus t.TempDir(), nicht von außen
	if err := os.WriteFile(ort, roh, 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// seitenVerwaisen nimmt einem Index seine Schemazeile und lässt seine Seiten
// stehen.
//
// **Das ist der leichte Schaden**: `integrity_check` gibt ihn als **Zeile**
// zurück („Page 162: never used"), und genau dafür steht das
// `WHERE integrity_check <> 'ok'` in der Abfrage.
//
// **Über eine zweite Verbindung**: das Schema liegt in der ersten im Speicher,
// und dort wäre der Index weiterhin da.
func seitenVerwaisen(t *testing.T, ort string) {
	t.Helper()

	db, fehler := storage.Oeffnen(probekontext(t), ort)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	for _, anweisung := range []string{
		`PRAGMA writable_schema=ON`,
		`DELETE FROM sqlite_schema WHERE name='ballast_y'`,
		`PRAGMA writable_schema=OFF`,
	} {
		if _, err := db.ExecContext(probekontext(t), anweisung); err != nil {
			t.Fatalf("%s: %v", anweisung, err)
		}
	}

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}
}

// eingearbeitet schreibt den Schreibvorlauf in die Datei und gibt sie frei.
func eingearbeitet(t *testing.T, ort string) {
	t.Helper()

	db, fehler := storage.Oeffnen(probekontext(t), ort)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	if _, err := db.ExecContext(probekontext(t),
		`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		t.Fatalf("den Vorlauf einarbeiten: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}
}
