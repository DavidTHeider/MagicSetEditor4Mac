// Die Buchführung des Migrators — was gelaufen ist, und ob es noch stimmt.
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2): `migrator.go`
// sagt, **wie** migriert wird; hier steht, **was verbucht ist** und wie eine
// nachträglich geänderte Migration auffällt. `migrator.go` ging über die
// Zeilengrenze, als die Vorbereitung dazukam
// (#939).

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// angewendeteMitSummen liest Namen und Prüfsummen der verbuchten Migrationen.
//
// Beides in **einer** Abfrage: zwei getrennte hätten zwei Fehlerpfade, und der
// zweite wäre in einem Test nur mit einer Attrappe zu erreichen, die dem
// Aufrufer nachstellt, was SQLite ohnehin tut.
func angewendeteMitSummen(ctx context.Context, db abfrager) (map[string]string, error) {
	return verbucht(ctx, db, `SELECT name, pruefsumme FROM migrationen ORDER BY name`)
}

// Ledger gibt denselben Stand nach außen — Werkstück O6 von Phase 5.2
// (#1430).
//
// # Wozu er außerhalb gebraucht wird
//
// **Weil Verträglichkeit kein Vergleich zweier Zahlen ist.** Ob eine Fassung
// diesen Bestand öffnen kann, entscheidet sich an Name **und** Prüfsumme jeder
// verbuchten Wanderung — eine nachträglich geänderte Migration trägt dieselbe
// Nummer und einen anderen Inhalt. `domain.Vertraeglichkeit` vergleicht das
// gegen das Manifest einer Freigabe; sie braucht dafür den **echten** Ledger
// und nicht die Migrationen im Bündel.
//
// **Nur lesend, und ohne etwas anzulegen.** Wer den Stand erfragt, migriert
// nicht — schon gar nicht, um die Frage beantworten zu können.
func Ledger(ctx context.Context, db *sql.DB) (map[string]string, error) {
	return angewendeteMitSummen(ctx, db)
}

// verbuchteNamen liest, welche Migrationen verbucht sind — **ohne etwas
// anzulegen und ohne die Prüfsummenspalte vorauszusetzen**.
//
// # Wozu es das gibt
//
// Es entscheidet, ob **vor** der Migration gesichert wird, und läuft deshalb
// vor `buchfuehrungAnlegen`. [angewendeteMitSummen] geht das nicht: es setzt
// die Tabelle samt Spalte voraus, und beides herzustellen ist selbst eine
// Schemaänderung — nach der eine Sicherung nicht mehr den Stand vor der
// Migration hält
// (#1170,
// Befund 8 des zweiten Freigabeblicks).
//
// # Keine Tabelle heißt: nichts verbucht
//
// Das ist ein frischer Bestand, und `gibtEsWasZuSichern` sagt dazu nein — es
// gibt nichts zu sichern.
//
// **Gefragt wird das Schema, nicht ein Fehlertext.** Der erste Anlauf verglich
// `err.Error()` gegen „no such table" — ein direkter Verstoß gegen §8, und
// derselbe, den `buchfuehrungAnlegen` daneben schon einmal begangen und
// abgelegt hat.
//
// # `''` statt der Spalte
//
// Damit derselbe Rumpf beide Fälle liest. Eine zweite Schleife wäre ein
// zweiter Satz Fehlerwege für dieselbe Sache — und die kein Lauf beträte.
func verbuchteNamen(ctx context.Context, db abfrager) (map[string]string, error) {
	if !buchfuehrungHat(ctx, db, "name") {
		return map[string]string{}, nil
	}

	return verbucht(ctx, db, `SELECT name, '' FROM migrationen ORDER BY name`)
}

// verbucht liest die Buchführung über eine Abfrage mit zwei Spalten — aus
// der Datenbank oder aus einer laufenden Transaktion ([anwendenAlle] liest
// sie unter ihrer Sperre).
func verbucht(ctx context.Context, db abfrager, abfrage string) (map[string]string, error) {
	zeilen, err := db.QueryContext(ctx, abfrage) //nolint:rowserrcheck // unten geprüft, der Linter sieht es über die Variable nicht
	if err != nil {
		return nil, uebersetze(err, "Buchführung lesen")
	}
	defer zeilen.Close()

	summen := map[string]string{}

	for zeilen.Next() {
		var name, summe string
		if fehler := zeilen.Scan(&name, &summe); fehler != nil {
			return nil, uebersetze(fehler, "Buchführung lesen")
		}

		summen[name] = summe
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "Buchführung lesen")
	}

	return summen, nil
}

// unbekannteAbweisen weist einen Bestand ab, der Migrationen führt, die es
// hier nicht gibt.
//
// Eine verschwundene Migration heißt: diese Datenbank ist mit einer Fassung
// entstanden, die es hier nicht gibt. Weiterzumigrieren hieße, auf einem
// Schema aufzubauen, das man nicht kennt.
//
// # Was die Meldung sagt und was nicht
//
// Sie nennt **alle** unbekannten Migrationen und den Stand, bis zu dem diese
// Fassung reicht.
//
// **Die Versionsnummer der Anwendung, die den Bestand geschrieben hat, kann
// sie nicht nennen.** Eine Binärdatei kennt die Nummer einer Fassung nicht,
// die es bei ihrer Übersetzung noch nicht gab; sie stünde nur dann hier, wenn
// der Bestand sie mitschriebe — und dann wäre sie eine Angabe *aus* der Datei
// und keine Prüfung. Was diese Fassung weiß, ist ihr eigener Schemastand, und
// das ist die Angabe, mit der sich etwas anfangen lässt.
//
// **Und sie ist bestimmt.** Der erste Anlauf lief mit `for name := range`
// über eine Map: fehlten zwei Migrationen, nannte die Meldung eine zufällige,
// bei jedem Start eine andere.
func unbekannteAbweisen(namen []string, bereits map[string]string) error {
	vorhanden := make(map[string]bool, len(namen))
	for _, name := range namen {
		vorhanden[name] = true
	}

	unbekannt := make([]string, 0, len(bereits))

	for name := range bereits {
		if !vorhanden[name] {
			unbekannt = append(unbekannt, name)
		}
	}

	if len(unbekannt) == 0 {
		return nil
	}

	sort.Strings(unbekannt)

	// **Der Stand dieser Fassung ist die letzte Migration, die sie kennt** —
	// nicht ihre Anzahl. `namen` ist sortiert, und eine Lücke darin fängt
	// `luecken` gleich darauf ab.
	stand := "keine einzige Migration"
	if len(namen) > 0 {
		stand = namen[len(namen)-1]
	}

	// **Zwei Fehler in einem** (#940):
	// `ErrMigrationen` sagt, worum es geht — daran hängt jede vorhandene
	// Prüfung —, und [domain.NeuereFassung] trägt die Namen, die die
	// Oberfläche nennen soll.
	return fmt.Errorf("%w: %w", ErrMigrationen,
		&domain.NeuereFassung{Unbekannt: unbekannt, Stand: stand})
}

// luecken weist eine Lücke in der Nummernfolge ab.
//
// `0001`, `0003` ohne `0002` liefe bisher durch. Der Fall entsteht beim
// Zusammenführen zweier Zweige, die je eine Migration hinzugefügt haben — und
// er heißt, dass eine Datenbank je nach Reihenfolge anders aussieht.
func luecken(namen []string) error {
	for i, name := range namen {
		nummer, err := strconv.Atoi(name[:4])
		if err != nil {
			return fmt.Errorf("%w: %q trägt keine lesbare Nummer", ErrMigrationen, name)
		}

		if nummer != i+1 {
			return fmt.Errorf("%w: nach %d folgt %d — eine Lücke oder eine doppelte Nummer",
				ErrMigrationen, i, nummer)
		}
	}

	return nil
}

// summeVergleichen hält eine bereits angewendete Migration gegen ihren Inhalt.
//
// # Reichweite
//
// Der Schutz gilt für Migrationen, die **mit** dieser Prüfung angewendet
// wurden. Für eine Datenbank aus der Zeit davor bleiben die bereits gelaufenen
// ungeprüft — dauerhaft. Der Zweitblick über Phase 2 hat zu Recht darauf
// hingewiesen, dass „gegen nachträgliche Änderung gesichert" ohne diesen Satz
// mehr verspricht, als es hält.
//
// **Eine leere Prüfsumme wird übergangen, nicht nachgetragen.** Sie stammt aus
// einer Datenbank, die vor dieser Prüfung entstanden ist; sie für gefälscht zu
// erklären hieße, jede vorhandene unbrauchbar zu machen.
//
// Nachgetragen wird sie auch nicht — das wäre ein zweiter Schreibweg mit einem
// eigenen Fehlerpfad, und er brächte nichts: eine nachgetragene Summe belegt
// nur, dass der Inhalt **heute** so aussieht, nicht dass er sich nie geändert
// hat. Jede **neue** Migration bekommt ihre Summe beim Anwenden; der Schutz
// wächst also von selbst hinein.
func summeVergleichen(quelle fs.FS, name, verbucht string) error {
	if verbucht == "" {
		return nil
	}

	inhalt, err := fs.ReadFile(quelle, "migrationen/"+name)
	if err != nil {
		return fmt.Errorf("Migration %s lesen: %w", name, err)
	}

	if jetzt := pruefsumme(inhalt); jetzt != verbucht {
		return fmt.Errorf("%w: %s hat sich seit der Anwendung geändert — "+
			"das Schema dieser Datenbank stimmt mit keiner Fassung überein", ErrMigrationen, name)
	}

	return nil
}

// anwenden spielt eine einzelne Migration in einer Transaktion ein.
//
// In einer Transaktion, damit eine Migration, die in der Mitte scheitert, keine
// halbe Datenbank hinterlässt — und damit ihr Eintrag in der Buchführung
// zusammen mit ihrer Wirkung gilt oder gar nicht.
// anwenden führt **eine** Migration in der laufenden Transaktion.
//
// **Sie schließt nicht ab.** Das tut [anwendenAlle], für die ganze Folge —
// an einem eigenen `Commit` hier hinge der halbe Stand (#938).
func anwenden(ctx context.Context, tx *sql.Tx, quelle fs.FS, name string) error {
	inhalt, err := fs.ReadFile(quelle, "migrationen/"+name)
	if err != nil {
		return fmt.Errorf("Migration %s lesen: %w", name, err)
	}

	if _, fehler := tx.ExecContext(ctx, string(inhalt)); fehler != nil {
		return fmt.Errorf("Migration %s anwenden: %w", name, fehler)
	}

	if _, fehler := tx.ExecContext(ctx,
		`INSERT INTO migrationen (name, angewendet, pruefsumme)
		 VALUES (?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'), ?)`,
		name, pruefsumme(inhalt),
	); fehler != nil {
		return fmt.Errorf("Migration %s verbuchen: %w", name, fehler)
	}

	return nil
}
