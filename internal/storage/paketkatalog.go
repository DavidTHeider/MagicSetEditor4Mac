package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Paketkatalog führt die installierten Pakete im Bestand.
//
// # Warum die Tabelle bis hierher leer war
//
// `packages` steht seit Phase 2 im Schema. Bis
// #54 hat
// niemand hineingeschrieben — der Zweitblick über Phase 9 hat es angestrichen:
// eine Tabelle, in die nur hineingeschrieben wird (hier: gar nicht), ist eine
// Zusage ohne Einlösung.
type Paketkatalog struct {
	db *sql.DB
}

// NeuerPaketkatalog baut den Katalog über einer Datenbank.
func NeuerPaketkatalog(db *sql.DB) *Paketkatalog { return &Paketkatalog{db: db} }

// Abgleichen bringt den Katalog auf den Stand einer Paketliste.
//
// **Abgleich und nicht Ergänzung.** Wer ein Paket aus dem Verzeichnis nimmt,
// hätte sonst einen Eintrag, der auf nichts zeigt — und die Stilauswahl böte
// einen Stil an, den es nicht mehr gibt.
//
// **In einer Transaktion**, aus demselben Grund wie beim Set: ein halb
// abgeglichener Katalog sieht wie ein Katalog aus.
//
// **Herkunft und Prüfsumme überleben.** Wer die Pakete von der Platte liest,
// weiß nicht, woher sie kommen; beide Spalten werden deshalb nur gesetzt, wenn
// die neue Angabe etwas sagt. Ohne diese Regel verlöre die erste Auffrischung
// genau die Angaben, wegen derer die Spalten geführt werden (§11, #56).
func (k *Paketkatalog) Abgleichen(ctx context.Context, pakete []domain.Paket) error {
	for _, paket := range pakete {
		if err := paket.Validate(); err != nil {
			return fmt.Errorf("das Paket %q: %w", paket.Name, err)
		}
	}

	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "Transaktion beginnen")
	}
	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach dem Commit ist sie zu

	vorher, err := aenderungen(ctx, tx)
	if err != nil {
		return err
	}

	behalten := make([]any, 0, len(pakete))

	for _, paket := range pakete {
		kennung, fehler := paketZeile(ctx, tx, paket)
		if fehler != nil {
			return fehler
		}

		if bedarf := bedarfSchreiben(ctx, tx, kennung, paket.Braucht); bedarf != nil {
			return bedarf
		}

		behalten = append(behalten, kennung)
	}

	if verwaist := verwaisteEntfernen(ctx, tx, behalten); verwaist != nil {
		return verwaist
	}

	// **Ohne Änderung wird nicht bestätigt** —
	// #1325.
	//
	// `App.Pakete` frischt bei **jedem** Aufruf auf, und ein Commit hebt
	// `PRAGMA data_version` für die übrigen Verbindungen. Das Fenster liest
	// daraus, ob ein **anderer** Prozess geschrieben hat — und lud daraufhin
	// alle Seiten neu. Die Kette ist ein Kreis: `Pakete` schreibt, der Stand
	// steigt, die Startseite lädt neu, sie ruft `Pakete`. Am 2026-09-03
	// gemessen: zehn Sekunden Leerlauf, zehnmal `Sets`, sichtbar als Flackern.
	//
	// **Dass der Kreis das eigene Fenster erreichte, war der zweite Fehler**
	// (#1485, behoben am 2026-09-09: eine Verbindung je Vorrat). Die Vorsicht
	// hier bleibt trotzdem: ein anderes Fenster liest einen leeren Commit
	// weiter als Änderung, und Arbeit ohne Änderung ist keine.
	//
	// **`total_changes` und nicht `RowsAffected`.** Der Zähler steht an der
	// Verbindung und summiert **alle** Schreibvorgänge dieser Transaktion —
	// die Katalogzeilen, die Abhängigkeiten und das Räumen. Ein
	// `RowsAffected` je Anweisung wäre dieselbe Zahl in vielen Stücken.
	geschrieben, err := aenderungenSeit(ctx, tx, vorher)
	if err != nil {
		return err
	}

	if !geschrieben {
		// **Zurückgerollt und nicht bestätigt.** Eine Transaktion, die nichts
		// geändert hat, hat auch nichts zu bestätigen; das `defer` oben
		// erledigt es.
		return nil
	}

	if abschluss := tx.Commit(); abschluss != nil {
		return uebersetze(abschluss, "den Katalog abschließen")
	}

	return nil
}

// paketZeile legt eine Katalogzeile an oder frischt sie auf.
func paketZeile(ctx context.Context, tx *sql.Tx, paket domain.Paket) (int64, error) {
	var kennung int64

	err := tx.QueryRowContext(ctx,
		`INSERT INTO packages (art, name, version, pfad, herkunft, pruefsumme, gruppe, rang, kurzname, abdruck)
		      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (art, name) DO UPDATE SET
		      version    = excluded.version,
		      pfad       = excluded.pfad,
		      gruppe     = excluded.gruppe,
		      rang       = excluded.rang,
		      kurzname   = excluded.kurzname,
		      herkunft   = CASE WHEN excluded.herkunft = '' THEN packages.herkunft
		                        ELSE excluded.herkunft END,
		      pruefsumme = CASE WHEN excluded.pruefsumme = '' THEN packages.pruefsumme
		                        ELSE excluded.pruefsumme END,
		      abdruck    = CASE WHEN excluded.abdruck = '' THEN packages.abdruck
		                        ELSE excluded.abdruck END
		    WHERE packages.version IS NOT excluded.version
		       OR packages.pfad IS NOT excluded.pfad
		       OR packages.gruppe IS NOT excluded.gruppe
		       OR packages.rang IS NOT excluded.rang
		       OR packages.kurzname IS NOT excluded.kurzname
		       OR (excluded.herkunft <> '' AND packages.herkunft IS NOT excluded.herkunft)
		       OR (excluded.pruefsumme <> '' AND packages.pruefsumme IS NOT excluded.pruefsumme)
		       OR (excluded.abdruck <> '' AND packages.abdruck IS NOT excluded.abdruck)
		   RETURNING id`,
		paket.Art, paket.Name, paket.Version, paket.Pfad,
		paket.Herkunft, paket.Pruefsumme, paket.Gruppe, paket.Rang,
		paket.Kurzname, paket.Abdruck).Scan(&kennung)

	// **Keine Zeile heißt: es gab nichts zu ändern.**
	//
	// Das `WHERE` am `DO UPDATE` verhindert den Schreibvorgang, wenn jeder
	// Wert schon so dasteht — und dann gibt `RETURNING` nichts zurück. Die
	// Kennung wird trotzdem gebraucht: die Abhängigkeiten hängen daran.
	if errors.Is(err, sql.ErrNoRows) {
		return kennungVon(ctx, tx, paket)
	}

	if err != nil {
		return 0, uebersetze(err, "das Paket %q eintragen", paket.Name)
	}

	return kennung, nil
}

// kennungVon liest die Kennung einer Katalogzeile, die schon so dasteht.
func kennungVon(ctx context.Context, tx *sql.Tx, paket domain.Paket) (int64, error) {
	var kennung int64

	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM packages WHERE art = ? AND name = ?`,
		paket.Art, paket.Name).Scan(&kennung); err != nil {
		return 0, uebersetze(err, "die Kennung des Pakets %q lesen", paket.Name)
	}

	return kennung, nil
}

// bedarfSchreiben ersetzt die Abhängigkeiten eines Pakets.
//
// **Ersetzen und nicht ergänzen**: ein Paket, das eine Abhängigkeit verliert,
// behielte sie sonst im Katalog — und #57 zöge daraus einen Schluss über einen
// Stand, den es nicht mehr gibt.
// **Und nur, wenn sie sich unterscheiden** (#1325). Ein `DELETE` samt
// `INSERT` bei jedem Abgleich schreibt auch dann, wenn nichts anders ist — und
// jeder Commit hebt den Datenstand, aus dem das Fenster auf eine **fremde**
// Änderung schließt.
func bedarfSchreiben(ctx context.Context, tx *sql.Tx, kennung int64, braucht []domain.Bedarf) error {
	gleich, err := bedarfGleich(ctx, tx, kennung, braucht)
	if err != nil {
		return err
	}

	if gleich {
		return nil
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM package_deps WHERE package_id = ?`, kennung); err != nil {
		return uebersetze(err, "die Abhängigkeiten räumen")
	}

	for stelle, bedarf := range braucht {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO package_deps (package_id, braucht, version, reihenfolge)
			      VALUES (?, ?, ?, ?)`,
			kennung, bedarf.Paket, bedarf.Version, stelle); err != nil {
			return uebersetze(err, "die Abhängigkeit %q eintragen", bedarf.Paket)
		}
	}

	return nil
}

// verwaisteEntfernen räumt, was in der abgeglichenen Liste nicht vorkam.
//
// **Über die Kennungen und nicht über (art, name)**: die Kennungen stehen
// gerade fest, und ein zweites Kriterium wäre eine zweite Gelegenheit, sich zu
// vertun. Eine leere Liste räumt alles — das ist der Fall „kein Paket mehr
// installiert" und kein Sonderfall.
func verwaisteEntfernen(ctx context.Context, tx *sql.Tx, behalten []any) error {
	// **Die Zusammensetzung besteht aus Fragezeichen und sonst nichts.**
	// gosec sieht eine SQL-Zeichenkette, die aus Teilen entsteht (G202); was
	// hinzukommt, ist `platzhalter` — eine Folge aus `, ?`, gebaut über eine
	// Zahl. Kein Wert aus dem Aufruf berührt den Text; die Kennungen reisen als
	// Argumente.
	frage := `DELETE FROM packages`
	if len(behalten) > 0 {
		frage += ` WHERE id NOT IN (?` + platzhalter(len(behalten)-1) + `)` // #nosec G202 -- der Grund steht darüber
	}

	if _, err := tx.ExecContext(ctx, frage, behalten...); err != nil {
		return uebersetze(err, "verwaiste Katalogzeilen räumen")
	}

	return nil
}

// platzhalter gibt `, ?` so oft, wie weitere Werte folgen.
func platzhalter(weitere int) string {
	rest := ""
	for range weitere {
		rest += ", ?"
	}

	return rest
}

// Pakete gibt den Katalog, nach Art und Name geordnet.
//
// **Geordnet, und das ist keine Zier**: eine Liste ohne Ordnung ist bei jedem
// Aufruf eine andere, und die Oberfläche zeigte sie springend.
//
// **Eine Abfrage und nicht zwei.** Der erste Anlauf las Pakete und
// Abhängigkeiten getrennt — das sind zwei Fehlerwege für eine Auskunft, und
// jeder davon will nachgestellt werden. Ein `LEFT JOIN` gibt je Abhängigkeit
// eine Zeile; zusammengefasst wird über die Kennung, die dabei ohnehin
// mitkommt.
func (k *Paketkatalog) Pakete(ctx context.Context) ([]domain.Paket, error) {
	zeilen, err := k.db.QueryContext(ctx,
		`SELECT p.id, p.art, p.name, p.version, p.pfad, p.herkunft, p.pruefsumme,
		        p.gruppe, p.rang, p.kurzname, p.abdruck,
		        coalesce(d.braucht, ''), coalesce(d.version, '')
		   FROM packages p
		   LEFT JOIN package_deps d ON d.package_id = p.id
		  ORDER BY p.art, p.name, d.reihenfolge`)
	if err != nil {
		return nil, uebersetze(err, "den Paketkatalog lesen")
	}
	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler kommt über zeilen.Err

	pakete := make([]domain.Paket, 0)
	letzte := int64(0)

	for zeilen.Next() {
		var (
			kennung int64
			paket   domain.Paket
			bedarf  domain.Bedarf
		)

		if err := zeilen.Scan(&kennung, &paket.Art, &paket.Name, &paket.Version,
			&paket.Pfad, &paket.Herkunft, &paket.Pruefsumme,
			&paket.Gruppe, &paket.Rang, &paket.Kurzname, &paket.Abdruck,
			&bedarf.Paket, &bedarf.Version); err != nil {
			return nil, uebersetze(err, "eine Katalogzeile lesen")
		}

		if kennung != letzte {
			letzte = kennung

			pakete = append(pakete, paket)
		}

		// **Ein leerer Name heißt „keine Abhängigkeit".** Der `LEFT JOIN` gibt
		// für ein Paket ohne `depends on:` eine Zeile mit NULL, und `coalesce`
		// macht daraus die leere Zeichenkette. Ein Bedarf ohne Namen weist die
		// Domäne ohnehin ab (`domain.Paket.Validate`) — er kann also nur von
		// hier kommen.
		if bedarf.Paket != "" {
			pakete[len(pakete)-1].Braucht = append(pakete[len(pakete)-1].Braucht, bedarf)
		}
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "den Paketkatalog lesen")
	}

	return pakete, nil
}

// PaketEntfernen nimmt einen Eintrag aus dem Katalog.
//
// **Ohne Zählung, und das ist eine Entscheidung.** Der erste Anlauf las
// `RowsAffected` und meldete bei null ein „nicht gefunden". Das ist ein
// Fehlerweg, den niemand betreten kann: wer entfernt, hat das Paket vorher
// gesucht — `service.Paketverwalter.Entfernen` braucht seinen Pfad, um die
// Dateien zu räumen, und meldet das Fehlen dort. Zwei Urteile über dieselbe
// Frage laufen auseinander, und das zweite hier wäre das ungeprüfte.
func (k *Paketkatalog) PaketEntfernen(ctx context.Context, art, name string) error {
	if _, err := k.db.ExecContext(ctx,
		`DELETE FROM packages WHERE art = ? AND name = ?`, art, name); err != nil {
		return uebersetze(err, "das Paket %q entfernen", name)
	}

	return nil
}

// PaketeNachMuster gibt die Pakete, die auf ein `match:`-Muster passen.
//
// # Wie MSE die Liste baut
//
// **ported.** Für eine `package choice` fragt MSE seinen Paketverwalter —
// `package_manager.findMatching(field().match, choices)` — und sortiert das
// Ergebnis nach `position_hint`, bei Gleichstand nach dem Namen
// (`src/render/value/package_choice.cpp`, Zeilen 18–43, Fassung f9d9356 vom
// 2021-07-06). Genau diese Ordnung steht im `ORDER BY`.
//
// # Wie das Muster zerfällt
//
// `match:` nennt einen Dateinamen samt Endung: `magic-overlay-*.mse-include`.
// Der Katalog führt beides getrennt — `art` ist die Endung, `name` der Rest —,
// und SQLites `GLOB` versteht denselben Platzhalter `*` wie MSE.
//
// **Ein Muster ohne Endung ist ein Befund.** Die Art stünde dann nirgends, und
// jede zu nehmen hieße, `magic-mana-*` auf Spielpakete passen zu lassen.
func (k *Paketkatalog) PaketeNachMuster(
	ctx context.Context, muster string,
) ([]domain.Paket, error) {
	punkt := strings.LastIndex(muster, ".")
	if punkt < 0 {
		return nil, fmt.Errorf("%w: das Muster %q nennt keine Paketart", domain.ErrValidation, muster)
	}

	namensmuster, art := muster[:punkt], muster[punkt+1:]

	zeilen, err := k.db.QueryContext(ctx,
		`SELECT art, name, version, pfad, herkunft, pruefsumme, gruppe, rang, kurzname
		   FROM packages
		  WHERE art = ? AND name GLOB ?
		  ORDER BY rang, name`, art, namensmuster)
	if err != nil {
		return nil, uebersetze(err, "die Pakete zum Muster %q suchen", muster)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler kommt über zeilen.Err

	pakete := make([]domain.Paket, 0)

	for zeilen.Next() {
		var paket domain.Paket

		if fehler := zeilen.Scan(&paket.Art, &paket.Name, &paket.Version, &paket.Pfad,
			&paket.Herkunft, &paket.Pruefsumme, &paket.Gruppe, &paket.Rang,
			&paket.Kurzname); fehler != nil {
			return nil, uebersetze(fehler, "eine Katalogzeile lesen")
		}

		pakete = append(pakete, paket)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "die Pakete zum Muster %q suchen", muster)
	}

	return pakete, nil
}
