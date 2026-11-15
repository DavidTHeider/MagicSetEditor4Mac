// **Die Sicherungsnaht für die Flächen ohne Fenster** —
// #1169.
//
// # Was hier geprüft wird, und was der Flächenlauf nicht sieht
//
// `internal/flaechen` fährt Kommandozeile, MCP und Dienst als **Prozesse** und
// sieht damit die Wirkung. Was er **nicht** sieht, ist der Lebenszyklus in
// diesem Prozess: dass der Aufrufer nach einem Zurückspielen eine **neue**
// Verbindung bekommt und nicht einen Griff auf eine Datei, die es nicht mehr
// gibt.
//
// **Genau das ist der Zeiger auf den Zeiger**, und ohne diesen Lauf wäre er
// eine Behauptung.
package app_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// offenerBestand gibt eine migrierte Datenbank samt ihrem Pfad.
func offenerBestand(t *testing.T) (db *sql.DB, ort string) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	ort = filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Migriere(probekontext(t), db,
		app.SicherungVorMigration(db, filepath.Dir(ort), app.Fassung)); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	return db, ort
}

// **Sichern, hineinsehen, zurückspielen — und die Verbindung lebt weiter.**
func TestDieSicherungsnahtFaehrtDenGanzenWeg(t *testing.T) {
	db, ort := offenerBestand(t)

	naht := app.BestandsnahtUeber(&db, ort)

	ziel := filepath.Join(t.TempDir(), "stand.msesicherung")

	// **Jede Fläche nennt sich** (OPS01, #2347) — ohne Fläche kein Vermerk.
	if fehler := naht.Sichern(app.MitSicherungswunsch(probekontext(t), domain.FlaecheCLI, "", false), ziel); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	schau, schaufehler := naht.Beschauen(probekontext(t), ziel)
	if schaufehler != nil {
		t.Fatalf("beschauen: %v", schaufehler)
	}

	if schau.AngelegtAm == "" {
		t.Error("die Sicherung nennt keinen Zeitpunkt")
	}

	// **Danach ändert sich der Bestand** — ohne diese Änderung sagte der Lauf
	// darunter nichts: ein Zurückspielen, das nichts verändert, sieht aus wie
	// eines, das wirkt.
	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO settings (schluessel, wert, profil_id)
		 VALUES ('probe', 'nachher', 1)`); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	beiseite, rueckfehler := naht.Zurueckspielen(probekontext(t), ziel)
	if rueckfehler != nil {
		t.Fatalf("zurückspielen: %v", rueckfehler)
	}

	if _, err := os.Stat(beiseite); err != nil {
		t.Errorf("der alte Bestand liegt nicht unter %q: %v", beiseite, err)
	}

	// **Und der Aufrufer hat eine lebende Verbindung** — auf den
	// zurückgespielten Stand, ohne die Einstellung von danach.
	var zahl int
	if err := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM settings WHERE schluessel = 'probe'`).Scan(&zahl); err != nil {
		t.Fatalf("lesen nach dem Tausch: %v", err)
	}

	if zahl != 0 {
		t.Error("die Einstellung von nach der Sicherung steht noch da")
	}
}

// **Eine Sicherung, die es nicht gibt, kommt als Fehler zurück.**
func TestDieSicherungsnahtMeldetEineFehlendeDatei(t *testing.T) {
	db, ort := offenerBestand(t)

	naht := app.BestandsnahtUeber(&db, ort)

	if _, fehler := naht.Beschauen(probekontext(t),
		filepath.Join(t.TempDir(), "gibtsnicht.msesicherung")); fehler == nil {
		t.Error("die fehlende Datei wurde angenommen")
	}

	if _, fehler := naht.Zurueckspielen(probekontext(t),
		filepath.Join(t.TempDir(), "gibtsnicht.msesicherung")); fehler == nil {
		t.Error("die fehlende Datei wurde zurückgespielt")
	}
}

// **Ohne Ort wird weder angehalten noch gesichert** — Nachlauf über Phase 2
// von 0.9.0, Befund 5 (#2323): eine Verbindung, die ihren Bestand nicht mehr
// nennt, gibt dem Wärter keine Datei, nach deren Haltern er fragen könnte.
func TestOhneOrtSichertDieVorsicherungNicht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort := t.TempDir()

	db, err := storage.Oeffnen(probekontext(t), filepath.Join(ort, "bestand.db"))
	if err != nil {
		t.Fatal(err)
	}

	_ = db.Close() //nolint:errcheck // geschlossen gesichert

	if fehler := app.SicherungVorMigration(db, ort, app.Fassung)(probekontext(t)); fehler == nil {
		t.Error("eine geschlossene Verbindung wurde ohne Fehler gesichert")
	}

	eintraege, _ := os.ReadDir(ort) //nolint:errcheck // gezählt
	for _, eintrag := range eintraege {
		if strings.HasPrefix(eintrag.Name(), "bestand-vor-") {
			t.Errorf("gesichert: %s", eintrag.Name())
		}
	}
}
