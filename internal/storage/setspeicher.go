package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// SetSpeicher legt Sets in der Datenbank ab.
type SetSpeicher struct {
	db *sql.DB
	// jetzt ist die Uhr. Als Feld, damit ein Test einen festen Zeitpunkt
	// vorgeben kann: eine Wanduhrzeit im Test macht ihn von der Uhr abhängig,
	// und dann ist er einmal im Jahr rot.
	jetzt func() time.Time
}

// NeuerSetSpeicher baut den Speicher über einer Datenbank.
func NeuerSetSpeicher(db *sql.DB) *SetSpeicher {
	return &SetSpeicher{db: db, jetzt: time.Now}
}

// MitUhr setzt die Zeitquelle.
//
// **Für Läufe, die eine Reihenfolge prüfen** (#891). `Zeitstempel` hat
// Millisekundenauflösung; zwei Aufrufe hintereinander landen im selben
// Augenblick, und dann entscheidet die Kennung statt der Zeit. Ein Lauf, der
// „zuletzt benutzt steht vorn" prüft, prüfte dann die Einfügereihenfolge.
//
// **Im Betrieb bleibt es `time.Now`** — die Naht ist eine Wahl, kein Zwang.
func (s *SetSpeicher) MitUhr(jetzt func() time.Time) *SetSpeicher {
	s.jetzt = jetzt

	return s
}

// SetAnlegen legt ein Set samt Karten, Werten und Rohblöcken ab.
//
// **In einer Transaktion**, und das ist keine Nebensache: ein Set, dessen
// Karten halb geschrieben sind, sieht wie ein Set aus. Wer es öffnet, findet
// weniger Karten als erwartet und keinen Hinweis darauf, warum.
func (s *SetSpeicher) SetAnlegen(
	ctx context.Context, set *domain.Set, roh []domain.RawBlock,
) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, uebersetze(err, "Transaktion beginnen")
	}
	defer func() { _ = tx.Rollback() }()

	kennung, fehler := setZeile(ctx, tx, set, Zeitstempel(s.jetzt()))
	if fehler != nil {
		return 0, fehler
	}

	if werte := werteSchreiben(ctx, tx, einfuegenSetwert, kennung, set.Values); werte != nil {
		return 0, werte
	}

	if karten := kartenSchreiben(ctx, tx, kennung, set.Cards); karten != nil {
		return 0, karten
	}

	if woerter := schluesselwoerterSchreiben(ctx, tx, kennung, set.Keywords); woerter != nil {
		return 0, woerter
	}

	if arten := setpackartenSchreiben(ctx, tx, kennung, set.Packs); arten != nil {
		return 0, arten
	}

	if bloecke := rohSchreiben(ctx, tx, kennung, roh); bloecke != nil {
		return 0, bloecke
	}

	if quelle := quelleSchreiben(ctx, tx, kennung, set.Quelle); quelle != nil {
		return 0, quelle
	}

	if abschluss := tx.Commit(); abschluss != nil {
		return 0, uebersetze(abschluss, "Set abschließen")
	}

	return kennung, nil
}

// setZeile legt die Setzeile an und gibt ihre Kennung.
//
// Über `RETURNING id` und nicht über `LastInsertId`. Der Unterschied ist nicht
// Geschmack: `LastInsertId` ist ein zweiter Aufruf mit einem zweiten
// Fehlerpfad, und dieser Pfad lässt sich mit einem Treiber, der ihn nie
// auslöst, nicht betreten. Ein Fehlerzweig, den kein Test je erreicht, ist eine
// Behauptung darüber, was im Ernstfall geschieht — hier gibt es ihn gar nicht.
func setZeile(
	ctx context.Context, tx *sql.Tx, set *domain.Set, zeitpunkt string,
) (int64, error) {
	var kennung int64

	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return 0, profilfehler
	}

	// **Die fünf Formatangaben stehen als Spalten** (#1325). Sie gehören zum
	// Format und nicht zum Spiel; was sie bedeuten, steht in
	// `domain/setkopf.go`.
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO sets (name, spiel, stil, herkunft, geoeffnet,
		                   mse_fassung, spielfassung, stilfassung,
		                   versionskontrolle, apprentice_code, profil_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		set.Name, set.Game, set.Style, set.Origin, zeitpunkt,
		set.Kopf.MSEFassung, set.Kopf.Spielfassung, set.Kopf.Stilfassung,
		set.Kopf.Versionskontrolle, set.Kopf.Apprenticecode,
		profilID).Scan(&kennung); err != nil {
		return 0, uebersetze(err, "Set schreiben")
	}

	return kennung, nil
}

// quelleSchreiben legt die Quelldatei ab — falls es eine gibt.
//
// **Ohne Quelle ist nichts zu tun**, und das ist kein Fehler: ein Set, das in
// der Anwendung entsteht, hat keine
// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
func quelleSchreiben(ctx context.Context, tx *sql.Tx, setID int64, quelle *domain.Quelle) error {
	if quelle == nil {
		return nil
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO set_quellen (set_id, datei, inhalt) VALUES (?, ?, ?)`,
		setID, quelle.Datei, quelle.Inhalt); err != nil {
		return uebersetze(err, "die Quelle des Sets schreiben")
	}

	return nil
}

func kartenSchreiben(
	ctx context.Context, tx *sql.Tx, setID int64, karten []*domain.Card,
) error {
	for nummer, karte := range karten {
		var kennung int64

		if err := tx.QueryRowContext(ctx,
			`INSERT INTO cards (set_id, reihenfolge) VALUES (?, ?) RETURNING id`,
			setID, nummer).Scan(&kennung); err != nil {
			return uebersetze(err, "Karte %d schreiben", nummer+1)
		}

		if fehler := werteSchreiben(ctx, tx, einfuegenKartenwert, kennung, karte.Values); fehler != nil {
			return fmt.Errorf("Karte %d: %w", nummer+1, fehler)
		}
	}

	return nil
}

// Die beiden Anweisungen für Feldwerte, ausgeschrieben.
//
// Nicht über `fmt.Sprintf` mit Tabellen- und Spaltennamen: Bezeichner lassen
// sich in SQL nicht binden, und ein zusammengesetzter Anweisungstext ist von
// einer Einschleusung nicht zu unterscheiden — gosec meldet ihn zu Recht.
// Zwei Konstanten sind länger und sagen dafür genau, was ausgeführt wird.
const (
	einfuegenSetwert    = `INSERT INTO set_values (set_id, feld, wert, reihenfolge) VALUES (?, ?, ?, ?)`
	einfuegenKartenwert = `INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES (?, ?, ?, ?)`
)

// werteSchreiben legt Feldwerte ab und behält ihre Reihenfolge.
func werteSchreiben(
	ctx context.Context, tx *sql.Tx, anweisung string, kennung int64, werte []domain.Value,
) error {
	for nummer, wert := range werte {
		if _, err := tx.ExecContext(ctx, anweisung,
			kennung, wert.Field, wert.Text, nummer); err != nil {
			return uebersetze(err, "Feld %q schreiben", wert.Field)
		}
	}

	return nil
}

func rohSchreiben(ctx context.Context, tx *sql.Tx, setID int64, roh []domain.RawBlock) error {
	for nummer, block := range roh {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO raw_blocks (set_id, datei, zeile, inhalt, reihenfolge)
			 VALUES (?, ?, ?, ?, ?)`,
			setID, block.File, block.Line, block.Text, nummer); err != nil {
			return uebersetze(err, "Rohblock aus Zeile %d schreiben", block.Line)
		}
	}

	return nil
}
