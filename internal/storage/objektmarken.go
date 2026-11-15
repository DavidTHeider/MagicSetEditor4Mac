package storage

// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk 0135.
//
// # Zwei Tafeln, eine Frage
//
// Die Karte trägt ihre Marken in `karten_marken` (0024), jede andere Art in
// `objektmarken` (0084). **Wer umbenennt, zusammenführt oder zählt, meint
// beide** — hier stehen sie deshalb unter einer Ablage, und jeder Weg, der
// eine Marke als Ganzes betrifft, geht über beide in **einem** Zug.
//
// # Das Profil
//
// Jede Frage geht an das Profil aus dem Kontext. **Eine Karte gehört dem
// Profil ihres Sets** — deshalb der Weg über `cards` und `sets`, auch zum
// Lesen: eine Marke an einer fremden Karte ist eine Auskunft über ein
// fremdes Profil (UX-12.AK-04).

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Objektmarkenablage liest und schreibt Marken jeder Art.
type Objektmarkenablage struct{ db *sql.DB }

// NeueObjektmarkenablage baut sie über einer Datenbank.
func NeueObjektmarkenablage(db *sql.DB) *Objektmarkenablage { return &Objektmarkenablage{db: db} }

// kartenmarkenImProfil ist die Marke einer Karte samt deren Profil — für jede Frage an
// `karten_marken`.
//
// **Das Profil steht in der Verknüpfung und nicht dahinter** — als `?1`, damit
// keine Frage es vergessen kann.
const kartenmarkenImProfil = `karten_marken km JOIN cards c ON c.id = km.card_id
	JOIN sets s ON s.id = c.set_id AND s.profil_id = ?1`

// vorhandenFrage ist je Art die Frage, ob es das Objekt **im Profil** gibt.
// Der erste Platzhalter ist die Kennung, der zweite das Profil.
var vorhandenFrage = map[domain.Markenart]string{
	domain.MarkenartKarte: `SELECT EXISTS (SELECT 1 FROM cards c JOIN sets s ON s.id = c.set_id
	                         WHERE c.id = ? AND s.profil_id = ?)`,
	domain.MarkenartSet:      `SELECT EXISTS (SELECT 1 FROM sets WHERE id = ? AND profil_id = ?)`,
	domain.MarkenartExemplar: `SELECT EXISTS (SELECT 1 FROM sammlung WHERE id = ? AND profil_id = ?)`,
	domain.MarkenartDeck:     `SELECT EXISTS (SELECT 1 FROM decks WHERE id = ? AND profil_id = ?)`,
	domain.MarkenartTurnier:  `SELECT EXISTS (SELECT 1 FROM turniere WHERE id = ? AND profil_id = ?)`,
	domain.MarkenartPartie:   `SELECT EXISTS (SELECT 1 FROM partien WHERE id = ? AND profil_id = ?)`,
	domain.MarkenartHandel:   `SELECT EXISTS (SELECT 1 FROM handelsfolge WHERE vorgang = ? AND profil_id = ?)`,
	domain.MarkenartSymbol:   `SELECT EXISTS (SELECT 1 FROM symbolbibliothek WHERE id = ? AND profil_id = ?)`,

	// **Ein Paket gehört dem Rechner** — das Profil fragt hier niemand; es
	// steht nur da, damit jede Frage dieselben zwei Platzhalter hat.
	domain.MarkenartPaket: `SELECT EXISTS (SELECT 1 FROM packages WHERE art || '/' || name = ?) AND ? IS NOT NULL`,
}

// Vorhanden sagt, ob es das Objekt im Profil gibt.
func (a *Objektmarkenablage) Vorhanden(ctx context.Context, art domain.Markenart, kennung string) (bool, error) {
	return vorhanden(ctx, a.db, art, kennung)
}

func vorhanden(ctx context.Context, frager abfrager, art domain.Markenart, kennung string) (bool, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return false, fehler
	}

	var gibt bool
	if fehler := frager.QueryRowContext(ctx, vorhandenFrage[art], kennung, profilID).Scan(&gibt); fehler != nil {
		return false, uebersetze(fehler, "nachsehen, ob es %s %s gibt", art, kennung)
	}

	return gibt, nil
}

// Setzen hängt eine Marke an — zweimal dieselbe ist nichts.
func (a *Objektmarkenablage) Setzen(ctx context.Context, art domain.Markenart, kennung, marke string) error {
	return setzen(ctx, a.db, art, kennung, marke)
}

func setzen(ctx context.Context, frager abfrager, art domain.Markenart, kennung, marke string) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	if art == domain.MarkenartKarte {
		_, fehler = frager.ExecContext(ctx,
			`INSERT OR IGNORE INTO karten_marken (card_id, marke) VALUES (?, ?)`, kennung, marke)
	} else {
		_, fehler = frager.ExecContext(ctx,
			`INSERT OR IGNORE INTO objektmarken (profil_id, art, kennung, marke) VALUES (?, ?, ?, ?)`,
			profilID, string(art), kennung, marke)
	}

	return uebersetze(fehler, "die Marke setzen")
}

// SetzenAlle vergibt eine Marke an mehrere Objekte — **ganz oder gar
// nicht**: gibt es eines nicht, bekommt keines sie (UX-12.AK-02).
func (a *Objektmarkenablage) SetzenAlle(ctx context.Context, art domain.Markenart, kennungen []string, marke string) error {
	return imGeschaeft(ctx, a.db, "die Marke vergeben", func(geschaeft *sql.Tx) error {
		for _, kennung := range kennungen {
			gibt, fehler := vorhanden(ctx, geschaeft, art, kennung)
			if fehler != nil {
				return fehler
			}

			if !gibt {
				return fmt.Errorf("%w: %s %s gibt es in diesem Profil nicht", domain.ErrNotFound, art, kennung)
			}

			if fehler := setzen(ctx, geschaeft, art, kennung, marke); fehler != nil {
				return fehler
			}
		}

		return nil
	})
}

// Entfernen nimmt eine Marke von einem Objekt.
func (a *Objektmarkenablage) Entfernen(ctx context.Context, art domain.Markenart, kennung, marke string) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	if art == domain.MarkenartKarte {
		_, fehler = a.db.ExecContext(ctx,
			`DELETE FROM karten_marken WHERE card_id = ? AND marke = ?
			    AND card_id IN (SELECT c.id FROM cards c JOIN sets s ON s.id = c.set_id WHERE s.profil_id = ?)`,
			kennung, marke, profilID)
	} else {
		_, fehler = a.db.ExecContext(ctx,
			`DELETE FROM objektmarken WHERE profil_id = ? AND art = ? AND kennung = ? AND marke = ?`,
			profilID, string(art), kennung, marke)
	}

	return uebersetze(fehler, "die Marke entfernen")
}

// An gibt die Marken eines Objekts, geordnet.
func (a *Objektmarkenablage) An(ctx context.Context, art domain.Markenart, kennung string) ([]string, error) {
	je, fehler := a.jeObjekt(ctx, art, kennung)

	return append([]string{}, je[kennung]...), fehler
}

// JeObjekt gibt die Marken jedes Objekts einer Art — nach Kennung.
func (a *Objektmarkenablage) JeObjekt(ctx context.Context, art domain.Markenart) (map[string][]string, error) {
	return a.jeObjekt(ctx, art, "")
}

// jeObjekt liest die Marken einer Art, auf Wunsch nur eines Objekts.
func (a *Objektmarkenablage) jeObjekt(ctx context.Context, art domain.Markenart, nur string) (map[string][]string, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	frage := `SELECT kennung, marke FROM objektmarken
	           WHERE profil_id = ? AND art = ? AND (? = '' OR kennung = ?) ORDER BY kennung, marke`
	argumente := []any{profilID, string(art), nur, nur}

	if art == domain.MarkenartKarte {
		frage = `SELECT CAST(km.card_id AS TEXT), km.marke FROM ` + kartenmarkenImProfil + `
		          WHERE (?2 = '' OR km.card_id = ?2) ORDER BY km.card_id, km.marke`
		argumente = []any{profilID, nur}
	}

	return a.markenzeilen(ctx, frage, argumente...)
}

// markenzeilen liest Kennung und Marke je Zeile — herausgelöst, damit ein
// Lauf den `Scan` gegen eine falsche Spaltenzahl prüfen kann.
func (a *Objektmarkenablage) markenzeilen(ctx context.Context, frage string, argumente ...any) (map[string][]string, error) {
	zeilen, fehler := a.db.QueryContext(ctx, frage, argumente...)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Marken lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur gelesen

	je := map[string][]string{}

	for zeilen.Next() {
		var kennung, marke string
		if fehler := zeilen.Scan(&kennung, &marke); fehler != nil {
			return nil, uebersetze(fehler, "eine Marke lesen")
		}

		je[kennung] = append(je[kennung], marke)
	}

	return je, uebersetze(zeilen.Err(), "die Marken lesen")
}

// profilmarken ist jede Marke des Profils mit dem Objekt, das sie trägt —
// über beide Tafeln.
const profilmarken = `SELECT marke FROM objektmarken WHERE profil_id = ?1
	UNION ALL SELECT km.marke FROM ` + kartenmarkenImProfil

// Katalog zählt die Marken des Profils — über alle Arten.
func (a *Objektmarkenablage) Katalog(ctx context.Context) ([]domain.Markenposten, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	return a.posten(ctx, `SELECT marke, count(*) FROM (`+profilmarken+`) GROUP BY marke ORDER BY marke`, profilID)
}

// posten liest Marke und Zahl je Zeile — herausgelöst wie [markenzeilen].
func (a *Objektmarkenablage) posten(ctx context.Context, frage string, argumente ...any) ([]domain.Markenposten, error) {
	zeilen, fehler := a.db.QueryContext(ctx, frage, argumente...)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Marken zählen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur gelesen

	katalog := []domain.Markenposten{}

	for zeilen.Next() {
		var posten domain.Markenposten
		if fehler := zeilen.Scan(&posten.Marke, &posten.Anzahl); fehler != nil {
			return nil, uebersetze(fehler, "eine Marke zählen")
		}

		katalog = append(katalog, posten)
	}

	return katalog, uebersetze(zeilen.Err(), "die Marken zählen")
}

// Umbenennen gibt einer Marke im Profil einen neuen Namen — auf einen, den
// es schon gibt, nur mit `zusammenfuehren` (UX-12.AK-02). Die Zahl ist die
// der Objekte, die die alte trugen.
func (a *Objektmarkenablage) Umbenennen(ctx context.Context, alt, neu string, zusammenfuehren bool) (int, error) {
	var zahl int

	fehler := imGeschaeft(ctx, a.db, "die Marke umbenennen", func(geschaeft *sql.Tx) error {
		profilID, profilfehler := ProfilAus(ctx)
		if profilfehler != nil {
			return profilfehler
		}

		neueZahl, zaehlfehler := markenZaehlen(ctx, geschaeft, profilID, neu)
		if zaehlfehler != nil {
			return zaehlfehler
		}

		if neueZahl > 0 && !zusammenfuehren {
			return fmt.Errorf("%w: »%s« gibt es schon — umbenennen hieße zusammenführen", domain.ErrConflict, neu)
		}

		for _, anweisung := range []string{
			`INSERT OR IGNORE INTO objektmarken (profil_id, art, kennung, marke)
			 SELECT profil_id, art, kennung, ?2 FROM objektmarken WHERE profil_id = ?1 AND marke = ?3`,
			`INSERT OR IGNORE INTO karten_marken (card_id, marke)
			 SELECT km.card_id, ?2 FROM ` + kartenmarkenImProfil + ` WHERE km.marke = ?3`,
		} {
			if _, schreibfehler := geschaeft.ExecContext(ctx, anweisung, profilID, neu, alt); schreibfehler != nil {
				return uebersetze(schreibfehler, "die Marke umbenennen")
			}
		}

		var entfernfehler error
		zahl, entfernfehler = entfernenUeberall(ctx, geschaeft, profilID, alt)

		return entfernfehler
	})

	return zahl, fehler
}

// UeberallEntfernen nimmt eine Marke von jedem Objekt des Profils.
func (a *Objektmarkenablage) UeberallEntfernen(ctx context.Context, marke string) (int, error) {
	var zahl int

	fehler := imGeschaeft(ctx, a.db, "die Marke entfernen", func(geschaeft *sql.Tx) error {
		profilID, profilfehler := ProfilAus(ctx)
		if profilfehler != nil {
			return profilfehler
		}

		var entfernfehler error
		zahl, entfernfehler = entfernenUeberall(ctx, geschaeft, profilID, marke)

		return entfernfehler
	})

	return zahl, fehler
}

// entfernenUeberall zählt und entfernt eine Marke aus beiden Tafeln.
func entfernenUeberall(ctx context.Context, geschaeft *sql.Tx, profilID int64, marke string) (int, error) {
	zahl, fehler := markenZaehlen(ctx, geschaeft, profilID, marke)
	if fehler != nil {
		return 0, fehler
	}

	for _, anweisung := range []string{
		`DELETE FROM objektmarken WHERE profil_id = ?1 AND marke = ?2`,
		`DELETE FROM karten_marken WHERE marke = ?2
		    AND card_id IN (SELECT c.id FROM cards c JOIN sets s ON s.id = c.set_id WHERE s.profil_id = ?1)`,
	} {
		if _, schreibfehler := geschaeft.ExecContext(ctx, anweisung, profilID, marke); schreibfehler != nil {
			return 0, uebersetze(schreibfehler, "die Marke entfernen")
		}
	}

	return zahl, nil
}

// markenZaehlen sagt, wie viele Objekte des Profils eine Marke tragen.
func markenZaehlen(ctx context.Context, frager abfrager, profilID int64, marke string) (int, error) {
	var zahl int
	if fehler := frager.QueryRowContext(ctx,
		`SELECT count(*) FROM (`+profilmarken+`) WHERE marke = ?2`, profilID, marke).Scan(&zahl); fehler != nil {
		return 0, uebersetze(fehler, "die Marke »%s« zählen", marke)
	}

	return zahl, nil
}
