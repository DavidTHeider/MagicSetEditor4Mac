// Die Prüfungen über die Migrationshistorie.
//
// Von `fehlerpfade_test.go` abgetrennt, weil die Datei über die Zeilengrenze aus
// AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, was ein
// einzelner Schritt bei einem Fehler tut, hier, was die Historie **als Ganzes**
// zusammenhält.
package storage

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

// Eine bereits angewendete Migration, deren Inhalt sich geändert hat.
//
// Die Buchführung sagt „erledigt", das Schema in der Datenbank stimmt mit keiner
// Fassung überein — und niemand merkt es. Das ist der Fall, für den die
// Prüfsumme da ist.
func TestEineNachtraeglichGeaenderteMigrationWirdAbgewiesen(t *testing.T) {
	db := offeneDatenbank(t)

	vorher := fstest.MapFS{"migrationen/0001_probe.sql": {Data: []byte("SELECT 1")}}
	if err := migriereAus(probekontext(t), db, vorher, nil); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	nachher := fstest.MapFS{"migrationen/0001_probe.sql": {Data: []byte("SELECT 2")}}

	err := migriereAus(probekontext(t), db, nachher, nil)
	if !errors.Is(err, ErrMigrationen) {
		t.Errorf("Fehler %v, erwartet war ErrMigrationen", err)
	}
}

// Ein unveränderter zweiter Lauf geht durch.
func TestEineUnveraenderteMigrationLaeuftDurch(t *testing.T) {
	db := offeneDatenbank(t)
	quelle := fstest.MapFS{"migrationen/0001_probe.sql": {Data: []byte("SELECT 1")}}

	for lauf := 1; lauf <= 2; lauf++ {
		if err := migriereAus(probekontext(t), db, quelle, nil); err != nil {
			t.Fatalf("Lauf %d: %v", lauf, err)
		}
	}
}

// Eine Datenbank ohne Prüfsummen wird übergangen, nicht abgewiesen.
//
// Sie stammt aus der Zeit vor dieser Prüfung. Sie für gefälscht zu erklären
// hieße, jede vorhandene unbrauchbar zu machen — und das ist genau der Fehler,
// vor dem `docs/entscheidungen/0002` warnt.
func TestEineDatenbankOhnePruefsummenLaeuftDurch(t *testing.T) {
	db := offeneDatenbank(t)
	quelle := fstest.MapFS{"migrationen/0001_probe.sql": {Data: []byte("SELECT 1")}}

	if err := migriereAus(probekontext(t), db, quelle, nil); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	// Den Stand von vorher herstellen: Prüfsumme leeren.
	if _, err := db.ExecContext(probekontext(t), `UPDATE migrationen SET pruefsumme = ''`); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := migriereAus(probekontext(t), db, quelle, nil); err != nil {
		t.Fatalf("eine Datenbank ohne Prüfsummen wurde abgewiesen: %v", err)
	}
}

// Eine Lücke in der Nummernfolge wird benannt.
//
// `0001`, `0003` ohne `0002` entsteht beim Zusammenführen zweier Zweige, die je
// eine Migration hinzugefügt haben — und heißt, dass eine Datenbank je nach
// Reihenfolge anders aussieht.
func TestEineLueckeInDerNummernfolgeWirdAbgewiesen(t *testing.T) {
	quelle := fstest.MapFS{
		"migrationen/0001_eins.sql": {Data: []byte("SELECT 1")},
		"migrationen/0003_drei.sql": {Data: []byte("SELECT 3")},
	}

	err := migriereAus(probekontext(t), offeneDatenbank(t), quelle, nil)
	if !errors.Is(err, ErrMigrationen) {
		t.Errorf("Fehler %v, erwartet war ErrMigrationen", err)
	}
}

// Der Zeitstempel der Buchführung unterliegt derselben Form wie angelegt_am.
func TestDerZeitstempelDerBuchfuehrungHatDieRichtigeForm(t *testing.T) {
	db := offeneDatenbank(t)
	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	var angewendet string
	if err := db.QueryRowContext(probekontext(t),
		`SELECT angewendet FROM migrationen ORDER BY name LIMIT 1`).Scan(&angewendet); err != nil {
		t.Fatalf("Buchführung nicht lesbar: %v", err)
	}

	if !zeitform.MatchString(angewendet) {
		t.Errorf("Zeitstempel %q entspricht nicht der Form %s", angewendet, Zeitform)
	}
}

// Eine Migration, deren Inhalt beim zweiten Lauf nicht mehr lesbar ist.
func TestEineUnlesbareMigrationBeimSummenpruefen(t *testing.T) {
	db := offeneDatenbank(t)
	quelle := fstest.MapFS{"migrationen/0001_probe.sql": {Data: []byte("SELECT 1")}}

	if err := migriereAus(probekontext(t), db, quelle, nil); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	// Ein Verzeichnis statt der Datei: ReadDir listet es, ReadFile scheitert.
	kaputt := fstest.MapFS{"migrationen/0001_probe.sql/darin.txt": {Data: []byte("x")}}

	if err := migriereAus(probekontext(t), db, kaputt, nil); err == nil {
		t.Error("eine unlesbare Migration wurde angenommen")
	}
}

func TestBuchfuehrungAnlegenMeldetEineGeschlosseneDatenbank(t *testing.T) {
	if err := buchfuehrungAnlegen(probekontext(t), geschlosseneDatenbank(t)); err == nil {
		t.Error("eine geschlossene Datenbank wurde angenommen")
	}
}

// Eine Buchführungstabelle in fremder Form: das ALTER scheitert an etwas
// anderem als „duplicate column".
func TestBuchfuehrungAnlegenMeldetEineFremdeTabelle(t *testing.T) {
	db := offeneDatenbank(t)

	// Eine Sicht statt einer Tabelle — CREATE TABLE IF NOT EXISTS geht durch,
	// ALTER TABLE nicht.
	if _, err := db.ExecContext(probekontext(t),
		`CREATE VIEW migrationen AS SELECT 1 AS name`); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := buchfuehrungAnlegen(probekontext(t), db); err == nil {
		t.Error("eine Sicht namens migrationen wurde angenommen")
	}
}

// Ein Migrationsname mit vier Zeichen, die keine Zahl sind.
//
// Die Namensprüfung lässt ihn nicht durch — dieser Test hält fest, dass auch
// `luecken` ihn nicht zum Absturz bringt, falls die beiden je auseinanderlaufen.
func TestEineUnlesbareNummerWirdAbgewiesen(t *testing.T) {
	if err := luecken([]string{"abcd_probe.sql"}); !errors.Is(err, ErrMigrationen) {
		t.Errorf("Fehler %v, erwartet war ErrMigrationen", err)
	}
}

// Eine Datenbank, deren Buchführung beim Summenprüfen nicht lesbar ist.
func TestSummenpruefenMeldetEineUnlesbareBuchfuehrung(t *testing.T) {
	db := offeneDatenbank(t)
	quelle := fstest.MapFS{"migrationen/0001_probe.sql": {Data: []byte("SELECT 1")}}

	if err := migriereAus(probekontext(t), db, quelle, nil); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}
	if _, err := db.ExecContext(probekontext(t), `DROP TABLE migrationen`); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}
	// Eine Tabelle ohne die Spalte pruefsumme, aber mit dem Eintrag.
	for _, anweisung := range []string{
		`CREATE TABLE migrationen (name TEXT PRIMARY KEY, angewendet TEXT NOT NULL) STRICT`,
		`INSERT INTO migrationen VALUES ('0001_probe.sql', '2026-08-08T10:00:00.000Z')`,
		`CREATE TABLE sperre (x TEXT)`,
	} {
		if _, err := db.ExecContext(probekontext(t), anweisung); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}
	}

	// Jetzt liegt eine alte Buchführung vor; buchfuehrungAnlegen trägt die
	// Spalte nach, und der Vergleich übergeht die leere Prüfsumme.
	if err := migriereAus(probekontext(t), db, quelle, nil); err != nil {
		t.Errorf("die alte Buchführung wurde abgewiesen: %v", err)
	}
}

// Eine Abfrage über eine Datenbank, die nicht mehr antwortet.
//
// `buchfuehrungHat` gibt dann `false` zurück, ohne abzustürzen. Das ist alles,
// was dieser Test zeigt — und die zweite Hälfte sagt, was er **nicht** zeigt.
func TestSpalteVorhandenAufEinerGeschlossenenDatenbank(t *testing.T) {
	db := geschlosseneDatenbank(t)

	if buchfuehrungHat(probekontext(t), db, "pruefsumme") {
		t.Error("eine geschlossene Datenbank meldete eine vorhandene Spalte")
	}

	// Und `buchfuehrungAnlegen` scheitert, statt still weiterzulaufen.
	//
	// **Aber nicht dort, wo der erste Anlauf es behauptete.** Der stand hier
	// mit der Meldung „er sollte spätestens am ALTER auftauchen" — und das
	// `ALTER` wird nie erreicht: das `CREATE TABLE` davor scheitert schon. Der
	// Nachlauf des Zweitblicks über Phase 2 hat es gefunden.
	//
	// Es ist dieselbe Klasse wie ein `go test -run` ohne Treffer: eine
	// Gegenprobe, die grün ist, weil sie den Fall nicht erreicht. Deshalb prüft
	// die Zusicherung jetzt die Stelle mit.
	err := buchfuehrungAnlegen(probekontext(t), db)
	if err == nil {
		t.Fatal("der Fehler ging verloren")
	}

	if !strings.Contains(err.Error(), "Migrationstabelle anlegen") {
		t.Errorf("der Fehler kam von woanders als erwartet: %v", err)
	}
}

// Dass ein Fehler des `ALTER` gemeldet wird, ist hier **nicht** geprüft.
//
// Es bräuchte eine Datenbank, an der `pragma_table_info` scheitert, während
// `CREATE TABLE IF NOT EXISTS` gelingt. Die gibt es nicht: was das eine
// verhindert, verhindert das andere. Der Kommentar an `buchfuehrungHat`
// behauptete diesen Weg als Absicherung — er ist eine Überlegung, kein Beleg,
// und steht dort jetzt als solche.
