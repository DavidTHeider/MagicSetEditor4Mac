// **Eine Sicherung wird gelesen, bevor sie den Bestand ersetzt** (#940).
//
// Der Grund für eine eigene Datei ist die Reihenfolge: `migrator.go` prüft
// dieselben Dinge **während** einer Migration, also an einem Bestand, der
// schon der eigene ist. Hier ist er es noch nicht — und genau deshalb darf
// nichts geschrieben werden, was die fremde Datei verändert.

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sicherungsbefund ist das, was eine Sicherung über sich hergibt.
//
// **Damit jemand entscheiden kann, nicht damit etwas geprüft wäre.** Die
// Prüfung steckt im Fehler von [PruefeSicherung]; die Angaben hier beantworten
// die Frage „ist das die Datei, die ich meine?".
type Sicherungsbefund struct {
	// AngelegtAm ist der Zeitpunkt, zu dem der Bestand entstanden ist —
	// **nicht** der Zeitpunkt der Sicherung. Er stammt aus `bestand`.
	AngelegtAm string

	// Schemastand ist die letzte Migration, die in der Datei verbucht ist.
	Schemastand string

	// Sets und Karten sagen, wie viel darin liegt.
	Sets   int
	Karten int
}

// PruefeSicherung liest eine Sicherungsdatei und weist ab, was nicht übernommen
// werden darf.
//
// # Drei Abweisungen, und sie sind nicht dieselbe
//
//  1. **Keine lesbare Datenbank** — die Datei fehlt, ist leer oder etwas
//     anderes. Ein gewöhnlicher Fehler.
//  2. **Kein Bestand dieser Anwendung** ([ErrBestand]) — `angelegt_am` fehlt
//     oder hat nicht die Form. Eine fremde SQLite-Datei kommt hier heraus.
//  3. **Aus einer neueren Fassung** ([ErrMigrationen]) — die Datei führt
//     Migrationen, die diese Fassung nicht kennt. Sie wird **benannt**
//     abgewiesen, mit dem Namen der unbekannten Migration und dem eigenen
//     Stand.
//
// # Was sie nicht prüft
//
// **Ob die Daten darin fachlich stimmen.** Ein Bestand mit widersprüchlichen
// Karten geht hier durch; die Prüfung gilt der Form und der Fassung, nicht dem
// Inhalt. Und sie sagt nichts darüber, ob die Datei **älter** ist als der
// laufende Bestand — das ist erlaubt und der Normalfall beim Zurückspielen.
//
// **Und sie migriert nicht.** Eine ältere Sicherung wird angenommen; migriert
// wird sie erst, wenn sie der eigene Bestand ist — dort, wo eine Sicherung
// davor entsteht.
func PruefeSicherung(ctx context.Context, pfad string) (Sicherungsbefund, error) {
	db, err := nurLesenOeffnen(ctx, pfad)
	if err != nil {
		return Sicherungsbefund{}, err
	}

	defer func() { _ = db.Close() }()

	// **Vor dem Zurückspielen jede Seite** — F05 (#2344): sonst ersetzte eine
	// Sicherung mit zerschriebener Datenseite einen heilen Bestand.
	if err := schnellPruefen(ctx, db); err != nil {
		return Sicherungsbefund{}, domain.MitGrund(domain.ErrSicherung, err,
			"%q ist beschädigt — wähle eine andere Sicherung", filepath.Base(pfad))
	}

	return lesenUndPruefen(ctx, db, migrationen, grundlinie, pfad)
}

// nurLesenOeffnen öffnet eine fremde Bestandsdatei, ohne sie zu verändern.
//
// **`mode=ro` und nicht [Oeffnen].** Jenes setzt `journal_mode = WAL` und
// **schreibt damit in die fremde Datei** — eine Prüfung, die ihren Gegenstand
// verändert, ist keine. Nur lesen heißt außerdem: eine Datei ohne Schreibrecht
// lässt sich prüfen.
//
// # Was sie trotzdem hinterlässt — nachgemessen am 2026-09-07
//
// **Ein `-shm` und ein `-wal` neben der Datei.** Die **Datenbankdatei** bleibt
// unangefasst, der **Ordner** nicht: SQLite braucht den geteilten Speicher, um
// einen WAL zu lesen, und
// legt ihn auch im Nur-Lese-Modus an, solange das Verzeichnis beschreibbar
// ist. Gemessen: ein Ordner mit `sicherung.db` trägt danach `sicherung.db`,
// `sicherung.db-shm` und `sicherung.db-wal`.
//
// **Nach dem Wegräumen stehen deshalb eine Sicherung und ihr `-shm` da** —
// zwei Dateien, aber ein Rückweg (Werkstück 5 von Phase 4.13,
// #1352).
// Der Begleiter ist **immer** der jüngste, denn er entsteht in dem
// Augenblick, in dem gefragt wird.
//
// **Genannt und nicht behoben.** Verhindern ließe es sich mit `immutable=1` —
// und das wäre eine Zusage, die bei einer Sicherung, an der jemand anderes
// gerade arbeitet, falsch ist. Wer die Begleiter unterscheiden muss, tut es am
// Namen: `app.ohneBegleiter`.
//
// **Eine Stelle für zwei Leser.** `PruefeSicherung` und [Bestandsmarke]
// öffnen dieselbe Art Datei zum selben Zweck; zwei Fassungen davon liefen mit
// der Zeit auseinander, und die eine schriebe dann doch.
func nurLesenOeffnen(ctx context.Context, pfad string) (*sql.DB, error) {
	if _, err := os.Stat(pfad); err != nil {
		return nil, domain.MitGrund(domain.ErrSicherung, err,
			"%q ließ sich nicht lesen — liegt die Datei noch dort, wo sie war?",
			filepath.Base(pfad))
	}

	// **Ohne Schlüssel, und das ist eine Entscheidung** (Vermerk 0072): eine
	// Sicherung trägt den Bestand im **Klartext**. Eine, die man nur auf dem
	// Rechner aufmachen kann, auf dem sie entstand, ist keine Sicherung —
	// `docs/ADMINISTRATION.md` sagt „eine einzelne Datei, aus der heraus die
	// Anwendung wieder aufsetzen kann", und das gilt auch nach einem
	// Rechnerwechsel.
	db, err := oeffner(AlsURI(pfad)+"?mode=ro&_pragma=busy_timeout(5000)", "")
	if err != nil {
		return nil, domain.MitGrund(domain.ErrSicherung, err,
			"%q ließ sich nicht öffnen — wähle eine andere Sicherung",
			filepath.Base(pfad))
	}

	if pingfehler := db.PingContext(ctx); pingfehler != nil {
		_ = db.Close()

		return nil, domain.MitGrund(domain.ErrSicherung, pingfehler,
			"%q ließ sich nicht öffnen — wähle eine andere Sicherung",
			filepath.Base(pfad))
	}

	return db, nil
}

// lesenUndPruefen ist die Prüfung selbst, an einer offenen Verbindung.
//
// **Die Migrationsquelle als Parameter**, wie bei `migriereAus`: sonst wäre der
// Fehlerpfad von [migrationsnamen] ein Zweig, den kein Test betreten kann —
// das eingebettete Verzeichnis ist immer in Ordnung (§3).
func lesenUndPruefen(
	ctx context.Context, db *sql.DB, quelle, basis fs.FS, pfad string,
) (Sicherungsbefund, error) {
	// **Die Form zuerst.** Eine Datei ohne `bestand` ist keine Sicherung
	// dieser Anwendung, und die Frage nach ihren Migrationen wäre dann eine
	// Frage an eine fremde Datenbank.
	angelegt, err := AngelegtAm(ctx, db)
	if err != nil {
		return Sicherungsbefund{}, domain.MitGrund(domain.ErrSicherung, err,
			"%q ist keine brauchbare Sicherung — wähle eine andere",
			filepath.Base(pfad))
	}

	// **Ohne Prüfsummenspalte nur die Namen** (Zweitblick über die Folgen vor
	// 0.8.0, Befund 2): eine sehr alte Sicherung scheiterte sonst an der
	// Spalte, bevor ihr Alter erkannt war, und bekam den falschen Rat. Der
	// Start liest dieselbe Buchführung ebenso (`verbuchteNamen`).
	lesen := angewendeteMitSummen
	if !buchfuehrungHat(ctx, db, "pruefsumme") {
		lesen = verbuchteNamen
	}

	verbucht, err := lesen(ctx, db)
	if err != nil {
		return Sicherungsbefund{}, domain.MitGrund(domain.ErrSicherung, err,
			"%q ist keine brauchbare Sicherung — wähle eine andere",
			filepath.Base(pfad))
	}

	namen, err := migrationsnamen(quelle)
	if err != nil {
		return Sicherungsbefund{}, err
	}

	// **Die Grundlinie zählt mit** (Vermerk 0189): ihre Wanderungen sind
	// fort als Dateien und da mit Name und Prüfsumme.
	var (
		umfang []string
		summen map[string]string
	)

	if basis != nil {
		if umfang, summen, err = grundlinienumfang(basis); err != nil {
			return Sicherungsbefund{}, err
		}
	}

	namen = append(slices.Clone(umfang), namen...)

	// **Dieselbe Abweisung wie beim Start**, und das ist der Punkt: eine
	// Sicherung, die diese Fassung nicht öffnen könnte, wird nicht erst
	// übernommen und dann verweigert.
	// **Hier bleibt das zweite `%w`**, und das ist eine Entscheidung: was
	// `unbekannteAbweisen` sagt, ist **die** Auskunft — der Name der
	// Migration, die diese Fassung nicht kennt. Ihn zu verdecken machte aus
	// „diese Sicherung stammt aus einer neueren Fassung" ein „irgendetwas
	// stimmt nicht". Der Wächter über die Meldungswege führt die Stelle mit
	// diesem Grund als Ausnahme.
	if err := unbekannteAbweisen(namen, verbucht); err != nil {
		return Sicherungsbefund{}, fmt.Errorf("%w: %q: %w",
			domain.ErrSicherung, filepath.Base(pfad), err)
	}

	// **Und eine Sicherung von vor 0.7.0 ebenso** — sie ließe sich
	// zurückspielen, aber nicht mehr wandern.
	if basis != nil {
		if err := grundlinieGefuehrt(umfang, summen, verbucht); err != nil {
			return Sicherungsbefund{}, fmt.Errorf("%w: %q: %w",
				domain.ErrSicherung, filepath.Base(pfad), err)
		}
	}

	// **Und die Mindestform** (#963): eine Datei, die Migrationen verbucht hat,
	// deren Tabellen fehlen, ist keine brauchbare Sicherung — und das fällt
	// sonst erst beim nächsten fachlichen Aufruf auf, nachdem die Oberfläche
	// „Zurückgespielt" gemeldet hat.
	if err := mindestformPruefen(ctx, db, quelle, basis, umfang, verbucht); err != nil {
		return Sicherungsbefund{}, fmt.Errorf("%q: %w", pfad, err)
	}

	return befundLesen(ctx, db, angelegt, verbucht), nil
}

// befundLesen sammelt die Angaben, die den Nutzenden die Entscheidung erlauben.
//
// **Ohne Fehler.** Was scheitern konnte, ist oben geschehen; was hier noch
// gelesen wird, ist Komfort — und ein Komfort, der die Prüfung kippen könnte,
// wäre keiner.
func befundLesen(
	ctx context.Context, db *sql.DB, angelegt string, verbucht map[string]string,
) Sicherungsbefund {
	befund := Sicherungsbefund{AngelegtAm: angelegt}

	for name := range verbucht {
		if name > befund.Schemastand {
			befund.Schemastand = name
		}
	}

	// **Zählen darf scheitern, ohne die Prüfung zu kippen.** Wer eine
	// Sicherung zurückspielt, tut es wegen ihres Inhalts; dass die Zahl
	// danebensteht, ist Komfort. Ein `sets`, das es in einer sehr alten
	// Fassung noch nicht gab, macht die Datei nicht unbrauchbar.
	befund.Sets = zaehlen(ctx, db, "sets")
	befund.Karten = zaehlen(ctx, db, "cards")

	return befund
}

// zaehlen gibt die Zeilenzahl einer Tabelle, oder −1, wenn es sie nicht gibt.
//
// **−1 und nicht 0.** „Diese Fassung kennt die Tabelle nicht" und „sie ist
// leer" sind zwei Auskünfte, und eine Oberfläche, die beide als `0` bekäme,
// zeigte für die erste eine Zahl, die nicht stimmt.
func zaehlen(ctx context.Context, db *sql.DB, tabelle string) int {
	var anzahl int

	// Der Tabellenname kommt aus dieser Datei und aus keiner Eingabe; ein
	// Platzhalter ist für einen Bezeichner ohnehin nicht vorgesehen.
	//
	// **Ein Zweig und nicht zwei.** `count(*)` gibt immer eine Zeile; ein
	// `sql.ErrNoRows` daneben wäre ein Fall, den kein Aufruf herstellt (§3).
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM `+tabelle).Scan(&anzahl); err != nil { //nolint:gosec // fester Bezeichner
		return -1
	}

	return anzahl
}
