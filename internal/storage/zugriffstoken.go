// **Zugriffstoken im Bestand** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// # Ohne Profilklammer
//
// **Ein Token hängt an einem Profil und gehört keinem.** Die Klammer aus
// Wanderung 0035 ordnet **Inhalte** dem gezeigten Profil zu; ein Token wird
// aber gerade dann gelesen, wenn **noch kein** Profil feststeht — es sagt ja
// erst, welches gemeint ist. Wer hier `ProfilAus` verlangte, hätte einen Ring
// ohne Anfang.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zugriffstokenablage führt die Token dieses Bestands.
type Zugriffstokenablage struct{ db *sql.DB }

// NeueZugriffstokenablage baut sie über einer Datenbank.
func NeueZugriffstokenablage(db *sql.DB) *Zugriffstokenablage {
	return &Zugriffstokenablage{db: db}
}

// tokenspalten sind die Spalten, die [domain.Zugriffstoken] füllt.
// **`#nosec G101`**: gosec sieht das Wort `token` in einem Konstantennamen
// und hält jede solche Zeile für ein eingebautes Geheimnis. Was hier steht,
// ist eine Spaltenliste — und **`gehuetet` steht ausdrücklich nicht darin**
// (siehe darüber).
const tokenspalten = `id, profil_id, bezeichnung, scope, ablauf, angelegt, fassung, principal, letzte_nutzung, vorgaenger_bis, sitz_partie, sitz` // #nosec G101 -- eine Spaltenliste

// TokenAnlegen legt ein Token an und gibt es samt seiner Kennung.
//
// **Die gehütete Form kommt von außen.** Wie ein Geheimnis gehütet wird,
// entscheidet `internal/service`; der Bestand legt ab, was er bekommt — und
// **nie** einen Klartext.
func (a *Zugriffstokenablage) TokenAnlegen(
	ctx context.Context, neu domain.Zugriffstoken, gehuetet string,
) (domain.Zugriffstoken, error) {
	if err := neu.Validate(); err != nil {
		return domain.Zugriffstoken{}, err
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`INSERT INTO zugriffstoken
			(profil_id, bezeichnung, gehuetet, scope, ablauf, angelegt, principal,
			 sitz_partie, sitz)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		neu.ProfilID, neu.Bezeichnung, gehuetet, string(neu.Scope),
		neu.Ablauf, neu.Angelegt, neu.Principal,
		neu.Sitzbindung.Partie, neu.Sitzbindung.Sitz)
	if err != nil {
		return domain.Zugriffstoken{}, uebersetze(err,
			"das Token %q anlegen", neu.Bezeichnung)
	}

	kennung, kennungsfehler := ergebnis.LastInsertId()
	if kennungsfehler != nil {
		return domain.Zugriffstoken{}, uebersetze(kennungsfehler,
			"die Kennung des Tokens lesen")
	}

	neu.ID = kennung
	neu.Fassung = 1

	return neu, nil
}

// Token gibt die Token eines Profils, nach Bezeichnung geordnet.
//
// **Ohne die gehütete Form.** Sie hat in einer Liste nichts zu suchen: was
// dort steht, wird angezeigt, und ein Geheimnis, das angezeigt wird, ist
// keines mehr.
func (a *Zugriffstokenablage) Token(ctx context.Context, profilID int64) (
	[]domain.Zugriffstoken, error,
) {
	zeilen, err := a.db.QueryContext(ctx,
		`SELECT `+tokenspalten+` FROM zugriffstoken
		  WHERE profil_id = ? ORDER BY bezeichnung`, profilID)
	if err != nil {
		return nil, uebersetze(err, "die Token lesen")
	}

	defer func() { _ = zeilen.Close() }()

	gelesen := []domain.Zugriffstoken{}

	for zeilen.Next() {
		eines, scanfehler := scanneToken(zeilen)
		if scanfehler != nil {
			return nil, uebersetze(scanfehler, "eine Tokenzeile lesen")
		}

		gelesen = append(gelesen, eines)
	}

	return gelesen, uebersetze(zeilen.Err(), "die Tokenzeilen lesen")
}

// TokenMitKennung gibt ein Token samt seiner gehüteten Form.
//
// **Die einzige Stelle, die sie herausgibt** — und sie heißt danach: wer sie
// ruft, will prüfen und nicht anzeigen.
func (a *Zugriffstokenablage) TokenMitKennung(ctx context.Context, tokenID int64) (
	domain.Zugriffstoken, domain.Gehuetetes, error,
) {
	var (
		eines    domain.Zugriffstoken
		scope    string
		gehuetet domain.Gehuetetes
	)

	// **Beide gehüteten Formen in einem Zug** — Werkstück O2 von Phase 5.2
	// (#1426): die geltende und die des Vorgängers aus einer Rotation. Ob
	// der Vorgänger noch zählt, entscheidet der Dienst an der Frist.
	fehler := a.db.QueryRowContext(ctx,
		`SELECT `+tokenspalten+`, gehuetet, vorgaenger_gehuetet
		   FROM zugriffstoken WHERE id = ?`,
		tokenID).Scan(&eines.ID, &eines.ProfilID, &eines.Bezeichnung, &scope,
		&eines.Ablauf, &eines.Angelegt, &eines.Fassung, &eines.Principal,
		&eines.LetzteNutzung, &eines.VorgaengerBis,
		&eines.Sitzbindung.Partie, &eines.Sitzbindung.Sitz,
		&gehuetet.Aktuell, &gehuetet.Vorgaenger)

	if errors.Is(fehler, sql.ErrNoRows) {
		return domain.Zugriffstoken{}, domain.Gehuetetes{}, fmt.Errorf(
			"%w: das Token %d gibt es nicht", domain.ErrNotFound, tokenID)
	}

	if fehler != nil {
		return domain.Zugriffstoken{}, domain.Gehuetetes{}, uebersetze(fehler,
			"das Token %d lesen", tokenID)
	}

	eines.Scope = domain.Scope(scope)

	return eines, gehuetet, nil
}

// TokenAendern schreibt Bezeichnung, Scope und Ablauf fort.
//
// **Das Geheimnis bleibt.** Wer ein anderes will, legt ein neues Token an —
// ein Schlüssel, der still seinen Bart wechselt, sperrt jeden aus, der ihn
// hat, ohne dass es jemand veranlasst hätte.
func (a *Zugriffstokenablage) TokenAendern(
	ctx context.Context, geaendert domain.Zugriffstoken,
) (domain.Zugriffstoken, error) {
	if err := geaendert.Validate(); err != nil {
		return domain.Zugriffstoken{}, err
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`UPDATE zugriffstoken
		    SET bezeichnung = ?, scope = ?, ablauf = ?, principal = ?,
		        fassung = fassung + 1
		  WHERE id = ? AND fassung = ?`,
		geaendert.Bezeichnung, string(geaendert.Scope), geaendert.Ablauf,
		geaendert.Principal, geaendert.ID, geaendert.Fassung)
	if err != nil {
		return domain.Zugriffstoken{}, uebersetze(err,
			"das Token %d ändern", geaendert.ID)
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return domain.Zugriffstoken{}, zaehlfehler
	}

	if zeilen == 0 {
		return domain.Zugriffstoken{}, a.tokenNichtGetroffen(ctx, geaendert.ID)
	}

	geaendert.Fassung++

	return geaendert, nil
}

// TokenLoeschen nimmt ein Token heraus.
//
// **Ohne Fassung und ohne Namensbestätigung.** Ein gelöschtes Token nimmt
// nichts mit: was damit geschrieben wurde, bleibt stehen, und wer sich
// vertippt, legt ein neues an. Das ist der Unterschied zum Löschen eines
// Nutzers, und deshalb steht dort ein Name davor.
func (a *Zugriffstokenablage) TokenLoeschen(ctx context.Context, tokenID int64) error {
	ergebnis, err := a.db.ExecContext(ctx,
		`DELETE FROM zugriffstoken WHERE id = ?`, tokenID)
	if err != nil {
		return uebersetze(err, "das Token %d löschen", tokenID)
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf("%w: das Token %d gibt es nicht",
			domain.ErrNotFound, tokenID)
	}

	return nil
}

// tokenNichtGetroffen sagt, **warum** ein Schreiben null Zeilen traf.
//
// **Dieselbe Unterscheidung wie beim Nutzer**: „gibt es nicht" heißt, du hast
// das falsche Token; „hat sich geändert" heißt, lies neu (§8).
func (a *Zugriffstokenablage) tokenNichtGetroffen(
	ctx context.Context, tokenID int64,
) error {
	var zahl int

	if err := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM zugriffstoken WHERE id = ?`, tokenID,
	).Scan(&zahl); err != nil {
		return uebersetze(err, "das Token %d nachschlagen", tokenID)
	}

	if zahl == 0 {
		return fmt.Errorf("%w: das Token %d gibt es nicht",
			domain.ErrNotFound, tokenID)
	}

	return fmt.Errorf("%w — das Token %d hat sich seit dem Lesen geändert",
		domain.ErrUeberholt, tokenID)
}

// scanneToken liest eine Zeile in ein Token.
func scanneToken(zeilen interface{ Scan(...any) error }) (
	domain.Zugriffstoken, error,
) {
	var (
		eines domain.Zugriffstoken
		scope string
	)

	if err := zeilen.Scan(&eines.ID, &eines.ProfilID, &eines.Bezeichnung,
		&scope, &eines.Ablauf, &eines.Angelegt, &eines.Fassung, &eines.Principal,
		&eines.LetzteNutzung, &eines.VorgaengerBis,
		&eines.Sitzbindung.Partie, &eines.Sitzbindung.Sitz); err != nil {
		return domain.Zugriffstoken{}, err
	}

	eines.Scope = domain.Scope(scope)

	return eines, nil
}
