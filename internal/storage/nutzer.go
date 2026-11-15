// **Nutzer und Profile im Bestand** — Werkstück 1 von Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Was hier **nicht** steht
//
// **Keine Prüfung fachlicher Art** (§2): der Speicher setzt um, was entschieden
// ist. Was der Bestand selbst hält, steht als `CHECK` in der Wanderung — und
// das ist die zweite Reihe, nicht die erste.
//
// **Und keine Trennung.** Dass ein Profil die Daten eines anderen nicht sieht,
// ist Werkstück 2.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nutzerablage führt Nutzer und ihre Profile.
type Nutzerablage struct{ db *sql.DB }

// NeueNutzerablage baut sie über einer Datenbank.
func NeueNutzerablage(db *sql.DB) *Nutzerablage { return &Nutzerablage{db: db} }

// Nutzer gibt alle Nutzer, nach Namen geordnet.
//
// **Nach Namen und nicht nach Kennung**: die Profilwahl zeigt sie, und dort
// ist die Reihenfolge des Anlegens keine, die jemand erwartet.
func (a *Nutzerablage) Nutzer(ctx context.Context) ([]domain.Nutzer, error) {
	zeilen, err := a.db.QueryContext(ctx,
		nutzerspalten+` FROM nutzer ORDER BY name`)
	if err != nil {
		return nil, uebersetze(err, "die Nutzer lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur gelesen

	var gelesen []domain.Nutzer

	for zeilen.Next() {
		var einer domain.Nutzer

		if err := scanneNutzer(zeilen, &einer); err != nil {
			return nil, err
		}

		gelesen = append(gelesen, einer)
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Nutzer lesen")
	}

	return gelesen, nil
}

// Profile gibt die Profile eines Nutzers, nach Namen geordnet.
func (a *Nutzerablage) Profile(ctx context.Context, nutzerID int64) (
	[]domain.Profil, error,
) {
	return a.profileLesen(ctx, profileDesNutzers, nutzerID)
}

// ProfileDesBesitzersVon gibt die Profile des Nutzers, dem dieses Profil
// gehört — das Profil selbst eingeschlossen, nach Namen geordnet.
//
// **Für eine Frage, die nur das Profil kennt, in dem ein Ruf arbeitet**
// (UXR-F21, #2041):
// Kommandozeile, MCP und Dienst tragen ihr Profil im Kontext und haben keinen
// „arbeitenden Nutzer", den ein Zeiger nennte. **Ein Profil, das es nicht
// gibt, hat keinen Besitzer** — dann ist die Liste leer und kein Fehler.
func (a *Nutzerablage) ProfileDesBesitzersVon(ctx context.Context, profilID int64) (
	[]domain.Profil, error,
) {
	return a.profileLesen(ctx, profileDesBesitzers, profilID)
}

// profileDesNutzers und profileDesBesitzers sind die beiden Abfragen — als
// Konstanten, damit der Lesepfad darunter einer ist und seine Fehlerwege
// einmal geprüft werden.
const (
	profileDesNutzers = `SELECT ` + profilspalten + ` FROM profile
		 WHERE nutzer_id = ? ORDER BY name`

	profileDesBesitzers = `SELECT ` + profilspalten + ` FROM profile
		 WHERE nutzer_id = (SELECT nutzer_id FROM profile WHERE id = ?)
		 ORDER BY name`
)

// profileLesen führt eine der beiden Abfragen aus und liest die Profile.
func (a *Nutzerablage) profileLesen(
	ctx context.Context, abfrage string, kennung int64,
) ([]domain.Profil, error) {
	zeilen, err := a.db.QueryContext(ctx, abfrage, kennung)
	if err != nil {
		return nil, uebersetze(err, "die Profile lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur gelesen

	var gelesen []domain.Profil

	for zeilen.Next() {
		var eines domain.Profil

		if err := scanneProfil(zeilen, &eines); err != nil {
			return nil, uebersetze(err, "eine Profilzeile lesen")
		}

		gelesen = append(gelesen, eines)
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Profile lesen")
	}

	return gelesen, nil
}

// Erstprofil ist der Name, den ein neuer Nutzer für sein erstes Profil bekommt.
//
// **Er steht hier und nicht in der Wanderung**, obwohl die Wanderung denselben
// Namen einträgt: dort ist es eine einmalige Zeile, hier eine Regel. Wer sie
// ändert, ändert nur, wie **neue** Nutzer heißen — die vorhandenen behalten,
// was sie haben.
const Erstprofil = "Mein Profil"

// NutzerAnlegen legt einen Nutzer **samt seinem ersten Profil** an.
//
// # Warum beides in einem Zug
//
// **Ein Nutzer ohne Profil hätte keine Arbeitsumgebung** — er könnte sich
// anmelden und nichts tun. Diesen Zustand gar nicht erst entstehen zu lassen
// ist billiger, als ihn überall zu behandeln.
//
// **Und deshalb in einer Transaktion**: ein Abbruch dazwischen ließe genau den
// Nutzer zurück, den es nicht geben soll.
func (a *Nutzerablage) NutzerAnlegen(ctx context.Context, neu domain.Nutzer) (
	domain.Nutzer, error,
) {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Nutzer{}, uebersetze(err, "den Nutzer anlegen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach Commit wirkungslos

	// **Mit `RETURNING`**: die Antwort trägt Kennung **und Fassung**, wie der
	// Bestand sie schrieb — bis zum Vertragsrundgang (B01) kam die Fassung als
	// `0` zurück, der Bestand führt `1`, und ein Ändern mit dieser Antwort
	// scheiterte an „Stand überholt".
	var kennung int64
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO nutzer (name, bild, email, zweitemail, telefon, kennwort, anonym)
		 VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id, fassung`,
		neu.Name, neu.Bild, neu.Email, neu.Zweitemail, neu.Telefon,
		neu.Kennwort, neu.Anonym).Scan(&kennung, &neu.Fassung); err != nil {
		return domain.Nutzer{}, uebersetze(err, "den Nutzer %q anlegen", neu.Name)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO profile (nutzer_id, name) VALUES (?, ?)`,
		kennung, Erstprofil); err != nil {
		return domain.Nutzer{}, uebersetze(err, "das erste Profil anlegen")
	}

	if err := tx.Commit(); err != nil {
		return domain.Nutzer{}, uebersetze(err, "den Nutzer anlegen")
	}

	neu.ID = kennung

	return neu, nil
}

// ProfilAnlegen legt ein weiteres Profil für einen Nutzer an.
//
// **Den Nutzer gibt es, oder es ist ein Nein.** SQLite meldete sonst nur einen
// Fremdschlüsselbruch, und wer das liest, sucht an der falschen Stelle (§8).
func (a *Nutzerablage) ProfilAnlegen(ctx context.Context, neu domain.Profil) (
	domain.Profil, error,
) {
	var da int

	if err := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM nutzer WHERE id = ?`, neu.NutzerID).Scan(&da); err != nil {
		return domain.Profil{}, uebersetze(err, "den Nutzer suchen")
	}

	if da == 0 {
		return domain.Profil{}, fmt.Errorf(
			"%w: den Nutzer %d führt der Bestand nicht", domain.ErrNotFound, neu.NutzerID)
	}

	// **Alle Angaben und nicht nur der Name** — Werkstück 10 (#1351).
	//
	// Bis dahin schrieb dieses `INSERT` `nutzer_id` und `name`, gab aber `neu`
	// **unverändert** zurück: wer eine Kurzbeschreibung mitgab, bekam sie in
	// der Antwort zurück und fand sie beim nächsten Lesen leer. Ein
	// Rückgabewert, der mehr behauptet als der Bestand trägt, ist ein Kopf,
	// der mehr verspricht als sein Rumpf hält (§10).
	//
	// **Und die Fassung dazu** (`RETURNING`) — derselbe Befund wie beim Nutzer.
	if err := a.db.QueryRowContext(ctx,
		`INSERT INTO profile (nutzer_id, name, subslogan, kurzebeschreibung,
			langebeschreibung, badge, technisch, ablauf, sichtbarkeit)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id, fassung, sichtbarkeit`,
		neu.NutzerID, neu.Name, neu.Subslogan, neu.Kurzebeschreibung,
		neu.Langebeschreibung, neu.Badge, neu.Technisch, neu.Ablauf,
		sichtbarOder(neu.Sichtbarkeit)).Scan(&neu.ID, &neu.Fassung, &neu.Sichtbarkeit); err != nil {
		return domain.Profil{}, uebersetze(err, "das Profil %q anlegen", neu.Name)
	}

	return neu, nil
}

// Profilzeigerschluessel ist der Eintrag im Bestand, der auf das aktive
// Profil zeigt.
//
// **Im Bestand und nicht im Arbeitsspeicher**: wer sein Profil wählt und die
// Anwendung schließt, findet es beim nächsten Mal wieder. Sonst wäre die Wahl
// eine Frage, die jeden Morgen neu gestellt wird.
const Profilzeigerschluessel = "profil_zeiger"

// Profilzeiger sagt, welches Profil gerade arbeitet.
//
// # Er steht nie auf nichts
//
// **Ein Zeiger auf ein Profil, das es nicht gibt, ließe jede Abfrage leer
// zurückkommen** — eine Anwendung, die aussieht, als wären alle Daten weg.
// Deshalb prüft diese Abfrage die Zugehörigkeit mit und fällt sonst auf das
// **erste vorhandene** Profil zurück.
//
// Das ist kein Beiwerk: die Kaskade räumt die Profile eines gelöschten
// Nutzers, und der Zeiger stünde danach auf einer Kennung, die es nicht mehr
// gibt.
func (a *Nutzerablage) Profilzeiger(ctx context.Context) (int64, error) {
	var gezeigt int64

	// **Ein Zug über beide Fragen.** Die gespeicherte Kennung gilt nur, wenn
	// es das Profil noch gibt; sonst nimmt `coalesce` das erste. Zwei
	// Abfragen gäben zwei Stände, zwischen denen sich etwas ändern kann.
	err := a.db.QueryRowContext(ctx,
		`SELECT coalesce(
		          (SELECT p.id FROM profile p
		            JOIN bestand b ON b.wert = CAST(p.id AS TEXT)
		           WHERE b.schluessel = ?),
		          (SELECT id FROM profile ORDER BY id LIMIT 1),
		          0)`, Profilzeigerschluessel).Scan(&gezeigt)
	if err != nil {
		return 0, uebersetze(err, "das aktive Profil lesen")
	}

	return gezeigt, nil
}

// ProfilzeigerSetzen stellt das aktive Profil um.
//
// **Das Profil gibt es, oder es ist ein Nein.** Ein Zeiger ins Leere ist der
// Zustand, den [Nutzerablage.Profilzeiger] gerade abfängt — ihn erst
// herzustellen und dann abzufangen wäre zweimal Arbeit für ein Nein, das hier
// billiger ist (§8).
func (a *Nutzerablage) ProfilzeigerSetzen(ctx context.Context, profilID int64) error {
	var da int

	if err := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM profile WHERE id = ?`, profilID).Scan(&da); err != nil {
		return uebersetze(err, "das Profil suchen")
	}

	if da == 0 {
		return fmt.Errorf(
			"%w: das Profil %d führt der Bestand nicht", domain.ErrNotFound, profilID)
	}

	if _, err := a.db.ExecContext(ctx,
		`INSERT INTO bestand (schluessel, wert) VALUES (?, ?)
		 ON CONFLICT (schluessel) DO UPDATE SET wert = excluded.wert`,
		Profilzeigerschluessel, strconv.FormatInt(profilID, 10)); err != nil {
		return uebersetze(err, "das aktive Profil setzen")
	}

	return nil
}

// nutzerspalten sind die Spalten, die ein [domain.Nutzer] trägt.
//
// **An einer Stelle und nicht in jeder Abfrage**: sie sind seit Werkstück 6
// zu siebt, und Werkstück 8 legt weitere daneben. Eine Aufzählung je Abfrage
// liefe beim ersten Feld auseinander, und die Reihenfolge müsste dreimal
// stimmen.
const nutzerspalten = `SELECT id, name, bild, email, zweitemail, telefon,
	kennwort, anonym, anrede, namenszusatz, vorname, weiterenamen, nachname,
	slogan, kurzebio, langebio, fassung`

// scanneNutzer liest eine Zeile in die Entität.
type zeilenleser interface{ Scan(ziele ...any) error }

func scanneNutzer(zeile zeilenleser, einer *domain.Nutzer) error {
	if err := zeile.Scan(&einer.ID, &einer.Name, &einer.Bild, &einer.Email,
		&einer.Zweitemail, &einer.Telefon, &einer.Kennwort, &einer.Anonym,
		&einer.Anrede, &einer.Namenszusatz, &einer.Vorname, &einer.Weiterenamen,
		&einer.Nachname, &einer.Slogan, &einer.Kurzebio, &einer.Langebio,
		&einer.Fassung); err != nil {
		return uebersetze(err, "eine Nutzerzeile lesen")
	}

	return nil
}

// NutzerZuKennung gibt den Nutzer zu einer Adresse oder einer Rufnummer.
//
// # Eine Abfrage für beides
//
// **Weil beide dasselbe tun**: sie sagen, wer sich anmeldet. Wer daraus zwei
// Wege machte, hätte zwei Fehlermeldungen für denselben Fall und müsste an der
// Fläche raten, welche gemeint ist.
//
// **Eine leere Kennung findet niemanden**, und das ist die halbe Zusage: ohne
// die Bedingung fände sie **jeden** Nutzer ohne Adresse — und wer sich mit
// einem leeren Feld anmeldete, käme als irgendjemand herein.
func (a *Nutzerablage) NutzerZuKennung(ctx context.Context, kennung string) (
	domain.Nutzer, error,
) {
	if strings.TrimSpace(kennung) == "" {
		return domain.Nutzer{}, fmt.Errorf(
			"%w: eine leere Kennung gehört niemandem", domain.ErrNotFound)
	}

	var einer domain.Nutzer

	zeile := a.db.QueryRowContext(ctx,
		nutzerspalten+` FROM nutzer WHERE email = ? OR telefon = ?`, kennung, kennung)

	if err := scanneNutzer(zeile, &einer); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Nutzer{}, fmt.Errorf(
				"%w: zu dieser Kennung gibt es keinen Nutzer", domain.ErrNotFound)
		}

		return domain.Nutzer{}, err
	}

	return einer, nil
}

// profilspalten sind die Spalten, die [domain.Profil] füllt.
//
// **Als Konstante und nicht dreimal geschrieben** — dieselbe Bauform wie
// `nutzerspalten`: die Reihenfolge muss zu [scanneProfil] passen, und zwei
// Aufzählungen liefen beim nächsten Feld auseinander.
const profilspalten = `id, nutzer_id, name, fassung, subslogan,
	kurzebeschreibung, langebeschreibung, badge, technisch, ablauf,
	sichtbarkeit`

// scanneProfil liest eine Zeile in ein Profil.
func scanneProfil(zeilen interface{ Scan(...any) error }, eines *domain.Profil) error {
	return zeilen.Scan(&eines.ID, &eines.NutzerID, &eines.Name, &eines.Fassung,
		&eines.Subslogan, &eines.Kurzebeschreibung, &eines.Langebeschreibung,
		&eines.Badge, &eines.Technisch, &eines.Ablauf, &eines.Sichtbarkeit)
}

// sichtbarOder gibt die Sichtbarkeit — oder die vorsichtige Vorgabe.
//
// **Leer heißt `privat`, und das steht hier statt in der Spalte.** Die
// Wanderung setzt denselben Wert als `DEFAULT`; wer ein Profil mit leerer
// Sichtbarkeit anlegt, meint aber nicht „nimm die Vorgabe der Spalte", sondern
// „ich lege mich nicht fest" — und die Antwort darauf ist dieselbe.
func sichtbarOder(gewaehlt string) string {
	if gewaehlt == "" {
		return string(domain.SichtbarPrivat)
	}

	return gewaehlt
}
