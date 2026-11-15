// Diese Datei führt Buch über die Migrationen und prüft den Bestand selbst.
//
// **Zwei Fragen, eine Datei**: welche Migrationen sind gelaufen, und
// entspricht der Bestand danach seiner eigenen Form? Sie stehen zusammen, weil
// beides Aussagen über den Zustand der Datei sind — `migrator.go` daneben
// beantwortet die andere: wie kommt eine Migration hinein.
//
// Getrennt am 2026-08-22, weil `migrator.go` über die 400 Zeilen aus §2 lief.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
)

// Die Buchführung wird in Go aufgebaut, nicht von einer Migration.
//
// Sie kann es nicht sein: sie entsteht **vor** jeder Migration, weil sie
// verbucht, welche gelaufen sind. Eine Migration, die die Tabelle erweitert,
// die sie selbst eintragen soll, ist ein Kreis.
//
// `ALTER TABLE … ADD COLUMN` deckt den Fall ab, dass eine Datenbank aus der
// Zeit vor der Prüfsumme stammt. Er scheitert, wenn die Spalte schon da ist —
// deshalb wird der Fehler hier gelesen und nicht weitergereicht.
const (
	migrationstabelle = `
CREATE TABLE IF NOT EXISTS migrationen (
  name        TEXT PRIMARY KEY,
  angewendet  TEXT NOT NULL,
  pruefsumme  TEXT NOT NULL DEFAULT ''
) STRICT`

	pruefsummenspalte = `ALTER TABLE migrationen ADD COLUMN pruefsumme TEXT NOT NULL DEFAULT ''`
)

// buchfuehrungAnlegen stellt die Buchführung her, gleich in welchem Stand sie ist.
func buchfuehrungAnlegen(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, migrationstabelle); err != nil {
		return uebersetze(err, "Migrationstabelle anlegen")
	}

	// Eine Datenbank aus der Zeit vor der Prüfsumme bekommt die Spalte
	// nachgetragen.
	//
	// **Gefragt wird das Schema, nicht der Fehlertext** (§8): ändert der
	// Treiber seinen Wortlaut, bräche ein Vergleich mit `"duplicate column"`
	// jeden Start an einer Spalte ab, die längst da ist. `PRAGMA table_info`
	// sagt dieselbe Sache als Tatsache statt als Text.
	if buchfuehrungHat(ctx, db, "pruefsumme") {
		return nil
	}

	if fehler := uebersetze(mustern(ctx, db, pruefsummenspalte),
		"Prüfsummenspalte nachtragen"); fehler != nil {
		return fehler
	}

	return nil
}

// buchfuehrungHat fragt das Schema, ob die Buchführung diese Spalte führt —
// und ob es sie überhaupt gibt: ohne Tabelle hat sie keine.
//
// Ohne Fehlerrückgabe, und das ist Absicht: scheitert die Abfrage, ist die
// Datenbank ohnehin nicht brauchbar. Ein zweiter Fehlerpfad für dieselbe
// Ursache wäre einer, den kein Test je beträte — und damit eine Behauptung
// darüber, was im Ernstfall geschieht.
//
// **Der Fehler wird nicht aufgefangen, er tritt nicht auf.** Wer hierher
// kommt, hat das `CREATE TABLE IF NOT EXISTS` bereits hinter sich; eine
// Datenbank, an der die Abfrage scheitert und jenes gelang, gibt es nicht.
func buchfuehrungHat(ctx context.Context, db abfrager, spalte string) bool {
	zeilen, err := db.QueryContext(ctx,
		`SELECT 1 FROM pragma_table_info('migrationen') WHERE name = ?`, spalte)
	if err != nil {
		return false
	}
	defer zeilen.Close()

	return zeilen.Next() && zeilen.Err() == nil
}

// mustern führt eine Anweisung aus und gibt nur ihren Fehler zurück.
func mustern(ctx context.Context, db *sql.DB, anweisung string, werte ...any) error {
	_, err := db.ExecContext(ctx, anweisung, werte...)

	return err
}

// ErrBestand meldet, dass der Bestand die eigene Prüfung nicht besteht.
var ErrBestand = errors.New("der Bestand entspricht nicht der erwarteten Form")

// zeitform prüft die Schreibweise aus Zeitform — drei Nachkommastellen, Z.
var zeitform = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)

// Zeilenquelle ist, was eine einzelne Zeile lesen kann.
//
// **`*sql.DB` und `*sql.Tx` erfüllen sie beide**, und genau darum gibt es sie:
// seit #938
// läuft die Prüfung **innerhalb** der Migrationstransaktion — dort ist eine
// Abweisung eine Rücknahme statt eines festgeschriebenen Schadens — und beim
// Start weiterhin darüber, wenn gar nicht migriert wurde.
type Zeilenquelle interface {
	QueryRowContext(ctx context.Context, abfrage string, argumente ...any) *sql.Row
}

// PruefeBestand liest den Bestand nach und weist ab, was nicht der Form
// entspricht.
//
// Sie prüft den **Inhalt**, nicht das Schema. Ein CHECK gilt nur für das, was
// durch ihn hindurchgeht; ein Wert kann aus einer Sicherung, einer fremden
// Fassung oder einem Lauf mit `ignore_check_constraints` stammen. Wer sich auf
// das Schema verlässt, prüft die Absicht statt der Tatsache.

func PruefeBestand(ctx context.Context, db Zeilenquelle) error {
	_, err := AngelegtAm(ctx, db)

	return err
}

// AngelegtAm liest den Zeitpunkt der Anlage und prüft dabei seine Form.
//
// **Lesen und Prüfen in einem Zug**, seit [PruefeSicherung] den Wert auch
// **braucht** (#940).
// Zwei getrennte Abfragen hätten dieselbe Zeile zweimal gelesen — und die
// zweite hätte einen Fehlerpfad bekommen, den kein Aufruf betreten kann, weil
// die erste ihn schon abgefangen hat (§3).
func AngelegtAm(ctx context.Context, db Zeilenquelle) (string, error) {
	var wert string

	err := db.QueryRowContext(ctx,
		`SELECT wert FROM bestand WHERE schluessel = 'angelegt_am'`).Scan(&wert)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: angelegt_am fehlt", ErrBestand)
	}

	if err != nil {
		return "", uebersetze(err, "Bestand lesen")
	}

	if !zeitform.MatchString(wert) {
		return "", fmt.Errorf("%w: angelegt_am ist %q, erwartet ist die Form %s",
			ErrBestand, wert, Zeitform)
	}

	return wert, nil
}
