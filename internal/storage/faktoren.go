// **Der zweite Faktor im Bestand** — Werkstück 7 von Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Was hier liegt und was nicht
//
// **Hier liegt, DASS jemand einen Faktor hat.** Das TOTP-Geheimnis liegt im
// Schlüsselbund — die Bestandsdatei ist die, die man weitergibt, und wer das
// Geheimnis hat, macht jeden Code.
//
// **Die Wiederherstellungscodes liegen dagegen hier**, gehütet wie ein
// Kennwort. Sie sind die Antwort darauf, dass der Schlüsselbund eines anderen
// Macs das Geheimnis nicht kennt: wer seinen Bestand zurückspielt, käme sonst
// nicht mehr an sein eigenes Profil.
//
// # Ohne Profilklammer
//
// **Ein Faktor gehört einem Nutzer und keinem Profil.** Wer das Profil
// wechselt, bleibt derselbe Mensch — und die Klammer aus Wanderung 0035 säße
// hier falsch: sie ordnet **Inhalte** einem Profil zu, nicht Personen.

package storage

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Faktoren gibt die zweiten Faktoren eines Nutzers.
func (a *Nutzerablage) Faktoren(ctx context.Context, nutzerID int64) (
	[]domain.Faktor, error,
) {
	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT id, nutzer_id, art, bezeichnung FROM nutzerfaktoren
		 WHERE nutzer_id = ? ORDER BY art`, nutzerID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die zweiten Faktoren lesen")
	}

	defer func() { _ = zeilen.Close() }()

	// **Nie `nil`** (§3): eine leere Liste ist eine Antwort.
	gelesen := []domain.Faktor{}

	for zeilen.Next() {
		var einer domain.Faktor

		if err := zeilen.Scan(
			&einer.ID, &einer.NutzerID, &einer.Art, &einer.Bezeichnung); err != nil {
			return nil, uebersetze(err, "eine Faktorzeile lesen")
		}

		gelesen = append(gelesen, einer)
	}

	return gelesen, uebersetze(zeilen.Err(), "die Faktorzeilen lesen")
}

// FaktorAnlegen trägt einen zweiten Faktor ein — **mit** dem Schritt, der
// schon galt (SEC-09).
func (a *Nutzerablage) FaktorAnlegen(ctx context.Context, neu domain.Faktor) (
	domain.Faktor, error,
) {
	ergebnis, fehler := a.db.ExecContext(ctx,
		`INSERT INTO nutzerfaktoren (nutzer_id, art, bezeichnung, letzter_schritt)
		 VALUES (?, ?, ?, ?)`,
		neu.NutzerID, string(neu.Art), neu.Bezeichnung, neu.LetzterSchritt)
	if fehler != nil {
		return domain.Faktor{}, uebersetze(fehler, "den zweiten Faktor anlegen")
	}

	kennung, kennungsfehler := ergebnis.LastInsertId()
	if kennungsfehler != nil {
		return domain.Faktor{}, uebersetze(kennungsfehler, "die Kennung lesen")
	}

	neu.ID = kennung

	return neu, nil
}

// FaktorLoeschen nimmt einen zweiten Faktor heraus.
//
// **„Gibt es nicht" ist ein Fehler und kein stiller Erfolg** (§8): wer ihn
// löschen wollte und nichts traf, hat den falschen Nutzer erwischt — und eine
// leere Antwort sähe aus wie getan.
func (a *Nutzerablage) FaktorLoeschen(
	ctx context.Context, nutzerID int64, art domain.Faktorart,
) error {
	ergebnis, fehler := a.db.ExecContext(ctx,
		`DELETE FROM nutzerfaktoren WHERE nutzer_id = ? AND art = ?`,
		nutzerID, string(art))
	if fehler != nil {
		return uebersetze(fehler, "den zweiten Faktor löschen")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf("%w: dieser Nutzer hat keinen Faktor der Art %q",
			domain.ErrNotFound, string(art))
	}

	return nil
}

// CodesSetzen ersetzt die Wiederherstellungscodes eines Nutzers.
//
// **Ersetzen und nicht ergänzen**, und beides in **einer** Transaktion: wer
// neue holt, verwirft die alten. Zwei Sätze nebeneinander wären zwei Wege
// hinein, und der ältere hinge an einem Zettel, den niemand mehr sucht.
func (a *Nutzerablage) CodesSetzen(
	ctx context.Context, nutzerID int64, gehuetete []string,
) error {
	vorgang, fehler := a.db.BeginTx(ctx, nil)
	if fehler != nil {
		return uebersetze(fehler, "die Codes setzen")
	}

	defer func() { _ = vorgang.Rollback() }()

	if _, err := vorgang.ExecContext(ctx,
		`DELETE FROM wiederherstellungscodes WHERE nutzer_id = ?`,
		nutzerID); err != nil {
		return uebersetze(err, "die alten Codes entfernen")
	}

	for _, gehuetet := range gehuetete {
		if _, err := vorgang.ExecContext(ctx,
			`INSERT INTO wiederherstellungscodes (nutzer_id, gehuetet) VALUES (?, ?)`,
			nutzerID, gehuetet); err != nil {
			return uebersetze(err, "einen Code eintragen")
		}
	}

	return uebersetze(vorgang.Commit(), "die Codes festschreiben")
}

// FreieCodes gibt die Codes, die noch nicht gebraucht wurden.
func (a *Nutzerablage) FreieCodes(ctx context.Context, nutzerID int64) (
	[]domain.Wiederherstellungscode, error,
) {
	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT id, gehuetet FROM wiederherstellungscodes
		 WHERE nutzer_id = ? AND gebraucht_am = '' ORDER BY id`, nutzerID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Wiederherstellungscodes lesen")
	}

	defer func() { _ = zeilen.Close() }()

	gelesen := []domain.Wiederherstellungscode{}

	for zeilen.Next() {
		var einer domain.Wiederherstellungscode

		if err := zeilen.Scan(&einer.ID, &einer.Gehuetet); err != nil {
			return nil, uebersetze(err, "eine Codezeile lesen")
		}

		gelesen = append(gelesen, einer)
	}

	return gelesen, uebersetze(zeilen.Err(), "die Codezeilen lesen")
}

// CodeVerbrauchen stempelt einen Code ab.
//
// **Die Bedingung steht im `WHERE` und nicht davor.** Ein Lesen, dann
// Entscheiden, dann Schreiben ließe zwei gleichzeitige Anmeldungen denselben
// Code zweimal verbrauchen; so entscheidet der Bestand, und der zweite trifft
// null Zeilen (§7, Regel 4 auf der anderen Seite der Naht).
func (a *Nutzerablage) CodeVerbrauchen(ctx context.Context, codeID int64) error {
	ergebnis, fehler := a.db.ExecContext(ctx,
		`UPDATE wiederherstellungscodes SET gebraucht_am = datetime('now')
		 WHERE id = ? AND gebraucht_am = ''`, codeID)
	if fehler != nil {
		return uebersetze(fehler, "den Code abstempeln")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf(
			"%w: diesen Wiederherstellungscode gibt es nicht mehr", domain.ErrNotFound)
	}

	return nil
}
