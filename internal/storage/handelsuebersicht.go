package storage

// **Die Liste der Handelsvorgänge und ihre Entwürfe** — TRADE-01
// (#1715).
//
// # Warum der Entwurf hier geprüft wird und nicht nur im Dienst
//
// **Weil zwei Prozesse schreiben.** Das Fenster könnte einen Entwurf
// löschen, während der MCP-Server ihn gerade anbietet; ein Lesen im Dienst
// und ein Löschen danach ließen das Angebot verschwinden, das schon
// hinausging. Die Bedingung steht deshalb **in** der Anweisung — was der
// Dienst vorher prüft, ist für den Satz, nicht für die Sicherheit.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// letzteRevision ist die Bedingung „diese Zeile ist die letzte ihres
// Vorgangs" — an einer Stelle, weil drei Anweisungen sie brauchen.
const letzteRevision = `h.revision = (SELECT MAX(revision) FROM handelsrevisionen
	                     WHERE profil_id = h.profil_id AND vorgang = h.vorgang)`

// nurEntwurf ist die Bedingung „der Vorgang ist ein Entwurf": seine erste
// Revision steht auf `entwurf`, und eine zweite gibt es nicht.
const nurEntwurf = `revision = 1 AND zustand = 'entwurf'
	AND NOT EXISTS (SELECT 1 FROM handelsrevisionen AS n
	                 WHERE n.profil_id = handelsrevisionen.profil_id
	                   AND n.vorgang = handelsrevisionen.vorgang AND n.revision > 1)`

// Uebersicht gibt je Vorgang die letzte Revision — mit dem Anfang des
// Vorgangs, der letzten Änderung, den Namen seiner Karten und dem Archiv.
//
// **Die Namen kommen über die Scryfall-Kennung und nicht über `DruckID`**:
// die DruckID eines empfangenen Angebots ist eine Zeile im Bestand des
// Gegenübers, und hier träfe sie einen anderen Druck.
//
// **`CROSS JOIN` hält die Reihenfolge fest**: erst die Positionen, dann je
// Position die Kennung über `druckkennungen_wert (system, wert)`. Ohne das
// wählte SQLite die Kennungen zuerst und las je Vorgang **alle**
// Scryfall-Kennungen des Bestands — kein Vollscan im Sinn des Zählers,
// aber ein Bereich so groß wie die Sammlung (gemessen am 2026-09-30,
// `ABFRAGEPLAENE.md`).
//
// **Eine Abfrage und nicht zwei**: eine zweite nach einer geglückten
// ersten scheitert an einer echten Datenbank nicht — ihr Fehlerzweig wäre
// eine Zusage, die niemand betritt (§10).
func (a *Handelsablage) Uebersicht(ctx context.Context) ([]domain.Handelsvorgangzeile, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, fehler := a.db.QueryContext(ctx, `
		SELECT h.angebot,
		       COALESCE((SELECT angelegt FROM handelsrevisionen
		                  WHERE profil_id = h.profil_id AND vorgang = h.vorgang
		                    AND revision = 1), h.angelegt),
		       h.geaendert,
		       EXISTS (SELECT 1 FROM handelsarchiv
		                WHERE profil_id = h.profil_id AND vorgang = h.vorgang),
		       (SELECT json_group_array(name) FROM (
		          SELECT DISTINCT d.name
		            FROM json_each(h.angebot, '$.Positionen') AS p
		            CROSS JOIN druckkennungen AS k
		              ON k.system = 'scryfall' AND k.semantik = 'druck'
		             AND k.wert = json_extract(p.value, '$.Printing')
		            JOIN drucke AS d ON d.id = k.druck_id
		           WHERE d.name <> ''
		           ORDER BY d.name))
		  FROM handelsrevisionen AS h
		 WHERE h.profil_id = ? AND `+letzteRevision+`
		 ORDER BY h.geaendert DESC, h.vorgang`, profilID)
	if fehler != nil {
		return nil, fmt.Errorf("die Übersicht der Handelsvorgänge ließ sich nicht lesen: %w", fehler)
	}

	return uebersichtLesen(zeilen)
}

// uebersichtLesen liest die Zeilen der Übersicht.
func uebersichtLesen(zeilen *sql.Rows) ([]domain.Handelsvorgangzeile, error) {
	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler steht unten

	liste := []domain.Handelsvorgangzeile{}

	for zeilen.Next() {
		var roh, angelegt, geaendert, karten string

		var zeile domain.Handelsvorgangzeile

		if fehler := zeilen.Scan(&roh, &angelegt, &geaendert, &zeile.Archiviert, &karten); fehler != nil {
			return nil, fmt.Errorf("ein Handelsvorgang ließ sich nicht lesen: %w", fehler)
		}

		angebot, fehler := entpackeAngebot(roh)
		if fehler != nil {
			return nil, fehler
		}

		// **Eine unlesbare Zeit wird zur Nullzeit und nicht zum Fehler** —
		// dieselbe Abwägung wie im Kostenjournal: sie ist die Auskunft
		// „wann", nicht der Vertrag. **Und die Namen schreibt SQLite** als
		// JSON-Feld; daran scheitert `encoding/json` nicht.
		zeile.Angebot = angebot
		zeile.Angelegt, _ = time.Parse(time.DateTime, angelegt)   //nolint:errcheck // siehe oben
		zeile.Geaendert, _ = time.Parse(time.DateTime, geaendert) //nolint:errcheck // dito
		_ = json.Unmarshal([]byte(karten), &zeile.Karten)         //nolint:errcheck,errchkjson // dito

		liste = append(liste, zeile)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, fmt.Errorf("die Übersicht der Handelsvorgänge brach ab: %w", fehler)
	}

	return liste, nil
}

// EntwurfErsetzen schreibt einen Entwurf neu — **nur**, solange er einer
// ist.
func (a *Handelsablage) EntwurfErsetzen(ctx context.Context, angebot domain.Handelsangebot) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	roh, _ := json.Marshal(angebot) //nolint:errcheck,errchkjson // der Grund steht bei revisionSchreiben

	ergebnis, fehler := a.db.ExecContext(ctx, `
		UPDATE handelsrevisionen
		   SET abdruck = ?, fassung = ?, angebot = ?, geaendert = datetime('now')
		 WHERE profil_id = ? AND vorgang = ? AND `+nurEntwurf,
		angebot.Abdruck(), angebot.Vertragsfassung, string(roh), profilID, angebot.Vorgang)

	return a.nurEntwurfGetroffen(ctx, profilID, angebot.Vorgang, ergebnis, fehler, "ersetzen")
}

// EntwurfLoeschen löscht einen Entwurf — **nur**, solange er einer ist.
// Was an ihm hing, nimmt der Auslöser aus Wanderung 0093 mit.
func (a *Handelsablage) EntwurfLoeschen(ctx context.Context, vorgang string) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, fehler := a.db.ExecContext(ctx,
		`DELETE FROM handelsrevisionen WHERE profil_id = ? AND vorgang = ? AND `+nurEntwurf,
		profilID, vorgang)

	return a.nurEntwurfGetroffen(ctx, profilID, vorgang, ergebnis, fehler, "löschen")
}

// nurEntwurfGetroffen sagt, warum eine Anweisung an einem Entwurf keine
// Zeile traf: es gibt den Vorgang nicht, oder er ist keiner mehr.
func (a *Handelsablage) nurEntwurfGetroffen(
	ctx context.Context, profilID int64, vorgang string,
	ergebnis sql.Result, fehler error, was string,
) error {
	if fehler != nil {
		return fmt.Errorf("den Entwurf %s: %w", was, fehler)
	}

	if getroffen, _ := ergebnis.RowsAffected(); getroffen > 0 { //nolint:errcheck // SQLite zählt immer
		return nil
	}

	if fehler := a.vorgangDa(ctx, profilID, vorgang); fehler != nil {
		return fehler
	}

	return domain.MitGrund(domain.ErrConflict, nil,
		"der Handel ist kein Entwurf mehr — er ist hinausgegangen")
}

// Archivieren legt einen Vorgang ins Archiv oder holt ihn heraus.
func (a *Handelsablage) Archivieren(ctx context.Context, vorgang string, an bool) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	if fehler := a.vorgangDa(ctx, profilID, vorgang); fehler != nil {
		return fehler
	}

	anweisung := `DELETE FROM handelsarchiv WHERE profil_id = ? AND vorgang = ?`
	if an {
		anweisung = `INSERT INTO handelsarchiv (profil_id, vorgang) VALUES (?, ?)
		             ON CONFLICT DO NOTHING`
	}

	if _, fehler := a.db.ExecContext(ctx, anweisung, profilID, vorgang); fehler != nil {
		return fmt.Errorf("das Archiv ließ sich nicht schreiben: %w", fehler)
	}

	return nil
}

// vorgangDa sagt, ob es den Vorgang in diesem Profil gibt.
func (a *Handelsablage) vorgangDa(ctx context.Context, profilID int64, vorgang string) error {
	var da bool

	if fehler := a.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM handelsrevisionen WHERE profil_id = ? AND vorgang = ?)`,
		profilID, vorgang).Scan(&da); fehler != nil {
		return fmt.Errorf("der Handelsvorgang ließ sich nicht finden: %w", fehler)
	}

	if !da {
		return domain.MitGrund(domain.ErrNotFound, nil, "den Handelsvorgang gibt es nicht")
	}

	return nil
}
