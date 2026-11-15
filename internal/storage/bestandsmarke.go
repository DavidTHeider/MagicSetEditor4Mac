// **Die Marke eines Bestands** —
// #967.
//
// # Die Frage
//
// **Woran erkennt ein Bestand seinen Stand?** `bestand.angelegt_am` taugt
// nicht — es ist über die ganze Lebensdauer konstant und nennt die Herkunft,
// nicht den Inhalt. Zwei Dateien mit demselben `angelegt_am` können sich um
// dreitausend Karten unterscheiden.
//
// # Die Antwort: gezählt, nicht geraten
//
// Über **jede** Tabelle des Bestands `count(*)` und `max(rowid)`, in einer
// Abfrage, und daraus eine Prüfsumme. Zwei Bestände mit derselben Marke haben
// in jeder Tabelle gleich viele Zeilen und dieselbe höchste Zeilennummer.
//
// # Was sie **nicht** sagt
//
// **Dass der Inhalt gleich ist.** Eine reine Änderung an einer bestehenden
// Zeile — ein umbenanntes Set, eine geänderte Einstellung — bewegt weder die
// Zahl noch die höchste Zeilennummer. Für den Fall, um den es geht, trägt sie
// trotzdem: **jede Kartenarbeit schreibt eine Zeile in `history`**, und die
// hebt `max(rowid)`. Wer die Marke für etwas anderes benutzen will, liest
// diesen Absatz zuerst.
//
// **Und sie ist keine Prüfung auf Unversehrtheit.** Eine beschädigte Datei
// kann eine gültige Marke haben; dafür gibt es `PruefeSicherung`.

package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// markenlaenge ist die Länge der Marke in Hexzeichen.
//
// **Zwölf, nicht vierundsechzig.** Sie steht in einem Dateinamen, den ein
// Mensch lesen soll; bei einer Handvoll Sicherungen desselben Bestands ist ein
// Zusammenstoß bei 48 Bit kein Fall, mit dem zu rechnen wäre.
const markenlaenge = 12

// BestandsmarkeVon gibt die Marke eines geöffneten Bestands.
func BestandsmarkeVon(ctx context.Context, db *sql.DB) (string, error) {
	tabellen, err := tabellennamen(ctx, db)
	if err != nil {
		return "", err
	}

	if len(tabellen) == 0 {
		return "", fmt.Errorf(
			"%w: die Datei führt keine einzige Tabelle", domain.ErrSicherung)
	}

	zeilen, err := zaehlungen(ctx, db, tabellen)
	if err != nil {
		return "", err
	}

	summe := sha256.Sum256([]byte(zeilen))

	return hex.EncodeToString(summe[:])[:markenlaenge], nil
}

// Bestandsmarke öffnet eine Datei und gibt ihre Marke.
//
// **Nur lesend und ohne Migration.** Eine Sicherung wird nicht angefasst; sie
// zu migrieren, um ihren Stand zu bestimmen, änderte genau das, was gemessen
// werden soll.
func Bestandsmarke(ctx context.Context, pfad string) (string, error) {
	db, err := nurLesenOeffnen(ctx, pfad)
	if err != nil {
		return "", err
	}

	defer func() { _ = db.Close() }()

	return BestandsmarkeVon(ctx, db)
}

// tabelle ist eine Tabelle des Bestands und ob sie eine Zeilennummer hat.
type tabelle struct {
	name string

	// mitZeilennummer sagt, ob `rowid` da ist.
	//
	// **Nicht jede Tabelle hat eine.** Die Schattentabellen der
	// Volltextsuche — `card_suche_data`, `card_suche_idx` — sind `WITHOUT
	// ROWID`, und ein `max(rowid)` darauf ist ein SQL-Fehler und keine Zahl.
	// Gefunden am 2026-08-25, als der erste Anlauf `rowid` für alle nahm.
	mitZeilennummer bool
}

// tabellennamen gibt die Tabellen des Bestands, ohne die von SQLite selbst.
//
// **`sqlite_master` und nicht die Migrationen.** Hier geht es um das, was
// **dasteht** — eine Sicherung eines älteren Standes führt weniger Tabellen,
// und dann ist ihre Marke zu Recht eine andere.
//
// # Eine Abfrage und nicht drei Fehlerwege
//
// Dieselbe Überlegung wie bei [fehlendeTabellen]: über `Query` und eine
// Schleife gäbe es **drei** Fehlerwege — Abfrage, Auslesen, Abschluss —, und
// zwei davon kann kein Lauf mit einer echten Datenbank betreten. Ein Zweig,
// den kein Aufruf betritt, ist eine Behauptung (§3).
//
// **Was das kostet**: die Zeilen kommen als **eine** Zeichenkette und werden
// hier zerlegt. Eine Tabelle, deren Name einen Zeilenumbruch oder einen
// Tabulator enthält, wird dabei falsch zerlegt. In diesem Schema gibt es sie
// nicht — die Namen stehen in den Migrationen —, und bei einer fremden Datei
// fällt der Fehler auf die sichere Seite: die Marke ist dann eine andere, und
// es entsteht eine zusätzliche Sicherung.
func tabellennamen(ctx context.Context, db *sql.DB) ([]tabelle, error) {
	var roh sql.NullString

	if err := db.QueryRowContext(ctx,
		`SELECT group_concat(zeile, char(10)) FROM (
		   SELECT name || char(9) ||
		          CASE WHEN upper(coalesce(sql, '')) LIKE '%WITHOUT%ROWID%'
		               THEN '0' ELSE '1' END AS zeile
		     FROM sqlite_master
		    WHERE type = 'table' AND name NOT LIKE 'sqlite_%'
		 )`,
	).Scan(&roh); err != nil {
		return nil, domain.MitGrund(domain.ErrSicherung, err,
			"die Tabellen dieser Datei ließen sich nicht lesen — wähle eine andere")
	}

	if !roh.Valid || roh.String == "" {
		return nil, nil
	}

	zeilen := strings.Split(roh.String, "\n")

	// **Sortiert wird hier und nicht im SQL.** `group_concat` sagt über die
	// Reihenfolge seiner Zeilen nichts zu; derselbe Bestand gäbe je nach
	// Laune des Planers verschiedene Marken.
	slices.Sort(zeilen)

	tabellen := make([]tabelle, 0, len(zeilen))

	for _, zeile := range zeilen {
		// **Über das Ende und nicht über einen Fehlerweg.**
		//
		// Ein `strings.Cut` bräuchte einen Zweig für „kein Tabulator darin",
		// und den kann kein Lauf betreten: die Zeile wird zwei Zeilen weiter
		// oben aus Name und Kennzeichen gebaut. Ein Zweig, den kein Aufruf
		// betritt, ist eine Behauptung (§3).
		//
		// **Beide Zweige hier betritt jeder echte Bestand**: `sets` hat eine
		// Zeilennummer, `card_suche_idx` nicht.
		name, mit := strings.CutSuffix(zeile, "\t1")

		tabellen = append(tabellen, tabelle{
			name:            strings.TrimSuffix(name, "\t0"),
			mitZeilennummer: mit,
		})
	}

	return tabellen, nil
}

// zaehlungen gibt je Tabelle eine Zeile `name:anzahl:hoechste`.
//
// # Ein Ausdruck und nicht eine Abfrage je Tabelle
//
// Vierzehn Tabellen gäben vierzehn Umläufe und vierzehn Fehlerwege, von denen
// dreizehn kein Lauf betreten kann. Zusammengesetzt ist es **ein** Ausdruck,
// **eine** Zeile und **ein** Fehlerweg.
//
// **Und nicht über `group_concat`.** Der erste Anlauf nahm es und musste die
// Reihenfolge der Zeilen behaupten — die ist bei einer Aggregatfunktion nicht
// zugesagt, und derselbe Bestand hätte je nach Laune des Planers verschiedene
// Marken gegeben. Aneinandergehängte Teilausdrücke stehen in der Reihenfolge,
// in der sie geschrieben sind.
//
// **Die Reihenfolge der Tabellen steht fest**, weil `tabellennamen` sortiert.
func zaehlungen(ctx context.Context, db *sql.DB, tabellen []tabelle) (string, error) {
	teile := make([]string, 0, len(tabellen))

	for _, t := range tabellen {
		// **Nur Zahlen im SQL, der Name kommt aus Go.**
		//
		// Der erste Anlauf setzte den Namen als `"tabelle"` in den Ausdruck —
		// und SQLite liest doppelte Anführungszeichen zuerst als **Bezeichner**.
		// Bei der Volltextsuche gibt es eine Spalte dieses Namens, über null
		// Zeilen ist sie `NULL`, und `NULL || irgendetwas` ist `NULL`. Die
		// ganze Marke war damit leer, und zwar für **jeden** Bestand.
		hoechste := `'-'`
		if t.mitZeilennummer {
			hoechste = `coalesce(max(rowid), 0)`
		}

		teile = append(teile, fmt.Sprintf(
			`(SELECT count(*) || ':' || %s FROM %q)`, hoechste, t.name))
	}

	var zahlen string
	if err := db.QueryRowContext(ctx,
		"SELECT "+strings.Join(teile, " || char(10) || "),
	).Scan(&zahlen); err != nil {
		return "", domain.MitGrund(domain.ErrSicherung, err,
			"der Stand dieser Datei ließ sich nicht zählen — wähle eine andere")
	}

	// **Die Namen kommen dazu**, damit eine umbenannte oder neue Tabelle die
	// Marke bewegt — und ohne dass die Zahlen dafür zerlegt werden müssten.
	namen := make([]string, 0, len(tabellen))
	for _, t := range tabellen {
		namen = append(namen, t.name)
	}

	return strings.Join(namen, "\n") + "\n" + zahlen, nil
}
