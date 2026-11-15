package storage

// **Wen dieses Profil kennt** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum eine Begegnung den Stand nicht anfasst
//
// **Weil sie nichts darüber weiß.** Ein Gegenüber, dem jemand einmal einen
// Fingerabdruck vorgelesen hat, bleibt bestätigt — auch wenn es sich beim
// nächsten Hallo anders nennt. Ein Einfügen mit Ersetzen setzte den Stand
// zurück, und jede Freigabe wäre nach dem nächsten Hallo fort, ohne dass
// jemand etwas entzogen hätte.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kontaktablage hält die Kontakte eines Profils.
type Kontaktablage struct{ db *sql.DB }

// NeueKontaktablage baut sie über einer Datenbank.
func NeueKontaktablage(db *sql.DB) *Kontaktablage {
	return &Kontaktablage{db: db}
}

// Liste gibt alle Kontakte dieses Profils, samt ihren Rechten.
func (k *Kontaktablage) Liste(
	ctx context.Context,
) ([]domain.Peerkontakt, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	zeilen, err := k.db.QueryContext(ctx, `
SELECT id, kennung, fingerabdruck, name, eigener_name, stand, gesehen,
       adresse, erstmals, notiz, vorgaenger
  FROM peerkontakte
 WHERE profil_id = ?
 ORDER BY kennung`, profilID)
	if err != nil {
		return nil, uebersetze(err, "die Kontakte lesen")
	}

	defer func() { _ = zeilen.Close() }()

	liste := make([]domain.Peerkontakt, 0)
	kennungen := make([]int64, 0)

	for zeilen.Next() {
		kontakt, id, fehler := lesKontakt(zeilen)
		if fehler != nil {
			return nil, fehler
		}

		liste = append(liste, kontakt)
		kennungen = append(kennungen, id)
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Kontakte lesen")
	}

	for i, id := range kennungen {
		rechte, fehler := k.rechte(ctx, id)
		if fehler != nil {
			return nil, fehler
		}

		liste[i].Rechte = rechte
	}

	return liste, nil
}

// Lesen gibt einen Kontakt dieses Profils.
func (k *Kontaktablage) Lesen(
	ctx context.Context, kennung string,
) (domain.Peerkontakt, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return domain.Peerkontakt{}, fehler
	}

	kontakt, id, fehler := lesKontakt(k.db.QueryRowContext(ctx, `
SELECT id, kennung, fingerabdruck, name, eigener_name, stand, gesehen,
       adresse, erstmals, notiz, vorgaenger
  FROM peerkontakte
 WHERE profil_id = ? AND kennung = ?`, profilID, kennung))
	if fehler != nil {
		return domain.Peerkontakt{}, fehler
	}

	rechte, fehler := k.rechte(ctx, id)
	if fehler != nil {
		return domain.Peerkontakt{}, fehler
	}

	kontakt.Rechte = rechte

	return kontakt, nil
}

// Begegnet trägt ein Gegenüber ein oder frischt seine Angaben auf.
//
// # Warum hier vier Werte stehen und keine Entität
//
// **Weil eine Begegnung weder Stand noch Rechte kennt.** Eine Entität
// entgegenzunehmen und zwei ihrer Felder stillschweigend fallen zu lassen
// wäre ein Kopf, der mehr verspricht als sein Rumpf hält (§10) — und der
// Aufrufer sähe nie, dass sein `Stand` nicht angekommen ist.
//
// **Neu ist ein Kontakt immer unbestätigt.** Was er darf, entscheidet dieses
// Profil, nicht die Nachricht, die gerade angekommen ist.
func (k *Kontaktablage) Begegnet(
	ctx context.Context, kennung, fingerabdruck, name string, jetzt time.Time,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	if fehler := (domain.Peerkontakt{
		Kennung: kennung, Stand: domain.KontaktUnbestaetigt,
	}).Validate(); fehler != nil {
		return fehler
	}

	// **Der Stand steht nicht im `DO UPDATE`.** Wer einmal bestätigt hat,
	// bleibt bestätigt; ein Einfügen mit Ersetzen setzte ihn zurück, und
	// jede Freigabe wäre nach dem nächsten Hallo fort.
	//
	// **`erstmals` und `notiz` aus demselben Grund** (T3, #1468): die erste
	// Begegnung ist die erste, und eine Notiz gehört dem, der sie
	// geschrieben hat — beides wäre nach dem nächsten Hallo fort.
	_, err := k.db.ExecContext(ctx, `
INSERT INTO peerkontakte
       (profil_id, kennung, fingerabdruck, name, eigener_name, stand, gesehen,
        erstmals)
VALUES (?, ?, ?, ?, '', ?, ?, ?)
ON CONFLICT (profil_id, kennung) DO UPDATE
   SET fingerabdruck = excluded.fingerabdruck,
       name          = excluded.name,
       gesehen       = excluded.gesehen,
       fassung       = peerkontakte.fassung + 1`,
		profilID, kennung, fingerabdruck, name,
		string(domain.KontaktUnbestaetigt), jetzt.Format(time.RFC3339Nano),
		jetzt.Format(time.RFC3339Nano))
	if err != nil {
		return uebersetze(err, "einen Kontakt vermerken")
	}

	return nil
}

// Stand setzt den Stand eines Kontakts.
func (k *Kontaktablage) Stand(
	ctx context.Context, kennung string, stand domain.Kontaktstand,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	if fehler := (domain.Peerkontakt{
		Kennung: kennung, Stand: stand,
	}).Validate(); fehler != nil {
		return fehler
	}

	// **Fällt die Bestätigung, fallen die Rechte mit.** Ein Recht an einem
	// ungeprüften Gegenüber gibt es nicht ([domain.Peerkontakt.Validate]);
	// eine Zeile, die es doch gäbe, wäre eine, die niemand mehr sieht und
	// die bei der nächsten Bestätigung wieder gölte.
	return k.inTransaktion(ctx, func(tx *sql.Tx) error {
		id, _, fehler := kontaktZeile(ctx, tx, profilID, kennung)
		if fehler != nil {
			return fehler
		}

		if _, err := tx.ExecContext(ctx,
			`UPDATE peerkontakte
			    SET stand = ?, fassung = fassung + 1
			  WHERE id = ?`, string(stand), id); err != nil {
			return uebersetze(err, "den Stand eines Kontakts setzen")
		}

		if stand == domain.KontaktBestaetigt {
			return nil
		}

		if _, err := tx.ExecContext(ctx,
			`DELETE FROM peerrechte WHERE kontakt_id = ?`, id); err != nil {
			return uebersetze(err, "die Rechte eines Kontakts räumen")
		}

		return nil
	})
}

// AdresseSetzen trägt ein, wo ein Gegenüber erreichbar ist.
//
// # Warum leer hier eine Änderung ist und kein Auslassen
//
// **Weil „ich kenne keine mehr" eine Aussage ist.** Wer umgezogen ist,
// bekommt eine neue; wessen Adresse nicht mehr gilt, bekommt keine. Ein
// leerer Wert, der still übersprungen würde, ließe die alte stehen — und der
// nächste Versuch ginge an einen Rechner, der jemand anderem gehört.
func (k *Kontaktablage) AdresseSetzen(
	ctx context.Context, kennung, adresse string,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	if fehler := (domain.Peerkontakt{
		Kennung: kennung, Stand: domain.KontaktUnbestaetigt, Adresse: adresse,
	}).Validate(); fehler != nil {
		return fehler
	}

	return k.inTransaktion(ctx, func(tx *sql.Tx) error {
		id, _, fehler := kontaktZeile(ctx, tx, profilID, kennung)
		if fehler != nil {
			return fehler
		}

		if _, err := tx.ExecContext(ctx,
			`UPDATE peerkontakte
			    SET adresse = ?, fassung = fassung + 1
			  WHERE id = ?`, adresse, id); err != nil {
			return uebersetze(err, "die Adresse eines Kontakts setzen")
		}

		return nil
	})
}

// lesKontakt liest eine Zeile in die Entität.
//
// **Die Zeilenkennung kommt mit heraus.** Sie steht in keiner Entität — die
// Fachschicht kennt einen Kontakt an seiner Peerkennung —, aber die Rechte
// hängen an ihr, und ein zweiter Blick in die Tabelle wäre ein zweiter Stand.
func lesKontakt(zeile zeilenleser) (domain.Peerkontakt, int64, error) {
	var (
		kontakt  domain.Peerkontakt
		id       int64
		stand    string
		gesehen  string
		erstmals string
	)

	err := zeile.Scan(&id, &kontakt.Kennung, &kontakt.Fingerabdruck,
		&kontakt.SelbstgewaehlterName, &kontakt.EigenerName, &stand, &gesehen,
		&kontakt.Adresse, &erstmals, &kontakt.Notiz, &kontakt.Vorgaenger)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.Peerkontakt{}, 0, fmt.Errorf(
			"%w: diesem Profil ist dieses Gegenüber noch nicht begegnet",
			domain.ErrNotFound)
	case err != nil:
		return domain.Peerkontakt{}, 0, uebersetze(
			err, "eine Kontaktzeile lesen")
	}

	gestempelt, fehler := time.Parse(time.RFC3339Nano, gesehen)
	if fehler != nil {
		return domain.Peerkontakt{}, 0, domain.MitGrund(domain.ErrValidation,
			fehler, "den Zeitpunkt einer Begegnung lesen")
	}

	// **Leer heißt: nicht bekannt** (T3, #1468). Ein Kontakt aus der Zeit vor
	// dieser Spalte hat keine erste Begegnung, und ein geratenes Datum sähe
	// aus wie ein gemessenes (§8).
	erstes, erstfehler := ausFrist(erstmals)
	if erstfehler != nil {
		return domain.Peerkontakt{}, 0, erstfehler
	}

	kontakt.Stand = domain.Kontaktstand(stand)
	kontakt.Gesehen = gestempelt
	kontakt.Erstmals = erstes

	return kontakt, id, nil
}

// inTransaktion führt eine Änderung in einem Zug aus.
func (k *Kontaktablage) inTransaktion(
	ctx context.Context, aendern func(*sql.Tx) error,
) (fehler error) {
	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "einen Kontakt ändern")
	}

	defer func() { fehler = zurueckgenommen(tx, fehler) }()

	if fehler := aendern(tx); fehler != nil {
		return fehler
	}

	return bestaetige(tx, "einen Kontakt ändern")
}

// NotizSetzen schreibt, was dieses Profil sich zu einem Gegenüber notiert.
//
// **Eine leere löscht die vorhandene** — „ich habe nichts mehr zu sagen" ist
// eine Aussage, dieselbe Form wie bei der Adresse.
//
// **Und sie geht nie hinaus**: eine Sichtkarte trägt sie nicht, und eine
// Freigabe erst recht nicht (T3, #1468).
func (k *Kontaktablage) NotizSetzen(
	ctx context.Context, kennung, notiz string,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	return k.inTransaktion(ctx, func(tx *sql.Tx) error {
		id, _, fehler := kontaktZeile(ctx, tx, profilID, kennung)
		if fehler != nil {
			return fehler
		}

		if _, err := tx.ExecContext(ctx,
			`UPDATE peerkontakte
			    SET notiz = ?, fassung = fassung + 1
			  WHERE id = ?`, notiz, id); err != nil {
			return uebersetze(err, "die Notiz zu einem Kontakt setzen")
		}

		return nil
	})
}
