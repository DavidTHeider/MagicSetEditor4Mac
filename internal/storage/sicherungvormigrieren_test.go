// **Die Sicherung hält den Stand vor der Migration — den ganzen** — der zweite
// Zweitblick über die Freigabe 0.3.0, Befund 8
// (#1170).
//
// # Der Befund
//
// `buchfuehrungAnlegen` lief **vor** dem Sicherungsrückruf. Bei einer alten
// Migrationstabelle trägt es per `ALTER TABLE` die Prüfsummenspalte nach — die
// Sicherung enthielt danach also **nicht** den Stand vor jeder Migration, wie
// `README.md` und `docs/ADMINISTRATION.md` zusagen.
//
// Betroffen sind Buchführungsmetadaten und keine Nutzdaten; deswegen war es
// kein Freigabestopp. **Eine Zusage ist es trotzdem**, und eine Zusage, die
// zu 99 % gilt, ist eine, die niemand nachprüfen kann.
//
// # Was hier geprüft wird
//
// Der Rückruf sieht das Schema, das er vorfindet. Seit es nur die Grundlinie
// gibt (Vermerk 0189), kommt eine Buchführung ohne Prüfsumme nur noch in
// einem Bestand vor der Grundlinie vor — und der wird abgewiesen, bevor
// überhaupt etwas geschieht (`TestEinBestandOhnePruefsummenspalteBleibtUnberuehrt`).
// Geprüft wird deshalb der Weg, den eine künftige Wanderung nimmt: über einem
// Bestand auf der Grundlinie sieht die Sicherung ihn **ohne** das, was die
// Wanderung anlegt.
package storage

import (
	"context"
	"database/sql"
	"testing"
)

func TestDieSicherungSiehtDenStandVorDerWanderung(t *testing.T) {
	db := leererBestand(t)
	if fehler := Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	weitere, _ := kuenftigeWanderung(t, "probe", "CREATE TABLE probe (x);")
	gelaufen := false

	fehler := migriereAusMitVorbereitung(probekontext(t), db,
		ueberlagert{heute: migrationen, weitere: weitere}, PruefeBestand,
		func(ctx context.Context) error {
			gelaufen = true

			if tabelleDa(ctx, t, db, "probe") {
				t.Error("die Sicherung sah die Tabelle der Wanderung — sie hält " +
					"nicht den Stand davor")
			}

			return nil
		}, grundlinie)
	if fehler != nil {
		t.Fatalf("wandern: %v", fehler)
	}

	if !gelaufen {
		t.Fatal("der Sicherungsrückruf lief gar nicht — dann sagt dieser Lauf nichts")
	}

	if !tabelleDa(probekontext(t), t, db, "probe") {
		t.Error("die Wanderung lief nach der Sicherung nicht")
	}
}

// tabelleDa sagt, ob es die Tabelle gibt.
func tabelleDa(ctx context.Context, t *testing.T, db *sql.DB, name string) bool {
	t.Helper()

	var anzahl int
	if fehler := db.QueryRowContext(ctx,
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, name).Scan(&anzahl); fehler != nil {
		t.Fatalf("Schema lesen: %v", fehler)
	}

	return anzahl > 0
}

// **Ein frischer Bestand sichert nicht.** Es gibt nichts zu sichern, und eine
// Kopie einer leeren Datenbank wäre eine Datei, die niemand versteht.
func TestEinFrischerBestandSichertNicht(t *testing.T) {
	db := offeneDatenbank(t)
	gelaufen := false

	if fehler := Migriere(probekontext(t), db, func(context.Context) error {
		gelaufen = true

		return nil
	}); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	if gelaufen {
		t.Error("ein frischer Bestand wurde gesichert")
	}
}

// **Eine unlesbare Buchführung hält die Migration an, bevor gesichert wird.**
//
// Der Weg dorthin ist derselbe wie in `fehlerpfade_test.go`: ohne `STRICT` und
// ohne `NOT NULL` lässt SQLite `NULL` in einer TEXT-Spalte zu, und das Scannen
// nach `string` scheitert daran.
func TestEineUnlesbareBuchfuehrungHaeltDieSicherungAn(t *testing.T) {
	db := offeneDatenbank(t)

	for _, anweisung := range []string{
		`CREATE TABLE migrationen (name TEXT PRIMARY KEY, angewendet TEXT)`,
		`INSERT INTO migrationen VALUES (NULL, 'x')`,
	} {
		if _, err := db.ExecContext(probekontext(t), anweisung); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}
	}

	if fehler := Migriere(probekontext(t), db, func(context.Context) error {
		t.Error("es wurde gesichert, obwohl die Buchführung unlesbar ist")

		return nil
	}); fehler == nil {
		t.Fatal("eine unlesbare Buchführung wurde übergangen")
	}
}
