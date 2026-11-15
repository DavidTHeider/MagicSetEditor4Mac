package storage

// **Die Grundlinie** — ein leerer Bestand in einem Zug statt Wanderung für
// Wanderung (J02, #2266).
//
// # Warum
//
// Jeder Start eines neuen Bestands und fast jeder Test fuhr alle Wanderungen
// einzeln — 97 am 2026-10-12, darunter Umbauten ganzer Tabellen. SQLite läuft
// hier als WebAssembly, und unter dem Race-Detektor kostete ein frischer
// Bestand 2,28 s statt 0,23 s über die Grundlinie
// (`frischerbestand_bench_test.go`).
//
// # Was gilt
//
// **Es gibt nur sie** — entschieden am 2026-10-12, noch vor 1.0.0
// ([0189](../../docs/entscheidungen/0189-nur-die-grundlinie.md)). Die 97
// Wanderungen, die sie enthält, liegen nicht mehr als Datei vor; ihr Umfang
// steht mit Name und Prüfsumme in `grundlinie/umfang.txt`, und ein Bestand
// führt sie in seiner Buchführung wie bisher.
//
//   - **Ein neuer Bestand** entsteht über sie in einem Zug.
//   - **Ein Bestand aus 0.7.0** führt genau diese 97 und öffnet sich ohne
//     Änderung.
//   - **Ein älterer** wird abgewiesen, bevor irgendetwas an ihm geschieht
//     ([domain.AeltereFassung]): einmal mit 0.7.0 geöffnet, steigt er dort auf.
//   - **Was nach [grundlinieBis] kommt**, ist wieder eine Wanderung als Datei
//     unter `migrationen/` und läuft wie bisher.
//
// **Dass sie denselben Bestand ergab wie die Wanderungen**, hielt bis zu
// deren Wegfall `TestDieGrundlinieGleichtDenWanderungen` (Commit
// `ebd92fa37`): Schema, Startzeilen und Folgezähler, verglichen Zeichen für
// Zeichen. Seither hält `TestDieGrundlinieIstUnveraendert` sie fest.

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

//go:embed grundlinie
var grundlinie embed.FS

// grundlinieBis ist die letzte Wanderung, die die Grundlinie enthält.
const grundlinieBis = "0097_faktorhaushalt.sql"

// bestandIstLeer sagt, ob außer der Buchführung noch nichts dasteht.
//
// **`sqlite_` als Präfix und nicht als Muster**: in `LIKE` steht `_` für ein
// beliebiges Zeichen, und eine Tabelle `sqliteX` galt als eine von SQLite
// selbst (Zweitblick über die Folgen vor 0.8.0, Befund 4).
func bestandIstLeer(ctx context.Context, quelle Zeilenquelle) (bool, error) {
	var anzahl int
	if fehler := quelle.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master
		WHERE substr(name, 1, 7) <> 'sqlite_' AND name <> 'migrationen'`).Scan(&anzahl); fehler != nil {
		return false, uebersetze(fehler, "Bestand prüfen")
	}

	return anzahl == 0, nil
}

// grundlinienumfang liest, welche Wanderungen die Grundlinie enthält — in
// ihrer Reihenfolge und mit ihrer Prüfsumme (`grundlinie/umfang.txt`).
func grundlinienumfang(basis fs.FS) (namen []string, summen map[string]string, err error) {
	inhalt, fehler := fs.ReadFile(basis, "grundlinie/umfang.txt")
	if fehler != nil {
		return nil, nil, fmt.Errorf("%w: der Umfang der Grundlinie: %w", ErrMigrationen, fehler)
	}

	summen = map[string]string{}

	for _, zeile := range strings.Split(string(inhalt), "\n") {
		if zeile == "" || strings.HasPrefix(zeile, "#") {
			continue
		}

		name, summe, ganz := strings.Cut(zeile, " ")
		if !ganz || !migrationsname.MatchString(name) {
			return nil, nil, fmt.Errorf("%w: der Umfang der Grundlinie nennt %q", ErrMigrationen, zeile)
		}

		namen = append(namen, name)
		summen[name] = summe
	}

	return namen, summen, nil
}

// grundlinieWennLeer legt einen Bestand ohne Buchführung über die Grundlinie
// an und gibt die verbuchten Wanderungen zurück.
//
// **Steht schon etwas da, ist es kein Bestand dieser Anwendung**: jeder, den
// sie je angelegt hat, führt seine Wanderungen. Bis 0.7.0 scheiterte ein
// solcher an der ersten Wanderung, die eine seiner Tabellen noch einmal
// anlegen wollte; ohne diese Abweisung bliebe er liegen, und erst der erste
// fachliche Aufruf fände seine Tabellen nicht.
func grundlinieWennLeer(
	ctx context.Context, tx *sql.Tx, basis fs.FS, namen []string, summen map[string]string,
) (map[string]string, error) {
	leer, fehler := bestandIstLeer(ctx, tx)
	if fehler != nil {
		return nil, fehler
	}

	if !leer {
		return nil, ohneBuchfuehrung()
	}

	return grundlinieAnwenden(ctx, tx, basis, namen, summen)
}

// grundlinieAnwenden legt den Bestand in einem Zug an — Tabellen, Saat, dann
// Indizes, Sichten, Auslöser und Folgezähler — und verbucht jede enthaltene
// Wanderung mit der Prüfsumme aus ihrem Umfang: dieselbe Buchführung, die
// 0.7.0 hinterlassen hat.
func grundlinieAnwenden(
	ctx context.Context, tx *sql.Tx, basis fs.FS, namen []string, summen map[string]string,
) (map[string]string, error) {
	tabellen, _ := fs.Glob(basis, "grundlinie/tabellen-*.sql") //nolint:errcheck // ein fester Ausdruck
	rest, _ := fs.Glob(basis, "grundlinie/rest-*.sql")         //nolint:errcheck // dito
	sort.Strings(tabellen)
	sort.Strings(rest)

	for _, datei := range append(append(tabellen, "grundlinie/saat.sql"), rest...) {
		inhalt, fehler := fs.ReadFile(basis, datei)
		if fehler != nil {
			return nil, fmt.Errorf("Grundlinie %s lesen: %w", datei, fehler)
		}

		if _, fehler := tx.ExecContext(ctx, string(inhalt)); fehler != nil {
			return nil, fmt.Errorf("Grundlinie %s anwenden: %w", datei, fehler)
		}
	}

	verbucht := make(map[string]string, len(namen))

	for _, name := range namen {
		if _, fehler := tx.ExecContext(ctx,
			`INSERT INTO migrationen (name, angewendet, pruefsumme)
			 VALUES (?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'), ?)`, name, summen[name]); fehler != nil {
			return nil, fmt.Errorf("Migration %s verbuchen: %w", name, fehler)
		}

		verbucht[name] = summen[name]
	}

	return verbucht, nil
}

// grundlinieGefuehrt prüft einen vorhandenen Bestand gegen die Grundlinie:
// **jede** ihrer Wanderungen muss er führen, und mit ihrer Prüfsumme. Fehlt
// eine, stammt er aus einer Fassung vor 0.7.0 ([domain.AeltereFassung]);
// weicht eine Summe ab, ist sein Schema keines, das es je gab.
func grundlinieGefuehrt(namen []string, summen, bereits map[string]string) error {
	if len(bereits) == 0 {
		return ohneBuchfuehrung()
	}

	var fehlend []string

	for _, name := range namen {
		verbucht, da := bereits[name]

		switch {
		case !da:
			fehlend = append(fehlend, name)
		case verbucht != "" && verbucht != summen[name]:
			return fmt.Errorf("%w: %s hat sich seit der Anwendung geändert — "+
				"das Schema dieser Datenbank stimmt mit keiner Fassung überein", ErrMigrationen, name)
		}
	}

	if len(fehlend) == 0 {
		return nil
	}

	return fmt.Errorf("%w: %w", ErrMigrationen,
		&domain.AeltereFassung{Fehlend: fehlend, Grundlinie: grundlinieBis})
}

// ohneBuchfuehrung weist Tabellen ab, die keine einzige Wanderung verbucht
// haben — sie stammen aus **keiner** Fassung, auch aus keiner älteren.
func ohneBuchfuehrung() error {
	return fmt.Errorf("%w: hier stehen Tabellen, aber keine einzige verbuchte "+
		"Wanderung — das ist kein Bestand dieser Anwendung", ErrMigrationen)
}

// vorblick liest Buchführung und Leere vor der Sperre — **in einem Stand**,
// einer lesenden Transaktion.
//
// Zwei einzelne Blicke reichten nicht: legte ein zweiter Prozess denselben
// neuen Bestand zwischen ihnen an, sah der erste die Buchführung leer und
// danach die Tabellen, und der frische Bestand galt als fremd
// (`grundlinienvorblick_intern_test.go`). In einem Stand ist er entweder
// leer oder geführt; was ein anderer inzwischen geschrieben hat, liest
// [anwendenAlle] unter der Sperre.
func vorblick(ctx context.Context, db *sql.DB) (verbucht map[string]string, leer bool, err error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, false, uebersetze(err, "Bestand lesen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nur gelesen

	if verbucht, err = verbuchteNamen(ctx, tx); err != nil {
		return nil, false, err
	}

	if zwischenDenBlicken != nil {
		zwischenDenBlicken()
	}

	leer, err = bestandIstLeer(ctx, tx)

	return verbucht, leer, err
}

// zwischenDenBlicken stellt in einem Lauf den Augenblick zwischen den beiden
// Blicken des [vorblick] — im Betrieb nil.
var zwischenDenBlicken func()

// vorDemFestschreiben stellt den Augenblick nach den Wanderungen und vor dem
// Festschreiben — im Betrieb nil, im Abnahmebau über
// `MSE_ABNAHME_WANDERUNGSHALT` ein Halt (`wanderungshalt_abnahme.go`, M02).
var vorDemFestschreiben func()

// grundlinieVorDerSperre weist ab, was vor jeder Schemaänderung abzuweisen
// ist: einen Bestand vor der Grundlinie (an den Namen) und Tabellen ohne
// Buchführung.
func grundlinieVorDerSperre(umfang []string, summen, verbucht map[string]string, leer bool) error {
	switch {
	case len(verbucht) > 0:
		return grundlinieGefuehrt(umfang, summen, verbucht)
	case !leer:
		return ohneBuchfuehrung()
	default:
		return nil
	}
}
