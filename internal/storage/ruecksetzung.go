package storage

// **Was eine Rücksetzung anfassen würde** — Werkstück O4 von Phase 5.2
// (#1428).
//
// **Rein lesend.** Was hier steht, ist die Vorschau; gelöscht wird über die
// Wege, die es dafür schon gibt (`Nutzerablage.ProfilLoeschen`,
// `NutzerLoeschen`, `Einstellungen`). Diese Datei zählt nur — und **nennt**.
//
// # Warum sie zählt und nicht schätzt
//
// „Dieses Profil und alles, was dazugehört" ist keine Auskunft, mit der jemand
// entscheidet. „Drei Sets, 107 Karten, zwei Token, ein Abzeichen" ist eine.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ruecksetzungsbeispiele ist, wie viele Namen eine Gruppe beispielhaft nennt.
const Ruecksetzungsbeispiele = 5

// zaehlquelle ist, woraus gezählt wird — die Verbindung **oder** ein Geschäft.
//
// **Die zweite Lesart ist der Grund für diese Zeile** (Befund 9 des Nachlaufs
// zum Zweitblick über Phase 5.2): der Vergleich mit der Vorschau lag vor den
// Schreibaufrufen und nicht in ihrem Geschäft. Wer denselben Zähler im
// Schreibgeschäft laufen lassen will, braucht ihn über etwas, das ein `*sql.Tx`
// auch ist.
type zaehlquelle interface {
	QueryRowContext(ctx context.Context, abfrage string, werte ...any) *sql.Row
	QueryContext(ctx context.Context, abfrage string, werte ...any) (*sql.Rows, error)
}

// zaehlstelle ist eine Gruppe samt ihren zwei Abfragen.
type zaehlstelle struct {
	was string

	// zahl zählt, namen nennt ein paar. **Zwei Abfragen und nicht eine**: die
	// Zahl gilt für alles, die Namen für die ersten fünf.
	zahl  string
	namen string
}

// profilstellen sind die Gruppen, die zu einem Profil gehören.
func profilstellen() []zaehlstelle {
	return []zaehlstelle{
		{
			was:   domain.BetroffenSets,
			zahl:  `SELECT count(*) FROM sets WHERE profil_id = ?`,
			namen: `SELECT name FROM sets WHERE profil_id = ? ORDER BY name LIMIT ?`,
		},
		{
			was: domain.BetroffenKarten,
			zahl: `SELECT count(*) FROM cards
			         WHERE set_id IN (SELECT id FROM sets WHERE profil_id = ?)`,
			namen: `SELECT '' FROM cards
			          WHERE set_id IN (SELECT id FROM sets WHERE profil_id = ?)
			          LIMIT ?`,
		},
		{
			was:   domain.BetroffenDecks,
			zahl:  `SELECT count(*) FROM decks WHERE profil_id = ?`,
			namen: `SELECT name FROM decks WHERE profil_id = ? ORDER BY name LIMIT ?`,
		},
		{
			was:   domain.BetroffenExemplare,
			zahl:  `SELECT coalesce(sum(menge), 0) FROM sammlung WHERE profil_id = ?`,
			namen: `SELECT name FROM sammlung WHERE profil_id = ? AND name <> '' ORDER BY name LIMIT ?`,
		},
		{
			// **Die Token werden benannt** — die Abnahme verlangt es:
			// „Profillöschung benennt und widerruft seine Token."
			was:   domain.BetroffenToken,
			zahl:  `SELECT count(*) FROM zugriffstoken WHERE profil_id = ?`,
			namen: `SELECT bezeichnung FROM zugriffstoken WHERE profil_id = ? ORDER BY bezeichnung LIMIT ?`,
		},
		{
			// **Ohne die Freigabe des Neuanfangs selbst** (G12, #2282): sie
			// erlaubt die Rücksetzung und ist nicht, was sie wegnimmt. Mit
			// ihr änderte das Erteilen die Vorschau und damit ihre Revision,
			// und die Ausführung wies sich selbst ab — gefunden im ersten
			// Lauf am nativen Fenster (2026-10-10). Der Zweck steht wörtlich
			// da und ist [domain.RuecksetzungsZweck]; ein Test hält beide
			// gleich.
			was: domain.BetroffenFreigaben,
			zahl: `SELECT count(*) FROM freigabenachweise
			         WHERE profil_id = ? AND zweck <> 'ruecksetzung.anwendung'`,
			namen: `SELECT zweck FROM freigabenachweise
			          WHERE profil_id = ? AND zweck <> 'ruecksetzung.anwendung'
			          ORDER BY zweck LIMIT ?`,
		},
		{
			// **Die Zugänge anderer an diesem Profil** (#2287): sie gehen mit
			// ihm (`ON DELETE CASCADE`, Wanderung 0044). Benannt wird, wer
			// sie hat — es ist das eigene Profil, und die Seite der Freigaben
			// nennt dieselben Namen.
			was:  domain.BetroffenZugaenge,
			zahl: `SELECT count(*) FROM profilfreigaben WHERE profil_id = ?`,
			namen: `SELECT n.name FROM profilfreigaben f JOIN nutzer n ON n.id = f.nutzer_id
			          WHERE f.profil_id = ? ORDER BY n.name LIMIT ?`,
		},
		{
			was:   domain.BetroffenEinstellungen,
			zahl:  `SELECT count(*) FROM settings WHERE profil_id = ?`,
			namen: `SELECT schluessel FROM settings WHERE profil_id = ? ORDER BY schluessel LIMIT ?`,
		},
	}
}

// Profilumfang zählt, was zu einem Profil gehört.
func Profilumfang(ctx context.Context, db zaehlquelle, profilID int64) (
	[]domain.Betroffenes, error,
) {
	betroffen := []domain.Betroffenes{}

	for _, stelle := range profilstellen() {
		gruppe, fehler := stelle.zaehlen(ctx, db, profilID)
		if fehler != nil {
			return nil, fehler
		}

		if gruppe.Anzahl == 0 {
			continue
		}

		betroffen = append(betroffen, gruppe)
	}

	return betroffen, nil
}

// zaehlen holt Zahl und Namen einer Gruppe.
func (z zaehlstelle) zaehlen(ctx context.Context, db zaehlquelle, profilID int64) (
	domain.Betroffenes, error,
) {
	gruppe := domain.Betroffenes{Was: z.was, Namen: []string{}}

	if fehler := db.QueryRowContext(ctx, z.zahl, profilID).Scan(&gruppe.Anzahl); fehler != nil {
		return gruppe, uebersetze(fehler, "%s zählen", z.was)
	}

	if gruppe.Anzahl == 0 {
		return gruppe, nil
	}

	zeilen, fehler := db.QueryContext(ctx, z.namen, profilID, Ruecksetzungsbeispiele)
	if fehler != nil {
		return gruppe, uebersetze(fehler, "%s benennen", z.was)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	for zeilen.Next() {
		var name string
		if lesefehler := zeilen.Scan(&name); lesefehler != nil {
			return gruppe, uebersetze(lesefehler, "einen Namen aus %s lesen", z.was)
		}

		// **Leere Namen kommen nicht in die Liste.** Eine Karte hat hier
		// keinen; die Zahl sagt trotzdem, wie viele es sind.
		if name != "" {
			gruppe.Namen = append(gruppe.Namen, name)
		}
	}

	if fehler := zeilen.Err(); fehler != nil {
		return gruppe, uebersetze(fehler, "%s benennen", z.was)
	}

	return gruppe, nil
}

// OffeneRueckkehrmengen sagt, ob über Mengen eines Profils nichts feststeht.
//
// **Das ist die Stelle, an der T2, T5 und T6 andocken** (#1428, Vermerk
// [0088](../../docs/entscheidungen/0088-die-epoche-einer-rueckkehr.md)): eine
// ungeklärte Verbindlichkeit sperrt die Löschung, und der belegwahrende Weg
// heraus ist die **Quittung**.
func OffeneRueckkehrmengen(ctx context.Context, db *sql.DB, profilID int64) (int, error) {
	var offen int

	if fehler := db.QueryRowContext(ctx,
		`SELECT count(*) FROM wiederherstellungsmengen m
		   JOIN wiederherstellungen w ON w.id = m.wiederherstellung_id
		  WHERE m.profil_id = ? AND w.klaerung = 'offen'`,
		profilID).Scan(&offen); fehler != nil {
		return 0, uebersetze(fehler, "die offenen Rückkehren zählen")
	}

	return offen, nil
}

// Nutzerzahl sagt, wie viele Menschen es gibt.
//
// **Der letzte lässt sich nicht löschen** — das weist `NutzerLoeschen` ab, und
// die Vorschau sagt es **vorher**.
func Nutzerzahl(ctx context.Context, db *sql.DB) (int, error) {
	var anzahl int

	if fehler := db.QueryRowContext(ctx,
		`SELECT count(*) FROM nutzer`).Scan(&anzahl); fehler != nil {
		return 0, uebersetze(fehler, "die Nutzer zählen")
	}

	return anzahl, nil
}

// EinstellungenZuruecksetzen nimmt die Wahl eines Profils zurück.
//
// **Gelöscht und nicht auf die Vorgabe gesetzt.** Eine Einstellung, die es
// nicht gibt, **erbt** — und genau das verlangt die Abnahme: „Profilreset
// stellt Vererbung her." Ein geschriebenes `erben` wäre eine Wahl, die wie
// keine aussieht.
//
// **Mit `schluessel` leer gilt es für alle** — sonst für die genannten. So
// fasst die Layoutrücksetzung genau ihre vier an und keinen mehr.
func EinstellungenZuruecksetzen(
	ctx context.Context, db *sql.DB, schluessel []string,
) (int, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return 0, profilfehler
	}

	ergebnis, fehler := loeschen(ctx, db,
		`DELETE FROM settings WHERE profil_id = ?`, profilID, schluessel)
	if fehler != nil {
		return 0, uebersetze(fehler, "die Einstellungen zurücksetzen")
	}

	betroffen, zaehlfehler := ergebnis.RowsAffected()
	if zaehlfehler != nil {
		return 0, uebersetze(zaehlfehler, "die zurückgesetzten Einstellungen zählen")
	}

	return int(betroffen), nil
}

// NutzereinstellungenZuruecksetzen nimmt die Wahl eines Menschen zurück.
func NutzereinstellungenZuruecksetzen(
	ctx context.Context, db *sql.DB, nutzerID int64,
) (int, error) {
	ergebnis, fehler := db.ExecContext(ctx,
		`DELETE FROM nutzereinstellungen WHERE nutzer_id = ?`, nutzerID)
	if fehler != nil {
		return 0, uebersetze(fehler, "die Nutzereinstellungen zurücksetzen")
	}

	betroffen, zaehlfehler := ergebnis.RowsAffected()
	if zaehlfehler != nil {
		return 0, uebersetze(zaehlfehler, "die zurückgesetzten Einstellungen zählen")
	}

	return int(betroffen), nil
}

// loeschen führt die Anweisung aus — mit oder ohne Schlüsselklausel.
//
// **Die Klausel entsteht aus Platzhaltern und nicht aus den Werten.** Ein
// `IN ('a','b')` zusammenzusetzen hieße, Eingaben in SQL zu schreiben; hier
// stehen nur Fragezeichen, und die Werte reisen daneben.
func loeschen(
	ctx context.Context, db *sql.DB, anweisung string, profilID int64, schluessel []string,
) (sql.Result, error) {
	werte := []any{profilID}

	if len(schluessel) == 0 {
		return db.ExecContext(ctx, anweisung, werte...) //nolint:wrapcheck // der Aufrufer sagt, worum es ging
	}

	platzhalter := make([]string, 0, len(schluessel))

	for _, einer := range schluessel {
		platzhalter = append(platzhalter, "?")
		werte = append(werte, einer)
	}

	return db.ExecContext(ctx, //nolint:wrapcheck // dito
		anweisung+" AND schluessel IN ("+strings.Join(platzhalter, ", ")+")", werte...)
}

// AlleProfile gibt jedes Profil des Bestands, nach Mensch und Namen geordnet.
//
// **Für den Neuanfang** — Befund 1 des Zweitblicks über Phase G (#2282): er
// legt die ganze Datei beiseite, und eine Vorschau über die Profile dessen,
// der ruft, zeigte nicht, was mit ihr geht.
func AlleProfile(ctx context.Context, db *sql.DB) ([]domain.Profil, error) {
	zeilen, fehler := db.QueryContext(ctx,
		`SELECT `+profilspalten+` FROM profile ORDER BY nutzer_id, name`)
	if fehler != nil {
		return nil, uebersetze(fehler, "alle Profile lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	var gelesen []domain.Profil

	for zeilen.Next() {
		var eines domain.Profil

		if lesefehler := scanneProfil(zeilen, &eines); lesefehler != nil {
			return nil, uebersetze(lesefehler, "eine Profilzeile lesen")
		}

		gelesen = append(gelesen, eines)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "alle Profile lesen")
	}

	return gelesen, nil
}

// FremdeZugaenge zählt die Zugänge eines Menschen an Profilen, die ihm nicht
// gehören — sie gehen mit ihm (#2287).
//
// **Gezählt und nicht benannt**: die Profile gehören anderen, und ihre Namen
// gehören nicht in die Vorschau dessen, der löscht — dieselbe Regel wie beim
// Neuanfang (`service.ohneNamen`).
func FremdeZugaenge(ctx context.Context, db zaehlquelle, nutzerID int64) (domain.Betroffenes, error) {
	gruppe := domain.Betroffenes{Was: domain.BetroffenZugaenge, Namen: []string{}}

	if fehler := db.QueryRowContext(ctx,
		`SELECT count(*) FROM profilfreigaben f JOIN profile p ON p.id = f.profil_id
		  WHERE f.nutzer_id = ? AND p.nutzer_id <> f.nutzer_id`,
		nutzerID).Scan(&gruppe.Anzahl); fehler != nil {
		return gruppe, uebersetze(fehler, "die fremden Zugänge zählen")
	}

	return gruppe, nil
}
