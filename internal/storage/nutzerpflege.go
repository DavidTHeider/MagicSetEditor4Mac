// **Nutzer und Profile ändern und löschen** — Werkstück 5 von Phase 4.12
// (#1351).
//
// # Zwei Regeln, und beide sind Zusagen und keine Bequemlichkeiten
//
// **Der letzte Nutzer bleibt.** Sonst stünde die Anwendung ohne Profil da, und
// jede Abfrage gäbe eine leere Liste — der Bestand sähe aus, als wäre er weg.
//
// **Und das letzte Profil eines Nutzers bleibt.** Ein Nutzer ohne
// Arbeitsumgebung könnte sich anmelden und nichts tun; genau deshalb legt
// [Nutzerablage.NutzerAnlegen] eines mit an.
//
// # Was hier **noch nicht** steht
//
// **Keine Fassung.** `decks`, `formate` und `sammlung` tragen eine und prüfen
// sie beim Schreiben (§7, Regel 2 über die Flächen hinweg); `nutzer` und
// `profile` nicht. Das ist eine **offene Stelle**, keine Entscheidung: wer
// gleichzeitig an zwei Flächen einen Namen ändert, verliert eine der beiden
// Änderungen, und niemand merkt es. Werkstück 8 gibt dem Konto siebzehn
// Felder — dort gehört sie hin, und zweimal zu wandern wäre die teurere Hälfte.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Erstnutzer ist der Name, den die Wanderung dem ersten Nutzer gibt.
//
// **Als Konstante und nicht als Zeichenkette im Test**: die Wanderung setzt
// ihn, und ein Test, der ihn noch einmal hinschreibt, prüft, dass zwei Kopien
// gleich sind.
const Erstnutzer = "Ich"

// NutzerAendern schreibt Name und Bild eines Nutzers fort.
func (a *Nutzerablage) NutzerAendern(ctx context.Context, geaendert domain.Nutzer) (
	domain.Nutzer, error,
) {
	if err := geaendert.Validate(); err != nil {
		return domain.Nutzer{}, err
	}

	// **Die Fassung steht im `WHERE` und nicht davor** — Werkstück 8 (#1351).
	//
	// **Weil ein Lesen, dann Entscheiden, dann Schreiben zwei gleichzeitige
	// Änderungen durchließe**: beide läsen dieselbe Fassung, beide fänden sie
	// gut, und die zweite überschriebe die erste. So entscheidet der Bestand,
	// und die zweite trifft null Zeilen (§7, Regel 2 über die Flächen hinweg).
	ergebnis, err := a.db.ExecContext(ctx,
		`UPDATE nutzer SET name = ?, bild = ?, anrede = ?, namenszusatz = ?,
			vorname = ?, weiterenamen = ?, nachname = ?, slogan = ?,
			kurzebio = ?, langebio = ?, fassung = fassung + 1
		 WHERE id = ? AND fassung = ?`,
		geaendert.Name, geaendert.Bild, geaendert.Anrede, geaendert.Namenszusatz,
		geaendert.Vorname, geaendert.Weiterenamen, geaendert.Nachname,
		geaendert.Slogan, geaendert.Kurzebio, geaendert.Langebio,
		geaendert.ID, geaendert.Fassung)
	if err != nil {
		return domain.Nutzer{}, uebersetze(err, "den Nutzer %d ändern", geaendert.ID)
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return domain.Nutzer{}, zaehlfehler
	}

	if zeilen == 0 {
		return domain.Nutzer{}, a.nutzerNichtGetroffen(ctx, geaendert.ID)
	}

	geaendert.Fassung++

	return geaendert, nil
}

// nutzerNichtGetroffen sagt, **warum** ein Schreiben null Zeilen traf.
//
// # Warum das eine zweite Abfrage kostet
//
// **Weil „gibt es nicht" und „hat sich geändert" zwei Antworten sind.** Die
// erste heißt: du hast den falschen Nutzer. Die zweite heißt: lies neu und
// versuch es noch einmal — und die Oberfläche tut dafür etwas anderes (§8).
//
// **Und weil das `WHERE` beides zugleich ausschließt**: nach einem `UPDATE …
// WHERE id = ? AND fassung = ?` mit null Zeilen weiß niemand, welche der
// beiden Bedingungen fehlschlug.
func (a *Nutzerablage) nutzerNichtGetroffen(ctx context.Context, nutzerID int64) error {
	var zahl int

	if err := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM nutzer WHERE id = ?`, nutzerID).Scan(&zahl); err != nil {
		return uebersetze(err, "den Nutzer %d nachschlagen", nutzerID)
	}

	if zahl == 0 {
		return fmt.Errorf("%w: den Nutzer %d gibt es nicht",
			domain.ErrNotFound, nutzerID)
	}

	return fmt.Errorf("%w — der Nutzer %d hat sich seit dem Lesen geändert",
		domain.ErrUeberholt, nutzerID)
}

// ProfilAendern schreibt den Namen eines Profils fort.
func (a *Nutzerablage) ProfilAendern(ctx context.Context, geaendert domain.Profil) (
	domain.Profil, error,
) {
	if err := geaendert.Validate(); err != nil {
		return domain.Profil{}, err
	}

	// **Mit Fassung, wie der Nutzer darüber** — Werkstück 8 (#1351).
	ergebnis, err := a.db.ExecContext(ctx,
		`UPDATE profile SET name = ?, subslogan = ?, kurzebeschreibung = ?,
			langebeschreibung = ?, badge = ?, technisch = ?, ablauf = ?,
			sichtbarkeit = ?, fassung = fassung + 1
		 WHERE id = ? AND fassung = ?`,
		geaendert.Name, geaendert.Subslogan, geaendert.Kurzebeschreibung,
		geaendert.Langebeschreibung, geaendert.Badge,
		geaendert.Technisch, geaendert.Ablauf,
		sichtbarOder(geaendert.Sichtbarkeit),
		geaendert.ID, geaendert.Fassung)
	if err != nil {
		return domain.Profil{}, uebersetze(err, "das Profil %d ändern", geaendert.ID)
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return domain.Profil{}, zaehlfehler
	}

	if zeilen == 0 {
		return domain.Profil{}, a.profilNichtGetroffen(ctx, geaendert.ID)
	}

	geaendert.Fassung++

	return geaendert, nil
}

// profilNichtGetroffen sagt, **warum** ein Schreiben null Zeilen traf —
// dieselbe Frage wie bei [Nutzerablage.nutzerNichtGetroffen].
func (a *Nutzerablage) profilNichtGetroffen(ctx context.Context, profilID int64) error {
	var zahl int

	if err := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM profile WHERE id = ?`, profilID).Scan(&zahl); err != nil {
		return uebersetze(err, "das Profil %d nachschlagen", profilID)
	}

	if zahl == 0 {
		return fmt.Errorf("%w: das Profil %d gibt es nicht",
			domain.ErrNotFound, profilID)
	}

	return fmt.Errorf("%w — das Profil %d hat sich seit dem Lesen geändert",
		domain.ErrUeberholt, profilID)
}

// NutzerLoeschen nimmt einen Nutzer samt seinen Profilen heraus.
//
// **Der Name bestätigt** — dieselbe Form wie beim Löschen eines Sets (#1218):
// eine Kennung allein ist eine Zahl, und wer sich vertippt, löscht den
// falschen Menschen.
//
// **Die Profile gehen mit** (`ON DELETE CASCADE`, Wanderung 0034). Was **in**
// ihnen liegt, geht **nicht** mit — dazu steht die Erklärung an
// [Nutzerablage.ProfilLoeschen].
func (a *Nutzerablage) NutzerLoeschen(
	ctx context.Context, nutzerID int64, name string,
	bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error,
) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "das Löschen beginnen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach dem Commit ist sie zu

	if pruefung := namePruefen(ctx, tx,
		`SELECT name FROM nutzer WHERE id = ?`, nutzerID, name,
		"den Nutzer"); pruefung != nil {
		return pruefung
	}

	// **Mit seinen Zugängen an fremden Profilen** (#2287) — sie gehen mit
	// ihm, und die Vorschau zählt sie.
	fremd, zugangsfehler := FremdeZugaenge(ctx, tx, nutzerID)
	if zugangsfehler != nil {
		return zugangsfehler
	}

	if gebunden := gebundenerUmfang(ctx, tx,
		`SELECT id, name FROM profile WHERE nutzer_id = ? ORDER BY id`,
		nutzerID, bindung, fremd); gebunden != nil {
		return gebunden
	}

	// **Der letzte bleibt.** Gezählt **in** der Transaktion, sonst zählte er
	// einen Stand, den ein zweiter Schreiber gerade ändert.
	var wieviele int64

	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM nutzer`).Scan(&wieviele); err != nil {
		return uebersetze(err, "die Nutzer zählen")
	}

	if wieviele <= 1 {
		return fmt.Errorf(
			"%w: der letzte Nutzer bleibt — ohne ihn hätte die Anwendung kein Profil",
			domain.ErrConflict)
	}

	if _, loeschfehler := tx.ExecContext(ctx,
		`DELETE FROM nutzer WHERE id = ?`, nutzerID); loeschfehler != nil {
		return uebersetze(loeschfehler, "den Nutzer %d löschen", nutzerID)
	}

	return uebersetze(tx.Commit(), "das Löschen abschließen")
}

// ProfilLoeschen nimmt ein Profil heraus.
//
// # Was mit seinem Inhalt geschieht
//
// **Nichts — und das ist Absicht.** Die zehn Wurzeln tragen eine `profil_id`
// ohne Fremdschlüssel (Wanderung 0035), und ein `ON DELETE CASCADE` darauf
// hieße: ein Klick löscht Sets, Decks, Sammlung und Turniere eines ganzen
// Profils. Für eine Anwendung, die die Kartensammlung eines Menschen hält, ist
// das die falsche Vorgabe.
//
// **Was bleibt, ist unsichtbar und nicht weg.** Wer das Profil neu anlegt,
// bekommt eine neue Kennung und findet es nicht wieder — das gehört gesagt,
// und Werkstück 9 gibt dem Profil die Einstellungen, an denen es hängt.
func (a *Nutzerablage) ProfilLoeschen(
	ctx context.Context, profilID int64, name string, ausweich int64,
	bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error,
) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "das Löschen beginnen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach dem Commit ist sie zu

	if pruefung := namePruefen(ctx, tx,
		`SELECT name FROM profile WHERE id = ?`, profilID, name,
		"das Profil"); pruefung != nil {
		return pruefung
	}

	if gebunden := gebundenerUmfang(ctx, tx,
		`SELECT id, name FROM profile WHERE id = ?`,
		profilID, bindung); gebunden != nil {
		return gebunden
	}

	// **Das letzte Profil seines Nutzers bleibt.**
	var wieviele int64

	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM profile
		  WHERE nutzer_id = (SELECT nutzer_id FROM profile WHERE id = ?)`,
		profilID).Scan(&wieviele); err != nil {
		return uebersetze(err, "die Profile zählen")
	}

	if wieviele <= 1 {
		return fmt.Errorf(
			"%w: das letzte Profil bleibt — ohne es hätte dieser Nutzer keine "+
				"Arbeitsumgebung", domain.ErrConflict)
	}

	// **Der Zeiger zieht in derselben Transaktion mit** — Befund 1 des
	// Zweitblicks über die Phasen 4.0 bis 4.13 (2026-09-07).
	//
	// Der Dienst stellte ihn **vorher** um, mit gutem Grund: nach dem Löschen
	// zeigte er sonst ins Leere. Nur wurde er damit auch dann bewegt, wenn
	// die Namensprüfung oder das „letzte Profil bleibt" gleich darauf ablehnt
	// — ein abgewiesenes Löschen ließ den Menschen in einem anderen Profil
	// zurück, ohne dass er etwas gelöscht hätte.
	//
	// **Hier ist beides ein Zug**: entweder das Profil ist fort und der Zeiger
	// steht auf dem Ausweich, oder nichts hat sich bewegt.
	if ausweich != 0 {
		if _, setzfehler := tx.ExecContext(ctx,
			`INSERT INTO bestand (schluessel, wert) VALUES (?, ?)
			 ON CONFLICT (schluessel) DO UPDATE SET wert = excluded.wert`,
			Profilzeigerschluessel, strconv.FormatInt(ausweich, 10)); setzfehler != nil {
			return uebersetze(setzfehler, "das aktive Profil setzen")
		}
	}

	if _, loeschfehler := tx.ExecContext(ctx,
		`DELETE FROM profile WHERE id = ?`, profilID); loeschfehler != nil {
		return uebersetze(loeschfehler, "das Profil %d löschen", profilID)
	}

	return uebersetze(tx.Commit(), "das Löschen abschließen")
}

// namePruefen sagt, ob der genannte Name zu dieser Kennung gehört.
//
// **Die Bestätigung und die Frage „gibt es das überhaupt" in einem Zug**: wer
// sie trennte, hätte zwei Stände und eine Meldung, die zum falschen gehört.
func namePruefen(
	ctx context.Context, tx *sql.Tx, abfrage string, kennung int64, name, was string,
) error {
	var vorhanden string

	err := tx.QueryRowContext(ctx, abfrage, kennung).Scan(&vorhanden)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("%w: %s %d gibt es nicht", domain.ErrNotFound, was, kennung)

	case err != nil:
		return uebersetze(err, "%s %d nachschlagen", was, kennung)
	}

	if vorhanden != name {
		return fmt.Errorf(
			"%w: %s %d heißt %q und nicht %q — der Name bestätigt das Löschen",
			domain.ErrValidation, was, kennung, vorhanden, name)
	}

	return nil
}
