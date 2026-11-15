// **Die Mindestform einer Sicherung** —
// #963.
//
// # Die Frage, die dahintersteht
//
// Was ist die Mindestform eines brauchbaren Bestands? „`angelegt_am` hat die
// richtige Gestalt" ist zu wenig — eine Datei ohne `sets` und `cards` bestand
// damit Auswahl **und** Start, und erst der nächste fachliche Aufruf scheiterte,
// nachdem die Oberfläche „Zurückgespielt" gemeldet hatte.
//
// „Jede Tabelle jeder Fassung" wäre zu viel: eine sehr alte Sicherung darf
// Tabellen nicht kennen, die es damals nicht gab, und wird migriert.
//
// # Der Schnitt: die Tabellen, die die **verbuchte** Folge angelegt hat
//
// Wer `0002_setbestand.sql` verbucht hat, muss `sets` haben — die Migration
// legt sie an. Wer sie nicht verbucht hat, muss es nicht.
//
// **Das hängt am Stand und nicht an einer Liste.** Eine feste Aufzählung
// müsste bei jeder neuen Migration mitwachsen, und die eine, die man vergisst,
// merkt niemand.

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// tabellenanlage findet `CREATE TABLE` und `CREATE VIRTUAL TABLE`.
//
// **Auch die virtuelle.** `card_suche` ist eine FTS5-Tabelle, und sie steht in
// `sqlite_master` wie jede andere — wer sie ausließe, ließe die Volltextsuche
// ungeprüft.
var tabellenanlage = regexp.MustCompile(
	`(?i)CREATE\s+(?:VIRTUAL\s+)?TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?([a-z_][a-z0-9_]*)`)

// tabellenabgang findet, was eine Migration wieder wegnimmt.
//
// # Warum es das braucht
//
// **Weil eine angelegte Tabelle nicht stehen bleiben muss.** SQLite kennt kein
// `ALTER TABLE … DROP CONSTRAINT`; wer eine Eindeutigkeit ändert, baut die
// Tabelle neu, kopiert, wirft die alte weg und benennt die neue um. Nach
// `0026_ordnung.sql` gibt es `sammlung_neu` nicht mehr — sie **ist** die
// `sammlung`.
//
// Ohne diese Zeile verlangte der Wächter eine Tabelle, die die Migration
// selbst wieder entfernt hat, und wies **jede** Sicherung ab
// (#1292).
//
// **Umbenannt heißt beides**: der alte Name geht, der neue kommt. Deshalb
// zwei Fänger.
var (
	tabellenabgang = regexp.MustCompile(
		`(?i)DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?["']?([a-z_][a-z0-9_]*)`)

	tabellenumbenennung = regexp.MustCompile(
		`(?i)ALTER\s+TABLE\s+["']?([a-z_][a-z0-9_]*)["']?\s+RENAME\s+TO\s+["']?([a-z_][a-z0-9_]*)`)
)

// mindestformPruefen weist eine Sicherung ab, der Tabellen ihres eigenen
// Standes fehlen.
//
// **Die Meldung nennt sie.** „Die Sicherung ist unbrauchbar" sagt niemandem,
// was zu tun ist (§8).
func mindestformPruefen(
	ctx context.Context, db *sql.DB, quelle, basis fs.FS, umfang []string, verbucht map[string]string,
) error {
	erwartet, err := tabellenAusMigrationen(quelle, basis, umfang, verbucht)
	if err != nil {
		return err
	}

	fehlend, err := fehlendeTabellen(ctx, db, erwartet)
	if err != nil {
		return err
	}

	if fehlend == "" {
		return nil
	}

	return fmt.Errorf(
		"%w: die Sicherung führt Migrationen, deren Tabellen fehlen: %s",
		domain.ErrSicherung, fehlend)
}

// tabellenAusMigrationen liest, welche Tabellen die verbuchten Migrationen
// angelegt haben.
//
// **Eine unbekannte Migration ist hier keine.** `unbekannteAbweisen` läuft
// vorher und hat sie schon zurückgewiesen; was hier ankommt, gibt es in der
// Quelle.
func tabellenAusMigrationen(quelle, basis fs.FS, umfang []string, verbucht map[string]string) ([]string, error) {
	// **Angelegt minus weggenommen.** Eine Migration darf eine Tabelle bauen
	// und wieder wegwerfen; verlangt wird, was am Ende dasteht.
	//
	// **Über eine Menge und nicht über die Reihenfolge.** Die Migrationen
	// kommen aus einer Map und damit ungeordnet; wer nach Reihenfolge
	// rechnete, bekäme bei jedem Lauf ein anderes Ergebnis. Eine Tabelle, die
	// **irgendeine** verbuchte Migration wegnimmt, wird nicht verlangt — und
	// wer sie danach neu anlegt, legt sie unter demselben Namen an, also steht
	// sie ohnehin in `sqlite_master`.
	angelegt := map[string]bool{}
	weg := map[string]bool{}

	texte, err := migrationstexte(quelle, basis, umfang, verbucht)
	if err != nil {
		return nil, err
	}

	for _, text := range texte {
		for _, treffer := range tabellenanlage.FindAllStringSubmatch(text, -1) {
			angelegt[strings.ToLower(treffer[1])] = true
		}

		for _, treffer := range tabellenabgang.FindAllStringSubmatch(text, -1) {
			weg[strings.ToLower(treffer[1])] = true
		}

		for _, treffer := range tabellenumbenennung.FindAllStringSubmatch(text, -1) {
			weg[strings.ToLower(treffer[1])] = true
			angelegt[strings.ToLower(treffer[2])] = true
		}
	}

	namen := make([]string, 0, len(angelegt))

	for name := range angelegt {
		if !weg[name] {
			namen = append(namen, name)
		}
	}

	slices.Sort(namen)

	return namen, nil
}

// fehlendeTabellen gibt die erwarteten Tabellen, die es nicht gibt — als eine
// Zeile, getrennt durch Komma.
//
// # Eine Abfrage und nicht drei Fehlerwege
//
// Der erste Anlauf las alle Tabellen und verglich in Go. Das gibt **drei**
// Fehlerwege — Abfrage, Auslesen, Abschluss —, und zwei davon kann kein Lauf
// betreten: eine Abfrage auf `sqlite_master`, die eine Zeile liefert, liefert
// eine Zeichenkette. Ein Zweig, den kein Aufruf betritt, ist eine Behauptung
// (§3).
//
// **Über `QueryRow` ist es einer.** SQLite rechnet den Vergleich selbst, und
// `group_concat` gibt die Namen für die Meldung zurück.
func fehlendeTabellen(ctx context.Context, db *sql.DB, erwartet []string) (string, error) {
	if len(erwartet) == 0 {
		return "", nil
	}

	stellen := strings.TrimSuffix(strings.Repeat("(?),", len(erwartet)), ",")

	werte := make([]any, 0, len(erwartet))
	for _, name := range erwartet {
		werte = append(werte, name)
	}

	var fehlend sql.NullString

	err := db.QueryRowContext(ctx, `
		WITH erwartet(name) AS (VALUES `+stellen+`)
		SELECT group_concat(name, ', ') FROM erwartet
		 WHERE name NOT IN (SELECT lower(name) FROM sqlite_master WHERE type = 'table')`,
		werte...).Scan(&fehlend)
	if err != nil {
		return "", domain.MitGrund(domain.ErrSicherung, err,
			"die Tabellen dieser Datei ließen sich nicht lesen — wähle eine andere")
	}

	return fehlend.String, nil
}

// migrationstexte gibt den Text jeder verbuchten Wanderung — und für die, die
// die Grundlinie enthält (`umfang`), deren Tabellen (Vermerk 0189): die
// Wanderungen bis zu ihr gibt es nicht mehr als Datei.
func migrationstexte(quelle, basis fs.FS, umfang []string, verbucht map[string]string) ([]string, error) {
	enthalten := map[string]bool{}

	var texte []string

	for _, name := range umfang {
		enthalten[name] = true
	}

	if _, aufDerGrundlinie := verbucht[grundlinieBis]; aufDerGrundlinie && basis != nil {
		tabellen, _ := fs.Glob(basis, "grundlinie/tabellen-*.sql") //nolint:errcheck // ein fester Ausdruck
		for _, datei := range tabellen {
			roh, _ := fs.ReadFile(basis, datei) //nolint:errcheck // eben gefunden
			texte = append(texte, string(roh))
		}
	}

	for migration := range verbucht {
		if enthalten[migration] {
			continue
		}

		// **Unter `migrationen/`**, wie `migrationsnamen` es liest — die
		// eingebettete Quelle trägt das Verzeichnis, nicht die Dateien allein.
		roh, err := fs.ReadFile(quelle, path.Join("migrationen", migration))
		if err != nil {
			return nil, domain.MitGrund(domain.ErrSicherung, err,
				"die Migration %q ließ sich nicht lesen — diese Fassung der "+
					"Anwendung ist unvollständig", migration)
		}

		texte = append(texte, string(roh))
	}

	return texte, nil
}
